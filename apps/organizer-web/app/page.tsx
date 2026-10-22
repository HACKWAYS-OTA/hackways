import Link from 'next/link';
import type { Event } from '@/types/api';
import { requireUser, privateApi, pageOffset } from '@/lib/account';
import { ApiClientError } from '@/lib/api';
import { dateLabel } from '@/lib/journeys';
import { site } from '@/lib/site';
import { eventStatus, hostViews, queryPath } from '@/lib/dashboard';
import { EventArtwork } from '@/components/management/EventArtwork';
import { Icon } from '@/components/ui/Icon';
export default async function ManageEvents({ searchParams }: { searchParams: { offset?: string; view?: string; q?: string } }) {
  const dashboardPath = '/';
  await requireUser(dashboardPath);
  const offset = pageOffset(searchParams.offset);
  const view = hostViews.find(([key]) => key === searchParams.view)?.[0] || 'all';
  const q = typeof searchParams.q === 'string' ? searchParams.q.slice(0, 200).trim() : '';
  let events: Event[] = [], error = '';
  try { events = (await privateApi<{ data: Event[] }>(queryPath('/events/organizer', { limit: 21, offset, view, q }))).data; }
  catch (err) { if (!(err instanceof ApiClientError && err.code === 'FORBIDDEN')) error = err instanceof Error ? err.message : 'Events could not be loaded.'; }
  return <div className="host-home"><header className="host-home-heading"><div><h1>Your events</h1><p>Good gatherings start with a little care.</p></div><a className="app-button app-button-primary" href={`${site.app}/create`}><Icon name="add" size={17} />Create event</a></header>
    <div className="host-list-toolbar"><nav className="host-list-views" aria-label="Hosted event views">{hostViews.map(([key, label]) => <Link key={key} href={queryPath(dashboardPath, { view: key, q })} aria-current={view === key ? 'page' : undefined}>{label}</Link>)}</nav><form key={`${view}-${q}`} className="host-search" role="search"><input type="hidden" name="view" value={view} /><Icon name="search" size={18} /><input aria-label="Search your events" type="search" name="q" defaultValue={q} maxLength={200} placeholder="Find an event" /><button type="submit">Search</button></form></div>
    {error ? <section className="host-empty"><h2>Your events couldn’t be loaded.</h2><p role="alert">{error}</p><Link className="app-button app-button-outline" href={queryPath(dashboardPath, { view, q, offset })}>Try again</Link></section> : events.length === 0 ? <section className="host-empty"><Icon name="calendar" size={32} /><h2>{q ? 'No events match your search.' : view === 'all' && !offset ? 'Your next gathering starts here.' : 'No events in this view.'}</h2><p>{q ? 'Try another event name or clear the search.' : 'Create on Hackways, then come here to welcome guests and take care of the details.'}</p>{q || view !== 'all' || offset ? <Link className="app-button app-button-outline" href={dashboardPath}>Show all events</Link> : <a className="app-button app-button-primary" href={`${site.app}/create`}>Create your first event</a>}</section> : <div className="host-events">{events.slice(0, 20).map(event => <article className="host-event-row" key={event.slug}><div className="host-date"><strong>{dateLabel(event.starts_at, event.timezone, { day: 'numeric' })}</strong><span>{dateLabel(event.starts_at, event.timezone, { month: 'short', year: 'numeric' })}</span></div><Link className="host-event-art" href={`/events/${event.slug}`} tabIndex={-1} aria-hidden="true"><EventArtwork event={event} compact /></Link><div className="host-event-copy">{event.status !== 'published' && <p className="host-event-state">{eventStatus(event)}</p>}<h2><Link href={`/events/${event.slug}`}>{event.title}</Link></h2><p>{dateLabel(event.starts_at, event.timezone, { weekday: 'short', hour: 'numeric', minute: '2-digit' })} · {event.format === 'online' ? 'Online' : event.location_name || 'Venue to be added'}</p><span>{event.current_rsvp_count} going{event.capacity != null ? ` · ${event.capacity} places` : ''}</span></div><Link className="host-event-open" href={`/events/${event.slug}`}>Manage event<Icon name="arrow" size={17} /></Link></article>)}</div>}
    {!error && (offset > 0 || events.length > 20) && <nav className="journey-pagination" aria-label="More hosted events">{offset > 0 && <Link className="app-button app-button-outline" href={queryPath(dashboardPath, { view, q, offset: Math.max(0, offset - 20) })}>Previous</Link>}{events.length > 20 && <Link className="app-button app-button-outline" href={queryPath(dashboardPath, { view, q, offset: offset + 20 })}>Next events</Link>}</nav>}
  </div>;
}
