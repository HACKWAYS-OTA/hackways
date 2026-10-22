package events

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/registrationforms"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) sendRegistrationForm(w http.ResponseWriter, r *http.Request, id pgtype.UUID) {
	row, err := h.queries.ReadRegistrationForm(r.Context(), id)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	form := registrationforms.Form{Revision: row.RegistrationFormRevision, NameRequired: row.RegistrationNameRequired}
	if err = json.Unmarshal(row.RegistrationFormSettings, &form.Settings); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = json.Unmarshal(row.Questions, &form.Questions); err != nil {
		lifecycleError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	platform.RespondJSON(w, 200, struct {
		Data registrationforms.Form `json:"data"`
	}{form})
}
func (h *Handler) PublicRegistrationForm(w http.ResponseWriter, r *http.Request) {
	event, err := h.queries.GetPublishedEventBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		lifecycleError(w, err)
		return
	}
	h.sendRegistrationForm(w, r, event.ID)
}
func (h *Handler) OwnerRegistrationForm(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	h.sendRegistrationForm(w, r, event.ID)
}
func (h *Handler) SaveRegistrationForm(w http.ResponseWriter, r *http.Request) {
	id, ok := eventID(w, r)
	if !ok {
		return
	}
	var form registrationforms.Form
	if platform.DecodeJSON(r, &form) != nil || form.Revision < 1 {
		platform.RespondError(w, 422, "INVALID_FORM", "Provide the current form revision and valid questions", nil)
		return
	}
	form.Settings.Title = strings.TrimSpace(form.Settings.Title)
	form.Settings.Description = strings.TrimSpace(form.Settings.Description)
	form.Settings.SubmitLabel = strings.TrimSpace(form.Settings.SubmitLabel)
	form.Settings.SuccessMessage = strings.TrimSpace(form.Settings.SuccessMessage)
	if err := registrationforms.ValidateForm(form); err != nil {
		platform.RespondError(w, 422, "INVALID_FORM", err.Error(), nil)
		return
	}
	tx, q, event, err := workflows.New(h.pool).Begin(r.Context(), id)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	owner, _ := auth.OrganizerFromContext(r.Context())
	if event.OrganizerID != owner.ID {
		lifecycleError(w, workflows.NotFound())
		return
	}
	if form.Revision != event.RegistrationFormRevision {
		platform.RespondError(w, 409, "FORM_CHANGED", "These questions changed in another window. Reload before saving", nil)
		return
	}
	if event.Status == db.EventStatusCancelled || !event.StartsAt.Time.After(event.DatabaseNow.Time) {
		platform.RespondError(w, 409, "EDIT_CLOSED", "Started or cancelled events cannot change their registration form", nil)
		return
	}
	// Deactivate rather than delete. The event lock serializes this with new registrations.
	if err = q.DeactivateRegistrationQuestions(r.Context(), id); err != nil {
		lifecycleError(w, err)
		return
	}
	for i, item := range form.Questions {
		validation, marshalErr := json.Marshal(item.Validation)
		if marshalErr != nil {
			lifecycleError(w, marshalErr)
			return
		}
		var code string
		if item.PublicCode == "" {
			code, err = q.CreateRegistrationQuestion(r.Context(), db.CreateRegistrationQuestionParams{EventID: id, Label: item.Label, Description: item.Description, Kind: item.Kind, Required: item.Required, Position: int32(i), Options: item.Options, Placeholder: item.Placeholder, Validation: validation})
		} else {
			code, err = q.UpdateRegistrationQuestion(r.Context(), db.UpdateRegistrationQuestionParams{EventID: id, PublicCode: item.PublicCode, Label: item.Label, Description: item.Description, Kind: item.Kind, Required: item.Required, Position: int32(i), Options: item.Options, Placeholder: item.Placeholder, Validation: validation})
		}
		if errors.Is(err, pgx.ErrNoRows) {
			platform.RespondError(w, 422, "INVALID_QUESTION", "A question does not belong to this event. Reload before saving", nil)
			return
		}
		if err != nil {
			lifecycleError(w, err)
			return
		}
		form.Questions[i].PublicCode = code
	}
	settings, err := json.Marshal(form.Settings)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if err = q.UpdateRegistrationForm(r.Context(), db.UpdateRegistrationFormParams{ID: id, RegistrationNameRequired: form.NameRequired, RegistrationFormSettings: settings}); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = workflows.Audit(r.Context(), q, owner.UserID, id, id, "event.registration_form_updated", platform.GetRequestID(r.Context())); err != nil {
		lifecycleError(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		lifecycleError(w, err)
		return
	}
	form.Revision++
	if form.Questions == nil {
		form.Questions = []registrationforms.Question{}
	}
	w.Header().Set("Cache-Control", "no-store")
	platform.RespondJSON(w, 200, struct {
		Data registrationforms.Form `json:"data"`
	}{form})
}
func (h *Handler) GuestDetail(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	row, err := h.queries.ManagedGuestDetail(r.Context(), db.ManagedGuestDetailParams{EventID: event.ID, PublicCode: chi.URLParam(r, "code")})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	answers := make([]registrationforms.Snapshot, 0)
	submission, err := h.queries.LatestRegistrationAnswers(r.Context(), row.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		lifecycleError(w, err)
		return
	}
	if err == nil {
		if err = json.Unmarshal(submission.Answers, &answers); err != nil {
			lifecycleError(w, err)
			return
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	now := time.Now()
	editable := event.Status == db.EventStatusPublished && event.StartsAt.Time.After(now)
	canConfirm := editable && (!event.RegistrationOpensAt.Valid || !event.RegistrationOpensAt.Time.After(now)) && (!event.RegistrationClosesAt.Valid || event.RegistrationClosesAt.Time.After(now))
	options := workflows.OwnerStatusOptions(row.Status, row.TicketStatus.String == "scanned", editable, canConfirm)
	platform.RespondJSON(w, 200, struct {
		Data          db.ManagedGuestDetailRow     `json:"data"`
		Answers       []registrationforms.Snapshot `json:"answers"`
		StatusOptions []workflows.StatusOption     `json:"status_options"`
	}{row, answers, options})
}
