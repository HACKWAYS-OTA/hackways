package mcpbridge

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/admin"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Mandatory CI exercises the real SDK, PostgreSQL grants and existing profile
// HTTP handler. No HTTP response or business operation is mocked.
func TestMCPProtocolAndConfirmedWrite(t *testing.T) {
	t.Setenv("MCP_ENABLED", "true")
	for _, tool := range catalog() {
		if tool.Name == "update_event" {
			var schema struct {
				Required []string `json:"required"`
			}
			if json.Unmarshal(inputSchema(tool), &schema) != nil {
				t.Fatal("invalid event schema")
			}
			for _, key := range []string{"capacity", "registration_mode", "waitlist_enabled", "registration_opens_at", "registration_closes_at"} {
				if !slices.Contains(schema.Required, key) {
					t.Fatal("replacement update could silently reset " + key)
				}
			}
		}
	}

	ctx := context.Background()
	pool, e := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	q := db.New(pool)
	nonce, e := auth.GenerateToken()
	if e != nil {
		t.Fatal(e)
	}
	user, e := q.CreateUser(ctx, db.CreateUserParams{Email: "mcp-protocol-" + nonce + "@example.invalid", PasswordHash: "unusable-isolated-fixture", DisplayName: "Protocol account"})
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM mcp_oauth_clients WHERE id=$1", "mcp-test-"+nonce)
	}()
	if e = q.MarkUserEmailVerified(ctx, user.ID); e != nil {
		t.Fatal(e)
	}
	handler := auth.NewHandler(pool, false)
	sessionToken, session, e := handler.SessionManager().CreateSession(ctx, user.ID, "Protocol CI", "", time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	client, e := q.MCPRegisterClient(ctx, db.MCPRegisterClientParams{ID: "mcp-test-" + nonce, Name: "Protocol CI", RedirectUris: []string{"http://127.0.0.1/callback"}})
	if e != nil {
		t.Fatal(e)
	}
	grant, e := q.MCPCreateGrant(ctx, db.MCPCreateGrantParams{ID: session.ID, ClientID: client.ID, Resource: auth.MCPResource(), Scopes: []string{"profile:read", "profile:write"}})
	if e != nil {
		t.Fatal(e)
	}
	rawAccess, e := auth.GenerateToken()
	if e != nil {
		t.Fatal(e)
	}
	access := "mcp_at_" + rawAccess
	if _, e = q.MCPInsertAccessToken(ctx, db.MCPInsertAccessTokenParams{ID: grant.ID, TokenHash: auth.HashToken(access)}); e != nil {
		t.Fatal(e)
	}
	api := chi.NewRouter()
	api.Mount("/api/v1/auth", handler.Routes())
	api.Mount("/api/v1/admin/global", admin.NewHandler(pool).Routes())
	upstream := httptest.NewServer(api)
	defer upstream.Close()
	bridge, e := New(pool, slog.New(slog.NewTextHandler(io.Discard, nil)), upstream.URL, upstream.URL)
	if e != nil {
		t.Fatal(e)
	}
	rpc := func(body, credential string) (int, map[string]json.RawMessage) {
		req := httptest.NewRequest("POST", "https://mcp.hackways.in/mcp", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		if credential != "" {
			req.Header.Set("Authorization", "Bearer "+credential)
		}
		rr := httptest.NewRecorder()
		bridge.ServeHTTP(rr, req)
		var envelope map[string]json.RawMessage
		if json.Unmarshal(rr.Body.Bytes(), &envelope) != nil {
			t.Fatalf("non-JSON protocol response: %d %s", rr.Code, rr.Body.String())
		}
		return rr.Code, envelope
	}
	status, init := rpc(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"CI","version":"1"}}}`, "")
	if status != 200 || len(init["result"]) == 0 {
		t.Fatal("MCP initialization failed")
	}
	status, list := rpc(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`, "")
	if status != 200 {
		t.Fatal("tool listing failed")
	}
	var listed struct {
		Tools []struct {
			Name        string                     `json:"name"`
			Security    []securityScheme           `json:"securitySchemes"`
			Meta        map[string]json.RawMessage `json:"_meta"`
			Annotations struct {
				Read bool `json:"readOnlyHint"`
			} `json:"annotations"`
		} `json:"tools"`
	}
	if json.Unmarshal(list["result"], &listed) != nil || len(listed.Tools) != len(catalog()) || len(listed.Tools) < 100 {
		t.Fatalf("catalog missing tools: %d", len(listed.Tools))
	}
	for _, tool := range listed.Tools {
		if len(tool.Security) != 1 || len(tool.Meta["securitySchemes"]) == 0 {
			t.Fatalf("missing security declaration: %s", tool.Name)
		}
	}
	_, unauthorized := rpc(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_my_profile","arguments":{}}}`, "")
	if !strings.Contains(string(unauthorized["result"]), "mcp/www_authenticate") {
		t.Fatal("missing runtime OAuth challenge")
	}
	_, profile := rpc(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_my_profile","arguments":{}}}`, access)
	if !strings.Contains(string(profile["result"]), "Protocol account") {
		t.Fatalf("real API profile not returned: %s", profile["result"])
	}
	_, invalid := rpc(`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"update_my_name","arguments":{"display_name":123}}}`, access)
	if !strings.Contains(string(invalid["result"]), `"isError":true`) {
		t.Fatal("schema accepted incorrect input type")
	}
	call := `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"update_my_name","arguments":{"display_name":"Protocol approved"}}}`
	_, prepared := rpc(call, access)
	var result struct {
		Structured Output `json:"structuredContent"`
	}
	_ = json.Unmarshal(prepared["result"], &result)
	if result.Structured.Status != "approval_required" || result.Structured.ApprovalID == "" {
		t.Fatalf("approval not prepared: %s", prepared["result"])
	}
	saved, e := q.GetUserByEmail(ctx, user.Email)
	if e != nil || saved.DisplayName != "Protocol account" {
		t.Fatal("prepare mutated the account")
	}
	req := httptest.NewRequest("POST", "/api/v1/auth/mcp/actions/"+result.Structured.ApprovalID+"/decision", strings.NewReader(`{"allow":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://hackways.in")
	req.AddCookie(&http.Cookie{Name: auth.DefaultSessionCookie, Value: sessionToken})
	decision := httptest.NewRecorder()
	api.ServeHTTP(decision, req)
	if decision.Code != 200 {
		t.Fatalf("browser approval: %d %s", decision.Code, decision.Body.String())
	}
	confirmed := strings.Replace(call, `"display_name":"Protocol approved"`, `"display_name":"Protocol approved","approval_id":"`+result.Structured.ApprovalID+`"`, 1)
	_, executed := rpc(confirmed, access)
	_ = json.Unmarshal(executed["result"], &result)
	if result.Structured.Status != "success" {
		t.Fatalf("confirmed tool failed: %s", executed["result"])
	}
	saved, e = q.GetUserByEmail(ctx, user.Email)
	if e != nil || saved.DisplayName != "Protocol approved" {
		t.Fatal("tool did not persist actual API change")
	}
	_, replayed := rpc(confirmed, access)
	if string(executed["result"]) != string(replayed["result"]) {
		t.Fatal("confirmed repeat must return saved result")
	}
	after, e := q.GetUserByEmail(ctx, user.Email)
	if e != nil || after.UpdatedAt != saved.UpdatedAt {
		t.Fatal("confirmed repeat executed the API again")
	}
	adminGrant, e := q.MCPCreateGrant(ctx, db.MCPCreateGrantParams{ID: session.ID, ClientID: client.ID, Resource: auth.MCPResource(), Scopes: []string{"admin:read", "profile:write"}})
	if e != nil {
		t.Fatal(e)
	}
	adminRandom, e := auth.GenerateToken()
	if e != nil {
		t.Fatal(e)
	}
	adminAccess := "mcp_at_" + adminRandom
	if _, e = q.MCPInsertAccessToken(ctx, db.MCPInsertAccessTokenParams{ID: adminGrant.ID, TokenHash: auth.HashToken(adminAccess)}); e != nil {
		t.Fatal(e)
	}
	adminRead := func() int {
		r := httptest.NewRequest("GET", "/api/v1/admin/global/overview", nil)
		r.Header.Set("Authorization", "Bearer "+adminAccess)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		return w.Code
	}
	if adminRead() != 403 {
		t.Fatal("OAuth admin scope bypassed real role grant")
	}
	if _, e = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES($1)", user.ID); e != nil {
		t.Fatal(e)
	}
	if adminRead() != 403 {
		t.Fatal("OAuth scope and role bypassed dedicated admin session")
	}
	if e = q.RecordPlatformAdminSession(ctx, session.ID); e != nil {
		t.Fatal(e)
	}
	if adminRead() != 200 {
		t.Fatal("verified actual admin could not read real overview")
	}
	_, adminPrepared := rpc(`{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"update_my_name","arguments":{"display_name":"Review only"}}}`, adminAccess)
	var adminApproval struct {
		Structured Output `json:"structuredContent"`
	}
	_ = json.Unmarshal(adminPrepared["result"], &adminApproval)
	if adminApproval.Structured.Status != "approval_required" || !strings.HasPrefix(adminApproval.Structured.ApprovalURL, "https://admin.hackways.in/connect/action?") {
		t.Fatal("an admin connection must review all its actions on the admin cookie origin")
	}

	if _, e = pool.Exec(ctx, "UPDATE platform_admin_sessions SET otp_verified_at=clock_timestamp()-INTERVAL '25 hours' WHERE session_id=$1", session.ID); e != nil {
		t.Fatal(e)
	}
	if adminRead() != 403 {
		t.Fatal("OAuth bypassed daily OTP expiry")
	}
	if e = q.MarkAdminSessionVerified(ctx, session.ID); e != nil {
		t.Fatal(e)
	}
	if adminRead() != 200 {
		t.Fatal("verified admin connection did not recover")
	}
	if _, e = pool.Exec(ctx, "UPDATE platform_administrators SET revoked_at=clock_timestamp() WHERE user_id=$1", user.ID); e != nil {
		t.Fatal(e)
	}
	if adminRead() != 403 {
		t.Fatal("OAuth retained a revoked admin grant")
	}
	if e = handler.SessionManager().RevokeSession(ctx, sessionToken); e != nil {
		t.Fatal(e)
	}
	_, revoked := rpc(`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"get_my_profile","arguments":{}}}`, access)
	if !strings.Contains(string(revoked["result"]), "mcp/www_authenticate") {
		t.Fatal("connection survived browser-session revocation")
	}

}
