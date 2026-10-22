package events

import (
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"time"
)

// DiscoverPromotions exposes only currently eligible public events. Future
// selections and administrative configuration never leave the private API.
func (h *Handler) DiscoverPromotions(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	at := pgtype.Timestamptz{Time: now, Valid: true}
	tx, err := h.pool.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	rows, err := q.ActiveDiscoverPromotions(r.Context(), at)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	next, err := q.NextDiscoverPromotionChange(r.Context(), at)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.ActiveDiscoverPromotionsRow{}
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	// Optional five-row feed has scheduled boundaries; no personalized cache is
	// involved. Clients rotate locally and revalidate at boundaries/on focus.
	w.Header().Set("Cache-Control", "no-store")
	platform.RespondJSON(w, 200, struct {
		Data         []db.ActiveDiscoverPromotionsRow `json:"data"`
		NextChangeAt pgtype.Timestamptz               `json:"next_change_at"`
		ServerTime   time.Time                        `json:"server_time"`
	}{rows, next, now})
}
