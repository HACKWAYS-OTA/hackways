import { AuthLayout } from '@hackways/ui/auth';
import { MCPApproval, type MCPConsent } from '@hackways/ui/mcp';
import { mcpData } from '@/lib/mcp';
export const dynamic = 'force-dynamic';
export const metadata = { title: 'Connect to Hackways', robots: { index: false, follow: false } };
export default async function Page({ searchParams }: { searchParams: { request?: string } }) {
  const id = typeof searchParams.request === 'string' && /^[a-fA-F0-9-]{36}$/.test(searchParams.request) ? searchParams.request : '';
  const destination = '/connect/authorize?request=' + encodeURIComponent(id);
  const result = id ? await mcpData<MCPConsent>('/requests/' + encodeURIComponent(id), destination) : { error: 'Start a new connection request from your app.' };
  return <AuthLayout title="Connect to Hackways" homeHref="https://hackways.in">{result.data ? <MCPApproval data={result.data} action={false} admin={false} /> : <div className="hw-mcp-approval"><p role="alert">{result.error}</p><a href="/connect/connections">Manage connected apps</a></div>}</AuthLayout>;
}
