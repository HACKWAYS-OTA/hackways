# CONTEXT.md — Hackways Current Engineering Context

**Last updated:** 2026-10-10  
**Current Milestone:** Milestone 0 (Production Repository & Architecture Baseline)  
**Active Focus:** Initializing production-grade repository standards, Git branching strategy, and GitHub integration.  

---

## Product Scope

Hackways is a production-grade event discovery, management, and RSVP registration platform.

### In Scope (MVP)
- Public event discovery, categorization, full-text search, and detail view.
- Event creation and management with state machine (draft -> published -> unpublished -> cancelled).
- Concurrency-safe attendee registration (guest and user) with capacity protection.
- Organizer dashboard for event and private attendee roster management.
- Transactional notification delivery (RSVP confirmation, event changes).
- Responsive, accessible Next.js UI for desktop and mobile.

### Deferred
- Paid ticketing and payment processing.
- Real-time chat.
- Redis (PostgreSQL satisfies caching, sessions, and queuing for early milestones).

---

## Active Repository State

- **GitHub Remote:** `https://github.com/varshith-dev/HACKWAYS-IN.git`
- **Default Branch:** `main` (Production)
- **Integration Branch:** `develop` (Integration)
- **Standards:** Defined in `AGENTS.md` and enforced via GitHub Actions CI.
