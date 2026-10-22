# Admin promotional campaigns

## Sender routing

| Mail purpose | Sender | Reply-To |
|---|---|---|
| Account verification, recovery, security and account welcome | `noreply@mail.hackways.in` | Omitted |
| Event updates, invitations, registration and ticket messages | `notify@mail.hackways.in` | `hello@hackways.in` |
| Admin promotional campaigns | `updates@mail.hackways.in` | `hello@hackways.in` |

The existing verified `mail.hackways.in` domain in the India email service retains noreply and notify; updates was created in Azure on 2026-10-22. All messages use the existing ACS endpoint, transactional PostgreSQL outbox and bounded two-send worker. No SMTP or external queue was added. Sender selection is assigned by the Go template catalog, never by a browser-supplied From field. `ACS_SENDER_ADDRESS`, `ACS_TRANSACTIONAL_SENDER_ADDRESS` and `ACS_PROMOTIONAL_SENDER_ADDRESS` configure these identities. An unconfigured promotional sender fails rather than falling back to another address.

Go applies the reply policy when serializing the [ACS send request](https://learn.microsoft.com/en-us/rest/api/communication/email/email/send?view=rest-communication-email-2023-03-31): account/security messages omit `replyTo` entirely; event, promotional and legacy event messages set exactly one address, `hello@hackways.in`. It is not an editable campaign field. Omitting Reply-To does not disable a mail client's Reply button; clients can still address the From mailbox. Reconciled messages already accepted by ACS retain their original headers.

## Admin workflow

`/campaigns` lists real saved campaigns. `/campaigns/new` creates a draft; `/campaigns/details/{numeric-reference}` edits and previews the saved version, sends a test only to the current verified admin email, reviews the audience, launches and cancels. The interface reuses platform admin authentication, network isolation, Hugeicons and shared controls. Creation has a stable idempotency key; edits use revisions; send/cancel replays cannot enqueue a duplicate campaign. Every mutation records the acting administrator and action in the audit log. Draft creation/editing and campaign actions no longer ask for a typed reason; supplied legacy reasons are still validated. The editor saves plain-text content and structured heading/paragraph/image/button/divider blocks. Go validates the bounded design configuration and escapes all content before producing email HTML.

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

Mutations require an approved Origin; an omitted reason receives an automatic action description. Creation also requires `Idempotency-Key`. Account APIs are `GET/POST /api/v1/auth/promotions/preference`; public confirmation is `POST /api/v1/auth/promotions/unsubscribe` with `{key}`. There is no unauthenticated campaign or recipient browsing endpoint.

Migration 15 adds contacts/consent, campaigns and per-recipient delivery references. Its down migration refuses to remove populated history; roll back runtime while retaining schema when populated. SQL is generated through sqlc for both Go engines. Production is now on clean schema 19 and source `dab8d1cd545b843bf20c2d8cdac16bb7feb058d8`, including the consent and campaign schema. No audience campaign was launched during deployment verification.

## Review and verification

Local management remains `http://localhost:3102/manage/login`; campaign admin is `http://localhost:3103/login?next=%2Fcampaigns`; main auth/account preview is `http://localhost:3101/login`. These use a separate local production-data copy. Only daily admin verification mail is enabled in that copy; other historical templates remain unleased. Production has no template filter. The original local previews are retained.

Mandatory isolated PostgreSQL checks cover admin/origin protection, idempotent draft creation/test/send, revision and audience conflicts, opted-in import, cancellation, unsubscribe and re-import suppression. Sender routing is checked at the actual ACS payload boundary. Real authorized India ACS checks accepted account messages from noreply and an admin test from updates; earlier authorized checks verified notify. The first promotional verification request was correctly rejected by an incomplete test CORS allowlist; the isolated check configuration was corrected and the real admin test succeeded. Production recipient/account records were not changed.

No connected browser is available, so interactive desktop/mobile layout and keyboard review remain pending. Linux PostgreSQL 16/race CI and all six image builds passed for release source `dab8d1cd545b843bf20c2d8cdac16bb7feb058d8` (run 39795523731): 62 core and 122 global results, zero failures/skips. The user authorized deployment; production activation and live HTTPS/auth-gate checks passed. Sender configuration is noreply for account/security, notify for event mail and updates for promotions. Reply-To is omitted for account/security and set to hello@hackways.in for event/promotional mail. Interactive review remains a separate verification limit.


## Campaign customization and scheduling (candidate, schema 21)

The editor has Content, Design, Recipients and Delivery sections plus an HTML preview from the same Go renderer used by the outbox. It supports inbox preview text, a separate heading, brand text, an HTTPS banner, five validated colors, standard email-safe fonts, alignment, plain/inset layouts, and up to twelve reorderable content blocks. Raw HTML/CSS is never accepted. Existing drafts receive presentation defaults; an older client that omits configuration during an edit preserves the saved design and schedule.

Recipients can be all subscribers, verified account subscribers or imported opt-in contacts. An optional list of up to 200 email addresses restricts that audience; it never subscribes an unknown person. Preview and saved-detail counts use the same SQL selection as queueing. The send transaction rechecks consent/count and the worker rechecks current subscription/account state before delivery. Draft edits disable send/test controls until the new revision is loaded.

A campaign can launch immediately or use a send time at least one minute ahead and within ninety days. Launch snapshots eligible recipients and writes that send time into the existing transactional email outbox. Test email is immediate. The existing claim query dispatches only due rows; no extra scheduler, polling process or queue is introduced. Cancellation continues to suppress queued messages. An expired saved schedule must be changed before launching.

## Custom messages from Messaging (candidate)

`/messaging/new` offers service/event updates and promotional campaigns. Promotions use the campaign workflow above and `updates@mail.hackways.in`. Service messages use `notify@mail.hackways.in`, with replies to `hello@hackways.in`, and support up to fifty explicitly entered active verified Hackways accounts. The service composer provides subject, message, optional HTTPS button, the actual Go-rendered preview and an explicit review checkbox.

New authenticated Go routes are `POST /api/v1/admin/global/messages/preview` and `/messages/send`; campaign live previews use `POST /campaigns/preview`. All retain admin role, daily OTP and Origin checks. Service sends require an idempotency key and confirmation. A single transaction creates the receipt, validates every recipient and queues every message; any unavailable recipient rolls back the whole batch. Replays return the original receipt, while changed content with the same key conflicts. Service recipients are checked again before provider dispatch. Queueing is reported as pending delivery and is visible in Messaging; the UI does not claim inbox delivery.

Migration 21 adds campaign configuration and service-message receipts/delivery links. The down migration intentionally retains these additive columns/tables to preserve saved content and history during a runtime rollback. Candidate tests reuse the mandatory isolated lifecycle and email-rendering suites, including saved design, scheduling, URL escaping, origin/role checks, full-batch rejection and idempotent service sends. No real recipient campaign or service batch was sent during implementation.
