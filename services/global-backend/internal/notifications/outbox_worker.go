package notifications

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MaxRetries       = 5
	BaseBackoffDelay = 15 * time.Second
	WorkerGoroutines = 2
	BatchSize        = 2
	DispatchTimeout  = 90 * time.Second
)

type OutboxWorker struct {
	queries      *db.Queries
	sender       EmailSender
	logger       *slog.Logger
	pollInterval time.Duration
	stopCh       chan struct{}
	stopOnce     sync.Once
	wg           sync.WaitGroup
	batchMu      sync.Mutex
}

func NewOutboxWorker(pool *pgxpool.Pool, sender EmailSender, logger *slog.Logger, interval time.Duration) *OutboxWorker {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return &OutboxWorker{queries: db.New(pool), sender: sender, logger: logger, pollInterval: interval, stopCh: make(chan struct{})}
}
func (w *OutboxWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case <-ticker.C:
				if _, err := w.ProcessNextBatch(ctx); err != nil {
					w.logger.Error("outbox claim failed", "error", err)
				}
			}
		}
	}()
}
func (w *OutboxWorker) Stop() { w.stopOnce.Do(func() { close(w.stopCh) }); w.wg.Wait() }
func (w *OutboxWorker) ProcessNextBatch(ctx context.Context) (int, error) {
	// Serialize manual triggers with the background loop, preserving the two-send bound.
	w.batchMu.Lock()
	defer w.batchMu.Unlock()
	entries, err := w.queries.ClaimEmailOutbox(ctx, BatchSize)
	if err != nil {
		return 0, err
	}
	var wg sync.WaitGroup
	for _, entry := range entries {
		wg.Add(1)
		go func(e db.ClaimEmailOutboxRow) { defer wg.Done(); w.dispatch(ctx, e) }(entry)
	}
	wg.Wait()
	return len(entries), nil
}
func uuid(id pgtype.UUID) string {
	b := hex.EncodeToString(id.Bytes[:])
	return b[:8] + "-" + b[8:12] + "-" + b[12:16] + "-" + b[16:20] + "-" + b[20:]
}
func render(e db.ClaimEmailOutboxRow) (EmailMessage, error) {
	var payload struct {
		Code           string             `json:"verification_code"`
		Title          string             `json:"event_title"`
		StartsAt       time.Time          `json:"starts_at"`
		OfferExpiresAt pgtype.Timestamptz `json:"offer_expires_at"`
		CustomMessage  string             `json:"custom_message"`
	}
	if err := json.Unmarshal(e.PayloadJson, &payload); err != nil {
		return EmailMessage{}, errors.New("notification payload is invalid")
	}
	title := payload.Title
	if title == "" {
		title = "your event"
	}
	var subject, body string
	switch e.TemplateType {
	case "email_verification":
		subject = "Verify your email"
		body = "Your Hackways email verification code is: " + payload.Code + ". This code expires 15 minutes after it was requested. Do not share it."
	case "rsvp_confirmation", "rsvp_confirmed":
		subject = "Your registration is confirmed"
		body = "Your registration for " + title + " is confirmed."
	case "rsvp_cancellation", "rsvp_cancelled":
		subject = "Your registration was cancelled"
		body = "Your registration for " + title + " was cancelled."
	case "rsvp_pending_approval":
		subject = "Your registration is awaiting review"
		body = "Your application for " + title + " is pending organizer approval. A place has not been reserved yet."
	case "rsvp_waitlisted":
		subject = "You joined the waitlist"
		body = "You joined the waitlist for " + title + ". This is not a confirmed place."
	case "rsvp_offered":
		subject = "A place is available"
		body = "A place is reserved temporarily for you at " + title + ". Sign in to accept your waitlist offer before " + payload.OfferExpiresAt.Time.UTC().Format(time.RFC1123) + "."
	case "rsvp_transferred":
		subject = "Your ticket was transferred"
		body = "Your ticket for " + title + " was transferred to the verified recipient. Your previous ticket is no longer valid."
	case "rsvp_expired":
		subject = "Registration application expired"
		body = "The review period for your application to " + title + " expired without approval. No place was confirmed."
	case "rsvp_rejected":
		subject = "Registration application update"
		body = "The organizer did not approve your application for " + title + ". Contact the organizer if you need more information."
	case "event_updated":
		subject = "Event details changed"
		body = "The organizer updated details for " + title + ". Review the current event page before attending."
	case "event_cancelled":
		subject = "Event cancelled"
		body = title + " was cancelled. Your registration is no longer valid."
	default:
		return EmailMessage{}, errors.New("unsupported notification template")
	}
	if !payload.StartsAt.IsZero() {
		body += "\nScheduled start: " + payload.StartsAt.UTC().Format(time.RFC1123)
	}
	if payload.CustomMessage != "" && strings.HasPrefix(e.TemplateType, "rsvp_") {
		body += "\n\nMessage from the organizer:\n" + payload.CustomMessage
	}
	return EmailMessage{Recipient: e.RecipientEmail, Subject: subject, BodyText: body, BodyHTML: "<p>" + html.EscapeString(body) + "</p>", OperationID: uuid(e.ID)}, nil
}
func (w *OutboxWorker) dispatch(ctx context.Context, e db.ClaimEmailOutboxRow) {
	msg, err := render(e)
	if err == nil {
		err = w.current(ctx, e)
	}
	if err == nil {
		// Configuration failures must not create a fictitious provider attempt.
		if _, missing := w.sender.(UnavailableEmailSender); missing {
			err = ErrEmailProviderUnavailable
		} else {
			first, beginErr := w.queries.BeginOutboxProviderAttempt(ctx, db.BeginOutboxProviderAttemptParams{ID: e.ID, LeaseToken: e.LeaseToken})
			if beginErr != nil {
				err = beginErr
			} else {
				msg.FirstSentAt = first.Time
				// Persist the operation timestamp before the first POST. A reclaimed
				// attempt must reconcile that identity, even with retry_count == 0.
				msg.FirstAttempt = !e.FirstSentAt.Valid
				sendCtx, cancel := context.WithTimeout(ctx, DispatchTimeout)
				err = w.sender.Send(sendCtx, msg)
				cancel()
			}
		}
	}
	// Record the outcome even when shutdown cancels an in-flight provider request.
	recordCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	if err == nil {
		affected, markErr := w.queries.MarkClaimedOutboxSent(recordCtx, db.MarkClaimedOutboxSentParams{ID: e.ID, LeaseToken: e.LeaseToken})
		if markErr != nil || affected != 1 {
			w.logger.Error("outbox acceptance not recorded", "id", e.ID, "error", markErr)
		}
		return
	}
	retry := e.RetryCount + 1
	status := db.NotificationStatusPending
	if retry >= MaxRetries || errors.Is(err, errSuperseded) {
		status = db.NotificationStatusFailed
	}
	exponent := e.RetryCount
	if exponent < 0 {
		exponent = 0
	}
	if exponent > 5 {
		exponent = 5
	}
	backoff := time.Duration(1<<exponent) * BaseBackoffDelay
	affected, markErr := w.queries.MarkClaimedOutboxFailed(recordCtx, db.MarkClaimedOutboxFailedParams{ID: e.ID, LeaseToken: e.LeaseToken, ScheduledAt: pgtype.Timestamptz{Time: time.Now().Add(backoff), Valid: true}, Status: status, ErrorMessage: pgtype.Text{String: err.Error(), Valid: true}})
	if markErr != nil || affected != 1 {
		w.logger.Error("outbox failure not recorded", "id", e.ID, "error", markErr)
	} else {
		w.logger.Warn("outbox send unconfirmed", "id", e.ID, "attempt", retry, "status", status, "error", fmt.Sprint(err))
	}
}

var errSuperseded = errors.New("notification superseded; no provider send was attempted")

func (w *OutboxWorker) current(ctx context.Context, e db.ClaimEmailOutboxRow) error {
	if e.TemplateType == "email_verification" {
		var payload struct {
			Code string `json:"verification_code"`
		}
		if err := json.Unmarshal(e.PayloadJson, &payload); err != nil {
			return err
		}
		sum := sha256.Sum256([]byte(payload.Code))
		active, err := w.queries.VerificationCodeActive(ctx, hex.EncodeToString(sum[:]))
		if err != nil {
			return err
		}
		if !active {
			return errSuperseded
		}
		return nil
	}
	state, err := w.queries.ReadOutboxCurrentState(ctx, e.ID)
	if err != nil {
		return err
	}
	if strings.HasPrefix(e.TemplateType, "rsvp_") && state.RegistrationStatus.Valid {
		expected := strings.TrimPrefix(e.TemplateType, "rsvp_")
		if expected == "confirmation" {
			expected = "confirmed"
		}
		if expected == "cancellation" {
			expected = "cancelled"
		}
		if string(state.RegistrationStatus.RegistrationStatus) != expected {
			return errSuperseded
		}
		if expected == "offered" && (!state.OfferExpiresAt.Valid || !state.OfferExpiresAt.Time.After(time.Now())) {
			return errSuperseded
		}
	}
	if e.TemplateType == "event_cancelled" && (!state.EventStatus.Valid || state.EventStatus.EventStatus != db.EventStatusCancelled) {
		return errSuperseded
	}
	if e.TemplateType == "event_updated" && (!state.EventStatus.Valid || state.EventStatus.EventStatus != db.EventStatusPublished) {
		return errSuperseded
	}
	return nil
}
