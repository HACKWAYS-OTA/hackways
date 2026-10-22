import { redirect } from 'next/navigation';
import { AuthLayout } from '@hackways/ui/auth';
import { AdminVerification } from '@/components/admin/AdminVerification';
import { getAdminAccess, safeAdminDestination } from '@/lib/admin-access';
export const metadata = { title: 'Verify device | Hackways Admin' };
export const dynamic = 'force-dynamic';
export default async function Verify({ searchParams }: { searchParams: { next?: string } }) {
  const next = safeAdminDestination(searchParams.next || '/'), access = await getAdminAccess();
  if (access === 'allowed') redirect(next);
  if (access === 'signed-out') redirect(`/login?next=${encodeURIComponent(next)}`);
  if (access === 'denied') redirect('/access-denied');
  if (access === 'unavailable') redirect('/auth-unavailable');
  return <main><AuthLayout title="Verify your admin sign-in" homeHref="https://hackways.in"><AdminVerification next={next} /></AuthLayout></main>;
}
