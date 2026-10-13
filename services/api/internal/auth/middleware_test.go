package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/hackways/api/db"
	"github.com/hackways/api/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRequireAuth_NegativeTests(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://hackways_user:hackways_password@localhost:5432/hackways_dev?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping middleware test: %v", err)
		return
	}
	defer pool.Close()

	queries := db.New(pool)
	sm := NewSessionManager(queries)
	authMiddleware := RequireAuth(sm)

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok || user == nil {
			t.Errorf("expected user in context")
		}
		platform.RespondJSON(w, http.StatusOK, map[string]string{"status": "authorized"})
	})

	protected := authMiddleware(dummyHandler)

	// 1. Negative Test: No cookie present -> 401 Unauthorized
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	rec1 := httptest.NewRecorder()
	protected.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 without cookie, got %d", rec1.Code)
	}

	var env1 platform.ErrorEnvelope
	_ = json.NewDecoder(rec1.Body).Decode(&env1)
	if env1.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %q", env1.Error.Code)
	}

	// 2. Negative Test: Tampered cookie -> 401 Unauthorized
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req2.AddCookie(&http.Cookie{
		Name:  DefaultSessionCookie,
		Value: "tampered_fake_hex_token_1234567890abcdef1234567890abcdef",
	})
	rec2 := httptest.NewRecorder()
	protected.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 with tampered cookie, got %d", rec2.Code)
	}

	// 3. Positive Test: Valid cookie -> 200 OK
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        "middleware-test@hackways.in",
		PasswordHash: "$2a$12$somevalidhashherethatislongenough",
		DisplayName:  "Middleware Tester",
	})
	if err != nil {
		fetched, _ := queries.GetUserByEmail(ctx, "middleware-test@hackways.in")
		user.ID = fetched.ID
	}

	token, _, err := sm.CreateSession(ctx, user.ID, "Test Agent", "127.0.0.1", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req3.AddCookie(&http.Cookie{
		Name:  DefaultSessionCookie,
		Value: token,
	})
	rec3 := httptest.NewRecorder()
	protected.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusOK {
		t.Errorf("expected status 200 with valid session, got %d", rec3.Code)
	}
}

func TestRequireOrganizer_NegativeTests(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://hackways_user:hackways_password@localhost:5432/hackways_dev?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping organizer test: %v", err)
		return
	}
	defer pool.Close()

	queries := db.New(pool)
	sm := NewSessionManager(queries)
	authMiddleware := RequireAuth(sm)
	organizerMiddleware := RequireOrganizer(queries)

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		platform.RespondJSON(w, http.StatusOK, map[string]string{"status": "organizer_ok"})
	})

	protected := authMiddleware(organizerMiddleware(dummyHandler))

	// User without organizer profile
	nonOrganizerUser, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        "attendee-only@hackways.in",
		PasswordHash: "$2a$12$somevalidhashherethatislongenough",
		DisplayName:  "Attendee Only",
	})
	if err != nil {
		fetched, _ := queries.GetUserByEmail(ctx, "attendee-only@hackways.in")
		nonOrganizerUser.ID = fetched.ID
	}

	token, _, err := sm.CreateSession(ctx, nonOrganizerUser.ID, "Test Agent", "127.0.0.1", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizer/studio", nil)
	req.AddCookie(&http.Cookie{
		Name:  DefaultSessionCookie,
		Value: token,
	})
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)

	// Negative Test: Must reject non-organizer with 403 Forbidden
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403 Forbidden for non-organizer, got %d", rec.Code)
	}

	var env platform.ErrorEnvelope
	_ = json.NewDecoder(rec.Body).Decode(&env)
	if env.Error.Code != "FORBIDDEN" {
		t.Errorf("expected error code FORBIDDEN, got %q", env.Error.Code)
	}
}
