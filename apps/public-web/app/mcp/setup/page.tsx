import { managedMetadata } from '@hackways/ui/seo';
import { pageSocialMetadata } from '@/lib/social-metadata';
import type { Metadata } from 'next';
import Link from 'next/link';
import { CopyEndpoint } from '../CopyEndpoint';
import '../mcp.css';
import './setup.css';

const pageDefaults: Metadata = { ...pageSocialMetadata('mcp-setup'),
  title: 'MCP setup | Hackways',
  description: 'Connect Hackways to ChatGPT with the exact MCP URL, OAuth endpoints, client registration options, scopes and PKCE settings.',
  alternates: { canonical: '/mcp/setup' },
};
export function generateMetadata() { return managedMetadata("/mcp/setup", pageDefaults); }

const issuer = 'https://mcp.hackways.in';
const endpoints = [
  ['Server URL', `${issuer}/mcp`],
  ['Auth URL', `${issuer}/oauth/authorize`],
  ['Token URL', `${issuer}/oauth/token`],
  ['Registration URL', `${issuer}/oauth/register`],
  ['Authorization server base', issuer],
  ['Resource', `${issuer}/mcp`],
  ['Authorization server metadata', `${issuer}/.well-known/oauth-authorization-server`],
  ['Protected resource metadata', `${issuer}/.well-known/oauth-protected-resource/mcp`],
] as const;
const scopes = [
  ['profile:read', 'Read your name and email.'],
  ['profile:write', 'Update your profile and preferences.'],
  ['registrations:read', 'Read your registrations and pass links.'],
  ['registrations:write', 'Register, accept invitations and manage your attendance.'],
  ['organizer:read', 'Read events, guests and Spaces you manage.'],
  ['organizer:write', 'Manage your events, questions, invitations and check-in.'],
  ['admin:read', 'Read platform records and analytics as an administrator.'],
  ['admin:write', 'Manage authorized platform operations as an administrator.'],
] as const;
const registrationCommand = `curl https://mcp.hackways.in/oauth/register \\
  -H 'Content-Type: application/json' \\
  --data '{"client_name":"ChatGPT","redirect_uris":["PASTE_YOUR_EXACT_CHATGPT_CALLBACK_URL"],"token_endpoint_auth_method":"none","grant_types":["authorization_code","refresh_token"],"response_types":["code"]}'`;

function Settings({ rows, copyable = false }: { rows: readonly (readonly [string, string])[]; copyable?: boolean }) {
  return <dl className="mcp-setup-settings">{rows.map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{copyable ? <code>{value}</code> : <span>{value}</span>}{copyable && <CopyEndpoint value={value} label={`Copy ${label.toLowerCase()}`} />}</dd></div>)}</dl>;
}

export default function MCPSetupPage() {
  return <div className="mcp-page mcp-setup site-container">
    <header className="mcp-setup-intro">
      <Link className="mcp-setup-back" href="/mcp">Hackways MCP</Link>
      <h1>Connect Hackways<br />to your workspace.</h1>
      <p>Use these settings when adding Hackways as a remote MCP server in ChatGPT. Start with automatic discovery; every manual field is listed below.</p>
    </header>

    <section className="mcp-setup-start" aria-labelledby="start-title">
      <h2 id="start-title">Quick setup</h2>
      <ol><li>Enter <code>{issuer}/mcp</code> as the server URL and choose OAuth.</li><li>Choose <strong>Client Identifier Metadata Document (CIMD)</strong> when available. Dynamic Client Registration (DCR) is also supported.</li><li>Use token endpoint auth method <code>none</code> and leave the client secret empty. Enable PKCE with <code>S256</code>.</li><li>Save, connect your Hackways account and review the requested permissions.</li></ol>
      <CopyEndpoint />
      <p className="mcp-setup-note">Use the full URL ending in <code>/mcp</code>. The bare domain opens the Hackways landing page.</p>
    </section>

    <section className="mcp-setup-section" aria-labelledby="registration-title">
      <h2 id="registration-title">Client registration</h2>
      <p>Choose one registration method. Your client ID depends on that choice.</p>
      <div className="mcp-setup-method"><h3>CIMD</h3><p>ChatGPT supplies its metadata URL as the OAuth client ID. No client secret or manual registration is needed. Hackways accepts ChatGPT’s published public-client method, <code>none</code>.</p><p>The stable metadata URL is <code>https://chatgpt.com/oauth/client.json</code>. Some connections use a callback-specific metadata URL supplied by ChatGPT; keep that exact value.</p></div>
      <div className="mcp-setup-method"><h3>Dynamic Client Registration (DCR)</h3><p>ChatGPT registers itself at <code>{issuer}/oauth/register</code> and saves the returned client ID. Leave manual client credentials empty.</p></div>
      <details className="mcp-setup-manual"><summary>User-Defined OAuth Client</summary><p>If you select this method, first register a public client for the exact callback URL shown in your ChatGPT connection. Replace the marked value in this command, run it once, then paste the returned <code>client_id</code> into ChatGPT. Keep that client ID for the connection.</p><pre><code>{registrationCommand}</code></pre><CopyEndpoint value={registrationCommand} label="Copy registration command" /><p>This command creates an OAuth client registration. It does not sign in or authorize your account.</p></details>
      <Settings rows={[
        ['Callback URL', 'Copy the exact callback URL shown in your ChatGPT connection.'],
        ['OAuth client ID', 'CIMD metadata URL, or the client_id returned by DCR.'],
        ['OAuth client secret', 'Leave empty.'],
        ['Token endpoint auth method', 'none'],
      ]} />
      <p className="mcp-setup-note">Callbacks are specific to your connection. Register the complete URL exactly, including its path. A client ID is not your Hackways email address.</p>
    </section>

    <section className="mcp-setup-section" aria-labelledby="endpoints-title"><h2 id="endpoints-title">OAuth endpoints</h2><p>Discovery fills these automatically. If ChatGPT asks for manual settings, use these exact values.</p><Settings rows={endpoints} copyable /></section>

    <section className="mcp-setup-section" aria-labelledby="scopes-title">
      <h2 id="scopes-title">Scopes</h2>
      <p>Tools declare the permissions they need. Use <code>profile:read</code> for base scopes so the connection can identify your account. Default scopes can stay empty when ChatGPT uses the tool declarations, or use the minimal read permissions below.</p>
      <Settings rows={[
        ['Base scopes', 'profile:read'],
        ['Default scopes (optional)', 'profile:read\nregistrations:read'],
      ]} copyable />
      <p>Enter scopes one per line or separated by commas in ChatGPT. Request organizer or admin scopes only for workflows you need; account roles and ownership still determine access.</p>
      <dl className="mcp-setup-scopes">{scopes.map(([scope, description]) => <div key={scope}><dt><code>{scope}</code></dt><dd>{description}</dd></div>)}</dl>
      <p className="mcp-setup-note">Administrator access requires a separate admin sign-in and daily verification. Every change also needs approval on Hackways.</p>
    </section>

    <section className="mcp-setup-section" aria-labelledby="identity-title">
      <h2 id="identity-title">PKCE &amp; OpenID support</h2>
      <Settings rows={[
        ['PKCE', 'Enabled; S256 required.'],
        ['OIDC enabled', 'Off.'],
        ['OIDC configuration URL', 'Leave empty.'],
        ['OIDC userinfo endpoint', 'Leave empty.'],
        ['OIDC scopes supported', 'None. Do not add openid or email scopes.'],
      ]} />
      <p>Hackways uses OAuth authorization and the authenticated <code>get_my_profile</code> tool to return your account ID, name and email. An OpenID Connect provider is not configured.</p>
    </section>

    <section className="mcp-setup-section" aria-labelledby="help-title">
      <h2 id="help-title">If discovery fails</h2>
      <ul><li>Check that the server URL is <code>{issuer}/mcp</code>, with HTTPS and the final path.</li><li>Open the <a href={`${issuer}/.well-known/oauth-authorization-server`}>authorization server metadata</a> and <a href={`${issuer}/.well-known/oauth-protected-resource/mcp`}>protected resource metadata</a>. Both must return JSON. If either is unavailable, retry after service recovery.</li><li>Retry discovery in ChatGPT. If it still fails, enter the endpoints above and select CIMD or DCR.</li><li>For User-Defined OAuth Client, use a registered client ID and matching callback. Keep client secret empty and token auth method <code>none</code>.</li><li>If admin permissions are omitted on account consent, use the separate admin authorization link shown there.</li></ul>
      <p>Settings follow the <a href="https://developers.openai.com/plugins/build/auth">official OpenAI authentication guide</a>. Manage or revoke access from <Link href="/connect/connections">connected apps</Link>.</p>
    </section>
  </div>;
}

export const revalidate = 30;
