// Package rollouts evaluates presentation features from an immutable snapshot.
// PostgreSQL owns configuration history; requests never query it for flags.
// Flags cannot grant permissions, bypass validation, or alter capacity rules.
package rollouts

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5/pgtype"
)

const CookieName = "__Host-hackways_rollout"

type Definition struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Surface string `json:"surface"`
}

func Definitions() []Definition {
	return []Definition{{"discover.promotions", "Discover promotional banner", "Discover"}, {"manage.overview.v2", "Expanded event overview", "Event management"}}
}

type Rule struct {
	Enabled    bool       `json:"enabled"`
	Percentage float64    `json:"percentage"`
	StartsAt   *time.Time `json:"starts_at,omitempty"`
	EndsAt     *time.Time `json:"ends_at,omitempty"`
}
type Rules map[string]Rule

func Validate(rules Rules) error {
	if len(rules) > len(Definitions()) {
		return errors.New("Only implemented features can be configured")
	}
	for key, rule := range rules {
		known := false
		for _, definition := range Definitions() {
			if key == definition.Key {
				known = true
				break
			}
		}
		if !known {
			return fmt.Errorf("Unknown feature: %s", key)
		}
		if rule.Percentage < 0 || rule.Percentage > 100 {
			return errors.New("Rollout percentage must be between 0 and 100")
		}
		if rule.StartsAt != nil && rule.EndsAt != nil && !rule.EndsAt.After(*rule.StartsAt) {
			return errors.New("Feature end must follow its start")
		}
	}
	return nil
}

type snapshot struct {
	revision int64
	version  int64
	label    string
	rules    Rules
	salt     []byte
	loaded   time.Time
}
type Cache struct {
	q       *db.Queries
	mu      sync.Mutex
	current atomic.Pointer[snapshot]
}

func New(q *db.Queries) *Cache { return &Cache{q: q} }
func (c *Cache) Refresh(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	row, err := c.q.ReadRuntimeRollout(ctx)
	if err != nil {
		return err
	}
	var rules Rules
	if err = json.Unmarshal(row.Flags, &rules); err != nil {
		return err
	}
	if err = Validate(rules); err != nil {
		return err
	}
	c.current.Store(&snapshot{row.Revision, row.Reference, row.Label, rules, append([]byte(nil), row.CohortSalt...), time.Now()})
	return nil
}
func (c *Cache) Run(ctx context.Context, logger *slog.Logger) {
	timer := time.NewTicker(10 * time.Second)
	defer timer.Stop()
	var lastLog time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if err := c.Refresh(ctx); err != nil && time.Since(lastLog) > time.Minute {
				logger.Warn("feature snapshot refresh failed; bounded last-known configuration applies")
				lastLog = time.Now()
			}
		}
	}
}
func bucket(salt []byte, key, cohort string) float64 {
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(key + "\x00" + cohort))
	sum := mac.Sum(nil)
	value := uint32(sum[0])<<24 | uint32(sum[1])<<16 | uint32(sum[2])<<8 | uint32(sum[3])
	return float64(value) / 4294967296.0 * 100
}
func enabled(rule Rule, salt []byte, key, cohort string, now time.Time) bool {
	return rule.Enabled && rule.Percentage > 0 && (rule.StartsAt == nil || !now.Before(*rule.StartsAt)) && (rule.EndsAt == nil || now.Before(*rule.EndsAt)) && bucket(salt, key, cohort) < rule.Percentage
}
func setCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: value, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 365 * 24 * 60 * 60})
}
func (c *Cache) AccountCookie(w http.ResponseWriter, id pgtype.UUID) {
	s := c.current.Load()
	if s == nil {
		return
	}
	mac := hmac.New(sha256.New, s.salt)
	mac.Write([]byte("account:"))
	mac.Write(id.Bytes[:])
	setCookie(w, hex.EncodeToString(mac.Sum(nil)))
}

type Evaluation struct {
	Version  int64           `json:"version"`
	Revision int64           `json:"revision"`
	Label    string          `json:"label"`
	Flags    map[string]bool `json:"flags"`
	Fresh    bool            `json:"fresh"`
}

type Status struct {
	Version    int64 `json:"version"`
	Revision   int64 `json:"revision"`
	AgeSeconds int64 `json:"age_seconds"`
	Fresh      bool  `json:"fresh"`
}

func (c *Cache) Status() Status {
	s := c.current.Load()
	if s == nil {
		return Status{}
	}
	age := time.Since(s.loaded)
	return Status{s.version, s.revision, int64(age.Seconds()), age <= 2*time.Minute}
}

func (c *Cache) Evaluate(cohort string, now time.Time) Evaluation {
	result := Evaluation{Flags: make(map[string]bool)}
	s := c.current.Load()
	fresh := s != nil && now.Sub(s.loaded) <= 2*time.Minute
	result.Fresh = fresh
	if s != nil {
		result.Version = s.version
		result.Revision = s.revision
		result.Label = s.label
	}
	for _, definition := range Definitions() {
		result.Flags[definition.Key] = fresh && enabled(s.rules[definition.Key], s.salt, definition.Key, cohort, now)
	}
	return result
}
func (c *Cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cohort := ""
	if cookie, err := r.Cookie(CookieName); err == nil {
		decoded, e := hex.DecodeString(cookie.Value)
		if e == nil && len(decoded) == 32 {
			cohort = cookie.Value
		}
	}
	if cohort == "" {
		random := make([]byte, 32)
		if _, err := rand.Read(random); err != nil {
			platform.RespondError(w, 503, "RUNTIME_UNAVAILABLE", "Feature configuration is temporarily unavailable", nil)
			return
		}
		cohort = hex.EncodeToString(random)
		setCookie(w, cohort)
	}
	result := c.Evaluate(cohort, time.Now())
	data, _ := json.Marshal(result)
	sum := sha256.Sum256(data)
	etag := `"` + hex.EncodeToString(sum[:]) + `"`
	w.Header().Set("Cache-Control", "private, no-cache")
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(304)
		return
	}
	platform.RespondJSON(w, 200, result)
}
