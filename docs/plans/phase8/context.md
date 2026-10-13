# Phase 8 Agent Pre-Context: Production Deployment & Operational Readiness

## 1. Product Objective & Technical Target
Phase 8 establishes the complete production operational environment on an Azure Virtual Machine (**Standard_B2s or Standard_D2s_v5: 2 vCPU, 4 GB RAM, 50 GB NVMe SSD**). It ensures the 6-service topology operates reliably within strict memory and disk ceilings.

---

## 2. Canonical Specifications to Obey
1. **[`AGENTS.md` Rule 1.7 & Rule 2.2](../../../AGENTS.md#L60-L85):**
   - Respect Azure VM limits: 2 vCPU, 4 GB RAM, 50 GB NVMe disk.
   - **Never run Next.js compilation or `npm run build` on the VM host.**
   - Total container Docker memory limit must not exceed **2,495 MB**.
   - NVMe Swap Safety Net: 3.0 GB swapfile with `vm.swappiness = 10`.
   - Docker log rotation: max 10MB x 3 files (max 210 MB total).
2. **[`AGENTS.md` Rule 2.6](../../../AGENTS.md#L217-L232):**
   - Nginx is the sole public HTTP/HTTPS port (80/443). Internal backend ports (8081, 8082) and PostgreSQL (5432) must never be exposed publicly.
   - `admin.hackways.in` must be restricted to approved office static IPs and VPN subnets (`allow <subnet>; deny all;`).

---

## 3. Verified Repository Facts Before Phase 8
- Root `docker-compose.yml` is configured for local development (Postgres + Mailpit).
- The production deployment directory `deploy/` does not yet exist on disk and must be created in this phase.

---

## 4. Expected Outputs
- `deploy/docker/docker-compose.prod.yml` orchestrating all 7 containers with memory limits.
- `deploy/nginx/nginx.conf` with rate limiting and subdomain routing.
- `deploy/scripts/backup-db.sh` streaming compressed backups to Azure Blob Storage.
