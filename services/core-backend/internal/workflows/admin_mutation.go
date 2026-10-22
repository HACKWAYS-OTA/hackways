package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
)

// ReplayAdminMutation is called only after locking the affected event or Space.
// Owner workflows remain unchanged when the private administrator marker is absent.
func ReplayAdminMutation(w http.ResponseWriter, r *http.Request, q *db.Queries) (bool, error) {
	m, ok := auth.AdminMutationFromContext(r.Context())
	if !ok {
		return false, nil
	}
	actor, _ := auth.UserFromContext(r.Context())
	allowed, err := q.IsPlatformAdminSession(r.Context(), db.IsPlatformAdminSessionParams{SessionID: actor.SessionID, UserID: actor.UserID})
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, &Failure{403, "ADMIN_ACCESS_REQUIRED", "Admin access required"}
	}
	row, err := q.GetIdempotentResponse(r.Context(), db.GetIdempotentResponseParams{ScopeHash: m.Scope, KeyHash: m.Key})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if row.RequestHash != m.RequestHash {
		return false, fail("IDEMPOTENCY_CONFLICT", "This request key was used with different changes")
	}
	w.Header().Set("Idempotency-Replayed", "true")
	platform.RespondJSON(w, 200, json.RawMessage(row.ResponseJson))
	return true, nil
}

func RememberAdminMutation[T any](ctx context.Context, q *db.Queries, data T) error {
	m, ok := auth.AdminMutationFromContext(ctx)
	if !ok {
		return nil
	}
	response, err := json.Marshal(struct {
		Data T `json:"data"`
	}{data})
	if err != nil {
		return err
	}
	return q.StoreIdempotentResponse(ctx, db.StoreIdempotentResponseParams{ScopeHash: m.Scope, KeyHash: m.Key, RequestHash: m.RequestHash, ResponseJson: response})
}
