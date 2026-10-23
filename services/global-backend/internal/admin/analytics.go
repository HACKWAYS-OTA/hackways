package admin

import (
	"context"
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
)

type globalAnalytics struct {
	Period       platform.AnalyticsWindow       `json:"period"`
	Traffic      db.AnalyticsTrafficSummaryRow  `json:"traffic"`
	TrafficDaily []db.AnalyticsTrafficDailyRow  `json:"traffic_daily"`
	Pages        []db.AnalyticsTopPagesRow      `json:"pages"`
	Sources      []db.AnalyticsAcquisitionRow   `json:"sources"`
	Devices      []db.AnalyticsDevicesRow       `json:"devices"`
	Accounts     db.AnalyticsAccountSummaryRow  `json:"accounts"`
	AccountDaily []db.AnalyticsAccountDailyRow  `json:"account_daily"`
	Collection   db.AnalyticsCollectionStateRow `json:"collection"`
	GeneratedAt  time.Time                      `json:"generated_at"`
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
	if err = q.SetStorageWriteTimeout(ctx); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	data := globalAnalytics{Period: period, GeneratedAt: time.Now().UTC()}
	data.Traffic, err = q.AnalyticsTrafficSummary(ctx, db.AnalyticsTrafficSummaryParams{PeriodStart: period.Start, PeriodEnd: period.End})
	if err == nil {
		data.TrafficDaily, err = q.AnalyticsTrafficDaily(ctx, db.AnalyticsTrafficDailyParams{PeriodStart: period.Start, PeriodEnd: period.End})
	}
	if err == nil {
		data.Pages, err = q.AnalyticsTopPages(ctx, db.AnalyticsTopPagesParams{PeriodStart: period.Start, PeriodEnd: period.End})
	}
	if err == nil {
		data.Sources, err = q.AnalyticsAcquisition(ctx, db.AnalyticsAcquisitionParams{PeriodStart: period.Start, PeriodEnd: period.End})
	}
	if err == nil {
		data.Devices, err = q.AnalyticsDevices(ctx, db.AnalyticsDevicesParams{PeriodStart: period.Start, PeriodEnd: period.End})
	}
	if err == nil {
		data.Accounts, err = q.AnalyticsAccountSummary(ctx, db.AnalyticsAccountSummaryParams{PeriodStart: period.Start, PeriodEnd: period.End})
	}
	if err == nil {
		data.AccountDaily, err = q.AnalyticsAccountDaily(ctx, db.AnalyticsAccountDailyParams{PeriodStart: period.Start, PeriodEnd: period.End})
	}
	if err == nil {
		data.Collection, err = q.AnalyticsCollectionState(ctx)
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
		writeGlobalAnalyticsCSV(w, data)
		return
	}
	if data.TrafficDaily == nil {
		data.TrafficDaily = []db.AnalyticsTrafficDailyRow{}
	}
	if data.Pages == nil {
		data.Pages = []db.AnalyticsTopPagesRow{}
	}
	if data.Sources == nil {
		data.Sources = []db.AnalyticsAcquisitionRow{}
	}
	if data.Devices == nil {
		data.Devices = []db.AnalyticsDevicesRow{}
	}
	if data.AccountDaily == nil {
		data.AccountDaily = []db.AnalyticsAccountDailyRow{}
	}
	platform.RespondJSON(w, 200, data)
}

func writeGlobalAnalyticsCSV(w http.ResponseWriter, data globalAnalytics) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="hackways-accounts-traffic.csv"`)
	writer := csv.NewWriter(w)
	defer writer.Flush()
	if writer.Write([]string{"Day (UTC)", "New accounts", "Provider-accepted emails", "Consented page views", "Consented visitors", "Tracked registration completions"}) != nil {
		return
	}
	traffic := make(map[string]db.AnalyticsTrafficDailyRow, len(data.TrafficDaily))
	for _, row := range data.TrafficDaily {
		traffic[row.Day] = row
	}
	for _, row := range data.AccountDaily {
		visit := traffic[row.Day]
		if writer.Write([]string{row.Day, strconv.FormatInt(row.NewAccounts, 10), strconv.FormatInt(row.AcceptedEmails, 10), strconv.FormatInt(visit.PageViews, 10), strconv.FormatInt(visit.Visitors, 10), strconv.FormatInt(visit.TrackedRegistrations, 10)}) != nil {
			return
		}
	}
}
