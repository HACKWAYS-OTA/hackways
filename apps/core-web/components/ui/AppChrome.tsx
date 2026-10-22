'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useEffect, useLayoutEffect, useRef, type ReactNode } from 'react';
import { Icon } from '@/components/ui/Icon';
import { SkipToContent } from '@/components/ui/SkipToContent';
import { AppNavigation } from '@/components/ui/AppNavigation';
import { site } from '@/lib/site';
import type { User } from '@/types/api';
import { Avatar as ProfileAvatar } from '@hackways/ui/avatar';

export function AppChrome({ children, user }: { children: ReactNode; user: User | null }) {
  const pathname = usePathname();
  const previousPath = useRef(pathname);
  const historyPath = useRef<string | null>(null);
  useEffect(() => {
    const rememberHistoryNavigation = () => { historyPath.current = window.location.pathname; };
    window.addEventListener('popstate', rememberHistoryNavigation);
    return () => window.removeEventListener('popstate', rememberHistoryNavigation);
  }, []);
  useLayoutEffect(() => {
    const changed = previousPath.current !== pathname;
    const restoringHistory = historyPath.current === pathname;
    previousPath.current = pathname;
    historyPath.current = null;
    // Event scenes can remain visible during Next's scroll heuristic. New visits
    // start at the top before paint; history, query updates and refreshes retain position.
    if (changed && !restoringHistory && /^\/events\/[^/]+(?:\/register)?\/?$/.test(pathname)) {
      window.scrollTo({ top: 0, left: 0, behavior: 'instant' });
    }
  }, [pathname]);
  const isAuthPage = pathname === '/login' || pathname.startsWith('/auth/');
  return <>
    <SkipToContent />
    {!isAuthPage && <header className="app-header"><div className="app-header-inner">
      <Link className="wordmark" href="/discover" aria-label="Hackways events">hackways.</Link>
      <AppNavigation />
      <div className="header-actions"><Link className="header-create" href="/create"><Icon name="add" size={17} />Create event</Link><Link className="header-account" href="/account" aria-label={user ? `Your profile, ${user.display_name}` : 'Sign in to your account'} title={user ? 'Your profile' : 'Sign in'} aria-current={pathname === '/account' ? 'page' : undefined}><ProfileAvatar name={user?.display_name} image={user?.avatar_url} /></Link></div>
    </div></header>}
    <main className="app-main" id="main-content" tabIndex={-1}>{children}</main>
    {!isAuthPage && <footer className="app-footer app-container"><p>© {new Date().getFullYear()} Hackways</p><div><a href={`${site.public}/about`}>About</a><a href={`${site.public}/privacy`}>Privacy</a><a href={`${site.public}/help`}>Help & FAQs</a></div></footer>}
  </>;
}
