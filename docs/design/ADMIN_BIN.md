# Admin Bin

## Scope and identity

Bin is an Operate extension of the existing admin console for finding and
recovering soft-deleted events. It retains the sidebar, independently scrolling
workspace, account controls, Hugeicons, FG Futurist headings/buttons, Google Sans
Flex body text, white paper, dark ink and muted violet. It introduces no new
visual identity, assets, global design artifacts or service boundaries.

## Navigation and list

The sidebar links to `/bin` and marks it as the current page. The topbar search
has the accessible name Search deleted events and submits `q` to `/bin` using
GET. Submitting a search resets pagination because the form omits `offset`;
clearing and submitting the query returns the unfiltered list.

The server fetches real records from core Go before rendering. It requests 21
records, displays up to 20 and uses the extra record to decide whether Next is
available. Previous/Next preserve the search query and use clean query-based
URLs. The page number describes the current offset. Empty states distinguish
No deleted events from no search matches.

The existing table uses Event, Publication and Recovery columns, scoped headers
and a screen-reader caption. Its named Deleted events region is keyboard
focusable and horizontally scrollable. Row separators, muted publication text
and outlined recovery actions follow incumbent styles. The shell adapts at
1000px and 700px; pagination wraps and the table retains overflow handling.

## Move and recover

Event details expose Move to Bin through the existing action-form pattern. Each
Bin row opens an inline Restore event form through a native disclosure, whose
accessible summary includes the event title. Both actions require a labeled
Reason field and the shared `@hackways/ui/checkbox` confirmation. The reason has
3–1000 character bounds; backend checks remain authoritative.

Move to Bin explains that active events must first be cancelled and that
bookings, tickets and history are retained. Restore explains that published
events return as unpublished and cancelled registrations stay cancelled.
Recovery never republishes, reactivates a cancelled booking or purges records.
The real [Go API contract](../backend/ADMIN_BIN.md) owns eligibility,
authorization, revision checks, persistence and audit behavior.

The shared form sends the revision, reason and idempotency key to the real admin
API. Pending state disables the fieldset and shows the existing saving indicator.
Errors remain inline with an alert and preserve entered values. A timeout permits
retrying the same change with its retained key. Navigation to Bin or the restored
event details follows confirmed API success. Login, daily verification and access
denial use the existing admin recovery destinations.

## Loading, errors and evidence

Initial pending navigation reuses the centered Hackways wordmark and shimmer
with a Loading Bin status; reduced motion disables its animations. The
console error boundary presents Admin data is unavailable and Try again, which
invokes the existing reset action. Shared focus outlines, native disclosure and
form semantics remain in place.

Recorded from source on 2026-10-27: the Bin page, AdminShell, AdminActions,
Bin loading and console error boundaries, global styles and core Go Bin
handler. Browser review was explicitly excluded. Rendered layout, keyboard/focus,
mobile overflow and real-service journeys remain unverified by this documentation
pass. No browser checks, tests, builds or deployment actions were performed.
