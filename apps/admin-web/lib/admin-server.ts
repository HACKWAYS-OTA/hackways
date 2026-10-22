import { adminVerificationRequired, adminVerificationPath } from './admin-verification';
import 'server-only';
import { cookies } from 'next/headers';
import { notFound, redirect } from 'next/navigation';

export async function adminServer<T>(service: 'core' | 'global', path: string): Promise<T> {
  const session = cookies().get('__Host-hackways_admin_session');
  if (!session) redirect('/login');
  const configured = service === 'core' ? process.env.CORE_API_URL : process.env.GLOBAL_API_URL;
  const base = configured || (process.env.NODE_ENV === 'production' ? `http://${service}-backend:${service === 'core' ? 8081 : 8082}/api/v1` : `http://127.0.0.1:${service === 'core' ? 8081 : 8082}/api/v1`);
  const controller = new AbortController(), timeout = setTimeout(() => controller.abort(), 7000);
  try {
    const response = await fetch(`${base.replace(/\/$/, '')}/admin/${service}${path}`, { headers: { Cookie: `${session.name}=${encodeURIComponent(session.value)}` }, cache: 'no-store', signal: controller.signal });
    if (response.status === 401) redirect('/login');
    if (response.status === 403) redirect(await adminVerificationRequired(response) ? adminVerificationPath('/') : '/access-denied');
    if (response.status === 404) notFound();
    if (!response.ok) throw new Error('Admin data could not be loaded. Try again.');
    return await response.json() as T;
  } finally { clearTimeout(timeout); }
}
