"use client";

import { useEffect, useState } from 'react';
import { adminModules, type AdminModule, type AdminRecord } from './admin-modules';
import { PAGE_SIZE, cataloguePath } from './catalogue-paths';

export function useCatalogue(resource: AdminModule, query: string, page: number, status: string) {
  const [attempt, setAttempt] = useState(0);
  const [state, setState] = useState<{ rows: AdminRecord[]; loading: boolean; error: string | null }>({ rows: [], loading: true, error: null });
  useEffect(() => {
    const controller = new AbortController();
    let active = true;
    const timeout = window.setTimeout(() => controller.abort(), 15000);
    const params = new URLSearchParams({ limit: String(PAGE_SIZE), offset: String((page - 1) * PAGE_SIZE) });
    if (query) params.set('q', query);
    if (status) params.set('status', status);
    setState({ rows: [], loading: true, error: null });
    void (async () => {
      try {
        const base = (process.env.NEXT_PUBLIC_API_URL || '/api/v1').replace(/\/$/, '');
        const response = await fetch(`${base}/admin/${adminModules[resource].engine}/${resource}?${params}`, { credentials: 'include', cache: 'no-store', signal: controller.signal });
        if (response.status === 401 || response.status === 403) {
          if (active) {
            setState({ rows: [], loading: true, error: null });
            window.location.replace(response.status === 403 ? '/access-denied'
              : `/login?next=${encodeURIComponent(cataloguePath(resource, query, page, status))}`);
          }
          return;
        }
        if (!response.ok) throw new Error('Request failed');
        const body = await response.json() as { data?: AdminRecord[] };
        if (!Array.isArray(body.data)) throw new Error('Missing collection');
        if (active) setState({ rows: body.data, loading: false, error: null });
      } catch {
        if (active) setState({ rows: [], loading: false, error: 'These records could not be loaded.' });
      } finally { window.clearTimeout(timeout); }
    })();
    return () => { active = false; window.clearTimeout(timeout); controller.abort(); };
  }, [resource, query, page, status, attempt]);
  return { ...state, refresh: () => setAttempt(value => value + 1) };
}
