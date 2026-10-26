import { AuthForm } from '@/components/account/AuthForm';
import { brandIcons } from '@hackways/ui/brand-metadata';
import { safeNext } from '@/lib/journeys';
export const metadata = { title: 'Sign in', icons: brandIcons('/manage', 'admin-auth') };
export default function Login({ searchParams }: { searchParams: { next?: string } }) { return <AuthForm next={safeNext(searchParams.next, '/')} />; }
