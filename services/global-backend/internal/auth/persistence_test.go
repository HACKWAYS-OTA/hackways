package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
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
}

func TestSessionActivityQueueIsBounded(t *testing.T) {
	recorder := &activityRecorder{queue: make(chan pgtype.UUID, 1)}
	recorder.Mark(pgtype.UUID{})
	recorder.Mark(pgtype.UUID{})
	if len(recorder.queue) != 1 || recorder.dropped.Load() != 1 {
		t.Fatal("activity queue failed its bound")
	}
}
