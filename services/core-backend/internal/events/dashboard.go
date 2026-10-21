package events

import (
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"net/http"
	"strings"
)

func managementSearch(w http.ResponseWriter, r *http.Request) (string, bool) {
	value := strings.TrimSpace(r.URL.Query().Get("q"))
	if len([]rune(value)) > 200 {
		platform.RespondError(w, 400, "INVALID_SEARCH", "Keep searches within 200 characters", nil)
		return "", false
	}
	return value, true
}
func (h *Handler) GuestSummary(w http.ResponseWriter, r *http.Request) {
	event, ok := h.owned(w, r)
	if !ok {
		return
	}
	result, err := h.queries.EventGuestSummary(r.Context(), event.ID)
	if err != nil {
		lifecycleError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.EventGuestSummaryRow `json:"data"`
	}{result})
}
func (h *Handler) managedList(w http.ResponseWriter, r *http.Request) {
	owner, _ := auth.OrganizerFromContext(r.Context())
	search, ok := managementSearch(w, r)
	if !ok {
		return
	}
	view := r.URL.Query().Get("view")
	if view == "" {
		view = "all"
	}
	switch view {
	case "all", "upcoming", "drafts", "past", "cancelled":
	default:
		platform.RespondError(w, 400, "INVALID_VIEW", "Choose all, upcoming, drafts, past or cancelled events", nil)
		return
	}
	limit, offset := pagination(r)
	rows, err := h.queries.ListManagedEvents(r.Context(), db.ListManagedEventsParams{OrganizerID: owner.ID, Search: search, View: view, PageLimit: limit, PageOffset: offset})
	if err != nil {
		lifecycleError(w, err)
		return
	}
	if rows == nil {
		rows = []db.Events{}
	}
	platform.RespondJSON(w, 200, struct {
		Data       []db.Events `json:"data"`
		Pagination struct {
			Limit  int32 `json:"limit"`
			Offset int32 `json:"offset"`
			Count  int   `json:"count"`
		} `json:"pagination"`
	}{rows, struct {
		Limit  int32 `json:"limit"`
		Offset int32 `json:"offset"`
		Count  int   `json:"count"`
	}{limit, offset, len(rows)}})
}
