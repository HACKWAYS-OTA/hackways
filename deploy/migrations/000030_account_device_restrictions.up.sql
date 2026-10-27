CREATE TABLE device_restrictions (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 reference bigint GENERATED ALWAYS AS IDENTITY UNIQUE NOT NULL,
 token_hash text UNIQUE NOT NULL CHECK(token_hash ~ '^[a-f0-9]{64}$'),
 user_agent text NOT NULL DEFAULT '' CHECK(length(user_agent)<=512),
 blocked boolean NOT NULL DEFAULT true,
 revision bigint NOT NULL DEFAULT 1 CHECK(revision>0),
 reason text NOT NULL CHECK(length(reason) BETWEEN 3 AND 1000),
 updated_by uuid NOT NULL REFERENCES users(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX device_restrictions_blocked_idx ON device_restrictions(updated_at DESC,id) WHERE blocked;
CREATE TABLE account_restrictions (
 user_id uuid PRIMARY KEY REFERENCES users(id),
 blocked boolean NOT NULL DEFAULT true,
 revision bigint NOT NULL DEFAULT 1 CHECK(revision>0),
 reason text NOT NULL CHECK(length(reason) BETWEEN 3 AND 1000),
 source_device_id uuid REFERENCES device_restrictions(id),
 updated_by uuid REFERENCES users(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX account_restrictions_blocked_idx ON account_restrictions(updated_at DESC,user_id) WHERE blocked;
CREATE TABLE restriction_device_accounts (
 device_id uuid NOT NULL REFERENCES device_restrictions(id),
 user_id uuid NOT NULL REFERENCES users(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(device_id,user_id)
);
CREATE INDEX restriction_device_accounts_user_idx ON restriction_device_accounts(user_id,device_id);
CREATE INDEX recognized_devices_token_idx ON recognized_devices(token_hash,user_id);
ALTER TABLE recognized_devices ADD COLUMN admin_reference bigint GENERATED ALWAYS AS IDENTITY UNIQUE NOT NULL;
ALTER TABLE sessions ADD COLUMN device_token_hash text CHECK(device_token_hash IS NULL OR device_token_hash ~ '^[a-f0-9]{64}$'),
 ADD COLUMN revoked_at timestamptz;
CREATE INDEX sessions_device_idx ON sessions(device_token_hash) WHERE device_token_hash IS NOT NULL AND revoked_at IS NULL;

ALTER TABLE oauth_states ADD COLUMN device_token_hash text CHECK(device_token_hash IS NULL OR device_token_hash ~ '^[a-f0-9]{64}$');
ALTER TABLE oauth_exchange_tokens ADD COLUMN device_token_hash text CHECK(device_token_hash IS NULL OR device_token_hash ~ '^[a-f0-9]{64}$'),
 ADD COLUMN callback_device_token_hash text CHECK(callback_device_token_hash IS NULL OR callback_device_token_hash ~ '^[a-f0-9]{64}$');

COMMENT ON TABLE recognized_devices IS 'Browser recognition for notification deduplication and moderation association; never an authentication factor or hardware identity.';
