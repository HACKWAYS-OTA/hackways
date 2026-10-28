import Link from 'next/link';
import { HugeiconsIcon } from '@hugeicons/react';
import { ArrowUpRight01Icon } from '@hugeicons/core-free-icons';
import { AdminShell } from '@/components/admin/AdminShell';
import { NoticeListSync } from '@/components/admin/NoticeListSync';
import { UtilityNavigation } from '@/components/admin/UtilityNavigation';
import { adminServer } from '@/lib/admin-server';
import type { NoticeListItem } from '@/lib/utilities';
import '../utilities.css';
export default async function Notices({ searchParams }: { searchParams: { q?: string; offset?: string } }) {
 const query = (searchParams.q || '').slice(0,200), raw = Number(searchParams.offset), offset = Number.isSafeInteger(raw) && raw > 0 && raw <= 100000 ? raw : 0;
 const result = await adminServer<{ data: NoticeListItem[]; has_next: boolean }>('global', `/utilities/notices?q=${encodeURIComponent(query)}&limit=20&offset=${offset}`);
 return <AdminShell active="/utilities/notice" query={query}><NoticeListSync/><UtilityNavigation active="notice" /><header className="utility-heading"><div><h1>Notices</h1><p className="admin-secondary-copy">Write, publish and manage official updates.</p></div><Link className="app-button app-button-primary" href="/utilities/notice/new" target="_blank" rel="noopener noreferrer">Create notice<HugeiconsIcon icon={ArrowUpRight01Icon} size={17}/></Link></header><div className="notice-card-list">{result.data.map(row=><a key={row.reference} className="notice-list-card" href={`/utilities/notice/${row.reference}`} target="_blank" rel="noopener noreferrer"><div><h2>{row.title||'Untitled notice'}</h2>{row.summary&&<p>{row.summary}</p>}<small>/notice/{row.slug}</small></div><span className="notice-card-state">{row.status}</span><HugeiconsIcon icon={ArrowUpRight01Icon} size={21}/><span className="sr-only">Open editor in a new tab</span></a>)}{!result.data.length&&<p className="utility-empty">{query?'No notices match your search.':'Create a notice when you have an update to share.'}</p>}</div><nav className="admin-pagination" aria-label="Notice pages">{offset > 0 && <Link href={`/utilities/notice?q=${encodeURIComponent(query)}&offset=${offset-20}`}>Previous</Link>}{result.has_next && <Link href={`/utilities/notice?q=${encodeURIComponent(query)}&offset=${offset+20}`}>Next</Link>}</nav></AdminShell>;
}
