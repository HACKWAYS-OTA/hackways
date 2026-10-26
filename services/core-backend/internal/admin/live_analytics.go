package admin

import (
	"context"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"net/http"
	"time"
)

type liveCoreAnalytics struct {
	Window     string                             `json:"window"`
	CapturedAt time.Time                          `json:"captured_at"`
	Summary    db.AnalyticsRegistrationSummaryRow `json:"summary"`
	Events     db.AnalyticsEventSummaryRow        `json:"events"`
	Activity   []db.LiveRegistrationActivityRow   `json:"activity"`
	TopEvents  []db.AnalyticsTopEventsRow         `json:"top_events"`
}

func (h *Handler) LiveAnalytics(w http.ResponseWriter, r *http.Request) {
	interval, seconds, bucket, ok := platform.LiveAnalyticsInterval(w, r)
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
	if err = q.SetWorkflowTimeout(ctx); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	bounds, err := q.LiveAnalyticsWindow(ctx, seconds)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	result := liveCoreAnalytics{Window: interval, CapturedAt: bounds.PeriodEnd.Time}
	result.Summary, err = q.AnalyticsRegistrationSummary(ctx, db.AnalyticsRegistrationSummaryParams{PeriodStart: bounds.PeriodStart, PeriodEnd: bounds.PeriodEnd})
	if err == nil {
		result.Events, err = q.AnalyticsEventSummary(ctx, db.AnalyticsEventSummaryParams{PeriodStart: bounds.PeriodStart, PeriodEnd: bounds.PeriodEnd})
	}
	if err == nil {
		result.Activity, err = q.LiveRegistrationActivity(ctx, db.LiveRegistrationActivityParams{PeriodStart: bounds.PeriodStart, PeriodEnd: bounds.PeriodEnd, BucketSeconds: bucket})
	}
	if err == nil {
		result.TopEvents, err = q.AnalyticsTopEvents(ctx, db.AnalyticsTopEventsParams{PeriodStart: bounds.PeriodStart, PeriodEnd: bounds.PeriodEnd})
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if result.Activity == nil {
		result.Activity = []db.LiveRegistrationActivityRow{}
	}
	if result.TopEvents == nil {
		result.TopEvents = []db.AnalyticsTopEventsRow{}
	}
	w.Header().Set("Cache-Control", "private, no-store")
	platform.RespondJSON(w, 200, result)
}
