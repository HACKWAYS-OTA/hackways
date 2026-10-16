# Phase 10 Task Breakdown: Azure Managed Cloud Services Migration & Scalability Hardening

## Task Matrix

| Task ID | Task Title | Owner | Target Artifacts | Acceptance Criteria |
|:---:|:---|:---:|:---|:---|
| `P10-001` | Production Compose & Configuration Refactor | Cloud Lead | `deploy/docker/docker-compose.prod.yml`, `.env.example` | Local postgres container removed; connection string pointed to Azure PostgreSQL Flexible Server with `sslmode=require`; Go backend memory limits elevated. |
| `P10-002` | Azure PostgreSQL Connection & Migration Alignment | DB Engineer | `deploy/migrations/`, `docs/DATABASE.md` | Connection pool configured for Azure PostgreSQL; schema migrations applied cleanly; `SET LOCAL statement_timeout = '2000ms'` verified. |
| `P10-003` | Off-Node CI/CD Deployment Script Update | SRE Lead | `deploy/scripts/deploy.sh` | On-node `docker build` removed; `deploy.sh` pulls pre-built standalone images from GHCR; zero CPU spikes during deploy. |
| `P10-004` | Ingress Microcaching & Rate Limiting Verification | Ingress Lead | `deploy/nginx/conf.d/*.conf`, `deploy/nginx/nginx.conf` | Public read endpoints (`/api/v1/events`) microcached; rate limits enforced (`auth` 5r/s, `rsvps` 10r/s). |
| `P10-005` | Verification Suite Execution & Operational Signoff | QA / Security | `docs/planning/WORKLOGS.md`, `docs/plans/phase10/status.md` | Full repository test suite passes with race detection; memory budget verified within < 1.7 GB total container allocation. |
