import type { ReactNode } from 'react';
import { Icon } from '@/components/ui/Icon';
import { AdminAccount } from './AdminAccount';
import { adminModules, adminModuleOrder, type AdminModule } from '@/lib/admin-modules';
import { getAdminAccess } from '@/lib/admin-access';
import { SectionJumpButton } from '@hackways/ui/section-jump';
import { AdminSessionRefresh } from './AdminSessionRefresh';

export async function AdminShell({ active, query = '', status = '', children }: { active: '/' | `/${AdminModule}` | '/campaigns' | '/discover-promotions' | '/rollouts' | '/preferences'; query?: string; status?: string; children: ReactNode }) {
  if (await getAdminAccess() !== 'allowed') return null;
  const currentModule = active === '/campaigns' ? 'messaging' : active === '/discover-promotions' || active === '/rollouts' || active === '/preferences' || active === '/' ? 'events' : active.slice(1) as AdminModule;
  const searchLabel = active === '/campaigns' ? 'Search campaigns' : `Search ${adminModules[currentModule].label.toLowerCase()}`;
  const snapshotAt = process.env.ADMIN_PREVIEW_SNAPSHOT_AT;
  return <div className="admin-frame">
    <AdminSessionRefresh />
    <SectionJumpButton targetId="admin-main-content" className="admin-skip-link">Skip to dashboard content</SectionJumpButton>
    <aside className="admin-sidebar" tabIndex={0} aria-label="Admin sidebar">
      <a href="/" className="wordmark">hackways.</a>
      <nav className="admin-navigation" aria-label="Admin navigation">
        <a href="/" aria-current={active === '/' ? 'page' : undefined}><Icon name="home" />Home</a>
        {adminModuleOrder.map((name) => <a key={name} href={`/${name}`} aria-current={active === `/${name}` ? 'page' : undefined}><Icon name={name} />{adminModules[name].label}</a>)}
        <a href="/campaigns" aria-current={active === '/campaigns' ? 'page' : undefined}><Icon name="messaging" />Campaigns</a>
        <a href="/discover-promotions" aria-current={active === '/discover-promotions' ? 'page' : undefined}><Icon name="events" />Discover promotions</a>
        <a href="/rollouts" aria-current={active === '/rollouts' ? 'page' : undefined}><Icon name="audit" />Feature rollouts & versions</a>
        <a href="/preferences" aria-current={active === '/preferences' ? 'page' : undefined}><Icon name="people" />Display & accessibility</a>
      </nav>
      <a href="https://hackways.in" className="admin-public-link">Public site<Icon name="external" size={17} /></a>
    </aside>
    <div className="admin-workspace" tabIndex={0} role="region" aria-label="Dashboard content">
      <header className="admin-topbar">
        <form action={active === '/campaigns' ? '/campaigns' : `/${currentModule}`} method="get" role="search" className="admin-search">
          <Icon name="search" size={22} />
          <label className="sr-only" htmlFor="admin-search">{searchLabel}</label>
          <input id="admin-search" name="q" type="search" maxLength={200} defaultValue={query} placeholder={searchLabel} />
          {status && <input type="hidden" name="status" value={status} />}
          <button type="submit" className="icon-button" aria-label={searchLabel}><Icon name="arrow" size={19} /></button>
        </form>
        <AdminAccount />
      </header>
      {snapshotAt && <p className="admin-data-source">Local data copy · {new Intl.DateTimeFormat('en-IN', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Kolkata' }).format(new Date(snapshotAt))}</p>}
      <main className="admin-content" id="admin-main-content" tabIndex={-1}>{children}</main>
    </div>
  </div>;
}
