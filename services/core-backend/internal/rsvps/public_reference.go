package rsvps

import (
	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"net/http"
)

func (h *Handler) GetByPublicCode(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	row, e := h.queries.GetAccountRegistrationByCode(r.Context(), db.GetAccountRegistrationByCodeParams{PublicCode: chi.URLParam(r, "code"), UserID: user.UserID})
	if e != nil {
		respondFailure(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Rsvps `json:"data"`
	}{row})
}
