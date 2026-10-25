DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM program_applications) OR EXISTS(SELECT 1 FROM program_activity) OR EXISTS(SELECT 1 FROM program_releases) THEN
  RAISE EXCEPTION 'Program data exists; preserve it and use a forward migration';
 END IF;
END $$;
DROP TABLE program_activity;
DROP TABLE program_releases;
DROP TABLE program_applications;
DROP TABLE programs;
