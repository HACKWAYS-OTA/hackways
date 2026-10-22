CREATE TABLE runtime_rollout_versions (
    reference bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    label varchar(80) NOT NULL,
    flags jsonb NOT NULL CHECK (jsonb_typeof(flags)='object' AND octet_length(flags::text)<=65536),
    restored_from bigint REFERENCES runtime_rollout_versions(reference),
    created_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
INSERT INTO runtime_rollout_versions(label,flags) VALUES ('Initial configuration',
'{"discover.promotions":{"enabled":true,"percentage":100},"manage.overview.v2":{"enabled":true,"percentage":100}}');
CREATE TABLE runtime_rollout_control (
    singleton boolean PRIMARY KEY DEFAULT true CHECK(singleton),
    id uuid NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    active_version bigint NOT NULL REFERENCES runtime_rollout_versions(reference),
    cohort_salt bytea NOT NULL DEFAULT gen_random_bytes(32) CHECK(octet_length(cohort_salt)=32),
    revision bigint NOT NULL DEFAULT 1 CHECK(revision>0),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
INSERT INTO runtime_rollout_control(singleton,active_version) SELECT true,reference FROM runtime_rollout_versions ORDER BY reference LIMIT 1;
