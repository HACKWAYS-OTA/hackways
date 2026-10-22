'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';
import { SkipToContent } from '@/components/ui/SkipToContent';
import { site } from '@/lib/site';
import { Avatar } from '@hackways/ui/avatar';
import type { User } from '@/types/api';

export function AppChrome({ children, user }: { children: ReactNode; user: User | null }) {
  const pathname = usePathname();
  const isAuthPage = pathname === '/login' || pathname.startsWith('/auth/');
  return <>
    <SkipToContent />
    {!isAuthPage && <header className="app-header"><div className="app-header-inner"><a className="wordmark" href={`${site.app}/my-events`}>hackways.</a><nav className="app-nav" aria-label="Event management"><Link href="/">Manage events</Link><a href={`${site.app}/my-events`}>My events</a></nav><div className="header-actions"><a className="header-create" href={`${site.app}/create`}>Create event</a><a className="header-account" href={`${site.app}/account`} aria-label={user ? `Your profile, ${user.display_name}` : 'Sign in to your account'} title="Your profile"><Avatar name={user?.display_name} image={user?.avatar_url} /></a></div></div></header>}
    <main className="app-main" id="main-content" tabIndex={-1}>{children}</main>
    {!isAuthPage && <footer className="app-footer app-container"><p>Hackways</p><a href={`${site.app}/discover`}>Discover something new</a></footer>}
  </>;
}
