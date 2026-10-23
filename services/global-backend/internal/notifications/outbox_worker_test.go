package notifications

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hackways/global-backend/internal/testsupport"
	"image/png"
	"io"
	"log/slog"
	"strings"
	"sync"
	"unicode/utf8"

	"testing"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type mockSender struct {
	mu         sync.Mutex
	shouldFail bool
	calls      int
	messages   []EmailMessage
}

func (m *mockSender) Send(ctx context.Context, msg EmailMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	m.messages = append(m.messages, msg)
	if m.shouldFail {
		return errors.New("network failure to ACS endpoint")
	}
	return nil
}

func TestOutboxWorker_StateTransitions_LivePostgres(t *testing.T) {
	// The existing notification suite also checks custom text is escaped for email HTML.
	rendered, renderErr := render(db.ClaimEmailOutboxRow{TemplateType: "rsvp_confirmed", PayloadJson: []byte(`{"event_title":"CI event","custom_message":"Welcome <back>"}`)})
	if renderErr != nil || !strings.Contains(rendered.BodyText, "Welcome <back>") || strings.Contains(rendered.BodyHTML, "Welcome <back>") || !strings.Contains(rendered.BodyHTML, "Welcome &lt;back&gt;") {
		t.Fatal("custom organizer message rendering is unsafe or missing")
	}
	dbURL := testsupport.DatabaseURL(t)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping live outbox test: %v", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skipping live outbox test: %v", err)
		return
	}

	queries := db.New(pool)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Clean email_outbox for isolated test execution
	_, _ = pool.Exec(ctx, "DELETE FROM email_outbox")

	// 1. Enqueue email for successful dispatch test
	outboxEntry, err := queries.EnqueueEmailOutbox(ctx, db.EnqueueEmailOutboxParams{
		RecipientEmail: "attendee@hackways.in",
		RecipientName:  pgtype.Text{String: "Hackways Attendee", Valid: true},
		TemplateType:   "rsvp_confirmation",
		PayloadJson:    []byte(`{"event":"Hackways Launch 2026"}`),
	})
	if err != nil {
		t.Fatalf("failed to enqueue test outbox email: %v", err)
	}

	sender := &mockSender{shouldFail: false}
	worker := NewOutboxWorker(pool, sender, logger, 1*time.Second)

	// Process batch
	count, err := worker.ProcessNextBatch(ctx)
	if err != nil {
		t.Fatalf("failed to process outbox batch: %v", err)
	}
	if count == 0 {
		t.Errorf("expected at least 1 processed email, got 0")
	}

	// Verify row status is now 'sent'
	var status db.NotificationStatus
	var sentAt *time.Time
	err = pool.QueryRow(ctx, "SELECT status, sent_at FROM email_outbox WHERE id = $1", outboxEntry.ID).Scan(&status, &sentAt)
	if err != nil {
		t.Fatalf("failed to query outbox state: %v", err)
	}
	if status != db.NotificationStatusSent {
		t.Errorf("expected status 'sent', got %v", status)
	}
	if sentAt == nil {
		t.Errorf("expected sent_at timestamp to be populated")
	}
	if len(sender.messages) != 1 || !sender.messages[0].FirstAttempt || sender.messages[0].FirstSentAt.IsZero() || sender.messages[0].OperationID != uuid(outboxEntry.ID) {
		t.Fatal("first provider attempt did not receive its persisted identity")
	}

	// 2. Enqueue email for failure & retry test
	failEntry, err := queries.EnqueueEmailOutbox(ctx, db.EnqueueEmailOutboxParams{
		RecipientEmail: "failing@hackways.in",
		RecipientName:  pgtype.Text{String: "Failing Attendee", Valid: true},
		TemplateType:   "rsvp_confirmation",
		PayloadJson:    []byte(`{"event":"Retry Test"}`),
	})
	if err != nil {
		t.Fatalf("failed to enqueue failing outbox email: %v", err)
	}

	failSender := &mockSender{shouldFail: true}
	failWorker := NewOutboxWorker(pool, failSender, logger, 1*time.Second)

	_, err = failWorker.ProcessNextBatch(ctx)
	if err != nil {
		t.Fatalf("unexpected error during failure processing: %v", err)
	}

	var failStatus db.NotificationStatus
	var retryCount int32
	var scheduledAt time.Time
	err = pool.QueryRow(ctx, "SELECT status, retry_count, scheduled_at FROM email_outbox WHERE id = $1", failEntry.ID).Scan(&failStatus, &retryCount, &scheduledAt)
	if err != nil {
		t.Fatalf("failed to query failed outbox row: %v", err)
	}

	if failStatus != db.NotificationStatusPending {
		t.Errorf("expected status 'pending' after first failure, got %v", failStatus)
	}
	if retryCount != 1 {
		t.Errorf("expected retry_count = 1, got %d", retryCount)
	}
	if scheduledAt.Before(time.Now()) {
		t.Errorf("expected scheduled_at to be in the future, got %v", scheduledAt)
	}
	if len(failSender.messages) != 1 || !failSender.messages[0].FirstAttempt || failSender.messages[0].FirstSentAt.IsZero() {
		t.Fatal("failed first attempt was not recorded before dispatch")
	}
	firstMessage := failSender.messages[0]
	// A regular retry must reconcile, preserving the ID and timestamp even if
	// the original provider call timed out after accepting the message.
	if _, err := pool.Exec(ctx, "UPDATE email_outbox SET scheduled_at=NOW() WHERE id=$1", failEntry.ID); err != nil {
		t.Fatal(err)
	}
	failSender.shouldFail = false
	if _, err := failWorker.ProcessNextBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if len(failSender.messages) != 2 {
		t.Fatal("scheduled retry was not dispatched")
	}
	retried := failSender.messages[1]
	if retried.FirstAttempt || retried.OperationID != firstMessage.OperationID || !retried.FirstSentAt.Equal(firstMessage.FirstSentAt) {
		t.Fatal("retry lost the persisted provider identity or was treated as a new send")
	}
	// Simulate termination after first_sent_at was persisted but before a
	// failure or acceptance was recorded. retry_count remains zero on restart.
	crashed, err := queries.EnqueueEmailOutbox(ctx, db.EnqueueEmailOutboxParams{
		RecipientEmail: "restart@example.invalid", TemplateType: "rsvp_confirmation", PayloadJson: []byte(`{}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := queries.ClaimEmailOutbox(ctx, db.ClaimEmailOutboxParams{BatchLimit: BatchSize})
	if err != nil {
		t.Fatal(err)
	}
	var crashedClaim db.ClaimEmailOutboxRow
	for _, row := range claimed {
		if row.ID == crashed.ID {
			crashedClaim = row
		}
	}
	if !crashedClaim.ID.Valid || crashedClaim.FirstSentAt.Valid || crashedClaim.RetryCount != 0 {
		t.Fatal("unexpected crash fixture claim state")
	}
	persisted, err := queries.BeginOutboxProviderAttempt(ctx, db.BeginOutboxProviderAttemptParams{ID: crashed.ID, LeaseToken: crashedClaim.LeaseToken})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "UPDATE email_outbox SET scheduled_at=NOW() WHERE id=$1", crashed.ID); err != nil {
		t.Fatal(err)
	}
	restartedSender := &mockSender{}
	restartedWorker := NewOutboxWorker(pool, restartedSender, logger, time.Second)
	if _, err := restartedWorker.ProcessNextBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if len(restartedSender.messages) != 1 {
		t.Fatal("reclaimed crash attempt was not dispatched")
	}
	recovered := restartedSender.messages[0]
	if recovered.FirstAttempt || recovered.OperationID != uuid(crashed.ID) || !recovered.FirstSentAt.Equal(persisted.Time) {
		t.Fatal("restart incorrectly created a fresh send from retry_count zero")
	}

	// Exercise the real generated query and worker with isolated PostgreSQL data.
	attendee, err := queries.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("confirmation-ci-%d@example.invalid", time.Now().UnixNano()), PasswordHash: "not-a-login-credential", DisplayName: "CI attendee"})
	if err != nil {
		t.Fatal(err)
	}
	var organizerID, eventID, ticketID pgtype.UUID
	if err = pool.QueryRow(ctx, "INSERT INTO organizers(user_id,display_name,slug) VALUES($1,'CI host',$2) RETURNING id", attendee.ID, fmt.Sprintf("confirmation-ci-%d", time.Now().UnixNano())).Scan(&organizerID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "INSERT INTO events(organizer_id,title,slug,description,status,format,starts_at,ends_at,timezone) VALUES($1,'CI confirmation',$2,'Isolated event','published','in_person',NOW()+INTERVAL '1 day',NOW()+INTERVAL '2 days','Asia/Calcutta') RETURNING id", organizerID, fmt.Sprintf("confirmation-ci-%d", time.Now().UnixNano())).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	registration, err := queries.CreateRSVP(ctx, db.CreateRSVPParams{EventID: eventID, UserID: attendee.ID, GuestName: pgtype.Text{String: "CI registration name", Valid: true}, GuestEmail: pgtype.Text{String: attendee.Email, Valid: true}})
	if err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, "INSERT INTO tickets(rsvp_id) VALUES($1) RETURNING id", registration.ID).Scan(&ticketID); err != nil {
		t.Fatal(err)
	}
	confirmation, err := queries.EnqueueEmailOutbox(ctx, db.EnqueueEmailOutboxParams{RsvpID: registration.ID, EventID: eventID, RecipientEmail: attendee.Email, TemplateType: "rsvp_confirmed", PayloadJson: []byte(`{"custom_message":"See you <soon>"}`)})
	if err != nil {
		t.Fatal(err)
	}
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TICKET_SIGNING_SEED", base64.StdEncoding.EncodeToString(key.Seed()))
	passSender := &mockSender{}
	passWorker := NewOutboxWorker(pool, passSender, logger, time.Second)
	if count, err = passWorker.ProcessNextBatch(ctx); err != nil || count != 1 || len(passSender.messages) != 1 {
		t.Fatalf("confirmation dispatch: count=%d err=%v", count, err)
	}
	if len(passSender.messages[0].Attachments) != 2 || !strings.Contains(passSender.messages[0].BodyHTML, "cid:hackways-entry-qr") || !strings.Contains(passSender.messages[0].BodyHTML, ">My pass</a>") || strings.Contains(passSender.messages[0].BodyHTML, "See you <soon>") {
		t.Fatal("confirmation QR, calendar, action or escaping missing")
	}
	entry := db.ClaimEmailOutboxRow{ID: confirmation.ID, RsvpID: registration.ID, EventID: eventID, RecipientEmail: attendee.Email, TemplateType: "rsvp_confirmed", PayloadJson: []byte(`{}`)}
	for _, mutation := range []string{
		"UPDATE rsvps SET status='cancelled' WHERE id=$1",
		"UPDATE rsvps SET status='confirmed' WHERE id=$1; UPDATE tickets SET status='revoked' WHERE rsvp_id=$1",
	} {
		// pgx simple statements are separate so parameterized extended queries remain valid.
		for _, statement := range strings.Split(mutation, "; ") {
			if _, err = pool.Exec(ctx, statement, registration.ID); err != nil {
				t.Fatal(err)
			}
		}
		if _, err = passWorker.confirmationPass(ctx, entry); !errors.Is(err, errSuperseded) {
			t.Fatal("cancelled or revoked pass was attached")
		}
	}
	if _, err = pool.Exec(ctx, "UPDATE tickets SET status='issued' WHERE rsvp_id=$1", registration.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "UPDATE email_outbox SET recipient_email='another@example.invalid' WHERE id=$1", confirmation.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = passWorker.confirmationPass(ctx, entry); !errors.Is(err, errSuperseded) {
		t.Fatal("pass leaked to a different recipient")
	}
}

func TestEmailTemplateCatalog_RendersCleanLayouts(t *testing.T) {
	catalog := EmailTemplateCatalog()
	if len(catalog) != 43 {
		t.Fatalf("expected 43 email templates, got %d", len(catalog))
	}
	seen := make(map[string]bool, len(catalog))
	for _, template := range catalog {
		if seen[template.ID] {
			t.Fatalf("duplicate template id %q", template.ID)
		}
		seen[template.ID] = true
		message, err := render(db.ClaimEmailOutboxRow{
			TemplateType: template.ID,
			PayloadJson:  []byte(`{"event_title":"A & B <launch>","custom_message":"Welcome <back>","location":"Main Hall","organizer_name":"Ada Organisers","space_name":"Builders Space","event_banner_url":"https://media.hackways.in/banner.jpg","action_url":"https://app.hackways.in/events/example","promotion_subject":"Campaign preview","promotion_body":"Campaign message"}`),
		})
		if err != nil {
			t.Fatalf("template %q did not render: %v", template.ID, err)
		}
		if message.Subject == "" || message.BodyText == "" {
			t.Fatalf("template %q is missing subject or text body", template.ID)
		}
		if strings.Contains(message.BodyHTML, "<table") || strings.Contains(message.BodyHTML, "border-radius") || strings.Contains(message.BodyHTML, "gradient") {
			t.Fatalf("template %q uses a card-like or decorative layout", template.ID)
		}
		if strings.Contains(message.BodyHTML, "A & B <launch>") || strings.Contains(message.BodyHTML, "Welcome <back>") {
			t.Fatalf("template %q did not escape HTML content", template.ID)
		}
		if !strings.Contains(message.BodyHTML, "HACKWAYS") {
			t.Fatalf("template %q is missing the email wordmark", template.ID)
		}
		if !strings.Contains(message.BodyHTML, "Organised by") || !strings.Contains(message.BodyHTML, "Builders Space") {
			t.Fatalf("template %q is missing organiser or space context", template.ID)
		}
		actionLabel := "Open in Hackways"
		if isConfirmation(template.ID) {
			actionLabel = "My pass"
		}
		if template.ID == "password_reset_requested" {
			actionLabel = "Reset password"
		}
		if !strings.Contains(message.BodyHTML, actionLabel) || !strings.Contains(message.BodyHTML, "banner.jpg") {
			t.Fatalf("template %q is missing its safe action link or banner", template.ID)
		}
	}
	t.Run("confirmation attachments", func(t *testing.T) {
		verifyConfirmationAttachments(t)
	})
	design := CampaignDesign{Heading: "A <safe> heading", Accent: "#123456", Blocks: []CampaignBlock{{Kind: "text", Text: "Block <script>"}}}
	preview, err := RenderCampaignPreview("Subject", "Body", "Read", "https://hackways.in", design)
	if err != nil || !strings.Contains(preview.BodyHTML, "#123456") || strings.Contains(preview.BodyHTML, "Block <script>") || !strings.Contains(preview.BodyHTML, "&lt;safe&gt;") {
		t.Fatalf("custom campaign rendering: %v", err)
	}
	design.Blocks = []CampaignBlock{{Kind: "button", Text: "Unsafe", URL: "javascript:alert(1)"}}
	if _, err = RenderCampaignPreview("Subject", "Body", "", "", design); err == nil {
		t.Fatal("unsafe custom block URL accepted")
	}

}

func TestEmailTemplateCatalog_ReturnsCopy(t *testing.T) {
	catalog := EmailTemplateCatalog()
	catalog[0].Subject = "mutated"
	if EmailTemplateCatalog()[0].Subject == "mutated" {
		t.Fatal("template catalog exposed mutable process-wide state")
	}
}

func TestAuthEmailPresentation(t *testing.T) {
	adminOTP, e := render(db.ClaimEmailOutboxRow{TemplateType: "admin_daily_otp", PayloadJson: []byte(`{"verification_code":"012345"}`)})
	if e != nil || adminOTP.SenderClass != "account" || !strings.Contains(adminOTP.BodyHTML, "text-align:center") || !strings.Contains(adminOTP.BodyText, "24 hours") {
		t.Fatal("admin daily email presentation or sender class is incorrect")
	}
	otp, err := render(db.ClaimEmailOutboxRow{TemplateType: "email_verification", PayloadJson: []byte(`{"verification_code":"012345"}`)})
	if err != nil || !strings.Contains(otp.BodyHTML, "text-align:center") || !strings.Contains(otp.BodyHTML, "font-size:38px") || !strings.Contains(otp.BodyHTML, "012345") || !strings.Contains(otp.BodyText, "012345") {
		t.Fatal("verification code presentation missing")
	}
	reset, err := render(db.ClaimEmailOutboxRow{TemplateType: "password_reset_requested", PayloadJson: []byte(`{"action_url":"https://hackways.in/auth/reset-password?key=example"}`)})
	if err != nil || !strings.Contains(reset.BodyHTML, ">Reset password</a>") || strings.Contains(reset.BodyHTML, "Reset password: https:") || !strings.Contains(reset.BodyText, "?key=example") || strings.Contains(strings.ToLower(reset.BodyText), "token") || strings.Contains(reset.BodyHTML, "one-time code") {
		t.Fatal("reset link presentation incorrect")
	}
	alert, err := render(db.ClaimEmailOutboxRow{TemplateType: "sign_in_alert", PayloadJson: []byte(`{"device_name":"Windows PC","browser_name":"Chrome 130","approx_location":"Bengaluru, Karnataka, IN","signed_in_at":"2026-10-22T12:00:00Z"}`)})
	if err != nil || !strings.Contains(alert.BodyText, "Device: Windows PC") || !strings.Contains(alert.BodyText, "Browser: Chrome 130") || !strings.Contains(alert.BodyText, "Near: Bengaluru, Karnataka, IN") || !strings.Contains(alert.BodyText, "approximate") {
		t.Fatal("new-device email context missing")
	}
}

// Extends the existing mandatory notification suite; all credentials below are
// ephemeral verification keys, never application data or production credentials.
func verifyConfirmationAttachments(t *testing.T) {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TICKET_SIGNING_SEED", base64.StdEncoding.EncodeToString(key.Seed()))
	t.Setenv("APP_PUBLIC_URL", "https://hackways.in")
	var ticketID, eventID pgtype.UUID
	_ = ticketID.Scan("8be85aa1-3bc6-48e6-a13e-4310f207a511")
	_ = eventID.Scan("78f9927e-b3bb-49b8-97a3-a26a55ea4bf7")
	now := time.Now().UTC().Truncate(time.Second)
	pass := db.ReadConfirmationPassRow{TicketID: ticketID, EventID: eventID, TicketCode: "T-ci-only", RegistrationCode: "R-ci-only", Title: "CI " + strings.Repeat("Hyderabad "+"\u0c39", 15) + ", event;safe", Slug: "ci-event", Description: "Event\nBEGIN:ATTENDEE\r\nline;comma,", RecipientName: "CI attendee", StartsAt: pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true}, EndsAt: pgtype.Timestamptz{Time: now.Add(2 * time.Hour), Valid: true}, QueuedAt: pgtype.Timestamptz{Time: now, Valid: true}, Timezone: "Asia/Calcutta"}
	entry := db.ClaimEmailOutboxRow{ID: ticketID, TemplateType: "rsvp_confirmed", RecipientEmail: "ci@example.invalid", PayloadJson: []byte(`{"custom_message":"<safe>"}`)}
	worker := &OutboxWorker{}
	message, err := worker.renderConfirmationPass(context.Background(), entry, pass)
	if err != nil || len(message.Attachments) != 2 {
		t.Fatalf("confirmation attachments: %v", err)
	}
	imageBytes, err := base64.StdEncoding.DecodeString(message.Attachments[0].ContentInBase64)
	if err != nil {
		t.Fatal(err)
	}
	imageInfo, err := png.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil || imageInfo.Width != 384 || imageInfo.Height != 384 || message.Attachments[0].ContentID != "hackways-entry-qr" {
		t.Fatal("inline QR is not a valid PNG")
	}
	code, err := confirmationCode(pass)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(code, ".")
	claims, _ := base64.RawURLEncoding.DecodeString(parts[0])
	signature, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var decoded struct {
		TicketID  pgtype.UUID `json:"ticket_id"`
		EventID   pgtype.UUID `json:"event_id"`
		ExpiresAt int64       `json:"expires_at"`
	}
	if !ed25519.Verify(key.Public().(ed25519.PublicKey), claims, signature) || json.Unmarshal(claims, &decoded) != nil || decoded.TicketID != ticketID || decoded.EventID != eventID || decoded.ExpiresAt != pass.EndsAt.Time.Add(24*time.Hour).Unix() {
		t.Fatal("QR wire contract differs from the core scanner")
	}
	calendar, err := base64.StdEncoding.DecodeString(message.Attachments[1].ContentInBase64)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(calendar), "\r\n") {
		if len(line) > 75 || !utf8.ValidString(line) {
			t.Fatal("calendar folding split Unicode or exceeded 75 octets")
		}
	}
	unfolded := strings.ReplaceAll(string(calendar), "\r\n ", "")
	if !strings.Contains(unfolded, "DTSTART:"+pass.StartsAt.Time.UTC().Format("20060102T150405Z")) || !strings.Contains(unfolded, `DESCRIPTION:Event\nBEGIN:ATTENDEE\nline\;comma\,`) || strings.Contains(unfolded, "\r\nBEGIN:ATTENDEE") || !strings.Contains(message.BodyText, "https://hackways.in/my-events/R-ci-only") || strings.Contains(message.BodyText, code) {
		t.Fatal("calendar escaping, date, pass link or privacy failed")
	}
	t.Setenv("TICKET_SIGNING_SEED", "")
	if _, err = worker.renderConfirmationPass(context.Background(), entry, pass); err == nil {
		t.Fatal("missing signing key silently sent a confirmation without QR")
	}
	t.Setenv("TICKET_SIGNING_SEED", base64.StdEncoding.EncodeToString(key.Seed()))

}
