# Phase 2 Implementation Plan: Dependency & Toolchain Baseline

## 1. Metadata
- **Phase Identifier:** Phase 2 (`phase2`)
- **Title:** Dependency and Toolchain Baseline
- **Owner Role:** Tech Lead / Tooling Engineer
- **Current Status:** `PLANNED`
- **Revision Date:** 2026-10-13

---

## 2. Purpose & Expected Outcome
Phase 2 establishes reproducible, isolated, and compliant toolchains for all Node.js/TypeScript frontend applications and Go backend engines. It purges strictly prohibited packages, configures independent builds with Next.js `output: 'standalone'`, integrates `pgx/v5` into Go modules, and establishes workspace linkage for shared libraries.

**Expected Outcome:**
- `lucide-react` completely removed from all frontend manifests and lockfiles.
- `apps/web/next.config.js` updated to enforce `output: 'standalone'`.
- `services/api/go.mod` updated to include `github.com/jackc/pgx/v5`.
- npm and Go build, lint, format, and type-check commands verified and reproducible.

---

## 3. Scope & Non-Goals

### In Scope:
- Purging prohibited dependencies: `lucide-react`, `SQLGlot`, `pgglot`, `Psycopg`, `psycosql` ([`AGENTS.md` Rule 1.2 & 3.12](../../AGENTS.md#L30)).
- Adding required database driver `github.com/jackc/pgx/v5` to backend Go modules.
- Enabling Next.js `output: 'standalone'` across all frontend configurations.
- Linking `packages/api-client` to frontend apps via npm/pnpm workspace or relative dependency.
- Establishing linting and formatting targets in root `Makefile`.

### Explicit Non-Goals:
- Writing database queries or applying migrations (Phase 3).
- Authoring application UI pages or components (Phase 6).
- Creating production Docker containers or Nginx configs (Phase 8).

---

## 4. Architectural Requirements & Constraints
- Must obey Azure VM memory limits: Next.js builds must produce standalone outputs (< 250 MB runtime RSS).
- Go engines must compile with `CGO_ENABLED=0` for minimal scratch/Alpine containers.
- Zero dependencies on prohibited UI icon libraries or runtime ORMs.

---

## 5. Prerequisites & Dependencies
- **Prerequisite Phase:** Phase 1 (Repository Discovery & Architecture Alignment).
- **Prerequisite Decisions:** DEC-003 (`lucide-react` removal approved).

---

## 6. Ordered Implementation Steps
1. **Step 1 (Purge Prohibited Dependencies):** Remove `lucide-react` from `apps/web/package.json`. Regenerate `package-lock.json`.
2. **Step 2 (Standalone Next.js Configuration):** Add `output: 'standalone'` to `apps/web/next.config.js`.
3. **Step 3 (Go Dependency Integration):** Add `github.com/jackc/pgx/v5 v5.7.1` to `services/api/go.mod`. Run `go mod tidy`.
4. **Step 4 (Shared Package Linkage):** Configure workspace resolution so `@hackways/api-client` can be imported by frontend apps.
5. **Step 5 (Validation):** Run `npm run type-check` in frontend and `go vet ./...` in backend.

---

## 7. Expected File Changes
- `apps/web/package.json` & `apps/web/package-lock.json`
- `apps/web/next.config.js`
- `services/api/go.mod` & `services/api/go.sum`
- Root `Makefile` & `package.json` (if workspace root is introduced)

---

## 8. Verification Commands & Evidence
- `npm ls lucide-react` -> Returns `empty` (proves package is not present).
- `cd apps/web && npm run type-check` -> Exit code 0.
- `cd services/api && go vet ./...` -> Exit code 0.
- `cd services/api && go test ./... -v` -> All unit tests pass.

---

## 9. Acceptance Criteria & Phase Exit Conditions
- [ ] `lucide-react` is absent from all manifests and lockfiles.
- [ ] `output: 'standalone'` is active in Next.js config.
- [ ] `pgx/v5` is present in `go.mod` and vendor/sum.
- [ ] Frontend type-check passes with zero errors.
- [ ] Go backend compiles cleanly with `pgx/v5`.
