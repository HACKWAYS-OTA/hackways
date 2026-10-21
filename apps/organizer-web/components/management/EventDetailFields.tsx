'use client';

import { Checkbox } from '@hackways/ui/checkbox';
import { RadioGroup } from '@hackways/ui/radio-group';
import Link from 'next/link';
import { useState } from 'react';
import type { EventFormat } from '@/types/api';
import type { ManagedEvent } from '@/lib/management';
import { DateTimeField } from '@/components/ui/DateTimeField';
import { Input } from '@/components/ui/Input';
import { Icon } from '@/components/ui/Icon';
import { EditorMenu } from '@/components/hosting/EditorMenu';
import { EventTextEditor } from './EventTextEditor';
import { EventTimeZoneField } from './EventTimeZoneField';
import { useEventEditSession } from './EventEditSession';

export function EventDetailFields({ event, format, onFormatChange, disabled }: { event: ManagedEvent; format: EventFormat; onFormatChange: (format: EventFormat) => void; disabled: boolean }) {
  const [description, setDescription] = useState(event.description);
  const [location, setLocation] = useState(event.location_name || ''), [address, setAddress] = useState(event.location_address || '');
  const [online, setOnline] = useState(event.online_url || ''), [visibility, setVisibility] = useState(event.visibility || 'public');
  const [zone, setZone] = useState(event.timezone), [venueInvalid, setVenueInvalid] = useState(false);
  const [faq, setFaq] = useState(event.faq_text || ''), [rewards, setRewards] = useState(event.rewards || '');
  const session = useEventEditSession();
  return <>
    <div className="event-edit-meta">
      <Link className="event-edit-host" href={`/events/${event.slug}/settings`}><Icon name="people" size={17} />{event.space_id ? 'Space hosting' : 'Personal event'}<Icon name="arrow" size={14} /></Link>
      <input type="hidden" name="visibility" value={visibility} />
      <EditorMenu disabled={disabled} label={<><Icon name={visibility === 'private' ? 'lock' : 'globe'} size={16} />{visibility === 'private' ? 'Private' : 'Public'}</>}>
        <button type="button" aria-pressed={visibility === 'public'} onClick={() => setVisibility('public')}><Icon name="globe" size={18} /><span>Public<small>Listed on Discover and your Space.</small></span>{visibility === 'public' && <Icon name="check" size={16} />}</button>
        <button type="button" aria-pressed={visibility === 'private'} onClick={() => setVisibility('private')}><Icon name="lock" size={18} /><span>Private<small>Guests can discover and register through your event link.</small></span>{visibility === 'private' && <Icon name="check" size={16} />}</button>
      </EditorMenu>
    </div>
    <div className="event-edit-title" id="event-details" tabIndex={-1}><label htmlFor="edit-event-name">Event name</label><input id="edit-event-name" name="title" defaultValue={event.title} required maxLength={200} autoComplete="off" /></div>
    <section className="event-edit-schedule" id="event-schedule" tabIndex={-1} aria-label="Event schedule">
      <div className="event-edit-dates">
        <div><span>Starts (UTC)</span><DateTimeField label="Start (UTC)" name="starts" defaultValue={new Date(event.starts_at).toISOString().slice(0, 19)} required disabled={disabled} /></div>
        <div><span>Ends (UTC)</span><DateTimeField label="End (UTC)" name="ends" defaultValue={new Date(event.ends_at).toISOString().slice(0, 19)} required disabled={disabled} /></div>
      </div>
      <EventTimeZoneField value={zone} onChange={setZone} zones={session?.timeZones || []} disabled={disabled} />
    </section>
    <details className="event-edit-disclosure" id="event-location">
      <summary><Icon name="location" size={19} /><span>Event location<small>{format === 'online' ? online || 'Online event' : location || 'Add a venue or joining link'}</small></span><Icon name="down" size={16} /></summary>
      <div className="event-edit-disclosure-body">
        <fieldset className="hw-choice-field"><legend>How will people join?</legend><RadioGroup label="How will people join?" name="event-format" options={[{ value: 'in_person', label: 'In person' }, { value: 'online', label: 'Online' }, { value: 'hybrid', label: 'Hybrid' }] as const} value={format} onValueChange={onFormatChange} disabled={disabled} /></fieldset>
        {format !== 'online' && <div className="event-venue-editor" data-invalid={venueInvalid || undefined}>
          <div className="event-venue-line"><label htmlFor="edit-event-venue">Venue</label><input id="edit-event-venue" name="location" value={location} onChange={e => { setLocation(e.target.value); setVenueInvalid(false); }} placeholder="Add a venue or meeting place" required maxLength={1000} autoComplete="off" aria-invalid={venueInvalid || undefined} aria-describedby={venueInvalid ? 'event-venue-error' : undefined} onInvalid={e => { e.preventDefault(); setVenueInvalid(true); e.currentTarget.focus(); }} /></div>
          <div className="event-venue-line"><label htmlFor="edit-event-address">Address (optional)</label><input id="edit-event-address" name="address" value={address} onChange={e => setAddress(e.target.value)} placeholder="Building, street and city" maxLength={2000} autoComplete="off" /></div>
          {venueInvalid && <p id="event-venue-error" className="event-control-error" role="alert">Add a venue for your in-person event.</p>}
        </div>}
        {format !== 'in_person' && <Input label="Online joining link" name="online" type="url" pattern="https://.*" value={online} onChange={e => setOnline(e.target.value)} maxLength={2048} required />}
      </div>
    </details>
    <details className="event-edit-disclosure">
      <summary><Icon name="calendar" size={19} /><span>Description<small>{description || 'What should your guests know?'}</small></span><Icon name="down" size={16} /></summary>
      <div className="event-edit-disclosure-body"><EventTextEditor label="About this event" name="description" value={description} onChange={setDescription} required maxLength={20000} disabled={disabled} hideLabel /></div>
    </details>
    <details className="event-edit-disclosure">
      <summary><Icon name="add" size={19} /><span>Additional details<small>FAQ, contact and rewards</small></span><Icon name="down" size={16} /></summary>
      <div className="event-edit-disclosure-body"><EventTextEditor label="Frequently asked questions" name="faq_text" value={faq} onChange={setFaq} maxLength={10000} disabled={disabled} /><Input label="Contact page (optional)" name="contact_url" type="url" pattern="https://.*" maxLength={2048} defaultValue={event.contact_url || ''} /><EventTextEditor label="Rewards or prizes" name="rewards" value={rewards} onChange={setRewards} maxLength={10000} disabled={disabled} /></div>
    </details>
    <Link className="event-edit-registration" href={`/events/${event.slug}/registration`}>Capacity, approval & registration<Icon name="arrow" size={16} /></Link>
    {event.status === 'published' && <label className="event-notify-control"><span><strong>Notify guests</strong><small>Send schedule and location updates to registered guests.</small></span><Checkbox name="ack" disabled={disabled} /></label>}
  </>;
}
