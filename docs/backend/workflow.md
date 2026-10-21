# RSVP Platform — Workflow Specification

Source: *RSVP Platform – End-to-End UX Possibility Map* (PDF), plus system-design research (Part C).
Product entry point named in the map: **Hackways** ("Open Hackways").

Contents
- **Part A — Workflows** (every flow in the map, step by step, with decisions, failure branches and system behaviour)
- **Part B — State machines** (event, registration, payment, ticket, waitlist, organiser verification)
- **Part C — System design** (architecture, data model, hard problems, stack, roadmap)
- **Part D — Gaps found in the map**

---

# PART A — WORKFLOWS

## 0. Actors and swimlanes

The map has three lanes plus shared system layers.

| Lane | Who | Primary goals |
|---|---|---|
| **Attendee** | Guest or signed-in user | Find an event, register or pay, get a ticket, attend, follow up |
| **Organiser** | Verified host or team member | Create, publish, promote, run, settle and close events |
| **Platform admin** | Staff | Trust and safety, payments and disputes, reliability, support, configuration |
| **System (cross-cutting)** | Platform services | Identity, notifications, idempotency, audit, accessibility, consent and retention |

Shared principle used all through the map: **every failure branch has a named recovery path** (retry, alternative, explanation, support route). No dead ends.

---

## 1. App entry and session resolution

**Trigger:** the user opens the product. The map asks "How is the product opened?"

| Entry type | What happens |
|---|---|
| Fresh install / first visit | Load app shell, then first-run experience |
| Returning session | Restore previous session, then validate |
| Deep link, invite or QR | Resolve destination link, then validate |
| Push or email notification | Open notification, then resolve its destination |
| Shared event or organiser page | Load shared public page destination |

**Decision: Session valid?**
- **Valid:** render intended destination, or Home and discovery.
- **Expired or revoked:** continue as guest and preserve the destination, or sign in.
- **Account disabled:** explain the restriction and give an appeal or support route.

**Decision: Destination valid and accessible?**
- **Yes:** render it.
- **Expired, private, deleted or malformed:** explain the state and route to relevant discovery or help.
- **Slow network or server error:** retry, show a cached preview, or a safe fallback.

**Decision: App state?**
- **Online:** continue normally.
- **Offline or degraded:** show cached content, retry state and an offline notice. On reconnect, "Connection restored?" triggers refresh and reconciliation of queued actions.
- **Maintenance or unsupported version:** show status, update path and retry guidance.
- Compatibility note in the map: mobile, desktop, browser and version.

**First-run experience:** "Intro or permissions needed?" Explain value and ask only necessary permissions (location, notifications). The user can skip.

**Design notes**
- Deep links must survive authentication. The map says "Prompt sign-in and preserve intended action". Store the intended action server-side or in signed state, not only in the client.
- Public pages should be server-rendered or cached so they work on slow networks and are shareable.

---

## 2. Identity and account flows

**Trigger:** any action that needs an account (register, save, follow, create community, organiser access), or an explicit sign-in.

**Step 1 — Identity method.** Options in the map:
1. Email magic link or OTP (with email verification)
2. Google or other OAuth (authorise provider)
3. Passkey or other supported method
4. Continue as guest (only where "Account or guest registration allowed?" is yes)

**Step 2 — Existing or new account?**
- **Existing:** sign in.
- **New:** create account and verify.

**Step 3 — Authentication outcome**
| Outcome | Next |
|---|---|
| Success | Return to original action and destination |
| Expired, invalid or blocked | Recover account or change email, or use alternate secure method |
| OTP, email or OAuth failure | Explain the issue, retry or pick another method |
| Abandon | Leave; preserve nothing sensitive |
| Forgot password or lost access | Account recovery, escalating to Account recovery support case |

**Account required vs guest:** if the event requires an account (or approval, invite, paid with identity), the flow says "Account required", then sign-in prompt, then resume. If guest is allowed, continue as guest and attach to an account later via email.

**Organiser access gate:** "Organiser workspace access?" asks whether the user is signed in or a verified organiser.
- **No:** sign in, request organiser access or verify identity.
- **Yes:** enter the organiser workspace.

**Edge cases to specify**
- Same email via magic link and via Google: merge into one account.
- Guest registrations made before sign-up: claim by verified email.
- Rate limit OTP requests and verification attempts per identifier and per IP.

---

## 3. Discovery and browsing (Attendee)

### 3.1 Discovery entry points
Home and discovery offers: **Browse listings and filters, Search, Near me, Trending or featured, Calendar or date, Categories, Community feed, Online events, Map view, For you (personalised feed)**.

**"What does the visitor want?"** routes to: Find an event, Find an organiser, Find a community, Event discovery by venue or city, Create or manage event, Profile or settings, Contact host or view FAQ, Save for later, View saved events.

### 3.2 Global search
**Search target:** event, organiser or host, community, or place and location.
- Search event names, topics, speakers.
- Search organisers and hosts, then organiser results.
- Search communities, then community results.
- Search places, then choose city or enter location.

**Results state**
| State | Behaviour |
|---|---|
| Results found | Show list; "Mixed quality or sponsored results" requires labelling promoted content and explaining ranking |
| No results | Suggest broader terms, nearby dates or categories |
| Loading or error | Retry, retain filters, show cached results |

**Refine:** filters (date, format, price, category, distance, language, access); sort (relevance, soonest, newest, popularity); clear filters; change query.

### 3.3 Location permission
Decision "Location permission?":
- **Granted:** "Near me" and distance sorting.
- **Denied or not requested:** choose a city or enter a location manually.

### 3.4 Card actions
Every event card offers: **Open details, Quick RSVP, Save, Share, Report event**.
- **Quick RSVP:** run "Check quick RSVP eligibility". If eligible, short path to registration; if not, open the detail page.
- **Save:** add to saved items, saved events and collections; user can remove or organise collections.
- **Share:** copy link, native share, QR.
- **Report:** report misleading, unsafe or spam event, into Trust and safety (flow 14).

### 3.4b Unavailable items
If the card target is removed or reported, show a **safe unavailable state with a discovery fallback**.

### 3.5 Organiser and host profiles
Organiser profile shows credibility and event catalogue. Visitor action: **Follow, Browse events, Share, Contact host**. Follow or unfollow feeds the personalised feed.

### 3.6 Communities (discovery side)
Community profile and event feed. Actions: join or leave, follow or unfollow, view events, discuss or connect, update feed preferences. "Creation permission and verification?" gates **Create community** (profile, rules, moderators, links). Not allowed leads to "Request access or learn requirements".

---

## 4. Event detail page

**Decision: Event status?** Determines what the page shows.

| Status | Page behaviour |
|---|---|
| Published and upcoming | Overview, agenda, speakers, venue, host, policies; Register or buy ticket, Join community, Save |
| Full or registration closed | Explain waitlist, sold out, closed or eligibility; show waitlist, notify-me, or similar events |
| Postponed or venue changed | Show latest details and attendee notice |
| Cancelled | Show cancellation, refund and replacement information |
| Past event | Recordings, materials, community, future editions; attendance history and certificates |
| Draft, private, hidden, ended or removed | Explain status and suggest alternatives; invite-only shows an explanation of unavailable access |
| Removed or reported | Safe unavailable state |

Also on this page: **Ask organiser**, **Contact host or view FAQ**, **Report event**, **Share**, **Add to calendar**, **Related events**.

**Gate: "Eligibility, age, region and capacity valid?"** Fails lead to "Explain requirements and possible alternatives".

---

## 5. Registration and ticket flow (the core transaction)

### 5.1 Pre-checks
1. **Signed in?** If no, prompt sign-in and preserve intent, or continue as guest if allowed.
2. **Eligible and registration open?** If no, explain rules and offer alternatives.
3. **Capacity still available?** If no, **"Waitlist or alternative allowed?"**
   - Yes: join waitlist or choose another ticket (flow 5.6).
   - No: explain sold out or closed; exit, save, share, or follow organiser.
4. **Participation mode** for hybrid events: online or in person.
5. **Choose ticket or RSVP type** (ticket, seat, group, RSVP type).

### 5.2 Branch by registration type
Decision "Free, paid, donation, approval or invite?" (variants appear in the map):

| Type | Path |
|---|---|
| **Free and instant** | Answer form and consent questions, validate, create registration, confirmation |
| **Approval required** | Answer form, submit for organiser approval, show pending status and next steps (see 5.5) |
| **Paid** | Review price, fees, tax, refund policy, then payment (see 5.4) |
| **Donation / pay what you want** | Enter amount and review terms, then payment |
| **Invite code** | Check "Code valid and unused?" Invalid: retry code, request invite or return |

### 5.3 Form and consent
- Answer required and optional fields.
- Decision "Form valid?" / "Validation and consent pass?" If no: **show field-level errors and preserve input** (inline errors).
- Consent questions captured with timestamp and version (feeds retention and export).

### 5.4 Payment
1. Review price, fees, currency, taxes, refund rules.
2. Choose a supported payment method.
3. **Create provisional registration** and **reserve idempotently**; hold the slot briefly.
4. **Payment result?**

| Result | Handling |
|---|---|
| **Success** | Commit registration, confirmation, ticket, QR, receipt |
| **Declined or timeout** | Retry, change method, or keep slot briefly |
| **Abandoned** | Resume checkout, or release inventory hold and save a recovery link |
| **Duplicate or uncertain** | **Verify provider status before retry**; check existing registration; recover without duplicate charge |
| **Provider conflict** | Show payment support case option |

5. **"Capacity and submission commit succeeds?"** If yes, confirm. On conflict or server error: **idempotent retry, restore answers, explain outcome**. Duplicate request, stale data and conflict are handled by checking the existing registration.
6. **"Registration committed?"** Yes leads to the confirmation. No leads to recovery.

### 5.5 Approval flow
- Attendee sees pending state and expected next step.
- **Organiser decision?**

| Decision | Effect |
|---|---|
| Approved | Notify attendee; if paid, prompt payment (or charge) before expiry; issue ticket |
| Rejected | Notify with reason where permitted; refund if applicable |
| No decision / timeout | Show pending status; auto-expire or escalate per event policy |
| Request edits | Attendee updates form and resubmits |

### 5.6 Waitlist
- **Join waitlist or accept offer.** Attendee can manage waitlist position and offer expiry; keep place, leave list, or edit preferences.
- **"Waitlist offer arrives?"** Yes: notify attendee, **accept before expiry**; no response: offer expires, next person is promoted.
- Accepted offer re-enters the payment or confirmation path.

### 5.7 Confirmation
Show confirmation, ticket, QR and receipt. Delivery by email, SMS, push, in-app. **"No notification delivered?"** leads to in-app status, resend, and preference controls.

### 5.8 Registration support
If the attendee is stuck, route to **Registration support case** or **Payment support case**.

---

## 6. Managing registrations (My events and tickets)

**Surface:** "My registrations", "My events and tickets", **Ticket wallet and attendee details**.

**Decision: What changes?**

| Action | Flow |
|---|---|
| **View QR or ticket** | Show ticket; works offline via cached or wallet copy |
| **Edit attendee details** | Validate changes and audit update; update registration |
| **Cancel or request refund** | "Cancellation allowed?" Yes: check policy and refund eligibility, confirm cancellation and refund timeline, cancel, release capacity, notify parties. No: explain deadline and offer support |
| **Transfer ticket** | "Transfer permitted?" Yes: verify recipient and transfer rules, **revoke old ticket and issue new ticket** (replacement QR). No: explain |
| **Add to calendar or share** | Calendar invite, share link, reminders; handle calendar sync and time-zone |
| **Accessibility or dietary request** | Submit request and **protect private data** (restricted visibility to organiser) |
| **Manage waitlist** | See 5.6 |
| **View past events** | Attendance history, certificates, materials |

---

## 7. Reminders and event changes

**Decision: Reminder or event change?**
- **Reminder sent:** scheduled reminder via email, SMS, push or in-app.
- **Time, venue or agenda change:** notify registered attendees and show updated details (also update calendar details).
- **"Change materially affects attendees?"** (organiser side, flow 11.6) If yes, confirm impact, refunds and notifications.
- **"No notification delivered?"** Show in-app status, allow resend and preference controls. Channel failure triggers fallback to another channel.

Preference controls: email, push, SMS, frequency.

---

## 8. Event day

### 8.1 Pre-event context
Arrival context: venue, directions, accessibility, check-in information. For online: join link, device check and access window.

### 8.2 Online event
Join link, device check and access window, then open meeting or streaming room. Attend, network, ask questions, join activities. Consent, privacy, abuse: block and report tools available.

### 8.3 In-person check-in
**Arrive and check in:** scan QR or staff lookup.

**Decision: Ticket valid and attendee matches?**

| Outcome | Action |
|---|---|
| **Valid** | Check in once; show success; admit and mark attendance once |
| **QR invalid, duplicated, offline** | Manual lookup, duplicate warning, offline sync |
| **Wrong event or already used** | Explain issue; **staff override with audit trail** |
| **Invalid or already used** | Staff lookup, duplicate warning and audited override |

### 8.4 Hybrid event
Participation mode (online or in person) chosen at registration; both check-in methods apply, and attendance is tracked per mode.

---

## 9. Post-event experience

- Feedback, certificate, recordings, resources.
- Similar events, organiser follow, community.
- **Discover related events or unsubscribe.**
- Past event page: recordings, materials, future editions.

---

## 10. Profile, settings, privacy and account deletion

**Profile task**
- Edit profile and interests (name, bio, topics, location, preferences).
- Notification preferences (email, push, SMS, frequency).
- Privacy and security (sessions, devices, consent, data export and deletion).
- Blocked users or content (manage blocks and safety settings).
- Language and accessibility (language, contrast, motion, assistive settings).
- **Delete account.**

**Delete account flow**
1. Confirm consequences and retention requirements.
2. **"Deletion allowed and confirmed?"**
   - **Yes:** revoke sessions and process deletion per policy.
   - **No, or legal hold:** explain retained records and support path.
3. Edge: user is an organiser with live events or unsettled payouts. Block deletion until resolved or transferred.

---

## 11. Organiser workflow

### 11.1 Access
"Organiser workspace access?" See flow 2. Verification: **"Verify organiser identity and ownership"**, with "Verification result?" of verified, needs more evidence (request documents and pause risky capabilities), or rejected.

### 11.2 Workspace
Select event workspace. **Workspace task** options:
Create event, Manage existing event, Review registrations, Manage team, Finance, Insights, Support and reports.

### 11.3 Create event
1. **Start event creation.** Decision: "Create from template or blank?"
   - **Template:** select template and event type (Hackathons, meetups, conferences, workshops, culture).
   - **Blank:** start empty.
   - **Duplicate or recurring:** clone safely or configure recurrence.
2. **Enter event basics:** title, description, type, schedule, time zone, format, venue, accessibility, capacity.
3. **Configure tickets and policies:** capacity, tickets, pricing, questions, approval, waitlist and refund/other policies.
4. **Preview attendee-facing page** and validate completeness.

### 11.4 Publish
**Decision: "Publish now, schedule or save draft?"**
- **Draft:** save draft; show missing fields and policy warnings.
- **Publish / Schedule:** run **"Ready and compliant to publish?"** (validation and moderation checks).

| Check result | Next |
|---|---|
| Pass | Publish event, or schedule publishing and show scheduled status and edit window |
| Needs changes | Explain errors, missing fields or policy issue |
| Review required | Queue platform review (flow 12) |
| Rejected or system error | Explain moderation or validation issue; retry |

**"Publish result?"** Live, Scheduled, or Failed (retry). **"Event visible and valid?"** confirms the public page is reachable.

### 11.5 Promote
Share link, QR, embed, social assets; campaigns, share links, embeds and referral tracking.

### 11.6 Edit a live event
Edit and preview changes. **"Material attendee impact?"** (time, venue, price, format)
- **Yes:** confirm impact, refunds and notifications; notify attendees and update calendar details.
- **No:** save quietly.
Validate changes and update registrations. Assess attendee impact before saving.

### 11.7 Cancel or postpone
Confirm impact, refunds and notifications. Cancel releases inventory, triggers refunds per policy, notifies all parties, and updates the public page to Cancelled with replacement information.

### 11.8 Operate event
- **Review registrations:** attendee list, approvals, waitlist and check-in; live attendee list and capacity update.
- **Capacity threshold** trigger: close sales, raise capacity, or enable waitlist.
- **Question or support ticket:** respond with attendee context.
- **Risk or abuse report:** review report and apply scoped action.
- **New registration:** live list and capacity update.
- **Payment issue or refund:** reconcile payment, refund and ledger.
- **Close registration** and **Run event.**
- **Incoming operational event?** (the map's event router): new registration, capacity threshold, payment issue, question, risk report, event changes.

### 11.9 Manage team
Invite team; configure roles and permissions. Roles include check-in staff.

### 11.10 Finance
Payments, refunds, payouts and invoices; tax records. Reconcile payment, refund and ledger.

### 11.11 Insights
Views, funnel conversion, RSVP, attendance and sources. Attendance, conversion, revenue, exports.

### 11.12 Support and reports
Handle attendee questions, incidents and reports.

### 11.13 End event and close operations
End event and close operations; payouts, invoices, refunds, tax records. **Archive, duplicate, or create recurring event.** Post-event follow-up: feedback, certificate, materials, community.

---

## 12. Platform admin workflow

**Admin responsibility** splits into: Trust and safety, Payments and disputes, Verify organisers, Moderate content, User and organiser support, Reliability, Platform configuration (roles, permissions, limits, feature flags).

### 12.1 Review new or reported events / content moderation
"Suitable for attendee?" and **Policy decision?**
- **Approve:** event goes or stays live.
- **Request edits:** organiser edits and resubmits.
- **Reject or suspend:** restrict event or organiser; record reason.
- **Appeal received?** If yes: independent review and outcome notice. Decision, reason, appeal and audit trail are recorded.

### 12.2 Organiser verification
Identity and ownership review. Verification result: Approved; More evidence (request evidence and restrict risky actions); Rejected.

### 12.3 Trust and safety (fraud, spam, bots, duplicate accounts)
**Risk threshold?** Low, Medium, High.
| Level | Action |
|---|---|
| Low | Allow and monitor |
| Medium | Rate limit, CAPTCHA, extra verification; or rate limit or require verification |
| High | Block transaction, suspend access, or remove content |

Spam, fraud, harassment and abuse response. Safety report and escalation; case status and communications.

### 12.4 Payments and disputes
Reconcile payments, refunds and chargebacks. Payment disputes and escalations: investigate, communicate, refund or contest.

### 12.5 Support
Review escalated cases and appeals. Support case types seen in the map: Accessibility support, Account recovery support, Safety or abuse, Platform bug, Payment support, Registration support. Help centre and support; technical incident report.

### 12.6 Reliability
Service health and incident response; monitor outages, degraded modes and recovery; status communication. Audit log and case closure.

### 12.7 Platform configuration
Roles, permissions, limits and feature flags.

---

## 13. Support and safety (user-facing)

**Get help / Help or report a problem** routes by type: Contact or report, Support issue, Safety or abuse, Accessibility, Account access, Platform bug, Payment or refund, Registration or ticket.

- **Report content or block account** (Report or block).
- **Safety report and escalation**, then **case status and communications**.
- **Explain policy and support escalation** when automated resolution is not possible.
- **Technical incident report** for bugs.

---

## 14. Cross-cutting system behaviours

| Concern | Requirement from the map |
|---|---|
| **Idempotency, audit logs, state consistency** | Every write that moves money or capacity is idempotent and audited |
| **Notifications** | Email, SMS, push, in-app delivery; handle delivery failures; resend; preferences |
| **Network** | Network loss, retries, duplicate requests handled without duplicates |
| **Accessibility** | Keyboard, screen reader, contrast, language, motion, mobile |
| **Privacy and data** | Consent, retention, export, deletion; protect private data in requests |
| **Identity** | Identity, sessions, account deletion |
| **Time** | Calendar sync and time-zone handling |
| **Fraud** | Rate limit, CAPTCHA, extra verification |
| **Compatibility** | Mobile, desktop, browser and version |

---

# PART B — STATE MACHINES (derived from the map)

The map implies states but never names them. These are the states to implement.

### Event
`DRAFT → IN_REVIEW → (NEEDS_CHANGES ↔ DRAFT) → SCHEDULED → LIVE → ENDED → ARCHIVED`
Side states: `POSTPONED`, `CANCELLED`, `SUSPENDED` (admin), `REMOVED` (admin). Private or hidden is a visibility flag, not a state.

### Registration
`PROVISIONAL (hold) → CONFIRMED → CHECKED_IN`
Branches: `PENDING_APPROVAL → APPROVED | REJECTED | EXPIRED`; `WAITLISTED → OFFERED → CONFIRMED | EXPIRED`; `CONFIRMED → CANCELLED | TRANSFERRED`; `PROVISIONAL → RELEASED` on timeout or abandon.

### Payment
`INITIATED → PENDING → SUCCEEDED | FAILED | ABANDONED`
Then `SUCCEEDED → REFUND_PENDING → REFUNDED | PARTIALLY_REFUNDED`, and `SUCCEEDED → DISPUTED → WON | LOST`. `UNCERTAIN` is a real state: reconcile with the provider before any retry.

### Ticket
`ISSUED → SCANNED (used) | REVOKED (transfer, cancel) | VOID (event cancelled)`

### Waitlist entry
`ACTIVE → OFFERED → ACCEPTED | EXPIRED | LEFT`

### Organiser verification
`UNVERIFIED → SUBMITTED → VERIFIED | NEEDS_EVIDENCE | REJECTED`, with `SUSPENDED` as an admin overlay.

### Account
`ACTIVE → DISABLED (appeal path)`, `ACTIVE → DELETION_PENDING → DELETED` or `LEGAL_HOLD`.

---

# PART C — SYSTEM DESIGN RESEARCH

## C.1 What published designs agree on

Research across ticketing system-design write-ups and event-platform documentation points to the same core problems:

1. **Inventory consistency is the crux.** Ticketing guides consistently say to lead with preventing double-booking: seat or capacity inventory needs strong consistency, while browsing can be eventually consistent.
2. **Hold vs confirmed are different states.** A reservation holds inventory for a limited window (typically around 10 minutes) while payment completes; only payment confirmation makes it permanent. Your map's "keep slot briefly" and "release hold" map directly to this.
3. **Split read and write paths (CQRS).** Browse and search hit replicas or a search index (Elasticsearch-style); only reservations hit the primary transactional store. A 2026 academic paper on a distributed booking system used a CP-oriented booking subsystem with Redis locking plus an AP-oriented search subsystem, with Kafka for async messaging and virtual queuing, and reported zero double-bookings in contention tests.
4. **Virtual waiting room for spikes.** For flash demand, a queue meters admission to match checkout capacity rather than letting a thundering herd hit inventory. Admission rate is tuned from remaining inventory, checkout capacity and conversion rate.
5. **Payment confirmation via provider webhooks.** Bookings move to CONFIRMED on the provider's success webhook, not on the client's redirect. This is why your map has "Verify provider status before retry".
6. **Signed QR tickets.** Tickets carry a signed token (event, ticket, expiry) and are marked scanned on first use to block reuse; duplicate scans and bulk-purchase patterns feed fraud detection.
7. **Event platforms expose webhooks.** Luma's public docs describe webhooks for event created, updated, canceled, and guest registered/updated, with signature headers and timestamp checks against replay. Worth copying as an integration surface.

Scale reality check: Ticketmaster-class flash sales (millions of requests in seconds) are not your problem on day one. Most RSVP platforms are read-heavy with modest write bursts around popular events. Design for correctness first, then add the queue only when a single event can saturate checkout.

## C.2 Recommended architecture

Start as a **modular monolith**, split by bounded context, with clean seams so services can be extracted later.

```
Clients (Web / Mobile / Staff scanner)
        │
   CDN + Edge cache  (public event pages, assets)
        │
   API Gateway / BFF  (auth, rate limit, WAF, bot checks)
        │
 ┌──────┼───────────────────────────────────────────────┐
 │ Identity │ Events │ Registration │ Payments │ Discovery │
 │ Notify   │ Check-in │ Community  │ Trust&Safety │ Admin │
 └──────┬───────────────────────────────────────────────┘
        │
 Postgres (system of record)   Redis (holds, rate limits, cache)
 Search index (events/orgs)    Object storage (media, exports)
 Event bus / queue (outbox → workers)
        │
 Workers: notifications, payouts, reconciliation, moderation, search indexing
```

| Context | Responsibility | Key map flows |
|---|---|---|
| Identity | Accounts, OTP/magic link, OAuth, passkeys, sessions, deletion | 2, 10 |
| Events | Event CRUD, publish pipeline, recurrence, templates | 11 |
| Registration | Capacity, holds, forms, approvals, waitlist, transfer, cancel | 5, 6 |
| Payments | Provider integration, ledger, refunds, payouts, tax | 5.4, 11.10, 12.4 |
| Discovery | Search, filters, geo, feed, recommendations | 3 |
| Check-in | Ticket validation, offline sync, override audit | 8 |
| Notifications | Templates, channels, preferences, delivery tracking | 7 |
| Community | Communities, membership, follow | 3.6 |
| Trust and safety | Risk scoring, moderation queue, reports, appeals | 12 |
| Admin | Console, feature flags, audit | 12 |

## C.3 Core data model (minimum)

- `users`, `identities` (method, provider id), `sessions`, `consents`
- `organisers`, `organiser_members` (role), `organiser_verifications`
- `events` (status, visibility, timezone, format, capacity), `event_versions` (audit of edits), `venues`
- `ticket_types` (price, currency, quota, sales window, approval flag), `questions`
- `registrations` (state, attendee, ticket_type, answers, idempotency_key), `registration_holds` (expires_at)
- `waitlist_entries`, `waitlist_offers` (expires_at)
- `tickets` (signed code, state), `checkins` (who, when, method, override reason)
- `payments`, `ledger_entries` (double-entry), `refunds`, `payouts`, `disputes`
- `communities`, `memberships`, `follows`, `saved_items`
- `notifications` (channel, status, attempts), `reports`, `moderation_cases`, `appeals`
- `audit_log` (append-only)
- `outbox` (transactional events to the bus)

## C.4 The hard problems and how to solve them

### 1. Never oversell capacity
- Keep a transactional capacity counter per ticket type in Postgres. Reserve with a single atomic statement (conditional update: decrement only if remaining > 0) or `SELECT ... FOR UPDATE` on the ticket type row.
- Redis can front it for speed (atomic decrement or a lock with expiry), but **Postgres stays the source of truth**. Reconcile counters periodically.
- Holds expire via TTL plus a sweeper job that releases expired holds and promotes the waitlist.

### 2. Idempotent registration and payment
- Client sends an `Idempotency-Key` per checkout attempt; store it with a unique constraint on `(user, event, key)`.
- Payment provider calls are idempotent with the same key.
- **Never trust the client redirect.** Confirm on the provider webhook, and also poll or reconcile for "duplicate or uncertain" outcomes.
- Use the **transactional outbox**: write state change and outgoing event in one DB transaction, then publish.

### 3. Payment and refund correctness
- Double-entry ledger; every money movement is two entries.
- Refund policy engine evaluates cancellation windows server-side.
- Daily reconciliation job against provider reports; surface mismatches to admin (map: "Reconcile payments, refunds and chargebacks").
- Handle currency and tax per region (relevant for India: GST on ticket sales and platform fees, and payout KYC; confirm with a tax professional).

### 4. Waitlist
- Ordered queue per ticket type. On capacity release, create an offer with an expiry timer, notify, and move on if it expires. Keep promotion in a single worker per ticket type to avoid races.

### 5. Search and discovery (read path)
- Index published events into a search engine (text, geo-distance, date, price, category, language, accessibility filters). Updates flow through the outbox, so the index is eventually consistent.
- Use read replicas and CDN caching for event pages. Invalidate on edit.
- Label sponsored results; keep ranking explainable (map requirement).

### 6. Check-in at scale and offline
- Signed QR payload (event id, ticket id, expiry, signature). The scanner verifies the signature **offline** and records scans locally.
- Pre-sync the attendee list to staff devices; sync scans back and resolve duplicates by earliest timestamp. Conflicts raise the "duplicate warning" and audited override from the map.

### 7. Notifications
- Preference-aware router across email, SMS, push, in-app; per-channel retry with backoff; fallback channel; delivery status stored so the UI can show "No notification delivered" and resend.
- Event changes that materially affect attendees are sent with priority.

### 8. Trust and safety
- Risk score at signup, event publish, and checkout (velocity, device, email age, payment signals).
- Tiered responses straight from the map: allow and monitor, rate limit/CAPTCHA/verify, block or suspend.
- Moderation queue with SLA, decision reasons, appeals, and immutable audit trail.

### 9. Traffic spikes (add when needed)
- Virtual waiting room in front of checkout for a single hot event; admit users at a rate tied to remaining inventory and checkout throughput.
- Static event pages from CDN so the spike hits only the reserve endpoint.

### 10. Privacy, retention, deletion
- Consent records versioned; data export job; deletion with legal-hold exceptions; accessibility and dietary data in a restricted table visible only to the necessary organiser roles.

### 11. Reliability and observability
- Health checks, SLOs (for example 99.9% on reserve/checkout, higher on browse), structured logs with correlation ids, tracing across reserve → payment → webhook, dashboards for hold expiry, webhook lag and reconciliation drift.
- Degraded modes from the map: read-only browsing, cached pages, queued notifications.

## C.5 Suggested stack

Choose by team familiarity; this fits a Node/FastAPI background:

| Layer | Option |
|---|---|
| API | NestJS (Node) or FastAPI; REST first, add GraphQL only if clients demand it |
| Database | PostgreSQL (primary), read replica later |
| Cache, locks, rate limit | Redis |
| Search | Elasticsearch/OpenSearch, or Postgres full-text + PostGIS to start |
| Queue / bus | Start with a Postgres-backed or Redis queue (BullMQ); move to Kafka/RabbitMQ at scale |
| Payments | Razorpay or Stripe (webhooks, idempotency keys) |
| Auth | Own OTP/magic link + OAuth + WebAuthn passkeys, or a managed provider |
| Media/exports | Object storage + CDN |
| Clients | React web, Flutter mobile and staff scanner (offline-capable) |
| Infra | Containers on your preferred cloud; IaC; feature flags |

## C.6 Build order (roadmap)

1. **Foundation:** identity, events CRUD, publish pipeline, event page, free RSVP, confirmation email.
2. **Registration depth:** capacity holds, forms, approval, waitlist, cancellation, My tickets.
3. **Payments:** checkout, webhooks, ledger, refunds, payouts, reconciliation.
4. **Discovery:** search index, filters, geo, saved items, follow, personalised feed.
5. **Event day:** signed QR, staff scanner with offline sync, online join links.
6. **Organiser tooling:** team roles, insights, promotion links, recurring events.
7. **Trust and safety and admin console:** verification, moderation, risk scoring, disputes.
8. **Scale and polish:** waiting room, communities, accessibility audit, localisation, SLO tuning.

---

# PART D — GAPS AND PROBLEMS IN THE MAP

The map is broad and well-structured, but these are real holes to close before building.

1. **No timeouts defined.** Approval "No decision", hold duration, waitlist offer expiry, and payment pending expiry have no stated durations or outcomes. Define each.
2. **Approval plus payment ordering is unclear.** Does an approved paid registration reserve capacity at approval time, at payment time, or both? Pick one and handle expiry.
3. **Capacity semantics for approval events.** Pending requests count against capacity or not? Over-approval handling is missing.
4. **Refund policy logic is only a decision diamond.** Needs a rules model: windows, partial refunds, organiser-initiated cancellation, fees (are platform and processor fees refunded?).
5. **Transfer abuse.** Resale and scalping controls are not described; at minimum add transfer limits, price rules, and a fraud check on transfers.
6. **Guest identity linkage.** Guest registrations need a claim path and a rule for duplicate detection by email.
7. **Organiser payout lifecycle is thin.** KYC, payout holds for new organisers, dispute reserves, and failed payouts are not drawn.
8. **Event edit vs registration data.** What happens to custom questions already answered when the organiser changes the form?
9. **Recurring events.** Only "clone safely or configure recurrence" is stated; series-level vs instance-level edits and cancellations are undefined.
10. **Hybrid events.** Capacity per mode (venue capacity vs stream capacity) and mode switching after registration are not specified.
11. **Check-in staff role.** Appears only implicitly; define permissions, device trust and offline data-at-rest protection.
12. **Moderation SLAs and appeal limits.** No timing or appeal count caps; no repeat-offender policy.
13. **Notification compliance.** SMS and marketing consent, quiet hours, unsubscribe handling are not covered beyond preferences.
14. **Data and legal.** Retention periods, data residency, minors' registration, and tax invoicing per region are unspecified.
15. **Sponsored results.** The map requires labelling promoted content but there is no advertiser or promotion flow anywhere. Either remove it or add it.
16. **Analytics and attribution.** "Referral tracking" is listed but there is no privacy/consent treatment of tracking.

**Bottom line:** the UX coverage is strong; the missing pieces are **time limits, money rules, and ownership of edge-case decisions**. Those are what break production systems, so settle them before writing code.

---

## Sources consulted (paraphrased above)
- Ticketing system-design guides: educative.io (Ticketmaster design), techinterview.org (ticketing, flash-sale and booking guides), techprep.app (ticket booking system)
- Harrisburg University paper: *Design and Evaluation of a Distributed Ticket Booking System Using CQRS and Hybrid Consistency Models* (digitalcommons.harrisburgu.edu)
- Luma help centre and API documentation (webhooks, check-in, payments, pricing)
