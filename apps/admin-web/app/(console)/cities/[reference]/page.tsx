import Link from 'next/link';
import { notFound } from 'next/navigation';
import { AdminShell } from '@/components/admin/AdminShell';
import { CityEditor } from '@/components/admin/CityEditor';
import { adminServer } from '@/lib/admin-server';
import type { AdminCity, CityEvent } from '@/lib/cities';
import '../cities.css';
export default async function EditCity({ params, searchParams }: { params: { reference: string }; searchParams: { offset?: string } }) {
 if (!/^[1-9]\d{0,14}$/.test(params.reference)) notFound();
 const raw = Number(searchParams.offset), offset = Number.isSafeInteger(raw) && raw > 0 && raw <= 100000 ? raw : 0;
 const result = await adminServer<{ data: AdminCity; events: CityEvent[] }>('core', `/cities/${params.reference}?limit=20&offset=${offset}`);
 return <AdminShell active="/cities"><Link className="quiet-link" href="/cities">Back to City Hubs</Link><header className="city-admin-heading"><div><h1>{result.data.name}</h1><p className="admin-secondary-copy">hackways.in/cities/{result.data.slug}</p></div></header><CityEditor initial={result.data} events={result.events} />{(offset > 0 || result.events.length === 20) && <nav className="admin-pagination" aria-label="Curated event pages">{offset > 0 && <Link href={`/cities/${params.reference}?offset=${Math.max(0, offset - 20)}`}>Previous events</Link>}{result.events.length === 20 && <Link href={`/cities/${params.reference}?offset=${offset + 20}`}>More events</Link>}</nav>}</AdminShell>;
}
