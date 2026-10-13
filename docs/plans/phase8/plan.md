# Phase 8 Implementation Plan: Production Deployment & Operational Readiness

## 1. Metadata
- **Phase Identifier:** Phase 8 (`phase8`)
- **Title:** Production Deployment and Operational Readiness
- **Owner Role:** DevOps Lead / Site Reliability Engineer (SRE)
- **Current Status:** `PLANNED`
- **Revision Date:** 2026-10-13

---

## 2. Purpose & Expected Outcome
Phase 8 establishes the complete production infrastructure, container orchestration, Nginx reverse proxy, automated backup pipelines, log rotation, and operational runbooks for deploying Hackways onto a single Azure Virtual Machine (**Standard_B2s or Standard_D2s_v5: 2 vCPU, 4 GB RAM, 50 GB NVMe SSD**).

**Expected Outcome:**
- `deploy/docker/docker-compose.prod.yml` configured with hard container memory caps totaling ~2.49 GB.
- `deploy/nginx/nginx.conf` configured with rate limiting zones, subdomain routing, and admin IP/VPN whitelisting.
- Automated database backup script (`deploy/scripts/backup-db.sh`) streaming compressed dumps directly to Azure Blob Storage `backups` container.
- Disaster recovery procedure (`deploy/scripts/restore-db.sh`) tested and verified.
- Docker daemon log rotation (`10m x 3`) configured to protect the 50 GB NVMe disk.

---

## 3. Scope & Non-Goals

### In Scope:
- Production Docker Compose topology for all 7 containers.
- Nginx reverse proxy with SSL termination, Gzip, and rate limits.
- Admin portal network isolation (IP/VPN whitelisting).
- Automated database backup and restore scripts.
- Host virtual memory tuning (`/swapfile` 3.0 GB, `vm.swappiness = 10`).

### Explicit Non-Goals:
- Building frontend or backend images directly on the production Azure VM host ([`AGENTS.md` Rule 2.2](../../AGENTS.md#L87-L92)).
- Modifying production DNS or running live traffic migrations without explicit user approval.
- Storing uncompressed database dumps on the VM disk.

---

## 4. Architectural Requirements & Constraints
- **Resource Budget:** Total container memory limit must not exceed **2,495 MB**, leaving at least **1.5 GB** free for Linux page cache and buffer spikes ([`AGENTS.md` Rule 2.2](../../AGENTS.md#L71-L85)).
- **Log Footprint:** Container logs must not exceed **210 MB total** across all containers.
- **Admin Ingress:** `admin.hackways.in` must be restricted to approved static IPs/VPNs at the Nginx layer.

---

## 5. Prerequisites & Dependencies
- **Prerequisite Phase:** Phase 7 (Testing & CI Enforcement complete; pre-built images available).
- **Prerequisite Infrastructure:** Azure VM provisioned with Azure Key Vault, Blob Storage, and ACS Email accounts.

---

## 6. Ordered Implementation Steps
1. **Step 1 (Docker Compose Production Topology):** Author `deploy/docker/docker-compose.prod.yml` specifying container memory and CPU limits.
2. **Step 2 (Nginx Ingress Configuration):** Author `deploy/nginx/nginx.conf` and subdomain routing configs in `deploy/nginx/conf.d/`.
3. **Step 3 (Backup & Restore Scripts):** Author `deploy/scripts/backup-db.sh` and `restore-db.sh`.
4. **Step 4 (Host Hardening Runbook):** Document swapfile configuration (3 GB NVMe) and Docker log rotation daemon settings.
5. **Step 5 (Simulation & Dry Run):** Validate Docker Compose syntax and Nginx config test (`nginx -t`).

---

## 7. Expected File Changes
- `deploy/docker/docker-compose.prod.yml`
- `deploy/nginx/nginx.conf`
- `deploy/nginx/conf.d/` (subdomain virtual hosts)
- `deploy/scripts/backup-db.sh`
- `deploy/scripts/restore-db.sh`
- `docs/DEPLOYMENT.md` (updated)

---

## 8. Verification Commands & Evidence
- `docker compose -f deploy/docker/docker-compose.prod.yml config` -> Valid syntax (exit code 0).
- `nginx -t -c deploy/nginx/nginx.conf` -> Syntax OK.
- Total memory limit calculation verified <= 2,495 MB.

---

## 9. Acceptance Criteria & Phase Exit Conditions
- [ ] Production Docker Compose file defines memory caps for all 7 containers.
- [ ] Nginx config enforces rate limiting and admin network isolation.
- [ ] Database backup streams directly to Azure Blob without local disk usage.
- [ ] Restore procedure documented and verified.
- [ ] Explicit user approval obtained prior to any live production deployment.
