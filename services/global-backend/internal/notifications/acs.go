package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ACSEmailSender submits a new operation before polling its status. Reclaimed
// attempts poll the persisted operation before considering a repeatable retry.
// A successful send operation is provider acceptance, not inbox delivery.
// Indeterminate requests older than ACS's five-minute replay window fail closed.
type ACSEmailSender struct {
	endpoint          *url.URL
	key               []byte
	sender            string
	accountSender     string
	promotionalSender string
	client            *http.Client
}

func NewACSEmailSender(endpoint, accessKey, sender string, alternateSenders ...string) (*ACSEmailSender, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Port() != "" || !strings.HasSuffix(u.Hostname(), ".communication.azure.com") || u.Path != "" && u.Path != "/" {
		return nil, errors.New("ACS_ENDPOINT must be a valid HTTPS Azure Communication Services endpoint")
	}
	key, err := base64.StdEncoding.DecodeString(accessKey)
	if err != nil || len(key) < 32 || sender == "" || strings.ContainsAny(sender, "\r\n") {
		return nil, errors.New("ACS access key and verified sender must be configured")
	}
	// Account/security mail must never inherit the event-notification sender.
	account, promotion := "noreply@mail.hackways.in", ""
	if len(alternateSenders) > 0 && alternateSenders[0] != "" {
		account = alternateSenders[0]
	}
	if len(alternateSenders) > 1 {
		promotion = alternateSenders[1]
	}
	for _, value := range []string{sender, account, promotion} {
		if value == "" {
			continue
		}
		address, e := mail.ParseAddress(value)
		if e != nil || address.Address != value || strings.ContainsAny(value, "\r\n") {
			return nil, errors.New("ACS sender addresses must be valid email addresses")
		}
	}
	return &ACSEmailSender{endpoint: u, key: key, sender: sender, accountSender: account, promotionalSender: promotion, client: &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

type providerResult struct {
	Status string `json:"status"`
}

func (s *ACSEmailSender) request(ctx context.Context, method, path string, body []byte, id string, first time.Time) (providerResult, int, time.Duration, error) {
	var result providerResult
	target := strings.TrimRight(s.endpoint.String(), "/") + path
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(body))
	if err != nil {
		return result, 0, 0, err
	}
	date := time.Now().UTC().Format(http.TimeFormat)
	sum := sha256.Sum256(body)
	contentHash := base64.StdEncoding.EncodeToString(sum[:])
	signing := method + "\n" + req.URL.RequestURI() + "\n" + date + ";" + req.URL.Host + ";" + contentHash
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(signing))
	req.Header.Set("x-ms-date", date)
	req.Header.Set("x-ms-content-sha256", contentHash)
	req.Header.Set("Authorization", "HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature="+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ms-client-request-id", id)
	if method == http.MethodPost {
		req.Header.Set("Operation-Id", id)
		req.Header.Set("Repeatability-Request-Id", id)
		req.Header.Set("Repeatability-First-Sent", first.UTC().Format(http.TimeFormat))
	}
	response, err := s.client.Do(req)
	if err != nil {
		return result, 0, 0, errors.New("ACS request did not return a confirmed outcome")
	}
	defer response.Body.Close()
	delay := 2 * time.Second
	if n, parseErr := strconv.Atoi(response.Header.Get("Retry-After")); parseErr == nil && n > 0 && n <= 60 {
		delay = time.Duration(n) * time.Second
	}
	if response.StatusCode == 200 || response.StatusCode == 202 {
		if err = json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result); err != nil {
			return result, response.StatusCode, delay, errors.New("ACS returned an unreadable operation result")
		}
	}
	return result, response.StatusCode, delay, nil
}
func (s *ACSEmailSender) Send(ctx context.Context, msg EmailMessage) error {
	from := s.sender
	switch msg.SenderClass {
	case "account":
		from = s.accountSender
	case "promotion":
		from = s.promotionalSender
		if from == "" {
			return errors.New("promotional sender is not configured")
		}
	case "event", "":
	default:
		return errors.New("unsupported email sender category")
	}
	if msg.Recipient == "" {
		return ErrEmptyRecipient
	}
	if msg.OperationID == "" || msg.FirstSentAt.IsZero() {
		return errors.New("persisted email operation identity is required")
	}
	path := "/emails/operations/" + url.PathEscape(msg.OperationID) + "?api-version=2023-03-31"
	var result providerResult
	var status int
	var delay time.Duration
	var err error
	if !msg.FirstAttempt {
		result, status, delay, err = s.request(ctx, "GET", path, nil, msg.OperationID, msg.FirstSentAt)
		if err != nil {
			// An ambiguous lookup is never permission to send another email.
			return err
		}
	}
	if msg.FirstAttempt || status == 404 {
		if time.Since(msg.FirstSentAt) > 4*time.Minute {
			return errors.New("ACS outcome requires operator reconciliation; replay window elapsed")
		}
		type address struct {
			Address string `json:"address"`
		}
		type content struct {
			Subject   string `json:"subject"`
			PlainText string `json:"plainText"`
			HTML      string `json:"html"`
		}
		type recipients struct {
			To []address `json:"to"`
		}
		// Account/security messages omit Reply-To. All event and promotional
		// senders direct replies to the shared support mailbox.
		var replyTo []address
		if msg.SenderClass != "account" {
			replyTo = []address{{"hello@hackways.in"}}
		}
		payload, marshalErr := json.Marshal(struct {
			Sender          string     `json:"senderAddress"`
			Content         content    `json:"content"`
			Recipients      recipients `json:"recipients"`
			ReplyTo         []address  `json:"replyTo,omitempty"`
			DisableTracking bool       `json:"userEngagementTrackingDisabled"`
		}{from, content{msg.Subject, msg.BodyText, msg.BodyHTML}, recipients{[]address{{msg.Recipient}}}, replyTo, true})
		if marshalErr != nil {
			return marshalErr
		}
		result, status, delay, err = s.request(ctx, "POST", "/emails:send?api-version=2023-03-31", payload, msg.OperationID, msg.FirstSentAt)
		if err != nil {
			return err
		}
	}
	for attempt := 0; attempt < 12; attempt++ {
		if status != 200 && status != 202 {
			return fmt.Errorf("ACS operation returned HTTP %d; delivery is unconfirmed", status)
		}
		switch result.Status {
		case "Succeeded":
			return nil
		case "Failed", "Canceled":
			return errors.New("ACS send operation failed")
		case "Running", "NotStarted":
		default:
			return errors.New("ACS operation returned an unknown state")
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		result, status, delay, err = s.request(ctx, "GET", path, nil, msg.OperationID, msg.FirstSentAt)
		if err != nil {
			return err
		}
	}
	return errors.New("ACS operation still pending; reconciliation will retry the same operation")
}
