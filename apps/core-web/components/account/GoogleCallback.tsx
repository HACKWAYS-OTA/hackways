'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
import { safeNext } from '@/lib/journeys';

export function GoogleCallback({ code, error, next }: { code?: string; error?: string; next?: string }) {
  const router = useRouter();
  const [message, setMessage] = useState(error || 'Completing Google sign-in…');
  useEffect(() => {
    if (!code || error) return;
    let cancelled = false;
    fetchApi('/auth/google/exchange', { method: 'POST', body: JSON.stringify({ code }) })
      .then(() => { if (!cancelled) { router.replace(safeNext(next, '/account')); router.refresh(); } })
      .catch((err: unknown) => { if (!cancelled) setMessage(err instanceof Error ? err.message : 'Google sign-in failed. Please try again.'); });
    return () => { cancelled = true; };
  }, [code, error, next, router]);
  return <section className="auth-page"><h1>{error ? 'Sign-in was not completed.' : 'Signing you in…'}</h1><p role={error ? 'alert' : undefined}>{message}</p>{error && <a className="quiet-link" href={`/login?next=${encodeURIComponent(safeNext(next, '/account'))}`}>Return to sign in</a>}</section>;
}
