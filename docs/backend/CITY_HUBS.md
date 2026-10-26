# City Hubs and community media

City Hubs are administrator-managed city pages, separate from owner-managed
Spaces and memberships. No cities or associations are seeded. Administrators
create real content and curate existing published public events. The public
directory contains published cities only; draft/archived URLs return 404. Event
reads recheck publication, visibility and deletion; curation cannot make private
events public.

## APIs

Core-backend owns `/api/v1/cities`, `/cities/by-slug/{slug}` and
`/cities/{id}/events` (upcoming or `period=past`, bounded limit/offset).
Verified dedicated admin sessions protect `/api/v1/admin/core/cities`:

- GET list, GET `/{reference}`, GET `/event-choices?q=...`.
- POST create, PUT `/{reference}`, POST `/{reference}/events` with `event_id`,
  current `revision` and optional `remove`.
- Writes use existing admin origin, request-key, reason, audit and replay rules.
  City rows are locked with a two-second statement timeout; stale revisions return
  409. Create retries are bound to actor, key and body hash.
- Slugs remain stable after creation. Admin URLs use positive numeric references.
  Creation keys/hashes are never serialized. Published cities require description;
  time zones must load through Go's IANA database.

Migration 27 adds nullable Space logo/banner URLs; migration 28 adds cities and
event associations. Both are additive. Down migrations retain content; roll back
application images rather than deleting community/city data.

## Direct media uploads

Global-backend's authenticated `/api/v1/storage/upload-intent` accepts
`space-logos` and `space-banners` only for the saved Space's owner. POST
`/storage/space-image/complete` accepts resource ID, kind, current revision and
either the issued blob key or `remove:true`.

Admin-only POST `/api/v1/admin/global/media/upload-intent` accepts `city-banners`;
POST `/admin/global/media/city-banner/complete` attaches/removes that banner. Both
require the dedicated verified admin session and approved browser origin, not a
public account cookie. City publication remains core-backend logic.

Uploads use 15-minute create-only Azure SAS tokens bound to a resource prefix.
Go does not proxy file bytes. Completion HEAD-verifies the configured Azure
origin, immutable key pattern, BlockBlob type, MIME/extension and 1-byte to 5-MB
size before a SQLc ownership/revision-guarded update and audit. Removal detaches
the reference without deleting the blob. An immediate completion retry can return
the same saved image at revision+1; newer changes return a conflict.

The shared frontend crop dialog produces a local File before requesting any SAS.
Square logos, 3:1 Space covers and 16:9 City banners support drag, keyboard zoom/
position, reset/cancel. Decode/export errors stay in the editor. Saving retains
the created record and completed upload steps on failure. Explicit reload fetches
the fresh revision and preserves draft entries for review.

## Release and verification

Apply migrations before starting new Go images. Replace engines individually
with readiness checks, then public/core/admin web images and scoped Nginx routes.
Organizer-web needs no replacement. Preserve existing environment/ACS settings and
database records; frontend compilation runs only off the production VM.

Existing mandatory integration cases were extended for admin authorization,
origins, publication, replay, revision conflicts, curation, private-event
exclusion, owner-scoped media and unconfigured storage. Local isolated PostgreSQL
checks do not prove live Azure uploads or inbox delivery. Browser review is
omitted at the user's request. See the worklog for executed checks and activation.
