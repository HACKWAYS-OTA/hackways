package platform

import "net/http"

// LiveAnalyticsInterval validates sliding windows; historical dates retain their separate report.
func LiveAnalyticsInterval(w http.ResponseWriter, r *http.Request) (string, float64, int32, bool) {
	key := r.URL.Query().Get("window")
	if key == "" {
		key = "5m"
	}
	switch key {
	case "5m":
		return key, 300, 15, true
	case "1h":
		return key, 3600, 180, true
	case "24h":
		return key, 86400, 3600, true
	}
	RespondError(w, 400, "INVALID_INTERVAL", "Choose 5 minutes, 1 hour or 24 hours", nil)
	return "", 0, 0, false
}
