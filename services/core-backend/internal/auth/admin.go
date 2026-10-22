package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
)

// RequirePlatformAdmin requires a dedicated admin cookie backed by an explicit
// database grant. An ordinary account session cannot authenticate admin reads.
func RequirePlatformAdmin(sm *SessionManager, queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "private, no-store")
			token := GetSessionToken(r, AdminSessionCookie)
			if token == "" {
				platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			user, err := sm.GetSession(ctx, token)
			if err != nil {
				if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrInvalidToken) {
					platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
				} else {
					platform.RespondDatabaseError(w, err)
				}
				return
			}
			allowed, err := queries.IsPlatformAdminSession(ctx, db.IsPlatformAdminSessionParams{SessionID: user.SessionID, UserID: user.UserID})
			if err != nil {
				platform.RespondDatabaseError(w, err)
				return
			}
			if !allowed {
				platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "This account does not have admin access", nil)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(context.WithValue(ctx, platformAdminKey, true), userContextKey, user)))
		})
	}
}
