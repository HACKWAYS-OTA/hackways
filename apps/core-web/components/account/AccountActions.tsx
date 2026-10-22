'use client';
import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
export function AccountActions({ verified }: { verified: boolean }) {
  const router = useRouter();
  const [busy, setBusy] = useState(''), [message, setMessage] = useState(''), [error, setError] = useState('');
  async function action(kind: 'claim' | 'logout') {
    setBusy(kind); setMessage(''); setError('');
    try {
      if (kind === 'logout') { await fetchApi('/auth/logout', { method: 'POST' }); router.replace('/discover'); router.refresh(); return; }
      if (kind === 'claim') { const result = await fetchApi<{ message: string; may_have_more: boolean }>('/rsvps/claim', { method: 'POST' }); setMessage(result.message + (result.may_have_more ? '. More registrations may remain; connect again to continue.' : '')); }

    } catch (err) { setError(err instanceof Error ? err.message : 'The request failed. Please try again.'); }
    finally { setBusy(''); }
  }
  return <div className="account-sections">{!verified && <section><h2>Verify your email</h2><Link className="app-button app-button-outline" href="/auth/verify-email?next=%2Faccount">Verify email</Link></section>}<section><h2>Password</h2><Link className="app-button app-button-outline" href="/auth/forgot-password?next=%2Faccount">Reset password</Link></section><section><h2>Registered without an account?</h2><p>Connect registrations made with your email address. Your email must be verified first.</p><button className="app-button app-button-outline" disabled={!!busy || !verified} onClick={() => void action('claim')}>{busy === 'claim' ? 'Connecting…' : 'Connect guest registrations'}</button></section>{error && <p className="journey-error" role="alert">{error}</p>}{message && <p className="journey-notice" role="status">{message}</p>}<section><h2>Sign out</h2><p>Sign out of Hackways on this browser.</p><button className="quiet-link" disabled={!!busy} onClick={() => void action('logout')}>{busy === 'logout' ? 'Signing out…' : 'Sign out'}</button></section></div>;
}
