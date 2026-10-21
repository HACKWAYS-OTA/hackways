package rsvps

import (
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/registrationforms"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	sessions *auth.SessionManager
	engine   *workflows.Engine
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	q := db.New(pool)
	return &Handler{pool, q, auth.NewSessionManager(q), workflows.New(pool)}
}
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(auth.OptionalAuth(h.sessions)).Post("/", h.CreateRSVP)
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.sessions))
		protected.Get("/", h.ListMine)
		protected.Post("/claim", h.ClaimGuest)
		protected.Get("/{id}", h.GetRSVP)
		protected.Get("/by-code/{code}", h.GetByPublicCode)
		protected.Post("/{id}/cancel", h.Change)
		protected.Post("/{id}/accept", h.Change)
		protected.Get("/{id}/ticket", h.Ticket)
		protected.Patch("/{id}", h.EditNotes)
		protected.Post("/{id}/transfer", h.Transfer)
		protected.With(auth.RequireOrganizer(h.queries)).Post("/{id}/approve", h.Change)
		protected.With(auth.RequireOrganizer(h.queries)).Post("/{id}/reject", h.Change)
	})
	return r
}

type CreateRSVPRequest struct {
	EventID      string                     `json:"event_id"`
	GuestName    string                     `json:"guest_name"`
	GuestEmail   string                     `json:"guest_email"`
	Notes        string                     `json:"notes"`
	Waitlist     bool                       `json:"waitlist"`
	FormRevision *int64                     `json:"form_revision"`
	Answers      []registrationforms.Answer `json:"answers"`
}

func (h *Handler) CreateRSVP(w http.ResponseWriter, r *http.Request) {
	var body CreateRSVPRequest
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_PAYLOAD", "Invalid or oversized JSON registration payload", nil)
		return
	}
	var id pgtype.UUID
	if err := id.Scan(body.EventID); err != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_EVENT_ID", "A valid event ID is required", nil)
		return
	}
	req := workflows.Request{EventID: id, Name: strings.TrimSpace(body.GuestName), Email: strings.ToLower(strings.TrimSpace(body.GuestEmail)), Notes: strings.TrimSpace(body.Notes), Key: r.Header.Get("Idempotency-Key"), RequestID: platform.GetRequestID(r.Context()), Waitlist: body.Waitlist}
	req.FormRevision = body.FormRevision
	req.Answers = body.Answers
	if user, ok := auth.UserFromContext(r.Context()); ok {
		req.UserID = user.UserID
		req.Name = user.DisplayName
		req.Email = strings.ToLower(user.Email)
	}
	parsed, err := mail.ParseAddress(req.Email)
	if err != nil || parsed.Address != req.Email || len(req.Email) > 320 || len(req.Name) > 200 || len(req.Notes) > 2000 {
		platform.RespondError(w, 400, "INVALID_DETAILS", "Provide a valid email and keep name/notes within their limits", nil)
		return
	}
	if !workflows.ValidateKey(req.Key) {
		platform.RespondError(w, 400, "INVALID_IDEMPOTENCY_KEY", "Use a random key of 16 to 128 printable characters", nil)
		return
	}
	result, replayed, err := h.engine.Register(r.Context(), req)
	if err != nil {
		respondFailure(w, err)
		return
	}
	status := 201
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
		status = 200
	}
	platform.RespondJSON(w, status, result)
}
func (h *Handler) GetRSVP(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "id")); err != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid registration ID", nil)
		return
	}
	row, err := h.queries.GetRSVPByIDAndUser(r.Context(), db.GetRSVPByIDAndUserParams{ID: id, UserID: user.UserID})
	if err != nil {
		respondFailure(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Rsvps `json:"data"`
	}{row})
}
func page(r *http.Request) (int32, int32) {
	limit, offset := int64(20), int64(0)
	if n, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32); err == nil && n > 0 && n <= 100 {
		limit = n
	}
	if n, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32); err == nil && n >= 0 {
		offset = n
	}
	return int32(limit), int32(offset)
}
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	limit, offset := page(r)
	rows, err := h.queries.ListAccountRegistrations(r.Context(), db.ListAccountRegistrationsParams{UserID: user.UserID, Limit: limit, Offset: offset})
	if err != nil {
		respondFailure(w, err)
		return
	}
	if rows == nil {
		rows = []db.ListAccountRegistrationsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data   []db.ListAccountRegistrationsRow `json:"data"`
		Limit  int32                            `json:"limit"`
		Offset int32                            `json:"offset"`
	}{rows, limit, offset})
}
func (h *Handler) Change(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "id")); err != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid registration ID", nil)
		return
	}
	action := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	var organizerID pgtype.UUID
	if organizer, ok := auth.OrganizerFromContext(r.Context()); ok {
		organizerID = organizer.ID
	}
	row, err := h.engine.Change(r.Context(), id, user.UserID, organizerID, action, platform.GetRequestID(r.Context()))
	if err != nil {
		respondFailure(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Rsvps `json:"data"`
	}{row})
}
func respondFailure(w http.ResponseWriter, err error) {
	var failure *workflows.Failure
	if errors.As(err, &failure) {
		platform.RespondError(w, failure.Status, failure.Code, failure.Message, nil)
	} else if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 404, "NOT_FOUND", "Resource not found", nil)
	} else {
		platform.RespondDatabaseError(w, err)
	}
}
