package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/events"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/spaces"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) mountActions(r chi.Router) {
	eventsHandler, spacesHandler := events.NewHandler(h.pool), spaces.NewHandler(h.pool)
	r.Group(func(writes chi.Router) {
		writes.Use(auth.RequireAdminMutation)
		writes.Put("/events/{id}", h.EditEvent)
		writes.Post("/events/{id}/bin", h.BinEvent)
		writes.Post("/bin/events/{id}/restore", h.RestoreEvent)
		writes.Post("/events/{id}/status", h.forEvent(eventsHandler.TransitionEvent))
		writes.Put("/spaces/{id}", h.forSpace(spacesHandler.Update))
		writes.Post("/spaces/{id}/status", h.forSpace(spacesHandler.Status))
		writes.Post("/registrations/{id}/status", h.registrationStatus(eventsHandler))
		writes.Post("/tickets/{id}/check-in", h.ticketCheckIn(eventsHandler))
	})
}

func (h *Handler) withOwner(w http.ResponseWriter, r *http.Request, ownerID pgtype.UUID, next http.HandlerFunc) {
	owner, err := h.queries.AdminOrganizerByID(r.Context(), ownerID)
	if err != nil {
		detailError(w, err)
		return
	}
	ctx, ok := auth.WithAdminManagedOrganizer(r.Context(), &owner)
	if !ok {
		platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "Admin access required", nil)
		return
	}
	next(w, r.WithContext(ctx))
}
func (h *Handler) forEvent(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.detailID(w, r, "events")
		if !ok {
			return
		}
		event, err := h.queries.AdminEventDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		h.withOwner(w, routeParams(r, id, ""), event.OrganizerID, next)
	}
}
func (h *Handler) forSpace(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.detailID(w, r, "spaces")
		if !ok {
			return
		}
		space, err := h.queries.AdminSpaceDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		h.withOwner(w, routeParams(r, id, ""), space.OrganizerID, next)
	}
}

// The delegated management handlers use event route parameters. Replace only
// their scoped lookup parameters; retain the original URL for retry receipts.
func routeParams(r *http.Request, id pgtype.UUID, code string) *http.Request {
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", uuidString(id))
	if code != "" {
		ctx.URLParams.Add("code", code)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}

func (h *Handler) registrationStatus(next *events.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.detailID(w, r, "registrations")
		if !ok {
			return
		}
		reg, err := h.queries.AdminRegistrationDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		event, err := h.queries.AdminEventDetail(r.Context(), reg.EventID)
		if err != nil {
			detailError(w, err)
			return
		}
		h.withOwner(w, routeParams(r, reg.EventID, reg.PublicCode), event.OrganizerID, next.ChangeGuestStatus)
	}
}
func (h *Handler) ticketCheckIn(next *events.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.detailID(w, r, "tickets")
		if !ok {
			return
		}
		ticket, err := h.queries.AdminTicketDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		event, err := h.queries.AdminEventDetail(r.Context(), ticket.EventID)
		if err != nil {
			detailError(w, err)
			return
		}
		var body struct{}
		if err = platform.DecodeJSON(r, &body); err != nil {
			platform.RespondError(w, 400, "INVALID_PAYLOAD", "Provide an empty check-in object", nil)
			return
		}
		mutation, _ := auth.AdminMutationFromContext(r.Context())
		payload, _ := json.Marshal(struct {
			TicketID string `json:"ticket_id"`
			Reason   string `json:"reason"`
		}{uuidString(id), mutation.Reason})
		r.Body = io.NopCloser(strings.NewReader(string(payload)))
		h.withOwner(w, routeParams(r, ticket.EventID, ""), event.OrganizerID, next.CheckIn)
	}
}

func mutationError(w http.ResponseWriter, err error) {
	var failure *workflows.Failure
	if errors.As(err, &failure) {
		platform.RespondError(w, failure.Status, failure.Code, failure.Message, nil)
		return
	}
	detailError(w, err)
}

func (h *Handler) EditEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := h.detailID(w, r, "events")
	if !ok {
		return
	}
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Capacity    *int32 `json:"capacity"`
		Visibility  string `json:"visibility"`
		Revision    int64  `json:"revision"`
	}
	if platform.DecodeJSON(r, &body) != nil {
		platform.RespondError(w, 422, "INVALID_EVENT", "Provide the event details and current revision", nil)
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	body.Description = strings.TrimSpace(body.Description)
	if utf8.RuneCountInString(body.Title) < 1 || utf8.RuneCountInString(body.Title) > 200 || utf8.RuneCountInString(body.Description) < 1 || utf8.RuneCountInString(body.Description) > 20000 || body.Revision < 1 || (body.Visibility != "public" && body.Visibility != "private") || (body.Capacity != nil && *body.Capacity < 1) {
		platform.RespondError(w, 422, "INVALID_EVENT", "Check title, description, capacity and visibility", nil)
		return
	}
	tx, q, current, err := workflows.New(h.pool).Begin(r.Context(), id)
	if err != nil {
		mutationError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	if replayed, e := workflows.ReplayAdminMutation(w, r, q); e != nil {
		mutationError(w, e)
		return
	} else if replayed {
		return
	}
	if current.Revision != body.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The event changed; reload before saving", nil)
		return
	}
	if current.Status == db.EventStatusCancelled || !current.StartsAt.Time.After(current.DatabaseNow.Time) {
		platform.RespondError(w, 409, "EDIT_CLOSED", "Cancelled or started events cannot be edited", nil)
		return
	}
	if body.Capacity != nil && int64(*body.Capacity) < int64(current.CurrentRsvpCount)+int64(current.HeldRsvpCount) {
		platform.RespondError(w, 409, "CAPACITY_RESERVED", "Capacity cannot be less than confirmed places and active offers", nil)
		return
	}
	capacity := pgtype.Int4{}
	if body.Capacity != nil {
		capacity = pgtype.Int4{Int32: *body.Capacity, Valid: true}
	}
	if _, err = q.AdminEditEvent(r.Context(), db.AdminEditEventParams{ID: id, Title: body.Title, Description: body.Description, Capacity: capacity, Visibility: body.Visibility, Revision: body.Revision}); err != nil {
		mutationError(w, err)
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	mutation, _ := auth.AdminMutationFromContext(r.Context())
	type editState struct {
		Title       string      `json:"title"`
		Description string      `json:"description"`
		Capacity    pgtype.Int4 `json:"capacity"`
		Visibility  string      `json:"visibility"`
		Revision    int64       `json:"revision"`
	}
	metadata, _ := json.Marshal(struct {
		Reason string    `json:"reason"`
		Before editState `json:"before"`
		After  editState `json:"after"`
	}{mutation.Reason, editState{current.Title, current.Description, current.Capacity, current.Visibility, current.Revision}, editState{body.Title, body.Description, capacity, body.Visibility, current.Revision + 1}})
	if err = q.AppendWorkflowAudit(r.Context(), db.AppendWorkflowAuditParams{ActorUserID: actor.UserID, EventID: id, EntityID: id, Action: "admin.event.updated", RequestID: platform.GetRequestID(r.Context()), Metadata: metadata}); err != nil {
		mutationError(w, err)
		return
	}
	if current.Status == db.EventStatusPublished {
		payload, _ := json.Marshal(struct {
			Title    string    `json:"event_title"`
			StartsAt time.Time `json:"starts_at"`
		}{body.Title, current.StartsAt.Time})
		if err = q.EnqueueEventNotification(r.Context(), db.EnqueueEventNotificationParams{EventID: id, TemplateType: "event_updated", PayloadJson: payload}); err != nil {
			mutationError(w, err)
			return
		}
	}
	updated, err := q.AdminEventDetail(r.Context(), id)
	if err != nil {
		mutationError(w, err)
		return
	}
	if err = workflows.RememberAdminMutation(r.Context(), q, updated); err != nil {
		mutationError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		mutationError(w, err)
		return
	}
	sendDetail(w, updated, detailPage{})
}

func uuidString(id pgtype.UUID) string {
	b := id.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}
