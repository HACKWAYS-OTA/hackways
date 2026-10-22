'use client';

import { RadioGroup } from '@hackways/ui/radio-group';
import { useEffect, useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import type { Event, EventFormat, Space } from '@/types/api';
import { fetchApi } from '@/lib/api';
import { site } from '@/lib/site';
import { Icon } from '@/components/ui/Icon';
import { uploadEventCover, type CoverAttempt } from '@/lib/cover-upload';
import { CoverPicker } from './CoverPicker';
import { EditorMenu } from './EditorMenu';
import { ThemePicker } from './ThemePicker';
import { eventFontFamily } from '@hackways/ui/event-fonts';
import { defaultAppearance } from '@/lib/appearance';
import { DateTimeField } from '@/components/ui/DateTimeField';
import { AmbientBackdrop } from '@/components/events/AmbientBackdrop';
import './create-event.css';

export function CreateEventForm({ organizer, name, spaces, initialSpace }: {
  organizer: { display_name: string } | null;
  name: string;
  spaces: Space[];
  initialSpace: string;
}) {
  const router = useRouter();
  const [host, setHost] = useState(organizer?.display_name || '');
  const [format, setFormat] = useState<EventFormat>('in_person');
  const [zone, setZone] = useState('');
  const [busy, setBusy] = useState('');
  const [error, setError] = useState('');
  const [created, setCreated] = useState<Event | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [visibility, setVisibility] = useState<'public' | 'private'>('public');
  const [approval, setApproval] = useState(false);
  const [description, setDescription] = useState('');
  const [location, setLocation] = useState('');
  const [online, setOnline] = useState('');
  const [capacity, setCapacity] = useState('');
  const [savedID, setSavedID] = useState('');
  const [spaceSlug, setSpaceSlug] = useState(initialSpace);
  const [appearance, setAppearance] = useState(defaultAppearance);
  const saved = useRef<{ event: Event; payload: string } | null>(null);
  const coverAttempt = useRef<CoverAttempt | null>(null);
  const submitting = useRef(false);
  const feedback = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setZone(Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC');
  }, []);
  useEffect(() => {
    const body = document.body;
    const theme = `theme-${appearance.theme}`;
    body.classList.add('event-theme', 'event-editor-theme', theme, 'event-display-' + appearance.theme_display);
    body.style.setProperty('--event-tint', appearance.theme_color);
    body.style.setProperty('--event-heading', eventFontFamily(appearance.theme_font));
    body.style.setProperty('--event-radius', appearance.theme_style === 'crisp' ? '8px' : '16px');
    return () => {
      body.classList.remove('event-theme', 'event-editor-theme', theme, 'event-display-' + appearance.theme_display);
      ['--event-tint', '--event-heading', '--event-radius'].forEach(key => body.style.removeProperty(key));
    };
  }, [appearance]);
  useEffect(() => {
    if (error) feedback.current?.focus({ preventScroll: true });
  }, [error]);

  async function create(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitting.current) return;
    const form = new FormData(e.currentTarget);
    const status = (e.nativeEvent as SubmitEvent).submitter?.getAttribute('value') === 'draft' ? 'draft' : 'published';
    submitting.current = true;
    setBusy(status);
    setError('');
    try {
      const starts = new Date(String(form.get('starts_at')));
      const ends = new Date(String(form.get('ends_at')));
      if (ends <= starts) throw new Error('Choose an end time after the start time.');
      if (starts.getTime() <= Date.now()) throw new Error('Choose a start time in the future.');
      if (!host) {
        const profile = await fetchApi<{ data: { display_name: string } }>('/events/organizer/profile', {
          method: 'POST', body: JSON.stringify({ display_name: String(form.get('host')).trim() }),
        });
        setHost(profile.data.display_name);
      }
      const fields = {
          visibility, title: String(form.get('title')).trim(),
          description: String(form.get('description')).trim(),
          format, starts_at: starts.toISOString(), ends_at: ends.toISOString(), timezone: zone,
          capacity: form.get('capacity') ? Number(form.get('capacity')) : null,
          location_name: String(form.get('location') || '').trim(),
          online_url: String(form.get('online') || '').trim(), registration_mode: approval ? 'approval' : 'instant', space_slug: spaceSlug, ...appearance,
      };
      const payload = JSON.stringify(fields);
      if (saved.current && saved.current.payload !== payload) throw new Error('Your event record is already saved. Open it from My events to change the details, or restore your entries here to retry the image upload.');
      if (!saved.current) {
        const result = await fetchApi<{ data: Event }>('/events', { method: 'POST', body: JSON.stringify({ ...fields, status: file ? 'draft' : status }) });
        saved.current = { event: result.data, payload };
        setSavedID(result.data.id);
      }
      let event = saved.current.event;
      if (file && event.status === 'draft') {
        setBusy('image');
        event = await uploadEventCover(file, event, coverAttempt.current, attempt => { coverAttempt.current = attempt; });
        saved.current.event = event;
      }
      if (status === 'published' && event.status === 'draft') {
        setBusy('published');
        try {
          event = (await fetchApi<{ data: Event }>(`/events/${event.id}/status`, { method: 'POST', body: JSON.stringify({ status: 'published', revision: event.revision, acknowledge_attendees: false }) })).data;
        } catch (error) {
          // Reconcile a possibly committed publish before telling the user to retry.
          const latest = await fetchApi<{ data: Event }>(`/events/${event.id}/manage`, { cache: 'no-store' });
          if (latest.data.status !== 'published') throw error;
          event = latest.data;
        }
        saved.current.event = event;
      }
      if (event.status === 'published') {
        router.replace(`/events/${encodeURIComponent(event.slug)}`);
        return;
      }
      setCreated(event);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'The event could not be saved. Your details are still here.');
    } finally {
      submitting.current = false;
      setBusy('');
    }
  }

  if (created) return <section className="compose-success" role="status">
    <AmbientBackdrop active={appearance.theme === 'ambient'} file={file} source={created.banner_url} />
    <Icon name="check" size={28} />
    <h1>{created.status === 'published' ? 'Your event is live.' : 'Your draft is saved.'}</h1>
    <h2>{created.title}</h2>
    <p>{created.status === 'published' ? 'You’re ready to bring people together. Share your event page and start inviting.' : 'Your event stays private until you publish it. Continue with the details whenever you’re ready.'}</p>
    <div className="journey-actions">
      {created.status === 'published' && <Link className="app-button app-button-primary" href={`/events/${encodeURIComponent(created.slug)}`}>View event <Icon name="arrow" size={16} /></Link>}
      <a className={`app-button app-button-${created.status === 'published' ? 'outline' : 'primary'}`} href={`${site.organizer}/events/${created.slug}`}>Manage event</a>
      <Link className="quiet-link" href="/my-events?view=hosting">My hosted events</Link>
    </div>
  </section>;

  return <>
    <AmbientBackdrop active={appearance.theme === 'ambient'} file={file} />
    <header className="compose-heading">
      <div><Link className="compose-back" href="/my-events?view=hosting">My hosted events</Link><h1>Create an event</h1></div>
      <button type="submit" form="create-event" value="draft" className="app-button app-button-outline" disabled={!!busy || !zone}>{busy === 'draft' ? 'Saving…' : 'Save draft'}</button>
    </header>
    <div className="compose-layout"><div className="compose-art"><CoverPicker file={file} onChange={selected => { setFile(selected); coverAttempt.current = null; }} disabled={!!busy} /><ThemePicker value={appearance} onChange={setAppearance} disabled={!!busy} /></div>
    <form id="create-event" onSubmit={create} className="compose-form" aria-busy={!!busy} onInvalidCapture={e => { if (e.target instanceof HTMLElement) { const details = e.target.closest('details'); if (details) details.open = true; } }}>
      <fieldset className="compose-fields" disabled={!!busy}>
        <legend className="visually-hidden">Event details</legend>
        <div className="compose-meta">
          {host ? <EditorMenu disabled={!!busy} label={<><Icon name="people" size={16} /><span>{spaces.find(space => space.slug === spaceSlug)?.name || host}</span></>}><p className="editor-menu-caption">Organize this event under</p><button type="button" aria-pressed={!spaceSlug} onClick={() => setSpaceSlug('')}><span>{host}<small>Personal event</small></span>{!spaceSlug && <Icon name="check" size={16} />}</button>{spaces.map(space => <button type="button" key={space.slug} aria-pressed={spaceSlug === space.slug} onClick={() => setSpaceSlug(space.slug)}><Icon name="people" size={17} /><span>{space.name}<small>Space</small></span>{spaceSlug === space.slug && <Icon name="check" size={16} />}</button>)}<Link href="/spaces/new" className="editor-menu-link">Create a Space</Link><Link href="/spaces?view=owned" className="editor-menu-link">All my Spaces</Link></EditorMenu> : <div className="compose-host"><label htmlFor="event-host">Hosted by</label><input id="event-host" name="host" defaultValue={name} required maxLength={100} autoComplete="organization" aria-describedby="host-help" /><span className="visually-hidden" id="host-help">Use your name or your community’s name. This will appear on the event page.</span></div>}
          <EditorMenu align="right" disabled={!!busy} label={<><Icon name={visibility === 'public' ? 'globe' : 'lock'} size={15} />{visibility === 'public' ? 'Public' : 'Private'}</>}><button type="button" aria-pressed={visibility === 'public'} onClick={() => setVisibility('public')}><Icon name="globe" size={18} /><span>Public<small>Publish on Discover so people can register.</small></span>{visibility === 'public' && <Icon name="check" size={16} />}</button><button type="button" aria-pressed={visibility === 'private'} onClick={() => setVisibility('private')}><Icon name="lock" size={18} /><span>Private<small>Unlisted. Share the link to invite people to view and register.</small></span>{visibility === 'private' && <Icon name="check" size={16} />}</button></EditorMenu>
        </div>

        <div className="compose-title">
          <label htmlFor="event-title">Event name</label>
          <input id="event-title" name="title" required maxLength={200} autoComplete="off" />
        </div>

        <div className="compose-schedule">
          <div className="compose-dates">
            <div className="compose-date"><span className="compose-date-label"><span className="date-marker" />Starts</span><DateTimeField name="starts_at" label="Start" required /></div>
            <div className="compose-date"><span className="compose-date-label"><span className="date-marker date-marker-end" />Ends</span><DateTimeField name="ends_at" label="End" required /></div>
          </div>
          <div className="compose-timezone" id="event-zone"><Icon name="globe" size={19} /><div><span>Time zone</span><p>{zone || 'Detecting…'}</p></div></div>
        </div>

        <details className="compose-place compose-disclosure">
          <summary><Icon name="location" size={19} /><span>{(format === 'online' ? online : location || online) || 'Add event location'}<small>{(format === 'online' ? online : location || online) ? 'Edit location and joining details' : 'A venue, an online link, or both'}</small></span><Icon name="down" size={15} /></summary>
          <div className="compose-disclosure-body">
          <div className="compose-section-top"><h2 id="where-heading"><Icon name="location" size={19} />Where is it happening?</h2><fieldset className="hw-choice-field"><legend className="visually-hidden">Event format</legend><RadioGroup label="Event format" name="format" options={[{ value: 'in_person', label: 'In person' }, { value: 'online', label: 'Online' }, { value: 'hybrid', label: 'Hybrid' }] as const} value={format} onValueChange={setFormat} disabled={!!busy} /></fieldset></div>
          <div className={`compose-location-fields${format === 'hybrid' ? ' is-hybrid' : ''}`}>
            {format !== 'online' && <div className="compose-line-field"><label htmlFor="event-location">Venue or location</label><input id="event-location" name="location" value={location} onChange={e => setLocation(e.target.value)} maxLength={1000} autoComplete="off" /></div>}
            {format !== 'in_person' && <div className="compose-line-field"><label htmlFor="event-online">Online joining link</label><input id="event-online" name="online" type="url" pattern="https://.*" value={online} onChange={e => setOnline(e.target.value)} maxLength={2048} aria-describedby="online-help" /><span className="compose-hint" id="online-help">Use a secure https:// link.</span></div>}
          </div>
          </div>
        </details>

        <details className="compose-description compose-disclosure"><summary><Icon name="calendar" size={18} /><span>{description ? 'Event description' : 'Add a description'}{description && <small>{description}</small>}</span><Icon name="down" size={15} /></summary><div className="compose-disclosure-body"><label htmlFor="event-description">About your event</label><textarea id="event-description" name="description" value={description} onChange={e => setDescription(e.target.value)} rows={4} required maxLength={20000} aria-describedby="description-help" /><p id="description-help">What’s happening, who’s it for, and what should people know?</p></div></details>

        <section className="compose-options" aria-labelledby="options-heading"><h2 id="options-heading">Event options</h2><div className="compose-option-surface"><div className="compose-option-row"><div><Icon name="check" size={18} /><span>Entry</span></div><span className="compose-option-value">Free</span></div><div className="compose-option-row"><label id="approval-label" htmlFor="event-approval"><Icon name="lock" size={18} />Require approval</label><button type="button" id="event-approval" className="compose-switch" role="switch" aria-checked={approval} aria-labelledby="approval-label" onClick={() => setApproval(!approval)}><span /></button></div>{approval && <p className="compose-approval-help">Guests sign in to apply. You approve applications in your event’s guest list.</p>}<details className="compose-capacity-details"><summary className="compose-option-row"><span><Icon name="people" size={18} />Capacity</span><span className="compose-option-value">{capacity || 'No limit'}<Icon name="down" size={14} /></span></summary><div className="compose-capacity"><label htmlFor="event-capacity">Number of places</label><input id="event-capacity" name="capacity" value={capacity} onChange={e => setCapacity(e.target.value)} type="number" min={1} max={2147483647} step={1} aria-describedby="capacity-help" /><span id="capacity-help">Leave empty for no limit</span></div></details></div></section>
      </fieldset>

      {error && <div className="journey-error compose-error" role="alert" ref={feedback} tabIndex={-1}><p>{error}</p><p>{savedID ? 'Your event record is saved. ' : 'If saving was interrupted, '}Check <Link href="/my-events?view=hosting">your hosted events</Link> before starting another event.</p>{savedID && <a href={`${site.organizer}/events/${saved.current?.event.slug}`}>Open saved event</a>}</div>}
      <footer className="compose-finish"><button type="submit" value="create" className="app-button app-button-primary" disabled={!!busy || !zone}>{busy === 'image' ? 'Uploading banner…' : busy ? 'Saving event…' : 'Create event'}{!busy && <Icon name="arrow" size={17} />}</button><p>{visibility === 'private' ? 'Only people with the event link can find it. Share the link to invite them.' : 'Your event will be published on Discover.'}</p></footer>
    </form>
    </div>
  </>;
}
