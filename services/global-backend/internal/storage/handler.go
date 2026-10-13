package storage

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/internal/platform"
)

// Handler serves HTTP endpoints for cloud storage intent and asset signing.
type Handler struct {
	signer *AzureBlobSASSigner
}

// NewHandler constructs a Storage Handler.
func NewHandler(signer *AzureBlobSASSigner) *Handler {
	return &Handler{signer: signer}
}

// Routes mounts storage intent routes.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/upload-intent", h.CreateUploadIntent)
	return r
}

// CreateUploadIntent handles client requests for direct Azure Blob SAS write tokens.
func (h *Handler) CreateUploadIntent(w http.ResponseWriter, r *http.Request) {
	var req UploadIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON body", nil)
		return
	}

	res, err := h.signer.CreateUploadIntent(req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidFileType):
			platform.RespondError(w, http.StatusBadRequest, "UNSUPPORTED_MEDIA_TYPE", err.Error(), nil)
		case errors.Is(err, ErrFileTooLarge):
			platform.RespondError(w, http.StatusBadRequest, "PAYLOAD_TOO_LARGE", err.Error(), nil)
		case errors.Is(err, ErrEmptyFilename):
			platform.RespondError(w, http.StatusBadRequest, "INVALID_FILENAME", err.Error(), nil)
		default:
			platform.RespondError(w, http.StatusInternalServerError, "SAS_SIGNING_ERROR", "Failed to sign upload intent", nil)
		}
		return
	}

	platform.RespondJSON(w, http.StatusOK, res)
}
