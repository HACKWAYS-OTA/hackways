package auth

import (
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
)

// UpdateProfile changes only the authenticated account's display name.
// Email, verification, permissions and organizer identity are separate records.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok || user == nil {
		platform.RespondError(w, 401, "UNAUTHORIZED", "Sign in to edit your profile", nil)
		return
	}
	var body struct {
		DisplayName string `json:"display_name"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Provide your display name", nil)
		return
	}
	name := strings.TrimSpace(body.DisplayName)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 100 || strings.ContainsFunc(name, unicode.IsControl) {
		platform.RespondError(w, 422, "INVALID_NAME", "Use a name between 1 and 100 characters without control characters", nil)
		return
	}
	updated, err := h.queries.UpdateOwnDisplayName(r.Context(), db.UpdateOwnDisplayNameParams{ID: user.UserID, DisplayName: name})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	var avatar *string
	if updated.AvatarUrl.Valid {
		avatar = &updated.AvatarUrl.String
	}
	w.Header().Set("Cache-Control", "private, no-store")
	platform.RespondJSON(w, 200, struct {
		User UserResponse `json:"user"`
	}{UserResponse{ID: formatUUID(updated.ID), Email: updated.Email, DisplayName: updated.DisplayName, AvatarURL: avatar, EmailVerified: updated.EmailVerified}})
}
