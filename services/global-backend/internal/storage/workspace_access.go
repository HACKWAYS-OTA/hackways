package storage

import (
	"context"
	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// Final attachment locks the resource and applicable grants before repeating
// authorization in SQL. A previously issued SAS never grants publication rights.
func lockEventMediaAccess(ctx context.Context, q *db.Queries, id, user pgtype.UUID) (db.LockEventMediaResourceRow, error) {
	state, err := q.LockEventMediaResource(ctx, id)
	if err != nil {
		return state, err
	}
	if state.SpaceID.Valid {
		if _, err = q.LockMediaSpaceGrant(ctx, db.LockMediaSpaceGrantParams{SpaceID: state.SpaceID, UserID: user}); err != nil {
			return state, err
		}
	}
	_, err = q.LockMediaEventGrant(ctx, db.LockMediaEventGrantParams{EventID: id, UserID: user})
	return state, err
}
func lockSpaceMediaAccess(ctx context.Context, q *db.Queries, id, user pgtype.UUID) error {
	if _, err := q.LockSpaceMediaResource(ctx, id); err != nil {
		return err
	}
	_, err := q.LockMediaSpaceGrant(ctx, db.LockMediaSpaceGrantParams{SpaceID: id, UserID: user})
	return err
}
