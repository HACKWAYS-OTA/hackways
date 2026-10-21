package spaces

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool     *pgxpool.Pool
	q        *db.Queries
	sessions *auth.SessionManager
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	q := db.New(pool)
	return &Handler{pool, q, auth.NewSessionManager(q)}
}
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/by-slug/{slug}", h.Read)
	r.Get("/{id}", h.ReadByID)
	r.Get("/{id}/events", h.Events)
	r.With(auth.RequireAuth(h.sessions)).Get("/mine", h.Mine)
	r.With(auth.RequireAuth(h.sessions)).Get("/{id}/membership", h.Membership)
	r.With(auth.RequireAuth(h.sessions)).Post("/{id}/join", h.Join)
	r.With(auth.RequireAuth(h.sessions)).Post("/{id}/leave", h.Leave)
	r.Group(func(owner chi.Router) {
		owner.Use(auth.RequireAuth(h.sessions), auth.RequireOrganizer(h.q))
		owner.Post("/", h.Create)
		owner.Get("/{id}/manage", h.Manage)
		owner.Put("/{id}", h.Update)
		owner.Post("/{id}/status", h.Status)
	})
	return r
}
func page(r *http.Request) (int32, int32) {
	limit, offset := int64(20), int64(0)
	if n, e := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32); e == nil && n > 0 && n <= 100 {
		limit = n
	}
	if n, e := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32); e == nil && n >= 0 {
		offset = n
	}
	return int32(limit), int32(offset)
}
func resource(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	var id pgtype.UUID
	if id.Scan(chi.URLParam(r, "id")) != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid Space reference", nil)
		return id, false
	}
	return id, true
}
func failure(w http.ResponseWriter, e error) {
	if errors.Is(e, pgx.ErrNoRows) {
		platform.RespondError(w, 404, "NOT_FOUND", "Space not found", nil)
	} else {
		platform.RespondDatabaseError(w, e)
	}
}
func (h *Handler) begin(r *http.Request) (pgx.Tx, *db.Queries, error) {
	tx, e := h.pool.Begin(r.Context())
	if e != nil {
		return nil, nil, e
	}
	q := h.q.WithTx(tx)
	if e = q.SetWorkflowTimeout(r.Context()); e != nil {
		tx.Rollback(r.Context())
		return nil, nil, e
	}
	return tx, q, nil
}

// Public metadata deliberately excludes the persisted creation key and hash.
type spaceRecord struct {
	ID          pgtype.UUID        `json:"id"`
	OrganizerID pgtype.UUID        `json:"organizer_id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description"`
	Guidelines  string             `json:"guidelines"`
	WebsiteURL  string             `json:"website_url"`
	Status      string             `json:"status"`
	Revision    int64              `json:"revision"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

func view(s db.Spaces) spaceRecord {
	return spaceRecord{s.ID, s.OrganizerID, s.Name, s.Slug, s.Description, s.Guidelines, s.WebsiteUrl, s.Status, s.Revision, s.CreatedAt, s.UpdatedAt}
}
func reply(w http.ResponseWriter, status int, s db.Spaces) {
	platform.RespondJSON(w, status, struct {
		Data spaceRecord `json:"data"`
	}{view(s)})
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := page(r)
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len([]rune(query)) > 200 {
		platform.RespondError(w, 400, "INVALID_SEARCH", "Keep searches under 200 characters", nil)
		return
	}
	rows, e := h.q.ListPublicSpaces(r.Context(), db.ListPublicSpacesParams{Q: query, PageLimit: limit, PageOffset: offset})
	if e != nil {
		failure(w, e)
		return
	}
	if rows == nil {
		rows = []db.ListPublicSpacesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.ListPublicSpacesRow `json:"data"`
	}{rows})
}
func (h *Handler) Read(w http.ResponseWriter, r *http.Request) {
	row, e := h.q.ReadPublicSpace(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.ReadPublicSpaceRow `json:"data"`
	}{row})
}
func (h *Handler) Mine(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	limit, offset := page(r)
	rows, e := h.q.ListMySpaces(r.Context(), db.ListMySpacesParams{UserID: user.UserID, OwnedOnly: r.URL.Query().Get("view") == "owned", PageLimit: limit, PageOffset: offset})
	if e != nil {
		failure(w, e)
		return
	}
	if rows == nil {
		rows = []db.ListMySpacesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.ListMySpacesRow `json:"data"`
	}{rows})
}
func (h *Handler) Membership(w http.ResponseWriter, r *http.Request) {
	id, ok := resource(w, r)
	if !ok {
		return
	}
	user, _ := auth.UserFromContext(r.Context())
	row, e := h.q.ReadSpaceMembership(r.Context(), db.ReadSpaceMembershipParams{ID: id, UserID: user.UserID})
	if e != nil {
		failure(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.ReadSpaceMembershipRow `json:"data"`
	}{row})
}
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	id, ok := resource(w, r)
	if !ok {
		return
	}
	if _, e := h.q.ReadSpaceID(r.Context(), id); e != nil {
		failure(w, e)
		return
	}
	limit, offset := page(r)
	rows, e := h.q.ListSpaceEvents(r.Context(), db.ListSpaceEventsParams{SpaceID: id, Past: r.URL.Query().Get("period") == "past", PageLimit: limit, PageOffset: offset})
	if e != nil {
		failure(w, e)
		return
	}
	if rows == nil {
		rows = []db.Events{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.Events `json:"data"`
	}{rows})
}
func (h *Handler) Manage(w http.ResponseWriter, r *http.Request) {
	id, ok := resource(w, r)
	if !ok {
		return
	}
	owner, _ := auth.OrganizerFromContext(r.Context())
	s, e := h.q.ReadOwnedSpace(r.Context(), db.ReadOwnedSpaceParams{ID: id, OrganizerID: owner.ID})
	if e != nil {
		failure(w, e)
		return
	}
	reply(w, 200, s)
}

type details struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Guidelines  string `json:"guidelines"`
	Website     string `json:"website_url"`
	Revision    int64  `json:"revision"`
	Status      string `json:"status"`
}

func (d *details) valid() bool {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
	d.Guidelines = strings.TrimSpace(d.Guidelines)
	d.Website = strings.TrimSpace(d.Website)
	valid := len([]rune(d.Name)) > 0 && len([]rune(d.Name)) <= 120 && len([]rune(d.Description)) > 0 && len([]rune(d.Description)) <= 6000 && len([]rune(d.Guidelines)) <= 10000 && len(d.Website) <= 2048
	if d.Website != "" {
		u, e := url.ParseRequestURI(d.Website)
		valid = valid && e == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil
	}
	return valid
}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body details
	if platform.DecodeJSON(r, &body) != nil || !body.valid() {
		platform.RespondError(w, 422, "INVALID_SPACE", "Provide a Space name, description and valid optional HTTPS website", nil)
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if !workflows.ValidateKey(key) || key == "" {
		platform.RespondError(w, 400, "INVALID_IDEMPOTENCY_KEY", "Provide a random creation key with 16 to 128 characters", nil)
		return
	}
	owner, _ := auth.OrganizerFromContext(r.Context())
	canonical, _ := json.Marshal([4]string{body.Name, body.Description, body.Guidelines, body.Website})
	digest := sha256.Sum256(canonical)
	hash := hex.EncodeToString(digest[:])
	tx, q, e := h.begin(r)
	if e != nil {
		failure(w, e)
		return
	}
	defer tx.Rollback(r.Context())
	s, e := q.CreateSpace(r.Context(), db.CreateSpaceParams{OrganizerID: owner.ID, Name: body.Name, Description: body.Description, Guidelines: body.Guidelines, WebsiteUrl: body.Website, CreationKey: key, CreationHash: hash})
	status := 201
	if errors.Is(e, pgx.ErrNoRows) {
		status = 200
		s, e = q.ReadSpaceCreation(r.Context(), db.ReadSpaceCreationParams{OrganizerID: owner.ID, CreationKey: key})
	}
	if e != nil {
		failure(w, e)
		return
	}
	if s.CreationHash != hash {
		platform.RespondError(w, 409, "IDEMPOTENCY_CONFLICT", "This creation key was used with different Space details", nil)
		return
	}
	if e = q.JoinSpace(r.Context(), db.JoinSpaceParams{SpaceID: s.ID, UserID: owner.UserID}); e != nil {
		failure(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		failure(w, e)
		return
	}
	reply(w, status, s)
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) { h.write(w, r, false) }
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) { h.write(w, r, true) }
func (h *Handler) write(w http.ResponseWriter, r *http.Request, statusOnly bool) {
	id, ok := resource(w, r)
	if !ok {
		return
	}
	var body details
	if platform.DecodeJSON(r, &body) != nil || body.Revision < 1 || (!statusOnly && !body.valid()) || (statusOnly && body.Status != "active" && body.Status != "archived") {
		platform.RespondError(w, 422, "INVALID_SPACE", "Check the Space details and current revision", nil)
		return
	}
	owner, _ := auth.OrganizerFromContext(r.Context())
	tx, q, e := h.begin(r)
	if e != nil {
		failure(w, e)
		return
	}
	defer tx.Rollback(r.Context())
	s, e := q.LockSpace(r.Context(), id)
	if e != nil {
		failure(w, e)
		return
	}
	if s.OrganizerID != owner.ID {
		failure(w, pgx.ErrNoRows)
		return
	}
	if s.Revision != body.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The Space changed. Reload its latest details before saving", nil)
		return
	}
	if statusOnly {
		if s.Status != body.Status {
			s, e = q.SetSpaceStatus(r.Context(), db.SetSpaceStatusParams{ID: id, OrganizerID: owner.ID, Status: body.Status, Revision: body.Revision})
		}
	} else {
		s, e = q.UpdateSpace(r.Context(), db.UpdateSpaceParams{ID: id, OrganizerID: owner.ID, Name: body.Name, Description: body.Description, Guidelines: body.Guidelines, WebsiteUrl: body.Website, Revision: body.Revision})
	}
	if e != nil {
		failure(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		failure(w, e)
		return
	}
	reply(w, 200, s)
}
func (h *Handler) Join(w http.ResponseWriter, r *http.Request)  { h.membershipChange(w, r, true) }
func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) { h.membershipChange(w, r, false) }
func (h *Handler) membershipChange(w http.ResponseWriter, r *http.Request, joining bool) {
	id, ok := resource(w, r)
	if !ok {
		return
	}
	user, _ := auth.UserFromContext(r.Context())
	tx, q, e := h.begin(r)
	if e != nil {
		failure(w, e)
		return
	}
	defer tx.Rollback(r.Context())
	s, e := q.LockSpace(r.Context(), id)
	if e != nil {
		failure(w, e)
		return
	}
	membership, e := q.ReadSpaceMembership(r.Context(), db.ReadSpaceMembershipParams{ID: id, UserID: user.UserID})
	if e != nil {
		failure(w, e)
		return
	}
	if joining {
		if s.Status != "active" && !membership.Joined {
			platform.RespondError(w, 409, "SPACE_ARCHIVED", "This Space is archived and is not accepting new members", nil)
			return
		}
		e = q.JoinSpace(r.Context(), db.JoinSpaceParams{SpaceID: id, UserID: user.UserID})
	} else {
		if membership.Owned {
			platform.RespondError(w, 409, "SPACE_OWNER", "The owner cannot leave their own Space", nil)
			return
		}
		e = q.LeaveSpace(r.Context(), db.LeaveSpaceParams{SpaceID: id, UserID: user.UserID})
	}
	if e != nil {
		failure(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		failure(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data struct {
			Joined bool `json:"joined"`
			Owned  bool `json:"owned"`
		} `json:"data"`
	}{struct {
		Joined bool `json:"joined"`
		Owned  bool `json:"owned"`
	}{joining, membership.Owned}})
}
