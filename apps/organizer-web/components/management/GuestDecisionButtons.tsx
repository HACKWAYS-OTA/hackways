"use client";

import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import type { EditableGuest, Guest, GuestStatusOption } from '@/lib/guests';
import { ApiClientError, fetchApi } from '@/lib/api';
import { decisionBody, saveGuestDecision, type DecisionAttempt, type GuestDecision } from '@/lib/guest-decisions';
import { Icon } from '@/components/ui/Icon';

export function GuestDecisionButtons({ eventId, guest, onUpdated }: { eventId: string; guest: Guest; onUpdated: (message: string) => void }) {
  const router = useRouter();
  const lock = useRef(false), mounted = useRef(true);
  const attempt = useRef<(DecisionAttempt & { status: GuestDecision }) | null>(null);
  const [busy, setBusy] = useState<GuestDecision | null>(null), [error, setError] = useState('');
  const [saved, setSaved] = useState<GuestDecision | null>(null), [changed, setChanged] = useState(false);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  const name = guest.guest_name || guest.guest_email || 'this guest';

  async function decide(status: GuestDecision) {
    if (lock.current || saved || changed) return;
    lock.current = true; setBusy(status); setError('');
    try {
      if (!attempt.current) {
        const detail = await fetchApi<{ data: EditableGuest; status_options: GuestStatusOption[] }>(
          `/events/${eventId}/registrations/by-code/${encodeURIComponent(guest.public_code)}`, { cache: 'no-store' },
        );
        if (detail.data.id !== guest.id || detail.data.status !== 'pending_approval' || !detail.status_options.some(option => option.value === status)) {
          setChanged(true); router.refresh();
          throw new Error('This application has changed or the action is no longer available. Updated guest details are being loaded.');
        }
        attempt.current = { status, body: decisionBody(detail.data, status, !!detail.data.guest_email), key: crypto.randomUUID() };
      }
      // An uncertain request is retried with its original version, body and key.
      // Do not allow the opposite action until its outcome has been resolved.
      if (attempt.current.status !== status) return;
      const notice = await saveGuestDecision(eventId, guest, status, attempt.current);
      if (mounted.current) { setSaved(status); onUpdated(`${name}: ${notice}`); router.refresh(); }
    } catch (err) {
      if (!mounted.current) return;
      const conflict = err instanceof ApiClientError && ['REGISTRATION_CHANGED', 'REGISTRATION_EXPIRED', 'STATUS_CHANGE_UNAVAILABLE', 'CAPACITY_FULL', 'EVENT_FULL'].includes(err.code);
      if (conflict) { attempt.current = null; setChanged(true); router.refresh(); }
      // Other failures keep the original request identity for a safe retry.
      setError(err instanceof Error ? err.message : 'The decision could not be saved. Please retry.');
    } finally {
      lock.current = false;
      if (mounted.current) setBusy(null);
    }
  }
  return <>
    <div className="guest-decisions" aria-label={`Application actions for ${name}`} aria-busy={!!busy}>
      {saved ? <span className={`guest-status status-${saved}`} role="status">{saved === 'confirmed' ? 'Going' : 'Declined'}</span> : <>
        <button type="button" className="guest-row-approve" aria-label={`Approve application from ${name}`} disabled={!!busy || changed || !!attempt.current && attempt.current.status !== 'confirmed'} onClick={() => void decide('confirmed')}><Icon name="check" size={16} />{busy === 'confirmed' ? 'Approving…' : 'Approve'}</button>
        <button type="button" className="guest-row-decline" aria-label={`Decline application from ${name}`} disabled={!!busy || changed || !!attempt.current && attempt.current.status !== 'rejected'} onClick={() => void decide('rejected')}><Icon name="close" size={16} />{busy === 'rejected' ? 'Declining…' : 'Decline'}</button>
      </>}
    </div>
    {error && <div className="guest-row-feedback" role="alert"><p>{error}</p>{changed && <button type="button" className="quiet-link" onClick={() => { setChanged(false); setError(''); attempt.current = null; router.refresh(); }}>Load current application</button>}</div>}
  </>;
}
