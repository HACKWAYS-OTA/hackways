package notifications

// EmailTemplate describes a supported outbox template. The catalog is kept in
// the global backend so producers can opt into a consistent voice and layout
// without adding presentation logic to the Go services that enqueue messages.
type EmailTemplate struct {
	ID       string
	Subject  string
	Category string
	Copy     string
}

var emailTemplateCatalog = []EmailTemplate{
	{ID: "program_application_received", Subject: "Thanks for applying", Category: "event", Copy: "Your program application has been received."},
	{ID: "program_results_released", Subject: "Your program application update", Category: "event", Copy: "An update to your program application is ready."},
	{ID: "admin_service_message", Subject: "A message from Hackways", Category: "event", Copy: "A service update from Hackways."},
	{ID: "admin_daily_otp", Subject: "Verify your Hackways admin sign-in", Category: "security", Copy: "Confirm this device for the next 24 hours. This code expires in 10 minutes. An incorrect or expired code signs this device out."},
	{ID: "admin_promotion", Subject: "Hackways updates", Category: "promotion", Copy: "News and offers from Hackways."},
	{ID: "email_verification", Subject: "Verify your email", Category: "account", Copy: "Use the verification code below to confirm your email address."},
	{ID: "login_otp", Subject: "Your Hackways sign-in code", Category: "account", Copy: "Use this one-time code to sign in securely."},
	{ID: "email_change_otp", Subject: "Confirm your new email", Category: "account", Copy: "Use this one-time code to confirm your new email address."},
	{ID: "guest_access_otp", Subject: "Your guest access code", Category: "account", Copy: "Use this one-time code to open your guest registration."},
	{ID: "password_reset_requested", Subject: "Reset your password", Category: "account", Copy: "A password reset was requested for your Hackways account."},
	{ID: "password_changed", Subject: "Your password was changed", Category: "security", Copy: "Your Hackways password was changed."},
	{ID: "email_changed", Subject: "Your email address was changed", Category: "security", Copy: "The email address on your Hackways account was changed."},
	{ID: "sign_in_alert", Subject: "New sign-in to your account", Category: "security", Copy: "A new sign-in to your Hackways account was recorded."},
	{ID: "session_revoked", Subject: "A session was signed out", Category: "security", Copy: "A session for your Hackways account was signed out."},
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
	{ID: "event_starting", Subject: "Your event is starting", Category: "event", Copy: "Your event is starting. Check the event details for the latest location or joining information."},
	{ID: "waitlist_promoted", Subject: "You moved off the waitlist", Category: "registration", Copy: "A place is now available for you. Sign in to review the registration details."},
	{ID: "ticket_issued", Subject: "Your ticket is ready", Category: "registration", Copy: "Your ticket is ready for the event. Keep this message for your records."},
	{ID: "ticket_revoked", Subject: "Your ticket was revoked", Category: "registration", Copy: "Your ticket is no longer valid. Contact the organizer if you believe this was unexpected."},
	{ID: "checkin_ready", Subject: "Check-in is ready", Category: "event", Copy: "Check-in is ready for your event. Follow the organizer's instructions when you arrive."},
	{ID: "checkin_confirmed", Subject: "You are checked in", Category: "event", Copy: "Your check-in was recorded successfully."},
	{ID: "organizer_attendee_question", Subject: "A guest sent you a question", Category: "organizer", Copy: "A guest sent a question about your event. Review it in the organizer workspace."},
	{ID: "export_ready", Subject: "Your export is ready", Category: "organizer", Copy: "Your requested export is ready to download. The link will expire according to your workspace policy."},
	{ID: "moderation_approved", Subject: "Your event was approved", Category: "moderation", Copy: "Your event passed review and can be published from the organizer workspace."},
	{ID: "event_invitation", Subject: "You are invited to an event", Category: "invitation", Copy: "You have been invited to join an event on Hackways."},
	{ID: "space_invitation", Subject: "You are invited to a Hackways space", Category: "invitation", Copy: "You have been invited to join a Hackways space."},
	{ID: "organizer_invitation", Subject: "You are invited to organize on Hackways", Category: "invitation", Copy: "You have been invited to collaborate as an organizer on Hackways."},
	{ID: "guest_invitation", Subject: "You are invited as a guest", Category: "invitation", Copy: "You have been invited to register as a guest for an event."},
	{ID: "team_invitation", Subject: "You were invited to a Hackways workspace", Category: "organizer", Copy: "You were invited to collaborate in a Hackways workspace. Sign in to review the invitation."},
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
