package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const (
	userContextKey      contextKey = "auth_user"
	organizerContextKey contextKey = "auth_organizer"
)

// UserFromContext retrieves the authenticated SessionUser from the request context.
func UserFromContext(ctx context.Context) (*SessionUser, bool) {
	u, ok := ctx.Value(userContextKey).(*SessionUser)
	return u, ok && u != nil
}

// OrganizerFromContext retrieves the authenticated Organizer from the request context.
func OrganizerFromContext(ctx context.Context) (*db.Organizers, bool) {
	o, ok := ctx.Value(organizerContextKey).(*db.Organizers)
	return o, ok && o != nil
}

// RequireAuth middleware verifies the presence and validity of an authenticated session cookie.
func RequireAuth(sm *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := GetSessionToken(r, DefaultSessionCookie)
			if token == "" {
				// Also check fallback or admin cookie
				token = GetSessionToken(r, AdminSessionCookie)
			}

			if token == "" {
				platform.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
				return
			}

			user, err := sm.GetSession(r.Context(), token)
			if err != nil {
				platform.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Session expired or invalid", nil)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireOrganizer middleware verifies that the authenticated user possesses an organizer profile.
func RequireOrganizer(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				platform.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
				return
			}

			organizer, err := queries.GetOrganizerByUserID(r.Context(), user.UserID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					platform.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Organizer profile required to access this resource", nil)
					return
				}
				platform.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify organizer permissions", nil)
				return
			}

			ctx := context.WithValue(r.Context(), organizerContextKey, &organizer)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ValidateEventOwnership verifies that an event belongs authoritatively to the requesting organizer.
// Returns ErrForbidden if ownership check fails (IDOR prevention).
func ValidateEventOwnership(ctx context.Context, queries *db.Queries, eventID, organizerID pgtype.UUID) (*db.Events, error) {
	event, err := queries.GetEventByIDAndOrganizer(ctx, db.GetEventByIDAndOrganizerParams{
		ID:          eventID,
		OrganizerID: organizerID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("event not found or access forbidden")
		}
		return nil, err
	}
	return &event, nil
}
