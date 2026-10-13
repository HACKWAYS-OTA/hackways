# Phase 2 Agent Pre-Context: Dependency & Toolchain Baseline

## 1. Product Objective & Technical Target
Phase 2 establishes clean, compliant, and isolated dependency manifests across all Hackways services. It eliminates prohibited dependencies and prepares the build configurations for the resource-constrained Azure VM.

---

## 2. Canonical Specifications to Obey
1. **[`AGENTS.md` Rule 1.2 & Rule 3.12](../../../AGENTS.md#L30):**
   - **"Never introduce `lucide-react` anywhere in Hackways."**
   - **"SQLGlot, `pgglot`, Psycopg, and `psycosql` are strictly prohibited."**
2. **[`AGENTS.md` Rule 2.2](../../../AGENTS.md#L100):**
   - Next.js must be configured with `output: 'standalone'` and built off-node in GitHub Actions.
   - Go engines must obey `GOMEMLIMIT` and compile with `CGO_ENABLED=0`.
3. **[`AGENTS.md` Rule 2.5](../../../AGENTS.md#L202-L215):**
   - PostgreSQL queries must use `pgx/v5` and `sqlc`.

---

## 3. Verified Repository Facts Before Phase 2
- [`apps/web/package.json`](../../../apps/web/package.json#L15) line 15 currently contains `"lucide-react": "^0.460.0"`.
- `apps/web/node_modules` does not exist on disk.
- [`services/api/go.mod`](../../../services/api/go.mod) only requires `github.com/go-chi/chi/v5 v5.3.2`. `github.com/jackc/pgx/v5` is missing.
- [`apps/web/next.config.js`](../../../apps/web/next.config.js) lacks `output: 'standalone'`.

---

## 4. Expected Outputs
- `apps/web/package.json` with zero occurrences of `lucide-react`.
- `apps/web/next.config.js` with `output: 'standalone'`.
- `services/api/go.mod` with `github.com/jackc/pgx/v5`.
- Passing `npm run type-check` and `go vet ./...`.
