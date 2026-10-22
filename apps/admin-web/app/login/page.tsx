import type { Metadata } from 'next';
import { AdminLogin } from '@/components/admin/AdminLogin';
import { redirect } from 'next/navigation';
import { getAdminAccess, safeAdminDestination } from '@/lib/admin-access';
import { AuthLayout } from '@hackways/ui/auth';

export const metadata: Metadata = { title: 'Sign in | Hackways Admin' };
export const dynamic = 'force-dynamic';
export default async function LoginPage({ searchParams }: { searchParams: { next?: string; switch?: string } }) {
  const next = safeAdminDestination(typeof searchParams.next === 'string' ? searchParams.next : '/');
  if (searchParams.switch !== '1') {
    const access = await getAdminAccess();
    if (access === 'allowed') redirect(next);
    if (access === 'verification-required') redirect(`/verify?next=${encodeURIComponent(next)}`);
    if (access === 'denied') redirect('/access-denied');
    if (access === 'unavailable') redirect('/auth-unavailable');
  }
  return <main><AuthLayout title="Admin sign in" homeHref="https://hackways.in"><AdminLogin next={next} /></AuthLayout></main>;
}
