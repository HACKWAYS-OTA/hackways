import { managedMetadata } from '@hackways/ui/seo';
import type { Metadata } from 'next';
import Link from 'next/link';
import { readEvent, readEventManagementAccess,readInvitationState } from '@/lib/event-details';
import { eventPath } from '@/lib/events';
import { EventCover } from '@/components/events/EventCover';
import { EventSidebar } from '@/components/events/EventSidebar';
import { EventInvitation } from '@/components/events/EventInvitation';
import { EventInformation } from '@/components/events/EventInformation';
import { EventScrollLayout } from '@/components/events/EventScrollLayout';
import { entitySocialMetadata, pageSocialMetadata } from '@/lib/social-metadata';
import { EventStructuredData } from '@/components/events/EventStructuredData';

export async function generateMetadata({ params }: { params: { slug: string } }): Promise<Metadata> {
  const event = await readEvent(params.slug);
  if (event.visibility !== 'public' || event.status !== 'published') return managedMetadata(eventPath(event), { ...pageSocialMetadata('events'), title: 'Event details', description: 'View this event on Hackways.', robots: { index: false, follow: false } });
  const description = (event.short_description || event.description).slice(0, 160);
  return managedMetadata(eventPath(event), { ...entitySocialMetadata(eventPath(event), event.title, description), title: event.title, description, alternates: { canonical: eventPath(event) } });
}

export default async function EventPage({ params }: { params: { slug: string } }) {
  const [event, managementAccess,invitation] = await Promise.all([readEvent(params.slug), readEventManagementAccess(params.slug),readInvitationState(params.slug)]);
  return <EventScrollLayout>
      <EventStructuredData event={event} />
      <header className="event-title-block"><div className="event-title-anchor"><h1 className={event.title.length > 85 ? 'event-title-long' : undefined}>{event.title}</h1></div>
        {event.short_description && <p className="event-intro-copy">{event.short_description}</p>}
        <div className="event-presented-by"><span className="event-host-initial" aria-hidden="true">{(event.space_name || event.organizer_name).slice(0, 1).toUpperCase()}</span><div>{!event.space_slug && <span>Hosted by</span>}{event.space_slug ? <Link href={`/spaces/${event.space_slug}`}>{event.space_name}</Link> : <strong>{event.organizer_name}</strong>}</div></div>
        <EventInvitation event={event} managementAccess={managementAccess} invited={invitation.invited} registrationCode={invitation.registration_code} />
      </header>
    <aside className="event-plan" aria-label="Event banner and visit details"><div className="event-plan-anchor"><EventCover event={event} priority className="event-hero-image" sizes="(max-width: 640px) calc(100vw - 40px), (max-width: 860px) calc(100vw - 48px), 43vw" /><EventSidebar event={event} /></div></aside>
    <div className="event-story"><section className="event-story-section" aria-labelledby="about-heading"><h2 id="about-heading">About this event</h2><div className="event-prose">{event.description.split(/\n\s*\n/).filter(paragraph => paragraph.trim()).map((paragraph, index) => <p key={index}>{paragraph}</p>)}</div></section><EventInformation event={event} /></div>
  </EventScrollLayout>;
}
