# IMPLEMENTATION_PLAN.md — Hackways Engineering Roadmap

**Project:** Hackways  
**Method:** Milestone-gated delivery  
**Status Legend:** `planned` | `in_progress` | `implemented` | `verified` | `accepted`  

---

## Milestone 0 — Repository & Architecture Baseline
**Exit condition:** Git repository initialized with `main` and `develop` branches; remote connected; CI workflows and engineering standards active.

| ID | Task | Owner | Status | Acceptance Criteria |
|---|---|---|---|---|
| M0-001 | Repository Setup & Git Init | Supervisor | `verified` | Git initialized, `.gitignore`, `.env.example`, `README.md` |
| M0-002 | GitHub CI & Issue Templates | DevOps | `verified` | Workflows (`ci.yml`, `codeql.yml`), PR/issue templates |
| M0-003 | Architecture & Engineering Docs | Architect | `verified` | ARCHITECTURE, API, DATABASE, SECURITY, DEPLOYMENT, TESTING in `docs/` |
| M0-004 | Branch Strategy Deployment | DevOps | `in_progress`| `main` and `develop` created and pushed to GitHub |

---

## Milestone 1 — Foundations
**Exit condition:** `docker compose up` starts services; Go API server boots and connects to PostgreSQL; Next.js frontend runs cleanly.

| ID | Task | Owner | Status |
|---|---|---|---|
| M1-001 | Go Backend Skeleton (`cmd/api`, `internal/platform`) | Backend Engineer | `planned` |
| M1-002 | PostgreSQL Migrations (`000001` - `000007`) | Data Engineer | `planned` |
| M1-003 | `sqlc.yaml` and Core DB Queries | Data Engineer | `planned` |
| M1-004 | Next.js 14 App Router Init & Tailwind Design System | Frontend Engineer | `planned` |
| M1-005 | Typed API Client Utility | Frontend Engineer | `planned` |

---

## Milestone 2 — Core Event Platform
**Exit condition:** Organizers can create and publish events; public visitors can search and view events from real PostgreSQL data.

---

## Milestone 3 — Concurrency-Safe RSVP & Registrations
**Exit condition:** Attendees can register with atomic capacity locking; notifications are queued and dispatched.
