package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/notifications"
	"github.com/hackways/global-backend/internal/platform"
)

type campaignConfiguration struct {
	Design      notifications.CampaignDesign `json:"design"`
	Audience    string                       `json:"audience"`
	Recipients  []string                     `json:"recipients"`
	ScheduledAt *time.Time                   `json:"scheduled_at"`
}

func (c *campaignConfiguration) normalize() error {
	if err := c.Design.Normalize(); err != nil {
		return err
	}
	if c.Audience == "" {
		c.Audience = "all"
	}
	if c.Audience != "all" && c.Audience != "account" && c.Audience != "import" {
		return fmt.Errorf("Choose all subscribers, verified account subscribers or imported contacts")
	}
	if len(c.Recipients) > 200 {
		return fmt.Errorf("Choose up to 200 specific recipients")
	}
	seen := make(map[string]bool, len(c.Recipients))
	emails := make([]string, 0, len(c.Recipients))
	for _, email := range c.Recipients {
		email = strings.ToLower(strings.TrimSpace(email))
		address, err := mail.ParseAddress(email)
		if err != nil || address.Address != email || len(email) > 254 {
			return fmt.Errorf("Enter valid recipient email addresses")
		}
		if !seen[email] {
			seen[email] = true
			emails = append(emails, email)
		}
	}
	c.Recipients = emails
	if c.ScheduledAt != nil && (c.ScheduledAt.Before(time.Now().Add(time.Minute)) || c.ScheduledAt.After(time.Now().Add(90*24*time.Hour))) {
		return fmt.Errorf("Schedule at least one minute ahead and within ninety days")
	}
	return nil
}

func readCampaignConfiguration(raw []byte) (campaignConfiguration, error) {
	var c campaignConfiguration
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &c); err != nil {
			return c, err
		}
	}
	// Old drafts acquire display defaults without silently changing their send time.
	if err := c.Design.Normalize(); err != nil {
		return c, err
	}
	if c.Audience == "" {
		c.Audience = "all"
	}
	if c.Recipients == nil {
		c.Recipients = []string{}
	}
	return c, nil
}

func (h *Handler) CampaignPreview(w http.ResponseWriter, r *http.Request) {
	var body campaignBody
	if !campaignOriginAllowed(r) {
		platform.RespondError(w, 403, "INVALID_ORIGIN", "Use the admin console", nil)
		return
	}
	if platform.DecodeJSON(r, &body) != nil || !campaignFields(&body) {
		platform.RespondError(w, 422, "INVALID_CAMPAIGN", "Complete the subject and message to preview", nil)
		return
	}
	if body.Configuration == nil {
		body.Configuration = &campaignConfiguration{}
	}
	if err := body.Configuration.normalize(); err != nil {
		platform.RespondError(w, 422, "INVALID_CONFIGURATION", err.Error(), nil)
		return
	}
	preview, err := notifications.RenderCampaignPreview(body.Subject, body.Body, body.ActionLabel, body.ActionURL, body.Configuration.Design)
	if err != nil {
		platform.RespondError(w, 422, "INVALID_PREVIEW", "Check the campaign content", nil)
		return
	}
	count, err := h.queries.CampaignAudienceCountSelected(r.Context(), db.CampaignAudienceCountSelectedParams{AudienceSource: body.Configuration.Audience, RecipientEmails: body.Configuration.Recipients})
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		HTML  string `json:"html"`
		Count int64  `json:"audience_count"`
	}{preview.BodyHTML, count})
}

func campaignOriginAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return origin == "https://admin.hackways.in" || os.Getenv("APP_ENV") == "development" && (origin == "http://localhost:3003" || origin == "http://localhost:3103" || origin == "http://127.0.0.1:3103")
}
