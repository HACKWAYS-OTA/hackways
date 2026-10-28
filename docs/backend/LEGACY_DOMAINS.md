# Legacy domain handoff and default-host isolation

hackways.me has its own static, crawlable ownership/handoff page in the existing
Nginx ingress. It says hackways.me is now owned by hackways.in and directs users
to the latest platform. It does not render application or admin forms. All links
are fixed official HTTPS destinations. No extra service, app, Go API, persistent
business store, tracking or frontend dependency is introduced.

The root is 200 with a self-canonical because this ownership announcement is a
distinct page, not a duplicate of the hackways.in homepage. Metadata, Open Graph,
Twitter summary, Organization/WebPage JSON-LD, robots and a one-URL XML sitemap
are included. No fake ratings, search volume, promises or keyword stuffing. The
main platform's canonical/SEO behavior is unchanged. Known /login redirects to
the legacy root, stripping its query. Other obsolete apex paths show the same
handoff with 404; they are not a collection of fabricated 200 pages.

www/auth/join/checkin.hackways.me permanently redirect all paths to
https://hackways.me/ with no query propagation. Only these explicit legacy hosts are accepted; every accepted TLS name is
covered by the dedicated certificate.
Unknown domains cannot fall through to admin: explicit default listeners reject
HTTP 444 and unknown TLS handshakes. Existing recognized .in vhosts and admin
network/session/Go authorization policies remain unchanged.

Content is a static HTML file in the existing conf.d directory mount. Only the
two actual font files proxy to public-web, without cookies or query forwarding;
there are no legacy API routes. General rate limits and security headers apply.
Legacy access logs are off so obsolete auth query tokens are not logged.

Deploy from verified source with recovery backup. HTTP01 challenge files use
conf.d/legacy/acme/.well-known/acme-challenge with exceptions on HTTP and HTTPS, including edge-forced HTTPS. Issue a
separate Certbot SAN certificate for hackways.me, www, auth, join and checkin;
never replace the existing .in certificate. Validate nginx before reload. The
scoped deploy hook reloads ingress after this cert renews. Retain challenge routing
and account settings. Other unconfigured subdomains are rejected by the default host.

The operator may use the existing origin certificate only during the bounded
HTTP01 bootstrap, matching the observed current edge behavior; completion requires
the dedicated certificate and real HTTPS checks for all named hosts. Unknown SNI
continues to be rejected. Do not relax verification or expose admin to obtain TLS.

## Verified activation (2026-10-28)

Candidate efd7440 passed Release Candidate run 40683353480: all eight jobs,
including four-app types/lint, shared controls/API checks, SQLc compile, source
hygiene, deployment checks and PostgreSQL 16 race coverage (63 core and 136 global
cases, zero failures/skips). Its unused public image was not deployed. Temporary
publication variables were removed.

The existing ingress received only the new default-host config, legacy vhost and
static HTML. Encrypted database/configuration backups were verified by SHA-256;
PostgreSQL restore catalog was checked at clean schema 33. Dedicated Certbot SAN
issuance passed real public HTTP01 checks for all five names. The certificate
covers exactly those hosts and passes a 30-day expiry check. Persisted webroot
renewal settings and the actual scoped reload hook were verified.

Runtime verification found that a file alias on exact / triggered Nginx index
processing. Source correction 878cda7 routes / internally to the same protected
static document. The correction was source-audited and verified with real
nginx -t, reload, trusted origin TLS and public HTTPS; the full CI run above
preceded this one-location correction. The protected release record stores its
exact source commit and config SHA separately from the pinned candidate.

Public edge and separate external HTTPS checks confirmed root 200 with ownership
copy and no admin form/cookie; all named aliases and /login return fixed 301,
discarding paths/queries. Robots, sitemap and both genuine WOFF2 assets return
200. Unknown apex paths return the handoff with 404. Unknown HTTP/TLS hosts and
mismatched Host requests are rejected (HTTP/2 may report the deliberate close as
a protocol error). Main landing/discover/notices/city/sitemap checks passed and
the established admin access boundary was retained.

Compared with the verified backup, every existing ingress file and .env is byte
identical. All six application services plus ingress retain their exact container
IDs and start times, remain healthy and have no OOM indication. No image swap,
Next.js production build, migration or business-data mutation occurred. Private
activation evidence resides in the legacy-domain-final-efd7440ddaf3 release.
Actual browser appearance and authenticated interaction approval are not claimed.
Google guidance: https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls
and https://developers.google.com/search/docs/crawling-indexing/site-move-with-url-changes


## Page-local poster theme

The subsequent user-approved .me design follows their purple/lavender reference,
Faculty Glyphic display font, cropped supplied glass globe and fine background
noise. It does not replace the other applications' visual system. The ownership
notice and official destinations remain. See ../design/legacy-domain.md.

Reproducible source assets, their hashes/origins and font license live under
assets/legacy-domain. Operator release uploads go directly to the existing Azure
media container under immutable, hash-bearing brand paths. Public image and grain
reads go straight to Blob Storage; there is no VM media directory or Go upload
proxy. The actual WOFF2 uses one exact same-origin font path mapped to its fixed
Azure object with upstream TLS trust verification, SNI and no cookies. Existing
public font routes remain compatible. The application SAS upload workflow is
unchanged. CSP adds only the necessary Azure image origin; no script is added.
