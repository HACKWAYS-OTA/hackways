package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/testsupport"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Mandatory CI uses a disposable _test database; no application fixtures run here.
func TestAdminDetailActions(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	sm := auth.NewSessionManager(q)
	user := func(name string) db.CreateUserRow {
		t.Helper()
		u, e := q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("admin-actions-%s-%d@example.invalid", name, time.Now().UnixNano()), DisplayName: name, PasswordHash: "isolated-ci-only"})
		if e != nil {
			t.Fatal(e)
		}
		return u
	}
	actor, owner, guest := user("Admin actor"), user("Resource owner"), user("Registration guest")
	var fixtureEventID, fixtureSpaceID pgtype.UUID
	defer func() {
		for _, id := range []pgtype.UUID{fixtureEventID, fixtureSpaceID} {
			if id.Valid {
				_, _ = pool.Exec(ctx, "DELETE FROM idempotency_records WHERE response_json::text LIKE '%' || $1::text || '%'", uuidString(id))
			}
		}
		_, _ = pool.Exec(ctx, "DELETE FROM audit_log WHERE actor_user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE event_id IN (SELECT id FROM events WHERE organizer_id IN (SELECT id FROM organizers WHERE user_id=$1))", owner.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM events WHERE organizer_id IN (SELECT id FROM organizers WHERE user_id=$1)", owner.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM spaces WHERE organizer_id IN (SELECT id FROM organizers WHERE user_id=$1)", owner.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM organizers WHERE user_id=$1", owner.ID)
		for _, id := range []pgtype.UUID{actor.ID, owner.ID, guest.ID} {
			_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", id)
			_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", id)
			_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
		}
	}()
	if _, err = pool.Exec(ctx, "UPDATE users SET email_verified=true WHERE id=$1", actor.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES($1)", actor.ID); err != nil {
		t.Fatal(err)
	}
	token, session, err := sm.CreateSession(ctx, actor.ID, "CI", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err = q.RecordPlatformAdminSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	org, err := q.CreateOrganizer(ctx, db.CreateOrganizerParams{UserID: owner.ID, DisplayName: owner.DisplayName})
	if err != nil {
		t.Fatal(err)
	}
	event, err := q.CreateEventRecord(ctx, db.CreateEventRecordParams{OrganizerID: org.ID, Status: db.EventStatusPublished, Format: db.EventFormatInPerson, Title: "Admin action CI", Description: "Isolated regression fixture", StartsAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}, EndsAt: pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true}, Timezone: "UTC", Capacity: pgtype.Int4{Int32: 1, Valid: true}, LocationName: pgtype.Text{String: "CI venue", Valid: true}})
	if err != nil {
		t.Fatal(err)
	}
	fixtureEventID = event.ID
	space, err := q.CreateSpace(ctx, db.CreateSpaceParams{OrganizerID: org.ID, Name: "Admin CI Space", Description: "CI only", CreationKey: fmt.Sprintf("admin-ci-%d", time.Now().UnixNano()), CreationHash: "isolated"})
	if err != nil {
		t.Fatal(err)
	}
	fixtureSpaceID = space.ID
	reg, _, err := workflows.New(pool).Register(ctx, workflows.Request{EventID: event.ID, UserID: guest.ID, Name: guest.DisplayName, Key: "register-admin-ci-key", RequestID: "CI"})
	if err != nil {
		t.Fatal(err)
	}
	router := NewHandler(pool).Routes()
	request := func(method, path, body, key, origin, reason, cookie string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Origin", origin)
		r.Header.Set("X-Admin-Reason", reason)
		r.Header.Set("Idempotency-Key", key)
		if cookie != "" {
			r.AddCookie(&http.Cookie{Name: cookie, Value: token})
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	call := func(method, path, body, key string) *httptest.ResponseRecorder {
		return request(method, path, body, key, "http://localhost:3003", "Review%20requested%20by%20support", auth.AdminSessionCookie)
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("expected %d got %d: %s", status, w.Code, w.Body.String())
		}
	}
	eventPath := "/events/" + uuidString(event.ID)
	edit := fmt.Sprintf(`{"title":"Reviewed event","description":"Reviewed description","capacity":2,"visibility":"private","revision":%d}`, event.Revision)
	check(request("PUT", eventPath, edit, "admin-event-edit-key", "https://attacker.invalid", "reason", auth.AdminSessionCookie), 403)
	check(request("PUT", eventPath, edit, "admin-event-edit-key", "http://localhost:3003", "", auth.AdminSessionCookie), 422)
	check(request("PUT", eventPath, edit, "admin-event-edit-key", "http://localhost:3003", "reason", auth.DefaultSessionCookie), 401)
	check(call("PUT", eventPath, edit, "admin-event-edit-key"), 200)
	var queuedUpdates int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox WHERE event_id=$1 AND template_type='event_updated'", event.ID).Scan(&queuedUpdates); err != nil || queuedUpdates != 1 {
		t.Fatalf("published update notification missing: %d %v", queuedUpdates, err)
	}

	w := call("PUT", eventPath, edit, "admin-event-edit-key")
	check(w, 200)
	if w.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("edit retry was not replayed")
	}
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox WHERE event_id=$1 AND template_type='event_updated'", event.ID).Scan(&queuedUpdates); err != nil || queuedUpdates != 1 {
		t.Fatalf("retry duplicated published update notification: %d %v", queuedUpdates, err)
	}
	for _, path := range []string{eventPath + "?section=activity", "/moderation/" + uuidString(event.ID), "/spaces/" + uuidString(space.ID), "/registrations/" + uuidString(reg.Data.ID)} {
		detail := call("GET", path, "", "")
		check(detail, 200)
		for _, forbidden := range []string{"password_hash", "creation_key", "creation_hash", "token_hash", "payload_json"} {
			if strings.Contains(detail.Body.String(), forbidden) {
				t.Fatalf("detail exposed %s", forbidden)
			}
		}
	}

	check(call("PUT", eventPath, strings.Replace(edit, "Reviewed event", "Changed payload", 1), "admin-event-edit-key"), 409)
	check(call("PUT", eventPath, edit, "different-edit-request"), 409)
	check(call("GET", eventPath+"?section=activity", "", ""), 200)
	var auditID int64
	if err = pool.QueryRow(ctx, "SELECT max(id) FROM audit_log WHERE actor_user_id=$1", actor.ID).Scan(&auditID); err != nil {
		t.Fatal(err)
	}
	check(call("GET", fmt.Sprintf("/audit/%d", auditID), "", ""), 200)
	var auditCount int
	var metadata string
	if err = pool.QueryRow(ctx, "SELECT count(*),min(metadata::text) FROM audit_log WHERE actor_user_id=$1 AND entity_id=$2 AND action='admin.event.updated'", actor.ID, event.ID).Scan(&auditCount, &metadata); err != nil || auditCount != 1 || !strings.Contains(metadata, "Review requested by support") {
		t.Fatalf("audit actor/reason or exact once failed: %d %v", auditCount, err)
	}
	current, err := q.GetEventByID(ctx, event.ID)
	if err != nil {
		t.Fatal(err)
	}
	low := fmt.Sprintf(`{"title":"Reviewed event","description":"Reviewed description","capacity":0,"visibility":"private","revision":%d}`, current.Revision)
	check(call("PUT", eventPath, low, "capacity-zero-request"), 422)
	second, _, err := workflows.New(pool).Register(ctx, workflows.Request{EventID: event.ID, UserID: owner.ID, Name: owner.DisplayName, Key: "second-capacity-admin-ci", RequestID: "CI"})
	if err != nil {
		t.Fatal(err)
	}
	current, _ = q.GetEventByID(ctx, event.ID)
	reserved := fmt.Sprintf(`{"title":"Reviewed event","description":"Reviewed description","capacity":1,"visibility":"private","revision":%d}`, current.Revision)
	check(call("PUT", eventPath, reserved, "reserved-capacity-request"), 409)
	secondCancel := fmt.Sprintf(`{"status":"cancelled","updated_at":%q,"notify_guest":false,"message":""}`, second.Data.UpdatedAt.Time.Format(time.RFC3339Nano))
	check(call("POST", "/registrations/"+uuidString(second.Data.ID)+"/status", secondCancel, "second-cancel-admin-ci"), 200)

	regPath := "/registrations/" + uuidString(reg.Data.ID) + "/status"
	cancel := fmt.Sprintf(`{"status":"cancelled","updated_at":%q,"notify_guest":false,"message":""}`, reg.Data.UpdatedAt.Time.Format(time.RFC3339Nano))
	check(call("POST", regPath, cancel, "admin-reg-cancel-key"), 200)
	w = call("POST", regPath, cancel, "admin-reg-cancel-key")
	check(w, 200)
	if w.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("registration retry not replayed")
	}
	current, _ = q.GetEventByID(ctx, event.ID)
	if current.CurrentRsvpCount != 0 {
		t.Fatal("admin cancellation did not release capacity")
	}
	regrow, err := q.GetWorkflowRSVP(ctx, reg.Data.ID)
	if err != nil {
		t.Fatal(err)
	}
	confirm := fmt.Sprintf(`{"status":"confirmed","updated_at":%q,"notify_guest":false,"message":""}`, regrow.UpdatedAt.Time.Format(time.RFC3339Nano))
	check(call("POST", regPath, confirm, "admin-reg-confirm-key"), 200)
	ticket, err := q.GetTicketByRSVP(ctx, reg.Data.ID)
	if err != nil {
		t.Fatal(err)
	}
	check(call("GET", "/tickets/"+uuidString(ticket.ID), "", ""), 200)
	ticketPath := "/tickets/" + uuidString(ticket.ID) + "/check-in"
	check(call("POST", ticketPath, `{}`, "admin-ticket-scan-key"), 200)
	check(call("POST", ticketPath, `{}`, "admin-ticket-scan-key"), 200)
	scannedTicket, scanErr := q.GetWorkflowTicket(ctx, ticket.ID)
	if scanErr != nil || scannedTicket.CheckedInBy != actor.ID {
		t.Fatal("check-in must record real administrator")
	}
	spacePath := "/spaces/" + uuidString(space.ID)
	spaceEdit := fmt.Sprintf(`{"name":"Reviewed Space","description":"Updated description","guidelines":"Keep it kind","website_url":"https://example.invalid","revision":%d}`, space.Revision)
	check(call("PUT", spacePath, spaceEdit, "admin-space-edit-key"), 200)
	check(call("PUT", spacePath, spaceEdit, "admin-space-edit-key"), 200)
	space, err = q.LockSpace(ctx, space.ID)
	if err != nil {
		t.Fatal(err)
	}
	archive := fmt.Sprintf(`{"status":"archived","revision":%d}`, space.Revision)
	check(call("POST", spacePath+"/status", archive, "admin-space-archive-key"), 200)
	current, _ = q.GetEventByID(ctx, event.ID)
	cancelEvent := fmt.Sprintf(`{"status":"cancelled","revision":%d,"acknowledge_attendees":true}`, current.Revision)
	check(call("POST", eventPath+"/status", cancelEvent, "admin-event-cancel-key"), 200)
	check(call("POST", eventPath+"/status", cancelEvent, "admin-event-cancel-key"), 200)
	current, _ = q.GetEventByID(ctx, event.ID)
	if current.Status != db.EventStatusCancelled || current.CurrentRsvpCount != 0 {
		t.Fatal("event cancellation invariant failed")
	}
	if _, err = pool.Exec(ctx, "UPDATE platform_administrators SET revoked_at=clock_timestamp() WHERE user_id=$1", actor.ID); err != nil {
		t.Fatal(err)
	}
	check(call("GET", eventPath, "", ""), 403)
	check(call("POST", eventPath+"/status", cancelEvent, "admin-event-cancel-key"), 403)
	// Verify safe detail serialization never leaks authentication or creation secrets.
	var envelope struct{ Data json.RawMessage }
	if json.Unmarshal(w.Body.Bytes(), &envelope) != nil {
		t.Fatal("invalid mutation response")
	}
}
