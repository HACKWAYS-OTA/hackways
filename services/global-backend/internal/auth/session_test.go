package auth

import (
	"context"
	"github.com/hackways/global-backend/internal/testsupport"
	"net/http"
	"net/http/httptest"
	"net/netip"

	"testing"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCookieHelpers(t *testing.T) {
	rec := httptest.NewRecorder()
	token := "dummy-session-token-64-characters-long-hex-encoded-value-string"

	// 1. Test Attendee Cookie
	SetSessionCookie(rec, DefaultSessionCookie, token, 24*time.Hour, true)
	resp := rec.Result()
	cookies := resp.Cookies()

	if len(cookies) == 0 {
		t.Fatalf("expected Set-Cookie header, got none")
	}

	cookie := cookies[0]
	if cookie.Name != DefaultSessionCookie {
		t.Errorf("expected cookie name %s, got %s", DefaultSessionCookie, cookie.Name)
	}
	if cookie.Value != token {
		t.Errorf("expected cookie value %s, got %s", token, cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Errorf("expected HttpOnly to be true")
	}
	if !cookie.Secure {
		t.Errorf("expected Secure to be true for __Host- cookie")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite Lax, got %v", cookie.SameSite)
	}

	// 2. Test Admin Cookie
	recAdmin := httptest.NewRecorder()
	SetSessionCookie(recAdmin, AdminSessionCookie, token, 1*time.Hour, true)
	adminCookie := recAdmin.Result().Cookies()[0]
	if adminCookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("expected SameSite Strict for admin cookie, got %v", adminCookie.SameSite)
	}

	// 3. Test Clear Cookie
	recClear := httptest.NewRecorder()
	ClearSessionCookie(recClear, DefaultSessionCookie)
	clearCookie := recClear.Result().Cookies()[0]
	if clearCookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1 on clear, got %d", clearCookie.MaxAge)
	}
}

func TestTokenGeneration(t *testing.T) {
	token1, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(token1) != 64 {
		t.Errorf("expected 64 hex characters (32 bytes), got %d", len(token1))
	}
	if token1 == token2 {
		t.Errorf("tokens must be randomly unique")
	}

	hash1 := HashToken(token1)
	if len(hash1) != 64 {
		t.Errorf("expected SHA-256 hash length 64, got %d", len(hash1))
	}
}

func TestSessionManager_LivePostgres(t *testing.T) {
	dbURL := testsupport.DatabaseURL(t)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping live postgres session test: %v", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skipping live postgres session test: %v", err)
		return
	}

	queries := db.New(pool)
	sm := NewSessionManager(queries)

	// Create test user
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        "session-test@hackways.in",
		PasswordHash: "$2a$12$somevalidhashherethatislongenough",
		DisplayName:  "Session Tester",
	})
	if err != nil {
		// If user exists, fetch user
		fetched, getErr := queries.GetUserByEmail(ctx, "session-test@hackways.in")
		if getErr != nil {
			t.Fatalf("failed to create or get test user: %v", err)
		}
		user.ID = fetched.ID
		user.Email = fetched.Email
		user.DisplayName = fetched.DisplayName
	}

	// 1. Create session
	token, session, err := sm.CreateSession(ctx, user.ID, "Mozilla/5.0 Test Agent", "127.0.0.1", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if token == "" || session == nil {
		t.Fatalf("expected non-empty token and session")
	}

	// 2. Lookup session with valid token
	sessionUser, err := sm.GetSession(ctx, token)
	if err != nil {
		t.Fatalf("failed to get active session: %v", err)
	}
	if sessionUser.Email != "session-test@hackways.in" {
		t.Errorf("expected email 'session-test@hackways.in', got %s", sessionUser.Email)
	}

	// 3. Lookup session with invalid token
	_, err = sm.GetSession(ctx, "invalid-tampered-token-1234567890")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for tampered token, got %v", err)
	}

	// 4. Revoke session
	if err := sm.RevokeSession(ctx, token); err != nil {
		t.Fatalf("failed to revoke session: %v", err)
	}

	// 5. Lookup after revocation -> must fail
	_, err = sm.GetSession(ctx, token)
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound after revocation, got %v", err)
	}
}

func TestSignInDeviceAndLocationMetadata(t *testing.T) {
	details := describeDevice("Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/130.0.0.0 Safari/537.36 Edg/130.0.2849.46")
	if details.device != "Windows PC" || details.browser != "Microsoft Edge 130.0.2849.46" {
		t.Fatal("device/browser not identified")
	}
	r := httptest.NewRequest("POST", "/login", nil)
	r.RemoteAddr = "192.0.2.8:1234"
	r.Header.Set("X-Hackways-City", "Bengaluru")
	r.Header.Set("X-Hackways-Region", "Karnataka")
	r.Header.Set("X-Hackways-Country", "IN")
	if approximateLocation(r, nil) != "Unavailable" {
		t.Fatal("untrusted location accepted")
	}
	proxies := []netip.Prefix{netip.MustParsePrefix("192.0.2.8/32")}
	if approximateLocation(r, proxies) != "Bengaluru, Karnataka, IN" {
		t.Fatal("trusted location missing")
	}
	r.Header.Set("X-Hackways-City", "<script>")
	if approximateLocation(r, proxies) != "Karnataka, IN" {
		t.Fatal("unsafe city accepted")
	}
	r.Header.Set("X-Hackways-Country", "XX")
	if approximateLocation(r, proxies) != "Unavailable" {
		t.Fatal("unknown geography invented")
	}
}
