// Package workflows owns the free registration state machine. Every inventory
// mutation locks the event first, then reads registration state under that lock.
package workflows

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/registrationforms"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Failure struct {
	Status        int
	Code, Message string
}

func (f *Failure) Error() string      { return f.Message }
func fail(code, message string) error { return &Failure{409, code, message} }
func NotFound() error                 { return &Failure{404, "NOT_FOUND", "Resource not found"} }

type Engine struct {
	Pool *pgxpool.Pool
	Q    *db.Queries
}

func New(pool *pgxpool.Pool) *Engine { return &Engine{pool, db.New(pool)} }

type Request struct {
	EventID                            pgtype.UUID
	UserID                             pgtype.UUID
	Email, Name, Notes, Key, RequestID string
	Waitlist                           bool
	FormRevision                       *int64
	Answers                            []registrationforms.Answer
}
type Reply struct {
	Data    db.Rsvps `json:"data"`
	Message string   `json:"message"`
}

func digest(value string) string { h := sha256.Sum256([]byte(value)); return hex.EncodeToString(h[:]) }
func (e *Engine) Begin(ctx context.Context, id pgtype.UUID) (pgx.Tx, *db.Queries, db.LockWorkflowEventRow, error) {
	tx, err := e.Pool.Begin(ctx)
	if err != nil {
		return nil, nil, db.LockWorkflowEventRow{}, err
	}
	q := e.Q.WithTx(tx)
	if err = q.SetWorkflowTimeout(ctx); err != nil {
		tx.Rollback(ctx)
		return nil, nil, db.LockWorkflowEventRow{}, err
	}
	event, err := q.LockWorkflowEvent(ctx, id)
	if err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			err = NotFound()
		}
		return nil, nil, event, err
	}
	return tx, q, event, nil
}
func Audit(ctx context.Context, q *db.Queries, actor, event, entity pgtype.UUID, action, requestID string) error {
	metadata := []byte("{}")
	if mutation, ok := auth.AdminMutationFromContext(ctx); ok {
		user, _ := auth.UserFromContext(ctx)
		actor = user.UserID
		action = "admin." + action
		metadata, _ = json.Marshal(struct {
			Reason string `json:"reason"`
		}{mutation.Reason})
	}
	return q.AppendWorkflowAudit(ctx, db.AppendWorkflowAuditParams{ActorUserID: actor, EventID: event, EntityID: entity, Action: action, RequestID: requestID, Metadata: metadata})
}
func Notify(ctx context.Context, q *db.Queries, event db.LockWorkflowEventRow, rsvp db.Rsvps, kind string) error {
	_, err := NotifyWithMessage(ctx, q, event, rsvp, kind, "")
	return err
}
func NotifyWithMessage(ctx context.Context, q *db.Queries, event db.LockWorkflowEventRow, rsvp db.Rsvps, kind, message string) (bool, error) {
	context, err := q.ReadEventNotificationContext(ctx, event.ID)
	if err != nil {
		return false, err
	}
	origin := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/")
	if origin == "" {
		origin = "https://hackways.in"
	}
	eventURL := origin + "/events/" + url.PathEscape(context.Slug)
	actionURL := origin + "/my-events/" + url.PathEscape(rsvp.PublicCode)
	payload, err := json.Marshal(struct {
		Title          string             `json:"event_title"`
		StartsAt       time.Time          `json:"starts_at"`
		EndsAt         time.Time          `json:"ends_at"`
		OfferExpiresAt pgtype.Timestamptz `json:"offer_expires_at"`
		CustomMessage  string             `json:"custom_message,omitempty"`
		OrganizerName  string             `json:"organizer_name"`
		SpaceName      string             `json:"space_name,omitempty"`
		EventBannerURL string             `json:"event_banner_url,omitempty"`
		EventURL       string             `json:"event_url"`
		ActionURL      string             `json:"action_url"`
		Location       string             `json:"location,omitempty"`
	}{event.Title, event.StartsAt.Time, event.EndsAt.Time, rsvp.OfferExpiresAt, message, context.OrganizerName, context.SpaceName.String, context.BannerUrl.String, eventURL, actionURL, context.LocationName.String})
	if err != nil {
		return false, err
	}
	count, err := q.EnqueueRegistrationNotification(ctx, db.EnqueueRegistrationNotificationParams{RsvpID: rsvp.ID, TemplateType: kind, PayloadJson: payload})
	return count == 1, err
}
func Inventory(ctx context.Context, q *db.Queries, id pgtype.UUID, confirmed, held int32) error {
	return q.AdjustWorkflowInventory(ctx, db.AdjustWorkflowInventoryParams{EventID: id, ConfirmedDelta: confirmed, HeldDelta: held})
}
func Open(event db.LockWorkflowEventRow) bool {
	now := event.DatabaseNow.Time
	return event.Status == db.EventStatusPublished && event.StartsAt.Time.After(now) && (!event.RegistrationOpensAt.Valid || !event.RegistrationOpensAt.Time.After(now)) && (!event.RegistrationClosesAt.Valid || event.RegistrationClosesAt.Time.After(now))
}
func Full(event db.LockWorkflowEventRow) bool {
	return event.Capacity.Valid && int64(event.CurrentRsvpCount)+int64(event.HeldRsvpCount) >= int64(event.Capacity.Int32)
}

// Register replays an exact successful request before evaluating current capacity.
// A missing key preserves legacy conflict behavior. Keys must contain >=128 bits
// of client-generated randomness; only their digests are persisted.
func (e *Engine) Register(ctx context.Context, req Request) (Reply, bool, error) {
	tx, q, event, err := e.Begin(ctx, req.EventID)
	if err != nil {
		return Reply{}, false, err
	}
	defer tx.Rollback(ctx)
	scope := "guest:" + req.Email
	if req.UserID.Valid {
		scope = "user:" + hex.EncodeToString(req.UserID.Bytes[:])
	}
	scope = digest(hex.EncodeToString(req.EventID.Bytes[:]) + ":" + scope)
	key := digest(req.Key)
	// Keep the pre-form canonical shape for existing idempotency keys.
	answers := append([]registrationforms.Answer(nil), req.Answers...)
	sort.Slice(answers, func(i, j int) bool { return answers[i].Code < answers[j].Code })
	canonical, _ := json.Marshal(struct {
		Name, Email, Notes string
		Waitlist           bool
		FormRevision       *int64                     `json:",omitempty"`
		Answers            []registrationforms.Answer `json:",omitempty"`
	}{req.Name, req.Email, req.Notes, req.Waitlist, req.FormRevision, answers})
	requestHash := digest(string(canonical))
	if req.Key != "" {
		stored, readErr := q.GetIdempotentResponse(ctx, db.GetIdempotentResponseParams{ScopeHash: scope, KeyHash: key})
		if readErr == nil {
			if stored.RequestHash != requestHash {
				return Reply{}, false, fail("IDEMPOTENCY_CONFLICT", "This key was used with different registration details")
			}
			var reply Reply
			if err = json.Unmarshal(stored.ResponseJson, &reply); err != nil {
				return Reply{}, false, err
			}
			current, readErr := q.GetWorkflowRSVP(ctx, reply.Data.ID)
			if readErr != nil {
				return Reply{}, false, readErr
			}
			if current.UserID.Valid && (!req.UserID.Valid || current.UserID != req.UserID) {
				return Reply{}, false, &Failure{401, "ACCOUNT_REQUIRED", "Sign in to manage this account-linked registration"}
			}
			reply.Data = current
			reply.Message = "Registration " + string(current.Status)
			return reply, true, nil
		}
		if !errors.Is(readErr, pgx.ErrNoRows) {
			return Reply{}, false, readErr
		}
	}
	questions, err := q.ListRegistrationQuestions(ctx, event.ID)
	if err != nil {
		return Reply{}, false, err
	}
	if (req.FormRevision != nil && *req.FormRevision != event.RegistrationFormRevision) || (req.FormRevision == nil && (len(questions) > 0 || event.RegistrationNameRequired)) {
		return Reply{}, false, fail("FORM_CHANGED", "The organizer updated the registration questions. Reload the questions and try again")
	}
	form := registrationforms.Form{Revision: event.RegistrationFormRevision, NameRequired: event.RegistrationNameRequired}
	for _, item := range questions {
		form.Questions = append(form.Questions, registrationforms.Question{PublicCode: item.PublicCode, Label: item.Label, Description: item.Description, Kind: item.Kind, Required: item.Required, Options: item.Options})
	}
	snapshots, err := registrationforms.ValidateAnswers(form, req.Name, req.Answers)
	if err != nil {
		return Reply{}, false, &Failure{422, "INVALID_ANSWERS", err.Error()}
	}
	var existing db.Rsvps
	if req.UserID.Valid {
		existing, err = q.GetAccountEventRSVP(ctx, db.GetAccountEventRSVPParams{EventID: req.EventID, UserID: req.UserID})
	} else {
		existing, err = q.GetGuestEventRSVP(ctx, db.GetGuestEventRSVPParams{EventID: req.EventID, Lower: req.Email})
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Reply{}, false, err
	}
	found := err == nil
	// An email match identifies a conflict, never ownership. A guest must not
	// revive or edit an account-linked registration, including terminal states.
	if found && existing.UserID.Valid && (!req.UserID.Valid || existing.UserID != req.UserID) {
		return Reply{}, false, &Failure{401, "ACCOUNT_REQUIRED", "Sign in to manage this account-linked registration"}
	}
	if found && existing.Status != db.RegistrationStatusCancelled && existing.Status != db.RegistrationStatusExpired {
		return Reply{}, false, fail("ALREADY_REGISTERED", "A registration already exists; open your registrations to manage it")
	}
	if !Open(event) {
		return Reply{}, false, fail("REGISTRATION_CLOSED", "Registration is not currently open")
	}
	if !req.UserID.Valid && (!event.AllowGuest || req.Waitlist || event.RegistrationMode == "approval") {
		return Reply{}, false, &Failure{401, "ACCOUNT_REQUIRED", "Sign in to register for this event or join its waitlist"}
	}
	if err = e.ExpireAndPromote(ctx, q, &event); err != nil {
		return Reply{}, false, err
	}
	state := db.RegistrationStatusConfirmed
	if event.RegistrationMode == "approval" {
		state = db.RegistrationStatusPendingApproval
	}
	if req.Waitlist {
		if event.RegistrationMode != "instant" {
			return Reply{}, false, fail("APPROVAL_REQUIRED", "Submit an application; approval events do not bypass review through waitlist offers")
		}
		if !event.WaitlistEnabled {
			return Reply{}, false, fail("WAITLIST_DISABLED", "This event does not offer a waitlist")
		}
		if !Full(event) {
			return Reply{}, false, fail("SPACE_AVAILABLE", "A place is available; register instead of joining the waitlist")
		}
		state = db.RegistrationStatusWaitlisted
	} else if Full(event) && state == db.RegistrationStatusConfirmed {
		return Reply{}, false, fail("CAPACITY_FULL", "This event is full")
	}
	var approvalExpiry pgtype.Timestamptz
	if state == db.RegistrationStatusPendingApproval {
		expiry := event.DatabaseNow.Time.Add(7 * 24 * time.Hour)
		if event.StartsAt.Time.Before(expiry) {
			expiry = event.StartsAt.Time
		}
		if event.RegistrationClosesAt.Valid && event.RegistrationClosesAt.Time.Before(expiry) {
			expiry = event.RegistrationClosesAt.Time
		}
		approvalExpiry = pgtype.Timestamptz{Time: expiry, Valid: true}
	}
	var registration db.Rsvps
	if found {
		registration, err = q.ResetWorkflowRSVP(ctx, db.ResetWorkflowRSVPParams{ID: existing.ID, Status: state, ApprovalExpiresAt: approvalExpiry, Notes: pgtype.Text{String: req.Notes, Valid: req.Notes != ""}})
	} else {
		registration, err = q.InsertWorkflowRSVP(ctx, db.InsertWorkflowRSVPParams{EventID: req.EventID, UserID: req.UserID, GuestEmail: pgtype.Text{String: req.Email, Valid: req.Email != ""}, GuestName: pgtype.Text{String: req.Name, Valid: req.Name != ""}, Notes: pgtype.Text{String: req.Notes, Valid: req.Notes != ""}, Status: state, ApprovalExpiresAt: approvalExpiry})
	}
	if err != nil {
		return Reply{}, false, err
	}
	if state == db.RegistrationStatusConfirmed {
		// Inventory remains authoritative under this same event row lock.
		if err = Inventory(ctx, q, event.ID, 1, 0); err != nil {
			return Reply{}, false, err
		}
		if _, err = q.EnsureTicket(ctx, registration.ID); err != nil {
			return Reply{}, false, err
		}
	}
	answerJSON, err := json.Marshal(snapshots)
	if err != nil {
		return Reply{}, false, err
	}
	if err = q.InsertRegistrationSubmission(ctx, db.InsertRegistrationSubmissionParams{RsvpID: registration.ID, FormRevision: event.RegistrationFormRevision, Answers: answerJSON}); err != nil {
		return Reply{}, false, err
	}
	if err = Audit(ctx, q, req.UserID, event.ID, registration.ID, "registration."+string(state), req.RequestID); err != nil {
		return Reply{}, false, err
	}
	if err = Notify(ctx, q, event, registration, "rsvp_"+string(state)); err != nil {
		return Reply{}, false, err
	}
	reply := Reply{Data: registration, Message: "Registration " + string(state)}
	if req.Key != "" {
		data, marshalErr := json.Marshal(reply)
		if marshalErr != nil {
			return Reply{}, false, marshalErr
		}
		if err = q.StoreIdempotentResponse(ctx, db.StoreIdempotentResponseParams{ScopeHash: scope, KeyHash: key, RequestHash: requestHash, ResponseJson: data}); err != nil {
			return Reply{}, false, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Reply{}, false, err
	}
	return reply, false, nil
}

// Change handles account cancellation, accepting a reserved offer, and owner-only approval.
func (e *Engine) Change(ctx context.Context, id, actor, organizer pgtype.UUID, action, requestID string) (db.Rsvps, error) {
	lookup, err := e.Q.GetWorkflowRSVP(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = NotFound()
		}
		return db.Rsvps{}, err
	}
	tx, q, event, err := e.Begin(ctx, lookup.EventID)
	if err != nil {
		return db.Rsvps{}, err
	}
	defer tx.Rollback(ctx)
	rsvp, err := q.GetWorkflowRSVP(ctx, id)
	if err != nil {
		return rsvp, err
	}
	ownerAction := action == "approve" || action == "reject"
	if ownerAction {
		if !organizer.Valid || event.OrganizerID != organizer {
			return rsvp, NotFound()
		}
	} else if rsvp.UserID != actor || !actor.Valid {
		return rsvp, NotFound()
	}
	if err = e.ExpireAndPromote(ctx, q, &event); err != nil {
		return rsvp, err
	}
	rsvp, err = q.GetWorkflowRSVP(ctx, id)
	if err != nil {
		return rsvp, err
	}
	target := rsvp.Status
	confirmed, held := int32(0), int32(0)
	switch action {
	case "cancel":
		if rsvp.Status == db.RegistrationStatusCancelled {
			return rsvp, nil
		}
		ticket, ticketErr := q.GetTicketByRSVP(ctx, rsvp.ID)
		if ticketErr != nil && !errors.Is(ticketErr, pgx.ErrNoRows) {
			return rsvp, ticketErr
		}
		if ticketErr == nil && ticket.Status == "scanned" {
			return rsvp, fail("ALREADY_CHECKED_IN", "A checked-in registration cannot be cancelled")
		}
		if !event.StartsAt.Time.After(event.DatabaseNow.Time) {
			return rsvp, fail("CANCELLATION_CLOSED", "The event has started; contact the organizer")
		}
		switch rsvp.Status {
		case db.RegistrationStatusConfirmed:
			confirmed = -1
		case db.RegistrationStatusOffered:
			held = -1
		case db.RegistrationStatusWaitlisted, db.RegistrationStatusPendingApproval:
		default:
			return rsvp, fail("INVALID_STATE", "This registration cannot be cancelled")
		}
		target = db.RegistrationStatusCancelled
	case "accept":
		if rsvp.Status == db.RegistrationStatusConfirmed {
			return rsvp, nil
		}
		if rsvp.Status != db.RegistrationStatusOffered || !Open(event) {
			return rsvp, fail("OFFER_UNAVAILABLE", "The offer is unavailable or has expired")
		}
		confirmed = 1
		held = -1
		target = db.RegistrationStatusConfirmed
	case "approve":
		if rsvp.Status == db.RegistrationStatusConfirmed {
			return rsvp, nil
		}
		if rsvp.Status != db.RegistrationStatusPendingApproval || !Open(event) {
			return rsvp, fail("INVALID_STATE", "This application cannot be approved")
		}
		if Full(event) {
			return rsvp, fail("CAPACITY_FULL", "No place is available to approve this application")
		}
		confirmed = 1
		target = db.RegistrationStatusConfirmed
	case "reject":
		if rsvp.Status == db.RegistrationStatusRejected {
			return rsvp, nil
		}
		if rsvp.Status != db.RegistrationStatusPendingApproval {
			return rsvp, fail("INVALID_STATE", "Only pending applications can be rejected")
		}
		target = db.RegistrationStatusRejected
	default:
		return rsvp, &Failure{400, "INVALID_ACTION", "Unsupported registration action"}
	}
	rsvp, err = q.SetWorkflowRSVPState(ctx, db.SetWorkflowRSVPStateParams{ID: id, Status: target})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rsvp, fail("REGISTRATION_EXPIRED", "The offer or approval window expired before confirmation; refresh its status")
		}
		return rsvp, err
	}
	if confirmed != 0 || held != 0 {
		if err = Inventory(ctx, q, event.ID, confirmed, held); err != nil {
			return rsvp, err
		}
		event.CurrentRsvpCount += confirmed
		event.HeldRsvpCount += held
	}
	if target == db.RegistrationStatusConfirmed {
		_, err = q.EnsureTicket(ctx, id)
	} else {
		err = q.RevokeWorkflowTicket(ctx, id)
	}
	if err != nil {
		return rsvp, err
	}
	if err = Audit(ctx, q, actor, event.ID, id, "registration."+action, requestID); err != nil {
		return rsvp, err
	}
	if err = Notify(ctx, q, event, rsvp, "rsvp_"+string(target)); err != nil {
		return rsvp, err
	}
	if err = e.ExpireAndPromote(ctx, q, &event); err != nil {
		return rsvp, err
	}
	if err = tx.Commit(ctx); err != nil {
		return rsvp, err
	}
	return rsvp, nil
}

// ExpireAndPromote runs only under the event lock. Unaccepted offers reserve
// held inventory; pending approvals never consume inventory. Promotion is FIFO.
func (e *Engine) ExpireAndPromote(ctx context.Context, q *db.Queries, event *db.LockWorkflowEventRow) error {
	applications, err := q.ExpiredEventApplications(ctx, event.ID)
	if err != nil {
		return err
	}
	for _, application := range applications {
		expired, updateErr := q.SetWorkflowRSVPState(ctx, db.SetWorkflowRSVPStateParams{ID: application.ID, Status: db.RegistrationStatusExpired})
		if updateErr != nil {
			return updateErr
		}
		if err = Audit(ctx, q, pgtype.UUID{}, event.ID, application.ID, "registration.approval_expired", "worker"); err != nil {
			return err
		}
		if err = Notify(ctx, q, *event, expired, "rsvp_expired"); err != nil {
			return err
		}
	}
	expired, err := q.ExpiredEventOffers(ctx, event.ID)
	if err != nil {
		return err
	}
	for _, rsvp := range expired {
		if _, err = q.SetWorkflowRSVPState(ctx, db.SetWorkflowRSVPStateParams{ID: rsvp.ID, Status: db.RegistrationStatusExpired}); err != nil {
			return err
		}
		if err = Inventory(ctx, q, event.ID, 0, -1); err != nil {
			return err
		}
		event.HeldRsvpCount--
		if err = Audit(ctx, q, pgtype.UUID{}, event.ID, rsvp.ID, "registration.offer_expired", "worker"); err != nil {
			return err
		}
	}
	if !Open(*event) || !event.WaitlistEnabled {
		return nil
	}
	for promoted := 0; promoted < 50 && !Full(*event); promoted++ {
		rsvp, readErr := q.FirstWaitlistedRSVP(ctx, event.ID)
		if errors.Is(readErr, pgx.ErrNoRows) {
			break
		}
		if readErr != nil {
			return readErr
		}
		expiry := event.DatabaseNow.Time.Add(time.Duration(event.OfferMinutes) * time.Minute)
		if event.RegistrationClosesAt.Valid && event.RegistrationClosesAt.Time.Before(expiry) {
			expiry = event.RegistrationClosesAt.Time
		}
		if event.StartsAt.Time.Before(expiry) {
			expiry = event.StartsAt.Time
		}
		rsvp, err = q.SetWorkflowRSVPState(ctx, db.SetWorkflowRSVPStateParams{ID: rsvp.ID, Status: db.RegistrationStatusOffered, OfferExpiresAt: pgtype.Timestamptz{Time: expiry, Valid: true}})
		if err != nil {
			return err
		}
		if err = Inventory(ctx, q, event.ID, 0, 1); err != nil {
			return err
		}
		event.HeldRsvpCount++
		if err = Audit(ctx, q, pgtype.UUID{}, event.ID, rsvp.ID, "registration.offered", "worker"); err != nil {
			return err
		}
		if err = Notify(ctx, q, *event, rsvp, "rsvp_offered"); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) RunMaintenance(ctx context.Context, logger *slog.Logger) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func() {
				jobCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
				defer cancel()
				ids, err := e.Q.WorkflowMaintenanceEvents(jobCtx)
				if err != nil {
					logger.Error("workflow maintenance read failed", "error", err)
					return
				}
				for _, id := range ids {
					if jobCtx.Err() != nil {
						break
					}
					tx, q, event, beginErr := e.Begin(jobCtx, id)
					if beginErr != nil {
						logger.Warn("workflow maintenance lock failed", "error", beginErr)
						continue
					}
					err = e.ExpireAndPromote(jobCtx, q, &event)
					if err == nil {
						err = tx.Commit(jobCtx)
					}
					tx.Rollback(jobCtx)
					if err != nil {
						logger.Error("workflow maintenance failed", "event_id", id, "error", err)
					}
				}
				if err = e.Q.CleanupWorkflowKeys(jobCtx); err != nil {
					logger.Warn("workflow key cleanup failed", "error", err)
				}
			}()
		}
	}
}
func ValidateKey(key string) bool {
	return key == "" || len(key) >= 16 && len(key) <= 128 && strings.IndexFunc(key, func(c rune) bool { return c < '!' || c > '~' }) < 0
}
func (f Failure) String() string { return fmt.Sprintf("%s: %s", f.Code, f.Message) }
