# GitHub Copilot Instructions — Hackways

All AI assistance, code generation, and pull request reviews in this repository must comply with the authoritative rules in:

👉 **[AGENTS.md](../AGENTS.md)**

### Mandatory Architectural Rules
- **6 Decoupled Services:** 4 Next.js 14 web applications in `apps/` and 2 Go 1.23 engines in `services/`. Do not combine or blur boundaries.
- **Resource Constraints (Azure VM 2 vCPU, 4 GB RAM):** Strict memory limits, off-node Next.js builds only (`output: 'standalone'`), log rotation (`10m x 3`).
- **PostgreSQL Database:** Maximum 60 connections (`pgxpool` max 28 for core-backend, 22 for global-backend). Compile-time parameterized SQL via `sqlc` only.
- **RSVP Concurrency:** Pessimistic row locking (`SELECT ... FOR UPDATE` with `statement_timeout = '2000ms'`). No double-booking.
- **Azure Managed Services:** Azure Blob Storage direct client SAS uploads (no uploads on VM disk). Azure Communication Services (ACS) Email transactional outbox.
- **Security:** Host-scoped session cookies (`__Host-hackways_session`), admin IP/VPN isolation, zero secrets in source control.
- **Verification:** Execute real tests (`go test ./... -race`, `npm test`) before declaring complete.

Refer to **[AGENTS.md](../AGENTS.md)** for complete specifications.
