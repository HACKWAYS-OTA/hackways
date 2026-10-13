# Phase 5 Verification & Evidence Plan: Azure Integration Foundations

## 1. Checks Required Before Implementation
- Confirm Phase 4 exit gate is verified.
- Confirm `services/api/internal/storage/store.go` no longer contains local file writes.

---

## 2. Mandatory Verification Protocol

### Check 1: Azure SAS URL Syntax & Permission Verification
- **Command:** `cd services/api && go test ./internal/storage -run TestGenerateUploadSAS -v`
- **Expected Outcome:** Emits valid Azure Blob URI containing signed parameters `sp=w`, `se=` (15-min timestamp), and `sig=`.

### Check 2: Upload Intent Validation & Rejection
- **Command:** `cd services/api && go test ./internal/storage -run TestUploadIntentValidation -v`
- **Expected Outcome:**
  - Files > 5 MB rejected with HTTP 400.
  - Non-image MIME types (e.g. `application/x-sh`, `text/html`) rejected with HTTP 400.
  - Image MIME types succeed with HTTP 200.

### Check 3: Zero Local Media Files on Disk
- **Inspection Command:**
  ```powershell
  Test-Path ./uploads
  ```
- **Expected Outcome:** Returns `False`. The `./uploads` folder must never exist.

### Check 4: Outbox Worker State Machine Verification
- **Command:** `cd services/api && go test ./internal/notifications -run TestOutboxWorker -v -race`
- **Expected Outcome:**
  - Pending emails are picked up and marked `sent`.
  - Simulated transient failures trigger exponential backoff.
  - 5 retries result in status marked `failed`.
  - Zero goroutine leaks detected.

---

## 3. Exit Criteria for Phase 5
- [ ] Direct Azure SAS generator operational.
- [ ] Local filesystem store driver completely purged.
- [ ] ACS email outbox worker polling `email_outbox` with exponential backoff.
- [ ] Zero local media files created on VM disk.
