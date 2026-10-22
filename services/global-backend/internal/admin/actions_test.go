package admin

// Mandatory CI regressions. testsupport refuses application/preview databases.
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/hackways/global-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAdminRecordActions(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	sm := auth.NewSessionManager(q)
	nonce, _ := auth.GenerateToken()
	var users []pgtype.UUID
	create := func(label string, verified bool) db.CreateUserRow {
		t.Helper()
		u, e := q.CreateUser(ctx, db.CreateUserParams{Email: label + nonce + "@example.invalid", PasswordHash: "isolated-ci-only", DisplayName: label})
		if e != nil {
			t.Fatal(e)
		}
		users = append(users, u.ID)
		if verified {
			if e = q.MarkUserEmailVerified(ctx, u.ID); e != nil {
				t.Fatal(e)
			}
		}
		return u
	}
	actor := create("detail-admin", true)
	target := create("detail-member", false)
	defer func() {
		for _, id := range users {
			_, _ = pool.Exec(ctx, "DELETE FROM audit_log WHERE actor_user_id=$1 OR entity_id=$1", id)
			_, _ = pool.Exec(ctx, "DELETE FROM idempotency_records WHERE response_json->'data'->>'id'=$1", uuidText(id))
		}
		for _, id := range users {
			_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", id)
			_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", id)
		}
		for _, id := range users {
			_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
		}
	}()
	if _, err = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES($1)", actor.ID); err != nil {
		t.Fatal(err)
	}
	token, session, err := sm.CreateSession(ctx, actor.ID, "admin-details-ci", "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err = q.RecordPlatformAdminSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	router := platform.RequestID(NewHandler(pool).Routes())
	request := func(method, path string, body actionBody, key, origin, reason, cookie string) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, strings.NewReader(string(raw)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", origin)
		req.Header.Set("Idempotency-Key", key)
		req.Header.Set("X-Admin-Reason", url.PathEscape(reason))
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: cookie, Value: token})
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	get := func(path string) *httptest.ResponseRecorder {
		return request("GET", path, actionBody{}, "", "", "", auth.AdminSessionCookie)
	}
	post := func(path string, body actionBody, key string) *httptest.ResponseRecorder {
		return request("POST", path, body, key, "http://localhost:3003", "CI record correction", auth.AdminSessionCookie)
	}
	version := func(id pgtype.UUID) time.Time {
		t.Helper()
		row, e := q.AdminPersonDetail(ctx, id)
		if e != nil {
			t.Fatal(e)
		}
		return row.UpdatedAt.Time
	}
	path := "/people/" + uuidText(target.ID)
	if _, err = q.AdminPersonDetail(ctx, target.ID); err != nil {
		t.Fatalf("person SQL contract: %v", err)
	}
	for _, cookie := range []string{"", auth.DefaultSessionCookie} {
		if got := request("GET", path, actionBody{}, "", "", "", cookie).Code; got != 401 {
			t.Fatalf("detail bypass got %d", got)
		}
	}
	for _, forbidden := range []string{"password_hash", "token_hash", "payload_json", "lease_token", "error_message"} {
		if strings.Contains(get(path).Body.String(), forbidden) {
			t.Fatalf("detail exposed %s", forbidden)
		}
	}
	if got := get("/people/not-an-id").Code; got != 400 {
		t.Fatalf("invalid id got %d", got)
	}
	if got := get(path + "?section=unknown").Code; got != 400 {
		t.Fatalf("invalid section got %d", got)
	}
	for _, section := range []string{"events", "registrations", "sessions", "messages", "activity"} {
		if rec := get(path + "?section=" + section); rec.Code != 200 {
			t.Fatalf("%s detail got %d %s", section, rec.Code, rec.Body.String())
		}
	}
	name := "Corrected name"
	body := actionBody{UpdatedAt: version(target.ID), DisplayName: &name}
	key := "edit-record-key-0001"
	if got := request("POST", path+"/edit", body, key, "https://attacker.example", "CI edit", auth.AdminSessionCookie).Code; got != 403 {
		t.Fatalf("cross-origin got %d", got)
	}
	if got := request("POST", path+"/edit", body, key, "http://localhost:3003", "", auth.AdminSessionCookie).Code; got != 400 {
		t.Fatalf("missing reason got %d", got)
	}
	if rec := post(path+"/edit", body, key); rec.Code != 200 {
		t.Fatalf("edit got %d %s", rec.Code, rec.Body.String())
	}
	if got := post(path+"/edit", body, key).Code; got != 200 {
		t.Fatalf("safe replay got %d", got)
	}
	var auditCount int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM audit_log WHERE actor_user_id=$1 AND entity_id=$2 AND action='admin.person.edit'", actor.ID, target.ID).Scan(&auditCount); err != nil || auditCount != 1 {
		t.Fatal("replay duplicated or omitted audit")
	}
	if got := post(path+"/edit", body, "edit-record-key-stale").Code; got != 409 {
		t.Fatalf("stale edit got %d", got)
	}
	changed := "Different"
	body.DisplayName = &changed
	if got := post(path+"/edit", body, key).Code; got != 409 {
		t.Fatalf("key reused with changes got %d", got)
	}
	admin, member := "admin", "member"
	if got := post(path+"/access", actionBody{UpdatedAt: version(target.ID), Access: &admin}, "unverified-admin-key").Code; got != 422 {
		t.Fatalf("unverified grant got %d", got)
	}
	if err = q.MarkUserEmailVerified(ctx, target.ID); err != nil {
		t.Fatal(err)
	}
	if rec := post(path+"/access", actionBody{UpdatedAt: version(target.ID), Access: &admin}, "grant-admin-key-0001"); rec.Code != 200 {
		t.Fatalf("grant got %d %s", rec.Code, rec.Body.String())
	}
	if got := post("/people/"+uuidText(actor.ID)+"/access", actionBody{UpdatedAt: version(actor.ID), Access: &member}, "self-revoke-key-0001").Code; got != 409 {
		t.Fatalf("self-lockout accepted: %d", got)
	}
	_, targetSession, err := sm.CreateSession(ctx, target.ID, "target-admin-ci", "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err = q.RecordPlatformAdminSession(ctx, targetSession.ID); err != nil {
		t.Fatal(err)
	}
	if rec := post(path+"/access", actionBody{UpdatedAt: version(target.ID), Access: &member}, "revoke-admin-key-0001"); rec.Code != 200 {
		t.Fatalf("revoke got %d %s", rec.Code, rec.Body.String())
	}
	allowed, err := q.IsPlatformAdminSession(ctx, db.IsPlatformAdminSessionParams{SessionID: targetSession.ID, UserID: target.ID})
	if err != nil || allowed {
		t.Fatal("revoked administrator retained access")
	}
	if got := post("/people/"+uuidText(actor.ID)+"/sessions/"+uuidText(session.ID)+"/revoke", actionBody{UpdatedAt: version(actor.ID)}, "self-session-key-0001").Code; got != 409 {
		t.Fatalf("current session revoke got %d", got)
	}
	_, ordinary, err := sm.CreateSession(ctx, target.ID, "target-member-ci", "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if got := post("/people/"+uuidText(actor.ID)+"/sessions/"+uuidText(ordinary.ID)+"/revoke", actionBody{UpdatedAt: version(actor.ID)}, "foreign-session-key-1").Code; got != 404 {
		t.Fatalf("cross-account revoke got %d", got)
	}
	if rec := post(path+"/sessions/"+uuidText(ordinary.ID)+"/revoke", actionBody{UpdatedAt: version(target.ID)}, "revoke-session-key-01"); rec.Code != 200 {
		t.Fatalf("session revoke got %d %s", rec.Code, rec.Body.String())
	}
	if _, err = pool.Exec(ctx, "UPDATE users SET avatar_url='https://cdn.hackways.in/media-uploads/ci.png?secret=private' WHERE id=$1", target.ID); err != nil {
		t.Fatal(err)
	}
	assetPath := "/storage/" + uuidText(target.ID)
	asset := get(assetPath + "?kind=avatar")
	if asset.Code != 200 || strings.Contains(asset.Body.String(), "secret=private") {
		t.Fatal("asset detail leaked URL credentials")
	}
	kind := "avatar"
	detach := actionBody{UpdatedAt: version(target.ID), Kind: &kind}
	if rec := post(assetPath+"/detach", detach, "detach-avatar-key-01"); rec.Code != 200 {
		t.Fatalf("detach got %d %s", rec.Code, rec.Body.String())
	}
	if got := post(assetPath+"/detach", detach, "detach-avatar-key-01").Code; got != 200 {
		t.Fatalf("detach replay got %d", got)
	}
	if got := get(assetPath + "?kind=avatar").Code; got != 404 {
		t.Fatalf("detached metadata still present %d", got)
	}
	var messageID pgtype.UUID
	if err = pool.QueryRow(ctx, "INSERT INTO email_outbox(recipient_email,template_type,payload_json,status) VALUES($1,'email_verification','{}','failed') RETURNING id", target.Email).Scan(&messageID); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM audit_log WHERE entity_id=$1", messageID)
		_, _ = pool.Exec(ctx, "DELETE FROM idempotency_records WHERE response_json->'data'->>'id'=$1", uuidText(messageID))
		_, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE id=$1", messageID)
	}()
	msg, err := q.AdminMessageDetail(ctx, messageID)
	if err != nil {
		t.Fatal(err)
	}
	messagePath := "/messaging/" + uuidText(messageID)
	retry := actionBody{UpdatedAt: msg.UpdatedAt.Time}
	if rec := post(messagePath+"/retry", retry, "retry-message-key-01"); rec.Code != 200 {
		t.Fatalf("retry got %d %s", rec.Code, rec.Body.String())
	}
	if got := post(messagePath+"/retry", retry, "retry-message-key-01").Code; got != 200 {
		t.Fatalf("retry replay got %d", got)
	}
	if got := post(messagePath+"/retry", retry, "retry-message-key-02").Code; got != 409 {
		t.Fatalf("pending mail retry got %d", got)
	}
	if _, err = pool.Exec(ctx, "UPDATE email_outbox SET status='failed',first_sent_at=clock_timestamp(),updated_at=clock_timestamp() WHERE id=$1", messageID); err != nil {
		t.Fatal(err)
	}
	msg, err = q.AdminMessageDetail(ctx, messageID)
	if err != nil {
		t.Fatal(err)
	}
	if msg.CanRetry {
		t.Fatal("ambiguous provider attempt exposed as retryable")
	}
	if got := post(messagePath+"/retry", actionBody{UpdatedAt: msg.UpdatedAt.Time}, "retry-ambiguous-key1").Code; got != 409 {
		t.Fatalf("ambiguous delivery replay allowed %d", got)
	}
}
func uuidText(id pgtype.UUID) string {
	b := id.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}
