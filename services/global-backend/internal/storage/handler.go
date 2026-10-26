package storage

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	signer  *AzureBlobSASSigner
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewHandler(signer *AzureBlobSASSigner, pools ...*pgxpool.Pool) *Handler {
	h := &Handler{signer: signer}
	if len(pools) > 0 && pools[0] != nil {
		h.queries = db.New(pools[0])
		h.pool = pools[0]
	}
	return h
}
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/upload-intent", h.CreateUploadIntent)
	r.Post("/event-banner/complete", h.CompleteEventBanner)
	r.Post("/space-image/complete", h.CompleteSpaceImage)
	return r
}
func (h *Handler) CreateUploadIntent(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		platform.RespondError(w, 401, "UNAUTHORIZED", "Authentication required", nil)
		return
	}
	var req struct {
		UploadIntentRequest
		ResourceID string `json:"resource_id"`
	}
	if err := platform.DecodeJSON(r, &req); err != nil {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Malformed JSON upload request", nil)
		return
	}
	category := strings.TrimSpace(req.Category)
	switch category {
	case "avatars":
		req.Category = "avatars/" + uuidPath(user.UserID)
	case "event-banners":
		var eventID pgtype.UUID
		if err := eventID.Scan(req.ResourceID); err != nil || !eventID.Valid {
			platform.RespondError(w, 400, "RESOURCE_REQUIRED", "A valid event resource_id is required", nil)
			return
		}
		if h.queries == nil {
			platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Storage authorization is unavailable", nil)
			return
		}
		_, err := h.queries.ReadOwnedUploadEvent(r.Context(), db.ReadOwnedUploadEventParams{ID: eventID, UserID: user.UserID})
		if errors.Is(err, pgx.ErrNoRows) {
			platform.RespondError(w, 404, "NOT_FOUND", "Event not found", nil)
			return
		}
		if err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
		req.Category = "event-banners/" + uuidPath(eventID)
	case "space-logos", "space-banners":
		var id pgtype.UUID
		if id.Scan(req.ResourceID) != nil || !id.Valid || h.queries == nil {
			platform.RespondError(w, 400, "INVALID_RESOURCE", "Choose a saved Space", nil)
			return
		}
		if _, err := h.queries.ReadOwnedUploadSpace(r.Context(), db.ReadOwnedUploadSpaceParams{ID: id, UserID: user.UserID}); err != nil {
			mediaFailure(w, err)
			return
		}
		req.Category = category + "/" + uuidPath(id)
	default:
		platform.RespondError(w, 400, "INVALID_CATEGORY", "Use avatars or event-banners", nil)
		return
	}
	if h.signer.cfg.AccountName == "" || h.signer.cfg.AccountKey == "" {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Azure Blob Storage is not configured", nil)
		return
	}
	response, err := h.signer.CreateUploadIntent(req.UploadIntentRequest)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidFileType):
			platform.RespondError(w, 415, "UNSUPPORTED_MEDIA_TYPE", err.Error(), nil)
		case errors.Is(err, ErrFileTooLarge):
			platform.RespondError(w, 413, "PAYLOAD_TOO_LARGE", err.Error(), nil)
		case errors.Is(err, ErrEmptyFilename):
			platform.RespondError(w, 400, "INVALID_FILENAME", err.Error(), nil)
		default:
			platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Upload signing is unavailable", nil)
		}
		return
	}
	platform.RespondJSON(w, 200, response)
}
func uuidPath(id pgtype.UUID) string {
	const hex = "0123456789abcdef"
	result := make([]byte, 32)
	for i, b := range id.Bytes {
		result[2*i] = hex[b>>4]
		result[2*i+1] = hex[b&15]
	}
	return string(result)
}
