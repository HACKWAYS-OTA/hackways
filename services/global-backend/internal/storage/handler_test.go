package storage

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStorageHandler_CreateUploadIntent(t *testing.T) {
	cfg := AzureSASConfig{
		AccountName: "hackwaystest",
		AccountKey:  "dGVzdC1kZXZlbG9wbWVudC1rZXktZm9yLWhhY2t3YXlzLWF6dXJlLXNhcw==",
		Container:   "media-uploads",
	}
	signer := NewAzureBlobSASSigner(cfg)
	handler := NewHandler(signer)

	t.Run("Valid image upload intent", func(t *testing.T) {
		body, _ := json.Marshal(UploadIntentRequest{
			Filename: "banner.png",
			MIMEType: "image/png",
			Size:     1024 * 500, // 500 KB
			Category: "event-banners",
		})

		req := httptest.NewRequest(http.MethodPost, "/upload-intent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var res UploadIntentResponse
		if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if res.UploadURL == "" || res.BlobKey == "" {
			t.Fatalf("expected non-empty upload_url and blob_key, got %+v", res)
		}
	})

	t.Run("Disallowed MIME type returns 400", func(t *testing.T) {
		body, _ := json.Marshal(UploadIntentRequest{
			Filename: "malicious.exe",
			MIMEType: "application/x-msdownload",
			Size:     1024,
			Category: "event-banners",
		})

		req := httptest.NewRequest(http.MethodPost, "/upload-intent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
		}
	})

	t.Run("Oversized payload returns 400", func(t *testing.T) {
		body, _ := json.Marshal(UploadIntentRequest{
			Filename: "giant.png",
			MIMEType: "image/png",
			Size:     10 * 1024 * 1024, // 10 MB > 5 MB limit
			Category: "event-banners",
		})

		req := httptest.NewRequest(http.MethodPost, "/upload-intent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
		}
	})
}
