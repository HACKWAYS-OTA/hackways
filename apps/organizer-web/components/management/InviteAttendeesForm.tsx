'use client';
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Checkbox } from '@hackways/ui/checkbox';
import { fetchApi } from '@/lib/api';

export function InviteAttendeesForm({ eventId, available }: { eventId: string; available: boolean }) {
  const router = useRouter(), lock = useRef(false), attempt = useRef<{ body: string; key: string } | null>(null);
  const [emails, setEmails] = useState(''), [message, setMessage] = useState(''), [consent, setConsent] = useState(false);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [notice, setNotice] = useState('');
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current || !available || !consent) return;
    const recipients = Array.from(new Set(emails.split(/[\s,;]+/).map(email => email.trim().toLowerCase()).filter(Boolean)));
    if (!recipients.length || recipients.length > 50) { setError('Enter between 1 and 50 email addresses.'); return; }
    const body = JSON.stringify({ emails: recipients.sort(), message: message.trim() });
    if (attempt.current?.body !== body) attempt.current = { body, key: crypto.randomUUID() };
    lock.current = true; setBusy(true); setError(''); setNotice('');
    try {
      const result = await fetchApi<{ queued: number; skipped: number }>(`/events/${eventId}/invitations`, { method: 'POST', body, headers: { 'Idempotency-Key': attempt.current.key } });
      setNotice(`${result.queued} invitation${result.queued === 1 ? '' : 's'} queued for email delivery.${result.skipped ? ` ${result.skipped} skipped because they were already invited or registered.` : ''}`);
      setEmails(''); setMessage(''); setConsent(false); attempt.current = null; router.refresh();
    } catch (err) { setError(err instanceof Error ? err.message : 'Invitations could not be queued. Your entries are still here.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <form className="invite-form" onSubmit={submit} aria-busy={busy}><fieldset disabled={busy || !available}><legend className="visually-hidden">Invitation recipients and message</legend><div className="field"><label htmlFor="invite-emails">Email addresses</label><textarea id="invite-emails" className="app-input" value={emails} onChange={event => setEmails(event.target.value)} required maxLength={13000} rows={5} placeholder="One email address per line" aria-describedby="invite-email-help" /><p id="invite-email-help" className="field-help">Add up to 50 people. Separate addresses with a new line or comma.</p></div><div className="field"><label htmlFor="invite-message">Personal message <span>(optional)</span></label><textarea id="invite-message" className="app-input" value={message} onChange={event => setMessage(event.target.value)} maxLength={2000} rows={4} /></div><label className="invite-consent"><Checkbox checked={consent} onCheckedChange={setConsent} disabled={busy || !available} required /><span>These people expect an invitation from me.</span></label></fieldset>{error && <p className="journey-error" role="alert">{error}</p>}{notice && <p className="journey-notice" role="status">{notice}</p>}<button className="app-button app-button-primary" disabled={busy || !available || !consent || !emails.trim()}>{busy ? 'Sending invitations…' : 'Send invitations'}</button></form>;
}
