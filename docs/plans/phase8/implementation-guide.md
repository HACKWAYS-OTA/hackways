# Phase 8 Implementation Guide: Production Deployment & Operational Readiness

## 1. Prerequisites
- Phase 7 complete (all test suites passing, CI green).
- Azure VM provisioned with Ubuntu 22.04 LTS / 24.04 LTS.
- Docker Engine v24+ and Docker Compose v2+ installed on Azure VM.
- Azure CLI (`az`) installed and authenticated via Managed Identity.

---

## 2. Step-by-Step Implementation

### Step 1: Author Production Docker Compose
1. Create `deploy/docker/docker-compose.prod.yml` matching the exact topology in `docs/ARCHITECTURE_PLAN.md` Section 8.
2. Define `edge_net` (external: false) and `backend_net` (internal: true).
3. Set CPU and memory resource limits for all 7 containers.

### Step 2: Author Nginx Ingress Configurations
1. Create `deploy/nginx/nginx.conf`:
   - Set `worker_processes 2;` and `worker_connections 4096;`.
   - Configure Gzip compression for JSON, JS, and CSS.
   - Define rate-limiting zones (`auth_zone`, `rsvp_zone`, `general_zone`).
2. Create `deploy/nginx/conf.d/`:
   - `public-web.conf` -> reverse proxies `hackways.in` to `public-web:3000`.
   - `core-web.conf` -> reverse proxies `app.hackways.in` to `core-web:3001`.
   - `organizer-web.conf` -> reverse proxies `organizer.hackways.in` to `organizer-web:3002`.
   - `admin-web.conf` -> reverse proxies `admin.hackways.in` to `admin-web:3003` with IP whitelisting.

### Step 3: Implement Automated Backup Script
1. Create `deploy/scripts/backup-db.sh`:
   ```bash
   #!/bin/bash
   set -euo pipefail
   TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
   BACKUP_NAME="hackways_db_${TIMESTAMP}.sql.gz"

   docker exec -t hackways_postgres pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists | gzip | \
   az storage blob upload \
     --account-name "$AZURE_STORAGE_ACCOUNT" \
     --account-key "$AZURE_STORAGE_KEY" \
     --container-name "backups" \
     --name "$BACKUP_NAME" \
     --type block \
     --data @-
   ```
2. Make executable: `chmod +x deploy/scripts/backup-db.sh`.

### Step 4: Configure Host NVMe Swap & Log Rotation
1. Run on host VM:
   ```bash
   sudo fallocate -l 3G /swapfile
   sudo chmod 600 /swapfile
   sudo mkswap /swapfile
   sudo swapon /swapfile
   echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
   ```
2. Configure `/etc/docker/daemon.json` with `max-size: "10m"`, `max-file: "3"`.

---

## 3. What Must NOT Be Changed in Phase 8
- Never compile Next.js images on the VM host (`npm run build` is strictly prohibited on the Azure VM).
- Never expose ports 8081, 8082, or 5432 directly to the public internet.
- Never store uncompressed database dumps on the 50 GB NVMe disk.
