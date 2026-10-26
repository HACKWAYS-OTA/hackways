CREATE TABLE city_hubs (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 reference bigint GENERATED ALWAYS AS IDENTITY UNIQUE NOT NULL,
 slug text UNIQUE NOT NULL CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(slug)<=100),
 name text NOT NULL CHECK (length(name) BETWEEN 1 AND 120),
 region text NOT NULL DEFAULT '' CHECK (length(region)<=120),
 country text NOT NULL CHECK (length(country) BETWEEN 1 AND 120),
 timezone text NOT NULL CHECK (length(timezone) BETWEEN 1 AND 100),
 description text NOT NULL DEFAULT '' CHECK (length(description)<=6000),
 banner_url text,
 status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','archived')),
 revision bigint NOT NULL DEFAULT 1 CHECK (revision>0),
 created_by uuid NOT NULL REFERENCES users(id),
 creation_key text NOT NULL, creation_hash text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(created_by,creation_key)
);
CREATE INDEX city_hubs_public_idx ON city_hubs(status,name,id);
CREATE TABLE city_hub_events (
 city_id uuid NOT NULL REFERENCES city_hubs(id),
 event_id uuid NOT NULL REFERENCES events(id),
 added_by uuid NOT NULL REFERENCES users(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(city_id,event_id)
);
CREATE INDEX city_hub_events_event_idx ON city_hub_events(event_id);
