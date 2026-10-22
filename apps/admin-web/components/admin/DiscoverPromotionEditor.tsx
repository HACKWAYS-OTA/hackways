'use client';

import { useEffect, useRef, useState, type FormEvent } from 'react';
import Link from 'next/link';
import { Checkbox } from '@hackways/ui/checkbox';
import { adminRequest, mutationHeaders } from '@/lib/admin-request';
import type { PromotionConfiguration, PromotionEvent, PromotionSelection } from '@/lib/discover-promotions';
import { AdminPending } from './AdminPending';

function localDate(value: string) { const date = new Date(value); return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16); }
function utcDate(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? '' : date.toISOString(); }

export function DiscoverPromotionEditor({ initial }: { initial: PromotionConfiguration }) {
  const [saved, setSaved] = useState(initial);
  const [selections, setSelections] = useState(initial.selections);
  const [query, setQuery] = useState('');
  const [offset, setOffset] = useState(0);
  const [choices, setChoices] = useState<PromotionEvent[]>([]);
  const [more, setMore] = useState(false);
  const [finding, setFinding] = useState(true);
  const [searchError, setSearchError] = useState('');
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [reason, setReason] = useState('');
  const [busy, setBusy] = useState(false);
  const receipt = useRef<{ signature: string; key: string }>();
  useEffect(() => {
    const controller = new AbortController(); setFinding(true); setSearchError('');
    const timer = setTimeout(async () => {
      try { const result = await adminRequest<{ data: PromotionEvent[]; has_next: boolean }>(`core/discover-promotions/events?q=${encodeURIComponent(query)}&offset=${offset}`, { signal: controller.signal }); if (!controller.signal.aborted) { setChoices(result.data); setMore(result.has_next); } }
      catch (failure) { if (!controller.signal.aborted) setSearchError(failure instanceof Error ? failure.message : 'Events could not be loaded.'); }
      finally { if (!controller.signal.aborted) setFinding(false); }
    }, 250);
    return () => { clearTimeout(timer); controller.abort(); };
  }, [query, offset]);
  const update = (reference: number, change: Partial<PromotionSelection>) => setSelections(rows => rows.map(row => row.reference === reference ? { ...row, ...change } : row));
  function move(index: number, delta: number) { setSelections(rows => { const next = [...rows]; [next[index], next[index + delta]] = [next[index + delta], next[index]]; return next; }); }
  function toggle(event: PromotionEvent) { setSelections(rows => rows.some(row => row.reference === event.reference) ? rows.filter(row => row.reference !== event.reference) : rows.length < 5 ? [...rows, { ...event, starts_at: new Date().toISOString(), ends_at: event.event_starts_at, eligible: true }] : rows); }
  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (busy) return;
    const payload = { revision: saved.revision, selections: selections.map(({ reference, starts_at, ends_at }) => ({ reference, starts_at, ends_at })) };
    const body = JSON.stringify(payload), signature = body + reason;
    if (receipt.current?.signature !== signature) receipt.current = { signature, key: crypto.randomUUID() };
    setBusy(true); setError(''); setMessage('');
    try { const result = await adminRequest<{ data: PromotionConfiguration }>('core/discover-promotions', { method: 'PUT', headers: mutationHeaders(reason, receipt.current.key), body }); setSaved(result.data); setSelections(result.data.selections); setMessage(result.data.selections.length ? 'Promotion schedule saved.' : 'Banner disabled. Discover will show the regular event list.'); receipt.current = undefined; }
    catch (failure) { setError(failure instanceof Error ? failure.message : 'The schedule could not be saved. Your selection is preserved.'); }
    finally { setBusy(false); }
  }
  return <div className="promotion-editor">
    <section aria-labelledby="promotion-choices"><h2 id="promotion-choices">Choose events</h2><p className="admin-secondary-copy">Published, public, upcoming events only. Choose up to five.</p><label className="field promotion-search">Search events<input className="app-input" type="search" value={query} maxLength={200} onChange={event => { setQuery(event.target.value); setOffset(0); }} /></label>
      <div aria-busy={finding} className="promotion-choices">{finding ? <AdminPending /> : searchError ? <p role="alert">{searchError}</p> : !choices.length ? <p>No eligible events found.</p> : choices.map(event => <label key={event.reference} className="promotion-choice"><Checkbox checked={selections.some(row => row.reference === event.reference)} disabled={busy || selections.length === 5 && !selections.some(row => row.reference === event.reference)} onCheckedChange={() => toggle(event)} /><span><strong>{event.title}</strong><small>{new Date(event.event_starts_at).toLocaleString()} · {event.timezone}</small></span></label>)}</div>
      <nav className="admin-pagination" aria-label="Event choices">{offset > 0 && <button className="app-button app-button-outline" disabled={finding} onClick={() => setOffset(value => Math.max(0, value - 20))}>Previous</button>}{more && <button className="app-button app-button-outline" disabled={finding} onClick={() => setOffset(value => value + 20)}>Next</button>}</nav>
    </section>
    <form onSubmit={save} aria-labelledby="promotion-schedule"><h2 id="promotion-schedule">Schedule <span className="admin-secondary-copy">{selections.length}/5</span></h2><p className="admin-secondary-copy">Times use your device’s local time. Active events rotate every 7 seconds.</p><fieldset disabled={busy} className="promotion-schedule">
      {!selections.length && <p className="promotion-empty">No promotions selected. Save this selection to hide the banner.</p>}
      {selections.map((row, index) => <section key={row.reference} className="promotion-slot" aria-label={`Promotion ${index + 1}: ${row.title}`}><div className="promotion-slot-heading"><Link className="admin-record-link" href={`/events/details/${row.reference}`}>{row.title}</Link><div className="campaign-actions"><button type="button" className="app-button app-button-outline app-button-sm" disabled={index === 0} aria-label={`Move ${row.title} earlier`} onClick={() => move(index, -1)}>Up</button><button type="button" className="app-button app-button-outline app-button-sm" disabled={index === selections.length - 1} aria-label={`Move ${row.title} later`} onClick={() => move(index, 1)}>Down</button><button type="button" className="app-button app-button-outline app-button-sm" aria-label={`Remove ${row.title}`} onClick={() => setSelections(rows => rows.filter(item => item.reference !== row.reference))}>Remove</button></div></div>{row.eligible === false && <p className="field-error">This selection has expired or is no longer eligible. Remove it before saving.</p>}<div className="campaign-fields"><label className="field">Show from<input className="app-input" type="datetime-local" required value={row.starts_at ? localDate(row.starts_at) : ''} max={localDate(row.event_starts_at)} onChange={event => update(row.reference, { starts_at: utcDate(event.target.value) })} /></label><label className="field">Show until<input className="app-input" type="datetime-local" required value={row.ends_at ? localDate(row.ends_at) : ''} min={row.starts_at ? localDate(row.starts_at) : undefined} max={localDate(row.event_starts_at)} onChange={event => update(row.reference, { ends_at: utcDate(event.target.value) })} /></label></div></section>)}
      <label className="field">Reason for this change<input className="app-input" value={reason} onChange={event => setReason(event.target.value)} minLength={3} maxLength={500} required /></label>{error && <p className="field-error" role="alert">{error}</p>}{message && <p className="admin-action-success" role="status">{message}</p>}<div className="campaign-actions"><button className="app-button app-button-primary" disabled={busy}>{busy ? 'Saving…' : 'Save schedule'}</button><button type="button" className="app-button app-button-outline" onClick={() => { setSelections(saved.selections); setMessage(''); }}>Discard changes</button></div>
    </fieldset></form>
  </div>;
}
