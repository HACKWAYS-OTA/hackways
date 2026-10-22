package admin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type actionBody struct {
	UpdatedAt   time.Time `json:"updated_at"`
	DisplayName *string   `json:"display_name,omitempty"`
	Access      *string   `json:"access,omitempty"`
	Kind        *string   `json:"kind,omitempty"`
}
type actionAudit struct {
	Reason    string `json:"reason"`
	Previous  string `json:"previous,omitempty"`
	Value     string `json:"value,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

func (h *Handler) mountActions(r chi.Router) {
	r.Post("/people/{id}/edit", h.action("person.edit"))
	r.Post("/people/{id}/access", h.action("person.access"))
	r.Post("/people/{id}/sessions/{sessionID}/revoke", h.action("session.revoke"))
	r.Post("/messaging/{id}/retry", h.action("message.retry"))
	r.Post("/storage/{id}/detach", h.action("asset.detach"))
}
func actionHash(value []byte) string { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }
func actionConflict(w http.ResponseWriter) {
	platform.RespondError(w, 409, "RECORD_CHANGED", "This record changed. Reload it before continuing.", nil)
}

func (h *Handler) action(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := origin == "https://admin.hackways.in"
		if os.Getenv("APP_ENV") == "development" {
			allowed = allowed || origin == "http://localhost:3103" || origin == "http://127.0.0.1:3103" || origin == "http://localhost:3003" || origin == "http://127.0.0.1:3003"
		}
		if !allowed {
			platform.RespondError(w, 403, "INVALID_ORIGIN", "Use the admin console to make changes", nil)
			return
		}
		reason, err := url.PathUnescape(r.Header.Get("X-Admin-Reason"))
		reason = strings.TrimSpace(reason)
		if err != nil || !utf8.ValidString(reason) || utf8.RuneCountInString(reason) < 3 || utf8.RuneCountInString(reason) > 1000 || strings.ContainsAny(reason, "\x00\r\n") {
			platform.RespondError(w, 400, "REASON_REQUIRED", "Enter a reason between 3 and 1,000 characters", nil)
			return
		}
		key := r.Header.Get("Idempotency-Key")
		if len(key) < 16 || len(key) > 128 || strings.ContainsAny(key, "\r\n\t ") {
			platform.RespondError(w, 400, "IDEMPOTENCY_KEY_REQUIRED", "A valid action key is required", nil)
			return
		}
		for _, char := range key {
			if char < 33 || char > 126 {
				platform.RespondError(w, 400, "IDEMPOTENCY_KEY_REQUIRED", "Use a printable action key", nil)
				return
			}
		}
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
			return
		}
		var body actionBody
		if platform.DecodeJSON(r, &body) != nil || body.UpdatedAt.IsZero() {
			platform.RespondError(w, 400, "INVALID_REQUEST", "A current record version is required", nil)
			return
		}
		// Reject fields belonging to a different action, even if they are otherwise known.
		valid := body.DisplayName == nil && body.Access == nil && body.Kind == nil
		switch action {
		case "person.edit":
			valid = body.DisplayName != nil && body.Access == nil && body.Kind == nil
		case "person.access":
			valid = body.Access != nil && body.DisplayName == nil && body.Kind == nil
		case "asset.detach":
			valid = body.Kind != nil && body.DisplayName == nil && body.Access == nil
		}
		if !valid {
			platform.RespondError(w, 400, "INVALID_REQUEST", "Invalid action fields", nil)
			return
		}
		resource, kind := "people", ""
		if action == "message.retry" {
			resource = "messaging"
		}
		if action == "asset.detach" {
			resource = "storage"
			kind = *body.Kind
		}
		id, ok := h.detailID(w, r, resource, kind)
		if !ok {
			return
		}
		canonical, _ := json.Marshal(struct {
			Body              actionBody
			Reason, SessionID string
		}{body, reason, chi.URLParam(r, "sessionID")})
		scope := actionHash([]byte(fmt.Sprintf("admin:%x:%x:%s:%s", user.UserID.Bytes, id.Bytes, action, chi.URLParam(r, "sessionID"))))
		keyHash, requestHash := actionHash([]byte(key)), actionHash(canonical)
		tx, err := h.pool.Begin(r.Context())
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		defer tx.Rollback(r.Context())
		q := h.queries.WithTx(tx)
		if err = q.SetStorageWriteTimeout(r.Context()); err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		// Serialize role changes with privileged writes and recheck the live grant
		// inside the transaction, including for idempotent replays.
		if err = q.AdminLockRoleChanges(r.Context()); err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		permitted, err := q.IsPlatformAdminSession(r.Context(), db.IsPlatformAdminSessionParams{SessionID: user.SessionID, UserID: user.UserID})
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		if !permitted {
			platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "This account does not have admin access", nil)
			return
		}
		if err = q.AdminSerializeAction(r.Context(), scope); err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		replay, err := q.AdminActionReplay(r.Context(), db.AdminActionReplayParams{ScopeHash: scope, KeyHash: keyHash})
		if err == nil {
			if replay.RequestHash != requestHash {
				platform.RespondError(w, 409, "IDEMPOTENCY_CONFLICT", "This action key was used with different changes", nil)
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
		metadata := actionAudit{Reason: reason}
		var eventID pgtype.UUID
		switch action {
		case "person.edit", "person.access", "session.revoke":
			target, readErr := q.AdminUserLock(r.Context(), id)
			if readErr != nil {
				detailError(w, readErr)
				return
			}
			if !target.UpdatedAt.Time.Equal(body.UpdatedAt) {
				actionConflict(w)
				return
			}
			switch action {
			case "person.edit":
				name := strings.TrimSpace(*body.DisplayName)
				if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 100 || strings.ContainsAny(name, "\r\n\t") {
					platform.RespondError(w, 422, "INVALID_NAME", "Enter a name between 1 and 100 characters", nil)
					return
				}
				metadata.Previous = target.DisplayName
				metadata.Value = name
				err = q.AdminUpdatePersonName(r.Context(), db.AdminUpdatePersonNameParams{ID: id, DisplayName: name})
			case "person.access":
				if id == user.UserID {
					platform.RespondError(w, 409, "SELF_ACCESS_CHANGE", "You cannot change your own admin access", nil)
					return
				}
				if *body.Access != "admin" && *body.Access != "member" {
					platform.RespondError(w, 422, "INVALID_ACCESS", "Choose admin or member access", nil)
					return
				}
				existing, accessErr := q.IsPlatformAdministrator(r.Context(), id)
				if accessErr != nil {
					platform.RespondDatabaseError(w, accessErr)
					return
				}
				metadata.Previous = "member"
				if existing {
					metadata.Previous = "admin"
				}
				metadata.Value = *body.Access
				if *body.Access == "admin" {
					if !target.EmailVerified {
						platform.RespondError(w, 422, "EMAIL_NOT_VERIFIED", "This account must verify its email before receiving admin access", nil)
						return
					}
					err = q.AdminGrantAccess(r.Context(), db.AdminGrantAccessParams{UserID: id, GrantedBy: user.UserID})
				} else {
					count, countErr := q.AdminActiveCount(r.Context())
					if countErr != nil {
						platform.RespondDatabaseError(w, countErr)
						return
					}
					if existing && count <= 1 {
						platform.RespondError(w, 409, "LAST_ADMIN", "The last administrator cannot be removed", nil)
						return
					}
					_, err = q.AdminRevokeAccess(r.Context(), id)
					if err == nil {
						err = q.AdminRevokeAdminSessions(r.Context(), id)
					}
				}
				if err == nil {
					err = q.AdminTouchPerson(r.Context(), id)
				}
			case "session.revoke":
				var sessionID pgtype.UUID
				rawSession := chi.URLParam(r, "sessionID")
				if reference, parseErr := strconv.ParseInt(rawSession, 10, 64); parseErr == nil && reference > 0 {
					sessionID, err = q.AdminResolveReferenceID(r.Context(), db.AdminResolveReferenceIDParams{Reference: reference, Resource: "sessions"})
					if err != nil {
						detailError(w, err)
						return
					}
				} else if sessionID.Scan(rawSession) != nil || !sessionID.Valid {
					platform.RespondError(w, 400, "INVALID_SESSION", "Invalid session ID", nil)
					return
				}
				if sessionID == user.SessionID {
					platform.RespondError(w, 409, "CURRENT_SESSION", "Use Sign out to end your current session", nil)
					return
				}
				var count int64
				count, err = q.AdminRevokePersonSession(r.Context(), db.AdminRevokePersonSessionParams{UserID: id, ID: sessionID, CurrentSessionID: user.SessionID})
				if err == nil && count == 0 {
					platform.RespondError(w, 404, "SESSION_NOT_FOUND", "Session not found for this account", nil)
					return
				}
				if err == nil {
					err = q.AdminTouchPerson(r.Context(), id)
				}
				metadata.SessionID = chi.URLParam(r, "sessionID")
			}
		case "message.retry":
			message, readErr := q.AdminMessageDetail(r.Context(), id)
			if readErr != nil {
				detailError(w, readErr)
				return
			}
			eventID = message.EventID
			_, err = q.AdminRetryMessage(r.Context(), db.AdminRetryMessageParams{ID: id, UpdatedAt: pgtype.Timestamptz{Time: body.UpdatedAt, Valid: true}})
			if errors.Is(err, pgx.ErrNoRows) {
				platform.RespondError(w, 409, "MESSAGE_NOT_RETRYABLE", "Only failed messages with no recorded provider attempt can be retried. Reload this record.", nil)
				return
			}
		case "asset.detach":
			switch *body.Kind {
			case "event_banner":
				eventID = id
				_, err = q.AdminDetachEventBanner(r.Context(), db.AdminDetachEventBannerParams{ID: id, UpdatedAt: pgtype.Timestamptz{Time: body.UpdatedAt, Valid: true}})
			case "avatar":
				_, err = q.AdminDetachAvatar(r.Context(), db.AdminDetachAvatarParams{ID: id, UpdatedAt: pgtype.Timestamptz{Time: body.UpdatedAt, Valid: true}})
			default:
				platform.RespondError(w, 422, "INVALID_ASSET_TYPE", "Choose an asset type", nil)
				return
			}
			metadata.Value = *body.Kind
			if errors.Is(err, pgx.ErrNoRows) {
				actionConflict(w)
				return
			}
		}
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		encodedAudit, _ := json.Marshal(metadata)
		if err = q.AppendAdminAudit(r.Context(), db.AppendAdminAuditParams{ActorUserID: user.UserID, EventID: eventID, EntityID: id, Action: "admin." + action, RequestID: platform.GetRequestID(r.Context()), Metadata: encodedAudit}); err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		response, _ := json.Marshal(struct {
			Data struct {
				ID     pgtype.UUID `json:"id"`
				Action string      `json:"action"`
			} `json:"data"`
		}{Data: struct {
			ID     pgtype.UUID `json:"id"`
			Action string      `json:"action"`
		}{id, action}})
		if err = q.AdminActionRemember(r.Context(), db.AdminActionRememberParams{ScopeHash: scope, KeyHash: keyHash, RequestHash: requestHash, ResponseJson: response}); err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		if err = tx.Commit(r.Context()); err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		platform.RespondJSON(w, 200, json.RawMessage(response))
	}
}
