import { adminVerificationRequired, adminVerificationPath } from './admin-verification';
import 'server-only';
import { cookies } from 'next/headers';
import { notFound, redirect } from 'next/navigation';

export type Campaign = { reference: number; title: string; subject: string; body: string; action_label: string; action_url: string; status: string; revision: number; updated_at: string };
export type CampaignRow = Pick<Campaign, 'reference' | 'title' | 'subject' | 'status' | 'revision' | 'updated_at'>;
export type CampaignDetail = { data: Campaign; stats: { total: number; pending: number; accepted: number; failed: number; tests: number }; audience_count: number; audience: { reference: number; email: string; display_name: string; source: string; consented_at: string }[]; deliveries: { reference: number; recipient_email: string; status: string; retry_count: number; sent_at: string | null; is_test: boolean }[]; has_next: boolean; preview_html: string; sender: string };
export type PromotionContact = { reference: number; email: string; display_name: string; source: string; consent_source: string; consented_at: string; subscribed: boolean; unsubscribed_at: string | null };
export async function campaignServer<T>(path: string): Promise<T> {
  const session = cookies().get('__Host-hackways_admin_session');
  if (!session) redirect('/login?next=%2Fcampaigns');
  const base = process.env.GLOBAL_API_URL || (process.env.NODE_ENV === 'production' ? 'http://global-backend:8082/api/v1' : 'http://127.0.0.1:8082/api/v1');
  const controller = new AbortController(), timeout = setTimeout(() => controller.abort(), 7000);
  try {
    const response = await fetch(`${base}/admin/global${path}`, { headers: { Cookie: `${session.name}=${encodeURIComponent(session.value)}` }, cache: 'no-store', signal: controller.signal });
    if (response.status === 401) redirect('/login?next=%2Fcampaigns');
    if (response.status === 403) redirect(await adminVerificationRequired(response) ? adminVerificationPath('/campaigns') : '/access-denied');
    if (response.status === 404) notFound();
    if (!response.ok) throw new Error('Campaign data could not be loaded.');
    return await response.json() as T;
  } finally { clearTimeout(timeout); }
}
