import { managedEvent } from '@/lib/management';
import { GuestWorkspace } from '@/components/management/GuestWorkspace';
export const metadata = { title: 'Guests' };
export default async function Guests({ params, searchParams }: { params: { id: string }; searchParams: { offset?: string; q?: string; status?: string } }) {
  return <GuestWorkspace event={await managedEvent(params.id)} searchParams={searchParams} />;
}
