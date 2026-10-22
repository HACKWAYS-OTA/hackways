# Account verification, recovery and sign-in alerts

## User journeys

- Signup creates an unverified account, session and six-digit email verification challenge atomically. The app opens `/auth/verify-email` and shows the centered code input; email displays a large centered code. Unverified password sign-ins also return to verification. Management provides the equivalent `/manage/auth/verify-email` route.
- Forgot password opens `/auth/forgot-password` (or the management equivalent). It requests a secure reset **link**, never a reset OTP. Known and unknown addresses receive the same generic HTTP 202 body. Requests are limited per identity and deduplicated for 60 seconds.
- Reset emails use the fixed trusted origin `https://hackways.in/auth/reset-password?key=...`. The 32-byte random secret expires after 15 minutes. Only its SHA-256 digest is stored in the recovery challenge. The email outbox necessarily holds the delivery link; it is privileged data.
- The reset page captures the link secret in component memory and removes it from the address bar. It uses `no-referrer` and no-store headers. Reloading after capture requires reopening the email link. Nginx strips the query before forwarding to Next and redacts recovery queries/referrers from access logs. Production app logs do not log request bodies. Do not paste live reset URLs into diagnostics.
- Submitting matching new passwords consumes the challenge under the account lock, replaces the password and revokes all regular/admin sessions, pending OAuth exchanges and remembered devices in one transaction. It also invalidates outstanding email verification challenges and queues a password-change notice. It never signs the user in automatically.
- Signup OTPs expire after 15 minutes, use bcrypt hashes, allow five incorrect attempts and have a 60-second resend cooldown. Previously issued long verification codes remain compatible until expiry. Signup reports delivery unavailability honestly; resend/recovery return 503 when ACS is not configured.

## Device recognition and approximate location

Host-only Secure/HttpOnly recognition cookies remember a browser for 180 days. Admin cookies use SameSite Strict; ordinary cookies use Lax. Recognition never grants authentication or authorization. PostgreSQL stores only the cookie digest, with at most 50 active entries per account and bounded expiry cleanup. A verified account receives one alert for a newly recognized browser; subsequent ordinary sign-ins with that cookie do not repeat it. First signup/first migration-era sign-in establishes a baseline. Clearing cookies or private browsing is a new browser, even on the same physical computer.

Alerts include a coarse device category (for example Windows PC), browser/version, time and approximate city/region/country when supplied by trusted ingress. Browser user agents cannot disclose a computer's actual hostname. Unknown device/browser/location is stated honestly.

Location has two trust boundaries: Nginx accepts Cloudflare geography only when the original socket peer is in its configured Cloudflare ranges, and overwrites all X-Hackways location headers. Go only reads those headers from explicitly configured `AUTH_LOCATION_PROXY_CIDRS`; empty configuration rejects them. Set that variable to the exact Nginx internal network peer IP/CIDR at rollout, and keep it synchronized when the container address changes. Never whitelist public clients or `0.0.0.0/0`.

For city/region, enable Cloudflare **hackways.in / Rules / Settings / Managed Transforms / Add visitor location headers**. Country alone may be available through IP Geolocation. This Cloudflare switch has not been changed or verified in this session; missing metadata displays `Unavailable`. No external geolocation lookup, GPS permission, location fabrication or physical-device fingerprinting is used. See [Cloudflare managed transform headers](https://developers.cloudflare.com/rules/transform/managed-transforms/reference/).

## Operations and verification

Migration 14 is additive. Its down migration refuses to discard live recovery/device records; roll back runtime while retaining schema when populated. Six independent services, pgx/sqlc, bounded ACS outbox workers and India processing are preserved. Package-level Go tests run serially in CI because existing notification integration tests claim the shared isolated test outbox; tests still exercise concurrent requests internally and retain the race detector.

The ordinary local review uses a separate production-data copy. Only daily admin verification emails are enabled; all other copied-data templates remain paused and unleased. Real email delivery checks use a fresh isolated database with only the expressly authorized recipient. ACS operation success proves provider acceptance, not mailbox delivery. Account/security mail defaults explicitly to `noreply@mail.hackways.in`, independently of the event sender `notify@mail.hackways.in`. Promotional mail requires the configured `updates@mail.hackways.in` identity. Production maintenance was removed at the user's subsequent request; the previously released application is online. New auth and design changes remain in local review. See [session persistence](session-persistence.md) for migrations 18-19 and daily admin verification.

Security references: [OWASP recovery guidance](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html) and [session guidance](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html).
