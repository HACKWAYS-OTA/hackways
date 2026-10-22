# Admin promotional campaigns

## Sender routing

| Mail purpose | Sender |
|---|---|
| Account verification, recovery, security and account welcome | `noreply@mail.hackways.in` |
| Event updates, invitations, registration and ticket messages | `notify@mail.hackways.in` |
| Admin promotional campaigns | `updates@mail.hackways.in` |

The existing verified `mail.hackways.in` domain in the India email service retains noreply and notify; updates was created in Azure on 2026-10-22. All messages use the existing ACS endpoint, transactional PostgreSQL outbox and bounded two-send worker. No SMTP or external queue was added. Sender selection is assigned by the Go template catalog, never by a browser-supplied From field. `ACS_SENDER_ADDRESS`, `ACS_TRANSACTIONAL_SENDER_ADDRESS` and `ACS_PROMOTIONAL_SENDER_ADDRESS` configure these identities. An unconfigured promotional sender fails rather than falling back to another address.

## Admin workflow

`/campaigns` lists real saved campaigns. `/campaigns/new` creates a draft; `/campaigns/details/{numeric-reference}` edits and previews the saved version, sends a test only to the current verified admin email, reviews the audience, launches and cancels. The interface reuses platform admin authentication, network isolation, Hugeicons and shared controls. Creation has a stable idempotency key; edits use revisions; send/cancel replays cannot enqueue a duplicate campaign. Every mutation records the admin and reason in the existing audit log. Saved draft content is plain text with optional HTTPS action button; Go renders escaped HTML using the production renderer.

Campaigns are bounded to 1,000 recipients per send. Sending requires explicit confirmation and a matching reviewed audience count. A single database transaction locks the campaign, snapshots eligible contacts, enqueues outbox records and recipient unsubscribe proofs, and marks the campaign queued. A changed count rolls back. Before provider dispatch, current campaign state and subscription are checked again. Cancellation/unsubscribe suppress future attempts; messages already accepted or in flight cannot be recalled. No campaign audience was launched during implementation verification.

Delivery history distinguishes queued, provider accepted and failed/suppressed states; provider acceptance does not prove inbox delivery. Pagination is 20 records. While visible and pending, the detail page refreshes on a 15-second bounded interval, pauses in hidden tabs and stops after 120 cycles; it does not overlap transitions. The existing worker handles bounded provider retries. Existing ambiguous delivery attempts must not be blindly retried.

## Audience and consent

Account holders can explicitly enable/disable promotional email from `/account`; enabling requires a verified email. No existing user is subscribed by migration. `/campaigns/contacts` imports CSV with `email,name` (name optional), at most 200 contacts and 64 KB per import. The admin must attest explicit opt-in and record the consent source and date. The entire file is validated before persistence. Existing contacts are preserved, including unsubscribes; re-import never silently resubscribes someone. Imported opt-in contacts may be external to Hackways accounts. Account-backed contacts require current verification and matching email.

Every delivered promotional email has a random recipient-specific unsubscribe link. Only its digest is stored in the delivery index; the private outbox necessarily holds the email URL. `/auth/unsubscribe` captures/removes the URL key, requires a confirmation click and posts to the Go API. Link scanners cannot unsubscribe someone just by issuing a GET. The endpoint is idempotent for a valid link and never displays an email address. Unknown links fail explicitly. Unsubscribe does not affect account/security or event messages. Test messages do not opt the admin in.

## API and storage

All admin routes are under `/api/v1/admin/global`, require the dedicated admin session and are protected by the existing Nginx IP/VPN gate:

- `GET/POST /campaigns`
- `GET /campaigns/{reference}`
- `POST /campaigns/{reference}/edit|test|send|cancel`
- `GET /campaign-contacts`
- `POST /campaign-contacts/import`

Mutations require an approved Origin and `X-Admin-Reason`; creation also requires `Idempotency-Key`. Account APIs are `GET/POST /api/v1/auth/promotions/preference`; public confirmation is `POST /api/v1/auth/promotions/unsubscribe` with `{key}`. There is no unauthenticated campaign or recipient browsing endpoint.

Migration 15 adds contacts/consent, campaigns and per-recipient delivery references. Its down migration refuses to remove populated history; roll back runtime while retaining schema when populated. SQL is generated through sqlc for both Go engines. No production schema migration or application deployment has been performed for this work.

## Review and verification

Local management remains `http://localhost:3102/manage/login`; campaign admin is `http://localhost:3103/login?next=%2Fcampaigns`; main auth/account preview is `http://localhost:3101/login`. These use a separate local production-data copy. Email remains paused in that copy to prevent dispatching historical messages. The original local previews are retained.

Mandatory isolated PostgreSQL checks cover admin/origin protection, idempotent draft creation/test/send, revision and audience conflicts, opted-in import, cancellation, unsubscribe and re-import suppression. Sender routing is checked at the actual ACS payload boundary. Real authorized India ACS checks accepted account messages from noreply and an admin test from updates; earlier authorized checks verified notify. The first promotional verification request was correctly rejected by an incomplete test CORS allowlist; the isolated check configuration was corrected and the real admin test succeeded. Production recipient/account records were not changed.

No connected browser is available, so interactive desktop/mobile layout and keyboard review remain pending. Linux PostgreSQL 16/race CI passed on source commit dbd9aa033329e38de0cf08c5c5a89d2a91512945 (run 39779987414): 61 core and 92 global results, zero failures/skips. User review and the eventual full release build/deployment checks remain required before application deployment. Production maintenance was removed at the user's subsequent request; the existing released application is online. New application changes remain held for local review.
