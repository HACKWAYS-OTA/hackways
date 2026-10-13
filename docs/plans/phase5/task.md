# Phase 5 Task Register: Azure Integration Foundations

## Tasks Checklist

| Task ID | Title | Prerequisites | Status | Owner | Blockers / References |
|:---:|:---|:---:|:---:|:---:|:---|
| `P5-AZURE-001` | Purge `LocalStore` local disk storage driver | Phase 4 Exit Gate | `COMPLETED` | Backend Lead | GAP-005, AD-008 |
| `P5-AZURE-002` | Implement Azure Blob SAS token generator (`azure_sas.go`) | `P5-AZURE-001` | `COMPLETED` | Cloud Lead | 15-min expiry |
| `P5-AZURE-003` | Implement `POST /api/v1/storage/upload-intent` endpoint logic | `P5-AZURE-002` | `COMPLETED` | Backend Lead | MIME validation |
| `P5-AZURE-004` | Implement blob confirmation & validation interface | `P5-AZURE-003` | `COMPLETED` | Backend Lead | StorageService |
| `P5-AZURE-005` | Implement PostgreSQL outbox poller with `SKIP LOCKED` | Phase 4 Exit Gate | `COMPLETED` | Backend Lead | AD-009 |
| `P5-AZURE-006` | Implement ACS email dispatcher with exponential backoff | `P5-AZURE-005` | `COMPLETED` | Cloud Lead | 2 goroutines |
| `P5-AZURE-007` | Implement unit and integration tests for SAS & outbox worker | `P5-AZURE-004`, `006` | `COMPLETED` | QA / Cloud | Automated tests |
| `P5-AZURE-008` | Execute Phase 5 exit gate verification | All P5 Tasks | `COMPLETED` | Tech Lead | Phase Gate 5 |

---

## Detailed Task Specifications & Verified Evidence

### `P5-AZURE-001`: Purge `LocalStore`
- **Description:** Delete local file writing logic in `services/api/internal/storage/store.go` that writes to `./uploads`.
- **Acceptance Criteria:** Zero filesystem write calls in storage package.
- **Status:** `COMPLETED`.
- **Evidence:** `store.go` refactored to `StorageService` interface with zero local disk I/O.

### `P5-AZURE-002`: Azure Blob SAS token generator
- **Description:** Implement `GenerateUploadSAS(blobName, container, contentType string) (string, error)` generating a Write-Only SAS URI valid for 15 minutes.
- **Acceptance Criteria:** Emits valid SAS query string containing signed permissions `sp=w` and expiration.
- **Status:** `COMPLETED`.
- **Evidence:** `internal/storage/azure_sas.go` implements HMAC-SHA256 signing for Azure Blob Storage with 15-minute write-only validity.

### `P5-AZURE-003`: `POST /api/v1/storage/upload-intent`
- **Description:** Endpoint validates payload (`fileName`, `fileSize <= 5MB`, `mimeType in [image/png, image/jpeg, image/webp]`), generates a UUID blob key, and returns SAS URL.
- **Acceptance Criteria:** Rejects unsupported MIME types with 400; returns 200 with `uploadUrl` and `blobKey`.
- **Status:** `COMPLETED`.
- **Evidence:** `CreateUploadIntent` validates MIME types, sizes <= 5MB, non-empty filename; tested in `azure_sas_test.go`.

### `P5-AZURE-004`: Blob confirmation validation
- **Description:** Validates uploaded blob before linking `banner_url` to an event.
- **Acceptance Criteria:** `StorageService` defines `ConfirmUpload` contract.
- **Status:** `COMPLETED`.
- **Evidence:** Defined in `internal/storage/store.go`.

### `P5-AZURE-005`: Outbox poller with `SKIP LOCKED`
- **Description:** Background goroutine executing: `SELECT id, recipient_email, recipient_name, template_type, payload_json FROM email_outbox WHERE status = 'pending' AND scheduled_at <= NOW() ORDER BY created_at ASC LIMIT 10 FOR UPDATE SKIP LOCKED`.
- **Acceptance Criteria:** Concurrent workers never pick the same outbox row.
- **Status:** `COMPLETED`.
- **Evidence:** Implemented in `internal/notifications/outbox_worker.go` via compile-time verified `FetchPendingOutbox` query.

### `P5-AZURE-006`: ACS email dispatcher
- **Description:** Dispatches email via Azure Communication Services / EmailSender. Updates status to `sent` on success; increments `retry_count` and applies backoff on failure; sets status to `failed` after 5 retries.
- **Acceptance Criteria:** State transitions verified under simulated failure and success.
- **Status:** `COMPLETED`.
- **Evidence:** Implemented in `internal/notifications/outbox_worker.go` with 2-goroutine semaphore and exponential backoff (`15s * 2^retry`).

### `P5-AZURE-007`: Test suite
- **Description:** Automated tests verifying SAS token generation and outbox worker state machine.
- **Acceptance Criteria:** 100% tests pass.
- **Status:** `COMPLETED`.
- **Evidence:** All tests pass in `storage/azure_sas_test.go` and `notifications/outbox_worker_test.go` against live PostgreSQL.

### `P5-AZURE-008`: Phase 5 exit gate verification
- **Description:** Verify all criteria in `docs/plans/phase5/verification.md`.
- **Acceptance Criteria:** Exit gate approved; Phase 5 marked `VERIFIED`.
- **Status:** `COMPLETED`.
- **Evidence:** All criteria verified.
