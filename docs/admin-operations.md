# Admin operations and Discover promotions

The existing network-isolated admin application retains its sidebar and independent Go authorization. The event management console at `/manage` retains its separate top navigation. New admin destinations are `/discover-promotions`, `/rollouts` and `/preferences`; all are rejected by middleware and the server layout before rendering for signed-out or unauthorized visitors.

## Discover promotions

Migration 16 adds one revisioned settings row and up to five ordered promotion slots. Both the slot range constraint and unique event constraint apply in PostgreSQL. Core Go owns selection, scheduling, eligibility and public reads.

Admin APIs under `/api/v1/admin/core`:

| Method and path | Contract |
|---|---|
| `GET /discover-promotions` | Current revision, update time and saved numeric event references/windows |
| `GET /discover-promotions/events?q=&offset=` | Twenty eligible published/public/upcoming events plus `has_next` |
| `PUT /discover-promotions` | `{revision,selections:[{reference,starts_at,ends_at}]}`; up to five unique events; empty selection hides the banner |

Writes require the dedicated admin session, live grant, approved Origin and `Idempotency-Key`; Go supplies an audit description when the optional legacy `X-Admin-Reason` is absent. The transaction locks the actor, singleton settings and selected events in numeric-reference order. It validates publication/visibility/deletion and event start, replaces only the saved placement, increments revision and appends an audit. Conflicting revisions return 409 without changing saved configuration. Retries replay the confirmed result.

Windows use explicit ISO timestamps; the editor displays device-local datetime inputs. The promotion must end after its start, in the future and no later than event start. Expired or newly ineligible selections remain visible to admins for cleanup; public reads omit them.

Public `GET /api/v1/events/promotions` returns at most five currently active public events, their display expiry, server time and the next boundary. Future selections are not exposed. Discover loads the optional feed alongside its catalogue on the server. There is no automatic first-event fallback. If no promotion is active, the banner is absent and regular catalogue rows remain.

Active slides rotate every seven seconds without polling on that interval. Manual controls, pause/play, keyboard focus, pointer hover, hidden-tab suspension and reduced-motion preferences are supported. Keyboard focus stops automatic playback until the user starts it again. The feed revalidates at boundaries and while visible at most once per minute. The carousel hides unavailable/expired placements instead of retaining them as current.

## Operational actions

- Home shows real core/global queue totals with independent errors if either engine is unavailable. It does not invent CPU, memory or throughput metrics.
- Tables retain URL-based filters and pagination, add saved visible columns and selection of up to twenty records.
- Bulk event/space/registration/check-in/email-retry controls call the existing authoritative Go endpoints sequentially. Each record is read for its current version, checked again by Go and receives its own audit/idempotency receipt. The result list reports partial failure explicitly. Role changes and destructive image detachment are not bundled into an indiscriminate bulk operation.
- People retains profile editing, explicit admin grants/revocation with self/last-admin protection, related sessions and individual session revocation. Session and record affordances are styled as actions.
- Audit history remains read-only.

## Shared controls

The former `/preferences` display module has been removed at the product owner's request. The admin layout no longer reads or applies its browser cookie. Table column customization remains available. Standard keyboard focus, semantic labels, skip-to-content and device reduced-motion behavior remain enabled.

The skip-to-content control uses the shared section jump behavior without adding URL fragments. Hidden radio roots and Base UI's companion form inputs are positioned inside their own option. The outer fixed dashboard frame uses `overflow: clip`, so focusing/validating a hidden input cannot programmatically scroll the entire clipped frame and leave a blank viewport. The workspace and sidebar retain their own scroll areas.

## Verification and release state

The mandatory core tests exercise admin-only access, six-event rejection, the database slot constraint, replay/revision conflicts, public eligibility, start/end boundaries, hiding an empty/future placement and concurrent saves. Both full local Go suites, vet/build and sqlc checks passed against isolated PostgreSQL 18 databases. Frontend/shared controls checks are recorded in the worklog. Anonymous HTTP checks redirect all new admin pages to sign-in. Local Discover SSR renders ordinary rows and zero promoted cards when unconfigured.

The local preview uses schema 19. Only daily admin verification emails may be delivered; other templates from the copied-data outbox remain paused and unleased. Production was activated at the user request on 2026-10-22 with all six images from `dab8d1cd545b843bf20c2d8cdac16bb7feb058d8` and clean schema 19. Release Candidate run 39795523731 passed the full Linux PostgreSQL 16/race pipeline and six image builds. Production has no review email filter. Live checks verified public pages, sign-in gates, current runtime configuration and served containment CSS. The desktop/browser automation connection was unavailable, so visual focus behavior and responsive presentation need the user's browser review; compilation and API tests do not certify those interactions. See [session persistence](auth/session-persistence.md) for the 90-day and daily admin verification policy.


The immutable production image manifest is `/opt/hackways/releases/admin-auth-rollouts-dab8d1cd545b/images.compose.json`; protected rollback files are under `/opt/hackways/release-backups/admin-auth-rollouts-20261022T143008Z`. The database backup was streamed directly into the private Azure Blob `backups` container and its size verified. No Next.js build ran on the VM.

An initial ingress check stopped activation after the additive migrations: the container retained an older inode for its individually bind-mounted `nginx.conf`. Recreated only Nginx using its existing image to bind the validated current file, re-derived its exact internal `/32` peer for trusted location headers, upgraded both backends, then the frontends and reopened admin ingress. A reload alone cannot refresh a stale file bind mount after host-side atomic replacement. Admin allowlists, secrets and the existing pass-signing identity were preserved. Never restore a backend that omits daily verification while leaving admin ingress open.
