package events

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/registrationforms"
	"github.com/hackways/core-backend/internal/rsvps"
	"github.com/hackways/core-backend/internal/spaces"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ownerFixture struct {
	pool      *pgxpool.Pool
	q         *db.Queries
	owner     db.CreateUserRow
	organizer db.Organizers
	event     db.Events
	cookie    *http.Cookie
}

func TestEventInvitations_OwnershipIdempotencyAndInventory(t *testing.T) {
	f, other := ownerSetup(t), ownerSetup(t)
	ctx := context.Background()
	if _, err := f.pool.Exec(ctx, "UPDATE users SET email_verified=TRUE WHERE id=$1", f.owner.ID); err != nil {
		t.Fatal(err)
	}
	api := NewHandler(f.pool).Routes()
	path := "/" + idString(f.event.ID) + "/invitations"
	request := func(method, body, key string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", key)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		return w
	}
	body := `{"emails":["invitee@example.invalid","INVITEE@example.invalid"],"message":"Welcome aboard"}`
	key := "invitation-integration-test-key"
	if w := request("POST", body, key, nil); w.Code != 401 {
		t.Fatalf("anonymous invitation: %d", w.Code)
	}
	if w := request("POST", body, key, other.cookie); w.Code != 404 {
		t.Fatalf("cross-owner invitation: %d", w.Code)
	}
	w := request("POST", body, key, f.cookie)
	if w.Code != 200 {
		t.Fatalf("invite: %d %s", w.Code, w.Body.String())
	}
	var result workflows.InvitationResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Queued != 1 || result.Skipped != 0 {
		t.Fatal("unexpected invitation result")
	}
	w = request("POST", body, key, f.cookie)
	if w.Code != 200 || w.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("invitation retry did not replay")
	}
	if w = request("POST", `{"emails":["different@example.invalid"]}`, key, f.cookie); w.Code != 409 {
		t.Fatal("changed retry body accepted")
	}
	w = request("POST", body, key+"-new", f.cookie)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Queued != 0 || result.Skipped != 1 {
		t.Fatal("duplicate invitation queued")
	}
	w = request("GET", "", "", f.cookie)
	var listed struct {
		Data []db.ListEventInvitationsRow `json:"data"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &listed) != nil || len(listed.Data) != 1 || listed.Data[0].RecipientEmail != "invitee@example.invalid" || listed.Data[0].Status != db.NotificationStatusPending {
		t.Fatal("invitation history missing persisted outbox state")
	}
	var count int
	var payload string
	if err := f.pool.QueryRow(ctx, "SELECT count(*),min(payload_json::text) FROM email_outbox WHERE event_id=$1 AND template_type='event_invitation'", f.event.ID).Scan(&count, &payload); err != nil || count != 1 || !strings.Contains(payload, "Welcome aboard") || !strings.Contains(payload, "/register") {
		t.Fatal("invitation context not persisted")
	}
	after, err := f.q.GetEventByID(ctx, f.event.ID)
	if err != nil || after.CurrentRsvpCount != f.event.CurrentRsvpCount || after.HeldRsvpCount != f.event.HeldRsvpCount {
		t.Fatal("invitation changed inventory")
	}
	if _, err = f.pool.Exec(ctx, "UPDATE events SET status='draft' WHERE id=$1", f.event.ID); err != nil {
		t.Fatal(err)
	}
	if w = request("POST", `{"emails":["closed@example.invalid"]}`, key+"-closed", f.cookie); w.Code != 409 {
		t.Fatal("draft invitation accepted")
	}
}

func ownerSetup(t *testing.T) ownerFixture {
	t.Helper()
	raw := os.Getenv("DATABASE_URL")
	if raw == "" {
		t.Skip("explicit isolated DATABASE_URL is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal("invalid isolated database configuration")
	}
	schema := parsed.Query().Get("search_path")
	if !strings.HasSuffix(parsed.Path, "_test") && !(strings.HasPrefix(schema, "workflow_") && strings.Contains(schema, "_ci_") && !strings.ContainsAny(schema, ", ")) {
		t.Fatal("refusing to create fixtures outside an isolated CI database/schema")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, raw)
	if err != nil {
		t.Fatal("isolated pool unavailable")
	}
	t.Cleanup(pool.Close)
	q := db.New(pool)
	owner, err := q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("owner-%d@example.invalid", time.Now().UnixNano()), DisplayName: "Isolated owner", PasswordHash: "not-a-login-credential"})
	if err != nil {
		t.Fatal(err)
	}
	organizer, err := q.CreateOrganizer(ctx, db.CreateOrganizerParams{UserID: owner.ID, DisplayName: "Isolated owner profile"})
	if err != nil {
		t.Fatal(err)
	}
	event, err := q.CreateEventRecord(ctx, db.CreateEventRecordParams{OrganizerID: organizer.ID, Status: db.EventStatusPublished, Format: db.EventFormatInPerson, Title: "Isolated owner workflow", Description: "CI workflow fixture", StartsAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}, EndsAt: pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true}, Timezone: "UTC", Capacity: pgtype.Int4{Int32: 2, Valid: true}, LocationName: pgtype.Text{String: "CI only", Valid: true}})
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := auth.NewSessionManager(q).CreateSession(ctx, owner.ID, "CI", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return ownerFixture{pool, q, owner, organizer, event, &http.Cookie{Name: auth.DefaultSessionCookie, Value: token}}
}
func idString(id pgtype.UUID) string {
	b := id.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Uses the existing isolated lifecycle fixtures; never the application schema.
func TestOwnerLifecycle_SpacesAndPublicInformation(t *testing.T) {
	f, other := ownerSetup(t), ownerSetup(t)
	spaceAPI, eventAPI := spaces.NewHandler(f.pool).Routes(), NewHandler(f.pool).Routes()
	body := []byte(`{"name":"Lifecycle community","description":"Isolated workflow community"}`)
	key := fmt.Sprintf("space-creation-%d", time.Now().UnixNano())
	create := func(payload []byte) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/", bytes.NewReader(payload))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", key)
		r.AddCookie(f.cookie)
		w := httptest.NewRecorder()
		spaceAPI.ServeHTTP(w, r)
		return w
	}
	w := create(body)
	if w.Code != 201 {
		t.Fatalf("create Space: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID       pgtype.UUID `json:"id"`
			Slug     string      `json:"slug"`
			Revision int64       `json:"revision"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if w = create(body); w.Code != 200 {
		t.Fatalf("Space replay: %d %s", w.Code, w.Body.String())
	}
	if w = create([]byte(`{"name":"Different name","description":"Isolated workflow community"}`)); w.Code != 409 {
		t.Fatalf("Space conflicting replay: %d", w.Code)
	}
	path := "/" + idString(created.Data.ID)
	if w = requestJSON(t, spaceAPI, "GET", "/mine?view=owned", nil, f.cookie); w.Code != 200 {
		t.Fatalf("owned Space route: %d %s", w.Code, w.Body.String())
	}
	if w = requestJSON(t, spaceAPI, "POST", path+"/join", nil, nil); w.Code != 401 {
		t.Fatalf("anonymous join: %d", w.Code)
	}
	for i := 0; i < 2; i++ {
		if w = requestJSON(t, spaceAPI, "POST", path+"/join", nil, other.cookie); w.Code != 200 {
			t.Fatalf("join/replay: %d %s", w.Code, w.Body.String())
		}
	}
	if w = requestJSON(t, spaceAPI, "GET", path+"/manage", nil, other.cookie); w.Code != 404 {
		t.Fatalf("member obtained owner rights: %d", w.Code)
	}
	if w = requestJSON(t, spaceAPI, "POST", path+"/leave", nil, f.cookie); w.Code != 409 {
		t.Fatalf("owner left Space: %d", w.Code)
	}
	assignment := []byte(fmt.Sprintf(`{"space_slug":%q,"revision":1}`, created.Data.Slug))
	if w = requestJSON(t, eventAPI, "PUT", "/"+idString(other.event.ID)+"/space", assignment, other.cookie); w.Code != 404 {
		t.Fatalf("member hosted under another owner's Space: %d %s", w.Code, w.Body.String())
	}
	if w = requestJSON(t, eventAPI, "PUT", "/"+idString(f.event.ID)+"/space", assignment, f.cookie); w.Code != 200 {
		t.Fatalf("owner association: %d %s", w.Code, w.Body.String())
	}
	newEvent := createEventRequest{Status: "published", Visibility: "private", Title: "Space event appearance", Description: "Isolated appearance persistence", Format: "in_person", StartsAt: time.Now().Add(3 * time.Hour).Format(time.RFC3339), EndsAt: time.Now().Add(4 * time.Hour).Format(time.RFC3339), Timezone: "UTC", Location: "CI only", SpaceSlug: created.Data.Slug, Theme: "ambient", ThemeColor: "#ced9da", ThemeFont: "google", ThemeStyle: "crisp", ThemeDisplay: "dark"}
	appearanceBody, _ := json.Marshal(newEvent)
	w = requestJSON(t, eventAPI, "POST", "/", appearanceBody, f.cookie)
	if w.Code != 201 {
		t.Fatalf("Space event creation: %d %s", w.Code, w.Body.String())
	}
	var themed struct {
		Data db.Events `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &themed); err != nil {
		t.Fatal(err)
	}
	if themed.Data.ThemeDisplay != "dark" || themed.Data.Theme != "ambient" || themed.Data.ThemeColor != "#ced9da" || themed.Data.SpaceID != created.Data.ID {
		t.Fatal("Space or appearance was not persisted")
	}
	if themed.Data.Visibility != "private" {
		t.Fatal("private visibility was not persisted")
	}
	if w = requestJSON(t, eventAPI, "GET", "/by-slug/"+themed.Data.Slug, nil, nil); w.Code != 200 {
		t.Fatalf("link-only event not accessible by link: %d", w.Code)
	}
	listed, err := f.q.ListSpaceEvents(context.Background(), db.ListSpaceEventsParams{SpaceID: created.Data.ID, PageLimit: 100})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range listed {
		if row.ID == themed.Data.ID {
			t.Fatal("private event leaked into Space feed")
		}
	}
	w = requestJSON(t, eventAPI, "GET", "/?q=Space%20event%20appearance&limit=100", nil, nil)
	if w.Code != 200 {
		t.Fatalf("discovery search: %d", w.Code)
	}
	var catalog struct {
		Data []db.Events `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	for _, row := range catalog.Data {
		if row.ID == themed.Data.ID {
			t.Fatal("private event leaked into Discover search")
		}
	}
	faq, contact, rewards := "Bring your laptop.", "https://example.invalid/contact", "Community recognition"
	update := updateRequest{eventContent: eventContent{FAQText: &faq, ContactURL: &contact, Rewards: &rewards}, Title: f.event.Title, Description: f.event.Description, Format: string(f.event.Format), StartsAt: f.event.StartsAt.Time, EndsAt: f.event.EndsAt.Time, Timezone: "UTC", Capacity: &f.event.Capacity.Int32, Location: f.event.LocationName.String, Mode: "instant", AllowGuest: true, OfferMinutes: 10, Revision: 2}
	encoded, _ := json.Marshal(update)
	if w = requestJSON(t, eventAPI, "PUT", "/"+idString(f.event.ID), encoded, f.cookie); w.Code != 200 {
		t.Fatalf("event information save: %d %s", w.Code, w.Body.String())
	}
	stored, err := f.q.GetPublishedEventBySlug(context.Background(), f.event.Slug)
	if err != nil || stored.FaqText != faq || stored.ContactUrl != contact || stored.Rewards != rewards || stored.SpaceSlug.String != created.Data.Slug {
		t.Fatalf("public information did not persist: %v", err)
	}
	archive := []byte(fmt.Sprintf(`{"status":"archived","revision":%d}`, created.Data.Revision))
	if w = requestJSON(t, spaceAPI, "POST", path+"/status", archive, f.cookie); w.Code != 200 {
		t.Fatalf("archive: %d %s", w.Code, w.Body.String())
	}
	if w = requestJSON(t, spaceAPI, "POST", path+"/leave", nil, other.cookie); w.Code != 200 {
		t.Fatalf("leave archived Space: %d", w.Code)
	}
	if w = requestJSON(t, spaceAPI, "POST", path+"/join", nil, other.cookie); w.Code != 409 {
		t.Fatalf("archived Space accepted new membership: %d", w.Code)
	}
	if w = requestJSON(t, eventAPI, "GET", "/by-slug/"+f.event.Slug, nil, nil); w.Code != 200 {
		t.Fatalf("archive hid existing published event: %d", w.Code)
	}
}
func requestJSON(t *testing.T, handler http.Handler, method, path string, body []byte, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}
func reserveOwner(t *testing.T, f ownerFixture) db.Rsvps {
	t.Helper()
	result, _, err := workflows.New(f.pool).Register(context.Background(), workflows.Request{EventID: f.event.ID, UserID: f.owner.ID, Email: f.owner.Email, Name: f.owner.DisplayName, Key: fmt.Sprintf("owner-request-%d", time.Now().UnixNano()), RequestID: "owner-ci"})
	if err != nil {
		t.Fatal(err)
	}
	return result.Data
}

// Extends the mandatory lifecycle suite using its existing isolated fixtures.
func TestOwnerLifecycle_RegistrationQuestionsAndPrivateAnswers(t *testing.T) {
	f, other := ownerSetup(t), ownerSetup(t)
	handler := NewHandler(f.pool).Routes()
	path := "/" + idString(f.event.ID)
	form := registrationforms.Form{Revision: 1, NameRequired: true, Questions: []registrationforms.Question{
		{Label: "What would you like to learn?", Kind: "text", Required: true, Options: []string{}},
		{Label: "I agree to the event guidelines", Kind: "checkbox", Required: true, Options: []string{}},
	}}
	body, _ := json.Marshal(form)
	if w := requestJSON(t, handler, "GET", path+"/registration-form", nil, nil); w.Code != 401 {
		t.Fatal("anonymous form management allowed")
	}
	if w := requestJSON(t, handler, "PUT", path+"/registration-form", body, other.cookie); w.Code != 404 {
		t.Fatal("cross-owner form edit allowed")
	}
	w := requestJSON(t, handler, "PUT", path+"/registration-form", body, f.cookie)
	if w.Code != 200 {
		t.Fatalf("save questions: %d %s", w.Code, w.Body.String())
	}
	var config struct {
		Data registrationforms.Form `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &config) != nil || config.Data.Revision != 2 || len(config.Data.Questions) != 2 {
		t.Fatal("form not persisted")
	}
	for _, question := range config.Data.Questions {
		if !strings.HasPrefix(question.PublicCode, "Q-") {
			t.Fatal("public question reference missing")
		}
	}
	if w = requestJSON(t, handler, "PUT", path+"/registration-form", body, f.cookie); w.Code != 409 {
		t.Fatal("stale form overwrote questions")
	}
	if w = requestJSON(t, handler, "GET", "/by-slug/"+f.event.Slug+"/registration-form", nil, nil); w.Code != 200 {
		t.Fatal("published form unavailable")
	}
	engine := workflows.New(f.pool)
	req := workflows.Request{EventID: f.event.ID, UserID: f.owner.ID, Email: f.owner.Email, Name: f.owner.DisplayName, Key: fmt.Sprintf("question-request-%d", time.Now().UnixNano())}
	if _, _, err := engine.Register(context.Background(), req); err == nil {
		t.Fatal("form revision bypassed")
	}
	revision := int64(2)
	req.FormRevision = &revision
	if _, _, err := engine.Register(context.Background(), req); err == nil {
		t.Fatal("required answers bypassed")
	}
	req.Answers = []registrationforms.Answer{{Code: config.Data.Questions[0].PublicCode, Text: "PostgreSQL transactions"}, {Code: config.Data.Questions[1].PublicCode, Checked: true}}
	result, replayed, err := engine.Register(context.Background(), req)
	if err != nil || replayed {
		t.Fatalf("registration: %v", err)
	}
	detailPath := path + "/registrations/by-code/" + result.Data.PublicCode
	if w = requestJSON(t, handler, "GET", detailPath, nil, other.cookie); w.Code != 404 {
		t.Fatal("other owner read private answers")
	}
	if w = requestJSON(t, handler, "GET", detailPath, nil, nil); w.Code != 401 {
		t.Fatal("anonymous read private answers")
	}
	if w = requestJSON(t, handler, "GET", "/"+idString(other.event.ID)+"/registrations/by-code/"+result.Data.PublicCode, nil, other.cookie); w.Code != 404 {
		t.Fatal("guest escaped event scope")
	}
	updated := config.Data
	updated.Questions = updated.Questions[:1]
	updated.Questions[0].Label = "Your learning goal"
	body, _ = json.Marshal(updated)
	if w = requestJSON(t, handler, "PUT", path+"/registration-form", body, f.cookie); w.Code != 200 {
		t.Fatalf("question edit: %d %s", w.Code, w.Body.String())
	}
	if replay, wasReplay, replayErr := engine.Register(context.Background(), req); replayErr != nil || !wasReplay || replay.Data.ID != result.Data.ID {
		t.Fatal("form edit broke successful idempotent replay")
	}
	w = requestJSON(t, handler, "GET", detailPath, nil, f.cookie)
	var detail struct {
		Answers []registrationforms.Snapshot `json:"answers"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &detail) != nil || len(detail.Answers) != 2 || detail.Answers[0].Label != "What would you like to learn?" || detail.Answers[0].Text != "PostgreSQL transactions" || !detail.Answers[1].Checked {
		t.Fatal("editing/removing questions changed historical answers")
	}
	stored, err := f.q.GetEventByID(context.Background(), f.event.ID)
	if err != nil || stored.CurrentRsvpCount != 1 || stored.RegistrationFormRevision != 3 {
		t.Fatal("question mutation changed inventory or failed revision")
	}
	for _, kind := range []string{"select", "url", "number", "phone"} {
		question := registrationforms.Question{PublicCode: "Q-ci", Label: "CI validation", Kind: kind, Required: true, Options: []string{"One", "Two"}}
		if _, err := registrationforms.ValidateAnswers(registrationforms.Form{Questions: []registrationforms.Question{question}}, "", []registrationforms.Answer{{Code: "Q-ci", Text: "invalid"}}); err == nil {
			t.Fatalf("invalid %s answer accepted", kind)
		}
	}
	// Exercise the owner status contract with this suite's existing isolated registration.
	statusRequest := func(target string, updated time.Time, notify bool, message, key string, cookie *http.Cookie) *httptest.ResponseRecorder {
		payload, _ := json.Marshal(struct {
			Status    string    `json:"status"`
			UpdatedAt time.Time `json:"updated_at"`
			Notify    bool      `json:"notify_guest"`
			Message   string    `json:"message"`
		}{target, updated, notify, message})
		r := httptest.NewRequest("POST", detailPath+"/status", bytes.NewReader(payload))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", key)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, r)
		return response
	}
	key := fmt.Sprintf("status-contract-%d", time.Now().UnixNano())
	if response := statusRequest("cancelled", result.Data.UpdatedAt.Time, false, "", key, nil); response.Code != 401 {
		t.Fatal("anonymous status edit allowed")
	}
	if response := statusRequest("cancelled", result.Data.UpdatedAt.Time, false, "", key, other.cookie); response.Code != 404 {
		t.Fatal("cross-owner status edit allowed")
	}
	var noticesBefore int
	if err = f.pool.QueryRow(context.Background(), "SELECT count(*) FROM email_outbox WHERE rsvp_id=$1", result.Data.ID).Scan(&noticesBefore); err != nil {
		t.Fatal(err)
	}
	w = statusRequest("cancelled", result.Data.UpdatedAt.Time, false, "", key, f.cookie)
	var change workflows.OwnerStatusResult
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &change) != nil || change.Data.Status != db.RegistrationStatusCancelled || change.NotificationQueued {
		t.Fatalf("silent cancellation: %d %s", w.Code, w.Body.String())
	}
	cancelledVersion := change.Data.UpdatedAt.Time
	if replay := statusRequest("cancelled", result.Data.UpdatedAt.Time, false, "", key, f.cookie); replay.Code != 200 || replay.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("status retry was not idempotent")
	}
	if response := statusRequest("confirmed", cancelledVersion, false, "", key, f.cookie); response.Code != 409 {
		t.Fatal("changed-payload retry allowed")
	}
	var noticesAfter int
	if err = f.pool.QueryRow(context.Background(), "SELECT count(*) FROM email_outbox WHERE rsvp_id=$1", result.Data.ID).Scan(&noticesAfter); err != nil || noticesAfter != noticesBefore {
		t.Fatal("silent status change enqueued a notification")
	}
	stored, err = f.q.GetEventByID(context.Background(), f.event.ID)
	ticket, ticketErr := f.q.GetTicketByRSVP(context.Background(), result.Data.ID)
	if err != nil || ticketErr != nil || stored.CurrentRsvpCount != 0 || ticket.Status != "revoked" {
		t.Fatal("cancellation failed inventory/ticket invariants")
	}
	if response := statusRequest("confirmed", result.Data.UpdatedAt.Time, true, "Welcome back", key+"-stale", f.cookie); response.Code != 409 {
		t.Fatal("stale guest version overwrote status")
	}
	w = statusRequest("confirmed", cancelledVersion, true, "Welcome <back>", key+"-restore", f.cookie)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &change) != nil || change.Data.Status != db.RegistrationStatusConfirmed || !change.NotificationQueued {
		t.Fatalf("restore with notification: %d %s", w.Code, w.Body.String())
	}
	var customMessage string
	if err = f.pool.QueryRow(context.Background(), "SELECT payload_json->>'custom_message' FROM email_outbox WHERE rsvp_id=$1 AND payload_json ? 'custom_message' ORDER BY created_at DESC LIMIT 1", result.Data.ID).Scan(&customMessage); err != nil || customMessage != "Welcome <back>" {
		t.Fatal("custom notification text was not persisted")
	}
	var actionURL, organizerName string
	if err = f.pool.QueryRow(context.Background(), "SELECT payload_json->>'action_url',payload_json->>'organizer_name' FROM email_outbox WHERE rsvp_id=$1 ORDER BY created_at DESC LIMIT 1", result.Data.ID).Scan(&actionURL, &organizerName); err != nil || !strings.Contains(actionURL, "/my-events/"+result.Data.PublicCode) || organizerName == "" {
		t.Fatal("notification missing real registration action or host")
	}
	resendBody, _ := json.Marshal(struct {
		Status    string    `json:"status"`
		UpdatedAt time.Time `json:"updated_at"`
		Notify    bool      `json:"notify_guest"`
		Message   string    `json:"message"`
		Resend    bool      `json:"resend_notification"`
	}{"confirmed", change.Data.UpdatedAt.Time, true, "Resend current status", true})
	resendRequest := func(cookie *http.Cookie) *httptest.ResponseRecorder {
		request := httptest.NewRequest("POST", detailPath+"/status", bytes.NewReader(resendBody))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", key+"-resend")
		if cookie != nil {
			request.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	if response := resendRequest(other.cookie); response.Code != 404 {
		t.Fatal("cross-owner notification allowed")
	}
	if response := resendRequest(f.cookie); response.Code != 200 {
		t.Fatalf("notification resend failed: %d %s", response.Code, response.Body.String())
	}
	var afterResend int
	if err = f.pool.QueryRow(context.Background(), "SELECT count(*) FROM email_outbox WHERE rsvp_id=$1", result.Data.ID).Scan(&afterResend); err != nil {
		t.Fatal(err)
	}
	if response := resendRequest(f.cookie); response.Code != 200 || response.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("notification retry did not replay")
	}
	var afterReplay int
	_ = f.pool.QueryRow(context.Background(), "SELECT count(*) FROM email_outbox WHERE rsvp_id=$1", result.Data.ID).Scan(&afterReplay)
	storedRSVP, readErr := f.q.GetWorkflowRSVP(context.Background(), result.Data.ID)
	if readErr != nil || storedRSVP.Status != "confirmed" || !storedRSVP.UpdatedAt.Time.Equal(change.Data.UpdatedAt.Time) || afterReplay != afterResend {
		t.Fatal("notification resend changed status or duplicated email")
	}
	// All five manual states must persist their actual inventory and admission effects.
	version := storedRSVP.UpdatedAt.Time
	for step, target := range []string{"pending_approval", "waitlisted", "rejected", "cancelled", "confirmed"} {
		response := statusRequest(target, version, true, "Manual status check", fmt.Sprintf("%s-manual-%d", key, step), f.cookie)
		if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &change) != nil || string(change.Data.Status) != target {
			t.Fatalf("manual %s: %d %s", target, response.Code, response.Body.String())
		}
		version = change.Data.UpdatedAt.Time
		row, readErr := f.q.GetWorkflowRSVP(context.Background(), result.Data.ID)
		eventState, eventErr := f.q.GetEventByID(context.Background(), f.event.ID)
		expectedCount := int32(0)
		if target == "confirmed" {
			expectedCount = 1
		}
		if readErr != nil || eventErr != nil || eventState.CurrentRsvpCount != expectedCount || (target == "pending_approval" && (!row.ApprovalExpiresAt.Valid || !row.ApprovalExpiresAt.Time.After(time.Now()))) {
			t.Fatal("manual status did not preserve expiry/capacity")
		}
	}
	// Existing automatic waiting-list behavior must report its actual offered state.
	if _, err = f.pool.Exec(context.Background(), "UPDATE events SET waitlist_enabled=true WHERE id=$1", f.event.ID); err != nil {
		t.Fatal(err)
	}
	response := statusRequest("waitlisted", version, true, "Waiting list check", key+"-auto-list", f.cookie)
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &change) != nil || change.Data.Status != db.RegistrationStatusOffered || !change.NotificationQueued {
		t.Fatalf("automatic offer response was stale: %d %s", response.Code, response.Body.String())
	}
	response = statusRequest("confirmed", change.Data.UpdatedAt.Time, true, "Confirm offered place", key+"-auto-confirm", f.cookie)
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &change) != nil || change.Data.Status != db.RegistrationStatusConfirmed {
		t.Fatal("manual confirmation did not consume held place")
	}
	ticket, ticketErr = f.q.GetTicketByRSVP(context.Background(), result.Data.ID)
	stored, err = f.q.GetEventByID(context.Background(), f.event.ID)
	if err != nil || ticketErr != nil || stored.CurrentRsvpCount != 1 || ticket.Status != "issued" {
		t.Fatal("restoration failed inventory/ticket invariants")
	}
	if _, err = f.q.CheckInWorkflowTicket(context.Background(), db.CheckInWorkflowTicketParams{TicketID: ticket.ID, EventID: f.event.ID, ActorID: f.owner.ID}); err != nil {
		t.Fatal(err)
	}
	if response := statusRequest("cancelled", change.Data.UpdatedAt.Time, false, "", key+"-scanned", f.cookie); response.Code != 409 {
		t.Fatal("checked-in guest status changed")
	}
}

func TestOwnerLifecycle_RevisionCancellationAndOwnership(t *testing.T) {
	f := ownerSetup(t)
	ctx := context.Background()
	registration := reserveOwner(t, f)
	handler := NewHandler(f.pool).Routes()
	path := "/" + idString(f.event.ID)
	update := updateRequest{Title: f.event.Title, Description: "Updated CI description", Format: string(f.event.Format), StartsAt: f.event.StartsAt.Time.Add(time.Hour), EndsAt: f.event.EndsAt.Time.Add(time.Hour), Timezone: "UTC", Capacity: &f.event.Capacity.Int32, Location: f.event.LocationName.String, Mode: "instant", AllowGuest: true, OfferMinutes: 10, Revision: 1}
	body, _ := json.Marshal(update)
	w := requestJSON(t, handler, "PUT", path, body, f.cookie)
	if w.Code != 409 {
		t.Fatalf("material impact gate: %d %s", w.Code, w.Body.String())
	}
	update.Acknowledge = true
	body, _ = json.Marshal(update)
	w = requestJSON(t, handler, "PUT", path, body, f.cookie)
	if w.Code != 200 {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	w = requestJSON(t, handler, "PUT", path, body, f.cookie)
	if w.Code != 409 {
		t.Fatalf("old revision overwrote event: %d", w.Code)
	}
	body, _ = json.Marshal(transitionRequest{Status: db.EventStatusCancelled, Revision: 2, Acknowledge: true})
	w = requestJSON(t, handler, "POST", path+"/status", body, f.cookie)
	if w.Code != 200 {
		t.Fatalf("cancel: %d %s", w.Code, w.Body.String())
	}
	current, err := f.q.GetEventByID(ctx, f.event.ID)
	if err != nil || current.CurrentRsvpCount != 0 || current.HeldRsvpCount != 0 || current.Status != db.EventStatusCancelled {
		t.Fatalf("cancel inventory/status: %+v %v", current, err)
	}
	record, err := f.q.GetWorkflowRSVP(ctx, registration.ID)
	if err != nil || record.Status != db.RegistrationStatusCancelled {
		t.Fatal("registration remained active")
	}
	ticket, err := f.q.GetTicketByRSVP(ctx, registration.ID)
	if err != nil || ticket.Status != "revoked" {
		t.Fatal("ticket remained valid")
	}
	public, err := f.q.GetPublishedEventBySlug(ctx, f.event.Slug)
	if err != nil || public.Status != db.EventStatusCancelled {
		t.Fatal("public cancellation vanished")
	}
	other, err := f.q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("other-%d@example.invalid", time.Now().UnixNano()), DisplayName: "Other isolated owner", PasswordHash: "not-a-login-credential"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.q.CreateOrganizer(ctx, db.CreateOrganizerParams{UserID: other.ID, DisplayName: "Other isolated profile"}); err != nil {
		t.Fatal(err)
	}
	token, _, err := auth.NewSessionManager(f.q).CreateSession(ctx, other.ID, "CI", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	w = requestJSON(t, handler, "GET", path+"/manage", nil, &http.Cookie{Name: auth.DefaultSessionCookie, Value: token})
	if w.Code != 404 {
		t.Fatalf("other organizer accessed event: %d", w.Code)
	}
}
func TestOwnerLifecycle_DashboardFiltersSummaryAndAppearance(t *testing.T) {
	f, other := ownerSetup(t), ownerSetup(t)
	reserveOwner(t, f)
	handler := NewHandler(f.pool).Routes()
	path := "/" + idString(f.event.ID)
	for _, suffix := range []string{"/summary", "/registrations?q=Isolated&status=confirmed"} {
		if w := requestJSON(t, handler, "GET", path+suffix, nil, other.cookie); w.Code != 404 {
			t.Fatalf("cross-owner dashboard read: %d", w.Code)
		}
		if w := requestJSON(t, handler, "GET", path+suffix, nil, nil); w.Code != 401 {
			t.Fatalf("anonymous dashboard read: %d", w.Code)
		}
	}
	w := requestJSON(t, handler, "GET", path+"/summary", nil, f.cookie)
	var summary struct {
		Data db.EventGuestSummaryRow `json:"data"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &summary) != nil || summary.Data.Confirmed != 1 || summary.Data.Total != 1 || summary.Data.CheckedIn != 0 {
		t.Fatalf("summary is not based on registrations: %d %s", w.Code, w.Body.String())
	}
	w = requestJSON(t, handler, "GET", path+"/registrations?q=Isolated&status=confirmed", nil, f.cookie)
	var roster struct {
		Data []db.EventRosterFilteredRow `json:"data"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &roster) != nil || len(roster.Data) != 1 || roster.Data[0].UserID != f.owner.ID {
		t.Fatalf("matching guest not found: %d %s", w.Code, w.Body.String())
	}
	w = requestJSON(t, handler, "GET", path+"/registrations?status=waitlisted", nil, f.cookie)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &roster) != nil || len(roster.Data) != 0 {
		t.Fatal("guest status filter ignored")
	}
	if w = requestJSON(t, handler, "GET", path+"/registrations?status=unrecognized", nil, f.cookie); w.Code != 400 {
		t.Fatal("invalid guest status accepted")
	}
	w = requestJSON(t, handler, "GET", "/organizer?view=upcoming&q=workflow", nil, f.cookie)
	var catalog struct {
		Data []db.Events `json:"data"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &catalog) != nil || len(catalog.Data) != 1 || catalog.Data[0].OrganizerID != f.organizer.ID {
		t.Fatalf("owner event filter: %d %s", w.Code, w.Body.String())
	}
	if w = requestJSON(t, handler, "GET", "/organizer?view=invalid", nil, f.cookie); w.Code != 400 {
		t.Fatal("invalid event view accepted")
	}
	appearance := []byte(`{"theme":"ambient","theme_color":"#ced9da","theme_font":"google","theme_style":"crisp","theme_display":"dark","revision":1,"remove_banner":true}`)
	if w = requestJSON(t, handler, "PUT", path+"/appearance", appearance, other.cookie); w.Code != 404 {
		t.Fatal("cross-owner appearance update allowed")
	}
	if w = requestJSON(t, handler, "PUT", path+"/appearance", appearance, f.cookie); w.Code != 200 {
		t.Fatalf("appearance save: %d %s", w.Code, w.Body.String())
	}
	stored, err := f.q.GetEventByID(context.Background(), f.event.ID)
	if err != nil || stored.Theme != "ambient" || stored.ThemeDisplay != "dark" || stored.ThemeFont != "google" || stored.Revision != 2 || stored.CurrentRsvpCount != 1 {
		t.Fatal("appearance save changed inventory or failed to persist")
	}
	if w = requestJSON(t, handler, "PUT", path+"/appearance", appearance, f.cookie); w.Code != 409 {
		t.Fatal("stale appearance overwrote event")
	}
}

func TestOwnerCheckIn_OneUseAndCancellationGuard(t *testing.T) {
	f := ownerSetup(t)
	registration := reserveOwner(t, f)
	ticket, err := f.q.GetTicketByRSVP(context.Background(), registration.ID)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(struct {
		ID     string `json:"ticket_id"`
		Reason string `json:"reason"`
	}{idString(ticket.ID), "CI manual lookup"})
	handler := NewHandler(f.pool).Routes()
	path := "/" + idString(f.event.ID) + "/check-in"
	w := requestJSON(t, handler, "POST", path, body, f.cookie)
	if w.Code != 200 {
		t.Fatalf("check-in: %d %s", w.Code, w.Body.String())
	}
	summaryResponse := requestJSON(t, handler, "GET", "/"+idString(f.event.ID)+"/summary", nil, f.cookie)
	var summary struct {
		Data db.EventGuestSummaryRow `json:"data"`
	}
	if summaryResponse.Code != 200 || json.Unmarshal(summaryResponse.Body.Bytes(), &summary) != nil || summary.Data.CheckedIn != 1 {
		t.Fatal("checked-in summary was not refreshed")
	}
	remaining := requestJSON(t, handler, "GET", "/"+idString(f.event.ID)+"/registrations?status=not_arrived", nil, f.cookie)
	var roster struct {
		Data []db.EventRosterFilteredRow `json:"data"`
	}
	if remaining.Code != 200 || json.Unmarshal(remaining.Body.Bytes(), &roster) != nil || len(roster.Data) != 0 {
		t.Fatal("checked-in guest remained in arrival queue")
	}
	w = requestJSON(t, handler, "POST", path, body, f.cookie)
	if w.Code != 409 {
		t.Fatalf("duplicate check-in: %d", w.Code)
	}
	w = requestJSON(t, rsvps.NewHandler(f.pool).Routes(), "POST", "/"+idString(registration.ID)+"/cancel", nil, f.cookie)
	if w.Code != 409 {
		t.Fatalf("cancel checked-in registration: %d", w.Code)
	}
}
func TestTransfer_RevokesOldTicketAndSignsRecipientTicket(t *testing.T) {
	f := ownerSetup(t)
	ctx := context.Background()
	source := reserveOwner(t, f)
	oldTicket, err := f.q.GetTicketByRSVP(ctx, source.ID)
	if err != nil {
		t.Fatal(err)
	}
	recipient, err := f.q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("recipient-%d@example.invalid", time.Now().UnixNano()), DisplayName: "Isolated recipient", PasswordHash: "not-a-login-credential"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.pool.Exec(ctx, "UPDATE users SET email_verified=TRUE WHERE id=$1 OR id=$2", f.owner.ID, recipient.ID); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(struct {
		Email string `json:"recipient_email"`
	}{recipient.Email})
	handler := rsvps.NewHandler(f.pool).Routes()
	path := "/" + idString(source.ID) + "/transfer"
	w := requestJSON(t, handler, "POST", path, body, f.cookie)
	if w.Code != 200 {
		t.Fatalf("transfer: %d %s", w.Code, w.Body.String())
	}
	w = requestJSON(t, handler, "POST", path, body, f.cookie)
	if w.Code != 200 {
		t.Fatalf("transfer retry: %d", w.Code)
	}
	inventory, err := f.q.GetEventByID(ctx, f.event.ID)
	if err != nil || inventory.CurrentRsvpCount != 1 || inventory.HeldRsvpCount != 0 {
		t.Fatal("transfer changed total inventory")
	}
	revoked, err := f.q.GetTicketByRSVP(ctx, source.ID)
	if err != nil || revoked.Status != "revoked" {
		t.Fatal("old ticket remained valid")
	}
	destination, err := f.q.GetAccountEventRSVP(ctx, db.GetAccountEventRSVPParams{EventID: f.event.ID, UserID: recipient.ID})
	if err != nil || destination.TransferCount != 1 {
		t.Fatalf("destination registration: %v", err)
	}
	ticket, err := f.q.GetTicketByRSVP(ctx, destination.ID)
	if err != nil || ticket.ID == oldTicket.ID || ticket.Status != "issued" {
		t.Fatal("replacement ticket missing")
	}
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TICKET_SIGNING_SEED", base64.StdEncoding.EncodeToString(private.Seed()))
	token, _, err := auth.NewSessionManager(f.q).CreateSession(ctx, recipient.ID, "CI", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	w = requestJSON(t, handler, "GET", "/"+idString(destination.ID)+"/ticket", nil, &http.Cookie{Name: auth.DefaultSessionCookie, Value: token})
	if w.Code != 200 {
		t.Fatalf("ticket: %d %s", w.Code, w.Body.String())
	}
	var response struct {
		Code    string `json:"code"`
		QRImage string `json:"qr_image"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(response.QRImage, "data:image/png;base64,"))
	if err != nil {
		t.Fatal("pass QR is not encoded PNG")
	}
	qr, err := png.DecodeConfig(bytes.NewReader(raw))
	if err != nil || qr.Width != 384 || qr.Height != 384 {
		t.Fatal("pass QR image missing or invalid")
	}
	claims, err := rsvps.VerifyTicketCode(response.Code)
	if err != nil || claims.EventID != f.event.ID || claims.TicketID != ticket.ID {
		t.Fatal("signed ticket failed verification")
	}
	if _, err = rsvps.VerifyTicketCode(response.Code + "tampered"); err == nil {
		t.Fatal("tampered ticket accepted")
	}
}
