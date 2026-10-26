'use client';
import { useEffect, useRef, useState, type FormEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Select } from '@hackways/ui/select';
import { MediaField } from '@hackways/ui/media-field';
import { adminRequest, mutationHeaders } from '@/lib/admin-request';
import type { AdminCity, CityEvent } from '@/lib/cities';

type Intent = { upload_url: string; blob_key: string; expires_at: string };
const states = [{ value: 'draft', label: 'Draft (hidden)' }, { value: 'published', label: 'Published' }, { value: 'archived', label: 'Archived (hidden)' }] as const;
const emptyEvents: CityEvent[] = [];
export function CityEditor({ initial, events: initialEvents = emptyEvents }: { initial?: AdminCity; events?: CityEvent[] }) {
  const router = useRouter();
  const [saved, setSaved] = useState(initial), [status, setStatus] = useState<AdminCity['status']>(initial?.status || 'draft');
  const [file, setFile] = useState<File | null>(null), [removed, setRemoved] = useState(false), [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  const [events, setEvents] = useState(initialEvents), [query, setQuery] = useState(''), [choices, setChoices] = useState<CityEvent[]>([]), [loading, setLoading] = useState(false), [searchError, setSearchError] = useState('');
  const receipt = useRef<{ body: string; key: string }>();
  const fieldsSaved = useRef('');
  const upload = useRef<{ file: File; intent: Intent; uploaded: boolean }>();
  const eventReceipt = useRef<{ body: string; key: string }>();
  const savedReference = saved?.reference;
  async function reloadLatest() {
    if (!saved || busy) return; setBusy(true);
    try { const result = await adminRequest<{ data: AdminCity; events: CityEvent[] }>(`core/cities/${saved.reference}`, {}, true); setSaved(result.data); setEvents(result.events); fieldsSaved.current = ''; setError(''); setMessage('Latest saved version loaded. Your unsaved entries are preserved; review them before saving.'); }
    catch (failure) { setError(failure instanceof Error ? failure.message : 'The latest City Hub could not be loaded. Your entries are still here.'); }
    finally { setBusy(false); }
  }
  useEffect(() => { setEvents(initialEvents); }, [initialEvents]);
  useEffect(() => {
    if (!savedReference) return;
    const controller = new AbortController(); setLoading(true); setSearchError('');
    const timer = setTimeout(async () => { try { const result = await adminRequest<{ data: CityEvent[] }>(`core/cities/event-choices?q=${encodeURIComponent(query)}`, { signal: controller.signal }, true); if (!controller.signal.aborted) setChoices(result.data); } catch (failure) { if (!controller.signal.aborted) setSearchError(failure instanceof Error ? failure.message : 'Events could not be loaded.'); } finally { if (!controller.signal.aborted) setLoading(false); } }, 250);
    return () => { clearTimeout(timer); controller.abort(); };
  }, [query, savedReference]);
  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (busy) return; setBusy(true); setError(''); setMessage('');
    const form = new FormData(event.currentTarget), fields = { name: String(form.get('name')).trim(), region: String(form.get('region')).trim(), country: String(form.get('country')).trim(), timezone: String(form.get('timezone')).trim(), description: String(form.get('description')).trim(), status };
    const signature = JSON.stringify(fields); let city = saved;
    try {
      if (fieldsSaved.current !== signature || !city) {
        const body = JSON.stringify({ ...fields, ...(city ? { revision: city.revision } : { slug: String(form.get('slug')).trim() }) });
        if (receipt.current?.body !== body) receipt.current = { body, key: crypto.randomUUID() };
        city = (await adminRequest<{ data: AdminCity }>(city ? `core/cities/${city.reference}` : 'core/cities', { method: city ? 'PUT' : 'POST', headers: mutationHeaders(receipt.current.key), body }, true)).data;
        fieldsSaved.current = signature; setSaved(city);
      }
      if (file || removed && city.banner_url) {
        if (file) {
          if (!upload.current || upload.current.file !== file || !upload.current.uploaded && Date.parse(upload.current.intent.expires_at) <= Date.now()) {
            const intent = await adminRequest<Intent>('global/media/upload-intent', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ resource_id: city.id, category: 'city-banners', filename: file.name, mime_type: file.type, size: file.size }) }, true);
            upload.current = { file, intent, uploaded: false };
          }
          if (!upload.current.uploaded) { const controller = new AbortController(), timer = setTimeout(() => controller.abort(), 60000); try { const response = await fetch(upload.current.intent.upload_url, { method: 'PUT', credentials: 'omit', body: file, signal: controller.signal, headers: { 'Content-Type': file.type, 'x-ms-blob-type': 'BlockBlob' } }); if (!response.ok && response.status !== 409 && response.status !== 412) throw new Error('The City Hub is saved. Retry to finish uploading its banner.'); upload.current.uploaded = true; } finally { clearTimeout(timer); } }
        }
        const result = await adminRequest<{ data: Pick<AdminCity, 'id' | 'reference' | 'revision' | 'banner_url'> }>('global/media/city-banner/complete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ resource_id: city.id, revision: city.revision, blob_key: file ? upload.current?.intent.blob_key : '', remove: !file }) }, true);
        city = { ...city, ...result.data }; setSaved(city); setFile(null); setRemoved(false);
      }
      setMessage('City Hub saved.'); receipt.current = undefined;
      if (!initial) { router.replace(`/cities/${city.reference}`); router.refresh(); }
    } catch (failure) { setError(failure instanceof Error ? failure.message : 'The City Hub could not be saved. Your entries are still here.'); }
    finally { setBusy(false); }
  }
  async function curate(item: CityEvent, remove: boolean) {
    if (!saved || busy) return; setBusy(true); setError(''); setMessage('');
    const body = JSON.stringify({ event_id: item.id, remove, revision: saved.revision });
    if (eventReceipt.current?.body !== body) eventReceipt.current = { body, key: crypto.randomUUID() };
    try { const result = await adminRequest<{ data: AdminCity }>(`core/cities/${saved.reference}/events`, { method: 'POST', headers: mutationHeaders(eventReceipt.current.key), body }, true); setSaved(result.data); eventReceipt.current = undefined; setMessage(remove ? 'Event removed from this City Hub.' : 'Event added to this City Hub.'); router.refresh(); }
    catch (failure) { setError(failure instanceof Error ? failure.message : 'The event selection could not be saved.'); }
    finally { setBusy(false); }
  }
  return <div className="city-editor"><form onSubmit={save}><fieldset disabled={busy}><legend className="sr-only">City Hub details</legend><div className="city-editor-fields"><label className="field">City name<input className="app-input" name="name" defaultValue={initial?.name} required maxLength={120} /></label>{!initial && !saved && <label className="field">Page address<span className="city-address-prefix">hackways.in/cities/</span><input className="app-input" name="slug" required pattern="[a-z0-9]+(-[a-z0-9]+)*" maxLength={100} /><small>Lowercase letters, numbers and hyphens. The address stays the same after creation.</small></label>}<label className="field">Region (optional)<input className="app-input" name="region" defaultValue={initial?.region} maxLength={120} /></label><label className="field">Country<input className="app-input" name="country" defaultValue={initial?.country} required maxLength={120} /></label><label className="field">Time zone<input className="app-input" name="timezone" defaultValue={initial?.timezone || 'Asia/Kolkata'} required maxLength={100} /><small>Use an IANA time zone, such as Asia/Kolkata.</small></label><label className="field">Publication<Select label="City publication" value={status} options={states} onValueChange={setStatus} /></label><label className="field city-editor-description">About this city<textarea className="app-input" name="description" rows={6} defaultValue={initial?.description} maxLength={6000} required={status === 'published'} /></label></div><MediaField aspectRatio={16 / 9} label="City banner (optional)" file={file} initialUrl={saved?.banner_url} removed={removed} disabled={busy} onChange={(next, remove) => { setFile(next); setRemoved(remove); }} /><div className="campaign-actions"><button className="app-button app-button-primary" disabled={busy}>{busy ? 'Saving…' : saved ? 'Save City Hub' : 'Create City Hub'}</button>{saved && <Link href={`https://hackways.in/cities/${saved.slug}`} className="app-button app-button-outline" target="_blank">{saved.status === 'published' ? 'Open public page' : 'Public address (hidden)'}</Link>}</div></fieldset></form>{error && <p className="field-error" role="alert">{error} <button className="quiet-link" type="button" disabled={busy} onClick={() => void reloadLatest()}>Reload latest data</button></p>}{message && <p className="admin-action-success" role="status">{message}</p>}{saved && <section className="city-curation"><h2>City events</h2><p className="admin-secondary-copy">Curate published public events for {saved.name}. Removing an event here leaves the event itself unchanged.</p><div className="city-linked-events">{events.length ? events.map(item => <div key={item.id}><span><strong>{item.title}</strong><small>{new Date(item.starts_at).toLocaleDateString()} · {item.status}{item.visibility !== 'public' ? ' · Hidden from public feed' : ''}</small></span><button type="button" className="app-button app-button-outline app-button-sm" disabled={busy} onClick={() => void curate(item, true)}>Remove</button></div>) : <p>No events curated yet.</p>}</div><label className="field">Find a public event<input className="app-input" type="search" value={query} maxLength={200} onChange={event => setQuery(event.target.value)} /></label><div className="city-event-choices" aria-busy={loading}>{loading ? <p role="status">Loading events…</p> : searchError ? <p role="alert">{searchError}</p> : choices.length ? choices.map(item => { const added = events.some(row => row.id === item.id); return <div key={item.id}><span><strong>{item.title}</strong><small>{new Date(item.starts_at).toLocaleDateString()}</small></span><button type="button" className="app-button app-button-outline app-button-sm" disabled={busy || added} onClick={() => void curate(item, false)}>{added ? 'Added' : 'Add event'}</button></div>; }) : <p>No published public events match your search.</p>}</div></section>}</div>;
}
