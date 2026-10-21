# Better Auth development setup

Better Auth is installed only in `apps/core-web` and is opt-in. The current Go
global backend remains the production authentication authority. This prevents a
second, untrusted session from being accepted by the core and organizer APIs.

## Enable locally

1. Apply `apps/core-web/migrations/better-auth-dev.sql` to the local
   development database.
2. Copy `apps/core-web/.env.example` to `.env.local`.
3. Set `BETTER_AUTH_ENABLED=true`, a generated `BETTER_AUTH_SECRET`, and the
   Google client values in `.env.local`.
4. Set Google Cloud's development redirect URI to:
   `http://localhost:3001/api/auth/callback/google`.
5. Start the core web app with `npm run dev --workspace=@hackways/core-web`.

The Better Auth endpoints live at `/api/auth/*`. They are not routed by the
production Nginx configuration and are not used by the existing `/api/v1/auth/*`
Go endpoints.

## Production gate

Do not set `BETTER_AUTH_ENABLED=true` in production until a reviewed exchange
turns a verified Better Auth identity into the existing Go session contract. The
development tables and cookies must never be treated as authorization by either
Go backend.
