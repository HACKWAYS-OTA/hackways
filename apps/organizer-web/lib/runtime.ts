import 'server-only';
import { cookies } from 'next/headers';
import { fetchApi } from './api';
export async function runtimeFeatures(): Promise<Record<string, boolean>> {
  const cohort = cookies().get('__Host-hackways_rollout');
  try { return (await fetchApi<{ flags: Record<string, boolean> }>('/auth/runtime', { headers: cohort ? { Cookie: `${cohort.name}=${encodeURIComponent(cohort.value)}` } : {}, cache: 'no-store' })).flags; }
  catch { return {}; }
}
