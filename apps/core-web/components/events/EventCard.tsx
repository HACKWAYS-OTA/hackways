import Link from 'next/link';
import type { Event } from '@/types/api';
import { Icon } from '@/components/ui/Icon';
import { eventDate, eventZone, eventPath, formatLabels, registrationState } from '@/lib/events';
import { EventCover } from './EventCover';

export interface EventCardProps { event: Event; prominent?: boolean }
export function EventCard({ event, prominent = false }: EventCardProps) {
  const state = registrationState(event);
  const href = eventPath(event);
  return <article className={`catalog-event${prominent ? ' catalog-event--prominent' : ''}`}>
    <EventCover event={event} href={href} priority={prominent} sizes={prominent ? '(max-width: 640px) 100vw, 440px' : '(max-width: 640px) 100vw, 248px'} />
    <div className="catalog-event-content">
      <time className="catalog-event-date" dateTime={event.starts_at}>{eventDate(event, { weekday: 'short', month: 'long', day: 'numeric', year: 'numeric' })}</time>
      <p className="catalog-event-meta"><span>{formatLabels[event.format]}</span><span>{eventDate(event, { hour: 'numeric', minute: '2-digit' })} {eventZone(event)}</span></p>
      <h3><Link href={href} prefetch={false}>{event.title}</Link></h3>
      {(event.location_name || event.location_address || event.format === 'online') && <p className="catalog-event-location"><Icon name={event.format === 'online' ? 'globe' : 'location'} size={16} /><span>{event.location_name || event.location_address || 'Online event'}</span></p>}
      {(event.short_description || event.description) && <p className="catalog-event-description">{event.short_description || event.description}</p>}
      {state.disabled && <p className="catalog-event-closed">{state.label}</p>}
    </div>
    <Link className="catalog-event-open" href={href} prefetch={false} aria-label={`View ${event.title}`}><span>View event</span><Icon name="arrow" size={19} /></Link>
  </article>;
}
