import { externalUrl } from '@/lib/events';
import type { EventDetails } from '@/types/api';
import { Icon } from '@/components/ui/Icon';

export function EventInformation({ event }: { event: EventDetails }) {
  const contact = externalUrl(event.contact_url) || externalUrl(event.organizer_website);
  if (!event.faq_text && !event.rewards && !contact) return null;
  return <div className="event-additional-details">
    {event.faq_text && <section className="event-story-section" aria-labelledby="event-faq"><h2 id="event-faq">Before you come</h2><div className="event-prose"><p>{event.faq_text}</p></div></section>}
    {event.rewards && <section className="event-story-section" aria-labelledby="event-rewards"><h2 id="event-rewards">Rewards & prizes</h2><div className="event-prose"><p>{event.rewards}</p></div></section>}
    {contact && <section className="event-contact-line"><div><h2>A question for the host?</h2><p>{event.organizer_name}</p></div><a className="event-visit-link" href={contact} target="_blank" rel="noopener noreferrer">Contact the host<Icon name="external" size={16} /></a></section>}
  </div>;
}