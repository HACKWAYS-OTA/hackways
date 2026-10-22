package events

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
)

func (h *Handler) ChangeGuestStatus(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	var body struct {
		Status             db.RegistrationStatus `json:"status"`
		UpdatedAt          time.Time             `json:"updated_at"`
		Notify             *bool                 `json:"notify_guest"`
		Message            string                `json:"message"`
		ResendNotification bool                  `json:"resend_notification"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil || body.UpdatedAt.IsZero() || body.Notify == nil {
		platform.RespondError(w, 422, "INVALID_STATUS_CHANGE", "Provide a status, the current guest version and a notification choice", nil)
		return
	}
	if body.Status != db.RegistrationStatusConfirmed && body.Status != db.RegistrationStatusCancelled && body.Status != db.RegistrationStatusRejected && body.Status != db.RegistrationStatusPendingApproval && body.Status != db.RegistrationStatusWaitlisted && !(body.ResendNotification && body.Status == db.RegistrationStatusOffered) {
		platform.RespondError(w, 422, "INVALID_STATUS", "Choose an available guest status", nil)
		return
	}
	body.Message = strings.TrimSpace(body.Message)
	if utf8.RuneCountInString(body.Message) > 2000 || (!*body.Notify && body.Message != "") {
		platform.RespondError(w, 422, "INVALID_MESSAGE", "Messages require Notify guest and must contain at most 2000 characters", nil)
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if key == "" || !workflows.ValidateKey(key) {
		platform.RespondError(w, 400, "INVALID_IDEMPOTENCY_KEY", "Provide a random request key of 16 to 128 printable characters", nil)
		return
	}
	guest, err := h.queries.ManagedGuestDetail(r.Context(), db.ManagedGuestDetailParams{EventID: event.ID, PublicCode: chi.URLParam(r, "code")})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	owner, _ := auth.OrganizerFromContext(r.Context())
	result, replayed, err := workflows.New(h.pool).ChangeOwnerStatus(r.Context(), workflows.OwnerStatusRequest{
		EventID: event.ID, RegistrationID: guest.ID, ActorID: actor.UserID, OrganizerID: owner.ID,
		ResendNotification: body.ResendNotification, Status: body.Status, ExpectedUpdatedAt: body.UpdatedAt, Notify: *body.Notify, Message: body.Message, Key: key, RequestID: platform.GetRequestID(r.Context()),
	})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
	}
	platform.RespondJSON(w, 200, result)
}
