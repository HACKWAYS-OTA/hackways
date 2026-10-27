import 'server-only';
import { cache } from 'react';
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import type { User } from '@/types/api';
import { ApiClientError, fetchApi } from './api';

export function privateApi<T>(path: string): Promise<T> {
  const transport = ['__Host-hackways_session','__Host-hackways_device'].flatMap(name => {const cookie=cookies().get(name);return cookie ? [`${cookie.name}=${encodeURIComponent(cookie.value)}`] : [];}).join('; ');
  return fetchApi<T>(path, { cache: 'no-store', headers: transport ? { Cookie: transport } : {} });
}
const accountState = cache(async (): Promise<{ user: User | null; restricted: boolean }> => {
  if (!cookies().has('__Host-hackways_session')) return {user:null,restricted:false};
  try { return {user:(await privateApi<{ user: User }>('/auth/me')).user,restricted:false}; }
  catch (error) {
    if (error instanceof ApiClientError && error.code === 'UNAUTHORIZED') return {user:null,restricted:false};
    if (error instanceof ApiClientError && error.code === 'ACCOUNT_RESTRICTED') return {user:null,restricted:true};
    throw error;
  }
});
export const currentUser = cache(async (): Promise<User | null> => (await accountState()).user);
export async function requireUser(next: string): Promise<User> {
  const state=await accountState();
  if (state.restricted) redirect('/not-allowed');
  if (!state.user) redirect(`/login?next=${encodeURIComponent(next)}`);
  return state.user;
}
export function pageOffset(value: string | string[] | undefined): number {
  const n = typeof value === 'string' ? Number(value) : 0;
  return Number.isSafeInteger(n) && n >= 0 && n <= 100000 ? n : 0;
}
