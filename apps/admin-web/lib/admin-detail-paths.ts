import { adminModuleOrder, type AdminModule } from './admin-modules';

export function isAdminModule(value: string): value is AdminModule {
  return adminModuleOrder.some(module => module === value);
}

export function detailPath(resource: AdminModule, id: string | number, kind?: string) {
  const path = `/${resource}/details/${encodeURIComponent(String(id))}`;
  return resource === 'storage' && kind ? `${path}?kind=${encodeURIComponent(kind)}` : path;
}

export function relatedPath(resource: AdminModule, id: string | number, section: string, offset: number, kind?: string) {
  const params = new URLSearchParams({ section, offset: String(offset) });
  if (kind) params.set('kind', kind);
  return `${detailPath(resource, id)}?${params}`;
}
