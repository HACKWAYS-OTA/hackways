package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/cities"
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
		return request(method, path, body, key, "http://localhost:3003", "", auth.AdminSessionCookie)
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("expected %d got %d: %s", status, w.Code, w.Body.String())
		}
	}

	check(request("GET", "/analytics", "", "", "", "", ""), 401)
	for _, path := range []string{"/analytics?from=bad", "/analytics?from=2000-01-01", "/analytics?to=2999-01-01"} {
		check(call("GET", path, "", ""), 400)
	}
	analyticsRead := call("GET", "/analytics", "", "")
	check(analyticsRead, 200)
	if analyticsRead.Header().Get("Cache-Control") != "private, no-store" || !strings.Contains(analyticsRead.Body.String(), `"top_events"`) {
		t.Fatal("analytics output or cache policy invalid")
	}
	export := call("GET", "/analytics?format=csv", "", "")
	check(export, 200)
	if !strings.HasPrefix(export.Header().Get("Content-Type"), "text/csv") || len(strings.Split(strings.TrimSpace(export.Body.String()), "\n")) != 31 {
		t.Fatal("analytics CSV range invalid")
	}
	eventPath := "/events/" + strconv.FormatInt(event.AdminReference, 10)
	// Extend the mandatory isolated admin workflow with real City Hub persistence,
	// publication, curation, replay and private-event exclusion.
	cityAdmin, cityPublic := cities.NewHandler(pool).AdminRoutes(), cities.NewHandler(pool).Routes()
	cityCall := func(method, path, body, key, origin, cookie string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		r.Header.Set("Idempotency-Key", key)
		if cookie != "" {
			r.AddCookie(&http.Cookie{Name: cookie, Value: token})
		}
		out := httptest.NewRecorder()
		cityAdmin.ServeHTTP(out, r)
		return out
	}
	cityRead := func(path string) *httptest.ResponseRecorder {
		out := httptest.NewRecorder()
		cityPublic.ServeHTTP(out, httptest.NewRequest("GET", path, nil))
		return out
	}
	citySlug := fmt.Sprintf("ci-city-%d", time.Now().UnixNano())
	cityBody := fmt.Sprintf(`{"slug":%q,"name":"CI city","country":"CI country","timezone":"UTC","description":"Isolated city workflow","status":"draft"}`, citySlug)
	check(cityCall("POST", "/", cityBody, "city-create-ci-key", "http://localhost:3003", ""), 401)
	check(cityCall("POST", "/", cityBody, "city-create-ci-key", "http://localhost:3003", auth.DefaultSessionCookie), 401)
	check(cityCall("POST", "/", cityBody, "city-create-ci-key", "https://untrusted.invalid", auth.AdminSessionCookie), 403)
	cityResult := cityCall("POST", "/", cityBody, "city-create-ci-key", "http://localhost:3003", auth.AdminSessionCookie)
	check(cityResult, 200)
	var cityEnvelope struct {
		Data struct {
			ID        pgtype.UUID `json:"id"`
			Reference int64       `json:"reference"`
			Revision  int64       `json:"revision"`
		}
	}
	if json.Unmarshal(cityResult.Body.Bytes(), &cityEnvelope) != nil || !cityEnvelope.Data.ID.Valid {
		t.Fatal("city creation did not return persisted identity")
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM idempotency_records WHERE response_json::text LIKE '%' || $1::text || '%'", uuidString(cityEnvelope.Data.ID))
		_, _ = pool.Exec(ctx, "DELETE FROM city_hub_events WHERE city_id=$1", cityEnvelope.Data.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM city_hubs WHERE id=$1", cityEnvelope.Data.ID)
	}()
	if strings.Contains(cityResult.Body.String(), "creation_key") || strings.Contains(cityResult.Body.String(), "creation_hash") {
		t.Fatal("city secrets leaked")
	}
	replay := cityCall("POST", "/", cityBody, "city-create-ci-key", "http://localhost:3003", auth.AdminSessionCookie)
	check(replay, 200)
	if replay.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("city creation retry did not replay")
	}
	check(cityRead("/by-slug/"+citySlug), 404)
	cityPath := "/" + strconv.FormatInt(cityEnvelope.Data.Reference, 10)
	publish := `{"name":"CI city","country":"CI country","timezone":"UTC","description":"Isolated city workflow","status":"published","revision":1}`
	check(cityCall("PUT", cityPath, strings.Replace(publish, "UTC", "Invalid/Zone", 1), "city-invalid-zone-key", "http://localhost:3003", auth.AdminSessionCookie), 422)
	check(cityCall("PUT", cityPath, publish, "city-publish-ci-key", "http://localhost:3003", auth.AdminSessionCookie), 200)
	check(cityCall("PUT", cityPath, publish, "city-stale-edit-key", "http://localhost:3003", auth.AdminSessionCookie), 409)
	check(cityRead("/by-slug/"+citySlug), 200)
	add := fmt.Sprintf(`{"event_id":%q,"revision":2}`, uuidString(event.ID))
	check(cityCall("POST", cityPath+"/events", add, "city-event-add-key", "http://localhost:3003", auth.AdminSessionCookie), 200)
	check(cityCall("POST", cityPath+"/events", add, "city-event-add-key", "http://localhost:3003", auth.AdminSessionCookie), 200)
	feedPath := "/" + uuidString(cityEnvelope.Data.ID) + "/events"
	if feed := cityRead(feedPath); feed.Code != 200 || !strings.Contains(feed.Body.String(), event.Slug) {
		t.Fatal("curated public event missing")
	}
	if _, err = pool.Exec(ctx, "UPDATE events SET visibility='private' WHERE id=$1", event.ID); err != nil {
		t.Fatal(err)
	}
	if feed := cityRead(feedPath); feed.Code != 200 || strings.Contains(feed.Body.String(), event.Slug) {
		t.Fatal("private event leaked into city feed")
	}
	if _, err = pool.Exec(ctx, "UPDATE events SET visibility='public' WHERE id=$1", event.ID); err != nil {
		t.Fatal(err)
	}
	remove := fmt.Sprintf(`{"event_id":%q,"revision":3,"remove":true}`, uuidString(event.ID))
	check(cityCall("POST", cityPath+"/events", remove, "city-event-remove-key", "http://localhost:3003", auth.AdminSessionCookie), 200)
	archiveCity := `{"name":"CI city","country":"CI country","timezone":"UTC","description":"Isolated city workflow","status":"archived","revision":4}`
	check(cityCall("PUT", cityPath, archiveCity, "city-archive-ci-key", "http://localhost:3003", auth.AdminSessionCookie), 200)
	check(cityRead("/by-slug/"+citySlug), 404)
	edit := fmt.Sprintf(`{"title":"Reviewed event","description":"Reviewed description","capacity":2,"visibility":"private","revision":%d}`, event.Revision)
	check(request("PUT", eventPath, edit, "admin-event-edit-key", "https://attacker.invalid", "reason", auth.AdminSessionCookie), 403)
	check(request("PUT", eventPath, edit, "admin-event-edit-key", "http://localhost:3003", "x", auth.AdminSessionCookie), 422)
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
	for _, path := range []string{eventPath + "?section=activity", "/moderation/" + strconv.FormatInt(event.AdminReference, 10), "/spaces/" + strconv.FormatInt(space.AdminReference, 10), "/registrations/" + strconv.FormatInt(reg.Data.AdminReference, 10)} {
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
	check(call("GET", "/events/"+uuidString(event.ID), "", ""), 200) // existing API callers remain compatible
	check(call("GET", "/events/0", "", ""), 400)
	check(call("GET", "/events/9223372036854775807", "", ""), 404)
	var detailEnvelope struct {
		Data struct {
			Reference         int64 `json:"reference"`
			HostUserReference int64 `json:"host_user_reference"`
		}
		Related struct {
			Rows []struct {
				Reference string `json:"reference"`
			}
		}
	}
	eventDetail := call("GET", eventPath, "", "")
	check(eventDetail, 200)
	if err = json.Unmarshal(eventDetail.Body.Bytes(), &detailEnvelope); err != nil || detailEnvelope.Data.Reference != event.AdminReference || detailEnvelope.Data.HostUserReference < 1 {
		t.Fatalf("numeric detail references missing: %v", err)
	}
	if len(detailEnvelope.Related.Rows) != 1 {
		t.Fatal("expected related registration")
	}
	if n, e := strconv.ParseInt(detailEnvelope.Related.Rows[0].Reference, 10, 64); e != nil || n < 1 {
		t.Fatal("related registration must expose numeric reference")
	}
	var listEnvelope struct {
		Data []struct {
			ID        string `json:"id"`
			Reference int64  `json:"reference"`
		}
	}
	eventList := call("GET", "/events?q=Reviewed%20event", "", "")
	check(eventList, 200)
	if err = json.Unmarshal(eventList.Body.Bytes(), &listEnvelope); err != nil {
		t.Fatal(err)
	}
	foundReference := false
	for _, row := range listEnvelope.Data {
		if row.ID == uuidString(event.ID) && row.Reference == event.AdminReference {
			foundReference = true
		}
	}
	if !foundReference {
		t.Fatal("numeric list reference missing for persisted fixture")
	}

	var auditID int64
	if err = pool.QueryRow(ctx, "SELECT max(id) FROM audit_log WHERE actor_user_id=$1", actor.ID).Scan(&auditID); err != nil {
		t.Fatal(err)
	}
	check(call("GET", fmt.Sprintf("/audit/%d", auditID), "", ""), 200)
	var auditCount int
	var metadata string
	if err = pool.QueryRow(ctx, "SELECT count(*),min(metadata::text) FROM audit_log WHERE actor_user_id=$1 AND entity_id=$2 AND action='admin.event.updated'", actor.ID, event.ID).Scan(&auditCount, &metadata); err != nil || auditCount != 1 || !strings.Contains(metadata, "Administrative change") {
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
	check(call("POST", "/registrations/"+strconv.FormatInt(second.Data.AdminReference, 10)+"/status", secondCancel, "second-cancel-admin-ci"), 200)

	regPath := "/registrations/" + strconv.FormatInt(reg.Data.AdminReference, 10) + "/status"
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
	check(call("GET", "/tickets/"+strconv.FormatInt(ticket.AdminReference, 10), "", ""), 200)
	ticketPath := "/tickets/" + strconv.FormatInt(ticket.AdminReference, 10) + "/check-in"
	check(call("POST", ticketPath, `{}`, "admin-ticket-scan-key"), 200)
	check(call("POST", ticketPath, `{}`, "admin-ticket-scan-key"), 200)
	scannedTicket, scanErr := q.GetWorkflowTicket(ctx, ticket.ID)
	if scanErr != nil || scannedTicket.CheckedInBy != actor.ID {
		t.Fatal("check-in must record real administrator")
	}
	spacePath := "/spaces/" + strconv.FormatInt(space.AdminReference, 10)
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
	if current.Status != db.EventStatusCancelled || current.CurrentRsvpCount != 0 || !current.ManagementClosed {
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
