import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { campaignServer, type CampaignRow } from '@/lib/campaigns';
export const metadata = { title: 'Campaigns | Hackways Admin' };
export default async function Campaigns({ searchParams }: { searchParams: { q?: string; offset?: string } }) {
 const offset = Math.max(0, Number(searchParams.offset) || 0), query = searchParams.q || '';
 const result = await campaignServer<{ data: CampaignRow[]; has_next: boolean }>(`/campaigns?offset=${offset}&q=${encodeURIComponent(query)}`);
 return <AdminShell active="/campaigns" query={query}><div className="admin-page-heading"><h1>Campaigns</h1><div className="campaign-actions"><Link className="app-button app-button-outline" href="/campaigns/contacts">Contacts</Link><Link className="app-button app-button-primary" href="/campaigns/new">New campaign</Link></div></div>
 {result.data.length ? <div className="admin-table-scroll"><table className="admin-table"><caption className="sr-only">Promotional campaigns</caption><thead><tr><th>Campaign</th><th>Subject</th><th>Status</th><th>Updated</th></tr></thead><tbody>{result.data.map(row => <tr key={row.reference}><th scope="row"><Link className="admin-record-link" href={`/campaigns/details/${row.reference}`}>{row.title}</Link></th><td>{row.subject}</td><td>{row.status}</td><td>{new Date(row.updated_at).toLocaleString('en-IN', { timeZone: 'Asia/Kolkata' })}</td></tr>)}</tbody></table></div> : <div className="admin-result-state">No campaigns yet. Save a draft to preview your first email.</div>}
 <nav className="admin-pagination" aria-label="Campaign pages">{offset > 0 && <Link className="app-button app-button-outline" href={`/campaigns?offset=${Math.max(0,offset-20)}&q=${encodeURIComponent(query)}`}>Previous</Link>}{result.has_next && <Link className="app-button app-button-outline" href={`/campaigns?offset=${offset+20}&q=${encodeURIComponent(query)}`}>Next</Link>}</nav></AdminShell>;
}
