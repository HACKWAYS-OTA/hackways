# Engineering Roadmap & Implementation Plan — Hackways

**Status Legend:** `planned` | `in_progress` | `implemented` | `verified` | `accepted`  
**Reference Document:** `docs/ARCHITECTURE_PLAN.md`

---

## Milestone 0 — Repository & Standards Baseline
**Exit condition:** Git repository initialized; GitHub CI workflows and repository standards active.

| ID | Task | Status | Acceptance Criteria |
|---|---|---|---|
| M0-001 | Repository Setup & Git Init | `verified` | Git initialized, `.gitignore`, `.env.example`, `LICENSE` |
| M0-002 | GitHub CI & Community Templates | `verified` | Workflows (`ci.yml`), PR/issue templates in `.github/` |
| M0-003 | Architecture Blueprint | `verified` | 6-service architecture document in `docs/ARCHITECTURE_PLAN.md` |
| M0-004 | Branch Strategy Deployment | `verified` | `main` and `develop` created, protected, and synchronized on GitHub |

---

## Milestone 1 — Production Topology & Ingress
**Exit condition:** Multi-container Docker Compose starts PostgreSQL, Nginx, and both backend skeletons cleanly within resource limits.

| ID | Task | Status | Acceptance Criteria |
|---|---|---|---|
| M1-001 | Docker Compose Production Topology | `planned` | `deploy/docker/docker-compose.prod.yml` with memory caps and health checks |
| M1-002 | Nginx Reverse Proxy & Subdomain Routing | `planned` | Nginx configuration with rate limit zones and admin IP isolation |
| M1-003 | PostgreSQL 16 Tuned Configuration | `planned` | `shared_buffers = 256MB`, `work_mem = 4MB`, `max_connections = 60` |
| M1-004 | Automated Database Backup Script | `planned` | `deploy/scripts/backup-db.sh` with compression and retention |

---

## Milestone 2 — Backend Engine Modularization
**Exit condition:** Core Backend (:8081) and Global API (:8082) running independently with dedicated connection pools.

| ID | Task | Status | Acceptance Criteria |
|---|---|---|---|
| M2-001 | Core Backend Engine (`services/core-backend`) | `planned` | Port 8081, Event aggregate, `pgxpool` max 28, health check |
| M2-002 | Global API Backend (`services/global-backend`) | `planned` | Port 8082, Auth, session store, `pgxpool` max 22, health check |
| M2-003 | Concurrency-Safe RSVP Capacity Engine | `planned` | Transactional `SELECT ... FOR UPDATE` with 2000ms timeout |
| M2-004 | In-Memory Session Cache | `planned` | LRU session cache with 60s TTL to reduce DB queries |

---

## Milestone 3 — Frontend Service Decoupling
**Exit condition:** 4 independent Next.js standalone containers running on ports 3000, 3001, 3002, 3003.

| ID | Task | Status | Acceptance Criteria |
|---|---|---|---|
| M3-001 | Public Web (`apps/public-web`) | `planned` | Port 3000, SSG/ISR marketing, SEO metadata |
| M3-002 | Core Web (`apps/core-web`) | `planned` | Port 3001, Discovery, Event Detail, RSVP flow |
| M3-003 | Organizer Web (`apps/organizer-web`) | `planned` | Port 3002, Event Studio, Attendee roster, CSV exports |
| M3-004 | Admin Web (`apps/admin-web`) | `planned` | Port 3003, Moderation console, audit logs, restricted access |

---

## Milestone 4 — Load Testing & Production Verification
**Exit condition:** Automated k6 load tests pass with zero double-bookings and sub-350ms latency.

| ID | Task | Status | Acceptance Criteria |
|---|---|---|---|
| M4-001 | k6 High-Concurrency RSVP Ticket Drop | `planned` | 1,000 VUs competing for 200 tickets; zero oversold |
| M4-002 | k6 Public Read Cache Storm | `planned` | 2,500 req/sec with Cloudflare/Nginx cache verification |
| M4-003 | Backup & Disaster Recovery Drill | `planned` | Automated restore from pg_dump backup verified |
