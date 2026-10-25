import 'server-only';
import { cache } from 'react';
import { notFound } from 'next/navigation';
import { ApiClientError, fetchApi } from './api';
import { privateApi } from './account';
import type { Program, ProgramMine } from '@hackways/api-client';
export const programRead = cache(async (slug: string) => {
  try { return await fetchApi<{ data: Program; accepting_applications: boolean }>(`/programs/${encodeURIComponent(slug)}`, { cache: 'no-store' }); }
  catch (error) { if (error instanceof ApiClientError && error.code === 'NOT_FOUND') notFound(); throw error; }
});
export async function programMine(slug: string): Promise<ProgramMine> {
  try { return await privateApi<ProgramMine>(`/programs/${encodeURIComponent(slug)}/application`); }
  catch (error) { if (error instanceof ApiClientError && error.code === 'NOT_FOUND') notFound(); throw error; }
}
