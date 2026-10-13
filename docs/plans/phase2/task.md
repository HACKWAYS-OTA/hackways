# Phase 2 Task Register: Dependency & Toolchain Baseline

## Tasks Checklist

| Task ID | Title | Prerequisites | Status | Owner | Blockers / References |
|:---:|:---|:---:|:---:|:---:|:---|
| `P2-TOOL-001` | Purge `lucide-react` from `apps/web/package.json` & lockfile | Phase 1 Exit Gate | `COMPLETED` | Frontend Lead | BLK-002, DEC-003 |
| `P2-TOOL-002` | Configure `output: 'standalone'` in `apps/web/next.config.js` | `P2-TOOL-001` | `COMPLETED` | Frontend Lead | AD-006, GAP-007 |
| `P2-TOOL-003` | Add `github.com/jackc/pgx/v5` to `services/api/go.mod` | Phase 1 Exit Gate | `COMPLETED` | Backend Lead | BLK-004 |
| `P2-TOOL-004` | Execute `go mod tidy` and verify backend builds with `pgx/v5` | `P2-TOOL-003` | `COMPLETED` | Backend Lead | `go build` check |
| `P2-TOOL-005` | Install frontend dependencies cleanly (`npm ci` or `install`) | `P2-TOOL-001` | `COMPLETED` | Frontend Lead | Node 20+ required |
| `P2-TOOL-006` | Link `@hackways/api-client` package to `apps/web` | `P2-TOOL-005` | `COMPLETED` | Tooling Lead | Workspace link |
| `P2-TOOL-007` | Verify TypeScript type-checking across frontend and shared package | `P2-TOOL-006` | `COMPLETED` | Frontend Lead | `npm run type-check` |
| `P2-TOOL-008` | Execute Phase 2 exit gate verification | All P2 Tasks | `COMPLETED` | Tech Lead | Phase Gate 2 |

---

## Detailed Task Specifications & Verified Evidence

### `P2-TOOL-001`: Purge `lucide-react` from frontend
- **Description:** Remove `"lucide-react": "^0.460.0"` from `apps/web/package.json` and remove from `package-lock.json`. Verify no imports exist in code.
- **Acceptance Criteria:** `npm ls lucide-react` returns no matches; grep confirms zero occurrences.
- **Status:** `COMPLETED`.
- **Evidence:** `npm ls lucide-react` outputs `@hackways/web@1.0.0` `-- (empty)`. Zero references across all frontend files.

### `P2-TOOL-002`: Configure Next.js standalone output
- **Description:** Add `output: 'standalone'` to `apps/web/next.config.js` so build output is minimized for Docker containers.
- **Acceptance Criteria:** `next.config.js` contains `output: 'standalone'`.
- **Status:** `COMPLETED`.
- **Evidence:** `apps/web/next.config.js` verified containing `output: "standalone"`.

### `P2-TOOL-003`: Add `pgx/v5` driver to Go backend
- **Description:** Add `github.com/jackc/pgx/v5` and `github.com/jackc/pgx/v5/pgxpool` to Go module.
- **Acceptance Criteria:** `go.mod` includes `github.com/jackc/pgx/v5`.
- **Status:** `COMPLETED`.
- **Evidence:** `services/api/go.mod` includes `github.com/jackc/pgx/v5 v5.11.0`.

### `P2-TOOL-004`: Verify Go backend compilation with `pgx/v5`
- **Description:** Run `go mod tidy` and compile backend binary.
- **Acceptance Criteria:** `go build -o tmp_build.exe ./cmd/api` succeeds with exit code 0.
- **Status:** `COMPLETED`.
- **Evidence:** `go build` exits 0; `go test ./...` passes 7/7 (0.955s); `go vet ./...` clean; `gofmt -l .` clean.

### `P2-TOOL-005`: Clean install frontend dependencies
- **Description:** Run clean installation of Node modules in `apps/web` without prohibited dependencies.
- **Acceptance Criteria:** `node_modules` generated; zero audit vulnerabilities for core stack.
- **Status:** `COMPLETED`.
- **Evidence:** `npm install` completed with exit code 0, 396 packages audited.

### `P2-TOOL-006`: Link shared API client package
- **Description:** Ensure `apps/web` can import types from `@hackways/api-client` rather than duplicating them in `apps/web/types/api.ts`.
- **Acceptance Criteria:** Types imported cleanly from package.
- **Status:** `COMPLETED`.
- **Evidence:** Added `"@hackways/api-client": "file:../../packages/api-client"` in `apps/web/package.json`. `apps/web/types/api.ts` re-exports from `@hackways/api-client`.

### `P2-TOOL-007`: Verify TypeScript compilation
- **Description:** Run `npm run type-check` in `apps/web` and `packages/api-client`.
- **Acceptance Criteria:** TypeScript reports 0 diagnostic errors.
- **Status:** `COMPLETED`.
- **Evidence:** `npm run type-check` in `apps/web` exits 0; `npm run type-check` and `npm run build` in `packages/api-client` exits 0. Next.js linter `npm run lint` exits 0 with zero warnings or errors.

### `P2-TOOL-008`: Phase 2 exit gate verification
- **Description:** Confirm all criteria in `docs/plans/phase2/verification.md` and complete `handoff.md`.
- **Acceptance Criteria:** Exit gate approved; Phase 2 marked `VERIFIED`.
- **Status:** `COMPLETED`.
- **Evidence:** All 4 verification protocol checks passed with verifiable system output.
