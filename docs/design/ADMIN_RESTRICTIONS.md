# Admin Restrictions

## Scope and identity

Restrictions is an Operate extension of the incumbent admin console for reviewing
account suspensions and recognized-browser blocks. It retains the sidebar,
scrolling workspace, FG Futurist headings/buttons, Google Sans Flex body text,
Hugeicons, paper/ink/violet palette, shared controls and existing form patterns.
This source guide records an in-progress feature, not a deployment or rendered
verification result.

## Find and review

The sidebar links to `/restrictions` with current-page semantics. Accounts is the
default view; Browsers selects `kind=devices`. Type navigation preserves `q` and
resets the offset. Search restrictions uses a labeled GET form that retains the
selected kind and resets pagination. Previous/Next preserve both filters.

The server requests real global Go records before rendering, fetching 21 and
displaying up to 20 to determine whether Next is available. The table shows
account name/email or browser label, reason and review action. Account names link
to person details. Browser labels use the recorded user agent or Recognized
browser. The restrictions list adds no timestamp columns or freshness labels.
Empty copy distinguishes no blocked records from no search matches.

The table retains scoped headers, a screen-reader caption and a named,
keyboard-focusable overflow region. Inline native disclosures expose Unblock
account or Unblock browser forms with a record-specific accessible summary.
Existing row separators, outlined actions, wrapping pagination and responsive
admin-shell styles remain.

## Restriction actions

Person details expose Suspend account for non-admin records and recognized
Browsers with Block browser actions. A blocked browser links to its restriction
review instead. Every restriction action requires an explicit Reason with
3-1000 character bounds and the shared `@hackways/ui/checkbox` confirmation.
The form uses the persisted update version and existing idempotency transport.
Pending state disables its fieldset; inline alerts preserve entered values after
failure. Navigation follows confirmed API success.

Blocking a recognized browser suspends its associated accounts for review.
Unblocking the browser leaves those accounts suspended; unblocking an account
leaves browser blocks intact and requires a new sign-in. Each needs an independent
review and explicit action. Recognition uses the known browser cookie, not a
physical hardware identity: clearing cookies or using another browser can change
recognition. The UI explains this limit rather than promising a device-wide ban.

Go owns authorization and enforcement, including ending existing account access,
restriction checks and audited persistence. Hiding controls is not a security
boundary. The backend rejects suspending the acting administrator or an account
with active admin access, and rejects blocking a browser associated with an
administrator. These conflicts direct review of admin access first. Existing
login, daily verification and access-denied destinations handle expired or lost
console authority.

## Denial and recovery

The public `/not-allowed` page says You are not allowed, explains that the account
or browser needs administrator review and offers Get help. It is excluded from
indexing. Public and managed server account helpers distinguish restricted access
from a signed-out account: protected journeys send restricted users to this page
instead of the login flow, while optional identity reads return no user so the
public denial/help shell can render without a redirect loop. Recovery remains an
administrator operation; the denial page does not promise self-unblocking.

Initial loading reuses the Hackways wordmark/shimmer with a Loading restrictions
status and reduced-motion support. Fetch failures use the existing console error
boundary and Try again reset. Native form/disclosure semantics and shared focus
styles remain; actual focus and overflow behavior require rendered verification.

## Evidence limits

Recorded from the restrictions page/loading state, AdminShell, AdminActions,
AdminDetail, public denial page, public/managed account helpers and global Go
restriction handlers on 2026-10-27. Browser checks were explicitly excluded. No
context script, tests, builds or deployment actions were performed for this
documentation pass. Rendered behavior and real-service journeys remain
unverified here; source inspection does not establish feature completion.
