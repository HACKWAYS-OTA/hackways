package auth

// MCP credentials are deliberately separate from Hackways browser sessions.
// The session that approved a grant remains its revocable identity boundary.
import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var MCPScopes = []string{"profile:read", "profile:write", "registrations:read", "registrations:write", "organizer:read", "organizer:write", "admin:read", "admin:write"}
var pkceVerifier = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)
var pkceChallenge = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
var chatGPTMetadataPath = regexp.MustCompile(`^/oauth/(client\.json|[A-Za-z0-9_-]{1,128}/client\.json)$`)

func MCPEnabled() bool    { return os.Getenv("MCP_ENABLED") == "true" }
func MCPIssuer() string   { return "https://mcp.hackways.in" }
func MCPResource() string { return MCPIssuer() + "/mcp" }
func mcpOrigin(admin bool) string {
	if admin {
		return "https://admin.hackways.in"
	}
	return "https://hackways.in"
}
func mcpAccountScopes(scopes []string) []string {
	out := []string{}
	for _, scope := range scopes {
		if !strings.HasPrefix(scope, "admin:") {
			out = append(out, scope)
		}
	}
	return out
}
func mcpConsentAdmin(r *http.Request, scopes []string) bool {
	return mcpAdmin(scopes) && (r.URL.Query().Get("admin") == "true" || len(mcpAccountScopes(scopes)) == 0)
}
func mcpAdmin(scopes []string) bool {
	return slices.Contains(scopes, "admin:read") || slices.Contains(scopes, "admin:write")
}
func mcpID(raw string) (pgtype.UUID, bool) {
	var id pgtype.UUID
	err := id.Scan(raw)
	return id, err == nil && id.Valid
}
func mcpIDString(id pgtype.UUID) string { value, _ := id.Value(); s, _ := value.(string); return s }
func mcpText(s string) pgtype.Text      { return pgtype.Text{String: s, Valid: true} }
func mcpJSON(w http.ResponseWriter, status int, value json.RawMessage) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(value)
}
func mcpReply[T any](w http.ResponseWriter, status int, value T) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "Response unavailable", 500)
		return
	}
	mcpJSON(w, status, data)
}
func mcpOAuthError(w http.ResponseWriter, status int, code, description string) {
	mcpReply(w, status, struct {
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}{code, description})
}

func (h *Handler) MCPRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/metadata/resource", func(w http.ResponseWriter, r *http.Request) {
		mcpReply(w, 200, struct {
			Resource string   `json:"resource"`
			Servers  []string `json:"authorization_servers"`
			Scopes   []string `json:"scopes_supported"`
			Methods  []string `json:"bearer_methods_supported"`
		}{MCPResource(), []string{MCPIssuer()}, MCPScopes, []string{"header"}})
	})
	r.Get("/metadata/server", func(w http.ResponseWriter, r *http.Request) {
		mcpReply(w, 200, struct {
			Issuer        string   `json:"issuer"`
			Authorization string   `json:"authorization_endpoint"`
			Token         string   `json:"token_endpoint"`
			Register      string   `json:"registration_endpoint"`
			Revoke        string   `json:"revocation_endpoint"`
			Scopes        []string `json:"scopes_supported"`
			Responses     []string `json:"response_types_supported"`
			Grants        []string `json:"grant_types_supported"`
			Auth          []string `json:"token_endpoint_auth_methods_supported"`
			PKCE          []string `json:"code_challenge_methods_supported"`
			Iss           bool     `json:"authorization_response_iss_parameter_supported"`
			CIMD          bool     `json:"client_id_metadata_document_supported"`
		}{MCPIssuer(), MCPIssuer() + "/oauth/authorize", MCPIssuer() + "/oauth/token", MCPIssuer() + "/oauth/register", MCPIssuer() + "/oauth/revoke", MCPScopes, []string{"code"}, []string{"authorization_code", "refresh_token"}, []string{"none"}, []string{"S256"}, true, true})
	})
	r.Post("/register", h.mcpRegister)
	r.Get("/authorize", h.mcpAuthorize)
	r.Post("/token", h.mcpToken)
	r.Post("/revoke", h.mcpRevoke)
	// Browser consent endpoints never accept a bearer token as a browser identity.
	r.Get("/requests/{id}", h.mcpConsent)
	r.Post("/requests/{id}/decision", h.mcpDecision)
	r.Get("/actions/{id}", h.mcpAction)
	r.Post("/actions/{id}/decision", h.mcpActionDecision)
	r.Get("/connections", h.mcpConnections)
	r.Delete("/connections/{id}", h.mcpDisconnect)
	return r
}

type mcpClientMetadata struct {
	MethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	ClientID         string   `json:"client_id"`
	Name             string   `json:"client_name"`
	Redirects        []string `json:"redirect_uris"`
	Method           string   `json:"token_endpoint_auth_method"`
	Grants           []string `json:"grant_types"`
	Responses        []string `json:"response_types"`
}

func validMCPRedirect(raw string) bool {
	if len(raw) > 2048 {
		return false
	}
	u, e := url.Parse(raw)
	if e != nil || u.Host == "" || u.User != nil || u.Fragment != "" {
		return false
	}
	if u.Scheme == "https" {
		return true
	}
	ip := net.ParseIP(u.Hostname())
	return u.Scheme == "http" && (u.Hostname() == "localhost" || (ip != nil && ip.IsLoopback()))
}

// ChatGPT's plural CIMD field lists supported methods, while the legacy
// singular field is only a preference (SEP-3149). Our AS accepts only none.
func validMCPCIMDClient(c *mcpClientMetadata) bool {
	if c.MethodsSupported != nil {
		if !slices.Contains(c.MethodsSupported, "none") {
			return false
		}
		c.Method = "none"
	}
	return validMCPClient(*c)
}

func validMCPClient(c mcpClientMetadata) bool {
	if len(strings.TrimSpace(c.Name)) < 1 || len(c.Name) > 100 || len(c.Redirects) < 1 || len(c.Redirects) > 10 || c.Method != "none" {
		return false
	}
	for _, v := range c.Redirects {
		if !validMCPRedirect(v) {
			return false
		}
	}
	for _, v := range c.Grants {
		if v != "authorization_code" && v != "refresh_token" {
			return false
		}
	}
	for _, v := range c.Responses {
		if v != "code" {
			return false
		}
	}
	return true
}
func (h *Handler) mcpSaveClient(ctx context.Context, c mcpClientMetadata, metadata string, verified bool) (db.McpOauthClients, error) {
	tx, e := h.pool.Begin(ctx)
	if e != nil {
		return db.McpOauthClients{}, e
	}
	defer tx.Rollback(ctx)
	q := h.queries.WithTx(tx)
	if e = q.MCPLockClientRegistry(ctx); e != nil {
		return db.McpOauthClients{}, e
	}
	row, e := q.MCPRegisterClient(ctx, db.MCPRegisterClientParams{ID: c.ClientID, Name: c.Name, RedirectUris: c.Redirects, MetadataUrl: metadata, Verified: verified})
	if e != nil {
		return row, e
	}
	e = tx.Commit(ctx)
	return row, e
}
func (h *Handler) mcpRegister(w http.ResponseWriter, r *http.Request) {
	var c mcpClientMetadata
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if json.NewDecoder(r.Body).Decode(&c) != nil || !validMCPClient(c) {
		mcpOAuthError(w, 400, "invalid_client_metadata", "Use a name, exact HTTPS or loopback redirects and public client authentication (none).")
		return
	}
	token, e := GenerateToken()
	if e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Registration unavailable")
		return
	}
	c.ClientID = "mcp_client_" + token
	c.MethodsSupported = nil
	if _, e = h.mcpSaveClient(r.Context(), c, "", false); e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Client registry unavailable")
		return
	}
	c.Grants = []string{"authorization_code", "refresh_token"}
	c.Responses = []string{"code"}
	mcpReply(w, 201, c)
}
func (h *Handler) mcpClient(ctx context.Context, id string) (db.McpOauthClients, error) {
	if !strings.HasPrefix(id, "https://") {
		return h.queries.MCPGetClient(ctx, id)
	}
	// CIMD is intentionally restricted to the documented ChatGPT metadata origin.
	// It cannot fetch arbitrary client-provided hosts, redirects or private networks.
	u, e := url.Parse(id)
	if e != nil || u.Scheme != "https" || u.Host != "chatgpt.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !chatGPTMetadataPath.MatchString(u.Path) {
		return db.McpOauthClients{}, pgx.ErrNoRows
	}
	client := http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, id, nil)
	if e != nil {
		return db.McpOauthClients{}, e
	}
	resp, e := client.Do(req)
	if e != nil {
		return db.McpOauthClients{}, e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return db.McpOauthClients{}, pgx.ErrNoRows
	}
	raw, e := io.ReadAll(io.LimitReader(resp.Body, (32<<10)+1))
	if e != nil || len(raw) > 32<<10 {
		return db.McpOauthClients{}, pgx.ErrNoRows
	}
	var c mcpClientMetadata
	if json.Unmarshal(raw, &c) != nil || c.ClientID != id || !validMCPCIMDClient(&c) {
		return db.McpOauthClients{}, pgx.ErrNoRows
	}
	for _, redirect := range c.Redirects {
		v, _ := url.Parse(redirect)
		if v.Scheme != "https" || v.Host != "chatgpt.com" {
			return db.McpOauthClients{}, pgx.ErrNoRows
		}
	}
	return h.mcpSaveClient(ctx, c, id, true)
}
func (h *Handler) mcpAuthorize(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	scopes := strings.Fields(v.Get("scope"))
	if v.Get("response_type") != "code" || v.Get("resource") != MCPResource() || v.Get("code_challenge_method") != "S256" || !pkceChallenge.MatchString(v.Get("code_challenge")) || len(v.Get("state")) > 2048 || len(scopes) == 0 || len(scopes) > 8 {
		mcpOAuthError(w, 400, "invalid_request", "A resource-bound authorization code request with S256 PKCE and scopes is required.")
		return
	}
	for _, s := range scopes {
		if !slices.Contains(MCPScopes, s) {
			mcpOAuthError(w, 400, "invalid_scope", "Unsupported permission")
			return
		}
	}
	client, e := h.mcpClient(r.Context(), v.Get("client_id"))
	if e != nil || !slices.Contains(client.RedirectUris, v.Get("redirect_uri")) {
		mcpOAuthError(w, 400, "invalid_client", "Client or redirect not registered")
		return
	}
	tx, e := h.pool.Begin(r.Context())
	if e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Authorization unavailable")
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	if e = q.MCPLockClientRegistry(r.Context()); e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Authorization unavailable")
		return
	}
	row, e := q.MCPCreateAuthorizationRequest(r.Context(), db.MCPCreateAuthorizationRequestParams{ClientID: client.ID, RedirectUri: v.Get("redirect_uri"), State: v.Get("state"), Resource: MCPResource(), CodeChallenge: v.Get("code_challenge"), Scopes: scopes})
	if e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Authorization unavailable")
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Authorization unavailable")
		return
	}
	http.Redirect(w, r, mcpOrigin(mcpAdmin(scopes) && len(mcpAccountScopes(scopes)) == 0)+"/connect/authorize?request="+mcpIDString(row.ID), http.StatusSeeOther)
}

func (h *Handler) mcpBrowser(w http.ResponseWriter, r *http.Request, admin, mutation bool) (*SessionUser, bool) {
	if !admin && !allowRequestDevice(w, r, h.queries) {
		return nil, false
	}
	if r.Header.Get("Authorization") != "" {
		platform.RespondError(w, 401, "BROWSER_SIGN_IN_REQUIRED", "Open this page in your browser and sign in", nil)
		return nil, false
	}
	if mutation && r.Header.Get("Origin") != mcpOrigin(admin) {
		platform.RespondError(w, 403, "ORIGIN_REQUIRED", "This action must be approved on Hackways", nil)
		return nil, false
	}
	cookie := DefaultSessionCookie
	if admin {
		cookie = AdminSessionCookie
	}
	user, e := h.sm.GetSession(r.Context(), GetSessionToken(r, cookie))
	if e != nil {
		if errors.Is(e, ErrNotAllowed) {
			respondNotAllowed(w)
		} else if errors.Is(e, ErrSessionNotFound) || errors.Is(e, ErrInvalidToken) {
			platform.RespondError(w, 401, "SIGN_IN_REQUIRED", "Sign in to continue", nil)
		} else {
			platform.RespondDatabaseError(w, e)
		}
		return nil, false
	}
	if !user.EmailVerified {
		platform.RespondError(w, 403, "EMAIL_VERIFICATION_REQUIRED", "Verify your email before connecting an app", nil)
		return nil, false
	}
	if admin {
		allowed, err := h.queries.IsPlatformAdminSession(r.Context(), db.IsPlatformAdminSessionParams{SessionID: user.SessionID, UserID: user.UserID})
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return nil, false
		}
		if !allowed {
			platform.RespondError(w, 403, "ADMIN_VERIFICATION_REQUIRED", "Sign in to admin and complete daily verification", nil)
			return nil, false
		}
	}
	return user, true
}

type mcpConsentView struct {
	AdminURL   string             `json:"admin_authorize_url,omitempty"`
	ID         pgtype.UUID        `json:"id"`
	ClientName string             `json:"client_name"`
	Verified   bool               `json:"client_verified"`
	ClientID   string             `json:"client_id"`
	Scopes     []string           `json:"scopes"`
	Expires    pgtype.Timestamptz `json:"expires_at"`
	Email      string             `json:"email"`
	Status     string             `json:"status"`
}

func (h *Handler) mcpConsent(w http.ResponseWriter, r *http.Request) {
	id, ok := mcpID(chi.URLParam(r, "id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	row, e := h.queries.MCPAuthorizationRequest(r.Context(), id)
	if e != nil {
		if errors.Is(e, pgx.ErrNoRows) {
			platform.RespondError(w, 410, "EXPIRED", "Connection request expired. Start again from your app.", nil)
		} else {
			platform.RespondDatabaseError(w, e)
		}
		return
	}
	user, ok := h.mcpBrowser(w, r, mcpConsentAdmin(r, row.Scopes), false)
	if !ok {
		return
	}
	scopes := row.Scopes
	adminURL := ""
	if !mcpConsentAdmin(r, row.Scopes) {
		scopes = mcpAccountScopes(row.Scopes)
		if mcpAdmin(row.Scopes) {
			adminURL = mcpOrigin(true) + "/connect/authorize?request=" + mcpIDString(row.ID)
		}
	}
	mcpReply(w, 200, mcpConsentView{ID: row.ID, ClientName: row.ClientName, Verified: row.ClientVerified, ClientID: row.ClientID, Scopes: scopes, Expires: row.ExpiresAt, Email: user.Email, Status: row.Status, AdminURL: adminURL})
}
func (h *Handler) mcpDecision(w http.ResponseWriter, r *http.Request) {
	id, ok := mcpID(chi.URLParam(r, "id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	row, e := h.queries.MCPAuthorizationRequest(r.Context(), id)
	if e != nil {
		if errors.Is(e, pgx.ErrNoRows) {
			platform.RespondError(w, 410, "EXPIRED", "Connection request expired", nil)
		} else {
			platform.RespondDatabaseError(w, e)
		}
		return
	}
	user, ok := h.mcpBrowser(w, r, mcpConsentAdmin(r, row.Scopes), true)
	if !ok {
		return
	}
	var input struct {
		Allow bool `json:"allow"`
	}
	if platform.DecodeJSON(r, &input) != nil {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Choose Allow or Cancel", nil)
		return
	}
	tx, e := h.pool.Begin(r.Context())
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	locked, e := q.MCPLockAuthorizationRequest(r.Context(), id)
	if e != nil && !errors.Is(e, pgx.ErrNoRows) {
		platform.RespondDatabaseError(w, e)
		return
	}
	if e != nil || locked.Status != "pending" {
		platform.RespondError(w, 409, "REQUEST_CLOSED", "Request already handled or expired", nil)
		return
	}
	callback, _ := url.Parse(row.RedirectUri)
	values := callback.Query()
	values.Set("state", row.State)
	values.Set("iss", MCPIssuer())
	if input.Allow {
		code, err := GenerateToken()
		if err != nil {
			platform.RespondError(w, 503, "UNAVAILABLE", "Authorization unavailable", nil)
			return
		}
		code = "mcp_ac_" + code
		_, e = q.MCPApproveAuthorizationRequest(r.Context(), db.MCPApproveAuthorizationRequestParams{ID: id, UserID: user.UserID, SessionID: user.SessionID, CodeHash: mcpText(HashToken(code)), Scopes: func() []string {
			if mcpConsentAdmin(r, row.Scopes) {
				return row.Scopes
			}
			return mcpAccountScopes(row.Scopes)
		}()})
		values.Set("code", code)
	} else {
		_, e = q.MCPDenyAuthorizationRequest(r.Context(), id)
		values.Set("error", "access_denied")
	}
	if e == nil {
		e = q.MCPAppendAudit(r.Context(), db.MCPAppendAuditParams{ActorUserID: user.UserID, EntityID: id, Action: "mcp.consent", RequestID: platform.GetRequestID(r.Context()), Metadata: []byte(`{}`)})
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	callback.RawQuery = values.Encode()
	mcpReply(w, 200, struct {
		Redirect string `json:"redirect_url"`
	}{callback.String()})
}

func (h *Handler) mcpToken(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if r.ParseForm() != nil || r.Form.Get("resource") != MCPResource() || r.Header.Get("Authorization") != "" || r.Form.Get("client_secret") != "" || r.Form.Get("client_assertion") != "" || r.Form.Get("client_assertion_type") != "" {
		mcpOAuthError(w, 400, "invalid_request", "Resource and public client credentials are required")
		return
	}
	ctx := r.Context()
	tx, e := h.pool.Begin(ctx)
	if e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Identity service unavailable")
		return
	}
	defer tx.Rollback(ctx)
	q := h.queries.WithTx(tx)
	var grantID pgtype.UUID
	var scopes []string
	switch r.Form.Get("grant_type") {
	case "authorization_code":
		code := r.Form.Get("code")
		verifier := r.Form.Get("code_verifier")
		if !strings.HasPrefix(code, "mcp_ac_") || !pkceVerifier.MatchString(verifier) {
			mcpOAuthError(w, 400, "invalid_grant", "Authorization code unavailable")
			return
		}
		row, err := q.MCPLockAuthorizationCode(ctx, mcpText(HashToken(code)))
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				mcpOAuthError(w, 503, "temporarily_unavailable", "Identity service unavailable")
			} else {
				mcpOAuthError(w, 400, "invalid_grant", "Authorization code unavailable")
			}
			return
		}
		digest := sha256.Sum256([]byte(verifier))
		challenge := base64.RawURLEncoding.EncodeToString(digest[:])
		if row.ClientID != r.Form.Get("client_id") || row.RedirectUri != r.Form.Get("redirect_uri") || row.Resource != MCPResource() || row.CodeUsedAt.Valid || subtle.ConstantTimeCompare([]byte(challenge), []byte(row.CodeChallenge)) != 1 {
			mcpOAuthError(w, 400, "invalid_grant", "Authorization code unavailable")
			return
		}
		if e = q.MCPLockAccountGrants(ctx, row.UserID); e == nil {
			var g db.McpOauthGrants
			g, e = q.MCPCreateGrant(ctx, db.MCPCreateGrantParams{ID: row.SessionID, ClientID: row.ClientID, Resource: row.Resource, Scopes: row.Scopes})
			grantID = g.ID
			scopes = g.Scopes
		}
		if e == nil {
			e = q.MCPConsumeAuthorizationCode(ctx, row.ID)
		}
	case "refresh_token":
		token := r.Form.Get("refresh_token")
		if !strings.HasPrefix(token, "mcp_rt_") {
			mcpOAuthError(w, 400, "invalid_grant", "Refresh credential unavailable")
			return
		}
		row, err := q.MCPLockRefreshToken(ctx, HashToken(token))
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				mcpOAuthError(w, 503, "temporarily_unavailable", "Identity service unavailable")
			} else {
				mcpOAuthError(w, 400, "invalid_grant", "Refresh credential unavailable")
			}
			return
		}
		if row.ClientID != r.Form.Get("client_id") || row.Resource != MCPResource() {
			mcpOAuthError(w, 400, "invalid_grant", "Refresh credential unavailable")
			return
		}
		if row.UsedAt.Valid {
			_, e = q.MCPRevokeGrant(ctx, db.MCPRevokeGrantParams{ID: row.GrantID, RevocationReason: "Refresh credential reused"})
			if e == nil {
				e = tx.Commit(ctx)
			}
			if e != nil {
				mcpOAuthError(w, 503, "temporarily_unavailable", "Identity service unavailable")
			} else {
				mcpOAuthError(w, 400, "invalid_grant", "Connection revoked. Connect again.")
			}
			return
		}
		if row.RevokedAt.Valid || time.Now().After(row.ExpiresAt.Time) {
			mcpOAuthError(w, 400, "invalid_grant", "Connection expired")
			return
		}
		if requested := r.Form.Get("scope"); requested != "" && requested != strings.Join(row.Scopes, " ") {
			mcpOAuthError(w, 400, "invalid_scope", "Reconnect to change permissions")
			return
		}
		grantID = row.GrantID
		scopes = row.Scopes
		e = q.MCPConsumeRefreshToken(ctx, HashToken(token))
	default:
		mcpOAuthError(w, 400, "unsupported_grant_type", "Use authorization_code or refresh_token")
		return
	}
	if e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Connection could not be issued")
		return
	}
	access, e := GenerateToken()
	if e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Connection unavailable")
		return
	}
	refresh, e := GenerateToken()
	if e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Connection unavailable")
		return
	}
	access = "mcp_at_" + access
	refresh = "mcp_rt_" + refresh
	a, e := q.MCPInsertAccessToken(ctx, db.MCPInsertAccessTokenParams{TokenHash: HashToken(access), ID: grantID})
	if e == nil {
		e = q.MCPInsertRefreshToken(ctx, db.MCPInsertRefreshTokenParams{TokenHash: HashToken(refresh), ID: grantID})
	}
	if e == nil {
		e = tx.Commit(ctx)
	}
	if e != nil {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Connection unavailable")
		return
	}
	mcpReply(w, 200, struct {
		Access  string `json:"access_token"`
		Type    string `json:"token_type"`
		Expires int    `json:"expires_in"`
		Refresh string `json:"refresh_token"`
		Scope   string `json:"scope"`
	}{access, "Bearer", max(1, int(time.Until(a.ExpiresAt.Time).Seconds())), refresh, strings.Join(scopes, " ")})
}
func (h *Handler) mcpRevoke(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if r.ParseForm() != nil {
		mcpOAuthError(w, 400, "invalid_request", "Invalid request")
		return
	}
	row, e := h.queries.MCPGrantForToken(r.Context(), HashToken(r.Form.Get("token")))
	if e != nil && !errors.Is(e, pgx.ErrNoRows) {
		mcpOAuthError(w, 503, "temporarily_unavailable", "Identity service unavailable")
		return
	}
	if e == nil && row.ClientID == r.Form.Get("client_id") {
		if _, e = h.queries.MCPRevokeGrant(r.Context(), db.MCPRevokeGrantParams{ID: row.ID, RevocationReason: "Client disconnected"}); e != nil {
			mcpOAuthError(w, 503, "temporarily_unavailable", "Identity service unavailable")
			return
		}
	}
	w.WriteHeader(200)
}
func (h *Handler) mcpConnections(w http.ResponseWriter, r *http.Request) {
	admin := r.URL.Query().Get("admin") == "true"
	user, ok := h.mcpBrowser(w, r, admin, false)
	if !ok {
		return
	}
	rows, e := h.queries.MCPListConnections(r.Context(), user.UserID)
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	if rows == nil {
		rows = []db.MCPListConnectionsRow{}
	}
	mcpReply(w, 200, rows)
}
func (h *Handler) mcpDisconnect(w http.ResponseWriter, r *http.Request) {
	admin := r.URL.Query().Get("admin") == "true"
	user, ok := h.mcpBrowser(w, r, admin, true)
	if !ok {
		return
	}
	id, ok := mcpID(chi.URLParam(r, "id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	n, e := h.queries.MCPRevokeOwnedGrant(r.Context(), db.MCPRevokeOwnedGrantParams{ID: id, UserID: user.UserID})
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	if n == 0 {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) mcpAction(w http.ResponseWriter, r *http.Request) {
	id, ok := mcpID(chi.URLParam(r, "id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	row, e := h.queries.MCPReadActionIntent(r.Context(), id)
	if e != nil {
		if errors.Is(e, pgx.ErrNoRows) {
			platform.RespondError(w, 410, "EXPIRED", "Action expired or connection revoked", nil)
		} else {
			platform.RespondDatabaseError(w, e)
		}
		return
	}
	user, ok := h.mcpBrowser(w, r, mcpAdmin(row.Scopes), false)
	if !ok {
		return
	}
	if row.UserID != user.UserID {
		http.NotFound(w, r)
		return
	}
	mcpReply(w, 200, struct {
		ID      pgtype.UUID        `json:"id"`
		Tool    string             `json:"tool"`
		Client  string             `json:"client_name"`
		Payload json.RawMessage    `json:"payload"`
		Path    string             `json:"path"`
		Status  string             `json:"status"`
		Expires pgtype.Timestamptz `json:"expires_at"`
	}{row.ID, row.ToolName, row.ClientName, row.PayloadJson, row.ApiPath, row.Status, row.ExpiresAt})
}
func (h *Handler) mcpActionDecision(w http.ResponseWriter, r *http.Request) {
	id, ok := mcpID(chi.URLParam(r, "id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	row, e := h.queries.MCPReadActionIntent(r.Context(), id)
	if e != nil {
		if errors.Is(e, pgx.ErrNoRows) {
			platform.RespondError(w, 410, "EXPIRED", "Action expired or connection revoked", nil)
		} else {
			platform.RespondDatabaseError(w, e)
		}
		return
	}
	user, ok := h.mcpBrowser(w, r, mcpAdmin(row.Scopes), true)
	if !ok {
		return
	}
	if row.UserID != user.UserID {
		http.NotFound(w, r)
		return
	}
	var input struct {
		Allow bool `json:"allow"`
	}
	if platform.DecodeJSON(r, &input) != nil {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Choose Approve or Deny", nil)
		return
	}
	var n int64
	if input.Allow {
		n, e = h.queries.MCPApproveActionIntent(r.Context(), db.MCPApproveActionIntentParams{ID: id, UserID: user.UserID, ApprovedSessionID: user.SessionID})
	} else {
		n, e = h.queries.MCPDenyActionIntent(r.Context(), db.MCPDenyActionIntentParams{ID: id, UserID: user.UserID})
	}
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	if n == 0 {
		platform.RespondError(w, 409, "ACTION_CLOSED", "Action already handled or expired", nil)
		return
	}
	mcpReply(w, 200, struct {
		Approved bool   `json:"approved"`
		Message  string `json:"message"`
	}{input.Allow, "Return to your app. It can now check this action."})
}
