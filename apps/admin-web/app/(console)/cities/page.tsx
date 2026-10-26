import Image from 'next/image';
import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { adminServer } from '@/lib/admin-server';
import type { AdminCity } from '@/lib/cities';
import './cities.css';
export default async function Cities({ searchParams }: { searchParams: { q?: string; offset?: string } }) {
 const query = (searchParams.q || '').slice(0, 200), rawOffset = Number(searchParams.offset), offset = Number.isSafeInteger(rawOffset) && rawOffset > 0 && rawOffset <= 100000 ? rawOffset : 0;
 const result = await adminServer<{ data: AdminCity[] }>('core', `/cities?q=${encodeURIComponent(query)}&limit=20&offset=${offset}`);
 return <AdminShell active="/cities" query={query}><header className="city-admin-heading"><div><h1>City Hubs</h1><p className="admin-secondary-copy">Local pages, curated events and a place for each city.</p></div><Link className="app-button app-button-primary" href="/cities/new">Create City Hub</Link></header><div className={result.data.length ? "city-admin-list" : "city-admin-list city-admin-list-empty"}>{result.data.length ? result.data.map(city => <Link key={city.id} href={`/cities/${city.reference}`}><span>{city.banner_url && <Image unoptimized src={city.banner_url} alt="" width={88} height={56} />}<span><strong>{city.name}</strong><small>{city.country} · /cities/{city.slug}</small></span></span><span>{city.status}</span></Link>) : <p className="city-admin-empty">{query ? 'No matching cities.' : 'Create your first City Hub to start curating local events.'}</p>}</div>{(offset > 0 || result.data.length === 20) && <nav className="admin-pagination" aria-label="City Hub pages">{offset > 0 && <Link href={`/cities?q=${encodeURIComponent(query)}&offset=${Math.max(0, offset - 20)}`}>Previous</Link>}{result.data.length === 20 && <Link href={`/cities?q=${encodeURIComponent(query)}&offset=${offset + 20}`}>Next</Link>}</nav>}</AdminShell>;
}
