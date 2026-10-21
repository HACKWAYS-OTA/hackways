# Google sign-in

Hackways keeps its existing Go/PostgreSQL session system. Google is an OAuth identity provider only; it does not create a second session system.

## Runtime configuration

Set these values in the protected runtime environment for `global-backend`:

- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URL` (`https://auth.hackways.in/api/v1/auth/google/callback` in production)
- `GOOGLE_ALLOWED_RETURN_ORIGINS` (`https://app.hackways.in,https://organizer.hackways.in` in production)

The client secret must never be committed or printed. If the three Google variables are partially configured, the backend refuses to start rather than silently exposing a broken sign-in flow.

## Flow

1. A sign-in screen sends the browser to `auth.hackways.in/api/v1/auth/google/start` with a return path on the initiating app.
2. The Go backend stores a signed, short-lived state/PKCE cookie and redirects to Google.
3. The callback validates state, exchanges the code, and reads Google’s verified user profile.
4. The provider subject is linked to the existing `users` row in `user_identities` inside PostgreSQL. Existing password accounts are linked only after Google reports a verified email.
5. The callback creates a two-minute, one-time exchange code. The initiating app consumes it and receives the normal host-only `__Host-hackways_session` cookie.

The one-time handoff is required because a host-only cookie set on `auth.hackways.in` cannot be shared with the app or organizer subdomains.

## Google Cloud setup

Register the production callback URI exactly as above and authorize the app origins `https://app.hackways.in` and `https://organizer.hackways.in`. Local development uses the localhost values in the relevant `.env.example` files.

Apply migration `000008_google_oauth` before enabling the runtime variables. Better Auth remains unconfigured; enabling a second session implementation would bypass the existing Go authorization and session checks.
