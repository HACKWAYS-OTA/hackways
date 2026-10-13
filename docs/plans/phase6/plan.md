# Phase 6 Implementation Plan: Frontend Foundations

## 1. Metadata
- **Phase Identifier:** Phase 6 (`phase6`)
- **Title:** Frontend Foundations
- **Owner Role:** Frontend Lead / UI Engineer
- **Current Status:** `PLANNED`
- **Revision Date:** 2026-10-13

---

## 2. Purpose & Expected Outcome
Phase 6 establishes the frontend presentation layer across the four independent Next.js 14 applications (`public-web`, `core-web`, `organizer-web`, `admin-web`). It eliminates all mock data and browser alerts, integrates typed API client communication with backend engines, enforces accessible UI primitives without `lucide-react`, and implements genuine loading skeletons, empty states, and error envelopes.

**Expected Outcome:**
- All 4 independent Next.js applications scaffolded with standalone outputs.
- `apps/web/app/page.tsx` purged of hardcoded `SAMPLE_EVENTS` and `alert()` calls.
- Real server-side data fetching from Go backend endpoints.
- Reusable UI component library (`Button`, `Badge`, `Card`, `Modal`, `Input`) without `lucide-react`.
- Accessible, responsive layouts supporting mobile, tablet, and desktop viewports.

---

## 3. Scope & Non-Goals

### In Scope:
- Scaffolding the 4 independent Next.js 14 applications:
  - `apps/public-web`: SSG/ISR marketing and SEO catalogs.
  - `apps/core-web`: Attendee discovery and interactive RSVP portal.
  - `apps/organizer-web`: Organizer studio and attendee roster.
  - `apps/admin-web`: Auditing and platform telemetry.
- Connecting frontends to backend APIs via `@hackways/api-client`.
- Eliminating all mock data, dummy records, and simulated success states.
- Implementing first-render UX with Server Components to prevent layout shifts.

### Explicit Non-Goals:
- Introducing unapproved UI libraries or `lucide-react`.
- Implementing duplicate authorization or capacity checks in Next.js.
- Merging frontends into a single monolithic Next.js application.

---

## 4. Architectural Requirements & Constraints
- **Preserve 4 Independent Apps:** Merging the frontends is strictly prohibited ([`AGENTS.md` Rule 2.1](../../AGENTS.md#L94-L117)).
- **Zero Mock Data:** Must use real API responses from PostgreSQL ([`AGENTS.md` Rule 1.1](../../AGENTS.md#L9-L16)).
- **Zero `lucide-react`:** Icons must use inline SVGs or approved alternatives ([`AGENTS.md` Rule 1.2](../../AGENTS.md#L30)).
- **Memory Limits:** Each Next.js container must configure `NODE_OPTIONS="--max-old-space-size=..."` and `output: 'standalone'` ([`AGENTS.md` Rule 2.2](../../AGENTS.md#L71-L85)).

---

## 5. Prerequisites & Dependencies
- **Prerequisite Phase:** Phase 5 (Azure Integration Foundations).
- **Prerequisite Decisions:** DEC-005 (Approved icon alternative).

---

## 6. Ordered Implementation Steps
1. **Step 1 (Scaffold 4 Applications):** Establish `apps/public-web`, `apps/core-web`, `apps/organizer-web`, and `apps/admin-web` with standalone Next.js configurations.
2. **Step 2 (Purge Mock Data):** Refactor `page.tsx` to fetch real published events from `GET /api/v1/events`.
3. **Step 3 (Connect Shared API Client):** Import `@hackways/api-client` across all frontend apps.
4. **Step 4 (Implement UI Primitives):** Standardize `Button`, `Badge`, `Input`, `Dialog`, and `EventCard` across apps.
5. **Step 5 (First-Render UX & Skeletons):** Implement server rendering for critical metadata and accessible loading skeletons for dynamic slots.

---

## 7. Expected File Changes
- `apps/public-web/` (created)
- `apps/core-web/` (refactored from `apps/web`)
- `apps/organizer-web/` (created)
- `apps/admin-web/` (created)
- `packages/api-client/src/index.ts` (expanded methods)

---

## 8. Verification Commands & Evidence
- `cd apps/core-web && npm run type-check` -> PASS (exit code 0).
- `cd apps/public-web && npm run build` -> Standalone output emitted without memory spikes.
- Visual inspection verifying zero mock events and genuine API data.

---

## 9. Acceptance Criteria & Phase Exit Conditions
- [ ] 4 independent Next.js applications running standalone.
- [ ] Hardcoded `SAMPLE_EVENTS` and browser `alert()` completely purged.
- [ ] Real events fetched from Go API backend.
- [ ] Zero instances of `lucide-react`.
- [ ] Clean type-checking across all applications.
