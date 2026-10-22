package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/hackways/global-backend/internal/rollouts"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type rolloutConfiguration struct {
	Revision  int64              `json:"revision"`
	Version   int64              `json:"version"`
	Label     string             `json:"label"`
	Flags     json.RawMessage    `json:"flags"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}
type rolloutHistory struct {
	Reference    int64              `json:"reference"`
	Label        string             `json:"label"`
	Flags        json.RawMessage    `json:"flags"`
	RestoredFrom pgtype.Int8        `json:"restored_from"`
	CreatedBy    string             `json:"created_by"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
}

func (h *Handler) GetRollouts(w http.ResponseWriter, r *http.Request) {
	_, _, _, offset, ok := pageParams(w, r)
	if !ok {
		return
	}
	tx, err := h.pool.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	current, err := q.ReadRuntimeRollout(r.Context())
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	rows, err := q.ListRuntimeRolloutVersions(r.Context(), offset)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	more := len(rows) > 20
	if more {
		rows = rows[:20]
	}
	history := make([]rolloutHistory, 0, len(rows))
	for _, row := range rows {
		history = append(history, rolloutHistory{row.Reference, row.Label, row.Flags, row.RestoredFrom, row.CreatedBy, row.CreatedAt})
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	source := os.Getenv("SOURCE_COMMIT")
	if source == "" {
		source = "unreported"
	}
	cacheStatus := rollouts.Status{}
	if h.rollouts != nil {
		cacheStatus = h.rollouts.Status()
	}
	platform.RespondJSON(w, 200, struct {
		Data         rolloutConfiguration  `json:"data"`
		Definitions  []rollouts.Definition `json:"definitions"`
		History      []rolloutHistory      `json:"history"`
		HasNext      bool                  `json:"has_next"`
		SourceCommit string                `json:"source_commit"`
		Cache        rollouts.Status       `json:"cache"`
	}{rolloutConfiguration{current.Revision, current.Reference, current.Label, current.Flags, current.UpdatedAt}, rollouts.Definitions(), history, more, source, cacheStatus})
}

func (h *Handler) SaveRollouts(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Revision       int64          `json:"revision"`
		Label          string         `json:"label"`
		Flags          rollouts.Rules `json:"flags"`
		RestoreVersion int64          `json:"restore_version"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil || body.Revision < 1 || body.RestoreVersion < 0 {
		platform.RespondError(w, 422, "INVALID_ROLLOUT", "Provide the current configuration and a version label", nil)
		return
	}
	body.Label = strings.TrimSpace(body.Label)
	if !utf8.ValidString(body.Label) || utf8.RuneCountInString(body.Label) < 1 || utf8.RuneCountInString(body.Label) > 80 || strings.ContainsAny(body.Label, "\x00\r\n") {
		platform.RespondError(w, 422, "INVALID_LABEL", "Use a version label of 1 to 80 characters", nil)
		return
	}
	if body.RestoreVersion == 0 {
		if body.Flags == nil {
			platform.RespondError(w, 422, "INVALID_FLAGS", "Provide feature settings", nil)
			return
		}
		if err := rollouts.Validate(body.Flags); err != nil {
			platform.RespondError(w, 422, "INVALID_FLAGS", err.Error(), nil)
			return
		}
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) < 16 || len(key) > 128 {
		platform.RespondError(w, 400, "INVALID_REQUEST_KEY", "Provide a request key", nil)
		return
	}
	for _, c := range key {
		if c < 33 || c > 126 {
			platform.RespondError(w, 400, "INVALID_REQUEST_KEY", "Provide a printable request key", nil)
			return
		}
	}
	tx, q, actor, reason, ok := h.campaignMutation(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	control, err := q.LockRuntimeRollout(r.Context())
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	canonical, _ := json.Marshal(struct {
		Body struct {
			Revision       int64          `json:"revision"`
			Label          string         `json:"label"`
			Flags          rollouts.Rules `json:"flags"`
			RestoreVersion int64          `json:"restore_version"`
		}
		Reason string
	}{body, reason})
	scope := actionHash([]byte(fmt.Sprintf("admin-rollout:%x", actor.UserID.Bytes)))
	hash := actionHash(canonical)
	keyHash := actionHash([]byte(key))
	replay, err := q.AdminActionReplay(r.Context(), db.AdminActionReplayParams{ScopeHash: scope, KeyHash: keyHash})
	if err == nil {
		if replay.RequestHash != hash {
			platform.RespondError(w, 409, "IDEMPOTENCY_CONFLICT", "This request key was used with other settings", nil)
			return
		}
		w.Header().Set("Idempotency-Replayed", "true")
		platform.RespondJSON(w, 200, json.RawMessage(replay.ResponseJson))
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		platform.RespondDatabaseError(w, err)
		return
	}
	if control.Revision != body.Revision {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "Configuration changed. Reload the saved version before publishing.", nil)
		return
	}
	if body.RestoreVersion > 0 {
		previous, e := q.GetRuntimeRolloutVersion(r.Context(), body.RestoreVersion)
		if errors.Is(e, pgx.ErrNoRows) {
			platform.RespondError(w, 404, "NOT_FOUND", "Configuration version not found", nil)
			return
		}
		if e != nil {
			platform.RespondDatabaseError(w, e)
			return
		}
		if e = json.Unmarshal(previous.Flags, &body.Flags); e != nil {
			platform.RespondDatabaseError(w, e)
			return
		}
		if e = rollouts.Validate(body.Flags); e != nil {
			platform.RespondError(w, 409, "VERSION_INCOMPATIBLE", "That version contains features unsupported by this release", nil)
			return
		}
	}
	flags, _ := json.Marshal(body.Flags)
	version, err := q.CreateRuntimeRolloutVersion(r.Context(), db.CreateRuntimeRolloutVersionParams{Label: body.Label, Flags: flags, RestoredFrom: pgtype.Int8{Int64: body.RestoreVersion, Valid: body.RestoreVersion > 0}, CreatedBy: actor.UserID})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = q.ActivateRuntimeRolloutVersion(r.Context(), version); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	metadata, _ := json.Marshal(struct {
		Reason       string `json:"reason"`
		Previous     int64  `json:"previous_version"`
		Version      int64  `json:"version"`
		RestoredFrom int64  `json:"restored_from,omitempty"`
	}{reason, control.ActiveVersion, version, body.RestoreVersion})
	if err = q.AppendAdminAudit(r.Context(), db.AppendAdminAuditParams{ActorUserID: actor.UserID, EntityID: control.ID, Action: "admin.rollout.published", RequestID: platform.GetRequestID(r.Context()), Metadata: metadata}); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	response, _ := json.Marshal(struct {
		Data struct {
			Version  int64 `json:"version"`
			Revision int64 `json:"revision"`
		} `json:"data"`
	}{struct {
		Version  int64 `json:"version"`
		Revision int64 `json:"revision"`
	}{version, control.Revision + 1}})
	if err = q.AdminActionRemember(r.Context(), db.AdminActionRememberParams{ScopeHash: scope, KeyHash: keyHash, RequestHash: hash, ResponseJson: response}); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	// Polling remains the recovery path if this eager refresh fails after commit.
	if h.rollouts != nil {
		if err := h.rollouts.Refresh(r.Context()); err != nil {
			slog.Warn("rollout saved; eager cache refresh failed, background refresh will retry")
		}
	}
	platform.RespondJSON(w, 200, json.RawMessage(response))
}
