export type GlobalAnalytics = {
  period: { from: string; to: string };
  traffic: { page_views: number; visitors: number; sessions: number; tracked_registrations: number };
  traffic_daily: { day: string; page_views: number; visitors: number; tracked_registrations: number }[];
  pages: { route: string; views: number; visitors: number }[];
  sources: { source: string; views: number }[];
  devices: { device: string; views: number }[];
  accounts: { new_accounts: number; verified_new_accounts: number; accepted_emails: number; failed_emails: number; subscribed_contacts: number };
  account_daily: { day: string; new_accounts: number; accepted_emails: number }[];
  collection: { first_event: string; reserved_today: number };
  generated_at: string;
};
export type CoreAnalytics = {
  period: { from: string; to: string };
  summary: { registrations: number; confirmed: number; pending: number; waiting: number; other: number };
  events: { events_created: number; check_ins: number };
  daily: { day: string; registrations: number; confirmed: number; check_ins: number }[];
  top_events: { reference: number; title: string; slug: string; registrations: number; confirmed: number }[];
  generated_at: string;
};
export const analyticsNumber = (value: number) => new Intl.NumberFormat('en-IN').format(value);
export type AnalyticsWindow = '5m' | '1h' | '24h';
export type LiveCoreAnalytics = Pick<CoreAnalytics, 'summary' | 'events' | 'top_events'> & {
  window: AnalyticsWindow;
  captured_at: string;
  activity: { bucket: string; registrations: number; check_ins: number }[];
};
export type LiveGlobalAnalytics = Pick<GlobalAnalytics, 'traffic' | 'pages' | 'sources'> & {
  window: AnalyticsWindow;
  captured_at: string;
  limited: boolean;
  accounts: Pick<GlobalAnalytics['accounts'], 'new_accounts' | 'accepted_emails' | 'failed_emails'>;
};
export function analyticsPeriod(params: { from?: string; to?: string }) {
  const today = new Date().toISOString().slice(0, 10);
  const to = params.to || today;
  const parsedTo = Date.parse(`${to}T00:00:00Z`);
  const from = params.from || new Date((Number.isFinite(parsedTo) ? parsedTo : Date.parse(`${today}T00:00:00Z`)) - 29 * 86400000).toISOString().slice(0, 10);
  const start = Date.parse(`${from}T00:00:00Z`), end = Date.parse(`${to}T00:00:00Z`);
  const valid = /^\d{4}-\d{2}-\d{2}$/.test(from) && /^\d{4}-\d{2}-\d{2}$/.test(to) && Number.isFinite(start) && Number.isFinite(end) && new Date(start).toISOString().slice(0, 10) === from && new Date(end).toISOString().slice(0, 10) === to && from <= to && to <= today && start >= Date.parse(`${today}T00:00:00Z`) - 89 * 86400000 && end - start <= 89 * 86400000;
  return { from, to, today, valid, query: new URLSearchParams({ from, to }).toString() };
}
