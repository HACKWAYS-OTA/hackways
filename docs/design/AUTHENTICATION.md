# Authentication surfaces

This records the approved split layout as an extension of Hackways' existing brand. It is a surface guide, not a replacement brand system. The implementation lives in [Auth.tsx](../../packages/ui/src/Auth.tsx) and [auth.css](../../packages/ui/src/auth.css).

## Overview

The brand panel sits beside a focused form with no site navigation or footer. Its copy, “Good things happen when people meet.”, comes from the public landing page. The shared presentation does not determine identity, permissions or session validity: Go and PostgreSQL remain authoritative.

| Surface | Local route | Available sign-in methods |
| --- | --- | --- |
| Core | `http://localhost:3001/login` | Existing Google sign-in first, then email/password; account creation is available in the same form |
| Management | `http://localhost:3002/login` | Existing Google sign-in first, then email/password; account creation links to core |
| Admin | `http://localhost:3003/login` | Dedicated admin email/password only |

Public/auth hosts do not gain an additional login application. Public account entry should use the existing core journey. The current auth-host Nginx configuration forwards auth API requests to Go and returns 404 for other paths; it does not currently redirect its root to the core login.

## Colors

The brand panel uses the existing pale lavender (`#eee7f5`) and plum ink (`#30243c`). Forms inherit each app's established paper, ink, muted, line, soft and violet variables. Input borders use `#9b80b1`, measured at approximately 3.43:1 against white, so the resting field boundary remains visible. Errors use `#92374f` and retain text messages.

## Typography

FG Futurist supplies the wordmark, brand statement and form heading; Google Sans Flex supplies body text and controls. Both remain the existing locally hosted regular faces. The statement scales from 40px to 64px, headings from 28px to 34px, labels use 14px, and inputs use 16px. Do not substitute fonts or add a new typography dependency for these forms.

## Layout

Desktop uses a nominal 45/55 split with a 320px minimum brand column. The form is centered in its column and capped at 400px. The brand statement is vertically centered beneath the wordmark.

At widths of 800px and below, the brand panel becomes a compact wordmark on the paper background and hides the statement. Form content has 24px horizontal padding. The layout uses the dynamic viewport height and allows vertical scrolling rather than clipping content.

## Elevation & Depth

The surface is flat. Background tone separates the brand panel from the form; input borders and focus outlines distinguish controls. No decorative card or shadow encloses the form.

## Shapes

Inputs have 10px corners, action buttons 12px corners, and both are at least 50px tall. The password reveal target is at least 44px square and remains inside the input boundary with space reserved for its icon.

## Components

Import `AuthLayout`, `AuthGoogleButton`, `AuthDivider`, `AuthPasswordInput` and `AuthSubmit` from `@hackways/ui/auth`; load `@hackways/ui/auth.css` once in the consuming app root. Existing app `Input` components provide the email/name fields. Hugeicons supply the Google, arrow and reveal icons; the reveal uses the shared `Toggle`.

Forms retain explicit native labels, required fields, field names and autocomplete. Login passwords use `current-password`; core account creation uses `new-password`. Password reveal changes only the input type and exposes an accessible Show/Hide password label and pressed state. Pending submissions disable controls, set `aria-busy`, and show pending text. Errors use `role="alert"`; recoverable submission failures leave typed values in place. Focus outlines remain visible, and reduced-motion preferences disable auth transitions and the shared toggle's press scaling.

Core and management continue using the existing Go password and Google transports. Admin continues using its dedicated admin login endpoint; this presentation adds no admin OAuth flow. Admin middleware checks Go access before console rendering, with server layout and shell checks as additional guards. Permission checks are uncached across requests; React cache only deduplicates within a server render. Signed-out requests retain a safe `next` destination, ungranted accounts reach the 403 access-denied surface, and failed backend checks reach the 503 unavailable surface. Ordinary account sessions never authorize admin access; Go requires the dedicated admin session and a current database grant.

## Do's and Don'ts

- Do reuse the shared auth exports and existing controls, fonts and palette across these three apps.
- Do retain native form contracts, accessible state feedback and real backend outcomes.
- Don't introduce a fourth login application, new auth framework, icon dependency or TypeScript authorization backend.
- Don't treat a visible form, password reveal or successful type check as proof of a complete authentication journey.

## Recorded verification

The local review captured core, management and admin at desktop (1440 × 1000) and mobile (390 × 844): [review artifacts](../../.impeccable/review/). The [interaction evidence](../../.impeccable/review/login-interaction-checks.json) records loaded fonts, no horizontal overflow, no navigation/footer, and password → text → password for all six combinations. The interaction file is authoritative for the reveal check; the separate layout file records geometry.

The initial finish review found one material input-border contrast issue. After correction and six recaptures, the focused verdict scored that fix resolved with a ship disposition; that verdict covers the correction only. Real Google OAuth, external email delivery and all authenticated user journeys were not exercised by those captures, and production has not been deployed by this change.
