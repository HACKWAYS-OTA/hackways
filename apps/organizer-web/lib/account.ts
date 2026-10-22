import 'server-only';
import { cache } from 'react';
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import type { User } from '@/types/api';
import { ApiClientError, fetchApi } from './api';

export function privateApi<T>(path: string): Promise<T> {
  const session = cookies().get('__Host-hackways_session');
  return fetchApi<T>(path, { cache: 'no-store', headers: session ? { Cookie: `${session.name}=${encodeURIComponent(session.value)}` } : {} });
}
export const currentUser = cache(async (): Promise<User | null> => {
  if (!cookies().has('__Host-hackways_session')) return null;
  try { return (await privateApi<{ user: User }>('/auth/me')).user; }
  catch (error) { if (error instanceof ApiClientError && error.code === 'UNAUTHORIZED') return null; throw error; }
});
export async function requireUser(next: string): Promise<User> {
  const user = await currentUser();
  if (!user) redirect(`/login?next=${encodeURIComponent(next)}`);
  return user;
}
export function pageOffset(value: string | string[] | undefined): number {
  const n = typeof value === 'string' ? Number(value) : 0;
  return Number.isSafeInteger(n) && n >= 0 && n <= 100000 ? n : 0;
}
