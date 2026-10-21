'use client';

import { useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
import { Input } from '@/components/ui/Input';
import { site } from '@/lib/site';
import './auth-reading.css';

export function AuthForm({ next, management = false }: { next: string; management?: boolean }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setBusy(true); setError('');
    const form = new FormData(e.currentTarget);
    try {
      await fetchApi('/auth/login', { method: 'POST', body: JSON.stringify({ email: String(form.get('email')).trim(), password: form.get('password') }) });
      router.replace(next); router.refresh();
    } catch (err) { setError(err instanceof Error ? err.message : 'Sign in failed. Please try again.'); }
    finally { setBusy(false); }
  }
  function continueWithGoogle() {
    const callback = new URL('/auth/google/callback', window.location.origin);
    callback.searchParams.set('next', next);
    const start = new URL('/api/v1/auth/google/start', window.location.origin);
    start.searchParams.set('return_to', callback.toString());
    window.location.assign(start.toString());
  }
  return <section className="auth-page"><h1>{management ? 'Organizer sign in' : 'Sign in'}</h1><p>{management ? 'Use your Hackways account to manage existing events, review registrations and check in guests.' : 'Sign in to your account.'}</p><form className="journey-form" onSubmit={submit} aria-busy={busy}>
    <Input label="Email address" name="email" type="email" autoComplete="email" maxLength={254} required />
    <Input label="Password" name="password" type="password" autoComplete="current-password" maxLength={128} required />
    {error && <p className="journey-error" role="alert">{error}</p>}
    <button className="app-button app-button-primary" disabled={busy}>{busy ? 'Please wait...' : 'Sign in'}</button>
  </form><div className="auth-divider" aria-hidden="true"><span>or</span></div><button type="button" className="app-button app-button-secondary google-button" disabled={busy} onClick={continueWithGoogle}>Continue with Google</button><p className="auth-switch">New to Hackways? <a className="quiet-link" href={`${site.app}/login?next=%2Fcreate`}>Create an account</a></p></section>;
}
