import { requireEventPermission } from '@/lib/management';
import { GuestWorkspace } from '@/components/management/GuestWorkspace';
export const metadata = { title: 'Guests' };
export default async function Guests({ params, searchParams }: { params: { id: string }; searchParams: { offset?: string; q?: string; status?: string } }) {
  return <GuestWorkspace event={(await requireEventPermission(params.id, 'guests.read')).data} searchParams={searchParams} />;
}
