# Phase 10 Architecture Specification: Azure Managed Services & High-Scalability Single-VM Blueprint

## 1. Architectural Philosophy

By offloading persistence and external I/O to native Azure managed cloud services (**Azure Database for PostgreSQL Flexible Server**, **Azure Blob Storage**, and **Azure Communication Services**), the production host (**Azure VM: 2 vCPU, 4 GB RAM, 50 GB NVMe**) is transformed into a pure, high-throughput compute and routing layer.

```
┌────────────────────────────────────────────────────────────────────────┐
│                        Cloudflare Edge CDN / WAF                       │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ HTTPS (Port 443 / TLS 1.3)
                        ┌───────────▼───────────┐
                        │  Nginx Ingress Proxy  │ Rate limiting & Microcaching
                        └─────┬───────────┬─────┘
                              │           │
            ┌─────────────────┴─┐       ┌─┴─────────────────┐
            │   Web Frontends   │       │  Backend Engines  │
            │(Next.js Standalone│       │(Go 1.23 + Chi)    │
            └───────┬───────────┘       └─────────┬─────────┘
                    │                             │
    ┌───────────────┼───────────────┐     ┌───────┴───────┐
    ▼               ▼               ▼     ▼               ▼
Public Web      Core Web       Org Web  Core API      Global API
(hackways.in)   (app.hackways) (organizer)(:8081)      (:8082)
                                          │               │
                                          │               │
                    ┌─────────────────────┴───────────────┼───────────────────────────────┐
                    │                                     │                               │
                    ▼                                     ▼                               ▼
       ┌────────────────────────┐             ┌────────────────────────┐      ┌────────────────────────┐
       │   Azure Database for   │             │   Azure Blob Storage   │      │  Azure Communication   │
       │   PostgreSQL Flexible  │             │   (Zero-Disk Storage)  │      │  Services (ACS) Email  │
       │ (Multi-AZ, PITR, SSL)  │             │  (Write-Only SAS 15m)  │      │  (Transactional Outbox)│
       └────────────────────────┘             └────────────────────────┘      └────────────────────────┘
```

---

## 2. Resource Reallocation on the 4 GB VM

With local PostgreSQL removed, the 4 GB RAM budget is rebalanced for extreme resilience:

| Container | Hard Limit | GOMEMLIMIT / Node Memory | Connection Pool | Scalability Benefit |
| :--- | :--- | :--- | :--- | :--- |
| **core-backend** | **180 MB** | `150MiB` | Max 32, Min 4 | 50% more memory headroom for high-concurrency RSVP registration locks. |
| **global-backend** | **140 MB** | `120MiB` | Max 24, Min 3 | Faster outbox worker polling & direct Azure SAS token signing. |
| **core-web** | **260 MB** | `--max-old-space-size=190` | N/A | Smooth SSR rendering under attendee spikes. |
| **organizer-web** | **240 MB** | `--max-old-space-size=180` | N/A | Rapid roster pagination and event editing. |
| **admin-web** | **170 MB** | `--max-old-space-size=130` | N/A | Network-isolated administration. |
| **public-web** | **140 MB** | `--max-old-space-size=100` | N/A | Static pre-rendering and marketing delivery. |
| **nginx** | **80 MB** | N/A | 4096 conns | Microcaching and rate-limiting enforcement. |
| **Ceiling Total** | **~1,210 MB** | — | — | **~2,430 MB Free Headroom on Host (59% of RAM)** |
