package programs

// These cases extend the existing mandatory PostgreSQL CI suite. They never
// create application fixtures outside an explicitly isolated _test database.
import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestProgramApplicationsPrivateReviewAndRelease(t *testing.T) {
	ctx := context.Background()
	cfg, e := pgxpool.ParseConfig(testsupport.DatabaseURL(t))
	if e != nil {
		t.Fatal(e)
	}
	cfg.MaxConns = 4
	cfg.MinConns = 0
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	q := db.New(pool)
	sm := auth.NewSessionManager(q)
	nonce, e := auth.GenerateToken()
	if e != nil {
		t.Fatal(e)
	}
	slug := "program-ci-" + nonce[:12]
	ids := []pgtype.UUID{}
	tokens := []string{}
	for i := 0; i < 3; i++ {
		u, err := q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("program-%d-%s@example.invalid", i, nonce), DisplayName: "Isolated program review", PasswordHash: "mandatory-ci-only"})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, u.ID)
		if err = q.MarkUserEmailVerified(ctx, u.ID); err != nil {
			t.Fatal(err)
		}
		token, session, err := sm.CreateSession(ctx, u.ID, "mandatory-ci-programs", "", time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		tokens = append(tokens, token)
		if i == 0 {
			if _, err = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES($1)", u.ID); err != nil {
				t.Fatal(err)
			}
			if err = q.RecordPlatformAdminSession(ctx, session.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM admin_custom_message_deliveries WHERE message_id IN (SELECT id FROM admin_custom_messages WHERE actor_user_id=$1)", ids[0])
		_, _ = pool.Exec(ctx, "DELETE FROM admin_custom_messages WHERE actor_user_id=$1", ids[0])
		_, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE recipient_email LIKE $1", "program-%-"+nonce+"@example.invalid")
		_, _ = pool.Exec(ctx, "DELETE FROM program_activity WHERE program_id IN (SELECT id FROM programs WHERE slug=$1)", slug)
		_, _ = pool.Exec(ctx, "DELETE FROM program_releases WHERE program_id IN (SELECT id FROM programs WHERE slug=$1)", slug)
		_, _ = pool.Exec(ctx, "DELETE FROM program_applications WHERE program_id IN (SELECT id FROM programs WHERE slug=$1)", slug)
		_, _ = pool.Exec(ctx, "DELETE FROM programs WHERE slug=$1", slug)
		for _, id := range ids {
			_, _ = pool.Exec(ctx, "DELETE FROM platform_admin_sessions WHERE session_id IN (SELECT id FROM sessions WHERE user_id=$1)", id)
			_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", id)
			_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", id)
			_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
		}
	}()
	h := New(pool)
	public, admin := h.Routes(), h.AdminRoutes()
	call := func(isAdmin bool, user int, method, path, body, key string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		origin, cookie := "https://hackways.in", auth.DefaultSessionCookie
		router := public
		if isAdmin {
			origin, cookie = "https://admin.hackways.in", auth.AdminSessionCookie
			router = admin
		}
		r.Header.Set("Origin", origin)
		r.Header.Set("Content-Type", "application/json")
		if key != "" {
			r.Header.Set("Idempotency-Key", key)
		}
		if user >= 0 {
			r.AddCookie(&http.Cookie{Name: cookie, Value: tokens[user]})
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("expected %d got %d: %s", status, w.Code, w.Body.String())
		}
	}
	mine := func(user int) mineView {
		t.Helper()
		w := call(false, user, "GET", "/"+slug+"/application", "", "")
		check(w, 200)
		var body struct {
			Application *mineView `json:"application"`
		}
		if json.Unmarshal(w.Body.Bytes(), &body) != nil || body.Application == nil {
			t.Fatal("missing application")
		}
		return *body.Application
	}
	adminProgramRead := func() adminProgram {
		t.Helper()
		w := call(true, 0, "GET", "/"+slug, "", "")
		check(w, 200)
		var body struct {
			Data adminProgram `json:"data"`
		}
		if e = json.Unmarshal(w.Body.Bytes(), &body); e != nil {
			t.Fatal(e)
		}
		return body.Data
	}
	check(call(true, -1, "GET", "/", "", ""), 401)
	check(call(true, 1, "GET", "/", "", ""), 403)
	definition := programBody{Slug: slug, Title: "Isolated CI program", Summary: "Program verification in an isolated database", Overview: "A real PostgreSQL integration workflow", Responsibilities: "Host events using the platform", Recognition: "Recognition for actual contributions", Status: "open", MinimumAge: 18, Questions: []Question{{ID: "profile", Label: "Community profile", Kind: "url", Required: true, MaxLength: 2048, Domain: "linkedin.com"}}}
	raw, _ := json.Marshal(definition)
	check(call(true, 0, "POST", "/", string(raw), ""), 200)
	check(call(false, -1, "GET", "/"+slug+"/application", "", ""), 401)
	p := adminProgramRead()
	profiles := []profileBody{}
	for user := 1; user <= 2; user++ {
		profiles = append(profiles, profileBody{FullName: fmt.Sprintf("Applicant %d", user), City: "CI Community", Biography: "Mandatory isolated integration", Age: 18, AgeConfirmed: true, Proposal: "A community workshop to share practical ideas and meet other builders.", ProgramRevision: p.Revision, Answers: map[string]string{"profile": "https://www.linkedin.com/in/ci"}})
	}
	profiles[0].Answers["profile"] = "https://linkedin.com.evil.example/in/ci"
	raw, _ = json.Marshal(profiles[0])
	check(call(false, 1, "POST", "/"+slug+"/application", string(raw), ""), 400)
	profiles[0].Answers["profile"] = "https://uk.linkedin.com/in/ci"
	profiles[0].Age = 17
	for user := 1; user <= 2; user++ {
		raw, _ = json.Marshal(profiles[user-1])
		check(call(false, user, "POST", "/"+slug+"/application", string(raw), ""), 200)
	}
	a := mine(1)
	raw, _ = json.Marshal(struct {
		Revision        int64 `json:"revision"`
		ProgramRevision int64 `json:"program_revision"`
	}{a.Revision, p.Revision})
	check(call(false, 1, "POST", "/"+slug+"/application/submit", string(raw), ""), 400)
	profiles[0].Age = 18
	profiles[0].Revision = a.Revision
	raw, _ = json.Marshal(profiles[0])
	check(call(false, 1, "POST", "/"+slug+"/application", string(raw), ""), 200)
	for user := 1; user <= 2; user++ {
		a = mine(user)
		raw, _ = json.Marshal(struct {
			Revision        int64 `json:"revision"`
			ProgramRevision int64 `json:"program_revision"`
		}{a.Revision, p.Revision})
		check(call(false, user, "POST", "/"+slug+"/application/submit", string(raw), ""), 200)
		check(call(false, user, "POST", "/"+slug+"/application/submit", string(raw), ""), 200)
	}
	a = mine(1)
	aID, _ := a.ID.Value()
	w := call(true, 0, "GET", "/"+slug+"/applications/"+aID.(string), "", "")
	check(w, 200)
	saveDecision := func(user int, decision string) {
		t.Helper()
		app := mine(user)
		id, _ := app.ID.Value()
		raw, _ := json.Marshal(struct {
			Revision int64  `json:"revision"`
			Decision string `json:"decision"`
			Notes    string `json:"notes"`
		}{app.Revision, decision, "PRIVATE_REVIEW_NOTE_MUST_NOT_LEAK"})
		check(call(true, 0, "POST", "/"+slug+"/applications/"+id.(string)+"/decision", string(raw), ""), 200)
	}
	saveDecision(1, "accepted")
	w = call(false, 1, "GET", "/"+slug+"/application", "", "")
	check(w, 200)
	for _, secret := range []string{"reviewer_notes", "PRIVATE_REVIEW_NOTE", "\"decision\":\"accepted\"", "reviewed_by"} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatal("private review leaked")
		}
	}
	var earlyResultMessages int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox WHERE recipient_email LIKE $1 AND template_type='program_results_released'", "program-%-"+nonce+"@example.invalid").Scan(&earlyResultMessages); err != nil || earlyResultMessages != 0 {
		t.Fatal("result mail created before release")
	}
	if mine(1).PublishedDecision.Valid {
		t.Fatal("decision visible before release")
	}
	release := func(p adminProgram, key string) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(struct {
			Revision  int64 `json:"review_revision"`
			Count     int64 `json:"count"`
			Confirmed bool  `json:"confirmed"`
		}{p.ReviewRevision, p.Stats.Ready, true})
		return call(true, 0, "POST", "/"+slug+"/release", string(raw), key)
	}
	p = adminProgramRead()
	check(release(p, "release-program-ci-first"), 409)
	saveDecision(2, "rejected")
	check(release(p, "release-program-ci-stale"), 409)
	p = adminProgramRead()
	if p.Stats.Ready != 2 || p.Stats.Pending != 0 {
		t.Fatal("incorrect release preview")
	}
	check(release(p, "release-program-ci-final"), 200)
	check(release(p, "release-program-ci-final"), 200)
	if got := mine(1).PublishedDecision; !got.Valid || got.String != "accepted" {
		t.Fatal("missing released acceptance")
	}
	if got := mine(2).PublishedDecision; !got.Valid || got.String != "rejected" {
		t.Fatal("cross-user result mismatch")
	}
	// Real authorized Excel response and communications use the same isolated PostgreSQL records.
	exportPath := "/" + slug + "/applications/export.xlsx"
	check(call(true, -1, "GET", exportPath, "", ""), 401)
	check(call(true, 1, "GET", exportPath, "", ""), 403)
	exported := call(true, 0, "GET", exportPath, "", "")
	check(exported, 200)
	if !strings.Contains(exported.Header().Get("Content-Type"), "spreadsheetml.sheet") {
		t.Fatal("not an Excel workbook")
	}
	archive, err := zip.NewReader(bytes.NewReader(exported.Body.Bytes()), int64(exported.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	sheet := ""
	for _, file := range archive.File {
		stream, e := file.Open()
		if e != nil {
			t.Fatal(e)
		}
		data, e := io.ReadAll(stream)
		stream.Close()
		if e != nil {
			t.Fatal(e)
		}
		if file.Name == "xl/worksheets/sheet1.xml" {
			sheet = string(data)
		}
	}
	if strings.Contains(sheet, "<f>") {
		t.Fatal("export must not execute applicant formulas")
	}
	for _, expected := range []string{"Applicant 1", "Applicant 2", "autoFilter", "state=\"frozen\"", "r=\"F5\" s=\"7\""} {
		if !strings.Contains(sheet, expected) {
			t.Fatalf("Excel missing %s", expected)
		}
	}
	t.Setenv("ACS_ENDPOINT", "https://unit.communication.azure.com")
	t.Setenv("ACS_ACCESS_KEY", "mandatory-ci-not-a-live-key")
	t.Setenv("ACS_SENDER_ADDRESS", "notify@example.invalid")
	t.Setenv("EMAIL_OUTBOX_TEMPLATE_FILTER", "")
	communication := communicationBody{Subject: "Hello {{name}}", Body: "Your {{program}} result is {{decision}}.", Scope: "selected", IDs: []string{aID.(string)}}
	raw, _ = json.Marshal(communication)
	check(call(true, 1, "POST", "/"+slug+"/communications/preview", string(raw), ""), 403)
	response := call(true, 0, "POST", "/"+slug+"/communications/preview", string(raw), "")
	check(response, 200)
	var preview struct {
		Count int    `json:"count"`
		Hash  string `json:"audience_hash"`
		HTML  string `json:"html"`
	}
	if json.Unmarshal(response.Body.Bytes(), &preview) != nil || preview.Count != 1 || !strings.Contains(preview.HTML, "Applicant 1") || !strings.Contains(preview.HTML, "accepted") {
		t.Fatal("actual recipient preview missing")
	}
	communication.Confirmed = true
	communication.ExpectedCount = preview.Count
	communication.AudienceHash = strings.Repeat("0", 64)
	raw, _ = json.Marshal(communication)
	check(call(true, 0, "POST", "/"+slug+"/communications/send", string(raw), "program-communication-ci-stale"), 409)
	communication.AudienceHash = preview.Hash
	raw, _ = json.Marshal(communication)
	check(call(true, 0, "POST", "/"+slug+"/communications/send", string(raw), "program-communication-ci-final"), 202)
	check(call(true, 0, "POST", "/"+slug+"/communications/send", string(raw), "program-communication-ci-final"), 200)
	communication.Body = "Different body"
	raw, _ = json.Marshal(communication)
	check(call(true, 0, "POST", "/"+slug+"/communications/send", string(raw), "program-communication-ci-final"), 409)
	communication.IDs = []string{"00000000-0000-0000-0000-000000000001"}
	raw, _ = json.Marshal(communication)
	check(call(true, 0, "POST", "/"+slug+"/communications/preview", string(raw), ""), 409)
	var outboxID pgtype.UUID
	if err = pool.QueryRow(ctx, "SELECT id FROM email_outbox WHERE recipient_email LIKE $1 AND template_type='program_communication'", "program-1-"+nonce+"@example.invalid").Scan(&outboxID); err != nil {
		t.Fatal(err)
	}
	mailContext, err := q.ProgramMailContext(ctx, outboxID)
	if err != nil || mailContext.FullName != "Applicant 1" || mailContext.PublishedDecision.String != "accepted" {
		t.Fatal("mail hydration lost actual applicant context", err)
	}
	var messages int
	if e = pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox WHERE recipient_email LIKE $1", "program-%-"+nonce+"@example.invalid").Scan(&messages); e != nil || messages != 5 {
		t.Fatal("expected two receipts, two release notices and one communication despite retries")
	}
	if got := adminProgramRead(); len(got.Releases) != 1 || got.Stats.Released != 2 {
		t.Fatal("release retry duplicated effects")
	}
	app := mine(1)
	raw, _ = json.Marshal(struct {
		Revision int64  `json:"revision"`
		Decision string `json:"decision"`
		Notes    string `json:"notes"`
	}{app.Revision, "rejected", ""})
	check(call(true, 0, "POST", "/"+slug+"/applications/"+aID.(string)+"/decision", string(raw), ""), 409)
	// An explicit published correction preserves private-save protection and sends one updated notice.
	correctionPath := "/" + slug + "/applications/" + aID.(string) + "/published-decision"
	correctionBody := struct {
		Decision  string `json:"decision"`
		Notes     string `json:"notes"`
		Revision  int64  `json:"revision"`
		Confirmed bool   `json:"confirmed"`
	}{"waitlisted", "Corrected after review", app.Revision, true}
	correctionRaw, _ := json.Marshal(correctionBody)
	check(call(true, -1, "POST", correctionPath, string(correctionRaw), "program-result-correction-ci"), 401)
	check(call(true, 1, "POST", correctionPath, string(correctionRaw), "program-result-correction-ci"), 403)
	correctionBody.Confirmed = false
	invalidCorrection, _ := json.Marshal(correctionBody)
	check(call(true, 0, "POST", correctionPath, string(invalidCorrection), "program-result-correction-ci"), 400)
	check(call(true, 0, "POST", correctionPath, string(correctionRaw), "program-result-correction-ci"), 200)
	check(call(true, 0, "POST", correctionPath, string(correctionRaw), "program-result-correction-ci"), 200)
	if got := mine(1).PublishedDecision.String; got != "waitlisted" {
		t.Fatal("updated published result missing")
	}
	correctionBody.Confirmed = true
	correctionBody.Decision = "rejected"
	changedCorrection, _ := json.Marshal(correctionBody)
	check(call(true, 0, "POST", correctionPath, string(changedCorrection), "program-result-correction-ci"), 409)
	check(call(true, 0, "POST", correctionPath, string(changedCorrection), "program-result-correction-stale"), 409)
	var correctionCount, supersededCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox WHERE recipient_email LIKE $1 AND payload_json->>'program_result_updated'='true'", "program-1-"+nonce+"@example.invalid").Scan(&correctionCount); err != nil || correctionCount != 1 {
		t.Fatal("result correction retry duplicated notice", err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox WHERE recipient_email LIKE $1 AND template_type='program_results_released' AND status='failed'", "program-1-"+nonce+"@example.invalid").Scan(&supersededCount); err != nil || supersededCount != 1 {
		t.Fatal("old unsent result not superseded", err)
	}
	app = mine(1)
	raw, _ = json.Marshal(struct {
		Revision int64 `json:"revision"`
	}{app.Revision})
	check(call(false, 1, "POST", "/"+slug+"/application/withdraw", string(raw), ""), 200)
	if active, e := q.ProgramNotificationActive(ctx, outboxID); e != nil || active {
		t.Fatal("withdrawn applicant remains eligible for email")
	}
	if mine(1).Status != "withdrawn" {
		t.Fatal("withdrawal did not persist")
	}
	if _, e = pool.Exec(ctx, "UPDATE platform_admin_sessions SET otp_verified_at=clock_timestamp()-INTERVAL '25 hours' WHERE session_id IN (SELECT id FROM sessions WHERE user_id=$1)", ids[0]); e != nil {
		t.Fatal(e)
	}
	check(call(true, 0, "GET", "/"+slug, "", ""), 403)
}

func TestAmbassadorRequiredProfile(t *testing.T) {
	program := db.Programs{Slug: "ambassadors", MinimumAge: 18, Questions: []byte(`[]`)}
	body := profileBody{FullName: "CI applicant", City: "Hyderabad", Age: 21, AgeConfirmed: true, Proposal: strings.Repeat("Community event idea ", 3), Biography: "I organize local meetups.", Phone: "+91 9876543210", ProfileURL: "https://www.linkedin.com/in/ci-only"}
	if err := body.validate(program, true); err != nil {
		t.Fatal(err)
	}
	body.ProfileURL = "https://linkedin.com.evil.example/in/ci-only"
	if body.validate(program, true) == nil {
		t.Fatal("lookalike LinkedIn domain accepted")
	}
	body.ProfileURL = "https://linkedin.com/in/ci-only"
	body.Phone = "9876543210"
	if err := body.validate(program, true); err != nil || body.Phone != "+919876543210" {
		t.Fatal("10-digit Indian number must normalize without requiring +91", err)
	}
	body.Phone = "919876543210"
	if err := body.validate(program, true); err != nil || body.Phone != "+919876543210" {
		t.Fatal("unprefixed country code normalization failed", err)
	}
	body.Phone = "+447700900123"
	if err := body.validate(program, true); err != nil || body.Phone != "+447700900123" {
		t.Fatal("international number changed", err)
	}
	body.Phone = "12345"
	if body.validate(program, true) == nil {
		t.Fatal("incomplete phone accepted")
	}
	body.Phone = "+919876543210"
	body.Biography = ""
	if body.validate(program, true) == nil {
		t.Fatal("missing introduction accepted")
	}
	if err := body.validate(program, false); err != nil {
		t.Fatal("incomplete draft must remain saveable", err)
	}
}
