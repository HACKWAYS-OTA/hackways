import { eventFonts, eventFontFamily, type EventFont } from '@hackways/ui/event-fonts';
import type { CSSProperties } from 'react';
import type { Event } from '@/types/api';
export const eventThemes = [
  { value: 'minimal', label: 'Minimal' }, { value: 'aurora', label: 'Aurora' },
  { value: 'ambient', label: 'Ambient' }, { value: 'paper', label: 'Paper' },
  { value: 'lilac', label: 'Lilac' }, { value: 'confetti', label: 'Confetti' },
  { value: 'pattern', label: 'Pattern' }, { value: 'playful', label: 'Playful' },
  { value: 'seasonal', label: 'Seasonal' },
] as const;
export interface Appearance { theme: typeof eventThemes[number]['value']; theme_color: string; theme_font: EventFont; theme_style: 'soft' | 'crisp'; theme_display: 'light' | 'dark' }
export const defaultAppearance: Appearance = { theme: 'minimal', theme_color: '#ddd4e9', theme_font: 'futurist', theme_style: 'soft', theme_display: 'light' };
export function appearanceOf(event: Partial<Event>): Appearance {
  return { theme: eventThemes.find(item => item.value === event.theme)?.value || 'minimal', theme_color: /^#[0-9a-f]{6}$/i.test(event.theme_color || '') ? event.theme_color! : defaultAppearance.theme_color, theme_font: eventFonts.find(font => font.value === event.theme_font)?.value || 'futurist', theme_style: event.theme_style === 'crisp' ? 'crisp' : 'soft', theme_display: event.theme_display === 'dark' ? 'dark' : 'light' };
}
interface ThemeVariables extends CSSProperties { '--event-tint': string; '--event-heading': string; '--event-radius': string }
export function appearanceStyle(appearance: Appearance): ThemeVariables {
  return { '--event-tint': appearance.theme_color, '--event-heading': eventFontFamily(appearance.theme_font), '--event-radius': appearance.theme_style === 'crisp' ? '8px' : '16px' };
}
