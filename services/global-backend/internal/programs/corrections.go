package programs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) correctPublishedDecision(w http.ResponseWriter, r *http.Request) {
	if !origin(w, r, true) {
		return
	}
	id, ok := applicationID(w, r)
	if !ok {
		return
	}
	var body struct {
		Decision  string `json:"decision"`
		Notes     string `json:"notes"`
		Revision  int64  `json:"revision"`
		Confirmed bool   `json:"confirmed"`
	}
	if platform.DecodeJSON(r, &body) != nil || !body.Confirmed || !length(body.Notes, 0, 3000) || body.Decision == "pending" || !validDecision(body.Decision) || body.Decision == "" {
		invalid(w, "Choose an updated result and confirm publication")
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) < 16 || len(key) > 128 || strings.ContainsAny(key, " \r\n\t") {
		invalid(w, "Use a valid result update action key")
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
	raw, _ := json.Marshal(struct {
		Application string
		Decision    string
		Notes       string
		Revision    int64
	}{chi.URLParam(r, "id"), body.Decision, body.Notes, body.Revision})
	digest := sha256.Sum256(raw)
	hash := hex.EncodeToString(digest[:])
	previous, err := q.ProgramReleaseLookup(r.Context(), db.ProgramReleaseLookupParams{ProgramID: p.ID, RequestKey: key})
	if err == nil {
		if previous.PayloadHash != hash {
			conflict(w)
			return
		}
		_ = tx.Rollback(r.Context())
		h.adminApplication(w, r)
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		failure(w, err)
		return
	}
	application, err := q.ProgramApplicationLockID(r.Context(), db.ProgramApplicationLockIDParams{ProgramID: p.ID, ID: id})
	if err != nil {
		failure(w, err)
		return
	}
	if application.Status != "submitted" || !application.PublishedAt.Valid || application.Revision != body.Revision {
		conflict(w)
		return
	}
	if application.PublishedDecision.String == body.Decision {
		invalid(w, "Choose a different result before publishing an update")
		return
	}
	if err = q.ProgramSupersedeUnsentResults(r.Context(), chi.URLParam(r, "id")); err != nil {
		failure(w, err)
		return
	}
	_, err = q.ProgramCorrectPublishedDecision(r.Context(), db.ProgramCorrectPublishedDecisionParams{ID: id, Revision: body.Revision, Decision: body.Decision, ReviewerNotes: body.Notes, ReviewedBy: actor(r).UserID})
	if err != nil {
		failure(w, err)
		return
	}
	scheduled, err := q.ProgramEnqueueCorrectedResult(r.Context(), id)
	if err != nil {
		failure(w, err)
		return
	}
	if scheduled != 1 {
		platform.RespondError(w, 409, "APPLICANT_UNAVAILABLE", "This applicant cannot currently receive a result update", nil)
		return
	}
	if _, err = q.ProgramReleaseCreate(r.Context(), db.ProgramReleaseCreateParams{ProgramID: p.ID, RequestKey: key, PayloadHash: hash, ApplicationCount: 1, ReleasedBy: actor(r).UserID}); err != nil {
		failure(w, err)
		return
	}
	if err = q.ProgramTouchReview(r.Context(), p.ID); err == nil {
		err = audit(q, r, p, id, "application.result_corrected."+application.PublishedDecision.String+"_to_"+body.Decision)
	}
	if err == nil {
		err = tx.Commit(r.Context())
	}
	if err != nil {
		failure(w, err)
		return
	}
	h.adminApplication(w, r)
}
