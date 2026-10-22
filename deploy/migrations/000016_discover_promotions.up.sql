CREATE TABLE discover_promotion_settings (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    id uuid NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    updated_by uuid REFERENCES users(id),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
INSERT INTO discover_promotion_settings(singleton) VALUES (true);

CREATE TABLE discover_promotions (
    slot smallint PRIMARY KEY CHECK (slot BETWEEN 1 AND 5),
    event_id uuid NOT NULL UNIQUE REFERENCES events(id),
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    CHECK (ends_at > starts_at)
);
