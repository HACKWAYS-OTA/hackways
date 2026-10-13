# Phase 5 Technical Architecture: Azure Managed Services Integration

## 1. Direct SAS Upload Flow Architecture

```mermaid
sequenceDiagram
    autonumber
    actor Client as Organizer Browser
    participant API as Global Backend (:8082)
    participant AzureBlob as Azure Blob Storage
    participant CDN as Edge CDN / Cloudflare

    Client->>API: POST /api/v1/storage/upload-intent<br/>{fileName: "banner.png", fileSize: 2450000, mimeType: "image/png"}
    Note over API: Validates image MIME type<br/>Verifies size <= 5 MB<br/>Generates UUID blob key
    API->>API: Sign User Delegation SAS (15 mins, Perm: Write)
    API-->>Client: 200 OK {uploadUrl: "https://...blob.core.windows.net/media-uploads/...", blobKey: "events/banner-uuid.png"}
    
    Client->>AzureBlob: PUT https://...blob.core.windows.net/... (Direct Binary Payload)
    AzureBlob-->>Client: 201 Created (ETag emitted)

    Client->>API: POST /api/v1/events/{id}/banner {blobKey: "events/banner-uuid.png"}
    Note over API: Performs HEAD request to verify blob exists<br/>Saves blob URL to PostgreSQL
    API-->>Client: 200 OK (Event Banner Linked)

    actor PublicUser as Attendee Browser
    PublicUser->>CDN: GET /events/banner-uuid.png
    CDN->>AzureBlob: Fetch from public blob container
    CDN-->>PublicUser: Cached 200 OK (Zero load on VM)
```

---

## 2. PostgreSQL Transactional Outbox Pattern

```mermaid
sequenceDiagram
    autonumber
    participant CoreAPI as Core Backend (:8081)
    participant DB as PostgreSQL (email_outbox)
    participant Worker as Outbox Worker (:8082)
    participant ACS as Azure Communication Services

    Note over CoreAPI,DB: Inside Transaction confirming RSVP
    CoreAPI->>DB: INSERT INTO email_outbox (recipient_email, template_type, payload_json, status)<br/>VALUES ($1, 'rsvp_confirmation', $2, 'pending')
    Note over CoreAPI,DB: Transaction Committed

    loop Polling Loop (Every 3-5 seconds)
        Worker->>DB: SELECT id, recipient_email, payload_json<br/>FROM email_outbox WHERE status = 'pending' AND scheduled_at <= NOW()<br/>ORDER BY created_at ASC LIMIT 10 FOR UPDATE SKIP LOCKED
        Worker->>ACS: POST /emails:send (JSON Payload)
        alt Success (202 Accepted)
            Worker->>DB: UPDATE email_outbox SET status = 'sent', sent_at = NOW() WHERE id = $1
        else Transient Failure (5xx / Timeout)
            Worker->>DB: UPDATE email_outbox SET retry_count = retry_count + 1,<br/>scheduled_at = NOW() + (2^retry_count * 15s) WHERE id = $1
        else Terminal Failure (5 retries reached)
            Worker->>DB: UPDATE email_outbox SET status = 'failed' WHERE id = $1
        end
    end
```

---

## 3. Worker Concurrency & Memory Budget
- The outbox worker runs strictly within `global-backend` as a bounded pool of **2 goroutines**.
- Polling uses `LIMIT 10 FOR UPDATE SKIP LOCKED` to prevent query thrashing and allow seamless multi-instance scaling in the future.
- Total worker RSS memory is capped at **< 10 MB**, well within the 110 MB container ceiling for `global-backend`.
