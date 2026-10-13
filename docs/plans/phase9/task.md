# Phase 9 Task Register: Dual Go Backend Engines Decoupling & API Completion

## Tasks Checklist

| Task ID | Title | Prerequisites | Status | Owner | Blockers / References |
|:---:|:---|:---:|:---:|:---|:---|
| `P9-ENG-001` | Scaffold and configure `services/core-backend` (:8081, pool 28) | Phase 8 Exit Gate | `NOT_STARTED` | Backend Lead | `AGENTS.md` Rule 2.1 |
| `P9-ENG-002` | Scaffold and configure `services/global-backend` (:8082, pool 22) | Phase 8 Exit Gate | `NOT_STARTED` | Backend Lead | `AGENTS.md` Rule 2.1 |
| `P9-ENG-003` | Author and test Auth HTTP handlers in `global-backend` | `P9-ENG-002` | `NOT_STARTED` | Security Lead | `/api/v1/auth/*` |
| `P9-ENG-004` | Author and test Storage SAS HTTP handler in `global-backend` | `P9-ENG-002` | `NOT_STARTED` | Backend Lead | `/api/v1/storage/*` |
| `P9-ENG-005` | Wire in-process ACS outbox worker lifecycle in `global-backend` | `P9-ENG-002` | `NOT_STARTED` | Backend Lead | Bounded 2-routine |
| `P9-ENG-006` | Establish `deploy/migrations/` and update multi-backend verification suite | `P9-ENG-001`, `P9-ENG-002` | `NOT_STARTED` | DevOps Lead | `verify.ps1`, `ci.yml` |

---

## Detailed Task Specifications

### `P9-ENG-001`: Core Backend Decoupling
- **Directory:** `services/core-backend/`
- **Module:** `github.com/hackways/core-backend`
- **Port:** 8081
- **DB Connection Limit:** Max 28 connections, Min 4 connections.
- **Packages:** `cmd/server`, `internal/events`, `internal/rsvps`, `internal/platform`, `db/`.
- **Exit Criteria:** `go test ./... -v -race -count=1` passes 100%; server starts on `:8081` with clean `/health`.

### `P9-ENG-002`: Global Backend Decoupling
- **Directory:** `services/global-backend/`
- **Module:** `github.com/hackways/global-backend`
- **Port:** 8082
- **DB Connection Limit:** Max 22 connections, Min 3 connections.
- **Packages:** `cmd/server`, `internal/auth`, `internal/storage`, `internal/notifications`, `internal/platform`, `db/`.
- **Exit Criteria:** `go test ./... -v -race -count=1` passes 100%; server starts on `:8082` with clean `/health`.

### `P9-ENG-003`: Auth HTTP Handlers & Session Management
- **Endpoints:**
  - `POST /api/v1/auth/register`: Takes `{email, password, name, role}`. Validates email, hashes with bcrypt cost 12, creates user and active session, sets `__Host-hackways_session` cookie.
  - `POST /api/v1/auth/login`: Takes `{email, password}`. Verifies hash, issues session, sets cookie.
  - `POST /api/v1/auth/logout`: Clears session from DB, expires cookie.
  - `GET /api/v1/auth/me`: Returns current authenticated user details.
  - `POST /api/v1/auth/verify`: Internal endpoint for other services to validate session cookies.
- **Exit Criteria:** Unit and integration tests verify registration, login, session validity, and invalid credential rejections.

### `P9-ENG-004`: Storage SAS HTTP Handler
- **Endpoint:**
  - `POST /api/v1/storage/upload-intent`: Takes `{fileName, fileSize, mimeType}`. Verifies size <= 5 MB and MIME type against approved image types; returns write-only SAS URL with 15-minute expiration.
- **Exit Criteria:** Negative tests reject oversized files and prohibited MIME types; valid requests produce signed Azure SAS URLs.

### `P9-ENG-005`: Outbox Worker Lifecycle
- **Behavior:**
  - Background goroutine worker starts with `global-backend` server.
  - Polls `email_outbox` table using `FOR UPDATE SKIP LOCKED`.
  - Dispatches via Azure ACS / EmailSender.
  - Handles graceful cancellation on SIGINT/SIGTERM without dropping in-flight dispatches.

### `P9-ENG-006`: Multi-Backend Verification & CI Integration
- Establish canonical schema migrations in `deploy/migrations/`.
- Update `scripts/verify.ps1`, `scripts/verify.sh`, and `.github/workflows/ci.yml` to compile and test both `services/core-backend` and `services/global-backend`.
