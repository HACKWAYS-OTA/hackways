'use client';
import { useState, type FormEvent } from 'react';
import { AuthSubmit } from '@hackways/ui/auth';
import { AdminAccount } from './AdminAccount';
export function AdminVerification({ next }: { next: string }) {
  const [code, setCode] = useState(''), [pending, setPending] = useState(false), [queued, setQueued] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  async function call(action: 'request' | 'confirm') {
    if (pending) return; setPending(true); setError('');
    const controller = new AbortController(), timer = setTimeout(() => controller.abort(), 15000);
    try {
      const response = await fetch(`/api/v1/auth/admin/verification/${action}`, { method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(action === 'confirm' ? { code } : {}), signal: controller.signal });
      if (response.status === 401) { window.location.replace(`/login?switch=1&next=${encodeURIComponent(next)}`); return; }
      if (response.status === 204) { window.location.replace(next); return; }
      const body: unknown = await response.json().catch(() => null);
      const problem = body && typeof body === 'object' && 'error' in body ? body.error : null;
      if (!response.ok) { if (response.status === 429) setQueued(true); throw new Error(problem && typeof problem === 'object' && 'message' in problem && typeof problem.message === 'string' ? problem.message : 'Verification is unavailable. Try again.'); }
      setQueued(true); setMessage('Verification email queued. Enter the six-digit code from your email.');
    } catch (failure) { setError(failure instanceof Error && failure.name !== 'AbortError' ? failure.message : 'The request timed out. Your sign-in has been preserved; try again.'); }
    finally { clearTimeout(timer); setPending(false); }
  }
  function submit(event: FormEvent) { event.preventDefault(); void call('confirm'); }
  return <div className="admin-verification"><p className="hw-auth-note">Confirm this device once every 24 hours. An incorrect or expired code signs this device out.</p><button type="button" className="app-button app-button-outline" disabled={pending} onClick={() => { void call('request'); }}>{pending && !queued ? 'Requesting…' : queued ? 'Request another code' : 'Email me a code'}</button><form className="hw-auth-form" onSubmit={submit}><label className="field" htmlFor="admin-otp">Verification code<input id="admin-otp" className="app-input admin-otp" value={code} onChange={event => setCode(event.target.value.replace(/\D/g, '').slice(0, 6))} inputMode="numeric" autoComplete="one-time-code" pattern="[0-9]{6}" minLength={6} maxLength={6} disabled={pending} required /></label>{message && <p className="hw-auth-note" role="status">{message}</p>}{error && <p className="hw-auth-error" role="alert">{error}</p>}<AuthSubmit disabled={pending || !queued}>{pending && queued ? 'Verifying…' : 'Verify device'}</AuthSubmit></form><AdminAccount label="Sign out of this device" /></div>;
}
