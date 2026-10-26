# Admin analytics

## Scope and evidence

This records the existing admin analytics surface and its interval extension from
source inspection. It is an Operate surface: administrators inspect actual event,
account, messaging and consented visitor activity. It retains the admin shell,
fonts, palette, shared controls and navigation. No new visual world or assets are
introduced.

Evidence is limited to the analytics page, live/history components, shared select,
admin shell, global/analytics styles, analytics contracts and
[backend analytics documentation](../backend/ANALYTICS.md). Browser use was excluded
by the user. Rendered layout, chart legibility, contrast, keyboard interaction,
screen-reader output, responsive behavior and real-service behavior remain
unverified by this documentation pass. Source inspection does not establish
deployment or a successful authenticated API integration.

The worktree has no root `DESIGN.md` or `.impeccable/design.json`. This scoped
document does not create either artifact or infer a new global design system.
`PRODUCT.md` includes event-management-specific composition guidance; the existing
admin sidebar remains the source of truth for this separate console surface.

## Structure and controls

`/analytics` opens live analytics, defaulting to the last five minutes. The heading
contains the shared `@hackways/ui/select` control with the accessible name
"Interval", displaying `5 min`, `1 hr` or `24 hr`, followed by Refresh and History.
Selecting an interval replaces the `window` query parameter without a page reload.
The interface does not render `captured_at`, `generated_at`, an update timestamp or
clock labels on the live chart.

The six live summary values are registrations, check-ins, events created, new
accounts, consented visitors and page views. Event data supplies registration
activity, registration-state breakdown and events ranked by registrations. Global
data supplies top pages, sources and messaging counts. Event names link to their
detail pages; All events and Delivery queue retain existing destinations.

The live activity chart uses paired bars, textual series legends, an interval
direction from Start to Now, and ordinal period descriptions. View values opens
the exact registration/check-in counts in a table indexed by period. It does not
relabel calendar-day aggregates as minute or hourly measurements. The backend
contract specifies exact sliding windows with 20 buckets for five minutes/one hour
and 24 buckets for twenty-four hours. Each engine captures its own independent
snapshot; the interface does not imply one shared capture instant.

History opens at `/analytics?view=history`; URLs with `from` or `to` also select it.
The historical view retains UTC date inputs, shared date-range presets, Apply,
Live analytics and separate event/traffic-and-account CSV exports. Date selection
supports up to ninety inclusive days within the latest ninety days, defaulting to
thirty. Export actions appear only for the engine whose report loaded. Calendar
dates remain meaningful historical chart/table labels; they are not live freshness
timestamps.

Historical reports retain daily registration and consented-visitor charts, visitor
coverage, current registration states, sources, viewport sizes, event rankings,
top pages and messaging/subscriptions. Their footnote distinguishes current
registration states and ticket arrival records from immutable historical events,
and provider email acceptance from delivery.

## Layout and visual continuity

The existing admin shell supplies sidebar navigation, account controls, a content
skip action and an independently scrolling workspace. Analytics is the active
navigation destination. Typography inherits Google Sans Flex for body text and
FG Futurist for headings/buttons. Panels use the existing pale border, white
surface and gently rounded corners (14px). Summaries display tabular numerals
(32px), muted labels and unboxed figures. Chart series retain violet (`#72548e`)
and teal (`#4b8581`) with text legends and value tables.

Main analytic groups use two equal columns; historical breakdowns use three.
Both collapse to one column at widths up to 1150px. Summary values use an
auto-fitting grid with a 150px minimum; up to 540px they use two equal columns.
Headers and their controls wrap. Historical filters become two columns up to
1150px; up to 540px the preset and Apply span the row, with Apply full width.
Panel padding decreases from 26px to 18px at that breakpoint. Table containers
allow overflow and cap their height at 420px. Long breakdown labels can wrap.
The incumbent shell adapts separately at 1000px and 700px.

## Loading, empty and failure states

The server requests core and global reports concurrently before handing live
results to the client. Successful initial results render immediately. The visible
live page refreshes fifteen seconds after the previous request completes. Hidden
tabs stop polling; returning clears old figures and requests fresh reports.
Interval changes clear previous figures, abort superseded requests and accept
only responses matching the selected interval. Refresh is disabled and reads
Refreshing... during a request; a manual refresh retries both engines.

Each engine can fail independently. Failed engine figures are removed and a
specific error explains how to retry; successful engine panels remain available.
Unavailable data is not represented as a zero count. Loading without reports has
a status message, and the report region declares `aria-busy`. Genuine empty
activity has explicit interval-specific copy. A visitor collection-budget status
warns that visitor counts are partial. History reports invalid ranges and partial
engine failures with recovery through Apply.

## Accessibility and maintenance

Headings, definition lists, real links/buttons and shared select controls preserve
the incumbent semantic patterns. Live SVGs have an accessible interval description
and per-period text. Expandable value tables provide a textual alternative to
colored bars, use row/column headers, and sit inside named keyboard-focusable
scroll regions. History retains chart descriptions and daily-data tables. Error
messages use alerts; loading/coverage warnings use status roles. Global styles
provide visible focus outlines and reduced-motion rules; the shared select uses
its established reduced-motion behavior.

Keep interval calculations and authorization in Go/PostgreSQL. Keep missing-data,
partial-coverage and provider-acceptance wording precise. Preserve the live
Start/Now and ordinal-period labels when changing chart presentation. Do not add
freshness timestamps, synthetic comparisons, fabricated defaults or online-user
claims to this surface. Further visual or interaction claims require rendered
evidence that was explicitly outside this source-only pass.
