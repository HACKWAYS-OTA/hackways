package events

import (
	"context"
	"github.com/hackways/core-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
	"strings"
)

type eventContent struct {
	FAQText    *string `json:"faq_text"`
	ContactURL *string `json:"contact_url"`
	Rewards    *string `json:"rewards"`
}

func (c *eventContent) valid() bool {
	if c.FAQText != nil {
		*c.FAQText = strings.TrimSpace(*c.FAQText)
		if len([]rune(*c.FAQText)) > 10000 {
			return false
		}
	}
	if c.Rewards != nil {
		*c.Rewards = strings.TrimSpace(*c.Rewards)
		if len([]rune(*c.Rewards)) > 10000 {
			return false
		}
	}
	if c.ContactURL != nil {
		*c.ContactURL = strings.TrimSpace(*c.ContactURL)
		if len(*c.ContactURL) > 2048 || (*c.ContactURL != "" && !validHTTPSURL(*c.ContactURL)) {
			return false
		}
	}
	return true
}
func contentText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}
func (c eventContent) apply(ctx context.Context, q *db.Queries, event db.Events) (db.Events, error) {
	if c.FAQText == nil && c.ContactURL == nil && c.Rewards == nil {
		return event, nil
	}
	return q.SetEventContent(ctx, db.SetEventContentParams{ID: event.ID, FaqText: contentText(c.FAQText), ContactUrl: contentText(c.ContactURL), Rewards: contentText(c.Rewards)})
}
