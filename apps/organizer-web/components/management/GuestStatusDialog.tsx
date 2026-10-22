'use client';

import { Checkbox } from '@hackways/ui/checkbox';
import { RadioGroup } from '@hackways/ui/radio-group';
import { useEffect, useRef, useState, type FormEvent } from 'react';
import { ApiClientError, fetchApi } from '@/lib/api';
import type { EditableGuest, GuestStatusOption } from '@/lib/guests';
import { registrationLabels } from '@/lib/journeys';
import { Icon } from '@/components/ui/Icon';
import './guest-status.css';

export function GuestStatusDialog({ eventId, guest, options, initialStatus, resend = false, onClose, onBusy, onUpdated, onReload }: { eventId: string; guest: EditableGuest; options: GuestStatusOption[]; initialStatus: string; resend?: boolean; onClose: () => void; onBusy: (busy: boolean) => void; onUpdated: (message: string) => void; onReload: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null), lock = useRef(false), attempt = useRef<{ body: string; key: string } | null>(null);
  const [status, setStatus] = useState(resend ? guest.status : options.some(option => option.value === initialStatus) ? initialStatus : ''), [notify, setNotify] = useState(!!guest.guest_email), [message, setMessage] = useState('');
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [conflict, setConflict] = useState(false);
  useEffect(() => { dialog.current?.showModal(); }, []);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current || (!resend && !options.some(option => option.value === status))) return;
    lock.current = true; setBusy(true); onBusy(true); setError(''); setConflict(false);
    try {
      const body = JSON.stringify({ status, updated_at: guest.updated_at, notify_guest: notify, message: notify ? message.trim() : '', ...(resend ? { resend_notification: true } : {}) });
      if (!attempt.current || attempt.current.body !== body) attempt.current = { body, key: crypto.randomUUID() };
      const result = await fetchApi<{ data: { id: string; status: string }; notification_queued: boolean }>(`/events/${eventId}/registrations/by-code/${encodeURIComponent(guest.public_code)}/status`, { method: 'POST', headers: { 'Idempotency-Key': attempt.current.key }, body });
      if (result.data.id !== guest.id || result.data.status !== status) throw new Error('The status response could not be confirmed. Retry without changing the form.');
      onUpdated(resend ? 'A new guest notification is queued.' : result.notification_queued ? 'Guest status updated. A notification is queued.' : 'Guest status updated without notifying the guest.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'The status could not be updated. Your entries are still here.');
      setConflict(err instanceof ApiClientError && ['REGISTRATION_CHANGED', 'REGISTRATION_EXPIRED', 'STATUS_CHANGE_UNAVAILABLE'].includes(err.code));
    } finally { lock.current = false; setBusy(false); onBusy(false); }
  }
  const name = guest.guest_name || guest.guest_email || 'Guest registration';
  return <dialog ref={dialog} className="guest-status-dialog" aria-labelledby="guest-status-title" onCancel={event => { event.preventDefault(); event.stopPropagation(); if (!lock.current) onClose(); }} onClick={event => { event.stopPropagation(); if (event.target === event.currentTarget && !lock.current) { const rect = event.currentTarget.getBoundingClientRect(); if (event.clientX < rect.left || event.clientX > rect.right || event.clientY < rect.top || event.clientY > rect.bottom) onClose(); } }}>
    <header><h2 id="guest-status-title">{resend ? 'Send guest notification' : 'Change guest status'}</h2><button type="button" className="icon-button" aria-label="Close status editor" disabled={busy} onClick={onClose}><Icon name="close" size={20} /></button></header>
    <div className="guest-status-person"><span className="host-guest-avatar" aria-hidden="true">{name.slice(0, 1).toUpperCase()}</span><div><h3>{name}</h3>{guest.guest_email && <p>{guest.guest_email}</p>}</div><span className={`guest-status-label status-${guest.status}`}>{registrationLabels[guest.status]}</span></div>
    <form onSubmit={submit} aria-busy={busy}><fieldset disabled={busy}><legend className="visually-hidden">Guest status and notification</legend>
      {!resend && <div className="guest-status-field"><span>Change status to</span><RadioGroup label="Change status to" name="guest-status" layout="stack" options={options} value={status} onValueChange={setStatus} required disabled={busy} /></div>}
      {!resend && <label className="guest-status-notify"><Checkbox checked={notify} onCheckedChange={setNotify} disabled={!guest.guest_email || busy} /><span>Notify guest</span></label>}
      {notify && <label className="guest-status-field"><span className="visually-hidden">Message to the guest (optional)</span><textarea rows={3} value={message} onChange={event => setMessage(event.target.value)} maxLength={2000} placeholder="Add an optional message…" /><small>Included in the status notification email.</small></label>}
    </fieldset>
    {error && <div className="guest-status-error" role="alert"><p>{error}</p>{conflict && <button type="button" className="quiet-link" disabled={busy} onClick={onReload}>Close and reload guest details</button>}</div>}
    <button className="app-button app-button-primary" disabled={busy || (!resend && !options.some(option => option.value === status)) || conflict}>{busy ? 'Saving...' : resend ? 'Send notification' : 'Update status'}</button></form>
  </dialog>;
}
