package notifications

import (
	"context"
	"errors"
	"time"
)

var ErrEmptyRecipient = errors.New("recipient email address cannot be empty")
var ErrEmailProviderUnavailable = errors.New("transactional email provider is unavailable")

type EmailMessage struct {
	SenderClass string // account, event, or promotion; assigned by server template catalog.
	OperationID string
	FirstSentAt time.Time
	// FirstAttempt comes from the claimed row's persisted first_sent_at state,
	// never from retry_count (a process can crash before recording a failure).
	FirstAttempt bool
	Recipient    string
	Subject      string
	BodyText     string
	BodyHTML     string
}
type EmailSender interface {
	Send(context.Context, EmailMessage) error
}

// UnavailableEmailSender is a dependency error, never a simulated send.
type UnavailableEmailSender struct{}

func (UnavailableEmailSender) Send(context.Context, EmailMessage) error {
	return ErrEmailProviderUnavailable
}
