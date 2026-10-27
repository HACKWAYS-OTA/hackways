package events

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
)

func (h *Handler) ManageBySlug(w http.ResponseWriter, r *http.Request) {
	owner, _ := auth.OrganizerFromContext(r.Context())
	event, e := h.queries.GetEventBySlugAndOrganizer(r.Context(), db.GetEventBySlugAndOrganizerParams{Slug: chi.URLParam(r, "slug"), OrganizerID: owner.ID})
	if e != nil {
		lifecycleError(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data   db.Events            `json:"data"`
		Access auth.WorkspaceAccess `json:"access"`
	}{event, workspaceAccess(r)})
}
func (h *Handler) ChangeSpace(w http.ResponseWriter, r *http.Request) {
	id, ok := eventID(w, r)
	if !ok {
		return
	}
	var body struct {
		SpaceSlug string `json:"space_slug"`
		Revision  int64  `json:"revision"`
	}
	if platform.DecodeJSON(r, &body) != nil || body.Revision < 1 {
		platform.RespondError(w, 422, "INVALID_PAYLOAD", "Choose a Space and provide the event revision", nil)
		return
	}
	owner, _ := auth.OrganizerFromContext(r.Context())
	tx, q, event, e := workflows.New(h.pool).Begin(r.Context(), id)
	if e != nil {
		lifecycleError(w, e)
		return
	}
	defer tx.Rollback(r.Context())
	if event.OrganizerID != owner.ID {
		lifecycleError(w, pgx.ErrNoRows)
		return
	}
	if event.Revision != body.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The event changed. Reload before changing its Space", nil)
		return
	}
	if event.ManagementClosed {
		platform.RespondError(w, 409, "EDIT_CLOSED", "An administrator has closed editing for this event", nil)
		return
	}
	var spaceID pgtype.UUID
	if body.SpaceSlug != "" {
		space, readErr := q.LockOwnedActiveSpaceBySlug(r.Context(), db.LockOwnedActiveSpaceBySlugParams{Slug: body.SpaceSlug, OrganizerID: owner.ID})
		if errors.Is(readErr, pgx.ErrNoRows) {
			platform.RespondError(w, 404, "SPACE_NOT_AVAILABLE", "Choose an active Space you own", nil)
			return
		}
		if readErr != nil {
			lifecycleError(w, readErr)
			return
		}
		spaceID = space.ID
	}
	updated, e := q.SetEventSpace(r.Context(), db.SetEventSpaceParams{ID: id, SpaceID: spaceID})
	if e != nil {
		lifecycleError(w, e)
		return
	}
	if e = workflows.Audit(r.Context(), q, owner.UserID, id, id, "event.space_changed", platform.GetRequestID(r.Context())); e != nil {
		lifecycleError(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		lifecycleError(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Events `json:"data"`
	}{updated})
}
