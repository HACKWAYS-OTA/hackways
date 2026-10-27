'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';
import { SkipToContent } from '@/components/ui/SkipToContent';
import { Icon } from '@/components/ui/Icon';
import { site } from '@/lib/site';
import { Avatar } from '@hackways/ui/avatar';
import { AnalyticsTracker, CookiePreferencesButton } from '@hackways/ui/analytics';
import type { User } from '@/types/api';

export function AppChrome({ children, user }: { children: ReactNode; user: User | null }) {
  const pathname = usePathname(), isAuthPage = pathname === '/login' || pathname.startsWith('/auth/');
  const analytics = <AnalyticsTracker path={`/manage${pathname === '/' ? '' : pathname}`} endpoint={`${(process.env.NEXT_PUBLIC_GLOBAL_API_URL || '/api/v1').replace(/\/$/, '')}/analytics/events`} />;
  const navigation = <><Link href="/" aria-current={!pathname.startsWith('/spaces') ? 'page' : undefined}><Icon name="calendar" size={20} />Events</Link><Link href="/spaces" aria-current={pathname.startsWith('/spaces') ? 'page' : undefined}><Icon name="people" size={20} />Spaces</Link></>;
  const preferences = <footer className="workspace-preferences"><CookiePreferencesButton /></footer>;
  if (isAuthPage) return <>{analytics}<SkipToContent /><main className="app-main" id="main-content" tabIndex={-1}>{children}</main>{preferences}</>;
  return <div className="workspace-frame">{analytics}<SkipToContent />
    <aside className="workspace-sidebar"><Link className="wordmark" href="/">hackways.</Link><nav aria-label="Managed workspace">{navigation}</nav><a className="workspace-public" href={`${site.app}/discover`}>Back to Hackways<Icon name="external" size={16} /></a></aside>
    <div className="workspace-main"><header className="workspace-topbar"><Link className="wordmark workspace-mobile-brand" href="/">hackways.</Link><div className="workspace-topbar-actions"><a className="app-button app-button-outline app-button-sm" href={`${site.app}/create`}><Icon name="add" size={16} />Create event</a><a className="header-account" href={`${site.app}/account`} aria-label={user ? `Your profile, ${user.display_name}` : 'Sign in'}><Avatar name={user?.display_name} image={user?.avatar_url} /></a></div><nav className="workspace-mobile-nav" aria-label="Managed workspace">{navigation}</nav></header><main className="workspace-content" id="main-content" tabIndex={-1}>{children}</main>{preferences}</div>
  </div>;
}
