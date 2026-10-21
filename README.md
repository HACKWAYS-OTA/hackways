# Hackways

Event discovery, free-event registration, and organizer tools.

This private monorepo contains four independent Next.js applications and two Go backend services. Source functionality is not a deployment guarantee: production database/migration recovery and release-readiness checks remain separate gates. Payments, admin moderation, and offline scanning are not delivered.

## Product and service map

| Component | Source | Responsibility |
| --- | --- | --- |
| Public web | [apps/public-web](apps/public-web) | Public website and product entry points |
| Attendee app | [apps/core-web](apps/core-web) | Discovery, registration, account pages, Spaces, and event creation |
| Organizer app | [apps/organizer-web](apps/organizer-web) | Existing-event management, guest review, and check-in interfaces |
| Admin app | [apps/admin-web](apps/admin-web) | Restricted health dashboard; moderation APIs are not implemented |
| Core API | [services/core-backend](services/core-backend) | Events, Spaces, registration inventory, tickets, and owner-scoped workflows |
| Global API | [services/global-backend](services/global-backend) | Authentication, sessions, Azure Blob upload authorization, and email outbox |

Configured public entry points are [hackways.in](https://hackways.in), [app.hackways.in](https://app.hackways.in), and [organizer.hackways.in](https://organizer.hackways.in). These links describe intended routing, not verified live availability.

Go owns business logic and authorization; PostgreSQL owns durable data and capacity constraints. Nginx routes requests to the independent services. Upload bytes go directly to Azure Blob Storage. Email uses the PostgreSQL outbox and ACS; provider acceptance is not proof of inbox delivery. See the [backend implementation guide](docs/backend/README.md) for implemented behavior and explicit gaps.

## Repository layout

- [apps/](apps/) — four web applications.
- [services/](services/) — two independent Go modules.
- [packages/](packages/) — shared API client, domain contracts, and [UI controls](packages/ui/README.md).
- [deploy/migrations/](deploy/migrations/) — ordered database migrations.
- [deploy/](deploy/) — container, ingress, and operational configuration.
- [scripts/](scripts/) — developer verification helpers.
- [docs/](docs/) — architecture, contracts, runbooks, and historical planning.

## Local development

### Prerequisites

- Node.js 22 and npm, matching current CI and frontend Dockerfiles.
- Go 1.23+, GNU Make/Bash for Makefile commands, and Docker Compose v2+.
- The golang-migrate CLI for Makefile migration targets.
- A dedicated local development database. Never point development or fixture tests at production.

### Install and configure

```bash
git clone https://github.com/HACKWAYS-OTA/hackways.git
cd hackways
npm ci
npm run build --workspace @hackways/api-client
cp .env.example .env
```

Repository access is required. Review the template and configure local values privately; do not commit credentials. The root template includes production-oriented settings and is not a ready-to-run local configuration.

Makefile commands and Go services require environment variables such as DATABASE_URL to be exported in the invoking shell. Copying .env alone does not export them. Use a trusted dotenv loader or your shell's environment settings; do not execute an arbitrary dotenv file as shell code. Configure frontend environment files for the local API endpoints.

### Database and servers

```bash
make db-up
make migrate-up
make dev
```

The database target starts the local Compose services. Migrations use the exported DATABASE_URL; confirm its destination before running. On an existing database, reconcile migration history first.

make dev starts the two Go backends and core-web, not all four frontends. Run the remaining applications separately as needed:

```bash
npm run dev --workspace @hackways/public-web
npm run dev --workspace @hackways/organizer-web
npm run dev --workspace @hackways/admin-web
```

Backend defaults are local ports 8081/8082; core-web uses 3001. Do not expose backend or database ports publicly. Production host-scoped secure cookies and CORS require the configured ingress/origins; a local dev server is not a production authentication test.

## Verification commands

| Command | Scope |
| --- | --- |
| make test-api | Both Go modules with the race detector; requires CGO/C toolchain and isolated database setup for database-backed cases |
| make test-web | Existing app type-check and lint scripts, not browser end-to-end tests |
| make lint | Go vet plus frontend lint/type checks |
| npm run type-check | Root workspace type checks where defined |
| npm run lint | Root workspace lint scripts where defined |
| npm run type-check --workspace @hackways/ui | Shared control exports |
| make db-logs | Local Compose logs |
| make db-down | Stop local Compose services |

See [TESTING.md](docs/TESTING.md) for test prerequisites and evidence. make migrate-down is not a general recovery command: some migrations have forward-only guards. Use a reviewed recovery plan for persistent data. make format changes source files and suppresses frontend formatter failures; it is not a verification gate.

## Branches, CI, and releases

- main is the designated production branch; its name does not certify a deployed or production-ready state.
- develop is the integration branch.
- feature/*, fix/*, docs/*, and release/* isolate proposed changes for review.
- Use pull requests rather than direct changes to main/develop. On the current private GitHub Free plan, this is team policy, not enforced branch protection.
- [CI](https://github.com/HACKWAYS-OTA/hackways/actions/workflows/ci.yml) selects affected checks for main/develop pushes and pull requests.
- [Release Candidate](https://github.com/HACKWAYS-OTA/hackways/actions/workflows/release-candidate.yml) runs full checks and publishes six SHA-tagged images only for explicit rc-* tags.
- Application images are built in GitHub Actions, never on the Azure VM. No checked-in workflow automatically deploys to Azure.
- A successful CI/image build does not authorize deployment. Release approval, immutable image identity, migration rehearsal, backup/recovery evidence, and runtime checks are separate requirements.

Keep the historical source repository and release-candidate refs intact. Consult [the organization manager guide](docs/handover/GITHUB_MANAGER_GUIDE.md) and [import/release plan](docs/handover/IMPORT_RELEASE_PLAN_2026-10-21.md) before changing release state.

## Architecture and operating constraints

RSVP inventory transitions use PostgreSQL transactions, event-row locks, a two-second statement timeout, uniqueness constraints, and transactional outbox writes. Frontend checks do not replace backend authorization or database invariants. No production throughput or latency benchmark is claimed.

The six service boundaries, CI-only image builds, bounded resource use, and secret-handling rules are defined in [AGENTS.md](AGENTS.md). Architecture plans describe targets; current code and checked evidence determine implementation status.

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

## Documentation index

| Guide | Use it for |
| --- | --- |
| [Architecture](docs/ARCHITECTURE.md) | Current boundaries and implementation gaps |
| [Backend README](docs/backend/README.md) | Free-event workflows and their limitations |
| [API](docs/API.md) | HTTP contracts |
| [Database](docs/DATABASE.md) | Schema and migration context |
| [Security](docs/SECURITY.md) | Sessions, isolation, and threat model |
| [Deployment](docs/DEPLOYMENT.md) | Operational prerequisites and deployment behavior |
| [Testing](docs/TESTING.md) | Verification requirements |
| [Shared UI](packages/ui/README.md) | Reusable controls and accessibility |
| [Planning index](docs/plans/README.md) | Historical phases and how to reconcile them with current evidence |
| [Contributing](.github/CONTRIBUTING.md) | Contribution conventions |
| [Worklog](docs/planning/WORKLOGS.md) | Dated changes and verification history |

## Proprietary and confidential

Copyright © 2026 Hackways. All rights reserved.
This software, documentation, and associated assets are proprietary and confidential. Unauthorized copying, distribution, modification, public display, or reverse engineering of any portion of this repository is strictly prohibited.
