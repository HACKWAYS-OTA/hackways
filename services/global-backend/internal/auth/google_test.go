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
	})
}

func TestGoogleStartRejectsUntrustedReturnURL(t *testing.T) {
	h := testGoogleHandler()
	req := httptest.NewRequest(http.MethodGet, "/google/start?return_to=https%3A%2F%2Fevil.example%2Fcallback", nil)
	rr := httptest.NewRecorder()
	h.GoogleStart(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGoogleReturnURLValidation(t *testing.T) {
	h := testGoogleHandler()
	if !h.google.validReturnTo("http://localhost:3001/auth/google/callback?next=%2Faccount") {
		t.Fatal("expected allowed return URL")
	}
	if h.google.validReturnTo("http://localhost:3001/auth/google/callback#fragment") {
		t.Fatal("fragments must be rejected")
	}
	if h.google.validReturnTo("http://localhost:3001.evil.example/callback") {
		t.Fatal("untrusted host must be rejected")
	}
}

func TestGoogleNonceCookieIsHttpOnlyAndShortLived(t *testing.T) {
	h := testGoogleHandler()
	rr := httptest.NewRecorder()
	h.setGoogleNonceCookie(rr, "nonce")
	cookie := rr.Result().Cookies()[0]
	if cookie.Name != googleNonceCookieDev || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge <= 0 {
		t.Fatalf("unexpected nonce cookie: %#v", cookie)
	}
	if !strings.Contains(rr.Header().Get("Set-Cookie"), "Path=/") {
		t.Fatal("expected root-scoped nonce cookie")
	}
}
