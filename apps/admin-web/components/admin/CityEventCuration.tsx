'use client';

import { useEffect, useId, useRef, useState } from 'react';
import { adminRequest } from '@/lib/admin-request';
import type { CityEvent } from '@/lib/cities';

type EventChoice = Pick<CityEvent, 'id' | 'title' | 'slug' | 'starts_at'> & { location_name: string | null };

function EventDate({ value, timezone }: { value: string; timezone: string }) {
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? <>Unavailable</> : <time dateTime={value}>{new Intl.DateTimeFormat('en-IN', { day: 'numeric', month: 'short', year: 'numeric', timeZone: timezone }).format(parsed)}</time>;
}

export function CityEventCuration({ reference, events, timezone, busy, onCurate }: { reference: number; events: CityEvent[]; timezone: string; busy: boolean; onCurate: (item: Pick<CityEvent, 'id'>, remove: boolean) => Promise<void> }) {
  const id = useId(), input = useRef<HTMLInputElement>(null);
  const [open, setOpen] = useState(false), [query, setQuery] = useState(''), [retry, setRetry] = useState(0);
  const [choices, setChoices] = useState<EventChoice[]>([]), [loading, setLoading] = useState(false), [error, setError] = useState('');
  useEffect(() => {
    if (!open) return;
    const controller = new AbortController(); setLoading(true); setError(''); setChoices([]);
    const timer = setTimeout(async () => {
      try { const result = await adminRequest<{ data: EventChoice[] }>(`core/cities/event-choices?q=${encodeURIComponent(query)}`, { signal: controller.signal }, true); if (!controller.signal.aborted) setChoices(result.data); }
      catch (failure) { if (!controller.signal.aborted) setError(failure instanceof Error ? failure.message : 'Events could not be loaded.'); }
      finally { if (!controller.signal.aborted) setLoading(false); }
    }, 250);
    return () => { clearTimeout(timer); controller.abort(); };
  }, [open, query, reference, retry]);
  useEffect(() => { if (open) input.current?.focus(); }, [open]);
  return <section className="city-curation" aria-labelledby={`${id}-title`}>
    <header className="city-curation-heading"><h2 id={`${id}-title`}>City events</h2><button type="button" className="app-button app-button-outline app-button-sm" disabled={busy} aria-expanded={open} aria-controls={`${id}-finder`} onClick={() => setOpen(value => !value)}>{open ? 'Close search' : 'Add events'}</button></header>
    <div className="city-events-table-wrap" role="region" aria-label="Curated city events" tabIndex={0}><table className="city-events-table"><caption className="sr-only">Events selected for this City Hub</caption><thead><tr><th scope="col">Event</th><th scope="col">Date</th><th scope="col">Visibility</th><th scope="col"><span className="sr-only">Action</span></th></tr></thead><tbody>
      {events.length ? events.map(item => <tr key={item.id}><th scope="row">{item.title}</th><td><EventDate value={item.starts_at} timezone={timezone} /></td><td>{item.status === 'published' && item.visibility === 'public' ? 'Public' : 'Hidden'}</td><td><button type="button" className="city-table-action" disabled={busy} onClick={() => void onCurate(item, true)} aria-label={`Remove ${item.title} from this City Hub`}>Remove</button></td></tr>) : <tr><td colSpan={4} className="city-table-empty">No city events yet.</td></tr>}
    </tbody></table></div>
    {open && <div className="city-event-finder" id={`${id}-finder`}><div className="city-event-search"><label htmlFor={`${id}-search`}>Find an event</label><input ref={input} id={`${id}-search`} className="app-input" type="search" placeholder="Search published public events" value={query} maxLength={200} onChange={event => setQuery(event.target.value)} /></div>
      <div className="city-events-table-wrap" role="region" aria-label="Available public events" tabIndex={0} aria-busy={loading}><table className="city-events-table"><caption className="sr-only">Published public events available to add</caption><thead><tr><th scope="col">Event</th><th scope="col">Date</th><th scope="col">Location</th><th scope="col"><span className="sr-only">Action</span></th></tr></thead><tbody>
        {loading ? <tr><td colSpan={4} className="city-table-empty" role="status">Loading events...</td></tr> : error ? <tr><td colSpan={4} className="city-table-empty"><p role="alert">{error}</p><button type="button" className="city-table-action" onClick={() => setRetry(value => value + 1)}>Try again</button></td></tr> : choices.length ? choices.map(item => { const added = events.some(row => row.id === item.id); return <tr key={item.id}><th scope="row">{item.title}</th><td><EventDate value={item.starts_at} timezone={timezone} /></td><td>{item.location_name || 'Not listed'}</td><td><button type="button" className="city-table-action" disabled={busy || added} onClick={() => void onCurate(item, false)} aria-label={added ? `${item.title} is already added` : `Add ${item.title} to this City Hub`}>{added ? 'Added' : 'Add event'}</button></td></tr>; }) : <tr><td colSpan={4} className="city-table-empty">{query ? 'No matching public events.' : 'No public events available.'}</td></tr>}
      </tbody></table></div>
    </div>}
  </section>;
}
