# Failure and recovery scenarios

These are design decisions and review cases, not claims that every scenario has passed an executed test.

| Scenario | Required behavior / implemented control | Verification or remaining limit |
| --- | --- | --- |
| Two requests compete for the last place | Both lock the event; committed counter plus holds stays within capacity; loser receives conflict | Existing 100-request/10-place case passed: 10 confirmed, 90 conflicts; no race instrumentation |
| Client loses connection after commit | Same random key/payload replays current registration without a second increment or email | Persisted idempotency record shares the booking transaction; mutation replay not exercised |
| Reused key with edited form | 409 rather than applying a different operation | Canonical digest mismatch verified in the isolated suite |
| Duplicate email with case differences | Normalize at HTTP boundary, case-insensitive lookup under lock and unique database index | Local migration found zero legacy collisions |
| Guest claims a booked email without verification | Deny; identity linkage requires verified session email | Backend gate; actual email verification depends on ACS |
| Guest retry key after real account claims registration | Do not expose account-linked status/notes through guest replay | Account ownership rechecked before response |
| Cancel races registration or approval | Serialize on the event; counters and outbox commit with the winning state | Transaction implementation; concurrency run pending |
| Organizer lowers capacity below reservations | 409; no counter rewriting | Compare confirmed plus held under lock; DB guard |
| Pending approval floods capacity | No seat is allocated until organizer approval | Unique account/event identity and bounded reads; abuse controls still need monitored policies |
| Approval arrives after deadline | Expire at seven days/start/close, then reject approval | Database timestamps and maintenance; no timing load claim |
| Waitlist offer expires while accepted | Event lock decides expiry/accept; no dual allocation | Offer expiry checked under the same lock |
| Several maintenance processes run | Same event lock serializes expiration/promotion | No worker-local inventory authority |
| More than one maintenance batch is needed | Bound per pass and continue later; no unbounded query | 50 events/pass, 100 expiries/event, 50 promotions/event |
| Approval event attempts waitlist bypass | Reject combination; require organizer review | Explicit mode guard |
| Old editor overwrites changed schedule | 409 revision conflict, preserve persisted state | Locked revision check verified through authenticated handler requests |
| Material live event change | Require explicit attendee-impact acknowledgement; enqueue notices atomically | Backend routes implemented; UI work remains |
| Cancelled event link is shared | Previously public cancellation remains readable; registration disabled | Previously published cancellation query verified in the isolated suite |
| An attacker guesses another registration/event ID | 404 on owner-only resources; no UUID-as-credential access | Account/organizer scopes in Go and queries |
| Forged/expired QR or wrong event | Signature, event binding, expiry and DB ticket status checks | Ed25519 payload; scanner UI/offline sync remaining |
| Duplicate scanners race | Only issued ticket can change to scanned once; loser sees duplicate/unavailable | Conditional UPDATE inside locked transaction; concurrent scanner run pending |
| Cancel/transfer after admission | Reject scanned ticket; never restore it via a duplicate registration request | Ticket state checks |
| Transfer is retried | Source remains transferred, count unchanged; no second destination issued | Source transition/recipient ticket are atomic |
| Transferred ticket is scanned | Old ticket revoked and old registration no longer confirmed | Revocation and replacement signature verified in the isolated suite |
| Cross-site credentialed write | Explicit origin allowlist, Fetch Metadata fallback, JSON content type and SameSite cookie | Live hostile-origin rejection can be read-only verified |
| Slow/oversized request or exhausted slots | Bounded header/body/time/concurrency; 413/415/503 with recovery guidance | Resource limits implemented; throughput not measured |
| User-auth DB lookup fails | 503/500 error, never silently authenticate or downgrade to a guest | Propagate storage failures separately from invalid session |
| Provider missing | Queue remains pending without consuming send retries; verification/upload routes unavailable | No dummy signer/provider success in mounted runtime |
| Sender accepted but DB mark failed | Reconcile same provider operation before retry, lease-token guards outcome writes | ACS request identity implemented; actual provider integration not executed |
| Provider operation outcome stays uncertain | Never send a new operation blindly beyond replay window | Operator reconciliation required |
| Worker restarts after claim | Five-minute lease releases row; two entries claimed for two sends | No ten-entry batch waiting longer than its lease |
| Old worker reports success after lease replacement | Conditional outcome write requires matching lease token | Prevents stale worker overwrites |
| Confirmation email queued before cancellation | Suppress when registration no longer matches message state | State read before send; unavoidable send/state-change race is documented |
| Untrusted upload category/event ID | Deny arbitrary paths and other owners; server picks namespace | No upload bytes pass through Go; properties finalization remains |
| Database fails during shutdown | Drain HTTP, cancel workers, wait, then close pool | Shutdown ordering corrected; process signal recovery not exercised |

## Primary sources

- [PostgreSQL 16 row locking](https://www.postgresql.org/docs/16/explicit-locking.html): row locks serialize competing updates; use a consistent event-first lock order.
- [PostgreSQL 16 SELECT / SKIP LOCKED](https://www.postgresql.org/docs/16/sql-select.html): skip-locked rows suit queue consumers, not authoritative capacity availability reads.
- [OWASP REST security](https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html): authorize endpoints/resources, validate requests, and constrain abusive request rates/work.
- [ACS email Send](https://learn.microsoft.com/en-us/rest/api/communication/email/email/send?view=rest-communication-email-2023-03-31): 202 is an asynchronous operation; poll its status before recording send success.
- [ACS authentication](https://learn.microsoft.com/en-us/rest/api/communication/authentication): sign REST requests in a trusted service environment.
- [ACS repeatability](https://learn.microsoft.com/en-us/rest/api/communication/repeatable-requests): repeatability uses request identity and first-sent timestamp with a limited replay window. Application reconciliation must handle outcomes outside it.

No third-party throughput estimates, claimed SLOs, extra search engine, distributed lock, or payment rules from the brief are treated as measured/approved Hackways behavior.
