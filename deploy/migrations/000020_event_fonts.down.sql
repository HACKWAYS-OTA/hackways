-- Refuse rollback while new font choices are in use; never rewrite saved themes.
ALTER TABLE events ADD CONSTRAINT events_theme_font_legacy CHECK (theme_font IN ('futurist','google'));
ALTER TABLE events DROP CONSTRAINT events_theme_font_check;
ALTER TABLE events RENAME CONSTRAINT events_theme_font_legacy TO events_theme_font_check;
