package admin

import (
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	sessions *auth.SessionManager
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	q := db.New(pool)
	return &Handler{pool: pool, queries: q, sessions: auth.NewSessionManager(q)}
}
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(auth.RequirePlatformAdmin(h.sessions, h.queries))
	r.Get("/overview", h.Overview)
	r.Get("/events", h.AdminListEvents)
	r.Get("/spaces", h.AdminListSpaces)
	r.Get("/registrations", h.AdminListRegistrations)
	r.Get("/tickets", h.AdminListTickets)
	r.Get("/moderation", h.AdminListModeration)
	r.Get("/audit", h.AdminListAudit)
	r.Get("/discover-promotions", h.GetDiscoverPromotions)
	r.Get("/discover-promotions/events", h.DiscoverPromotionChoices)
	r.With(auth.RequireAdminMutation).Put("/discover-promotions", h.SaveDiscoverPromotions)
	r.Get("/{resource}/{id}", h.Detail)
	h.mountActions(r)
	return r
}

func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	result, err := h.queries.AdminCoreOverview(r.Context())
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.AdminCoreOverviewRow `json:"data"`
	}{result})
}
func pageParams(w http.ResponseWriter, r *http.Request) (string, string, int32, int32, bool) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	state := strings.TrimSpace(r.URL.Query().Get("status"))
	if utf8.RuneCountInString(q) > 200 || len(state) > 60 {
		platform.RespondError(w, 400, "INVALID_SEARCH", "Search exceeds its limit", nil)
		return "", "", 0, 0, false
	}
	limit, offset := int64(20), int64(0)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 1 || n > 100 {
			platform.RespondError(w, 400, "INVALID_PAGE", "Invalid page size", nil)
			return "", "", 0, 0, false
		}
		limit = n
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 0 {
			platform.RespondError(w, 400, "INVALID_PAGE", "Invalid page offset", nil)
			return "", "", 0, 0, false
		}
		offset = n
	}
	return q, state, int32(limit), int32(offset), true
}
func (h *Handler) AdminListEvents(w http.ResponseWriter, r *http.Request) {
	q, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListEvents(r.Context(), db.AdminListEventsParams{Q: q, State: state, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListEventsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListEventsRow `json:"data"`
	}{rows})
}
func (h *Handler) AdminListSpaces(w http.ResponseWriter, r *http.Request) {
	q, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListSpaces(r.Context(), db.AdminListSpacesParams{Q: q, State: state, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListSpacesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListSpacesRow `json:"data"`
	}{rows})
}
func (h *Handler) AdminListRegistrations(w http.ResponseWriter, r *http.Request) {
	q, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListRegistrations(r.Context(), db.AdminListRegistrationsParams{Q: q, State: state, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListRegistrationsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListRegistrationsRow `json:"data"`
	}{rows})
}
func (h *Handler) AdminListTickets(w http.ResponseWriter, r *http.Request) {
	q, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListTickets(r.Context(), db.AdminListTicketsParams{Q: q, State: state, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListTicketsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListTicketsRow `json:"data"`
	}{rows})
}
func (h *Handler) AdminListModeration(w http.ResponseWriter, r *http.Request) {
	q, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListModeration(r.Context(), db.AdminListModerationParams{Q: q, State: state, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListModerationRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListModerationRow `json:"data"`
	}{rows})
}
func (h *Handler) AdminListAudit(w http.ResponseWriter, r *http.Request) {
	q, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListAudit(r.Context(), db.AdminListAuditParams{Q: q, State: state, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListAuditRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListAuditRow `json:"data"`
	}{rows})
}
