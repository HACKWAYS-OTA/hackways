import Link from 'next/link';
import type { Event } from '@/types/api';
import { currentUser, privateApi, pageOffset } from '@/lib/account';
import { ApiClientError } from '@/lib/api';
import { dateLabel, registrationLabels, type MyRegistration } from '@/lib/journeys';
import { site } from '@/lib/site';
import { Icon } from '@/components/ui/Icon';
export const metadata = { title: 'My events', robots: { index: false, follow: false } };

export default async function MyEvents({ searchParams }: { searchParams: { view?: string; offset?: string } }) {
  const user = await currentUser();
  const hosting = searchParams.view === 'hosting';
  const offset = pageOffset(searchParams.offset);
  let hosted: Event[] = [], attending: MyRegistration[] = [], error = '';
  if (user) {
    try {
      if (hosting) hosted = (await privateApi<{ data: Event[] }>(`/events/organizer?limit=20&offset=${offset}`)).data;
      else attending = (await privateApi<{ data: MyRegistration[] }>(`/rsvps?limit=20&offset=${offset}`)).data;
    } catch (err) {
      if (!(hosting && err instanceof ApiClientError && err.code === 'FORBIDDEN')) error = err instanceof Error ? err.message : 'Your events could not be loaded.';
    }
  }
  const count = hosting ? hosted.length : attending.length;
  const base = `/my-events?view=${hosting ? 'hosting' : 'attending'}`;
  return <div className="desk-page">
    <div className="desk-heading"><div><h1>My events</h1><p>Your plans. Your people. All in one place.</p></div><Link className="app-button app-button-primary" href="/create"><Icon name="add" size={17} />Create an event</Link></div>
    <nav className="desk-tabs" aria-label="Your events"><Link href="/my-events" aria-current={!hosting ? 'page' : undefined}>I’m attending</Link><Link href="/my-events?view=hosting" aria-current={hosting ? 'page' : undefined}>I’m hosting</Link></nav>
    {!user ? <section className="desk-empty"><Icon name="calendar" size={32} /><h2>A little space for your plans.</h2><p>Sign in to see your registrations and the events you host. You can still explore events without an account.</p><div className="journey-actions"><Link className="app-button app-button-primary" href={`/login?next=${encodeURIComponent(base)}`}>Sign in</Link><Link className="quiet-link" href="/discover">Explore events <Icon name="arrow" size={16} /></Link></div></section>
    : error ? <section className="desk-empty"><h2>We couldn’t load your events.</h2><p role="alert">{error}</p><Link href={base} className="app-button app-button-outline">Try again</Link></section>
    : count === 0 ? <section className="desk-empty"><Icon name={hosting ? 'add' : 'calendar'} size={32} /><h2>{offset ? 'You’ve reached the end.' : hosting ? 'Your next gathering starts here.' : 'Make room for something good.'}</h2><p>{hosting ? 'Create an event, share the details, and welcome your community. Drafts stay private until you publish.' : 'Find a gathering you’re curious about. Your registrations will appear here when you use your account.'}</p><Link className="app-button app-button-primary" href={hosting ? '/create' : '/discover'}>{hosting ? 'Create your first event' : 'Discover events'}<Icon name="arrow" size={16} /></Link></section>
    : <div className="personal-event-list">{hosting ? hosted.map(event => <article className="personal-event" key={event.id}><DateStamp start={event.starts_at} zone={event.timezone} /><div className="personal-event-copy">{event.status !== 'published' && <p>{event.status === 'unpublished' ? 'Registration paused' : event.status === 'cancelled' ? 'Cancelled' : 'Draft'}</p>}<h2>{event.title}</h2><p>{dateLabel(event.starts_at, event.timezone)} · {event.timezone}</p><p>{event.current_rsvp_count} going{event.capacity != null ? ` · ${event.capacity} places` : ''}</p></div><a className="app-button app-button-outline" href={`${site.organizer}/events/${event.slug}`}>Manage event<Icon name="arrow" size={16} /></a></article>) : attending.map(reg => <article className="personal-event" key={reg.id}><DateStamp start={reg.starts_at} zone={reg.timezone} /><div className="personal-event-copy"><span className={`journey-status status-${reg.status}`}>{reg.event_status === 'cancelled' ? 'Event cancelled' : registrationLabels[reg.status]}</span><h2><Link href={`/my-events/${reg.public_code}`}>{reg.title}</Link></h2><p>{dateLabel(reg.starts_at, reg.timezone)} · {reg.timezone}</p>{reg.status === 'offered' && reg.offer_expires_at && <p>Accept before {dateLabel(reg.offer_expires_at, reg.timezone)}</p>}</div><Link href={`/my-events/${reg.public_code}`} className="app-button app-button-outline">View registration<Icon name="arrow" size={16} /></Link></article>)}</div>}
    {user && !error && (offset > 0 || count === 20) && <nav className="journey-pagination" aria-label="More events">{offset > 0 && <Link href={`${base}&offset=${Math.max(0, offset - 20)}`} className="app-button app-button-outline">Previous</Link>}{count === 20 && <Link href={`${base}&offset=${offset + 20}`} className="app-button app-button-outline">Next</Link>}</nav>}
    <aside className="desk-note"><p>{hosting ? 'Everything you need to host. One event at a time.' : 'Looking for your guest registration? Use the same email when you create an account, then visit Account to connect it.'}</p><Link className="quiet-link" href={hosting ? '/create' : '/account'}>{hosting ? 'Plan a gathering' : 'Go to account'}<Icon name="arrow" size={16} /></Link></aside>
  </div>;
}
function DateStamp({ start, zone }: { start: string; zone: string }) {
  return <div className="event-date-stamp" aria-hidden="true"><span>{dateLabel(start, zone, { month: 'short' })}</span><strong>{dateLabel(start, zone, { day: 'numeric' })}</strong><span>{dateLabel(start, zone, { weekday: 'short' })}</span></div>;
}
