# Spaces and City Hubs

## Overview

This surface guide records an extension of Hackways' existing application identity. It covers the community Space directory, profile and creation/settings workspace in core-web, the separate public City Hub directory/profile, and City Hub editing/curation in admin-web. It does not replace the global brand or authorize changes to unrelated surfaces.

The approved direction preserves FG Futurist, Google Sans Flex, light paper, dark ink and soft violet, Hugeicons, and the shared `@hackways/ui` controls. Spaces remain communities with owners and membership; City Hubs are admin-managed local pages with curated public events. A City Hub is not a membership Space, and the public pages offer no City Hub editing controls.

Visitor mode is **Operate**: discover a community or local event, create an identifiable community, or curate a city page. Space creation is a page composer: branding sits beside directly editable community copy, with the create/save action in the top toolbar. Empty directories and event lists explain their real state rather than supplying example communities, cities or events.

## Colors

Inherit the consuming app's existing paper, ink, muted, line, soft and violet variables. Both inspected application roots currently define white paper (`#fff`), dark ink (`#1f1f23`), muted plum (`#655d70`), soft lavender (`#f8f5fc`) and violet (`#72548e`). Dark ink carries primary actions; violet supports focus and quiet actions. These are observed application values, not a new surface-specific palette.

Images carry community and city identity. Empty image regions use the existing soft violet field, an image icon or a name initial. They must not imply an uploaded asset exists. Explicit image-picker actions and real error/loading text communicate state independently of color.

## Typography

FG Futurist supplies community/city names, directory/profile headings, initials and the established action-button character. Google Sans Flex supplies descriptions, metadata, ordinary form controls and the understated New Space/Edit Space toolbar heading. Keep the existing self-hosted faces.

The Space toolbar heading is quiet (20px). The editable Space name carries the page hierarchy in FG Futurist (36px, reducing to 30px at 900px and 29px at 680px), while description and supporting fields use Google Sans Flex (15px). Directory community titles are compact (23px, reducing to 20px), with descriptions and owner/member metadata subordinate. City profile names are intentionally larger, scaling from 40px to 72px. These sizes describe the current scoped CSS, not a global hierarchy change.

## Layout

### Space directory and profile

The directory stays a centered, readable column, capped at 1120px with at least 40px outer margins on each side at desktop widths. A heading and Create a Space action lead into Explore Spaces, My communities and Spaces I host navigation. Public exploration has a compact search field; all views retain bounded pagination.

Each community is a compact horizontal card: logo/initial, member count, name, two-line description, host and Visit Space action. A large decorative banner does not dominate the directory. On mobile, the logo shrinks and the visit action moves beneath the copy.

Profiles may begin with a saved cover image, followed by logo, name, membership count, host, description and the appropriate membership/owner actions. Upcoming and past events occupy the main body; community information and guidelines sit alongside on wide screens and below at 1000px and narrower. Mobile stacks identity and actions. Archived state remains visible and explains paused membership/event creation.

### Space creation and settings

The centered composer page is capped at 1040px. Its toolbar pairs back navigation and New Space/Edit Space at left with Create Space/Save changes at right. The action sits outside the form element and submits that form through its native `form` attribute; pending state disables it and changes its label. There is no bottom save row.

Below the toolbar, a 300px branding column sits left of the editable copy with a 64px gap. The branding column stacks the cover control and a compact square logo control. The right column begins with a large borderless Space name, then a directly editable description, a website field accompanied by a globe icon, and collapsible community guidelines. Accessible labels remain present even where they are visually hidden. This is the editing surface itself; the rejected community-page preview card and sequential section headings are absent. Archive/reactivate controls remain a separate settings action.

At 900px and below, the branding column narrows to 250px and the gap to 40px. At 680px and below, branding and copy stack with a 28px gap; cover and logo sit side by side within the branding row, and the toolbar action remains at the top with tighter padding. At 640px and below, page outer margins reduce from at least 40px to 20px. Image selection and adjustment show the actual selected image inside the shared media controls.

### Public City Hubs

The directory uses image-led cards with region/country, city name and the API-supplied upcoming-event count. Three columns become two at 960px and one at 640px. Missing banners use an initial rather than invented city photography.

The profile gives the city name, region/country and time zone a clear opening. A saved banner sits alongside the title on desktop and below it on mobile; without a banner, no empty hero photo is rendered. Upcoming/past event navigation and event results lead the body. City description and the community-directory link occupy the secondary column, which moves below at 960px and narrower. City pages link to ordinary event creation; they do not grant public curation rights.

### Admin City Hub workspace

Admin retains its existing shell. The list exposes each city name, public address, optional thumbnail and publication state, with a Create City Hub action. The editor is capped at 900px; identity, region, country, time zone and publication fields use a two-column grid that becomes one column at 640px. The description spans the full grid, followed by the shared image picker and save/public-address actions.

Event curation follows the saved city details. Existing associations are listed with Remove actions, and search provides eligible published public event choices with Add event/Added state. Curation appears only after a City Hub has a persisted identity. Removing an association leaves the event itself unchanged.

## Elevation & Depth

Use the established light surface with restrained tonal layering. Community cards, City Hub cards and media fields rely on their boundaries and image regions rather than decorative shadow stacks. The crop dialog is an elevated, modal task layer with a dimmed backdrop and a single diffuse shadow.

Space composer fields, guidelines and archive settings, plus Space/City profile sections, use spacing instead of structural divider lines. Composer text fields are borderless at rest and gain an explicit violet focus outline while editing. Admin curation also loses its section-top divider; list-row separators remain for scanning. Preserve keyboard focus and necessary control/card boundaries. This is a scoped divider cleanup, not a global ban on borders.

## Shapes

Cards use gently rounded corners (14px). Media pickers use rounded frames (12px), with a wide cover/banner shape and a square logo shape. Composer text fields are flat, transparent and square at rest, with a small corner radius on their focus outline (3px). The crop dialog uses a softer outer frame (16px) around the cropped image region (12px). Existing pill actions remain part of the app language.

## Components

### Shared image field

Use `MediaField` from `@hackways/ui/media-field` for Space cover/logo and the admin City Hub banner. The visible image control is a custom button with a Hugeicons image glyph; a hidden native file input supplies the real file chooser. It supports keyboard activation and a visible focus outline. Do not display the browser's unstyled file-input row as the upload interface.

The inspected control accepts PNG, JPEG and WebP files, rejects empty files and files larger than 5 MB, shows a local preview and filename, and exposes Change image and Remove actions. Removal of a saved image can be reversed with Keep current before submitting. Client validation is feedback; backend validation remains authoritative. Image selection alone does not upload or persist the asset.

Selecting a valid image opens the shared crop-and-adjust dialog before the field accepts it. A square crop serves logos; Space banners use the default wide crop (3:1) and City banners specify a wider-than-square crop (16:9). Dragging and labeled Zoom, Horizontal position and Vertical position range controls adjust the crop; the ranges provide a keyboard alternative to dragging. Reset restores centered framing and the original zoom. Unneeded movement controls disable when the image already fills that axis.

Use image exports the adjustment to a new browser `File` and updates the image field; it does not upload or save the parent form. Cancel, close and Escape discard the pending adjustment while retaining the previous field selection. The native modal has a named heading/description; its setup moves focus to the heading and cleanup returns focus to the previously active element. These are source-backed behaviors awaiting browser verification.

The cropper reports image decoding/export failures, rejects decoded dimensions above 40 megapixels, and checks the adjusted output against the existing 5 MB limit. Output width is capped at 768px for square logos or 1600px for other crops and does not exceed the sampled source width. The modal remains scrollable within the viewport and wraps its footer actions on narrow screens. Do not imply that this local image utility replaces backend media validation or Azure completion.

### Image selection and persistence

The shared image field uses browser object URLs for temporary selected-image display and revokes them during component cleanup. A displayed selection is not an Azure asset, a saved profile or a confirmed upload. Saved fields and media follow the real API flow only when the form submits.

Space saving retains the persisted identity/revision and image-attempt state to support retry after partial success. Image uploads go directly from browser to Azure using the existing signed-upload transport, followed by completion through Go. Failures keep recoverable entries and report an error; navigation follows completed saving. Reload latest updates saved revision information while preserving unsaved form entries for review. Do not describe that action as discarding the draft or automatically merging concurrent edits.

### Membership and lifecycle actions

Owners receive event creation and settings actions; members receive membership status and a leave confirmation; other signed-in visitors can join active Spaces; signed-out visitors enter the existing login journey with a return destination. Archiving has an explicit confirmation and explains that existing memberships/events remain. Reactivation is a separate real status mutation. Go owns permissions and lifecycle validity.

### City publication and curation

Publication uses the existing shared `Select`, with Draft (hidden), Published and Archived (hidden) options. The page address is supplied at creation and remains stable thereafter. Publishing requires a description in the editor; real backend validation decides acceptance.

Event search is delayed briefly (250ms), cancels older requests and distinguishes loading, failure and no matching eligible events. Add/remove and save actions disable while pending. Success text follows the actual API result. Revision/idempotency transport remains in the existing application calls; visible controls do not establish authorization.

### Accessibility and first render

Preserve associated field labels, native form validation, semantic buttons/links, alert messages and live status messages. Composer name, description, website and guidelines retain accessible labels; borderless presentation must preserve the visible focus outline. Keep the toolbar submit action associated with the form's generated ID so native validation and submission still apply. Use clean page paths and the existing query-based tabs/pagination; do not introduce fragment navigation. Keep icons decorative when adjacent text supplies the action name. Shared checkbox, radio, select and toggle exports remain mandatory wherever those controls are needed.

Directories and profiles fetch their essential real records on the server. Space profile event and authenticated membership reads are concurrent after resolving identity; personalized membership is not a fabricated client default. Reserved image ratios and explicit dimensions describe intended layout stability. Actual focus behavior, reduced motion, font loading and absence of overflow still require rendered verification.

## Do's and Don'ts

- Do extend the existing Hackways fonts, colors, Hugeicons and shared controls.
- Do keep community identity, event discovery and the create/save action easy to scan.
- Do distinguish a local draft, uploaded media and confirmed persisted state.
- Do show the real empty, archived, loading, error and retry states.
- Do retain the toolbar create/save action, mobile branding-before-copy order and keyboard access as content stacks.
- Don't restore the rejected community-page preview card alongside or below the creation form.
- Don't restore sequential section headings, boxed text fields or a bottom save row in the Space composer.
- Don't turn City Hubs into public membership Spaces or expose public editing/curation controls.
- Don't invent community records, city photography, membership totals or events to fill empty regions.
- Don't introduce a new visual identity, font/icon dependency or frontend business backend.
- Don't claim that source inspection, type checks or HTTP responses establish a visual or complete user-journey pass.

## Direction contract

**THESIS:** Extend Hackways with recognizable communities and separately curated local event pages; community creation directly composes the page's branding and copy instead of presenting a sequential details form.

**OWN-WORLD:** Retain the existing paper/dark-ink/soft-violet language, FG Futurist and Google Sans Flex, rounded controls and Hugeicons.

**STORY:** Visitors discover real communities or city events. Owners identify their community and save it. Administrators prepare and curate City Hubs through the separate console.

**FIRST VIEWPORT:** The top toolbar places New Space/Edit Space and back navigation at left, Create Space/Save changes at right. A compact branding column sits beside the large editable name and description, followed by website and guidelines. Mobile stacks branding above copy while retaining the action in the toolbar.

**FORM:** Local extension of the established application world. No concept-seed or replacement identity was requested for this narrowly specified direction.

**FINISH:** unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance

For this extension, existing global design files are preserved. No shipping raster was generated for the guide. The rendered-review limitation below remains explicit; this contract is not a shipping verdict.

## Evidence and verification limits

Recorded from source on 2026-10-27. Reviewed `AGENTS.md`, `PRODUCT.md`, the architecture and verification guides, existing authentication/brand-icon/management design notes, core/admin global CSS, Space directory/profile/form/membership and upload helper, public City directory/profile/CSS, admin City list/new/detail/editor/CSS, and the shared media/crop controls/CSS and UI usage guide. The final Space form and CSS were reread after replacement of the rejected preview/sequential-form arrangement with the toolbar-led page composer. Root `DESIGN.md` was absent in this checkout during inspection; this task creates only the local surface guide and does not repair global design artifacts or sidecars.

Primary implementation references: [Space form](../../apps/core-web/components/spaces/SpaceForm.tsx), [Space layout](../../apps/core-web/components/spaces/spaces.css), [public City profile](../../apps/core-web/app/cities/[slug]/page.tsx), [City layout](../../apps/core-web/components/cities/cities.css), [City editor](../../apps/admin-web/components/admin/CityEditor.tsx), [admin City layout](../../apps/admin-web/app/(console)/cities/cities.css), [shared media field](../../packages/ui/src/MediaField.tsx), [crop dialog](../../packages/ui/src/ImageCropDialog.tsx) and [media/crop styles](../../packages/ui/src/media-field.css).

Browser review was explicitly excluded at the user's request, so no current rendered review was performed. User feedback rejected the preview card and subsequent sequential form; the resulting composer was checked in source. Earlier screenshots do not verify the final composer layout, crop flow, divider cleanup or responsive result. This document records source-backed behavior and responsive intent, not a visual pass. Browser keyboard/focus, mobile overflow, reduced motion, crop interaction, authenticated saving/curation and external Azure completion remain unverified. No browser, build or deployment actions were performed by this documentation work. No new tests, fixtures, generated assets or application functionality were added.
