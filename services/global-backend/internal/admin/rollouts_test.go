package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/rollouts"
	"github.com/hackways/global-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRolloutVersionsAuthorizationAndRestore(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	original, err := q.ReadRuntimeRollout(ctx)
	if err != nil {
		t.Fatal(err)
	}
	actor, err := q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("rollout-%d@example.invalid", time.Now().UnixNano()), PasswordHash: "isolated-ci-only", DisplayName: "Rollout CI"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "UPDATE runtime_rollout_control SET active_version=$1,revision=$2", original.Reference, original.Revision)
		_, _ = pool.Exec(ctx, "DELETE FROM audit_log WHERE actor_user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM runtime_rollout_versions WHERE created_by=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM idempotency_records WHERE scope_hash=$1", actionHash([]byte(fmt.Sprintf("admin-rollout:%x", actor.ID.Bytes))))
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", actor.ID)
	}()
	sm := auth.NewSessionManager(q)
	token, session, err := sm.CreateSession(ctx, actor.ID, "CI", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	cache := rollouts.New(q)
	if err = cache.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	router := NewHandler(pool, cache).Routes()
	call := func(method, body, key, cookie, origin string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/rollouts", strings.NewReader(body))
		if cookie != "" {
			r.AddCookie(&http.Cookie{Name: cookie, Value: token})
		}
		r.Header.Set("Origin", origin)
		r.Header.Set("X-Admin-Reason", "Mandatory%20CI%20configuration%20review")
		r.Header.Set("Idempotency-Key", key)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("expected %d got %d: %s", status, w.Code, w.Body.String())
		}
	}
	check(call("GET", "", "", auth.DefaultSessionCookie, ""), 401)
	check(call("GET", "", "", auth.AdminSessionCookie, ""), 403)
	_, err = pool.Exec(ctx, "UPDATE users SET email_verified=true WHERE id=$1", actor.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES($1)", actor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = q.RecordPlatformAdminSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	read := call("GET", "", "", auth.AdminSessionCookie, "")
	check(read, 200)
	if strings.Contains(read.Body.String(), "cohort_salt") || strings.Contains(read.Body.String(), "password_hash") {
		t.Fatal("private configuration leaked")
	}
	body := fmt.Sprintf(`{"revision":%d,"label":"CI 10 percent","flags":{"manage.overview.v2":{"enabled":true,"percentage":10}}}`, original.Revision)
	check(call("POST", body, "rollout-invalid-origin-key", auth.AdminSessionCookie, "https://evil.example"), 403)
	check(call("POST", fmt.Sprintf(`{"revision":%d,"label":"CI","flags":{"unknown":{"enabled":true,"percentage":100}}}`, original.Revision), "rollout-unknown-feature-key", auth.AdminSessionCookie, "https://admin.hackways.in"), 422)
	written := call("POST", body, "rollout-publish-ci-key", auth.AdminSessionCookie, "https://admin.hackways.in")
	check(written, 200)
	replay := call("POST", body, "rollout-publish-ci-key", auth.AdminSessionCookie, "https://admin.hackways.in")
	check(replay, 200)
	if replay.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("write retry did not replay")
	}
	check(call("POST", body, "rollout-stale-version-key", auth.AdminSessionCookie, "https://admin.hackways.in"), 409)
	var response struct {
		Data struct {
			Version  int64 `json:"version"`
			Revision int64 `json:"revision"`
		} `json:"data"`
	}
	if err = json.Unmarshal(written.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if cache.Evaluate("cohort", time.Now()).Version != response.Data.Version {
		t.Fatal("committed version did not refresh snapshot")
	}
	restore := fmt.Sprintf(`{"revision":%d,"label":"CI restore","restore_version":%d}`, response.Data.Revision, original.Reference)
	check(call("POST", restore, "rollout-restore-ci-key", auth.AdminSessionCookie, "https://admin.hackways.in"), 200)
	current, err := q.ReadRuntimeRollout(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(current.Flags) != string(original.Flags) || current.Reference == original.Reference {
		t.Fatal("restore must copy the prior configuration into a new immutable version")
	}
	_, err = pool.Exec(ctx, "UPDATE platform_administrators SET revoked_at=clock_timestamp() WHERE user_id=$1", actor.ID)
	if err != nil {
		t.Fatal(err)
	}
	check(call("POST", restore, "rollout-revoked-ci-key", auth.AdminSessionCookie, "https://admin.hackways.in"), 403)
}
