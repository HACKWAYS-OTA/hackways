package workflows

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type StatusOption struct {
	Value db.RegistrationStatus `json:"value"`
	Label string                `json:"label"`
}

// OwnerStatusOptions is presentation advice; ChangeOwnerStatus repeats it under the event lock.
func OwnerStatusOptions(status db.RegistrationStatus, scanned, editable, canConfirm bool) []StatusOption {
	options := make([]StatusOption, 0, 3)
	if scanned || !editable || status == db.RegistrationStatusTransferred {
		return options
	}
	if canConfirm {
		switch status {
		case db.RegistrationStatusPendingApproval, db.RegistrationStatusOffered, db.RegistrationStatusCancelled, db.RegistrationStatusExpired, db.RegistrationStatusRejected:
			options = append(options, StatusOption{db.RegistrationStatusConfirmed, "Going"})
		}
	}
	if status == db.RegistrationStatusPendingApproval {
		options = append(options, StatusOption{db.RegistrationStatusRejected, "Declined"})
	}
	switch status {
	case db.RegistrationStatusConfirmed, db.RegistrationStatusPendingApproval, db.RegistrationStatusWaitlisted, db.RegistrationStatusOffered:
		options = append(options, StatusOption{db.RegistrationStatusCancelled, "Not going"})
	}
	return options
}

type OwnerStatusRequest struct {
	EventID, RegistrationID, ActorID, OrganizerID pgtype.UUID
	Status                                        db.RegistrationStatus
	ExpectedUpdatedAt                             time.Time
	Notify                                        bool
	Message, Key, RequestID                       string
}
type OwnerStatusResult struct {
	Data               db.Rsvps `json:"data"`
	NotificationQueued bool     `json:"notification_queued"`
}

func (e *Engine) ChangeOwnerStatus(ctx context.Context, req OwnerStatusRequest) (OwnerStatusResult, bool, error) {
	tx, q, event, err := e.Begin(ctx, req.EventID)
	if err != nil {
		return OwnerStatusResult{}, false, err
	}
	defer tx.Rollback(ctx)
	if !req.OrganizerID.Valid || event.OrganizerID != req.OrganizerID {
		return OwnerStatusResult{}, false, NotFound()
	}
	rsvp, err := q.GetWorkflowRSVP(ctx, req.RegistrationID)
	if err != nil || rsvp.EventID != event.ID {
		if err == nil || errors.Is(err, pgx.ErrNoRows) {
			err = NotFound()
		}
		return OwnerStatusResult{}, false, err
	}
	canonical, err := json.Marshal(struct {
		Status            db.RegistrationStatus
		ExpectedUpdatedAt time.Time
		Notify            bool
		Message           string
	}{req.Status, req.ExpectedUpdatedAt, req.Notify, req.Message})
	if err != nil {
		return OwnerStatusResult{}, false, err
	}
	scope := digest("owner-status:" + hex.EncodeToString(event.ID.Bytes[:]) + ":" + hex.EncodeToString(rsvp.ID.Bytes[:]) + ":" + hex.EncodeToString(req.OrganizerID.Bytes[:]))
	key, requestHash := digest(req.Key), digest(string(canonical))
	stored, err := q.GetIdempotentResponse(ctx, db.GetIdempotentResponseParams{ScopeHash: scope, KeyHash: key})
	if err == nil {
		if stored.RequestHash != requestHash {
			return OwnerStatusResult{}, false, fail("IDEMPOTENCY_CONFLICT", "This request key was used with different changes")
		}
		var result OwnerStatusResult
		if err = json.Unmarshal(stored.ResponseJson, &result); err != nil {
			return result, false, err
		}
		return result, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return OwnerStatusResult{}, false, err
	}
	if err = e.ExpireAndPromote(ctx, q, &event); err != nil {
		return OwnerStatusResult{}, false, err
	}
	rsvp, err = q.GetWorkflowRSVP(ctx, req.RegistrationID)
	if err != nil {
		return OwnerStatusResult{}, false, err
	}
	if !rsvp.UpdatedAt.Time.Equal(req.ExpectedUpdatedAt) {
		return OwnerStatusResult{}, false, fail("REGISTRATION_CHANGED", "This registration changed. Reload the guest details before updating its status")
	}
	ticket, ticketErr := q.GetTicketByRSVP(ctx, rsvp.ID)
	if ticketErr != nil && !errors.Is(ticketErr, pgx.ErrNoRows) {
		return OwnerStatusResult{}, false, ticketErr
	}
	scanned := ticketErr == nil && ticket.Status == "scanned"
	editable := event.Status == db.EventStatusPublished && event.StartsAt.Time.After(event.DatabaseNow.Time)
	allowed := false
	for _, option := range OwnerStatusOptions(rsvp.Status, scanned, editable, Open(event)) {
		if option.Value == req.Status {
			allowed = true
		}
	}
	if !allowed {
		return OwnerStatusResult{}, false, fail("STATUS_CHANGE_UNAVAILABLE", "This status change is unavailable. Checked-in, transferred and closed-event registrations cannot be changed")
	}
	confirmed, held := int32(0), int32(0)
	if rsvp.Status == db.RegistrationStatusConfirmed {
		confirmed--
	}
	if rsvp.Status == db.RegistrationStatusOffered {
		held--
	}
	if req.Status == db.RegistrationStatusConfirmed {
		if rsvp.Status != db.RegistrationStatusOffered && Full(event) {
			return OwnerStatusResult{}, false, fail("CAPACITY_FULL", "No place is available to mark this guest as going")
		}
		confirmed++
	}
	rsvp, err = q.SetWorkflowRSVPState(ctx, db.SetWorkflowRSVPStateParams{ID: rsvp.ID, Status: req.Status})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = fail("REGISTRATION_EXPIRED", "The application or offer expired. Reload the guest details")
		}
		return OwnerStatusResult{}, false, err
	}
	if err = Inventory(ctx, q, event.ID, confirmed, held); err != nil {
		return OwnerStatusResult{}, false, err
	}
	event.CurrentRsvpCount += confirmed
	event.HeldRsvpCount += held
	if req.Status == db.RegistrationStatusConfirmed {
		_, err = q.EnsureTicket(ctx, rsvp.ID)
	} else {
		err = q.RevokeWorkflowTicket(ctx, rsvp.ID)
	}
	if err != nil {
		return OwnerStatusResult{}, false, err
	}
	if err = Audit(ctx, q, req.ActorID, event.ID, rsvp.ID, "registration.owner_status_"+string(req.Status), req.RequestID); err != nil {
		return OwnerStatusResult{}, false, err
	}
	queued := false
	if req.Notify {
		queued, err = NotifyWithMessage(ctx, q, event, rsvp, "rsvp_"+string(req.Status), req.Message)
		if err != nil {
			return OwnerStatusResult{}, false, err
		}
		if !queued {
			return OwnerStatusResult{}, false, fail("RECIPIENT_UNAVAILABLE", "This guest has no notification email. Turn off Notify guest to continue")
		}
	}
	if err = e.ExpireAndPromote(ctx, q, &event); err != nil {
		return OwnerStatusResult{}, false, err
	}
	result := OwnerStatusResult{Data: rsvp, NotificationQueued: queued}
	response, err := json.Marshal(result)
	if err != nil {
		return OwnerStatusResult{}, false, err
	}
	if err = q.StoreIdempotentResponse(ctx, db.StoreIdempotentResponseParams{ScopeHash: scope, KeyHash: key, RequestHash: requestHash, ResponseJson: response}); err != nil {
		return OwnerStatusResult{}, false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return OwnerStatusResult{}, false, err
	}
	return result, false, nil
}
