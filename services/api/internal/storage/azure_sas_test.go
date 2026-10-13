package storage

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestAzureBlobSASSigner_CreateUploadIntent(t *testing.T) {
	cfg := AzureSASConfig{
		AccountName: "hackwaysstorage",
		AccountKey:  "aGVsbG93b3JsZGF6dXJlc3RvcmFnZWtleTEyMzQ1Njc4OQ==", // base64 key
		Container:   "media-uploads",
		CDNDomain:   "https://cdn.hackways.in",
	}

	signer := NewAzureBlobSASSigner(cfg)

	// 1. Success case: PNG banner upload
	req := UploadIntentRequest{
		Filename: "banner.png",
		MIMEType: "image/png",
		Size:     2 * 1024 * 1024, // 2 MB
		Category: "event-banners",
	}

	resp, err := signer.CreateUploadIntent(req)
	if err != nil {
		t.Fatalf("expected successful upload intent, got error: %v", err)
	}

	if !strings.HasPrefix(resp.BlobKey, "event-banners/") || !strings.HasSuffix(resp.BlobKey, ".png") {
		t.Errorf("unexpected blobKey format: %s", resp.BlobKey)
	}

	if !strings.HasPrefix(resp.PublicURL, "https://cdn.hackways.in/media-uploads/event-banners/") {
		t.Errorf("unexpected publicURL: %s", resp.PublicURL)
	}

	parsedURL, err := url.Parse(resp.UploadURL)
	if err != nil {
		t.Fatalf("failed to parse upload URL: %v", err)
	}

	query := parsedURL.Query()
	if query.Get("sp") != "w" {
		t.Errorf("expected write-only permission sp=w, got %s", query.Get("sp"))
	}
	if query.Get("spr") != "https" {
		t.Errorf("expected spr=https, got %s", query.Get("spr"))
	}
	if query.Get("sr") != "b" {
		t.Errorf("expected blob resource sr=b, got %s", query.Get("sr"))
	}
	if query.Get("sig") == "" {
		t.Errorf("expected non-empty HMAC signature sig")
	}

	expiresAt, err := time.Parse(time.RFC3339, resp.ExpiresAt)
	if err != nil {
		t.Fatalf("failed to parse expiresAt: %v", err)
	}
	remaining := time.Until(expiresAt)
	if remaining < 14*time.Minute || remaining > 16*time.Minute {
		t.Errorf("expected expiry in ~15 minutes, got %v", remaining)
	}

	// 2. Failure case: Disallowed MIME type
	_, err = signer.CreateUploadIntent(UploadIntentRequest{
		Filename: "payload.exe",
		MIMEType: "application/x-msdownload",
		Size:     1024,
	})
	if err != ErrInvalidFileType {
		t.Errorf("expected ErrInvalidFileType for exe, got %v", err)
	}

	// 3. Failure case: Disallowed file size (> 5MB)
	_, err = signer.CreateUploadIntent(UploadIntentRequest{
		Filename: "huge.jpg",
		MIMEType: "image/jpeg",
		Size:     6 * 1024 * 1024,
	})
	if err != ErrFileTooLarge {
		t.Errorf("expected ErrFileTooLarge for 6MB file, got %v", err)
	}

	// 4. Failure case: Empty filename
	_, err = signer.CreateUploadIntent(UploadIntentRequest{
		Filename: "",
		MIMEType: "image/webp",
		Size:     1024,
	})
	if err != ErrEmptyFilename {
		t.Errorf("expected ErrEmptyFilename, got %v", err)
	}
}
