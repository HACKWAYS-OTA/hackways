package events

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

type invitationListItem struct {
	RecipientEmail     string                 `json:"recipient_email"`
	Status             db.NotificationStatus  `json:"status"`
	CreatedAt          pgtype.Timestamptz     `json:"created_at"`
	SentAt             pgtype.Timestamptz     `json:"sent_at"`
	RetryCount         int32                  `json:"retry_count"`
	RegistrationCode   *string                `json:"registration_code"`
	RegistrationStatus *db.RegistrationStatus `json:"registration_status"`
	CheckedInAt        pgtype.Timestamptz     `json:"checked_in_at"`
	GuestName          string                 `json:"guest_name"`
}

func (h *Handler) Invitations(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	limit, offset := pagination(r)
	search, valid := managementSearch(w, r)
	if !valid {
		return
	}
	state := r.URL.Query().Get("status")
	if state == "" {
		state = "all"
	}
	switch state {
	case "all", "awaiting", "responded", "sent", "failed", "queued":
	default:
		platform.RespondError(w, 400, "INVALID_STATUS", "Choose an invitation filter", nil)
		return
	}
	rows, err := h.queries.ListEventInvitations(r.Context(), db.ListEventInvitationsParams{EventID: event.ID, PageLimit: limit, PageOffset: offset, Search: search, StateFilter: state})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	items := make([]invitationListItem, 0, len(rows))
	for _, row := range rows {
		var code *string
		var status *db.RegistrationStatus
		if row.RegistrationCode != "" {
			value := row.RegistrationCode
			code = &value
		}
		if row.RegistrationStatus != "" {
			value := db.RegistrationStatus(row.RegistrationStatus)
			status = &value
		}
		items = append(items, invitationListItem{RecipientEmail: row.RecipientEmail, Status: row.Status, CreatedAt: row.CreatedAt, SentAt: row.SentAt, RetryCount: row.RetryCount, RegistrationCode: code, RegistrationStatus: status, CheckedInAt: row.CheckedInAt, GuestName: row.GuestName})
	}
	summary, err := h.queries.EventInvitationSummary(r.Context(), event.ID)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	platform.RespondJSON(w, 200, struct {
		Data    []invitationListItem         `json:"data"`
		Summary db.EventInvitationSummaryRow `json:"summary"`
	}{items, summary})
}

func (h *Handler) AccountInvitation(w http.ResponseWriter, r *http.Request) {
	event, err := h.queries.GetPublishedEventBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		lifecycleError(w, err)
		return
	}
	user, _ := auth.UserFromContext(r.Context())
	invited := false
	if user.EmailVerified {
		invited, err = h.queries.AccountEventInvitation(r.Context(), db.AccountEventInvitationParams{EventID: event.ID, Column2: strings.ToLower(user.Email)})
		if err != nil {
			lifecycleError(w, err)
			return
		}
	}
	var registrationCode *string
	registration, readErr := h.queries.GetAccountEventRSVP(r.Context(), db.GetAccountEventRSVPParams{EventID: event.ID, UserID: user.UserID})
	if readErr != nil && !errors.Is(readErr, pgx.ErrNoRows) {
		lifecycleError(w, readErr)
		return
	}
	if readErr == nil && registration.Status != db.RegistrationStatusCancelled && registration.Status != db.RegistrationStatusExpired {
		registrationCode = &registration.PublicCode
	}
	w.Header().Set("Cache-Control", "private, no-store")
	platform.RespondJSON(w, 200, struct {
		Invited          bool    `json:"invited"`
		RegistrationCode *string `json:"registration_code"`
	}{invited, registrationCode})
}
