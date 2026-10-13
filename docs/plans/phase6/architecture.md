# Phase 6 Technical Architecture: Frontend Services & Ingress Boundaries

## 1. Multi-Frontend Domain & Routing Architecture

```mermaid
flowchart TD
    subgraph IngressRouting ["Nginx Subdomain Ingress"]
        Nginx["Nginx Reverse Proxy"]
    end

    subgraph WebApps ["4 Independent Next.js 14 Applications"]
        PublicWeb["1. Public Web\n(Port 3000 | hackways.in)\nNext.js 14 SSG/ISR\nMarketing & SEO Catalogs\nMemory Limit: 140 MB"]
        CoreWeb["2. Core Web\n(Port 3001 | app.hackways.in)\nNext.js 14 App Router\nAttendee Discovery & RSVP\nMemory Limit: 250 MB"]
        OrgWeb["3. Organizer Web\n(Port 3002 | organizer.hackways.in)\nNext.js 14 App Router\nEvent Studio & Rosters\nMemory Limit: 230 MB"]
        AdminWeb["4. Admin Web\n(Port 3003 | admin.hackways.in)\nNext.js 14 App Router\nAudits & Telemetry (VPN Only)\nMemory Limit: 170 MB"]
    end

    subgraph GoEngines ["Authoritative Backend Engines"]
        CoreAPI["Core Backend (:8081)\nEvents & RSVPs"]
        GlobalAPI["Global Backend (:8082)\nAuth, Sessions, Uploads"]
    end

    Nginx -->|hackways.in| PublicWeb
    Nginx -->|app.hackways.in| CoreWeb
    Nginx -->|organizer.hackways.in| OrgWeb
    Nginx -->|admin.hackways.in [VPN]| AdminWeb

    CoreWeb -.->|Fetch /api/v1/events| CoreAPI
    CoreWeb -.->|Fetch /api/v1/rsvps| CoreAPI
    OrgWeb -.->|Fetch /api/v1/storage| GlobalAPI
    OrgWeb -.->|Mutate /api/v1/events| CoreAPI
```

---

## 2. Server-Side Rendering & Data Fetching Boundaries
- **Server Components (RSC):** Initial page loads fetch event data on the server via internal backend networking (`http://core-backend:8081/api/v1/events`). This guarantees zero layout shifts, high SEO scores, and fast first paint.
- **Client Components:** Used only where user interactivity is required (e.g. interactive capacity progress bar, modal dialogs, search filter dropdowns).
- **Session Cookie Forwarding:** Server Actions and Server Components read `__Host-hackways_session` from the incoming browser request and forward it directly to the Go API in the `Cookie` header.

---

## 3. UI Component System & Iconography
- All UI primitives reuse Tailwind CSS design tokens defined in `tailwind.config.ts`.
- **Iconography Rule:** Since `lucide-react` is strictly prohibited, all iconography is implemented as optimized, inline SVG React components (e.g. `CalendarIcon`, `MapPinIcon`, `TicketIcon`, `CheckIcon`, `SpinnerIcon`).
