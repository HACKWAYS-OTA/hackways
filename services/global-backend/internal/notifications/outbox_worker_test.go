package notifications

import (
	"context"
	"errors"
	"github.com/hackways/global-backend/internal/testsupport"
	"io"
	"log/slog"
	"strings"
	"sync"

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
	claimed, err := queries.ClaimEmailOutbox(ctx, BatchSize)
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
}

func TestEmailTemplateCatalog_RendersCleanLayouts(t *testing.T) {
	catalog := EmailTemplateCatalog()
	if len(catalog) != 40 {
		t.Fatalf("expected 40 email templates, got %d", len(catalog))
	}
	seen := make(map[string]bool, len(catalog))
	for _, template := range catalog {
		if seen[template.ID] {
			t.Fatalf("duplicate template id %q", template.ID)
		}
		seen[template.ID] = true
		message, err := render(db.ClaimEmailOutboxRow{
			TemplateType: template.ID,
			PayloadJson:  []byte(`{"event_title":"A & B <launch>","custom_message":"Welcome <back>","location":"Main Hall","action_url":"https://hackways.in/events/example"}`),
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
		if template.TextOnly && message.BodyHTML != "" {
			t.Fatalf("template %q should remain text-only", template.ID)
		}
		if !template.TextOnly && !strings.Contains(message.BodyHTML, "HACKWAYS") {
			t.Fatalf("template %q is missing the email wordmark", template.ID)
		}
	}
}

func TestEmailTemplateCatalog_ReturnsCopy(t *testing.T) {
	catalog := EmailTemplateCatalog()
	catalog[0].Subject = "mutated"
	if EmailTemplateCatalog()[0].Subject == "mutated" {
		t.Fatal("template catalog exposed mutable process-wide state")
	}
}
