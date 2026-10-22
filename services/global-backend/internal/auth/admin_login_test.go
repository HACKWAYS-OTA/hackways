package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDedicatedAdminSignIn(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	nonce, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	password := "isolated-admin-" + nonce[:12]
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	user, err := q.CreateUser(ctx, db.CreateUserParams{Email: "admin-login-" + nonce + "@example.invalid", PasswordHash: hash, DisplayName: "Admin login fixture"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", user.ID)
	}()
	h := NewHandler(pool, false)
	ordinaryToken, _, err := h.sm.CreateSession(ctx, user.ID, "admin-entry-ci", "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	entry := func(cookie *http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		rec := httptest.NewRecorder()
		h.Routes().ServeHTTP(rec, req)
		if rec.Header().Get("Cache-Control") != "private, no-store" {
			t.Fatal("entry response must not be cached")
		}
		return rec
	}
	ordinaryCookie := &http.Cookie{Name: DefaultSessionCookie, Value: ordinaryToken}
	if got := entry(nil).Code; got != 401 {
		t.Fatalf("signed-out entry got %d", got)
	}
	if got := entry(ordinaryCookie).Code; got != 403 {
		t.Fatalf("ordinary account without grant got %d", got)
	}
	if got := entry(&http.Cookie{Name: AdminSessionCookie, Value: ordinaryToken}).Code; got != 403 {
		t.Fatalf("ordinary session relabelled as admin got %d", got)
	}
	if got := entry(&http.Cookie{Name: AdminSessionCookie, Value: "invalid"}).Code; got != 401 {
		t.Fatalf("invalid admin session got %d", got)
	}
	login := func(pass string) *httptest.ResponseRecorder {
		body, encodeErr := json.Marshal(LoginRequest{Email: user.Email, Password: pass})
		if encodeErr != nil {
			t.Fatal(encodeErr)
		}
		req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(string(body)))
		rec := httptest.NewRecorder()
		h.AdminLogin(rec, req)
		return rec
	}
	if got := login("wrong-password").Code; got != 401 {
		t.Fatalf("wrong credentials got %d", got)
	}
	if got := login(password).Code; got != 403 {
		t.Fatalf("account without grant got %d", got)
	}
	if _, err = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES ($1)", user.ID); err != nil {
		t.Fatal(err)
	}
	if got := login(password).Code; got != 403 {
		t.Fatalf("unverified account got %d", got)
	}
	if err = q.MarkUserEmailVerified(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	if got := entry(ordinaryCookie).Code; got != 401 {
		t.Fatalf("administrator without a dedicated admin session got %d", got)
	}
	rec := login(password)
	if rec.Code != 200 {
		t.Fatalf("verified administrator got %d", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatal("expected an admin session cookie and a recognition cookie")
	}
	recognition := cookies[1]
	if recognition.Name != adminDeviceCookie || !recognition.Secure || !recognition.HttpOnly || recognition.Domain != "" || recognition.Path != "/" || recognition.SameSite != http.SameSiteStrictMode || recognition.MaxAge != int(deviceLifetime.Seconds()) {
		t.Fatal("incorrect admin device recognition cookie policy")
	}
	if strings.Contains(rec.Body.String(), recognition.Value) {
		t.Fatal("device recognition token exposed in response")
	}
	cookie := cookies[0]
	if cookie.Name != AdminSessionCookie || !cookie.Secure || !cookie.HttpOnly || cookie.Domain != "" || cookie.Path != "/" || cookie.SameSite != http.SameSiteStrictMode || cookie.MaxAge != 1800 {
		t.Fatal("incorrect admin cookie policy")
	}
	if strings.Contains(rec.Body.String(), cookie.Value) {
		t.Fatal("admin token exposed in response")
	}
	session, err := h.sm.GetSession(ctx, cookie.Value)
	if err != nil {
		t.Fatal(err)
	}
	allowed, err := q.IsPlatformAdminSession(ctx, db.IsPlatformAdminSessionParams{SessionID: session.SessionID, UserID: user.ID})
	if err != nil || !allowed {
		t.Fatal("dedicated admin session missing")
	}
	meReq := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
	meReq.AddCookie(cookie)
	meRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(meRec, meReq)
	if meRec.Code != 200 {
		t.Fatalf("admin identity got %d", meRec.Code)
	}
	if strings.Contains(meRec.Body.String(), cookie.Value) {
		t.Fatal("identity exposed admin token")
	}
	logoutReq := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	logoutReq.AddCookie(cookie)
	logoutRec := httptest.NewRecorder()
	h.AdminLogout(logoutRec, logoutReq)
	if logoutRec.Code != 204 {
		t.Fatalf("logout got %d", logoutRec.Code)
	}
	expiredMe := httptest.NewRecorder()
	h.Routes().ServeHTTP(expiredMe, meReq)
	if expiredMe.Code != 401 {
		t.Fatalf("logged-out identity got %d", expiredMe.Code)
	}
	if _, err = h.sm.GetSession(ctx, cookie.Value); err == nil {
		t.Fatal("logged-out session remains active")
	}
}
