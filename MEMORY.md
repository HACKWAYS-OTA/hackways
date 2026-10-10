# MEMORY.md — Hackways Durable Engineering Knowledge

---

## Architectural Decisions

### AD-001 — Modular Monolith Architecture
**Decision:** Build as a single Go application (`services/api`) backed by PostgreSQL 16, coupled with a Next.js 14 App Router frontend (`apps/web`).  
**Rationale:** Prevents distributed system failure modes and premature operational overhead while preserving clean domain boundaries.

### AD-002 — Chi Router & slog
**Decision:** Standardize on `github.com/go-chi/chi/v5` and stdlib `log/slog`.  
**Rationale:** Standard library compatibility, zero memory overhead, structured JSON logs.

### AD-003 — pgx/v5 and sqlc
**Decision:** Use `pgx/v5` with connection pooling and `sqlc` for SQL code generation.  
**Rationale:** Compile-time verified SQL queries without ORM magic or runtime overhead.

### AD-004 — Server-Side Stateful Sessions in PostgreSQL
**Decision:** Store sessions server-side with opaque cookie tokens, avoiding JWTs.  
**Rationale:** Instant server-side revocation on logout, zero token refresh vulnerabilities.

### AD-005 — Atomic RSVP Capacity Locks
**Decision:** Pessimistic locking via `SELECT ... FOR UPDATE` inside transactions for registration counts.  
**Rationale:** Prevents overbooking race hazards during high-demand registration spikes.

### AD-006 — Git Flow Branching Strategy
**Decision:** Standardize on `main` (production), `develop` (integration), `feature/*`, `fix/*`, `release/*`, and `hotfix/*`.  
**Rationale:** Enforces clean separation between ongoing development and stable production releases.
