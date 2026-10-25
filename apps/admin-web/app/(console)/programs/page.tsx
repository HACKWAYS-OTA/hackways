import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { adminServer } from '@/lib/admin-server';
import type { ProgramSummary } from '@hackways/api-client';
export const metadata = { title: 'Programs | Hackways Admin' };
export default async function Programs({ searchParams }: { searchParams: { q?: string; offset?: string } }) {
 const offset = Math.min(100000, Math.max(0, Number(searchParams.offset) || 0));
 const result = await adminServer<{ data: ProgramSummary[]; has_next: boolean }>('global', `/programs?q=${encodeURIComponent(searchParams.q || '')}&offset=${offset}`);
 return <AdminShell active="/programs" query={searchParams.q || ''}><div className="admin-page-heading"><div><h1>Programs</h1><p className="program-admin-caption">Create a program. Review the people behind each application.</p></div><Link className="app-button app-button-primary" href="/programs/new">Create program</Link></div><div className="program-admin-directory">{result.data.map(p => <Link key={p.slug} href={`/programs/${p.slug}`} className="program-admin-directory-row"><div><span className="program-admin-badge">{p.status}</span><h2>{p.title}</h2><p>{p.summary}</p></div><dl><div><dt>Applications</dt><dd>{p.applications}</dd></div><div><dt>Awaiting decision</dt><dd>{p.pending}</dd></div></dl></Link>)}{!result.data.length && <div className="admin-result-state">No programs found. Create a program to start accepting applications.</div>}</div><nav className="admin-pagination" aria-label="Program pages">{offset > 0 && <Link className="app-button app-button-outline" href={`/programs?offset=${Math.max(0, offset - 20)}&q=${encodeURIComponent(searchParams.q || '')}`}>Previous</Link>}{result.has_next && <Link className="app-button app-button-outline" href={`/programs?offset=${offset + 20}&q=${encodeURIComponent(searchParams.q || '')}`}>Next</Link>}</nav></AdminShell>;
}
