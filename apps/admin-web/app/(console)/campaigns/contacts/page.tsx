import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { PromotionImport } from '@/components/admin/CampaignEditor';
import { campaignServer, type PromotionContact } from '@/lib/campaigns';
import { Icon } from '@/components/ui/Icon';
export const metadata = { title: 'Campaign contacts | Hackways Admin' };
export default async function Contacts({ searchParams }: { searchParams: { offset?: string } }) {
 const offset = Math.max(0, Number(searchParams.offset) || 0), result = await campaignServer<{ data: PromotionContact[]; has_next: boolean }>(`/campaign-contacts?offset=${offset}`);
 return <AdminShell active="/campaigns"><Link className="app-button app-button-outline campaign-back" href="/campaigns"><Icon name="back" />Back</Link><div className="admin-page-heading"><h1>Campaign contacts</h1></div><div className="campaign-import"><PromotionImport /></div><section className="campaign-section"><h2>Contacts</h2>{result.data.length ? <div className="admin-table-scroll"><table className="admin-table"><thead><tr><th>Email</th><th>Name</th><th>Preference</th><th>Source</th><th>Consent record</th></tr></thead><tbody>{result.data.map(row => <tr key={row.reference}><th scope="row">{row.email}</th><td>{row.display_name}</td><td>{row.subscribed ? 'Opted in' : 'Not subscribed'}</td><td>{row.source}</td><td>{row.consent_source}<br />{new Date(row.consented_at).toLocaleString('en-IN', { timeZone: 'Asia/Kolkata' })}</td></tr>)}</tbody></table></div> : <p>No promotional contacts have been added.</p>}<nav className="admin-pagination" aria-label="Contact pages">{offset > 0 && <Link className="app-button app-button-outline" href={`/campaigns/contacts?offset=${Math.max(0,offset-20)}`}>Previous</Link>}{result.has_next && <Link className="app-button app-button-outline" href={`/campaigns/contacts?offset=${offset+20}`}>Next</Link>}</nav></section></AdminShell>;
}
