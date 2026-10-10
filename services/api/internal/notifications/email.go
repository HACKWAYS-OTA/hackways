package notifications

import (
	"context"
	"errors"
	"log/slog"
)

// ErrEmptyRecipient indicates missing recipient email.
var ErrEmptyRecipient = errors.New("recipient email address cannot be empty")

// EmailMessage contains notification payload fields.
type EmailMessage struct {
	Recipient string
	Subject   string
	BodyText  string
	BodyHTML  string
}

// EmailSender abstracts transactional email dispatch.
type EmailSender interface {
	Send(ctx context.Context, msg EmailMessage) error
}

// LogEmailSender logs emails to structured log stream (for local development).
type LogEmailSender struct {
	logger *slog.Logger
}

// NewLogEmailSender creates an EmailSender that outputs to slog.
func NewLogEmailSender(logger *slog.Logger) *LogEmailSender {
	return &LogEmailSender{logger: logger}
}

// Send logs the outbound email without requiring an external SMTP server.
func (s *LogEmailSender) Send(ctx context.Context, msg EmailMessage) error {
	if msg.Recipient == "" {
		return ErrEmptyRecipient
	}

	s.logger.Info("dispatched transactional email",
		"recipient", msg.Recipient,
		"subject", msg.Subject,
	)
	return nil
}
