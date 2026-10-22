import type { ReactNode } from 'react';
import { headers } from 'next/headers';
import { redirect } from 'next/navigation';
import { getAdminAccess, safeAdminDestination } from '@/lib/admin-access';

export const dynamic = 'force-dynamic';
export const revalidate = 0;

export default async function ConsoleLayout({ children }: { children: ReactNode }) {
  const access = await getAdminAccess();
  if (access === 'signed-out') {
    const next = safeAdminDestination(headers().get('x-hackways-admin-path') || '/');
    redirect(`/login?next=${encodeURIComponent(next)}`);
  }
  if (access === 'denied') redirect('/access-denied');
  if (access === 'unavailable') redirect('/auth-unavailable');
  return children;
}
