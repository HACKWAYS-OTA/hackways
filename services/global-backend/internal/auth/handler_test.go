package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/testsupport"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"

	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
		var wg sync.WaitGroup
		for i := range statuses {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				statuses[index] = call("/password-reset/confirm", map[string]string{"token": token, "password": nextPassword}).Code
			}(i)
		}
		wg.Wait()
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
