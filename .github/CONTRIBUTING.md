# Contributing to Hackways

Thank you for contributing to **Hackways**! This document provides engineering standards, git workflows, and contribution guidelines for our production-grade event discovery and RSVP platform.

---

## 1. Code of Conduct

All contributors, maintainers, and reviewers are expected to uphold the standards described in [.github/CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

---

## 2. Monorepo Architecture Overview

Hackways is organized as a decoupled monorepo supporting 6 core production services running on a constrained VM target (2 vCPU, 4 GB RAM):

- **Web Applications (`apps/`):**
  - `apps/public-web/`: Landing pages, marketing, SEO (Next.js 14 SSG/ISR)
  - `apps/core-web/`: Event discovery, attendee registration, RSVP status (Next.js 14 App Router)
  - `apps/organizer-web/`: Event studio, attendee roster, analytics (Next.js 14 App Router)
  - `apps/admin-web/`: System moderation, audit logs, tenant controls (Next.js 14 App Router)
- **Backend Engines (`services/`):**
  - `services/core-backend/`: Event lifecycle, ticket inventory, atomic RSVP concurrency engine (Go 1.23, Chi, pgx/v5)
  - `services/global-backend/`: Authentication, sessions, RBAC, notifications, media storage (Go 1.23, Chi, pgx/v5)
- **Shared Packages (`packages/`):**
  - `packages/api-client/`: Generated TypeScript client
  - `packages/shared-types/`: Shared domain contracts
- **Deployment & Infra (`deploy/`):**
  - `deploy/docker/`: Production Docker Compose topologies
  - `deploy/nginx/`: Edge reverse proxy routing and rate limiting

---

## 3. Git Branching Strategy

The project follows this branch and PR policy. Required reviews and branch
protection are not technically enforced on the current private GitHub Free plan.
Contributors must use PRs; do not push new work directly to `main` or enable
automatic merging/deployment as a substitute for operator approval.

| Branch | Purpose | Base Branch | Merges Into |
|---|---|---|---|
| `main` | Production-ready, stable releases | — | — |
| `develop` | Ongoing integration branch | `main` | `main` (via release) |
| `feature/<slug>` | New capabilities or enhancements | `develop` | `develop` |
| `fix/<slug>` | Bug fixes | `develop` | `develop` |
| `release/<version>`| Stabilization for release candidate | `develop` | `main` and `develop` |
| `hotfix/<slug>` | Urgent production fixes | `main` | `main` and `develop` |

### Rules:
1. **Direct commits to `main` and `develop` are prohibited.** All changes must arrive via reviewed Pull Requests.
2. Branch names must follow `<type>/<kebab-case-slug>` (e.g., `feature/rsvp-capacity-lock`, `fix/organizer-auth-cookie`).
3. Keep feature branches short-lived and rebased against `develop`.

---

## 4. Commit Message Conventions

We enforce [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description in imperative mood>

[optional body explaining rationale, trade-offs, and design decisions]

[optional footer referencing issue or task ID, e.g. Fixes #42 or Refs M2-001]
```

### Allowed Types:
- `feat`: New feature or user-facing capability
- `fix`: Bug fix
- `docs`: Documentation updates or additions
- `refactor`: Code reorganization with zero behavior change
- `perf`: Performance optimizations
- `test`: Adding or updating test suites
- `chore`: Tooling, build scripts, or dependency updates
- `ci`: CI/CD workflows and automation

---

## 5. Local Development & Verification

1. **Start Database & Mail Services:**
   ```bash
   make db-up
   ```
2. **Run Schema Migrations:**
   ```bash
   make migrate-up
   ```
3. **Execute Automated Quality Checks:**
   ```bash
   # Go backend checks (race detector enabled)
   make test-api
   make lint

   # Frontend checks
   make test-web
   ```

---

## 6. Definition of Done

A task or PR is considered complete only when:
- [ ] Business logic implemented and verified against PostgreSQL.
- [ ] Database transactions enforce concurrency safety (`SELECT ... FOR UPDATE` on RSVP reservations).
- [ ] Server-side input validation and authorization checks are enforced (no cosmetic frontend checks).
- [ ] Zero mock data or placeholder controls on production paths.
- [ ] Tests written and passing (`go test -race`, `npm test`).
- [ ] All code formatted (`gofmt`, Prettier) and linted cleanly.
- [ ] Documentation and migration files updated.
