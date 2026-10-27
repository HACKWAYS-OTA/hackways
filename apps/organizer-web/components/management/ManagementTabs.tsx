'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { WorkspacePermission } from '@/types/api';
export function ManagementTabs({ slug, permissions }: { slug: string; permissions: WorkspacePermission[] }) {
  const pathname = usePathname();
  return <nav className="host-tabs" aria-label="Manage this event">{([['', 'Overview', 'view'], ['/guests', 'Guests', 'guests.read'], ['/invitations', 'Invitations', 'guests.manage'], ['/registration', 'Registration', 'content'], ['/check-in', 'Check-in', 'checkin'], ['/edit', 'Event page', 'content'], ['/activity', 'Activity', 'guests.manage'], ['/settings', 'Settings', 'lifecycle'], ['/team', 'Team access', 'access.manage']] as const).filter(([, , permission]) => permissions.includes(permission)).map(([suffix, label]) => <Link key={suffix} href={`/events/${slug}${suffix}`} aria-current={(pathname === `/events/${slug}${suffix}` || (suffix === '/edit' && pathname === `/events/${slug}/appearance`) || (suffix === '/registration' && pathname.startsWith(`/events/${slug}/registration/`))) ? 'page' : undefined}>{label}</Link>)}</nav>;
}
