package cities

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool, db.New(pool)} }
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/by-slug/{slug}", h.Read)
	r.Get("/{id}/events", h.Events)
	return r
}
func (h *Handler) AdminRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(auth.RequirePlatformAdmin(auth.NewSessionManager(h.q), h.q))
	r.Get("/", h.AdminList)
	r.Get("/event-choices", h.Choices)
	r.Get("/{reference}", h.AdminRead)
	r.Group(func(m chi.Router) {
		m.Use(auth.RequireAdminMutation)
		m.Post("/", h.Create)
		m.Put("/{reference}", h.Update)
		m.Post("/{reference}/events", h.EventChange)
	})
	return r
}
func send[T any](w http.ResponseWriter, data T) {
	platform.RespondJSON(w, 200, struct {
		Data T `json:"data"`
	}{data})
}
func fail(w http.ResponseWriter, e error) {
	var p *pgconn.PgError
	var f *workflows.Failure
	if errors.As(e, &f) {
		platform.RespondError(w, f.Status, f.Code, f.Message, nil)
	} else if errors.Is(e, pgx.ErrNoRows) {
		platform.RespondError(w, 404, "NOT_FOUND", "City Hub not found", nil)
	} else if errors.As(e, &p) && p.Code == "23505" {
		platform.RespondError(w, 409, "SLUG_TAKEN", "This city address is already in use", nil)
	} else {
		platform.RespondDatabaseError(w, e)
	}
}
func pagination(r *http.Request) (int32, int32) {
	limit, offset := int64(20), int64(0)
	if n, e := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32); e == nil && n > 0 && n <= 100 {
		limit = n
	}
	if n, e := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32); e == nil && n >= 0 {
		offset = n
	}
	return int32(limit), int32(offset)
}
func reference(w http.ResponseWriter, r *http.Request) (int64, bool) {
	n, e := strconv.ParseInt(chi.URLParam(r, "reference"), 10, 64)
	if e != nil || n < 1 {
		platform.RespondError(w, 400, "INVALID_REFERENCE", "Invalid City Hub reference", nil)
		return 0, false
	}
	return n, true
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	l, o := pagination(r)
	rows, e := h.q.ListPublicCities(r.Context(), db.ListPublicCitiesParams{Limit: l, Offset: o})
	if e != nil {
		fail(w, e)
		return
	}
	if rows == nil {
		rows = []db.ListPublicCitiesRow{}
	}
	send(w, rows)
}
func (h *Handler) Read(w http.ResponseWriter, r *http.Request) {
	row, e := h.q.ReadPublicCity(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		fail(w, e)
		return
	}
	send(w, row)
}
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	var id pgtype.UUID
	if id.Scan(chi.URLParam(r, "id")) != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid city", nil)
		return
	}
	l, o := pagination(r)
	rows, e := h.q.ListCityEvents(r.Context(), db.ListCityEventsParams{ID: id, Past: r.URL.Query().Get("period") == "past", PageLimit: l, PageOffset: o})
	if e != nil {
		fail(w, e)
		return
	}
	if rows == nil {
		rows = []db.Events{}
	}
	send(w, rows)
}
func search(r *http.Request) string {
	return string([]rune(strings.TrimSpace(r.URL.Query().Get("q")))[:min(200, len([]rune(strings.TrimSpace(r.URL.Query().Get("q")))))])
}
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	l, o := pagination(r)
	rows, e := h.q.AdminListCities(r.Context(), db.AdminListCitiesParams{Q: search(r), PageLimit: l, PageOffset: o})
	if e != nil {
		fail(w, e)
		return
	}
	if rows == nil {
		rows = []db.AdminListCitiesRow{}
	}
	send(w, rows)
}

type cityView struct {
	ID          pgtype.UUID `json:"id"`
	Reference   int64       `json:"reference"`
	Slug        string      `json:"slug"`
	Name        string      `json:"name"`
	Region      string      `json:"region"`
	Country     string      `json:"country"`
	Timezone    string      `json:"timezone"`
	Description string      `json:"description"`
	BannerURL   pgtype.Text `json:"banner_url"`
	Status      string      `json:"status"`
	Revision    int64       `json:"revision"`
}

func view(c db.CityHubs) cityView {
	return cityView{c.ID, c.Reference, c.Slug, c.Name, c.Region, c.Country, c.Timezone, c.Description, c.BannerUrl, c.Status, c.Revision}
}
func (h *Handler) AdminRead(w http.ResponseWriter, r *http.Request) {
	ref, ok := reference(w, r)
	if !ok {
		return
	}
	c, e := h.q.ReadCityAdmin(r.Context(), ref)
	if e != nil {
		fail(w, e)
		return
	}
	l, o := pagination(r)
	events, e := h.q.AdminCityEvents(r.Context(), db.AdminCityEventsParams{CityID: c.ID, Limit: l, Offset: o})
	if e != nil {
		fail(w, e)
		return
	}
	if events == nil {
		events = []db.AdminCityEventsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data   cityView                `json:"data"`
		Events []db.AdminCityEventsRow `json:"events"`
	}{view(c), events})
}
func (h *Handler) Choices(w http.ResponseWriter, r *http.Request) {
	rows, e := h.q.CityEventChoices(r.Context(), search(r))
	if e != nil {
		fail(w, e)
		return
	}
	if rows == nil {
		rows = []db.CityEventChoicesRow{}
	}
	send(w, rows)
}

type details struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	Country     string `json:"country"`
	Timezone    string `json:"timezone"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Revision    int64  `json:"revision"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func (d *details) valid(create bool) bool {
	d.Name = strings.TrimSpace(d.Name)
	d.Region = strings.TrimSpace(d.Region)
	d.Country = strings.TrimSpace(d.Country)
	d.Description = strings.TrimSpace(d.Description)
	_, e := time.LoadLocation(d.Timezone)
	return e == nil && len(d.Timezone) <= 100 && len([]rune(d.Name)) > 0 && len([]rune(d.Name)) <= 120 && len([]rune(d.Region)) <= 120 && len([]rune(d.Country)) > 0 && len([]rune(d.Country)) <= 120 && len([]rune(d.Description)) <= 6000 && (d.Status == "draft" || d.Status == "published" || d.Status == "archived") && (!create || (len(d.Slug) <= 100 && slugPattern.MatchString(d.Slug))) && (d.Status != "published" || d.Description != "")
}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) { h.write(w, r, true) }
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) { h.write(w, r, false) }
func (h *Handler) write(w http.ResponseWriter, r *http.Request, create bool) {
	var b details
	if platform.DecodeJSON(r, &b) != nil || !b.valid(create) || (!create && b.Revision < 1) {
		platform.RespondError(w, 422, "INVALID_CITY", "Provide a name, country, valid time zone and description for published cities", nil)
		return
	}
	tx, e := h.pool.Begin(r.Context())
	if e != nil {
		fail(w, e)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.q.WithTx(tx)
	if e = q.SetWorkflowTimeout(r.Context()); e != nil {
		fail(w, e)
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	var c db.CityHubs
	if create {
		mutation, _ := auth.AdminMutationFromContext(r.Context())
		canonical, _ := json.Marshal(b)
		digest := sha256.Sum256(canonical)
		c, e = q.CreateCity(r.Context(), db.CreateCityParams{Slug: b.Slug, Name: b.Name, Region: b.Region, Country: b.Country, Timezone: b.Timezone, Description: b.Description, Status: b.Status, CreatedBy: actor.UserID, CreationKey: mutation.Key, CreationHash: hex.EncodeToString(digest[:])})
		if e != nil {
			fail(w, e)
			return
		}
		if c.CreationHash != hex.EncodeToString(digest[:]) {
			platform.RespondError(w, 409, "IDEMPOTENCY_CONFLICT", "This request key was used with different city details", nil)
			return
		}
		if replayed, err := workflows.ReplayAdminMutation(w, r, q); err != nil {
			fail(w, err)
			return
		} else if replayed {
			return
		}
	} else {
		ref, ok := reference(w, r)
		if !ok {
			return
		}
		c, e = q.LockCity(r.Context(), ref)
		if e != nil {
			fail(w, e)
			return
		}
		if replayed, err := workflows.ReplayAdminMutation(w, r, q); err != nil {
			fail(w, err)
			return
		} else if replayed {
			return
		}
		if c.Revision != b.Revision {
			platform.RespondError(w, 409, "REVISION_CONFLICT", "This City Hub changed. Reload before saving; your entries are still here", nil)
			return
		}
		c, e = q.UpdateCity(r.Context(), db.UpdateCityParams{Reference: ref, Name: b.Name, Region: b.Region, Country: b.Country, Timezone: b.Timezone, Description: b.Description, Status: b.Status, Revision: b.Revision})
		if e != nil {
			fail(w, e)
			return
		}
	}
	action := "city.updated"
	if create {
		action = "city.created"
	}
	if e = workflows.Audit(r.Context(), q, actor.UserID, pgtype.UUID{}, c.ID, action, platform.GetRequestID(r.Context())); e != nil {
		fail(w, e)
		return
	}
	if e = workflows.RememberAdminMutation(r.Context(), q, view(c)); e != nil {
		fail(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		fail(w, e)
		return
	}
	send(w, view(c))
}
func (h *Handler) EventChange(w http.ResponseWriter, r *http.Request) {
	ref, ok := reference(w, r)
	if !ok {
		return
	}
	var b struct {
		EventID  pgtype.UUID `json:"event_id"`
		Revision int64       `json:"revision"`
		Remove   bool        `json:"remove"`
	}
	if platform.DecodeJSON(r, &b) != nil || !b.EventID.Valid || b.Revision < 1 {
		platform.RespondError(w, 422, "INVALID_EVENT", "Choose an event and the current city revision", nil)
		return
	}
	tx, e := h.pool.Begin(r.Context())
	if e != nil {
		fail(w, e)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.q.WithTx(tx)
	if e = q.SetWorkflowTimeout(r.Context()); e != nil {
		fail(w, e)
		return
	}
	c, e := q.LockCity(r.Context(), ref)
	if e != nil {
		fail(w, e)
		return
	}
	if replayed, err := workflows.ReplayAdminMutation(w, r, q); err != nil {
		fail(w, err)
		return
	} else if replayed {
		return
	}
	if c.Revision != b.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "This City Hub changed. Reload before curating events", nil)
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	if b.Remove {
		e = q.RemoveCityEvent(r.Context(), db.RemoveCityEventParams{CityID: c.ID, EventID: b.EventID})
	} else {
		eligible, err := q.CityEventEligible(r.Context(), b.EventID)
		if err != nil {
			fail(w, err)
			return
		}
		if !eligible {
			platform.RespondError(w, 422, "EVENT_UNAVAILABLE", "Choose a published public event", nil)
			return
		}
		e = q.AddCityEvent(r.Context(), db.AddCityEventParams{CityID: c.ID, ID: b.EventID, AddedBy: actor.UserID})
	}
	if e != nil {
		fail(w, e)
		return
	}
	if e = q.BumpCityRevision(r.Context(), c.ID); e != nil {
		fail(w, e)
		return
	}
	c.Revision++
	if e = workflows.Audit(r.Context(), q, actor.UserID, pgtype.UUID{}, c.ID, "city.events_updated", platform.GetRequestID(r.Context())); e != nil {
		fail(w, e)
		return
	}
	if e = workflows.RememberAdminMutation(r.Context(), q, view(c)); e != nil {
		fail(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		fail(w, e)
		return
	}
	send(w, view(c))
}
