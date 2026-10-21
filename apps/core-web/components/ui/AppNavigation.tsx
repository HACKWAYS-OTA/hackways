"use client";
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Icon } from './Icon';
export function AppNavigation({ compact = false }: { compact?: boolean }) {
  const pathname = usePathname();
  return <nav className={`app-nav${compact ? ' app-nav-compact' : ''}`} aria-label="Main navigation"><Link href="/discover" aria-label="Discover" title="Discover" aria-current={pathname === '/discover' ? 'page' : undefined}><Icon name="discover" size={18} /><span>Discover</span></Link><Link href="/spaces" aria-label="Spaces" title="Spaces" aria-current={pathname.startsWith('/spaces') ? 'page' : undefined}><Icon name="people" size={18} /><span>Spaces</span></Link><Link href="/my-events" aria-label="My events" title="My events" aria-current={pathname.startsWith('/my-events') ? 'page' : undefined}><Icon name="calendar" size={18} /><span>My events</span></Link></nav>;
}
