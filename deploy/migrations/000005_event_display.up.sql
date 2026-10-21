ALTER TABLE events ADD COLUMN theme_display TEXT NOT NULL DEFAULT 'light' CHECK(theme_display IN ('light','dark'));
