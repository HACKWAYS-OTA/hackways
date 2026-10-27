'use client';
import { useEffect } from 'react';

/** Renew cookie transport from Go's existing fixed server expiry. Never stores
 * credentials in JavaScript or turns an unavailable service into a logout. */
export function SessionRefresh({ enabled, apiBase, restrictedUrl = '/not-allowed' }: { enabled: boolean; apiBase: string; restrictedUrl?: string }) {
  useEffect(() => {
    if (!enabled) return;
    let disposed = false, lastAttempt = 0, timer: ReturnType<typeof setTimeout>;
    let request: AbortController | null = null;
    function schedule() { clearTimeout(timer); if (!disposed && !document.hidden) timer = setTimeout(() => { void refresh(); }, Math.max(1000, 6 * 60 * 60 * 1000 - (Date.now() - lastAttempt))); }
    async function refresh() {
      if (disposed || document.hidden || request) return;
      lastAttempt = Date.now(); request = new AbortController();
      const timeout = setTimeout(() => request?.abort(), 5000);
      try {
        const response = await fetch(`${apiBase.replace(/\/$/, '')}/auth/session`, { credentials: 'include', cache: 'no-store', signal: request.signal });
        if (!disposed && response.status === 403) {
          const body: unknown = await response.json().catch(() => null);
          if (body && typeof body === 'object' && 'error' in body && body.error && typeof body.error === 'object' && 'code' in body.error && body.error.code === 'ACCOUNT_RESTRICTED') window.location.assign(restrictedUrl);
        }
      }
      catch { /* Network recovery retries on visibility/online with bounded cadence. */ }
      finally { clearTimeout(timeout); request = null; schedule(); }
    }
    function resume() { if (document.hidden) { clearTimeout(timer); request?.abort(); } else if (Date.now() - lastAttempt >= 10 * 60 * 1000) { void refresh(); } else schedule(); }
    void refresh(); document.addEventListener('visibilitychange', resume); window.addEventListener('online', resume);
    return () => { disposed = true; clearTimeout(timer); request?.abort(); document.removeEventListener('visibilitychange', resume); window.removeEventListener('online', resume); };
  }, [enabled, apiBase, restrictedUrl]);
  return null;
}
