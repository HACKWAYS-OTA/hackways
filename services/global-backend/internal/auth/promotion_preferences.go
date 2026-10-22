package auth

import (
	"net/http"
	"strings"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
)

func (h *Handler) PromotionPreference(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	subscribed, err := h.queries.ReadPromotionPreference(r.Context(), user.Email)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Subscribed bool `json:"subscribed"`
	}{subscribed})
}
func (h *Handler) SetPromotionPreference(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Subscribed *bool `json:"subscribed"`
	}
	if platform.DecodeJSON(r, &body) != nil || body.Subscribed == nil {
		platform.RespondError(w, 400, "INVALID_PREFERENCE", "Choose whether to receive promotional email", nil)
		return
	}
	user, _ := UserFromContext(r.Context())
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
	if *body.Subscribed && !locked.EmailVerified {
		platform.RespondError(w, 403, "EMAIL_NOT_VERIFIED", "Verify your email before subscribing", nil)
		return
	}
	err = q.SetPromotionPreference(r.Context(), db.SetPromotionPreferenceParams{Email: strings.ToLower(locked.Email), DisplayName: locked.DisplayName, UserID: locked.ID, Subscribed: *body.Subscribed})
	if err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Subscribed bool `json:"subscribed"`
	}{*body.Subscribed})
}
func (h *Handler) UnsubscribePromotion(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key string `json:"key"`
	}
	if platform.DecodeJSON(r, &body) != nil || !validResetToken(body.Key) {
		platform.RespondError(w, 400, "INVALID_LINK", "Open the unsubscribe link from your email", nil)
		return
	}
	if !h.allowIdentityAttempt(w, r, HashToken(body.Key)) {
		return
	}
	changed, err := h.queries.UnsubscribePromotion(r.Context(), HashToken(body.Key))
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if changed != 1 {
		platform.RespondError(w, 400, "INVALID_LINK", "This unsubscribe link could not be verified", nil)
		return
	}
	// Generic/idempotent response never discloses an email address.
	platform.RespondJSON(w, 200, struct {
		Message string `json:"message"`
	}{"Promotional email is turned off for this address. Account and event emails are unchanged."})
}
