package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type relatedItem struct {
	ID         string             `json:"id"`
	Label      string             `json:"label"`
	Detail     string             `json:"detail"`
	Resource   string             `json:"resource"`
	Status     string             `json:"status"`
	OccurredAt pgtype.Timestamptz `json:"occurred_at"`
}
type detailTab struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}
type detailPage struct {
	Tabs    []detailTab   `json:"tabs"`
	Section string        `json:"section"`
	Rows    []relatedItem `json:"rows"`
	Offset  int32         `json:"offset"`
	HasNext bool          `json:"has_next"`
}

func detailError(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 404, "NOT_FOUND", "Record not found", nil)
	} else {
		platform.RespondDatabaseError(w, err)
	}
}
func detailID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	var id pgtype.UUID
	if id.Scan(chi.URLParam(r, "id")) != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid record ID", nil)
		return id, false
	}
	return id, true
}
func sendDetail[T any](w http.ResponseWriter, data T, page detailPage) {
	if page.Rows == nil {
		page.Rows = []relatedItem{}
	}
	if page.Tabs == nil {
		page.Tabs = []detailTab{}
	}
	if len(page.Rows) > 20 {
		page.HasNext = true
		page.Rows = page.Rows[:20]
	}
	platform.RespondJSON(w, 200, struct {
		Data    T          `json:"data"`
		Related detailPage `json:"related"`
	}{data, page})
}
func (h *Handler) detailPage(w http.ResponseWriter, r *http.Request, id pgtype.UUID, resource string) (detailPage, bool) {
	_, _, _, offset, ok := pageParams(w, r)
	if !ok {
		return detailPage{}, false
	}
	page := detailPage{Offset: offset, Tabs: []detailTab{{"activity", "Activity"}}}
	switch resource {
	case "events", "moderation":
		page.Tabs = append([]detailTab{{"registrations", "Registrations"}, {"tickets", "Tickets"}}, page.Tabs...)
	case "spaces":
		page.Tabs = append([]detailTab{{"events", "Events"}, {"members", "Members"}}, page.Tabs...)
	}
	page.Section = r.URL.Query().Get("section")
	if page.Section == "" {
		page.Section = page.Tabs[0].Key
	}
	valid := false
	for _, tab := range page.Tabs {
		if tab.Key == page.Section {
			valid = true
		}
	}
	if !valid {
		platform.RespondError(w, 400, "INVALID_SECTION", "Unknown detail section", nil)
		return page, false
	}
	var err error
	switch page.Section {
	case "registrations":
		var rows []db.AdminEventRegistrationsRow
		rows, err = h.queries.AdminEventRegistrations(r.Context(), db.AdminEventRegistrationsParams{EventID: id, Limit: 21, Offset: offset})
		for _, v := range rows {
			page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
		}
	case "tickets":
		var rows []db.AdminEventTicketsRow
		rows, err = h.queries.AdminEventTickets(r.Context(), db.AdminEventTicketsParams{EventID: id, Limit: 21, Offset: offset})
		for _, v := range rows {
			page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
		}
	case "events":
		var rows []db.AdminSpaceEventsRow
		rows, err = h.queries.AdminSpaceEvents(r.Context(), db.AdminSpaceEventsParams{SpaceID: id, Limit: 21, Offset: offset})
		for _, v := range rows {
			page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
		}
	case "members":
		var rows []db.AdminSpaceMembersRow
		rows, err = h.queries.AdminSpaceMembers(r.Context(), db.AdminSpaceMembersParams{SpaceID: id, Limit: 21, Offset: offset})
		for _, v := range rows {
			page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
		}
	default:
		if resource == "events" || resource == "moderation" {
			var rows []db.AdminEventHistoryRow
			rows, err = h.queries.AdminEventHistory(r.Context(), db.AdminEventHistoryParams{EventID: id, Limit: 21, Offset: offset})
			for _, v := range rows {
				page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
			}
		} else {
			var rows []db.AdminEntityHistoryRow
			rows, err = h.queries.AdminEntityHistory(r.Context(), db.AdminEntityHistoryParams{EntityID: id, Limit: 21, Offset: offset})
			for _, v := range rows {
				page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
			}
		}
	}
	if err != nil {
		detailError(w, err)
		return page, false
	}
	return page, true
}

func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	resource := chi.URLParam(r, "resource")
	if resource == "audit" {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id < 1 {
			platform.RespondError(w, 400, "INVALID_ID", "Invalid audit ID", nil)
			return
		}
		row, err := h.queries.AdminAuditDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		sendDetail(w, row, detailPage{})
		return
	}
	id, ok := detailID(w, r)
	if !ok {
		return
	}
	switch resource {
	case "events", "moderation":
		row, err := h.queries.AdminEventDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		page, ok := h.detailPage(w, r, id, resource)
		if ok {
			sendDetail(w, row, page)
		}
	case "spaces":
		row, err := h.queries.AdminSpaceDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		page, ok := h.detailPage(w, r, id, resource)
		if ok {
			sendDetail(w, row, page)
		}
	case "registrations":
		row, err := h.queries.AdminRegistrationDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		page, ok := h.detailPage(w, r, id, resource)
		if ok {
			sendDetail(w, row, page)
		}
	case "tickets":
		row, err := h.queries.AdminTicketDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		page, ok := h.detailPage(w, r, id, resource)
		if ok {
			sendDetail(w, row, page)
		}
	default:
		platform.RespondError(w, 404, "NOT_FOUND", "Record not found", nil)
	}
}
