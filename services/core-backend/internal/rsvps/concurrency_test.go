package rsvps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getTestPool(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://hackways_user:hackways_password@localhost:5432/hackways_dev?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping live postgres test: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skipping live postgres test: %v", err)
	}

	return pool
}

// TestRSVP_CapacityConcurrencyRace validates the core Hackways invariant:
// Under extreme concurrent contention, an event with capacity C will confirm
// EXACTLY C attendees, and all excess requests receive 409 Conflict (zero oversold tickets).
func TestRSVP_CapacityConcurrencyRace(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()

	ctx := context.Background()
	queries := db.New(pool)
	handler := NewHandler(pool)

	// 1. Create a test user and organizer profile
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        fmt.Sprintf("organizer_%d@hackways.in", time.Now().UnixNano()),
		PasswordHash: "fake_hash_for_test",
		DisplayName:  "Test Organizer",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	var organizerID pgtype.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO organizers (id, user_id, display_name, slug)
		VALUES (gen_random_uuid(), $1, 'Test Organizer', $2)
		RETURNING id
	`, user.ID, fmt.Sprintf("org-%d", time.Now().UnixNano())).Scan(&organizerID)
	if err != nil {
		t.Fatalf("failed to insert test organizer: %v", err)
	}

	// 2. Create a published event with capacity = 10
	targetCapacity := int32(10)
	var eventUUID pgtype.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO events (id, organizer_id, status, format, title, slug, description, starts_at, ends_at, capacity, timezone)
		VALUES (gen_random_uuid(), $1, 'published', 'in_person', $2, $3, $4, NOW() + INTERVAL '1 day', NOW() + INTERVAL '2 days', $5, 'UTC')
		RETURNING id
	`, organizerID, fmt.Sprintf("Contention Test Event %d", time.Now().UnixNano()),
		fmt.Sprintf("contention-test-%d", time.Now().UnixNano()),
		"Stress testing pessimistic row locking and zero overbooking invariant",
		targetCapacity,
	).Scan(&eventUUID)
	if err != nil {
		t.Fatalf("failed to insert test event: %v", err)
	}

	// 3. Launch 100 concurrent goroutines competing for the 10 seats
	totalRequests := 100
	var successCount int64
	var conflictCount int64
	var otherCount int64

	var wg sync.WaitGroup
	wg.Add(totalRequests)

	// Start gate so all goroutines trigger simultaneously
	startGate := make(chan struct{})

	for i := 0; i < totalRequests; i++ {
		attendeeNum := i
		go func() {
			defer wg.Done()
			<-startGate // wait for synchronization barrier

			reqBody := CreateRSVPRequest{
				EventID:    fmt.Sprintf("%x-%x-%x-%x-%x", eventUUID.Bytes[0:4], eventUUID.Bytes[4:6], eventUUID.Bytes[6:8], eventUUID.Bytes[8:10], eventUUID.Bytes[10:16]),
				GuestName:  fmt.Sprintf("Attendee %d", attendeeNum),
				GuestEmail: fmt.Sprintf("attendee_%d_%d@hackways.in", time.Now().UnixNano(), attendeeNum),
			}
			jsonBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/rsvps", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateRSVP(w, req)

			switch w.Code {
			case http.StatusCreated:
				atomic.AddInt64(&successCount, 1)
			case http.StatusConflict:
				atomic.AddInt64(&conflictCount, 1)
			default:
				atomic.AddInt64(&otherCount, 1)
			}
		}()
	}

	// Release all 100 goroutines concurrently
	close(startGate)
	wg.Wait()

	// 4. Invariant Assertions
	t.Logf("Concurrency results: Success=%d (expected %d), Conflict=%d (expected %d), Unexpected=%d",
		successCount, targetCapacity, conflictCount, int64(totalRequests)-int64(targetCapacity), otherCount)

	if otherCount > 0 {
		t.Errorf("expected 0 unexpected errors, got %d", otherCount)
	}

	if successCount != int64(targetCapacity) {
		t.Fatalf("INVARIANT VIOLATION: expected exactly %d confirmed RSVPs, but confirmed %d! Overbooking detected!",
			targetCapacity, successCount)
	}

	expectedConflicts := int64(totalRequests) - int64(targetCapacity)
	if conflictCount != expectedConflicts {
		t.Errorf("expected %d conflict responses, got %d", expectedConflicts, conflictCount)
	}

	// 5. Authoritative Database Verification
	var currentRsvpCount int32
	err = pool.QueryRow(ctx, "SELECT current_rsvp_count FROM events WHERE id = $1", eventUUID).Scan(&currentRsvpCount)
	if err != nil {
		t.Fatalf("failed to query event count from db: %v", err)
	}
	if currentRsvpCount != targetCapacity {
		t.Errorf("db current_rsvp_count = %d, expected %d", currentRsvpCount, targetCapacity)
	}

	var totalRsvpRows int64
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM rsvps WHERE event_id = $1", eventUUID).Scan(&totalRsvpRows)
	if err != nil {
		t.Fatalf("failed to query total rsvps rows: %v", err)
	}
	if totalRsvpRows != int64(targetCapacity) {
		t.Errorf("total rsvp rows in db = %d, expected %d", totalRsvpRows, targetCapacity)
	}

	var totalOutboxRows int64
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM email_outbox WHERE event_id = $1", eventUUID).Scan(&totalOutboxRows)
	if err != nil {
		t.Fatalf("failed to query total email_outbox rows: %v", err)
	}
	if totalOutboxRows != int64(targetCapacity) {
		t.Errorf("total email outbox rows = %d, expected %d", totalOutboxRows, targetCapacity)
	}
}

// TestRSVP_IdempotentDuplicateRegistration tests double registration protection.
func TestRSVP_IdempotentDuplicateRegistration(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()

	ctx := context.Background()
	queries := db.New(pool)
	handler := NewHandler(pool)

	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        fmt.Sprintf("organizer_dup_%d@hackways.in", time.Now().UnixNano()),
		PasswordHash: "fake_hash",
		DisplayName:  "Dup Organizer",
	})
	if err != nil {
		t.Fatalf("failed to create organizer user: %v", err)
	}

	var organizerID pgtype.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO organizers (id, user_id, display_name, slug)
		VALUES (gen_random_uuid(), $1, 'Dup Organizer', $2)
		RETURNING id
	`, user.ID, fmt.Sprintf("org-dup-%d", time.Now().UnixNano())).Scan(&organizerID)
	if err != nil {
		t.Fatalf("failed to insert organizer: %v", err)
	}

	var eventUUID pgtype.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO events (id, organizer_id, status, format, title, slug, description, starts_at, ends_at, capacity, timezone)
		VALUES (gen_random_uuid(), $1, 'published', 'online', 'Idempotency Test Event', $2, 'Testing idempotent duplicate registration', NOW() + INTERVAL '1 day', NOW() + INTERVAL '2 days', 50, 'UTC')
		RETURNING id
	`, organizerID, fmt.Sprintf("idempotent-test-%d", time.Now().UnixNano())).Scan(&eventUUID)
	if err != nil {
		t.Fatalf("failed to insert test event: %v", err)
	}

	eventIDStr := fmt.Sprintf("%x-%x-%x-%x-%x", eventUUID.Bytes[0:4], eventUUID.Bytes[4:6], eventUUID.Bytes[6:8], eventUUID.Bytes[8:10], eventUUID.Bytes[10:16])

	// Create attendee user
	attendeeUser, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        fmt.Sprintf("dup_attendee_%d@hackways.in", time.Now().UnixNano()),
		PasswordHash: "fake_hash",
		DisplayName:  "Duplicate Attendee",
	})
	if err != nil {
		t.Fatalf("failed to create attendee user: %v", err)
	}

	reqBody := CreateRSVPRequest{
		EventID:    eventIDStr,
		GuestEmail: attendeeUser.Email,
		GuestName:  attendeeUser.DisplayName,
	}
	jsonBytes, _ := json.Marshal(reqBody)

	// First RSVP: Must succeed (201 Created)
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/rsvps", bytes.NewReader(jsonBytes))
	w1 := httptest.NewRecorder()
	handler.CreateRSVP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first registration failed with status %d: %s", w1.Code, w1.Body.String())
	}

	// Second RSVP from same identity: Must be rejected with 409 Conflict
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/rsvps", bytes.NewReader(jsonBytes))
	w2 := httptest.NewRecorder()
	handler.CreateRSVP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected second registration to fail with 409 Conflict, got %d", w2.Code)
	}

	// Verify current_rsvp_count did NOT increment twice
	var count int32
	_ = pool.QueryRow(ctx, "SELECT current_rsvp_count FROM events WHERE id = $1", eventUUID).Scan(&count)
	if count != 1 {
		t.Errorf("current_rsvp_count = %d, expected 1 after duplicate registration", count)
	}
}
