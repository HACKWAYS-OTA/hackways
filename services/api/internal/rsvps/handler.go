package rsvps

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/api/db"
	"github.com/hackways/api/internal/auth"
	"github.com/hackways/api/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler serves HTTP endpoints for RSVPs.
type Handler struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewHandler constructs an RSVPs Handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:    pool,
		queries: db.New(pool),
	}
}

// Routes mounts RSVP endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.CreateRSVP)
	r.Get("/{id}", h.GetRSVP)
	return r
}

// CreateRSVPRequest represents incoming reservation payload.
type CreateRSVPRequest struct {
	EventID    string `json:"event_id"`
	GuestName  string `json:"guest_name"`
	GuestEmail string `json:"guest_email"`
	Notes      string `json:"notes"`
}

// CreateRSVP handles POST /api/v1/rsvps using the authoritative pessimistic locking capacity engine.
func (h *Handler) CreateRSVP(w http.ResponseWriter, r *http.Request) {
	var req CreateRSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Invalid request JSON payload", nil)
		return
	}

	if req.EventID == "" {
		platform.RespondError(w, http.StatusBadRequest, "MISSING_EVENT_ID", "Event ID is required", nil)
		return
	}

	var eventUUID pgtype.UUID
	if err := eventUUID.Scan(req.EventID); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Event ID format is invalid", nil)
		return
	}

	var userUUID pgtype.UUID
	userEmail := req.GuestEmail
	userName := req.GuestName

	// If authenticated session exists, link user ID directly
	if user, ok := auth.UserFromContext(r.Context()); ok && user != nil {
		userUUID = user.UserID
		if userEmail == "" {
			userEmail = user.Email
		}
		if userName == "" {
			userName = user.DisplayName
		}
	}

	if !userUUID.Valid && userEmail == "" {
		platform.RespondError(w, http.StatusBadRequest, "IDENTITY_REQUIRED", "Either authentication or guest_email is required", nil)
		return
	}

	ctx := r.Context()

	// 1. Begin isolated transaction
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	defer tx.Rollback(ctx)

	// 2. Enforce statement_timeout = '2000ms' as per AGENTS.md Rule 2.4
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = '2000ms';"); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}

	qtx := h.queries.WithTx(tx)

	// 3. Pessimistic lock row with SELECT ... FOR UPDATE
	event, err := qtx.LockEventForRSVP(ctx, eventUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			platform.RespondError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Event not found", nil)
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}

	// 4. Verify event is published and open
	if event.Status != db.EventStatusPublished {
		platform.RespondError(w, http.StatusBadRequest, "EVENT_NOT_OPEN", "Event is not open for registration", nil)
		return
	}

	// 5. Invariant check: capacity full
	if event.Capacity.Valid && event.CurrentRsvpCount >= event.Capacity.Int32 {
		platform.RespondError(w, http.StatusConflict, "CAPACITY_FULL", "Event is at full capacity", nil)
		return
	}

	// 6. Idempotent reservation insert (Database constraint ON CONFLICT DO NOTHING)
	params := db.CreateRSVPParams{
		EventID:    eventUUID,
		UserID:     userUUID,
		GuestName:  pgtype.Text{String: userName, Valid: userName != ""},
		GuestEmail: pgtype.Text{String: userEmail, Valid: userEmail != ""},
		Notes:      pgtype.Text{String: req.Notes, Valid: req.Notes != ""},
	}

	rsvp, err := qtx.CreateRSVP(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// ON CONFLICT DO NOTHING returned zero rows -> user is already registered!
			platform.RespondError(w, http.StatusConflict, "ALREADY_REGISTERED", "You are already registered for this event", nil)
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}

	// 7. Atomic capacity counter increment
	if err := qtx.IncrementRSVPCount(ctx, eventUUID); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}

	// 8. Enqueue confirmation email atomically into transactional outbox
	payloadJSON, _ := json.Marshal(map[string]any{
		"event_id":   req.EventID,
		"rsvp_id":    rsvp.ID,
		"guest_name": userName,
	})

	_, err = qtx.EnqueueEmailOutbox(ctx, db.EnqueueEmailOutboxParams{
		RsvpID:         rsvp.ID,
		EventID:        eventUUID,
		RecipientEmail: userEmail,
		RecipientName:  pgtype.Text{String: userName, Valid: userName != ""},
		TemplateType:   "rsvp_confirmation",
		PayloadJson:    payloadJSON,
	})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}

	// 9. Commit transaction atomically
	if err := tx.Commit(ctx); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}

	platform.RespondJSON(w, http.StatusCreated, map[string]any{
		"data":    rsvp,
		"message": "RSVP confirmed successfully",
	})
}

// GetRSVP handles GET /api/v1/rsvps/{id}.
func (h *Handler) GetRSVP(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid RSVP ID format", nil)
		return
	}

	rsvp, err := h.queries.GetRSVPByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			platform.RespondError(w, http.StatusNotFound, "NOT_FOUND", "RSVP record not found", nil)
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}

	platform.RespondJSON(w, http.StatusOK, map[string]any{
		"data": rsvp,
	})
}
