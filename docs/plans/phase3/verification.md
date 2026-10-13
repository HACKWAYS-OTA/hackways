# Phase 3 Verification & Evidence Plan: Database & SQL Engineering

## 1. Checks Required Before Implementation
- Confirm Phase 2 exit gate is verified.
- Confirm local PostgreSQL 16 instance is running and accepts connections on port 5432.

---

## 2. Mandatory Verification Protocol

### Check 1: Migration Up and Down Verification
- **Command:**
  ```bash
  migrate -path services/api/migrations -database "$DATABASE_URL" up
  migrate -path services/api/migrations -database "$DATABASE_URL" down 1
  migrate -path services/api/migrations -database "$DATABASE_URL" up
  ```
- **Expected Outcome:** Schema applies, rolls back cleanly, and reapplies without errors.

### Check 2: sqlc Generation Verification
- **Command:** `cd services/api && sqlc generate`
- **Expected Outcome:** Generates Go files with exit code 0.
- **Evidence:** File inspection of `services/api/db/models.go` showing `Rsvp` and `EmailOutbox` structs.

### Check 3: Connection Pool & Health Check Integration
- **Command:** Launch backend with valid `DATABASE_URL` and query:
  ```bash
  curl -s http://localhost:8080/health
  ```
- **Expected Outcome:** Returns JSON payload with HTTP 200 indicating database ping succeeded.

### Check 4: Backend Test Suite Verification
- **Command:** `cd services/api && go test ./... -v -count=1`
- **Expected Outcome:** All unit and database integration tests pass.

---

## 3. Exit Criteria for Phase 3
- [ ] Schema naming harmonized (`rsvps` and `email_outbox`).
- [ ] Migrations apply cleanly and reversibly.
- [ ] `sqlc` generates valid Go files with zero warnings.
- [ ] `pgxpool` active in backend respecting 28-connection cap.
- [ ] Health check verifies database connectivity.
