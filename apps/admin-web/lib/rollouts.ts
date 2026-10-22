export type FeatureRule = { enabled: boolean; percentage: number; starts_at?: string; ends_at?: string };
export type RolloutConfiguration = { revision: number; version: number; label: string; flags: Record<string, FeatureRule>; updated_at: string };
export type RolloutVersion = { reference: number; label: string; flags: Record<string, FeatureRule>; restored_from: number | null; created_by: string; created_at: string };
export type RolloutData = { data: RolloutConfiguration; definitions: { key: string; name: string; surface: string }[]; history: RolloutVersion[]; has_next: boolean; source_commit: string; cache: { version: number; revision: number; age_seconds: number; fresh: boolean } };
