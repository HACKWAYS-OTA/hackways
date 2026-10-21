# Database and SQL Workflow

## Authority

PostgreSQL is the source of truth. The current schema is defined by sequential migrations in `deploy/migrations/`; do not create independent migration histories inside the backend services. The initial schema creates users, sessions, organizers, events, RSVPs, and the transactional email outbox.

Migration `000002_free_event_workflows` adds workflow policies/revisions, held inventory, approval/offer deadlines, transfer history, tickets, audit, request idempotency, authentication attempt limits, email verification and outbox claim tokens. Existing confirmed registrations receive ticket records. The normalized guest identity index rejects collisions instead of merging/deleting rows. Its down migration deliberately refuses destructive rollback; see [rollout and recovery notes](backend/README.md).

## Query workflow

Each Go service has its own sqlc configuration and generated `db` package. Both configs compile against `../../deploy/migrations/` and the service's `db/queries/` sources. Change query SQL first, then run `sqlc generate` from each service directory whose query set changed. Commit generated code together with the query source. Runtime ORM, string-built SQL, and unparameterized queries are prohibited.

```powershell
Push-Location services/core-backend
sqlc generate
Pop-Location

Push-Location services/global-backend
sqlc generate
Pop-Location
```

## RSVP integrity

RSVP creation must execute in one PostgreSQL transaction:

1. Set `SET LOCAL statement_timeout = '2000ms'`.
2. Lock the active event row using `SELECT ... FOR UPDATE`.
3. Check event state and capacity while holding the lock.
4. Insert idempotently under database uniqueness constraints.
5. Increment the event's confirmed RSVP counter only when the insert succeeds.
6. Insert the email outbox row in the same transaction.
7. Commit before returning success.

The current schema uses unique constraints on `(event_id, user_id)` and `(event_id, guest_email)`. PostgreSQL permits multiple NULL values under ordinary unique constraints, so authenticated idempotency is enforced for users; guest email uniqueness also applies when provided. Verify any change to identity rules against these constraints and the handler.

The normalized index also covers `(event_id, lower(guest_email))` for non-null emails. Confirmed registrations and held offers use separate counters; their sum is protected by a CHECK constraint. Pending applications and waitlist entries consume no capacity. Capacity transitions lock the event first and commit registration, counter, ticket, audit and outbox changes together. Transfers keep total inventory constant. Idempotency keys expire after 24 hours; identity uniqueness does not expire. Offer/approval confirmation also checks its deadline in SQL.

## Connection limits

The production budget is 60 total connections: core backend up to 28, global backend up to 22, with ten reserved for migrations and administration. Keep local and production environment values within those ceilings. Pool defaults in code are not proof that deployed environment overrides respect them.

## Spaces and event presentation migrations

- `000003_spaces`: Spaces, unique memberships, creation replay keys/hashes, revision and archive state, event association, appearance, and random unique registration/ticket codes. The `(space_id,organizer_id)` foreign key enforces owner-only event hosting under a Space. Ticket reissue rotates its public code.
- `000004_event_information`: optional FAQ, HTTPS contact link and rewards text; empty defaults preserve existing records.
- `000005_event_display`: persisted light/dark event display, default light.
- `000006_event_visibility`: public/link-only visibility, default public, and a partial public-discovery index. Listing SQL excludes private events while slug reads remain available. Publication and visibility are separate.

Both engines regenerate sqlc models against the full migration sequence. Down files deliberately refuse destructive rollback; preserve persisted content and visibility when reverting application code. These additive migrations were applied to the local application schema and isolated verification schemas without deleting application records. The local database has no historical migration-tool version metadata; reconcile that metadata before introducing automated migrate-up to this existing database.

## Operational rules

- Use backward-compatible, numbered migrations. Do not apply destructive production DDL without a reviewed transition plan.
- Never use real production data in tests or local development.
- Verify query plans and indexes when measured query latency identifies a problem; do not add speculative indexes.
- Production database connection strings require TLS and secrets must be injected at runtime.
- Media belongs in Azure Blob Storage. PostgreSQL stores metadata and references, not uploaded file bodies.

## Registration form persistence (000007)

Events have an independent registration_form_revision and optional registration_name_required flag. event_registration_questions stores bounded metadata, ordering, options, active state and unique Q-codes. Owner mutations and new RSVP validation serialize using the same event row lock and 2-second statement timeout. Missing questions are deactivated, never deleted by the application.

registration_submissions stores append-only answer snapshots with the corresponding form revision and RSVP. It is indexed by RSVP/latest submission. Re-registration adds a new submission rather than replacing historical answers. Transfers do not copy answer snapshots. Latest answers are available only through event-owner guest detail authorization. Payloads are capped at 20 answers, 2,000 characters per answer and 20,000 combined text bytes. No binary media or local JSON persistence is introduced.

Both engines regenerate sqlc against this additive migration. Its down file deliberately refuses data-destroying rollback, consistent with the existing content/visibility migrations; revert application code without deleting submissions, or prepare a separately reviewed compatible transition.