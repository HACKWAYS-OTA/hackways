'use client';
import { useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
import { Input } from '@/components/ui/Input';
export function AccountActions({ verified }: { verified: boolean }) {
  const router = useRouter();
  const [busy, setBusy] = useState(''), [message, setMessage] = useState(''), [error, setError] = useState(''), [sent, setSent] = useState(false);
  async function action(kind: 'verify' | 'claim' | 'logout', code?: string) {
    setBusy(kind); setMessage(''); setError('');
    try {
      if (kind === 'logout') { await fetchApi('/auth/logout', { method: 'POST' }); router.replace('/discover'); router.refresh(); return; }
      if (kind === 'claim') { const result = await fetchApi<{ message: string; may_have_more: boolean }>('/rsvps/claim', { method: 'POST' }); setMessage(result.message + (result.may_have_more ? '. More registrations may remain; connect again to continue.' : '')); }
      if (kind === 'verify') {
        await fetchApi(code ? '/auth/email-verification/confirm' : '/auth/email-verification/request', { method: 'POST', ...(code ? { body: JSON.stringify({ code }) } : {}) });
        if (code) { setMessage('Your email is verified.'); router.refresh(); } else { setSent(true); setMessage('A verification code has been queued for your email.'); }
      }
    } catch (err) { setError(err instanceof Error ? err.message : 'The request failed. Please try again.'); }
    finally { setBusy(''); }
  }
  function confirm(e: FormEvent<HTMLFormElement>) { e.preventDefault(); void action('verify', String(new FormData(e.currentTarget).get('code')).trim()); }
  return <div className="account-sections">{!verified && <section><h2>Verify your email</h2><p>Confirm that this email belongs to you before connecting guest registrations.</p><button className="app-button app-button-outline" disabled={!!busy} onClick={() => void action('verify')}>{busy === 'verify' ? 'Please wait…' : sent ? 'Send another code' : 'Send verification code'}</button>{sent && <form className="journey-form" onSubmit={confirm}><Input label="Verification code" name="code" autoComplete="one-time-code" required maxLength={128} /><button className="app-button app-button-primary" disabled={!!busy}>Verify email</button></form>}</section>}<section><h2>Registered without an account?</h2><p>Connect registrations made with your email address. Your email must be verified first.</p><button className="app-button app-button-outline" disabled={!!busy || !verified} onClick={() => void action('claim')}>{busy === 'claim' ? 'Connecting…' : 'Connect guest registrations'}</button></section>{error && <p className="journey-error" role="alert">{error}</p>}{message && <p className="journey-notice" role="status">{message}</p>}<section><h2>Sign out</h2><p>Sign out of Hackways on this browser.</p><button className="quiet-link" disabled={!!busy} onClick={() => void action('logout')}>{busy === 'logout' ? 'Signing out…' : 'Sign out'}</button></section></div>;
}
