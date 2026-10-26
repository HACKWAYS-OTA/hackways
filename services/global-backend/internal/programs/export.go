package programs

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
)

var exportSlot = make(chan struct{}, 1)

func reviewLabel(value string) string {
	switch value {
	case "accepted":
		return "Accepted"
	case "rejected":
		return "Not selected"
	case "waitlisted":
		return "Waitlisted"
	case "pending":
		return "Needs review"
	default:
		return "Not released"
	}
}
func validDecision(value string) bool {
	return value == "" || value == "pending" || value == "accepted" || value == "rejected" || value == "waitlisted"
}

func (h *Handler) exportApplicants(w http.ResponseWriter, r *http.Request) {
	select {
	case exportSlot <- struct{}{}:
		defer func() { <-exportSlot }()
	default:
		platform.RespondError(w, 429, "EXPORT_BUSY", "Another export is being prepared. Try again shortly.", nil)
		return
	}
	search, decision := strings.TrimSpace(r.URL.Query().Get("q")), r.URL.Query().Get("decision")
	if len(search) > 200 || !validDecision(decision) {
		invalid(w, "Use valid applicant filters")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		failure(w, err)
		return
	}
	defer tx.Rollback(ctx)
	q := h.q.WithTx(tx)
	if err = q.SetStorageWriteTimeout(ctx); err != nil {
		failure(w, err)
		return
	}
	p, err := q.ProgramRead(ctx, chi.URLParam(r, "slug"))
	if err != nil {
		failure(w, err)
		return
	}
	count, err := q.ProgramExportCount(ctx, db.ProgramExportCountParams{ProgramID: p.ID, DecisionFilter: decision, Search: search})
	if err != nil {
		failure(w, err)
		return
	}
	if count > 5000 {
		platform.RespondError(w, 422, "EXPORT_LIMIT", "Narrow your filters to export at most 5,000 applicants per workbook.", nil)
		return
	}
	zone := time.FixedZone("IST", 19800)
	book, err := newApplicantWorkbook(p.Title+" - Applicants", fmt.Sprintf("%d submitted applicants | Exported %s IST", count, time.Now().In(zone).Format("02 Jan 2006, 15:04")))
	if err != nil {
		failure(w, err)
		return
	}
	for offset := int32(0); int64(offset) < count; offset += 100 {
		rows, readErr := q.ProgramExportPage(ctx, db.ProgramExportPageParams{ProgramID: p.ID, DecisionFilter: decision, Search: search, PageOffset: offset})
		if readErr != nil {
			failure(w, readErr)
			return
		}
		for _, a := range rows {
			style := 3
			if book.row%2 == 0 {
				style = 4
			}
			submitted := time.Time{}
			if a.SubmittedAt.Valid {
				submitted = a.SubmittedAt.Time
			}
			book.addDatedRow([]string{a.FullName, a.Email, a.Phone, a.City, a.ProfileUrl, "", reviewLabel(a.Decision), reviewLabel(a.PublishedDecision.String), a.Biography, a.Proposal}, style, 60, submitted)
		}
		if book.err != nil {
			break
		}
	}
	data, err := book.finish()
	if errors.Is(err, errWorkbookLimit) {
		platform.RespondError(w, 422, "EXPORT_SIZE_LIMIT", "This workbook is too large. Narrow your applicant filters and try again.", nil)
		return
	}
	if err != nil {
		failure(w, err)
		return
	}
	if err = tx.Commit(ctx); err != nil {
		failure(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": p.Slug + "-applicants-" + time.Now().In(zone).Format("20060102") + ".xlsx"}))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
