package auth

import (
	"encoding/json"
	"errors"
	"github.com/hackways/global-backend/internal/notifications"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) Sessions(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	rows, err := h.queries.ListAccountSessions(r.Context(), user.UserID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.ListAccountSessionsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data    []db.ListAccountSessionsRow `json:"data"`
		Current pgtype.UUID                 `json:"current_session_id"`
	}{rows, user.SessionID})
}
func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "sessionID")); err != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid session ID", nil)
		return
	}
	affected, err := h.queries.RevokeOwnedSession(r.Context(), db.RevokeOwnedSessionParams{ID: id, UserID: user.UserID})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if affected != 1 {
		platform.RespondError(w, 404, "NOT_FOUND", "Session not found", nil)
		return
	}
	if id == user.SessionID {
		ClearSessionCookie(w, DefaultSessionCookie)
	}
	w.WriteHeader(204)
}
func (h *Handler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	if err := h.sm.RevokeAllUserSessions(r.Context(), user.UserID); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	ClearSessionCookie(w, DefaultSessionCookie)
	w.WriteHeader(204)
}
func (h *Handler) Notifications(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	limit, offset := int32(20), int32(0)
	if n, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32); err == nil && n > 0 && n <= 100 {
		limit = int32(n)
	}
	if n, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32); err == nil && n >= 0 {
		offset = int32(n)
	}
	rows, err := h.queries.ListOwnedNotifications(r.Context(), db.ListOwnedNotificationsParams{UserID: user.UserID, Limit: limit, Offset: offset})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.ListOwnedNotificationsRow{}
	}
	platform.RespondJSON(w, 200, struct {
		Data []db.ListOwnedNotificationsRow `json:"data"`
	}{rows})
}
func (h *Handler) RetryNotification(w http.ResponseWriter, r *http.Request) {
	if !notifications.EmailConfigured() {
		platform.RespondError(w, 503, "EMAIL_UNAVAILABLE", "Email delivery is unavailable until ACS is configured", nil)
		return
	}
	user, _ := UserFromContext(r.Context())
	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "notificationID")); err != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid notification ID", nil)
		return
	}
	if !h.allowIdentityAttempt(w, r, user.Email) {
		return
	}
	_, err := h.queries.RetryOwnedNotification(r.Context(), db.RetryOwnedNotificationParams{ID: id, UserID: user.UserID})
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 409, "RETRY_UNAVAILABLE", "Only your failed notifications older than five minutes can be retried", nil)
		return
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 202, struct {
		Message string `json:"message"`
	}{"Notification queued; delivery is not yet confirmed"})
}
func (h *Handler) RequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	if user.EmailVerified {
		platform.RespondJSON(w, 200, struct {
			Message string `json:"message"`
		}{"Email already verified"})
		return
	}
	if !notifications.EmailConfigured() {
		platform.RespondError(w, 503, "EMAIL_UNAVAILABLE", "Email verification is unavailable until ACS is configured", nil)
		return
	}
	if !h.allowIdentityAttempt(w, r, user.Email) {
		return
	}
	token, err := GenerateToken()
	if err != nil {
		platform.RespondError(w, 503, "TOKEN_UNAVAILABLE", "Verification is temporarily unavailable", nil)
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	if err = q.CreateEmailVerification(r.Context(), db.CreateEmailVerificationParams{UserID: user.UserID, Email: user.Email, TokenHash: HashToken(token)}); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	payload, _ := json.Marshal(struct {
		Code string `json:"verification_code"`
	}{token})
	_, err = q.EnqueueEmailOutbox(r.Context(), db.EnqueueEmailOutboxParams{RecipientEmail: user.Email, RecipientName: pgtype.Text{String: user.DisplayName, Valid: true}, TemplateType: "email_verification", PayloadJson: payload})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 202, struct {
		Message string `json:"message"`
	}{"Verification email queued; the code expires in 15 minutes"})
}
func (h *Handler) ConfirmEmailVerification(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil || len(body.Code) != 64 {
		platform.RespondError(w, 400, "INVALID_CODE", "Provide the verification code from your email", nil)
		return
	}
	user, _ := UserFromContext(r.Context())
	if !h.allowIdentityAttempt(w, r, user.Email) {
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	verification, err := q.ConsumeEmailVerification(r.Context(), HashToken(body.Code))
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 409, "CODE_EXPIRED", "The code expired or was used; request a new one", nil)
		return
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if verification.UserID != user.UserID || !strings.EqualFold(verification.Email, user.Email) {
		platform.RespondError(w, 409, "CODE_MISMATCH", "The code does not belong to this signed-in account", nil)
		return
	}
	affected, err := q.VerifyAccountEmail(r.Context(), db.VerifyAccountEmailParams{ID: user.UserID, Lower: user.Email})
	if err != nil || affected != 1 {
		if err == nil {
			err = errors.New("account email changed")
		}
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Message string `json:"message"`
	}{"Email verified"})
}
