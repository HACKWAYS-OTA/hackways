import Link from 'next/link';
import type { WorkspaceEvent } from '@/types/api';
import { managedSpace } from '@/lib/workspace';
import { privateApi, pageOffset } from '@/lib/account';
import { dateLabel } from '@/lib/journeys';
import { eventStatus } from '@/lib/dashboard';
import { site } from '@/lib/site';

export const metadata = { title: 'Manage Space' };
export default async function SpacePage({ params, searchParams }: { params: { slug: string }; searchParams: { offset?: string } }) {
  const { data: space, access } = await managedSpace(params.slug), offset = pageOffset(searchParams.offset);
  let events: WorkspaceEvent[] = [], error = '';
  try { events = (await privateApi<{ data: WorkspaceEvent[] }>(`/workspace/events?space=${encodeURIComponent(space.slug)}&limit=21&offset=${offset}`)).data; }
  catch (failure) { error = failure instanceof Error ? failure.message : 'Events could not be loaded.'; }
  return <><div className="workspace-space-actions">{access.permissions.includes('create') && space.status === 'active' && <a className="app-button app-button-primary" href={`${site.app}/create?space=${encodeURIComponent(space.slug)}`}>Create event</a>}{access.permissions.includes('content') && <a className="app-button app-button-outline" href={`${site.app}/spaces/${space.slug}/settings`}>Edit Space</a>}</div>
    {error ? <p className="journey-error" role="alert">{error} <Link href={`/spaces/${space.slug}`}>Try again</Link></p> : <div className="workspace-table-scroll" role="region" aria-label="Space events" tabIndex={0}><table className="workspace-table workspace-event-table"><caption className="visually-hidden">Events in {space.name}</caption><thead><tr><th scope="col">Event</th><th scope="col">Date</th><th scope="col">Status</th><th scope="col">Going</th></tr></thead><tbody>{events.length ? events.slice(0, 20).map(event => <tr key={event.id}><th scope="row"><Link href={`/events/${event.slug}`}>{event.title}</Link></th><td>{dateLabel(event.starts_at, event.timezone)}</td><td>{eventStatus(event)}</td><td>{event.current_rsvp_count}</td></tr>) : <tr><td colSpan={4} className="workspace-table-empty">No events in this Space yet.</td></tr>}</tbody></table></div>}
    {!error && (offset > 0 || events.length > 20) && <nav className="journey-pagination" aria-label="Space event pages">{offset > 0 && <Link className="app-button app-button-outline" href={`/spaces/${space.slug}?offset=${Math.max(0, offset - 20)}`}>Previous</Link>}{events.length > 20 && <Link className="app-button app-button-outline" href={`/spaces/${space.slug}?offset=${offset + 20}`}>Next</Link>}</nav>}
  </>;
}
