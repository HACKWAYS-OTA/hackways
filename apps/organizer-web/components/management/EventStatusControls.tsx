'use client';
import { Checkbox } from '@hackways/ui/checkbox';
import { useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import type { ManagedEvent } from '@/lib/management';
import { fetchApi } from '@/lib/api';
export function EventStatusControls({ event }: { event: ManagedEvent }) {
  const router = useRouter(), lock = useRef(false);
  const [intent, setIntent] = useState(''), [ack, setAck] = useState(false), [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function apply(status: string) {
    if (lock.current) return; lock.current = true; setBusy(true); setError('');
    try {
      if (status === 'delete') { await fetchApi(`/events/${event.id}`, { method: 'DELETE', headers: { 'If-Match': String(event.revision) } }); router.replace('/'); }
      else { await fetchApi(`/events/${event.id}/status`, { method: 'POST', body: JSON.stringify({ status, revision: event.revision, acknowledge_attendees: ack }) }); router.refresh(); }
      setIntent('');
    } catch (err) { setError(err instanceof Error ? err.message : 'The change could not be saved.'); }
    finally { lock.current = false; setBusy(false); }
  }
  if (event.status === 'cancelled') return <section><h2>This event is cancelled.</h2><p>Registration is closed and its entry tickets are no longer valid. Guest records and event history are retained.</p></section>;
  const ended = Date.now() >= Date.parse(event.ends_at);
  return <section><h2>Publication & cancellation</h2><div className="journey-actions">{!ended && (event.status !== 'published' ? <button type="button" className="app-button app-button-primary" disabled={busy || Date.now() >= Date.parse(event.starts_at)} onClick={() => void apply('published')}>{busy ? 'Saving…' : 'Publish event'}</button> : <button type="button" className="app-button app-button-outline" disabled={busy} onClick={() => setIntent('unpublished')}>Pause registration</button>)}{!ended && <button type="button" className="quiet-link" disabled={busy} onClick={() => { setIntent('cancelled'); setAck(false); }}>Cancel event</button>}{['draft', 'unpublished'].includes(event.status) && <button type="button" className="quiet-link" disabled={busy} onClick={() => setIntent('delete')}>Remove event</button>}</div>{intent && <div className="journey-form"><p>{intent === 'delete' ? 'Remove this inactive event? Active registrations must be resolved first. This cannot be undone from the dashboard.' : intent === 'cancelled' ? 'Cancellation closes registration and invalidates all tickets. This cannot be undone.' : 'New registration pauses. Existing bookings remain, and the event page is unavailable until you publish again.'}</p>{intent === 'cancelled' && <label className="check-label"><Checkbox checked={ack} disabled={busy} onCheckedChange={setAck} />I understand registrations will be cancelled and attendee notifications queued.</label>}<div className="journey-actions"><button type="button" className="app-button app-button-primary" disabled={busy || (intent === 'cancelled' && !ack)} onClick={() => void apply(intent)}>{busy ? 'Saving…' : intent === 'delete' ? 'Remove this event' : intent === 'cancelled' ? 'Cancel this event' : 'Pause registration'}</button><button type="button" className="quiet-link" disabled={busy} onClick={() => setIntent('')}>Keep as it is</button></div></div>}{error && <p role="alert" className="journey-error">{error} <button type="button" className="quiet-link" disabled={busy} onClick={() => router.refresh()}>Reload latest event</button></p>}</section>;
}
