CREATE TABLE space_access (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 space_id uuid NOT NULL REFERENCES spaces(id),
 user_id uuid NOT NULL REFERENCES users(id),
 role text NOT NULL CHECK (role IN ('manager','editor','checkin')),
 revision bigint NOT NULL DEFAULT 1 CHECK (revision>0),
 granted_by uuid NOT NULL REFERENCES users(id),
 updated_by uuid NOT NULL REFERENCES users(id),
 revoked_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(space_id,user_id)
);
CREATE INDEX space_access_user_idx ON space_access(user_id,space_id) WHERE revoked_at IS NULL;
CREATE TABLE event_access (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 event_id uuid NOT NULL REFERENCES events(id),
 user_id uuid NOT NULL REFERENCES users(id),
 role text NOT NULL CHECK (role IN ('manager','editor','checkin')),
 revision bigint NOT NULL DEFAULT 1 CHECK (revision>0),
 granted_by uuid NOT NULL REFERENCES users(id),
 updated_by uuid NOT NULL REFERENCES users(id),
 revoked_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(event_id,user_id)
);
CREATE INDEX event_access_user_idx ON event_access(user_id,event_id) WHERE revoked_at IS NULL;
