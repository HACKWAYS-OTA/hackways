package storage

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

var (
	ErrInvalidFileType = errors.New("unsupported file type: only image/jpeg, image/png, and image/webp are allowed")
	ErrFileTooLarge    = errors.New("file size exceeds 5MB limit")
	ErrEmptyFilename   = errors.New("filename cannot be empty")
)

const (
	MaxUploadSizeBytes = 5 * 1024 * 1024  // 5 MB
	SASExpiryDuration  = 15 * time.Minute // 15 minutes as per AGENTS.md Rule 2.3
	AzureAPIVersion    = "2020-12-06"
)

// AllowedMIMETypes maps permitted MIME types to file extensions.
var AllowedMIMETypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// AzureSASConfig contains credentials and configuration for Azure Blob Storage.
type AzureSASConfig struct {
	AccountName string
	AccountKey  string
	Container   string
	CDNDomain   string
}

// AzureBlobSASSigner generates direct client-to-blob Write-Only SAS upload tokens.
type AzureBlobSASSigner struct {
	cfg AzureSASConfig
}

// NewAzureBlobSASSigner creates a new AzureBlobSASSigner.
func NewAzureBlobSASSigner(cfg AzureSASConfig) *AzureBlobSASSigner {
	return &AzureBlobSASSigner{cfg: cfg}
}

// UploadIntentRequest represents a client request to initiate a media upload.
type UploadIntentRequest struct {
	Filename string `json:"filename"`
	MIMEType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Category string `json:"category"` // e.g. "event-banners", "avatars"
}

// UploadIntentResponse returns the direct Azure Blob SAS upload URL and public read URL.
type UploadIntentResponse struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
	BlobKey   string `json:"blob_key"`
	ExpiresAt string `json:"expires_at"`
}

// CreateUploadIntent validates upload constraints and issues a 15-minute write-only SAS token.
func (s *AzureBlobSASSigner) CreateUploadIntent(req UploadIntentRequest) (*UploadIntentResponse, error) {
	ext, ok := AllowedMIMETypes[strings.ToLower(req.MIMEType)]
	if !ok {
		return nil, ErrInvalidFileType
	}

	if req.Size <= 0 || req.Size > MaxUploadSizeBytes {
		return nil, ErrFileTooLarge
	}

	if req.Filename == "" {
		return nil, ErrEmptyFilename
	}

	// Generate UUID / random hex blob key to prevent collisions and path traversals
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, fmt.Errorf("failed to generate blob key: %w", err)
	}

	category := strings.Trim(req.Category, "/")
	if category == "" {
		category = "uploads"
	}
	blobKey := fmt.Sprintf("%s/%s%s", category, hex.EncodeToString(randBytes), ext)

	now := time.Now().UTC()
	start := now.Add(-5 * time.Minute) // 5m clock skew allowance
	expiry := now.Add(SASExpiryDuration)

	permissions := "w"
	if strings.HasPrefix(category, "event-banners/") || strings.HasPrefix(category, "space-logos/") || strings.HasPrefix(category, "space-banners/") || strings.HasPrefix(category, "city-banners/") {
		permissions = "c"
	}
	sasToken, err := s.generateBlobSAS(blobKey, start, expiry, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SAS token: %w", err)
	}

	uploadURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s",
		s.cfg.AccountName,
		s.cfg.Container,
		blobKey,
		sasToken,
	)

	publicBase := s.cfg.CDNDomain
	if publicBase == "" {
		publicBase = fmt.Sprintf("https://%s.blob.core.windows.net", s.cfg.AccountName)
	}
	publicURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(publicBase, "/"), s.cfg.Container, blobKey)

	return &UploadIntentResponse{
		UploadURL: uploadURL,
		PublicURL: publicURL,
		BlobKey:   blobKey,
		ExpiresAt: expiry.Format(time.RFC3339),
	}, nil
}

// generateBlobWriteSAS signs a Write-Only SAS query string using Azure Storage Shared Key HMAC-SHA256.
func (s *AzureBlobSASSigner) generateBlobWriteSAS(blobName string, start, expiry time.Time) (string, error) {
	return s.generateBlobSAS(blobName, start, expiry, "w")
}

func (s *AzureBlobSASSigner) generateBlobSAS(blobName string, start, expiry time.Time, permissions string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(s.cfg.AccountKey)
	if err != nil {
		return "", fmt.Errorf("invalid storage signing key encoding")
	}

	signedResource := "b" // blob
	startTimeStr := start.Format("2006-01-02T15:04:05Z")
	expiryTimeStr := expiry.Format("2006-01-02T15:04:05Z")
	canonicalizedResource := fmt.Sprintf("/blob/%s/%s/%s", s.cfg.AccountName, s.cfg.Container, path.Clean(blobName))

	// Azure Blob SAS StringToSign structure for version 2020-12-06
	stringToSign := strings.Join([]string{
		permissions,
		startTimeStr,
		expiryTimeStr,
		canonicalizedResource,
		"",      // signedIdentifier
		"",      // signedIP
		"https", // signedProtocol
		AzureAPIVersion,
		signedResource,
		"", // signedSnapshotTime
		"", // signedEncryptionScope
		"", // rscc
		"", // rscd
		"", // rsce
		"", // rscl
		"", // rsct
	}, "\n")

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	params := url.Values{}
	params.Set("sp", permissions)
	params.Set("st", startTimeStr)
	params.Set("se", expiryTimeStr)
	params.Set("spr", "https")
	params.Set("sv", AzureAPIVersion)
	params.Set("sr", signedResource)
	params.Set("sig", signature)

	return params.Encode(), nil
}
