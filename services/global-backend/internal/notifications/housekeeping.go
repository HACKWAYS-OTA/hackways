package notifications

import (
	"context"
	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

// Housekeeping remains active when external notification delivery is disabled.
func Housekeeping(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) {
	q := db.New(pool)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			job, cancel := context.WithTimeout(ctx, 2*time.Second)
			if err := q.CleanupAuthAttempts(job); err != nil {
				logger.Warn("authentication counter cleanup failed", "error", err)
			}
			if err := q.CleanupExpiredRecognizedDevices(job); err != nil {
				logger.Warn("expired device cleanup failed", "error", err)
			}
			if err := q.CleanupAnalyticsEvents(job); err != nil {
				logger.Warn("analytics retention cleanup failed", "error", err)
			}
			if err := q.CleanupAnalyticsBudget(job); err != nil {
				logger.Warn("analytics quota cleanup failed", "error", err)
			}
			cancel()
		}
	}
}
