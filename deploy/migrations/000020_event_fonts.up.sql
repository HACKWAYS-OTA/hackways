ALTER TABLE events DROP CONSTRAINT events_theme_font_check;
ALTER TABLE events ADD CONSTRAINT events_theme_font_check CHECK (theme_font IN ('futurist','google','dm-sans','manrope','space-grotesk','outfit','sora','fraunces','playfair-display','lora','dm-mono','bebas-neue'));
