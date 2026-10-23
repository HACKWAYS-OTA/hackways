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

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const mcpResourceURL = "https://mcp.hackways.in/mcp"
const mcpService = "core"

type mcpRoute struct {
	method, pattern, scope string
	read                   bool
}

var mcpRoutes = []mcpRoute{
	{"POST", `/api/v1/events/organizer/profile`, "organizer:write", false},
	{"PUT", `/api/v1/admin/core/spaces/[A-Za-z0-9-]{1,80}`, "admin:write", false},
	{"POST", `/api/v1/rsvps/`, "registrations:write", false},
	{"GET", `/api/v1/spaces/(mine|[a-fA-F0-9-]{36}/membership)`, "profile:read", true},
	{"POST", `/api/v1/spaces/[a-fA-F0-9-]{36}/(join|leave)`, "profile:write", false},
	{"GET", `/api/v1/spaces/[a-fA-F0-9-]{36}/manage`, "organizer:read", true},
	{"POST", `/api/v1/spaces/`, "organizer:write", false},
	{"PUT", `/api/v1/spaces/[a-fA-F0-9-]{36}`, "organizer:write", false},
	{"POST", `/api/v1/spaces/[a-fA-F0-9-]{36}/status`, "organizer:write", false},
	{"PUT", `/api/v1/admin/core/events/[A-Za-z0-9-]{1,80}`, "admin:write", false},
	{"GET", `/api/v1/events/by-slug/[^/]+/invitation`, "registrations:read", true},
	{"GET", `/api/v1/rsvps(/|/[a-fA-F0-9-]{36}|/by-code/[^/]+)`, "registrations:read", true},
	{"POST", `/api/v1/rsvps/(claim|[a-fA-F0-9-]{36}/(cancel|accept|transfer))`, "registrations:write", false},
	{"GET", `/api/v1/events/(organizer|organizer/profile|manage-by-slug/[^/]+|[a-fA-F0-9-]{36}/(manage|registrations|invitations|registration-form|summary|audit)|[a-fA-F0-9-]{36}/registrations/by-code/[^/]+)`, "organizer:read", true},
	{"POST", `/api/v1/events/`, "organizer:write", false},
	{"PUT", `/api/v1/events/[a-fA-F0-9-]{36}(/(space|appearance|registration-form))?`, "organizer:write", false},
	{"POST", `/api/v1/events/[a-fA-F0-9-]{36}/(status|invitations|check-in|check-in/undo|registrations/by-code/[^/]+/status)`, "organizer:write", false},
	{"DELETE", `/api/v1/events/[a-fA-F0-9-]{36}`, "organizer:write", false},
	{"GET", `/api/v1/admin/core/(overview|analytics|events|spaces|registrations|tickets|moderation|audit|discover-promotions|discover-promotions/events)`, "admin:read", true},
	{"GET", `/api/v1/admin/core/(events|spaces|registrations|tickets|moderation)/[A-Za-z0-9-]{1,80}`, "admin:read", true},
	{"POST", `/api/v1/admin/core/(events|spaces|registrations|tickets|moderation)/[A-Za-z0-9-]{1,80}/(status|check-in)`, "admin:write", false},
	{"PUT", `/api/v1/admin/core/discover-promotions`, "admin:write", false},
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
