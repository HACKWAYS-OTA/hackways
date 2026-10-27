package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPersistentSessionsMultipleDevicesAndCookieRefresh(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	user, err := q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("persistent-session-%d@example.invalid", time.Now().UnixNano()), PasswordHash: "isolated-ci-only", DisplayName: "Session CI"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM recognized_devices WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", user.ID)
	}()
	handler := NewHandler(pool, false)
	sm := NewSessionManager(q)
	var tokens []string
	for i := 0; i < 2; i++ {
		r := httptest.NewRequest("POST", "/login", nil)
		r.Header.Set("User-Agent", fmt.Sprint("CI browser ", i))
		w := httptest.NewRecorder()
		if err = handler.establishSession(w, r, user.ID, "isolated-ci-only", false); err != nil {
			t.Fatal(err)
		}
		for _, cookie := range w.Result().Cookies() {
			if cookie.Name == DefaultSessionCookie {
				tokens = append(tokens, cookie.Value)
				if cookie.MaxAge != 90*24*60*60 || !cookie.Secure || !cookie.HttpOnly {
					t.Fatal("invalid persistent cookie")
				}
			}
		}
	}
	if len(tokens) != 2 {
		t.Fatal("two browsers must receive independent sessions")
	}
	for _, token := range tokens {
		session, e := sm.GetSession(ctx, token)
		if e != nil || time.Until(session.ExpiresAt) < UserSessionLifetime-time.Minute {
			t.Fatal("new device invalidated an earlier session or shortened expiry")
		}
	}
	first, err := q.GetSessionByTokenHash(ctx, HashToken(tokens[0]))
	if err != nil {
		t.Fatal(err)
	}
	// Migration extends active legacy ordinary sessions only.
	_, err = pool.Exec(ctx, "UPDATE sessions SET expires_at=clock_timestamp()+interval '7 days' WHERE id=$1", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, adminSession, err := sm.CreateSession(ctx, user.ID, "CI admin", "", 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err = q.RecordPlatformAdminSession(ctx, adminSession.ID); err != nil {
		t.Fatal(err)
	}
	expiredToken, _, err := sm.CreateSession(ctx, user.ID, "CI expired", "", -time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../../../../deploy/migrations/000018_persistent_user_sessions.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, string(migration)); err != nil {
		t.Fatal(err)
	}
	if _, err = sm.GetSession(ctx, expiredToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatal("expired session was revived")
	}
	var adminExpiry time.Time
	if err = pool.QueryRow(ctx, "SELECT expires_at FROM sessions WHERE id=$1", adminSession.ID).Scan(&adminExpiry); err != nil {
		t.Fatal(err)
	}
	if adminExpiry.Sub(adminSession.ExpiresAt.Time) > time.Second {
		t.Fatal("ordinary policy extended dedicated admin session")
	}
	r := httptest.NewRequest("GET", "/session", nil)
	r.AddCookie(&http.Cookie{Name: DefaultSessionCookie, Value: tokens[0]})
	w := httptest.NewRecorder()
	handler.Routes().ServeHTTP(w, r)
	if w.Code != 204 {
		t.Fatalf("refresh got %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != tokens[0] || cookies[0].MaxAge < int((UserSessionLifetime-time.Minute).Seconds()) {
		t.Fatal("refresh changed credential or lost legacy cookie persistence")
	}
	activityCtx, cancel := context.WithCancel(ctx)
	done := StartSessionActivity(activityCtx, q, slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer func() { cancel(); <-done }()
	_, err = pool.Exec(ctx, "UPDATE sessions SET last_seen=clock_timestamp()-interval '10 minutes' WHERE id=$1", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = sm.GetSession(ctx, tokens[0]); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(4 * time.Second)
	recorded := false
	for time.Now().Before(deadline) {
		var seen time.Time
		if err = pool.QueryRow(ctx, "SELECT last_seen FROM sessions WHERE id=$1", first.ID).Scan(&seen); err != nil {
			t.Fatal(err)
		}
		if time.Since(seen) < time.Minute {
			recorded = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !recorded {
		t.Fatal("bounded asynchronous heartbeat was not recorded")
	}
	if err = sm.RevokeSession(ctx, tokens[0]); err != nil {
		t.Fatal(err)
	}
	if err = q.UpdateSessionActivityBatch(ctx, []pgtype.UUID{first.ID}); err != nil {
		t.Fatal(err)
	}
	r = httptest.NewRequest("GET", "/session", nil)
	r.AddCookie(&http.Cookie{Name: DefaultSessionCookie, Value: tokens[0]})
	w = httptest.NewRecorder()
	handler.Routes().ServeHTTP(w, r)
	if w.Code != 401 || len(w.Result().Cookies()) != 0 {
		t.Fatal("revoked session was renewed")
	}
	if _, err = sm.GetSession(ctx, tokens[1]); err != nil {
		t.Fatal("revoking one browser affected the other")
	}
	t.Run("blocked browser account switch persists quarantine", func(t *testing.T) {
		password := "RestrictionProof123!"
		hash, e := HashPassword(password)
		if e != nil {
			t.Fatal(e)
		}
		other, e := q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("restriction-switch-%d@example.invalid", time.Now().UnixNano()), PasswordHash: hash, DisplayName: "Switch CI"})
		if e != nil {
			t.Fatal(e)
		}
		browser, e := GenerateToken()
		if e != nil {
			t.Fatal(e)
		}
		deviceHash := HashToken(browser)
		device, e := q.AdminBlockDevice(ctx, db.AdminBlockDeviceParams{TokenHash: deviceHash, UserAgent: "Isolated browser", Reason: "Mandatory isolated restriction regression", UpdatedBy: user.ID})
		if e != nil {
			t.Fatal(e)
		}
		defer func() {
			_, _ = pool.Exec(ctx, "DELETE FROM audit_log WHERE entity_id=$1", other.ID)
			_, _ = pool.Exec(ctx, "DELETE FROM restriction_device_accounts WHERE device_id=$1", device.ID)
			_, _ = pool.Exec(ctx, "DELETE FROM account_restrictions WHERE user_id=$1", other.ID)
			_, _ = pool.Exec(ctx, "DELETE FROM device_restrictions WHERE id=$1", device.ID)
			_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", other.ID)
			_, _ = pool.Exec(ctx, "DELETE FROM recognized_devices WHERE user_id=$1", other.ID)
			_, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE recipient_email=$1", other.Email)
			_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", other.ID)
		}()
		credential, _, e := sm.CreateSession(ctx, other.ID, "Other browser", "", time.Hour)
		if e != nil {
			t.Fatal(e)
		}
		call := func(path, password string) *httptest.ResponseRecorder {
			payload := map[string]string{"email": other.Email, "password": password}
			if path == "/register" {
				payload["display_name"] = "Blocked signup"
			}
			raw, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", path, strings.NewReader(string(raw)))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: deviceCookie, Value: browser})
			out := httptest.NewRecorder()
			handler.Routes().ServeHTTP(out, req)
			return out
		}
		if rec := call("/login", "wrong-password"); rec.Code != 401 {
			t.Fatalf("wrong proof %d", rec.Code)
		}
		if flag, e := q.IsRestrictedAccount(ctx, other.ID); e != nil || flag {
			t.Fatal("wrong credentials suspended supplied account")
		}
		rec := call("/login", password)
		if rec.Code != 403 || !strings.Contains(rec.Body.String(), "You are not allowed") {
			t.Fatalf("blocked login %d %s", rec.Code, rec.Body.String())
		}
		if flag, e := q.IsRestrictedAccount(ctx, other.ID); e != nil || !flag {
			t.Fatal("quarantine rolled back on denial")
		}
		if _, e = sm.GetSession(ctx, credential); !errors.Is(e, ErrNotAllowed) {
			t.Fatal("old account2 session stayed active")
		}
		if e = q.AdminSetAccountRestriction(ctx, db.AdminSetAccountRestrictionParams{UserID: other.ID, Blocked: false, Reason: "Isolated account review", UpdatedBy: user.ID}); e != nil {
			t.Fatal(e)
		}
		if _, e = sm.GetSession(ctx, credential); !errors.Is(e, ErrSessionNotFound) {
			t.Fatal("unblock revived retired credential")
		}
		// A valid OAuth exchange on the same blocked browser must persist the same
		// quarantine, consume its proof and never issue a session.
		newGoogle := googleUserInfo{Sub: "blocked-new-" + browser, Email: "new-" + other.Email, Name: "Blocked new Google", VerifiedEmail: true}
		if _, e = handler.upsertGoogleIdentity(ctx, newGoogle, HashToken("clean-"+browser), deviceHash); !errors.Is(e, ErrNotAllowed) {
			t.Fatalf("blocked Google signup %v", e)
		}
		var created int
		if e = pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE email=$1", newGoogle.Email).Scan(&created); e != nil || created != 0 {
			t.Fatal("blocked Google signup persisted user")
		}
		existingGoogle := googleUserInfo{Sub: "blocked-existing-" + browser, Email: other.Email, Name: other.DisplayName, VerifiedEmail: true}
		if found, e := handler.upsertGoogleIdentity(ctx, existingGoogle, deviceHash); e != nil || found != other.ID {
			t.Fatal("existing Google proof could not reach quarantine")
		}
		if e = pool.QueryRow(ctx, "SELECT count(*) FROM user_identities WHERE provider='google' AND subject=$1", existingGoogle.Sub).Scan(&created); e != nil || created != 0 {
			t.Fatal("blocked Google proof linked identity")
		}
		exchange, nonce := "ci-exchange-"+browser, "ci-nonce-"+browser
		if e = q.CreateOAuthExchangeToken(ctx, db.CreateOAuthExchangeTokenParams{TokenHash: HashToken(exchange), UserID: other.ID, NonceHash: HashToken(nonce), ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Minute), Valid: true}, DeviceTokenHash: pgtype.Text{String: HashToken("clean-" + browser), Valid: true}, CallbackDeviceTokenHash: pgtype.Text{String: deviceHash, Valid: true}}); e != nil {
			t.Fatal(e)
		}
		req := httptest.NewRequest("POST", "/google/exchange", nil)
		out := httptest.NewRecorder()
		e = handler.establishSession(out, req, other.ID, "", false, db.ConsumeOAuthExchangeTokenParams{TokenHash: HashToken(exchange), NonceHash: HashToken(nonce)})
		if !errors.Is(e, ErrNotAllowed) || len(out.Result().Cookies()) != 0 {
			t.Fatalf("OAuth restriction bypass %v", e)
		}
		if flag, e := q.IsRestrictedAccount(ctx, other.ID); e != nil || !flag {
			t.Fatal("OAuth quarantine did not commit")
		}
		if e = q.AdminUnblockDevice(ctx, db.AdminUnblockDeviceParams{ID: device.ID, Reason: "Isolated browser review", UpdatedBy: user.ID}); e != nil {
			t.Fatal(e)
		}
		if rec = call("/login", password); rec.Code != 403 {
			t.Fatal("browser unblock reactivated account")
		}
		if e = q.AdminSetAccountRestriction(ctx, db.AdminSetAccountRestrictionParams{UserID: other.ID, Blocked: false, Reason: "Isolated final account review", UpdatedBy: user.ID}); e != nil {
			t.Fatal(e)
		}
		if rec = call("/login", password); rec.Code != 200 {
			t.Fatalf("reviewed sign-in failed %d %s", rec.Code, rec.Body.String())
		}
		if _, e = sm.GetSession(ctx, credential); !errors.Is(e, ErrSessionNotFound) {
			t.Fatal("new sign-in revived old credential")
		}
		logout := call("/logout", password)
		for _, cookie := range logout.Result().Cookies() {
			if cookie.Name == deviceCookie && cookie.MaxAge < 0 {
				t.Fatal("logout cleared recognition")
			}
		}
		// A blocked browser cannot create even an unverified account.
		if _, e = q.AdminBlockDevice(ctx, db.AdminBlockDeviceParams{TokenHash: deviceHash, UserAgent: "Isolated browser", Reason: "Isolated signup rejection", UpdatedBy: user.ID}); e != nil {
			t.Fatal(e)
		}
		if rec = call("/register", password); rec.Code != 403 {
			t.Fatal("blocked browser registration was not denied first")
		}
	})

}

func TestSessionActivityQueueIsBounded(t *testing.T) {
	recorder := &activityRecorder{queue: make(chan pgtype.UUID, 1)}
	recorder.Mark(pgtype.UUID{})
	recorder.Mark(pgtype.UUID{})
	if len(recorder.queue) != 1 || recorder.dropped.Load() != 1 {
		t.Fatal("activity queue failed its bound")
	}
}
