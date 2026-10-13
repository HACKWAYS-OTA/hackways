package events

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler serves HTTP endpoints for events.
type Handler struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewHandler constructs an events Handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:    pool,
		queries: db.New(pool),
	}
}

// Routes mounts event endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListEvents)
	r.Get("/{id}", h.GetEvent)
	return r
}

// ListEvents handles GET /api/v1/events.
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	query := r.URL.Query().Get("q")

	limit := int32(20)
	offset := int32(0)

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = int32(l)
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = int32(o)
	}

	var events []db.Events
	var err error

	if query != "" {
		events, err = h.queries.SearchPublishedEvents(r.Context(), db.SearchPublishedEventsParams{
			PlaintoTsquery: query,
			Limit:          limit,
			Offset:         offset,
		})
	} else {
		events, err = h.queries.ListPublishedEvents(r.Context(), db.ListPublishedEventsParams{
			Limit:  limit,
			Offset: offset,
		})
	}

	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}

	if events == nil {
		events = []db.Events{}
	}

	platform.RespondJSON(w, http.StatusOK, map[string]any{
		"data": events,
		"pagination": map[string]any{
			"limit":  limit,
			"offset": offset,
			"count":  len(events),
		},
	})
}

// GetEvent handles GET /api/v1/events/{id}.
func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid event ID format", nil)
		return
	}

	event, err := h.queries.GetEventByID(r.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			platform.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Event not found", nil)
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}

	platform.RespondJSON(w, http.StatusOK, map[string]any{
		"data": event,
	})
}
