package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/notifications"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const campaignAudienceLimit = 1000

type campaignBody struct {
	Configuration *campaignConfiguration `json:"configuration"`
	Title         string                 `json:"title"`
	Subject       string                 `json:"subject"`
	Body          string                 `json:"body"`
	ActionLabel   string                 `json:"action_label"`
	ActionURL     string                 `json:"action_url"`
	Revision      int64                  `json:"revision"`
	AudienceCount int64                  `json:"audience_count"`
	Confirmed     bool                   `json:"confirmed"`
}
type campaignView struct {
	Configuration json.RawMessage `json:"configuration"`
	Reference     int64           `json:"reference"`
	Title         string          `json:"title"`
	Subject       string          `json:"subject"`
	Body          string          `json:"body"`
	ActionLabel   string          `json:"action_label"`
	ActionURL     string          `json:"action_url"`
	Status        string          `json:"status"`
	Revision      int64           `json:"revision"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func viewCampaign(c db.EmailCampaigns) campaignView {
	return campaignView{json.RawMessage(c.Configuration), c.Reference, c.Title, c.Subject, c.Body, c.ActionLabel, c.ActionUrl, c.Status, c.Revision, c.UpdatedAt.Time}
}
func (h *Handler) mountCampaigns(r chi.Router) {
	r.Post("/messages/preview", h.AdminMessagePreview)
	r.Post("/messages/send", h.AdminMessageSend)
	r.Get("/campaigns", h.CampaignList)
	r.Post("/campaigns", h.CampaignCreate)
	r.Post("/campaigns/preview", h.CampaignPreview)
	r.Get("/campaigns/{id}", h.CampaignDetail)
	r.Post("/campaigns/{id}/{action}", h.CampaignChange)
	r.Get("/campaign-contacts", h.PromotionContacts)
	r.Post("/campaign-contacts/import", h.ImportPromotionContacts)
}
func campaignReference(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		platform.RespondError(w, 400, "INVALID_REFERENCE", "Use a valid campaign reference", nil)
		return 0, false
	}
	return id, true
}
func campaignOffset(r *http.Request) int32 {
	n, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if err != nil || n < 0 {
		return 0
	}
	return int32(n)
}
func (h *Handler) CampaignList(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(query) > 200 {
		platform.RespondError(w, 400, "INVALID_SEARCH", "Search is too long", nil)
		return
	}
	rows, err := h.queries.CampaignList(r.Context(), db.CampaignListParams{Q: query, PageOffset: campaignOffset(r)})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.CampaignListRow{}
	}
	next := len(rows) > 20
	if next {
		rows = rows[:20]
	}
	platform.RespondJSON(w, 200, struct {
		Data    []db.CampaignListRow `json:"data"`
		HasNext bool                 `json:"has_next"`
	}{rows, next})
}
func (h *Handler) CampaignDetail(w http.ResponseWriter, r *http.Request) {
	ref, ok := campaignReference(w, r)
	if !ok {
		return
	}
	c, err := h.queries.CampaignRead(r.Context(), ref)
	if err != nil {
		detailError(w, err)
		return
	}
	stats, err := h.queries.CampaignStats(r.Context(), c.ID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	configuration, err := readCampaignConfiguration(c.Configuration)
	if err != nil {
		platform.RespondError(w, 500, "CAMPAIGN_UNAVAILABLE", "Saved campaign settings could not be read", nil)
		return
	}
	count, err := h.queries.CampaignAudienceCountSelected(r.Context(), db.CampaignAudienceCountSelectedParams{AudienceSource: configuration.Audience, RecipientEmails: configuration.Recipients})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	audience, err := h.queries.CampaignAudiencePreviewSelected(r.Context(), db.CampaignAudiencePreviewSelectedParams{AudienceSource: configuration.Audience, RecipientEmails: configuration.Recipients})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if audience == nil {
		audience = []db.CampaignAudiencePreviewSelectedRow{}
	}
	deliveries, err := h.queries.CampaignDeliveryList(r.Context(), db.CampaignDeliveryListParams{CampaignID: c.ID, PageOffset: campaignOffset(r)})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if deliveries == nil {
		deliveries = []db.CampaignDeliveryListRow{}
	}
	next := len(deliveries) > 20
	if next {
		deliveries = deliveries[:20]
	}
	preview, err := notifications.RenderCampaignPreview(c.Subject, c.Body, c.ActionLabel, c.ActionUrl, configuration.Design)
	if err != nil {
		platform.RespondError(w, 500, "PREVIEW_UNAVAILABLE", "The email preview could not be rendered", nil)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data          campaignView                            `json:"data"`
		Stats         db.CampaignStatsRow                     `json:"stats"`
		AudienceCount int64                                   `json:"audience_count"`
		Audience      []db.CampaignAudiencePreviewSelectedRow `json:"audience"`
		Deliveries    []db.CampaignDeliveryListRow            `json:"deliveries"`
		HasNext       bool                                    `json:"has_next"`
		PreviewHTML   string                                  `json:"preview_html"`
		Sender        string                                  `json:"sender"`
	}{viewCampaign(c), stats, count, audience, deliveries, next, preview.BodyHTML, "updates@mail.hackways.in"})
}
func (h *Handler) campaignMutation(w http.ResponseWriter, r *http.Request) (pgx.Tx, *db.Queries, *auth.SessionUser, string, bool) {
	origin := r.Header.Get("Origin")
	allowed := origin == "https://admin.hackways.in"
	if os.Getenv("APP_ENV") == "development" {
		allowed = allowed || origin == "http://localhost:3003" || origin == "http://localhost:3103" || origin == "http://127.0.0.1:3103"
	}
	reason, err := url.PathUnescape(r.Header.Get("X-Admin-Reason"))
	reason = strings.TrimSpace(reason)
	if err == nil && reason == "" {
		reason = "Campaign " + chi.URLParam(r, "action")
		if chi.URLParam(r, "action") == "" {
			reason = "Campaign draft created"
		}
		if strings.HasSuffix(r.URL.Path, "/campaign-contacts/import") {
			reason = "Opted-in contacts imported"
		}
	}
	if !allowed {
		platform.RespondError(w, 403, "INVALID_ORIGIN", "Use the admin console to make changes", nil)
		return nil, nil, nil, "", false
	}
	if err != nil || !utf8.ValidString(reason) || utf8.RuneCountInString(reason) < 3 || utf8.RuneCountInString(reason) > 1000 || strings.ContainsAny(reason, "\x00\r\n") {
		platform.RespondError(w, 400, "REASON_REQUIRED", "Enter a reason between 3 and 1,000 characters", nil)
		return nil, nil, nil, "", false
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
		return nil, nil, nil, "", false
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return nil, nil, nil, "", false
	}
	q := h.queries.WithTx(tx)
	if err = q.SetStorageWriteTimeout(r.Context()); err == nil {
		_, err = q.LockAuthenticationUser(r.Context(), user.UserID)
	}
	var granted bool
	if err == nil {
		granted, err = q.IsPlatformAdminSession(r.Context(), db.IsPlatformAdminSessionParams{SessionID: user.SessionID, UserID: user.UserID})
	}
	if err != nil {
		tx.Rollback(r.Context())
		platform.RespondDatabaseError(w, err)
		return nil, nil, nil, "", false
	}
	if !granted {
		tx.Rollback(r.Context())
		platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "Admin access is required", nil)
		return nil, nil, nil, "", false
	}
	return tx, q, user, reason, true
}
func campaignFields(body *campaignBody) bool {
	body.Title = strings.TrimSpace(body.Title)
	body.Subject = strings.TrimSpace(body.Subject)
	body.Body = strings.TrimSpace(body.Body)
	body.ActionLabel = strings.TrimSpace(body.ActionLabel)
	body.ActionURL = strings.TrimSpace(body.ActionURL)
	if !utf8.ValidString(body.Title+body.Subject+body.Body+body.ActionLabel) || strings.ContainsRune(body.Title+body.Subject+body.Body+body.ActionLabel, 0) || utf8.RuneCountInString(body.Title) < 1 || utf8.RuneCountInString(body.Title) > 150 || utf8.RuneCountInString(body.Subject) < 1 || utf8.RuneCountInString(body.Subject) > 150 || utf8.RuneCountInString(body.Body) < 1 || utf8.RuneCountInString(body.Body) > 20000 || utf8.RuneCountInString(body.ActionLabel) > 80 || len(body.ActionURL) > 2000 || strings.ContainsAny(body.Subject, "\r\n\x00") {
		return false
	}
	if body.ActionURL != "" {
		u, err := url.Parse(body.ActionURL)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || body.ActionLabel == "" {
			return false
		}
	} else if body.ActionLabel != "" {
		return false
	}
	return true
}
func campaignAudit(r *http.Request, q *db.Queries, user *auth.SessionUser, id pgtype.UUID, action, reason string, count int64) error {
	metadata, _ := json.Marshal(struct {
		Reason string `json:"reason"`
		Count  int64  `json:"count"`
	}{reason, count})
	return q.AppendAdminAudit(r.Context(), db.AppendAdminAuditParams{ActorUserID: user.UserID, EntityID: id, Action: "admin.campaign." + action, RequestID: platform.GetRequestID(r.Context()), Metadata: metadata})
}
func (h *Handler) CampaignCreate(w http.ResponseWriter, r *http.Request) {
	var body campaignBody
	if platform.DecodeJSON(r, &body) != nil || !campaignFields(&body) {
		platform.RespondError(w, 422, "INVALID_CAMPAIGN", "Enter a title, subject, message and optional HTTPS action link", nil)
		return
	}
	if body.Configuration == nil {
		body.Configuration = &campaignConfiguration{}
	}
	if err := body.Configuration.normalize(); err != nil {
		platform.RespondError(w, 422, "INVALID_CONFIGURATION", err.Error(), nil)
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) < 16 || len(key) > 128 || strings.ContainsAny(key, " \r\n\t") {
		platform.RespondError(w, 400, "IDEMPOTENCY_KEY_REQUIRED", "A campaign creation key is required", nil)
		return
	}
	tx, q, user, reason, ok := h.campaignMutation(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	canonical, _ := json.Marshal(body)
	hash := actionHash(canonical)
	configurationJSON, _ := json.Marshal(body.Configuration)
	c, err := q.CampaignCreate(r.Context(), db.CampaignCreateParams{Configuration: configurationJSON, Title: body.Title, Subject: body.Subject, Body: body.Body, ActionLabel: body.ActionLabel, ActionUrl: body.ActionURL, CreatedBy: user.UserID, CreationKey: key, CreationHash: hash})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if c.CreationHash != hash {
		actionConflict(w)
		return
	}
	if err = campaignAudit(r, q, user, c.ID, "draft", reason, 0); err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 201, struct {
		Data campaignView `json:"data"`
	}{viewCampaign(c)})
}
func (h *Handler) CampaignChange(w http.ResponseWriter, r *http.Request) {
	ref, ok := campaignReference(w, r)
	if !ok {
		return
	}
	action := chi.URLParam(r, "action")
	if action != "edit" && action != "test" && action != "send" && action != "cancel" {
		platform.RespondError(w, 404, "NOT_FOUND", "Action not found", nil)
		return
	}
	var body campaignBody
	if platform.DecodeJSON(r, &body) != nil || body.Revision < 1 {
		platform.RespondError(w, 400, "INVALID_REQUEST", "A current campaign revision is required", nil)
		return
	}
	if action == "edit" && !campaignFields(&body) {
		platform.RespondError(w, 422, "INVALID_CAMPAIGN", "Enter valid campaign content and optional HTTPS link", nil)
		return
	}
	if action == "edit" && body.Configuration != nil {
		if err := body.Configuration.normalize(); err != nil {
			platform.RespondError(w, 422, "INVALID_CONFIGURATION", err.Error(), nil)
			return
		}
	}
	tx, q, user, reason, ok := h.campaignMutation(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	c, err := q.CampaignLock(r.Context(), ref)
	if err != nil {
		detailError(w, err)
		return
	}
	if c.Revision != body.Revision {
		actionConflict(w)
		return
	}
	if action == "send" && c.Status == "queued" {
		platform.RespondJSON(w, 200, struct {
			Message string `json:"message"`
		}{"Campaign is already queued"})
		return
	}
	if action == "cancel" && c.Status == "cancelled" {
		platform.RespondJSON(w, 200, struct {
			Message string `json:"message"`
		}{"Campaign is already cancelled"})
		return
	}
	if c.Status == "cancelled" || (action != "cancel" && c.Status != "draft") {
		actionConflict(w)
		return
	}
	configuration, err := readCampaignConfiguration(c.Configuration)
	if err != nil {
		platform.RespondError(w, 500, "CAMPAIGN_UNAVAILABLE", "Saved settings could not be read", nil)
		return
	}
	if body.Configuration == nil {
		body.Configuration = &configuration
	}
	var count int64
	switch action {
	case "edit":
		configurationJSON, _ := json.Marshal(body.Configuration)
		_, err = q.CampaignEdit(r.Context(), db.CampaignEditParams{Configuration: configurationJSON, ID: c.ID, Title: body.Title, Subject: body.Subject, Body: body.Body, ActionLabel: body.ActionLabel, ActionUrl: body.ActionURL, Revision: c.Revision})
	case "cancel":
		err = q.CampaignCancel(r.Context(), c.ID)
	case "test", "send":
		if !notifications.EmailConfigured() || os.Getenv("ACS_PROMOTIONAL_SENDER_ADDRESS") == "" {
			platform.RespondError(w, 503, "EMAIL_UNAVAILABLE", "Promotional email is not configured", nil)
			return
		}
		var contact pgtype.UUID
		if action == "test" {
			contact, err = q.PromotionTestContact(r.Context(), db.PromotionTestContactParams{Email: strings.ToLower(user.Email), DisplayName: user.DisplayName, UserID: user.UserID})
		} else {
			if !body.Confirmed {
				platform.RespondError(w, 422, "SEND_CONFIRMATION_REQUIRED", "Review the audience and confirm sending", nil)
				return
			}
			if configuration.ScheduledAt != nil && configuration.ScheduledAt.Before(time.Now()) {
				platform.RespondError(w, 422, "SCHEDULE_PASSED", "Choose a new send time or send now", nil)
				return
			}
			count, err = q.CampaignAudienceCountSelected(r.Context(), db.CampaignAudienceCountSelectedParams{AudienceSource: configuration.Audience, RecipientEmails: configuration.Recipients})
			if err != nil {
				break
			}
			if count < 1 || count > campaignAudienceLimit || count != body.AudienceCount {
				platform.RespondError(w, 409, "AUDIENCE_CHANGED", "Review the current opted-in audience. Campaigns require 1 to 1,000 eligible recipients.", nil)
				return
			}
		}
		if err != nil {
			break
		}
		payload, _ := json.Marshal(struct {
			Design      notifications.CampaignDesign `json:"campaign_design"`
			Subject     string                       `json:"promotion_subject"`
			Body        string                       `json:"promotion_body"`
			ActionLabel string                       `json:"action_label"`
			ActionURL   string                       `json:"action_url"`
		}{configuration.Design, c.Subject, c.Body, c.ActionLabel, c.ActionUrl})
		var queued int64
		sendAt := pgtype.Timestamptz{}
		if action == "send" && configuration.ScheduledAt != nil {
			sendAt = pgtype.Timestamptz{Time: *configuration.ScheduledAt, Valid: true}
		}
		queued, err = q.QueueCampaignMessages(r.Context(), db.QueueCampaignMessagesParams{AudienceSource: configuration.Audience, RecipientEmails: configuration.Recipients, SendAt: sendAt, TestOnly: action == "test", TestContactID: contact, CampaignID: c.ID, Revision: c.Revision, MaxRecipients: campaignAudienceLimit + 1, BasePayload: payload})
		if err != nil {
			break
		}
		if action == "send" && queued != count {
			platform.RespondError(w, 409, "AUDIENCE_CHANGED", "The audience changed. Review it and try again.", nil)
			return
		}
		count = queued
		if action == "send" {
			err = q.CampaignSetQueued(r.Context(), c.ID)
		}
	}
	if err == nil {
		err = campaignAudit(r, q, user, c.ID, action, reason, count)
	}
	if err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			actionConflict(w)
		} else {
			platform.RespondDatabaseError(w, err)
		}
		return
	}
	platform.RespondJSON(w, 200, struct {
		Message string `json:"message"`
		Queued  int64  `json:"queued"`
	}{"Campaign action saved. Queued email is not yet confirmed delivered.", count})
}
