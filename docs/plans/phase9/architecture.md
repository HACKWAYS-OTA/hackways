# Phase 9 Architecture Specification: Dual Go Backend Engines

## 1. Domain & Runtime Boundaries

```mermaid
flowchart TD
    subgraph ClientLayer ["Client Layer"]
        PublicWeb["public-web (:3000)\n(hackways.in)"]
        CoreWeb["core-web (:3001)\n(app.hackways.in)"]
        OrgWeb["organizer-web (:3002)\n(organizer.hackways.in)"]
        AdminWeb["admin-web (:3003)\n(admin.hackways.in)"]
    end

    subgraph IngressLayer ["Ingress Layer (Nginx)"]
        Nginx["Nginx Reverse Proxy (:80 / :443)\n• SSL Termination\n• Rate Limiting Zones\n• Subdomain Routing"]
    end

    subgraph BackendEngines ["Dual Go Backend Engines"]
        subgraph CoreEngine ["services/core-backend (:8081)"]
            CoreRouter["Chi Router (:8081)"]
            EventsHandler["/api/v1/events/*"]
            RSVPHandler["/api/v1/rsvps/*\n(Pessimistic Lock Engine)"]
            CorePool["pgxpool (Max 28 Conns)"]
        end

        subgraph GlobalEngine ["services/global-backend (:8082)"]
            GlobalRouter["Chi Router (:8082)"]
            AuthHandler["/api/v1/auth/*\n(Register, Login, Session)"]
            StorageHandler["/api/v1/storage/*\n(Azure SAS Signer)"]
            OutboxWorker["Outbox Worker Pool\n(2 Goroutines)"]
            GlobalPool["pgxpool (Max 22 Conns)"]
        end
    end

    subgraph CloudPersistence ["Persistence & Cloud"]
        Postgres["PostgreSQL 16 (Max 60 Conns)"]
        AzureBlob["Azure Blob Storage\n(Direct Client Upload via SAS)"]
        ACSEmail["Azure Communication Services (ACS)\n(Transactional Outbox Dispatch)"]
    end

    ClientLayer --> Nginx
    Nginx -->|/api/v1/events, /api/v1/rsvps| CoreRouter
    Nginx -->|/api/v1/auth, /api/v1/storage| GlobalRouter

    CoreRouter --> EventsHandler
    CoreRouter --> RSVPHandler
    RSVPHandler --> CorePool

    GlobalRouter --> AuthHandler
    GlobalRouter --> StorageHandler
    AuthHandler --> GlobalPool
    StorageHandler --> AzureBlob
    OutboxWorker --> GlobalPool
    OutboxWorker --> ACSEmail

    CorePool --> Postgres
    GlobalPool --> Postgres
```

---

## 2. PostgreSQL Connection Pool Allocation
- Total Instance Cap: **60 connections**
- `services/core-backend`: **Max 28 connections** (Min 4, Idle 15m)
- `services/global-backend`: **Max 22 connections** (Min 3, Idle 15m)
- Admin, migrations, CLI: **10 connections reserved**

---

## 3. Memory & Resource Governance
- `core-backend`: `GOMEMLIMIT=100MiB`, `GOGC=100`, Docker memory cap **120 MB**, CPU quota 0.70.
- `global-backend`: `GOMEMLIMIT=90MiB`, `GOGC=100`, Docker memory cap **110 MB**, CPU quota 0.40.
