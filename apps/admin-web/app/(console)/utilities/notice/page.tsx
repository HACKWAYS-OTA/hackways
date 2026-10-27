import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { UtilityNavigation } from '@/components/admin/UtilityNavigation';
import { adminServer } from '@/lib/admin-server';
import type { NoticeListItem } from '@/lib/utilities';
import '../utilities.css';
export default async function Notices({ searchParams }: { searchParams: { q?: string; offset?: string } }) {
 const query = (searchParams.q || '').slice(0,200), raw = Number(searchParams.offset), offset = Number.isSafeInteger(raw) && raw > 0 && raw <= 100000 ? raw : 0;
 const result = await adminServer<{ data: NoticeListItem[]; has_next: boolean }>('global', `/utilities/notices?q=${encodeURIComponent(query)}&limit=20&offset=${offset}`);
 return <AdminShell active="/utilities/notice" query={query}><UtilityNavigation active="notice" /><header className="utility-heading"><div><h1>Notices</h1><p className="admin-secondary-copy">Write, publish and manage official updates.</p></div><Link className="app-button app-button-primary" href="/utilities/notice/new">Create notice</Link></header><div className="utility-table-scroll"><table className="utility-table"><thead><tr><th scope="col">Notice</th><th scope="col">Visibility</th><th scope="col">Manage</th></tr></thead><tbody>{result.data.map(row => <tr key={row.reference}><td><strong>{row.title||'Untitled notice'}</strong><small>/notice/{row.slug}</small></td><td>{row.status}</td><td><Link href={`/utilities/notice/${row.reference}`}>Edit notice</Link></td></tr>)}</tbody></table>{!result.data.length && <p className="utility-empty">{query ? 'No notices match your search.' : 'Create a notice when you have an update to share.'}</p>}</div><nav className="admin-pagination" aria-label="Notice pages">{offset > 0 && <Link href={`/utilities/notice?q=${encodeURIComponent(query)}&offset=${offset-20}`}>Previous</Link>}{result.has_next && <Link href={`/utilities/notice?q=${encodeURIComponent(query)}&offset=${offset+20}`}>Next</Link>}</nav></AdminShell>;
}
