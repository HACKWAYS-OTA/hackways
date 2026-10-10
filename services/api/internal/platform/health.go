package platform

import (
	"net/http"
	"time"
)

// HealthResponse represents service health and runtime diagnostic metadata.
type HealthResponse struct {
	Status      string `json:"status"`
	Version     string `json:"version"`
	Timestamp   string `json:"timestamp"`
	Uptime      string `json:"uptime"`
	Environment string `json:"environment"`
}

var serverStartTime = time.Now()

// HealthHandler returns an HTTP handler serving system health status.
func HealthHandler(env string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is supported", nil)
			return
		}

		res := HealthResponse{
			Status:      "ok",
			Version:     "1.0.0",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Uptime:      time.Since(serverStartTime).Round(time.Second).String(),
			Environment: env,
		}

		RespondJSON(w, http.StatusOK, res)
	}
}
