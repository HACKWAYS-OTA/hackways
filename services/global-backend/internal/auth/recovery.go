package auth

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/notifications"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
)

type resetLinkPayload struct {
	ActionURL string `json:"action_url"`
	Hash      string `json:"verification_hash"`
}

func validResetToken(token string) bool {
	raw, err := hex.DecodeString(token)
	return err == nil && len(raw) == 32
}

func recoveryEmail(raw string) (string, bool) {
	email := strings.ToLower(strings.TrimSpace(raw))
	address, err := mail.ParseAddress(email)
	return email, err == nil && address.Address == email && len(email) <= 320
}
func recoveryRequested(w http.ResponseWriter) {
	platform.RespondJSON(w, 202, struct {
		Message string `json:"message"`
	}{"If this account exists, a password reset link is queued. Check your email. The link expires in 15 minutes."})
}
func invalidRecoveryToken(w http.ResponseWriter) {
	platform.RespondError(w, 400, "INVALID_RESET_TOKEN", "This reset link is invalid, expired or already used. Request a new link.", nil)
}

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Enter your email address", nil)
		return
	}
	email, valid := recoveryEmail(body.Email)
	if !valid {
		platform.RespondError(w, 400, "INVALID_EMAIL", "Enter a valid email address", nil)
		return
	}
	if !h.allowIdentityAttempt(w, r, email) {
		return
	}
	if !notifications.EmailConfigured() {
		platform.RespondError(w, 503, "EMAIL_UNAVAILABLE", "Password recovery email is temporarily unavailable. Please try again later.", nil)
		return
	}
	// Both unknown and known accounts perform the same cryptographic work. Never
	// call the email provider inline or disclose account existence in the response.
	token, err := GenerateToken()
	if err != nil {
		platform.RespondError(w, 503, "TOKEN_UNAVAILABLE", "Recovery is temporarily unavailable", nil)
		return
	}
	hash := HashToken(token)
	user, err := h.queries.GetUserByEmail(r.Context(), email)
	if errors.Is(err, pgx.ErrNoRows) {
		recoveryRequested(w)
		return
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
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
	locked, err := q.LockAuthenticationUser(r.Context(), user.ID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	previous, err := q.ReadPasswordResetChallenge(r.Context(), user.ID)
	if err == nil && time.Since(previous.CreatedAt.Time) < time.Minute {
		recoveryRequested(w)
		return
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = q.CreatePasswordResetChallenge(r.Context(), db.CreatePasswordResetChallengeParams{UserID: user.ID, TokenHash: hash}); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	payload, err := json.Marshal(resetLinkPayload{ActionURL: "https://hackways.in/auth/reset-password?key=" + token, Hash: hash})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = enqueueAccountEmail(r.Context(), q, locked.Email, locked.DisplayName, "password_reset_requested", payload); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	recoveryRequested(w)
}

func (h *Handler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Open your reset link and enter a new password", nil)
		return
	}
	body.Token = strings.TrimSpace(body.Token)
	if !validResetToken(body.Token) {
		invalidRecoveryToken(w)
		return
	}
	if len(body.Password) < 8 || len(body.Password) > 72 {
		platform.RespondError(w, 400, "INVALID_PASSWORD", "Use a password between 8 and 72 bytes", nil)
		return
	}
	if !h.allowIdentityAttempt(w, r, HashToken(body.Token)) {
		return
	}
	userID, err := h.queries.FindPasswordResetUser(r.Context(), HashToken(body.Token))
	if errors.Is(err, pgx.ErrNoRows) {
		invalidRecoveryToken(w)
		return
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	password, err := HashPassword(body.Password)
	if err != nil {
		platform.RespondError(w, 503, "PASSWORD_UNAVAILABLE", "The password could not be secured. Try again.", nil)
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
	locked, err := q.LockAuthenticationUser(r.Context(), userID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	challenge, err := q.ReadPasswordResetChallenge(r.Context(), userID)
	if errors.Is(err, pgx.ErrNoRows) {
		invalidRecoveryToken(w)
		return
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if challenge.TokenHash != HashToken(body.Token) || challenge.ConsumedAt.Valid || !challenge.ExpiresAt.Time.After(time.Now()) {
		invalidRecoveryToken(w)
		return
	}
	changed, err := q.ResetAccountPassword(r.Context(), db.ResetAccountPasswordParams{ID: userID, PasswordHash: password})
	if err != nil || changed != 1 {
		if err == nil {
			err = errors.New("password reset account changed")
		}
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = q.ConsumePasswordResetChallenge(r.Context(), userID); err == nil {
		err = q.InvalidateUserEmailCodes(r.Context(), userID)
	}
	if err == nil {
		err = q.DeleteSessionsByUserID(r.Context(), userID)
	}
	if err == nil {
		err = q.RevokeUserOAuthExchanges(r.Context(), userID)
	}
	if err == nil {
		err = q.ForgetUserDevices(r.Context(), userID)
	}
	if err == nil {
		err = enqueueAccountEmail(r.Context(), q, locked.Email, locked.DisplayName, "password_changed", []byte(`{}`))
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	ClearSessionCookie(w, DefaultSessionCookie)
	ClearSessionCookie(w, AdminSessionCookie)
	ClearSessionCookie(w, deviceCookie)
	ClearSessionCookie(w, adminDeviceCookie)
	platform.RespondJSON(w, 200, struct {
		Message string `json:"message"`
	}{"Password changed. Sign in again on your devices."})
}
