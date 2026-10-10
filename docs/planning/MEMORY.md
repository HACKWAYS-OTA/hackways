# Durable Engineering Knowledge & Architecture Decisions — Hackways

---

## Architectural Decisions Log

### AD-001 — 6-Service Decoupled Monorepo Architecture
- **Decision:** Structure Hackways as 6 logically independent services within a unified monorepo: 4 Next.js 14 frontends (`public-web`, `core-web`, `organizer-web`, `admin-web`) and 2 Go 1.23 backend engines (`core-backend`, `global-backend`), running under Docker Compose.
- **Rationale:** Preserves monorepo development velocity while maintaining strict domain separation, enabling independent scaling or splitting to multiple VMs in the future with zero code changes.

### AD-002 — Go Engines with Chi Router & slog
- **Decision:** Standardize on Go 1.23+ with `github.com/go-chi/chi/v5` and standard library `log/slog`.
- **Rationale:** Minimal runtime footprint (~35MB RSS each), high throughput, zero external framework overhead, and structured JSON logs.

### AD-003 — Connection Pooling & PostgreSQL Resource Budget
- **Decision:** Cap PostgreSQL connections to 60. Allocate 28 max conns to `core-backend`, 22 to `global-backend`, and reserve 10 for migrations/admin. Tune `shared_buffers = 256MB` and `work_mem = 4MB`.
- **Rationale:** Prevents database connection exhaustion and memory thrashing on a 4 GB RAM production VM.

### AD-004 — Cryptographic Host-Scoped Cookie Sessions
- **Decision:** Avoid JWTs in favor of server-side stateful sessions in PostgreSQL. Issue `__Host-hackways_session` for attendee/organizer apps and `__Host-hackways_admin_session` for admin web.
- **Rationale:** Prevents session leakage across subdomains, eliminates token refresh vulnerabilities, and enables instant server-side revocation on logout or compromise.

### AD-005 — Pessimistic Row-Level Locking for RSVP Registration
- **Decision:** Enforce capacity invariants via transactional `SELECT ... FOR UPDATE` with a 2000ms statement timeout.
- **Rationale:** Eliminates ticket overbooking race hazards during high-concurrency ticket drops. Lock hold duration is ~1.5–2.8ms on NVMe SSD, serializing 350–500 reservations/sec safely.

### AD-006 — Off-Node CI Image Builds & Strict Memory Ceilings
- **Decision:** Next.js applications MUST be pre-built in GitHub Actions (`output: 'standalone'`) and pulled as lightweight runtimes. Enforce hard Docker memory limits totaling ~2.54 GB across all containers.
- **Rationale:** Next.js build steps consume up to 2.5 GB RAM, which would trigger kernel OOM kills on a 4 GB host. Pre-building off-node maintains a ~1.46 GB safety buffer on the production host.
