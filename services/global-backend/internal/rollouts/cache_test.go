package rollouts

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
)

// Pure evaluation uses no database, including under concurrent HTTP reads.
func TestStableCohortsSchedulesAndFallback(t *testing.T) {
	now := time.Now()
	salt := []byte("isolated-cohort-test-salt-32-bytes")
	c := New(nil)
	s := &snapshot{revision: 1, version: 1, label: "CI", loaded: now, salt: salt, rules: Rules{"manage.overview.v2": {Enabled: true, Percentage: 10}}}
	c.current.Store(s)
	for i := 0; i < 1000; i++ {
		cohort := fmt.Sprint(i)
		a := c.Evaluate(cohort, now).Flags["manage.overview.v2"]
		b := c.Evaluate(cohort, now).Flags["manage.overview.v2"]
		if a != b {
			t.Fatal("cohort changed without configuration change")
		}
		rule := s.rules["manage.overview.v2"]
		rule.Percentage = 50
		if a && !enabled(rule, salt, "manage.overview.v2", cohort, now) {
			t.Fatal("increasing rollout removed an eligible user")
		}
	}
	start, end := now.Add(time.Hour), now.Add(2*time.Hour)
	rule := Rule{Enabled: true, Percentage: 100, StartsAt: &start, EndsAt: &end}
	if enabled(rule, salt, "key", "cohort", now) || !enabled(rule, salt, "key", "cohort", start) || enabled(rule, salt, "key", "cohort", end) {
		t.Fatal("schedule boundaries must be start-inclusive and end-exclusive")
	}
	rule.Enabled = false
	if enabled(rule, salt, "key", "cohort", start) {
		t.Fatal("disabled flag was enabled")
	}
	if c.Evaluate("cohort", now.Add(3*time.Minute)).Flags["manage.overview.v2"] {
		t.Fatal("expired snapshot did not fall back")
	}
	if Validate(Rules{"unknown.flag": {Enabled: true, Percentage: 100}}) == nil || Validate(Rules{"manage.overview.v2": {Percentage: 101}}) == nil {
		t.Fatal("invalid configuration accepted")
	}
	if New(nil).Evaluate("cohort", now).Fresh {
		t.Fatal("missing snapshot reported fresh")
	}
}
func TestRuntimeHTTPHasNoDatabaseDependency(t *testing.T) {
	now := time.Now()
	c := New(nil)
	c.current.Store(&snapshot{revision: 1, version: 1, label: "CI", loaded: now, salt: []byte("ci-salt"), rules: Rules{"discover.promotions": {Enabled: true, Percentage: 100}}})
	r := httptest.NewRequest("GET", "/auth/runtime", nil)
	w := httptest.NewRecorder()
	c.ServeHTTP(w, r)
	if w.Code != 200 || len(w.Result().Cookies()) != 1 || w.Header().Get("ETag") == "" {
		t.Fatal("runtime did not issue a cohort and cached evaluation")
	}
	cookie := w.Result().Cookies()[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.Name != CookieName {
		t.Fatal("cohort cookie lost transport protections")
	}
	r = httptest.NewRequest("GET", "/auth/runtime", nil)
	r.AddCookie(cookie)
	r.Header.Set("If-None-Match", w.Header().Get("ETag"))
	next := httptest.NewRecorder()
	c.ServeHTTP(next, r)
	if next.Code != 304 {
		t.Fatalf("unchanged snapshot got %d", next.Code)
	}
	t.Run("concurrent reads", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				t.Parallel()
				rec := httptest.NewRecorder()
				c.ServeHTTP(rec, r)
				if rec.Code != 304 {
					t.Error("concurrent evaluation changed")
				}
			})
		}
	})
}
