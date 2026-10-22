package admin

import (
	"encoding/csv"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) PromotionContacts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.PromotionContactsList(r.Context(), campaignOffset(r))
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if rows == nil {
		rows = []db.PromotionContactsListRow{}
	}
	next := len(rows) > 20
	if next {
		rows = rows[:20]
	}
	platform.RespondJSON(w, 200, struct {
		Data    []db.PromotionContactsListRow `json:"data"`
		HasNext bool                          `json:"has_next"`
	}{rows, next})
}
func (h *Handler) ImportPromotionContacts(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CSV           string    `json:"csv"`
		ConsentSource string    `json:"consent_source"`
		ConsentedAt   time.Time `json:"consented_at"`
		Confirmed     bool      `json:"confirmed"`
	}
	if platform.DecodeJSON(r, &body) != nil || len(body.CSV) > 65536 || !body.Confirmed || strings.TrimSpace(body.ConsentSource) == "" || !utf8.ValidString(body.ConsentSource) || strings.ContainsAny(body.ConsentSource, "\x00\r\n") || utf8.RuneCountInString(body.ConsentSource) > 1000 || body.ConsentedAt.IsZero() || body.ConsentedAt.After(time.Now().Add(time.Minute)) {
		platform.RespondError(w, 422, "CONSENT_REQUIRED", "Provide CSV, consent source/date, and confirm that every imported contact opted in", nil)
		return
	}
	reader := csv.NewReader(strings.NewReader(body.CSV))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil || len(header) < 1 || len(header) > 2 || strings.ToLower(strings.TrimSpace(strings.TrimPrefix(header[0], "\ufeff"))) != "email" || (len(header) == 2 && strings.ToLower(strings.TrimSpace(header[1])) != "name") {
		platform.RespondError(w, 422, "INVALID_CSV", "Use a CSV header of email,name (name is optional)", nil)
		return
	}
	type contact struct{ email, name string }
	contacts := []contact{}
	seen := map[string]bool{}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) < 1 || len(row) > 2 || len(contacts) >= 200 {
			platform.RespondError(w, 422, "INVALID_CSV", "Import at most 200 contacts per file using email,name columns", nil)
			return
		}
		email := strings.ToLower(strings.TrimSpace(row[0]))
		address, err := mail.ParseAddress(email)
		name := ""
		if len(row) == 2 {
			name = strings.TrimSpace(row[1])
		}
		if err != nil || address.Address != email || len(email) > 320 || utf8.RuneCountInString(name) > 100 || !utf8.ValidString(name) || strings.ContainsAny(name, "\x00\r\n") {
			platform.RespondError(w, 422, "INVALID_CONTACT", "A row has an invalid email or name; no contacts were imported", nil)
			return
		}
		if !seen[email] {
			contacts = append(contacts, contact{email, name})
			seen[email] = true
		}
	}
	if len(contacts) == 0 {
		platform.RespondError(w, 422, "EMPTY_IMPORT", "Add at least one opted-in contact", nil)
		return
	}
	tx, q, user, reason, ok := h.campaignMutation(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	var added int64
	for _, c := range contacts {
		n, err := q.ImportPromotionContact(r.Context(), db.ImportPromotionContactParams{Email: c.email, DisplayName: c.name, ConsentSource: strings.TrimSpace(body.ConsentSource), ConsentedAt: pgtype.Timestamptz{Time: body.ConsentedAt, Valid: true}})
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		added += n
	}
	if err = campaignAudit(r, q, user, user.UserID, "contacts.import", reason, added); err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Imported  int64  `json:"imported"`
		Unchanged int64  `json:"unchanged"`
		Message   string `json:"message"`
	}{added, int64(len(contacts)) - added, "Imported new contacts. Existing contacts and unsubscribes were preserved."})
}
