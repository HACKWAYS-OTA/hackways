'use client';

import { useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
import { Checkbox } from '@hackways/ui/checkbox';
import { Input } from '@/components/ui/Input';
import { AuthLayout, AuthGoogleButton, AuthDivider, AuthPasswordInput, AuthSubmit } from '@hackways/ui/auth';

export function AuthForm({ next }: { next: string }) {
  const router = useRouter();
  const [register, setRegister] = useState(false);
  const [promotions, setPromotions] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setBusy(true); setError('');
    const form = new FormData(e.currentTarget);
    try {
      const result=await fetchApi<{user:{email_verified:boolean};email_delivery_available?:boolean}>(register ? '/auth/register' : '/auth/login', { method: 'POST', body: JSON.stringify({ email: String(form.get('email')).trim(), password: form.get('password'), ...(register ? { display_name: String(form.get('name')).trim(), promotional_email: promotions } : {}) }) });
      if(!result.user.email_verified) router.replace(`/auth/verify-email?next=${encodeURIComponent(next)}${register ? '&queued=1' : ''}${register && result.email_delivery_available===false ? '&delivery=unavailable' : ''}`); else router.replace(next); router.refresh();
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
  return <AuthLayout title={register ? 'Create your account' : 'Sign in to Hackways'} homeHref="/discover">
    {!register && <><AuthGoogleButton disabled={busy} onClick={continueWithGoogle} /><AuthDivider /></>}
    <form className="hw-auth-form" onSubmit={submit} aria-busy={busy}>
      {register && <Input label="Your name" name="name" autoComplete="name" maxLength={100} disabled={busy} required />}
      <Input label="Email address" name="email" type="email" autoComplete={register ? 'email' : 'username'} maxLength={254} disabled={busy} required />
      <AuthPasswordInput name="password" autoComplete={register ? 'new-password' : 'current-password'} minLength={register ? 8 : undefined} maxLength={128} disabled={busy} required helperText={register ? 'Use at least 8 characters.' : undefined} />
      {!register && <a className="hw-auth-forgot" href={`/auth/forgot-password?next=${encodeURIComponent(next)}`}>Forgot password?</a>}
      {register && <label className="hw-auth-preference"><Checkbox checked={promotions} onCheckedChange={setPromotions} disabled={busy} /><span>Send me Hackways news and offers. You can unsubscribe anytime.</span></label>}
      {error && <p className="hw-auth-error" role="alert">{error}</p>}
      <AuthSubmit disabled={busy}>{busy ? 'Please wait...' : register ? 'Create account' : 'Sign in'}</AuthSubmit>
    </form>
    <p className="hw-auth-switch">{register ? 'Already have an account?' : 'New to Hackways?'} <button type="button" disabled={busy} onClick={() => { setRegister(!register); setError(''); }}>{register ? 'Sign in' : 'Create an account'}</button></p>
  </AuthLayout>;
}
