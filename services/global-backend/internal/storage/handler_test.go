package storage

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStorageHandler_UnauthenticatedUploadIntents(t *testing.T) {
	cfg := AzureSASConfig{
		AccountName: "hackwaystest",
		AccountKey:  "dGVzdC1kZXZlbG9wbWVudC1rZXktZm9yLWhhY2t3YXlzLWF6dXJlLXNhcw==",
		Container:   "media-uploads",
	}
	signer := NewAzureBlobSASSigner(cfg)
	handler := NewHandler(signer)

	t.Run("Even valid metadata requires a session", func(t *testing.T) {
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

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d: %s", rr.Code, rr.Body.String())
		}

	})

	t.Run("Invalid metadata does not bypass authentication", func(t *testing.T) {
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

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rr.Code)
		}
	})

	t.Run("Oversized upload intent does not bypass authentication", func(t *testing.T) {
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

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rr.Code)
		}
	})
}
