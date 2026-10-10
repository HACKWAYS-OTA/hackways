# Contributing to Hackways

Thank you for your interest in contributing to **Hackways**! This document provides instructions for collaborating on the platform in a structured, maintainable, and production-grade manner.

---

## 1. Code of Conduct

All contributors and maintainers are expected to uphold the standards described in [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Respectful, constructive collaboration is mandatory.

---

## 2. Git Branching Strategy

We follow a Git Flow-inspired branching workflow:

| Branch | Purpose | Base Branch | Merged Into |
|---|---|---|---|
| `main` | Production-ready, stable code | — | — |
| `develop` | Ongoing integration branch | `main` | `main` (via release) |
| `feature/<slug>` | New capabilities or enhancements | `develop` | `develop` |
| `fix/<slug>` | Non-critical bug fixes | `develop` | `develop` |
| `release/<version>`| Stabilization for release candidate | `develop` | `main` and `develop` |
| `hotfix/<slug>` | Urgent production fixes | `main` | `main` and `develop` |

### Rules:
1. **Never commit directly to `main` or `develop`.** All changes must arrive via a Pull Request.
2. Branch names must follow the format `<type>/<short-descriptive-slug>` (e.g., `feature/event-filtering`, `fix/rsvp-capacity-lock`).
3. Keep feature branches short-lived and regularly rebased or synchronized with `develop`.

---

## 3. Commit Message Conventions

We enforce [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description in imperative mood>

[optional body explaining rationale and context]

[optional footer referencing issue or task ID, e.g. Fixes #123 or Refs M1-002]
```

### Allowed Types:
- `feat`: A new feature for users or organizers
- `fix`: A bug fix
- `docs`: Documentation updates or additions
- `refactor`: Code changes that neither fix a bug nor add a feature
- `test`: Adding or updating test suites
- `chore`: Tooling, build scripts, or dependency updates
- `perf`: Code changes that improve performance
- `ci`: Changes to GitHub Actions or automation workflows

### Example:
```
feat(events): add full-text GIN search query for event discovery

Implements PostgreSQL tsvector search on title and description columns
to provide fast discovery across large event catalogs.

Refs M2-004
```

---

## 4. Development & Pull Request Workflow

1. **Create your branch:**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/organizer-dashboard
   ```
2. **Implement changes with tests:**
   - Every backend feature must have matching unit and/or integration tests.
   - Frontend changes must pass component and type tests.
   - Do NOT introduce mock persistence or fake data on production paths.
3. **Run local quality checks:**
   ```bash
   # Go formatting & linting
   gofmt -w ./...
   go vet ./...

   # Go tests with race detection
   cd services/api && go test ./... -race -count=1

   # Frontend linting & type checks
   cd apps/web && npm run lint && npm run type-check
   ```
4. **Push branch to GitHub:**
   ```bash
   git push origin feature/organizer-dashboard
   ```
5. **Open a Pull Request:**
   - Target branch: `develop` (unless hotfix targeting `main`).
   - Fill out all sections of the [Pull Request Template](.github/pull_request_template.md).
   - Link any associated GitHub issues or implementation tasks.
6. **Code Review:**
   - At least one approval is required prior to merge.
   - All automated CI checks must pass.
   - Squash-and-merge or rebase-and-merge as designated by maintainers.

---

## 5. Definition of Done

A contribution is considered complete only when:
- [ ] End-to-end implementation is verified against PostgreSQL.
- [ ] Backend and frontend schemas agree.
- [ ] Input validation and authorization are enforced on the server.
- [ ] Edge cases (empty, capacity limit, duplicate, unauthorized) are covered.
- [ ] All automated tests pass with real output (`go test -race`, `npm test`).
- [ ] No secrets or sensitive data are logged or committed.
- [ ] Migration scripts exist for all schema modifications.
