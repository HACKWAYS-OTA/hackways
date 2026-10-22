package events

import (
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"net/http"
	"net/mail"
	"sort"
	"strings"
	"unicode/utf8"
)

func (h *Handler) InviteAttendees(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	actor, _ := auth.UserFromContext(r.Context())
	owner, _ := auth.OrganizerFromContext(r.Context())
	if !actor.EmailVerified {
		platform.RespondError(w, 403, "EMAIL_VERIFICATION_REQUIRED", "Verify your account email before sending invitations", nil)
		return
	}
	var body struct {
		Emails  []string `json:"emails"`
		Message string   `json:"message"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil || len(body.Emails) < 1 || len(body.Emails) > 50 {
		platform.RespondError(w, 422, "INVALID_RECIPIENTS", "Enter 1 to 50 email addresses", nil)
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if !workflows.ValidateKey(key) {
		platform.RespondError(w, 400, "INVALID_IDEMPOTENCY_KEY", "Provide a random request key of 16 to 128 printable characters", nil)
		return
	}
	body.Message = strings.TrimSpace(body.Message)
	if !utf8.ValidString(body.Message) || utf8.RuneCountInString(body.Message) > 2000 {
		platform.RespondError(w, 422, "INVALID_MESSAGE", "Use a message with at most 2000 characters", nil)
		return
	}
	emails := make([]string, 0, len(body.Emails))
	seen := make(map[string]bool, len(body.Emails))
	for _, raw := range body.Emails {
		email := strings.ToLower(strings.TrimSpace(raw))
		address, err := mail.ParseAddress(email)
		if err != nil || address.Address != email || len(email) > 254 {
			platform.RespondError(w, 422, "INVALID_EMAIL", "Each recipient must be a valid email address", nil)
			return
		}
		if !seen[email] {
			seen[email] = true
			emails = append(emails, email)
		}
	}
	sort.Strings(emails)
	result, replayed, err := workflows.New(h.pool).Invite(r.Context(), workflows.InvitationRequest{EventID: event.ID, OrganizerID: owner.ID, ActorID: actor.UserID, Emails: emails, Message: body.Message, Key: key, RequestID: platform.GetRequestID(r.Context())})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
	}
	platform.RespondJSON(w, 200, result)
}

func (h *Handler) Invitations(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	limit, offset := pagination(r)
	rows, err := h.queries.ListEventInvitations(r.Context(), db.ListEventInvitationsParams{EventID: event.ID, Limit: limit, Offset: offset})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if rows == nil {
		rows = []db.ListEventInvitationsRow{}
	}
	w.Header().Set("Cache-Control", "private, no-store")
	platform.RespondJSON(w, 200, struct {
		Data []db.ListEventInvitationsRow `json:"data"`
	}{rows})
}
