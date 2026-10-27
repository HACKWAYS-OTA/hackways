# First-party analytics

## Admin reports

The default console view uses `GET /api/v1/admin/{core,global}/analytics/live`.
`window=5m` (default), `1h`, and `24h` select exact sliding intervals; other values
return 400. PostgreSQL `transaction_timestamp()` defines each engine's interval
inside a read-only repeatable-read transaction. The engines retain independent
snapshots. Core activity has 20 buckets for 5m/1h and 24 for 24h, including real
zero counts, with the same bounds as the totals. Existing indexed queries provide
registration states, check-ins, event creation, new accounts, provider acceptance
and consented traffic. No schema migration is required.

The page loads both engines concurrently on the server. While visible, it refreshes
15 seconds after the preceding request completes; hidden tabs stop polling and
refresh on return. Interval changes cancel older requests and clear prior figures.
Failed sections clear their figures and offer retry; missing data never becomes
zero. Existing admin session, grant and daily OTP checks apply to both endpoints.
Responses remain private/no-store, with 5-second requests and 2-second statements.
`captured_at` is transport metadata only: the interface has no update timestamp.
Traffic is labeled as consented visitors, not online users. A collection-budget
warning appears when a selected day reached its ingestion limit. History and its
date-range exports remain available separately under `/analytics?view=history`.

`GET /api/v1/admin/core/analytics` reports registrations, current registration states,
events created, current ticket check-in timestamps, daily activity and ten leading
events. `GET /api/v1/admin/global/analytics` reports accounts, provider-accepted and
currently failed email, current subscribed contacts and consented visitor activity.
Both require the existing dedicated admin session, database grant and daily OTP.

Optional `from` and `to` are inclusive UTC dates within the latest ninety days;
the default is thirty days. `format=csv` exports daily figures. Each engine uses
a read-only repeatable-read transaction with bounded statement/request timeouts.
The admin page requests the engines concurrently and reports partial failures
explicitly. Responses are private/no-store. Dates are dense in daily exports.

Business totals use actual PostgreSQL records, including registration updates and
transfers. They are not immutable submission history. Current registration states
can change after the selected period. Reissuing a ticket can clear its arrival
timestamp. Email provider acceptance is not inbox delivery. There is no fabricated
comparison period, historical backfill or inferred conversion rate.

## Optional visitor collection

`POST /api/v1/analytics/events` is handled by global-backend through the main Nginx
vhost. It accepts only the production website origin (localhost in development),
JSON up to 8 KiB, affirmative consent, random browser/visit UUIDs and one to ten
events. Allowed events are `page_view`, `registration_step` and
`registration_complete`. All current non-admin pages are covered: public marketing,
programs (including ambassadors), cities, discovery, account/authentication,
registration and managed Space/event workflows. Redirect-only pages do not emit a
view before the destination renders. The collector accepts only the reviewed route
catalog, not arbitrary URL paths. Query strings, fragments, unknown fields,
answers and account identifiers are rejected. Admin-console origins and routes
remain excluded.

Events contain a random idempotency UUID, approved path, viewport category,
referring hostname and optional registration step. The service stores purpose-
separated HMAC-SHA256 browser/visit identifiers using the existing session secret.
It does not store raw browser identifiers, account identity, IP addresses or full
referrers in analytics tables. Existing infrastructure logs remain separate.
Public ingestion is untrusted and best effort; figures do not prove human visits.

The shared `@hackways/ui/analytics` component is mounted by public/core/organizer
web, never admin-web. Organizer paths are prefixed with `/manage`. Private dynamic
routes are normalized before collection: `/my-events/:registration`,
`/spaces/:space/settings`, `/programs/:program/application`,
`/programs/:program/apply/:step`, `/manage/events/:event` and its approved workflow
suffixes, and `/manage/spaces/:space` (including `/team`). Public entity slugs
remain available for public page comparisons. Private IDs, ticket codes,
application steps/answers and authentication tokens never enter stored routes.
Go independently validates this catalog and rejects raw private identifiers.
When adding a new page, update the shared normalization and Go route catalog
together. Navigation between different resources counts a new page view even
when both normalize to the same private template; window focus does not.

Before this coverage change, the allowlist excluded programs, cities and private
workflows. Those missing historical views cannot be reconstructed as exact
consented analytics and are not backfilled. Reports continue to represent only
recorded consented visits; blockers, declined consent and collection limits can
make them lower than actual traffic.
Collection and browser identifiers start only after Allow analytics. Footer Cookie
preferences can switch to Essential only, stop queued collection and clear the
analytics identifiers without touching sign-in. Choices last one year, browser
identifiers ninety days refreshed during use, visit identifiers thirty minutes of
inactivity. Requests have a ten-second browser timeout and one bounded retry with
the same event IDs; there is no background polling. Queue/batch limits are 20/10.

Migration 22 adds indexed analytics tables and report indices. An atomic PostgreSQL
daily budget limits reservations to 25,000 events; duplicate retries consume budget
but cannot duplicate stored events. Reaching the budget returns 429 and pauses
the browser collector for an hour. Existing housekeeping deletes up to 500 expired
events each minute after ninety days. It also removes expired quota records.
These are storage/ingestion bounds, not measured throughput guarantees.

## Campaign recipients and account preferences

`GET /api/v1/admin/global/campaigns/recipients?source=all|account|import&q=prefix`
returns up to twenty eligible subscribers plus a `more` flag. It requires the same
admin authorization. It never exposes unsubscribe credentials. Configuration now
includes `recipient_mode: all|selected`; selected mode requires 1-200 valid email
addresses. Empty selected mode is rejected rather than interpreted as everyone.
Legacy drafts derive their mode from their saved recipient list.

Source, selection and subscribed/verified eligibility are applied by generated SQL
to preview, count and queue operations. The worker rechecks eligibility before
provider submission. Existing campaign revision/idempotency/audience-count checks
and scheduling remain in place. This release does not send a campaign.

New email accounts default `promotional_email` to true; the signup checkbox can
submit false. A newly created Google account gets the same default. The preference
is written in the identity transaction. Existing identities and unsubscribes are
not resubscribed by sign-in or default initialization. Existing users are not
backfilled. Promotional messages retain their unsubscribe link and configured
updates sender; account/event senders remain unchanged.

## Activation and rollback

Build images externally through mandatory CI. Apply additive migration 22 before
activating the Go engines, then the affected frontends and the reviewed main-vhost
analytics location. Preserve the current protected environment, session/signing
identity, admin ingress policy and independent service boundaries. Keep a private
database backup and previous immutable runtime manifest. Runtime rollback can
retain the additive schema; the down migration deliberately refuses to discard
collected analytics. CDN activation is separately deferred by the operator.


## Event management and arrival corrections

Migration 23 adds `events.management_closed`, default false. Existing current
administrator cancellations/unpublishes are initialized from the most recent
recorded lifecycle audit action. Administrator cancellation/unpublish closes
owner metadata editing; administrator publication clears that closure. Owners
cannot clear it through details, appearance, questions, Space, banner or status
requests. Ownership, revision checks and transactional event locking remain in Go.
Starting or owner-cancelling an event no longer closes its metadata editor. A past
start can be retained while correcting details; changing the start still requires
a future date. Cancellation remains terminal and never restores tickets or RSVPs.
Registration/invitation/check-in eligibility windows remain separate.

Successful `POST /events/{id}/check-in` retains its ticket `data` and adds an
`attendee` object: ticket ID, name, email, registration code and checked-in time.
Repeat admission returns 409 `ALREADY_CHECKED_IN` with the same scoped attendee in
error details. Private information is returned only after organizer authorization
and event scoping; responses are private/no-store.

`POST /events/{id}/check-in/undo` requires the ticket ID and exact recorded
`checked_in_at`. It runs under the same event lock, changes a confirmed scanned
ticket back to issued, clears its arrival fields and records actor/time in audit.
An already-undone retry succeeds without another audit entry; an old timestamp
cannot undo a subsequent admission. Cancelled, revoked, unrelated or administrator-
closed admissions are not eligible. Corrections remain available after the event
ends while it is published. Scanner, roster and guest details share the action.
