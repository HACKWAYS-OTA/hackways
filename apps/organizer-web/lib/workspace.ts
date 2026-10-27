import 'server-only';
import { cache } from 'react';
import { notFound } from 'next/navigation';
import type { Space, WorkspaceAccess, WorkspaceRole } from '@/types/api';
import { privateApi, requireUser } from './account';
import { ApiClientError } from './api';

export const roleLabel: Record<WorkspaceRole, string> = { owner: 'Owner', manager: 'Manager', editor: 'Editor', checkin: 'Check-in staff' };
export const managedSpace = cache(async (slug: string) => {
  if (!/^[a-z0-9-]{1,230}$/.test(slug)) notFound();
  await requireUser(`/spaces/${slug}`);
  try {
    const result = await privateApi<{ data: Space; access: WorkspaceAccess }>(`/spaces/manage-by-slug/${slug}`);
    if (!result.access || !Array.isArray(result.access.permissions)) throw new Error('Workspace access could not be loaded. Try again.');
    return result;
  } catch (error) { if (error instanceof ApiClientError && error.code === 'NOT_FOUND') notFound(); throw error; }
});
