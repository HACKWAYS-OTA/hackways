import type { RegistrationStatus } from '@/types/api';
export const registrationLabels: Record<RegistrationStatus, string> = {
  confirmed: 'Going', cancelled: 'Cancelled', waitlisted: 'On the waiting list', pending_approval: 'Awaiting approval', offered: 'A place is available', rejected: 'Not approved', expired: 'Expired', transferred: 'Transferred',
};
export interface MyRegistration {
  id: string; public_code: string; event_id: string; status: RegistrationStatus; offer_expires_at: string | null;
  registered_at: string; title: string; slug: string; starts_at: string; ends_at: string; timezone: string; event_status: string;
}
export function safeNext(value: string | undefined, fallback = '/my-events'): string {
  if (!value || !value.startsWith('/') || value.startsWith('//') || /[\\\u0000-\u0020]/.test(value)) return fallback;
  const url = new URL(value, 'https://local.invalid');
  return url.origin === 'https://local.invalid' && !['/login', '/onboarding'].includes(url.pathname) ? url.pathname + url.search : fallback;
}
export function dateLabel(value: string, timezone: string, options: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' }): string {
  return new Intl.DateTimeFormat('en', { ...options, timeZone: timezone }).format(new Date(value));
}
