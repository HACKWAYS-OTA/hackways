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
	legacy := get(path)
	if legacy.Code != 200 {
		t.Fatal("legacy internal lookup failed")
	}
	targetRecord, err := q.AdminPersonDetail(ctx, target.ID)
	if err != nil {
		t.Fatal(err)
	}
	path = "/people/" + fmt.Sprint(targetRecord.Reference)
	if get(path).Code != 200 {
		t.Fatal("numeric account reference did not resolve")
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
	if rec := post(path+"/sessions/"+fmt.Sprint(ordinary.AdminReference)+"/revoke", actionBody{UpdatedAt: version(target.ID)}, "revoke-session-key-01"); rec.Code != 200 {
		t.Fatalf("session revoke got %d %s", rec.Code, rec.Body.String())
	}
	if _, err = pool.Exec(ctx, "UPDATE users SET avatar_url='https://cdn.hackways.in/media-uploads/ci.png?secret=private' WHERE id=$1", target.ID); err != nil {
		t.Fatal(err)
	}
	assetPath := "/storage/" + fmt.Sprint(targetRecord.Reference)
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
	messagePath := "/messaging/" + fmt.Sprint(msg.Reference)
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

// Campaign regressions extend the existing mandatory isolated PostgreSQL suite.
func TestAdminCampaignLifecycle(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("ACS_ENDPOINT", "https://ci.example.invalid")
	t.Setenv("ACS_ACCESS_KEY", "not-a-provider-key")
	t.Setenv("ACS_SENDER_ADDRESS", "notify@mail.hackways.in")
	t.Setenv("ACS_PROMOTIONAL_SENDER_ADDRESS", "updates@mail.hackways.in")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	nonce, _ := auth.GenerateToken()
	email := "campaign-admin-" + nonce + "@example.invalid"
	actor, err := q.CreateUser(ctx, db.CreateUserParams{Email: email, PasswordHash: "isolated-mandatory-ci", DisplayName: "Campaign CI"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "WITH removed AS (DELETE FROM email_campaign_deliveries d USING email_campaigns c WHERE d.campaign_id=c.id AND c.created_by=$1 RETURNING d.outbox_id) DELETE FROM email_outbox WHERE id IN(SELECT outbox_id FROM removed)", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM audit_log WHERE actor_user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM email_campaigns WHERE created_by=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM promotion_contacts WHERE email LIKE '%'||$1||'@example.invalid'", nonce)
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", actor.ID)
	}()
	if err = q.MarkUserEmailVerified(ctx, actor.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES($1)", actor.ID); err != nil {
		t.Fatal(err)
	}
	sm := auth.NewSessionManager(q)
	token, session, err := sm.CreateSession(ctx, actor.ID, "Campaign CI", "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err = q.RecordPlatformAdminSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	router := platform.RequestID(NewHandler(pool).Routes())
	call := func(method, path string, body json.RawMessage, origin string, signed bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", origin)
		req.Header.Set("X-Admin-Reason", url.PathEscape("Authorized isolated campaign regression"))
		req.Header.Set("Idempotency-Key", "campaign-create-"+nonce)
		if signed {
			req.AddCookie(&http.Cookie{Name: auth.AdminSessionCookie, Value: token})
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	marshal := func(body campaignBody) json.RawMessage {
		raw, e := json.Marshal(body)
		if e != nil {
			t.Fatal(e)
		}
		return raw
	}
	if got := call("GET", "/campaigns", nil, "", false); got.Code != 401 {
		t.Fatalf("anonymous campaign access %d", got.Code)
	}
	body := campaignBody{Title: "CI campaign", Subject: "CI subject", Body: "Hello <unsafe>", ActionLabel: "Discover", ActionURL: "https://hackways.in/discover"}
	if got := call("POST", "/campaigns", marshal(body), "https://untrusted.example.invalid", true); got.Code != 403 {
		t.Fatal("untrusted origin accepted")
	}
	created := call("POST", "/campaigns", marshal(body), "http://localhost:3103", true)
	if created.Code != 201 {
		t.Fatalf("create %d %s", created.Code, created.Body.String())
	}
	var result struct {
		Data campaignView `json:"data"`
	}
	if err = json.Unmarshal(created.Body.Bytes(), &result); err != nil || result.Data.Reference < 1 {
		t.Fatal("missing numeric campaign reference")
	}
	ref := result.Data.Reference
	path := "/campaigns/" + fmt.Sprint(ref)
	again := call("POST", "/campaigns", marshal(body), "http://localhost:3103", true)
	var replay struct {
		Data campaignView `json:"data"`
	}
	_ = json.Unmarshal(again.Body.Bytes(), &replay)
	if again.Code != 201 || replay.Data.Reference != ref {
		t.Fatal("draft creation duplicated")
	}
	changed := body
	changed.Subject = "Changed"
	if got := call("POST", "/campaigns", marshal(changed), "http://localhost:3103", true); got.Code != 409 {
		t.Fatal("creation key reused for different content")
	}
	contactEmail := "campaign-contact-" + nonce + "@example.invalid"
	importPayload := struct {
		CSV           string    `json:"csv"`
		ConsentSource string    `json:"consent_source"`
		ConsentedAt   time.Time `json:"consented_at"`
		Confirmed     bool      `json:"confirmed"`
	}{"email,name\n" + contactEmail + ",Contact CI", "Existing explicit opt-in record", time.Now().Add(-time.Hour), true}
	raw, _ := json.Marshal(importPayload)
	imported := call("POST", "/campaign-contacts/import", raw, "http://localhost:3103", true)
	if imported.Code != 200 {
		t.Fatalf("import %d %s", imported.Code, imported.Body.String())
	}
	detail := call("GET", path, nil, "", true)
	if detail.Code != 200 || strings.Contains(detail.Body.String(), "unsubscribe_hash") {
		t.Fatal("campaign detail failed or exposed secrets")
	}
	c, err := q.CampaignRead(ctx, ref)
	if err != nil {
		t.Fatal(err)
	}
	action := campaignBody{Revision: 1}
	test := call("POST", path+"/test", marshal(action), "http://localhost:3103", true)
	if test.Code != 200 {
		t.Fatalf("test queue %d %s", test.Code, test.Body.String())
	}
	if got := call("POST", path+"/test", marshal(action), "http://localhost:3103", true); got.Code != 200 || !strings.Contains(got.Body.String(), `"queued":0`) {
		t.Fatal("test email duplicated")
	}
	var firstTest pgtype.UUID
	if err = pool.QueryRow(ctx, "SELECT outbox_id FROM email_campaign_deliveries WHERE campaign_id=$1 AND is_test", c.ID).Scan(&firstTest); err != nil {
		t.Fatal(err)
	}
	body.Revision = 1
	body.Subject = "Reviewed subject"
	if got := call("POST", path+"/edit", marshal(body), "http://localhost:3103", true); got.Code != 200 {
		t.Fatalf("edit %d %s", got.Code, got.Body.String())
	}
	if active, e := q.CampaignMessageActive(ctx, firstTest); e != nil || active {
		t.Fatal("old draft test remains active")
	}
	count, err := q.CampaignAudienceCount(ctx)
	if err != nil || count < 1 {
		t.Fatal("imported audience missing")
	}
	action = campaignBody{Revision: 2, AudienceCount: count + 1, Confirmed: true}
	if got := call("POST", path+"/send", marshal(action), "http://localhost:3103", true); got.Code != 409 {
		t.Fatal("stale audience accepted")
	}
	action.AudienceCount = count
	// A replay is successful but can never add another recipient delivery.
	first := call("POST", path+"/send", marshal(action), "http://localhost:3103", true)
	if first.Code != 200 {
		t.Fatalf("send %d %s", first.Code, first.Body.String())
	}
	if got := call("POST", path+"/send", marshal(action), "http://localhost:3103", true); got.Code != 200 {
		t.Fatal("send replay failed")
	}
	stats, err := q.CampaignStats(ctx, c.ID)
	if err != nil || stats.Total != count {
		t.Fatal("campaign sent duplicates or missed recipients")
	}
	if got := call("POST", path+"/edit", marshal(body), "http://localhost:3103", true); got.Code != 409 {
		t.Fatal("queued campaign edited")
	}
	var outbox pgtype.UUID
	var payload []byte
	if err = pool.QueryRow(ctx, "SELECT o.id,o.payload_json FROM email_outbox o JOIN email_campaign_deliveries d ON d.outbox_id=o.id WHERE d.campaign_id=$1 AND NOT d.is_test AND o.recipient_email=$2", c.ID, contactEmail).Scan(&outbox, &payload); err != nil {
		t.Fatal(err)
	}
	if active, e := q.CampaignMessageActive(ctx, outbox); e != nil || !active {
		t.Fatal("valid campaign suppressed")
	}
	var link struct {
		URL string `json:"unsubscribe_url"`
	}
	_ = json.Unmarshal(payload, &link)
	parsed, e := url.Parse(link.URL)
	if e != nil || parsed.Host != "hackways.in" {
		t.Fatal("invalid unsubscribe destination")
	}
	key := parsed.Query().Get("key")
	public := auth.NewHandler(pool, false).Routes()
	unsubscribe := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/promotions/unsubscribe", strings.NewReader(`{"key":"`+key+`"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		public.ServeHTTP(w, req)
		return w
	}
	if unsubscribe().Code != 200 || unsubscribe().Code != 200 {
		t.Fatal("unsubscribe/replay failed")
	}
	if active, e := q.CampaignMessageActive(ctx, outbox); e != nil || active {
		t.Fatal("unsubscribed recipient still eligible")
	}
	if got := call("POST", "/campaign-contacts/import", raw, "http://localhost:3103", true); got.Code != 200 {
		t.Fatal("reimport failed")
	}
	subscribed, err := q.ReadPromotionPreference(ctx, contactEmail)
	if err != nil || subscribed {
		t.Fatal("import resubscribed an opt-out")
	}
	if got := call("POST", path+"/cancel", marshal(campaignBody{Revision: 2}), "http://localhost:3103", true); got.Code != 200 {
		t.Fatal("cancel failed")
	}
	if active, e := q.CampaignMessageActive(ctx, outbox); e != nil || active {
		t.Fatal("cancelled campaign still active")
	}

}
