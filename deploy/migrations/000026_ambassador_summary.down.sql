UPDATE programs SET summary='Bring your community together. Open to anyone aged 18 or older.',
  revision=revision+1, updated_at=clock_timestamp()
WHERE slug='ambassadors' AND summary='Host events that bring your community together, using Hackways.';

UPDATE programs SET recognition='A letter of appreciation recognizes your work. The highest contributors receive swag. No additional benefits are currently announced.',
  revision=revision+1, updated_at=clock_timestamp()
WHERE slug='ambassadors' AND recognition='Your contributions are recognised with a letter of appreciation, with swag for the highest contributors.';
