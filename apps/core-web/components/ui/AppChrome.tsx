'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';
import { Icon } from '@/components/ui/Icon';
import { SkipToContent } from '@/components/ui/SkipToContent';
import { AppNavigation } from '@/components/ui/AppNavigation';
import { site } from '@/lib/site';

export function AppChrome({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const isAuthPage = pathname === '/login' || pathname.startsWith('/auth/');
  return <>
    <SkipToContent />
    {!isAuthPage && <header className="app-header"><div className="app-header-inner">
      <Link className="wordmark" href="/discover" aria-label="Hackways events">hackways.</Link>
      <AppNavigation />
      <div className="header-actions"><Link className="header-create" href="/create"><Icon name="add" size={17} />Create event</Link><Link className="header-account" href="/account">Account</Link></div>
    </div></header>}
    <main className="app-main" id="main-content" tabIndex={-1}>{children}</main>
    {!isAuthPage && <footer className="app-footer app-container"><p>© {new Date().getFullYear()} Hackways</p><div><a href={`${site.public}/about`}>About</a><a href={`${site.public}/privacy`}>Privacy</a><a href={`${site.public}/help`}>Help & FAQs</a></div></footer>}
  </>;
}
