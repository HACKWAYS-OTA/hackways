'use client';

import { useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
import { Input } from '@/components/ui/Input';
import { AuthLayout, AuthGoogleButton, AuthDivider, AuthPasswordInput, AuthSubmit } from '@hackways/ui/auth';
import { site } from '@/lib/site';

export function AuthForm({ next }: { next: string }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setBusy(true); setError('');
    const form = new FormData(e.currentTarget);
    try {
      const result=await fetchApi<{user:{email_verified:boolean}}>('/auth/login', { method: 'POST', body: JSON.stringify({ email: String(form.get('email')).trim(), password: form.get('password') }) });
      router.replace(result.user.email_verified ? next : `/auth/verify-email?next=${encodeURIComponent(next)}`); router.refresh();
    } catch (err) { setError(err instanceof Error ? err.message : 'Sign in failed. Please try again.'); }
    finally { setBusy(false); }
  }
  function continueWithGoogle() {
    const callback = new URL(process.env.NODE_ENV === 'production' ? '/manage/auth/google/callback' : '/auth/google/callback', window.location.origin);
    callback.searchParams.set('next', next);
    const start = new URL('/api/v1/auth/google/start', window.location.origin);
    start.searchParams.set('return_to', callback.toString());
    window.location.assign(start.toString());
  }
  return <AuthLayout title="Sign in to Hackways" homeHref={`${site.app}/discover`}>
    <AuthGoogleButton disabled={busy} onClick={continueWithGoogle} /><AuthDivider />
    <form className="hw-auth-form" onSubmit={submit} aria-busy={busy}>
      <Input label="Email address" name="email" type="email" autoComplete="username" maxLength={254} disabled={busy} required />
      <AuthPasswordInput name="password" autoComplete="current-password" maxLength={128} disabled={busy} required />
      <a className="hw-auth-forgot" href={`/manage/auth/forgot-password?next=${encodeURIComponent(next)}`}>Forgot password?</a>
      {error && <p className="hw-auth-error" role="alert">{error}</p>}
      <AuthSubmit disabled={busy}>{busy ? 'Please wait...' : 'Sign in'}</AuthSubmit>
    </form>
    <p className="hw-auth-switch">New to Hackways? <a href={`${site.app}/login?next=%2Fcreate`}>Create an account</a></p>
  </AuthLayout>;
}
