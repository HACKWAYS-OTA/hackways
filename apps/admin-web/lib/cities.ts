import type { CityHub } from '@hackways/api-client';
export type AdminCity = CityHub & { reference: number; revision: number; status: 'draft' | 'published' | 'archived' };
export type CityEvent = { id: string; title: string; slug: string; starts_at: string; status: string; visibility: string };
