package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testGoogleHandler() *Handler {
	return NewHandler(nil, false, GoogleOAuthConfig{
		ClientID: "client-id", ClientSecret: "client-secret",
		RedirectURL:   "http://localhost:8082/api/v1/auth/google/callback",
		ReturnOrigins: []string{"http://localhost:3001", "http://localhost:3002"},
		SessionSecret: "development-session-secret",
	})
}

func TestGoogleStartRejectsUntrustedReturnURL(t *testing.T) {
	h := testGoogleHandler()
	req := httptest.NewRequest(http.MethodGet, "/google/start?return_to=https%3A%2F%2Fevil.example%2Fauth%2Fgoogle%2Fcallback", nil)
	rr := httptest.NewRecorder()
	h.GoogleStart(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGoogleStartSetsSignedStateAndPKCE(t *testing.T) {
	h := testGoogleHandler()
	req := httptest.NewRequest(http.MethodGet, "/google/start?return_to=http%3A%2F%2Flocalhost%3A3001%2Fauth%2Fgoogle%2Fcallback%3Fnext%3D%252Faccount", nil)
	rr := httptest.NewRecorder()
	h.GoogleStart(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Location"), "accounts.google.com") || !strings.Contains(rr.Header().Get("Location"), "code_challenge=") {
		t.Fatalf("expected Google authorization URL with PKCE, got %q", rr.Header().Get("Location"))
	}
	cookies := rr.Result().Cookies()
	if len(cookies) != 1 || cookies[0].HttpOnly != true || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected a secure, http-only OAuth state cookie, got %#v", cookies)
	}
	payload, err := h.readGoogleStateCookie(httptest.NewRequest(http.MethodGet, "/", nil))
	_ = payload
	_ = err
	if err == nil {
		t.Fatal("expected missing state cookie to be rejected")
	}
}

func TestGoogleStateCookieTamperRejected(t *testing.T) {
	h := testGoogleHandler()
	rr := httptest.NewRecorder()
	if err := h.setGoogleStateCookie(rr, googleState{State: "state", Verifier: "verifier", ReturnTo: "http://localhost:3001/account", ExpiresAt: 2_000_000_000}); err != nil {
		t.Fatal(err)
	}
	cookie := rr.Result().Cookies()[0]
	cookie.Value = cookie.Value + "tampered"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	if _, err := h.readGoogleStateCookie(req); err == nil {
		t.Fatal("expected tampered OAuth state cookie to be rejected")
	}
}
