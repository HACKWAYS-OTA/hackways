package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/hackways/global-backend/internal/sitecontent"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

type noticeBody struct {
	Reference int64  `json:"reference"`
	Revision  int64  `json:"revision"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Body      string `json:"body"`
	Status    string `json:"status"`
}
type seoBody struct {
	Revision    int64    `json:"revision"`
	Path        string   `json:"path"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Canonical   string   `json:"canonical_path"`
	Noindex     bool     `json:"noindex"`
	Keywords    []string `json:"planning_keywords"`
	Notes       string   `json:"notes"`
}
type utilitySaved struct {
	ID        pgtype.UUID `json:"id"`
	Reference int64       `json:"reference"`
	Revision  int64       `json:"revision"`
	Slug      string      `json:"slug,omitempty"`
	Path      string      `json:"path,omitempty"`
	Status    string      `json:"status,omitempty"`
}
type utilityConflict struct{}

func (utilityConflict) Error() string {
	return "The saved version changed. Reload it before saving again."
}

func (h *Handler) mountUtilities(r chi.Router) {
	r.Get("/utilities/notices", h.NoticeList)
	r.Get("/utilities/notices/{id}", h.NoticeDetail)
	r.Get("/utilities/notices/by-slug/{slug}", h.NoticeBySlug)
	r.Post("/utilities/notices", h.SaveNotice)
	r.Get("/utilities/seo", h.SeoList)
	r.Get("/utilities/seo/page", h.SeoDetail)
	r.Get("/utilities/seo/suggestions", h.SeoSuggestions)
	r.Post("/utilities/seo", h.SaveSeo)
}
func (h *Handler) NoticeBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !sitecontent.Slug.MatchString(slug) {
		platform.RespondError(w, 404, "NOT_FOUND", "Notice not found", nil)
		return
	}
	row, e := h.queries.AdminNoticeBySlug(r.Context(), slug)
	if e != nil {
		detailError(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.SiteNotices `json:"data"`
	}{row})
}
func (h *Handler) SeoSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	rows, e := h.queries.AdminSeoSuggestions(ctx)
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	public := make([]db.AdminSeoSuggestionsRow, 0, 10)
	for _, row := range rows {
		if sitecontent.PublicPath(row.Route) {
			public = append(public, row)
			if len(public) == 10 {
				break
			}
		}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminSeoSuggestionsRow `json:"data"`
	}{public})
}
func (h *Handler) NoticeList(w http.ResponseWriter, r *http.Request) {
	search, state, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, e := h.queries.AdminNotices(r.Context(), db.AdminNoticesParams{Q: search, State: state, PageLimit: limit + 1, PageOffset: offset})
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	more := len(rows) > int(limit)
	if more {
		rows = rows[:limit]
	}
	if rows == nil {
		rows = []db.AdminNoticesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data    []db.AdminNoticesRow `json:"data"`
		HasNext bool                 `json:"has_next"`
	}{rows, more})
}
func (h *Handler) NoticeDetail(w http.ResponseWriter, r *http.Request) {
	reference, ok := campaignReference(w, r)
	if !ok {
		return
	}
	row, e := h.queries.AdminNotice(r.Context(), reference)
	if e != nil {
		detailError(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.SiteNotices `json:"data"`
	}{row})
}
func (h *Handler) SeoList(w http.ResponseWriter, r *http.Request) {
	search, _, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, e := h.queries.AdminSeoPages(r.Context(), db.AdminSeoPagesParams{Q: search, PageLimit: limit + 1, PageOffset: offset})
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	more := len(rows) > int(limit)
	if more {
		rows = rows[:limit]
	}
	if rows == nil {
		rows = []db.AdminSeoPagesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data    []db.AdminSeoPagesRow `json:"data"`
		HasNext bool                  `json:"has_next"`
	}{rows, more})
}
func (h *Handler) SeoDetail(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if !sitecontent.PublicPath(path) {
		platform.RespondError(w, 422, "INVALID_PATH", "Choose a public page path without queries or private identifiers", nil)
		return
	}
	row, e := h.queries.AdminSeoPage(r.Context(), path)
	if errors.Is(e, pgx.ErrNoRows) {
		platform.RespondJSON(w, 200, struct {
			Data *db.SiteSeoPages `json:"data"`
		}{nil})
		return
	}
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.SiteSeoPages `json:"data"`
	}{row})
}
func textWithin(value string, min, max int) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) && utf8.RuneCountInString(value) >= min && utf8.RuneCountInString(value) <= max
}

// Reuse the established admin origin, current-session/grant, timeout, audit and
// idempotency contracts. The resource lock serializes edits by different admins.
func (h *Handler) utilityWrite(w http.ResponseWriter, r *http.Request, kind, resource string, body []byte, apply func(*db.Queries, pgtype.UUID) (utilitySaved, pgtype.UUID, error)) {
	if r.Header.Get("X-Admin-Reason") == "" {
		platform.RespondError(w, 400, "REASON_REQUIRED", "Add a short change note before saving", nil)
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) < 16 || len(key) > 128 || strings.IndexFunc(key, func(c rune) bool { return c < 33 || c > 126 }) >= 0 {
		platform.RespondError(w, 400, "INVALID_REQUEST_KEY", "Provide a printable request key", nil)
		return
	}
	tx, q, actor, reason, ok := h.campaignMutation(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	scope := actionHash([]byte(fmt.Sprintf("admin-utility:%x:%s:%s", actor.UserID.Bytes, kind, resource)))
	keyHash := actionHash([]byte(key))
	hash := actionHash(append(body, []byte(reason)...))
	if e := q.LockSiteUtility(r.Context(), "site-utility:"+kind+":"+resource); e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	replay, e := q.AdminActionReplay(r.Context(), db.AdminActionReplayParams{ScopeHash: scope, KeyHash: keyHash})
	if e == nil {
		if replay.RequestHash != hash {
			platform.RespondError(w, 409, "IDEMPOTENCY_CONFLICT", "This request key was used for another change", nil)
			return
		}
		w.Header().Set("Idempotency-Replayed", "true")
		platform.RespondJSON(w, 200, json.RawMessage(replay.ResponseJson))
		return
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		platform.RespondDatabaseError(w, e)
		return
	}
	saved, id, e := apply(q, actor.UserID)
	saved.ID = id
	if e != nil {
		var conflict utilityConflict
		var pgerr *pgconn.PgError
		if errors.As(e, &conflict) || errors.Is(e, pgx.ErrNoRows) {
			platform.RespondError(w, 409, "REVISION_CONFLICT", "The saved version changed or the path already exists. Reload before saving again.", nil)
		} else if errors.As(e, &pgerr) && pgerr.Code == "23505" {
			platform.RespondError(w, 409, "PATH_EXISTS", "That page address already exists", nil)
		} else {
			platform.RespondDatabaseError(w, e)
		}
		return
	}
	metadata, _ := json.Marshal(struct {
		Reason   string `json:"reason"`
		Revision int64  `json:"revision"`
	}{reason, saved.Revision})
	if e = q.AppendAdminAudit(r.Context(), db.AppendAdminAuditParams{ActorUserID: actor.UserID, EntityID: id, Action: "admin.utility." + kind + ".saved", RequestID: platform.GetRequestID(r.Context()), Metadata: metadata}); e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	response, _ := json.Marshal(struct {
		Data utilitySaved `json:"data"`
	}{saved})
	if e = q.AdminActionRemember(r.Context(), db.AdminActionRememberParams{ScopeHash: scope, KeyHash: keyHash, RequestHash: hash, ResponseJson: response}); e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	platform.RespondJSON(w, 200, json.RawMessage(response))
}
func (h *Handler) SaveNotice(w http.ResponseWriter, r *http.Request) {
	var body noticeBody
	if platform.DecodeJSON(r, &body) != nil {
		platform.RespondError(w, 422, "INVALID_NOTICE", "Check notice fields", nil)
		return
	}
	body.Slug = strings.TrimSpace(body.Slug)
	body.Title = strings.TrimSpace(body.Title)
	body.Summary = strings.TrimSpace(body.Summary)
	body.Body = strings.TrimSpace(body.Body)
	if body.Reference < 0 || body.Revision < 0 || body.Reference == 0 && body.Revision != 0 || body.Reference > 0 && body.Revision < 1 || !sitecontent.Slug.MatchString(body.Slug) || !textWithin(body.Title, 1, 180) || !textWithin(body.Summary, 0, 500) || !textWithin(body.Body, 0, 100000) || (body.Status != "draft" && body.Status != "published" && body.Status != "archived") || body.Status == "published" && (body.Body == "" || body.Summary == "") {
		platform.RespondError(w, 422, "INVALID_NOTICE", "Provide a title and valid address. Published notices also need a summary and content.", nil)
		return
	}
	canonical, _ := json.Marshal(body)
	h.utilityWrite(w, r, "notice", body.Slug, canonical, func(q *db.Queries, actor pgtype.UUID) (utilitySaved, pgtype.UUID, error) {
		var row db.SiteNotices
		var e error
		if body.Reference == 0 {
			row, e = q.CreateNotice(r.Context(), db.CreateNoticeParams{Slug: body.Slug, Title: body.Title, Summary: body.Summary, Body: body.Body, Status: body.Status, CreatedBy: actor})
		} else {
			previous, err := q.AdminNotice(r.Context(), body.Reference)
			if err != nil {
				return utilitySaved{}, pgtype.UUID{}, err
			}
			if previous.Slug != body.Slug {
				return utilitySaved{}, pgtype.UUID{}, utilityConflict{}
			}
			row, e = q.UpdateNotice(r.Context(), db.UpdateNoticeParams{Reference: body.Reference, Revision: body.Revision, Title: body.Title, Summary: body.Summary, Body: body.Body, Status: body.Status, Actor: actor})
		}
		return utilitySaved{Reference: row.Reference, Revision: row.Revision, Slug: row.Slug, Status: row.Status}, row.ID, e
	})
}
func (h *Handler) SaveSeo(w http.ResponseWriter, r *http.Request) {
	var body seoBody
	if platform.DecodeJSON(r, &body) != nil {
		platform.RespondError(w, 422, "INVALID_SEO", "Check page settings", nil)
		return
	}
	body.Path = strings.TrimSpace(body.Path)
	body.Title = strings.TrimSpace(body.Title)
	body.Description = strings.TrimSpace(body.Description)
	body.Canonical = strings.TrimSpace(body.Canonical)
	body.Notes = strings.TrimSpace(body.Notes)
	if body.Revision < 0 || !sitecontent.PublicPath(body.Path) || body.Canonical != "" && !sitecontent.PublicPath(body.Canonical) || !textWithin(body.Title, 0, 90) || !textWithin(body.Description, 0, 320) || !textWithin(body.Notes, 0, 4000) || len(body.Keywords) > 2000 {
		platform.RespondError(w, 422, "INVALID_SEO", "Use a public path, title up to 90 characters and description up to 320 characters. Private pages cannot be indexed.", nil)
		return
	}
	keywords := make([]string, 0, len(body.Keywords))
	seen := map[string]bool{}
	for _, keyword := range body.Keywords {
		keyword = strings.TrimSpace(keyword)
		if !textWithin(keyword, 1, 64) || strings.ContainsAny(keyword, "\r\n") {
			platform.RespondError(w, 422, "INVALID_KEYWORDS", "Use up to 2,000 planning keywords of 1 to 64 characters, one per line", nil)
			return
		}
		key := strings.ToLower(keyword)
		if !seen[key] {
			seen[key] = true
			keywords = append(keywords, keyword)
		}
	}
	body.Keywords = keywords
	canonical, _ := json.Marshal(body)
	h.utilityWrite(w, r, "seo", body.Path, canonical, func(q *db.Queries, actor pgtype.UUID) (utilitySaved, pgtype.UUID, error) {
		var row db.SiteSeoPages
		var e error
		if body.Revision == 0 {
			row, e = q.CreateSeoPage(r.Context(), db.CreateSeoPageParams{Path: body.Path, Title: body.Title, Description: body.Description, CanonicalPath: body.Canonical, Noindex: body.Noindex, PlanningKeywords: body.Keywords, Notes: body.Notes, UpdatedBy: actor})
		} else {
			row, e = q.UpdateSeoPage(r.Context(), db.UpdateSeoPageParams{Path: body.Path, Revision: body.Revision, Title: body.Title, Description: body.Description, CanonicalPath: body.Canonical, Noindex: body.Noindex, PlanningKeywords: body.Keywords, Notes: body.Notes, Actor: actor})
		}
		return utilitySaved{Reference: row.Reference, Revision: row.Revision, Path: row.Path}, row.ID, e
	})
}
