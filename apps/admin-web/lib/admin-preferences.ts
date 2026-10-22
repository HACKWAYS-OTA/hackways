export type AdminPreferences = { text: 'standard' | 'large'; density: 'comfortable' | 'compact'; contrast: 'standard' | 'high'; motion: 'system' | 'reduced'; scrollbars: 'hidden' | 'visible' };
export const defaultPreferences: AdminPreferences = { text: 'standard', density: 'comfortable', contrast: 'standard', motion: 'system', scrollbars: 'hidden' };
export function parsePreferences(value?: string): AdminPreferences {
  try { const data: unknown = JSON.parse(value || '{}'); if (!data || typeof data !== 'object') return defaultPreferences; return { text: 'text' in data && data.text === 'large' ? 'large' : 'standard', density: 'density' in data && data.density === 'compact' ? 'compact' : 'comfortable', contrast: 'contrast' in data && data.contrast === 'high' ? 'high' : 'standard', motion: 'motion' in data && data.motion === 'reduced' ? 'reduced' : 'system', scrollbars: 'scrollbars' in data && data.scrollbars === 'visible' ? 'visible' : 'hidden' }; }
  catch { return defaultPreferences; }
}
