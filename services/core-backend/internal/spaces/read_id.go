package spaces

import (
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/platform"
	"net/http"
)

func (h *Handler) ReadByID(w http.ResponseWriter, r *http.Request) {
	id, ok := resource(w, r)
	if !ok {
		return
	}
	row, e := h.q.ReadSpaceSummaryByID(r.Context(), id)
	if e != nil {
		failure(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.ReadSpaceSummaryByIDRow `json:"data"`
	}{row})
}
