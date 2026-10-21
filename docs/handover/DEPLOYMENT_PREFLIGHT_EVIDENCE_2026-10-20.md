# Deployment preflight evidence — 20 October 2026

**Release status: BLOCKED. No deletion, credential change, container removal/restart, database mutation or deployment was performed.** This record supersedes earlier statements that all backend readiness checks were healthy or that no VM backup existed.

## Subsequent deployment preparation

The user subsequently authorized deployment and reiterated cleanup first. [The updated cleanup/deployment record](CLEANUP_DEPLOYMENT_STATUS_2026-10-20.md) supersedes the earlier DB backup/access blockers below: the current credential authenticated, a private managed DB backup was verified, and its isolated restore matched production counts. Migration rehearsal now stops on Azure's missing pgcrypto prerequisite. No cleanup or production deployment has occurred.

## Updated credential verification - 14:36:53 UTC

The user-provided credential was read from the approved local file and transported only through encrypted SSH standard input. It was not printed, placed in command-line arguments, persisted on the VM or applied to running configuration.

- Authentication to `hackways_prod`: **SUCCEEDED**.
- PostgreSQL server: **16.15**; TLS active and session transaction read-only confirmed.
- Updated password versus existing protected VM password: **different**, using an equality result only. The earlier rejected stored credential remains configured in the running applications.
- No password reset, configuration change, container restart, deployment, deletion or database write was performed.
- The direct diagnostic authentication blocker is resolved. Application credential synchronization still requires authorization. Database backup/restore, migration/rollback rehearsal and the other release gates remain pending; the earlier authentication failures below are historical evidence.

Sanitized result: `logs/release-review/updated-credential-auth-result.json`. Existing proposal bundles predate this verification and were not rebuilt.

## Confirmed cleanup boundary

- Database work is restricted to `hackways_prod` on the user-approved managed PostgreSQL host. Do not clean, restore over or drop `postgres` or unrelated databases.
- Preserve `hackways-nginx`, `hackways-core-backend` and `hackways-global-backend`, their current images, SSH, networking, credentials, required services and existing backups.
- Only the exact test record IDs/files agreed after inventory may be removed. A test-like name alone is not sufficient evidence.
- The roughly 831 MB reclaimable Docker build cache is a proposed candidate, not approved/executed deletion. No stopped containers, unused images or Docker volumes were found in the initial inventory.
- Test-data candidates remain unidentified because current stored DB credentials fail authentication. No blanket deletes, table truncation, database drops or volume/system pruning are approved.

## Source and release identity

| Item | Evidence |
| --- | --- |
| Current branch / HEAD | `develop`, `95388304079f32a3c5d205e6ac7699e7747a9e42` |
| HEAD date / message | 16 October 2026; ingress egress/storage routing change |
| Working tree | 268 status entries before this evidence file was added; includes recent modified/untracked application work |
| Reviewed source commit containing recent work and Compose fix | **Not available. Changes have not been committed as a reviewed candidate.** The historical HEAD must not be supplied as that release SHA |
| Current corrected Compose file SHA-256 | `1a8004b6b6384ee7473d91106208b6ec7827e90a48806beda6b6c2af095c0030` — a file checksum, not a Git commit |
| Destination/source decision | Product destination and private-repository protection plan remain unsettled; visible organization repo is still only `.github` |

No commit was fabricated, remotely pushed, approved or deployed. A reviewed source snapshot and its CI evidence must precede a release designation.

## Full Compose configuration validation

The corrected local Compose bytes were passed in memory to the existing VM Docker Compose CLI, using `/opt/hackways/.env` and the existing production project directory for path resolution. Rendered configuration containing secrets was captured in memory and never printed or saved.

- Docker Compose version: **5.6.0**.
- Normal interpolation and complete `config --quiet` validation: **exit 0**.
- JSON model parsing: **passed**.
- Default model contains seven services: six product services plus Nginx. The optional local PostgreSQL profile is not enabled.
- Both backend model targets were checked against the exact managed host and database `hackways_prod`: **matched**.
- Both backend production models contain the required session configuration field, without recording its value.
- Both frontend server API URL entries are correctly nested under `environment` in the source file.
- Existing Nginx `nginx -t`: **exit 0**.

This proves Compose interpolation/model validity, not container-image availability, successful application startup, database access, public URL correctness or release health. No `up`, `down`, `restart`, image pull or deployment command was run.

## Database authentication investigation

All diagnostic SQL attempts explicitly selected `PGDATABASE=hackways_prod`, required TLS and set transaction read-only, a five-second statement timeout and a one-second lock timeout. No secret was placed in SQL, a command-line connection URI, console output or this document.

| Check | Result |
| --- | --- |
| Database endpoint and port | Expected managed PostgreSQL endpoint, port 5432 |
| Database/user | `hackways_prod` / `hackways_admin`, matching protected host settings |
| DNS lookup | Succeeded |
| TCP connection | Succeeded |
| URI parsing | No fragment; `sslmode=require`; password percent-decoding performed |
| Host DATABASE_URL versus POSTGRES_PASSWORD field | Stored passwords match; user fields also match |
| Core container versus host source | Host, port, DB, username and stored password all match |
| Global container versus host source | Host, port, DB, username and stored password all match |
| Fresh PostgreSQL authentication | **Rejected: password authentication failed** |
| Approved local updated-password file | Absent during inspection; its contents were not available to test |

**Finding:** DNS and TCP connectivity are working. PostgreSQL rejects the credential currently stored consistently in the protected VM source and both containers. This is consistent with the user-reported external password change and stale runtime configuration; the approved updated credential still needs a secure source before it can be tested. No password reset or credential propagation was performed.

The stored password was not printed, compared in public logs, copied to GitHub or sent to team members. Runtime versus host comparison emitted equality booleans only. A duplicate authentication attempt using the identical POSTGRES_PASSWORD field was skipped.

### Current backend readiness

- Core backend `/health`: **HTTP 503**, `degraded`, database `unhealthy`.
- Global backend `/health`: **HTTP 503**, `degraded`, database `unhealthy` on the latest diagnostic. Its earlier HTTP 200/Docker healthy result is no longer evidence of current DB readiness.
- Both running containers have `APP_ENV` unset and legacy `ENV=production`, and report `environment=development`. The reviewed production configuration must use `APP_ENV=production`; the corrected local Compose model already does. No live environment was changed.
- Both containers remain running. No session was killed and no connection pool was restarted to try to bypass authentication failure.

## Verified VM application/configuration backup

Created a new backup without overwriting existing objects or writing an archive to VM disk:

| Item | Evidence |
| --- | --- |
| Storage account/container | Approved `hackwaysstorage` / `backups` |
| Container privacy check | HTTP 200; no public-access ACL enabled |
| New object | `predeploy/20261020T135129Z-f2fbc9a2/hackways-vm-config.tar.gz` |
| Size | 448,544 bytes |
| SHA-256 | `8808ff923c1ec55c924903c36e2c2b678effcc06b25ce13b6e316ec9af4a1777` |
| Upload | HTTP 201; create-only conditional request |
| Azure response | Server encryption confirmed |
| Independent download | Exact byte count and SHA-256 matched |
| Archive inspection | 390 members; no absolute or parent-traversal archive paths |
| Round-trip checks | Protected application configuration matched the source; three-container metadata manifest decoded and matched expected count |

Scope: `/opt/hackways`, existing `/etc/letsencrypt` and a manifest of the three running containers' image IDs/mounts/ports. The private archive includes sensitive recovery configuration; access must remain restricted. No credential contents or public/SAS download link appear here.

Verification was a read-back and **in-memory archive round trip**. This is not a managed-database backup, database restore test, bootable VM snapshot or full service rollback rehearsal. No live files were overwritten during verification. Current images remain preserved on the VM; their recoverability after future migrations is not yet established.

Backup transport/integrity implementation followed Microsoft's [Shared Key request authorization](https://learn.microsoft.com/en-us/rest/api/storageservices/authorize-with-shared-key) and [Put Blob](https://learn.microsoft.com/en-us/rest/api/storageservices/put-blob) contracts; all authenticated requests used HTTPS and redirect following was disabled.

## Database backup, migration and rollback evidence

- Managed `hackways_prod` dump: **BLOCKED by authentication**.
- Isolated restore of a managed DB backup: **NOT RUN**.
- Live schema/migration version: **NOT READ** because authentication failed.
- Core backend `sqlc compile`: **exit 0**.
- Global backend `sqlc compile`: **exit 0**.
- Migration rehearsal against a restored copy: **NOT RUN**. Static compilation is not a migration rehearsal.
- Application rollback rehearsal: **NOT RUN**. No reviewed new release or approved compatible rollback release exists yet.
- Migration `000002_free_event_workflows.down.sql` explicitly refuses destructive rollback. Do not bypass it or force migration version metadata.

Once secure access is working: obtain a consistent database-only dump, verify private off-host integrity, restore into an isolated nonproduction database, compare schema/table counts and constraints, apply only pending reviewed migrations, verify capacity/data invariants, and test a schema-compatible application rollback or forward repair. Keep the rehearsal isolated from production notification delivery and media writes. Do not use the production database as a rehearsal target.

## Ownership and proposed checkout/deployer setup

- GitHub manager: **Devara**, as named by the user; no GitHub account handle or additional operational role is inferred.
- Release manager: **awaiting the user's separate assignment**.
- Infrastructure owner: **awaiting the user's separate assignment**.

Proposed setup, not installed:

1. Preserve the existing non-Git `/opt/hackways` tree and running images as recovery material. Do not reset it, replace it with a clone or overwrite the protected environment while the release is blocked.
2. After the repository decision, prepare a separate clean staging checkout, such as `/opt/hackways-checkout`, checked out to the reviewed SHA. Configure approved private Git/GHCR read access with the infrastructure owner; record identity/expiry, never credential values.
3. Resolve the current path contract before activation: the deployer supports REPO_DIR, but `deploy/scripts/deploy.sh` currently hardcodes `/opt/hackways`. Merely setting REPO_DIR to a staging path is insufficient. A reviewed path/configuration change or explicitly approved final checkout migration is required.
4. Install the signed-workflow webhook service only after owners, branch policy, CI-built immutable images, private pull access, backup/restore and rollback evidence are ready. Align webhook and host secrets privately; do not expose the service directly to public ingress.
5. Keep the production branch consistent across repository rules and the deployer. Existing source defaults to develop; the final production branch remains a release-policy decision.
6. Build off the production VM. Preserve SSH/network configuration and admin restrictions. Perform any eventual deployment in an agreed window with explicit readiness checks for all six services and Nginx.

The GitHub Project should keep release readiness blocked on current credential access, reviewed source SHA/CI, database backup/restore and migration/rollback evidence, named owners, repository migration/protection support and actual deployer/pull-access setup.
