package auth

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
	"time"

	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
)

// AdminLogin establishes a separate host-only admin session requiring daily OTP.
// Only explicitly granted, verified accounts can create one.
func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	var req LoginRequest
	if err := platform.DecodeJSON(r, &req); err != nil {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Enter your email and password", nil)
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if len(req.Email) == 0 || len(req.Email) > 320 || len(req.Password) == 0 || len(req.Password) > 72 {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Enter your email and password", nil)
		return
	}
	if !h.allowIdentityAttempt(w, r, req.Email) {
		return
	}
	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			platform.RespondError(w, 401, "INVALID_CREDENTIALS", "Sign in failed", nil)
		} else {
			platform.RespondDatabaseError(w, err)
		}
		return
	}
	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		platform.RespondError(w, 401, "INVALID_CREDENTIALS", "Sign in failed", nil)
		return
	}
	if err = h.establishSession(w, r, user.ID, user.PasswordHash, true); err != nil {
		switch {
		case errors.Is(err, ErrNotAllowed):
			respondNotAllowed(w)
		case errors.Is(err, errAdminAccess):
			platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "This account does not have admin access", nil)
		case errors.Is(err, errCredentialsChanged):
			platform.RespondError(w, 401, "INVALID_CREDENTIALS", "Sign in again with your current password", nil)
		default:
			platform.RespondDatabaseError(w, err)
		}
		return
	}

	platform.RespondJSON(w, 200, struct {
		DisplayName string `json:"display_name"`
	}{user.DisplayName})
}

func (h *Handler) AdminLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	token := GetSessionToken(r, AdminSessionCookie)
	if token != "" {
		if err := h.sm.RevokeSession(r.Context(), token); err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
	}
	ClearSessionCookie(w, AdminSessionCookie)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminMe(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
		return
	}
	platform.RespondJSON(w, 200, struct {
		DisplayName string `json:"display_name"`
	}{user.DisplayName})
}

// AdminEntry distinguishes a signed-in account without an admin grant from a
// visitor who needs the dedicated admin sign-in. Ordinary sessions never grant
// admin access, including ordinary sessions belonging to administrators.
func (h *Handler) AdminEntry(w http.ResponseWriter, r *http.Request) {
	if GetSessionToken(r, AdminSessionCookie) != "" {
		RequirePlatformAdmin(h.sm, h.queries)(http.HandlerFunc(h.AdminMe)).ServeHTTP(w, r)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	if token := GetSessionToken(r, DefaultSessionCookie); token != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		user, err := h.sm.GetSession(ctx, token)
		if errors.Is(err, ErrNotAllowed) {
			respondNotAllowed(w)
			return
		}
		if err != nil && !errors.Is(err, ErrSessionNotFound) && !errors.Is(err, ErrInvalidToken) {
			platform.RespondDatabaseError(w, err)
			return
		}
		if err == nil {
			allowed, queryErr := h.queries.IsPlatformAdministrator(ctx, user.UserID)
			if queryErr != nil {
				platform.RespondDatabaseError(w, queryErr)
				return
			}
			if !allowed {
				platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "This account does not have admin access", nil)
				return
			}
		}
	}
	platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
}

func (h *Handler) AdminRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.AdminLogin)
	r.Post("/logout", h.AdminLogout)
	r.Get("/me", h.AdminEntry)
	r.With(RequireAdminIdentity(h.sm, h.queries)).Post("/verification/request", h.RequestAdminVerification)
	r.With(RequireAdminIdentity(h.sm, h.queries)).Post("/verification/confirm", h.ConfirmAdminVerification)
	r.With(RequirePlatformAdmin(h.sm, h.queries)).Get("/session", h.RefreshAdminSessionCookie)
	return r
}
