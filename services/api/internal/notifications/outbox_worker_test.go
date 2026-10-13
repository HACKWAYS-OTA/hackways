package notifications

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/hackways/api/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type mockSender struct {
	shouldFail bool
	calls      int
}

func (m *mockSender) Send(ctx context.Context, msg EmailMessage) error {
	m.calls++
	if m.shouldFail {
		return errors.New("network failure to ACS endpoint")
	}
	return nil
}

func TestOutboxWorker_StateTransitions_LivePostgres(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://hackways_user:hackways_password@localhost:5432/hackways_dev?sslmode=disable"
	}

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
}
