package admin

import (
	"context"
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
)

type coreAnalytics struct {
	Period      platform.AnalyticsWindow           `json:"period"`
	Summary     db.AnalyticsRegistrationSummaryRow `json:"summary"`
	Events      db.AnalyticsEventSummaryRow        `json:"events"`
	Daily       []db.AnalyticsRegistrationDailyRow `json:"daily"`
	TopEvents   []db.AnalyticsTopEventsRow         `json:"top_events"`
	GeneratedAt time.Time                          `json:"generated_at"`
}

func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	period, ok := platform.ReadAnalyticsWindow(w, r)
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
	data := coreAnalytics{Period: period, GeneratedAt: time.Now().UTC()}
	data.Summary, err = q.AnalyticsRegistrationSummary(ctx, db.AnalyticsRegistrationSummaryParams{PeriodStart: period.Start, PeriodEnd: period.End})
	if err == nil {
		data.Events, err = q.AnalyticsEventSummary(ctx, db.AnalyticsEventSummaryParams{PeriodStart: period.Start, PeriodEnd: period.End})
	}
	if err == nil {
		data.Daily, err = q.AnalyticsRegistrationDaily(ctx, db.AnalyticsRegistrationDailyParams{PeriodStart: period.Start, PeriodEnd: period.End})
	}
	if err == nil {
		data.TopEvents, err = q.AnalyticsTopEvents(ctx, db.AnalyticsTopEventsParams{PeriodStart: period.Start, PeriodEnd: period.End})
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	if r.URL.Query().Get("format") == "csv" {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="hackways-event-analytics.csv"`)
		writer := csv.NewWriter(w)
		defer writer.Flush()
		if writer.Write([]string{"Day (UTC)", "Registrations recorded", "Currently confirmed in this cohort", "Check-ins"}) != nil {
			return
		}
		for _, row := range data.Daily {
			if writer.Write([]string{row.Day, strconv.FormatInt(row.Registrations, 10), strconv.FormatInt(row.Confirmed, 10), strconv.FormatInt(row.CheckIns, 10)}) != nil {
				return
			}
		}
		return
	}
	if data.Daily == nil {
		data.Daily = []db.AnalyticsRegistrationDailyRow{}
	}
	if data.TopEvents == nil {
		data.TopEvents = []db.AnalyticsTopEventsRow{}
	}
	platform.RespondJSON(w, 200, data)
}
