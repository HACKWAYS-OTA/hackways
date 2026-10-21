'use client';
import { Checkbox } from '@hackways/ui/checkbox';
import { RadioGroup } from '@hackways/ui/radio-group';
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import type { ManagedEvent } from '@/lib/management';
import type { EventFormat } from '@/types/api';
import { fetchApi } from '@/lib/api';
import { Input } from '@/components/ui/Input';
import { DateTimeField } from '@/components/ui/DateTimeField';
import { EventDetailFields } from './EventDetailFields';
import { useEventEditSession } from './EventEditSession';
const utcInput = (value: string | null) => value ? new Date(value).toISOString().slice(0, 19) : '';
export function EventSettings({ event, section = 'details', saveAppearance, onSaved, onError }: { event: ManagedEvent; section?: 'details' | 'registration'; saveAppearance?: () => Promise<void>; onSaved?: () => void; onError?: (message: string) => void }) {
  const router = useRouter(), lock = useRef(false), validationShown = useRef(false);
  const session = useEventEditSession();
  const [format, setFormat] = useState<EventFormat>(event.format), [mode, setMode] = useState(event.registration_mode), [busy, setBusy] = useState(false), [error, setError] = useState(''), [saved, setSaved] = useState(false), [waitlist, setWaitlist] = useState(event.waitlist_enabled);
  const details = section === 'details';
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); if (lock.current || (session && !session.begin('details'))) return;
    const base = session?.latest() || event;
    lock.current = true; setBusy(true); setError(''); setSaved(false);
    const data = new FormData(e.currentTarget);
    const value = (name: string, fallback: string | null | undefined) => data.has(name) ? String(data.get(name)).trim() : fallback || '';
    const date = (name: string, fallback: string | null) => !data.has(name) ? fallback : data.get(name) ? new Date(`${String(data.get(name))}Z`).toISOString() : null;
    let detailsCommitted = false;
    try {
      const result = await fetchApi<{ data: ManagedEvent }>(`/events/${base.id}`, { method: 'PUT', body: JSON.stringify({
        title: value('title', base.title), description: value('description', base.description), format,
        visibility: value('visibility', base.visibility || 'public'), faq_text: value('faq_text', base.faq_text), contact_url: value('contact_url', base.contact_url), rewards: value('rewards', base.rewards),
        starts_at: date('starts', base.starts_at), ends_at: date('ends', base.ends_at), timezone: value('timezone', base.timezone), capacity: details ? base.capacity : data.get('capacity') ? Number(data.get('capacity')) : null,
        location_name: details && format === 'online' ? '' : value('location', base.location_name), location_address: details && format === 'online' ? '' : value('address', base.location_address), online_url: details && format === 'in_person' ? '' : value('online', base.online_url),
        registration_opens_at: date('opens', base.registration_opens_at), registration_closes_at: date('closes', base.registration_closes_at), registration_mode: mode, allow_guest: details ? base.allow_guest : mode === 'instant' && data.get('guest') === 'on', waitlist_enabled: mode === 'instant' && waitlist,
        offer_minutes: details ? base.offer_minutes : Number(data.get('offer')), revision: base.revision, acknowledge_attendees: data.get('ack') === 'on',
      }) });
      session?.update(result.data); detailsCommitted = true;
      await saveAppearance?.();
      setSaved(true); onSaved?.(); router.refresh();
    } catch (err) { const message = err instanceof Error ? err.message : 'The changes could not be saved. Your entries are still here.'; const detail = detailsCommitted && saveAppearance ? `Event details were saved. Cover and theme could not be confirmed: ${message}` : message; if (onError) onError(detail); else setError(detail); }
    finally { lock.current = false; setBusy(false); session?.finish(); }
  }
  return <form id={details && session ? 'event-edit-details-form' : undefined} onSubmit={submit} className={`management-form host-edit-form${details ? ' event-edit-details' : ''}`} aria-busy={busy || !!session?.busyArea} onInput={() => setSaved(false)} onClickCapture={e => { if (e.target instanceof Element && e.target.closest('button[type="button"], [role="option"], summary')) setSaved(false); }} onInvalidCapture={e => { if (onError && (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement)) { e.preventDefault(); e.stopPropagation(); if (!validationShown.current) { validationShown.current = true; const label = e.target.labels?.[0]?.textContent?.trim() || e.target.name; onError(`${label ? `${label}: ` : ''}${e.target.validationMessage || 'Check this field before saving.'}`); queueMicrotask(() => { validationShown.current = false; }); } } if (e.target instanceof HTMLElement) { const disclosure = e.target.closest('details'); if (disclosure) disclosure.open = true; } }}><fieldset className="host-form-fields" disabled={busy || !!session?.busyArea}><legend className="visually-hidden">{details ? 'Event page details' : 'Registration settings'}</legend>
    {details ? <EventDetailFields event={event} format={format} onFormatChange={setFormat} disabled={busy || !!session?.busyArea} /> : <>
      <fieldset className="form-section" id="registration-visibility" tabIndex={-1}><legend>Event visibility</legend><RadioGroup label="Event visibility" name="visibility" layout="stack" defaultValue={event.visibility === 'private' ? 'private' : 'public'} onValueChange={() => setSaved(false)} disabled={busy} options={[{ value: 'public', label: 'Public', description: 'Listed on Discover and your Space.' }, { value: 'private', label: 'Private', description: 'Guests can find and register through your event link.' }]} /></fieldset>
      <fieldset className="form-section" id="registration-places" tabIndex={-1}><legend>Capacity & joining</legend><Input label="Number of places" name="capacity" type="number" min={1} max={2147483647} step={1} defaultValue={event.capacity ?? ''} helperText="Leave empty for no limit. Places already reserved cannot be removed." /><fieldset className="hw-choice-field"><legend>Registration</legend><RadioGroup label="Registration" name="registration-mode" value={mode} disabled={busy} options={[{ value: 'instant', label: 'Join directly' }, { value: 'approval', label: 'I approve applications' }] as const} onValueChange={key => { setSaved(false); setMode(key); if (key === 'approval') setWaitlist(false); }} /></fieldset><label className="check-label"><Checkbox name="guest" onCheckedChange={() => setSaved(false)} defaultChecked={event.allow_guest} disabled={busy || mode === 'approval'} />Let people register without an account (direct registration only)</label><label className="check-label"><Checkbox checked={waitlist} disabled={busy || mode === 'approval'} onCheckedChange={checked => { setSaved(false); setWaitlist(checked); }} />Offer a waiting list when the event fills up</label><Input label="Minutes to accept an offered place" type="number" name="offer" min={5} max={60} step={1} defaultValue={event.offer_minutes} required /></fieldset>
      <fieldset className="form-section" id="registration-window" tabIndex={-1}><legend>Registration times</legend><p className="field-help">Optional. Times below are in UTC. Registration still closes when the event starts.</p><div className="form-pair"><div className="field"><span>Opens (UTC)</span><DateTimeField label="Registration opens" name="opens" defaultValue={utcInput(event.registration_opens_at)} disabled={busy} /></div><div className="field"><span>Closes (UTC)</span><DateTimeField label="Registration closes" name="closes" defaultValue={utcInput(event.registration_closes_at)} disabled={busy} /></div></div></fieldset>
    </>}
  </fieldset>{error && !onError && <div className="journey-error" role="alert"><p>{error}</p><p>Your entries are still here. If the event changed elsewhere, reload the latest version before saving.</p><button type="button" className="quiet-link" disabled={busy || !!session?.busyArea} onClick={() => { if (session) void session.reload().catch(err => setError(err instanceof Error ? err.message : 'The event could not be reloaded.')); else router.refresh(); }}>{session ? 'Reload saved event (discard edits)' : 'Reload latest event'}</button></div>}{saved && !session && <p className="journey-notice" role="status">Your changes are saved.</p>}{!session && <footer className="host-save"><span>{details ? 'Changes update the event page.' : 'Settings apply to new registrations.'}</span><button className="app-button app-button-primary" disabled={busy}>{busy ? 'Saving…' : 'Save changes'}</button></footer>}</form>;
}
