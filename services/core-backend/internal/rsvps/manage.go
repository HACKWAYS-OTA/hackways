package rsvps

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
	"strings"
)

func (h *Handler) EditNotes(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "id")); err != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid registration ID", nil)
		return
	}
	var body struct {
		Notes string `json:"notes"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil || len(body.Notes) > 2000 {
		platform.RespondError(w, 422, "INVALID_NOTES", "Notes must be at most 2000 characters", nil)
		return
	}
	existing, err := h.queries.GetRSVPByIDAndUser(r.Context(), db.GetRSVPByIDAndUserParams{ID: id, UserID: user.UserID})
	if err != nil {
		respondFailure(w, err)
		return
	}
	tx, q, event, err := h.engine.Begin(r.Context(), existing.EventID)
	if err != nil {
		respondFailure(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	existing, err = q.GetRSVPByIDAndUser(r.Context(), db.GetRSVPByIDAndUserParams{ID: id, UserID: user.UserID})
	if err != nil {
		respondFailure(w, err)
		return
	}
	if !event.StartsAt.Time.After(event.DatabaseNow.Time) || event.Status == db.EventStatusCancelled || existing.Status != db.RegistrationStatusConfirmed && existing.Status != db.RegistrationStatusPendingApproval && existing.Status != db.RegistrationStatusWaitlisted && existing.Status != db.RegistrationStatusOffered {
		platform.RespondError(w, 409, "EDIT_CLOSED", "This registration can no longer be edited", nil)
		return
	}
	updated, err := q.EditRegistrationNotes(r.Context(), db.EditRegistrationNotesParams{ID: id, Notes: pgtype.Text{String: strings.TrimSpace(body.Notes), Valid: strings.TrimSpace(body.Notes) != ""}})
	if err != nil {
		respondFailure(w, err)
		return
	}
	if err = workflows.Audit(r.Context(), q, user.UserID, event.ID, id, "registration.notes_updated", platform.GetRequestID(r.Context())); err != nil {
		respondFailure(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		respondFailure(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Rsvps `json:"data"`
	}{updated})
}
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	if !user.EmailVerified {
		platform.RespondError(w, 403, "EMAIL_VERIFICATION_REQUIRED", "Verify your email before transferring a ticket", nil)
		return
	}
	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "id")); err != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid registration ID", nil)
		return
	}
	var body struct {
		Email string `json:"recipient_email"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.RespondError(w, 400, "INVALID_PAYLOAD", "Provide the recipient email", nil)
		return
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	address, parseErr := mail.ParseAddress(body.Email)
	if parseErr != nil || address.Address != body.Email || len(body.Email) > 320 {
		platform.RespondError(w, 422, "INVALID_EMAIL", "Provide a valid recipient email", nil)
		return
	}
	existing, err := h.queries.GetRSVPByIDAndUser(r.Context(), db.GetRSVPByIDAndUserParams{ID: id, UserID: user.UserID})
	if err != nil {
		respondFailure(w, err)
		return
	}
	tx, q, event, err := h.engine.Begin(r.Context(), existing.EventID)
	if err != nil {
		respondFailure(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	existing, err = q.GetRSVPByIDAndUser(r.Context(), db.GetRSVPByIDAndUserParams{ID: id, UserID: user.UserID})
	if err != nil {
		respondFailure(w, err)
		return
	}
	if existing.Status == db.RegistrationStatusTransferred {
		platform.RespondJSON(w, 200, struct {
			Data db.Rsvps `json:"data"`
		}{existing})
		return
	}
	if existing.Status != db.RegistrationStatusConfirmed || existing.TransferCount >= 1 || !event.StartsAt.Time.After(event.DatabaseNow.Time) || event.Status != db.EventStatusPublished || event.RegistrationMode != "instant" {
		platform.RespondError(w, 409, "TRANSFER_UNAVAILABLE", "Only untransferred, confirmed instant-registration tickets can be transferred before the event", nil)
		return
	}
	ticket, ticketErr := q.GetTicketByRSVP(r.Context(), id)
	if ticketErr != nil && !errors.Is(ticketErr, pgx.ErrNoRows) {
		respondFailure(w, ticketErr)
		return
	}
	if ticketErr == nil && ticket.Status == "scanned" {
		platform.RespondError(w, 409, "ALREADY_CHECKED_IN", "A checked-in ticket cannot be transferred", nil)
		return
	}
	recipient, err := q.FindVerifiedTransferRecipient(r.Context(), body.Email)
	if errors.Is(err, pgx.ErrNoRows) || err == nil && recipient.ID == user.UserID {
		platform.RespondError(w, 422, "RECIPIENT_UNAVAILABLE", "Use a different recipient with a verified Hackways account", nil)
		return
	}
	if err != nil {
		respondFailure(w, err)
		return
	}
	if _, readErr := q.GetAccountEventRSVP(r.Context(), db.GetAccountEventRSVPParams{EventID: event.ID, UserID: recipient.ID}); readErr == nil {
		platform.RespondError(w, 409, "RECIPIENT_REGISTERED", "The recipient already has a registration for this event", nil)
		return
	} else if !errors.Is(readErr, pgx.ErrNoRows) {
		respondFailure(w, readErr)
		return
	}
	source, err := q.SetWorkflowRSVPState(r.Context(), db.SetWorkflowRSVPStateParams{ID: id, Status: db.RegistrationStatusTransferred})
	if err != nil {
		respondFailure(w, err)
		return
	}
	if err = q.RevokeWorkflowTicket(r.Context(), id); err != nil {
		respondFailure(w, err)
		return
	}
	destination, err := q.CreateTransferredRegistration(r.Context(), db.CreateTransferredRegistrationParams{EventID: event.ID, UserID: recipient.ID, GuestName: pgtype.Text{String: recipient.DisplayName, Valid: true}, GuestEmail: pgtype.Text{String: recipient.Email, Valid: true}, TransferredFrom: id})
	if err != nil {
		respondFailure(w, err)
		return
	}
	if _, err = q.EnsureTicket(r.Context(), destination.ID); err != nil {
		respondFailure(w, err)
		return
	}
	for _, change := range []struct {
		Registration     db.Rsvps
		Action, Template string
	}{{source, "registration.transferred", "rsvp_transferred"}, {destination, "registration.transfer_received", "rsvp_confirmed"}} {
		if err = workflows.Audit(r.Context(), q, user.UserID, event.ID, change.Registration.ID, change.Action, platform.GetRequestID(r.Context())); err != nil {
			respondFailure(w, err)
			return
		}
		if err = workflows.Notify(r.Context(), q, event, change.Registration, change.Template); err != nil {
			respondFailure(w, err)
			return
		}
	}
	if err = tx.Commit(r.Context()); err != nil {
		respondFailure(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Rsvps `json:"data"`
	}{source})
}
