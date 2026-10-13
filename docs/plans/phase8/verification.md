# Phase 8 Verification & Evidence Plan: Production Deployment & Operational Readiness

## 1. Checks Required Before Implementation
- Confirm Phase 7 exit gate is verified.
- Confirm pre-built container images exist in registry.

---

## 2. Mandatory Verification Protocol

### Check 1: Production Compose Syntax & Resource Cap Audit
- **Command:**
  ```bash
  docker compose -f deploy/docker/docker-compose.prod.yml config
  ```
- **Expected Outcome:** Valid syntax with zero schema warnings.
- **Resource Sum Check:** Sum of all `deploy.resources.limits.memory` must equal 2,495 MB.

### Check 2: Nginx Configuration Test
- **Command:**
  ```bash
  nginx -t -c deploy/nginx/nginx.conf
  ```
- **Expected Outcome:** `syntax is ok` and `test is successful`.

### Check 3: Admin Network Restriction Test
- **Command:** Send curl request from non-whitelisted IP:
  ```bash
  curl -I https://admin.hackways.in
  ```
- **Expected Outcome:** HTTP 403 Forbidden emitted directly by Nginx.

### Check 4: Disaster Recovery Dry Run
- **Command:** Run `deploy/scripts/backup-db.sh`, verify blob upload in Azure Blob, then run `deploy/scripts/restore-db.sh` into drill database.
- **Expected Outcome:** 100% records restored cleanly.

### Check 5: Host Memory & Swap Verification
- **Commands:**
  ```bash
  free -m
  swapon --show
  ```
- **Expected Outcome:** Swap shows 3072 MB active. Memory buffer shows > 1.5 GB available.

---

## 3. Exit Criteria for Phase 8
- [ ] Production Compose file validates with ~2.49 GB total memory ceiling.
- [ ] Nginx rate limiting and admin isolation verified.
- [ ] Direct Azure Blob backup and restore pipeline tested and verified.
- [ ] Host swap and log rotation verified active.
- [ ] Formal project owner approval received for production launch.
