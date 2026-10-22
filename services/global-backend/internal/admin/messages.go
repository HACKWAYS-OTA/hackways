package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/notifications"
	"github.com/hackways/global-backend/internal/platform"
)

type adminMessageBody struct {
	Subject     string                       `json:"subject"`
	Body        string                       `json:"body"`
	ActionLabel string                       `json:"action_label"`
	ActionURL   string                       `json:"action_url"`
	Recipients  []string                     `json:"recipients"`
	Design      notifications.CampaignDesign `json:"design"`
	Confirmed   bool                         `json:"confirmed"`
}

func decodeAdminMessage(w http.ResponseWriter, r *http.Request) (adminMessageBody, bool) {
	var body adminMessageBody
	if !campaignOriginAllowed(r) {
		platform.RespondError(w, 403, "INVALID_ORIGIN", "Use the admin console", nil)
		return body, false
	}
	if platform.DecodeJSON(r, &body) != nil {
		platform.RespondError(w, 400, "INVALID_MESSAGE", "Check the message fields", nil)
		return body, false
	}
	fields := campaignBody{Title: body.Subject, Subject: body.Subject, Body: body.Body, ActionLabel: body.ActionLabel, ActionURL: body.ActionURL}
	if !campaignFields(&fields) {
		platform.RespondError(w, 422, "INVALID_MESSAGE", "Add a subject, message and optional HTTPS button", nil)
		return body, false
	}
	body.Subject, body.Body, body.ActionLabel, body.ActionURL = fields.Subject, fields.Body, fields.ActionLabel, fields.ActionURL
	if len(body.Recipients) > 50 {
		platform.RespondError(w, 422, "RECIPIENT_LIMIT", "Choose up to fifty recipients", nil)
		return body, false
	}
	c := campaignConfiguration{Recipients: body.Recipients, Design: body.Design}
	if err := c.normalize(); err != nil {
		platform.RespondError(w, 422, "INVALID_MESSAGE", err.Error(), nil)
		return body, false
	}
	body.Recipients, body.Design = c.Recipients, c.Design
	return body, true
}

func (h *Handler) AdminMessagePreview(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeAdminMessage(w, r)
	if !ok {
		return
	}
	preview, err := notifications.RenderAdminMessagePreview(body.Subject, body.Body, body.ActionLabel, body.ActionURL, body.Design)
	if err != nil {
		platform.RespondError(w, 422, "PREVIEW_UNAVAILABLE", "Check the email content", nil)
		return
	}
	platform.RespondJSON(w, 200, struct {
		HTML   string `json:"html"`
		Sender string `json:"sender"`
	}{preview.BodyHTML, "notify@mail.hackways.in"})
}

func (h *Handler) AdminMessageSend(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeAdminMessage(w, r)
	if !ok {
		return
	}
	if !body.Confirmed || len(body.Recipients) == 0 {
		platform.RespondError(w, 422, "CONFIRMATION_REQUIRED", "Choose recipients, review the message and confirm sending", nil)
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) < 16 || len(key) > 128 || strings.ContainsAny(key, " \r\n\t") {
		platform.RespondError(w, 400, "IDEMPOTENCY_KEY_REQUIRED", "A message creation key is required", nil)
		return
	}
	if !notifications.EmailConfigured() {
		platform.RespondError(w, 503, "EMAIL_UNAVAILABLE", "Email delivery is not configured", nil)
		return
	}
	tx, q, user, _, ok := h.campaignMutation(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	canonical, _ := json.Marshal(body)
	hash := actionHash(canonical)
	message, err := q.AdminMessageCreate(r.Context(), db.AdminMessageCreateParams{ActorUserID: user.UserID, CreationKey: key, CreationHash: hash, Subject: body.Subject})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if message.CreationHash != hash {
		actionConflict(w)
		return
	}
	count, err := q.AdminMessageQueuedCount(r.Context(), message.ID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if count > 0 {
		platform.RespondJSON(w, 200, struct {
			Queued    int64 `json:"queued"`
			Reference int64 `json:"reference"`
		}{count, message.Reference})
		return
	}
	payload, _ := json.Marshal(struct {
		Subject     string                       `json:"promotion_subject"`
		Body        string                       `json:"promotion_body"`
		ActionLabel string                       `json:"action_label"`
		ActionURL   string                       `json:"action_url"`
		Design      notifications.CampaignDesign `json:"campaign_design"`
	}{body.Subject, body.Body, body.ActionLabel, body.ActionURL, body.Design})
	rows, err := q.QueueAdminCustomMessages(r.Context(), db.QueueAdminCustomMessagesParams{RecipientEmails: body.Recipients, Payload: payload, MessageID: message.ID})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if len(rows) != len(body.Recipients) {
		platform.RespondError(w, 422, "RECIPIENT_UNAVAILABLE", "Every service-message recipient must have an active verified Hackways account. No messages were queued.", nil)
		return
	}
	metadata, _ := json.Marshal(struct {
		Count  int    `json:"count"`
		Sender string `json:"sender"`
	}{len(rows), "notify@mail.hackways.in"})
	if err = q.AppendAdminAudit(r.Context(), db.AppendAdminAuditParams{ActorUserID: user.UserID, EntityID: message.ID, Action: "admin.message.queued", RequestID: platform.GetRequestID(r.Context()), Metadata: metadata}); err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 202, struct {
		Queued    int   `json:"queued"`
		Reference int64 `json:"reference"`
	}{len(rows), message.Reference})
}
