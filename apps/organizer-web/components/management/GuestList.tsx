'use client';
import { useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import type { ManagedEvent } from '@/lib/management';
import type { Guest } from '@/lib/guests';
import { dateLabel, registrationLabels } from '@/lib/journeys';
import { Icon } from '@/components/ui/Icon';
import { GuestDrawer } from './GuestDrawer';
import './guest-list.css';
export function GuestList({ event, guests, checkIn = false, checkinOpen = false }: { event: ManagedEvent; guests: Guest[]; checkIn?: boolean; checkinOpen?: boolean }) {
  const params = useSearchParams(), requestedCode = params.get('guest');
  const [code, setCode] = useState(requestedCode);
  useEffect(() => { setCode(requestedCode); }, [requestedCode]);
  const [decision, setDecision] = useState<'approve' | 'reject' | 'check-in' | null>(checkIn && params.get('checkin')==='1' ? 'check-in' : null);
  const approvalOpen = event.status === 'published' && Date.now() < Date.parse(event.starts_at);
  function select(code: string | null, action: 'approve' | 'reject' | 'check-in' | null = null) { setDecision(action); setCode(code); const url = new URL(window.location.href); if (code) url.searchParams.set('guest', code); else url.searchParams.delete('guest'); if (action === 'check-in') url.searchParams.set('checkin','1'); else url.searchParams.delete('checkin'); window.history.replaceState(null, '', url.pathname + url.search); }
  return <>
    {guests.length === 0 ? <section className="host-empty"><Icon name="people" size={29} /><h2>No guests in this view.</h2><p>Try another name, email or status. New registrations appear here after guests sign up.</p></section> : <ul className="guest-list" aria-label={checkIn ? 'Guests for check-in' : 'Event guests'}>{guests.map(guest => {
      const name = guest.guest_name || guest.guest_email || `Registration ${guest.public_code}`;
      const canDecide = guest.status === 'pending_approval' && approvalOpen;
      const status = guest.checked_in_at && guest.status === 'confirmed' ? 'Checked in' : registrationLabels[guest.status];
      return <li className="guest-row" key={guest.public_code}>
        <button type="button" className="guest-open" aria-haspopup="dialog" onClick={() => select(guest.public_code)}>
          <span className="guest-identity"><span className="host-guest-avatar" aria-hidden="true">{name.slice(0, 1).toUpperCase()}</span><span className="guest-person"><span className="guest-name">{name}</span>{guest.guest_email && guest.guest_name && <span className="guest-email" title={guest.guest_email}>{guest.guest_email}</span>}</span></span>
          <span className={canDecide ? 'visually-hidden' : `guest-status status-${guest.status}`}>{status}</span>
          <time className="guest-time" dateTime={guest.registered_at} title={`Registered ${dateLabel(guest.registered_at, event.timezone)} (${event.timezone})`}>{dateLabel(guest.registered_at, event.timezone, { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' })}</time>
        </button>
        {checkIn && guest.status === 'confirmed' && !guest.checked_in_at && <button type="button" className="app-button app-button-outline guest-row-checkin" disabled={!checkinOpen} title={!checkinOpen ? 'Check-in opens 24 hours before the event.' : undefined} onClick={() => select(guest.public_code, 'check-in')}><Icon name="check" size={16} />Check in</button>}
        {canDecide && <div className="guest-decisions"><button type="button" className="host-guest-approve" aria-label={`Approve application from ${name}`} onClick={() => select(guest.public_code, 'approve')}><Icon name="check" size={16} />Approve</button><button type="button" className="host-guest-decline" aria-label={`Decline application from ${name}`} onClick={() => select(guest.public_code, 'reject')}><Icon name="close" size={16} />Decline</button></div>}
      </li>;
    })}</ul>}
    <GuestDrawer event={event} code={code} codes={guests.map(g => g.public_code)} onSelect={select} onClose={() => select(null)} checkinOpen={checkinOpen} requestedDecision={decision} />
  </>;
}
