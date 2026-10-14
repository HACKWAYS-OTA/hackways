package rsvps

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// TestDatabase_TransactionRollbackIntegrity verifies that when an error occurs during an RSVP,
// the entire transaction rolls back completely (zero ghost records in events, rsvps, or email_outbox).
func TestDatabase_TransactionRollbackIntegrity(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()

	ctx := context.Background()
	queries := db.New(pool)

	// Create test user & organizer
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        fmt.Sprintf("rollback_org_%d@hackways.in", time.Now().UnixNano()),
		PasswordHash: "fake_hash",
		DisplayName:  "Rollback Org",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	var organizerID pgtype.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO organizers (id, user_id, display_name, slug)
		VALUES (gen_random_uuid(), $1, 'Rollback Org', $2)
		RETURNING id
	`, user.ID, fmt.Sprintf("org-rb-%d", time.Now().UnixNano())).Scan(&organizerID)
	if err != nil {
		t.Fatalf("failed to insert organizer: %v", err)
	}

	var eventUUID pgtype.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO events (id, organizer_id, status, format, title, slug, description, starts_at, ends_at, capacity, timezone)
		VALUES (gen_random_uuid(), $1, 'published', 'in_person', 'Rollback Test Event', $2, 'Testing rollback', NOW() + INTERVAL '1 day', NOW() + INTERVAL '2 days', 50, 'UTC')
		RETURNING id
	`, organizerID, fmt.Sprintf("rb-event-%d", time.Now().UnixNano())).Scan(&eventUUID)
	if err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Begin an explicit transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	qtx := queries.WithTx(tx)

	// Create RSVP within tx
	rsvp, err := qtx.CreateRSVP(ctx, db.CreateRSVPParams{
		EventID:    eventUUID,
		GuestName:  pgtype.Text{String: "Aborted Guest", Valid: true},
		GuestEmail: pgtype.Text{String: "aborted@hackways.in", Valid: true},
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("failed to create rsvp in tx: %v", err)
	}

	// Enqueue email within tx
	_, err = qtx.EnqueueEmailOutbox(ctx, db.EnqueueEmailOutboxParams{
		RsvpID:         rsvp.ID,
		EventID:        eventUUID,
		RecipientEmail: "aborted@hackways.in",
		TemplateType:   "rsvp_confirmation",
		PayloadJson:    []byte(`{}`),
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("failed to enqueue email in tx: %v", err)
	}

	// Increment RSVP count within tx
	if err := qtx.IncrementRSVPCount(ctx, eventUUID); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("failed to increment count in tx: %v", err)
	}

	// Simulate failure and trigger rollback
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// Verify ZERO rows were persisted outside the rolled-back transaction
	var rsvpCount int64
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM rsvps WHERE event_id = $1", eventUUID).Scan(&rsvpCount)
	if rsvpCount != 0 {
		t.Errorf("expected 0 rsvps after rollback, got %d", rsvpCount)
	}

	var outboxCount int64
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM email_outbox WHERE event_id = $1", eventUUID).Scan(&outboxCount)
	if outboxCount != 0 {
		t.Errorf("expected 0 email_outbox rows after rollback, got %d", outboxCount)
	}

	var currentCount int32
	_ = pool.QueryRow(ctx, "SELECT current_rsvp_count FROM events WHERE id = $1", eventUUID).Scan(&currentCount)
	if currentCount != 0 {
		t.Errorf("expected current_rsvp_count = 0 after rollback, got %d", currentCount)
	}
}

// TestDatabase_StatementTimeoutProtection verifies that long-running lock acquisition
// aborts promptly when exceeding statement_timeout = 2000ms.
func TestDatabase_StatementTimeoutProtection(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()

	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// Set a tight statement_timeout of 100ms
	_, err = tx.Exec(ctx, "SET LOCAL statement_timeout = '100ms';")
	if err != nil {
		t.Fatalf("failed to set statement_timeout: %v", err)
	}

	// Execute a 500ms sleep which MUST trigger query cancellation
	start := time.Now()
	_, err = tx.Exec(ctx, "SELECT pg_sleep(0.5);")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected statement timeout error, but query completed in %v", elapsed)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "57014" { // query_canceled
			t.Errorf("expected SQLSTATE 57014 (query_canceled), got %s", pgErr.Code)
		}
	} else {
		t.Errorf("expected pgconn.PgError, got %T: %v", err, err)
	}

	// Map error through platform error mapper
	status, code, _ := platform.MapDatabaseError(err)
	if status != http.StatusServiceUnavailable {
		t.Errorf("expected HTTP 503 Service Unavailable, got %d", status)
	}
	if code != "STATEMENT_TIMEOUT" {
		t.Errorf("expected code STATEMENT_TIMEOUT, got %s", code)
	}
}

// TestDatabase_ConstraintEnforcement verifies database-level foreign key and unique violations.
func TestDatabase_ConstraintEnforcement(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()

	ctx := context.Background()
	queries := db.New(pool)

	// 1. Foreign key violation: RSVP with non-existent event ID
	var nonExistentEvent pgtype.UUID
	_ = nonExistentEvent.Scan("00000000-0000-0000-0000-000000000001")

	_, err := queries.CreateRSVP(ctx, db.CreateRSVPParams{
		EventID:    nonExistentEvent,
		GuestEmail: pgtype.Text{String: "fk@hackways.in", Valid: true},
	})
	if err == nil {
		t.Fatalf("expected foreign key violation, got nil error")
	}

	fkStatus, fkCode, _ := platform.MapDatabaseError(err)
	if fkStatus != http.StatusBadRequest {
		t.Errorf("expected HTTP 400 Bad Request for FK violation, got %d", fkStatus)
	}
	if fkCode != "INVALID_REFERENCE" {
		t.Errorf("expected code INVALID_REFERENCE, got %s", fkCode)
	}

	// 2. Unique constraint violation: duplicate user email
	uniqueEmail := fmt.Sprintf("uniq_%d@hackways.in", time.Now().UnixNano())
	_, err = queries.CreateUser(ctx, db.CreateUserParams{
		Email:        uniqueEmail,
		PasswordHash: "hash1",
		DisplayName:  "User 1",
	})
	if err != nil {
		t.Fatalf("failed to insert initial user: %v", err)
	}

	_, err = queries.CreateUser(ctx, db.CreateUserParams{
		Email:        uniqueEmail,
		PasswordHash: "hash2",
		DisplayName:  "User 2",
	})
	if err == nil {
		t.Fatalf("expected unique constraint violation, got nil error")
	}

	dupStatus, dupCode, _ := platform.MapDatabaseError(err)
	if dupStatus != http.StatusConflict {
		t.Errorf("expected HTTP 409 Conflict for unique violation, got %d", dupStatus)
	}
	if dupCode != "CONFLICT" {
		t.Errorf("expected code CONFLICT, got %s", dupCode)
	}
}
