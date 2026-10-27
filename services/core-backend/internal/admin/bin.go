package admin

import (
	"net/http"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// EventBin retains registrations, tickets, grants, curation, media and history.
// There is intentionally no purge operation.
func (h *Handler) EventBin(w http.ResponseWriter, r *http.Request) {
	query, _, limit, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.AdminListEventBin(r.Context(), db.AdminListEventBinParams{Q: query, PageLimit: limit, PageOffset: offset})
	if err != nil {
		detailError(w, err)
		return
	}
	if rows == nil {
		rows = []db.AdminListEventBinRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.AdminListEventBinRow `json:"data"`
	}{rows})
}

func (h *Handler) BinEvent(w http.ResponseWriter, r *http.Request)     { h.changeEventBin(w, r, false) }
func (h *Handler) RestoreEvent(w http.ResponseWriter, r *http.Request) { h.changeEventBin(w, r, true) }

func (h *Handler) changeEventBin(w http.ResponseWriter, r *http.Request, restore bool) {
	id, ok := h.detailID(w, r, "events")
	if !ok {
		return
	}
	var body struct {
		Revision int64 `json:"revision"`
	}
	if platform.DecodeJSON(r, &body) != nil || body.Revision < 1 {
		platform.RespondError(w, 422, "REVISION_REQUIRED", "Provide the current event revision", nil)
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		mutationError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	if err = q.SetWorkflowTimeout(r.Context()); err != nil {
		mutationError(w, err)
		return
	}
	event, err := q.LockAdminBinEvent(r.Context(), id)
	if err != nil {
		mutationError(w, err)
		return
	}
	// Replay under the same resource lock, including after soft deletion. A lost
	// response must never repeat a lifecycle action or bypass fresh admin checks.
	if replayed, e := workflows.ReplayAdminMutation(w, r, q); e != nil {
		mutationError(w, e)
		return
	} else if replayed {
		return
	}
	if event.Revision != body.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The event changed. Reload before continuing", nil)
		return
	}
	if restore != event.DeletedAt.Valid {
		platform.RespondError(w, 409, "BIN_STATE_CHANGED", "The event is already in a different state. Reload before continuing", nil)
		return
	}
	if !restore && event.Status != db.EventStatusCancelled && (!event.EndsAt.Valid || event.EndsAt.Time.After(event.DatabaseNow.Time)) {
		activity, e := q.EventWorkflowActivity(r.Context(), id)
		if e != nil {
			mutationError(w, e)
			return
		}
		if event.Status == db.EventStatusPublished || activity.RegistrationActive {
			platform.RespondError(w, 409, "CANCEL_BEFORE_BIN", "Cancel this active event before moving it to Bin so guests and reserved places are handled correctly", nil)
			return
		}
	}
	// Restoration never publishes an event or reactivates a cancelled booking.
	if restore {
		updated, e := q.RestoreAdminBinEvent(r.Context(), id)
		if e != nil {
			mutationError(w, e)
			return
		}
		err = commitBinChange(r, q, tx, updated, id, "event.restored")
		if err == nil {
			platform.RespondJSON(w, 200, struct {
				Data db.RestoreAdminBinEventRow `json:"data"`
			}{updated})
		}
	} else {
		updated, e := q.MoveAdminEventToBin(r.Context(), id)
		if e != nil {
			mutationError(w, e)
			return
		}
		err = commitBinChange(r, q, tx, updated, id, "event.deleted")
		if err == nil {
			platform.RespondJSON(w, 200, struct {
				Data db.MoveAdminEventToBinRow `json:"data"`
			}{updated})
		}
	}
	if err != nil {
		mutationError(w, err)
	}
}

func commitBinChange[T any](r *http.Request, q *db.Queries, tx pgx.Tx, updated T, id pgtype.UUID, action string) error {
	actor, _ := auth.UserFromContext(r.Context())
	if err := workflows.Audit(r.Context(), q, actor.UserID, id, id, action, platform.GetRequestID(r.Context())); err != nil {
		return err
	}
	if err := workflows.RememberAdminMutation(r.Context(), q, updated); err != nil {
		return err
	}
	return tx.Commit(r.Context())
}
