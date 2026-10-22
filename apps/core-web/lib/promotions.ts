import type { Event } from '@/types/api';
import { fetchApi } from './api';

export type PromotionFeed = { data: { events: Event; slot: number; visible_until: string }[]; next_change_at: string | null; server_time: string };
export async function listPromotions(signal?: AbortSignal): Promise<PromotionFeed> {
  return fetchApi<PromotionFeed>('/events/promotions', { signal, cache: 'no-store' });
}
export async function activePromotions(signal?: AbortSignal): Promise<PromotionFeed> {
  const [feed, runtime] = await Promise.all([fetchApi<PromotionFeed>('/events/promotions', { signal, cache: 'no-store' }), fetchApi<{ flags: Record<string, boolean> }>('/auth/runtime', { signal, cache: 'no-store' })]);
  return runtime.flags['discover.promotions'] ? feed : { ...feed, data: [] };
}
