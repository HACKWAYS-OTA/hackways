# Phase 9 Implementation Plan: Dual Go Backend Engines Decoupling & API Completion

## 1. Metadata
- **Phase Identifier:** Phase 9 (`phase9`)
- **Title:** Dual Go Backend Engines Decoupling & API Completion
- **Owner Role:** Principal Backend Architect / SRE Lead
- **Current Status:** `IN_PROGRESS`
- **Revision Date:** 2026-10-13

---

## 2. Purpose & Expected Outcome
Phase 9 enforces strict architectural discipline by splitting the single prototype backend into two independent Go engines as specified in [`AGENTS.md` Section 2.1](../../../AGENTS.md#L94-L117):
- **Core Backend (`services/core-backend`)**: Owns event lifecycle and pessimistic RSVP capacity locking on port `8081`.
- **Global Backend (`services/global-backend`)**: Owns auth, sessions, Azure Blob SAS signing, and ACS email delivery on port `8082`.

**Expected Outcome:**
- Two completely decoupled Go modules with dedicated connection pools (28 vs 22 max connections).
- Working HTTP endpoints for auth registration, login, logout, me, session validation, and SAS upload intent.
- In-process ACS transactional outbox worker running in `global-backend` with graceful shutdown.
- All unit, race detector, and live integration tests passing across both engines.

---

## 3. Scope & Non-Goals

### In Scope:
- Scaffolding `services/core-backend` with independent `go.mod` (Go 1.23+), port 8081, pool max 28.
- Scaffolding `services/global-backend` with independent `go.mod` (Go 1.23+), port 8082, pool max 22.
- Establishing canonical database migrations in `deploy/migrations/`.
- Implementing `global-backend` HTTP handlers for Auth (`register`, `login`, `logout`, `me`, `verify`) and Storage (`upload-intent`).
- Integrating the bounded 2-goroutine outbox worker into `global-backend` server lifecycle.
- Updating developer verification tooling (`verify.ps1`, `verify.sh`, `Makefile`) and GitHub Actions CI.

### Explicit Non-Goals:
- Merging the two backend engines into a single binary.
- Introducing unapproved ORMs or runtime SQL translators (GORM, Ent, SQLGlot, pgglot).
- Adding Redis, Kafka, or RabbitMQ for queueing (PostgreSQL transactional outbox is authoritative).
- Mocking database interactions in production endpoints.

---

## 4. Architectural Requirements & Constraints
- **Core Backend Memory Cap:** `GOMEMLIMIT=100MiB`, Docker limit 120 MB.
- **Global Backend Memory Cap:** `GOMEMLIMIT=90MiB`, Docker limit 110 MB.
- **PostgreSQL Connection Budget:** Core Backend: 28 conns; Global Backend: 22 conns; 10 reserved. Total instance connection cap = 60.
- **Pessimistic Locking:** `statement_timeout = '2000ms'` inside `SELECT ... FOR UPDATE` transactions.
- **Cookie Security:** `__Host-hackways_session` (Lax) and `__Host-hackways_admin_session` (Strict).

---

## 5. Ordered Implementation Steps
1. **Step 1 (`P9-ENG-001`):** Scaffold `services/core-backend` directory, `go.mod`, `sqlc.yaml`, `db/`, `internal/platform/`, `internal/events/`, `internal/rsvps/`, and `cmd/server/main.go` on port 8081 with pool limit 28.
2. **Step 2 (`P9-ENG-002`):** Scaffold `services/global-backend` directory, `go.mod`, `sqlc.yaml`, `db/`, `internal/platform/`, `internal/auth/`, `internal/storage/`, `internal/notifications/`, and `cmd/server/main.go` on port 8082 with pool limit 22.
3. **Step 3 (`P9-ENG-003`):** Implement full HTTP handlers for Auth in `global-backend` (`/register`, `/login`, `/logout`, `/me`, `/verify`).
4. **Step 4 (`P9-ENG-004`):** Implement HTTP handler for Azure Blob Storage SAS write intent in `global-backend` (`/upload-intent`).
5. **Step 5 (`P9-ENG-005`):** Wire outbox worker background goroutine into `global-backend` lifecycle with graceful SIGTERM termination.
6. **Step 6 (`P9-ENG-006`):** Establish `deploy/migrations/`, update `scripts/verify.ps1`, `scripts/verify.sh`, and `.github/workflows/ci.yml` to verify both engines with `-race`.
