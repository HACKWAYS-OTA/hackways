package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5"
)

type ProgramMailDetails struct {
	Name, Title, Slug, City, Proposal, ProfileURL, Decision string
	SubmittedAt                                             time.Time
}

func isProgramMail(template string) bool {
	return template == "program_application_received" || template == "program_results_released" || template == "program_communication"
}

func (w *OutboxWorker) prepareProgramMail(ctx context.Context, entry db.ClaimEmailOutboxRow) (db.ClaimEmailOutboxRow, error) {
	if !isProgramMail(entry.TemplateType) {
		return entry, nil
	}
	if entry.TemplateType == "program_communication" {
		active, err := w.queries.AdminMessageRecipientActive(ctx, entry.ID)
		if err != nil {
			return entry, err
		}
		if !active {
			return entry, errSuperseded
		}
	}
	actual, err := w.queries.ProgramMailContext(ctx, entry.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return entry, errSuperseded
	}
	if err != nil {
		return entry, err
	}
	var payload templatePayload
	if err = json.Unmarshal(entry.PayloadJson, &payload); err != nil {
		return entry, err
	}
	applyProgramDetails(&payload, ProgramMailDetails{actual.FullName, actual.Title, actual.Slug, actual.City, actual.Proposal, actual.ProfileUrl, actual.PublishedDecision.String, actual.SubmittedAt.Time})
	entry.PayloadJson, err = json.Marshal(payload)
	return entry, err
}

func applyProgramDetails(payload *templatePayload, details ProgramMailDetails) {
	payload.RecipientName, payload.ProgramTitle, payload.ProgramSlug = details.Name, details.Title, details.Slug
	payload.ProgramDecision, payload.ProgramCity, payload.ProgramProposal, payload.ProgramProfileURL = details.Decision, details.City, details.Proposal, details.ProfileURL
	payload.ProgramSubmittedAt = details.SubmittedAt
	payload.ActionURL = programApplicationURL(details.Slug)
	if details.Slug == "ambassadors" {
		payload.EventBannerURL = "https://hackways.in/programs/ambassadors.jpg"
	}
}

func programMerge(text string, payload templatePayload) string {
	decision := map[string]string{"accepted": "accepted", "rejected": "not selected", "waitlisted": "waitlisted"}[payload.ProgramDecision]
	if decision == "" {
		decision = "under review"
	}
	return strings.NewReplacer("{{name}}", payload.RecipientName, "{{program}}", payload.ProgramTitle, "{{decision}}", decision, "{{application_url}}", programApplicationURL(payload.ProgramSlug)).Replace(text)
}

func RenderProgramCommunicationPreview(details ProgramMailDetails, subject, body string) (EmailMessage, error) {
	payload := templatePayload{PromotionSubject: subject, PromotionBody: body}
	applyProgramDetails(&payload, details)
	data, err := json.Marshal(payload)
	if err != nil {
		return EmailMessage{}, err
	}
	return render(db.ClaimEmailOutboxRow{TemplateType: "program_communication", PayloadJson: data})
}
