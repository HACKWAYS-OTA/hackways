import type { AdminModule } from './admin-modules';
export const PAGE_SIZE = 20;
export type CatalogueSearch = { q?: string | string[]; page?: string | string[]; status?: string | string[] };
export function catalogueParams(params: CatalogueSearch) {
  const query = typeof params.q === 'string' ? params.q.trim().slice(0, 200) : '';
  const status = typeof params.status === 'string' ? params.status.trim().slice(0, 60) : '';
  const parsed = typeof params.page === 'string' ? Number(params.page) : 1;
  const page = Number.isSafeInteger(parsed) && parsed > 0 && parsed <= Math.floor(2147483647 / PAGE_SIZE) ? parsed : 1;
  return { query, page, status };
}
export function cataloguePath(resource: AdminModule, query: string, page: number, status = '') {
  const params = new URLSearchParams();
  if (query) params.set('q', query);
  if (page > 1) params.set('page', String(page));
  if (status) params.set('status', status);
  const suffix = params.toString();
  return `/${resource}${suffix ? `?${suffix}` : ''}`;
}
