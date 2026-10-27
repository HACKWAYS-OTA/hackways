package admin

// These database-backed authorization checks run in the mandatory CI workflow.
import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAdminCatalogueAccess(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	sm := auth.NewSessionManager(q)
	nonce, err := auth.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	user, err := q.CreateUser(ctx, db.CreateUserParams{Email: "admin-review-" + nonce + "@example.invalid", PasswordHash: "isolated-fixture", DisplayName: "Admin review fixture"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", user.ID)
	}()
	token, session, err := sm.CreateSession(ctx, user.ID, "mandatory-ci", "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	router := NewHandler(pool).Routes()
	request := func(path, cookie string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: cookie, Value: token})
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	routes := []string{"events", "spaces", "registrations", "tickets", "moderation", "audit", "bin"}
	for _, route := range routes {
		for _, cookie := range []string{"", auth.DefaultSessionCookie} {
			if got := request("/"+route, cookie).Code; got != 401 {
				t.Fatalf("%s ordinary request got %d", route, got)
			}
		}
		if got := request("/"+route, auth.AdminSessionCookie).Code; got != 403 {
			t.Fatalf("%s without grant got %d", route, got)
		}
	}
	if _, err = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES ($1)", user.ID); err != nil {
		t.Fatal(err)
	}
	if err = q.RecordPlatformAdminSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}

	if got := request("/"+routes[0], auth.AdminSessionCookie).Code; got != 403 {
		t.Fatalf("unverified account got %d", got)
	}
	if _, err = pool.Exec(ctx, "UPDATE users SET email_verified=TRUE WHERE id=$1", user.ID); err != nil {
		t.Fatal(err)
	}
	for _, route := range routes {
		rec := request("/"+route+"?limit=1&offset=0", auth.AdminSessionCookie)
		if rec.Code != 200 {
			t.Fatalf("%s granted read got %d", route, rec.Code)
		}
		if rec.Header().Get("Cache-Control") != "private, no-store" {
			t.Fatal("private reads must not be cached")
		}
		var envelope struct {
			Data []json.RawMessage `json:"data"`
		}
		if json.Unmarshal(rec.Body.Bytes(), &envelope) != nil || envelope.Data == nil || len(envelope.Data) > 1 {
			t.Fatalf("%s invalid paginated collection", route)
		}
		for _, forbidden := range []string{"password_hash", "token_hash", "payload_json", "lease_token", "error_message"} {
			if strings.Contains(rec.Body.String(), forbidden) {
				t.Fatalf("%s exposed private transport field", route)
			}
		}
		if got := request("/"+route+"?limit=1000", auth.AdminSessionCookie).Code; got != 400 {
			t.Fatalf("%s accepted unbounded page: %d", route, got)
		}
		if got := request("/"+route+"?q="+strings.Repeat("x", 201), auth.AdminSessionCookie).Code; got != 400 {
			t.Fatalf("%s accepted oversized query: %d", route, got)
		}
	}
	if _, err = pool.Exec(ctx, "DELETE FROM platform_admin_sessions WHERE session_id=$1", session.ID); err != nil {
		t.Fatal(err)
	}
	if got := request("/"+routes[0], auth.AdminSessionCookie).Code; got != 403 {
		t.Fatalf("ordinary session relabelled as admin got %d", got)
	}
	if err = q.RecordPlatformAdminSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	// Both engines independently gate an overdue daily verification.
	if _, err = pool.Exec(ctx, "UPDATE platform_admin_sessions SET otp_verified_at=clock_timestamp()-interval '25 hours' WHERE session_id=$1", session.ID); err != nil {
		t.Fatal(err)
	}
	stepup := request("/"+routes[0], auth.AdminSessionCookie)
	if stepup.Code != 403 || !strings.Contains(stepup.Body.String(), "ADMIN_VERIFICATION_REQUIRED") {
		t.Fatal("core admin API did not require daily verification")
	}
	if _, err = pool.Exec(ctx, "UPDATE platform_admin_sessions SET otp_verified_at=clock_timestamp() WHERE session_id=$1", session.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "UPDATE platform_administrators SET revoked_at=NOW() WHERE user_id=$1", user.ID); err != nil {
		t.Fatal(err)
	}
	if got := request("/"+routes[0], auth.AdminSessionCookie).Code; got != 403 {
		t.Fatalf("revoked grant got %d", got)
	}
}
