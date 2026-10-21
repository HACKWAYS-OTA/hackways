# Review of the GitHub manager's readiness audit

20 October 2026. Reviewed `GITHUB_READINESS_AUDIT_2026-10-20.md` supplied from Downloads against available GitHub APIs, the working copy and read-only VM inventory. No release approval is granted by this review.

## Verdict

The audit is consistent with the evidence available and correctly distinguishes historical CI/images, organization setup and live production readiness. **Organization preparation has progressed; cleanup and deployment are not ready to execute.**

## Confirmed and accepted

- HACKWAYS-OTA exists. The visible organization repository list still contains only public `.github`; the product destination has not been created/transferred.
- Both `meridbase` and `varshith-dev` were independently confirmed as organization administrators through the GitHub API. Their specific release/infrastructure responsibilities remain unassigned.
- Historical source SHA `95388304079f32a3c5d205e6ac7699e7747a9e42` and successful CI are correctly described as historical evidence. They do not include substantial local-only changes or establish the live version.
- All six reported container digest strings have the correct SHA-256 length. This is syntax validation only, not registry-content verification.
- The Project configuration and four blocked draft items are useful preparation. Keeping assignees, release dates and release filtering unset until agreed avoids inventing readiness.
- The audit correctly leaves transfer, permissions, build variables, release identity, backup verification and production rollout pending.

## Independent verification limits

- Project configuration is manager-reported: this CLI lacks `read:project`. The review does not assert an independent field/view inspection.
- Package visibility and digest-to-tag mappings are manager-reported: a read-only package API check returned 403 for missing `read:packages`. No token scope was expanded. Verify these mappings and pull permissions with an authorized registry identity before release.
- A valid-looking image digest or successful CI run is not evidence that a VM can pull that package or is running it.

## New deployment preflight evidence to add to the audit

| Area | Observed state | Required action |
| --- | --- | --- |
| Compose working copy | The two frontend server API URL indentation errors are now corrected locally; YAML parsing succeeds | Validate with Docker Compose and approved environment inputs, review and commit the fix, then validate the release configuration. No production Compose change has been applied |
| Current VM containers | Only Nginx, core-backend and global-backend are running; all three report Docker healthy; no stopped containers or Docker volumes found | Preserve these required containers. Do not label them obsolete or remove them as test applications |
| Current VM images | Two backend containers use local `hackways/...:latest` image names; no four frontend containers were found | Historical GHCR evidence does not describe the running six-service release. Record actual image IDs/digests and prepare the agreed six-service rollout |
| Deployment checkout | `/opt/hackways` exists but is not a Git checkout | Establish the approved checkout/release path without overwriting current configuration or losing recovery material |
| Deployer | `hackways-deployer.service` and `/etc/hackways/deployer.env` are absent | The source-level develop webhook policy is not an installed live deployment path. Agree the operator and configure/verify the selected deployment mechanism |
| Database target | Existing container configuration targets `hackways_prod` on the managed PostgreSQL host named by the user | Limit application cleanup to the confirmed database and approved record IDs; preserve unrelated databases |
| Database readiness | Read-only connection using the running container's stored credential failed authentication after the user reported a password change | Use the approved updated credential source and validate actual database readiness. Docker healthy alone is insufficient |
| Backup/cleanup | No cleanup performed; no verified backups created yet. Exact cleanup boundary confirmation and secure updated credential access are pending | Confirm scope, complete private off-host backups and an isolated restore check, then obtain approval for individual deletion candidates |
| Storage reclamation | Approximately 831 MB reclaimable Docker build cache; no unused images reported | Cache is only a cleanup candidate until the user confirms the boundary and backups/recovery requirements are satisfied |

SSH access was successful with the supplied private-key path and strict known-host verification. The key content and credential values were not printed or added to repository files. Read-only inventory changed no running services, networking or credentials.

## Items for the GitHub manager now

1. Confirm the final product repository transfer/name and the Free-plan protection decision with the organization owners. Do not use `.github` as the product repository or as a place for private operational files.
2. Assign a release manager and infrastructure operator. Organization ownership does not automatically assign either role.
3. Coordinate review and a safe commit of the accepted local source snapshot. Provide the resulting release SHA; do not choose the old successful SHA by default.
4. On the confirmed product destination, configure the three public origin variables, eligible teams/Code Owners and the three required PR checks.
5. Run CI for the accepted snapshot, publish the six release images and record their immutable digests and the host's private pull-access evidence.
6. Convert the four blocked Project drafts into actionable repository issues once the product repository exists. Add the relevant setup backlog items and acceptance criteria. Keep the release blocked until engineering and operations evidence is attached.
7. Receive the supplied companion files and validate the actual CODEOWNERS teams before activating the template.
8. Update the preflight draft with the local Compose correction and the observed live-host differences above. Neither the Compose parser result nor this review authorizes production deployment.

The confirmed source package namespace and historical digests should remain recorded for traceability. Do not relabel them as destination release artifacts or assume they are a schema-compatible rollback target for later local changes.

## Items for engineering and the infrastructure operator

- Obtain the user's exact cleanup scope confirmation and identify genuine records before preparing deletion candidates. Names such as "test" are not sufficient proof that a record is disposable.
- Obtain secure current DB access, inspect schema/migration state read-only, and back up the actual managed database. Existing Docker-local PostgreSQL backup scripts do not cover that host as written.
- Preserve current VM configuration, SSH/network access, required services, credentials and existing backups. Establish private backup integrity and isolated restore evidence before deleting anything.
- Rehearse the reviewed migrations against an isolated restored database, verify data/inventory invariants and validate a compatible application rollback. Do not indiscriminately run migration downs: the workflow foundation migration is intentionally forward-only.
- Validate the complete Compose configuration, runtime environment wiring, public URLs, Nginx routes, private admin ingress and real service readiness before rollout. Build images off the production VM.

## Forwardable response to the manager

> Thank you. The audit correctly separates GitHub setup from release readiness. Please proceed with the agreed repository/plan decision, named release and infrastructure owners, and preparation of the reviewed source snapshot. Then provide the final repo URL, release branch/SHA, current-source CI run and six immutable image references with host pull access verified. Please attach these to the blocked Project items and convert them to repository issues once the product repo exists.
>
> Engineering has corrected the local Compose indentation issue; full release validation remains pending. Read-only VM inspection found three required running containers, a non-Git application directory and no installed deployer service. Current stored database credentials fail authentication after the reported password change. Cleanup confirmation, secure DB access, verified backups, migration rehearsal and rollback evidence are still required. Nothing has been cleaned up or deployed.
