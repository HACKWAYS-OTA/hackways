---
version: 1
slug: "apps-public-web-app-ambassadors-page-tsx"
primary_target: "apps/public-web/app/ambassadors/page.tsx"
related_targets: ["apps/public-web/app/ambassadors/ambassadors.css","apps/public-web/components/ambassadors/GradientASCII.tsx","apps/public-web/components/ambassadors/AmbassadorsHeroArt.tsx","apps/public-web/components/ambassadors/AmbassadorsHeader.tsx","apps/public-web/components/ambassadors/AmbassadorsHeader.css"]
---

# Ambassadors and program journeys

## Scope and authority

Public-web owns `/ambassadors`. The user selected a community story and supplied a
black/white poster as style reference, explicitly not a layout to copy. Subsequent
instructions replaced email joining with a real application workflow and private
admin review. The final hero instruction replaces the supplied photograph with
React Bits GradientASCII. The photograph and generated WebP derivatives are not
used. These decisions supersede the initial email-only page brief.

Confirmed program facts: anyone aged 18+ may apply; ambassadors host their events
using Hackways; work receives a letter of appreciation; highest contributors can
receive swag. The user announced results at the end of November. Compensation or additional benefits are not claimed.

## Direction contract

Mode: Persuade for the public landing; Operate for applications and admin review.
Seed c9425aa5 and the user's community-story choice established the reading path.
This is an extension scoped to the ambassador/program routes. Public-web keeps its main landing navigation on other routes. Ambassador routes
use a distinct responsive header with only the Hackways logo and a top-right Sign in
action. The header overlays the ASCII hero instead of reserving a white strip.
Foreground colors adapt across the bright upper field and dark lower field; after
the hero it gains a solid white content surface. While the renderer is initializing,
unavailable or context-lost, explicit fallback state keeps header text light over
the dark CSS fallback. Shader intro fade remains disabled so ready-state brightness and
header foreground agree immediately. A subtle 850ms scale settles the ready canvas;
reduced motion and the shared pause control disable it. It opens normal core sign-in with the application profile route as the safe
post-login destination; already signed-in users continue immediately. The footer
remains shared. Admin-web retains its existing console navigation.
Core-web program routes use a dedicated ProgramFrame applicant header/footer
(Programs, Your application and Your account), replacing AppChrome only for this
scoped journey. FG Futurist / Google Sans Flex and shared controls remain inherited.
Missing global DESIGN.md is pre-existing and is not repaired here.

The first viewport uses the requested blue/white/cyan ASCII horizon on deep navy,
centered invitation text, a pale primary action and a quieter section jump.
The initial art spans the entire available hero width with square corners. Scrolling
gradually reveals white gutters and rounded corners, matching the user-supplied
scrolled-state screenshot. A stable-height inset clip mask keeps the canvas mounted
and avoids resizing its drawing buffers during the transformation. A passive scroll
listener schedules only the settling animation, with no React renders per frame.
The effect reverses when scrolling back up, restores on pageshow and uses a static
boxed presentation for reduced motion. This is the campaign's one authored focal
moment, connecting the immersive invitation to the reading sections.
Gradient headline emphasis was explicitly requested. A soft oval radial scrim surrounds the copy instead of a full-width linear fade,
preserving the natural curved horizon and luminous edges of the source reference.
Hero spacing fits the overlaid header and brings actions upward. The source demo's
refresh watermark is intentionally not reproduced. No essential content waits for WebGL. The page scrollbar chrome is hidden only
while the ambassador page is present; wheel, touch and keyboard scrolling stay
enabled, with no overflow lock or custom scrolling implementation.
The remaining sections use white space, blue accents, ruled lists and a dark final
invitation. Reading order: invitation, community story, hosting role, application
steps, age eligibility, recognition, FAQs, joining details. Mobile preserves this
order and stacks two-column sections without horizontal scrolling by design.

## Hero implementation

`components/ambassadors/GradientASCII.tsx` adapts the upstream TypeScript version of
the user-supplied JavaScript/CSS component. Shaders and appearance come from React
Bits; source and license are retained beside the component. No runtime dependency
was added. The landing mounts one horizon instance with the requested palette,
18px cells and 0.8 density. Speed, flicker, glow and grain use the supplied defaults;
the previous near-opaque full-field dimmer was removed. DPR is 1.25 for the hero, rendering is capped at 30fps
and 2.5 million canvas pixels, with half-resolution field and quarter-resolution
halo buffers. It pauses offscreen and in hidden tabs; reduced-motion disables drift,
flicker, intro and pointer effects. A shared Toggle pauses/resumes animation. Context
loss reveals the static CSS fallback; restoration rebuilds WebGL resources. Cleanup
releases programs, textures, framebuffers, observers and event handlers.

The shader is decorative and aria-hidden. Text and links are server-rendered.
Forced-colors hides the art and uses system foreground/background colors. Shared
SectionJumpButton focuses the role without generating a fragment URL. FAQs use
native details/summary with reduced-motion marker behavior.

## Application and review contract

Join actions open core-web `/programs/ambassadors/apply/profile`, not mailto.
Separate profile, proposal and review routes save real drafts through global-backend.
Verified sign-in is required. Application name is a profile snapshot, not an account
identity edit. Extra questions are program-configurable. Before submission, the Go
backend checks the current program revision, application window, age attestation and
question validation. The application status page shows only the user's own record.

Admin-web Programs supports creation/settings, bounded searchable applications,
profile review, private notes and pending/accepted/rejected/waitlisted decisions.
Private decisions are invisible to applicants until explicit result release. Release
requires all submitted applications reviewed, a matching revision and recipient
count, confirmation and a durable idempotency key. Published decisions are locked.
Submission sends a confirmation and explicit result release sends an email update through the existing outbox. Private review still sends no decision notice. Withdrawal is explicit and irreversible in
this version. Input remains available after recoverable errors; expired-session
recovery uses another tab so editors retain answers.

## Verification boundaries

Local review uses public 4000, core 4001, global backend 4002 and admin 4003 on
loopback. The user subsequently authorized CI and production deployment. Actual release
state, checks and image identities are recorded in WORKLOGS.md; authorization is not
evidence of activation. The
preview database is an isolated copy of approved existing data; mandatory Go tests
use a separate `_test` database. Browser automation currently fails during kernel
initialization, so source/HTTP/type/lint/database checks do not imply rendered,
keyboard or WebGL visual approval. See WORKLOGS.md for executed results.

## Directory and overview refinement

The Programs directory is a compact browse view: 32px heading, 16px intro, left-aligned
program text and a 192px right thumbnail (104px on narrow screens). The user-supplied
`AMBASSADORS-HW-M.jpg` is preserved as its program image. It is not repeated at large
size on the overview. The overview describes the open role, community contribution
and recognition in short prose. Eligibility is not promoted, internal review steps
are omitted, and confirmed email updates replace the earlier no-email policy.
Application controls retain full touch targets and required LinkedIn/WhatsApp fields.


Mobile refinement keeps the existing type and ASCII palette but uses content height
instead of a viewport-height minimum. The headline begins after 164px of top space;
body copy is left aligned at 16px/1.65. The compact 44px pause toggle sits separately
at the upper-right, with a stable accessible toggle label. Mobile navigation adds a
contrasting translucent background. High-contrast mode preserves system colors.
The requested gradient headline remains the one intentional detector exception.
