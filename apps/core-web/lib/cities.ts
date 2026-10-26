import 'server-only';
import { cache } from 'react';
import { notFound } from 'next/navigation';
import type { CityHub } from '@hackways/api-client';
import { fetchApi, ApiClientError } from './api';
export const readCity = cache(async (slug: string): Promise<CityHub> => {
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(slug) || slug.length > 100) notFound();
  try { return (await fetchApi<{ data: CityHub }>(`/cities/by-slug/${encodeURIComponent(slug)}`, { cache: 'no-store' })).data; }
  catch (error) { if (error instanceof ApiClientError && ['NOT_FOUND', 'HTTP_404'].includes(error.code)) notFound(); throw error; }
});
