'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
export function ManagementTabs({ slug }: { slug: string }) {
  const pathname = usePathname();
  return <nav className="host-tabs" aria-label="Manage this event">{[['', 'Overview'], ['/guests', 'Guests'], ['/registration', 'Registration'], ['/check-in', 'Check-in'], ['/edit', 'Event page'], ['/activity', 'Activity'], ['/settings', 'Settings']].map(([suffix, label]) => <Link key={suffix} href={`/events/${slug}${suffix}`} aria-current={(pathname === `/events/${slug}${suffix}` || (suffix === '/edit' && pathname === `/events/${slug}/appearance`) || (suffix === '/registration' && pathname.startsWith(`/events/${slug}/registration/`))) ? 'page' : undefined}>{label}</Link>)}</nav>;
}
