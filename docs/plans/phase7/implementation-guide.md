# Phase 7 Implementation Guide: Testing & CI Enforcement

## 1. Prerequisites
- Phase 6 complete.
- Local PostgreSQL instance running for integration tests.
- k6 CLI installed for load testing.

---

## 2. Step-by-Step Implementation

### Step 1: Implement Concurrency Test (`concurrency_test.go`)
1. Create `services/core-backend/concurrency_test.go`:
   ```go
   package main_test

   import (
       "sync"
       "sync/atomic"
       "testing"
   )

   func TestRSVPConcurrencyCapacity_NoOverbooking(t *testing.T) {
       // Setup event with capacity = 10
       const totalRequests = 100
       const capacity = 10

       var successCount int64
       var conflictCount int64
       var wg sync.WaitGroup

       for i := 0; i < totalRequests; i++ {
           wg.Add(1)
           go func(userID int) {
               defer wg.Done()
               resp := makeRSVPRequest(eventID, userID)
               if resp.StatusCode == 201 {
                   atomic.AddInt64(&successCount, 1)
               } else if resp.StatusCode == 409 {
                   atomic.AddInt64(&conflictCount, 1)
               }
           }(i)
       }
       wg.Wait()

       if successCount != capacity {
           t.Fatalf("expected exactly %d successes, got %d", capacity, successCount)
       }
   }
   ```
2. Run with race detector: `go test -v -race -run TestRSVPConcurrencyCapacity`.

### Step 2: Configure Frontend Component Tests
1. Install Vitest in `apps/core-web`:
   ```bash
   npm install -D vitest @testing-library/react @testing-library/jest-dom jsdom
   ```
2. Author `EventCard.test.tsx` verifying capacity progress bar color transitions.

### Step 3: Author k6 Load Test Script
1. Create `deploy/tests/k6-rsvp-drop.js`.
2. Configure 1,000 VUs competing for 200 tickets.
3. Validate response status distributions and latencies.

### Step 4: Update CI Pipeline
1. Update `.github/workflows/ci.yml` matrix to include both Go engines and all 4 frontends.

---

## 3. What Must NOT Be Changed in Phase 7
- Never disable the Go race detector (`-race`) to make tests run faster or pass flaky tests.
- Never mock the database inside concurrency tests (must run against real PostgreSQL).
- Never report tests as passing without running them.
