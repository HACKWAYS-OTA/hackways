import { redirect } from 'next/navigation';
import { currentUser } from '@/lib/account';
export default async function HomePage() { redirect(await currentUser() ? '/my-events' : '/discover'); }
