# Managed workspace design reference

## Overview

This is a source-based reference for the managed Space and event surfaces implemented on 2026-10-27. It extends Hackways' existing application identity. It is scoped to this workflow and does not establish a replacement global design system. Root `DESIGN.md` and `.impeccable/design.json` remain absent.

The direction is a compact console for operating real resources: open Events or Spaces, enter a resource, then use the actions allowed by the persisted role. The organizer application remains independent and is intended to be mounted at `/manage`. Component route strings below are relative to that application. Creation, public Space pages, profile and Space editing remain in core-web. Platform administration, moderation, infrastructure and campaign modules are outside this shell.

This document records source behavior, not a rendered or runtime review. No browser inspection, viewport verification or production release is claimed. No new raster assets were introduced by this surface.

## Colors

The organizer root stylesheet remains authoritative: white paper (`--paper: #fff`), dark ink (`--ink: #1f1f23`), muted purple-gray (`--muted: #655d70`), pale dividers (`--line: #e6e0ec`), soft lavender (`--soft: #f8f5fc`) and violet interaction color (`--violet: #72548e`). Primary action fill uses dark ink through `--accent`.

The workspace sidebar uses a pale lavender surface (`#f7f4fa`). Active navigation uses a stronger lavender fill (`#eae2f2`) and dark plum text (`#392a48`); selection also carries `aria-current`. Table headings use the existing soft surface and muted text. Saved-access feedback uses green text (`#2a6256`) with a status announcement. Errors reuse the existing journey error treatment and alert semantics.

## Typography

Existing local FG Futurist and Google Sans Flex font files remain in use. Headings, the wordmark and pill buttons use FG Futurist; body text, navigation, search and tables use Google Sans Flex. The root body is regular-weight (15px, line-height 1.6). No additional font family is introduced.

Workspace page and Space headings use 34px, reducing to 30px at the mobile breakpoint. Space card headings use 23px. Team headings use 26px; inherited-access headings use 24px. Tables use 13px with muted 12px account email and role detail. Long names and emails wrap rather than supplying fabricated abbreviations. Existing event metric values use regular 32px tabular numerals.

## Layout

The desktop shell has a fixed, independently scrollable left sidebar (232px), with the wordmark, Events and Spaces navigation, and a bottom Back to Hackways link. The main area offsets by the sidebar width. The toolbar contains Create event and a direct account avatar link; it does not add an account menu. Content width is capped at 1280px with 64px total horizontal allowance. The toolbar has an 84px minimum height.

At 1050px and below, the sidebar narrows to 200px and the grant-access form becomes two columns with the submit action spanning both. At 760px and below, the sidebar disappears, the main offset resets, and the toolbar shows the wordmark plus two visible Events/Spaces links above the same resource content. These are ordinary labeled navigation links, not a drawer or hamburger menu. Mobile content has 40px total horizontal allowance. At 480px and below, the access form becomes one column and table cells reduce padding.

Resource tables sit inside keyboard-focusable horizontal scroll regions. The source retains semantic tables, captions, column headers and person row headers rather than turning tabular access assignments into unlabeled cards. Space cards use a logo or the shared Avatar fallback, description, effective role, archived label when applicable and a directional icon.

Events and Spaces fetch 21 records to display pages of 20. Space event tables follow the same pattern. Guest lists fetch 26 to display 25. Team lists request 50 with previous/next actions. Search and paging use clean paths and query parameters; guest selection updates query parameters without creating fragments.

## Elevation & Depth

The new workspace shell, Space cards and team tables use surface color, borders and spacing for separation. Their workspace stylesheet adds no shadow system. Existing guest detail dialogs retain their established drawer treatment; this extension does not redefine that component's elevation.

## Shapes

Keep the incumbent pill buttons and restrained rounded containers. Sidebar items use 9px corners; resource tabs use 8px corners; table scroll containers and Space logos use 12px corners; Space cards use 14px corners. Cards use a thin divider-colored border, changing to violet with a near-white lavender surface on hover. These are observed implementation values, not a new global radius scale.

## Components

### Shell and resource navigation

`AppChrome` provides a skip-to-content link and focusable main target. Login/authentication pages omit the workspace shell. Desktop and mobile navigation share Events and Spaces destinations and active-state logic. The create action goes directly to core-web `/create`; the avatar goes directly to core-web `/account`; Back to Hackways goes to `/discover`.

The account page adds Managed Spaces alongside My events and Hosted events. Its destination is the organizer Space list. Space resource navigation contains Events and, when `access.manage` is present, Team access. The Space heading shows the actual effective role and archived state, with a direct View Space link. Edit Space opens core-web settings when `content` is allowed. Create event is shown for an active Space with `create` permission.

Event tabs are filtered by returned permissions: Overview (`view`), Guests (`guests.read`), Invitations and Activity (`guests.manage`), Registration and Event page (`content`), Check-in (`checkin`), Settings (`lifecycle`) and Team access (`access.manage`). Resource page loaders also require the corresponding permission; hiding a tab is not the authorization boundary.

### Roles and team access

The Go permission map and shared workspace DTOs define the role vocabulary:

| Role | Allowed operations |
| --- | --- |
| Owner | All resource permissions, including team access, deletion and event/Space association. Ownership is not a selectable team role. |
| Manager | View, content, lifecycle, guest reading/management/export, check-in and creation; no team assignment, deletion or association permission. |
| Editor | View and content editing. |
| Check-in staff | View, guest reading and check-in. |

Team access uses an exact account email field, the shared `@hackways/ui/select` role control and Grant access. It assigns an existing verified account; the UI does not promise an invitation or email delivery. Assignable values are Manager, Editor and Check-in staff. The owner is not represented as an editable assignment.

The direct-assignment table shows Person, Role, Access and Action. Each person row includes the persisted display name and email. An active row supports role selection and Remove; a removed row disables role editing and offers Restore. Event rows explicitly say Remove override and Override removed. Removing an explicit event override restores the person's Space role if one exists; it does not promise removal of all effective access. A separate read-only Space access table displays inherited assignments returned by the API. Explicit event roles override Space roles rather than adding their permissions together.

Team mutations POST email, role and revision, or DELETE the assignment with its revision, to `/workspace/{spaces|events}/{resourceID}/access`. Controls lock while a request is pending. A confirmed mutation is followed by a fresh team read. If the mutation result or subsequent refresh leaves state uncertain, further changes are disabled until Reload team succeeds. The email input remains controlled and retained. Saved and removed messages follow the mutation response; a failed refresh explicitly says access was saved and requests a reload. Removed assignments remain distinguishable and restorable.

### Membership and creation

Public community membership and management access are independent. A delegated operator can see Manage Space while still having Join Space or Leave Space controls. Joining does not grant management access; leaving does not revoke an assignment. Owners omit join/leave controls. Archived Spaces pause new membership, and the leave confirmation states that existing event registrations remain.

The global create page concurrently loads the personal organizer profile and managed Spaces. Its Organize under selector is available when a personal profile or managed Space exists, including delegated managers with no personal organizer profile. Active Spaces with owner/manager roles are offered. A requested Space outside the initial list is checked through its management context and must permit creation. A personal event can collect Hosted by and create the personal organizer profile when needed. Selecting a delegated Space does not require that personal profile first.

Space settings load the real management context in core-web and require `content`. Archive/reactivation controls additionally require the owner-sensitive permission supplied as `canArchive`. Managed edits return to the organizer Space after success. Reloading the latest Space revision preserves unsaved entries for review. Archiving explains that new members/events pause while existing events and memberships remain.

### Guests and lifecycle actions

Guest roster and summary requests run concurrently. A summary failure omits totals and reports that they are unavailable; a roster failure shows a retry action rather than an empty successful list. Guest actions follow `guests.manage` and `checkin`, with custom answers and notes limited in the UI to guest managers. Check-in controls also observe event and ticket state; the source opens check-in 24 hours before the start and closes it at event end.

The existing native dialog drawer loads the selected guest with an abortable no-store request, clears old details while loading and checks the returned guest code before rendering. It offers previous/next and close controls with accessible labels. Status changes, notifications, registration answers, notes and check-in correction are shown only for the applicable permissions and state. Manual check-in requires a reason. Existing confirmations cover cancellation, pausing and inactive-event removal; removal additionally requires `delete`. Mutations display success only after their API call and refresh relevant persisted views.

### Data loading and safe failures

Organizer server reads use the existing authenticated `privateApi` with `cache: 'no-store'`. Managed event/Space context comes from Go manage-by-slug endpoints and must include a permissions array; missing access information produces an error instead of fabricated owner rights. Unauthenticated users are redirected to sign in. Invalid or unavailable resource contexts and disallowed page permissions use the existing not-found handling where implemented.

Events, Spaces and resource tables distinguish no records, no search matches and API failure. Retry links preserve the relevant search, filter and offset. API calls use the existing bounded timeout and structured error handling for network failure, unreadable responses and service errors. TypeScript presents permissions; Go and PostgreSQL remain authoritative for access, revisions, transactions, capacity and audit actors. Auth/storage requests continue through global-backend, while resource/team operations continue through core-backend. No Next.js business backend or simulated data layer is introduced.

## Do's and Don'ts

- **Do** reuse the existing wordmark, fonts, Hugeicons, shared controls and real resource DTOs.
- **Do** preserve direct account/public links, clean route paths, visible mobile navigation, labeled forms, focus states and alert/status announcements.
- **Do** derive available actions from persisted permissions and preserve explicit inherited-versus-direct access language.
- **Do** retain resource input and require a reload before another access mutation when the saved state is uncertain.
- **Don't** turn public Space membership into management authorization or hide join/leave solely because someone is a delegated operator.
- **Don't** describe removing an event override as guaranteed loss of effective access.
- **Don't** add platform administration modules, invented dashboard totals, invitation-delivery promises or replacement identity assets.
- **Don't** treat this source reference as evidence of browser, mobile, production or final-check verification. Release verification is recorded separately in the planning worklog.
