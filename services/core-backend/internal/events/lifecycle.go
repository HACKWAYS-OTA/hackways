package events

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/rsvps"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) mountManagement(r chi.Router) {
	r.Group(func(owner chi.Router) {
		owner.Use(auth.RequireAuth(h.sessions), auth.RequireOrganizer(h.queries))
		owner.Get("/{id}/manage", h.ManageEvent)
		owner.Get("/manage-by-slug/{slug}", h.ManageBySlug)
		owner.Put("/{id}/space", h.ChangeSpace)
		owner.Put("/{id}/appearance", h.ChangeAppearance)
		owner.Put("/{id}", h.UpdateEvent)
		owner.Delete("/{id}", h.DeleteDraft)
		owner.Post("/{id}/status", h.TransitionEvent)
		owner.Get("/{id}/registrations", h.Roster)
		owner.Get("/{id}/invitations", h.Invitations)
		owner.Post("/{id}/invitations", h.InviteAttendees)
		owner.Get("/{id}/registrations/by-code/{code}", h.GuestDetail)
		owner.Post("/{id}/registrations/by-code/{code}/status", h.ChangeGuestStatus)
		owner.Get("/{id}/registration-form", h.OwnerRegistrationForm)
		owner.Put("/{id}/registration-form", h.SaveRegistrationForm)
		owner.Get("/{id}/summary", h.GuestSummary)
		owner.Get("/{id}/audit", h.AuditTrail)
		owner.Post("/{id}/check-in", h.CheckIn)
		owner.Post("/{id}/check-in/undo", h.UndoCheckIn)
	})
}
func eventID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "id")); err != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid event ID", nil)
		return id, false
	}
	return id, true
}
func lifecycleError(w http.ResponseWriter, err error) {
	var failure *workflows.Failure
	if errors.As(err, &failure) {
		platform.RespondError(w, failure.Status, failure.Code, failure.Message, nil)
	} else if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 404, "NOT_FOUND", "Event not found", nil)
	} else {
		platform.RespondDatabaseError(w, err)
	}
}
func (h *Handler) owned(w http.ResponseWriter, r *http.Request) (db.Events, bool) {
	id, ok := eventID(w, r)
	if !ok {
		return db.Events{}, false
	}
	organizer, _ := auth.OrganizerFromContext(r.Context())
	event, err := h.queries.GetEventByIDAndOrganizer(r.Context(), db.GetEventByIDAndOrganizerParams{ID: id, OrganizerID: organizer.ID})
	if err != nil {
		lifecycleError(w, err)
		return event, false
	}
	return event, true
}
func (h *Handler) ManageEvent(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Events `json:"data"`
	}{event})
}
func (h *Handler) Roster(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	limit, offset := pagination(r)
	search, valid := managementSearch(w, r)
	if !valid {
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "all"
	}
	switch status {
	case "all", "confirmed", "pending_approval", "waitlisted", "offered", "cancelled", "rejected", "expired", "transferred", "checked_in", "not_arrived":
	default:
		platform.RespondError(w, 400, "INVALID_STATUS", "Choose a valid guest status", nil)
		return
	}
	rows, err := h.queries.EventRosterFiltered(r.Context(), db.EventRosterFilteredParams{EventID: event.ID, PageLimit: limit, PageOffset: offset, Search: search, StatusFilter: status})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if rows == nil {
		rows = []db.EventRosterFilteredRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data   []db.EventRosterFilteredRow `json:"data"`
		Limit  int32                       `json:"limit"`
		Offset int32                       `json:"offset"`
	}{rows, limit, offset})
}
func (h *Handler) AuditTrail(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	limit, offset := pagination(r)
	rows, err := h.queries.ListWorkflowAudit(r.Context(), db.ListWorkflowAuditParams{EventID: event.ID, Limit: limit, Offset: offset})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if rows == nil {
		rows = []db.ListWorkflowAuditRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.ListWorkflowAuditRow `json:"data"`
	}{rows})
}

type updateRequest struct {
	eventContent
	Visibility   *string    `json:"visibility"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Format       string     `json:"format"`
	StartsAt     time.Time  `json:"starts_at"`
	EndsAt       time.Time  `json:"ends_at"`
	Timezone     string     `json:"timezone"`
	Capacity     *int32     `json:"capacity"`
	Location     string     `json:"location_name"`
	Address      string     `json:"location_address"`
	OnlineURL    string     `json:"online_url"`
	Opens        *time.Time `json:"registration_opens_at"`
	Closes       *time.Time `json:"registration_closes_at"`
	Mode         string     `json:"registration_mode"`
	AllowGuest   bool       `json:"allow_guest"`
	Waitlist     bool       `json:"waitlist_enabled"`
	OfferMinutes int32      `json:"offer_minutes"`
	Revision     int64      `json:"revision"`
	Acknowledge  bool       `json:"acknowledge_attendees"`
}

func text(value string) pgtype.Text { return pgtype.Text{String: value, Valid: value != ""} }
func stamp(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}
func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := eventID(w, r)
	if !ok {
		return
	}
	var body updateRequest
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_PAYLOAD", "Provide one valid event JSON object", nil)
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	body.Description = strings.TrimSpace(body.Description)
	body.Location = strings.TrimSpace(body.Location)
	body.OnlineURL = strings.TrimSpace(body.OnlineURL)
	format := db.EventFormat(body.Format)
	valid := len([]rune(body.Title)) >= 1 && len([]rune(body.Title)) <= 200 && len([]rune(body.Description)) >= 1 && len([]rune(body.Description)) <= 20000 && len(body.Location) <= 1000 && len(body.Address) <= 2000 && len(body.OnlineURL) <= 2048
	valid = valid && body.EndsAt.After(body.StartsAt) && (format == db.EventFormatOnline || format == db.EventFormatInPerson || format == db.EventFormatHybrid) && (body.Capacity == nil || *body.Capacity > 0)
	_, zoneErr := time.LoadLocation(body.Timezone)
	valid = valid && body.Timezone != "" && zoneErr == nil && (body.Mode == "instant" || body.Mode == "approval") && body.OfferMinutes >= 5 && body.OfferMinutes <= 60
	if body.OnlineURL != "" {
		valid = valid && validHTTPSURL(body.OnlineURL)
	}
	if body.Opens != nil && body.Closes != nil {
		valid = valid && body.Closes.After(*body.Opens)
	}
	if body.Closes != nil {
		valid = valid && !body.Closes.After(body.StartsAt)
	}
	if body.Opens != nil {
		valid = valid && body.Opens.Before(body.StartsAt)
	}
	if body.Mode == "approval" && body.Waitlist {
		valid = false
	}
	if !valid || !body.eventContent.valid() || (body.Visibility != nil && *body.Visibility != "public" && *body.Visibility != "private") {
		platform.RespondError(w, 422, "INVALID_EVENT", "Check event details, time window, capacity, registration mode, and offer duration (5 to 60 minutes)", nil)
		return
	}
	engine := workflows.New(h.pool)
	tx, q, current, err := engine.Begin(r.Context(), id)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	organizer, _ := auth.OrganizerFromContext(r.Context())
	if current.OrganizerID != organizer.ID {
		lifecycleError(w, workflows.NotFound())
		return
	}
	if current.Revision != body.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The event changed; reload before saving", nil)
		return
	}
	if current.ManagementClosed {
		platform.RespondError(w, 409, "EDIT_CLOSED", "An administrator has closed editing for this event", nil)
		return
	}
	if !body.StartsAt.After(current.DatabaseNow.Time) && !body.StartsAt.Equal(current.StartsAt.Time) {
		platform.RespondError(w, 422, "INVALID_START", "Choose a future start time when changing the schedule", nil)
		return
	}
	if body.Capacity != nil && int64(*body.Capacity) < int64(current.CurrentRsvpCount)+int64(current.HeldRsvpCount) {
		platform.RespondError(w, 409, "CAPACITY_RESERVED", "Capacity cannot be less than confirmed places and active offers", nil)
		return
	}
	activity, activityErr := q.EventWorkflowActivity(r.Context(), id)
	if activityErr != nil {
		lifecycleError(w, activityErr)
		return
	}
	if !body.Waitlist && current.WaitlistEnabled && activity.WaitlistActive {
		platform.RespondError(w, 409, "WAITLIST_ACTIVE", "Resolve existing waitlist entries and offers before disabling the waitlist", nil)
		return
	}
	if body.Mode != current.RegistrationMode && activity.RegistrationActive {
		platform.RespondError(w, 409, "REGISTRATION_MODE_LOCKED", "Registration mode cannot change after places are reserved", nil)
		return
	}
	material := !body.StartsAt.Equal(current.StartsAt.Time) || !body.EndsAt.Equal(current.EndsAt.Time) || body.Timezone != current.Timezone || body.Location != current.LocationName.String || body.Address != current.LocationAddress.String || body.OnlineURL != current.OnlineUrl.String || format != current.Format
	if material && current.Status == db.EventStatusPublished && !body.Acknowledge {
		platform.RespondError(w, 409, "ATTENDEE_IMPACT", "Confirm attendee notification before changing schedule, format, or location", nil)
		return
	}
	var capacity pgtype.Int4
	if body.Capacity != nil {
		capacity = pgtype.Int4{Int32: *body.Capacity, Valid: true}
	}
	event, err := q.UpdateWorkflowEvent(r.Context(), db.UpdateWorkflowEventParams{ID: id, Title: body.Title, Description: body.Description, Format: format, StartsAt: stamp(&body.StartsAt), EndsAt: stamp(&body.EndsAt), Timezone: body.Timezone, Capacity: capacity, LocationName: text(body.Location), LocationAddress: text(body.Address), OnlineUrl: text(body.OnlineURL), RegistrationOpensAt: stamp(body.Opens), RegistrationClosesAt: stamp(body.Closes), RegistrationMode: body.Mode, AllowGuest: body.AllowGuest, WaitlistEnabled: body.Waitlist, OfferMinutes: body.OfferMinutes})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	event, err = body.eventContent.apply(r.Context(), q, event)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if body.Visibility != nil {
		event, err = q.SetEventVisibility(r.Context(), db.SetEventVisibilityParams{ID: id, Visibility: *body.Visibility})
		if err != nil {
			lifecycleError(w, err)
			return
		}
	}
	if material && current.Status == db.EventStatusPublished {
		payload, _ := json.Marshal(struct {
			Title    string    `json:"event_title"`
			StartsAt time.Time `json:"starts_at"`
		}{body.Title, body.StartsAt})
		err = q.EnqueueEventNotification(r.Context(), db.EnqueueEventNotificationParams{EventID: id, TemplateType: "event_updated", PayloadJson: payload})
		if err != nil {
			lifecycleError(w, err)
			return
		}
	}
	if err = workflows.Audit(r.Context(), q, organizer.UserID, id, id, "event.updated", platform.GetRequestID(r.Context())); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		lifecycleError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Events `json:"data"`
	}{event})
}

type transitionRequest struct {
	Status      db.EventStatus `json:"status"`
	Revision    int64          `json:"revision"`
	Acknowledge bool           `json:"acknowledge_attendees"`
}

func (h *Handler) TransitionEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := eventID(w, r)
	if !ok {
		return
	}
	var body transitionRequest
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_PAYLOAD", "Invalid status request", nil)
		return
	}
	engine := workflows.New(h.pool)
	tx, q, event, err := engine.Begin(r.Context(), id)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	if replayed, replayErr := workflows.ReplayAdminMutation(w, r, q); replayErr != nil {
		lifecycleError(w, replayErr)
		return
	} else if replayed {
		return
	}

	organizer, _ := auth.OrganizerFromContext(r.Context())
	if event.OrganizerID != organizer.ID {
		lifecycleError(w, workflows.NotFound())
		return
	}
	_, isAdmin := auth.AdminMutationFromContext(r.Context())
	if event.ManagementClosed && !isAdmin {
		platform.RespondError(w, 409, "EDIT_CLOSED", "An administrator has closed editing for this event", nil)
		return
	}
	adminClosingSame := isAdmin && !event.ManagementClosed && event.Status == body.Status && body.Status != db.EventStatusPublished
	if event.Status == body.Status && !adminClosingSame {
		current, readErr := q.GetEventByID(r.Context(), id)
		if readErr != nil {
			lifecycleError(w, readErr)
			return
		}
		platform.RespondJSON(w, 200, struct {
			Data db.Events `json:"data"`
		}{current})
		return
	}
	if event.Revision != body.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The event changed; reload before updating its status", nil)
		return
	}
	allowed := adminClosingSame || event.Status != db.EventStatusCancelled && (body.Status == db.EventStatusCancelled || body.Status == db.EventStatusPublished && (event.Status == db.EventStatusDraft || event.Status == db.EventStatusUnpublished) || body.Status == db.EventStatusUnpublished && event.Status == db.EventStatusPublished)
	if !allowed {
		platform.RespondError(w, 409, "INVALID_TRANSITION", "This event status transition is not allowed", nil)
		return
	}
	if body.Status == db.EventStatusPublished && !publishable(event) {
		platform.RespondError(w, 422, "NOT_PUBLISHABLE", "Complete event details and use a future schedule before publishing", nil)
		return
	}
	if body.Status == db.EventStatusCancelled && !body.Acknowledge {
		platform.RespondError(w, 409, "ATTENDEE_IMPACT", "Confirm cancellation and attendee notification", nil)
		return
	}
	if body.Status == db.EventStatusPublished && event.SpaceID.Valid {
		if _, spaceErr := q.LockOwnedActiveSpace(r.Context(), db.LockOwnedActiveSpaceParams{ID: event.SpaceID, OrganizerID: organizer.ID}); spaceErr != nil {
			if errors.Is(spaceErr, pgx.ErrNoRows) {
				platform.RespondError(w, 409, "SPACE_ARCHIVED", "Reactivate the Space or move this draft before publishing", nil)
			} else {
				lifecycleError(w, spaceErr)
			}
			return
		}
	}
	if body.Status == db.EventStatusCancelled {
		payload, _ := json.Marshal(struct {
			Title string `json:"event_title"`
		}{event.Title})
		if err = q.EnqueueEventNotification(r.Context(), db.EnqueueEventNotificationParams{EventID: id, TemplateType: "event_cancelled", PayloadJson: payload}); err != nil {
			lifecycleError(w, err)
			return
		}
		if err = q.CancelEventRegistrations(r.Context(), id); err != nil {
			lifecycleError(w, err)
			return
		}
		if err = q.RevokeEventTickets(r.Context(), id); err != nil {
			lifecycleError(w, err)
			return
		}
		if err = workflows.Inventory(r.Context(), q, id, -event.CurrentRsvpCount, -event.HeldRsvpCount); err != nil {
			lifecycleError(w, err)
			return
		}
	}
	updated, err := q.TransitionWorkflowEvent(r.Context(), db.TransitionWorkflowEventParams{ID: id, Status: body.Status, IsAdmin: isAdmin})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if err = workflows.Audit(r.Context(), q, organizer.UserID, id, id, "event."+string(body.Status), platform.GetRequestID(r.Context())); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = workflows.RememberAdminMutation(r.Context(), q, updated); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		lifecycleError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Events `json:"data"`
	}{updated})
}
func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	id, ok := eventID(w, r)
	if !ok {
		return
	}
	var body struct {
		Code     string `json:"code"`
		TicketID string `json:"ticket_id"`
		Reason   string `json:"reason"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_PAYLOAD", "Invalid check-in request", nil)
		return
	}
	var ticketID pgtype.UUID
	if strings.HasPrefix(body.Code, "T-") {
		var readErr error
		ticketID, readErr = h.queries.GetEventTicketByCode(r.Context(), db.GetEventTicketByCodeParams{PublicCode: body.Code, EventID: id})
		if readErr != nil {
			if errors.Is(readErr, pgx.ErrNoRows) {
				platform.RespondError(w, 422, "INVALID_TICKET", "This entry code is not valid for this event", nil)
			} else {
				lifecycleError(w, readErr)
			}
			return
		}
	} else if body.Code != "" {
		claims, err := rsvps.VerifyTicketCode(body.Code)
		if err != nil || claims.EventID != id {
			platform.RespondError(w, 422, "INVALID_TICKET", "The ticket is invalid, expired, or for another event", nil)
			return
		}
		ticketID = claims.TicketID
	} else if err := ticketID.Scan(body.TicketID); err != nil || !ticketID.Valid || !validManualReason(r, body.Reason) {
		platform.RespondError(w, 422, "MANUAL_REASON_REQUIRED", "Manual check-in requires a valid ticket ID and reason", nil)
		return
	}
	engine := workflows.New(h.pool)
	tx, q, event, err := engine.Begin(r.Context(), id)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	if replayed, replayErr := workflows.ReplayAdminMutation(w, r, q); replayErr != nil {
		lifecycleError(w, replayErr)
		return
	} else if replayed {
		return
	}

	organizer, _ := auth.OrganizerFromContext(r.Context())
	if event.OrganizerID != organizer.ID {
		lifecycleError(w, workflows.NotFound())
		return
	}
	now := event.DatabaseNow.Time
	if event.ManagementClosed || event.Status != db.EventStatusPublished || now.Before(event.StartsAt.Time.Add(-24*time.Hour)) || now.After(event.EndsAt.Time) {
		platform.RespondError(w, 409, "CHECK_IN_CLOSED", "Check-in opens 24 hours before the event and closes when it ends", nil)
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	ticket, err := q.CheckInWorkflowTicket(r.Context(), db.CheckInWorkflowTicketParams{ActorID: actor.UserID, TicketID: ticketID, EventID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		existing, readErr := q.GetWorkflowTicket(r.Context(), ticketID)
		if readErr == nil && existing.EventID == id && existing.Status == "scanned" {
			attendee, attendeeErr := q.ReadCheckInAttendee(r.Context(), db.ReadCheckInAttendeeParams{ID: ticketID, EventID: id})
			if attendeeErr != nil {
				lifecycleError(w, attendeeErr)
				return
			}
			platform.RespondJSON(w, 409, checkInErrorResponse{checkInError{"ALREADY_CHECKED_IN", "This attendee is already checked in. This scan was not counted again.", checkInDetails{attendee}}})
			return
		}
		lifecycleError(w, workflows.NotFound())
		return
	}
	if err != nil {
		lifecycleError(w, err)
		return
	}
	action := "ticket.scanned"
	if body.Code == "" {
		action = "ticket.manual_check_in"
	}
	if _, isAdmin := auth.AdminMutationFromContext(r.Context()); isAdmin {
		action = "admin." + action
	}
	metadata, _ := json.Marshal(struct {
		Reason string `json:"reason,omitempty"`
	}{body.Reason})
	if err = q.AppendWorkflowAudit(r.Context(), db.AppendWorkflowAuditParams{ActorUserID: actor.UserID, EventID: id, EntityID: ticketID, Action: action, RequestID: platform.GetRequestID(r.Context()), Metadata: metadata}); err != nil {
		lifecycleError(w, err)
		return
	}
	attendee, err := q.ReadCheckInAttendee(r.Context(), db.ReadCheckInAttendeeParams{ID: ticketID, EventID: id})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if err = workflows.RememberAdminMutation(r.Context(), q, ticket); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		lifecycleError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data     db.Tickets                `json:"data"`
		Attendee db.ReadCheckInAttendeeRow `json:"attendee"`
	}{ticket, attendee})
}

// DeleteDraft only hides inactive drafts; registrations and audit history are retained.
func (h *Handler) DeleteDraft(w http.ResponseWriter, r *http.Request) {
	id, ok := eventID(w, r)
	if !ok {
		return
	}
	revision, err := strconv.ParseInt(strings.Trim(r.Header.Get("If-Match"), "\""), 10, 64)
	if err != nil || revision < 1 {
		platform.RespondError(w, 428, "REVISION_REQUIRED", "Send the current revision in If-Match", nil)
		return
	}
	engine := workflows.New(h.pool)
	tx, q, event, err := engine.Begin(r.Context(), id)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	organizer, _ := auth.OrganizerFromContext(r.Context())
	if event.OrganizerID != organizer.ID {
		lifecycleError(w, workflows.NotFound())
		return
	}
	if event.Revision != revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The event changed; reload before removing it", nil)
		return
	}
	if event.ManagementClosed {
		platform.RespondError(w, 409, "EDIT_CLOSED", "An administrator has closed editing for this event", nil)
		return
	}
	activity, err := q.EventWorkflowActivity(r.Context(), id)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if event.Status != db.EventStatusDraft && event.Status != db.EventStatusUnpublished || activity.RegistrationActive {
		platform.RespondError(w, 409, "CANNOT_DELETE", "Only inactive draft or unpublished events can be removed; cancel live events instead", nil)
		return
	}
	if err = q.SoftDeleteWorkflowEvent(r.Context(), id); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = workflows.Audit(r.Context(), q, organizer.UserID, id, id, "event.deleted", platform.GetRequestID(r.Context())); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		lifecycleError(w, err)
		return
	}
	w.WriteHeader(204)
}

func validManualReason(r *http.Request, reason string) bool {
	limit := 500
	if _, admin := auth.AdminMutationFromContext(r.Context()); admin {
		limit = 4000
	}
	return len(strings.TrimSpace(reason)) >= 3 && len(reason) <= limit
}

// Location details may be supplied after publishing in every supported format.
// Supplied links still require HTTPS; core event and schedule checks stay strict.
func publishable(event db.LockWorkflowEventRow) bool {
	title, description := strings.TrimSpace(event.Title), strings.TrimSpace(event.Description)
	link := strings.TrimSpace(event.OnlineUrl.String)
	_, zoneErr := time.LoadLocation(event.Timezone)
	validFormat := event.Format == db.EventFormatInPerson || event.Format == db.EventFormatOnline || event.Format == db.EventFormatHybrid
	return len([]rune(title)) >= 1 && len([]rune(title)) <= 200 && len([]rune(description)) >= 1 && len([]rune(description)) <= 20000 &&
		validFormat && event.StartsAt.Valid && event.EndsAt.Valid && event.StartsAt.Time.After(event.DatabaseNow.Time) && event.EndsAt.Time.After(event.StartsAt.Time) &&
		event.Timezone != "" && zoneErr == nil && len(event.LocationName.String) <= 1000 && len(event.LocationAddress.String) <= 2000 && len(link) <= 2048 &&
		(link == "" || validHTTPSURL(link)) && (!event.Capacity.Valid || event.Capacity.Int32 > 0)
}
