package notifications

// EmailTemplate describes a supported outbox template. The catalog is kept in
// the global backend so producers can opt into a consistent voice and layout
// without adding presentation logic to the Go services that enqueue messages.
type EmailTemplate struct {
	ID       string
	Subject  string
	Category string
	TextOnly bool
	Copy     string
}

var emailTemplateCatalog = []EmailTemplate{
	{ID: "email_verification", Subject: "Verify your email", Category: "account", TextOnly: true, Copy: "Use the verification code below to confirm your email address."},
	{ID: "password_reset_requested", Subject: "Reset your password", Category: "account", TextOnly: true, Copy: "A password reset was requested for your Hackways account."},
	{ID: "password_changed", Subject: "Your password was changed", Category: "security", TextOnly: true, Copy: "Your Hackways password was changed."},
	{ID: "email_changed", Subject: "Your email address was changed", Category: "security", TextOnly: true, Copy: "The email address on your Hackways account was changed."},
	{ID: "sign_in_alert", Subject: "New sign-in to your account", Category: "security", TextOnly: true, Copy: "A new sign-in to your Hackways account was recorded."},
	{ID: "session_revoked", Subject: "A session was signed out", Category: "security", TextOnly: true, Copy: "A session for your Hackways account was signed out."},
	{ID: "account_welcome", Subject: "Welcome to Hackways", Category: "account", Copy: "Your Hackways account is ready. You can now discover events and keep your registrations in one place."},
	{ID: "organizer_welcome", Subject: "Welcome to Hackways for organizers", Category: "organizer", Copy: "Your organizer account is ready. Create an event when you are ready to invite people."},
	{ID: "rsvp_confirmation", Subject: "Your registration is confirmed", Category: "registration", Copy: "Your registration is confirmed."},
	{ID: "rsvp_confirmed", Subject: "Your registration is confirmed", Category: "registration", Copy: "Your registration is confirmed."},
	{ID: "rsvp_cancellation", Subject: "Your registration was cancelled", Category: "registration", Copy: "Your registration was cancelled."},
	{ID: "rsvp_cancelled", Subject: "Your registration was cancelled", Category: "registration", Copy: "Your registration was cancelled."},
	{ID: "rsvp_pending_approval", Subject: "Your registration is awaiting review", Category: "registration", Copy: "Your application is awaiting organizer review. A place has not been reserved yet."},
	{ID: "rsvp_waitlisted", Subject: "You joined the waitlist", Category: "registration", Copy: "You joined the waitlist. This is not a confirmed place."},
	{ID: "rsvp_offered", Subject: "A place is available", Category: "registration", Copy: "A place is temporarily available for you. Sign in before the offer expires."},
	{ID: "rsvp_transferred", Subject: "Your ticket was transferred", Category: "registration", Copy: "Your ticket was transferred to the verified recipient. Your previous ticket is no longer valid."},
	{ID: "rsvp_expired", Subject: "Your registration application expired", Category: "registration", Copy: "The review period for your application ended without approval. No place was confirmed."},
	{ID: "rsvp_rejected", Subject: "Registration application update", Category: "registration", Copy: "The organizer did not approve your application. Contact the organizer if you need more information."},
	{ID: "registration_reminder_24h", Subject: "Your event is tomorrow", Category: "registration", Copy: "This is a reminder that your confirmed event is tomorrow."},
	{ID: "registration_reminder_1h", Subject: "Your event starts soon", Category: "registration", Copy: "Your confirmed event starts soon. Keep this message nearby for the event details."},
	{ID: "event_updated", Subject: "Event details changed", Category: "event", Copy: "The organizer updated the event details. Review the current event page before attending."},
	{ID: "event_cancelled", Subject: "Event cancelled", Category: "event", Copy: "The event was cancelled. Any registration for it is no longer valid."},
	{ID: "event_published", Subject: "Your event is published", Category: "organizer", Copy: "Your event is now published and ready to share."},
	{ID: "event_unpublished", Subject: "Your event is no longer public", Category: "organizer", Copy: "Your event was taken out of public view. Existing registration records are unchanged unless you update them."},
	{ID: "event_starting", Subject: "Your event is starting", Category: "event", Copy: "Your event is starting. Check the event details for the latest location or joining information."},
	{ID: "event_ended", Subject: "Thanks for joining the event", Category: "event", Copy: "The event has ended. Thank you for taking part."},
	{ID: "event_capacity_reached", Subject: "Your event has reached capacity", Category: "organizer", Copy: "Your event has reached its current capacity. New applications will follow the event's registration policy."},
	{ID: "waitlist_promoted", Subject: "You moved off the waitlist", Category: "registration", Copy: "A place is now available for you. Sign in to review the registration details."},
	{ID: "ticket_issued", Subject: "Your ticket is ready", Category: "registration", Copy: "Your ticket is ready for the event. Keep this message for your records."},
	{ID: "ticket_revoked", Subject: "Your ticket was revoked", Category: "registration", Copy: "Your ticket is no longer valid. Contact the organizer if you believe this was unexpected."},
	{ID: "checkin_ready", Subject: "Check-in is ready", Category: "event", Copy: "Check-in is ready for your event. Follow the organizer's instructions when you arrive."},
	{ID: "checkin_confirmed", Subject: "You are checked in", Category: "event", Copy: "Your check-in was recorded successfully."},
	{ID: "organizer_attendee_question", Subject: "A guest sent you a question", Category: "organizer", Copy: "A guest sent a question about your event. Review it in the organizer workspace."},
	{ID: "organizer_attendee_update", Subject: "A guest updated their registration", Category: "organizer", Copy: "A guest updated their registration details. Review the current response in the organizer workspace."},
	{ID: "export_ready", Subject: "Your export is ready", Category: "organizer", Copy: "Your requested export is ready to download. The link will expire according to your workspace policy."},
	{ID: "moderation_review", Subject: "Your event is under review", Category: "moderation", TextOnly: true, Copy: "Your event was submitted for review. We will update you when its review status changes."},
	{ID: "moderation_approved", Subject: "Your event was approved", Category: "moderation", Copy: "Your event passed review and can be published from the organizer workspace."},
	{ID: "moderation_rejected", Subject: "Your event needs changes", Category: "moderation", TextOnly: true, Copy: "Your event needs changes before it can be published. Review the organizer workspace for the next steps."},
	{ID: "team_invitation", Subject: "You were invited to a Hackways workspace", Category: "organizer", Copy: "You were invited to collaborate in a Hackways workspace. Sign in to review the invitation."},
	{ID: "platform_notice", Subject: "A notice from Hackways", Category: "platform", TextOnly: true, Copy: "This is a notice about your Hackways account or a service you use."},
}

// EmailTemplateCatalog returns a copy so callers cannot mutate the renderer's
// process-wide definitions.
func EmailTemplateCatalog() []EmailTemplate {
	result := make([]EmailTemplate, len(emailTemplateCatalog))
	copy(result, emailTemplateCatalog)
	return result
}

func emailTemplate(id string) (EmailTemplate, bool) {
	for _, template := range emailTemplateCatalog {
		if template.ID == id {
			return template, true
		}
	}
	return EmailTemplate{}, false
}
