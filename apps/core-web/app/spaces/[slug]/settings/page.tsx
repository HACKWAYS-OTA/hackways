import { notFound } from 'next/navigation';
import type { Space, WorkspaceAccess } from '@/types/api';
import { requireUser, privateApi } from '@/lib/account';
import { SpaceForm } from '@/components/spaces/SpaceForm';
import '@/components/spaces/spaces.css';
export const metadata = { title: 'Space settings', robots: { index: false, follow: false } };
export default async function Settings({ params }: { params: { slug: string } }) {
  const user = await requireUser(`/spaces/${params.slug}/settings`);
  const result = await privateApi<{data: Space; access: WorkspaceAccess}>(`/spaces/manage-by-slug/${encodeURIComponent(params.slug)}`);
  if (!result.access.permissions.includes('content')) notFound();
  return <SpaceForm space={result.data} hostName={user.display_name} hasProfile canArchive={result.access.permissions.includes('delete')} managed />;
}
