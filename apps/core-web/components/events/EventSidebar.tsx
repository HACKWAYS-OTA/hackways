import type { EventDetails } from '@/types/api';
import { EventActions } from './EventActions';
import { Icon } from '@/components/ui/Icon';
import { eventDate, eventEndDate, eventZone, externalUrl } from '@/lib/events';

export function EventSidebar({ event, showActions = true }: { event: EventDetails; showActions?: boolean }) {
  const location = [event.location_name, event.location_address].filter(Boolean).join(', ');
  const online = externalUrl(event.online_url);
  const sameDay = eventDate(event, { year: 'numeric', month: 'numeric', day: 'numeric' }) === eventEndDate(event, { year: 'numeric', month: 'numeric', day: 'numeric' });
  return <div className="event-visit-details">
    <section className="event-visit-section" aria-labelledby="event-when"><Icon name="calendar" size={22} /><div><h2 id="event-when">Date & time</h2><p className="event-visit-primary"><time dateTime={event.starts_at}>{eventDate(event, { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}</time></p><p><time dateTime={event.starts_at}>{eventDate(event, { hour: 'numeric', minute: '2-digit' })}</time> – <time dateTime={event.ends_at}>{eventEndDate(event, sameDay ? { hour: 'numeric', minute: '2-digit' } : { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })}</time></p><span className="event-visit-secondary">{eventZone(event)}</span></div></section>
    <section className="event-visit-section" aria-labelledby="event-where"><Icon name={event.format === 'online' ? 'globe' : 'location'} size={22} /><div><h2 id="event-where">{event.format === 'online' ? 'Join from anywhere' : 'The meeting place'}</h2>{event.format !== 'online' && <><p className="event-visit-primary">{event.location_name || 'Venue to be announced'}</p>{event.location_address && <p>{event.location_address}</p>}{location && <a className="event-visit-link" href={`https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(location)}`} target="_blank" rel="noopener noreferrer">Get directions<Icon name="external" size={15} /></a>}</>}{event.format === 'online' && <p className="event-visit-primary">Online event</p>}{online && <a className="event-visit-link" href={online} target="_blank" rel="noopener noreferrer">Open online event<Icon name="external" size={15} /></a>}{!online && event.format === 'online' && <p>The host has not added joining details yet.</p>}</div></section>
    {showActions && <EventActions event={event} />}
  </div>;
}
