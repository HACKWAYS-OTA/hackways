import { SectionJumpButton } from '@hackways/ui/section-jump';
import Link from 'next/link';
import { managedEvent } from '@/lib/management';
import { EventSettings } from '@/components/management/EventSettings';
import { RegistrationTabs } from '@/components/management/RegistrationTabs';
import { Icon } from '@/components/ui/Icon';
export const metadata = { title: 'Event registration' };
export default async function Registration({ params }: { params: { id: string } }) {
  const event = await managedEvent(params.id);
  const editable = !event.management_closed;
  return <><RegistrationTabs slug={event.slug} /><div className="host-settings-layout host-registration-layout">
    <section><h2>Registration settings</h2>{editable ? <EventSettings key={event.revision} event={event} section="registration" /> : <p className="journey-notice">An administrator has closed editing for this event.</p>}</section>
    <aside className="host-registration-aside" aria-labelledby="registration-overview-heading">
      <h2 id="registration-overview-heading">Registration overview</h2>
      <dl><div><dt>Confirmed places</dt><dd><Link href={`/events/${event.slug}/guests?status=confirmed`} aria-label={`View ${event.current_rsvp_count} confirmed guests`}>{event.current_rsvp_count}{event.capacity ? ` / ${event.capacity}` : ' · no limit'}<Icon name="arrow" size={14} /></Link></dd></div><div><dt>Reserved offers</dt><dd><Link href={`/events/${event.slug}/guests?status=offered`} aria-label={`View ${event.held_rsvp_count} reserved offers`}>{event.held_rsvp_count}<Icon name="arrow" size={14} /></Link></dd></div></dl>
      <nav className="host-registration-shortcuts" aria-label="Registration actions">
        {editable && <><SectionJumpButton targetId="registration-visibility">Edit visibility<Icon name="arrow" size={15} /></SectionJumpButton><SectionJumpButton targetId="registration-places">Change capacity & joining<Icon name="arrow" size={15} /></SectionJumpButton><SectionJumpButton targetId="registration-window">Set registration times<Icon name="arrow" size={15} /></SectionJumpButton></>}
        <Link href={`/events/${event.slug}/registration/questions`}>{editable ? 'Edit registration questions' : 'View registration questions'}<Icon name="arrow" size={15} /></Link>
      </nav>

    </aside>
  </div></>;
}
