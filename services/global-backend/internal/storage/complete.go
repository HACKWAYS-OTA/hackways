package storage

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var bannerObject = regexp.MustCompile(`^[a-f0-9]{32}\.(jpg|png|webp)$`)

// CompleteEventBanner reads Azure properties only. Media bytes never pass through Go.
func (h *Handler) CompleteEventBanner(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		platform.RespondError(w, 401, "UNAUTHORIZED", "Authentication required", nil)
		return
	}
	var req struct {
		ResourceID string `json:"resource_id"`
		BlobKey    string `json:"blob_key"`
		Revision   int64  `json:"revision"`
	}
	if platform.DecodeJSON(r, &req) != nil {
		platform.RespondError(w, 400, "INVALID_REQUEST", "Provide the uploaded banner and current event revision", nil)
		return
	}
	var id pgtype.UUID
	if id.Scan(req.ResourceID) != nil || !id.Valid || req.Revision < 1 {
		platform.RespondError(w, 400, "INVALID_RESOURCE", "A valid event and revision are required", nil)
		return
	}
	prefix := "event-banners/" + uuidPath(id) + "/"
	if !strings.HasPrefix(req.BlobKey, prefix) || !bannerObject.MatchString(strings.TrimPrefix(req.BlobKey, prefix)) {
		platform.RespondError(w, 400, "INVALID_BLOB", "The banner must belong to this event", nil)
		return
	}
	if h.queries == nil || h.pool == nil || h.signer.cfg.AccountName == "" || h.signer.cfg.AccountKey == "" {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Azure Blob Storage is not configured", nil)
		return
	}
	_, err := h.queries.ReadOwnedUploadEvent(r.Context(), db.ReadOwnedUploadEventParams{ID: id, UserID: user.UserID})
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 404, "NOT_FOUND", "Event not found", nil)
		return
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	now := time.Now().UTC()
	token, err := h.signer.generateBlobSAS(req.BlobKey, now.Add(-time.Minute), now.Add(time.Minute), "r")
	if err != nil {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Storage verification is unavailable", nil)
		return
	}
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", h.signer.cfg.AccountName, h.signer.cfg.Container, req.BlobKey)
	request, err := http.NewRequestWithContext(r.Context(), http.MethodHead, blobURL+"?"+token, nil)
	if err != nil {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Storage verification is unavailable", nil)
		return
	}
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		platform.RespondError(w, 502, "STORAGE_VERIFICATION_FAILED", "The banner could not be verified. Retry the upload.", nil)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		platform.RespondError(w, 409, "UPLOAD_NOT_READY", "The uploaded banner is not available. Retry the upload.", nil)
		return
	}
	contentType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	extension, allowed := AllowedMIMETypes[contentType]
	if err != nil || !allowed || path.Ext(req.BlobKey) != extension || response.ContentLength <= 0 || response.ContentLength > MaxUploadSizeBytes || response.Header.Get("x-ms-blob-type") != "BlockBlob" {
		platform.RespondError(w, 422, "INVALID_BANNER", "Use a PNG, JPEG or WebP block blob up to 5 MB", nil)
		return
	}
	publicBase := strings.TrimRight(h.signer.cfg.CDNDomain, "/")
	if publicBase == "" {
		publicBase = fmt.Sprintf("https://%s.blob.core.windows.net", h.signer.cfg.AccountName)
	}
	publicURL := publicBase + "/" + h.signer.cfg.Container + "/" + req.BlobKey
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	queries := h.queries.WithTx(tx)
	if err = queries.SetStorageWriteTimeout(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	// Repeat ownership, editable event state and revision checks atomically.
	before, err := lockEventMediaAccess(r.Context(), queries, id, user.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The event changed or editing closed. Reload before continuing.", nil)
		return
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	event, err := queries.AttachOwnedEventBanner(r.Context(), db.AttachOwnedEventBannerParams{ID: id, UserID: user.UserID, BannerUrl: pgtype.Text{String: publicURL, Valid: true}, Revision: req.Revision})
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 409, "REVISION_CONFLICT", "The event changed or editing closed. Reload the event before continuing.", nil)
		return
	}
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	if before.BannerUrl.String != publicURL {
		if err = queries.AuditEventMedia(r.Context(), db.AuditEventMediaParams{ActorUserID: user.UserID, EventID: id, RequestID: platform.GetRequestID(r.Context())}); err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
	}
	if err = tx.Commit(r.Context()); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, 200, struct {
		Data db.Events `json:"data"`
	}{event})
}
