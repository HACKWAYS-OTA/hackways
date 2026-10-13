# Phase 5 Agent Pre-Context: Azure Integration Foundations

## 1. Product Objective & Technical Target
Phase 5 integrates the Hackways backend with Azure Blob Storage and Azure Communication Services (ACS) Email. The core objective is **zero local persistent media on the VM disk** and **zero blocking external HTTP calls during the RSVP transaction cycle**.

---

## 2. Canonical Specifications to Obey
1. **[`AGENTS.md` Rule 2.3](../../../AGENTS.md#L125-L135):**
   - **NO PERSISTENT MEDIA ON LOCAL VM DISK.** The 50 GB NVMe disk is strictly reserved for OS, Docker images, and PostgreSQL data.
   - Uploads MUST use **direct client-to-blob upload via Write-Only SAS tokens** (15-min expiry).
   - Uploaded media MUST NEVER be proxied or buffered through Go application RAM.
2. **[`AGENTS.md` Rule 2.3](../../../AGENTS.md#L143-L158):**
   - **NO LOCAL SMTP CONTAINERS.**
   - Transactional emails MUST use the PostgreSQL Transactional Outbox Pattern with a bounded 2-goroutine worker in `global-backend`.
3. **[`AGENTS.md` Rule 3.5](../../../AGENTS.md#L240):**
   - Prohibited shortcut: Creating local `/uploads` folders on the VM.

---

## 3. Verified Repository Facts Before Phase 5
- Existing file [`services/api/internal/storage/store.go`](../../../services/api/internal/storage/store.go) currently implements `LocalStore` writing files to `./uploads`. **This must be completely purged.**
- Existing file [`services/api/internal/notifications/email.go`](../../../services/api/internal/notifications/email.go) implements a mock `LogEmailSender`.

---

## 4. Expected Outputs
- `AzureBlobSASSigner` emitting 15-minute write-only SAS tokens.
- Bounded 2-goroutine ACS email outbox worker.
- Complete removal of `LocalStore`.
