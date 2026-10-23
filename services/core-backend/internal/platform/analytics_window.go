package platform

import (
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"time"
)

type AnalyticsWindow struct {
	From  string             `json:"from"`
	To    string             `json:"to"`
	Start pgtype.Timestamptz `json:"-"`
	End   pgtype.Timestamptz `json:"-"`
}

func ReadAnalyticsWindow(w http.ResponseWriter, r *http.Request) (AnalyticsWindow, bool) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	last := today
	var err error
	if raw := r.URL.Query().Get("to"); raw != "" {
		last, err = time.Parse("2006-01-02", raw)
		if err != nil {
			RespondError(w, 400, "INVALID_PERIOD", "Choose valid dates", nil)
			return AnalyticsWindow{}, false
		}
	}
	first := last.AddDate(0, 0, -29)
	if raw := r.URL.Query().Get("from"); raw != "" {
		first, err = time.Parse("2006-01-02", raw)
		if err != nil {
			RespondError(w, 400, "INVALID_PERIOD", "Choose valid dates", nil)
			return AnalyticsWindow{}, false
		}
	}
	if first.After(last) || last.After(today) || first.Before(today.AddDate(0, 0, -89)) || last.Sub(first) > 89*24*time.Hour {
		RespondError(w, 400, "INVALID_PERIOD", "Choose up to ninety days within the latest ninety days", nil)
		return AnalyticsWindow{}, false
	}
	return AnalyticsWindow{first.Format("2006-01-02"), last.Format("2006-01-02"), pgtype.Timestamptz{Time: first, Valid: true}, pgtype.Timestamptz{Time: last.AddDate(0, 0, 1), Valid: true}}, true
}
