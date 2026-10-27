import type { WorkspaceTeam } from '@/types/api';
import { requireEventPermission } from '@/lib/management';
import { privateApi } from '@/lib/account';
import { TeamAccess } from '@/components/management/TeamAccess';

export const metadata = { title: 'Event team access' };
export default async function EventTeam({ params }: { params: { id: string } }) {
  const { data: event } = await requireEventPermission(params.id, 'access.manage');
  const team = await privateApi<WorkspaceTeam>(`/workspace/events/${event.id}/access?limit=50`);
  return <TeamAccess kind="event" resourceID={event.id} initial={team} />;
}
