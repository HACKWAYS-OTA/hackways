# Circular Hackways favicon

The favicon uses the existing FG Futurist lowercase `h`, outlined to SVG paths,
with the wordmark's terminal dot. A solid navy circle (`#173b7a`) carries white
lettering and a pale blue dot (`#a5d8ff`). The square asset has transparent corners;
the visible mark is circular. No font request, external image, animation or runtime
icon renderer is needed.

The canonical source is `apps/public-web/public/brand/hackways-icon.svg`.
The same checked assets are copied into each of the four independent apps:

- `favicon.ico`: transparent 16, 32 and 48 pixel frames.
- `brand/hackways-icon.svg`: scalable source with outlined lettering.
- `brand/hackways-icon-{48,96,180,192,512}.png`: transparent raster exports.

All roots use `brandIcons()` from `@hackways/ui/brand-metadata` for browser,
shortcut and Apple touch links. Organizer-web passes `/manage` because of its
existing Next.js base path. Admin links to the public brand origin so its login
page does not fetch protected static paths or require an authentication exception.
Public/core use root paths. Existing
private-page noindex policy and social preview artwork remain unchanged.

The public homepage advertises a crawlable 96px PNG and SVG plus the ICO fallback.
See [Google's favicon guidance](https://developers.google.com/search/docs/appearance/favicon-in-search).
This supports favicon eligibility; Google controls whether and when it appears.
Browser/search caches may continue showing older icons until refreshed.

Export tooling used the already available FG Futurist font, fontTools, Sharp and
Pillow as local asset utilities. No app dependency or build-time rendering step
was added. Future edits should change the canonical SVG, regenerate the raster
sizes/ICO, copy the exact results to all apps and verify matching hashes.

The user also has copies and a scale preview under
`C:\Users\ADMIN\Downloads\HACKWAYS LOGOS\Favicon`.
