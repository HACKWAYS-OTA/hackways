# Claude Code Instructions — Hackways

All AI agents, including Claude Code, working on the Hackways repository **MUST** strictly adhere to the canonical repository rules specified in:

👉 **[AGENTS.md](AGENTS.md)**

### Critical Non-Negotiable Rules Summary
1. **Inspect Before Modifying:** Always inspect relevant code, schemas, and architecture rules before creating, editing, or deleting files.
2. **6 Independent Production Services:** Maintain 4 Next.js web apps (`public-web`, `core-web`, `organizer-web`, `admin-web`) and 2 Go engines (`core-backend`, `global-backend`). Never merge them.
3. **Azure VM Limits (2 vCPU, 4 GB RAM, 50 GB NVMe):** Never run Next.js builds on the VM host. Obey hard container memory caps (~2.49 GB total ceiling).
4. **Cloud Managed Services:** Use Azure Blob Storage direct SAS tokens for media uploads (zero disk storage on VM). Use Azure Communication Services (ACS) Email transactional outbox (no local SMTP daemons).
5. **RSVP Concurrency Invariant:** Reservations must enforce pessimistic locking via PostgreSQL `SELECT ... FOR UPDATE` with `statement_timeout = '2000ms'`.
6. **No Mocks or Raw SQL:** Compile-time `sqlc` queries only. No in-memory mocks on production paths.
7. **Verifiable Claims Only:** Never claim tests or commands passed without executing them and verifying the exit code and output.

Refer to **[AGENTS.md](AGENTS.md)** for complete specifications.
