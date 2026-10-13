# Engineering Context — Hackways

**Last Updated:** 2026-10-10  
**Current Milestone:** Milestone 1 (Multi-Service Production Topology & Infrastructure)  
**Target Environment:** Single VM (2 vCPU, 4 GB RAM, 50 GB NVMe SSD) with Cloudflare Edge CDN  

---

## 1. Product Scope

Hackways is a high-reliability event discovery, registration, and RSVP platform with zero tolerance for overbooking race conditions or unverified persistence.

### Active Scope (6-Service Production Architecture)
- **4 Next.js Web Runtimes:**
  - `Public Web` (`hackways.in`): High-speed marketing landing, SEO metadata, event catalogs.
  - `Core Web` (`app.hackways.in`): Discovery, event details, attendee registration, ticket portal.
  - `Organizer Web` (`organizer.hackways.in`): Event creation, attendee roster management, exports.
  - `Admin Web` (`admin.hackways.in`): Isolated platform auditing, user moderation, health telemetry.
- **2 Go Backend Engines:**
  - `Core Backend` (`core-backend:8081`): High-concurrency event lifecycle, ticket capacity limits, transactional RSVP locks (`SELECT ... FOR UPDATE`).
  - `Global API Backend` (`global-backend:8082`): Authentication, sessions, RBAC, background email notifications, media uploads.
- **Persistence & Ingress:**
  - PostgreSQL 16 (shared buffer 256MB, max 60 conns).
  - Nginx Reverse Proxy with rate limiting and subdomain routing.

---

## 2. Active Repository State

- **GitHub Remote:** `https://github.com/varshith-dev/HACKWAYS-IN.git`
- **Default Production Branch:** `main`
- **Active Integration Branch:** `develop`
- **Branching Policy:** Git Flow (`main`, `develop`, `feature/*`, `fix/*`, `release/*`, `hotfix/*`)
- **Key Reference:** `docs/ARCHITECTURE_PLAN.md` defines the complete system specification.
