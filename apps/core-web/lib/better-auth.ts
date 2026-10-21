import { betterAuth, type Auth, type BetterAuthOptions } from 'better-auth';
import { Kysely, PostgresDialect } from 'kysely';
import pg from 'pg';

type BetterAuthDatabase = Kysely<Record<string, unknown>>;

declare global {
  // eslint-disable-next-line no-var
  var __hackwaysBetterAuth: Auth<BetterAuthOptions> | undefined;
  // eslint-disable-next-line no-var
  var __hackwaysBetterAuthDatabase: BetterAuthDatabase | undefined;
}

function required(name: string): string {
  const value = process.env[name]?.trim();
  if (!value) throw new Error(`${name} is required when Better Auth is enabled`);
  return value;
}

/**
 * Development Better Auth instance.
 *
 * This is deliberately opt-in. The existing Go auth service remains the
 * production authority until the reviewed session-exchange migration is
 * complete. Better Auth uses its own tables from migrations/better-auth-dev.sql.
 */
export function getBetterAuth() {
  if (process.env.BETTER_AUTH_ENABLED !== 'true') return null;
  if (globalThis.__hackwaysBetterAuth) return globalThis.__hackwaysBetterAuth;

  const databaseUrl = required('BETTER_AUTH_DATABASE_URL');
  const secret = required('BETTER_AUTH_SECRET');
  const baseURL = required('BETTER_AUTH_URL');
  const googleClientId = required('GOOGLE_CLIENT_ID');
  const googleClientSecret = required('GOOGLE_CLIENT_SECRET');

  const { Pool } = pg;
  const pool = new Pool({
    connectionString: databaseUrl,
    max: 4,
    idleTimeoutMillis: 10_000,
    connectionTimeoutMillis: 5_000,
  });
  const database = new Kysely<Record<string, unknown>>({
    dialect: new PostgresDialect({ pool }),
  });
  globalThis.__hackwaysBetterAuthDatabase = database;

  globalThis.__hackwaysBetterAuth = betterAuth({
    appName: 'Hackways',
    baseURL,
    secret,
    database: { db: database, type: 'postgres' },
    trustedOrigins: [
      'http://localhost:3001',
      'https://auth.hackways.in',
      'https://app.hackways.in',
    ],
    socialProviders: {
      google: {
        clientId: googleClientId,
        clientSecret: googleClientSecret,
        scopes: ['openid', 'email', 'profile'],
      },
    },
  }) as unknown as Auth<BetterAuthOptions>;

  return globalThis.__hackwaysBetterAuth;
}
