package admin

// Mandatory CI regressions. testsupport refuses application/preview databases.
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/analytics"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/notifications"
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
		_, _ = pool.Exec(ctx, "DELETE FROM idempotency_records WHERE response_json->'data'->>'id' IN(SELECT id::text FROM device_restrictions WHERE updated_by=$1)", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM restriction_device_accounts WHERE device_id IN(SELECT id FROM device_restrictions WHERE updated_by=$1) OR user_id=ANY($2::uuid[])", actor.ID, users)
		_, _ = pool.Exec(ctx, "DELETE FROM account_restrictions WHERE user_id=ANY($1::uuid[]) OR source_device_id IN(SELECT id FROM device_restrictions WHERE updated_by=$2)", users, actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM device_restrictions WHERE updated_by=$1", actor.ID)
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
		return request("POST", path, body, key, "http://localhost:3003", "", auth.AdminSessionCookie)
	}

	// Analytics extends this mandatory real-PostgreSQL administration regression.
	if got := request("GET", "/analytics/live", actionBody{}, "", "", "", "").Code; got != 401 {
		t.Fatalf("anonymous live analytics access %d", got)
	}
	if got := request("GET", "/analytics/live", actionBody{}, "", "", "", auth.DefaultSessionCookie).Code; got != 401 {
		t.Fatalf("ordinary live analytics access %d", got)
	}
	for _, value := range []string{"invalid", "0", "90d"} {
		if got := get("/analytics/live?window=" + value).Code; got != 400 {
			t.Fatalf("unbounded live interval %s accepted: %d", value, got)
		}
	}
	for _, window := range []string{"5m", "1h", "24h"} {
		rec := get("/analytics/live?window=" + window)
		var live liveGlobalAnalytics
		if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &live) != nil || live.Window != window || live.Pages == nil || live.Sources == nil || rec.Header().Get("Cache-Control") != "private, no-store" || time.Since(live.CapturedAt) > 10*time.Second || time.Until(live.CapturedAt) > time.Second {
			t.Fatalf("invalid live account analytics %d %s", rec.Code, rec.Body.String())
		}
		if live.Accounts.NewAccounts < 2 {
			t.Fatal("live report did not include persisted accounts")
		}
	}
	if got := request("GET", "/analytics", actionBody{}, "", "", "", "").Code; got != 401 {
		t.Fatalf("anonymous analytics access %d", got)
	}
	for _, path := range []string{"/analytics?from=not-a-date", "/analytics?from=2000-01-01", "/analytics?to=2999-01-01"} {
		if got := get(path).Code; got != 400 {
			t.Fatalf("invalid analytics period %s got %d", path, got)
		}
	}
	if rec := get("/analytics"); rec.Code != 200 || rec.Header().Get("Cache-Control") != "private, no-store" || !strings.Contains(rec.Body.String(), `"account_daily"`) {
		t.Fatalf("analytics read %d %s", rec.Code, rec.Body.String())
	}
	if rec := get("/analytics?format=csv"); rec.Code != 200 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/csv") || len(strings.Split(strings.TrimSpace(rec.Body.String()), "\n")) != 31 {
		t.Fatalf("analytics export %d %s", rec.Code, rec.Body.String())
	}
	t.Run("consented collection and deduplication", func(t *testing.T) {
		collector := analytics.NewHandler(pool, strings.Repeat("c", 32), true).Routes()
		payload := fmt.Sprintf(`{"consent":true,"visitor_id":%q,"session_id":%q,"events":[{"id":%q,"event_name":"page_view","route":"/discover","device":"mobile","referrer_host":"search.example"}]}`, uuidText(actor.ID), uuidText(target.ID), uuidText(actor.ID))
		record := func(body, origin string) *httptest.ResponseRecorder {
			req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
			req.Header.Set("Origin", origin)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			collector.ServeHTTP(rec, req)
			return rec
		}
		defer func() { _, _ = pool.Exec(ctx, "DELETE FROM analytics_events WHERE id=$1", actor.ID) }()
		for _, invalid := range []string{strings.Replace(payload, `"consent":true`, `"consent":false`, 1), strings.Replace(payload, "/discover", "/reset-password", 1), strings.Replace(payload, "/discover", "/discover?email=private", 1), strings.Replace(payload, `"device":"mobile"`, `"device":"mobile","answer":"private"`, 1)} {
			if got := record(invalid, "http://localhost:3101").Code; got != 400 {
				t.Fatalf("private or unconsented event accepted %d", got)
			}
		}
		if got := record(payload, "https://untrusted.example").Code; got != 403 {
			t.Fatalf("foreign collection origin %d", got)
		}
		for i := 0; i < 2; i++ {
			if rec := record(payload, "http://localhost:3101"); rec.Code != 204 {
				t.Fatalf("collection %d %s", rec.Code, rec.Body.String())
			}
		}
		var count int
		var visitor, session string
		if e := pool.QueryRow(ctx, "SELECT count(*),min(visitor_hash),min(session_hash) FROM analytics_events WHERE id=$1", actor.ID).Scan(&count, &visitor, &session); e != nil || count != 1 || len(visitor) != 64 || len(session) != 64 || visitor == session || strings.Contains(visitor, uuidText(actor.ID)) {
			t.Fatalf("collection deduplication/hash failure: %d %v", count, e)
		}
		var reserved int
		if e := pool.QueryRow(ctx, "SELECT reserved_events FROM analytics_ingest_days WHERE day=(clock_timestamp() AT TIME ZONE 'UTC')::date").Scan(&reserved); e != nil {
			t.Fatal(e)
		}
		if _, e := pool.Exec(ctx, "UPDATE analytics_ingest_days SET reserved_events=25000 WHERE day=(clock_timestamp() AT TIME ZONE 'UTC')::date"); e != nil {
			t.Fatal(e)
		}
		defer func() {
			_, _ = pool.Exec(ctx, "UPDATE analytics_ingest_days SET reserved_events=$1 WHERE day=(clock_timestamp() AT TIME ZONE 'UTC')::date", reserved)
		}()
		if got := record(payload, "http://localhost:3101").Code; got != 429 {
			t.Fatalf("collection quota bypass %d", got)
		}
	})
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
	for _, section := range []string{"events", "registrations", "sessions", "devices", "messages", "activity"} {
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
	if got := request("POST", path+"/edit", body, key, "http://localhost:3003", "x", auth.AdminSessionCookie).Code; got != 400 {
		t.Fatalf("invalid legacy reason got %d", got)
	}
	if rec := post(path+"/edit", body, key); rec.Code != 200 {
		t.Fatalf("edit got %d %s", rec.Code, rec.Body.String())
	}
	if got := post(path+"/edit", body, key).Code; got != 200 {
		t.Fatalf("safe replay got %d", got)
	}
	var auditCount int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM audit_log WHERE actor_user_id=$1 AND entity_id=$2 AND action='admin.person.edit' AND metadata->>'reason'='Admin person.edit'", actor.ID, target.ID).Scan(&auditCount); err != nil || auditCount != 1 {
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
	// Extend mandatory real-PostgreSQL admin coverage with account/device review.
	blocked, unblocked := true, false
	if rec := post("/people/"+uuidText(actor.ID)+"/restriction", actionBody{UpdatedAt: version(actor.ID), Blocked: &blocked}, "restriction-self-key"); rec.Code != 409 {
		t.Fatalf("admin recovery bypass %d %s", rec.Code, rec.Body.String())
	}
	linked := create("linked-member", true)
	browser, _ := auth.GenerateToken()
	deviceHash := auth.HashToken(browser)
	for _, id := range []pgtype.UUID{target.ID, linked.ID, actor.ID} {
		if err = q.RememberRecognizedDevice(ctx, db.RememberRecognizedDeviceParams{UserID: id, TokenHash: deviceHash, UserAgent: "CI restriction browser"}); err != nil {
			t.Fatal(err)
		}
	}
	devices, e := q.AdminPersonDevices(ctx, db.AdminPersonDevicesParams{UserID: target.ID, Limit: 20})
	if e != nil || len(devices) != 1 {
		t.Fatalf("recognized device missing %v", e)
	}
	ref, e := strconv.ParseInt(devices[0].Reference, 10, 64)
	if e != nil {
		t.Fatal(e)
	}
	if _, err = pool.Exec(ctx, "DELETE FROM recognized_devices WHERE user_id=$1 AND token_hash=$2", actor.ID, deviceHash); err != nil {
		t.Fatal(err)
	}
	if err = q.BindSessionDevice(ctx, db.BindSessionDeviceParams{ID: session.ID, DeviceTokenHash: pgtype.Text{String: deviceHash, Valid: true}}); err != nil {
		t.Fatal(err)
	}
	browserBody := actionBody{UpdatedAt: version(target.ID), DeviceReference: &ref}
	if rec := post(path+"/devices/block", browserBody, "restriction-admin-browser"); rec.Code != 409 {
		t.Fatalf("admin-associated browser blocked %d", rec.Code)
	}
	if err = q.BindSessionDevice(ctx, db.BindSessionDeviceParams{ID: session.ID}); err != nil {
		t.Fatal(err)
	}
	targetToken, _, e := sm.CreateSession(ctx, target.ID, "legacy ordinary", "", time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	linkedToken, linkedSession, e := sm.CreateSession(ctx, linked.ID, "linked ordinary", "", time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	if err = q.BindSessionDevice(ctx, db.BindSessionDeviceParams{ID: linkedSession.ID, DeviceTokenHash: pgtype.Text{String: deviceHash, Valid: true}}); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "DELETE FROM recognized_devices WHERE user_id=$1 AND token_hash=$2", linked.ID, deviceHash); err != nil {
		t.Fatal(err)
	}
	if rec := post(path+"/devices/block", browserBody, "restriction-browser-key"); rec.Code != 200 {
		t.Fatalf("browser block %d %s", rec.Code, rec.Body.String())
	}
	if rec := post(path+"/devices/block", browserBody, "restriction-browser-key"); rec.Code != 200 || rec.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("browser restriction retry failed")
	}
	for _, id := range []pgtype.UUID{target.ID, linked.ID} {
		if flag, e := q.IsRestrictedAccount(ctx, id); e != nil || !flag {
			t.Fatal("associated account not restricted")
		}
	}
	for _, credential := range []string{targetToken, linkedToken} {
		if _, e := sm.GetSession(ctx, credential); !errors.Is(e, auth.ErrNotAllowed) {
			t.Fatalf("restricted old session allowed %v", e)
		}
	}
	browserRows, e := q.AdminListRestrictedDevices(ctx, db.AdminListRestrictedDevicesParams{Q: "CI restriction browser", PageLimit: 20})
	if e != nil || len(browserRows) != 1 {
		t.Fatal("blocked browser list missing")
	}
	deviceRecord := browserRows[0]
	privateList := get("/restrictions/devices?q=CI%20restriction%20browser")
	if privateList.Code != 200 || strings.Contains(privateList.Body.String(), deviceHash) || strings.Contains(privateList.Body.String(), "token_hash") {
		t.Fatal("browser list exposed identifier")
	}
	if rec := post(path+"/restriction", actionBody{UpdatedAt: version(target.ID), Blocked: &unblocked}, "restriction-account-review"); rec.Code != 200 {
		t.Fatalf("account unblock %d %s", rec.Code, rec.Body.String())
	}
	if _, e = sm.GetSession(ctx, targetToken); !errors.Is(e, auth.ErrSessionNotFound) {
		t.Fatal("unblocking account revived its retired session")
	}
	if flag, e := q.IsRestrictedDevice(ctx, deviceHash); e != nil || !flag {
		t.Fatal("account unblock changed browser restriction")
	}
	devicePath := "/devices/" + fmt.Sprint(deviceRecord.Reference) + "/restriction"
	if rec := post(devicePath, actionBody{UpdatedAt: deviceRecord.UpdatedAt.Time, Blocked: &unblocked}, "restriction-device-review"); rec.Code != 200 {
		t.Fatalf("browser unblock %d %s", rec.Code, rec.Body.String())
	}
	if flag, e := q.IsRestrictedAccount(ctx, linked.ID); e != nil || !flag {
		t.Fatal("browser unblock reactivated associated account")
	}
	if rec := post(path+"/restriction", actionBody{UpdatedAt: version(target.ID), Blocked: &blocked}, "restriction-manual-account"); rec.Code != 200 {
		t.Fatal("manual suspension failed")
	}
	if rec := post(path+"/access", actionBody{UpdatedAt: version(target.ID), Access: &admin}, "restriction-admin-grant"); rec.Code != 409 {
		t.Fatal("restricted account received admin grant")
	}
	if rec := post(path+"/restriction", actionBody{UpdatedAt: version(target.ID), Blocked: &unblocked}, "restriction-manual-review"); rec.Code != 200 {
		t.Fatal("manual account review failed")
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
		// Campaign actions retain automatic audit metadata without requiring a reason.
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
	scheduled := time.Now().Add(2 * time.Hour).UTC()
	body.Configuration = &campaignConfiguration{Design: notifications.CampaignDesign{Heading: "CI <heading>", Accent: "#123456", Blocks: []notifications.CampaignBlock{{Kind: "text", Text: "CI additional paragraph"}}}, Audience: "all", ScheduledAt: &scheduled}

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

	searchPath := "/campaigns/recipients?source=import&q=" + url.QueryEscape(contactEmail)
	if got := call("GET", searchPath, nil, "", false); got.Code != 401 {
		t.Fatalf("anonymous recipient search %d", got.Code)
	}
	if got := call("GET", searchPath, nil, "", true); got.Code != 200 || !strings.Contains(got.Body.String(), contactEmail) || strings.Contains(got.Body.String(), "unsubscribe_hash") {
		t.Fatalf("recipient search %d %s", got.Code, got.Body.String())
	}
	if got := call("GET", strings.Replace(searchPath, "source=import", "source=account", 1), nil, "", true); got.Code != 200 || strings.Contains(got.Body.String(), contactEmail) {
		t.Fatal("recipient search crossed audience source")
	}
	invalidBody := body
	invalidBody.Configuration = &campaignConfiguration{RecipientMode: "selected", Audience: "all"}
	if got := call("POST", "/campaigns/preview", marshal(invalidBody), "http://localhost:3103", true); got.Code != 422 {
		t.Fatalf("empty selected audience accepted %d", got.Code)
	}
	selectedCount, selectionErr := q.CampaignAudienceCountSelected(ctx, db.CampaignAudienceCountSelectedParams{AudienceSource: "import", RecipientEmails: []string{contactEmail}})
	if selectionErr != nil || selectedCount != 1 {
		t.Fatal("specific imported audience was not selected")
	}
	selectedCount, selectionErr = q.CampaignAudienceCountSelected(ctx, db.CampaignAudienceCountSelectedParams{AudienceSource: "account", RecipientEmails: []string{contactEmail}})
	if selectionErr != nil || selectedCount != 0 {
		t.Fatal("audience source restriction was bypassed")
	}
	if got := call("POST", "/campaigns/preview", marshal(body), "http://localhost:3103", true); got.Code != 200 || !strings.Contains(got.Body.String(), "#123456") {
		t.Fatalf("live campaign preview: %d %s", got.Code, got.Body.String())
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
	body.Configuration = nil // Old clients must preserve saved design and delivery settings.
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
	var planned time.Time
	if err = pool.QueryRow(ctx, "SELECT scheduled_at FROM email_outbox WHERE id=$1", outbox).Scan(&planned); err != nil || planned.Before(scheduled.Add(-time.Second)) {
		t.Fatal("campaign schedule was not persisted into the real outbox")
	}
	if !strings.Contains(string(payload), "#123456") {
		t.Fatal("custom design was not preserved through old-client edit and queue")
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
	if got := call("GET", searchPath, nil, "", true); got.Code != 200 || strings.Contains(got.Body.String(), contactEmail) {
		t.Fatal("unsubscribed address returned in search")
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

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM admin_custom_message_deliveries WHERE message_id IN (SELECT id FROM admin_custom_messages WHERE actor_user_id=$1)", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM admin_custom_messages WHERE actor_user_id=$1", actor.ID)
	}()
	custom := adminMessageBody{Subject: "CI service update", Body: "Existing isolated actor message", Recipients: []string{actor.Email, contactEmail}, Confirmed: true}
	messageJSON, _ := json.Marshal(custom)
	if got := call("POST", "/messages/send", messageJSON, "http://localhost:3103", true); got.Code != 422 {
		t.Fatalf("unverified recipient accepted: %d", got.Code)
	}
	custom.Recipients = []string{actor.Email}
	messageJSON, _ = json.Marshal(custom)
	if got := call("POST", "/messages/preview", messageJSON, "https://untrusted.example.invalid", true); got.Code != 403 {
		t.Fatal("message preview allowed untrusted origin")
	}
	if got := call("POST", "/messages/send", messageJSON, "http://localhost:3103", false); got.Code != 401 {
		t.Fatal("anonymous message queued")
	}
	if got := call("POST", "/messages/send", messageJSON, "http://localhost:3103", true); got.Code != 202 {
		t.Fatalf("service queue: %d %s", got.Code, got.Body.String())
	}
	if got := call("POST", "/messages/send", messageJSON, "http://localhost:3103", true); got.Code != 200 {
		t.Fatal("service replay did not succeed")
	}
	var serviceCount int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM admin_custom_message_deliveries d JOIN admin_custom_messages m ON m.id=d.message_id WHERE m.actor_user_id=$1", actor.ID).Scan(&serviceCount); err != nil || serviceCount != 1 {
		t.Fatal("service message was duplicated")
	}
	custom.Body = "Changed message"
	messageJSON, _ = json.Marshal(custom)
	if got := call("POST", "/messages/send", messageJSON, "http://localhost:3103", true); got.Code != 409 {
		t.Fatal("message receipt reused for different content")
	}

}
