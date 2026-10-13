# Phase 8 Technical Architecture: Production Ingress & VM Governance

## 1. Production Docker Network & Ingress Topology

```mermaid
flowchart TD
    subgraph Internet ["Public Internet & Cloudflare Edge"]
        PublicReq["Public HTTPS Traffic (Port 443)"]
        AdminReq["Admin Traffic (VPN / Static Office IP)"]
    end

    subgraph AzureHost ["Azure Host (Standard_B2s: 2 vCPU / 4 GB RAM / 50 GB NVMe)"]
        subgraph EdgeIngress ["Ingress (edge_net)"]
            Nginx["Nginx Reverse Proxy\n(Port 80/443)\nLimit: 75 MB RAM"]
        end

        subgraph EdgeWebServices ["Web Tier (edge_net & backend_net)"]
            PublicWeb["public-web:3000\nLimit: 140 MB"]
            CoreWeb["core-web:3001\nLimit: 250 MB"]
            OrgWeb["organizer-web:3002\nLimit: 230 MB"]
            AdminWeb["admin-web:3003\nLimit: 170 MB"]
        end

        subgraph BackendServices ["Backend Tier (backend_net internal only)"]
            CoreAPI["core-backend:8081\nLimit: 120 MB | CPU 0.70"]
            GlobalAPI["global-backend:8082\nLimit: 110 MB | CPU 0.40"]
            Postgres["PostgreSQL 16\nLimit: 700 MB | CPU 0.60"]
        end
    end

    PublicReq --> Nginx
    AdminReq -->|IP Whitelist| Nginx

    Nginx --> PublicWeb & CoreWeb & OrgWeb & AdminWeb
    Nginx -->|/api/v1/events, /api/v1/rsvps| CoreAPI
    Nginx -->|/api/v1/auth, /api/v1/storage| GlobalAPI

    CoreAPI --> Postgres
    GlobalAPI --> Postgres
```

---

## 2. Single-VM Failure Domain & Swap Protection
- Operating all 6 applications and PostgreSQL on a single VM means any memory leak could trigger an out-of-memory kernel panic.
- **Defensive Safeguards:**
  1. Every container enforces `deploy.resources.limits.memory`.
  2. The sum of all container memory caps is **2,495 MB**, leaving **1,601 MB** of physical RAM available for OS kernel, disk page cache, and buffer headroom.
  3. A **3.0 GB NVMe swapfile** with `vm.swappiness = 10` serves as a safety buffer against sudden spikes without causing premature disk thrashing.

---

## 3. Rate Limiting Zones (Nginx)
To shield the 2 vCPU VM from volumetric denial-of-service and brute force attacks:
- `auth_zone`: `5r/s` (burst=5 nodelay) on `/api/v1/auth/*`.
- `rsvp_zone`: `10r/s` (burst=5 nodelay) on `/api/v1/rsvps/*`.
- `general_zone`: `30r/s` (burst=20 nodelay) across all other API routes.
