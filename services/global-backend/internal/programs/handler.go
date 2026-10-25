package programs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool *pgxpool.Pool
	q    *db.Queries
	sm   *auth.SessionManager
}

func New(pool *pgxpool.Pool) *Handler {
	q := db.New(pool)
	return &Handler{pool, q, auth.NewSessionManager(q)}
}
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listPublic)
	r.Get("/{slug}", h.readPublic)
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(h.sm))
		r.Use(private)
		r.Get("/{slug}/application", h.mine)
		r.Post("/{slug}/application", h.saveDraft)
		r.Post("/{slug}/application/{action}", h.applicationAction)
	})
	return r
}
func (h *Handler) AdminRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(auth.RequirePlatformAdmin(h.sm, h.q))
	r.Get("/", h.adminList)
	r.Post("/", h.adminCreate)
	r.Get("/{slug}", h.adminRead)
	r.Post("/{slug}", h.adminUpdate)
	r.Get("/{slug}/applications", h.adminApplications)
	r.Get("/{slug}/applications/{id}", h.adminApplication)
	r.Post("/{slug}/applications/{id}/decision", h.decision)
	r.Post("/{slug}/release", h.release)
	return r
}
func private(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		next.ServeHTTP(w, r)
	})
}
func failure(w http.ResponseWriter, e error) {
	if errors.Is(e, pgx.ErrNoRows) {
		platform.RespondError(w, 404, "NOT_FOUND", "This program or application was not found", nil)
		return
	}
	var p *pgconn.PgError
	if errors.As(e, &p) && p.Code == "23505" {
		platform.RespondError(w, 409, "ALREADY_EXISTS", "This record already exists. Open it instead of creating a duplicate", nil)
		return
	}
	platform.RespondDatabaseError(w, e)
}
func invalid(w http.ResponseWriter, message string) {
	platform.RespondError(w, 400, "INVALID_PROGRAM_REQUEST", message, nil)
}
func conflict(w http.ResponseWriter) {
	platform.RespondError(w, 409, "RECORD_CHANGED", "This record changed. Reload it before continuing; your unsaved input is still available", nil)
}
func origin(w http.ResponseWriter, r *http.Request, admin bool) bool {
	allowed := r.Header.Get("Origin") == "https://hackways.in"
	if admin {
		allowed = r.Header.Get("Origin") == "https://admin.hackways.in"
	}
	if os.Getenv("APP_ENV") == "development" {
		ports := []string{"4000", "4001", "3001", "3101"}
		if admin {
			ports = []string{"4003", "3003", "3103"}
		}
		for _, port := range ports {
			allowed = allowed || r.Header.Get("Origin") == "http://localhost:"+port || r.Header.Get("Origin") == "http://127.0.0.1:"+port
		}
	}
	if !allowed {
		platform.RespondError(w, 403, "INVALID_ORIGIN", "Use the Hackways program workspace for this action", nil)
	}
	return allowed
}
func (h *Handler) begin(w http.ResponseWriter, r *http.Request) (pgx.Tx, *db.Queries, bool) {
	tx, e := h.pool.Begin(r.Context())
	if e != nil {
		failure(w, e)
		return nil, nil, false
	}
	q := h.q.WithTx(tx)
	if e = q.SetStorageWriteTimeout(r.Context()); e != nil {
		_ = tx.Rollback(r.Context())
		failure(w, e)
		return nil, nil, false
	}
	return tx, q, true
}
func open(p db.Programs) bool {
	now := time.Now()
	return p.Status == "open" && (!p.OpensAt.Valid || !now.Before(p.OpensAt.Time)) && (!p.ClosesAt.Valid || now.Before(p.ClosesAt.Time))
}
func actor(r *http.Request) *auth.SessionUser { u, _ := auth.UserFromContext(r.Context()); return u }
func audit(q *db.Queries, r *http.Request, p db.Programs, id pgtype.UUID, action string) error {
	return q.ProgramActivityAppend(r.Context(), db.ProgramActivityAppendParams{ProgramID: p.ID, ApplicationID: id, ActorID: actor(r).UserID, Action: action})
}
func offset(r *http.Request) int32 {
	n, e := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if e != nil || n < 0 || n > 100000 {
		return 0
	}
	return int32(n)
}
func applicationID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	var id pgtype.UUID
	e := id.Scan(chi.URLParam(r, "id"))
	if e != nil || !id.Valid {
		invalid(w, "Use a valid application reference")
		return id, false
	}
	return id, true
}
func (h *Handler) listPublic(w http.ResponseWriter, r *http.Request) {
	rows, e := h.q.ProgramListPublic(r.Context(), offset(r))
	if e != nil {
		failure(w, e)
		return
	}
	data := []Program{}
	more := len(rows) > 50
	if more {
		rows = rows[:50]
	}
	for _, p := range rows {
		data = append(data, view(p))
	}
	platform.RespondJSON(w, 200, struct {
		Data    []Program `json:"data"`
		HasNext bool      `json:"has_next"`
	}{data, more})
}
func (h *Handler) readPublic(w http.ResponseWriter, r *http.Request) {
	p, e := h.q.ProgramRead(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	if p.Status == "draft" || p.Status == "archived" {
		failure(w, pgx.ErrNoRows)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data      Program `json:"data"`
		Accepting bool    `json:"accepting_applications"`
	}{view(p), open(p)})
}

type mineView struct {
	db.ProgramApplicationMineRow
	Answers          json.RawMessage `json:"answers"`
	QuestionSnapshot json.RawMessage `json:"question_snapshot"`
}

func (h *Handler) mine(w http.ResponseWriter, r *http.Request) {
	p, e := h.q.ProgramRead(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	a, e := h.q.ProgramApplicationMine(r.Context(), db.ProgramApplicationMineParams{ProgramID: p.ID, UserID: actor(r).UserID})
	var data *mineView
	if e == nil {
		data = &mineView{a, json.RawMessage(a.Answers), json.RawMessage(a.QuestionSnapshot)}
	} else if !errors.Is(e, pgx.ErrNoRows) {
		failure(w, e)
		return
	}
	if data == nil && (p.Status == "draft" || p.Status == "archived") {
		failure(w, pgx.ErrNoRows)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Program     Program   `json:"program"`
		Application *mineView `json:"application"`
		Accepting   bool      `json:"accepting_applications"`
	}{view(p), data, open(p)})
}
func (h *Handler) saveDraft(w http.ResponseWriter, r *http.Request) {
	if !origin(w, r, false) {
		return
	}
	if !actor(r).EmailVerified {
		platform.RespondError(w, 403, "EMAIL_NOT_VERIFIED", "Verify your email before saving an application", nil)
		return
	}
	var body profileBody
	if platform.DecodeJSON(r, &body) != nil {
		invalid(w, "Check your application fields")
		return
	}
	tx, q, ok := h.begin(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	p, e := q.ProgramLock(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	if !open(p) {
		platform.RespondError(w, 409, "APPLICATIONS_CLOSED", "Applications are not open for this program", nil)
		return
	}
	if p.Revision != body.ProgramRevision {
		conflict(w)
		return
	}
	if e = body.validate(p, false); e != nil {
		invalid(w, e.Error())
		return
	}
	a, e := q.ProgramApplicationCreate(r.Context(), db.ProgramApplicationCreateParams{ProgramID: p.ID, UserID: actor(r).UserID})
	if e != nil {
		failure(w, e)
		return
	}
	if a.Status != "draft" {
		conflict(w)
		return
	}
	expected := body.Revision
	if expected == 0 && a.Revision == 1 {
		expected = 1
	}
	if a.Revision != expected {
		conflict(w)
		return
	}
	if body.Answers == nil {
		body.Answers = map[string]string{}
	}
	answers, e := json.Marshal(body.Answers)
	if e != nil {
		invalid(w, "Answers could not be read")
		return
	}
	a, e = q.ProgramApplicationSave(r.Context(), db.ProgramApplicationSaveParams{ID: a.ID, Revision: expected, FullName: body.FullName, City: body.City, Phone: body.Phone, Biography: body.Biography, ProfileUrl: body.ProfileURL, Proposal: body.Proposal, Age: body.Age, AgeConfirmed: body.AgeConfirmed, Answers: answers})
	if e != nil {
		failure(w, e)
		return
	}
	if e = audit(q, r, p, a.ID, "application.draft_saved"); e != nil {
		failure(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		failure(w, e)
		return
	}
	h.mine(w, r)
}
func (h *Handler) applicationAction(w http.ResponseWriter, r *http.Request) {
	if !origin(w, r, false) {
		return
	}
	var body struct {
		Revision        int64 `json:"revision"`
		ProgramRevision int64 `json:"program_revision"`
	}
	if platform.DecodeJSON(r, &body) != nil {
		invalid(w, "Use the current application revision")
		return
	}
	action := chi.URLParam(r, "action")
	if action != "submit" && action != "withdraw" {
		failure(w, pgx.ErrNoRows)
		return
	}
	tx, q, ok := h.begin(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	p, e := q.ProgramLock(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	a, e := q.ProgramApplicationLock(r.Context(), db.ProgramApplicationLockParams{ProgramID: p.ID, UserID: actor(r).UserID})
	if e != nil {
		failure(w, e)
		return
	}
	if (action == "submit" && a.Status == "submitted") || (action == "withdraw" && a.Status == "withdrawn") {
		h.mine(w, r)
		return
	}
	if a.Revision != body.Revision {
		conflict(w)
		return
	}
	if action == "submit" {
		if !actor(r).EmailVerified {
			platform.RespondError(w, 403, "EMAIL_NOT_VERIFIED", "Verify your email before submitting", nil)
			return
		}
		if !open(p) || p.Revision != body.ProgramRevision {
			platform.RespondError(w, 409, "PROGRAM_CHANGED", "Program availability or questions changed. Review the current form", nil)
			return
		}
		b := profile(a)
		if e = b.validate(p, true); e != nil {
			invalid(w, e.Error())
			return
		}
		if a.Status != "draft" {
			conflict(w)
			return
		}
		_, e = q.ProgramApplicationSubmit(r.Context(), db.ProgramApplicationSubmitParams{ID: a.ID, QuestionSnapshot: p.Questions})
		if e == nil {
			_, e = q.ProgramEnqueueReceipt(r.Context(), a.ID)
		}
	} else {
		_, e = q.ProgramApplicationWithdraw(r.Context(), a.ID)
	}
	if e != nil {
		failure(w, e)
		return
	}
	if e = q.ProgramTouchReview(r.Context(), p.ID); e != nil {
		failure(w, e)
		return
	}
	if e = audit(q, r, p, a.ID, "application."+action); e != nil {
		failure(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		failure(w, e)
		return
	}
	h.mine(w, r)
}

type adminProgram struct {
	Program
	ReviewRevision int64                      `json:"review_revision"`
	Stats          db.ProgramReviewStatsRow   `json:"stats"`
	Releases       []db.ProgramReleaseListRow `json:"releases"`
}

func (h *Handler) adminRead(w http.ResponseWriter, r *http.Request) {
	p, e := h.q.ProgramRead(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	s, e := h.q.ProgramReviewStats(r.Context(), p.ID)
	if e != nil {
		failure(w, e)
		return
	}
	releases, e := h.q.ProgramReleaseList(r.Context(), p.ID)
	if e != nil {
		failure(w, e)
		return
	}
	if releases == nil {
		releases = []db.ProgramReleaseListRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data adminProgram `json:"data"`
	}{adminProgram{view(p), p.ReviewRevision, s, releases}})
}
func (h *Handler) adminList(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	if !length(search, 0, 200) {
		invalid(w, "Search is too long")
		return
	}
	rows, e := h.q.ProgramAdminList(r.Context(), db.ProgramAdminListParams{Search: search, PageOffset: offset(r)})
	if e != nil {
		failure(w, e)
		return
	}
	type row struct {
		Slug         string             `json:"slug"`
		Title        string             `json:"title"`
		Summary      string             `json:"summary"`
		Status       string             `json:"status"`
		Applications int64              `json:"applications"`
		Pending      int64              `json:"pending"`
		UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
	}
	data := []row{}
	more := len(rows) > 20
	if more {
		rows = rows[:20]
	}
	for _, p := range rows {
		data = append(data, row{p.Slug, p.Title, p.Summary, p.Status, p.ApplicationCount, p.PendingCount, p.UpdatedAt})
	}
	platform.RespondJSON(w, 200, struct {
		Data    []row `json:"data"`
		HasNext bool  `json:"has_next"`
	}{data, more})
}
func (h *Handler) adminCreate(w http.ResponseWriter, r *http.Request) { h.saveProgram(w, r, true) }
func (h *Handler) adminUpdate(w http.ResponseWriter, r *http.Request) { h.saveProgram(w, r, false) }
func (h *Handler) saveProgram(w http.ResponseWriter, r *http.Request, create bool) {
	if !origin(w, r, true) {
		return
	}
	var body programBody
	if platform.DecodeJSON(r, &body) != nil {
		invalid(w, "Check the program fields")
		return
	}
	if e := body.validate(create); e != nil {
		invalid(w, e.Error())
		return
	}
	if body.Questions == nil {
		body.Questions = []Question{}
	}
	questions, e := json.Marshal(body.Questions)
	if e != nil {
		invalid(w, "Questions could not be saved")
		return
	}
	tx, q, ok := h.begin(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	var p db.Programs
	if create {
		p, e = q.ProgramCreate(r.Context(), db.ProgramCreateParams{Slug: body.Slug, Title: body.Title, Summary: body.Summary, Overview: body.Overview, Responsibilities: body.Responsibilities, Recognition: body.Recognition, Status: body.Status, MinimumAge: body.MinimumAge, Questions: questions, OpensAt: timestamp(body.OpensAt), ClosesAt: timestamp(body.ClosesAt), CreatedBy: actor(r).UserID})
	} else {
		p, e = q.ProgramLock(r.Context(), chi.URLParam(r, "slug"))
		if e != nil {
			failure(w, e)
			return
		}
		if p.Revision != body.Revision {
			conflict(w)
			return
		}
		p, e = q.ProgramUpdate(r.Context(), db.ProgramUpdateParams{ID: p.ID, Revision: body.Revision, Title: body.Title, Summary: body.Summary, Overview: body.Overview, Responsibilities: body.Responsibilities, Recognition: body.Recognition, Status: body.Status, MinimumAge: body.MinimumAge, Questions: questions, OpensAt: timestamp(body.OpensAt), ClosesAt: timestamp(body.ClosesAt)})
	}
	if e != nil {
		failure(w, e)
		return
	}
	if e = audit(q, r, p, pgtype.UUID{}, "program.saved"); e != nil {
		failure(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		failure(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data Program `json:"data"`
	}{view(p)})
}
func (h *Handler) adminApplications(w http.ResponseWriter, r *http.Request) {
	p, e := h.q.ProgramRead(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	if !length(search, 0, 200) {
		invalid(w, "Search is too long")
		return
	}
	rows, e := h.q.ProgramApplicationsAdmin(r.Context(), db.ProgramApplicationsAdminParams{ProgramID: p.ID, State: r.URL.Query().Get("status"), DecisionFilter: r.URL.Query().Get("decision"), Search: search, PageOffset: offset(r)})
	if e != nil {
		failure(w, e)
		return
	}
	type item struct {
		ID          pgtype.UUID        `json:"id"`
		Name        string             `json:"full_name"`
		City        string             `json:"city"`
		Email       string             `json:"email"`
		Status      string             `json:"status"`
		Decision    string             `json:"decision"`
		PublishedAt pgtype.Timestamptz `json:"published_at"`
		SubmittedAt pgtype.Timestamptz `json:"submitted_at"`
	}
	data := []item{}
	more := len(rows) > 20
	if more {
		rows = rows[:20]
	}
	for _, a := range rows {
		data = append(data, item{a.ID, a.FullName, a.City, a.Email, a.Status, a.Decision, a.PublishedAt, a.SubmittedAt})
	}
	platform.RespondJSON(w, 200, struct {
		Data    []item `json:"data"`
		HasNext bool   `json:"has_next"`
	}{data, more})
}
func (h *Handler) adminApplication(w http.ResponseWriter, r *http.Request) {
	id, ok := applicationID(w, r)
	if !ok {
		return
	}
	p, e := h.q.ProgramRead(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	a, e := h.q.ProgramApplicationAdminRead(r.Context(), db.ProgramApplicationAdminReadParams{ProgramID: p.ID, ID: id})
	if e != nil {
		failure(w, e)
		return
	}
	type app struct {
		db.ProgramApplicationAdminReadRow
		Answers          json.RawMessage `json:"answers"`
		QuestionSnapshot json.RawMessage `json:"question_snapshot"`
	}
	platform.RespondJSON(w, 200, struct {
		Program     Program `json:"program"`
		Application app     `json:"application"`
	}{view(p), app{a, json.RawMessage(a.Answers), json.RawMessage(a.QuestionSnapshot)}})
}
func (h *Handler) decision(w http.ResponseWriter, r *http.Request) {
	if !origin(w, r, true) {
		return
	}
	id, ok := applicationID(w, r)
	if !ok {
		return
	}
	var b struct {
		Decision string `json:"decision"`
		Notes    string `json:"notes"`
		Revision int64  `json:"revision"`
	}
	if platform.DecodeJSON(r, &b) != nil || !length(b.Notes, 0, 3000) || (b.Decision != "pending" && b.Decision != "accepted" && b.Decision != "rejected" && b.Decision != "waitlisted") {
		invalid(w, "Choose a decision and keep private notes within 3,000 characters")
		return
	}
	tx, q, ok := h.begin(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	p, e := q.ProgramLock(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	a, e := q.ProgramApplicationLockID(r.Context(), db.ProgramApplicationLockIDParams{ProgramID: p.ID, ID: id})
	if e != nil {
		failure(w, e)
		return
	}
	if a.Revision != b.Revision || a.Status != "submitted" || a.PublishedAt.Valid {
		conflict(w)
		return
	}
	_, e = q.ProgramDecisionSave(r.Context(), db.ProgramDecisionSaveParams{ID: id, Revision: b.Revision, Decision: b.Decision, ReviewerNotes: b.Notes, ReviewedBy: actor(r).UserID})
	if e != nil {
		failure(w, e)
		return
	}
	if e = q.ProgramTouchReview(r.Context(), p.ID); e != nil {
		failure(w, e)
		return
	}
	if e = audit(q, r, p, id, "application.decision_saved"); e != nil {
		failure(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		failure(w, e)
		return
	}
	h.adminApplication(w, r)
}
func (h *Handler) release(w http.ResponseWriter, r *http.Request) {
	if !origin(w, r, true) {
		return
	}
	var b struct {
		ReviewRevision int64 `json:"review_revision"`
		Count          int64 `json:"count"`
		Confirmed      bool  `json:"confirmed"`
	}
	if platform.DecodeJSON(r, &b) != nil || !b.Confirmed || b.Count < 1 {
		invalid(w, "Review and confirm the decisions before releasing them")
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) < 16 || len(key) > 128 || strings.IndexFunc(key, func(c rune) bool { return c < 33 || c > 126 }) >= 0 {
		invalid(w, "Use a valid release action key")
		return
	}
	raw, _ := json.Marshal(b)
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	tx, q, ok := h.begin(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	p, e := q.ProgramLock(r.Context(), chi.URLParam(r, "slug"))
	if e != nil {
		failure(w, e)
		return
	}
	previous, e := q.ProgramReleaseLookup(r.Context(), db.ProgramReleaseLookupParams{ProgramID: p.ID, RequestKey: key})
	if e == nil {
		if previous.PayloadHash != hash {
			conflict(w)
			return
		}
		platform.RespondJSON(w, 200, struct {
			Released int32 `json:"released"`
		}{previous.ApplicationCount})
		return
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		failure(w, e)
		return
	}
	s, e := q.ProgramReviewStats(r.Context(), p.ID)
	if e != nil {
		failure(w, e)
		return
	}
	if p.ReviewRevision != b.ReviewRevision || s.Ready != b.Count {
		conflict(w)
		return
	}
	if s.Pending != 0 {
		platform.RespondError(w, 409, "REVIEW_INCOMPLETE", "Record a decision for every submitted application before releasing results", nil)
		return
	}
	n, e := q.ProgramReleaseDecisions(r.Context(), p.ID)
	if e != nil {
		failure(w, e)
		return
	}
	if n != b.Count {
		conflict(w)
		return
	}
	_, e = q.ProgramReleaseCreate(r.Context(), db.ProgramReleaseCreateParams{ProgramID: p.ID, RequestKey: key, PayloadHash: hash, ApplicationCount: int32(n), ReleasedBy: actor(r).UserID})
	if e != nil {
		failure(w, e)
		return
	}
	if e = q.ProgramTouchReview(r.Context(), p.ID); e != nil {
		failure(w, e)
		return
	}
	if e = audit(q, r, p, pgtype.UUID{}, "program.results_released"); e != nil {
		failure(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		failure(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Released int64 `json:"released"`
	}{n})
}
