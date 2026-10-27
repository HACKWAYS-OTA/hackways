package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
)

// RequirePlatformAdmin requires a dedicated admin cookie backed by an explicit
// database grant. An ordinary account session cannot authenticate admin reads.
func RequirePlatformAdmin(sm *SessionManager, queries *db.Queries) func(http.Handler) http.Handler {
	return requireAdminSession(sm, queries, false)
}

// Identity-only access is restricted to requesting and confirming daily OTPs.
func RequireAdminIdentity(sm *SessionManager, queries *db.Queries) func(http.Handler) http.Handler {
	return requireAdminSession(sm, queries, true)
}

func requireAdminSession(sm *SessionManager, queries *db.Queries, allowVerification bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "private, no-store")
			var oauthUser *SessionUser
			if r.Header.Get("Authorization") != "" {
				if allowVerification {
					platform.RespondError(w, 403, "BROWSER_REQUIRED", "Admin verification requires a browser", nil)
					return
				}
				var ok bool
				oauthUser, ok = mcpRequestUser(w, r, sm)
				if !ok {
					return
				}
			}
			token := GetSessionToken(r, AdminSessionCookie)
			if token == "" && oauthUser == nil {
				platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			user := oauthUser
			var err error
			if user == nil {
				user, err = sm.GetSession(ctx, token)
			}
			if err != nil {
				if errors.Is(err, ErrNotAllowed) {
					respondNotAllowed(w)
				} else if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrInvalidToken) {
					platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
				} else {
					platform.RespondDatabaseError(w, err)
				}
				return
			}
			allowed, err := queries.IsPlatformAdminIdentitySession(ctx, db.IsPlatformAdminIdentitySessionParams{SessionID: user.SessionID, UserID: user.UserID})
			if err != nil {
				platform.RespondDatabaseError(w, err)
				return
			}
			if !allowed {
				platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "This account does not have admin access", nil)
				return
			}
			if !allowVerification {
				fresh, e := queries.IsPlatformAdminSession(ctx, db.IsPlatformAdminSessionParams{SessionID: user.SessionID, UserID: user.UserID})
				if e != nil {
					platform.RespondDatabaseError(w, e)
					return
				}
				if !fresh {
					platform.RespondError(w, 403, "ADMIN_VERIFICATION_REQUIRED", "Verify your admin sign-in with the daily email code", nil)
					return
				}
			}
			// The three-second budget bounds identity reads, not the authenticated operation.
			cancel()
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
		})
	}
}
