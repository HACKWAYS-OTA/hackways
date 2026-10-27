// Package sitecontent serves published platform notices and reviewed SEO metadata.
// Editorial notes, keyword plans and draft content never enter public responses.
package sitecontent

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var Slug = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,119}$`)
var publicPath = regexp.MustCompile(`^/(?:discover|spaces|cities|programs|ambassadors|notices|about|contact|acceptable-use|cookies|terms|privacy|help|organizers|mcp(?:/setup)?|(?:events|spaces|cities|programs|notice|help)/[a-zA-Z0-9-]{1,180})?$`)

func PublicPath(path string) bool { return publicPath.MatchString(path) && path != "/spaces/new" }

type Handler struct{ queries *db.Queries }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{db.New(pool)} }
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Timeout(3 * time.Second))
	r.Get("/notices", h.Notices)
	r.Get("/notices/{slug}", h.Notice)
	r.Get("/seo", h.Seo)
	r.Get("/seo-rules", h.Rules)
	return r
}
func page(w http.ResponseWriter, r *http.Request) (int32, int32, bool) {
	limit, offset := int64(50), int64(0)
	for name, target := range map[string]*int64{"limit": &limit, "offset": &offset} {
		if raw := r.URL.Query().Get(name); raw != "" {
			value, e := strconv.ParseInt(raw, 10, 32)
			if e != nil {
				platform.RespondError(w, 400, "INVALID_PAGE", "Choose a valid page", nil)
				return 0, 0, false
			}
			*target = value
		}
	}
	if limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		platform.RespondError(w, 400, "INVALID_PAGE", "Choose a valid page", nil)
		return 0, 0, false
	}
	return int32(limit), int32(offset), true
}
func (h *Handler) Notices(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := page(w, r)
	if !ok {
		return
	}
	rows, e := h.queries.PublicNotices(r.Context(), db.PublicNoticesParams{PageLimit: limit + 1, PageOffset: offset})
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	more := len(rows) > int(limit)
	if more {
		rows = rows[:limit]
	}
	if rows == nil {
		rows = []db.PublicNoticesRow{}
	}
	w.Header().Set("Cache-Control", "no-store")
	platform.RespondJSON(w, 200, struct {
		Data    []db.PublicNoticesRow `json:"data"`
		HasNext bool                  `json:"has_next"`
	}{rows, more})
}
func (h *Handler) Notice(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !Slug.MatchString(slug) {
		platform.RespondError(w, 404, "NOT_FOUND", "Notice not found", nil)
		return
	}
	row, e := h.queries.PublicNotice(r.Context(), slug)
	if errors.Is(e, pgx.ErrNoRows) {
		platform.RespondError(w, 404, "NOT_FOUND", "Notice not found", nil)
		return
	}
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	images, e := h.queries.PublicNoticeImages(r.Context(), slug)
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	if images == nil {
		images = []db.PublicNoticeImagesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data   db.PublicNoticeRow         `json:"data"`
		Images []db.PublicNoticeImagesRow `json:"images"`
	}{row, images})
}
func (h *Handler) Seo(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if !PublicPath(path) {
		platform.RespondError(w, 400, "INVALID_PATH", "Choose a public page path", nil)
		return
	}
	row, e := h.queries.PublicSeoPage(r.Context(), path)
	if errors.Is(e, pgx.ErrNoRows) {
		platform.RespondJSON(w, 200, struct {
			Data *db.PublicSeoPageRow `json:"data"`
		}{nil})
		return
	}
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=30")
	platform.RespondJSON(w, 200, struct {
		Data db.PublicSeoPageRow `json:"data"`
	}{row})
}
func (h *Handler) Rules(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := page(w, r)
	if !ok {
		return
	}
	rows, e := h.queries.PublicSeoRules(r.Context(), db.PublicSeoRulesParams{PageLimit: limit + 1, PageOffset: offset})
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	more := len(rows) > int(limit)
	if more {
		rows = rows[:limit]
	}
	if rows == nil {
		rows = []db.PublicSeoRulesRow{}
	}
	w.Header().Set("Cache-Control", "public, max-age=30")
	platform.RespondJSON(w, 200, struct {
		Data    []db.PublicSeoRulesRow `json:"data"`
		HasNext bool                   `json:"has_next"`
	}{rows, more})
}
