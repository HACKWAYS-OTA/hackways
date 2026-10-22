// Session activity is best-effort telemetry, separate from authorization.
package auth

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type activityRecorder struct {
	queue   chan pgtype.UUID
	dropped atomic.Uint64
}

var sessionActivity atomic.Pointer[activityRecorder]

func (a *activityRecorder) Mark(id pgtype.UUID) {
	select {
	case a.queue <- id:
	default:
		a.dropped.Add(1)
	}
}

// StartSessionActivity runs one bounded worker per engine, never per request.
// Losing an activity sample cannot log anyone out or revive a deleted session.
func StartSessionActivity(ctx context.Context, q *db.Queries, logger *slog.Logger) <-chan struct{} {
	recorder := &activityRecorder{queue: make(chan pgtype.UUID, 256)}
	sessionActivity.Store(recorder)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer sessionActivity.CompareAndSwap(recorder, nil)
		timer := time.NewTicker(time.Second)
		defer timer.Stop()
		pending := make(map[pgtype.UUID]struct{}, 64)
		var lastReport time.Time
		flush := func() {
			if len(pending) == 0 {
				return
			}
			ids := make([]pgtype.UUID, 0, len(pending))
			for id := range pending {
				ids = append(ids, id)
			}
			writeCtx, cancel := context.WithTimeout(ctx, time.Second)
			err := q.UpdateSessionActivityBatch(writeCtx, ids)
			cancel()
			clear(pending)
			if err != nil && ctx.Err() == nil && time.Since(lastReport) > time.Minute {
				logger.Warn("session activity heartbeat could not be recorded; authorization is unaffected")
				lastReport = time.Now()
			}
		}
		for {
			select {
			case <-ctx.Done():
				return
			case id := <-recorder.queue:
				pending[id] = struct{}{}
				if len(pending) >= 64 {
					flush()
				}
			case <-timer.C:
				flush()
				if dropped := recorder.dropped.Swap(0); dropped > 0 {
					logger.Warn("bounded session activity queue dropped telemetry samples", "count", dropped)
				}
			}
		}
	}()
	return done
}
