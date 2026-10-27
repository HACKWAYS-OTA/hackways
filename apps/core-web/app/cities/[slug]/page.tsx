import { managedMetadata } from '@hackways/ui/seo';
import Image from 'next/image';
import { SectionJumpButton } from '@hackways/ui/section-jump';
import Link from 'next/link';
import { readCity } from '@/lib/cities';
import { fetchApi } from '@/lib/api';
import { pageOffset } from '@/lib/account';
import type { Event } from '@/types/api';
import { EventCard } from '@/components/events/EventCard';
import { Icon } from '@/components/ui/Icon';
import '@/components/cities/cities.css';
export async function generateMetadata({ params }: { params: { slug: string } }) {
  const city = await readCity(params.slug), title = `Hackathons & Tech Events in ${city.name}`, description = `Discover public hackathons, tech meetups and community events in ${city.name}. Explore published gatherings and connect through Hackways.`, path = `/cities/${city.slug}`;
  return managedMetadata(path, { title, description, alternates: { canonical: path }, openGraph: { title, description, url: path, ...(city.banner_url ? { images: [{ url: city.banner_url, alt: city.name }] } : {}) }, twitter: { card: 'summary_large_image' as const, title, description, ...(city.banner_url ? { images: [city.banner_url] } : {}) } });
}
export default async function CityPage({ params, searchParams }: { params: { slug: string }; searchParams: { period?: string; offset?: string } }) {
  const city = await readCity(params.slug), past = searchParams.period === 'past', offset = pageOffset(searchParams.offset), path = `/cities/${city.slug}`;
  let events: Event[] = [], error = '';
  try { events = (await fetchApi<{ data: Event[] }>(`/cities/${city.id}/events?period=${past ? 'past' : 'upcoming'}&limit=20&offset=${offset}`, { cache: 'no-store' })).data; }
  catch (failure) { error = failure instanceof Error ? failure.message : 'Events could not be loaded.'; }
  const summary = city.description.trim().split(/\n\s*\n/)[0];
  const intro = summary.length <= 280 ? summary : summary.slice(0, 280).replace(/\s+\S*$/, '') + '...';
  return <div className="city-profile-page">
    <header className={`city-profile-hero${city.banner_url ? ' city-profile-hero-image' : ''}`}>
      {city.banner_url && <Image className="city-profile-photo" unoptimized src={city.banner_url} alt="" fill sizes="100vw" priority />}
      <div className="city-profile-title"><div className="city-profile-copy"><h1>{city.name}</h1>{intro && <p>{intro}</p>}<SectionJumpButton className="app-button city-hero-action" targetId="city-events">Explore events<Icon name="arrow" size={18} /></SectionJumpButton></div></div>
    </header>
    <div className="city-profile-body"><section className="city-events" id="city-events" aria-labelledby="city-events-title">
      <div className="city-events-heading"><h2 id="city-events-title">Events</h2><Link className="quiet-link" href="/create">Create event<Icon name="add" size={16} /></Link></div>
      <nav className="desk-tabs" aria-label="City events"><Link href={path} aria-current={!past ? 'page' : undefined}>Upcoming</Link><Link href={`${path}?period=past`} aria-current={past ? 'page' : undefined}>Past events</Link></nav>
      {error ? <div className="desk-empty"><p role="alert">{error}</p><Link className="quiet-link" href={`${path}?period=${past ? 'past' : 'upcoming'}`}>Try again</Link></div> : events.length ? <div className="city-event-list">{events.map(event => <EventCard event={event} key={event.id} />)}</div> : <div className="city-events-empty"><p>{past ? 'No past events yet.' : 'No upcoming events yet.'}</p><Link className="quiet-link" href="/discover">Explore all events<Icon name="arrow" size={16} /></Link></div>}
      {!error && (offset > 0 || events.length === 20) && <nav className="journey-pagination" aria-label="More city events">{offset > 0 && <Link href={`${path}?period=${past ? 'past' : 'upcoming'}&offset=${Math.max(0, offset - 20)}`}>Previous</Link>}{events.length === 20 && <Link href={`${path}?period=${past ? 'past' : 'upcoming'}&offset=${offset + 20}`}>Next</Link>}</nav>}
    </section>{city.description.length > intro.length && <aside className="city-about"><h2>About {city.name}</h2><p>{city.description}</p></aside>}</div>
  </div>;
}
