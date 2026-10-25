import type { EditableGuest } from './guests';
import { fetchApi } from './api';

export type GuestDecision = 'confirmed' | 'rejected';
export type DecisionAttempt = { body: string; key: string };

export function decisionBody(guest: EditableGuest, status: GuestDecision, notify: boolean, message = ''): string {
  return JSON.stringify({ status, updated_at: guest.updated_at, notify_guest: notify, message: notify ? message.trim() : '' });
}

/** UI transport only: Go owns permissions, revision checks, inventory and notifications. */
export async function saveGuestDecision(eventId: string, guest: Pick<EditableGuest, 'id' | 'public_code'>, status: GuestDecision, attempt: DecisionAttempt) {
  const result = await fetchApi<{ data: { id: string; status: string }; notification_queued: boolean }>(
    `/events/${eventId}/registrations/by-code/${encodeURIComponent(guest.public_code)}/status`,
    { method: 'POST', body: attempt.body, headers: { 'Idempotency-Key': attempt.key } },
  );
  if (result.data.id !== guest.id || result.data.status !== status) throw new Error('The decision could not be confirmed. Please retry.');
  return `${status === 'confirmed' ? 'Guest approved.' : 'Application declined.'}${result.notification_queued ? ' An email update is on its way.' : ''}`;
}
