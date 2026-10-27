# Notices and public SEO controls

Global Go owns editorial persistence, publication, route eligibility and admin
authorization. The four Next apps remain separate presentation clients. Additive
migration 31 adds notices, verified image references, per-page SEO plans and the
public-entity publication projection. It creates no notices or keyword records.
Its down migration retains content for forward recovery.

Phase 2 adds migration 32's admin-only PostgreSQL editorial topic library: 120
records across 12 themes, six phrases each (720 phrases). Phase 3 adds migration
33's nullable JSONB `document` and `body_format` (`markdown`/`richtext`), preserving
existing Markdown rows. Rich-text documents must be objects of at most 1 MiB;
Markdown rows have no document. The title constraint now permits titleless drafts
while requiring a nonblank published title. Down guards retain authored content.
These describe checked-in source. The supplied live baseline remains d0/schema
31; neither later migration nor the phase-3 frontend is claimed deployed here.

## Administration

The existing dedicated admin session, explicit grant and daily verification protect
all `/api/v1/admin/global/utilities/*` routes. Writes additionally require the
approved admin Origin, a short `X-Admin-Reason`, printable Idempotency-Key and the
current revision. Existing transaction timeouts, user/grant rechecks, a resource
advisory lock, confirmed audit entry and small persisted receipt protect edits.
Different admins cannot silently overwrite newer revisions. Replaying a confirmed
request cannot create another notice or revision.

- GET/POST `/utilities/notices`: bounded search/list and create/update.
- GET `/utilities/notices/{reference}` and `/utilities/notices/by-slug/{slug}`:
  private editing and recovery after uncertain creation, including drafts.
- GET/POST `/utilities/seo`: bounded saved-page list and upsert.
- GET `/utilities/seo/page?path=`: private full editorial plan, or null if absent.
- GET `/utilities/seo/suggestions`: actual recorded consented public page views
  over the last 7 days, bounded to 10 eligible results from 30 grouped candidates.
  These are not Google trends, search positions or search-volume estimates.
- GET `/utilities/seo/library`: admin-only editorial topic records, bounded to
  150 with a two-second query timeout. The manager browses/searches/filters them;
  per-page editors can append case-insensitively deduplicated additions to the
  local keyword draft, refusing an addition above 2,000. Save is still required.
  No library phrases are injected into pages or turned into empty topic pages.

Notice slugs use lowercase letters/digits/hyphens and at most 120 characters.
They remain editable until first publication, then stay fixed through archive and
republication. An initially blank slug is derived deterministically from the
receipt key, so image-first/titleless draft creation can replay exactly. Title
is at most 180 characters, summary 500 and fallback body 100,000. Publishing
requires a real title and meaningful text or image content. A blank summary is
derived from actual document text/image alt text, capped at 180 characters, with
the actual title as fallback. Draft and archived notices are absent from anonymous
detail/list responses. Archiving is reversible; it does not delete content, image
references or the first publication date.

Draft autosave waits two seconds after authored changes and serializes writes.
Published/archived records do not autosave. Manual Save draft, Publish changes and
Archive use the same revision-aware transport. The studio generates an audited
operation reason when its optional change note is blank. Before any newer write,
an uncertain save replays the exact original request body, reason and receipt
key; it does not substitute newly edited content under that key. Confirmed
snapshots update the revision without overwriting edits made in flight. Retry save
and Check saved version retain unsaved content; checking is not automatic merge.

SEO plans support title 90, description 320, canonical public path 240, noindex,
editorial notes 4,000 and up to 2,000 unique planning keywords of 1-64 characters.
Keywords and notes are private planning data, never public metadata, HTML keyword
blocks or meta-keyword tags. No fabricated keyword lists are inserted. Blank
presentation fields restore the page's own defaults. Private account, application,
registration, management and admin routes cannot be indexed through this API.

## Public content and metadata

Anonymous Go routes live under `/api/v1/site`: `/notices`, `/notices/{slug}`,
`/seo?path=` and `/seo-rules`. Lists are paginated. Notice body reads use no-store;
only currently published content is returned. Public SEO projections omit all
internal keywords, notes, actor IDs and revisions. Both source and canonical
entity targets must be public before their paths appear in metadata/sitemap rules.
Draft SEO plans can be prepared without exposing their titles or target slugs.

The public directory is `/notices`; individual pages use `/notice/{slug}`, such as
`/notice/notice-2026-10-1`. Public entry links appear only in footers. No new top
navigation item, homepage promotion or app banner is introduced. Administrators
write the initial notice themselves; an unknown or unpublished slug returns 404.

Shared server-only SEO presentation reads reviewed Go settings with a 30-second
revalidation interval and bounded timeout. Existing accurate page metadata remains
the fallback on service failure, with a server warning. Public/static pages use
ISR 30; private core routes inherit noindex, as management/admin already do.
Public entity readers still enforce their existing visibility before metadata is
constructed. SEO cannot publish content or bypass resource access checks.

Static and catalog sitemaps exclude noindex and noncanonical source URLs. Catalog
data is bounded/cached for up to 5 minutes and includes only published Go notice
records. Sitemap failure returns an error, not a fabricated empty feed. Normal
metadata, Open Graph and Twitter defaults remain available. Public notices have
Article JSON-LD with actual publication/modification dates and platform identity;
public published events have Event JSON-LD from actual event fields. Meeting URLs,
tickets and Space-owner account attribution stay out of structured data. There
are no invented prices, ratings or eligibility/ranking guarantees.

## Documents, Markdown and media

Tiptap packages are pinned to 3.32.0 in admin-web only. Its Docs-like studio edits
structured JSON with shared Select/Toggle controls, fonts/sizes/colors, emphasis,
alignment/spacing, lists, quotes, code, links, resizable tables, undo/redo and
images. Markdown import inserts at the cursor; export downloads the compatible
Markdown representation and leaves richer styling in the saved document. No
editor library is required by the public reader.

Go's typed validator accepts a `doc` root and explicit node/mark allowlists. It
checks structure, at most 10,000 nodes, depth 32, 100,000 text characters, bounded
table/image attributes, supported fonts, sizes 8-72px/pt, line height 1-3 and hex
colors. Links reject credentials, executable schemes and fragments; image URLs
reject signed SAS credentials. Unsupported content fails with `INVALID_DOCUMENT`.
Go derives the legacy fallback body from validated content instead of trusting a
second client body. Existing revision, authorization and audited receipt rules
apply to document saves.

Public rich text is rendered through `@hackways/ui/document` as explicit React
elements and allowed styles, without arbitrary HTML/CSS. Legacy Markdown uses
react-markdown 10.1.0 and remark-gfm 4.0.1 with raw HTML disabled, safe URLs,
shared task checkboxes and section-jump actions without fragment navigation.
Both formats retain focusable overflow regions for tables and image dimensions.

New image uploads first save actual entered notice fields as a titleless draft
when needed. Picker, pasted and dropped images use the same upload path. Inserted
image descriptions start from the filename and are editable; width presets,
aspect-preserving resize and removal update the document rather than the blob.
`/api/v1/admin/global/media/notice-image/intent` issues a 15-minute create-only SAS
under `notice-images/{notice-id}/` for PNG/JPEG/WebP up to 5 MB. The browser uploads
directly to Azure, omitting application credentials. No media is proxied through
Go RAM or stored on VM disk. `/notice-image/complete` verifies scoped key, actual
blob existence, content type and size through bounded HEAD, then rechecks current
admin access before recording the unsigned public URL and audit in a transaction.
Repeated completion does not duplicate attachment/audit. Rendering dimensions are
bounded client-provided layout metadata, not a server pixel-content certification.
Public body responses include at most 100 associated image dimensions, restricted
to URLs referenced by the currently saved published body; pending image uploads
are not exposed by the public API. Existing
Azure configuration/CORS remains required; missing configuration returns 503.

Two shared branded 404 families provide presentation recovery: platform
public/core/management routes return to discovery or the workspace, while admin
uses a separate console composition with console/event/Space/Bin links. Missing
or unpublished notice slugs still return the actual 404 path. These components
do not weaken authorization or synthesize records.

## Verification and operation

Existing mandatory PostgreSQL administration regressions cover publication,
archive/restore, private draft SEO and canonical targets, large private keyword
plans, revision conflicts, receipts, admin/Origin access and missing storage.
Existing SAS regression covers scoped create-only notice URLs. No fixture notices
or keyword plans are created in production. Source-only Markdown SSR checks cover
formatting, tables, image dimensions and script/fragment rejection. User excludes
browser checks; no visual or authenticated production editor claim is made.
This phase-3 documentation pass inspected source only and ran no tests, builds,
browser checks or authenticated cloud uploads. Autosave/retry interaction,
formatting parity, image resize, responsive layout, 404 presentation and Azure
completion are not independently verified by this pass.

For a future release, deploy external CI images only. Back up the database and
configuration, verify the actual migration baseline, and apply approved 32/33 in
sequence after 31 before starting a schema-33 global runtime. Frontends consuming
documents follow the matching Go engine. Preserve the scoped public
`/api/v1/site/` Nginx location, environment and independent core engine. Retain
additive content schemas on runtime rollback; choose a compatible enforcing
runtime rather than dropping authored data. Never roll back account restriction
enforcement to a pre-retirement engine once decisions exist.

Guidance: [Google SEO basics](https://developers.google.com/search/docs/fundamentals/seo-starter-guide),
[Article data](https://developers.google.com/search/docs/appearance/structured-data/article),
[Event data](https://developers.google.com/search/docs/appearance/structured-data/event),
[safe Markdown rendering](https://github.com/remarkjs/react-markdown).
