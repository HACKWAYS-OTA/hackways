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

The attendee cookie is `__Host-hackways_session` with `Secure`, `HttpOnly`, `Path=/`, and `SameSite=Lax`. Dedicated admin sign-in uses `__Host-hackways_admin_session` with `SameSite=Strict`. The `__Host-` prefix requires HTTPS in browsers. Production cookie behavior must be checked through the HTTPS ingress, not assumed from unit-level construction.

## Response and errors

Successful handlers return JSON. Errors use the backend's standard envelope with a machine-readable code and user-facing message. Exact status and response fields are defined at each handler; older examples in planning documents are not normative.

## Not implemented in the mounted route set

The free-event lifecycle, organizer roster/audit/check-in, registration cancellation/notes/approval/waitlist/transfer/claim/tickets, session management, email verification and notification status/retry routes are mounted. See [the detailed API contract and implementation boundaries](backend/README.md). Payments are deferred by the user. Independent moderation reports/appeals, account recovery/deletion, custom forms and offline scanner workflows remain unimplemented; do not infer them from the product brief. The admin publication/actions extension is described below.

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


## Admin extension (release preparation; not deployed)

The user has approved deploying the accumulated shared sign-in, admin record/actions and email-renderer work. The revised release candidate, image publication, production migrations and activation remain pending; this section describes the current contract, not a completed deployment.

The two existing Go engines mount protected administration collections. Core serves `GET /api/v1/admin/core/events`, `/spaces`, `/registrations`, `/tickets`, `/moderation` and `/audit`; Global serves `GET /api/v1/admin/global/people`, `/messaging` and `/storage`. Each accepts `q` (maximum 200 characters), `status`, `limit` (1-100, default 20), and `offset` (nonnegative int32), returning `{data:[...]}`. An invalid search/page returns 400; missing/expired admin authentication returns 401 and a missing/revoked grant returns 403. Private rows are never returned through ordinary attendee sessions. `moderation` lists only drafts/unpublished events, and storage returns attached media metadata with URL queries/fragments removed.

Global adds `POST /api/v1/auth/admin/login` with `{email,password}`, `POST /api/v1/auth/admin/logout` and guarded `GET /api/v1/auth/admin/me`. Login requires a verified account and explicit database grant; it creates a separate, marker-backed 30-minute Strict/Secure/HttpOnly host-only admin session. Response identity contains display name, never cookie/token values. Migration 000009 supplies the grant/session-marker tables and provisions no users or privileges. Production migration, grants and Nginx routing were not changed. See [the admin pages, scope and verification limits](planning/ADMIN_DASHBOARD.md).

The admin identity endpoint returns 401 for missing/invalid dedicated authentication and 403 for authenticated accounts without an admin grant or dedicated session marker. If only an ordinary cookie is present, Go may check that account's grant to distinguish denial from the need for dedicated admin sign-in; this never grants admin access or returns identity success. Backend/database failures remain failures, not sign-out or success. Admin web gates documents using these responses before rendering and keeps all admin APIs independently guarded.

### Record details

`GET /api/v1/admin/{engine}/{resource}/{id}` supports all nine collection resources on their existing engine. The admin UI uses positive numeric references for record routes; audit keeps its existing positive integer reference. Storage requires `kind=event_banner|avatar` and the reference of its attached event/user. Both Go engines resolve numeric references to internal UUIDs through sqlc before the existing authoritative reads/workflows; UUID API inputs remain compatible. Invalid references, sections or asset kinds return 400; missing records return 404.

The response is `{data:{id,reference,...},related:{tabs:[{key,label}],section,rows:[{id,reference,label,detail,resource,status,occurred_at}],offset,has_next}}`. Record `reference` is numeric; related `reference` is a positive decimal string. Internal UUID IDs remain in DTOs for compatibility, but the admin UI renders numeric references and creates numeric record/related links, not UUID fields or detail-page URLs. Detail DTOs select permitted fields through sqlc and exclude passwords, credential/session tokens, email bodies and provider errors. Asset URLs strip query strings/fragments. Audit detail includes recorded metadata and has no write endpoint.

Related sections use nonnegative `offset`, fetch 21 rows and return at most 20 with `has_next`; the detail page size is fixed independently of collection `limit`. Events/moderation offer `registrations`, `tickets`, `activity`; spaces offer `events`, `members`, `activity`; people offer `events`, `registrations`, `sessions`, `messages`, `activity`. Other non-audit details offer `activity`; audit returns empty related arrays. Omitted `section` selects the first tab. Active person sessions expose record IDs and bounded user-agent text, never session credentials.

Admin web renders `/{resource}/details/{reference}` from a server fetch with a six-second timeout, no cached data and safe DTO validation. Existing UUID detail-page links redirect to the corresponding numeric route. `section`, `offset` and storage `kind` remain in clean query paths; related tabs/pagination retain scroll-disabled links. Signed-out and forbidden requests leave the console; missing records render 404 and fetch failures provide a retry path. Details provide a Back button; filters apply immediately through query links without Apply/dropdown controls, action choices use the shared RadioGroup, and no Refresh button is presented. Catalogue loading shows the animated Hackways wordmark/line with a static reduced-motion fallback.

### Reviewed record actions

All paths below are prefixed by `/api/v1/admin/`. The UI supplies numeric record/session references in the `{id}`/`{sessionID}` positions; existing UUID API callers remain compatible. Each write requires the dedicated admin session/current grant, JSON, an allowed admin `Origin`, `X-Admin-Reason` containing a percent-encoded 3-1000 character reason, and a random printable `Idempotency-Key` of 16-128 non-whitespace ASCII characters. Production origin is `https://admin.hackways.in`; development additionally permits localhost/127.0.0.1 on port 3003. The console currently limits its reason field to 500 characters.

| Method and suffix | JSON fields | Persisted action and constraints |
| --- | --- | --- |
| `PUT core/events/{id}` | `title,description,capacity,visibility,revision` | Edit future, non-cancelled event; nullable capacity means unlimited, otherwise at least confirmed plus held places; visibility is `public` or `private` |
| `POST core/events/{id}/status` | `status,revision,acknowledge_attendees` | Existing publish/unpublish/cancel lifecycle; targets `published`, `unpublished`, `cancelled`; backend transition/acknowledgement rules still apply |
| `PUT core/spaces/{id}` | `name,description,guidelines,website_url,revision` | Existing validated space edit |
| `POST core/spaces/{id}/status` | `status,revision` | Archive/restore with `archived` or `active` |
| `POST core/registrations/{id}/status` | `status,updated_at,notify_guest,message` | Existing guest status workflow: `confirmed`, `rejected` or `cancelled`; explicit notification boolean and optional message up to 2000 characters; current state, queue, capacity and check-in constraints remain authoritative |
| `POST core/tickets/{id}/check-in` | `{}` | Existing manual check-in for the route's ticket; Go validates event/registration/ticket state and check-in window under lock |
| `POST global/people/{id}/edit` | `display_name,updated_at` | Trimmed display name of 1-100 characters |
| `POST global/people/{id}/access` | `access,updated_at` | `admin` or `member`; verified target required to grant, no self access changes or removal of the last active admin; removal revokes dedicated admin sessions |
| `POST global/people/{id}/sessions/{sessionID}/revoke` | `updated_at` | Revoke only a session belonging to that person; current acting session must use sign-out |
| `POST global/messaging/{id}/retry` | `updated_at` | Requeue only a failed message with `first_sent_at` still null; resets retry scheduling, never directly sends mail |
| `POST global/storage/{id}/detach` | `kind,updated_at` | Detach `event_banner` or `avatar` metadata only; retain the blob; banner detachment also increments the event revision |

Use the exact current `revision` or timestamp `updated_at` from detail reads. A stale record/version returns a conflict rather than overwriting newer changes. Core mutations reuse event/space locks and existing lifecycle/inventory rules; event inventory transactions retain `statement_timeout='2000ms'`. Go preserves the administrator as audit actor while scoping delegated management to the resource owner. Mutation, audit and replay receipt commit together. Global actions serialize role changes and action receipts, then recheck the live dedicated session/grant inside the transaction.

Successful global actions return `{data:{id,action}}`; core actions return their workflow data. Exact successful retries return the stored result with `Idempotency-Replayed: true`; changed requests under the same scoped key return 409 `IDEMPOTENCY_CONFLICT`. Reuse the same key, body and reason after an uncertain response. Audit rows remain read-only through these APIs. Queued notifications/retries are not evidence of provider acceptance or inbox delivery.

Migration `000010_admin_detail_indexes` adds reversible entity/actor audit and recipient-message timeline indexes; it adds no privileges or seeded data. It has been applied only to isolated local test databases and the restored local preview, not production.

Migration `000011_admin_record_references` adds `admin_reference BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL UNIQUE` to users, events, spaces, RSVPs, tickets, email outbox and sessions. Each table owns its sequence; resource/kind disambiguates references. UUID keys, foreign keys, ownership and state rules remain unchanged. Its down migration deliberately refuses column removal to preserve saved numeric links: roll back application images while retaining the additive schema. Migrations 000009, 000010 and 000011 have not been applied to production at this documentation handoff.

### Isolated real-data preview runtime

With `ADMIN_PREVIEW_ONLY=true`, configuration rejects production mode, non-loopback listeners/database hosts, and any database name other than `hackways_admin_preview`. Core exposes its admin collection/detail/action routes and health/service metadata; Global exposes its admin routes, `/auth/admin/{login,logout,me}` and health/service metadata. Approved admin actions persist only in this restored local copy. Ordinary application mutation routes are absent and all background maintenance/email workers remain disabled, so an outbox retry queues without sending. Production backends remain unchanged; the dedicated local admin cookie and explicit local grant are still required.

## QR passes and notification retries

The authenticated RSVP ticket endpoint adds `qr_image`, a PNG data URL containing the signed admission code. It retains existing machine contracts (`code`, `short_code`, public key and algorithm); the UI displays the QR and human pass reference. Confirmed registration pages fetch the pass on the server. A persistent 32-byte base64 `TICKET_SIGNING_SEED` must be injected into core-backend. QR images and responses are private and not cached. Existing event/owner/one-use checks remain authoritative at check-in.

The owner registration-status endpoint accepts optional `resend_notification: true` with current `status`, `updated_at`, `notify_guest: true`, optional message and an Idempotency-Key. This queues one fresh current-status notification without altering registration, ticket or capacity. Resends use a separate idempotency scope and audit action `registration.notification_requested`; cross-owner, stale and closed-event requests fail. The response reports queued state, not inbox delivery.

Registration notifications include the actual event banner, host/Space context and a canonical registration action URL. `APP_PUBLIC_URL` can override the canonical origin for an approved environment.

Owner and admin registration controls support the five manual states: `confirmed` (Going), `pending_approval` (Awaiting approval), `waitlisted` (Waiting list), `rejected` (Declined), and `cancelled` (Cancelled). The current state is omitted from change choices. Checked-in/transferred and closed-event restrictions remain. Pending review receives an expiry bounded by seven days, event start and any future registration closing time. Moving out of confirmed/offered releases confirmed/held capacity and revokes admission; confirming issues admission under the existing lock and capacity checks. Existing automatic waiting-list offers remain; if that policy immediately promotes a manual waitlist entry, the response returns the actual `offered` state and the UI explains it. No manual timed-offer status was added.
