package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type promotionConfig struct {
	Revision   int64                           `json:"revision"`
	UpdatedAt  pgtype.Timestamptz              `json:"updated_at"`
	Selections []db.AdminDiscoverPromotionsRow `json:"selections"`
}

func readPromotionConfig(r *http.Request, q *db.Queries) (promotionConfig, error) {
	settings, err := q.DiscoverPromotionSettings(r.Context())
	if err != nil {
		return promotionConfig{}, err
	}
	rows, err := q.AdminDiscoverPromotions(r.Context())
	if rows == nil {
		rows = []db.AdminDiscoverPromotionsRow{}
	}
	return promotionConfig{settings.Revision, settings.UpdatedAt, rows}, err
}

func (h *Handler) GetDiscoverPromotions(w http.ResponseWriter, r *http.Request) {
	tx, err := h.pool.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		detailError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	result, err := readPromotionConfig(r, h.queries.WithTx(tx))
	if err != nil {
		detailError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		detailError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data promotionConfig `json:"data"`
	}{result})
}

func (h *Handler) DiscoverPromotionChoices(w http.ResponseWriter, r *http.Request) {
	query, _, _, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	rows, err := h.queries.PromotionEventChoices(r.Context(), db.PromotionEventChoicesParams{Q: query, PageOffset: offset})
	if err != nil {
		detailError(w, err)
		return
	}
	more := len(rows) > 20
	if more {
		rows = rows[:20]
	}
	if rows == nil {
		rows = []db.PromotionEventChoicesRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data    []db.PromotionEventChoicesRow `json:"data"`
		HasNext bool                          `json:"has_next"`
	}{rows, more})
}

type promotionSelection struct {
	Reference int64     `json:"reference"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
}

func (h *Handler) SaveDiscoverPromotions(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Revision   int64                `json:"revision"`
		Selections []promotionSelection `json:"selections"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil || body.Revision < 1 || len(body.Selections) > 5 {
		platform.RespondError(w, 422, "INVALID_PROMOTIONS", "Select up to five events and provide the current revision", nil)
		return
	}
	seen := map[int64]bool{}
	for _, selection := range body.Selections {
		if selection.Reference < 1 || seen[selection.Reference] || selection.StartsAt.IsZero() || !selection.EndsAt.After(selection.StartsAt) {
			platform.RespondError(w, 422, "INVALID_PROMOTIONS", "Use distinct events and an end after each start", nil)
			return
		}
		seen[selection.Reference] = true
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		detailError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	if err = q.SetWorkflowTimeout(r.Context()); err != nil {
		detailError(w, err)
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	// Serialize with access revocation, then lock the singleton before checking
	// the receipt/revision. Every save is atomic and the database caps slots at 5.
	if _, err = q.LockPromotionActor(r.Context(), actor.UserID); err != nil {
		detailError(w, err)
		return
	}
	settings, err := q.LockDiscoverPromotionSettings(r.Context())
	if err != nil {
		detailError(w, err)
		return
	}
	if replayed, e := workflows.ReplayAdminMutation(w, r, q); e != nil {
		mutationError(w, e)
		return
	} else if replayed {
		return
	}
	if settings.Revision != body.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "Promotions changed. Reload the saved selection before editing again.", nil)
		return
	}
	before, err := q.AdminDiscoverPromotions(r.Context())
	if err != nil {
		detailError(w, err)
		return
	}
	ordered := append([]promotionSelection(nil), body.Selections...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Reference < ordered[j].Reference })
	ids := make(map[int64]pgtype.UUID, len(ordered))
	for _, selection := range ordered {
		event, e := q.LockPromotionEvent(r.Context(), selection.Reference)
		if errors.Is(e, pgx.ErrNoRows) {
			platform.RespondError(w, 422, "EVENT_UNAVAILABLE", "A selected event is no longer available", nil)
			return
		}
		if e != nil {
			detailError(w, e)
			return
		}
		if event.Status != db.EventStatusPublished || event.Visibility != "public" || !event.StartsAt.Time.After(settings.DatabaseNow.Time) || !selection.EndsAt.After(settings.DatabaseNow.Time) || selection.EndsAt.After(event.StartsAt.Time) {
			platform.RespondError(w, 422, "EVENT_INELIGIBLE", "Choose published, public, upcoming events. End each promotion before the event starts.", nil)
			return
		}
		ids[selection.Reference] = event.ID
	}
	if err = q.ClearDiscoverPromotions(r.Context()); err != nil {
		detailError(w, err)
		return
	}
	for index, selection := range body.Selections {
		if err = q.InsertDiscoverPromotion(r.Context(), db.InsertDiscoverPromotionParams{Slot: int16(index + 1), EventID: ids[selection.Reference], StartsAt: pgtype.Timestamptz{Time: selection.StartsAt, Valid: true}, EndsAt: pgtype.Timestamptz{Time: selection.EndsAt, Valid: true}}); err != nil {
			detailError(w, err)
			return
		}
	}
	if err = q.UpdateDiscoverPromotionSettings(r.Context(), actor.UserID); err != nil {
		detailError(w, err)
		return
	}
	mutation, _ := auth.AdminMutationFromContext(r.Context())
	metadata, _ := json.Marshal(struct {
		Reason string                          `json:"reason"`
		Before []db.AdminDiscoverPromotionsRow `json:"before"`
		After  []promotionSelection            `json:"after"`
	}{mutation.Reason, before, body.Selections})
	if err = q.AppendWorkflowAudit(r.Context(), db.AppendWorkflowAuditParams{ActorUserID: actor.UserID, EntityID: settings.ID, Action: "admin.discover_promotions.updated", RequestID: platform.GetRequestID(r.Context()), Metadata: metadata}); err != nil {
		detailError(w, err)
		return
	}
	result, err := readPromotionConfig(r, q)
	if err != nil {
		detailError(w, err)
		return
	}
	if err = workflows.RememberAdminMutation(r.Context(), q, result); err != nil {
		detailError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		detailError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data promotionConfig `json:"data"`
	}{result})
}
