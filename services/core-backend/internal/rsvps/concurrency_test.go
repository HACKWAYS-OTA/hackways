package rsvps

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hackways/core-backend/internal/testsupport"
	"net/http"
	"net/http/httptest"

	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getTestPool(t *testing.T) *pgxpool.Pool {
	dbURL := testsupport.DatabaseURL(t)

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

// These checks extend the existing mandatory database-backed CI suite. Their
// fixtures must only run with an explicitly isolated DATABASE_URL.
func workflowFixture(t *testing.T, mode string, waitlist bool) (*pgxpool.Pool, *workflows.Engine, db.Events, db.Organizers, []db.CreateUserRow) {
	t.Helper()
	pool := getTestPool(t)
	t.Cleanup(pool.Close)
	ctx := context.Background()
	q := db.New(pool)
	users := make([]db.CreateUserRow, 3)
	for i := range users {
		user, err := q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("workflow-%d-%d@example.invalid", time.Now().UnixNano(), i), DisplayName: "Isolated workflow account", PasswordHash: "not-a-login-credential"})
		if err != nil {
			t.Fatal(err)
		}
		users[i] = user
	}
	organizer, err := q.CreateOrganizer(ctx, db.CreateOrganizerParams{UserID: users[0].ID, DisplayName: "Isolated workflow owner"})
	if err != nil {
		t.Fatal(err)
	}
	event, err := q.CreateEventRecord(ctx, db.CreateEventRecordParams{OrganizerID: organizer.ID, Status: db.EventStatusPublished, Format: db.EventFormatInPerson, Title: "Isolated workflow verification", Description: "Mandatory CI fixture, outside application data", StartsAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}, EndsAt: pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true}, Timezone: "UTC", Capacity: pgtype.Int4{Int32: 1, Valid: true}, LocationName: pgtype.Text{String: "CI only", Valid: true}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "UPDATE events SET registration_mode=$2,waitlist_enabled=$3 WHERE id=$1", event.ID, mode, waitlist); err != nil {
		t.Fatal(err)
	}
	return pool, workflows.New(pool), event, organizer, users
}
func workflowRequest(event db.Events, user db.CreateUserRow, key string, waitlist bool) workflows.Request {
	return workflows.Request{EventID: event.ID, UserID: user.ID, Email: user.Email, Name: user.DisplayName, Key: key, Waitlist: waitlist, RequestID: "workflow-ci"}
}

func TestWorkflow_WaitlistCancellationAcceptanceAndReplay(t *testing.T) {
	pool, engine, event, _, users := workflowFixture(t, "instant", true)
	ctx := context.Background()
	q := db.New(pool)
	firstRequest := workflowRequest(event, users[1], "first-request-key-0001", false)
	first, _, err := engine.Register(ctx, firstRequest)
	if err != nil || first.Data.Status != db.RegistrationStatusConfirmed {
		t.Fatalf("first registration: %v", err)
	}
	waiting, _, err := engine.Register(ctx, workflowRequest(event, users[2], "waitlist-request-key-0002", true))
	if err != nil || waiting.Data.Status != db.RegistrationStatusWaitlisted {
		t.Fatalf("waitlist: %v", err)
	}
	if _, err = engine.Change(ctx, first.Data.ID, users[1].ID, pgtype.UUID{}, "cancel", "workflow-ci"); err != nil {
		t.Fatal(err)
	}
	offered, err := q.GetWorkflowRSVP(ctx, waiting.Data.ID)
	if err != nil || offered.Status != db.RegistrationStatusOffered {
		t.Fatalf("offer not persisted: %v", err)
	}
	inventory, err := q.GetEventByID(ctx, event.ID)
	if err != nil || inventory.CurrentRsvpCount != 0 || inventory.HeldRsvpCount != 1 {
		t.Fatalf("offer inventory incorrect: %+v %v", inventory, err)
	}
	accepted, err := engine.Change(ctx, offered.ID, users[2].ID, pgtype.UUID{}, "accept", "workflow-ci")
	if err != nil || accepted.Status != db.RegistrationStatusConfirmed {
		t.Fatalf("accept: %v", err)
	}
	inventory, err = q.GetEventByID(ctx, event.ID)
	if err != nil || inventory.CurrentRsvpCount != 1 || inventory.HeldRsvpCount != 0 {
		t.Fatalf("confirmed inventory incorrect: %+v %v", inventory, err)
	}
	replay, replayed, err := engine.Register(ctx, firstRequest)
	if err != nil || !replayed || replay.Data.Status != db.RegistrationStatusCancelled {
		t.Fatalf("replay must reflect current state: %+v %v", replay, err)
	}
	firstRequest.Notes = "changed payload"
	if _, _, err = engine.Register(ctx, firstRequest); err == nil {
		t.Fatal("key reuse with changed payload succeeded")
	}
	revoked, err := q.GetTicketByRSVP(ctx, first.Data.ID)
	if err != nil || revoked.Status != "revoked" {
		t.Fatalf("cancelled ticket was not revoked: %v", err)
	}
}
func TestWorkflow_GuestCannotResetAccountRegistration(t *testing.T) {
	for _, status := range []db.RegistrationStatus{db.RegistrationStatusCancelled, db.RegistrationStatusExpired} {
		t.Run(string(status), func(t *testing.T) {
			pool, engine, event, _, users := workflowFixture(t, "instant", false)
			ctx := context.Background()
			q := db.New(pool)
			accountRequest := workflowRequest(event, users[1], "account-ownership-key-0001", false)
			accountRequest.Notes = "Private account note"
			registered, _, err := engine.Register(ctx, accountRequest)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = engine.Change(ctx, registered.Data.ID, users[1].ID, pgtype.UUID{}, "cancel", "ownership-ci"); err != nil {
				t.Fatal(err)
			}
			if status == db.RegistrationStatusExpired {
				if _, err = pool.Exec(ctx, "UPDATE rsvps SET status='expired' WHERE id=$1", registered.Data.ID); err != nil {
					t.Fatal(err)
				}
			}
			before, err := q.GetWorkflowRSVP(ctx, registered.Data.ID)
			if err != nil {
				t.Fatal(err)
			}
			ticketBefore, err := q.GetTicketByRSVP(ctx, before.ID)
			if err != nil {
				t.Fatal(err)
			}
			answersBefore, err := q.LatestRegistrationAnswers(ctx, before.ID)
			if err != nil {
				t.Fatal(err)
			}
			eventBefore, err := q.GetEventByID(ctx, event.ID)
			if err != nil {
				t.Fatal(err)
			}
			var submissionsBefore, outboxBefore, auditBefore int64
			if err = pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM registration_submissions WHERE rsvp_id=$1),(SELECT count(*) FROM email_outbox WHERE rsvp_id=$1),(SELECT count(*) FROM audit_log WHERE entity_id=$1)", before.ID).Scan(&submissionsBefore, &outboxBefore, &auditBefore); err != nil {
				t.Fatal(err)
			}
			guest := workflows.Request{EventID: event.ID, Email: users[1].Email, Name: "Unauthenticated caller", Notes: "Unauthorized replacement", Key: "guest-ownership-key-00002", RequestID: "ownership-ci"}
			if _, _, err = engine.Register(ctx, guest); err == nil {
				t.Fatal("guest reset an account-owned registration")
			} else {
				var failure *workflows.Failure
				if !errors.As(err, &failure) || failure.Status != 401 || failure.Code != "ACCOUNT_REQUIRED" {
					t.Fatalf("expected account ownership rejection, got %v", err)
				}
			}
			after, err := q.GetWorkflowRSVP(ctx, before.ID)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatalf("registration changed: %v", err)
			}
			ticketAfter, err := q.GetTicketByRSVP(ctx, before.ID)
			if err != nil || !reflect.DeepEqual(ticketBefore, ticketAfter) {
				t.Fatalf("ticket changed: %v", err)
			}
			answersAfter, err := q.LatestRegistrationAnswers(ctx, before.ID)
			if err != nil || !reflect.DeepEqual(answersBefore, answersAfter) {
				t.Fatalf("answers changed: %v", err)
			}
			eventAfter, err := q.GetEventByID(ctx, event.ID)
			if err != nil || !reflect.DeepEqual(eventBefore, eventAfter) {
				t.Fatalf("event/inventory changed: %v", err)
			}
			var submissionsAfter, outboxAfter, auditAfter int64
			if err = pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM registration_submissions WHERE rsvp_id=$1),(SELECT count(*) FROM email_outbox WHERE rsvp_id=$1),(SELECT count(*) FROM audit_log WHERE entity_id=$1)", before.ID).Scan(&submissionsAfter, &outboxAfter, &auditAfter); err != nil {
				t.Fatal(err)
			}
			if submissionsBefore != submissionsAfter || outboxBefore != outboxAfter || auditBefore != auditAfter {
				t.Fatal("rejected request persisted side effects")
			}
			// The authenticated owner can still re-register through the same path.
			accountRequest.Key = "account-reregister-key-0002"
			if _, _, err = engine.Register(ctx, accountRequest); err != nil {
				t.Fatalf("owner re-registration failed: %v", err)
			}
		})
	}
}

func TestWorkflow_ApprovalAllocatesOnlyOnDecision(t *testing.T) {
	pool, engine, event, organizer, users := workflowFixture(t, "approval", false)
	ctx := context.Background()
	q := db.New(pool)
	first, _, err := engine.Register(ctx, workflowRequest(event, users[1], "approval-request-key-0001", false))
	if err != nil || first.Data.Status != db.RegistrationStatusPendingApproval {
		t.Fatalf("application: %v", err)
	}
	second, _, err := engine.Register(ctx, workflowRequest(event, users[2], "approval-request-key-0002", false))
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := q.GetEventByID(ctx, event.ID)
	if err != nil || inventory.CurrentRsvpCount != 0 || inventory.HeldRsvpCount != 0 {
		t.Fatal("pending applications consumed capacity")
	}
	if _, err = engine.Change(ctx, first.Data.ID, users[0].ID, organizer.ID, "approve", "workflow-ci"); err != nil {
		t.Fatal(err)
	}
	if _, err = engine.Change(ctx, second.Data.ID, users[0].ID, organizer.ID, "approve", "workflow-ci"); err == nil {
		t.Fatal("approval oversold capacity")
	}
	if _, err = engine.Change(ctx, second.Data.ID, users[1].ID, users[1].ID, "reject", "workflow-ci"); err == nil {
		t.Fatal("nonowner rejected application")
	}
	if _, err = engine.Change(ctx, second.Data.ID, users[0].ID, organizer.ID, "reject", "workflow-ci"); err != nil {
		t.Fatal(err)
	}
}
func TestWorkflow_ExpiredOfferAndDuplicateCheckIn(t *testing.T) {
	pool, engine, event, organizer, users := workflowFixture(t, "instant", true)
	ctx := context.Background()
	q := db.New(pool)
	first, _, err := engine.Register(ctx, workflowRequest(event, users[1], "expiry-request-key-0001", false))
	if err != nil {
		t.Fatal(err)
	}
	waiting, _, err := engine.Register(ctx, workflowRequest(event, users[2], "expiry-request-key-0002", true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = engine.Change(ctx, first.Data.ID, users[1].ID, pgtype.UUID{}, "cancel", "workflow-ci"); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, "UPDATE rsvps SET offer_expires_at=NOW()-INTERVAL '1 minute' WHERE id=$1", waiting.Data.ID); err != nil {
		t.Fatal(err)
	}
	tx, qtx, current, err := engine.Begin(ctx, event.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err = engine.ExpireAndPromote(ctx, qtx, &current); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	expired, err := q.GetWorkflowRSVP(ctx, waiting.Data.ID)
	if err != nil || expired.Status != db.RegistrationStatusExpired {
		t.Fatalf("offer did not expire: %v", err)
	}
	inventory, err := q.GetEventByID(ctx, event.ID)
	if err != nil || inventory.HeldRsvpCount != 0 || inventory.CurrentRsvpCount != 0 {
		t.Fatal("expired offer retained inventory")
	}
	if _, err = engine.Change(ctx, expired.ID, users[2].ID, pgtype.UUID{}, "accept", "workflow-ci"); err == nil {
		t.Fatal("expired offer accepted")
	}
	restored, _, err := engine.Register(ctx, workflowRequest(event, users[1], "reregister-request-key-0003", false))
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := q.GetTicketByRSVP(ctx, restored.Data.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = q.CheckInWorkflowTicket(ctx, db.CheckInWorkflowTicketParams{TicketID: ticket.ID, EventID: event.ID, ActorID: organizer.UserID}); err != nil {
		t.Fatal(err)
	}
	if _, err = q.CheckInWorkflowTicket(ctx, db.CheckInWorkflowTicketParams{TicketID: ticket.ID, EventID: event.ID, ActorID: organizer.UserID}); err == nil {
		t.Fatal("duplicate check-in succeeded")
	}
	if _, err = engine.Change(ctx, restored.Data.ID, users[1].ID, pgtype.UUID{}, "cancel", "workflow-ci"); err == nil {
		t.Fatal("checked-in registration cancelled")
	}
}
