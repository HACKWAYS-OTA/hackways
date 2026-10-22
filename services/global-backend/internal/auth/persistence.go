package auth

import (
	"github.com/hackways/global-backend/internal/platform"
	"net/http"
	"time"
)

const UserSessionLifetime = 90 * 24 * time.Hour

// RefreshSessionCookie preserves the existing opaque credential and its fixed
// server expiry. It never creates a new session or extends a revoked credential.
func (h *Handler) RefreshSessionCookie(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	token := GetSessionToken(r, DefaultSessionCookie)
	if !ok || token == "" {
		platform.RespondError(w, 401, "UNAUTHORIZED", "Sign in to continue", nil)
		return
	}
	remaining := time.Until(user.ExpiresAt)
	if remaining < time.Second {
		platform.RespondError(w, 401, "UNAUTHORIZED", "Session expired", nil)
		return
	}
	SetSessionCookie(w, DefaultSessionCookie, token, remaining, h.isProduction)
	w.WriteHeader(http.StatusNoContent)
}
