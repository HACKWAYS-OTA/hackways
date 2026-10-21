import 'server-only';
import { cache } from 'react';
import { notFound } from 'next/navigation';
import type { Event } from '@/types/api';
import { requireUser, privateApi } from './account';
import { ApiClientError } from './api';
export type ManagedEvent = Event & { revision: number; held_rsvp_count: number; registration_opens_at: string | null; registration_closes_at: string | null; registration_mode: 'instant' | 'approval'; allow_guest: boolean; waitlist_enabled: boolean; offer_minutes: number };
export const managedEvent = cache(async (slug: string) => {
  if (!/^[a-z0-9-]{1,230}$/.test(slug)) notFound();
  await requireUser(`/events/${slug}`);
  try { return (await privateApi<{ data: ManagedEvent }>(`/events/manage-by-slug/${slug}`)).data; }
  catch (error) { if (error instanceof ApiClientError && ['NOT_FOUND', 'FORBIDDEN'].includes(error.code)) notFound(); throw error; }
});
