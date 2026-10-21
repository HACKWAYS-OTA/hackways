# Main app and event management

## Agreed routing

The main product runs on port 3001. Port 3002 is reserved for managing events that already exist. The four Next applications and two Go engines remain independent.

| Address | Purpose |
| --- | --- |
| `3001/discover` | Public event discovery and search |
| `3001/events/[slug]` | Event details |
| `3001/events/[slug]/register` | Guest or account registration; application and waiting-list states |
| `3001/my-events` | Account registrations and hosted events in separate views |
| `3001/my-events/[code]` | Own registration, entry code, note, acceptance and cancellation |
| `3001/create` | Host profile setup, event details, save draft or publish |
| `3001/login`, `3001/account` | Account access, verification, guest-registration claiming and logout |
| `3002/` | Search hosted events; all/upcoming/draft/past/cancelled views |
| `3002/events/[slug]` | Real attendance summary, review shortcut, sharing and recent history |
| `3002/events/[slug]/guests` | Server-filtered/searchable guest list and application decisions |
| `3002/events/[slug]/check-in` | Entry codes, recorded manual check-in reasons and arrival queue |
| `3002/events/[slug]/edit` | Event copy, schedule, location, FAQ, contact and rewards |
| `3002/events/[slug]/appearance` | Actual cover replacement/removal and saved theme/display editing |
| `3002/events/[slug]/activity` | Paginated persisted audit history |
| `3002/events/[slug]/registration` | Visibility, capacity, joining policy, waiting list and registration windows |
| `3002/events/[slug]/registration/questions` | Name requirement and persistent ordered custom questions |
| `3002/events/[slug]/settings` | Space, publication, pause and cancellation |

## Interface decisions

Mobile event details follow the supplied reference: Back, the original 16:10 banner, title, full-width registration and authorized Manage event actions, then stacked date/time and location. Calendar/share remain with visit details; host and story follow. There is one normal document scroll. On event detail and registration routes, a compact mobile masthead provides the Hackways home mark, icon links to Discover/Spaces/My events, Create event and Account. It inherits the saved event tint and light/dark appearance directly; there is no separate theme initialization or fabricated notification/avatar state. Other routes keep their existing header.

Mobile navigation controls have 44-pixel touch targets. Below 360 pixels the create label shortens while keeping its accessible name. Mobile detail pages use safe-area-aware 20-pixel gutters, fluid title sizing, clear primary/outline action hierarchy, stacked facts and aligned calendar/share actions. Poster-title clamping affects only the generated cover; the full event title remains readable below it.

Use plain language, a short top navigation, event lists and contextual actions. No analytics-card scaffold or sidebar. Mobile places the navigation on a second header row and puts row actions below event details. The updated creation editor places the requested banner image picker on the left and a compact editor on the right. Host and visibility menus, a grouped schedule, expandable location/description, approval and capacity options retain clear labels. Mobile stacks a wide banner picker above the editor. This is a banner selection preview, not a duplicate preview of the whole event page. Host setup is inline when no profile exists. Both apps use local FG Futurist and Google Sans Flex fonts. The compact theme sheet persists appearance and light/dark mode. Ambient uses the actual event banner as a static blurred background across creation and the saved event page; a missing image leaves a neutral surface.

## Data and security

Event detail pages show Manage event only after a server-side, current-session read of the existing Go `GET /events/manage-by-slug/{slug}` endpoint succeeds. The link targets the same slug on the organizer origin (3002 locally). Public event and permission reads run concurrently; permission results are request-scoped with no shared cache. Anonymous/unauthorized visitors do not receive the link. Unexpected access-check failures keep the public page usable and show a retry. This exposes existing owner access and does not introduce staff/delegated permissions.

Public event details retain normal document scrolling. On desktop, once the banner reaches a 40-pixel top inset, the existing title and banner/visit details hold their measured positions while host, Register action and story continue scrolling. The right column releases before the footer; calendar/share actions sit beneath the date/location group on the right. There is no wheel interception, nested scroll region, popup frame or scrollbar masking. The banner keeps its original 16:10 aspect ratio. Optional FAQ, rewards and contact content is omitted when absent. Mobile uses the banner-first order described above. Registration keeps a separate form and event summary. Saved themes and banner-derived Ambient continue across both routes. Decorative publication badges are omitted from hosted-event lists and management headings; draft, paused and cancelled states remain readable as plain text.

The introduction groups title, host and the Register action without generic helper sentences. The left story contains the event description and optional event information; calendar/share actions remain on the right. Date/time and location sit side by side beneath the right banner and stack on phones. Page length follows actual content; short descriptions do not create artificial scroll distance. Detail-only layout rules leave registration summaries unchanged.

Management has seven task tabs, a Details/Cover & theme sub-navigation and Joining & places/Registration questions sub-navigation. Guest review opens a focus-contained drawer by short registration code, with cancellable private reads, historical answers and real approval/check-in actions. The overview uses an attendance sentence and actionable counts rather than analytics cards. Guest lists use rows on desktop and compact stacked identities/actions on mobile. Search and status filters run in owner-scoped Go/sqlc queries and survive pagination. Summary totals come from PostgreSQL; failures show an unavailable state, never invented zeroes. Independent resources load concurrently. Refresh is explicit and there is no polling loop. Manual check-in requires a recorded reason and retains backend one-use admission checks.

Management workspace headings use the supplied Google Sans regular face without synthesized bold or text stroke. The compact event header offers Edit event when editable and View event when published. Registration groups editable choices beside a smaller summary with real guest-filter links and anchors to the actual settings; Save follows the complete form rather than sticking over fields. Guest totals are filter actions. The guest list uses compact clickable rows with real initials, inline name/email, actual status/check-in badges, exact registration timestamps and mobile stacking. Pending guest actions open explicit approve/decline confirmation, and the overview exposes contextual editing for schedule, location and joining policy. Website answers use validated HTTPS links. Submitted-answer modification remains unsupported. The guest drawer now has a compact editable status badge and a centered status popup using Go-provided options. Notify guest reveals an optional custom message; confirmed writes update real capacity/tickets/audit and optionally enqueue email. Timestamp conflicts and idempotency preserve safe retries. Checked-in/transferred and closed-event registrations remain protected; there are no arbitrary client-only status assignments.

The management Edit page follows the create studio composition: real cover/theme controls at left and populated title, visibility, schedule, time zone, location and description at right, stacking on mobile. Additional content and notification acknowledgment remain available. The Edit toolbar exposes Save changes and Discard changes; Discard returns to the management overview without saving. The rendered appearance wrapper and decorative canvas carry the actual saved theme from the initial render, with styles loaded by the server route, instead of mutating body classes after mount. Date/time triggers have stronger themed hover/focus feedback. Save serializes the existing Go detail and appearance operations, including a real selected-banner upload, using confirmed revisions. Partial failures retain selections and identify what actually saved. A single themed dialog reports errors and confirmed success; the success tick respects reduced motion. The Edit route omits management tabs, Refresh, duplicate Edit and separate cover-save controls. A local edit-session context shares confirmed revisions and blocks overlapping saves; the forms reset only on explicit discard/reload, preserving unsaved text or cover selections during the other operation. UTC schedule semantics remain explicit. The dedicated appearance route and separate Registration controls remain available.

Appearance edits use a revision-checked core endpoint; event publication remains in core. Banner bytes still upload directly to Azure. The existing global finalizer now permits owned future draft/unpublished/published events and atomically checks state/time/revision after external verification. Started/cancelled events cannot replace their banner. Save errors retain local selections; external upload requires working Azure configuration. This does not add payments, staff/team permissions, offline scanning, bulk emails or Azure roster exports.

There are no Next.js API handlers or server actions implementing business operations. Server components fetch essential private data with the current host session and no shared caching. Browser mutations call the existing Go APIs with credentials and actual response/error handling. PostgreSQL locks, capacity checks, ownership and revisions remain authoritative in Go.

Event and registration layouts render one decorative saved-theme canvas behind the entire viewport, including the navbar/footer. The shared desktop chrome, mobile navbar and event scene are transparent over the same gradients/patterns or real Ambient cover wash. Route-scoped text/focus colors follow light/dark mode. The pinned title has no opaque mask; moving left content is clipped below it while native page scrolling and keyboard reveal are retained. The canvas and theme selectors cease applying after leaving the event layout.

Production browser API URLs are `/api/v1`, routed by Nginx. Server reads use `CORE_API_URL` and `GLOBAL_API_URL` on the private network. Local development uses Go ports 8081 and 8082. Build-time app origins must match the deployment environment. No production origin is used as a development fallback.

Production host-scoped cookies require management sign-in on the organizer origin when no session exists there. Localhost cookies are shared across ports. No cross-origin token handoff is introduced.

## Limits and recovery

Payments remain deferred. Banner selection uses the actual chosen file; upload goes directly to Azure and is verified/attached by the global Go storage service. The editor first saves a draft when a banner is selected, then uploads, attaches, and publishes only after confirmed operations. Upload attempts and the saved event are reused for unchanged in-page retries. Azure/ACS credentials remain unavailable locally, so external delivery/upload is unverified and actual service failures are shown. Private visibility now means an unlisted, published event accessible by its link; drafts remain unpublished separately. Public discovery and Space feeds exclude private events in SQL. Spaces support owner-only hosting and community membership; shared calendar/team hosting is not implemented. Registration pages and visible ticket references use persisted short codes; signed ticket APIs remain compatible. Guest claiming reports the backend message without inventing a claimed count.

Initial event creation has no automatic retry: after an ambiguous initial-create failure, users are directed to check hosted events first because that endpoint has no idempotency contract. Once a draft ID is confirmed, unchanged in-page upload retries reuse it; completion accepts an immediate identical revision retry, and a lost publish response is reconciled with an actual owner-scoped read. Registration retries retain their stable idempotency key. Edit conflicts preserve input and offer an explicit reload. Destructive status changes require confirmation; backend ownership and revision checks apply to every operation.

Signed-in browser operation and visual/device checks remain separate from compilation, HTTP checks and the database-backed backend suite. See WORKLOGS.md for the actual verification results.
