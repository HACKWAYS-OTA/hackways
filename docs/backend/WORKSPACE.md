# Managed workspace and delegated access

Organizer web retains `/manage` and contains only Space/event workflows. Core web
retains creation, public profiles and Space content editing. Core Go owns resource
authorization and teams; global Go independently authorizes direct Azure media
signing and finalization. There are still four web apps and two Go engines.

## Access model

Migration 29 adds `space_access` and `event_access`, with resource/user uniqueness,
constrained roles, optimistic revisions, actor references and revocation records.
Ownership remains derived from the existing organizer/user relation and cannot be
assigned, transferred or revoked through team endpoints. Existing ownership works
without assignment rows. Only existing verified accounts can receive assignments,
looked up by exact normalized email; there is no global user-directory search.

| Role | Allowed work |
|---|---|
| Owner | All resource operations, team assignment/removal and ownership-sensitive actions |
| Manager | Resource details, publication/cancellation, registrations, invitations, check-in; Space event creation |
| Editor | Resource details, appearance and registration configuration; no guest management or lifecycle actions |
| Check-in staff | Admission roster, ticket check-in/correction; no registration changes, messages, private notes or form answers |

Space assignments apply to associated events. An active explicit event assignment
overrides the inherited role. Removing an event override restores any active Space
assignment; it does not deny inherited access. The owner remains owner regardless
of assignments. Public Space membership remains independent of management access.

## APIs

All endpoints require the ordinary authenticated account, not a platform-admin
session. Responses are private/no-store. Catalogues are bounded to 100 records
per page (default 30), search to 200 characters, and reads to five seconds.

- `GET /api/v1/workspace/spaces`: owned or assigned Spaces, `q`, `limit`, `offset`.
- `GET /api/v1/workspace/events`: owned, explicitly assigned or inherited events;
  optional `space` slug, `q`, `view=all|upcoming|drafts|past|cancelled`, pagination.
- `GET /api/v1/{spaces,events}/manage-by-slug/{slug}`: private resource envelope
  includes authoritative `access` with role and permissions.
- `GET /api/v1/workspace/{spaces,events}/{id}/access`: owner-only direct and
  inherited assignments, bounded pagination. Revoked direct rows remain available.
- `POST` to the same access collection: `{email, role, revision}`. Use revision 0
  for a new assignment, current revision for a role change or restoration.
- `DELETE .../access/{assignment_id}`: `{revision}` revokes a direct assignment.

Identical desired grants and already-revoked removals are safe retries without
duplicate audit records. Stale conflicting revisions return 409. Owners cannot
assign themselves a restricted role. Non-owner team mutation is forbidden; other
accounts cannot discover private resources. UI visibility never grants access.

## Mutation safety and attribution

Each event route checks a specific backend permission. It scopes the resource's
existing organizer workflow while retaining the actual authenticated actor.
Event mutation locks the event first, then rechecks applicable grant rows with
share locks; revocations serialize on those grant records. Space writes similarly
lock the Space and recheck its grant. Existing RSVP inventory locks, idempotency,
capacity checks and two-second statement timeouts remain in place.

A Space manager can create an associated event without a personal organizer
profile. Creation locks the Space and grant, verifies active status and permission,
and persists the Space owner's organizer ID to preserve the association invariant.
Audit records identify the teammate who performed the operation, including content,
team access, image and admission changes.

Media bytes still travel directly to Blob using a 15-minute write-only SAS. Both
signing and final attachment check the actual actor's effective role. Finalization
locks the resource and grant rows, checks revision/permission again and retains the
existing blob verification. A previously issued SAS never grants publication
rights after access revocation. No new email behavior is introduced.

## Release and recovery

Apply the additive migration before either new Go image. Old applications remain
compatible with the retained schema. Roll back application images while retaining
assignment/audit data; the down migration deliberately preserves these records.
Do not run permanent data purges or delete production assignments for verification.

Mandatory isolated CI cases extend the existing event lifecycle and program/media
suites. They cover verified-account assignment, owner-only mutation, cross-account
denial, inheritance/override, manager creation without a personal profile, actor
audit, staff admission/privacy and revocation, plus independent media enforcement.
Local checks use an isolated PostgreSQL database; hosted PostgreSQL 16/race checks
and immutable image verification are required before production activation.
