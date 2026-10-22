# Hackways backend guide

[Repository overview](../../README.md) · [API reference](../API.md) · [Architecture](../ARCHITECTURE.md) · [Database](../DATABASE.md)

## Scope and readiness

This guide describes source behavior in the two Go services, not a claim that the platform is deployed or production-ready. External delivery, uploads, migration/recovery readiness, and browser journeys require their own evidence. Historical local test results below are dated context, not the latest release verdict.

Source brief: [user-supplied workflow](workflow.md). Repository constraints in [AGENTS.md](../../AGENTS.md) override its optional stack recommendations. This implements the free-event foundation and registration workflows inside the existing core/global Go engines. Payments are explicitly deferred by the user. No Redis, extra API service, ORM, or TypeScript backend was introduced.

## Implementation and boundaries

| Area | Implemented backend behavior | Remaining work |
| --- | --- | --- |
| HTTP boundary | Exact production origins, CSRF origin checks, strict typed JSON decoding, 1 MB bodies, 16 KB headers, request IDs, bounded admission/deadlines, structured errors | Load measurements, ingress configuration execution |
| Identity | Existing password sessions; per-identifier PostgreSQL attempt limits; owned session list/revocation; 15-minute, one-use email verification | Magic-link login, OAuth, passkeys, password recovery, disabled-account and deletion policy |
| Events | Complete validated drafts and published creation; revision-checked full updates; publish/unpublish/cancel; inactive draft soft deletion; public cancellation state | Incomplete draft editor model, scheduled publishing, recurring events, moderation pipeline |
| Registration | Instant free account/guest registration; explicit random idempotency keys; ownership-scoped reads/listing/notes/cancellation; verified-email guest claim | Complete browser-journey verification; broader consent/accessibility data policy |
| Approval | Account-only applications, no inventory until approval; owner decisions; expiry after the earliest of seven days, event start, or registration close | Request-edits workflow; per-event approval duration UI |
| Waitlist | Account-only FIFO waitlist for full instant-registration events; inventory-backed offers; 5-60 minute offers, default 10; accept/leave/expiry; bounded maintenance | Frontend waitlist controls; approval-plus-waitlist composition |
| Tickets | Real persisted tickets; Ed25519-signed payloads without PII; owner-authorized one-use check-in; audited manual lookup; cancellation revocation | Scanner UI, offline roster/sync conflict handling, separate staff/team roles |
| Transfers | Verified account to verified account; one transfer maximum; before start; unscanned instant-registration tickets only; old ticket revoked, new ticket issued atomically, count unchanged | Invitations for recipients without verified accounts; approval-event transfer review |
| Organizer operations | Owned event detail, paginated roster/audit; material schedule/location changes require acknowledgement and enqueue notices | Team roles/invitations, exports to Azure, insights, promotion tools |
| Spaces | Public directory/profile/feed, owner-only creation/edit/archive/reactivation, idempotent join/leave, owned Space selection and event association | Moderation, team hosting roles and automated invitations |
| Notifications | ACS HMAC REST send with stable operation ID, status reconciliation, two concurrent sends, leased rows, bounded retries, no logged delivery success; owned status/retry | Actual ACS delivery verification, inbox delivery receipts, preferences, SMS/push |
| Media | Owned SAS issuance; create-only banner uploads; Azure HEAD properties check and revision-scoped editable-event cover attachment; actual-file picker and recoverable create flow | External Azure upload/CORS verification pending credentials; content scanning and orphan-object cleanup |
| Other platform workflows | Preserved explicit unavailable/ownership/error responses | Space team roles, saves/follows, support/moderation/admin workflows, retention/export/deletion and legal policy |

The remaining columns are not implemented functionality. The frontend keeps its accepted design and now supplies a stable key for unchanged in-page retries and distinguishes an application from a confirmed place. Backend approval/waitlist/management APIs do not imply the corresponding frontend screens are complete.

## State and inventory rules

- Event wire statuses remain `draft`, `published`, `unpublished`, `cancelled` for compatibility. Ended is derived from the schedule; scheduled publishing, postponed labels, and archive states have not been invented.
- Registration states: `confirmed`, `pending_approval`, `waitlisted`, `offered`, `cancelled`, `rejected`, `expired`, `transferred`. Check-in lives on the ticket.
- `current_rsvp_count` counts confirmed registrations; `held_rsvp_count` counts active offers. Their sum may not exceed capacity. Pending approval and waitlist rows consume no inventory.
- Every inventory transition locks the same event first, uses sqlc-generated `SET LOCAL statement_timeout = '2000ms'`, and commits the state, counters, ticket change, audit, and outbox together. Request context bounds total work. No external network request occurs while holding that lock.
- The maintenance loop wakes every five seconds, considers at most 50 event IDs per pass, expires at most 100 offers/applications per event, and promotes at most 50 entries. Lease/approval timestamps remain authoritative even when a worker is delayed. No throughput or fairness-latency guarantee has been measured.
- Approval and instant waitlists are separate modes. A waitlist cannot bypass organizer approval. Changing mode with active registrations or disabling a nonempty waitlist is rejected.
- Cancellation after check-in or after the event starts is rejected. Re-registering after cancellation/expiry preserves registration identity and issues a fresh ticket ID if the old ticket was revoked. Rejected/transferred registrations cannot re-enter by retrying creation.
- Idempotency scope is event plus authenticated account or normalized guest identity; only key digests are stored. Keys are 16-128 printable characters and clients must generate at least 128 random bits (the attendee form uses UUIDv4). Exact request retries return current persisted registration status, never recreate the booking. Payload mismatch returns 409. Guest replay cannot read an account-linked record after a claim. Keys expire after 24 hours; identity uniqueness remains after key expiry.
- Transfers keep inventory constant and copy no private notes. A transfer recipient must not already have any registration for that event. Verified-email claims do not silently merge competing account/guest records.
- Audit is append-only through application queries. It contains actions and IDs, not passwords, tokens, email bodies, or notes. Database access administration/retention controls are still an operational responsibility.

## API contract

All paths below are relative to `/api/v1`; existing public catalog and auth routes remain.

| Method | Path | Access / request |
| --- | --- | --- |
| POST | `/events` | Organizer. Existing full creation body; optional `status`: `draft` or `published` (default preserves published behavior). |
| GET | `/events/{id}/manage` | Event owner; includes revision and workflow policy. |
| PUT | `/events/{id}` | Event owner; full replacement fields listed below, matching `revision`. |
| POST | `/events/{id}/status` | Event owner; `{status,revision,acknowledge_attendees}`; cancellation requires acknowledgement. |
| DELETE | `/events/{id}` | Event owner; current revision in `If-Match`; only inactive draft/unpublished events. Soft deletion retains data. |
| GET | `/events/{id}/registrations` | Event owner; `limit` 1-100, `offset`. Private roster including notes. |
| GET | `/events/{id}/audit` | Event owner; bounded pagination. |
| POST | `/events/{id}/check-in` | Event owner; signed `code`, or `ticket_id` plus manual `reason`. Wrong/revoked/duplicate ticket denied. |
| POST | `/rsvps` | Existing body plus optional `waitlist: true`; `Idempotency-Key` strongly recommended. Approval/waitlist requires account. |
| GET | `/rsvps` | Account; paginated own event registrations. |
| GET | `/rsvps/{id}` | Account owning that registration. |
| PATCH | `/rsvps/{id}` | Account owner; `{notes}` up to 2000 characters, before start. |
| POST | `/rsvps/{id}/cancel` | Account owner; idempotent cancellation/leave. |
| POST | `/rsvps/{id}/accept` | Account owner; accepts a still-valid offer. |
| POST | `/rsvps/{id}/approve` | Event owner; only pending, unexpired application and available capacity. |
| POST | `/rsvps/{id}/reject` | Event owner; only pending application. |
| POST | `/rsvps/{id}/transfer` | Verified account owner; `{recipient_email}` with restrictions above. |
| POST | `/rsvps/claim` | Verified account; bounded linking of guest registrations with its verified email. Repeat for batches over 50. |
| GET | `/rsvps/{id}/ticket` | Account owner; ticket and signed code; 503 if signing seed is absent. |
| GET | `/auth/sessions` | Account; owned session metadata, never token hashes. |
| PATCH | `/auth/me` | Account; saves its own `display_name` (1–100 characters, no control characters). Email, verification and roles are not writable here. |
| GET | `/events/{id}/invitations` | Event owner; paginated recipients and persisted outbox delivery state. |
| POST | `/events/{id}/invitations` | Verified event owner; 1–50 emails, optional 2,000-character message, required `Idempotency-Key`. Upcoming published event only; 1,000 invitations per event. Duplicate invitations and active registrants skipped. Does not reserve capacity or bypass approval. |
| DELETE | `/auth/sessions/{sessionID}` | Account owning the session. |
| DELETE | `/auth/sessions` | Account; revoke all owned sessions and clear current cookie. |
| POST | `/auth/email-verification/request` | Account; queues code only when ACS configured; response does not claim delivery. |
| POST | `/auth/email-verification/confirm` | Same account; `{code}`; single-use and unexpired. |
| GET | `/auth/notifications` | Account; bounded owned notification status without addresses/payload/tokens. |
| POST | `/auth/notifications/{notificationID}/retry` | Owner; failed message older than five minutes, rate-limited; not a delivery success. |
| POST | `/storage/upload-intent` | Account; `avatars`, or `event-banners` plus owned event `resource_id`. |
| POST | `/storage/event-banner/complete` | Account; `{resource_id, blob_key, revision}`. Verifies owned draft namespace, Azure block blob properties, MIME/extension and 5 MB size, then attaches the URL with a two-second write timeout and revision check. |

Event PUT fields: `title`, `description`, `format`, RFC3339 `starts_at`/`ends_at`, IANA `timezone`, nullable `capacity`, `location_name`, `location_address`, `online_url`, nullable `registration_opens_at`/`registration_closes_at`, `registration_mode` (`instant`/`approval`), `allow_guest`, `waitlist_enabled`, `offer_minutes` (5-60), `revision`, `acknowledge_attendees`. Material live changes queue notices in the same transaction. A revision conflict returns 409 without replacing saved data.

## Notifications and missing services

Configure `ACS_ENDPOINT`, base64 `ACS_ACCESS_KEY`, and verified `ACS_SENDER_ADDRESS` in global-backend. The existing AZURE_ACS_CONNECTION_STRING and AZURE_ACS_SENDER_ADDRESS variables remain supported; explicit ACS_* values take precedence. Missing configuration pauses delivery without consuming queued-message retries. Failed/uncertain sends never become `sent` until ACS reports send-operation success. That status means provider send acceptance, not inbox receipt.

Each attempt checks the current registration/event state and suppresses superseded notices or expired verification codes. A state can change after this read while the provider accepts a message; messages must still direct attendees to current in-app state. There is no claim of exactly-once inbox delivery. Stable operation IDs and ACS repeatability headers narrow duplicate risk; an unknown operation beyond the replay window requires operator reconciliation, not a blind resend.

Azure Blob signing needs `AZURE_STORAGE_ACCOUNT` and a base64 `AZURE_STORAGE_ACCOUNT_KEY`; no raw/development key fallback is accepted. Banner SAS tokens grant create-only access to a random owned-event object for 15 minutes. Completion privately signs a short-lived read SAS for a bounded, non-redirecting HEAD request and checks actual Blob size/type; the read token is never returned to the client. It does not inspect image bytes or malware. Attachment repeats ownership, editable event state, future start and revision checks atomically and accepts an unchanged immediate retry. Neither engine buffers media bytes or stores uploads on local disk. Azure CORS must allow the app origin, PUT, Content-Type and x-ms-blob-type; public read/CDN configuration must match the configured media URL. External Azure behavior remains unverified pending credentials. Incomplete objects may need operational cleanup.

Event creation additionally accepts optional `registration_mode` (`instant` by default, or `approval`). Approval mode disables guest registration at creation; applications require an account. This is stored atomically with the new event. Existing callers omitting the field retain instant registration.

Azure contract references: [Put Blob permissions](https://learn.microsoft.com/en-us/rest/api/storageservices/put-blob) distinguish create-only from overwrite permission, and [Get Blob Properties](https://learn.microsoft.com/en-us/rest/api/storageservices/get-blob-properties) documents the metadata-only HEAD response and read permission.

## Migration and local operation

Apply the full ordered migration sequence in [deploy/migrations](../../deploy/migrations/), not just one historical migration. Inspect the target database and its migration state before any write; production rehearsal and recovery approval are separate from local setup.

1. The foundational `000002_free_event_workflows.up.sql` migration must precede later workflow migrations. It extends registration enums, adds inventory/policy/revision fields, tickets, audit, idempotency, rate-limit and verification tables, and outbox lease fields. Existing confirmed registrations get real ticket rows.
2. Case-insensitive guest collisions make the migration fail for operator reconciliation; it never deletes/merges records. The inventory CHECK is initially `NOT VALID`, so new writes are protected without silently repairing historical drift. Validate existing rows and then validate the constraint during a controlled deployment.
3. The migration has an explicit forward-only down guard because enum values and persisted ticket/audit records must not be deleted. Rollback requires a reviewed compatible application image or forward repair; never blindly run destructive downs or force a dirty migration version. The local legacy DB was migrated with an atomic psql invocation, without rewriting migration-runner metadata. Reconcile that metadata before using golang-migrate on this local database.
4. `DATABASE_URL` is required. Credentials and local signing material are in ignored `.env`; never copy them into source. Production secrets belong in Key Vault/injected environment. Do not use local development credentials in deployment.
5. `TICKET_SIGNING_SEED` is core-only, base64 encoding of 32 cryptographically random bytes. Keep it stable; rotation/old-public-key retention requires a planned rollout. Missing seed disables signed ticket retrieval while preserving registrations.
6. Run `deploy/scripts/start-local-backends.ps1` to load local environment and start bounded-memory Go processes in hidden windows. It leaves occupied ports unchanged. Go listeners default to loopback outside production; production containers bind internally. Only Nginx exposes ingress. Stop old verified service processes before restarting a changed engine.
7. Organizer-host Nginx routes include RSVP decisions; all authorization remains in Go. Proxy deployment still requires operator validation.

## Historical local verification and current checks

Current automated verification is defined by [CI](../../.github/workflows/ci.yml). Release-candidate CI uses isolated PostgreSQL 16 and Go race detection; inspect the exact candidate run rather than extrapolating from a historical local result. This documentation change does not rerun or certify application tests.

See [failure/recovery scenarios](scenarios.md). Local verification and exact command results are recorded in [WORKLOGS](../planning/WORKLOGS.md). Local PostgreSQL is **18.6**, not the target production 16. Schema-only migration verification rolled back an isolated schema; the additive migration was then applied locally. Existing real event/registration data was retained. Both engines passed normal Go suites against isolated schemas, including the 100-request/10-place capacity case. New checks cover waitlist cancellation/acceptance/expiry, approval allocation, current-state replay, revision conflicts, event cancellation/ownership, one-use check-in, transfers and signed/tampered ticket payloads. These checks exposed and fixed PostgreSQL enum-parameter ambiguity. Race-instrumented execution is blocked by the local CGO/C-compiler setup. Production/PostgreSQL-16 rollout, external sends/uploads, browser interaction, multi-process worker stress and offline scanning remain unverified.



## Owner status controls

The existing core engine now supports owner-scoped guest status edits with server-provided choices, expected guest timestamps and exact idempotent retries. It preserves capacity/held-inventory locks, ticket replacement/revocation, FIFO promotion and audit. Custom optional status messages use the same PostgreSQL outbox and escaped global-worker templates. Checked-in/transferred and closed-event registrations remain protected. See [API.md](../API.md) for targets and the request/response contract. No new service, SQL framework or migration was introduced.

## Sectioned registration and invitation response

Migration 13 extends the existing registration-question records with section breaks, input hints and typed validation JSON, plus event-level page title/introduction/button/confirmation settings. Existing questions remain valid. Up to 40 input questions and 10 section breaks are accepted; submissions store at most 40 answer snapshots, retaining the section label and multiple-choice selections. Form revisions, ownership, the event lock and historical answer retention remain authoritative in Go.

Supported inputs: short/long text, email, HTTPS URL, phone, number, date, time, single choice, multiple choice and confirmation. Go validates required answers, supported choices, numeric ranges/whole numbers, text lengths and valid formats. No custom executable regular expressions or scripts are accepted. The attendee /events/[slug]/register page has identity, organizer-defined sections and review steps; unsectioned legacy questions are split into groups of four. Back/Next preserves entries and final submission validates every section on the backend.

GET /events/by-slug/{slug}/invitation requires an account session. Invitation recognition requires a verified matching email; it never grants an attendee an organizer role or reserves capacity. The response can return the account's existing registration reference. Email invitation actions open registration with an invitation hint; that query string is not authorization. Invited registrations follow the event's approval rules. Owner invitation lists support q and status filters (all, awaiting, responded, queued, sent, failed) and return real registration/check-in state with matching summary counts.

QR check-in uses a lazy-loaded, pinned jsqr 1.4.0 browser decoder, camera permission requested only after Start scanner, and local image selection as a fallback. No camera frames/images are uploaded. Scanning pauses after a result; tracks stop on cancellation, unmount or a hidden tab. The existing signed-pass check-in endpoint remains the authority for event, expiry, ticket state and one-use checks. Manual check-in still requires identity verification and a recorded reason.

Migration 13's down migration preserves columns/history and refuses an incompatible downgrade when expanded questions/answers exist. Review that data before downgrading; do not delete it to make a rollback pass. The UI and Go runtime must support the expanded question kinds once hosts use them.

Decoder reference: [jsQR](https://github.com/cozmo/jsQR); camera lifecycle reference: [MediaDevices.getUserMedia](https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getUserMedia).
