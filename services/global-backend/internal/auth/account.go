package auth

import (
	"errors"
	"github.com/hackways/global-backend/internal/notifications"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	}{"Check your inbox for the notification."})
}
func (h *Handler) RequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	if user.EmailVerified {
		platform.RespondJSON(w, 200, struct {
			Message string `json:"message"`
		}{"Email already verified"})
		return
	}
	if !notifications.EmailConfiguredFor("email_verification") {
		platform.RespondError(w, 503, "EMAIL_UNAVAILABLE", "Verification email is temporarily unavailable. Please try again later.", nil)
		return
	}
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
	if err = q.SetStorageWriteTimeout(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	locked, err := q.LockAuthenticationUser(r.Context(), user.UserID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if locked.EmailVerified {
		platform.RespondJSON(w, 200, struct {
			Message string `json:"message"`
		}{"Email already verified"})
		return
	}
	previous, err := q.ReadEmailVerification(r.Context(), user.UserID)
	if err == nil && time.Since(previous.CreatedAt.Time) < time.Minute {
		wait := int(math.Ceil(time.Until(previous.CreatedAt.Time.Add(time.Minute)).Seconds()))
		if wait < 1 {
			wait = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(wait))
		platform.RespondError(w, 429, "CODE_RECENT", "Wait a minute before requesting another code", nil)
		return
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = queueEmailVerification(r.Context(), q, user.UserID, locked.Email, locked.DisplayName); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 202, struct {
		Message string `json:"message"`
	}{"Check your inbox for the six-digit verification code. It expires in 15 minutes."})
}

func (h *Handler) ConfirmEmailVerification(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_CODE", "Enter the code from your email", nil)
		return
	}
	body.Code = strings.TrimSpace(body.Code)
	if !validEmailCode(body.Code) && len(body.Code) != 64 {
		platform.RespondError(w, 400, "INVALID_CODE", "Enter the six-digit code from your email", nil)
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
	if err = q.SetStorageWriteTimeout(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	locked, err := q.LockAuthenticationUser(r.Context(), user.UserID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	challenge, err := q.ReadEmailVerification(r.Context(), user.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 409, "CODE_EXPIRED", "Request a new verification code", nil)
		return
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if challenge.ConsumedAt.Valid || challenge.Attempts >= 5 || !challenge.ExpiresAt.Time.After(time.Now()) || !strings.EqualFold(challenge.Email, locked.Email) {
		platform.RespondError(w, 409, "CODE_EXPIRED", "The code expired or was used. Request a new one.", nil)
		return
	}
	if !matchesEmailCode(body.Code, challenge.TokenHash) {
		if err = q.FailEmailVerificationAttempt(r.Context(), user.UserID); err == nil {
			err = tx.Commit(r.Context())
		}
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		platform.RespondError(w, 400, "INVALID_CODE", "The code is incorrect. Check your email and try again.", nil)
		return
	}
	if _, err = q.ConsumeEmailVerification(r.Context(), challenge.TokenHash); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	affected, err := q.VerifyAccountEmail(r.Context(), db.VerifyAccountEmailParams{ID: user.UserID, Lower: locked.Email})
	if err != nil || affected != 1 {
		if err == nil {
			err = errors.New("account changed during verification")
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
