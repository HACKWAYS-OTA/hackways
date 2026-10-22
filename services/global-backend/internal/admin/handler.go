package admin

import (
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
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
	r.Get("/people", h.AdminListPeople)
	r.Get("/messaging", h.AdminListMessages)
	r.Get("/storage", h.AdminListAssets)
	r.Get("/{resource}/{id}", h.Detail)
	h.mountActions(r)
	return r
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
func (h *Handler) AdminListPeople(w http.ResponseWriter, r *http.Request) {
	q, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListPeople(r.Context(), db.AdminListPeopleParams{Q: q, State: state, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListPeopleRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListPeopleRow `json:"data"`
	}{rows})
}
func (h *Handler) AdminListMessages(w http.ResponseWriter, r *http.Request) {
	q, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListMessages(r.Context(), db.AdminListMessagesParams{Q: q, State: state, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListMessagesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListMessagesRow `json:"data"`
	}{rows})
}
func (h *Handler) AdminListAssets(w http.ResponseWriter, r *http.Request) {
	q, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListAssets(r.Context(), db.AdminListAssetsParams{Q: q, State: state, PageLimit: limit, PageOffset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListAssetsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListAssetsRow `json:"data"`
	}{rows})
}
