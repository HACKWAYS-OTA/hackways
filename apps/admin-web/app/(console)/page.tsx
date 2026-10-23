import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { Icon } from '@/components/ui/Icon';
import { adminServer } from '@/lib/admin-server';
import '@/components/admin/overview.css';

export const metadata = { title: 'Platform overview | Hackways Admin' };
type CoreOverview = { upcoming_events: number; moderation_queue: number; pending_approvals: number; waiting_list: number; active_promotions: number };
type GlobalOverview = { accounts: number; administrators: number; pending_emails: number; failed_emails: number; active_sessions: number };
const count = (value: number) => new Intl.NumberFormat('en-IN').format(value);

function QueueLink({ label, value, href }: { label: string; value: number; href: string }) {
  return <Link className="overview-queue" href={href}><span>{label}</span><strong>{count(value)}</strong><Icon name="arrow" size={18} /></Link>;
}
function Unavailable() {
  return <p className="overview-unavailable" role="alert">Overview unavailable. Open the module to try again.</p>;
}

export default async function AdminHomePage() {
  const [core, global] = await Promise.allSettled([
    adminServer<{ data: CoreOverview }>('core', '/overview'),
    adminServer<{ data: GlobalOverview }>('global', '/overview'),
  ]);
  const events = core.status === 'fulfilled' ? core.value.data : null;
  const accounts = global.status === 'fulfilled' ? global.value.data : null;
  return <AdminShell active="/">
    <div className="admin-page-heading overview-heading"><h1>Platform overview</h1><Link className="app-button app-button-outline" href="/audit"><Icon name="audit" size={17} />Audit history</Link></div>
    <div className="overview-bento">
      <section className="overview-panel overview-events" aria-labelledby="overview-events-title">
        <div className="overview-panel-heading"><h2 id="overview-events-title">Events & registration</h2><Icon name="events" size={25} /></div>
        {events ? <>
          <div className="overview-feature-count"><div><strong>{count(events.upcoming_events)}</strong><span>Upcoming events</span></div><Link className="app-button app-button-primary" href="/events?status=published">Manage events<Icon name="arrow" size={17} /></Link></div>
          <div className="overview-queues">
            <QueueLink label="Moderation queue" value={events.moderation_queue} href="/moderation" />
            <QueueLink label="Awaiting approval" value={events.pending_approvals} href="/registrations?status=pending_approval" />
            <QueueLink label="Waiting list" value={events.waiting_list} href="/registrations?status=waitlisted" />
          </div>
        </> : <><Unavailable /><Link className="app-button app-button-primary" href="/events">Open events<Icon name="arrow" size={17} /></Link></>}
      </section>
      <section className="overview-panel overview-delivery" aria-labelledby="overview-delivery-title">
        <div className="overview-panel-heading"><h2 id="overview-delivery-title">Email delivery</h2><Icon name="messaging" size={25} /></div>
        {accounts ? <>
          <dl className="overview-delivery-counts"><div className={accounts.failed_emails ? 'overview-delivery-failed' : undefined}><dt>Failed emails</dt><dd>{count(accounts.failed_emails)}</dd></div><div><dt>Pending emails</dt><dd>{count(accounts.pending_emails)}</dd></div></dl>
          <div className="overview-panel-actions"><Link className="app-button app-button-outline" href={`/messaging?status=${accounts.failed_emails ? 'failed' : 'pending'}`}>{accounts.failed_emails ? 'Review failed emails' : 'View delivery queue'}<Icon name="arrow" size={17} /></Link><Link className="overview-text-action" href="/messaging/new">Compose message<Icon name="add" size={17} /></Link></div>
        </> : <><Unavailable /><Link className="app-button app-button-outline" href="/messaging">Open messaging<Icon name="arrow" size={17} /></Link></>}
      </section>
      <section className="overview-panel overview-accounts" aria-labelledby="overview-accounts-title">
        <div className="overview-panel-heading"><h2 id="overview-accounts-title">Accounts & access</h2><Icon name="people" size={25} /></div>
        {accounts ? <div className="overview-queues"><QueueLink label="Accounts" value={accounts.accounts} href="/people" /><QueueLink label="Administrators" value={accounts.administrators} href="/people?status=admin" /><QueueLink label="Active sessions" value={accounts.active_sessions} href="/people" /></div> : <Unavailable />}
        <Link className="overview-text-action" href="/people?status=admin">Manage access<Icon name="arrow" size={17} /></Link>
      </section>
      <section className="overview-panel overview-publishing" aria-labelledby="overview-publishing-title">
        <div className="overview-panel-heading"><h2 id="overview-publishing-title">Publishing & releases</h2><Icon name="external" size={25} /></div>
        {events ? <div className="overview-promotion-count"><strong>{count(events.active_promotions)}</strong><span>Active Discover promotions</span></div> : <Unavailable />}
        <nav className="overview-publishing-actions" aria-label="Publishing tools"><Link className="app-button app-button-primary" href="/discover-promotions">Schedule banners<Icon name="arrow" size={17} /></Link><Link className="app-button app-button-outline" href="/campaigns">Campaigns<Icon name="arrow" size={17} /></Link></nav>
        <Link className="overview-release-action" href="/rollouts"><span>Feature rollouts & versions</span><Icon name="arrow" size={20} /></Link>
      </section>
    </div>
  </AdminShell>;
}
