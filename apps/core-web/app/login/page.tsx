import { AuthForm } from '@/components/account/AuthForm';
import { safeNext } from '@/lib/journeys';
export const metadata = { title: 'Sign in', robots: { index: false, follow: false } };
export default function Login({ searchParams }: { searchParams: { next?: string } }) {
  return <AuthForm next={safeNext(searchParams.next)} />;
}
