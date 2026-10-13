package platform

import (
	"context"
	"net/http"
	"time"
)

// Pinger defines an interface for components capable of health pinging (e.g. pgxpool.Pool).
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthResponse represents service health and runtime diagnostic metadata.
type HealthResponse struct {
	Status      string `json:"status"`
	Database    string `json:"database,omitempty"`
	Version     string `json:"version"`
	Timestamp   string `json:"timestamp"`
	Uptime      string `json:"uptime"`
	Environment string `json:"environment"`
}

var serverStartTime = time.Now()

// HealthHandler returns an HTTP handler serving system health status.
func HealthHandler(env string, db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is supported", nil)
			return
		}

		dbStatus := "healthy"
		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := db.Ping(ctx); err != nil {
				dbStatus = "unhealthy"
				res := HealthResponse{
					Status:      "degraded",
					Database:    dbStatus,
					Version:     "1.0.0",
					Timestamp:   time.Now().UTC().Format(time.RFC3339),
					Uptime:      time.Since(serverStartTime).Round(time.Second).String(),
					Environment: env,
				}
				RespondJSON(w, http.StatusServiceUnavailable, res)
				return
			}
		}

		res := HealthResponse{
			Status:      "ok",
			Database:    dbStatus,
			Version:     "1.0.0",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Uptime:      time.Since(serverStartTime).Round(time.Second).String(),
			Environment: env,
		}

		RespondJSON(w, http.StatusOK, res)
	}
}
