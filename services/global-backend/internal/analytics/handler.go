// Package analytics collects consented, first-party product events. It never
// reads account/session identity, IP addresses, form answers or full referrers.
package analytics

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	queries     *db.Queries
	key         []byte
	development bool
}

func NewHandler(pool *pgxpool.Pool, secret string, development bool) *Handler {
	return &Handler{queries: db.New(pool), key: []byte(secret), development: development}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/events", h.Record)
	return r
}

type event struct {
	ID           string `json:"id"`
	Name         string `json:"event_name"`
	Route        string `json:"route"`
	Device       string `json:"device"`
	ReferrerHost string `json:"referrer_host"`
	Step         *int16 `json:"step,omitempty"`
}

// Only published entity slugs and canonical private route templates are accepted.
// Never accept arbitrary paths, account IDs, registration codes or query strings.
var pageRoute = regexp.MustCompile(`^/(?:discover|events|spaces(?:/new)?|cities|programs|create|my-events|account|login|not-allowed|about|contact|acceptable-use|cookies|terms|privacy|help|organizers|ambassadors|mcp(?:/setup)?|connect/(?:action|authorize|connections)|(?:manage/)?auth/(?:forgot-password|reset-password|verify-email|google/callback|unsubscribe)|events/[a-zA-Z0-9-]{1,180}(?:/register)?|(?:spaces|cities|programs|help)/[a-zA-Z0-9-]{1,180}|my-events/:registration|spaces/:space/settings|programs/:program/(?:application|apply/:step)|manage(?:/(?:spaces|onboarding|create|login|spaces/:space(?:/team)?|events/:event(?:/(?:team|settings|registration(?:/questions)?|check-in|invitations|appearance|edit|guests|activity))?))?)?$`)
var domain = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]{0,251}[a-z0-9])?$`)

func validID(raw string) bool {
	var id pgtype.UUID
	return len(raw) == 36 && id.Scan(raw) == nil && id.Valid && id.Bytes != [16]byte{}
}

func validEvent(value *event) bool {
	if !validID(value.ID) || len(value.Route) > 240 || !pageRoute.MatchString(value.Route) {
		return false
	}
	switch value.Name {
	case "page_view":
		if value.Step != nil {
			return false
		}
	case "registration_step", "registration_complete":
		if !strings.HasPrefix(value.Route, "/events/") || !strings.HasSuffix(value.Route, "/register") {
			return false
		}
		if value.Name == "registration_step" && (value.Step == nil || *value.Step < 0 || *value.Step > 50) {
			return false
		}
		if value.Name == "registration_complete" && value.Step != nil {
			return false
		}
	default:
		return false
	}
	if value.Device != "desktop" && value.Device != "tablet" && value.Device != "mobile" {
		return false
	}
	value.ReferrerHost = strings.ToLower(strings.TrimSpace(value.ReferrerHost))
	if len(value.ReferrerHost) > 253 || value.ReferrerHost != "" && (!domain.MatchString(value.ReferrerHost) || !strings.Contains(value.ReferrerHost, ".")) {
		return false
	}
	return true
}

func (h *Handler) hash(purpose, value string) string {
	mac := hmac.New(sha256.New, h.key)
	mac.Write([]byte("hackways-analytics-v1\x00" + purpose + "\x00" + value))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *Handler) Record(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	origin, err := url.Parse(r.Header.Get("Origin"))
	allowed := err == nil && origin.String() == "https://hackways.in"
	if h.development && err == nil && (origin.Hostname() == "localhost" || origin.Hostname() == "127.0.0.1") && origin.Scheme == "http" && origin.Path == "" && origin.RawQuery == "" && origin.Fragment == "" && origin.User == nil {
		allowed = true
	}
	if !allowed {
		platform.RespondError(w, 403, "INVALID_ORIGIN", "Use the Hackways website", nil)
		return
	}
	if len(h.key) < 32 {
		platform.RespondError(w, 503, "ANALYTICS_UNAVAILABLE", "Analytics collection is unavailable", nil)
		return
	}
	var body struct {
		Consent bool    `json:"consent"`
		Visitor string  `json:"visitor_id"`
		Session string  `json:"session_id"`
		Events  []event `json:"events"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	if platform.DecodeJSON(r, &body) != nil || !body.Consent || !validID(body.Visitor) || !validID(body.Session) || len(body.Events) < 1 || len(body.Events) > 10 {
		platform.RespondError(w, 400, "INVALID_ANALYTICS", "Analytics requires consent and a small valid event batch", nil)
		return
	}
	seen := make(map[string]bool, len(body.Events))
	for i := range body.Events {
		if !validEvent(&body.Events[i]) || seen[body.Events[i].ID] {
			platform.RespondError(w, 400, "INVALID_ANALYTICS", "Check analytics event data", nil)
			return
		}
		seen[body.Events[i].ID] = true
	}
	encoded, err := json.Marshal(body.Events)
	if err != nil {
		platform.RespondError(w, 400, "INVALID_ANALYTICS", "Analytics could not be prepared", nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	result, err := h.queries.RecordAnalyticsBatch(ctx, db.RecordAnalyticsBatchParams{BatchSize: int32(len(body.Events)), VisitorHash: h.hash("visitor", body.Visitor), SessionHash: h.hash("session", body.Session), Events: encoded})
	if err != nil {
		slog.Error("analytics collection failed", "request_id", platform.GetRequestID(r.Context()))
		platform.RespondError(w, 503, "ANALYTICS_UNAVAILABLE", "Analytics could not be recorded", nil)
		return
	}
	if !result.WithinBudget {
		platform.RespondError(w, 429, "ANALYTICS_LIMIT", "Analytics collection has reached today's limit", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
