export type AdminAccess = 'allowed' | 'signed-out' | 'denied' | 'unavailable';

type SessionCookie = { name: string; value: string };

// Usable by both the request gate and the server-render guard. Authorization is
// performed by Go; the browser is never trusted to assert roles or session state.
export async function verifyAdminAccess(sessionCookies: SessionCookie[]): Promise<AdminAccess> {
  if (sessionCookies.length === 0) return 'signed-out';
  const base = (process.env.GLOBAL_API_URL || (process.env.NODE_ENV === 'development'
    ? 'http://127.0.0.1:8082/api/v1' : 'http://global-backend:8082/api/v1')).replace(/\/$/, '');
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 4000);
  try {
    const response = await fetch(`${base}/auth/admin/me`, {
      headers: { Cookie: sessionCookies.map(cookie => `${cookie.name}=${cookie.value}`).join('; '), Accept: 'application/json' },
      cache: 'no-store', redirect: 'error', signal: controller.signal,
    });
    if (response.status === 401) return 'signed-out';
    if (response.status === 403) return 'denied';
    if (response.status !== 200) return 'unavailable';
    const body: unknown = await response.json();
    return body !== null && typeof body === 'object' && 'display_name' in body && typeof body.display_name === 'string'
      ? 'allowed' : 'unavailable';
  } catch {
    return 'unavailable';
  } finally {
    clearTimeout(timeout);
  }
}
