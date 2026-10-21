'use client';
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import type { ManagedEvent } from '@/lib/management';
import { fetchApi } from '@/lib/api';
import { Icon } from '@/components/ui/Icon';
export function CheckInForm({ event, open }: { event: ManagedEvent; open: boolean }) {
  const router = useRouter();
  const lock = useRef(false);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  async function checkIn(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); if (lock.current) return;
    const form = e.currentTarget, code = String(new FormData(form).get('code') || '').trim();
    lock.current = true; setBusy(true); setError(''); setMessage('');
    try { await fetchApi(`/events/${event.id}/check-in`, { method: 'POST', body: JSON.stringify({ code }) }); setMessage('Guest checked in. Their ticket has been marked as used.'); form.reset(); router.refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : 'Check-in failed. The code is still here so you can retry.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <section className="host-entry"><div><Icon name="check" size={23} /><h2>Welcome them in.</h2><p>Ask for the entry code from the guest’s My events page.</p></div>{open ? <form onSubmit={checkIn} aria-busy={busy}><label htmlFor="entry-code">Entry code</label><div className="host-entry-field"><input id="entry-code" name="code" required maxLength={2048} autoComplete="off" autoCapitalize="none" spellCheck={false} disabled={busy} /><button className="app-button app-button-primary" disabled={busy}>{busy ? 'Checking in…' : 'Check in'}<Icon name="arrow" size={16} /></button></div><p className="field-help">You can also find a guest below and check them in manually with a reason.</p></form> : <p className="host-entry-closed">Check-in opens 24 hours before the event and closes when it ends. The event must be published.</p>}{error && <p className="journey-error" role="alert">{error}</p>}{message && <p className="journey-notice" role="status">{message}</p>}</section>;
}
