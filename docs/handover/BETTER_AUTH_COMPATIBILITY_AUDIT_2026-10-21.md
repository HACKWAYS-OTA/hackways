# Better Auth compatibility audit

Date: 21 October 2026

## Current authentication authority

Hackways currently authenticates users in the Go global backend at
`services/global-backend/internal/auth`. It owns registration, password login,
logout, email verification, session creation, session revocation and account
session listing. The core backend independently validates the same opaque
session cookie before authorizing event, RSVP and organizer operations.

The current contract is:

- PostgreSQL `users` and `sessions` tables.
- `pgx` database access with generated `sqlc` queries.
- SHA-256 hashes of opaque session tokens stored in `sessions.token_hash`.
- Host-scoped `__Host-hackways_session` cookies for attendee and organizer
  applications, plus a separate admin cookie.
- Frontend requests to the global backend `/api/v1/auth/*` endpoints with
  credentials included.

Existing verification on 21 October 2026:

```text
go test ./internal/auth -count=1   PASS
go vet ./internal/auth             PASS
```

## Better Auth capabilities relevant to Hackways

Better Auth can use PostgreSQL, but its built-in adapter uses Kysely and its
core schema expects `user`, `session`, `account` and `verification` records.
It supports custom model and field names, Google OAuth, and versioned secret
rotation. These capabilities do not make the existing Go session contract
compatible automatically.

## Compatibility findings

1. The Better Auth session `token` field is a unique session credential, while
   Hackways stores only a SHA-256 digest in `sessions.token_hash`. Reusing the
   current table without a deliberate token contract would either break Go
   validation or store a credential in the wrong form.
2. Better Auth stores email-password credentials in an `account` record. The
   current schema stores a bcrypt password hash directly on `users`.
3. Better Auth's default session cookie on `auth.hackways.in` cannot be read by
   the host-scoped cookies used on `app.hackways.in` and
   `organizer.hackways.in`. A cross-host session must use a reviewed exchange
   protocol or an explicitly reviewed cookie-domain change.
4. Running Better Auth in a Next.js app would add a TypeScript auth server and
   a Kysely/Node database path alongside the Go `pgx`/`sqlc` path. It must not
   be deployed as a disconnected login that creates users the Go backends
   cannot authorize.
5. `auth.hackways.in` currently resolves to the production VM, but no Nginx
   upstream or Better Auth runtime is configured for that host.

## Safe migration shape

The first implementation should remain non-production and use a dedicated
Better Auth schema. It must include:

1. A route handler hosted at `auth.hackways.in` with an explicit
   `BETTER_AUTH_URL`.
2. Google sign-in with the least-privilege `openid`, `email` and `profile`
   identity scopes.
3. A one-time, short-lived exchange from the Better Auth session to the
   existing Go session contract, or a reviewed full session migration. Do not
   expose Better Auth cookies directly to protected Go endpoints.
4. Existing-user matching by verified email with collision and account-linking
   checks. Never create a second Hackways account silently.
5. Numbered, reviewable PostgreSQL migrations and a rollback procedure. No
   generated CLI migration may be applied directly to production.
6. Regression tests for first sign-in, repeat sign-in, account linking,
   logout, revocation, callback replay, invalid state, and an unauthenticated
   request to each protected Go backend.

## Credentials

The protected file
`C:\Users\ADMIN\Documents\hackways-better-auth-secrets.json` was found and
contains the names `BETTER_AUTH_SECRET`, `GOOGLE_CLIENT_ID` and
`GOOGLE_CLIENT_SECRET`. Only presence was checked; values were not printed,
logged or committed.

## Development work completed after the audit

`better-auth@1.7.12` and `pg@8.16.3` are installed only in `apps/core-web`.
The opt-in Next.js route is available at `/api/auth/*` when
`BETTER_AUTH_ENABLED=true`, with a separate development schema in
`apps/core-web/migrations/better-auth-dev.sql`. Production Nginx routing,
Go session validation, and production environment values were not changed.

The route is intentionally not connected to the existing Go session cookie yet;
the reviewed exchange protocol remains the prerequisite for production use.
