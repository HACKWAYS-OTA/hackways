# Managed workspace (requested 2026-10-27)

Status: implemented locally; full verification and production release pending. This is a separate release from
the Spaces/City Hubs rollout. The user requested a profile shortcut, shared Space
and event operation, owner-controlled roles/access, and an organizer dashboard
with an admin-console-like shell containing only Space/event workflows.

## Intended boundaries

- Keep the four web apps and two Go engines. Management stays organizer-web at
  `/manage`; event/Space creation stays core-web. Core Go owns authorization and
  teams. Global Go owns the existing direct Azure media authorization/finalizer.
- Add an additive SQLc migration for explicit Space/event access assignments.
  The owning organizer/user remains immutable; team assignment never transfers
  ownership or grants platform administration.
- Proposed roles: owner, manager, editor, check-in. Owners alone assign/revoke
  access and perform ownership-sensitive operations. Managers operate the
  resource and its registrations; editors edit content; check-in staff can read
  the attendee information needed for admission and scan/correct check-in.
- Backend permissions must cover every read/write/export/media path. Public
  directory membership is separate from management access. Revocation must be
  checked on subsequent API requests; frontend visibility is not authorization.
- Space roles apply to associated events unless the owner assigns an explicit
  event role. Guest exports, invitations, cancellation, access management and
  destruction need deliberate restrictions. Audit the actual signed-in actor;
  never impersonate the owner in audit records.
- Assign access to existing verified accounts by exact email. Never expose a
  searchable global user directory, alter someone else's profile, or fabricate
  invitations/delivery. New mail behavior is not part of this request yet.
- Use the existing Go event/Space workflows, transaction locks, RSVP capacity,
  idempotency and failure handling. Do not duplicate those engines in TypeScript.

## Dashboard direction

The user's new instruction supersedes the previous management preference for no
sidebar. Use a responsive organizer shell with workspace home, managed Spaces,
events, resource settings and team access. Reuse Hackways typography, colors,
Hugeicons and shared controls. Do not include platform audit/moderation/users,
campaign or infrastructure modules. Accessible mobile navigation and role-aware
actions are required. Show real persisted resources and honest empty/error states.

## Required verification before release

Extend mandatory isolated integration coverage for cross-account denial, roles,
revocation, inheritance/override, team mutation authorization, guest export/media
permissions and actual audit actors. Verify existing owner workflows remain valid.
Publish immutable CI images only after successful checks; retain the live
application and outbox flows. Record completed work and limitations in WORKLOGS.

## Direction contract

THESIS: Resource operation stays in one managed workspace; owners delegate only the work a teammate needs.
OWN-WORLD: Existing Hackways paper, navy/violet, FG Futurist and Google Sans Flex, shared controls and Hugeicons. Extend the current application identity.
STORY: Open managed Spaces/events, enter a resource, and use only actions allowed by the persisted role. Owners manage teams by verified account email.
FIRST VIEWPORT: A compact desktop sidebar, account/create toolbar, real resource list and clear creation action. Mobile uses two accessible navigation links above the same content.
FORM: Code-led extension of the incumbent console as explicitly requested; no replacement visual world, invented metrics, or administrator modules.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance.
