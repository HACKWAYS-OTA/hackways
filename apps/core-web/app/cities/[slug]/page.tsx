import Image from 'next/image';
import Link from 'next/link';
import { readCity } from '@/lib/cities';
import { fetchApi } from '@/lib/api';
import { pageOffset } from '@/lib/account';
import type { Event } from '@/types/api';
import { EventCard } from '@/components/events/EventCard';
import { Icon } from '@/components/ui/Icon';
import '@/components/cities/cities.css';
export async function generateMetadata({ params }: { params: { slug: string } }) {
  const city = await readCity(params.slug), title = `Events in ${city.name}`, description = city.description.slice(0, 160), path = `/cities/${city.slug}`;
  return { title, description, alternates: { canonical: path }, openGraph: { title, description, url: path, ...(city.banner_url ? { images: [{ url: city.banner_url, alt: city.name }] } : {}) }, twitter: { card: 'summary_large_image' as const, title, description, ...(city.banner_url ? { images: [city.banner_url] } : {}) } };
}
export default async function CityPage({ params, searchParams }: { params: { slug: string }; searchParams: { period?: string; offset?: string } }) {
  const city = await readCity(params.slug), past = searchParams.period === 'past', offset = pageOffset(searchParams.offset), path = `/cities/${city.slug}`;
  let events: Event[] = [], error = '';
  try { events = (await fetchApi<{ data: Event[] }>(`/cities/${city.id}/events?period=${past ? 'past' : 'upcoming'}&limit=20&offset=${offset}`, { cache: 'no-store' })).data; }
  catch (failure) { error = failure instanceof Error ? failure.message : 'Events could not be loaded.'; }
  return <div className="cities-page city-profile-page"><Link className="quiet-link" href="/cities">All cities<Icon name="arrow" size={16} /></Link><header className={`city-profile-hero${city.banner_url ? ' city-profile-hero-image' : ''}`}><div className="city-profile-title"><p>{[city.region, city.country].filter(Boolean).join(', ')}</p><h1>{city.name}</h1><span><Icon name="location" size={16} />{city.timezone.replaceAll('_', ' ')}</span></div>{city.banner_url && <div className="city-profile-photo"><Image unoptimized src={city.banner_url} alt="" width={1200} height={700} priority /></div>}</header><div className="city-profile-body"><section className="city-events"><div className="city-events-heading"><h2>On the calendar</h2><Link className="quiet-link" href="/create">Create an event<Icon name="add" size={16} /></Link></div><nav className="desk-tabs" aria-label="City events"><Link href={path} aria-current={!past ? 'page' : undefined}>Upcoming</Link><Link href={`${path}?period=past`} aria-current={past ? 'page' : undefined}>Past events</Link></nav>{error ? <div className="desk-empty"><p role="alert">{error}</p><Link className="quiet-link" href={`${path}?period=${past ? 'past' : 'upcoming'}`}>Try again</Link></div> : events.length ? <div className="city-event-list">{events.map(event => <EventCard event={event} key={event.id} />)}</div> : <div className="desk-empty"><h3>{past ? 'A new chapter for this city.' : 'More gatherings are on the way.'}</h3><p>{past ? 'Past public events will appear here.' : 'Check back for the next local event.'}</p><Link className="quiet-link" href="/discover">Explore all events<Icon name="arrow" size={16} /></Link></div>}{!error && (offset > 0 || events.length === 20) && <nav className="journey-pagination" aria-label="More city events">{offset > 0 && <Link href={`${path}?period=${past ? 'past' : 'upcoming'}&offset=${Math.max(0, offset - 20)}`}>Previous</Link>}{events.length === 20 && <Link href={`${path}?period=${past ? 'past' : 'upcoming'}&offset=${offset + 20}`}>Next</Link>}</nav>}</section><aside className="city-about"><h2>Around {city.name}</h2><p>{city.description}</p><Link className="quiet-link" href="/spaces">Meet the communities<Icon name="arrow" size={16} /></Link></aside></div></div>;
}
