import type { Event } from '@/types/api';
import { eventDate, formatLabels } from '@/lib/events';

/** A branded event poster, composed from the actual published event details. */
export function EventPoster({ event }: { event: Event }) {
  return <div className={`event-poster event-poster--${event.format}`} aria-hidden="true">
    <div className="event-poster-top"><span>{eventDate(event, { month: 'short', day: 'numeric', year: 'numeric' })}</span><span className="event-poster-brand">hackways.</span></div>
    <span className="event-poster-title">{event.title}</span>
    <svg className="event-poster-art" viewBox="0 0 320 180" fill="none" focusable="false">
      <path d="M72 180V110a88 88 0 0 1 176 0v70" stroke="currentColor" strokeWidth="22" />
      <path d="M116 180v-70a44 44 0 0 1 88 0v70" stroke="currentColor" strokeWidth="22" />
      <path d="M160 180v-70" stroke="currentColor" strokeWidth="22" />
    </svg>
    <span className="event-poster-format">{formatLabels[event.format]}</span>
  </div>;
}
