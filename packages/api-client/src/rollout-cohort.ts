/** Obtain only the opaque presentation cohort cookie from the Go runtime API.
 * This transport helper does not evaluate flags or authorize application access. */
export async function requestRolloutCookie(base: string): Promise<string | null> {
  const controller = new AbortController(), timer = setTimeout(() => controller.abort(), 1000);
  try {
    const response = await fetch(`${base.replace(/\/$/, '')}/auth/runtime`, { cache: 'no-store', signal: controller.signal });
    if (!response.ok) return null;
    const cookie = response.headers.get('set-cookie');
    return cookie && /^__Host-hackways_rollout=[0-9a-f]{64};/i.test(cookie) ? cookie : null;
  } catch { return null; } finally { clearTimeout(timer); }
}
