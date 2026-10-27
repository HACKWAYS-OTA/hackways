package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/hackways/global-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"

	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Extends the mandatory PostgreSQL CI suite; all records live in its isolated
// database. This does not send email or touch production identities.
func TestMCPOAuthAndApprovedProfileWrite(t *testing.T) {
	t.Setenv("MCP_ENABLED", "true")
	ctx := context.Background()
	pool, e := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	q := db.New(pool)
	nonce, e := GenerateToken()
	if e != nil {
		t.Fatal(e)
	}
	user, e := q.CreateUser(ctx, db.CreateUserParams{Email: "mcp-" + nonce + "@example.invalid", PasswordHash: "unusable-isolated-fixture", DisplayName: "MCP account"})
	if e != nil {
		t.Fatal(e)
	}
	if e = q.MarkUserEmailVerified(ctx, user.ID); e != nil {
		t.Fatal(e)
	}
	h := NewHandler(pool, false)
	session, _, e := h.sm.CreateSession(ctx, user.ID, "MCP CI", "", time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	defer func() { _, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", user.ID) }()
	router := chi.NewRouter()
	router.Use(platform.RequestProtection(24))
	router.Mount("/api/v1/auth", h.Routes())
	request := func(method, path, body, content, origin, bearer string, cookie bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if content != "" {
			req.Header.Set("Content-Type", content)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		if cookie {
			req.AddCookie(&http.Cookie{Name: DefaultSessionCookie, Value: session})
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	registration := request("POST", "/api/v1/auth/mcp/register", `{"client_name":"Isolated MCP client","redirect_uris":["http://127.0.0.1:49123/callback"],"token_endpoint_auth_method":"none"}`, "application/json", "", "", false)
	if registration.Code != 201 {
		t.Fatalf("registration: %d %s", registration.Code, registration.Body.String())
	}
	var client mcpClientMetadata
	if json.Unmarshal(registration.Body.Bytes(), &client) != nil || client.ClientID == "" {
		t.Fatal("missing client ID")
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM mcp_oauth_requests WHERE client_id=$1", client.ClientID)
		_, _ = pool.Exec(ctx, "DELETE FROM mcp_oauth_grants WHERE client_id=$1", client.ClientID)
		_, _ = pool.Exec(ctx, "DELETE FROM mcp_oauth_clients WHERE id=$1", client.ClientID)
	}()
	verifier := strings.Repeat("x", 64)
	digest := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(digest[:])
	values := url.Values{"client_id": {client.ClientID}, "redirect_uri": {client.Redirects[0]}, "response_type": {"code"}, "resource": {MCPResource()}, "code_challenge_method": {"S256"}, "code_challenge": {challenge}, "scope": {"profile:read profile:write admin:read"}, "state": {"state-check"}}
	values.Set("redirect_uri", "https://unregistered.example.invalid/callback")
	if got := request("GET", "/api/v1/auth/mcp/authorize?"+values.Encode(), "", "", "", "", false).Code; got != 400 {
		t.Fatalf("redirect mismatch accepted: %d", got)
	}
	values.Set("redirect_uri", client.Redirects[0])
	authorization := request("GET", "/api/v1/auth/mcp/authorize?"+values.Encode(), "", "", "", "", false)
	if authorization.Code != 303 {
		t.Fatalf("authorize: %d %s", authorization.Code, authorization.Body.String())
	}
	destination, _ := url.Parse(authorization.Header().Get("Location"))
	if destination.Host != "hackways.in" {
		t.Fatal("ordinary account was forced into admin login")
	}
	id := destination.Query().Get("request")
	consent := request("GET", "/api/v1/auth/mcp/requests/"+id, "", "", "", "", true)
	if consent.Code != 200 {
		t.Fatalf("consent: %d", consent.Code)
	}
	var view mcpConsentView
	_ = json.Unmarshal(consent.Body.Bytes(), &view)
	if mcpAdmin(view.Scopes) || view.AdminURL == "" {
		t.Fatal("ordinary consent must omit admin scopes and offer a separate admin entry")
	}
	if got := request("POST", "/api/v1/auth/mcp/requests/"+id+"/decision", `{"allow":true}`, "application/json", "https://evil.example.invalid", "", true).Code; got != 403 {
		t.Fatalf("cross-origin approval accepted: %d", got)
	}
	decision := request("POST", "/api/v1/auth/mcp/requests/"+id+"/decision", `{"allow":true}`, "application/json", "https://hackways.in", "", true)
	if decision.Code != 200 {
		t.Fatalf("decision: %d %s", decision.Code, decision.Body.String())
	}
	var returned struct {
		URL string `json:"redirect_url"`
	}
	_ = json.Unmarshal(decision.Body.Bytes(), &returned)
	callback, _ := url.Parse(returned.URL)
	if callback.Query().Get("iss") != MCPIssuer() || callback.Query().Get("state") != "state-check" {
		t.Fatal("callback must bind issuer and preserve state")
	}
	form := url.Values{"grant_type": {"authorization_code"}, "client_id": {client.ClientID}, "redirect_uri": {client.Redirects[0]}, "resource": {MCPResource()}, "code": {callback.Query().Get("code")}, "code_verifier": {strings.Repeat("z", 64)}}
	if got := request("POST", "/api/v1/auth/mcp/token", form.Encode(), "application/x-www-form-urlencoded", "", "", false).Code; got != 400 {
		t.Fatalf("bad PKCE accepted: %d", got)
	}
	form.Set("code_verifier", verifier)
	exchanged := request("POST", "/api/v1/auth/mcp/token", form.Encode(), "application/x-www-form-urlencoded", "", "", false)
	if exchanged.Code != 200 {
		t.Fatalf("token: %d %s", exchanged.Code, exchanged.Body.String())
	}
	var tokens struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
		Scope   string `json:"scope"`
	}
	_ = json.Unmarshal(exchanged.Body.Bytes(), &tokens)
	if strings.Contains(tokens.Scope, "admin:") {
		t.Fatal("ordinary consent escalated scopes")
	}
	if got := request("POST", "/api/v1/auth/mcp/token", form.Encode(), "application/x-www-form-urlencoded", "", "", false).Code; got != 400 {
		t.Fatalf("code replay accepted: %d", got)
	}
	if got := request("GET", "/api/v1/auth/me", "", "", "", tokens.Access, false).Code; got != 200 {
		t.Fatalf("real profile API denied credential: %d", got)
	}
	body := `{"display_name":"Approved MCP name"}`
	if got := request("PATCH", "/api/v1/auth/me", body, "application/json", "", tokens.Access, false).Code; got != 403 {
		t.Fatalf("unapproved write accepted: %d", got)
	}
	if got := request("GET", "/api/v1/auth/mcp/connections", "", "", "", tokens.Access, false).Code; got != 401 {
		t.Fatalf("OAuth credential substituted for browser cookie: %d", got)
	}
	identity, e := MCPIdentity(ctx, q, "Bearer "+tokens.Access)
	if e != nil {
		t.Fatal(e)
	}
	var intentID pgtype.UUID
	if e = intentID.Scan(nonce[:8] + "-" + nonce[8:12] + "-" + nonce[12:16] + "-" + nonce[16:20] + "-" + nonce[20:32]); e != nil {
		t.Fatal(e)
	}
	_, e = q.MCPCreateActionIntent(ctx, db.MCPCreateActionIntentParams{ID: intentID, GrantID: identity.GrantID, ToolName: "update_my_name", ApiService: "global", Method: "PATCH", ApiPath: "/api/v1/auth/me", PayloadJson: []byte(body), RequestHash: HashToken(body)})
	if e != nil {
		t.Fatal(e)
	}
	approve := request("POST", "/api/v1/auth/mcp/actions/"+mcpIDString(intentID)+"/decision", `{"allow":true}`, "application/json", "https://hackways.in", "", true)
	if approve.Code != 200 {
		t.Fatalf("action approval: %d", approve.Code)
	}
	lease, e := GenerateToken()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = q.MCPClaimActionIntent(ctx, db.MCPClaimActionIntentParams{ID: intentID, GrantID: identity.GrantID, LeaseHash: mcpText(HashToken(lease))}); e != nil {
		t.Fatal(e)
	}
	execute := func(payload string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("PATCH", "/api/v1/auth/me", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokens.Access)
		req.Header.Set("X-MCP-Intent", mcpIDString(intentID))
		req.Header.Set("X-MCP-Lease", lease)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr
	}
	if got := execute(`{"display_name":"Changed by model after approval"}`).Code; got != 403 {
		t.Fatalf("changed approved fields accepted: %d", got)
	}
	if got := execute(body); got.Code != 200 {
		t.Fatalf("approved real write: %d %s", got.Code, got.Body.String())
	}
	saved, e := q.GetUserByEmail(ctx, user.Email)
	if e != nil || saved.DisplayName != "Approved MCP name" {
		t.Fatal("confirmed operation was not persisted")
	}
	refresh := url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientID}, "resource": {MCPResource()}, "refresh_token": {tokens.Refresh}}
	first := request("POST", "/api/v1/auth/mcp/token", refresh.Encode(), "application/x-www-form-urlencoded", "", "", false)
	if first.Code != 200 {
		t.Fatalf("refresh: %d %s", first.Code, first.Body.String())
	}
	if got := request("POST", "/api/v1/auth/mcp/token", refresh.Encode(), "application/x-www-form-urlencoded", "", "", false).Code; got != 400 {
		t.Fatalf("refresh reuse accepted: %d", got)
	}
	if got := request("GET", "/api/v1/auth/me", "", "", "", tokens.Access, false).Code; got != 401 {
		t.Fatalf("refresh reuse did not revoke grant: %d", got)
	}
	// Extend the existing mandatory OAuth journey: session-bound recognition
	// blocks bearer clients even without cookies, and review cannot revive grants.
	_, freshSession, e := h.sm.CreateSession(ctx, user.ID, "Restricted MCP CI", "", time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	freshGrant, e := q.MCPCreateGrant(ctx, db.MCPCreateGrantParams{ID: freshSession.ID, ClientID: client.ClientID, Resource: MCPResource(), Scopes: []string{"profile:read"}})
	if e != nil {
		t.Fatal(e)
	}
	opaque, _ := GenerateToken()
	access := "mcp_at_" + opaque
	if _, e = q.MCPInsertAccessToken(ctx, db.MCPInsertAccessTokenParams{TokenHash: HashToken(access), ID: freshGrant.ID}); e != nil {
		t.Fatal(e)
	}
	browser, _ := GenerateToken()
	deviceHash := HashToken(browser)
	if e = q.BindSessionDevice(ctx, db.BindSessionDeviceParams{ID: freshSession.ID, DeviceTokenHash: pgtype.Text{String: deviceHash, Valid: true}}); e != nil {
		t.Fatal(e)
	}
	if _, e = MCPIdentity(ctx, q, "Bearer "+access); e != nil {
		t.Fatal("fresh MCP credential invalid")
	}
	restrictedDevice, e := q.AdminBlockDevice(ctx, db.AdminBlockDeviceParams{TokenHash: deviceHash, UserAgent: "MCP isolated browser", Reason: "Isolated bearer recognition check", UpdatedBy: user.ID})
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM account_restrictions WHERE user_id=$1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM device_restrictions WHERE id=$1", restrictedDevice.ID)
	}()
	if got := request("GET", "/api/v1/auth/me", "", "", "", access, false).Code; got != 401 {
		t.Fatalf("bound browser bearer bypass %d", got)
	}
	if e = q.AdminUnblockDevice(ctx, db.AdminUnblockDeviceParams{ID: restrictedDevice.ID, Reason: "Isolated browser read check complete", UpdatedBy: user.ID}); e != nil {
		t.Fatal(e)
	}
	tx, e := pool.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer tx.Rollback(ctx)
	txq := q.WithTx(tx)
	if e = txq.AdminSetAccountRestriction(ctx, db.AdminSetAccountRestrictionParams{UserID: user.ID, Blocked: true, Reason: "Isolated MCP session retirement", UpdatedBy: user.ID}); e != nil {
		t.Fatal(e)
	}
	if e = txq.RetireRestrictedAccountSessions(ctx, []pgtype.UUID{user.ID}); e != nil {
		t.Fatal(e)
	}
	if e = txq.RetireRestrictedAccountGrants(ctx, []pgtype.UUID{user.ID}); e != nil {
		t.Fatal(e)
	}
	if e = tx.Commit(ctx); e != nil {
		t.Fatal(e)
	}
	if got := request("GET", "/api/v1/auth/me", "", "", "", access, false).Code; got != 401 {
		t.Fatal("restricted MCP bearer accepted")
	}
	if e = q.AdminSetAccountRestriction(ctx, db.AdminSetAccountRestrictionParams{UserID: user.ID, Blocked: false, Reason: "Isolated MCP account review", UpdatedBy: user.ID}); e != nil {
		t.Fatal(e)
	}
	if got := request("GET", "/api/v1/auth/me", "", "", "", access, false).Code; got != 401 {
		t.Fatal("account review revived retired MCP grant")
	}

}

func TestAuthHandler_LivePostgres(t *testing.T) {
	dbURL := testsupport.DatabaseURL(t)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping live postgres auth handler test: %v", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skipping live postgres auth handler test: %v", err)
		return
	}

	handler := NewHandler(pool, false)
	uniqueEmail := fmt.Sprintf("auth-test-%d@hackways.in", time.Now().UnixNano())
	password := "SecureP@ssw0rd123"
	defer func() { _, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE recipient_email=$1", uniqueEmail) }()
	displayName := "Live Auth Tester"

	var sessionToken string

	t.Run("Register success", func(t *testing.T) {
		body, _ := json.Marshal(RegisterRequest{
			Email:       uniqueEmail,
			Password:    password,
			DisplayName: displayName,
		})

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rr.Code, rr.Body.String())
		}

		cookies := rr.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatalf("expected session cookie set upon registration")
		}
		sessionToken = cookies[0].Value
		if cookies[0].MaxAge != int(UserSessionLifetime.Seconds()) {
			t.Fatal("signup cookie must persist for ninety days")
		}
	})

	t.Run("New account promotion defaults preserve an unsubscribe", func(t *testing.T) {
		q := db.New(pool)
		user, err := q.GetUserByEmail(ctx, uniqueEmail)
		if err != nil {
			t.Fatal(err)
		}
		subscribed, err := q.ReadPromotionPreference(ctx, uniqueEmail)
		if err != nil || !subscribed {
			t.Fatal("new account promotional email did not default on")
		}
		if err = q.SetPromotionPreference(ctx, db.SetPromotionPreferenceParams{UserID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Subscribed: false}); err != nil {
			t.Fatal(err)
		}
		if err = q.SetNewAccountPromotionPreference(ctx, db.SetNewAccountPromotionPreferenceParams{UserID: user.ID, Subscribed: true}); err != nil {
			t.Fatal(err)
		}
		if subscribed, err = q.ReadPromotionPreference(ctx, uniqueEmail); err != nil || subscribed {
			t.Fatal("default preference resubscribed an existing unsubscribe")
		}
	})
	t.Run("Signup can turn promotional email off", func(t *testing.T) {
		email := "off-" + uniqueEmail
		defer func() { _, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE recipient_email=$1", email) }()
		off := false
		raw, _ := json.Marshal(RegisterRequest{Email: email, Password: password, DisplayName: displayName, PromotionalEmail: &off})
		request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(raw))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.Routes().ServeHTTP(response, request)
		if response.Code != 201 {
			t.Fatalf("signup opt-out: %d", response.Code)
		}
		if subscribed, err := db.New(pool).ReadPromotionPreference(ctx, email); err != nil || subscribed {
			t.Fatal("signup opt-out ignored")
		}
	})
	t.Run("Google defaults only apply to newly created accounts", func(t *testing.T) {
		q := db.New(pool)
		email := "google-" + uniqueEmail
		info := googleUserInfo{Sub: "ci-google-" + uniqueEmail, Email: email, VerifiedEmail: true, Name: displayName}
		id, err := handler.upsertGoogleIdentity(ctx, info)
		if err != nil {
			t.Fatal(err)
		}
		if subscribed, err := q.ReadPromotionPreference(ctx, email); err != nil || !subscribed {
			t.Fatal("new Google account missing default promotion preference")
		}
		if err = q.SetPromotionPreference(ctx, db.SetPromotionPreferenceParams{UserID: id, Email: email, DisplayName: displayName, Subscribed: false}); err != nil {
			t.Fatal(err)
		}
		if _, err = handler.upsertGoogleIdentity(ctx, info); err != nil {
			t.Fatal(err)
		}
		if subscribed, err := q.ReadPromotionPreference(ctx, email); err != nil || subscribed {
			t.Fatal("Google login resubscribed an unsubscribe")
		}
	})
	t.Run("Duplicate email returns 409", func(t *testing.T) {
		body, _ := json.Marshal(RegisterRequest{
			Email:       uniqueEmail,
			Password:    password,
			DisplayName: displayName,
		})

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 Conflict, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Login success", func(t *testing.T) {
		body, _ := json.Marshal(LoginRequest{
			Email:    uniqueEmail,
			Password: password,
		})

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		cookies := rr.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatalf("expected session cookie set upon login")
		}
		sessionToken = cookies[0].Value
		if cookies[0].MaxAge != int(UserSessionLifetime.Seconds()) {
			t.Fatal("sign-in cookie must persist for ninety days")
		}
	})

	t.Run("Login invalid password returns 401", func(t *testing.T) {
		body, _ := json.Marshal(LoginRequest{
			Email:    uniqueEmail,
			Password: "WrongPassword123",
		})

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Me returns authenticated user with valid cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.AddCookie(&http.Cookie{
			Name:  DefaultSessionCookie,
			Value: sessionToken,
		})
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var res struct {
			User UserResponse `json:"user"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if res.User.Email != uniqueEmail || res.User.DisplayName != displayName {
			t.Fatalf("unexpected user in response: %+v", res.User)
		}
	})

	t.Run("Verify endpoint validates token", func(t *testing.T) {
		body, _ := json.Marshal(VerifyRequest{Token: sessionToken})
		req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Logout revokes session", func(t *testing.T) {
		// Profile mutations must require a session and preserve account identity.
		profile := func(payload string, signedIn bool) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodPatch, "/me", bytes.NewBufferString(payload))
			req.Header.Set("Content-Type", "application/json")
			if signedIn {
				req.AddCookie(&http.Cookie{Name: DefaultSessionCookie, Value: sessionToken})
			}
			rr := httptest.NewRecorder()
			handler.Routes().ServeHTTP(rr, req)
			return rr
		}
		if rr := profile(`{"display_name":"Someone else"}`, false); rr.Code != 401 {
			t.Fatalf("anonymous profile mutation: %d", rr.Code)
		}
		if rr := profile(`{"display_name":"  "}`, true); rr.Code != 422 {
			t.Fatalf("empty display name: %d", rr.Code)
		}
		if rr := profile(`{"display_name":"Updated profile"}`, true); rr.Code != 200 {
			t.Fatalf("profile mutation: %d %s", rr.Code, rr.Body.String())
		}
		me := httptest.NewRequest(http.MethodGet, "/me", nil)
		me.AddCookie(&http.Cookie{Name: DefaultSessionCookie, Value: sessionToken})
		result := httptest.NewRecorder()
		handler.Routes().ServeHTTP(result, me)
		var updated struct {
			User UserResponse `json:"user"`
		}
		if err := json.Unmarshal(result.Body.Bytes(), &updated); err != nil || updated.User.DisplayName != "Updated profile" || updated.User.Email != uniqueEmail {
			t.Fatal("profile did not persist with identity preserved")
		}
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.AddCookie(&http.Cookie{
			Name:  DefaultSessionCookie,
			Value: sessionToken,
		})
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rr.Code)
		}

		// Subsequent /me should fail with 401
		reqMe := httptest.NewRequest(http.MethodGet, "/me", nil)
		reqMe.AddCookie(&http.Cookie{
			Name:  DefaultSessionCookie,
			Value: sessionToken,
		})
		rrMe := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rrMe, reqMe)

		if rrMe.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized after logout, got %d", rrMe.Code)
		}
	})
}

// Extends the mandatory PostgreSQL auth suite; never uses the application database.
func TestAuthRecoveryAndRecognizedDevices_LivePostgres(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err = pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ACS_ENDPOINT", "https://email-ci.example.invalid")
	t.Setenv("ACS_ACCESS_KEY", "ci-does-not-send")
	t.Setenv("ACS_SENDER_ADDRESS", "sender@example.invalid")
	h := NewHandler(pool, false)
	q := db.New(pool)
	email := fmt.Sprintf("auth-recovery-%d@example.invalid", time.Now().UnixNano())
	password := "RecoveryOriginal123!"
	nextPassword := "RecoveryUpdated123!"
	defer func() { _, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE recipient_email=$1", email) }()
	call := func(path string, body map[string]string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("User-Agent", "Mozilla/5.0 Windows Chrome/130.0")
		for _, c := range cookies {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		h.Routes().ServeHTTP(w, r)
		return w
	}
	cookie := func(w *httptest.ResponseRecorder, name string) *http.Cookie {
		for _, c := range w.Result().Cookies() {
			if c.Name == name {
				return c
			}
		}
		t.Fatalf("missing %s cookie", name)
		return nil
	}
	readCode := func(template string) otpPayload {
		var payload []byte
		if e := pool.QueryRow(ctx, "SELECT payload_json FROM email_outbox WHERE recipient_email=$1 AND template_type=$2 ORDER BY created_at DESC,id DESC LIMIT 1", email, template).Scan(&payload); e != nil {
			t.Fatal(e)
		}
		var code otpPayload
		if e := json.Unmarshal(payload, &code); e != nil {
			t.Fatal(e)
		}
		if !validEmailCode(code.Code) || code.Hash == code.Code || !matchesEmailCode(code.Code, code.Hash) {
			t.Fatal("OTP generation or secure storage contract failed")
		}
		return code
	}
	readResetToken := func() string {
		var payload []byte
		if e := pool.QueryRow(ctx, "SELECT payload_json FROM email_outbox WHERE recipient_email=$1 AND template_type='password_reset_requested' ORDER BY created_at DESC,id DESC LIMIT 1", email).Scan(&payload); e != nil {
			t.Fatal(e)
		}
		var link resetLinkPayload
		if e := json.Unmarshal(payload, &link); e != nil {
			t.Fatal(e)
		}
		parsed, e := url.Parse(link.ActionURL)
		if e != nil || parsed.Scheme != "https" || parsed.Host != "hackways.in" || parsed.Path != "/auth/reset-password" {
			t.Fatal("unsafe reset URL")
		}
		token := parsed.Query().Get("key")
		if !validResetToken(token) || link.Hash != HashToken(token) || strings.Contains(string(payload), "verification_code") {
			t.Fatal("reset link token contract failed")
		}
		var stored string
		if e := pool.QueryRow(ctx, "SELECT token_hash FROM password_reset_challenges WHERE user_id=(SELECT id FROM users WHERE email=$1)", email).Scan(&stored); e != nil || stored != HashToken(token) {
			t.Fatal("reset token must be hashed at rest in challenge")
		}
		return token
	}
	countAlerts := func() int {
		var n int
		if e := pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox WHERE recipient_email=$1 AND template_type='sign_in_alert'", email).Scan(&n); e != nil {
			t.Fatal(e)
		}
		return n
	}
	resetRate := func() {
		sum := sha256.Sum256([]byte("identity:" + email))
		if _, e := pool.Exec(ctx, "DELETE FROM auth_rate_limits WHERE scope_hash=$1", hex.EncodeToString(sum[:])); e != nil {
			t.Fatal(e)
		}
	}
	registered := call("/register", map[string]string{"email": email, "password": password, "display_name": "Auth CI"})
	if registered.Code != 201 {
		t.Fatalf("registration: %d", registered.Code)
	}
	session, device := cookie(registered, DefaultSessionCookie), cookie(registered, deviceCookie)
	code := readCode("email_verification")
	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	if countAlerts() != 0 {
		t.Fatal("signup baseline sent a new-device alert")
	}
	t.Run("verification is one use and owned", func(t *testing.T) {
		wrong := "000000"
		if code.Code == wrong {
			wrong = "111111"
		}
		if w := call("/email-verification/confirm", map[string]string{"code": wrong}, session); w.Code != 400 {
			t.Fatalf("wrong OTP status %d", w.Code)
		}
		if w := call("/email-verification/confirm", map[string]string{"code": code.Code}, session); w.Code != 200 {
			t.Fatalf("verify status %d", w.Code)
		}
		if w := call("/email-verification/confirm", map[string]string{"code": code.Code}, session); w.Code != 409 {
			t.Fatalf("OTP replay %d", w.Code)
		}
		verified, e := q.GetUserByEmail(ctx, email)
		if e != nil || !verified.EmailVerified {
			t.Fatal("verification did not persist")
		}
	})
	t.Run("regular browser is not repeatedly alerted", func(t *testing.T) {
		resetRate()
		if w := call("/logout", map[string]string{}, session); w.Code != 200 {
			t.Fatal("logout")
		}
		regular := call("/login", map[string]string{"email": email, "password": password}, device)
		if regular.Code != 200 || countAlerts() != 0 {
			t.Fatal("regular browser generated alert")
		}
		session = cookie(regular, DefaultSessionCookie)
		fresh := call("/login", map[string]string{"email": email, "password": password})
		if fresh.Code != 200 || countAlerts() != 1 {
			t.Fatal("new browser did not generate exactly one alert")
		}
		nextDevice := cookie(fresh, deviceCookie)
		for i := 0; i < 2; i++ {
			if w := call("/login", map[string]string{"email": email, "password": password}, nextDevice); w.Code != 200 {
				t.Fatal("recognized login failed")
			}
		}
		if countAlerts() != 1 {
			t.Fatal("recognized device produced repeated alerts")
		}
		if w := call("/login", map[string]string{"email": email, "password": "incorrect"}); w.Code != 401 || countAlerts() != 1 {
			t.Fatal("failed login created alert")
		}
	})
	t.Run("reset requests do not enumerate accounts or spam", func(t *testing.T) {
		resetRate()
		known := call("/password-reset/request", map[string]string{"email": email})
		unknown := call("/password-reset/request", map[string]string{"email": "unknown-" + email})
		if known.Code != 202 || unknown.Code != 202 || known.Body.String() != unknown.Body.String() {
			t.Fatal("recovery request enumerated accounts")
		}
		if w := call("/password-reset/request", map[string]string{"email": email}); w.Code != 202 {
			t.Fatal("cooldown request response changed")
		}
		var n int
		if e := pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox WHERE recipient_email=$1 AND template_type='password_reset_requested'", email).Scan(&n); e != nil || n != 1 {
			t.Fatal("reset cooldown did not deduplicate email")
		}
	})
	t.Run("invalid expired and superseded links are rejected", func(t *testing.T) {
		resetRate()
		token := readResetToken()
		wrong, e := GenerateToken()
		if e != nil {
			t.Fatal(e)
		}
		if w := call("/password-reset/confirm", map[string]string{"token": wrong, "password": nextPassword}); w.Code != 400 {
			t.Fatal("invalid token accepted")
		}
		if _, e := pool.Exec(ctx, "UPDATE password_reset_challenges SET expires_at=clock_timestamp()-INTERVAL '1 minute',created_at=clock_timestamp()-INTERVAL '2 minutes' WHERE user_id=$1", user.ID); e != nil {
			t.Fatal(e)
		}
		if w := call("/password-reset/confirm", map[string]string{"token": token, "password": nextPassword}); w.Code != 400 {
			t.Fatal("expired token accepted")
		}
		if w := call("/password-reset/request", map[string]string{"email": email}); w.Code != 202 {
			t.Fatal("replacement link failed")
		}
		if replacement := readResetToken(); replacement == token {
			t.Fatal("reset token reused")
		}
		if w := call("/password-reset/confirm", map[string]string{"token": token, "password": nextPassword}); w.Code != 400 {
			t.Fatal("superseded token accepted")
		}
		current, e := q.GetUserByEmail(ctx, email)
		if e != nil || !CheckPasswordHash(password, current.PasswordHash) {
			t.Fatal("invalid link changed password")
		}
	})
	t.Run("one reset wins and all sessions are revoked", func(t *testing.T) {
		resetRate()
		if _, e := pool.Exec(ctx, "UPDATE password_reset_challenges SET created_at=clock_timestamp()-INTERVAL '2 minutes' WHERE user_id=$1", user.ID); e != nil {
			t.Fatal(e)
		}
		if w := call("/password-reset/request", map[string]string{"email": email}); w.Code != 202 {
			t.Fatal("new reset request failed")
		}
		token := readResetToken()
		var statuses [2]int
		var clearedRecognition [2]bool
		var wg sync.WaitGroup
		for i := range statuses {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				result := call("/password-reset/confirm", map[string]string{"token": token, "password": nextPassword}, device)
				statuses[index] = result.Code
				for _, cookie := range result.Result().Cookies() {
					if cookie.Name == deviceCookie && cookie.MaxAge < 0 {
						clearedRecognition[index] = true
					}
				}
			}(i)
		}
		wg.Wait()
		if clearedRecognition[0] || clearedRecognition[1] {
			t.Fatal("password recovery cleared browser recognition")
		}
		if !(statuses[0] == 200 && statuses[1] == 400 || statuses[1] == 200 && statuses[0] == 400) {
			t.Fatalf("concurrent reset statuses %v", statuses)
		}
		var sessions, devices int
		if e := pool.QueryRow(ctx, "SELECT count(*) FROM sessions WHERE user_id=$1", user.ID).Scan(&sessions); e != nil {
			t.Fatal(e)
		}
		if e := pool.QueryRow(ctx, "SELECT count(*) FROM recognized_devices WHERE user_id=$1", user.ID).Scan(&devices); e != nil || sessions != 0 || devices != 0 {
			t.Fatal("reset did not revoke sessions and devices")
		}
		if _, e := h.sm.GetSession(ctx, session.Value); e == nil {
			t.Fatal("old session survived reset")
		}
		if w := call("/password-reset/confirm", map[string]string{"token": token, "password": password}); w.Code != 400 {
			t.Fatal("reset replay accepted")
		}
		if w := call("/login", map[string]string{"email": email, "password": password}, device); w.Code != 401 {
			t.Fatal("old password accepted")
		}
		if w := call("/login", map[string]string{"email": email, "password": nextPassword}, device); w.Code != 200 {
			t.Fatalf("new password failed %d", w.Code)
		}
		var notices int
		if e := pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox WHERE recipient_email=$1 AND template_type='password_changed'", email).Scan(&notices); e != nil || notices != 1 {
			t.Fatal("password change notification missing or duplicated")
		}
	})
	t.Run("verification ownership and failure limit", func(t *testing.T) {
		otherEmail := fmt.Sprintf("otp-limit-%d@example.invalid", time.Now().UnixNano())
		defer func() { _, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE recipient_email=$1", otherEmail) }()
		account := call("/register", map[string]string{"email": otherEmail, "password": password, "display_name": "OTP CI"})
		if account.Code != 201 {
			t.Fatal("OTP account registration failed")
		}
		ownSession := cookie(account, DefaultSessionCookie)
		var payload []byte
		if e := pool.QueryRow(ctx, "SELECT payload_json FROM email_outbox WHERE recipient_email=$1 AND template_type='email_verification'", otherEmail).Scan(&payload); e != nil {
			t.Fatal(e)
		}
		var ownCode otpPayload
		if e := json.Unmarshal(payload, &ownCode); e != nil {
			t.Fatal(e)
		}
		foreign := code.Code
		if foreign == ownCode.Code {
			foreign = "999999"
			if foreign == ownCode.Code {
				foreign = "111111"
			}
		}
		for i := 0; i < 5; i++ {
			if w := call("/email-verification/confirm", map[string]string{"code": foreign}, ownSession); w.Code != 400 {
				t.Fatalf("wrong verification attempt %d", w.Code)
			}
		}
		if w := call("/email-verification/confirm", map[string]string{"code": ownCode.Code}, ownSession); w.Code != 409 {
			t.Fatal("locked verification code accepted")
		}
		if w := call("/email-verification/request", map[string]string{}, ownSession); w.Code != 429 {
			t.Fatal("verification resend cooldown not enforced")
		}
		saved, e := q.GetUserByEmail(ctx, otherEmail)
		if e != nil || saved.EmailVerified {
			t.Fatal("foreign/locked OTP verified account")
		}
	})
	t.Run("email unavailable is explicit", func(t *testing.T) {
		resetRate()
		t.Setenv("ACS_ENDPOINT", "")
		t.Setenv("ACS_ACCESS_KEY", "")
		t.Setenv("ACS_SENDER_ADDRESS", "")
		t.Setenv("AZURE_ACS_CONNECTION_STRING", "")
		t.Setenv("AZURE_ACS_SENDER_ADDRESS", "")
		if w := call("/password-reset/request", map[string]string{"email": email}); w.Code != 503 || !strings.Contains(w.Body.String(), "EMAIL_UNAVAILABLE") {
			t.Fatal("unconfigured delivery reported success")
		}
	})
}

// Uses the current documented production CIMD transition shape. This is part of
// mandatory CI and never replaces real client metadata in application paths.
func TestMCPCIMDPublicMethodIntersection(t *testing.T) {
	for _, test := range []struct {
		name    string
		methods []string
		legacy  string
		allowed bool
	}{
		{"production plural supports public", []string{"none", "private_key_jwt"}, "private_key_jwt", true},
		{"private assertion only", []string{"private_key_jwt"}, "private_key_jwt", false},
		{"empty plural", []string{}, "none", false},
		{"legacy public", nil, "none", true},
		{"legacy unsupported", nil, "private_key_jwt", false},
	} {
		c := mcpClientMetadata{ClientID: "https://chatgpt.com/oauth/client.json", Name: "ChatGPT", Redirects: []string{"https://chatgpt.com/connector_platform_oauth_redirect"}, Method: test.legacy, MethodsSupported: test.methods, Grants: []string{"authorization_code", "refresh_token"}, Responses: []string{"code"}}
		if validMCPCIMDClient(&c) != test.allowed {
			t.Fatalf("method selection: %s", test.name)
		}
		if test.allowed && c.Method != "none" {
			t.Fatal("selected unsupported client authentication")
		}
	}
}
