package events

import (
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"net/http"
)

func (h *Handler) ChangeAppearance(w http.ResponseWriter, r *http.Request) {
	id, ok := eventID(w, r)
	if !ok {
		return
	}
	var body struct {
		Theme        string `json:"theme"`
		ThemeColor   string `json:"theme_color"`
		ThemeFont    string `json:"theme_font"`
		ThemeStyle   string `json:"theme_style"`
		ThemeDisplay string `json:"theme_display"`
		RemoveBanner bool   `json:"remove_banner"`
		Revision     int64  `json:"revision"`
	}
	if platform.DecodeJSON(r, &body) != nil || body.Revision < 1 {
		platform.RespondError(w, 422, "INVALID_APPEARANCE", "Choose a theme and provide the current event revision", nil)
		return
	}
	appearance := createEventRequest{Theme: body.Theme, ThemeColor: body.ThemeColor, ThemeFont: body.ThemeFont, ThemeStyle: body.ThemeStyle, ThemeDisplay: body.ThemeDisplay}
	if !validAppearance(&appearance) {
		platform.RespondError(w, 422, "INVALID_APPEARANCE", "Choose a valid event theme, colour, font and display mode", nil)
		return
	}
	tx, q, event, err := workflows.New(h.pool).Begin(r.Context(), id)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	owner, _ := auth.OrganizerFromContext(r.Context())
	if event.OrganizerID != owner.ID {
		lifecycleError(w, workflows.NotFound())
		return
	}
	if event.Revision != body.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The event changed. Reload before changing its appearance", nil)
		return
	}
	if event.Status == db.EventStatusCancelled || !event.StartsAt.Time.After(event.DatabaseNow.Time) {
		platform.RespondError(w, 409, "EDIT_CLOSED", "Started or cancelled events cannot change appearance", nil)
		return
	}
	updated, err := q.SetEventAppearance(r.Context(), db.SetEventAppearanceParams{ID: id, Theme: appearance.Theme, ThemeColor: appearance.ThemeColor, ThemeFont: appearance.ThemeFont, ThemeStyle: appearance.ThemeStyle, ThemeDisplay: appearance.ThemeDisplay, RemoveBanner: body.RemoveBanner})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if err = workflows.Audit(r.Context(), q, owner.UserID, id, id, "event.appearance_updated", platform.GetRequestID(r.Context())); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		lifecycleError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Events `json:"data"`
	}{updated})
}
