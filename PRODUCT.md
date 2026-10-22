# Hackways

<!-- impeccable:product-schema 1 -->

## Platform
web

## Users
Event organizers manage their events, invite and review attendees, configure registration questions and check guests in. Attendees discover events and register. Platform administrators use a separate console.

## Product Purpose
Manage real event registration and attendance with persisted records and authorized operations.

## Capabilities and Constraints
Four independent Next.js apps and two Go engines remain separate. Go and PostgreSQL own authorization, validation, event capacity, registration transitions and the transactional email outbox. Shared controls come from packages/ui; icons come from Hugeicons. Clean page paths are required. Production builds run off the VM.

## Brand Commitments
Hackways' existing wordmark, Google Sans Flex/FG Futurist fonts and application palette remain. The user requests a Kaggle-inspired, centered event management console at hackways.in/manage with the existing top navigation, no sidebar, concise content and recognizable action buttons. The public event footer is explicitly outside this redesign.

## Product Principles
- Use real APIs and persisted data; confirm success only after the operation succeeds.
- Make frequent guest and event actions clear and accessible.
- Preserve approval, capacity, ownership and audit guarantees.
- Keep forms understandable with meaningful input labels; remove editorial filler.

## Evidence on Hand
AGENTS.md, docs/ARCHITECTURE.md, the implemented six-service codebase and user-provided production screenshots. Interactive browser review requires a connected browser.
