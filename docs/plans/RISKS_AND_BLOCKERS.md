# Technical Risks, Blockers & Specification Conflicts Register

## 1. Overview

This document maintains an active, prioritized register of technical risks, specification contradictions, missing capabilities, and operational dependencies across all phases of the Hackways engineering plan.

---

## 2. Active Risks & Blockers Register

| Risk ID | Severity | Category | Affected Phase | Summary | Status | Owner Role | Required Resolution |
|:---:|:---:|:---|:---:|:---|:---:|:---|:---|
| **BLK-001** | **CRITICAL** | Specification Conflict | Phase 1, Phase 2 | Directory structure mismatch: Canonical rules mandate 6 decoupled services, but only `apps/web` and `services/api` exist on disk. | `ACTIVE` | Chief Architect | Approve whether to scaffold the 6 folders immediately or execute Phase 1–3 in modular transitional format. |
| **BLK-002** | **CRITICAL** | Rule Violation | Phase 2, Phase 6 | Prohibited package `"lucide-react"` is present in `apps/web/package.json` line 15 and `package-lock.json`. | `ACTIVE` | Frontend Lead | Remove `lucide-react` from `package.json` and lockfiles; regenerate dependencies. |
| **BLK-003** | **HIGH** | Schema Inconsistency | Phase 1, Phase 3 | DDL migration defines `registrations` & `notifications`, while blueprint mandates `rsvps` & `email_outbox`. | `ACTIVE` | Database Lead | Update migration script before running migrations to prevent breaking changes later. |
| **BLK-004** | **HIGH** | Missing Dependency | Phase 2, Phase 3 | `services/api/go.mod` only has Chi v5; `pgx/v5` is missing, and `sqlc` generated Go files do not exist. | `ACTIVE` | Backend Lead | Install `github.com/jackc/pgx/v5` and execute `sqlc generate`. |
| **BLK-005** | **HIGH** | Tooling Deficiency | Phase 7, Phase 8 | Docker Engine is not detected on the Windows PATH; local container startup (`make db-up`) fails. | `ACTIVE` | DevOps Lead | Ensure Docker Desktop / daemon is installed and operational on development workstations. |
| **RSK-001** | **MEDIUM** | Resource Constraint | Phase 8 | Azure `Standard_B2s` VM CPU credit depletion under prolonged high load (throttles to 0.40 vCPU). | `MONITORED`| Infrastructure Lead | Rely on Cloudflare edge caching for >85% reads; document upgrade path to `Standard_D2s_v5`. |
| **RSK-002** | **MEDIUM** | Operational Hazard | Phase 5, Phase 8 | 50 GB NVMe disk exhaustion from unrotated Docker logs or database backups. | `MITIGATED`| Infrastructure Lead | Enforce Docker log rotation (`10m x 3` files = max 210 MB) and ship backups directly to Azure Blob. |
| **RSK-003** | **LOW** | Code Formatting | Phase 1, Phase 2 | `gofmt -l .` flags all 10 Go files on Windows due to CRLF line endings. | `ACTIVE` | Tech Lead | Add `.gitattributes` to enforce `* text=auto eol=lf`. |
| **RSK-004** | **LOW** | Repository Hygiene | Phase 1 | Untracked file `LICENSE` contains unrelated OpenAI Apps SDK guidelines. | `ACTIVE` | Repository Admin | Remove file and update repository proprietary notice. |

---

## 3. Detailed Impact Analysis & Mitigation Strategies

### BLK-001: 6-Service Directory Structure Conflict
- **Impact:** Authoring features in `apps/web` or `services/api` creates significant refactoring debt if the repository must be split into 6 folders (`public-web`, `core-web`, `organizer-web`, `admin-web`, `core-backend`, `global-backend`).
- **Mitigation:** In Phase 1, formally establish the 6-folder directory tree with clean skeletons so foundational development happens in the permanent target locations.

### BLK-002: Prohibited Dependency (`lucide-react`)
- **Impact:** Any automated CI check enforcing `AGENTS.md` Rule 1.2 & Rule 3.12 will fail immediately upon PR creation.
- **Mitigation:** Purge the package from `package.json` and lockfiles in Phase 2 before any frontend component development begins.

### BLK-003: Schema Table Naming Mismatch
- **Impact:** If `000001_initial_schema.up.sql` is run against a database with table names `registrations` and `notifications`, all backend queries, documentation, and agent rules expecting `rsvps` and `email_outbox` will encounter SQL syntax errors.
- **Mitigation:** Harmonize the DDL script in Phase 3 before initializing the local development or test databases.
