# Hackways email templates

The global backend owns transactional email rendering. Producers enqueue a
template ID and a JSON payload in `email_outbox`; the bounded ACS worker keeps
the existing operation identity, duplicate-send protection, retry schedule and
status reconciliation.

## Design rules

- HTML uses a simple editorial flow: wordmark, subject, readable paragraphs,
  optional text link, divider and a short footer. There are no cards, gradients,
  decorative illustrations or generated marketing copy.
- Every message has a useful plain-text body. Security and moderation notices
  deliberately send plain text only.
- Dynamic values are escaped before HTML rendering. Missing optional fields do
  not become fabricated values.
- Existing registration and event lifecycle messages retain their current-state
  checks in the outbox worker. New catalog entries are opt-in; adding a catalog
  entry does not enqueue mail by itself.

## Catalog

The catalog currently contains 40 IDs in
`services/global-backend/internal/notifications/templates.go`:

| Area | Template IDs |
| --- | --- |
| Account and security | `email_verification`, `password_reset_requested`, `password_changed`, `email_changed`, `sign_in_alert`, `session_revoked`, `account_welcome` |
| Organizer | `organizer_welcome`, `event_published`, `event_unpublished`, `event_capacity_reached`, `organizer_attendee_question`, `organizer_attendee_update`, `export_ready`, `team_invitation` |
| Registration | `rsvp_confirmation`, `rsvp_confirmed`, `rsvp_cancellation`, `rsvp_cancelled`, `rsvp_pending_approval`, `rsvp_waitlisted`, `rsvp_offered`, `rsvp_transferred`, `rsvp_expired`, `rsvp_rejected`, `registration_reminder_24h`, `registration_reminder_1h`, `waitlist_promoted`, `ticket_issued`, `ticket_revoked` |
| Event | `event_updated`, `event_cancelled`, `event_starting`, `event_ended`, `checkin_ready`, `checkin_confirmed` |
| Review and platform | `moderation_review`, `moderation_approved`, `moderation_rejected`, `platform_notice` |

The producers currently exercised in production are the existing verification,
registration and event update/cancellation IDs. The remaining entries are a
reviewed copy and layout catalog for future, explicitly wired workflows; they
are not reported as sent or active until a producer queues them.

