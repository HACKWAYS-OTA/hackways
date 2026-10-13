package storage

import (
	"context"
)

// StorageService defines the interface for cloud media upload management.
// Under AGENTS.md Rule 2.3, persistent media must NEVER reside on the VM NVMe disk.
// All client media uploads are executed directly browser-to-blob via Write-Only SAS tokens.
type StorageService interface {
	CreateUploadIntent(req UploadIntentRequest) (*UploadIntentResponse, error)
	ConfirmUpload(ctx context.Context, blobKey string) (bool, error)
}
