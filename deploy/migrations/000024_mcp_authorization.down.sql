DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM mcp_oauth_grants) OR EXISTS(SELECT 1 FROM mcp_action_intents) THEN RAISE EXCEPTION 'MCP connections and action records must be explicitly resolved before removing their storage'; END IF;
END $$;
DROP TABLE mcp_action_intents;
DROP TABLE mcp_refresh_tokens;
DROP TABLE mcp_access_tokens;
DROP TABLE mcp_oauth_grants;
DROP TABLE mcp_oauth_requests;
DROP TABLE mcp_oauth_clients;
