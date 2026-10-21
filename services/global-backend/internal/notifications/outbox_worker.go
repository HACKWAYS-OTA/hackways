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

type templatePayload struct {
	Code           string             `json:"verification_code"`
	Title          string             `json:"event_title"`
	StartsAt       time.Time          `json:"starts_at"`
	EndsAt         time.Time          `json:"ends_at"`
	OfferExpiresAt pgtype.Timestamptz `json:"offer_expires_at"`
	CustomMessage  string             `json:"custom_message"`
	RecipientName  string             `json:"recipient_name"`
	OrganizerName  string             `json:"organizer_name"`
	Location       string             `json:"location"`
	EventURL       string             `json:"event_url"`
	ActionURL      string             `json:"action_url"`
	Reason         string             `json:"reason"`
	TicketCode     string             `json:"ticket_code"`
}

var registrationStateTemplates = map[string]struct{}{
	"rsvp_confirmation": {}, "rsvp_confirmed": {}, "rsvp_cancellation": {}, "rsvp_cancelled": {},
	"rsvp_pending_approval": {}, "rsvp_waitlisted": {}, "rsvp_offered": {}, "rsvp_transferred": {},
	"rsvp_expired": {}, "rsvp_rejected": {},
}

func isRegistrationStateTemplate(id string) bool {
	_, ok := registrationStateTemplates[id]
	return ok
}

func formatTemplateDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC1123)
}

func appendTemplateContext(body string, payload templatePayload) string {
	if date := formatTemplateDate(payload.StartsAt); date != "" {
		body += "\nScheduled start: " + date
	}
	if location := strings.TrimSpace(payload.Location); location != "" {
		body += "\nLocation: " + location
	}
	if code := strings.TrimSpace(payload.TicketCode); code != "" {
		body += "\nTicket: " + code
	}
	if message := strings.TrimSpace(payload.CustomMessage); message != "" {
		body += "\n\nMessage from the organizer:\n" + message
	}
	return body
}

func renderHTML(template EmailTemplate, subject, body string, payload templatePayload) string {
	if template.TextOnly {
		return ""
	}
	paragraphs := strings.Split(body, "\n\n")
	var content strings.Builder
	for _, paragraph := range paragraphs {
		lines := strings.Split(paragraph, "\n")
		for index, line := range lines {
			if index > 0 {
				content.WriteString("<br>")
			}
			content.WriteString(html.EscapeString(line))
		}
		content.WriteString("</p><p style=\"margin:0 0 18px;line-height:1.65;color:#3f3648;\">")
	}
	if len(paragraphs) > 0 {
		contentStr := content.String()
		contentStr = strings.TrimSuffix(contentStr, "</p><p style=\"margin:0 0 18px;line-height:1.65;color:#3f3648;\">")
		content.Reset()
		content.WriteString(contentStr)
	}
	actionURL := strings.TrimSpace(payload.ActionURL)
	if actionURL == "" {
		actionURL = strings.TrimSpace(payload.EventURL)
	}
	action := ""
	if actionURL != "" {
		action = `<p style="margin:26px 0 22px;"><a href="` + html.EscapeString(actionURL) + `" style="color:#201a2b;text-decoration:underline;font-weight:600;">Open Hackways</a></p>`
	}
	return `<div style="margin:0;padding:32px 20px;background:#ffffff;color:#201a2b;font-family:Arial,Helvetica,sans-serif;font-size:16px;line-height:1.65;">` +
		`<div style="max-width:600px;margin:0 auto;">` +
		`<p style="margin:0 0 42px;font-size:14px;letter-spacing:.08em;font-weight:700;color:#6d587f;">HACKWAYS</p>` +
		`<p style="margin:0 0 10px;font-size:13px;letter-spacing:.04em;text-transform:uppercase;color:#806a91;">` + html.EscapeString(template.Category) + `</p>` +
		`<h1 style="margin:0 0 22px;font-size:28px;line-height:1.2;font-weight:600;color:#201a2b;">` + html.EscapeString(subject) + `</h1>` +
		`<p style="margin:0 0 18px;line-height:1.65;color:#3f3648;">` + content.String() + `</p>` +
		action +
		`<div style="height:1px;background:#e8e1ed;margin:34px 0 18px;"></div>` +
		`<p style="margin:0;font-size:13px;line-height:1.55;color:#766b7f;">You are receiving this message because it relates to your Hackways account or registration.</p>` +
		`</div></div>`
}

func render(e db.ClaimEmailOutboxRow) (EmailMessage, error) {
	var payload templatePayload
	if err := json.Unmarshal(e.PayloadJson, &payload); err != nil {
		return EmailMessage{}, errors.New("notification payload is invalid")
	}
	template, ok := emailTemplate(e.TemplateType)
	if !ok {
		return EmailMessage{}, errors.New("unsupported notification template")
	}
	title := payload.Title
	if title == "" {
		title = "your event"
	}
	subject, body := template.Subject, template.Copy
	switch e.TemplateType {
	case "email_verification":
		body = "Your Hackways email verification code is: " + payload.Code + ". This code expires 15 minutes after it was requested. Do not share it."
	case "rsvp_confirmation", "rsvp_confirmed":
		body = "Your registration for " + title + " is confirmed."
	case "rsvp_cancellation", "rsvp_cancelled":
		body = "Your registration for " + title + " was cancelled."
	case "rsvp_pending_approval":
		body = "Your application for " + title + " is pending organizer approval. A place has not been reserved yet."
	case "rsvp_waitlisted":
		body = "You joined the waitlist for " + title + ". This is not a confirmed place."
	case "rsvp_offered":
		body = "A place is reserved temporarily for you at " + title + ". Sign in to accept your waitlist offer."
		if payload.OfferExpiresAt.Valid {
			body += " The offer expires " + payload.OfferExpiresAt.Time.UTC().Format(time.RFC1123) + "."
		}
	case "rsvp_transferred":
		body = "Your ticket for " + title + " was transferred to the verified recipient. Your previous ticket is no longer valid."
	case "rsvp_expired":
		body = "The review period for your application to " + title + " expired without approval. No place was confirmed."
	case "rsvp_rejected":
		body = "The organizer did not approve your application for " + title + ". Contact the organizer if you need more information."
	case "event_updated":
		body = "The organizer updated details for " + title + ". Review the current event page before attending."
	case "event_cancelled":
		body = title + " was cancelled. Your registration is no longer valid."
	}
	if title != "your event" && e.TemplateType != "email_verification" && !strings.Contains(body, title) {
		body = "For " + title + ":\n\n" + body
	}
	if isRegistrationStateTemplate(e.TemplateType) {
		body = appendTemplateContext(body, payload)
	} else {
		body = appendTemplateContext(body, templatePayload{StartsAt: payload.StartsAt, Location: payload.Location, TicketCode: payload.TicketCode})
	}
	return EmailMessage{Recipient: e.RecipientEmail, Subject: subject, BodyText: body, BodyHTML: renderHTML(template, subject, body, payload), OperationID: uuid(e.ID)}, nil
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
	if isRegistrationStateTemplate(e.TemplateType) && state.RegistrationStatus.Valid {
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
