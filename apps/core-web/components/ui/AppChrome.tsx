'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';
import { Icon } from '@/components/ui/Icon';
import { SkipToContent } from '@/components/ui/SkipToContent';
import { AppNavigation } from '@/components/ui/AppNavigation';
import { site } from '@/lib/site';
import type { User } from '@/types/api';
import { Avatar as ProfileAvatar } from '@hackways/ui/avatar';

export function AppChrome({ children, user }: { children: ReactNode; user: User | null }) {
  const pathname = usePathname();
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
