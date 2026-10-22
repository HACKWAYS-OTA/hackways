import Link from 'next/link';
import { managedEvent } from '@/lib/management';
import { privateApi, pageOffset } from '@/lib/account';
import type { Activity } from '@/lib/dashboard';
import { ActivityList } from '@/components/management/ActivityList';
export const metadata = { title: 'Event activity' };
export default async function EventActivity({ params, searchParams }: { params: { id: string }; searchParams: { offset?: string } }) {
  const event = await managedEvent(params.id), offset = pageOffset(searchParams.offset);
  const result = await privateApi<{ data: Activity[] }>(`/events/${event.id}/audit?limit=26&offset=${offset}`);
  return <section className="host-activity-page"><header className="host-section-heading"><div><h2>Activity</h2></div></header><ActivityList items={result.data.slice(0, 25)} zone={event.timezone} />{(offset > 0 || result.data.length > 25) && <nav className="journey-pagination" aria-label="More event history">{offset > 0 && <Link className="app-button app-button-outline" href={`?offset=${Math.max(0, offset - 25)}`}>Newer activity</Link>}{result.data.length > 25 && <Link className="app-button app-button-outline" href={`?offset=${offset + 25}`}>Older activity</Link>}</nav>}</section>;
}
