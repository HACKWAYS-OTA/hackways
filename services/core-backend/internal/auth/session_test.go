package auth

import (
	"context"
	"errors"
	"github.com/hackways/core-backend/internal/testsupport"
	"net/http"
	"net/http/httptest"
	"strings"

	"testing"
	"time"

	"github.com/hackways/core-backend/db"
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

	// Mandatory isolated persistence coverage: both engines reject restriction
	// state, old sessions remain retired after review, and guests cannot bypass it.
	if _, err = pool.Exec(ctx, "INSERT INTO account_restrictions(user_id,reason,updated_by) VALUES($1,'Isolated core account restriction',$1) ON CONFLICT(user_id) DO UPDATE SET blocked=true", user.ID); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = pool.Exec(ctx, "DELETE FROM account_restrictions WHERE user_id=$1", user.ID) }()
	if _, err = sm.GetSession(ctx, token); !errors.Is(err, ErrNotAllowed) {
		t.Fatal("core accepted restricted account")
	}
	request := httptest.NewRequest("POST", "/api/v1/rsvps", nil)
	request.AddCookie(&http.Cookie{Name: DefaultSessionCookie, Value: token})
	denied := httptest.NewRecorder()
	RequireAuth(sm)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("restricted account reached action") })).ServeHTTP(denied, request)
	if denied.Code != 403 || !strings.Contains(denied.Body.String(), "You are not allowed") {
		t.Fatal("core restriction response missing")
	}
	if _, err = pool.Exec(ctx, "UPDATE account_restrictions SET blocked=false WHERE user_id=$1;", user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "UPDATE sessions SET revoked_at=clock_timestamp() WHERE id=$1", session.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = sm.GetSession(ctx, token); !errors.Is(err, ErrSessionNotFound) {
		t.Fatal("core revived retired session")
	}
	browser, _ := GenerateToken()
	deviceHash := HashToken(browser)
	if _, err = pool.Exec(ctx, "INSERT INTO device_restrictions(token_hash,reason,updated_by) VALUES($1,'Isolated core browser restriction',$2)", deviceHash, user.ID); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = pool.Exec(ctx, "DELETE FROM device_restrictions WHERE token_hash=$1", deviceHash) }()
	request = httptest.NewRequest("POST", "/api/v1/rsvps", nil)
	request.AddCookie(&http.Cookie{Name: "__Host-hackways_device", Value: browser})
	denied = httptest.NewRecorder()
	RestrictionGuard(queries)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("blocked guest reached action") })).ServeHTTP(denied, request)
	if denied.Code != 403 {
		t.Fatalf("blocked guest got%d", denied.Code)
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
