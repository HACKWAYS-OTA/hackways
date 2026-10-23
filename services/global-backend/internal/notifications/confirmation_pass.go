package notifications

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/skip2/go-qrcode"
)

func isConfirmation(template string) bool {
	return template == "rsvp_confirmation" || template == "rsvp_confirmed"
}

// This wire format is shared with core-backend's scanner. Admission remains an
// event-locked core transaction; a signature never overrides revoked tickets.
func confirmationCode(pass db.ReadConfirmationPassRow) (string, error) {
	seed, err := base64.StdEncoding.DecodeString(os.Getenv("TICKET_SIGNING_SEED"))
	if err != nil || len(seed) != ed25519.SeedSize {
		return "", errors.New("confirmation pass signing is not configured")
	}
	claims, err := json.Marshal(struct {
		TicketID  pgtype.UUID `json:"ticket_id"`
		EventID   pgtype.UUID `json:"event_id"`
		ExpiresAt int64       `json:"expires_at"`
	}{pass.TicketID, pass.EventID, pass.EndsAt.Time.Add(24 * time.Hour).Unix()})
	if err != nil {
		return "", errors.New("confirmation pass could not be encoded")
	}
	key := ed25519.NewKeyFromSeed(seed)
	return base64.RawURLEncoding.EncodeToString(claims) + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(key, claims)), nil
}

func attachment(name, contentType, contentID string, data []byte) EmailAttachment {
	return EmailAttachment{Name: name, ContentType: contentType, ContentID: contentID, ContentInBase64: base64.StdEncoding.EncodeToString(data)}
}

func passLinks(pass db.ReadConfirmationPassRow) (string, string, error) {
	origin := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/")
	if origin == "" {
		origin = "https://hackways.in"
	}
	parsed, err := url.Parse(origin)
	if err != nil || safeHTTPSURL(origin) == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", errors.New("confirmation pass public origin is invalid")
	}
	return origin + "/events/" + url.PathEscape(pass.Slug), origin + "/my-events/" + url.PathEscape(pass.RegistrationCode), nil
}

func (w *OutboxWorker) confirmationPass(ctx context.Context, entry db.ClaimEmailOutboxRow) (EmailMessage, error) {
	readCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	pass, err := w.queries.ReadConfirmationPass(readCtx, entry.ID)
	cancel()
	if errors.Is(err, pgx.ErrNoRows) {
		return EmailMessage{}, errSuperseded
	}
	if err != nil {
		return EmailMessage{}, errors.New("confirmation pass details could not be read")
	}
	return w.renderConfirmationPass(ctx, entry, pass)
}

func (w *OutboxWorker) renderConfirmationPass(ctx context.Context, entry db.ClaimEmailOutboxRow, pass db.ReadConfirmationPassRow) (EmailMessage, error) {
	code, err := confirmationCode(pass)
	if err != nil {
		return EmailMessage{}, err
	}
	qr, err := qrcode.Encode(code, qrcode.Medium, 384)
	if err != nil {
		return EmailMessage{}, errors.New("confirmation QR could not be rendered")
	}
	eventURL, passURL, err := passLinks(pass)
	if err != nil {
		return EmailMessage{}, err
	}
	var payload templatePayload
	if err = json.Unmarshal(entry.PayloadJson, &payload); err != nil {
		return EmailMessage{}, errors.New("confirmation payload is invalid")
	}
	payload.Title, payload.StartsAt, payload.EndsAt = pass.Title, pass.StartsAt.Time, pass.EndsAt.Time
	payload.EventURL, payload.ActionURL = eventURL, passURL
	payload.EventBannerURL, payload.Location = pass.BannerUrl.String, passLocation(pass)
	// Do not print the machine admission credential as text in the message.
	payload.TicketCode = ""
	entry.PayloadJson, err = json.Marshal(payload)
	if err != nil {
		return EmailMessage{}, errors.New("confirmation content could not be prepared")
	}
	message, err := render(entry)
	if err != nil {
		return EmailMessage{}, err
	}
	payload.PassQR = true
	template, _ := emailTemplate(entry.TemplateType)
	message.BodyHTML = renderHTML(template, message.Subject, message.BodyText, payload)
	message.BodyText += "\n\nMy pass: " + passURL + "\nShow your pass at check-in. Keep it private. The attached calendar file adds the event to your calendar."
	message.Attachments = []EmailAttachment{
		attachment("entry-qr.png", "image/png", "hackways-entry-qr", qr),
		attachment("event.ics", "text/calendar", "", confirmationCalendar(pass, eventURL, passURL)),
	}
	return message, nil
}

func passLocation(pass db.ReadConfirmationPassRow) string {
	var parts []string
	for _, value := range []string{pass.LocationName.String, pass.LocationAddress.String} {
		if strings.TrimSpace(value) != "" {
			parts = append(parts, strings.TrimSpace(value))
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	return safeHTTPSURL(pass.OnlineUrl.String)
}

// RFC 5545: escape TEXT properties, fold by UTF-8 octets and use CRLF.
// PUBLISH exports an event without sending an RSVP invitation or attendee PII.
func calendarText(value string) string {
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\r' && r != '\n' {
			return -1
		}
		return r
	}, value)
	return strings.NewReplacer("\\", "\\\\", "\r\n", "\\n", "\r", "\\n", "\n", "\\n", ";", "\\;", ",", "\\,").Replace(value)
}

func calendarLine(value string) string {
	var out strings.Builder
	size := 0
	for _, r := range value {
		width := utf8.RuneLen(r)
		if size+width > 75 {
			out.WriteString("\r\n ")
			size = 1
		}
		out.WriteRune(r)
		size += width
	}
	return out.String()
}

func confirmationCalendar(pass db.ReadConfirmationPassRow, eventURL, passURL string) []byte {
	stamp := func(t time.Time) string { return t.UTC().Format("20060102T150405Z") }
	// Match the existing public calendar export UID so importing both does not
	// create two unrelated event identities. Times are UTC, independent of DST.
	lines := []string{"BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//Hackways//Events//EN", "CALSCALE:GREGORIAN", "METHOD:PUBLISH", "BEGIN:VEVENT",
		"UID:" + calendarText(pass.Slug) + "@hackways", "DTSTAMP:" + stamp(pass.QueuedAt.Time),
		"DTSTART:" + stamp(pass.StartsAt.Time), "DTEND:" + stamp(pass.EndsAt.Time),
		"SUMMARY:" + calendarText(pass.Title), "DESCRIPTION:" + calendarText(pass.Description+"\n\nMy pass: "+passURL),
		"URL:" + calendarText(eventURL), "STATUS:CONFIRMED"}
	if location := passLocation(pass); location != "" {
		lines = append(lines, "LOCATION:"+calendarText(location))
	}
	lines = append(lines, "END:VEVENT", "END:VCALENDAR")
	for i := range lines {
		lines[i] = calendarLine(lines[i])
	}
	return []byte(strings.Join(lines, "\r\n") + "\r\n")
}
