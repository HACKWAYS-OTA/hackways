import Link from 'next/link';
import { readSpace } from '@/lib/spaces';
import { currentUser, privateApi, pageOffset } from '@/lib/account';
import { fetchApi } from '@/lib/api';
import { externalUrl } from '@/lib/events';
import type { Event } from '@/types/api';
import { EventCard } from '@/components/events/EventCard';
import { SpaceMembership } from '@/components/spaces/SpaceMembership';
import { Icon } from '@/components/ui/Icon';
import '@/components/spaces/spaces.css';
import { entitySocialMetadata } from '@/lib/social-metadata';
export async function generateMetadata({ params }: { params: { slug: string } }) { const space = await readSpace(params.slug); const description = space.description.slice(0, 160); return { ...entitySocialMetadata(`/spaces/${encodeURIComponent(space.slug)}`, space.name, description), title: space.name, description }; }
export default async function SpacePage({ params, searchParams }: { params: { slug: string }; searchParams: { period?: string; offset?: string } }) {
  const [space, user] = await Promise.all([readSpace(params.slug), currentUser()]);
  const past = searchParams.period === 'past', offset = pageOffset(searchParams.offset), base = `/spaces/${space.slug}`;
  const [events, membership] = await Promise.all([
    fetchApi<{ data: Event[] }>(`/spaces/${space.id}/events?period=${past ? 'past' : 'upcoming'}&limit=20&offset=${offset}`, { cache: 'no-store' }),
    user ? privateApi<{ data: { owned: boolean; joined: boolean } }>(`/spaces/${space.id}/membership`) : Promise.resolve(null),
  ]);
  const website = externalUrl(space.website_url);
  return <div className="spaces-page"><Link className="quiet-link" href="/spaces">Back to Spaces</Link><header className="space-profile"><div className="space-profile-copy"><div className="space-profile-name"><span className="space-mark" aria-hidden="true">{space.name.trim().slice(0, 1)}</span><div><h1>{space.name}</h1><p>{space.member_count} {space.member_count === 1 ? 'member' : 'members'} · Hosted by {space.owner_name}{space.status === 'archived' ? ' · Archived' : ''}</p></div></div><p className="space-description">{space.description}</p>{website && <a className="quiet-link" href={website} target="_blank" rel="noopener noreferrer">Community website<Icon name="external" size={15} /></a>}</div><SpaceMembership space={space} membership={membership?.data || null} signedIn={!!user} /></header>{space.status === 'archived' && <p className="space-archive-notice">This Space is archived. Existing events are still available; new members and events are paused.</p>}<div className="space-body"><section className="space-events"><nav className="desk-tabs" aria-label="Space events"><Link href={base} aria-current={!past ? 'page' : undefined}>Upcoming events</Link><Link href={`${base}?period=past`} aria-current={past ? 'page' : undefined}>Past events</Link></nav>{events.data.length ? <div className="space-event-list">{events.data.map(event => <EventCard key={event.slug} event={event} />)}</div> : <div className="desk-empty"><h2>{past ? 'The story is just beginning.' : 'Something to look forward to.'}</h2><p>{past ? 'Past events will appear here after they finish.' : 'Upcoming events will appear here when the owner publishes them.'}</p>{membership?.data.owned && space.status === 'active' && <Link className="app-button app-button-outline" href={`/create?space=${space.slug}`}>Create the first event</Link>}</div>}{(offset > 0 || events.data.length === 20) && <nav className="journey-pagination" aria-label="More community events">{offset > 0 && <Link className="app-button app-button-outline" href={`${base}?period=${past ? 'past' : 'upcoming'}&offset=${Math.max(0, offset - 20)}`}>Previous</Link>}{events.data.length === 20 && <Link className="app-button app-button-outline" href={`${base}?period=${past ? 'past' : 'upcoming'}&offset=${offset + 20}`}>Next</Link>}</nav>}</section><aside className="space-about"><h2>A place to belong.</h2><p>Join the community and take part in its events. Each event has its own registration.</p>{space.guidelines && <><h2>Community guidelines</h2><p className="space-guidelines">{space.guidelines}</p></>}</aside></div></div>;
}
