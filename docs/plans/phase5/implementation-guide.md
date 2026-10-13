# Phase 5 Implementation Guide: Azure Integration Foundations

## 1. Prerequisites
- Phase 4 complete (Auth middleware available).
- Azure SDK dependencies: `github.com/Azure/azure-sdk-for-go/sdk/storage/azblob`.
- Local or test Azure Storage emulator / connection credentials.

---

## 2. Step-by-Step Implementation

### Step 1: Purge LocalStore & Create Azure SAS Signer
1. Delete `LocalStore` implementation from `services/api/internal/storage/store.go`.
2. Create `services/api/internal/storage/azure_sas.go`:
   - Initialize Azure Blob client using account key or Managed Identity.
   - Implement `GenerateUploadSAS`:
     - Scope permissions to write-only (`sp=w`).
     - Set expiry: `time.Now().Add(15 * time.Minute)`.
     - Set container: `media-uploads`.
     - Return full pre-signed upload URL.

### Step 2: Implement Upload Intent Endpoint
1. Create HTTP handler for `POST /api/v1/storage/upload-intent`:
   - Enforce authentication via `auth.RequireAuth`.
   - Validate `mimeType` in `["image/jpeg", "image/png", "image/webp"]`.
   - Validate `fileSize <= 5242880` (5 MB).
   - Generate UUID blob key: `events/{id}/covers/{uuid}.{ext}`.
   - Return `{ uploadUrl, blobKey }`.

### Step 3: Implement ACS Email Outbox Worker
1. Create `services/api/internal/notifications/outbox_worker.go`:
   - Define `OutboxWorker` struct accepting `*pgxpool.Pool` and ACS HTTP client.
   - Start 2 worker goroutines with graceful shutdown context.
   - Query pending rows using `FOR UPDATE SKIP LOCKED`.
   - Format email template and call ACS REST endpoint.
   - Handle exponential backoff: `2^retry_count * 15 * time.Second`.
   - Mark terminal failures when `retry_count >= 5`.

---

## 3. What Must NOT Be Changed in Phase 5
- Never store media files in a local directory (`./uploads` or `/tmp/uploads`).
- Never buffer the multi-megabyte media payload through Go RAM.
- Never add local SMTP daemons (Postfix, Exim, Sendmail) or Redis Celery queues to the VM host.
