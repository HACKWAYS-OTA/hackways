import { managedMetadata } from '@hackways/ui/seo';
import { pageSocialMetadata } from '@/lib/social-metadata';
import type { Metadata } from 'next';
import Link from 'next/link';
import { CopyEndpoint } from './CopyEndpoint';
import './mcp.css';

const pageDefaults: Metadata = { ...pageSocialMetadata('mcp'),
  title: 'Hackways MCP', description: 'Connect your AI tools to real Hackways events, registrations and organizer workflows with permissions you control.',
  alternates: { canonical: '/mcp' },
};
export function generateMetadata() { return managedMetadata("/mcp", pageDefaults); }
const workflows = [
  { title: 'Find your next event', text: 'Search published events and Spaces, read registration questions, follow your applications and open your event pass.', example: 'Find upcoming meetups and show me their registration requirements.' },
  { title: 'Keep your event moving', text: 'Create and update events, build registration sections, invite attendees, review applications and manage check-in.', example: 'Show the applications awaiting approval for my event.' },
  { title: 'Run the platform', text: 'Work with authorized admin records, analytics, campaigns, Discover promotions and feature rollouts.', example: 'Review the current campaign audience before scheduling it.' },
];
export default function MCPPage() {
  return <div className="mcp-page site-container">
    <section className="mcp-intro" aria-labelledby="mcp-title">
      <div><h1 id="mcp-title">Your events.<br />In the conversation.</h1><p>Connect your AI workspace to Hackways. Find a gathering, manage your guests and keep the details moving with the Model Context Protocol.</p><Link className="button button-primary" href="/mcp/setup">Set up your connection</Link></div>
      <div className="mcp-endpoint"><h2>Connect to Hackways</h2><p>Add this remote server URL in an MCP-compatible client.</p><code>https://mcp.hackways.in/mcp</code><CopyEndpoint /><p className="mcp-endpoint-note">Streamable HTTP · OAuth authorization</p></div>
    </section>
    <section className="mcp-workflows" aria-labelledby="mcp-workflows-title"><h2 id="mcp-workflows-title">From discovery to the door.</h2>{workflows.map(workflow => <article key={workflow.title}><div><h3>{workflow.title}</h3><p>{workflow.text}</p></div><blockquote>{workflow.example}</blockquote></article>)}</section>
    <section className="mcp-control" aria-labelledby="mcp-control-title"><div><h2 id="mcp-control-title">Your account.<br />Your permission.</h2><p>Public discovery works without signing in. Personal, organizer and admin tools ask you to connect your Hackways account.</p><a className="button button-outline" href="/connect/connections">Manage connected apps</a></div><ol><li><h3>Choose what to share</h3><p>A separate Hackways authorization page shows the app, your account and its requested permissions.</p></li><li><h3>Review every change</h3><p>Before a tool changes a record or sends a message, approve the exact action in your browser.</p></li><li><h3>Stay in control</h3><p>Disconnect whenever you choose. Event ownership, capacity rules and daily admin verification still apply.</p></li></ol></section>
    <section className="mcp-help" aria-labelledby="mcp-help-title"><h2 id="mcp-help-title">A few practical details</h2><details><summary>Does this give an AI app admin access?</summary><p>Only an existing administrator can authorize admin permissions. The dedicated admin sign-in and daily verification remain required. An account connection cannot grant itself administrator access.</p></details><details><summary>Can I use it with ChatGPT?</summary><p>Use an account and workspace that supports adding remote MCP servers. Add the URL above and follow the authorization flow. Availability and workspace permissions are controlled by your client.</p></details><details><summary>Are my password and verification codes shared?</summary><p>No. Sign in on Hackways. Your connected app uses a separate, revocable credential and never needs your password, email verification code or browser session cookie.</p></details><details><summary>What happens when a tool needs approval?</summary><p>Open its Hackways approval link, review the request and approve or deny it. Return to your app to complete the approved action. If an operation cannot be confirmed, check the saved record before trying again.</p></details><p>Read our <Link href="/privacy">privacy notice</Link> or <Link href="/help">get help</Link>.</p></section>
  </div>;
}

export const revalidate = 30;
