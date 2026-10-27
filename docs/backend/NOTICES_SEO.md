# Notices and public SEO controls

Global Go owns editorial persistence, publication, route eligibility and admin
authorization. The four Next apps remain separate presentation clients. Additive
migration 31 adds notices, verified image references, per-page SEO plans and the
public-entity publication projection. It creates no notices or keyword records.
Its down migration retains content for forward recovery.

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

Notice slugs are immutable after creation, lowercase letters/digits/hyphens and
at most 120 characters. Title is required and at most 180; summary 500; Markdown
body 100,000. Publishing requires a nonempty summary/body. Draft and archived
notices are absent from anonymous detail/list responses. Archiving is reversible;
it does not delete text, image references or the first publication date.

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

## Markdown and media

The shared renderer uses react-markdown 10.1.0 and remark-gfm 4.0.1. The editor
supports formatting controls, write/split/preview, bounded undo/redo, Markdown
tables, links, lists, code and image descriptions. Preview uses deferred input
and loads its parser on demand. Raw HTML is disabled; executable links and signed
image URLs are not rendered. GFM tasks use the shared checkbox. In-page actions
use the shared section jump, without creating fragment navigation URLs.

New image uploads first save actual entered notice fields as a draft when needed.
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

## Verification and operation

Existing mandatory PostgreSQL administration regressions cover publication,
archive/restore, private draft SEO and canonical targets, large private keyword
plans, revision conflicts, receipts, admin/Origin access and missing storage.
Existing SAS regression covers scoped create-only notice URLs. No fixture notices
or keyword plans are created in production. Source-only Markdown SSR checks cover
formatting, tables, image dimensions and script/fragment rejection. User excludes
browser checks; no visual or authenticated production editor claim is made.

Deploy external CI images only. Back up the database/configuration, verify schema 30,
apply approved 31, start global Go before the four frontends, and install the scoped
public `/api/v1/site/` Nginx location. Preserve environment and the core engine.
Rollback may retain 31 and the prior enforcing runtime; never roll back account
restriction enforcement to a pre-retirement engine once decisions exist.

Guidance: [Google SEO basics](https://developers.google.com/search/docs/fundamentals/seo-starter-guide),
[Article data](https://developers.google.com/search/docs/appearance/structured-data/article),
[Event data](https://developers.google.com/search/docs/appearance/structured-data/event),
[safe Markdown rendering](https://github.com/remarkjs/react-markdown).
