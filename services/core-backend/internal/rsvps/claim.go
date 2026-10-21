package rsvps

import (
	"errors"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"net/http"
)

// ClaimGuest links only the signed-in account's verified email. It never merges
// competing account/guest reservations or changes capacity during identity linkage.
func (h *Handler) ClaimGuest(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	if !user.EmailVerified {
		platform.RespondError(w, 403, "EMAIL_VERIFICATION_REQUIRED", "Verify your email before claiming guest registrations", nil)
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		respondFailure(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	if err = q.SetWorkflowTimeout(r.Context()); err != nil {
		respondFailure(w, err)
		return
	}
	ids, err := q.GuestClaimEvents(r.Context(), db.GuestClaimEventsParams{Email: user.Email, UserID: user.UserID})
	if err != nil {
		respondFailure(w, err)
		return
	}
	for _, id := range ids {
		if _, err = q.LockWorkflowEvent(r.Context(), id); err != nil {
			respondFailure(w, err)
			return
		}
		// Recheck after acquiring the same event lock used by registration creation.
		if _, readErr := q.GetAccountEventRSVP(r.Context(), db.GetAccountEventRSVPParams{EventID: id, UserID: user.UserID}); readErr == nil {
			continue
		} else if !errors.Is(readErr, pgx.ErrNoRows) {
			respondFailure(w, readErr)
			return
		}
		registration, claimErr := q.ClaimGuestEventRegistration(r.Context(), db.ClaimGuestEventRegistrationParams{UserID: user.UserID, EventID: id, Email: user.Email})
		if claimErr != nil {
			respondFailure(w, claimErr)
			return
		}
		if err = workflows.Audit(r.Context(), q, user.UserID, id, registration.ID, "registration.claimed", platform.GetRequestID(r.Context())); err != nil {
			respondFailure(w, err)
			return
		}
	}
	if err = tx.Commit(r.Context()); err != nil {
		respondFailure(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Message   string `json:"message"`
		Processed int    `json:"processed_events"`
		HasMore   bool   `json:"may_have_more"`
	}{"Matching guest registrations linked; existing account conflicts were left unchanged", len(ids), len(ids) == 50})
}
