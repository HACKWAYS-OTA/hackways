# Notices and SEO utilities

## Scope and continuity

These are incumbent admin Operate surfaces: administrators publish official
updates and manage metadata for eligible public pages. Utilities adds Notices and
SEO Manager navigation inside the existing console, preserving its sidebar,
fonts, paper/ink/violet palette, Hugeicons, shared controls and scrolling workspace.
The public notice reader uses the established public identity. No replacement
design system or new service boundary is implied.

The directories use real server-fetched Go data, search and bounded 20-record
pagination, with explicit empty states. Editors pair content with a 300px settings
column; at 1050px they stack and settings stop being sticky. Tables retain
horizontal overflow, and toolbar/navigation actions wrap. These are source-backed
layout intentions, not rendered findings.

## Notice authoring and recovery

Notices have a title, summary, Markdown content, fixed-after-creation page address,
shared visibility Select and required change note. Draft, Published and Archived
are persisted states. Publishing requires summary and content in addition to the
title/address; Go validation decides acceptance. Draft and archived notices are
private. A public-notice link appears only for confirmed published state.

The content editor offers shared radio controls for Write, Side by side and
Preview. Formatting buttons insert Markdown for emphasis, headings, lists, quotes,
links, code and tables. Undo/Redo retain a bounded local history; keyboard
shortcuts support bold, italic and history actions. Preview uses the same deferred,
lazy-loaded renderer as public notices, with a genuine loading state. Side by
side stacks below 720px. Raw HTML is disabled, executable URLs are filtered, and
in-page links use the shared section-jump behavior without creating fragment URLs.
Task items use shared disabled checkboxes; wide content tables have focusable
overflow regions.

Image inserts accept PNG, JPEG or WebP up to 5 MB and require an image description
before upload. A new notice first saves a real draft to establish its identity.
The browser uploads directly to Azure using the signed intent, then completes
through Go before inserting the returned image URL and alt text. Uploaded image
dimensions accompany rendering. The interface explicitly asks for another save
after insertion: an uploaded asset is not a saved notice-body change. Failed
uploads preserve recoverable file/description state and expose an error.

Saves send the persisted revision, change note and an idempotency key; success
copy and navigation follow the actual API result. Pending state disables mutation
controls. Failures retain typed content. Reload saved version updates the saved
revision while preserving unsaved content for review; it does not discard or
automatically merge the draft. Go owns authorization, validation and persistence.

## Public reading

`/notices` lists published notices with summaries, publication dates, pagination
and a real empty state. `/notice/[slug]` is read-only, showing title, summary,
publication date and rendered Markdown in a readable column capped at 840px.
Draft/archived content is not exposed through these public reads. Failures are
not disguised as empty data.

Public discovery adds Notices to the public/core footers, without a new primary
navigation destination. The public pages provide no editing controls. Notice
metadata uses the actual notice title/summary and publication information,
including Article structured data; indexing eligibility remains separate from
publication status.

## SEO settings and evidence

SEO Manager edits an eligible public path's search title, description, canonical
path and noindex setting, plus internal planning keywords and editorial notes.
The shared checkbox controls exclusion. Blank metadata fields retain page defaults;
Use page defaults clears the overrides locally and requires Save to apply. Up to
2,000 planning terms are allowed, one per line. Keywords and notes stay internal:
they are never injected into page content or meta-keyword tags.

Private account, registration, application, management and admin pages remain
excluded from indexing. Metadata settings do not publish private content or
guarantee search rankings. Go restricts eligible paths; the metadata helper
preserves an existing explicit noindex and retains base metadata on service
failure. Public override reads use a 30-second revalidation policy. Save/reload
recovery follows the persisted revision and retains unsaved entries for review.

Most visited public pages uses actual recorded consented page views from the last
seven days. It distinguishes no visits from a failed suggestion request and links
to page settings. These are traffic observations, not search positions, ranking
improvements or invented keyword recommendations.

## Verification limits

Recorded from the current Utilities pages/editors/styles, shared Markdown and SEO
helpers, public notice routes/footer links, and Go utilities/site-content/storage
handlers. Browser review was excluded. No browser checks, tests, builds or
deployment actions were performed by this documentation pass. Keyboard/focus,
responsive rendering, authenticated saves, Azure completion and public freshness
remain unverified here. This guide describes in-progress source behavior and does
not establish activation or feature completion.
