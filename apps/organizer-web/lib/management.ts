import 'server-only';
import { cache } from 'react';
import { notFound } from 'next/navigation';
import type { Event, WorkspaceAccess, WorkspacePermission } from '@/types/api';
import { requireUser, privateApi } from './account';
import { ApiClientError } from './api';
export type ManagedEvent = Event & { management_closed: boolean; revision: number; held_rsvp_count: number; registration_opens_at: string | null; registration_closes_at: string | null; registration_mode: 'instant' | 'approval'; allow_guest: boolean; waitlist_enabled: boolean; offer_minutes: number };
export const managedEventContext = cache(async (slug: string) => {
  if (!/^[a-z0-9-]{1,230}$/.test(slug)) notFound();
  await requireUser(`/events/${slug}`);
  try {
    const result = await privateApi<{ data: ManagedEvent; access: WorkspaceAccess }>(`/events/manage-by-slug/${slug}`);
    if (!result.access || !Array.isArray(result.access.permissions)) throw new Error('Workspace access could not be loaded. Try again.');
    return result;
  }
  catch (error) { if (error instanceof ApiClientError && ['NOT_FOUND', 'FORBIDDEN'].includes(error.code)) notFound(); throw error; }
});
export const managedEvent = cache(async (slug: string) => (await managedEventContext(slug)).data);
export async function requireEventPermission(slug: string, permission: WorkspacePermission) { const result = await managedEventContext(slug); if (!result.access.permissions.includes(permission)) notFound(); return result; }
