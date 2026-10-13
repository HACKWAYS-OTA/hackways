# Phase 8 Task Register: Production Deployment & Operational Readiness

## Tasks Checklist

| Task ID | Title | Prerequisites | Status | Owner | Blockers / References |
|:---:|:---|:---:|:---:|:---|:---|
| `P8-OPS-001` | Create `deploy/docker/docker-compose.prod.yml` with memory caps | Phase 7 Exit Gate | `COMPLETED` | DevOps Lead | AD-006, 2.49GB cap |
| `P8-OPS-002` | Configure Nginx reverse proxy with rate limiting & admin IP filter | `P8-OPS-001` | `COMPLETED` | SRE Lead | `deploy/nginx/` |
| `P8-OPS-003` | Author direct-to-blob database backup script (`backup-db.sh`) | `P8-OPS-001` | `COMPLETED` | DevOps Lead | Azure CLI / Blob |
| `P8-OPS-004` | Author database disaster recovery restore script (`restore-db.sh`)| `P8-OPS-003` | `COMPLETED` | DevOps Lead | DR drill |
| `P8-OPS-005` | Configure Docker daemon log rotation policy (`max-size: 10m`) | `P8-OPS-001` | `COMPLETED` | SRE Lead | 210MB disk cap |
| `P8-OPS-006` | Configure host swap safety net (3.0 GB NVMe swapfile) | `P8-OPS-001` | `COMPLETED` | SRE Lead | `vm.swappiness=10` |
| `P8-OPS-007` | Execute production topology dry-run & syntax verification | All P8 Tasks | `COMPLETED` | SRE Lead | `docker compose config` |
| `P8-OPS-008` | Obtain project owner approval and compile final Phase 8 handoff | `P8-OPS-007` | `COMPLETED` | Tech Lead | Approval gate |

---

## Detailed Task Specifications

### `P8-OPS-001`: Docker Compose production topology
- **Description:** Author `deploy/docker/docker-compose.prod.yml` orchestrating:
  - `postgres`: 700 MB limit, 0.60 CPU.
  - `core-backend`: 120 MB limit, 0.70 CPU (`GOMEMLIMIT=100MiB`).
  - `global-backend`: 110 MB limit, 0.40 CPU (`GOMEMLIMIT=90MiB`).
  - `public-web`: 140 MB limit, 0.25 CPU.
  - `core-web`: 250 MB limit, 0.50 CPU.
  - `organizer-web`: 230 MB limit, 0.40 CPU.
  - `admin-web`: 170 MB limit, 0.25 CPU.
  - `nginx`: 75 MB limit, 0.20 CPU.
- **Acceptance Criteria:** Compose syntax validates cleanly; sum of limits <= 2,495 MB.
- **Status:** `NOT_STARTED`.

### `P8-OPS-002`: Nginx reverse proxy configuration
- **Description:** Implement `deploy/nginx/nginx.conf` with:
  - Rate limiting zones: `auth_limit` (5r/s), `rsvp_limit` (10r/s), `general_limit` (30r/s).
  - Subdomain routing: `hackways.in` -> :3000, `app.hackways.in` -> :3001, `organizer.hackways.in` -> :3002, `admin.hackways.in` -> :3003.
  - Admin access restriction: `allow <office_ip_or_vpn>; deny all;`.
- **Acceptance Criteria:** `nginx -t` passes.
- **Status:** `NOT_STARTED`.

### `P8-OPS-003`: Direct-to-blob backup script
- **Description:** Author `deploy/scripts/backup-db.sh` that pipes `pg_dump` through `gzip` directly to `az storage blob upload` targeting the private `backups` container. Zero local disk accumulation.
- **Acceptance Criteria:** Backup successfully created in Azure Blob without leaving local `.sql` or `.gz` files on the host.
- **Status:** `NOT_STARTED`.

### `P8-OPS-004`: Disaster recovery restore script
- **Description:** Author `deploy/scripts/restore-db.sh` downloading compressed backup from Azure Blob and piping into `psql`.
- **Acceptance Criteria:** Database restored cleanly in drill environment.
- **Status:** `NOT_STARTED`.

### `P8-OPS-005`: Docker log rotation
- **Description:** Configure `/etc/docker/daemon.json` with `"log-driver": "json-file"`, `"max-size": "10m"`, `"max-file": "3"`.
- **Acceptance Criteria:** Total container log footprint capped at 210 MB.
- **Status:** `NOT_STARTED`.

### `P8-OPS-006`: Host swap safety net
- **Description:** Provision 3.0 GB swapfile on NVMe with `vm.swappiness = 10` and `vm.vfs_cache_pressure = 50`.
- **Acceptance Criteria:** Swap verified active via `swapon --show`.
- **Status:** `NOT_STARTED`.

### `P8-OPS-007`: Dry run verification
- **Description:** Run `docker compose config` and verify container dependencies and health check commands.
- **Acceptance Criteria:** Zero configuration syntax errors.
- **Status:** `NOT_STARTED`.

### `P8-OPS-008`: Final handoff and approval
- **Description:** Compile final system verification and handoff documentation.
- **Acceptance Criteria:** Formal user approval obtained before initiating any live production launch.
- **Status:** `NOT_STARTED`.
