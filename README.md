# Hackways

[![CI Pipeline](https://github.com/varshith-dev/HACKWAYS-IN/actions/workflows/ci.yml/badge.svg)](https://github.com/varshith-dev/HACKWAYS-IN/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.23](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://golang.org)
[![Next.js 14](https://img.shields.io/badge/Next.js-14%20App%20Router-black?logo=next.js)](https://nextjs.org)
[![PostgreSQL 16](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)](https://www.postgresql.org)

**Hackways** is a production-grade, highly reliable event discovery, organizer management, and attendee RSVP registration platform. Designed to run on a constrained single-VM production target (**2 vCPU, 4 GB RAM**), Hackways decouples frontend and domain logic into six independent services while enforcing strict transactional invariants (`SELECT ... FOR UPDATE`) to eliminate double-booking and capacity race conditions.

---

## Architecture Overview

Hackways separates concerns into **four independent Next.js web runtimes** and **two high-performance Go backend engines** fronted by an Nginx reverse proxy and Cloudflare Edge CDN:

```
┌────────────────────────────────────────────────────────────────────────┐
│                        Cloudflare Edge CDN / WAF                       │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ HTTPS (Port 443)
                        ┌───────────▼───────────┐
                        │   Nginx Reverse Proxy │ Rate limiting, SSL, Gzip
                        └─────┬───────────┬─────┘
                              │           │
            ┌─────────────────┴─┐       ┌─┴─────────────────┐
            │   Web Frontends   │       │  Backend Engines  │
            │(Next.js Standalone│       │(Go 1.23 + Chi)    │
            └───────┬───────────┘       └─────────┬─────────┘
                    │                             │
    ┌───────────────┼───────────────┐     ┌───────┴───────┐
    ▼               ▼               ▼     ▼               ▼
Public Web      Core Web       Org Web  Core API      Global API
(hackways.in)   (app.hackways) (organizer)(:8081)      (:8082)
    │               │               │     │               │
    └───────────────┴───────┬───────┴─────┘               │
                            ▼                             ▼
                    ┌──────────────┐              ┌───────────────┐
                    │ PostgreSQL 16│              │ Object Storage│
                    │(Max 60 conns)│              │  (Local/R2)   │
                    └──────────────┘              └───────────────┘
```

### The Six Application Services

| Service | Technology | Domain & Routing | Memory Limit |
| :--- | :--- | :--- | :--- |
| **1. Public Web** | Next.js 14 SSG/ISR | `hackways.in` — Marketing landing, SEO, event discovery | 150 MB |
| **2. Core Web** | Next.js 14 App Router | `app.hackways.in` — Attendee event details, tickets, RSVP portal | 260 MB |
| **3. Organizer Web** | Next.js 14 App Router | `organizer.hackways.in` — Event creation studio, roster & analytics | 240 MB |
| **4. Admin Web** | Next.js 14 App Router | `admin.hackways.in` — System audits, moderation (VPN/IP restricted) | 180 MB |
| **5. Core Backend Engine**| Go 1.23 (`chi`, `pgx/v5`) | `core-backend:8081` — Event lifecycle, ticket capacity & RSVP locks | 120 MB |
| **6. Global API Engine** | Go 1.23 (`chi`, `pgx/v5`) | `global-backend:8082` — Auth, sessions, RBAC, background notifications | 110 MB |

> Detailed architectural specifications, memory budgets, and migration strategies are documented in **[docs/ARCHITECTURE_PLAN.md](docs/ARCHITECTURE_PLAN.md)**.

---

## 4 GB RAM Production Resource Budget

| Component | Resident Memory (RSS) | Docker Memory Limit | CPU Quota (2 vCPU = 2.0) |
| :--- | :--- | :--- | :--- |
| **PostgreSQL 16 (Alpine)** | ~220 MB | **700 MB** | 0.60 |
| **Core Backend (Go)** | ~35 MB | **120 MB** | 0.70 |
| **Global Backend (Go)** | ~35 MB | **110 MB** | 0.40 |
| **Core Web (Next.js)** | ~110 MB | **260 MB** | 0.50 |
| **Organizer Web (Next.js)** | ~100 MB | **240 MB** | 0.40 |
| **Admin Web (Next.js)** | ~80 MB | **180 MB** | 0.25 |
| **Public Web (Next.js)** | ~60 MB | **150 MB** | 0.25 |
| **Nginx Ingress** | ~25 MB | **80 MB** | 0.20 |
| **Host Linux OS & System** | ~350 MB | ~500 MB | 0.20 |
| **Ceiling Total** | **~1,215 MB** | **~2,540 MB** | **Safe Headroom: ~1,556 MB** |

---

## Concurrency & RSVP Correctness

Hackways guarantees that capacity limits are never breached during registration spikes:
- **Pessimistic Row-Level Lock:** `SELECT ... FOR UPDATE` acquires an exclusive lock on the event record with `statement_timeout = '2000ms'`.
- **Atomic Serialization:** Lock duration is ~1.5–2.8ms on NVMe storage, cleanly serializing 350–500 reservations per second per hot event.
- **Strict Idempotency:** Duplicate user registrations are prevented via database constraints on `(event_id, user_id)`.

---

## Quick Start (Local Development)

### Prerequisites
- **Docker & Docker Compose** (v2+)
- **Go 1.23+**
- **Node.js 20+** and **npm**

### 1. Clone & Configure Environment
```bash
git clone https://github.com/varshith-dev/HACKWAYS-IN.git
cd HACKWAYS-IN
cp .env.example .env
```

### 2. Start PostgreSQL & Mailpit
```bash
make db-up
```
- PostgreSQL: `localhost:5432`
- Mailpit Web UI: `http://localhost:8025`

### 3. Run Migrations & Start Servers
```bash
make migrate-up
make dev
```

---

## Available Commands (`Makefile`)

| Command | Description |
| :--- | :--- |
| `make dev` | Start database containers and launch development servers |
| `make db-up` | Start local PostgreSQL and Mailpit containers |
| `make db-down` | Stop local database and mail containers |
| `make db-logs` | Stream database container logs |
| `make migrate-up` | Run pending PostgreSQL schema migrations |
| `make migrate-down`| Roll back the latest database migration |
| `make test` | Run all backend and frontend tests |
| `make test-api` | Run Go backend tests with race detection (`-race`) |
| `make test-web` | Run frontend test suites |
| `make lint` | Run Go vet, ESLint, and TypeScript type-checks |
| `make format` | Auto-format Go (`gofmt`) and TypeScript code |

---

## Git Branching Strategy

We enforce a strict Git Flow model. Direct commits to `main` and `develop` are prohibited.

- `main`: Production-ready, stable releases.
- `develop`: Integration branch for active development.
- `feature/*`: New capabilities and enhancements.
- `fix/*`: Bug fixes.
- `release/*`: Release candidate stabilization.
- `hotfix/*`: Critical production hotfixes.

Consult **[.github/CONTRIBUTING.md](.github/CONTRIBUTING.md)** for branch conventions and Pull Request guidelines.

---

## Documentation Index

- **[docs/ARCHITECTURE_PLAN.md](docs/ARCHITECTURE_PLAN.md)** — Production architecture specification & resource allocations
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** — Domain architecture & service contracts
- **[docs/DATABASE.md](docs/DATABASE.md)** — Database schema, indexes, and connection tuning
- **[docs/SECURITY.md](docs/SECURITY.md)** — Threat model, session security, and isolation policies
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** — Production deployment, swap configuration, and backups
- **[docs/TESTING.md](docs/TESTING.md)** — Quality assurance, integration tests, and k6 load tests
- **[docs/planning/](docs/planning/)** — Milestone roadmaps, engineering context, and progress logs

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
