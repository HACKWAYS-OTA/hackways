# Persistent sessions and daily admin verification

## Ordinary sessions

Hackways retains custom Go authentication with opaque random session credentials, PostgreSQL authority and HttpOnly cookies. This change does not introduce Better Auth, JWT sessions or browser-stored credentials.

Password signup, password sign-in and Google sign-in create a fixed 90-day session. Signing in on another device does not revoke earlier devices. Explicit sign-out, administrator revocation and password recovery still invalidate sessions. A database or network failure is reported as unavailability and does not delete a valid credential.

Migration 18 extends still-active ordinary sessions to at least 90 days from their original creation time. It does not revive expired or deleted sessions. `GET /api/v1/auth/session` validates the existing credential and refreshes its cookie with the remaining server lifetime; it never creates a session or extends the fixed server expiry. Core and management request this on mount and at bounded six-hour/visibility intervals. Previously issued seven-day browser cookies can only be renewed when that browser returns before its old cookie expires. An offline browser cookie cannot be extended remotely.

Remembered-device cookies remain host-only, Secure and HttpOnly with a 180-day recognition lifetime. They suppress duplicate new-browser notices, not authentication checks. Every protected request still checks its live PostgreSQL session and permissions. Only non-authoritative activity writes are asynchronous: one worker per engine, a 256-entry queue, batches of at most 64, one-second flush/deadline and a five-minute heartbeat threshold. Dropped activity samples cannot grant access, revive a session or sign a user out.

## Dedicated admin sessions

Migration 19 extends active dedicated admin sessions to 90 days and adds per-session verification timestamps and expiring challenges. Existing sessions receive no implicit verification and require a code before admin access. Password authentication creates an unverified session; it cannot read or mutate admin data until verified. Both Go engines enforce the same live grant and 24-hour verification window, including a recheck inside administrative mutations.

| Endpoint | Access and result |
|---|---|
| `POST /api/v1/auth/admin/verification/request` | Dedicated session, live admin grant, approved Origin; queues one six-digit code with a 60-second resend cooldown |
| `POST /api/v1/auth/admin/verification/confirm` | Same identity checks; `{code}` verifies an unconsumed challenge within its ten-minute expiry |
| `GET /api/v1/auth/admin/session` | Fully verified admin only; refreshes cookie transport and returns `X-Admin-Verification-Seconds` |

Challenges are hashed using the existing bcrypt code workflow. The restricted transactional email outbox holds the code for delivery. Account, session and challenge locks serialize resend/confirmation. A correct code enables only that session for 24 hours. An incorrect or expired submitted code deletes only that admin session and clears its cookie; other devices remain signed in. Malformed requests and delivery/database outages return explicit errors, without revoking otherwise valid sessions. A consumed successful submission can safely replay while the session remains verified.

Admin APIs return `403 ADMIN_VERIFICATION_REQUIRED` when the daily window expires. Middleware, server layouts and client actions direct the browser to `/verify` before showing protected content. The mounted timer uses the server's remaining window and rechecks on visibility/network recovery. Security decisions remain authoritative in Go even if a browser is suspended or its clock is inaccurate.

The `admin_daily_otp` template uses the account sender `noreply@mail.hackways.in`. A narrow `EMAIL_OUTBOX_TEMPLATE_FILTER=admin_daily_otp` runtime setting is available for local review; filtering happens in the generated claim query before leasing, so other templates retain status and attempt counts. Omit the filter for normal production processing. Invalid template names fail startup.

## Deployment and verification

Apply additive migrations 18-19 before the compatible application release. Activate both backend enforcement implementations before allowing admin traffic to the new UI. Do not roll back to a backend that omits daily verification while continuing to allow admin traffic. The down migrations intentionally retain the additive verification schema and issued expiry dates to preserve live state.

Local PostgreSQL integration tests cover 90-day cookies, active-only migration, multiple devices, cookie renewal, revocation, bounded asynchronous activity, daily expiry, wrong/expired-code revocation, resend cooldown and delivery unavailability. Release Candidate run 39795523731 passed the mandatory Linux PostgreSQL 16/race checks: 62 core and 122 global results, zero failures/skips, plus all six image builds. A real authorized delivery check completed provider acceptance and OTP-confirmed admin access with a 24-hour deadline; no unrelated templates were sent. The earlier isolated runtime inherited `notify` as its account sender and the user reported that incorrect identity. The fallback and explicit review configuration were corrected to `noreply`, with a regression test covering absent transactional sender configuration. The user subsequently authorized deployment. All six production services now run source `dab8d1cd545b843bf20c2d8cdac16bb7feb058d8` with clean schema 19; both backend readiness checks and anonymous admin gates passed. Active sessions were extended without a blanket session revocation. Only the separate copied-data local preview retains the narrow email filter.
