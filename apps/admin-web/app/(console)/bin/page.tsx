import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { EventBinAction } from '@/components/admin/AdminActions';
import { adminServer } from '@/lib/admin-server';
import type { AdminRecord } from '@/lib/admin-modules';

export const metadata = { title: 'Bin | Hackways Admin' };

export default async function BinPage({ searchParams }: { searchParams: { q?: string; offset?: string } }) {
  const query = (searchParams.q || '').slice(0, 200);
  const parsed = Number(searchParams.offset || 0);
  const offset = Number.isSafeInteger(parsed) && parsed >= 0 && parsed <= 2147483600 ? parsed : 0;
  const result = await adminServer<{ data: AdminRecord[] }>('core', `/bin?${new URLSearchParams({ q: query, limit: '21', offset: String(offset) })}`);
  const rows = result.data.slice(0, 20);
  const path = (next: number) => `/bin?${new URLSearchParams({ q: query, offset: String(next) })}`;
  return <AdminShell active="/bin" query={query}>
    <div className="admin-page-heading"><h1>Bin</h1></div>
    {!rows.length ? <div className="admin-result-state"><p>{query ? 'No deleted events match your search.' : 'No deleted events.'}</p></div> : <>
      <div className="admin-table-scroll" role="region" aria-label="Deleted events" tabIndex={0}><table className="admin-table">
        <caption className="sr-only">Deleted events available for recovery</caption>
        <thead><tr><th scope="col">Event</th><th scope="col">Publication</th><th scope="col">Recovery</th></tr></thead>
        <tbody>{rows.map(record => <tr key={record.id}><th scope="row">{record.title}</th><td>{String(record.status).replaceAll('_', ' ')}</td><td><details><summary className="app-button app-button-outline app-button-sm">Restore event<span className="sr-only">: {record.title}</span></summary><EventBinAction record={record} restore /></details></td></tr>)}</tbody>
      </table></div>
      <nav className="admin-pagination" aria-label="Bin pages">
        {offset > 0 && <Link className="app-button app-button-outline app-button-sm" href={path(Math.max(0, offset - 20))}>Previous</Link>}
        <span>Page {Math.floor(offset / 20) + 1}</span>
        {result.data.length > 20 && <Link className="app-button app-button-outline app-button-sm" href={path(offset + 20)}>Next</Link>}
      </nav>
    </>}
  </AdminShell>;
}
