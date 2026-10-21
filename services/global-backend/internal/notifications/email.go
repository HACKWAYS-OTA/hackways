package notifications

import (
	"context"
	"errors"
	"time"
)

var ErrEmptyRecipient = errors.New("recipient email address cannot be empty")
var ErrEmailProviderUnavailable = errors.New("transactional email provider is unavailable")

type EmailMessage struct {
	OperationID string
	FirstSentAt time.Time
	Recipient   string
	Subject     string
	BodyText    string
	BodyHTML    string
}
type EmailSender interface {
	Send(context.Context, EmailMessage) error
}

// UnavailableEmailSender is a dependency error, never a simulated send.
type UnavailableEmailSender struct{}

func (UnavailableEmailSender) Send(context.Context, EmailMessage) error {
	return ErrEmailProviderUnavailable
}
