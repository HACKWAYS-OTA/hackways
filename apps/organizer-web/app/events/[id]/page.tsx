import Link from 'next/link';
import { managedEvent } from '@/lib/management';
import { privateApi } from '@/lib/account';
import { dateLabel } from '@/lib/journeys';
import type { GuestSummary, Activity } from '@/lib/dashboard';
import { EventArtwork } from '@/components/management/EventArtwork';
import { ActivityList } from '@/components/management/ActivityList';
import { ShareEvent } from '@/components/management/ShareEvent';
import { EventStatusControls } from '@/components/management/EventStatusControls';
import { Icon } from '@/components/ui/Icon';
import { runtimeFeatures } from '@/lib/runtime';
export const metadata = { title: 'Event overview' };
export default async function Overview({ params }: { params: { id: string } }) {
  const event = await managedEvent(params.id);
  const [summary, activity, runtime] = await Promise.allSettled([
    privateApi<{ data: GuestSummary }>(`/events/${event.id}/summary`),
    privateApi<{ data: Activity[] }>(`/events/${event.id}/audit?limit=5`),
    runtimeFeatures(),
  ]);
  const counts = summary.status === 'fulfilled' ? summary.value.data : null;
  const base = `/events/${event.slug}`;
  const expanded = runtime.status === 'fulfilled' && runtime.value['manage.overview.v2'];
  const metrics = counts ? [
    ['Going', counts.confirmed, `${base}/guests?status=confirmed`],
    ['Awaiting approval', counts.pending, `${base}/guests?status=pending_approval`],
    ['Waiting list', counts.waitlisted, `${base}/guests?status=waitlisted`],
    ['Checked in', counts.checked_in, `${base}/check-in`],
  ] as const : [];
  return <div className={`management-overview${expanded ? '' : ' management-overview--compact'}`}>
    {counts ? <nav className="management-metrics" aria-label="Attendance">
      {metrics.map(([label, count, href]) => <Link href={href} key={label}><span>{label}</span><strong>{count}</strong><Icon name="arrow" size={16} /></Link>)}
    </nav> : <p className="journey-error" role="alert">Attendance could not be loaded.</p>}
    <nav className="management-actions" aria-label="Event actions">
      <Link className="app-button app-button-primary" href={`${base}/invitations`}><Icon name="add" size={18} />Invite attendees</Link>
      <Link className="app-button app-button-outline" href={`${base}/guests`}><Icon name="people" size={18} />Manage guests</Link>
      <Link className="app-button app-button-outline" href={`${base}/check-in`}><Icon name="scan" size={18} />Scan passes</Link>
      <Link className="app-button app-button-outline" href={`${base}/registration/questions`}><Icon name="edit" size={18} />Edit registration</Link>
    </nav>
    {expanded && <div className="management-overview-grid">
      <section className="management-activity"><header className="host-section-heading"><h2>Recent activity</h2><Link className="quiet-link" href={`${base}/activity`}>View all activity<Icon name="arrow" size={16} /></Link></header>
        {activity.status === 'fulfilled' ? <ActivityList items={activity.value.data} zone={event.timezone} /> : <p className="journey-error">Activity could not be loaded.</p>}
      </section>
      <section className="management-event"><EventArtwork event={event} /><h2>Event details</h2>
        <dl><div><dt>Schedule</dt><dd>{dateLabel(event.starts_at, event.timezone)}<span>– {dateLabel(event.ends_at, event.timezone)}</span></dd></div><div><dt>Location</dt><dd>{event.format === 'online' ? 'Online event' : event.location_name || 'Venue to be announced'}</dd></div><div><dt>Capacity</dt><dd>{event.capacity == null ? 'No limit' : event.capacity}</dd></div><div><dt>Registration</dt><dd>{event.registration_mode === 'approval' ? 'Approval required' : 'Open registration'}</dd></div></dl>
        <Link className="app-button app-button-outline" href={`${base}/edit`}><Icon name="edit" size={16} />Edit event</Link>
      </section>
    </div>}
    <ShareEvent event={event} />
    {event.status !== 'published' && event.status !== 'cancelled' && <div className="management-publication"><EventStatusControls event={event} /></div>}
  </div>;
}
