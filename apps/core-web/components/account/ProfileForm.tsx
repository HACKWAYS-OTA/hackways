'use client';
import { useState, useRef, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
import type { User } from '@/types/api';

export function ProfileForm({ user }: { user: User }) {
  const router = useRouter(), lock = useRef(false);
  const [name, setName] = useState(user.display_name), [saved, setSaved] = useState(user.display_name);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current) return;
    lock.current = true; setBusy(true); setError(''); setMessage('');
    try { const result = await fetchApi<{ user: User }>('/auth/me', { method: 'PATCH', body: JSON.stringify({ display_name: name.trim() }) }); setName(result.user.display_name); setSaved(result.user.display_name); setMessage('Your profile is saved.'); router.refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : 'Your profile could not be saved. Please try again.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <section className="profile-section"><div><h2>Personal details</h2><p>Your name appears on your event registrations.</p></div><form className="profile-form" onSubmit={submit} aria-busy={busy}><div className="field"><label htmlFor="profile-name">Display name</label><input id="profile-name" className="app-input" value={name} onChange={event => setName(event.target.value)} required maxLength={100} autoComplete="name" disabled={busy} /></div><div className="field"><label htmlFor="profile-email">Email address</label><input id="profile-email" className="app-input" value={user.email} readOnly type="email" aria-describedby="profile-email-note" /><p id="profile-email-note" className="field-help">Your sign-in email is linked to this account.</p></div>{error && <p className="journey-error" role="alert">{error}</p>}{message && <p className="journey-notice" role="status">{message}</p>}<button className="app-button app-button-primary" disabled={busy || name.trim() === saved}>{busy ? 'Saving…' : 'Save changes'}</button></form></section>;
}
