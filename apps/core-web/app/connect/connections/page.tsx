import { AuthLayout } from '@hackways/ui/auth';
import { MCPConnections, type MCPConnection } from '@hackways/ui/mcp';
import { mcpData } from '@/lib/mcp';
export const dynamic = 'force-dynamic';
export const metadata = { title: 'Connected apps', robots: { index: false, follow: false } };
export default async function Page() {
  const result = await mcpData<MCPConnection[]>('/connections', '/connect/connections');
  return <AuthLayout title="Connected apps" description="Review and revoke access to your Hackways account." homeHref="https://hackways.in">{result.data ? <MCPConnections initial={result.data} admin={false} /> : <p role="alert">{result.error}</p>}</AuthLayout>;
}
