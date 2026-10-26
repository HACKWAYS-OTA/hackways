# Original Hackways favicons

The user's three PNG originals in `C:\Users\ADMIN\Downloads\HACKWAYS-FAVICONS`
are the visual authority. They supersede the generated monogram. Their shape,
gradients, colors, white center and transparency are unchanged.

| Variant | Original | Use |
| --- | --- | --- |
| general | GENERAL-FAVICON.png | Public marketing, discovery, Spaces and programs |
| event | EVENT PAGE FAVICON.png | Event details, registration and event management |
| admin-auth | ADMIN FAVICON  ALSO FOR AUTHINTICATION.png | All public auth pages, sign-in, connections and the entire admin console |

All sources are 400x400 RGBA. Native 400px platform PNGs are byte-for-byte copies.
Source SHA256 values are recorded in the user's export manifest. No recoloring,
sharpening, tracing, AI enhancement or redrawing is used. PNG exports above 400px
are Lanczos resamples, not newly recovered source detail.

Each app has `public/brand/{general,event,admin-auth}/` with:

- `favicon.svg`: a self-contained SVG embedding the exact original PNG. This
  preserves the source pixels and is not described as traced vector artwork.
- `favicon.ico`: transparent 16/32/48/64/128/256px frames.
- `icon-{48,96,180,192,400,512}.png`: transparent derived/native sizes.

Legacy `brand/hackways-icon*` paths now contain the supplied general artwork.
Root favicon.ico uses each app's default variant. The shared
`brandIcons(basePath, variant)` helper supplies browser, shortcut and Apple links.
Organizer assets use `/manage`; admin links use the public origin so login needs
no authentication exception for assets. Metadata-only public auth/connection
layouts and sign-in metadata select green. The event layout selects purple.
Existing form, session and application logic is unchanged.

Google supports one search favicon per hostname; page tab variants do not promise
different search-result icons on the same hostname. The homepage exposes a 96px
square PNG plus SVG/ICO. Google controls recrawling and search display. See
[Google's favicon guidance](https://developers.google.com/search/docs/appearance/favicon-in-search).

The original Downloads files remain unchanged. General, Events and
Admin-Authentication subfolders contain original copies, transparent PNGs from
16px to 1024px, embedded-image SVGs, ICO, lossless transparent WebP and quality98
white-background JPGs. A manifest, README, preview and ZIP are alongside them.
JPEG has a white background because it cannot store transparency.

Exports use the already available Pillow utility. No app dependency or production
build-time image renderer was added. Future artwork changes must start from
approved originals, then regenerate/check identical asset hashes in all four apps.
