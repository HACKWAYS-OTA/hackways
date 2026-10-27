# Admin live metrics, Bin and restrictions

Requested 2026-10-27. Live analytics was deployed as 4ddadc1 after hosted PostgreSQL 16/race checks. Bin is deployed as7a6cf1b after hosted run40500732330; restrictions are implemented as a separate candidate and remain undeployed. Keep them separate from the managed-workspace release.

## Findings from existing code

- Original inspection before the implemented live extension: both analytics handlers read real PostgreSQL in read-only repeatable-read
  transactions with bounded timeouts and no-store responses. The admin page is a
  server-rendered historical date-range report without automatic refresh.
- Existing browser analytics is consent based and bounded. Session activity is
  batched at five-minute granularity. Neither is an exact online-person counter.
- Auth already sets a 180-day HttpOnly/Secure host-scoped recognition cookie,
  hashes it and stores per-account recognition records. It currently deduplicates
  sign-in alerts; it grants no permissions and is not a device blacklist.
- Account sessions currently check expiry and deleted user state. Existing admin
  access changes concern platform administrator grants, not account suspension.
- Event soft-deletion exists, but no administrator Bin/restore workflow exists.

## Planned architecture

Keep Go/PostgreSQL authoritative and the six services intact. Global Go owns
account/device restriction state and session issuance; both engines enforce that
state on every protected operation. Browser visibility is never the boundary.

1. Live analytics is implemented: database-clock sliding 5 min, 1 hr and 24 hr windows, bounded visible-tab refresh, independent error/retry handling and separate history. No update timestamps are rendered. PostgreSQL snapshots, actual registrations/check-ins/accounts/provider acceptance and consented visits remain authoritative; no inferred exact online-person counter. See docs/backend/ANALYTICS.md and the production worklog.
2. Add recoverable event deletion and an admin Bin. Preserve registrations,
   tickets, audit and media. Published events require the existing cancellation/
   attendee handling before deletion. Restoration must not silently republish or
   reinstate cancelled registrations. No permanent purge is included.
3. Extend the existing random device-cookie recognition system with a global
   device block record and account review restrictions. Keep account and device
   unblock actions distinct. Persist reasons, versions and audit actors.
4. Logging out retains the device cookie. Signing into another verified account
   with a blocked identifier records the association, restricts that account for
   review and denies session/action issuance. Existing sessions and guest actions
   must also check the block; OAuth/MCP paths must not bypass account restrictions.
5. Preserve current accounts/sessions during migration with compatible defaults.
   Protect administrator recovery/last-admin invariants. Check transaction lock
   ordering before combining device, account and session writes. Restrictions and
   revocations must commit before returning a denial; never roll back the very
   quarantine being reported as enforced.

## Limits and verification

The identifier identifies a browser profile, not physical hardware. Cookie
clearing, a different browser/profile or a new device can evade identification;
IP/user-agent matches alone must not automatically suspend unrelated people.
No hidden hardware fingerprint, MAC-address claim or guaranteed evasion prevention.

Extend mandatory isolated PostgreSQL/race cases for blocked-account/device
actions, logout/account switching, OAuth/MCP, revocation/review/unblock, deletion/
restore and live-metric truth/freshness. No production accounts or devices are
blocked for verification. Preserve live applications, auth recovery, capacity
invariants and the email outbox throughout staged rollout.

## Restriction implementation decisions (after Bin candidate7a6cf1b)

- Add schema30: standalone device_restrictions(UUID, numeric reference, unique SHA256token_hash, user_agent, blocked, revision, reason, actor/time); account_restrictions(user PK,blocked,revision,reason,source_device optional FK, actor/time); restriction_device_accounts(device,user PK) for permanent review associations; recognized_devices numeric admin_reference identity; sessions.device_token_hash(nullable),revoked_at(nullable). No prior accounts/session permissions change by default.
- Keep actual random recognition cookie. No hardware/IP/fingerprint claims. Password reset must retain device cookies. Admin account protection: reject suspension of self or any active platform admin; revoke grant first with existing last-admin controls. Reject blocking a browser associated with active administrator to preserve recovery.
- Strict lock order: existing global admin-role advisory -> action receipt advisory -> device-hash advisory (if any) -> sorted user rows -> device/account restriction rows -> sessions/MCP grants. Sign-in must take device-hash advisory BEFORE user lock; never acquire admin-role lock while holding a user lock. This serializes first block insertion against login without gap locks or inverted device/user row locks.
- Admin device block takes recognized-device reference+target account version; resolve hash only server-side, take device advisory, lock associated users in UUID order (union recognized and prior review associations, limit101; reject over100 before mutation). Restrict all associated accounts, persist device review associations, retire their sessions/MCP grants and touch user versions using bounded array SQL statements. Independent device unblock changes only device record; account unblock changes only account record. Reblocking should retain an existing manual restriction reason/source.
- Valid password/OAuth proof from blocked cookie: lock device then user, validate current credentials/exchange, persist account review restriction + device association, retire account sessions/MCP grants and COMMIT before returning403. Wrong password must never restrict the supplied email. Sign-up from known blocked cookie is rejected before creating records. No blanket guard before password/OAuth exchange because it would skip account2 quarantine.
- Bind newly issued sessions to recognized hash in same TX. Session reads expose blocked/revoked booleans internally: blocked gives403, revoked denies; account/device unblock never revives retired tokens. Ordinary session APIs/admin queries/MCP identity, refresh, grant and execution checks must honor revoked/account/device state. Revoking MCP grants is authoritative even after unblock.
- Core/global Go read-only device guard covers all non-admin mutations including guest registration, applications, uploads, media attachment. Protected session middleware also checks device on reads. Allow logout and credential-proof exchange; retain recovery cookies. No duplicate TS business logic. MCP uses session-bound hash since bearer clients lack browser cookie.
- Extend existing global admin action framework with person.restriction {updated_at,blocked}, person.device {updated_at,device_reference}, device.restriction {updated_at,blocked:false}; canonical receipts, Origin/reason/role lock and audit retained. Device resolve/list SQL never exposes token hashes. Person detail has restriction state; related Browsers table offers Block browser. New Restrictions module offers paginated blocked accounts/browsers and independent review/unblock actions.
- Frontend forbidden response should show exact You are not allowed with code403 on protected navigation/interactive calls. Inspect layout optional-user fetch before adding redirects to avoid forbidden-page redirect loops; do not rewrite auth architecture. User-facing public auth forms may display the backend message directly.
- Mandatory isolated PG cases: existing session/action denial in both engines, correct-password account switch quarantines+commits, wrong-password no quarantine, logout cookie survival, OAuth-proof denial, MCP/session retirement survives unblock, guest routes denied, independent unblocks, admin recovery, safe admin replay/versions. Never block production accounts/devices to verify.

## Final implementation reference

The implemented contract supersedes earlier planning details: [account/browser restrictions](../backend/ACCOUNT_DEVICE_RESTRICTIONS.md). It includes session-bound associations after recognition pruning, all known OAuth source/callback hashes and guest-email enforcement. Bin is live; restriction candidate verification/activation is recorded separately in WORKLOGS.
