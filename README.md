# Hackways

[![CI Pipeline](https://github.com/varshith-dev/HACKWAYS-IN/actions/workflows/ci.yml/badge.svg)](https://github.com/varshith-dev/HACKWAYS-IN/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hackways/api)](https://goreportcard.com/report/github.com/hackways/api)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.23](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://golang.org)
[![Next.js 14](https://img.shields.io/badge/Next.js-14%20App%20Router-black?logo=next.js)](https://nextjs.org)
[![PostgreSQL 16](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)](https://www.postgresql.org)

**Hackways** is a production-grade, highly reliable event discovery, event management, and RSVP registration platform. Built as a modular monolith in Go with a responsive Next.js frontend, Hackways eliminates concurrency issues, race-condition overbooking, and fake persistence through strict type safety and transaction-level guarantees.

---

## Table of Contents

- [System Architecture](#system-architecture)
- [Technology Stack](#technology-stack)
- [Repository Structure](#repository-structure)
- [Prerequisites](#prerequisites)
- [Local Setup & Quick Start](#local-setup--quick-start)
- [Environment Configuration](#environment-configuration)
- [Git Branching Strategy](#git-branching-strategy)
- [Development Workflow](#development-workflow)
- [Testing & Quality Assurance](#testing--quality-assurance)
- [CI/CD Automation](#cicd-automation)
- [Deployment Strategy](#deployment-strategy)
- [Documentation Index](#documentation-index)
- [Contributing](#contributing)
- [Security](#security)

---

## System Architecture

Hackways adopts a **modular monolith** architecture: a single, high-performance Go application (`services/api`) backed by PostgreSQL 16, coupled with a Next.js 14 App Router frontend (`apps/web`).

```
┌────────────────────────────────────────────────────────────────────────┐
│                              Internet / CDN                            │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ HTTPS
                        ┌───────────▼───────────┐
                        │    Next.js Web App    │  apps/web/
                        │   (App Router, SSR)   │  Port 3000
                        └───────────┬───────────┘
                                    │ REST /api/v1 (JSON)
                        ┌───────────▼───────────┐
                        │     Go API Server     │  services/api/
                        │  (Chi router, slog)   │  Port 8080
                        └─────┬───────────┬─────┘
                              │           │
                     pgx pool │           │ storage.Store
                              ▼           ▼
                      ┌──────────────┐  ┌───────────────────┐
                      │PostgreSQL 16 │  │  Object Storage   │
                      │(Single schema│  │ (Local FS / S3)   │
                      └──────────────┘  └───────────────────┘
```

### Core Architecture Highlights

1. **Atomic Concurrency for RSVPs:** Prevents overbooking race conditions via explicit row-level locking (`SELECT ... FOR UPDATE`) and database-level unique constraints.
2. **Compile-Time Type-Safe SQL:** Powered by `sqlc` to generate Go models and repositories directly from PostgreSQL schema migrations without ORM overhead.
3. **Strict Domain Boundaries:** Domain modules (`auth`, `users`, `organizers`, `events`, `rsvps`, `notifications`, `platform`, `storage`) enforce clear separation of concerns under `services/api/internal/`.
4. **Server-Validated Sessions:** Cookie-authenticated stateful sessions stored in PostgreSQL with SHA-256 token hashing; no sensitive data or JWTs exposed client-side.

---

## Technology Stack

| Layer | Technology | Purpose |
|---|---|---|
| **Backend API** | Go 1.23+ | High-throughput, concurrent modular monolith |
| **HTTP Routing** | Chi v5 (`net/http`) | Idiomatic, low-overhead REST routing and middleware pipeline |
| **Database Driver** | `pgx/v5` | Native PostgreSQL connection pooling and binary protocol |
| **Data Persistence** | PostgreSQL 16 | Relational consistency, full-text GIN search, row locks |
| **Query Generator** | `sqlc` v1.27+ | Type-safe Go query generation from standard SQL |
| **Migrations** | `golang-migrate` | Numbered, versioned, reversible schema migrations |
| **Frontend Framework** | Next.js 14+ (App Router) | React Server Components (RSC), SSR for SEO, Server Actions |
| **Styling** | Tailwind CSS | Utility-first design tokens and responsive primitives |
| **Form Management** | `react-hook-form` + `zod` | Client & server schema validation and form state |
| **Containerization** | Docker & Docker Compose | Uniform local development and containerized deployments |

---

## Repository Structure

```
hackways/
├── .github/                  # GitHub Actions CI/CD workflows and issue templates
│   ├── workflows/            # Automated CI, linting, testing, and security scanning
│   ├── ISSUE_TEMPLATE/       # Structured bug and feature issue templates
│   └── pull_request_template.md # Comprehensive PR review checklist
├── apps/
│   └── web/                  # Next.js 14 App Router frontend
├── services/
│   └── api/                  # Go modular monolith application
│       ├── cmd/api/          # Main entrypoint (main.go)
│       ├── internal/         # Private domain modules
│       │   ├── auth/         # Authentication, bcrypt, session middleware
│       │   ├── users/        # User accounts and profiles
│       │   ├── organizers/   # Organizer identities and permissions
│       │   ├── events/       # Event aggregate, state machine, search
│       │   ├── rsvps/        # RSVP registration, capacity enforcement
│       │   ├── notifications/# Transactional email dispatch
│       │   ├── platform/     # Config, health checks, shared middleware
│       │   └── storage/      # File upload interface (Local / S3)
│       ├── migrations/       # Numbered SQL migration files (.up.sql / .down.sql)
│       └── db/               # sqlc-generated query models
├── packages/
│   └── api-client/           # Shared TypeScript API client
├── docs/                     # Authoritative engineering documentation
│   ├── ARCHITECTURE.md       # Comprehensive system design
│   ├── API.md                # REST API contract specification
│   ├── DATABASE.md           # Schema, indexing, and migration specifications
│   ├── SECURITY.md           # Threat model and security policies
│   ├── DEPLOYMENT.md         # Infrastructure and deployment procedures
│   └── TESTING.md            # Testing standards and coverage expectations
├── docker-compose.yml        # Local development container topology
├── .env.example              # Documented environment variables template
├── .gitignore                # Production exclusions
├── AGENTS.md                 # Engineering standards and agent ownership map
├── CONTRIBUTING.md           # Contribution guidelines & branching model
├── README.md                 # This repository overview
└── WORKLOGS.md               # Auditable development progress log
```

---

## Prerequisites

Before starting local development, ensure the following tools are installed:

- **Go:** `1.23+` ([Installation Guide](https://go.dev/doc/install))
- **Node.js:** `22+` (LTS recommended, includes `npm 10+`)
- **Docker & Docker Compose:** Docker Desktop or Docker Engine v25+
- **sqlc:** `v1.27+` (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
- **golang-migrate CLI:** `v4.18+` ([Installation Guide](https://github.com/golang-migrate/migrate))

---

## Local Setup & Quick Start

### 1. Clone the Repository
```bash
git clone https://github.com/varshith-dev/HACKWAYS-IN.git
cd HACKWAYS-IN
```

### 2. Configure Environment Variables
```bash
cp .env.example .env
```
*Review `.env` and adjust database credentials or session secrets if needed.*

### 3. Start Database via Docker Compose
```bash
docker compose up -d postgres
```

### 4. Run Database Migrations
```bash
migrate -path services/api/migrations -database "$DATABASE_URL" up
```

### 5. Launch the Backend API
```bash
cd services/api
go run ./cmd/api
# API server starts on http://localhost:8080
# Verify health: curl http://localhost:8080/health
```

### 6. Launch the Frontend
```bash
cd apps/web
npm install
npm run dev
# Frontend web application starts on http://localhost:3000
```

---

## Git Branching Strategy

Hackways follows a robust Git Flow-inspired branching model suited for production releases and team collaboration:

```
feature/*  fix/*
   │         │
   ▼         ▼
┌─────────────────────────────────┐
│             develop             │  <-- Integration branch (all PRs merge here)
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│            release/*            │  <-- Release stabilization & staging tests
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│              main               │  <-- Production branch (always deployable)
└────────────────┬────────────────┘
                 ▲
                 │
             hotfix/*                <-- Urgent production patches
```

### Branch Definitions

- **`main`**: Production branch. Contains stable, thoroughly tested code running in production. No direct commits allowed; all changes arrive via pull requests from `release/*` or `hotfix/*`.
- **`develop`**: Integration branch. Aggregates completed features and fixes for the upcoming release.
- **`feature/<short-slug>`**: Feature branches branched from `develop` and merged back into `develop` via PR upon review.
- **`fix/<short-slug>`**: Non-critical bug fix branches branched from `develop` and merged back into `develop`.
- **`release/<version>`**: Release candidates branched from `develop`, tested, and merged into both `main` and `develop`.
- **`hotfix/<short-slug>`**: Critical production fixes branched directly from `main` and merged into both `main` and `develop`.

---

## Development Workflow

1. Create a branch from `develop`:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/event-capacity-guard
   ```
2. Write clean, tested code following standards in `AGENTS.md`.
3. Format and lint code:
   ```bash
   # Go backend
   gofmt -w ./... && go vet ./...

   # Frontend
   npm run lint && npm run type-check
   ```
4. Verify tests pass locally:
   ```bash
   cd services/api && go test ./... -race -count=1
   ```
5. Commit using **Conventional Commits**:
   ```bash
   git commit -m "feat(rsvps): enforce atomic row lock during registration"
   ```
6. Push branch and open a Pull Request against `develop`.

---

## Testing & Quality Assurance

All features require automated verification before merging.

### Backend Tests
```bash
cd services/api

# Run unit tests with race detection
go test ./... -race -count=1

# Run integration tests against real PostgreSQL
go test ./... -race -count=1 -run Integration
```

### Frontend Tests
```bash
cd apps/web

# Unit & component tests (Vitest)
npm test

# End-to-end tests (Playwright)
npm run test:e2e
```

---

## CI/CD Automation

Every pull request and push to `main` and `develop` triggers automated GitHub Actions workflows:

1. **Backend Verification:** Go build, formatting checks (`gofmt`), static analysis (`go vet`, `golangci-lint`), and unit/integration test execution with the race detector.
2. **Frontend Verification:** TypeScript compiler checks (`tsc --noEmit`), ESLint validation, and component tests.
3. **Security Audits:** Secret detection and dependency vulnerability scans.
4. **Database Migration Checks:** Automated verification that migrations apply cleanly forwards and rollback backwards.

---

## Documentation Index

- [System Architecture](docs/ARCHITECTURE.md)
- [REST API Specification](docs/API.md)
- [Database Schema & Migrations](docs/DATABASE.md)
- [Security Threat Model](docs/SECURITY.md)
- [Deployment & Operations](docs/DEPLOYMENT.md)
- [Testing Standards](docs/TESTING.md)
- [Engineering Standards (AGENTS.md)](AGENTS.md)

---

## Contributing

Please review [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on pull requests, code reviews, and commit standards.

## Security

To report security vulnerabilities, please refer to [SECURITY.md](SECURITY.md). Never submit vulnerability reports through public GitHub issues.

---

## License

This project is licensed under the [MIT License](LICENSE).
