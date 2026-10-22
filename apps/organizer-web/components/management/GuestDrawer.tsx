'use client';
import { useEffect, useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import type { ManagedEvent } from '@/lib/management';
import type { EditableGuest, GuestStatusOption } from '@/lib/guests';
import type { RegistrationAnswerSnapshot } from '@/types/api';
import { fetchApi } from '@/lib/api';
import { dateLabel, registrationLabels } from '@/lib/journeys';
import { Icon } from '@/components/ui/Icon';
import { safeLink } from '@/lib/dashboard';
import { GuestStatusDialog } from './GuestStatusDialog';
type Detail = { data: EditableGuest; answers: RegistrationAnswerSnapshot[]; status_options: GuestStatusOption[] };
export function GuestDrawer({ event, code, codes, onSelect, onClose, checkinOpen, requestedDecision = null }: { event: ManagedEvent; code: string | null; codes: string[]; onSelect: (code: string) => void; onClose: () => void; checkinOpen: boolean; requestedDecision?: 'approve' | 'reject' | null }) {
  const router = useRouter(), dialog = useRef<HTMLDialogElement>(null), lock = useRef(false);
  const [detail, setDetail] = useState<Detail | null>(null), [loading, setLoading] = useState(true), [error, setError] = useState(''), [message, setMessage] = useState(''), [version, setVersion] = useState(0);
  const [intent, setIntent] = useState<'check-in' | null>(null), [busy, setBusy] = useState(false);
  const [statusOpen, setStatusOpen] = useState(false), [statusChoice, setStatusChoice] = useState(''), [resend, setResend] = useState(false);
  useEffect(() => { if (code && !dialog.current?.open) dialog.current?.showModal(); else if (!code && dialog.current?.open) dialog.current.close(); }, [code]);
  useEffect(() => {
    if (!code) return;
    const controller = new AbortController(); setDetail(null); setLoading(true); setError(''); setIntent(null); setStatusOpen(false); setResend(false);
    fetchApi<Detail>(`/events/${event.id}/registrations/by-code/${encodeURIComponent(code)}`, { signal: controller.signal, cache: 'no-store' })
      .then(result => { if (!controller.signal.aborted) { setDetail(result); if (requestedDecision && (result.status_options || []).some(option => option.value === (requestedDecision === 'approve' ? 'confirmed' : 'rejected'))) { setStatusChoice(requestedDecision === 'approve' ? 'confirmed' : 'rejected'); setStatusOpen(true); } } })
      .catch(err => { if (!controller.signal.aborted) setError(err instanceof Error ? err.message : 'Guest details could not be loaded.'); })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [code, event.id, event.status, event.starts_at, version, requestedDecision]);
  useEffect(() => { setMessage(''); }, [code]);
  async function act(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); if (!detail || detail.data.public_code !== code || !intent || lock.current) return;
    const reason = String(new FormData(e.currentTarget).get('reason') || '').trim();
    lock.current = true; setBusy(true); setError(''); setMessage('');
    try {
      if (intent === 'check-in') await fetchApi(`/events/${event.id}/check-in`, { method: 'POST', body: JSON.stringify({ ticket_id: detail.data.ticket_id, reason }) });
      setMessage('Guest checked in.');
      setIntent(null); setVersion(current => current + 1); router.refresh();
    } catch (err) { setError(err instanceof Error ? err.message : 'This change could not be saved. Please try again.'); }
    finally { lock.current = false; setBusy(false); }
  }
  const guest = detail?.data.public_code === code ? detail.data : null, index = code ? codes.indexOf(code) : -1;
  return <dialog ref={dialog} className="host-guest-drawer" aria-labelledby="guest-drawer-heading" onCancel={e => { e.preventDefault(); if (!lock.current && !busy) onClose(); }} onClick={e => { if (e.target === dialog.current && !lock.current && !busy) onClose(); }}>
    <div className="host-drawer-surface"><header className="host-drawer-header"><h2 id="guest-drawer-heading">Guest details</h2><div><button type="button" disabled={busy || index <= 0} aria-label="Previous guest" onClick={() => onSelect(codes[index - 1])}><Icon name="arrow" className="host-arrow-back" size={18} /></button><button type="button" disabled={busy || index < 0 || index >= codes.length - 1} aria-label="Next guest" onClick={() => onSelect(codes[index + 1])}><Icon name="arrow" size={18} /></button><button type="button" disabled={busy} aria-label="Close guest details" onClick={onClose}><Icon name="close" size={20} /></button></div></header>
      <div className="host-drawer-content" aria-busy={loading || busy}>
        {message && <p className="journey-notice" role="status">{message}</p>}
        {error && <div className="journey-error" role="alert"><p>{error}</p>{!guest && <button type="button" className="quiet-link" onClick={() => setVersion(current => current + 1)}>Try again</button>}</div>}
        {loading && <div className="host-loading" role="status"><span /><span /><p>Loading guest details…</p></div>}
        {guest && <><section className="host-drawer-person"><span className="host-guest-avatar" aria-hidden="true">{(guest.guest_name || guest.guest_email || 'G').slice(0, 1).toUpperCase()}</span><div><h3>{guest.guest_name || 'Guest registration'}</h3><p>{guest.guest_email}</p></div>{(detail?.status_options || []).length ? <button type="button" className={`guest-status-trigger status-${guest.status}`} aria-haspopup="dialog" aria-label={`Change status for ${guest.guest_name || guest.guest_email || guest.public_code}`} disabled={busy} onClick={() => { setResend(false); setStatusChoice(''); setStatusOpen(true); }}>{registrationLabels[guest.status]}<Icon name="edit" size={15} /></button> : <span className={`guest-status-label status-${guest.status}`} title={guest.checked_in_at ? 'Checked-in registrations cannot change status.' : 'No status changes are available for this registration.'}>{guest.checked_in_at && guest.status === 'confirmed' ? 'Checked in' : registrationLabels[guest.status]}</span>}</section>
          <dl className="host-guest-facts"><div><dt>Registration</dt><dd>{guest.public_code}</dd></div><div><dt>Registered</dt><dd>{dateLabel(guest.registered_at, event.timezone)}</dd></div>{guest.checked_in_at && <div><dt>Checked in</dt><dd>{dateLabel(guest.checked_in_at, event.timezone)}</dd></div>}{guest.offer_expires_at && guest.status === 'offered' && <div><dt>Offer expires</dt><dd>{dateLabel(guest.offer_expires_at, event.timezone)}</dd></div>}</dl>
          {guest.guest_email && ['confirmed', 'rejected', 'cancelled'].includes(guest.status) && event.status === 'published' && Date.now() < Date.parse(event.starts_at) && <button type="button" className="app-button app-button-outline" disabled={busy} onClick={() => { setResend(true); setStatusChoice(guest.status); setStatusOpen(true); }}>Send notification</button>}
          {guest.status === 'confirmed' && guest.ticket_status === 'issued' && guest.ticket_id && checkinOpen && !intent && <section className="host-drawer-decision"><h3>Manual check-in</h3><p>Verify their identity before checking them in manually.</p><button className="app-button app-button-primary" onClick={() => setIntent('check-in')}>Check in this guest</button></section>}
          {intent && <form className="host-drawer-decision" onSubmit={act}><h3>Confirm manual check-in?</h3><p>Your reason is recorded in the event history.</p>{intent === 'check-in' && <div className="field"><label htmlFor="drawer-reason">Reason for manual check-in</label><input id="drawer-reason" className="app-input" name="reason" required minLength={3} maxLength={500} disabled={busy} /></div>}<div className="journey-actions"><button className="app-button app-button-primary" disabled={busy}>{busy ? 'Saving…' : 'Confirm check-in'}</button><button type="button" className="quiet-link" disabled={busy} onClick={() => setIntent(null)}>Go back</button></div></form>}
          <section className="host-drawer-answers"><h3>Registration answers</h3>{detail && detail.answers.length ? <dl>{detail.answers.map(answer => { const url = answer.kind === 'url' ? safeLink(answer.text) : null; return <div key={answer.code}><dt>{answer.label}</dt><dd>{url ? <a className="host-visit-answer" href={url} target="_blank" rel="noopener noreferrer">{answer.text}<Icon name="external" size={15} /></a> : answer.kind === 'checkbox' ? answer.checked ? 'Confirmed' : 'Not selected' : answer.text || 'Not provided'}</dd></div>; })}</dl> : <p className="host-muted">No custom answers were submitted with this registration.</p>}</section>
          {guest.notes && <section className="host-drawer-answers"><h3>A note from the guest</h3><p className="host-answer-text">{guest.notes}</p></section>}
        </>}
      </div>
    </div>
    {statusOpen && guest && detail && <GuestStatusDialog key={`${guest.public_code}-${guest.updated_at}`} eventId={event.id} guest={guest} options={detail.status_options || []} initialStatus={statusChoice} resend={resend} onClose={() => setStatusOpen(false)} onBusy={setBusy} onUpdated={notice => { setStatusOpen(false); setMessage(notice); setVersion(value => value + 1); router.refresh(); }} onReload={() => { setStatusOpen(false); setVersion(value => value + 1); }} />}
  </dialog>;
}
