# Automatic public URL discovery

The root `sitemap.xml` retains 21 static marketing/help/directory URLs. The separate
root `catalog-sitemap.xml` is an SEO metadata adapter for the approved anonymous
Go APIs. Both are advertised by robots.txt and can be submitted in Search Console.
Keeping static and catalog files separate lets static discovery remain available
if an upstream catalog is temporarily unavailable.

## Data and safety

- Existing `/events`, `/spaces` and `/programs` read endpoints own visibility and
  publication filtering. The Next.js adapter never queries PostgreSQL, accepts
  credentials or forwards a visitor's session. Private/unlisted events, draft
  programs, applications and account records are not enumerated.
- Event enumeration currently includes upcoming public events, following the
  existing Go listing contract. Past events are not a complete archive feed.
- Read in pages of50, at most20 pages/1,000 URLs per catalog. Three feeds run in
  parallel, with pages sequential inside each. A shared8-second deadline and
  2MiB decoded response limit bound a refresh. Reaching the cap logs a warning;
  partitioned feeds should replace the cap if actual catalog growth requires it.
- Next's data cache revalidates after300 seconds on demand, and one process shares
  an active refresh promise to avoid duplicate cold loads. This is not a scheduled
  crawler or a guarantee of a refresh exactly every five minutes. Successful cached
  public URLs can remain available while a background revalidation fails.
- Only actual API `updated_at` values become lastmod timestamps. Programs do not
  currently expose that field, so they omit lastmod instead of inventing dates.
- XML escapes reserved characters; URLs use canonical origin and encoded slugs.
  Entries are deduplicated. No private answer, recipient or review data is cached.
- An uncached upstream/validation failure returns503 with Retry-After300, rather
  than pretending an empty success. Static sitemap reads remain independent.

## Deployment and verification

No Go/API/schema/authorization contract changes are needed. Production uses the
existing internal Go hostnames; optional CORE_API_URL/GLOBAL_API_URL overrides are
server-only and reject credential-bearing URLs. Local preview can use the real
approved public HTTP read endpoints. Builds never fetch live catalog data.

Verify both XML files, their robots declarations, and a known published record
after release. Compare entries with the anonymous APIs; use no production test
records. An empty genuine event list is valid and must not be filled with invented
events. Google still decides crawling/indexing. Ordinary pages are not eligible
for Google's restricted Instant Indexing API.
