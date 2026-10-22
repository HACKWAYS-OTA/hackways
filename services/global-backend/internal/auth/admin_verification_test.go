package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAdminDailyVerificationAndPerDeviceRevocation(t *testing.T) {
	t.Setenv("ACS_ENDPOINT", "https://admin-otp-ci.example.invalid")
	t.Setenv("ACS_ACCESS_KEY", "ci-does-not-send")
	t.Setenv("ACS_SENDER_ADDRESS", "sender@example.invalid")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	user, err := q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("admin-otp-%d@example.invalid", time.Now().UnixNano()), PasswordHash: "isolated-ci-only", DisplayName: "Admin OTP CI"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE recipient_email=$1", user.Email)
		_, _ = pool.Exec(ctx, "DELETE FROM audit_log WHERE actor_user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM recognized_devices WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", user.ID)
	}()
	if err = q.MarkUserEmailVerified(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES($1)", user.ID); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(pool, false)
	var devices []*http.Cookie
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/admin/login", nil)
		if err = h.establishSession(w, r, user.ID, "isolated-ci-only", true); err != nil {
			t.Fatal(err)
		}
		for _, cookie := range w.Result().Cookies() {
			if cookie.Name == AdminSessionCookie {
				devices = append(devices, cookie)
				if cookie.MaxAge != 90*24*60*60 {
					t.Fatal("admin persistence is not ninety days")
				}
			}
		}
	}
	call := func(method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Origin", "https://admin.hackways.in")
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		h.Routes().ServeHTTP(w, r)
		return w
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("expected %d got %d: %s", status, w.Code, w.Body.String())
		}
	}
	initial := call("GET", "/admin/me", "", devices[0])
	check(initial, 403)
	if !strings.Contains(initial.Body.String(), "ADMIN_VERIFICATION_REQUIRED") {
		t.Fatal("missing daily verification distinction")
	}
	readCode := func() string {
		t.Helper()
		var raw []byte
		if err = pool.QueryRow(ctx, "SELECT payload_json FROM email_outbox WHERE recipient_email=$1 AND template_type='admin_daily_otp' ORDER BY created_at DESC LIMIT 1", user.Email).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var payload otpPayload
		if err = json.Unmarshal(raw, &payload); err != nil {
			t.Fatal(err)
		}
		if !validEmailCode(payload.Code) {
			t.Fatal("invalid generated challenge")
		}
		return payload.Code
	}
	check(call("POST", "/admin/verification/request", "{}", devices[0]), 202)
	code := readCode()
	check(call("POST", "/admin/verification/request", "{}", devices[0]), 429)
	bad := "000000"
	if bad == code {
		bad = "999999"
	}
	check(call("POST", "/admin/verification/confirm", `{"code":"`+bad+`"}`, devices[0]), 401)
	if _, err = h.sm.GetSession(ctx, devices[0].Value); !errors.Is(err, ErrSessionNotFound) {
		t.Fatal("incorrect code did not revoke that device")
	}
	if _, err = h.sm.GetSession(ctx, devices[1].Value); err != nil {
		t.Fatal("incorrect code revoked another device")
	}
	check(call("POST", "/admin/verification/request", "{}", devices[1]), 202)
	code = readCode()
	verified := call("POST", "/admin/verification/confirm", `{"code":"`+code+`"}`, devices[1])
	check(verified, 204)
	if verified.Header().Get("X-Admin-Verification-Seconds") == "" {
		t.Fatal("verification deadline is missing")
	}
	check(call("GET", "/admin/me", "", devices[1]), 200)
	check(call("POST", "/admin/verification/confirm", `{"code":"`+code+`"}`, devices[1]), 204)
	current, err := h.sm.GetSession(ctx, devices[1].Value)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "UPDATE platform_admin_sessions SET otp_verified_at=clock_timestamp()-interval '25 hours' WHERE session_id=$1", current.SessionID); err != nil {
		t.Fatal(err)
	}
	check(call("GET", "/admin/me", "", devices[1]), 403)
	check(call("GET", "/admin/session", "", devices[1]), 403)
	t.Setenv("ACS_ENDPOINT", "")
	t.Setenv("ACS_ACCESS_KEY", "")
	check(call("POST", "/admin/verification/request", "{}", devices[1]), 503)
	if _, err = h.sm.GetSession(ctx, devices[1].Value); err != nil {
		t.Fatal("email outage revoked a valid session")
	}
	t.Setenv("ACS_ENDPOINT", "https://admin-otp-ci.example.invalid")
	t.Setenv("ACS_ACCESS_KEY", "ci-does-not-send")
	if _, err = pool.Exec(ctx, "UPDATE admin_session_challenges SET consumed_at=NULL,expires_at=clock_timestamp()-interval '1 minute' WHERE session_id=$1", current.SessionID); err != nil {
		t.Fatal(err)
	}
	check(call("POST", "/admin/verification/confirm", `{"code":"`+code+`"}`, devices[1]), 401)
	check(call("GET", "/admin/me", "", devices[1]), 401)
}
