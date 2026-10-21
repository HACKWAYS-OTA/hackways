package events

import (
	"errors"

	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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

// Handler serves HTTP endpoints for events.
type Handler struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	sessions *auth.SessionManager
}

// NewHandler constructs an events Handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:     pool,
		queries:  db.New(pool),
		sessions: auth.NewSessionManager(db.New(pool)),
	}
}

// Routes mounts event endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(auth.RequireAuth(h.sessions)).Post("/organizer/profile", h.CreateOrganizerProfile)
	r.With(auth.RequireAuth(h.sessions), auth.RequireOrganizer(h.queries)).Get("/organizer/profile", h.GetOrganizerProfile)
	r.With(auth.RequireAuth(h.sessions), auth.RequireOrganizer(h.queries)).Get("/organizer", h.ListOrganizerEvents)
	r.With(auth.RequireAuth(h.sessions), auth.RequireOrganizer(h.queries)).Post("/", h.CreateEvent)
	r.Get("/", h.ListEvents)
	r.Get("/by-slug/{slug}", h.GetEventBySlug)
	r.Get("/by-slug/{slug}/registration-form", h.PublicRegistrationForm)
	r.Get("/{id}", h.GetEvent)
	h.mountManagement(r)
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

	event, err := h.queries.GetPublishedEventByID(r.Context(), id)
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

// GetEventBySlug reads published event details and public organizer information.
func (h *Handler) GetEventBySlug(w http.ResponseWriter, r *http.Request) {
	slug, err := url.PathUnescape(chi.URLParam(r, "slug"))
	if err != nil || len(slug) == 0 || len(slug) > 512 || !utf8.ValidString(slug) || strings.ContainsRune(slug, 0) {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_SLUG", "Invalid event slug", nil)
		return
	}
	event, err := h.queries.GetPublishedEventBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			platform.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Event not found", nil)
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, http.StatusOK, map[string]any{"data": event})
}

type createOrganizerRequest struct {
	DisplayName string `json:"display_name"`
}

// CreateOrganizerProfile creates the signed-in user's organizer profile.
func (h *Handler) CreateOrganizerProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		platform.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req createOrganizerRequest
	if err := decodeJSON(w, r, &req); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed or oversized JSON body", nil)
		return
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if len([]rune(req.DisplayName)) < 1 || len([]rune(req.DisplayName)) > 150 {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Organizer name must contain 1 to 150 characters", nil)
		return
	}

	organizer, err := h.queries.GetOrganizerByUserID(r.Context(), user.UserID)
	if err == nil {
		platform.RespondJSON(w, http.StatusOK, map[string]any{"data": organizer})
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		platform.RespondDatabaseError(w, err)
		return
	}

	organizer, err = h.queries.CreateOrganizer(r.Context(), db.CreateOrganizerParams{
		UserID:      user.UserID,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Another request may have created this user's profile concurrently.
			organizer, err = h.queries.GetOrganizerByUserID(r.Context(), user.UserID)
		}
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		platform.RespondJSON(w, http.StatusOK, map[string]any{"data": organizer})
		return
	}

	platform.RespondJSON(w, http.StatusCreated, map[string]any{"data": organizer})
}

// GetOrganizerProfile returns only the profile bound to the current session.
func (h *Handler) GetOrganizerProfile(w http.ResponseWriter, r *http.Request) {
	organizer, ok := auth.OrganizerFromContext(r.Context())
	if !ok {
		platform.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Organizer profile required", nil)
		return
	}
	platform.RespondJSON(w, http.StatusOK, map[string]any{"data": organizer})
}

// ListOrganizerEvents returns events owned by the authenticated organizer.
func (h *Handler) ListOrganizerEvents(w http.ResponseWriter, r *http.Request) {
	h.managedList(w, r)
}

type createEventRequest struct {
	eventContent
	Status           string `json:"status"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Format           string `json:"format"`
	StartsAt         string `json:"starts_at"`
	EndsAt           string `json:"ends_at"`
	Timezone         string `json:"timezone"`
	Capacity         *int32 `json:"capacity"`
	Location         string `json:"location_name"`
	OnlineURL        string `json:"online_url"`
	RegistrationMode string `json:"registration_mode"`
	SpaceSlug        string `json:"space_slug"`
	Theme            string `json:"theme"`
	ThemeColor       string `json:"theme_color"`
	ThemeFont        string `json:"theme_font"`
	ThemeStyle       string `json:"theme_style"`
	ThemeDisplay     string `json:"theme_display"`
	Visibility       string `json:"visibility"`
}

// CreateEvent publishes an event owned by the authenticated organizer.
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	organizer, ok := auth.OrganizerFromContext(r.Context())
	if !ok {
		platform.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Organizer profile required", nil)
		return
	}

	var req createEventRequest
	if err := decodeJSON(w, r, &req); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed or oversized JSON body", nil)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Timezone = strings.TrimSpace(req.Timezone)
	req.Location = strings.TrimSpace(req.Location)
	req.OnlineURL = strings.TrimSpace(req.OnlineURL)
	if len([]rune(req.Title)) < 1 || len([]rune(req.Title)) > 200 {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Event title must contain 1 to 200 characters", nil)
		return
	}
	if len([]rune(req.Description)) < 1 || len([]rune(req.Description)) > 20000 {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Event description must contain 1 to 20000 characters", nil)
		return
	}
	format := db.EventFormat(req.Format)
	if format != db.EventFormatInPerson && format != db.EventFormatOnline && format != db.EventFormatHybrid {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Event format must be in_person, online, or hybrid", nil)
		return
	}
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Start time must be an RFC3339 timestamp with an offset", nil)
		return
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil || !endsAt.After(startsAt) || !startsAt.After(time.Now()) {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "End time must be after a future start time", nil)
		return
	}
	if _, err := time.LoadLocation(req.Timezone); err != nil || req.Timezone == "" {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Timezone must be a valid IANA timezone", nil)
		return
	}
	if req.Capacity != nil && *req.Capacity < 1 {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Capacity must be a positive whole number", nil)
		return
	}
	if len(req.Location) > 1000 || len(req.OnlineURL) > 2048 {
		platform.RespondError(w, 400, "VALIDATION_ERROR", "Event location or URL exceeds its length limit", nil)
		return
	}
	if (format == db.EventFormatInPerson || format == db.EventFormatHybrid) && req.Location == "" {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Location is required for in-person events", nil)
		return
	}
	if (format == db.EventFormatOnline || format == db.EventFormatHybrid) && !validHTTPSURL(req.OnlineURL) {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "A valid HTTPS event URL is required for online events", nil)
		return
	}

	var capacity pgtype.Int4
	if req.Capacity != nil {
		capacity = pgtype.Int4{Int32: *req.Capacity, Valid: true}
	}
	status := db.EventStatus(req.Status)
	if status == "" {
		status = db.EventStatusPublished
	}
	if status != db.EventStatusDraft && status != db.EventStatusPublished {
		platform.RespondError(w, 400, "INVALID_STATUS", "New events must be draft or published", nil)
		return
	}
	mode := strings.TrimSpace(req.RegistrationMode)
	if mode == "" {
		mode = "instant"
	}
	if mode != "instant" && mode != "approval" {
		platform.RespondError(w, 400, "INVALID_REGISTRATION_MODE", "Choose instant registration or application approval", nil)
		return
	}
	if !validAppearance(&req) {
		platform.RespondError(w, 422, "INVALID_APPEARANCE", "Choose a supported theme, color, font and style", nil)
		return
	}
	if !req.eventContent.valid() {
		platform.RespondError(w, 422, "INVALID_EVENT_CONTENT", "Check FAQ, reward and HTTPS contact details", nil)
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	qtx := h.queries.WithTx(tx)
	if err = qtx.SetWorkflowTimeout(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	var spaceID pgtype.UUID
	if req.SpaceSlug != "" {
		space, readErr := qtx.LockOwnedActiveSpaceBySlug(r.Context(), db.LockOwnedActiveSpaceBySlugParams{Slug: req.SpaceSlug, OrganizerID: organizer.ID})
		if errors.Is(readErr, pgx.ErrNoRows) {
			platform.RespondError(w, 404, "SPACE_NOT_AVAILABLE", "Choose an active Space you own", nil)
			return
		}
		if readErr != nil {
			platform.RespondDatabaseError(w, readErr)
			return
		}
		spaceID = space.ID
	}
	event, err := qtx.CreateEventRecord(r.Context(), db.CreateEventRecordParams{
		Status:           status,
		OrganizerID:      organizer.ID,
		Format:           format,
		Title:            req.Title,
		Description:      req.Description,
		StartsAt:         pgtype.Timestamptz{Time: startsAt, Valid: true},
		EndsAt:           pgtype.Timestamptz{Time: endsAt, Valid: true},
		Timezone:         req.Timezone,
		Capacity:         capacity,
		LocationName:     pgtype.Text{String: req.Location, Valid: req.Location != ""},
		OnlineUrl:        pgtype.Text{String: req.OnlineURL, Valid: req.OnlineURL != ""},
		RegistrationMode: mode,
		SpaceID:          spaceID, Theme: req.Theme, ThemeColor: req.ThemeColor, ThemeFont: req.ThemeFont, ThemeStyle: req.ThemeStyle, ThemeDisplay: req.ThemeDisplay, Visibility: req.Visibility,
	})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	event, err = req.eventContent.apply(r.Context(), qtx, event)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = workflows.Audit(r.Context(), qtx, organizer.UserID, event.ID, event.ID, "event.created", platform.GetRequestID(r.Context())); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, http.StatusCreated, map[string]any{"data": event})
}

func decodeJSON[T any](w http.ResponseWriter, r *http.Request, dst *T) error {
	return platform.DecodeJSON(r, dst)
}

func pagination(r *http.Request) (int32, int32) {
	limit, offset := 20, 0
	if parsed, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && parsed > 0 && parsed <= 100 {
		limit = parsed
	}
	if parsed, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && parsed >= 0 {
		offset = parsed
	}
	return int32(limit), int32(offset)
}

func validHTTPSURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && strings.EqualFold(parsed.Scheme, "https") && parsed.Hostname() != "" && parsed.User == nil
}
