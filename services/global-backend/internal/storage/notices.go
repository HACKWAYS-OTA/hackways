package storage

import (
	"encoding/json"
	"errors"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
)

func (h *Handler) NoticeImageIntent(w http.ResponseWriter, r *http.Request) {
	var b struct {
		UploadIntentRequest
		ResourceID pgtype.UUID `json:"resource_id"`
	}
	if platform.DecodeJSON(r, &b) != nil || !b.ResourceID.Valid {
		platform.RespondError(w, 422, "INVALID_UPLOAD", "Save a notice draft and choose an image", nil)
		return
	}
	if _, e := h.queries.ReadNoticeUpload(r.Context(), b.ResourceID); e != nil {
		mediaFailure(w, e)
		return
	}
	if h.signer.cfg.AccountKey == "" || h.signer.cfg.AccountName == "" {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Azure Blob Storage is not configured", nil)
		return
	}
	b.Category = "notice-images/" + uuidPath(b.ResourceID)
	intent, e := h.signer.CreateUploadIntent(b.UploadIntentRequest)
	if e != nil {
		platform.RespondError(w, 422, "INVALID_IMAGE", "Use a PNG, JPEG or WebP image up to 5 MB", nil)
		return
	}
	platform.RespondJSON(w, 200, intent)
}
func (h *Handler) CompleteNoticeImage(w http.ResponseWriter, r *http.Request) {
	var b struct {
		ResourceID pgtype.UUID `json:"resource_id"`
		BlobKey    string      `json:"blob_key"`
		Width      int32       `json:"width"`
		Height     int32       `json:"height"`
	}
	if platform.DecodeJSON(r, &b) != nil || !b.ResourceID.Valid || b.Width < 1 || b.Width > 20000 || b.Height < 1 || b.Height > 20000 {
		platform.RespondError(w, 422, "INVALID_IMAGE", "Choose a saved notice and uploaded image", nil)
		return
	}
	image, ok := h.verifiedImage(w, r, b.BlobKey, "notice-images/"+uuidPath(b.ResourceID)+"/")
	if !ok {
		return
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		platform.RespondError(w, 401, "ADMIN_AUTH_REQUIRED", "Admin sign in required", nil)
		return
	}
	tx, e := h.pool.Begin(r.Context())
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	if e = q.SetStorageWriteTimeout(r.Context()); e == nil {
		_, e = q.LockAuthenticationUser(r.Context(), user.UserID)
	}
	var granted bool
	if e == nil {
		granted, e = q.IsPlatformAdminSession(r.Context(), db.IsPlatformAdminSessionParams{SessionID: user.SessionID, UserID: user.UserID})
	}
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	if !granted {
		platform.RespondError(w, 403, "ADMIN_ACCESS_REQUIRED", "Admin access is required", nil)
		return
	}
	if _, e = q.ReadNoticeUpload(r.Context(), b.ResourceID); e != nil {
		mediaFailure(w, e)
		return
	}
	url, e := q.RecordNoticeImage(r.Context(), db.RecordNoticeImageParams{NoticeID: b.ResourceID, BlobKey: b.BlobKey, Url: image.String, CreatedBy: user.UserID, Width: b.Width, Height: b.Height})
	if errors.Is(e, pgx.ErrNoRows) {
		url, e = q.ReadNoticeImage(r.Context(), db.ReadNoticeImageParams{NoticeID: b.ResourceID, BlobKey: b.BlobKey})
		if e != nil {
			platform.RespondDatabaseError(w, e)
			return
		}
		platform.RespondJSON(w, 200, struct {
			URL string `json:"url"`
		}{url})
		return
	}
	if e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	metadata, _ := json.Marshal(struct {
		BlobKey string `json:"blob_key"`
	}{b.BlobKey})
	if e = q.AppendAdminAudit(r.Context(), db.AppendAdminAuditParams{ActorUserID: user.UserID, EntityID: b.ResourceID, Action: "admin.notice.image.attached", RequestID: platform.GetRequestID(r.Context()), Metadata: metadata}); e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	if e = tx.Commit(r.Context()); e != nil {
		platform.RespondDatabaseError(w, e)
		return
	}
	platform.RespondJSON(w, 200, struct {
		URL string `json:"url"`
	}{url})
}
