package auth

import (
	"net/http"
	"time"
)

const (
	// DefaultSessionCookie is the host-scoped cookie name for attendee and public sessions.
	DefaultSessionCookie = "__Host-hackways_session"
	// AdminSessionCookie is the host-scoped cookie name for platform administration.
	AdminSessionCookie = "__Host-hackways_admin_session"
)

// SetSessionCookie emits an HTTP-only, host-scoped session cookie on the response.
func SetSessionCookie(w http.ResponseWriter, cookieName, token string, maxAge time.Duration, isProduction bool) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // __Host- prefix strictly requires Secure: true (supported on localhost in modern browsers)
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge.Seconds()),
		Expires:  time.Now().Add(maxAge),
	}

	if cookieName == AdminSessionCookie {
		cookie.SameSite = http.SameSiteStrictMode
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookie deletes the session cookie from the client browser.
func ClearSessionCookie(w http.ResponseWriter, cookieName string) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}

	if cookieName == AdminSessionCookie {
		cookie.SameSite = http.SameSiteStrictMode
	}

	http.SetCookie(w, cookie)
}

// GetSessionToken extracts the session token from the specified cookie.
func GetSessionToken(r *http.Request, cookieName string) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return cookie.Value
}
