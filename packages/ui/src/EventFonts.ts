export const eventFonts = [
  { value: 'futurist', label: 'Futurist', family: 'FG Futurist' },
  { value: 'google', label: 'Google Sans', family: 'Google Sans Flex' },
  { value: 'dm-sans', label: 'DM Sans', family: 'DM Sans' },
  { value: 'manrope', label: 'Manrope', family: 'Manrope' },
  { value: 'space-grotesk', label: 'Space Grotesk', family: 'Space Grotesk' },
  { value: 'outfit', label: 'Outfit', family: 'Outfit' },
  { value: 'sora', label: 'Sora', family: 'Sora' },
  { value: 'fraunces', label: 'Fraunces', family: 'Fraunces' },
  { value: 'playfair-display', label: 'Playfair Display', family: 'Playfair Display' },
  { value: 'lora', label: 'Lora', family: 'Lora' },
  { value: 'dm-mono', label: 'DM Mono', family: 'DM Mono' },
  { value: 'bebas-neue', label: 'Bebas Neue', family: 'Bebas Neue' },
] as const;
export type EventFont = typeof eventFonts[number]['value'];
export function eventFontFamily(value: string) { return '"' + (eventFonts.find(font => font.value === value)?.family || 'FG Futurist') + '", sans-serif'; }
