'use client';
import { useEffect } from 'react';
import { adminVerificationPath, adminVerificationRequired } from '@/lib/admin-verification';
export function AdminSessionRefresh() {
  useEffect(() => {
    let disposed = false, lastAttempt = 0, nextAttempt = 0, timer: ReturnType<typeof setTimeout>, request: AbortController | null = null;
    async function check() {
      if (disposed || document.hidden || request) return;
      lastAttempt = Date.now(); request = new AbortController();
      const timeout = setTimeout(() => request?.abort(), 7000);
      let wait = 60000;
      try {
        const response = await fetch('/api/v1/auth/admin/session', { credentials: 'include', cache: 'no-store', signal: request.signal });
        const next = window.location.pathname + window.location.search;
        if (response.status === 401) { window.location.replace(`/login?next=${encodeURIComponent(next)}`); return; }
        if (await adminVerificationRequired(response)) { window.location.replace(adminVerificationPath(next)); return; }
        if (response.status === 403) { window.location.replace('/access-denied'); return; }
        if (response.ok) { const seconds = Number(response.headers.get('X-Admin-Verification-Seconds')); wait = Math.min(6 * 60 * 60 * 1000, seconds > 0 ? seconds * 1000 + 250 : 6 * 60 * 60 * 1000); }
      } catch { /* Availability failures preserve the credential and retry later. */ }
      finally { clearTimeout(timeout); request = null; nextAttempt = Date.now() + wait; clearTimeout(timer); if (!disposed && !document.hidden) timer = setTimeout(() => { void check(); }, Math.max(1000, wait)); }
    }
    function resume() { if (document.hidden) { clearTimeout(timer); request?.abort(); } else if (Date.now() - lastAttempt >= 60000 || Date.now() >= nextAttempt) { void check(); } else { clearTimeout(timer); timer = setTimeout(() => { void check(); }, Math.max(1000, nextAttempt - Date.now())); } }
    void check(); document.addEventListener('visibilitychange', resume); window.addEventListener('online', resume);
    return () => { disposed = true; clearTimeout(timer); request?.abort(); document.removeEventListener('visibilitychange', resume); window.removeEventListener('online', resume); };
  }, []);
  return null;
}
