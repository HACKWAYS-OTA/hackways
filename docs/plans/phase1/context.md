# Phase 1 Agent Pre-Context: Repository Discovery & Architecture Alignment

## 1. Product Objective & Technical Target
You are working on **Hackways** (`hackways.in`), a high-concurrency event discovery, ticketing, and RSVP platform engineered to run on a resource-constrained Azure Virtual Machine (**2 vCPU, 4 GB RAM, 50 GB NVMe SSD**). The core domain requirement is that an event with capacity $C$ must NEVER confirm more than $C$ attendees under any concurrency load.

---

## 2. Canonical Specifications to Obey
1. **[`AGENTS.md`](../../../AGENTS.md):** Canonical, non-negotiable engineering rules.
   - Rule 1.1: No mock data or fake success states in application logic.
   - Rule 1.2: Maintain 6 independent production services (4 Next.js apps, 2 Go engines). **Never introduce `lucide-react`.**
   - Rule 2.2: Azure VM memory limits (~2,495 MB container ceiling). No frontend builds on host VM.
   - Rule 2.3: Zero persistent media on VM disk. Direct Azure Blob SAS uploads only.
   - Rule 2.4: Pessimistic locking mandatory: `SELECT ... FOR UPDATE` with `statement_timeout = '2000ms'`.
2. **[`docs/ARCHITECTURE_PLAN.md`](../../ARCHITECTURE_PLAN.md):** Target Azure VM blueprint.

---

## 3. Verified Current Repository Facts
- **Git Branch:** `develop` (clean sync with `origin/develop`).
- **Physical Layout:** Only [`apps/web`](../../../apps/web) (Next.js 14) and [`services/api`](../../../services/api) (Go Chi v5) exist on disk.
- **Untracked File:** An untracked file [`LICENSE`](../../../LICENSE) exists at root containing OpenAI Apps SDK guidelines.
- **Installed Toolchain on Host:** Go `1.26.4`, Node.js `v24.13.0`, npm `11.6.2`. Docker is **not found on PATH**.
- **Backend Tests:** 7 unit tests pass in `services/api/internal/platform/`.
- **Frontend Dependencies:** `node_modules` does not exist in `apps/web`.

---

## 4. Known Defects to Fix in Phase 1 / Phase 2
- Prohibited dependency `"lucide-react": "^0.460.0"` in `apps/web/package.json` line 15.
- Hardcoded `SAMPLE_EVENTS` mock data in `apps/web/app/page.tsx`.
- Local disk upload driver in `services/api/internal/storage/store.go` (writes to `./uploads`).
- Missing `output: 'standalone'` in `apps/web/next.config.js`.

---

## 5. Prohibited Actions
- Do NOT begin implementing application business logic or routes.
- Do NOT fabricate passing test results or mock backend integrations.
- Do NOT introduce unapproved packages.
