package workflows

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type InvitationRequest struct {
	EventID, OrganizerID, ActorID pgtype.UUID
	Emails                        []string
	Message, Key, RequestID       string
}
type InvitationResult struct {
	Queued  int64 `json:"queued"`
	Skipped int64 `json:"skipped"`
}

func (e *Engine) Invite(ctx context.Context, req InvitationRequest) (InvitationResult, bool, error) {
	result := InvitationResult{}
	tx, q, event, err := e.Begin(ctx, req.EventID)
	if err != nil {
		return result, false, err
	}
	defer tx.Rollback(ctx)
	if event.OrganizerID != req.OrganizerID {
		return result, false, NotFound()
	}
	canonical, err := json.Marshal(struct {
		Emails  []string
		Message string
	}{req.Emails, req.Message})
	if err != nil {
		return result, false, err
	}
	scope := digest("event-invitation:" + hex.EncodeToString(event.ID.Bytes[:]) + ":" + hex.EncodeToString(req.OrganizerID.Bytes[:]))
	key, hash := digest(req.Key), digest(string(canonical))
	stored, err := q.GetIdempotentResponse(ctx, db.GetIdempotentResponseParams{ScopeHash: scope, KeyHash: key})
	if err == nil {
		if stored.RequestHash != hash {
			return result, false, fail("IDEMPOTENCY_CONFLICT", "This request key was used with different invitations")
		}
		err = json.Unmarshal(stored.ResponseJson, &result)
		return result, true, err
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return result, false, err
	}
	if event.Status != db.EventStatusPublished || !event.StartsAt.Time.After(event.DatabaseNow.Time) || event.RegistrationClosesAt.Valid && !event.RegistrationClosesAt.Time.After(event.DatabaseNow.Time) {
		return result, false, fail("INVITATIONS_CLOSED", "Publish an upcoming event with registration open before inviting attendees")
	}
	count, err := q.CountEventInvitations(ctx, event.ID)
	if err != nil {
		return result, false, err
	}
	newCount, err := q.CountNewEventInvitationRecipients(ctx, db.CountNewEventInvitationRecipientsParams{EventID: event.ID, Emails: req.Emails})
	if err != nil {
		return result, false, err
	}
	if count+newCount > 1000 {
		return result, false, &Failure{429, "INVITATION_LIMIT", "This event has reached its invitation limit of 1000 recipients"}
	}
	info, err := q.ReadEventNotificationContext(ctx, event.ID)
	if err != nil {
		return result, false, err
	}
	origin := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/")
	if origin == "" {
		origin = "https://hackways.in"
	}
	eventURL := origin + "/events/" + url.PathEscape(info.Slug)
	payload, err := json.Marshal(struct {
		Title         string    `json:"event_title"`
		StartsAt      time.Time `json:"starts_at"`
		EndsAt        time.Time `json:"ends_at"`
		OrganizerName string    `json:"organizer_name"`
		SpaceName     string    `json:"space_name,omitempty"`
		Banner        string    `json:"event_banner_url,omitempty"`
		Location      string    `json:"location,omitempty"`
		EventURL      string    `json:"event_url"`
		ActionURL     string    `json:"action_url"`
		Message       string    `json:"custom_message,omitempty"`
	}{event.Title, event.StartsAt.Time, event.EndsAt.Time, info.OrganizerName, info.SpaceName.String, info.BannerUrl.String, info.LocationName.String, eventURL, eventURL + "/register?invitation=1", req.Message})
	if err != nil {
		return result, false, err
	}
	for _, email := range req.Emails {
		inserted, insertErr := q.EnqueueEventInvitation(ctx, db.EnqueueEventInvitationParams{EventID: event.ID, Email: email, Payload: payload})
		if insertErr != nil {
			return result, false, insertErr
		}
		result.Queued += inserted
	}
	result.Skipped = int64(len(req.Emails)) - result.Queued
	if result.Queued > 0 {
		if err = Audit(ctx, q, req.ActorID, event.ID, event.ID, "event.attendees_invited", req.RequestID); err != nil {
			return result, false, err
		}
	}
	response, err := json.Marshal(result)
	if err != nil {
		return result, false, err
	}
	if err = q.StoreIdempotentResponse(ctx, db.StoreIdempotentResponseParams{ScopeHash: scope, KeyHash: key, RequestHash: hash, ResponseJson: response}); err != nil {
		return result, false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return result, false, err
	}
	return result, false, nil
}
