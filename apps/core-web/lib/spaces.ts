import 'server-only';
import { cache } from 'react';
import { notFound } from 'next/navigation';
import type { Space } from '@/types/api';
import { ApiClientError, fetchApi } from './api';
export const readSpace = cache(async (slug: string): Promise<Space> => {
  if (!/^[a-z0-9-]{1,230}$/.test(slug)) notFound();
  try { return (await fetchApi<{ data: Space }>(`/spaces/by-slug/${encodeURIComponent(slug)}`, { cache: 'no-store' })).data; }
  catch (error) { if (error instanceof ApiClientError && error.code === 'NOT_FOUND') notFound(); throw error; }
});
