# Testing Standards & Strategy — Hackways

---

## 1. Quality Gates & Acceptance Requirements

A task or feature cannot be marked verified without passing real automated tests with real commands and output.

```bash
# Go Backend Unit Tests (Mandatory -race)
cd services/api && go test ./... -race -count=1

# Go Backend Integration Tests (Real PostgreSQL)
cd services/api && go test ./... -race -count=1 -run Integration

# Frontend Component Tests
cd apps/web && npm test

# Frontend End-to-End Tests
cd apps/web && npm run test:e2e
```

---

## 2. Test Hierarchy

### Unit Tests
- Co-located with implementation: `<file>_test.go` or `<Component>.test.tsx`.
- Test domain logic, edge case handling, and schema validation.
- Mock external boundaries (email senders, S3 storage), never in-process business logic.

### Integration Tests
- Run against real PostgreSQL 16 test database (via Docker).
- Test database transactions, constraint violation handling, row locking, and sqlc queries.
- Zero mock persistence on integration test paths.

### End-to-End Tests
- Playwright automated browser tests covering critical attendee and organizer user journeys.
