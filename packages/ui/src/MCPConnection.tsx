'use client';

import { useRef, useState } from 'react';

export interface MCPConsent {
  id: string; client_name: string; client_verified: boolean; client_id: string;
  scopes: string[]; admin_authorize_url?: string; expires_at: string; email: string; status: string;
}
export interface MCPAction {
  id: string; tool: string; client_name: string; payload: Record<string, unknown>;
  path: string; status: string; expires_at: string;
}
export interface MCPConnection {
  id: string; client_name: string; client_verified: boolean; scopes: string[];
  created_at: string; expires_at: string; admin_connection: boolean;
}
const permissions: Record<string, string> = {
  'profile:read': 'Read your profile, devices, notifications and Space memberships',
  'profile:write': 'Update your name, email preferences, devices and memberships',
  'registrations:read': 'Read your registrations, invitations and pass links',
  'registrations:write': 'Register, accept, cancel or transfer your registrations',
  'organizer:read': 'Read events, questions, guests and Spaces you manage',
  'organizer:write': 'Manage your events, questions, invitations and attendee check-in',
  'admin:read': 'Read platform records, analytics, campaigns and rollout settings',
  'admin:write': 'Change platform records, access, messaging, promotions and rollouts',
};
function apiBase(admin: boolean) {
  if (admin) return '/api/v1/auth/mcp';
  return `${(process.env.NEXT_PUBLIC_GLOBAL_API_URL || '/api/v1').replace(/\/$/, '')}/auth/mcp`;
}
async function responseError(response: Response) {
  const body: unknown = await response.json().catch(() => null);
  if (body && typeof body === 'object' && 'error' in body && body.error && typeof body.error === 'object' && 'message' in body.error && typeof body.error.message === 'string') return body.error.message;
  return 'The request could not be completed. Refresh this page and try again.';
}

export function MCPApproval({ data, action = false, admin = false }: { data: MCPConsent | MCPAction; action?: boolean; admin?: boolean }) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const [decision, setDecision] = useState<boolean | null>(null);
  const inFlight = useRef(false);
  const isConsent = 'scopes' in data;
  async function decide(allow: boolean) {
    if (inFlight.current) return;
    inFlight.current = true; setPending(true); setError('');
    const controller = new AbortController(); const timeout = setTimeout(() => controller.abort(), 12000);
    try {
      const response = await fetch(`${apiBase(admin)}/${action ? 'actions' : 'requests'}/${encodeURIComponent(data.id)}/decision${admin ? '?admin=true' : ''}`, {
        method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ allow }), signal: controller.signal,
      });
      if (!response.ok) { setError(await responseError(response)); return; }
      if (action) { setDecision(allow); return; }
      const result: unknown = await response.json();
      if (result && typeof result === 'object' && 'redirect_url' in result && typeof result.redirect_url === 'string') window.location.assign(result.redirect_url);
      else setError('The return address was unavailable. Restart the connection from your app.');
    } catch { setError('The connection could not be confirmed. Refresh to check its status before trying again.'); }
    finally { clearTimeout(timeout); inFlight.current = false; setPending(false); }
  }
  if (decision !== null) return <div className="hw-mcp-result" role="status"><h2>{decision ? 'Action approved' : 'Action denied'}</h2><p>{decision ? 'Return to your app to complete this exact action before the request expires.' : 'Your app cannot execute this action.'}</p><a href="/connect/connections">Manage connected apps</a></div>;
  return <div className="hw-mcp-approval" aria-busy={pending}>
    <p><strong>{data.client_name}</strong>{isConsent && <span className="hw-mcp-identity">{data.client_verified ? 'Verified client metadata' : 'Client identity has not been verified'}</span>}</p>
    {isConsent ? <>
      <p className="hw-mcp-account">Connecting as <strong>{data.email}</strong></p>
      {data.admin_authorize_url && <p><a href={data.admin_authorize_url}>Connect an administrator account instead</a></p>}<h2>Allow this app to</h2>
      <ul className="hw-mcp-permissions">{data.scopes.map(scope => <li key={scope}>{permissions[scope] || scope}</li>)}</ul>
      <p>You review every change in Hackways before it can run. You can disconnect this app at any time.</p>
      <details><summary>Client details</summary><p className="hw-mcp-client-id">{data.client_id}</p></details>
    </> : <>
      <p>Review the exact request before allowing {data.client_name} to continue.</p>
      <h2>{data.tool.replaceAll('_', ' ')}</h2>
      <dl className="hw-mcp-operation"><dt>Destination</dt><dd>{data.path}</dd></dl>
      <pre className="hw-mcp-payload" tabIndex={0} aria-label="Exact action fields">{JSON.stringify(data.payload, null, 2)}</pre>
      <p>Approval applies only to these fields and this connected app. It does not approve future changes.</p>
    </>}
    {error && <p className="hw-auth-error" role="alert">{error}</p>}
    {data.status === 'pending' ? <div className="hw-mcp-actions"><button type="button" className="hw-mcp-secondary" disabled={pending} onClick={() => void decide(false)}>{action ? 'Deny action' : 'Cancel'}</button><button type="button" className="hw-mcp-primary" disabled={pending} onClick={() => void decide(true)}>{pending ? 'Saving...' : action ? 'Approve action' : 'Allow connection'}</button></div> : <p role="status">This request is {data.status}. Return to your app or start a new request.</p>}
    <a className="hw-mcp-manage" href="/connect/connections">Manage connected apps</a>
  </div>;
}

export function MCPConnections({ initial, admin = false }: { initial: MCPConnection[]; admin?: boolean }) {
  const [connections, setConnections] = useState(initial);
  const [pending, setPending] = useState<string | null>(null);
  const [error, setError] = useState('');
  async function disconnect(id: string) {
    if (pending) return;
    setPending(id); setError('');
    const controller = new AbortController(); const timer = setTimeout(() => controller.abort(), 12000);
    try {
      const response = await fetch(`${apiBase(admin)}/connections/${encodeURIComponent(id)}${admin ? '?admin=true' : ''}`, { method: 'DELETE', credentials: 'include', signal: controller.signal });
      if (!response.ok) { setError(await responseError(response)); return; }
      setConnections(current => current.filter(item => item.id !== id));
    } catch { setError('Disconnect could not be confirmed. Refresh to check the connection.'); }
    finally { clearTimeout(timer); setPending(null); }
  }
  return <div className="hw-mcp-connections">
    {error && <p className="hw-auth-error" role="alert">{error}</p>}
    {!connections.length && <p>No apps are connected. Start from your MCP client using <code>https://mcp.hackways.in/mcp</code>.</p>}
    {connections.map(item => <article className="hw-mcp-connection" key={item.id}><h2>{item.client_name}</h2><p>{item.admin_connection ? 'Administrator connection' : 'Account connection'}</p><ul>{item.scopes.map(scope => <li key={scope}>{permissions[scope] || scope}</li>)}</ul><button type="button" className="hw-mcp-secondary" disabled={pending !== null} onClick={() => void disconnect(item.id)}>{pending === item.id ? 'Disconnecting...' : 'Disconnect app'}</button></article>)}
    <a href={admin ? '/' : '/account'}>{admin ? 'Back to admin console' : 'Back to your account'}</a>
  </div>;
}
