package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MaxRetries       = 5
	BaseBackoffDelay = 15 * time.Second
	WorkerGoroutines = 2 // Bounded 2-routine worker pool as per AGENTS.md Rule 2.2 (< 10 MB RAM)
	BatchSize        = 10
)

// OutboxWorker coordinates asynchronous dispatch of transactional emails.
type OutboxWorker struct {
	pool         *pgxpool.Pool
	queries      *db.Queries
	sender       EmailSender
	logger       *slog.Logger
	pollInterval time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewOutboxWorker constructs an outbox worker engine.
func NewOutboxWorker(pool *pgxpool.Pool, sender EmailSender, logger *slog.Logger, pollInterval time.Duration) *OutboxWorker {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	return &OutboxWorker{
		pool:         pool,
		queries:      db.New(pool),
		sender:       sender,
		logger:       logger,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
	}
}

// Start launches the outbox processing loop.
func (w *OutboxWorker) Start(ctx context.Context) {
	w.logger.Info("starting transactional email outbox worker", "goroutines", WorkerGoroutines)
	w.wg.Add(1)
	go w.runLoop(ctx)
}

// Stop gracefully signals the outbox worker to terminate and awaits in-flight dispatches.
func (w *OutboxWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	w.logger.Info("transactional email outbox worker stopped cleanly")
}

func (w *OutboxWorker) runLoop(ctx context.Context) {
	defer w.wg.Done()
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Worker semaphore enforcing bounded concurrency
	sem := make(chan struct{}, WorkerGoroutines)

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx, sem)
		}
	}
}

// ProcessNextBatch processes up to BatchSize pending emails.
// Exported for deterministic testing and manual triggers.
func (w *OutboxWorker) ProcessNextBatch(ctx context.Context) (int, error) {
	sem := make(chan struct{}, WorkerGoroutines)
	return w.processBatch(ctx, sem)
}

func (w *OutboxWorker) processBatch(ctx context.Context, sem chan struct{}) (int, error) {
	// 1. Fetch pending rows with FOR UPDATE SKIP LOCKED
	entries, err := w.queries.FetchPendingOutbox(ctx, BatchSize)
	if err != nil {
		w.logger.Error("failed to query email outbox", "error", err)
		return 0, err
	}

	if len(entries) == 0 {
		return 0, nil
	}

	var batchWg sync.WaitGroup
	for _, entry := range entries {
		batchWg.Add(1)
		sem <- struct{}{} // acquire routine slot

		go func(e db.FetchPendingOutboxRow) {
			defer func() {
				<-sem // release routine slot
				batchWg.Done()
			}()

			w.dispatchEntry(ctx, e)
		}(entry)
	}

	batchWg.Wait()
	return len(entries), nil
}

func (w *OutboxWorker) dispatchEntry(ctx context.Context, e db.FetchPendingOutboxRow) {
	msg := EmailMessage{
		Recipient: e.RecipientEmail,
		Subject:   fmt.Sprintf("Hackways Notification: %s", e.TemplateType),
		BodyText:  string(e.PayloadJson),
		BodyHTML:  string(e.PayloadJson),
	}

	err := w.sender.Send(ctx, msg)
	if err != nil {
		w.handleFailure(ctx, e, err)
		return
	}

	// Success: mark sent
	if markErr := w.queries.MarkOutboxSent(ctx, e.ID); markErr != nil {
		w.logger.Error("failed to mark outbox entry as sent", "id", e.ID, "error", markErr)
	} else {
		w.logger.Info("dispatched outbox email successfully", "id", e.ID, "recipient", e.RecipientEmail)
	}
}

func (w *OutboxWorker) handleFailure(ctx context.Context, e db.FetchPendingOutboxRow, sendErr error) {
	newRetryCount := e.RetryCount + 1
	var newStatus db.NotificationStatus = db.NotificationStatusPending

	// Exponential backoff: 15s * 2^retryCount
	backoffMultiplier := 1 << e.RetryCount
	backoffDuration := time.Duration(backoffMultiplier) * BaseBackoffDelay
	nextScheduledAt := time.Now().Add(backoffDuration)

	if newRetryCount >= MaxRetries {
		newStatus = db.NotificationStatusFailed
		w.logger.Warn("outbox email reached max retries; marking failed",
			"id", e.ID,
			"retries", newRetryCount,
			"error", sendErr,
		)
	} else {
		w.logger.Warn("outbox email dispatch failed; scheduling retry",
			"id", e.ID,
			"retry", newRetryCount,
			"next_retry", nextScheduledAt,
			"error", sendErr,
		)
	}

	errText := pgtype.Text{String: sendErr.Error(), Valid: true}
	err := w.queries.MarkOutboxFailed(ctx, db.MarkOutboxFailedParams{
		ID:           e.ID,
		ScheduledAt:  pgtype.Timestamptz{Time: nextScheduledAt, Valid: true},
		Status:       newStatus,
		ErrorMessage: errText,
	})
	if err != nil {
		w.logger.Error("failed to record outbox failure state", "id", e.ID, "error", err)
	}
}
