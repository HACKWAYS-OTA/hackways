package events

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type checkInDetails struct {
	Attendee db.ReadCheckInAttendeeRow `json:"attendee"`
}
type checkInError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details checkInDetails `json:"details"`
}
type checkInErrorResponse struct {
	Error checkInError `json:"error"`
}

// UndoCheckIn corrects one specific admission under the same event-row lock as scanning.
func (h *Handler) UndoCheckIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	id, ok := eventID(w, r)
	if !ok {
		return
	}
	var body struct {
		TicketID    string    `json:"ticket_id"`
		CheckedInAt time.Time `json:"checked_in_at"`
	}
	var ticketID pgtype.UUID
	if platform.DecodeJSON(r, &body) != nil || ticketID.Scan(body.TicketID) != nil || !ticketID.Valid || body.CheckedInAt.IsZero() {
		platform.RespondError(w, 422, "INVALID_CHECK_IN", "Provide the current checked-in ticket", nil)
		return
	}
	tx, q, event, err := workflows.New(h.pool).Begin(r.Context(), id)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	owner, _ := auth.OrganizerFromContext(r.Context())
	if event.OrganizerID != owner.ID {
		lifecycleError(w, workflows.NotFound())
		return
	}
	if event.ManagementClosed || event.Status != db.EventStatusPublished {
		platform.RespondError(w, 409, "CHECK_IN_CLOSED", "Check-in corrections are unavailable for this event", nil)
		return
	}
	if _, err = q.ReadCheckInAttendee(r.Context(), db.ReadCheckInAttendeeParams{ID: ticketID, EventID: id}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			lifecycleError(w, workflows.NotFound())
		} else {
			lifecycleError(w, err)
		}
		return
	}
	ticket, err := q.UndoCheckInWorkflowTicket(r.Context(), db.UndoCheckInWorkflowTicketParams{TicketID: ticketID, EventID: id, CheckedInAt: pgtype.Timestamptz{Time: body.CheckedInAt, Valid: true}})
	if errors.Is(err, pgx.ErrNoRows) {
		existing, e := q.GetWorkflowTicket(r.Context(), ticketID)
		if e != nil {
			if errors.Is(e, pgx.ErrNoRows) {
				lifecycleError(w, workflows.NotFound())
			} else {
				lifecycleError(w, e)
			}
			return
		}
		if existing.EventID != id {
			lifecycleError(w, workflows.NotFound())
			return
		}
		if existing.Status == "issued" && !existing.CheckedInAt.Valid {
			platform.RespondJSON(w, 200, struct {
				Data db.GetWorkflowTicketRow `json:"data"`
			}{existing})
			return
		}
		platform.RespondError(w, 409, "CHECK_IN_CHANGED", "This check-in changed. Refresh before trying again.", nil)
		return
	}
	if err != nil {
		lifecycleError(w, err)
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	metadata, _ := json.Marshal(struct {
		CheckedInAt time.Time `json:"checked_in_at"`
	}{body.CheckedInAt})
	if err = q.AppendWorkflowAudit(r.Context(), db.AppendWorkflowAuditParams{ActorUserID: actor.UserID, EventID: id, EntityID: ticketID, Action: "ticket.check_in_undone", RequestID: platform.GetRequestID(r.Context()), Metadata: metadata}); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		lifecycleError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Tickets `json:"data"`
	}{ticket})
}
