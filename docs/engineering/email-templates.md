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



## Confirmed event passes (2026-10-23)

`rsvp_confirmation` and `rsvp_confirmed` messages tied to a real registration
include a 384px signed QR PNG embedded using an ACS content ID, a **My pass**
link to the authenticated `/my-events/{registration_code}` page, and an
`event.ics` file attachment. The admission credential is never printed in the
email text or included in the link. Guest registrations retain the existing
verified-email account-claim flow for opening My pass.

The global worker reads one eligible pass through the generated
`ReadConfirmationPass` query. The registration must still be confirmed, its
ticket issued/scanned, the event published/not deleted and the queued recipient
must match the current account/guest email. Revoked/cancelled/transferred passes
are not attached. Core's event-locked scanner remains authoritative for admission;
a delivered QR does not override revocation, capacity or one-use check-in.

Both Go engines receive the **same existing** `TICKET_SIGNING_SEED`, injected
through protected runtime configuration. Do not rotate or fabricate another seed
when deploying the notification worker. It uses the existing Ed25519 claim format
and end-of-event plus 24-hour expiry. Rendering failures remain retryable outbox
failures; no provider attempt or delivery success is fabricated.

The calendar uses UTC timestamps, the existing public calendar UID, escaped
text, UTF-8-safe 75-octet folding and CRLF endings. It exports a published event,
not an attendee invitation. No email address or admission credential is included.
Generated QR/calendar bytes stay in the existing two-job worker; no persistent
files, media proxy or external QR service is added.

ACS REST moves to stable **2025-09-01**, whose attachment contract supports
`contentId`. The QR uses `cid:hackways-entry-qr`; ICS is `text/calendar` without
an inline content ID. Requests remain bounded below the provider's 10 MB limit,
with the existing persisted operation/reconciliation identity, five retries,
sender addresses and Reply-To behavior preserved. Engagement tracking stays off.

Apple Wallet `.pkpass` was explicitly deferred by the operator because no Apple
Pass Type ID signing credentials are available. No unsigned Wallet file, signing
helper or Wallet UI is shipped.

Sources: [ACS inline images](https://learn.microsoft.com/en-us/azure/communication-services/concepts/email/email-attachment-inline),
[stable attachment contract](https://github.com/Azure/azure-rest-api-specs/blob/main/specification/communication/data-plane/Email/stable/2025-09-01/CommunicationServicesEmail.json),
[ACS attachment types](https://learn.microsoft.com/en-us/azure/communication-services/concepts/email/email-attachment-allowed-mime-types).
