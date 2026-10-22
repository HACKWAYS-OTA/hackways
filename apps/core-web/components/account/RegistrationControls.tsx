'use client';
import { useEffect, useState, type FormEvent } from 'react';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import type { Registration } from '@/types/api';
import { fetchApi } from '@/lib/api';
export function RegistrationControls({ registration, startsAt, initialTicket, initialTicketError }: { registration: Registration; startsAt?: string; initialTicket?: { qr_image: string; short_code: string } | null; initialTicketError?: string }) {
  const router = useRouter();
  const [busy, setBusy] = useState(''), [error, setError] = useState(initialTicketError || ''), [message, setMessage] = useState(''), [confirm, setConfirm] = useState(false), [ticket, setTicket] = useState<{ qr_image: string; short_code: string } | null>(initialTicket || null);
  useEffect(() => { setTicket(initialTicket || null); }, [initialTicket]);
  const beforeStart = !startsAt || Date.parse(startsAt) > Date.now();
  const active = ['confirmed', 'pending_approval', 'waitlisted', 'offered'].includes(registration.status);
  async function action(kind: string, notes?: string) {
    setBusy(kind); setError(''); setMessage('');
    try {
      if (kind === 'ticket') { setTicket(await fetchApi<{ qr_image: string; short_code: string }>(`/rsvps/${registration.id}/ticket`, { cache: 'no-store' })); }
      else { await fetchApi(`/rsvps/${registration.id}${kind === 'notes' ? '' : `/${kind}`}`, { method: kind === 'notes' ? 'PATCH' : 'POST', ...(kind === 'notes' ? { body: JSON.stringify({ notes }) } : {}) }); setMessage(kind === 'cancel' ? 'Your registration is cancelled.' : kind === 'accept' ? 'Your place is confirmed.' : 'Your note is saved.'); setConfirm(false); setTicket(null); router.refresh(); }
    } catch (err) { setError(err instanceof Error ? err.message : 'The request failed. Please try again.'); }
    finally { setBusy(''); }
  }
  function save(e: FormEvent<HTMLFormElement>) { e.preventDefault(); void action('notes', String(new FormData(e.currentTarget).get('notes'))); }
  return <section className="registration-controls"><div className="journey-actions">{registration.status === 'confirmed' && !ticket && <button className="app-button app-button-primary" disabled={!!busy} onClick={() => void action('ticket')}>{busy === 'ticket' ? 'Loading…' : 'Show my pass'}</button>}{registration.status === 'offered' && beforeStart && <button className="app-button app-button-primary" disabled={!!busy} onClick={() => void action('accept')}>{busy === 'accept' ? 'Confirming…' : 'Accept this place'}</button>}{active && beforeStart && <button className="quiet-link" disabled={!!busy} onClick={() => setConfirm(!confirm)}>{registration.status === 'waitlisted' ? 'Leave waiting list' : 'Cancel registration'}</button>}</div>
    {confirm && <div className="inline-confirm"><h2>Cancel your registration?</h2><p>Your place will be released. You may need to register again if your plans change.</p><div className="journey-actions"><button className="app-button app-button-outline" disabled={!!busy} onClick={() => setConfirm(false)}>Keep my registration</button><button className="app-button app-button-primary" disabled={!!busy} onClick={() => void action('cancel')}>{busy === 'cancel' ? 'Cancelling…' : 'Yes, cancel'}</button></div></div>}
    {ticket && registration.status === 'confirmed' && <section className="ticket-pass" aria-label="Your event pass"><h2>Your entry pass</h2><p>Show this QR code at check-in.</p><Image unoptimized src={ticket.qr_image} alt="Your unique event entry QR code" width={384} height={384} /><p className="ticket-pass-reference">Pass reference <strong>{ticket.short_code}</strong></p><button className="quiet-link" onClick={() => setTicket(null)}>Hide pass</button></section>}
    {active && beforeStart && <form className="journey-form" onSubmit={save}><div className="field"><label htmlFor="registration-notes">A note for the host (optional)</label><textarea className="app-input" id="registration-notes" name="notes" defaultValue={registration.notes || ''} maxLength={2000} rows={3} /></div><button className="app-button app-button-outline" disabled={!!busy}>{busy === 'notes' ? 'Saving…' : 'Save note'}</button></form>}
    {error && <p className="journey-error" role="alert">{error}</p>}{message && <p className="journey-notice" role="status">{message}</p>}
  </section>;
}
