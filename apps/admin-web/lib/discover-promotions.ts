export type PromotionEvent = { reference: number; title: string; slug: string; event_starts_at: string; timezone: string };
export type PromotionSelection = PromotionEvent & { starts_at: string; ends_at: string; eligible?: boolean };
export type PromotionConfiguration = { revision: number; updated_at: string; selections: PromotionSelection[] };
