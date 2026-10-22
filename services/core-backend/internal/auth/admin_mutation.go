package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
)

type adminContextKey uint8

const (
	platformAdminKey adminContextKey = iota
	adminMutationKey
)

type AdminMutation struct{ Reason, Key, Scope, RequestHash string }

func AdminMutationFromContext(ctx context.Context) (AdminMutation, bool) {
	m, ok := ctx.Value(adminMutationKey).(AdminMutation)
	return m, ok && ctx.Value(platformAdminKey) == true
}

// WithAdminManagedOrganizer scopes an already authorized administrator to one
// resource owner. It never changes the authenticated user or audit actor.
func WithAdminManagedOrganizer(ctx context.Context, owner *db.Organizers) (context.Context, bool) {
	if ctx.Value(platformAdminKey) != true || owner == nil {
		return ctx, false
	}
	return context.WithValue(ctx, organizerContextKey, owner), true
}

// RequireAdminMutation rejects cross-origin writes and binds retries to their
// authenticated actor, clean route, exact request body and recorded reason.
func RequireAdminMutation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := UserFromContext(r.Context())
		if !ok || r.Context().Value(platformAdminKey) != true {
			platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "Admin access required", nil)
			return
		}
		origin := r.Header.Get("Origin")
		allowed := origin == "https://admin.hackways.in"
		if os.Getenv("APP_ENV") == "development" {
			allowed = allowed || origin == "http://localhost:3003" || origin == "http://127.0.0.1:3003"
		}
		if !allowed {
			platform.RespondError(w, 403, "INVALID_ORIGIN", "This request origin is not allowed", nil)
			return
		}
		reason, err := url.PathUnescape(r.Header.Get("X-Admin-Reason"))
		reason = strings.TrimSpace(reason)
		if err != nil || !utf8.ValidString(reason) || utf8.RuneCountInString(reason) < 3 || utf8.RuneCountInString(reason) > 1000 || strings.ContainsAny(reason, "\x00\r\n") {
			platform.RespondError(w, 422, "REASON_REQUIRED", "Provide a reason of 3 to 1000 characters", nil)
			return
		}
		key := r.Header.Get("Idempotency-Key")
		if len(key) < 16 || len(key) > 128 {
			platform.RespondError(w, 400, "INVALID_IDEMPOTENCY_KEY", "Provide a random request key of 16 to 128 printable characters", nil)
			return
		}
		for _, c := range key {
			if c < 33 || c > 126 {
				platform.RespondError(w, 400, "INVALID_IDEMPOTENCY_KEY", "Provide a printable request key", nil)
				return
			}
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 128*1024))
		if err != nil {
			platform.RespondError(w, 413, "PAYLOAD_TOO_LARGE", "Request exceeds its limit", nil)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		digest := func(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
		m := AdminMutation{Reason: reason, Key: digest(key), Scope: digest("admin:" + hex.EncodeToString(actor.UserID.Bytes[:]) + ":" + r.Method + ":" + r.URL.Path), RequestHash: digest(reason + "\n" + string(body))}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), adminMutationKey, m)))
	})
}
