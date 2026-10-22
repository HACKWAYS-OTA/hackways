import { AuthForm } from '@/components/account/AuthForm';
import { safeNext } from '@/lib/journeys';
import { currentUser } from '@/lib/account';
import { redirect } from 'next/navigation';
export const metadata = { title: 'Sign in', robots: { index: false, follow: false } };
export default async function Login({ searchParams }: { searchParams: { next?: string } }) {
  const next = safeNext(searchParams.next);
  if (await currentUser()) redirect(next);
  return <AuthForm next={next} />;
}
