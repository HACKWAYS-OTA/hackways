'use client';

import { useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
import { Input } from '@/components/ui/Input';
import { site } from '@/lib/site';

export function AuthForm({ next, management = false }: { next: string; management?: boolean }) {
  const router = useRouter();
  const [register, setRegister] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setBusy(true); setError('');
    const form = new FormData(e.currentTarget);
    try {
      await fetchApi(register ? '/auth/register' : '/auth/login', { method: 'POST', body: JSON.stringify({ email: String(form.get('email')).trim(), password: form.get('password'), ...(register ? { display_name: String(form.get('name')).trim() } : {}) }) });
      router.replace(next); router.refresh();
    } catch (err) { setError(err instanceof Error ? err.message : 'Sign in failed. Please try again.'); }
    finally { setBusy(false); }
  }
  function continueWithGoogle() {
    const callback = new URL('/auth/google/callback', window.location.origin);
    callback.searchParams.set('next', next);
    const start = new URL('/api/v1/auth/google/start', site.auth);
    start.searchParams.set('return_to', callback.toString());
    window.location.assign(start.toString());
  }
  return <section className="auth-page"><h1>{register ? 'Make yourself at home.' : 'Welcome back.'}</h1><p>{management ? 'Sign in to manage your events and welcome your guests.' : 'Keep your plans together. Join an event or bring people together yourself.'}</p><form className="journey-form" onSubmit={submit} aria-busy={busy}>
    {register && <Input label="Your name" name="name" autoComplete="name" maxLength={100} required />}
    <Input label="Email address" name="email" type="email" autoComplete="email" maxLength={254} required />
    <Input label="Password" name="password" type="password" autoComplete={register ? 'new-password' : 'current-password'} minLength={register ? 8 : undefined} maxLength={128} required helperText={register ? 'Use at least 8 characters.' : undefined} />
    {error && <p className="journey-error" role="alert">{error}</p>}
    <button className="app-button app-button-primary" disabled={busy}>{busy ? 'Please wait...' : register ? 'Create account' : 'Sign in'}</button>
  </form>{!register && <><div className="auth-divider" aria-hidden="true"><span>or</span></div><button type="button" className="app-button app-button-secondary google-button" disabled={busy} onClick={continueWithGoogle}>Continue with Google</button></>}<p className="auth-switch">{register ? 'Already have an account?' : 'New to Hackways?'} <button type="button" className="quiet-link" disabled={busy} onClick={() => { setRegister(!register); setError(''); }}>{register ? 'Sign in' : 'Create an account'}</button></p></section>;
}
