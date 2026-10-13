# Canonical Deployment & Operations Guide — Hackways

This document is the authoritative operations specification for deploying the Hackways platform onto an Azure Virtual Machine, enforcing the memory governance, networking, and cloud persistence rules defined in [`AGENTS.md`](../AGENTS.md).

---

## 1. Production Architecture Topology

```
                              ┌──────────────────────────────────────────────┐
                              │ Cloudflare Edge CDN / Azure DNS (Port 443)   │
                              └──────────────────────┬───────────────────────┘
                                                     │ HTTPS (TLS 1.3 / HTTP/2)
                              ┌──────────────────────▼───────────────────────┐
                              │ Nginx Reverse Proxy (Port 80/443, 75 MB cap) │
                              │ Rate Limits: Auth 5r/s | RSVP 10r/s | Gen 30r│
                              └──────┬──────────┬──────────┬──────────┬──────┘
                                     │          │          │          │
         hackways.in (Port 3000)     │          │          │          │ admin.hackways.in (Port 3003)
      ┌──────────────────────────────▼─┐        │          │       ┌──▼──────────────────────────┐
      │ Public Web (140 MB cap)        │        │          │       │ Admin Web (170 MB cap)      │
      │ Next.js 14 SSG/ISR Marketing   │        │          │       │ VPN/Static IP Allowlist Only│
      └────────────────────────────────┘        │          │       └─────────────────────────────┘
                                                │          │
                 app.hackways.in (Port 3001)    │          │ organizer.hackways.in (Port 3002)
              ┌─────────────────────────────────▼┐        ┌▼────────────────────────────────┐
              │ Core Web (250 MB cap)            │        │ Organizer Web (230 MB cap)      │
              │ Next.js 14 Attendee App & RSVP   │        │ Next.js 14 Event Studio & Roster│
              └────────────────┬─────────────────┘        └─────────────────┬───────────────┘
                               │                                            │
                               └──────────────────────┬─────────────────────┘
                                                      │ REST APIs
                         ┌────────────────────────────┴───────────────────────────┐
                         │                                                        │
          /api/v1/events, /api/v1/rsvps                            /api/v1/auth, /api/v1/storage
                         │                                                        │
           ┌─────────────▼────────────────┐                        ┌──────────────▼──────────────┐
           │ Core Backend (Go 1.23 Engine)│                        │ Global Backend (Go Engine)  │
           │ Port 8081 | 120 MB RAM cap   │                        │ Port 8082 | 110 MB RAM cap  │
           │ Concurrency Locks & Capacity │                        │ Sessions, SAS Signer, Worker│
           └─────────────┬────────────────┘                        └──────────────┬──────────────┘
                         │                                                        │
                         │ pgxpool (Max 28 conns)                  pgxpool (Max 22 conns)
                         └────────────────────────────┬───────────────────────────┘
                                                      │
                                        ┌─────────────▼──────────────┐
                                        │ PostgreSQL 16 (700 MB cap) │
                                        │ Max 60 conns, 2000ms locks │
                                        └────────────────────────────┘
```

---

## 2. Azure VM Resource Budget (2 vCPU, 4.0 GB RAM, 50 GB NVMe)

The host environment runs on a single Azure Virtual Machine (`Standard_B2s` or `Standard_D2s_v5`). All containers enforce strict resource caps in `deploy/docker/docker-compose.prod.yml`:

| Component | Resident RSS | Docker Hard Limit | CPU Quota (Max 2.0) | Memory Governance Strategy |
| :--- | :--- | :--- | :--- | :--- |
| **PostgreSQL 16** | ~220 MB | **700 MB** | 0.60 | `shared_buffers = 256MB`, `work_mem = 4MB`, `max_connections = 60` |
| **Core Backend (Go)** | ~35 MB | **120 MB** | 0.70 | `GOMEMLIMIT=100MiB`, `GOGC=100` |
| **Global Backend (Go)** | ~35 MB | **110 MB** | 0.40 | `GOMEMLIMIT=90MiB`, `GOGC=100`, bounded 2-routine worker |
| **Core Web (Next.js)** | ~110 MB | **250 MB** | 0.50 | `NODE_OPTIONS="--max-old-space-size=190"`, standalone output |
| **Organizer Web (Next.js)**| ~100 MB | **230 MB** | 0.40 | `NODE_OPTIONS="--max-old-space-size=180"`, standalone output |
| **Admin Web (Next.js)** | ~80 MB | **170 MB** | 0.25 | `NODE_OPTIONS="--max-old-space-size=130"`, standalone output |
| **Public Web (Next.js)** | ~60 MB | **140 MB** | 0.25 | `NODE_OPTIONS="--max-old-space-size=100"`, standalone output |
| **Nginx Ingress** | ~25 MB | **75 MB** | 0.20 | Event loop worker processes: 2, worker_connections: 4096 |
| **Linux Host & System** | ~350 MB | ~500 MB | 0.20 | Systemd, journald, Docker daemon |
| **Ceiling Total** | **~1,215 MB** | **~2,495 MB** | **~3.50 (Burstable)** | **Safe Headroom: ~1,600 MB for disk page cache & spikes** |

---

## 3. Host Hardening & Disk Protection

### 3.1 NVMe Swap Safety Net (3.0 GB)
Run `deploy/scripts/setup-host-swap.sh` to provision an emergency NVMe buffer:
```bash
fallocate -l 3G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo "/swapfile none swap sw 0 0" >> /etc/fstab
```
Kernel settings in `/etc/sysctl.d/99-hackways-memory.conf`:
- `vm.swappiness = 10`
- `vm.vfs_cache_pressure = 50`
- `vm.dirty_ratio = 10`

### 3.2 Docker Daemon Log Rotation (210 MB Ceiling)
Configured in `/etc/docker/daemon.json` (or `deploy/docker/daemon.json`):
```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```
Ensures container logs across all 7 services never exceed 210 MB on the 50 GB NVMe disk.

### 3.3 Zero-Disk Upload Architecture
- No media files touch the VM disk.
- All uploads use 15-minute write-only SAS tokens generated by `global-backend` via HMAC-SHA256.
- Database backups stream directly to the private `backups` Azure Blob Storage container via `deploy/scripts/backup-db.sh`.

---

## 4. Disaster Recovery & Operations Runbooks

### 4.1 Automated Database Backup (Daily Cron)
Run daily via cron:
```bash
0 2 * * * /opt/hackways/deploy/scripts/backup-db.sh >> /var/log/hackways-backup.log 2>&1
```

### 4.2 Disaster Recovery Restore
In the event of database failure or corrupted state:
```bash
/opt/hackways/deploy/scripts/restore-db.sh hackways_db_YYYYMMDD_HHMMSSZ.sql.gz
```

### 4.3 Off-Node Next.js Compilation Policy
**Strict Rule:** Never run `npm run build` or `next build` on the Azure VM host.
All container images must be built in GitHub Actions CI with `output: 'standalone'` and pushed to Azure Container Registry (`hackways.azurecr.io`). The VM only pulls pre-built lightweight runtimes.
