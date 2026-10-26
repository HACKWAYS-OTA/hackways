import Image from 'next/image';
import Link from 'next/link';
import type { Metadata } from 'next';
import type { CityHub } from '@hackways/api-client';
import { fetchApi } from '@/lib/api';
import { pageOffset } from '@/lib/account';
import { Icon } from '@/components/ui/Icon';
import '@/components/cities/cities.css';
export const metadata: Metadata = { title: 'Cities', description: 'Explore local events and communities in cities on Hackways.', alternates: { canonical: '/cities' } };
export default async function Cities({ searchParams }: { searchParams: { offset?: string } }) {
  const offset = pageOffset(searchParams.offset);
  let cities: CityHub[] = [], error = '';
  try { cities = (await fetchApi<{ data: CityHub[] }>(`/cities?limit=20&offset=${offset}`, { cache: 'no-store' })).data; }
  catch (failure) { error = failure instanceof Error ? failure.message : 'Cities could not be loaded.'; }
  return <div className="cities-page"><header className="city-directory-heading"><h1>Explore cities.</h1><Link className="quiet-link" href="/discover">All events<Icon name="arrow" size={18} /></Link></header>{error ? <section className="desk-empty"><h2>We couldn’t load cities.</h2><p role="alert">{error}</p><Link className="app-button app-button-outline" href="/cities">Try again</Link></section> : cities.length ? <div className="city-directory">{cities.map(city => <Link className="city-card" href={`/cities/${city.slug}`} key={city.id}><div className="city-card-image">{city.banner_url ? <Image unoptimized src={city.banner_url} alt="" width={720} height={420} loading="lazy" /> : <span aria-hidden="true">{city.name.slice(0, 1)}</span>}</div><div className="city-card-copy"><h2>{city.name}<Icon name="arrow" size={22} /></h2>{typeof city.event_count === 'number' && <p>{city.event_count} upcoming {city.event_count === 1 ? 'event' : 'events'}</p>}</div></Link>)}</div> : <section className="desk-empty"><Icon name="location" size={30} /><h2>New cities are on the way.</h2><p>Explore all public events while we prepare local City Hubs.</p><Link className="app-button app-button-outline" href="/discover">Explore events</Link></section>}{(offset > 0 || cities.length === 20) && <nav className="journey-pagination" aria-label="More cities">{offset > 0 && <Link href={`/cities?offset=${Math.max(0, offset - 20)}`}>Previous</Link>}{cities.length === 20 && <Link href={`/cities?offset=${offset + 20}`}>Next</Link>}</nav>}</div>;
}
