import 'server-only';

import { cache } from 'react';
import { cookies } from 'next/headers';
import { adminModuleOrder } from './admin-modules';
import { verifyAdminAccess, type AdminAccess } from './admin-session';

export function safeAdminDestination(requested = '/') {
  if (requested.includes('\\') || requested.includes('#') || /[\u0000-\u001f\u007f]/.test(requested)) return '/';
  const path = requested.split('?')[0];
  return path === '/' || ['/discover-promotions','/rollouts','/campaigns','/campaigns/new','/campaigns/contacts'].includes(path) || /^\/campaigns\/details\/[1-9][0-9]*$/.test(path) || adminModuleOrder.some(name => path === `/${name}`
    || new RegExp(`^/${name}/details/[1-9][0-9]*$`).test(path)
    || name !== 'audit' && new RegExp(`^/${name}/details/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`).test(path)) ? requested : '/';
}

// React cache deduplicates within this render only. Permissions are never shared
// between requests; the Go service verifies the live session and database grant.
export const getAdminAccess = cache(async (): Promise<AdminAccess> => {
  const cookieStore = cookies();
  const sessionCookies = ['__Host-hackways_admin_session', '__Host-hackways_session']
    .map(name => cookieStore.get(name))
    .filter((cookie): cookie is { name: string; value: string } => Boolean(cookie?.value));
  return verifyAdminAccess(sessionCookies);
});
