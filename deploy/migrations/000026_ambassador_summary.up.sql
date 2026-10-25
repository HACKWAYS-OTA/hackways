-- Update only the original default; preserve administrator-authored program copy.
UPDATE programs SET summary='Host events that bring your community together, using Hackways.',
  revision=revision+1, updated_at=clock_timestamp()
WHERE slug='ambassadors' AND summary='Bring your community together. Open to anyone aged 18 or older.';

UPDATE programs SET recognition='Your contributions are recognised with a letter of appreciation, with swag for the highest contributors.',
  revision=revision+1, updated_at=clock_timestamp()
WHERE slug='ambassadors' AND recognition='A letter of appreciation recognizes your work. The highest contributors receive swag. No additional benefits are currently announced.';
