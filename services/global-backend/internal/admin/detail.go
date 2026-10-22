package admin

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
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
	if resource == "people" {
		page.Tabs = append([]detailTab{{"events", "Events"}, {"registrations", "Registrations"}, {"sessions", "Sessions"}, {"messages", "Messages"}}, page.Tabs...)
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
	case "events":
		var rows []db.AdminPersonEventsRow
		rows, err = h.queries.AdminPersonEvents(r.Context(), db.AdminPersonEventsParams{UserID: id, Limit: 21, Offset: offset})
		for _, v := range rows {
			page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
		}
	case "registrations":
		var rows []db.AdminPersonRegistrationsRow
		rows, err = h.queries.AdminPersonRegistrations(r.Context(), db.AdminPersonRegistrationsParams{UserID: id, Limit: 21, Offset: offset})
		for _, v := range rows {
			page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
		}
	case "sessions":
		var rows []db.AdminPersonSessionsRow
		user, authenticated := auth.UserFromContext(r.Context())
		if !authenticated {
			platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
			return page, false
		}
		rows, err = h.queries.AdminPersonSessions(r.Context(), db.AdminPersonSessionsParams{UserID: id, CurrentSessionID: user.SessionID, PageLimit: 21, PageOffset: offset})
		for _, v := range rows {
			page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
		}
	case "messages":
		var rows []db.AdminPersonMessagesRow
		rows, err = h.queries.AdminPersonMessages(r.Context(), db.AdminPersonMessagesParams{ID: id, Limit: 21, Offset: offset})
		for _, v := range rows {
			page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
		}
	default:
		var rows []db.AdminGlobalEntityHistoryRow
		rows, err = h.queries.AdminGlobalEntityHistory(r.Context(), db.AdminGlobalEntityHistoryParams{EntityID: id, Limit: 21, Offset: offset})
		for _, v := range rows {
			page.Rows = append(page.Rows, relatedItem{v.ID, v.Label, v.Detail, v.Resource, v.Status, v.OccurredAt})
		}
	}
	if err != nil {
		detailError(w, err)
		return page, false
	}
	return page, true
}
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id, ok := detailID(w, r)
	if !ok {
		return
	}
	resource := chi.URLParam(r, "resource")
	switch resource {
	case "people":
		row, err := h.queries.AdminPersonDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		page, ok := h.detailPage(w, r, id, resource)
		if ok {
			sendDetail(w, row, page)
		}
	case "messaging":
		row, err := h.queries.AdminMessageDetail(r.Context(), id)
		if err != nil {
			detailError(w, err)
			return
		}
		page, ok := h.detailPage(w, r, id, resource)
		if ok {
			sendDetail(w, row, page)
		}
	case "storage":
		kind := r.URL.Query().Get("kind")
		if kind != "event_banner" && kind != "avatar" {
			platform.RespondError(w, 400, "INVALID_ASSET_TYPE", "Choose an asset type", nil)
			return
		}
		row, err := h.queries.AdminAssetDetail(r.Context(), db.AdminAssetDetailParams{ID: id, Kind: kind})
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
