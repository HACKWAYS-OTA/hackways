package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hackways/global-backend/internal/testsupport"
	"net/http"
	"net/http/httptest"

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
