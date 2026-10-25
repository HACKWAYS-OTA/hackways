package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/notifications"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func adminVerificationOrigin(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	allowed := origin == "https://admin.hackways.in"
	if os.Getenv("APP_ENV") == "development" {
		allowed = allowed || origin == "http://localhost:4003" || origin == "http://127.0.0.1:4003" || origin == "http://localhost:3103" || origin == "http://127.0.0.1:3103" || origin == "http://localhost:3003" || origin == "http://127.0.0.1:3003"
	}
	if !allowed {
		platform.RespondError(w, 403, "INVALID_ORIGIN", "Use the admin sign-in page", nil)
	}
	return allowed
}

func (h *Handler) RequestAdminVerification(w http.ResponseWriter, r *http.Request) {
	if !adminVerificationOrigin(w, r) {
		return
	}
	if !notifications.EmailConfiguredFor("admin_daily_otp") {
		platform.RespondError(w, 503, "EMAIL_UNAVAILABLE", "Admin verification email is unavailable. Your sign-in is preserved; try again later.", nil)
		return
	}
	user, _ := UserFromContext(r.Context())
	// Bcrypt must not hold a database transaction open on constrained hosts.
	code, hash, err := newEmailCode()
	if err != nil {
		platform.RespondError(w, 503, "CODE_UNAVAILABLE", "A verification code could not be created. Try again.", nil)
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
	actor, err := q.LockAuthenticationUser(r.Context(), user.UserID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	allowed, err := q.IsPlatformAdminIdentitySession(r.Context(), db.IsPlatformAdminIdentitySessionParams{SessionID: user.SessionID, UserID: user.UserID})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if !allowed {
		platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "Admin access required", nil)
		return
	}
	session, err := q.LockAdminVerificationSession(r.Context(), db.LockAdminVerificationSessionParams{SessionID: user.SessionID, UserID: user.UserID})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if session.OtpVerifiedAt.Valid && session.OtpVerifiedAt.Time.After(session.DatabaseNow.Time.Add(-24*time.Hour)) {
		w.WriteHeader(204)
		return
	}
	previous, err := q.ReadAdminSessionChallenge(r.Context(), user.SessionID)
	if err == nil && session.DatabaseNow.Time.Sub(previous.CreatedAt.Time) < time.Minute {
		platform.RespondError(w, 429, "CODE_COOLDOWN", "Wait one minute before requesting another code", nil)
		return
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = q.CreateAdminSessionChallenge(r.Context(), db.CreateAdminSessionChallengeParams{SessionID: user.SessionID, CodeHash: hash}); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	payload, _ := json.Marshal(otpPayload{Code: code, Hash: hash})
	if err = enqueueAccountEmail(r.Context(), q, actor.Email, actor.DisplayName, "admin_daily_otp", payload); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 202, struct {
		Message string `json:"message"`
	}{"Check your inbox for your verification code. It expires in 10 minutes."})
}

func (h *Handler) ConfirmAdminVerification(w http.ResponseWriter, r *http.Request) {
	if !adminVerificationOrigin(w, r) {
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Enter the six-digit verification code", nil)
		return
	}
	body.Code = strings.TrimSpace(body.Code)
	user, _ := UserFromContext(r.Context())
	// Compare before acquiring row locks. Recheck the exact persisted challenge below.
	candidate, candidateErr := h.queries.ReadAdminSessionChallenge(r.Context(), user.SessionID)
	if candidateErr != nil && !errors.Is(candidateErr, pgx.ErrNoRows) {
		platform.RespondDatabaseError(w, candidateErr)
		return
	}
	candidateMatches := candidateErr == nil && matchesEmailCode(body.Code, candidate.CodeHash)
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
	if _, err = q.LockAuthenticationUser(r.Context(), user.UserID); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	allowed, err := q.IsPlatformAdminIdentitySession(r.Context(), db.IsPlatformAdminIdentitySessionParams{SessionID: user.SessionID, UserID: user.UserID})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if !allowed {
		platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "Admin access required", nil)
		return
	}
	session, err := q.LockAdminVerificationSession(r.Context(), db.LockAdminVerificationSessionParams{SessionID: user.SessionID, UserID: user.UserID})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if session.OtpVerifiedAt.Valid && session.OtpVerifiedAt.Time.After(session.DatabaseNow.Time.Add(-24*time.Hour)) {
		h.RefreshAdminSessionCookie(w, r)
		return
	}
	challenge, err := q.ReadAdminSessionChallenge(r.Context(), user.SessionID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		platform.RespondDatabaseError(w, err)
		return
	}
	valid := err == nil && !challenge.ConsumedAt.Valid && challenge.ExpiresAt.Time.After(session.DatabaseNow.Time) && candidateMatches && candidate.CodeHash == challenge.CodeHash && challenge.ExpiresAt.Time.After(time.Now())
	action := "admin.session.verified"
	if valid {
		if err = q.ConsumeAdminSessionChallenge(r.Context(), user.SessionID); err == nil {
			err = q.MarkAdminSessionVerified(r.Context(), user.SessionID)
		}
	} else {
		action = "admin.session.verification_failed"
		err = q.DeleteSessionByTokenHash(r.Context(), HashToken(GetSessionToken(r, AdminSessionCookie)))
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = q.AppendAdminAudit(r.Context(), db.AppendAdminAuditParams{ActorUserID: user.UserID, EntityID: user.SessionID, EventID: pgtype.UUID{}, Action: action, RequestID: platform.GetRequestID(r.Context()), Metadata: []byte(`{}`)}); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if !valid {
		ClearSessionCookie(w, AdminSessionCookie)
		platform.RespondError(w, 401, "ADMIN_VERIFICATION_FAILED", "The code was incorrect or expired. This device has been signed out. Sign in again.", nil)
		return
	}
	h.RefreshAdminSessionCookie(w, r)
}

func (h *Handler) RefreshAdminSessionCookie(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	remaining := time.Until(user.ExpiresAt)
	if remaining < time.Second {
		platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
		return
	}
	deadline, err := h.queries.AdminSessionVerificationDeadline(r.Context(), user.SessionID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if !deadline.Valid || !deadline.Time.After(time.Now()) {
		platform.RespondError(w, 403, "ADMIN_VERIFICATION_REQUIRED", "Daily admin verification is required", nil)
		return
	}
	w.Header().Set("X-Admin-Verification-Seconds", fmt.Sprint(int64(time.Until(deadline.Time).Seconds())))
	SetSessionCookie(w, AdminSessionCookie, GetSessionToken(r, AdminSessionCookie), remaining, true)
	w.WriteHeader(204)
}
