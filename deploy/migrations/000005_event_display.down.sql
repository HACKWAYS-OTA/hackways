DO $$ BEGIN RAISE EXCEPTION 'Preserve saved appearance. Roll back application code without dropping theme_display.'; END $$;
