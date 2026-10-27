package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type workspaceKey struct{}
type WorkspaceAccess struct {
	ResourceID  pgtype.UUID `json:"resource_id"`
	Kind        string      `json:"kind"`
	Role        string      `json:"role"`
	Permissions []string    `json:"permissions"`
	permission  string
}

var ErrWorkspaceAccessDenied = errors.New("workspace access was revoked or changed")

// RecheckWorkspace is called under the resource's write lock. Grant-row share
// locks serialize concurrent revocations until this mutation commits.
func RecheckWorkspace(ctx context.Context, q *db.Queries, kind string, id pgtype.UUID) error {
	access, scoped := WorkspaceFromContext(ctx)
	if !scoped {
		return nil
	}
	user, ok := UserFromContext(ctx)
	if !ok || access.Kind != kind || access.ResourceID != id {
		return ErrWorkspaceAccessDenied
	}
	if kind == "event" {
		space, err := q.EventSpaceForAccess(ctx, id)
		if err != nil {
			return err
		}
		if space.Valid {
			if _, err = q.LockSpaceAccessForActor(ctx, db.LockSpaceAccessForActorParams{SpaceID: space, UserID: user.UserID}); err != nil {
				return err
			}
		}
		if _, err = q.LockEventAccessForActor(ctx, db.LockEventAccessForActorParams{EventID: id, UserID: user.UserID}); err != nil {
			return err
		}
		row, err := q.EventWorkspaceAccess(ctx, db.EventWorkspaceAccessParams{ActorID: user.UserID, ResourceID: id})
		if errors.Is(err, pgx.ErrNoRows) || err == nil && !RoleAllows(row.AccessRole, access.permission) {
			return ErrWorkspaceAccessDenied
		}
		return err
	}
	if _, err := q.LockSpaceAccessForActor(ctx, db.LockSpaceAccessForActorParams{SpaceID: id, UserID: user.UserID}); err != nil {
		return err
	}
	row, err := q.SpaceWorkspaceAccess(ctx, db.SpaceWorkspaceAccessParams{ActorID: user.UserID, ResourceID: id})
	if errors.Is(err, pgx.ErrNoRows) || err == nil && !RoleAllows(row.AccessRole, access.permission) {
		return ErrWorkspaceAccessDenied
	}
	return err
}

var workspacePermissions = []string{"view", "content", "lifecycle", "guests.read", "guests.manage", "guests.export", "checkin", "create", "access.manage", "delete", "association"}

func RoleAllows(role, permission string) bool {
	switch role {
	case "owner":
		for _, value := range workspacePermissions {
			if value == permission {
				return true
			}
		}
	case "manager":
		return permission == "view" || permission == "content" || permission == "lifecycle" || permission == "guests.read" || permission == "guests.manage" || permission == "guests.export" || permission == "checkin" || permission == "create"
	case "editor":
		return permission == "view" || permission == "content"
	case "checkin":
		return permission == "view" || permission == "guests.read" || permission == "checkin"
	}
	return false
}
func RolePermissions(role string) []string {
	result := []string{}
	for _, value := range workspacePermissions {
		if RoleAllows(role, value) {
			result = append(result, value)
		}
	}
	return result
}
func WorkspaceFromContext(ctx context.Context) (WorkspaceAccess, bool) {
	access, ok := ctx.Value(workspaceKey{}).(WorkspaceAccess)
	return access, ok
}
func RequireEventPermission(q *db.Queries, permission string) func(http.Handler) http.Handler {
	return requireWorkspacePermission(q, "event", permission)
}
func RequireSpacePermission(q *db.Queries, permission string) func(http.Handler) http.Handler {
	return requireWorkspacePermission(q, "space", permission)
}
func requireWorkspacePermission(q *db.Queries, kind, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "private, no-store")
			user, ok := UserFromContext(r.Context())
			if !ok {
				platform.RespondError(w, 401, "UNAUTHORIZED", "Sign in to continue", nil)
				return
			}
			var id pgtype.UUID
			slug := chi.URLParam(r, "slug")
			raw := chi.URLParam(r, "id")
			if raw != "" && (id.Scan(raw) != nil || !id.Valid) {
				platform.RespondError(w, 400, "INVALID_ID", "Invalid workspace resource", nil)
				return
			}
			var owner db.Organizers
			var role string
			var resourceID pgtype.UUID
			var err error
			if kind == "event" {
				var row db.EventWorkspaceAccessRow
				row, err = q.EventWorkspaceAccess(r.Context(), db.EventWorkspaceAccessParams{ActorID: user.UserID, ResourceID: id, Slug: slug})
				owner, role, resourceID = row.Organizers, row.AccessRole, row.ResourceID
			} else {
				var row db.SpaceWorkspaceAccessRow
				row, err = q.SpaceWorkspaceAccess(r.Context(), db.SpaceWorkspaceAccessParams{ActorID: user.UserID, ResourceID: id, Slug: slug})
				owner, role, resourceID = row.Organizers, row.AccessRole, row.ResourceID
			}
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					platform.RespondError(w, 404, "NOT_FOUND", "This resource is not available to your account", nil)
				} else {
					platform.RespondDatabaseError(w, err)
				}
				return
			}
			if !RoleAllows(role, permission) {
				platform.RespondError(w, 403, "WORKSPACE_ACCESS_DENIED", "Your workspace role does not allow this action", nil)
				return
			}
			// Delegate only this checked resource's organizer scope. The authenticated user
			// remains the real actor, and the context marker is private to this package.
			access := WorkspaceAccess{ResourceID: resourceID, Kind: kind, Role: role, Permissions: RolePermissions(role), permission: permission}
			ctx := context.WithValue(context.WithValue(r.Context(), organizerContextKey, &owner), workspaceKey{}, access)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
