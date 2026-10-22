'use client';
import { adminVerificationRequired, adminVerificationPath } from './admin-verification';
import type { Campaign } from './campaigns';
import type { CampaignConfiguration, CampaignDesign } from './campaign-design';
export type CampaignValues = Record<string, string | number | boolean | CampaignConfiguration | CampaignDesign | string[]>;
export async function campaignRequest(path: string, values: CampaignValues, reason = '', key?: string, signal?: AbortSignal) {
  const base = (process.env.NEXT_PUBLIC_API_URL || '/api/v1').replace(/\/$/, '');
  const controller = new AbortController(), timeout = window.setTimeout(() => controller.abort(), 15000);
  const abort = () => controller.abort(); signal?.addEventListener('abort', abort, { once: true });
  try {
    if (signal?.aborted) controller.abort();
    const response = await fetch(`${base}/admin/global${path}`, { method: 'POST', credentials: 'include', headers: {
      'Content-Type': 'application/json', ...(reason ? { 'X-Admin-Reason': encodeURIComponent(reason) } : {}), ...(key ? { 'Idempotency-Key': key } : {}),
    }, body: JSON.stringify(values), signal: controller.signal });
    if (response.status === 401 || response.status === 403) {
      window.location.assign(response.status === 401 ? '/login?next=%2Fcampaigns' : await adminVerificationRequired(response) ? adminVerificationPath('/campaigns') : '/access-denied');
      throw new Error('Admin sign-in is required.');
    }
    const result = await response.json() as { data?: Campaign; error?: { message?: string }; queued?: number; imported?: number; unchanged?: number; html?: string; audience_count?: number };
    if (!response.ok) throw new Error(result.error?.message || 'The action could not be saved. Your input is still here.');
    return result;
  } finally { window.clearTimeout(timeout); signal?.removeEventListener('abort', abort); }
}
