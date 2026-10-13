# Phase 6 Implementation Guide: Frontend Foundations

## 1. Prerequisites
- Phase 5 complete (Go backend APIs active).
- Node.js 20+ and npm installed.
- Prohibited dependency `lucide-react` purged.

---

## 2. Step-by-Step Implementation

### Step 1: Establish 4 App Directory Layout
1. Relocate `apps/web` to `apps/core-web`.
2. Update port in `apps/core-web/package.json`: `"dev": "next dev -p 3001"`.
3. Scaffold `apps/public-web`: `"dev": "next dev -p 3000"`.
4. Scaffold `apps/organizer-web`: `"dev": "next dev -p 3002"`.
5. Scaffold `apps/admin-web`: `"dev": "next dev -p 3003"`.
6. Ensure every `next.config.js` specifies `output: 'standalone'`.

### Step 2: Refactor `core-web/app/page.tsx`
1. Remove `SAMPLE_EVENTS` array and `"use client"` directive from the page root.
2. Convert `HomePage` to an async Server Component:
   ```tsx
   import { fetchEvents } from '@/lib/api';
   import { EventCard } from '@/components/events/EventCard';

   export default async function HomePage() {
     const events = await fetchEvents();
     return (
       <div className="max-w-7xl mx-auto px-4 py-10">
         <h1 className="text-4xl font-bold text-white mb-6">Upcoming Events</h1>
         <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
           {events.map((evt) => (
             <EventCard key={evt.id} event={evt} />
           ))}
         </div>
       </div>
     );
   }
   ```
3. Wrap dynamic sections in `<Suspense fallback={<EventCardSkeleton />}>`.

### Step 3: Implement SVG Icon Set
1. Create `components/ui/icons.tsx` containing SVG primitives.
2. Replace any icon placeholders with native SVG components.

### Step 4: Verification Checks
1. Run `npm run type-check` across all 4 frontend directories.
2. Verify zero runtime errors.

---

## 3. What Must NOT Be Changed in Phase 6
- Never install `lucide-react`.
- Never replicate Go capacity locking logic inside Next.js Server Actions.
- Never write hardcoded mock arrays to simulate backend APIs.
