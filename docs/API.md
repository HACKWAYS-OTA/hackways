# HTTP API: Implemented Routes

**Prefix:** `/api/v1`
**Format:** JSON
**Ingress:** Nginx routes requests to one of the two Go backends.

This page documents routes mounted by the current server code. Features described in planning documents are not part of this API until a handler is mounted and implemented.

## Core backend (`core-backend:8081`)

| Method and path | Access | Behavior |
|---|---|---|
| `GET /health` | Public health probe | Reports database readiness; returns 503 when the database is unavailable |
| `GET /api/v1/` | Public | Core service/version metadata; readiness is reported by `/health` |
| `GET /api/v1/events` | Public | Lists published future events; supports `limit`, `offset`, and `q` |
| `GET /api/v1/events/{id}` | Public | Reads a public event by UUID, including previously published cancellations; drafts and unpublished events return 404 |
| `GET /api/v1/events/by-slug/{slug}` | Public | Reads a published or previously published cancelled, non-deleted event by its stored unique slug, including public organizer name, bio, website, and verification state; missing/unpublished events return 404 |
| `POST /api/v1/events/organizer/profile` | Authenticated user | Creates the current user's organizer profile idempotently; returns an existing profile if present |
| `GET /api/v1/events/organizer/profile` | Organizer | Returns the profile associated with the current session |
| `GET /api/v1/events/organizer` | Organizer | Lists only events owned by the current organizer; supports `limit` and `offset` |
| `POST /api/v1/events` | Organizer | Validates and creates a complete draft or published event under the current organizer ID |
| `POST /api/v1/rsvps` | Guest or authenticated | Creates an instant RSVP, application, or explicit waitlist entry under event row lock; workflow policies are enforced in Go |
| `GET /api/v1/rsvps/{id}` | Authenticated owner | Reads an RSVP only when its `user_id` matches the session user |

Guest RSVP creation does not create an account-bound session. A guest cannot use the owner-only detail route until signing in, verifying that email, and linking through POST /api/v1/rsvps/claim. Competing account records are not silently merged. Do not broaden detail access based on possession of an RSVP UUID.

The attendee web application uses `/discover`, `/events/{slug}`, and `/events/{slug}/register`. Root `/` and `/events` redirect to `/discover`. Event pages resolve their stored slug through the Go endpoint; they do not scan the catalog or query PostgreSQL from Next.js. Detail metadata, layouts, and pages share a request-scoped read and do not persist an authorization or availability cache. The registration page posts to the existing RSVP endpoint; a confirmed state appears only for a matching confirmed response, and a pending application is presented separately. The form supplies a random Idempotency-Key for unchanged in-page retries. Guest confirmations are not exposed through a new public lookup URL. Calendar downloads and copied event links are browser presentation of the already-public event data.

When the attendee session cookie is present, RSVP creation validates it and binds the RSVP and confirmation recipient to the account. An invalid supplied session returns `401` instead of silently downgrading to guest identity. Without a session, the handler requires a valid guest email. Both paths enforce the event's `registration_opens_at` and `registration_closes_at` values while holding the event row lock.

## Global backend (`global-backend:8082`)

| Method and path | Access | Behavior |
|---|---|---|
| `GET /health` | Public health probe | Reports database readiness; returns 503 when the database is unavailable |
| `GET /api/v1/` | Public | Global service/version metadata; readiness is reported by `/health` |
| `POST /api/v1/auth/register` | Public | Creates account and session cookie |
| `POST /api/v1/auth/login` | Public | Authenticates and issues session cookie |
| `POST /api/v1/auth/logout` | Cookie | Revokes the provided session if present and clears session cookies |
| `POST /api/v1/auth/verify` | Cookie or token body | Validates a session and returns user identity |
| `GET /api/v1/auth/me` | Authenticated | Returns the current user |
| `POST /api/v1/storage/upload-intent` | Authenticated | Issues an Azure Blob upload SAS from the request metadata |
| `POST /api/v1/storage/event-banner/complete` | Authenticated event owner | Verifies Azure object properties and attaches a banner to the matching editable event revision |

Upload intent checks event ownership and uses server-controlled user/event namespaces. `event-banners` requests require `resource_id` and grant create-only object access; `avatars` bind to the session user. Missing Azure credentials return 503. Banner completion accepts `{resource_id,blob_key,revision}`, reads Azure HEAD properties with a private read SAS, enforces the owned namespace, block-blob type, allowed image MIME/extension and 5 MB bound, then attaches the public URL only to an owned matching editable event revision. Media bytes are never proxied through Go. Actual Azure/CORS/public-read behavior is unverified locally. Event creation accepts optional `registration_mode: instant|approval`; approval requires account-backed registrations.

## Spaces, appearance and public references

### Event management dashboard extensions

- `GET /events/organizer`: adds optional `q` (literal title substring, maximum 200 characters) and `view=all|upcoming|drafts|past|cancelled`; existing pagination and organizer ownership remain.
- `GET /events/{id}/summary`: event owner only; real total/confirmed/pending/waitlisted/offered/cancelled/rejected/expired/transferred/checked_in counts. Scanned tickets count only with a confirmed registration.
- `GET /events/{id}/registrations`: adds `q` (name/email/registration-code substring, maximum 200 characters) and `status`, including `all`, registration states, `checked_in` and `not_arrived`. Invalid filters return 400. Pages remain bounded at 100 rows. Dashboard pages request one extra row to determine whether a next page exists.
- `PUT /events/{id}/appearance`: owner only; accepts theme/color/font/style/display, `revision`, and optional `remove_banner`. The core transaction locks the event with the existing two-second timeout, checks future/non-cancelled state and revision, updates appearance and writes audit history without changing inventory.
- Existing banner completion now attaches verified media to owned future draft/unpublished/published events. It still enforces object namespace, Azure HEAD properties, type/size, transaction timeout and revision; immediate identical completion retries remain safe. Publication still belongs to core-backend. Started/cancelled events reject attachment. Actual Azure behavior remains unverified locally without credentials.

Core-backend owns `/api/v1/spaces`: public directory (`GET /`, paginated/searchable), slug detail (`GET /by-slug/{slug}`), public event feed (`GET /{id}/events`), authenticated owned/joined directory (`GET /mine`), membership read/join/leave, and owner-only create/manage/edit/status operations. Creation requires an `Idempotency-Key`; changed-payload replay returns 409. Members gain no event management rights. Archived Spaces reject new members, event associations and publication while retaining existing published events. Owners cannot leave their Space.

Event creation accepts `space_slug`, `visibility: public|private`, `theme`, `theme_color`, `theme_font`, `theme_style`, and `theme_display: light|dark`. A private published event is link-only: slug details and registration work, but Discover/search/public Space feeds exclude it. This is not an email allowlist or outbound invitation service. Publication state is independent of discoverability. Owner updates may change visibility with the current revision. Optional `faq_text`, HTTPS `contact_url` and `rewards` persist in the same create/edit transaction; omitted fields are preserved and empty strings clear them.

`PUT /events/{id}/space` accepts `{space_slug,revision}`; an empty slug detaches. Go ownership checks and a composite foreign key prevent cross-organizer associations. `GET /events/manage-by-slug/{slug}` is owner-scoped. Browser event/management links use slugs. Registration pages use `R-` codes through owner-scoped `/rsvps/by-code/{code}`. Ticket responses add `T-` codes; check-in remains event/owner-scoped and one-use. Existing UUID API transport identifiers and signed ticket contracts remain compatible; new visible browser navigation and references use slugs/codes.

## Session cookies

The attendee cookie is `__Host-hackways_session` with `Secure`, `HttpOnly`, `Path=/`, and `SameSite=Lax`. The admin cookie name is reserved in code as `__Host-hackways_admin_session` and uses `SameSite=Strict`. The `__Host-` prefix requires HTTPS in browsers. Production cookie behavior must be checked through the HTTPS ingress, not assumed from unit-level construction.

## Response and errors

Successful handlers return JSON. Errors use the backend's standard envelope with a machine-readable code and user-facing message. Exact status and response fields are defined at each handler; older examples in planning documents are not normative.

## Not implemented in the mounted route set

The free-event lifecycle, organizer roster/audit/check-in, registration cancellation/notes/approval/waitlist/transfer/claim/tickets, session management, email verification and notification status/retry routes are mounted. See [the detailed API contract and implementation boundaries](backend/README.md). Payments are deferred by the user. Admin moderation, account recovery/deletion, custom forms and offline scanner workflows remain unimplemented; do not infer them from the product brief.

## Registration forms and private guest review

| Method and path | Access | Behavior |
| --- | --- | --- |
| `GET /api/v1/events/by-slug/{slug}/registration-form` | Public event | Current form revision, name requirement and active ordered questions; no answers; no-store |
| `GET /api/v1/events/{id}/registration-form` | Event owner | Current owner-editable configuration, including draft events |
| `PUT /api/v1/events/{id}/registration-form` | Event owner | Replace active question order/configuration under event row lock; revision conflict returns 409; started/cancelled events reject edits |
| `GET /api/v1/events/{id}/registrations/by-code/{code}` | Event owner | Actual guest identity/state/ticket facts and latest submitted answer snapshots; cross-owner or event mismatch returns 404 |

The form payload is `{revision,name_required,questions:[{public_code,label,description,kind,required,options}]}`. New questions send an empty public_code; the server returns a persisted Q-code. Kinds are text, textarea, url, number, checkbox, select and phone. Maximum 20 active questions; choice questions require 2–20 unique options. Omitted questions are deactivated, preserving historical responses. Form edits increment both form and event revisions. Reads obtain question metadata and the form revision in one SQL statement.

`POST /api/v1/rsvps` additionally accepts `form_revision` and `answers:[{code,text,checked}]`. Go validates current required answers, types, codes, lengths and choices before inventory changes, inside the existing locked transaction. FORM_CHANGED returns 409; invalid answers return 422. Successful exact idempotent retries replay their committed result even if the organizer subsequently edits questions. Legacy requests remain compatible when the event has no custom questions or name requirement; old canonical idempotency hashes are preserved when the new fields are absent.

Submitted responses are append-only snapshots, stored transactionally with registration and its outbox entry. Guest review returns snapshot labels and types rather than relabeling older answers. Transfers do not copy the original attendee's responses to the recipient. No public event/catalog endpoint exposes responses. Browser management URLs use event slugs and `?guest=R-...`; UUIDs stay internal transport identifiers.
## Owner guest-status editor

- `GET /api/v1/events/{id}/registrations/by-code/{code}` additionally returns the guest `updated_at` and `status_options:[{value,label}]`. Options are backend presentation advice and are rechecked transactionally at mutation time.
- `POST /api/v1/events/{id}/registrations/by-code/{code}/status` requires the owning organizer session and a random `Idempotency-Key` (16-128 printable characters). JSON: `{status,updated_at,notify_guest,message}`. `notify_guest` is an explicit boolean; message is optional, trimmed, limited to 2000 Unicode characters, and rejected when notifications are off. Response: `{data,notification_queued}`. An exact replay sets `Idempotency-Replayed: true`; a changed payload under the same key returns 409.
- Allowed targets are `confirmed` (Going), `cancelled` (Not going), and `rejected` (Declined), constrained by current state. Pending applications can be confirmed or declined; active registrations can be cancelled; cancelled/expired/rejected registrations can be restored while registration is open. Active reserved offers can be confirmed. Waitlisted guests cannot skip the offer queue. Checked-in, transferred, started and non-published-event registrations cannot be changed through this editor. Options may be empty.
- Status writes acquire the existing event row lock and local 2-second statement timeout, compare the current guest timestamp, check actual capacity including held offers, maintain inventory/tickets, audit the actor, optionally enqueue the existing status-email template with custom text, and store the idempotent result in one transaction. Cancellation promotes the FIFO waitlist through the existing engine. Conflict codes include `REGISTRATION_CHANGED`, `REGISTRATION_EXPIRED`, `CAPACITY_FULL`, `STATUS_CHANGE_UNAVAILABLE` and `IDEMPOTENCY_CONFLICT`.
- The optional message is persisted in the transactional outbox and HTML-escaped by global-backend. Notification queued means persisted for the existing ACS worker, never proof of delivery. Disabling notifications suppresses this guest's status email; normal notifications to other guests offered a freed place retain their existing behavior. No schema migration is required.
