# hackways.me ownership handoff

## Overview

This brief applies **only to hackways.me**. The user explicitly approved the
supplied lavender/violet poster direction, cropped supplied glass globe, Faculty
Glyphic lettering and noise for this static ownership handoff. Main Hackways apps
retain their existing branding, fonts and tokens. This is a page-local theme,
not a replacement product design system.

The page explains that hackways.me is owned by hackways.in and directs visitors
to the real current platform. Preserve the ownership statement, official HTTPS
links, SEO metadata, self-canonical, structured data, robots/sitemap and existing
fixed legacy-host redirects. It adds no app, API, tracking, service or dependency.

## Colors

| Local role | Source value |
| --- | --- |
| Ink and primary action | `#29004e` |
| Theme color | `#c5a4ed` |
| Background field | `#b28aef` |
| Poster gradient | `#c5aae9` → `#bb94ee` → `#aa7bf3`, 156 degrees |
| Upper-right radial light | `#ddc7ef`, fading by 58% |
| Primary hover | `#50147e` |
| Focus outline / outer ring | `#fff8ff` / `#4b1a71` |

These values belong to this document's inline styles; do not promote them to
global tokens.

## Typography

Faculty Glyphic regular (400) supplies the domain lettering, `HACKWAYS.IN` heading
and decorative brace. Google Sans Flex supplies ownership copy, supporting text
and actions. Body text is 17px/1.65; ownership is 20px/1.55. The heading scales
from 1.85rem to 5.5rem with 8vw, line height 1.18 and tracking −.035em. The
masthead domain is 23px; the primary action is 16px. Both faces use font-display
swap and preloads; this font choice is not authorized for other app surfaces.

## Layout

A masthead anchors the page, a centered ownership announcement occupies the main
area, and help/privacy links close the footer. The shell caps at 1280px with
48px minimum desktop side margins. Ownership/supporting copy use 42ch/46ch
measures. Main padding is 70px above and 90px below; actions follow by 32px.

At 600px and below, side margins become 20px, ownership/supporting copy become
18px/16px with 27ch/30ch measures, actions stack, the masthead domain becomes
19px and footer padding respects the bottom safe area. The page uses minimum
100svh height and permits content growth rather than clipping the announcement.

## Elevation & Depth

The fixed decorative backdrop uses the supplied transparent glass globe, rotated
18 degrees and cropped beyond the upper-left viewport. Its width scales
540–1020px and opacity is .43; mobile uses 620px and .35. Repeated grain is .24
opacity with soft-light blending. A large rotated outline brace sits beyond the
lower-right edge. These layers are decorative, pointer-inert and hidden from
assistive technology; they do not obscure the link interaction layer.

## Shapes

The primary link is a dark pill with 999px radius, minimum 52px height and
13px/24px padding. Other masthead/footer/secondary links have minimum 44px hit
areas. The composition uses open space rather than card frames or shadow stacks.

## Components and asset delivery

Open Hackways.in is the primary action; Discover events is secondary. Both are
ordinary real links. Semantic header/main/footer, one h1, the named announcement
section, decorative icon hiding and a 3px focus outline with an outer ring are
present in source. Reduced motion removes the action transition. Forced-colors
and print rules remove the decorative backdrop. Contrast, reflow and font loading
still require rendered verification.

The globe and grain load directly from public Azure Blob URLs. The exact versioned
Faculty Glyphic font URL is proxied through the existing Nginx ingress with
upstream TLS verification, explicit SNI/trust store and no forwarded cookies.
No wildcard media/font proxy or new service is introduced. The
[asset provenance](../../assets/legacy-domain/provenance.json) records supplied
globe origin, procedural grain seed, hashes and font origin/license. CSS performs
the globe cropping; keep those assets scoped to this handoff.

## Do's and Don'ts

- Keep the user-approved poster composition local to hackways.me.
- Preserve real ownership, official destinations and established host isolation.
- Keep decorative assets separate from semantic copy and keyboard interaction.
- Do not change the main apps' identity or invent application/SEO claims.

Source reviewed: [HTML](../../deploy/nginx/conf.d/legacy/hackways-me.html),
[Nginx configuration](../../deploy/nginx/conf.d/hackways.me.conf), provenance and
`PRODUCT.md`. Root `DESIGN.md` and `.impeccable/design.json` are absent, a
pre-existing documentation gap left unchanged. This brief records current source;
it does not claim this theme deployed, browser-approved or visually verified.
No browser or screenshot review was performed.
