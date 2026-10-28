# Legacy domain handoff and default-host isolation

hackways.me has its own static, crawlable ownership/handoff page in the existing
Nginx ingress. It says hackways.me is now owned by hackways.in and directs users
to the latest platform. It does not render application or admin forms. All links
are fixed official HTTPS destinations. No extra service, app, Go API, persistent
business store, tracking or frontend dependency is introduced.

The root is200 with a self-canonical because this ownership announcement is a
distinct page, not a duplicate of the hackways.in homepage. Metadata, Open Graph,
Twitter summary, Organization/WebPage JSON-LD, robots and a one-URL XML sitemap
are included. No fake ratings, search volume, promises or keyword stuffing. The
main platform's canonical/SEO behavior is unchanged. Known /login redirects to
the legacy root, stripping its query. Other obsolete apex paths show the same
handoff with404; they are not a collection of fabricated200 pages.

www/auth/join/checkin.hackways.me permanently redirect all paths to
https://hackways.me/ with no query propagation. The redirect vhost also matches
other legacy subdomains; supported TLS names are covered by the dedicated cert.
Unknown domains cannot fall through to admin: explicit default listeners reject
HTTP444 and unknown TLS handshakes. Existing recognized .in vhosts and admin
network/session/Go authorization policies remain unchanged.

Content is a static HTML file in the existing conf.d directory mount. Only the
two actual font files proxy to public-web, without cookies or query forwarding;
there are no legacy API routes. General rate limits and security headers apply.
Legacy access logs are off so obsolete auth query tokens are not logged.

Deploy from verified source with recovery backup. HTTP01 challenge files use
conf.d/legacy/acme/.well-known/acme-challenge with an exception on port80. Issue a
separate Certbot SAN certificate for hackways.me, www, auth, join and checkin;
never replace the existing .in certificate. Validate nginx before reload. The
scoped deploy hook reloads ingress after this cert renews. Retain challenge routing
and account settings. A wildcard redirect match is not a wildcard certificate.

The operator may use the existing origin certificate only during the bounded
HTTP01 bootstrap, matching the observed current edge behavior; completion requires
the dedicated certificate and real HTTPS checks for all named hosts. Unknown SNI
continues to be rejected. Do not relax verification or expose admin to obtain TLS.

No deployment or actual browser approval is claimed by this source document.
Google guidance: https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls
and https://developers.google.com/search/docs/crawling-indexing/site-move-with-url-changes
