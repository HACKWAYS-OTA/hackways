# Mobile navigation and event reading

Core-web keeps its existing four destinations and account/create actions. At
900px and below, the shared app header shows icons with accessible names/titles
and current-page state. The compact brand uses h.; desktop retains hackways.
The header allocates44px to the brand, a flexible four-icon nav, and88px to
create/account. Six-pixel minimum safe-area margins allow seven44px controls
at320px. Labels are visually hidden only; navigation remains semantic links.

The event-specific mobile header, active through640px, uses the same three-slot
layout. Brand/nav/actions explicitly occupy columns1/2/3 in row1. This overrides
the old .app-nav full-width second-row placement. The former width100% nav plus
nonshrinking siblings overflowed the viewport and could make zoomed-out content
look compressed into a narrow column. Content retains width100% and min-width0;
no overflow clipping, viewport zoom lock or user-scalable restriction was added.
Existing desktop event pinning and theme presentation remain unchanged.

Source was traced against the supplied screenshots. Actual browser/pinch-zoom,
responsive rendering and keyboard interaction have not been exercised in this
pass, following the user's earlier instruction to build without browser checks.
