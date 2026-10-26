# Social preview images

All four frontend apps provide Open Graph and Twitter metadata. Fixed pages use
pre-rendered 1200x630 and 1200x1200 PNGs in public-web/public/social. Landscape is
listed first and used for Twitter large-image cards; square is an alternate for
clients that prefer it. Actual cropping and refresh caching remain client-controlled.
Absolute image URLs use the configured canonical origin. Images contain existing
brand fonts, strong margins, title, short summary and Hackways identification.
Ambassadors use the approved ASCII/blue direction; other fixed pages use light cards.

Core-web public event, Space and program image routes read actual Go API data.
Publication is checked on every request before reuse of rendered bytes. Private or
unpublished events use generic branding; missing/draft programs return no public
record. Account, admin and organizer metadata contain no private record details.
The Go services remain the authority for data and permissions; the Next image
modules are presentation only. No remote event image is fetched during rendering.

Dynamic image rendering allows one concurrent generation, up to two bounded waiting
requests, and a disposable 12-entry/8MiB/10-minute byte cache. Queues fail with 503
and Retry-After rather than grow without bound. Data reads are not stored in this
cache. Font files are existing brand assets converted from the supplied WOFF2
files, with no font identity change.

Regenerate fixed cards on Linux with:

    node deploy/scripts/render-social-images.cjs

The installed Next 14 node image package has an invalid file-URL path on Windows;
local dynamic image requests return 503 there. Linux image generation succeeds.
Production Linux runtime returned both required PNG dimensions at the initial launch;
the core canary peaked at 136,036,352 bytes under its 250 MiB cap. Private metadata
contains generic artwork, never applicant or result data. Static PNGs are committed
artifacts; no production build runs on the production VM.


Ambassador campaign and program link previews use the supplied
`apps/core-web/public/programs/ambassadors.jpg` unchanged (3616 x 2396 JPEG).
The program file-based OG route also returns that same JPEG, so Next metadata
precedence cannot replace it with generated campaign artwork. Other public entity
images remain dynamic. The shared artwork contains no applicant or decision data.


WhatsApp adaptation: the same supplied artwork is center-framed with `objectFit: cover`
at 1200 x 630 in the existing React OG rendering utility. No additional title or border
is painted into the asset. The resulting `ambassadors-og.png` fills wide preview cards,
and its new URL avoids reusing the earlier tall-image cache. Both campaign and program
metadata use it. The original JPEG remains the directory/email asset. Regenerate via
`node deploy/scripts/render-social-images.cjs --ambassadors-only` on Linux.
