package admin

import (
	"context"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"net/http"
	"time"
)

type liveGlobalAnalytics struct {
	Window     string                        `json:"window"`
	CapturedAt time.Time                     `json:"captured_at"`
	Traffic    db.AnalyticsTrafficSummaryRow `json:"traffic"`
	Accounts   db.LiveAccountSummaryRow      `json:"accounts"`
	Pages      []db.AnalyticsTopPagesRow     `json:"pages"`
	Sources    []db.AnalyticsAcquisitionRow  `json:"sources"`
	Limited    bool                          `json:"limited"`
}

func (h *Handler) LiveAnalytics(w http.ResponseWriter, r *http.Request) {
	interval, seconds, _, ok := platform.LiveAnalyticsInterval(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	defer tx.Rollback(ctx)
	q := h.queries.WithTx(tx)
	if err = q.SetStorageWriteTimeout(ctx); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	bounds, err := q.LiveAnalyticsWindow(ctx, seconds)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	result := liveGlobalAnalytics{Window: interval, CapturedAt: bounds.PeriodEnd.Time}
	result.Traffic, err = q.AnalyticsTrafficSummary(ctx, db.AnalyticsTrafficSummaryParams{PeriodStart: bounds.PeriodStart, PeriodEnd: bounds.PeriodEnd})
	if err == nil {
		result.Accounts, err = q.LiveAccountSummary(ctx, db.LiveAccountSummaryParams{PeriodStart: bounds.PeriodStart, PeriodEnd: bounds.PeriodEnd})
	}
	if err == nil {
		result.Pages, err = q.AnalyticsTopPages(ctx, db.AnalyticsTopPagesParams{PeriodStart: bounds.PeriodStart, PeriodEnd: bounds.PeriodEnd})
	}
	if err == nil {
		result.Sources, err = q.AnalyticsAcquisition(ctx, db.AnalyticsAcquisitionParams{PeriodStart: bounds.PeriodStart, PeriodEnd: bounds.PeriodEnd})
	}
	if err == nil {
		result.Limited, err = q.LiveAnalyticsLimited(ctx, db.LiveAnalyticsLimitedParams{PeriodStart: bounds.PeriodStart, PeriodEnd: bounds.PeriodEnd})
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if result.Pages == nil {
		result.Pages = []db.AnalyticsTopPagesRow{}
	}
	if result.Sources == nil {
		result.Sources = []db.AnalyticsAcquisitionRow{}
	}
	w.Header().Set("Cache-Control", "private, no-store")
	platform.RespondJSON(w, 200, result)
}
