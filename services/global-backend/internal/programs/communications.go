package programs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/notifications"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5/pgtype"
)

type communicationBody struct {
	Subject       string   `json:"subject"`
	Body          string   `json:"body"`
	Scope         string   `json:"scope"`
	IDs           []string `json:"ids"`
	Decision      string   `json:"decision"`
	Search        string   `json:"search"`
	AudienceHash  string   `json:"audience_hash"`
	ExpectedCount int      `json:"expected_count"`
	Confirmed     bool     `json:"confirmed"`
}

func decodeCommunication(w http.ResponseWriter, r *http.Request) (communicationBody, []pgtype.UUID, bool) {
	var body communicationBody
	if !origin(w, r, true) {
		return body, nil, false
	}
	if platform.DecodeJSON(r, &body) != nil {
		invalid(w, "Check your message and recipients")
		return body, nil, false
	}
	body.Subject, body.Body, body.Search = strings.TrimSpace(body.Subject), strings.TrimSpace(body.Body), strings.TrimSpace(body.Search)
	if !length(body.Subject, 2, 200) || strings.ContainsAny(body.Subject, "\r\n") || !length(body.Body, 2, 5000) || len(body.Search) > 200 || !validDecision(body.Decision) || (body.Scope != "selected" && body.Scope != "all" && body.Scope != "filtered") {
		invalid(w, "Add a subject and message, then choose your recipients")
		return body, nil, false
	}
	if body.Scope != "filtered" {
		body.Decision, body.Search = "", ""
	}
	ids := []pgtype.UUID{}
	if body.Scope == "selected" {
		if len(body.IDs) < 1 || len(body.IDs) > 200 {
			invalid(w, "Choose between one and 200 applicants")
			return body, nil, false
		}
		sort.Strings(body.IDs)
		for i, value := range body.IDs {
			var id pgtype.UUID
			if id.Scan(value) != nil || !id.Valid {
				invalid(w, "Use valid applicant selections")
				return body, nil, false
			}
			if i > 0 && value == body.IDs[i-1] {
				invalid(w, "Select each applicant only once")
				return body, nil, false
			}
			ids = append(ids, id)
		}
	} else {
		body.IDs = nil
	}
	return body, ids, true
}

func audienceFingerprint(rows []db.ProgramCommunicationAudienceRow) string {
	hash := sha256.New()
	for _, row := range rows {
		_, _ = fmt.Fprintf(hash, "%x:%d:%s\n", row.ID.Bytes, row.Revision, row.Email)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func audienceValid(w http.ResponseWriter, body communicationBody, rows []db.ProgramCommunicationAudienceRow) bool {
	if len(rows) > 2000 {
		platform.RespondError(w, 422, "AUDIENCE_LIMIT", "Use filters to message at most 2,000 applicants at a time.", nil)
		return false
	}
	if body.Scope == "selected" && len(rows) != len(body.IDs) {
		platform.RespondError(w, 409, "APPLICANT_UNAVAILABLE", "Some selected applicants are no longer eligible for this program message. Refresh your selection.", nil)
		return false
	}
	return true
}

func (h *Handler) communicationPreview(w http.ResponseWriter, r *http.Request) {
	body, ids, ok := decodeCommunication(w, r)
	if !ok {
		return
	}
	p, err := h.q.ProgramRead(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		failure(w, err)
		return
	}
	rows, err := h.q.ProgramCommunicationAudience(r.Context(), db.ProgramCommunicationAudienceParams{ProgramID: p.ID, Selected: body.Scope == "selected", ApplicationIds: ids, DecisionFilter: body.Decision, Search: body.Search})
	if err != nil {
		failure(w, err)
		return
	}
	if !audienceValid(w, body, rows) {
		return
	}
	type preview struct {
		Count     int    `json:"count"`
		Hash      string `json:"audience_hash"`
		HTML      string `json:"html"`
		Subject   string `json:"subject"`
		Recipient string `json:"recipient"`
	}
	result := preview{Count: len(rows), Hash: audienceFingerprint(rows)}
	if len(rows) > 0 {
		a := rows[0]
		message, renderErr := notifications.RenderProgramCommunicationPreview(notifications.ProgramMailDetails{Name: a.FullName, Title: p.Title, Slug: p.Slug, City: a.City, Proposal: a.Proposal, ProfileURL: a.ProfileUrl, Decision: a.PublishedDecision.String, SubmittedAt: a.SubmittedAt.Time}, body.Subject, body.Body)
		if renderErr != nil {
			invalid(w, "The message preview could not be prepared")
			return
		}
		result.HTML, result.Subject, result.Recipient = message.BodyHTML, message.Subject, a.Email
	}
	platform.RespondJSON(w, 200, result)
}

func (h *Handler) communicationSend(w http.ResponseWriter, r *http.Request) {
	body, ids, ok := decodeCommunication(w, r)
	if !ok {
		return
	}
	if !body.Confirmed || body.ExpectedCount < 1 || len(body.AudienceHash) != 64 {
		invalid(w, "Preview the message and confirm your recipients before sending")
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) < 16 || len(key) > 128 || strings.ContainsAny(key, " \r\n\t") {
		invalid(w, "Use a valid message action key")
		return
	}
	if !notifications.EmailConfiguredFor("program_communication") {
		platform.RespondError(w, 503, "EMAIL_UNAVAILABLE", "Email delivery is not available. Your message has not been sent.", nil)
		return
	}
	tx, q, ok := h.begin(w, r)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	p, err := q.ProgramLock(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		failure(w, err)
		return
	}
	canonical, _ := json.Marshal(struct {
		Program pgtype.UUID
		Message communicationBody
	}{p.ID, body})
	digest := sha256.Sum256(canonical)
	message, err := q.AdminMessageCreate(r.Context(), db.AdminMessageCreateParams{ActorUserID: actor(r).UserID, CreationKey: key, CreationHash: hex.EncodeToString(digest[:]), Subject: body.Subject})
	if err != nil {
		failure(w, err)
		return
	}
	if message.CreationHash != hex.EncodeToString(digest[:]) {
		conflict(w)
		return
	}
	count, err := q.AdminMessageQueuedCount(r.Context(), message.ID)
	if err != nil {
		failure(w, err)
		return
	}
	if count > 0 {
		platform.RespondJSON(w, 200, struct {
			Scheduled int64 `json:"scheduled"`
			Reference int64 `json:"reference"`
		}{count, message.Reference})
		return
	}
	rows, err := q.ProgramCommunicationAudience(r.Context(), db.ProgramCommunicationAudienceParams{ProgramID: p.ID, Selected: body.Scope == "selected", ApplicationIds: ids, DecisionFilter: body.Decision, Search: body.Search})
	if err != nil {
		failure(w, err)
		return
	}
	if !audienceValid(w, body, rows) {
		return
	}
	if len(rows) != body.ExpectedCount || audienceFingerprint(rows) != body.AudienceHash {
		platform.RespondError(w, 409, "AUDIENCE_CHANGED", "The recipient list changed. Preview again before sending.", nil)
		return
	}
	ids = make([]pgtype.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	count, err = q.ProgramQueueCommunications(r.Context(), db.ProgramQueueCommunicationsParams{ProgramID: p.ID, ApplicationIds: ids, Subject: body.Subject, Body: body.Body, MessageID: message.ID})
	if err != nil {
		failure(w, err)
		return
	}
	if count != int64(len(rows)) {
		conflict(w)
		return
	}
	if err = audit(q, r, p, pgtype.UUID{}, "program.communication_scheduled"); err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		failure(w, err)
		return
	}
	platform.RespondJSON(w, 202, struct {
		Scheduled int64 `json:"scheduled"`
		Reference int64 `json:"reference"`
	}{count, message.Reference})
}
