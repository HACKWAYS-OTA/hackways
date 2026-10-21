import type { Event } from '@/types/api';
export interface GuestSummary { total: number; confirmed: number; pending: number; waitlisted: number; offered: number; cancelled: number; rejected: number; expired: number; transferred: number; checked_in: number }
export interface Activity { id: number; action: string; created_at: string }
export const hostViews = [['all', 'All events'], ['upcoming', 'Upcoming'], ['drafts', 'Drafts'], ['past', 'Past'], ['cancelled', 'Cancelled']] as const;
export function eventStatus(event: Event): string {
  if (event.status === 'cancelled') return 'Cancelled';
  if (event.status === 'draft') return 'Draft';
  if (Date.parse(event.ends_at) <= Date.now()) return 'Ended';
  if (event.status === 'unpublished') return 'Registration paused';
  return event.visibility === 'private' ? 'Link only' : 'Published';
}
export function safeLink(value?: string | null): string | null {
  try { const url = new URL(value || ''); return url.protocol === 'https:' && !url.username && !url.password ? url.href : null; } catch { return null; }
}
const actions: Record<string, string> = {
 'event.registration_form_updated': 'Registration questions updated',
 'registration.confirmed': 'A guest registered', 'registration.pending_approval': 'An application arrived', 'registration.waitlisted': 'A guest joined the waiting list', 'registration.cancel': 'Registration cancelled', 'registration.accept': 'An offered place was accepted', 'registration.approve': 'Application approved', 'registration.reject': 'Application declined', 'registration.offered': 'A place was offered', 'registration.offer_expired': 'An offered place expired', 'registration.approval_expired': 'An application expired', 'registration.transferred': 'Registration transferred',
 'event.appearance_updated': 'Cover or theme updated', 'registration.claimed': 'Registration linked to an account',
 'event.created': 'Event created', 'event.updated': 'Event details updated', 'event.published': 'Event published', 'event.unpublished': 'Registration paused', 'event.cancelled': 'Event cancelled', 'event.space_changed': 'Event Space changed',
 'ticket.scanned': 'Guest checked in', 'ticket.manual_check_in': 'Guest checked in manually', 'rsvp.confirmed': 'Registration confirmed', 'rsvp.cancelled': 'Registration cancelled', 'rsvp.approved': 'Application approved', 'rsvp.rejected': 'Application declined', 'rsvp.waitlisted': 'Guest joined the waiting list', 'rsvp.offered': 'A place was offered', 'rsvp.transferred': 'Registration transferred', 'rsvp.claimed': 'Guest registration linked to an account',
};
export function activityLabel(action: string): string { return actions[action] || action.replaceAll('.', ' · ').replaceAll('_', ' '); }
export function queryPath(path: string, values: Record<string, string | number>): string {
  const query = new URLSearchParams();
  Object.entries(values).forEach(([key, value]) => { if (String(value)) query.set(key, String(value)); });
  return `${path}?${query}`;
}
