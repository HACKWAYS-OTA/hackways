import 'server-only';
import { cookies } from 'next/headers';
export async function mcpData<T>(path: string, _destination: string): Promise<{ data?: T; error?: string }> {
  const base = (process.env.GLOBAL_API_URL || 'http://global-backend:8082/api/v1').replace(/\/$/, '');
  const session = cookies().get('__Host-hackways_admin_session');
  const controller = new AbortController(); const timer = setTimeout(() => controller.abort(), 8000);
  try {
    const response = await fetch(`${base}/auth/mcp${path}`, { headers: { Accept: 'application/json', ...(session ? { Cookie: `${session.name}=${session.value}` } : {}) }, cache: 'no-store', redirect: 'error', signal: controller.signal });
    if (!response.ok) return { error: response.status === 410 ? 'This request expired. Start again from your app.' : 'The request is unavailable. Check your admin sign-in and daily verification, then retry.' };
    return { data: await response.json() as T };
  } catch { return { error: 'The connection service is unavailable. Please try again shortly.' }; }
  finally { clearTimeout(timer); }
}
