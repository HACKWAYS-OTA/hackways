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
| M0-002 | GitHub CI & Issue Templates | DevOps | `verified` | Workflows (`ci.yml`), PR/issue templates |
| M0-003 | Architecture & Engineering Docs | Architect | `verified` | ARCHITECTURE, API, DATABASE, SECURITY, DEPLOYMENT, TESTING in `docs/` |
| M0-004 | Branch Strategy Deployment | DevOps | `verified` | `main` and `develop` created, protected, and pushed to GitHub |

---

## Milestone 1 — Foundations
**Exit condition:** `docker compose up` starts services; Go API server boots and connects to PostgreSQL; Next.js frontend runs cleanly.

| ID | Task | Owner | Status | Acceptance Criteria |
|---|---|---|---|---|
| M1-001 | Go Backend Skeleton (`cmd/api`, `internal/platform`) | Backend Engineer | `verified` | Config, health check, middleware, error envelope, unit tests |
| M1-002 | PostgreSQL Migrations (`000001_initial_schema`) | Data Engineer | `verified` | Extensions, tables, constraints, triggers, GIN search index |
| M1-003 | `sqlc.yaml` and Core DB Queries | Data Engineer | `verified` | sqlc v2 config and queries in `services/api/db/queries/` |
| M1-004 | Next.js 14 App Router Init & Design System | Frontend Engineer | `verified` | App Router, Tailwind tokens, Button, Badge, EventCard |
| M1-005 | Typed API Client Utility | Frontend Engineer | `verified` | `apps/web/lib/api.ts` and `@hackways/api-client` package |

---

## Milestone 2 — Core Event Platform
**Exit condition:** Organizers can create and publish events; public visitors can search and view events from real PostgreSQL data.

| ID | Task | Owner | Status | Acceptance Criteria |
|---|---|---|---|---|
| M2-001 | Event Domain Service & Handlers | Backend Engineer | `planned` | CRUD operations with organizer ownership enforcement |
| M2-002 | Event State Machine | Backend Engineer | `planned` | draft -> published -> unpublished -> cancelled validation |
| M2-003 | Event Full-Text Search Handler | Backend Engineer | `planned` | Search endpoints executing GIN tsvector queries |
| M2-004 | Event Detail & Filter Views | Frontend Engineer | `planned` | Frontend event view with interactive filtering |

---

## Milestone 3 — Concurrency-Safe RSVP & Registrations
**Exit condition:** Attendees can register with atomic capacity locking; notifications are queued and dispatched.
