package auth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const mcpResourceURL = "https://mcp.hackways.in/mcp"
const mcpService = "global"

type mcpRoute struct {
	method, pattern, scope string
	read                   bool
}

var mcpRoutes = []mcpRoute{
	{"GET", `/api/v1/auth/(me|sessions|notifications|promotions/preference)`, "profile:read", true},
	{"PATCH", `/api/v1/auth/me`, "profile:write", false},
	{"POST", `/api/v1/auth/promotions/preference`, "profile:write", false},
	{"DELETE", `/api/v1/auth/sessions/[A-Za-z0-9-]{1,80}`, "profile:write", false},
	{"GET", `/api/v1/admin/global/(overview|analytics|people|messaging|storage|rollouts|campaigns|campaign-contacts)`, "admin:read", true},
	{"GET", `/api/v1/admin/global/(people|messaging|storage|campaigns)/[A-Za-z0-9-]{1,80}`, "admin:read", true},
	{"GET", `/api/v1/admin/global/campaigns/recipients`, "admin:read", true},
	{"POST", `/api/v1/admin/global/(campaigns/preview|messages/preview)`, "admin:read", true},
	{"POST", `/api/v1/admin/global/(rollouts|campaigns|campaign-contacts/import|messages/send)`, "admin:write", false},
	{"POST", `/api/v1/admin/global/campaigns/[A-Za-z0-9-]{1,80}/(edit|send|cancel|test)`, "admin:write", false},
	{"POST", `/api/v1/admin/global/(people/[A-Za-z0-9-]{1,80}/(edit|access)|people/[A-Za-z0-9-]{1,80}/sessions/[A-Za-z0-9-]{1,80}/revoke|messaging/[A-Za-z0-9-]{1,80}/retry|storage/[A-Za-z0-9-]{1,80}/detach)`, "admin:write", false},
}
var mcpRoutePatterns = func() []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(mcpRoutes))
	for i, v := range mcpRoutes {
		out[i] = regexp.MustCompile("^" + v.pattern + "$")
	}
	return out
}()

// MCPIdentity validates an opaque OAuth access credential against its real grant,
// session, user and audience. Browser cookies never enter this path.
func MCPIdentity(ctx context.Context, q *db.Queries, header string) (db.MCPAccessIdentityRow, error) {
	if os.Getenv("MCP_ENABLED") != "true" || !strings.HasPrefix(header, "Bearer mcp_at_") || len(header) != len("Bearer mcp_at_")+64 {
		return db.MCPAccessIdentityRow{}, ErrInvalidToken
	}
	token := strings.TrimPrefix(header, "Bearer ")
	row, e := q.MCPAccessIdentity(ctx, HashToken(token))
	if e != nil {
		return row, e
	}
	if row.Resource != mcpResourceURL {
		return row, ErrInvalidToken
	}
	return row, nil
}
func mcpSession(row db.MCPAccessIdentityRow) *SessionUser {
	var avatar *string
	if row.AvatarUrl.Valid {
		v := row.AvatarUrl.String
		avatar = &v
	}
	return &SessionUser{SessionID: row.SessionID, UserID: row.UserID, Email: row.Email, DisplayName: row.DisplayName, AvatarURL: avatar, EmailVerified: row.EmailVerified, ExpiresAt: row.ExpiresAt.Time}
}
func mcpRequestUser(w http.ResponseWriter, r *http.Request, sm *SessionManager) (*SessionUser, bool) {
	row, e := MCPIdentity(r.Context(), sm.queries, r.Header.Get("Authorization"))
	if e != nil {
		if errors.Is(e, pgx.ErrNoRows) || errors.Is(e, ErrInvalidToken) {
			w.Header().Set("WWW-Authenticate", `Bearer resource_metadata="https://mcp.hackways.in/.well-known/oauth-protected-resource/mcp", error="invalid_token"`)
			platform.RespondError(w, 401, "MCP_AUTH_REQUIRED", "Reconnect your Hackways account", nil)
		} else {
			platform.RespondDatabaseError(w, e)
		}
		return nil, false
	}
	var route *mcpRoute
	for i, v := range mcpRoutes {
		if v.method == r.Method && mcpRoutePatterns[i].MatchString(r.URL.Path) {
			route = &mcpRoutes[i]
			break
		}
	}
	if route == nil || !slices.Contains(row.Scopes, route.scope) {
		platform.RespondError(w, 403, "MCP_SCOPE_REQUIRED", "This connection does not allow this operation", nil)
		return nil, false
	}
	if !route.read {
		var id pgtype.UUID
		if id.Scan(r.Header.Get("X-MCP-Intent")) != nil || !id.Valid {
			platform.RespondError(w, 403, "MCP_APPROVAL_REQUIRED", "Approve this action in your browser", nil)
			return nil, false
		}
		body, e := io.ReadAll(io.LimitReader(r.Body, (128<<10)+1))
		if e != nil || len(body) > 128<<10 {
			platform.RespondError(w, 413, "REQUEST_TOO_LARGE", "Action is too large", nil)
			return nil, false
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		allowed, e := sm.queries.MCPExecutionAllowed(r.Context(), db.MCPExecutionAllowedParams{IntentID: id, GrantID: row.GrantID, RequestHash: HashToken(string(body)), ApiService: mcpService, Method: r.Method, ApiPath: r.URL.RequestURI(), LeaseHash: pgtype.Text{String: HashToken(r.Header.Get("X-MCP-Lease")), Valid: true}})
		if e != nil {
			platform.RespondDatabaseError(w, e)
			return nil, false
		}
		if !allowed {
			platform.RespondError(w, 403, "MCP_APPROVAL_REQUIRED", "Action approval is missing, expired or does not match", nil)
			return nil, false
		}
	}
	return mcpSession(row), true
}
