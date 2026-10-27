# Notices and SEO utilities

## Scope and continuity

These are incumbent admin Operate surfaces: administrators publish official
updates and manage metadata for eligible public pages. Utilities adds Notices and
SEO Manager navigation inside the existing console, preserving its sidebar,
fonts, paper/ink/violet palette, Hugeicons, shared controls and scrolling workspace.
The public notice reader uses the established public identity. No replacement
design system or new service boundary is implied.

The directories use real server-fetched Go data, search and bounded 20-record
pagination, with explicit empty states. SEO editors retain their 300px settings
column, stacking at 1050px. Notices use a document studio with title/save state and
publishing actions above a wrapping formatting toolbar and centered white page.
Publishing settings live in a dialog. At 760px the studio header wraps and page
padding reduces. Tables retain horizontal overflow. These are source-backed
layout intentions, not rendered findings.

## Notice authoring and recovery

The Docs-like notice studio uses Tiptap packages pinned to 3.32.0 in admin-web
only. The public reader does not load the editor. Drafts can start without a title
or with an image first; Untitled notice is an input hint, not stored title text.
Publishing requires a real title and meaningful document content. The optional
search summary is derived from actual authored text/image descriptions by Go when
blank, with the real title as fallback. A draft address can change until first
publication; its public address stays permanent thereafter. Draft, Published and
Archived are persisted states, and draft/archived reads remain private.

The direct-editing toolbar uses shared Select and Toggle controls for text style,
font, size, emphasis, alignment and line spacing, plus labeled color controls,
lists, quotes, code, links, tables, images, clear style and undo/redo. Selected
tables expose row/column/header controls. Markdown import inserts at the cursor
without replacing the document; export downloads `.md` and explains that fonts,
colors and alignment remain in the saved rich-text document. Existing Markdown
notices remain readable. A real loading state precedes editor initialization.

Image inserts accept PNG, JPEG or WebP up to 5 MB through the picker, paste or
drop. A new notice first saves a real titleless draft to establish its identity.
The browser uploads directly to Azure using the signed intent, then completes
through Go before inserting the returned image URL and alt text. Uploaded image
dimensions accompany rendering. Images support aspect-preserving resize, width
presets, editable descriptions and removal; the initial description comes from
the actual filename and should be reviewed. Upload and document persistence are
separate operations: inserted content follows draft autosave or explicit publish.
Failed uploads retain the file for Retry upload and report an error.

Authored drafts autosave after two seconds of inactivity; published or archived
records require an explicit action. Save draft, Publish/Publish changes and
Archive use serialized writes with persisted revisions and audited reasons. A
change note is optional in the studio; operation-specific reasons are sent when
it is blank. An uncertain response must replay the exact original body, reason
and receipt key before a newer write proceeds. Edits made during an in-flight save
remain in the draft. Save-state text distinguishes saving, unsaved, unpublished
changes, confirmed saved/published/archived and errors. Retry save and Check saved
version retain unsaved entries for review. A before-unload guard covers
dirty/saving state. Go owns authorization, validation and persistence.

Migration 33 stores a bounded structured rich-text document alongside the legacy
Markdown format. Go validates allowed nodes, marks, nesting, styles and safe URLs
and derives the fallback body from the document. The public React renderer uses
explicit elements and allowed styles, without arbitrary HTML/CSS or Tiptap.

## Public reading

`/notices` lists published notices with summaries, publication dates, pagination
and a real empty state. `/notice/[slug]` is read-only, showing title, summary,
publication date and the saved document or legacy Markdown in a readable column
capped at 840px.
Draft/archived content is not exposed through these public reads. Failures are
not disguised as empty data.

Public discovery adds Notices to the public/core footers, without a new primary
navigation destination. The public pages provide no editing controls. Notice
metadata uses the actual notice title/summary and publication information,
including Article structured data; indexing eligibility remains separate from
publication status.

Two branded 404 families preserve context: public/core and managed routes share
the platform treatment with discovery/workspace recovery links; admin uses a
separate console composition with Back to console and event/Space/Bin links.
Missing or unpublished notice slugs use the public missing-page path rather than
inventing notice content.

## SEO settings and evidence

SEO Manager edits an eligible public path's search title, description, canonical
path and noindex setting, plus internal planning keywords and editorial notes.
The shared checkbox controls exclusion. Blank metadata fields retain page defaults;
Use page defaults clears the overrides locally and requires Save to apply. Up to
2,000 planning terms are allowed, one per line. Keywords and notes stay internal:
they are never injected into page content or meta-keyword tags.

Keyword ideas is an admin-only editorial library served by Go from PostgreSQL.
Its checked-in configuration defines 120 topic records across 12 themes, with six
phrases per topic (720 phrases). These are actual editorial records, not measured
search-volume or ranking data. The manager offers browse-only access; page
editors also offer Add terms. A shared theme Select and labeled topic search
filter titles, themes and phrases. Topic disclosures show their terms; Show more
topics expands the initial 18 results in bounded batches. No-match and unavailable
library states are distinct, and a failed library read leaves the page plan
editable.

Add terms changes only the current page's draft plan. It preserves existing
entries and appends terms not already present, comparing case-insensitively and
deduplicating incoming additions. An addition that would exceed 2,000 entries is
rejected with an explanation rather than partially applied. Status copy reports
added or already-present terms and requires Save to persist the plan. Browsing or
adding ideas does not change public metadata or create content automatically.

Private account, registration, application, management and admin pages remain
excluded from indexing. Metadata settings do not publish private content or
guarantee search rankings. Go restricts eligible paths; the metadata helper
preserves an existing explicit noindex and retains base metadata on service
failure. Public override reads use a 30-second revalidation policy. Save/reload
recovery follows the persisted revision and retains unsaved entries for review.

Public home/discovery fallback metadata describes the existing technical-event
scope more broadly: hackathons, developer meetups, workshops and conferences,
including AI, software, data, cloud, cybersecurity, startups, design and community
interests. This describes the platform's discovery purpose; it does not fabricate
available events. The library introduces no empty public topic pages or automatic
keyword landing pages.

Most visited public pages uses actual recorded consented page views from the last
seven days. It distinguishes no visits from a failed suggestion request and links
to page settings. These are traffic observations, not search positions, ranking
improvements or invented keyword recommendations.

## Verification limits

Recorded from the current Utilities pages/editors/styles, shared Markdown and SEO
helpers, public notice routes/footer links and home/discovery metadata, Go
utilities/site-content/storage handlers, rich-text validator/renderer and migration
33, studio editor and shared 404 families, plus the phase-2 editorial topic-library
migration 32 and admin component. The 120-topic/12-theme/720-phrase count was
inspected in migration source and subsequently verified against production schema33.
Release6b3d46c98eb1 passed full hosted CI40549822456 and activated all five targeted
images. Public HTTP checks verified metadata, footer-only notice discovery and
branded/noindex public recovery pages; running admin assets contain the new canvas,
font, image and Markdown controls. Both404 families also passed one-off Node SSR
checks. All six services are ready; existing authored content and environment
were preserved. Browser review was excluded. Keyboard/focus interaction, visual
responsive review, authenticated notice editing and actual Azure upload completion
were not exercised in production; compilation/asset checks do not establish those
interaction results.
