import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { AccountRestrictionAction, BrowserUnblockAction } from '@/components/admin/AdminActions';
import { adminServer } from '@/lib/admin-server';
import type { AdminRecord } from '@/lib/admin-modules';

export const metadata = { title: 'Restrictions | Hackways Admin' };

export default async function RestrictionsPage({ searchParams }: { searchParams: { kind?: string; q?: string; offset?: string } }) {
  const kind = searchParams.kind === 'devices' ? 'devices' : 'accounts';
  const query = (searchParams.q || '').slice(0, 200);
  const parsed = Number(searchParams.offset || 0);
  const offset = Number.isSafeInteger(parsed) && parsed >= 0 && parsed <= 2147483600 ? parsed : 0;
  const result = await adminServer<{ data: AdminRecord[] }>('global', `/restrictions/${kind}?${new URLSearchParams({ q: query, limit: '21', offset: String(offset) })}`);
  const rows = result.data.slice(0, 20);
  const path = (nextKind: string, next = 0) => `/restrictions?${new URLSearchParams({ kind: nextKind, q: query, offset: String(next) })}`;
  return <AdminShell active="/restrictions" query={query} searchKind={kind}>
    <div className="admin-page-heading"><h1>Restrictions</h1></div>
    <nav className="admin-filters" aria-label="Restriction types"><Link href={path('accounts')} aria-current={kind === 'accounts' ? 'page' : undefined}>Accounts</Link><Link href={path('devices')} aria-current={kind === 'devices' ? 'page' : undefined}>Browsers</Link></nav>
    {!rows.length ? <div className="admin-result-state"><p>{query ? 'No restrictions match your search.' : `No blocked ${kind === 'devices' ? 'browsers' : 'accounts'}.`}</p></div> : <>
      <div className="admin-table-scroll" role="region" aria-label="Restrictions awaiting review" tabIndex={0}><table className="admin-table">
        <caption className="sr-only">Restrictions awaiting administrator review</caption>
        <thead><tr><th scope="col">{kind === 'devices' ? 'Browser' : 'Account'}</th>{kind === 'accounts' && <th scope="col">Email</th>}<th scope="col">Reason</th><th scope="col">Review</th></tr></thead>
        <tbody>{rows.map(record => <tr key={record.id}><th scope="row">{kind === 'accounts' ? <Link className="admin-record-link" href={`/people/details/${record.reference}`}>{record.display_name}</Link> : String(record.user_agent || 'Recognized browser')}</th>{kind === 'accounts' && <td>{record.email}</td>}<td>{record.reason}</td><td><details><summary className="app-button app-button-outline app-button-sm">{kind === 'devices' ? 'Unblock browser' : 'Unblock account'}<span className="sr-only">: {kind === 'devices' ? record.reference : record.display_name}</span></summary>{kind === 'devices' ? <BrowserUnblockAction record={record} /> : <AccountRestrictionAction record={record} />}</details></td></tr>)}</tbody>
      </table></div>
      <nav className="admin-pagination" aria-label="Restriction pages">{offset > 0 && <Link className="app-button app-button-outline app-button-sm" href={path(kind, Math.max(0, offset - 20))}>Previous</Link>}<span>Page {Math.floor(offset / 20) + 1}</span>{result.data.length > 20 && <Link className="app-button app-button-outline app-button-sm" href={path(kind, offset + 20)}>Next</Link>}</nav>
    </>}
  </AdminShell>;
}
