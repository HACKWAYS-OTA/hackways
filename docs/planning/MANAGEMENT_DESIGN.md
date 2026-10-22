# Management console local review

Scope: event management at `/manage`, confirmed by the user. Keep top navigation without a sidebar, centered page headings, Hackways fonts and Hugeicons, and explicit action buttons. Preserve meaningful input labels, real data and consequential action warnings. Remove editorial filler.

The overview now groups attendance, guest actions, activity and event details. Guest-list Approve/Decline persist directly through the existing Go status API with revision and idempotency checks. Identity buttons still open details. Forward navigation into event details/registration resets scroll; history and query-only navigation retain their existing behavior. The public event footer is unchanged.

Local management: `http://localhost:3102/manage/login`. Attendee review: `http://localhost:3101/discover`. The apps use a separate local PostgreSQL copy through the two Go engines. Email delivery, uploads and Google OAuth are not configured in this preview. Production deployment is held for the user's review.

Both affected applications passed TypeScript and lint. Sign-in, catalogue, health and management font HTTP probes passed. No browser is connected; visual and interactive review remain pending.
