'use client';
import { useState } from 'react';
import type { ManagedEvent } from '@/lib/management';
import { site } from '@/lib/site';
import { Icon } from '@/components/ui/Icon';
export function ShareEvent({ event }: { event: ManagedEvent }) {
  const [message, setMessage] = useState('');
  const url = `${site.app}/events/${event.slug}`;
  async function copy() {
    try { await navigator.clipboard.writeText(url); setMessage('Event link copied.'); }
    catch { setMessage('Copy the event address from the field below.'); }
  }
  return <section className="host-share"><h2>Event link</h2><label htmlFor="share-event-address" className="visually-hidden">Event page address</label><div className="host-share-field"><input id="share-event-address" value={url} readOnly onFocus={e => e.target.select()} /><button type="button" onClick={() => void copy()} disabled={event.status !== 'published'}>Copy link</button></div>{event.status === 'published' && <a className="quiet-link" href={url} target="_blank" rel="noopener noreferrer">Open event page<Icon name="external" size={15} /></a>}{message && <p className="field-help" role="status">{message}</p>}</section>;
}
