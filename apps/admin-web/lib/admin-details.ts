import { adminVerificationRequired, adminVerificationPath } from './admin-verification';
import 'server-only';

import { cookies } from 'next/headers';
import { redirect, notFound } from 'next/navigation';
import { adminModules, type AdminModule, type AdminRecord } from './admin-modules';
import { relatedPath } from './admin-detail-paths';

export type RelatedRecord = { id: string; reference: string; label: string; detail: string; resource: string; status: string; occurred_at: string | null };
export type AdminDetail = { data: AdminRecord; related: { tabs: { key: string; label: string }[]; section: string; rows: RelatedRecord[]; offset: number; has_next: boolean } };

function object(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isDetail(value: unknown): value is AdminDetail {
  if (!object(value) || !object(value.data) || !object(value.related)) return false;
  if (!['string', 'number'].includes(typeof value.data.id)) return false;
  if (typeof value.data.reference !== 'number' || !Number.isSafeInteger(value.data.reference) || value.data.reference < 1) return false;
  if (!Object.values(value.data).every(cell => cell === null || ['string', 'number', 'boolean'].includes(typeof cell))) return false;
  const related = value.related;
  return Array.isArray(related.tabs) && related.tabs.every(tab => object(tab) && typeof tab.key === 'string' && typeof tab.label === 'string')
    && typeof related.section === 'string' && typeof related.offset === 'number' && typeof related.has_next === 'boolean'
    && Array.isArray(related.rows) && related.rows.every(row => object(row)
      && ['id', 'reference', 'label', 'detail', 'resource', 'status'].every(key => typeof row[key] === 'string')
      && /^[1-9][0-9]*$/.test(String(row.reference))
      && (row.occurred_at === null || typeof row.occurred_at === 'string'));
}

export async function getAdminDetail(resource: AdminModule, id: string, section: string, offset: number, kind: string): Promise<AdminDetail | null> {
  const engine = adminModules[resource].engine;
  const base = (engine === 'core' ? process.env.CORE_API_URL : process.env.GLOBAL_API_URL)
    || (process.env.NODE_ENV === 'development' ? `http://127.0.0.1:${engine === 'core' ? 8081 : 8082}/api/v1` : `http://${engine}-backend:${engine === 'core' ? 8081 : 8082}/api/v1`);
  const session = cookies().get('__Host-hackways_admin_session');
  const destination = relatedPath(resource, id, section, offset, kind);
  if (!session) redirect(`/login?next=${encodeURIComponent(destination)}`);
  const params = new URLSearchParams({ offset: String(offset) });
  if (section) params.set('section', section);
  if (kind) params.set('kind', kind);
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 6000);
  let response: Response;
  let body: unknown = null;
  try {
    response = await fetch(`${base.replace(/\/$/, '')}/admin/${engine}/${resource}/${encodeURIComponent(id)}?${params}`, {
      headers: { Cookie: `${session.name}=${session.value}`, Accept: 'application/json' },
      cache: 'no-store', redirect: 'error', signal: controller.signal,
    });
    if (response.ok) body = await response.json();
  } catch {
    return null;
  } finally {
    clearTimeout(timeout);
  }
  if (response.status === 401) redirect(`/login?next=${encodeURIComponent(destination)}`);
  if (response.status === 403) redirect(await adminVerificationRequired(response) ? adminVerificationPath(destination) : '/access-denied');
  if (response.status === 404) notFound();
  if (!response.ok) return null;
  return isDetail(body) ? body : null;
}
