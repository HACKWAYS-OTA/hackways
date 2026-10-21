import type { Event } from '@/types/api';
import { ApiClientError, fetchApi } from './api';

export const PAGE_SIZE = 24;
export interface EventPage { data: Event[]; pagination?: { limit: number; offset: number; count: number } }
export interface InitialEvents { events: Event[]; error: string | null; hasMore: boolean }

export async function listEvents(query = '', offset = 0, signal?: AbortSignal): Promise<EventPage> {
  const params = new URLSearchParams({ limit: String(PAGE_SIZE), offset: String(offset) });
  if (query.trim()) params.set('q', query.trim());
  const response = await fetchApi<EventPage>(`/events?${params}`, { signal, cache: 'no-store' });
  if (!Array.isArray(response.data)) throw new ApiClientError('The event service returned an unreadable response. Please try again.', 'INVALID_RESPONSE');
  return response;
}
export const formatLabels = { in_person: 'In person', online: 'Online', hybrid: 'Hybrid' } as const;
export function eventPath(event: Pick<Event, 'slug'>): string { return `/events/${encodeURIComponent(event.slug)}`; }
export function externalUrl(value: string | null | undefined): string | null {
  if (!value) return null;
  try { const url = new URL(value); return ['http:', 'https:'].includes(url.protocol) && !url.username && !url.password ? url.href : null; } catch { return null; }
}
export function eventZone(event: Event): string {
  try { new Intl.DateTimeFormat('en', { timeZone: event.timezone }).format(); return event.timezone; }
  catch { return 'UTC'; }
}
export function eventDate(event: Event, options: Intl.DateTimeFormatOptions): string {
  return new Intl.DateTimeFormat('en', { ...options, timeZone: eventZone(event) }).format(new Date(event.starts_at));
}
export function eventEndDate(event: Event, options: Intl.DateTimeFormatOptions): string {
  return new Intl.DateTimeFormat('en', { ...options, timeZone: eventZone(event) }).format(new Date(event.ends_at));
}
export function registrationState(event: Event): { label: string; disabled: boolean } {
  const now = Date.now();
  if (event.status !== 'published') return { label: 'Unavailable', disabled: true };
  if (new Date(event.ends_at).getTime() <= now) return { label: 'Event ended', disabled: true };
  if (new Date(event.starts_at).getTime() <= now) return { label: 'Registration closed', disabled: true };
  if (event.registration_closes_at && new Date(event.registration_closes_at).getTime() <= now) return { label: 'Registration closed', disabled: true };
  if (event.registration_opens_at && new Date(event.registration_opens_at).getTime() > now) return { label: 'Opening soon', disabled: true };
  if (event.registration_mode !== 'approval' && event.capacity != null && event.current_rsvp_count + (event.held_rsvp_count || 0) >= event.capacity) return event.waitlist_enabled ? { label: 'Join waiting list', disabled: false } : { label: 'Fully booked', disabled: true };
  return { label: 'View event', disabled: false };
}
