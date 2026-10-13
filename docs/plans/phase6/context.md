# Phase 6 Agent Pre-Context: Frontend Foundations

## 1. Product Objective & Technical Target
Phase 6 establishes the frontend presentation layer across the **four independent Next.js 14 web applications**. It enforces true server-side rendering, zero mock data, zero layout shifts, and accessible component primitives while respecting strict memory caps.

---

## 2. Canonical Specifications to Obey
1. **[`AGENTS.md` Rule 1.1](../../../AGENTS.md#L9-L16):**
   - **Never use mock data, dummy records, or simulated success states in application functionality.**
   - All features must connect to actual Go APIs and PostgreSQL.
2. **[`AGENTS.md` Rule 1.2 & Rule 3.12](../../../AGENTS.md#L30):**
   - **Never introduce `lucide-react` anywhere in Hackways.**
   - Do not introduce duplicate TypeScript backend routing or business logic inside Next.js.
3. **[`AGENTS.md` Rule 1.4](../../../AGENTS.md#L52-L65):**
   - Diagnose actual rendering paths; fetch essential data on server before rendering.
   - Use clear loading skeletons only when data is genuinely pending. Prevent content shifts.
4. **[`AGENTS.md` Rule 2.1](../../../AGENTS.md#L94-L117):**
   - Preserve all 4 independent applications: `public-web`, `core-web`, `organizer-web`, `admin-web`. Never merge them.

---

## 3. Verified Repository Facts Before Phase 6
- Current frontend exists in a single directory `apps/web`.
- `apps/web/app/page.tsx` renders hardcoded `SAMPLE_EVENTS` and fires `alert()`.
- Component library has primitives `Button.tsx` and `Badge.tsx`.

---

## 4. Expected Outputs
- 4 clean Next.js 14 standalone applications running on ports 3000, 3001, 3002, 3003.
- Complete removal of `SAMPLE_EVENTS` and `alert()`.
- Real event data rendered via Server Components.
