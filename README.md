# Hackways

Organization setup: [HACKWAYS-OTA GitHub manager guide](docs/handover/GITHUB_MANAGER_GUIDE.md), with proposed ownership, setup backlog and a sanitized environment inventory. It distinguishes the existing remote repository from local changes and unverified production deployment state.

[![CI Pipeline](https://github.com/varshith-dev/HACKWAYS-IN/actions/workflows/ci.yml/badge.svg)](https://github.com/varshith-dev/HACKWAYS-IN/actions/workflows/ci.yml)
[![Go 1.23](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://golang.org)
[![Next.js 14](https://img.shields.io/badge/Next.js-14%20App%20Router-black?logo=next.js)](https://nextjs.org)
[![PostgreSQL 16](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)](https://www.postgresql.org)

**Hackways is an in-development event discovery and RSVP platform.** The repository has four Next.js applications and two Go backend services. Free-event management, registration forms, ownership-checked uploads and ACS sending are implemented in the working source, but production readiness is not established. Current database authentication, reviewed release identity, database backup/restore, migration rehearsal and deployment ownership remain blockers. Admin moderation, payments and offline scanning are not delivered. See [the implementation audit](docs/ARCHITECTURE.md), [current backend behavior](docs/backend/README.md) and [deployment evidence](docs/handover/DEPLOYMENT_PREFLIGHT_EVIDENCE_2026-10-20.md).

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
                    │(Max 60 conns)│              │  (Azure Blob)   │
                    └──────────────┘              └───────────────┘
```

### Deployment Boundaries and Implemented Scope

The six deployable directories are present, but the product capabilities in the architecture plan are not all implemented. The route inventory in [docs/API.md](docs/API.md) is the source of truth for current backend functionality.

| Service | Current implementation | Production memory limit |
| :--- | :--- | :---: |
| Public Web | Marketing app exists | 140 MB |
| Core Web | Attendee app exists; RSVP UI is present | 250 MB |
| Organizer Web | App scaffold; organizer workflows are not implemented | 230 MB |
| Admin Web | App scaffold; admin workflows are not implemented | 170 MB |
| Core Backend | Event reads, RSVP create, owner-scoped RSVP detail | 120 MB |
| Global Backend | Auth, upload SAS intent, outbox worker; ACS delivery unavailable | 110 MB |

> The architecture plan is a target blueprint, not proof of deployed behavior. See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for checked-in deployment behavior and known configuration hazards.

---

## 4 GB RAM Resource Budget (Target)

The following is the canonical target budget. It is not a measurement of current production usage, and it does not exactly match the checked-in Compose limits; see [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

| Component | Resident Memory (RSS) | Docker Memory Limit | CPU Quota (2 vCPU = 2.0) |
| :--- | :--- | :--- | :--- |
| **PostgreSQL 16 (Alpine)** | ~220 MB | **700 MB** | 0.60 |
| **Core Backend (Go)** | ~35 MB | **120 MB** | 0.70 |
| **Global Backend (Go)** | ~35 MB | **110 MB** | 0.40 |
| **Core Web (Next.js)** | ~110 MB | **250 MB** | 0.50 |
| **Organizer Web (Next.js)** | ~100 MB | **230 MB** | 0.40 |
| **Admin Web (Next.js)** | ~80 MB | **170 MB** | 0.25 |
| **Public Web (Next.js)** | ~60 MB | **140 MB** | 0.25 |
| **Nginx Ingress** | ~25 MB | **75 MB** | 0.20 |
| **Host Linux OS & System** | ~350 MB | ~500 MB | 0.20 |
| **Calculated sum of listed limits** |  | **2,295 MB** | **~3.5 CPU quota** |

The modeled total includes 500 MB of host/system allowance. PostgreSQL placement is unresolved, and configured Compose container limits differ from this target; see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md#service-boundaries).

---

## Concurrency & RSVP Correctness

RSVP creation uses a PostgreSQL transaction, a 2-second statement timeout, an event row lock, unique constraints, and an outbox insert before commit. The repository has an integration test for concurrent capacity; no production throughput or latency benchmark has been measured, so this README makes no capacity claims.

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

### 2. Start PostgreSQL
```bash
make db-up
```
- PostgreSQL: `localhost:5432`

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
| `make db-up` | Start the local PostgreSQL container |
| `make db-down` | Stop the local database container |
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

## Proprietary & Confidential

Copyright © 2026 Hackways. All rights reserved.  
This software, documentation, and associated assets are proprietary and confidential. Unauthorized copying, distribution, modification, public display, or reverse engineering of any portion of this repository is strictly prohibited.
