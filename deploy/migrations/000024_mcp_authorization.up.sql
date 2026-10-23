CREATE TABLE mcp_oauth_clients (
 id TEXT PRIMARY KEY CHECK(length(id) BETWEEN 16 AND 2048),
 name TEXT NOT NULL CHECK(length(name) BETWEEN 1 AND 100),
 redirect_uris TEXT[] NOT NULL CHECK(cardinality(redirect_uris) BETWEEN 1 AND 10),
 metadata_url TEXT NOT NULL DEFAULT '' CHECK(length(metadata_url)<=2048),
 verified BOOLEAN NOT NULL DEFAULT FALSE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TABLE mcp_oauth_requests (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),client_id TEXT NOT NULL REFERENCES mcp_oauth_clients(id),
 redirect_uri TEXT NOT NULL CHECK(length(redirect_uri)<=2048),state TEXT NOT NULL CHECK(length(state)<=2048),
 resource TEXT NOT NULL CHECK(length(resource)<=300),code_challenge TEXT NOT NULL CHECK(length(code_challenge)=43),
 scopes TEXT[] NOT NULL CHECK(cardinality(scopes) BETWEEN 1 AND 8 AND scopes <@ ARRAY['profile:read','profile:write','registrations:read','registrations:write','organizer:read','organizer:write','admin:read','admin:write']::text[]),
 status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','denied')),
 user_id UUID REFERENCES users(id) ON DELETE CASCADE,session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
 code_hash TEXT UNIQUE CHECK(code_hash IS NULL OR length(code_hash)=64),code_used_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),expires_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()+INTERVAL '5 minutes'
);
CREATE INDEX mcp_requests_expiry_idx ON mcp_oauth_requests(expires_at);
CREATE TABLE mcp_oauth_grants (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,client_id TEXT NOT NULL REFERENCES mcp_oauth_clients(id),
 resource TEXT NOT NULL CHECK(length(resource)<=300),
 scopes TEXT[] NOT NULL CHECK(cardinality(scopes) BETWEEN 1 AND 8 AND scopes <@ ARRAY['profile:read','profile:write','registrations:read','registrations:write','organizer:read','organizer:write','admin:read','admin:write']::text[]),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),expires_at TIMESTAMPTZ NOT NULL,
 revoked_at TIMESTAMPTZ,revocation_reason TEXT NOT NULL DEFAULT '' CHECK(length(revocation_reason)<=100)
);
CREATE INDEX mcp_grants_user_idx ON mcp_oauth_grants(user_id,created_at DESC,id);
CREATE INDEX mcp_grants_session_idx ON mcp_oauth_grants(session_id);
CREATE INDEX mcp_grants_expiry_idx ON mcp_oauth_grants(expires_at);
CREATE TABLE mcp_access_tokens (
 token_hash TEXT PRIMARY KEY CHECK(length(token_hash)=64),grant_id UUID NOT NULL REFERENCES mcp_oauth_grants(id) ON DELETE CASCADE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX mcp_access_grant_idx ON mcp_access_tokens(grant_id);
CREATE INDEX mcp_access_expiry_idx ON mcp_access_tokens(expires_at);
CREATE TABLE mcp_refresh_tokens (
 token_hash TEXT PRIMARY KEY CHECK(length(token_hash)=64),grant_id UUID NOT NULL REFERENCES mcp_oauth_grants(id) ON DELETE CASCADE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),expires_at TIMESTAMPTZ NOT NULL,used_at TIMESTAMPTZ
);
CREATE INDEX mcp_refresh_grant_idx ON mcp_refresh_tokens(grant_id);
CREATE TABLE mcp_action_intents (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),grant_id UUID NOT NULL REFERENCES mcp_oauth_grants(id) ON DELETE CASCADE,
 tool_name TEXT NOT NULL CHECK(length(tool_name) BETWEEN 1 AND 100),api_service TEXT NOT NULL CHECK(api_service IN ('core','global')),
 method TEXT NOT NULL CHECK(method IN ('POST','PUT','PATCH','DELETE')),api_path TEXT NOT NULL CHECK(length(api_path) BETWEEN 1 AND 1000),
 payload_json JSONB NOT NULL CHECK(jsonb_typeof(payload_json)='object'),request_hash TEXT NOT NULL CHECK(length(request_hash)=64),
 status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','denied','executing','finished','uncertain')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),expires_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()+INTERVAL '10 minutes',
 approved_at TIMESTAMPTZ,approved_session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
 lease_hash TEXT CHECK(lease_hash IS NULL OR length(lease_hash)=64),lease_expires_at TIMESTAMPTZ,
 response_status INTEGER CHECK(response_status BETWEEN 100 AND 599),result_json JSONB,finished_at TIMESTAMPTZ
);
CREATE INDEX mcp_actions_grant_idx ON mcp_action_intents(grant_id,created_at DESC);
CREATE INDEX mcp_actions_expiry_idx ON mcp_action_intents(expires_at);

CREATE INDEX mcp_requests_client_idx ON mcp_oauth_requests(client_id);
CREATE INDEX mcp_grants_client_idx ON mcp_oauth_grants(client_id);
