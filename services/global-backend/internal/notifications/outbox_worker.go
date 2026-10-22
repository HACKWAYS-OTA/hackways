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
	"net/url"
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
	queries        *db.Queries
	sender         EmailSender
	logger         *slog.Logger
	pollInterval   time.Duration
	stopCh         chan struct{}
	stopOnce       sync.Once
	wg             sync.WaitGroup
	batchMu        sync.Mutex
	templateFilter []string
}

// SetTemplateFilter is configured before Start. The SQL claim itself excludes
// other templates, so a scoped review worker never leases historical emails.
func (w *OutboxWorker) SetTemplateFilter(values []string) error {
	if len(values) > 10 {
		return errors.New("too many outbox template filters")
	}
	for _, value := range values {
		if _, ok := emailTemplate(value); !ok {
			return errors.New("unknown outbox template filter")
		}
	}
	w.templateFilter = append([]string(nil), values...)
	return nil
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
	entries, err := w.queries.ClaimEmailOutbox(ctx, db.ClaimEmailOutboxParams{BatchLimit: BatchSize, TemplateTypes: w.templateFilter})
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
	CampaignDesign   *CampaignDesign    `json:"campaign_design,omitempty"`
	PromotionSubject string             `json:"promotion_subject"`
	PromotionBody    string             `json:"promotion_body"`
	ActionLabel      string             `json:"action_label"`
	UnsubscribeURL   string             `json:"unsubscribe_url"`
	IsTest           bool               `json:"is_test"`
	VerificationHash string             `json:"verification_hash"`
	DeviceName       string             `json:"device_name"`
	BrowserName      string             `json:"browser_name"`
	ApproxLocation   string             `json:"approx_location"`
	DeviceLabel      string             `json:"device_label"`
	SignedInAt       time.Time          `json:"signed_in_at"`
	Code             string             `json:"verification_code"`
	Title            string             `json:"event_title"`
	StartsAt         time.Time          `json:"starts_at"`
	EndsAt           time.Time          `json:"ends_at"`
	OfferExpiresAt   pgtype.Timestamptz `json:"offer_expires_at"`
	CustomMessage    string             `json:"custom_message"`
	RecipientName    string             `json:"recipient_name"`
	OrganizerName    string             `json:"organizer_name"`
	SpaceName        string             `json:"space_name"`
	Location         string             `json:"location"`
	EventURL         string             `json:"event_url"`
	ActionURL        string             `json:"action_url"`
	EventBannerURL   string             `json:"event_banner_url"`
	Reason           string             `json:"reason"`
	TicketCode       string             `json:"ticket_code"`
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
	if organizer := strings.TrimSpace(payload.OrganizerName); organizer != "" {
		body += "\nOrganiser: " + organizer
	}
	if space := strings.TrimSpace(payload.SpaceName); space != "" {
		body += "\nSpace: " + space
	}
	if code := strings.TrimSpace(payload.TicketCode); code != "" {
		body += "\nTicket: " + code
	}
	if message := strings.TrimSpace(payload.CustomMessage); message != "" {
		body += "\n\nMessage from the organizer:\n" + message
	}
	return body
}

func safeHTTPSURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return ""
	}
	return parsed.String()
}

func renderHTML(template EmailTemplate, subject, body string, payload templatePayload) string {
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
	actionURL = safeHTTPSURL(actionURL)
	action := ""
	if actionURL != "" {
		label := "Open in Hackways"
		if template.ID == "admin_promotion" && payload.ActionLabel != "" {
			label = payload.ActionLabel
		}
		if template.ID == "password_reset_requested" {
			label = "Reset password"
		}
		action = `<p style="margin:26px 0 22px;"><a href="` + html.EscapeString(actionURL) + `" style="display:inline-block;padding:11px 18px;background:#201a2b;color:#ffffff;text-decoration:none;font-weight:600;">` + label + `</a></p>`
	}
	if isOTPTemplate(template.ID) && payload.Code != "" {
		action = `<div style="margin:30px 0;text-align:center;"><p style="margin:0 0 12px;font-size:14px;color:#655d70;">Verification code</p><p style="margin:0;padding:22px 12px;background:#f3eef8;border:1px solid #e8e1ed;font-size:38px;line-height:1.3;font-weight:700;letter-spacing:8px;color:#201a2b;">` + html.EscapeString(payload.Code) + `</p></div>`
	}
	banner := ""
	if bannerURL := safeHTTPSURL(payload.EventBannerURL); bannerURL != "" {
		banner = `<img src="` + html.EscapeString(bannerURL) + `" alt="" style="display:block;width:100%;max-height:260px;object-fit:cover;margin:0 0 28px;">`
	}
	unsubscribe := ""
	if template.ID == "admin_promotion" {
		if link := safeHTTPSURL(payload.UnsubscribeURL); link != "" {
			unsubscribe = `<p style="margin:18px 0;font-size:13px;"><a href="` + html.EscapeString(link) + `" style="color:#655d70;text-decoration:underline;">Unsubscribe from promotional emails</a></p>`
		}
		if payload.IsTest {
			unsubscribe += `<p style="font-size:13px;color:#655d70;">Test email requested by an administrator. No subscription was added.</p>`
		}
	}
	byline := ""
	if organizer := strings.TrimSpace(payload.OrganizerName); organizer != "" {
		byline = `<p style="margin:0 0 18px;font-size:14px;color:#766b7f;">Organised by <strong style="color:#3f3648;">` + html.EscapeString(organizer) + `</strong>`
		if space := strings.TrimSpace(payload.SpaceName); space != "" {
			byline += ` in <strong style="color:#3f3648;">` + html.EscapeString(space) + `</strong>`
		}
		byline += `</p>`
	}
	footer := "You are receiving this message because it relates to your Hackways account or registration."
	if template.ID == "admin_promotion" {
		footer = "You are receiving Hackways news and offers because you opted in. Unsubscribe at any time."
		if payload.IsTest {
			footer = "This preview email was requested by an administrator. No subscription was added."
		}
	}
	return `<div style="margin:0;padding:32px 20px;background:#ffffff;color:#201a2b;font-family:Arial,Helvetica,sans-serif;font-size:16px;line-height:1.65;">` +
		`<div style="max-width:600px;margin:0 auto;">` +
		`<p style="margin:0 0 42px;font-size:14px;letter-spacing:.08em;font-weight:700;color:#6d587f;">HACKWAYS</p>` +
		banner +
		`<p style="margin:0 0 10px;font-size:13px;letter-spacing:.04em;text-transform:uppercase;color:#806a91;">` + html.EscapeString(template.Category) + `</p>` +
		`<h1 style="margin:0 0 22px;font-size:28px;line-height:1.2;font-weight:600;color:#201a2b;">` + html.EscapeString(subject) + `</h1>` +
		byline +
		`<p style="margin:0 0 18px;line-height:1.65;color:#3f3648;">` + content.String() + `</p>` +
		action + unsubscribe +
		`<div style="height:1px;background:#e8e1ed;margin:34px 0 18px;"></div>` +
		`<p style="margin:0;font-size:13px;line-height:1.55;color:#766b7f;">` + html.EscapeString(footer) + `</p>` +
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
	case "admin_promotion", "admin_service_message":
		subject = payload.PromotionSubject
		body = payload.PromotionBody
		if subject == "" || body == "" {
			return EmailMessage{}, errors.New("campaign content is missing")
		}
		if payload.IsTest {
			subject = "[Test] " + subject
		}
	case "admin_daily_otp":
		body = template.Copy
	case "email_verification", "login_otp", "email_change_otp", "guest_access_otp":
		body = "Use this code to verify your email. It expires 15 minutes after it was requested. Do not share it."
		if e.TemplateType != "email_verification" {
			body = "Your Hackways one-time code expires 15 minutes after it was requested. Do not share it."
		}
	case "password_reset_requested":
		body = "Use this secure link to reset your Hackways password. It expires 15 minutes after it was requested and can only be used once. If you did not request this, ignore this email."
	case "sign_in_alert":
		device, browser, location := payload.DeviceName, payload.BrowserName, payload.ApproxLocation
		if device == "" {
			device = payload.DeviceLabel
		}
		if browser == "" {
			browser = "Not identified"
		}
		if location == "" {
			location = "Unavailable"
		}
		body = "A new device signed in to your Hackways account.\n\nDevice: " + device + "\nBrowser: " + browser + "\nNear: " + location + "\nTime: " + payload.SignedInAt.UTC().Format(time.RFC1123) + "\n\nLocation is approximate and can differ when using a VPN or mobile network. If this was not you, reset your password and sign out of other sessions from your account."
	case "password_changed":
		body = "Your Hackways password was changed. All existing sessions and remembered devices have been signed out. If this was not you, recover your account immediately."
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
	case "event_invitation":
		body = "You are invited to " + title + ". View the event and register if you would like to attend. This invitation does not reserve a place."
	}
	if title != "your event" && !isOTPTemplate(e.TemplateType) && !strings.Contains(body, title) {
		body = "For " + title + ":\n\n" + body
	}
	if isRegistrationStateTemplate(e.TemplateType) || e.TemplateType == "event_invitation" {
		body = appendTemplateContext(body, payload)
	} else {
		body = appendTemplateContext(body, templatePayload{StartsAt: payload.StartsAt, Location: payload.Location, OrganizerName: payload.OrganizerName, SpaceName: payload.SpaceName, TicketCode: payload.TicketCode})
	}
	htmlBody := renderHTML(template, subject, body, payload)
	if e.TemplateType == "admin_promotion" && safeHTTPSURL(payload.UnsubscribeURL) != "" {
		body += "\n\nUnsubscribe: " + safeHTTPSURL(payload.UnsubscribeURL)
	}
	if isOTPTemplate(e.TemplateType) {
		body += "\n\nVerification code: " + payload.Code
	}
	if e.TemplateType == "password_reset_requested" {
		body += "\n\nReset password: " + safeHTTPSURL(payload.ActionURL)
	}
	senderClass := "event"
	if template.Category == "account" || template.Category == "security" || template.ID == "organizer_welcome" {
		senderClass = "account"
	}
	if template.Category == "promotion" {
		senderClass = "promotion"
	}
	if (e.TemplateType == "admin_promotion" || e.TemplateType == "admin_service_message") && payload.CampaignDesign != nil {
		var err error
		htmlBody, body, err = renderCampaignDesign(subject, payload.PromotionBody, payload.ActionLabel, payload.ActionURL, safeHTTPSURL(payload.UnsubscribeURL), payload.IsTest, *payload.CampaignDesign, e.TemplateType == "admin_promotion")
		if err != nil {
			return EmailMessage{}, err
		}
	}
	return EmailMessage{SenderClass: senderClass, Recipient: e.RecipientEmail, Subject: subject, BodyText: body, BodyHTML: htmlBody, OperationID: uuid(e.ID)}, nil
}

func isOTPTemplate(id string) bool {
	switch id {
	case "admin_daily_otp", "email_verification", "login_otp", "email_change_otp", "guest_access_otp":
		return true
	default:
		return false
	}
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
 if e.TemplateType=="admin_service_message" { active,err:=w.queries.AdminMessageRecipientActive(ctx,e.ID);if err!=nil { return err };if !active { return errSuperseded };return nil }

	if e.TemplateType == "admin_promotion" {
		active, err := w.queries.CampaignMessageActive(ctx, e.ID)
		if err != nil {
			return err
		}
		if !active {
			return errSuperseded
		}
		return nil
	}
	if e.TemplateType == "admin_daily_otp" || e.TemplateType == "email_verification" || e.TemplateType == "password_reset_requested" {
		var payload templatePayload
		if err := json.Unmarshal(e.PayloadJson, &payload); err != nil {
			return err
		}
		hash := payload.VerificationHash
		if hash == "" {
			sum := sha256.Sum256([]byte(payload.Code))
			hash = hex.EncodeToString(sum[:])
		}
		var active bool
		var err error
		if e.TemplateType == "admin_daily_otp" {
			active, err = w.queries.AdminSessionCodeActive(ctx, hash)
		} else if e.TemplateType == "email_verification" {
			active, err = w.queries.VerificationCodeActive(ctx, hash)
		} else {
			active, err = w.queries.PasswordResetTokenActive(ctx, hash)
		}
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

// Reuses the exact production renderer for the admin's saved-draft preview.
func RenderCampaignPreview(subject, body, label, actionURL string, design ...CampaignDesign) (EmailMessage, error) {
	value := templatePayload{PromotionSubject: subject, PromotionBody: body, ActionLabel: label, ActionURL: actionURL}
	if len(design) > 0 {
		value.CampaignDesign = &design[0]
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return EmailMessage{}, err
	}
	return render(db.ClaimEmailOutboxRow{TemplateType: "admin_promotion", PayloadJson: payload})
}

func RenderAdminMessagePreview(subject, body, label, actionURL string, design CampaignDesign) (EmailMessage,error) {
 payload,err:=json.Marshal(templatePayload{PromotionSubject:subject,PromotionBody:body,ActionLabel:label,ActionURL:actionURL,CampaignDesign:&design})
 if err!=nil { return EmailMessage{},err }
 return render(db.ClaimEmailOutboxRow{TemplateType:"admin_service_message",PayloadJson:payload})
}
