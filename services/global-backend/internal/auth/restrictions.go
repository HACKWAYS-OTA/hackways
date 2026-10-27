package auth

import (
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
)

func respondNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-store")
	platform.RespondError(w, 403, "ACCOUNT_RESTRICTED", "You are not allowed", nil)
}

func requestDeviceHash(r *http.Request) string {
	token := GetSessionToken(r, "__Host-hackways_device")
	decoded, err := hex.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		return ""
	}
	return HashToken(token)
}

func allowRequestDevice(w http.ResponseWriter, r *http.Request, q *db.Queries) bool {
	hash := requestDeviceHash(r)
	if hash == "" {
		return true
	}
	blocked, err := q.IsRestrictedDevice(r.Context(), hash)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return false
	}
	if blocked {
		respondNotAllowed(w)
		return false
	}
	return true
}

// Public actions also include anonymous registration/application paths. Sign-in
// credential proofs reach their transaction so account-switch quarantine can
// commit before denial. Recovery never clears the recognized browser cookie.
func RestrictionGuard(q *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions || strings.HasPrefix(path, "/api/v1/admin/") || strings.HasPrefix(path, "/api/v1/auth/admin/") {
				next.ServeHTTP(w, r)
				return
			}
			switch path {
			case "/api/v1/auth/login", "/api/v1/auth/google/exchange", "/api/v1/auth/logout", "/api/v1/auth/password-reset/request", "/api/v1/auth/password-reset/confirm":
				next.ServeHTTP(w, r)
				return
			}
			if allowRequestDevice(w, r, q) {
				next.ServeHTTP(w, r)
			}
		})
	}
}
