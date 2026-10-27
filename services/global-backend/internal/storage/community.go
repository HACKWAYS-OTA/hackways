package storage

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func mediaFailure(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		platform.RespondError(w, 404, "NOT_FOUND", "Resource not found", nil)
	} else {
		platform.RespondDatabaseError(w, err)
	}
}
func (h *Handler) AdminRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(auth.RequirePlatformAdmin(auth.NewSessionManager(h.queries), h.queries))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := origin == "https://admin.hackways.in"
			if os.Getenv("APP_ENV") == "development" {
				allowed = allowed || origin == "http://localhost:3003" || origin == "http://127.0.0.1:3003" || origin == "http://localhost:4003" || origin == "http://127.0.0.1:4003"
			}
			if !allowed {
				platform.RespondError(w, 403, "INVALID_ORIGIN", "This request origin is not allowed", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Post("/upload-intent", h.CityUploadIntent)
	r.Post("/city-banner/complete", h.CompleteCityBanner)
	return r
}
func (h *Handler) CityUploadIntent(w http.ResponseWriter, r *http.Request) {
	var b struct {
		UploadIntentRequest
		ResourceID pgtype.UUID `json:"resource_id"`
	}
	if platform.DecodeJSON(r, &b) != nil || !b.ResourceID.Valid || b.Category != "city-banners" {
		platform.RespondError(w, 422, "INVALID_UPLOAD", "Choose a saved City Hub and an image", nil)
		return
	}
	if _, e := h.queries.ReadCityUpload(r.Context(), b.ResourceID); e != nil {
		mediaFailure(w, e)
		return
	}
	if h.signer.cfg.AccountKey == "" || h.signer.cfg.AccountName == "" {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Azure Blob Storage is not configured", nil)
		return
	}
	b.Category = "city-banners/" + uuidPath(b.ResourceID)
	intent, e := h.signer.CreateUploadIntent(b.UploadIntentRequest)
	if e != nil {
		platform.RespondError(w, 422, "INVALID_IMAGE", "Use a PNG, JPEG or WebP image up to 5 MB", nil)
		return
	}
	platform.RespondJSON(w, 200, intent)
}

type communityImage struct {
	ResourceID pgtype.UUID `json:"resource_id"`
	BlobKey    string      `json:"blob_key"`
	Revision   int64       `json:"revision"`
	Kind       string      `json:"kind"`
	Remove     bool        `json:"remove"`
}

func (h *Handler) verifiedImage(w http.ResponseWriter, r *http.Request, key, prefix string) (pgtype.Text, bool) {
	if !strings.HasPrefix(key, prefix) || !bannerObject.MatchString(strings.TrimPrefix(key, prefix)) {
		platform.RespondError(w, 400, "INVALID_BLOB", "This image does not belong to the resource", nil)
		return pgtype.Text{}, false
	}
	if h.signer.cfg.AccountKey == "" || h.signer.cfg.AccountName == "" {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Azure Blob Storage is not configured", nil)
		return pgtype.Text{}, false
	}
	now := time.Now().UTC()
	token, e := h.signer.generateBlobSAS(key, now.Add(-time.Minute), now.Add(time.Minute), "r")
	if e != nil {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Image verification is unavailable", nil)
		return pgtype.Text{}, false
	}
	blob := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", h.signer.cfg.AccountName, h.signer.cfg.Container, key)
	request, e := http.NewRequestWithContext(r.Context(), http.MethodHead, blob+"?"+token, nil)
	if e != nil {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Image verification is unavailable", nil)
		return pgtype.Text{}, false
	}
	client := http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, e := client.Do(request)
	if e != nil {
		platform.RespondError(w, 502, "UPLOAD_UNVERIFIED", "Could not verify the image. Retry to finish saving", nil)
		return pgtype.Text{}, false
	}
	defer response.Body.Close()
	content, _, parseErr := mime.ParseMediaType(response.Header.Get("Content-Type"))
	ext, allowed := AllowedMIMETypes[content]
	if response.StatusCode != 200 || parseErr != nil || !allowed || path.Ext(key) != ext || response.ContentLength <= 0 || response.ContentLength > MaxUploadSizeBytes || response.Header.Get("x-ms-blob-type") != "BlockBlob" {
		platform.RespondError(w, 422, "INVALID_IMAGE", "Upload a PNG, JPEG or WebP block blob up to 5 MB", nil)
		return pgtype.Text{}, false
	}
	base := strings.TrimRight(h.signer.cfg.CDNDomain, "/")
	if base == "" {
		base = fmt.Sprintf("https://%s.blob.core.windows.net", h.signer.cfg.AccountName)
	}
	return pgtype.Text{String: base + "/" + h.signer.cfg.Container + "/" + key, Valid: true}, true
}
func (h *Handler) CompleteSpaceImage(w http.ResponseWriter, r *http.Request) {
	h.completeCommunity(w, r, false)
}
func (h *Handler) CompleteCityBanner(w http.ResponseWriter, r *http.Request) {
	h.completeCommunity(w, r, true)
}
func (h *Handler) completeCommunity(w http.ResponseWriter, r *http.Request, city bool) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		platform.RespondError(w, 401, "UNAUTHORIZED", "Sign in to continue", nil)
		return
	}
	var b communityImage
	if platform.DecodeJSON(r, &b) != nil || !b.ResourceID.Valid || b.Revision < 1 || (!city && b.Kind != "logo" && b.Kind != "banner") {
		platform.RespondError(w, 422, "INVALID_IMAGE", "Choose an image and the current resource revision", nil)
		return
	}
	if h.queries == nil || h.pool == nil {
		platform.RespondError(w, 503, "STORAGE_UNAVAILABLE", "Storage authorization unavailable", nil)
		return
	}
	category := "space-" + b.Kind + "s"
	if city {
		category = "city-banners"
		if _, e := h.queries.ReadCityUpload(r.Context(), b.ResourceID); e != nil {
			mediaFailure(w, e)
			return
		}
	} else {
		if _, e := h.queries.ReadOwnedUploadSpace(r.Context(), db.ReadOwnedUploadSpaceParams{ID: b.ResourceID, UserID: user.UserID}); e != nil {
			mediaFailure(w, e)
			return
		}
	}
	image := pgtype.Text{}
	if !b.Remove {
		var valid bool
		image, valid = h.verifiedImage(w, r, b.BlobKey, category+"/"+uuidPath(b.ResourceID)+"/")
		if !valid {
			return
		}
	}
	tx, e := h.pool.Begin(r.Context())
	if e != nil {
		mediaFailure(w, e)
		return
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	if e = q.SetStorageWriteTimeout(r.Context()); e != nil {
		mediaFailure(w, e)
		return
	}
	if city {
		state, err := q.ReadCityUpload(r.Context(), b.ResourceID)
		if err != nil {
			mediaFailure(w, err)
			return
		}
		if state.Revision == b.Revision+1 && state.BannerUrl == image {
			platform.RespondJSON(w, 200, struct {
				Data db.ReadCityUploadRow `json:"data"`
			}{state})
			return
		}
		row, err := q.AttachCityBanner(r.Context(), db.AttachCityBannerParams{ID: b.ResourceID, Revision: b.Revision, ImageUrl: image})
		if errors.Is(err, pgx.ErrNoRows) {
			imageConflict(w)
			return
		}
		if err != nil {
			mediaFailure(w, err)
			return
		}
		if e = q.AuditCommunityMedia(r.Context(), db.AuditCommunityMediaParams{ActorUserID: user.UserID, EntityID: b.ResourceID, Action: "city.banner_updated", RequestID: platform.GetRequestID(r.Context())}); e != nil {
			mediaFailure(w, e)
			return
		}
		if e = tx.Commit(r.Context()); e != nil {
			mediaFailure(w, e)
			return
		}
		platform.RespondJSON(w, 200, struct {
			Data db.AttachCityBannerRow `json:"data"`
		}{row})
	} else {
		if e = lockSpaceMediaAccess(r.Context(), q, b.ResourceID, user.UserID); e != nil {
			mediaFailure(w, e)
			return
		}
		state, err := q.ReadSpaceImageState(r.Context(), db.ReadSpaceImageStateParams{ID: b.ResourceID, UserID: user.UserID})
		if err != nil {
			mediaFailure(w, err)
			return
		}
		current := state.BannerUrl
		if b.Kind == "logo" {
			current = state.LogoUrl
		}
		if state.Revision == b.Revision+1 && current == image {
			platform.RespondJSON(w, 200, struct {
				Data db.ReadSpaceImageStateRow `json:"data"`
			}{state})
			return
		}
		row, err := q.AttachSpaceImage(r.Context(), db.AttachSpaceImageParams{ID: b.ResourceID, UserID: user.UserID, Revision: b.Revision, Kind: b.Kind, ImageUrl: image})
		if errors.Is(err, pgx.ErrNoRows) {
			imageConflict(w)
			return
		}
		if err != nil {
			mediaFailure(w, err)
			return
		}
		if e = q.AuditCommunityMedia(r.Context(), db.AuditCommunityMediaParams{ActorUserID: user.UserID, EntityID: b.ResourceID, Action: "space." + b.Kind + "_updated", RequestID: platform.GetRequestID(r.Context())}); e != nil {
			mediaFailure(w, e)
			return
		}
		if e = tx.Commit(r.Context()); e != nil {
			mediaFailure(w, e)
			return
		}
		platform.RespondJSON(w, 200, struct {
			Data db.AttachSpaceImageRow `json:"data"`
		}{row})
	}
}
func imageConflict(w http.ResponseWriter) {
	platform.RespondError(w, 409, "REVISION_CONFLICT", "This page changed. Reload before saving the image", nil)
}
