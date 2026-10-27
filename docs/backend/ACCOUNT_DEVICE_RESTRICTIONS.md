# Account and browser restrictions

Global Go owns moderation decisions and authentication issuance; both Go engines
read PostgreSQL on protected requests. The console is not an authorization
boundary. Migration 30 is additive and changes no existing account or session
permission by default. Four Next applications and two Go engines remain separate.

## Identity and persistence

The existing Secure/HttpOnly host-scoped 180-day random browser cookie is hashed
with SHA256. It is a browser-profile identifier, not hardware identity. Clearing
cookies or using another profile/device can evade recognition. IP/user-agent
matches never automatically suspend unrelated accounts. Recognition is not an
authentication factor.

`device_restrictions` retains the hash, internal identity, numeric console
reference, reason, revision and audit actor independently of recognition cleanup.
`account_restrictions` records independent review decisions and their source.
`restriction_device_accounts` retains associations discovered during review.
New sessions bind the recognized hash and have nullable `revoked_at`. Legacy
sessions retain compatible defaults; unavailable historical associations cannot
be inferred. API responses never expose hashes or raw recognition cookies.

## Moderation operations

These are dedicated-admin routes under `/api/v1/admin/global`, protected by the
existing database grant, daily verification, Origin, reason, version and stable
idempotency receipt. Reasons are audited, and missing/current record versions
cannot silently overwrite newer decisions.

- `GET /restrictions/accounts?q=&limit=20&offset=0`
- `GET /restrictions/devices?q=&limit=20&offset=0`
- `GET /people/{reference}?section=devices` lists recognized browser references.
- `POST /people/{reference}/restriction` accepts `updated_at, blocked`.
- `POST /people/{reference}/devices/block` accepts `updated_at, device_reference`.
- `POST /devices/{reference}/restriction` accepts `updated_at, blocked:false`.

Suspension retires all sessions and MCP grants. Blocking a browser also suspends
associated accounts discovered from recognition, active session bindings and
prior review associations. Rows are locked in UUID order. The operation rejects
more than 100 associated accounts before changing anything; those accounts need
individual review to keep the transaction bounded. Existing manual reasons are
retained. Active administrator accounts, the acting administrator and browsers
associated with active administrators cannot be suspended through this path.
Existing grant-revocation and last-admin protections must be used first.

Account unblock changes only the account. Browser unblock changes only the
browser. Neither revives sessions, refresh/access grants or cancelled bookings.
A reviewed person signs in again. Existing RSVPs, tickets, media and ownership
remain persisted; restrictions do not silently cancel events or registrations.

## Sign-in and account switching

Password proof is checked before quarantine, and the password version is rechecked
under the account lock. A wrong password cannot suspend the submitted email.
Google proof retains known source/callback hashes through the validated OAuth
state/exchange. Signup/linking checks every known hash before creating records.
Existing verified provider proof can still reach the quarantine transaction.

Sign-in locks sorted unique known browser hashes before the account row. A valid
sign-in from a blocked browser records the new account association, suspends the
account, retires existing sessions/grants, consumes exchanges and commits before
returning403 `ACCOUNT_RESTRICTED` / `You are not allowed`. Logout and password
recovery retain recognition cookies. Existing-account-only restriction denial
also consumes its proof and retires old access before returning.

The administrator lock order is role advisory, action receipt advisory, browser
advisory, sorted user rows, restriction rows and access retirement. No auth path
acquires a browser lock after holding a user row. Device advisory locks serialize
first block insertion with sign-in even before a restriction row exists.

## Enforcement and user experience

Both engines reject restricted account/session bindings and retired sessions.
MCP identity, code/refresh exchange and approved execution queries check account,
browser and revocation state. Public mutation guards also cover guest actions;
registration checks blocked account email after logout. Credential-proof
endpoints reach their quarantine logic, while logout/recovery preserve their
existing purposes. Approved admin routes remain protected independently.

Restrictions apply to authorization checks for requests; they do not undo
completed operations. Direct Azure upload architecture is unchanged: new SAS
issuance and final attachment require allowed access. No media passes through Go.
Existing transactional email/outbox and RSVP capacity rules remain unchanged.

Private SSR requests forward the session and browser cookies. Layouts treat a
restricted identity as signed out without redirecting; protected page helpers
redirect to the canonical public `/not-allowed` page. This avoids redirect loops.
Interactive clients and background cookie refresh use that same destination.
The new Nginx core route is required. Normal service errors remain real failures,
not invented restriction decisions. Restriction lists use bounded pagination and
the established global loader, with no visible freshness timestamps.

## Release and recovery

Run mandatory isolated PostgreSQL 16/race integration before activation. Apply
migration 30, start core Go before global Go, then the three changed frontends and
the scoped Nginx route. Verify readiness and retain all existing environment and
untargeted services. Do not create production blocks/accounts to verify behavior.

The down migration retains restriction records and retired credentials. Once
moderation decisions exist, rolling Go back to code that ignores restrictions
would bypass them: retain enforcing Go images and recover forward. Initial image
rollback is safe only while restriction records remain empty. Never delete
moderation decisions or unretire credentials to make rollback appear successful.
