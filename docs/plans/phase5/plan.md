# Phase 5 Implementation Plan: Azure Integration Foundations

## 1. Metadata
- **Phase Identifier:** Phase 5 (`phase5`)
- **Title:** Azure Integration Foundations
- **Owner Role:** Cloud Architect / Backend Engineer
- **Current Status:** `PLANNED`
- **Revision Date:** 2026-10-13

---

## 2. Purpose & Expected Outcome
Phase 5 implements direct client-to-blob media uploads via Azure Blob Storage SAS tokens and reliable transactional email delivery via Azure Communication Services (ACS) Email through the PostgreSQL outbox pattern. It eliminates the prohibited local disk upload driver (`services/api/internal/storage/store.go`) and local SMTP dependencies.

**Expected Outcome:**
- `LocalStore` replaced with `AzureBlobSASSigner` issuing 15-minute write-only SAS URLs.
- Zero user media stored on or proxied through the Azure VM host.
- Transactional email worker in `global-backend` (2 goroutines) polling `email_outbox` and dispatching via ACS REST API / Go SDK.
- Exponential backoff retry logic (up to 5 retries) and dead-letter handling active.

---

## 3. Scope & Non-Goals

### In Scope:
- Azure Blob Storage SAS token generation endpoint (`POST /api/v1/storage/upload-intent`).
- MIME type inspection and size validation (max 5 MB).
- Purging local filesystem storage driver (`./uploads`).
- ACS Email dispatch worker polling `email_outbox` table with `SKIP LOCKED`.
- Exponential backoff retry handler: `scheduled_at = NOW() + (2^retry_count * 15s)`.

### Explicit Non-Goals:
- Storing uploads on VM local disk.
- Running local SMTP daemons or Redis queue workers.
- Building frontend file upload UI components (Phase 6).

---

## 4. Architectural Requirements & Constraints
- **Zero-Disk Upload Architecture:** Uploaded media must never be buffered or proxied through Go RAM or VM disk ([`AGENTS.md` Rule 2.3](../../AGENTS.md#L125-L135)).
- **Email Delivery:** Transactional emails must use the PostgreSQL Transactional Outbox Pattern ([`AGENTS.md` Rule 2.3](../../AGENTS.md#L143-L158)).
- **Memory Ceiling:** Email worker must run in a bounded 2-goroutine pool consuming < 10 MB RAM.

---

## 5. Prerequisites & Dependencies
- **Prerequisite Phase:** Phase 4 (Security Foundations for authenticated upload requests).
- **Prerequisite Decisions:** AD-008 (Azure Blob SAS upload) and AD-009 (ACS Email outbox).
- **External Dependencies:** Azure Blob Storage account credentials/Managed Identity and Azure Communication Services endpoint.

---

## 6. Ordered Implementation Steps
1. **Step 1 (Purge Local Disk Storage):** Deprecate and remove `LocalStore` in `services/api/internal/storage/store.go`.
2. **Step 2 (Azure Blob SAS Signer):** Implement `storage/azure_sas.go` generating Write-Only User Delegation SAS tokens (15-min expiry).
3. **Step 3 (Upload Intent Endpoint):** Implement `POST /api/v1/storage/upload-intent` validating MIME types and emitting SAS URLs.
4. **Step 4 (ACS Outbox Worker):** Implement background polling worker in `notifications/outbox_worker.go` processing `email_outbox`.
5. **Step 5 (Unit & Mock Tests):** Test SAS URL structure and outbox worker state transitions.

---

## 7. Expected File Changes
- `services/api/internal/storage/store.go` (purged/replaced)
- `services/api/internal/storage/azure_sas.go` & `azure_sas_test.go`
- `services/api/internal/notifications/email.go`
- `services/api/internal/notifications/outbox_worker.go` & `outbox_worker_test.go`
- `services/api/internal/platform/config.go` (Azure config fields)

---

## 8. Verification Commands & Evidence
- `cd services/api && go test ./internal/storage/... -v -count=1` -> PASS.
- `cd services/api && go test ./internal/notifications/... -v -count=1` -> PASS.
- Verification that `./uploads` directory is never created on disk.

---

## 9. Acceptance Criteria & Phase Exit Conditions
- [ ] Direct Azure Blob SAS generator operational.
- [ ] Local filesystem store driver completely purged.
- [ ] ACS email outbox worker polling `email_outbox` with exponential backoff.
- [ ] Memory footprint of background worker remains < 10 MB.
