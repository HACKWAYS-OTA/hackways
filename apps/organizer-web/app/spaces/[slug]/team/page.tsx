import { notFound } from 'next/navigation';
import type { WorkspaceTeam } from '@/types/api';
import { managedSpace } from '@/lib/workspace';
import { privateApi } from '@/lib/account';
import { TeamAccess } from '@/components/management/TeamAccess';

export const metadata = { title: 'Space team access' };
export default async function SpaceTeam({ params }: { params: { slug: string } }) {
  const { data: space, access } = await managedSpace(params.slug);
  if (!access.permissions.includes('access.manage')) notFound();
  const team = await privateApi<WorkspaceTeam>(`/workspace/spaces/${space.id}/access?limit=50`);
  return <TeamAccess kind="space" resourceID={space.id} initial={team} />;
}
