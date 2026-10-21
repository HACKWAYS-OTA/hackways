# Cleanup and deployment status - 20 October 2026

## Current state

The user requested cleanup before full deployment. No cleanup or production release has been executed. Three existing containers, production records, credentials, networking, SSH and backups remain intact. The approved current database credential works through the protected local file but has not been propagated into the old application runtime.

## Verified recovery

- Existing private VM configuration backup remains preserved.
- Managed database: `hackways_prod` only, PostgreSQL 16.15, approximately 8.5 MB including database overhead.
- A custom-format, no-owner/no-privileges database dump was uploaded create-only into the private `backups` container with server encryption confirmed.
- Blob: `predeploy/20261020T144151Z-5c0fcc89/hackways_prod.dump`.
- Size: 24,149 bytes.
- SHA-256: `cb61d9bba203075e3cd4b2382198ab1185fa8c94a0fa0315cdbe3b77602c9b17`.
- Independent download matched the exact bytes/hash. No database archive was persisted on VM disk.
- Restored into isolated `hackways_restore_20261020t144151z_5c0fcc89_test`. All seven table counts matched production: one user, one session, four reference categories, and zero organizers/events/RSVPs/outbox rows.
- Isolated databases are retained pending rehearsal completion and explicit cleanup approval. No notification worker is connected to them.

## Proposed cleanup boundary - approval pending

| Target | Finding | Proposed action |
| --- | --- | --- |
| Docker build cache | Approximately 831.1 MB reclaimable | Remove unused build cache only after scope confirmation |
| Containers | Three running required services; no stopped containers | Preserve all three service definitions and recovery images |
| Images / volumes | Three active images; no unused images or volumes | Preserve |
| Application records | One user/session; no evidence they are disposable | Preserve; no guessed seed deletion |
| Categories | Four canonical reference records from initial migration | Preserve |
| Azure media container | Empty | Preserve container and configuration |
| Azure backups container | Two verified backup objects | Preserve both and container access controls |
| Host application directories | Only `/opt/hackways`; `/srv` and visible home application inventory empty | Preserve live source/configuration; no obsolete app identified |
| Other Azure resources | Inventory unavailable without management-plane access | Do not infer that any resource is obsolete or delete it |

No general Docker prune, recursive application directory deletion, database wipe, unrelated database operation or Azure resource deletion is approved by this candidate list.

## Migration rehearsal findings

1. Production has no `schema_migrations` history and only the `plpgsql` extension. Do not blindly run all migrations or force a migration version.
2. Created isolated `hackways_baseline_20261020t144151z_5c0fcc89_test` to compare the initial schema. Azure rejected the unmodified migration's `CREATE EXTENSION pgcrypto` because the extension is not allow-listed.
3. For schema comparison only, omitted the two extension declarations in the isolated baseline input. The checked-in migration was not modified. Normalized schema dumps then matched the restored production schema exactly. This does not prove the extension prerequisites are satisfied.
4. Exact migration `000002_free_event_workflows.up.sql` applied successfully in a transaction on the restored copy. SHA-256: `2db500a822ffbdfd5c28ce6ad30e20955da08bbb78008e437462a7626d206a6c`.
5. Migration `000003_spaces.up.sql` failed and rolled back on the isolated restore. It depends on `public.gen_random_bytes`, supplied by the missing pgcrypto extension. Production was not migrated.
6. Allow the required `pgcrypto` and `uuid-ossp` extensions through the server's existing `azure.extensions` allow-list, preserving other entries, then complete baseline registration, remaining migration rehearsal and compatible application rollback evidence before production migration. Migration 000002 has an intentionally forward-only down script; do not bypass it.

Microsoft documents the allow-list prerequisite and Portal/CLI procedure in [Allow extensions](https://learn.microsoft.com/en-us/azure/postgresql/extensions/how-to-allow-extensions). Management-plane authorization is currently missing: no local Azure CLI/session, no connected browser, and the VM managed-identity token request returned HTTP 400. Database access and storage access do not grant Azure resource-management access.

## Source and delivery

- Current Git HEAD is still historical `95388304079f32a3c5d205e6ac7699e7747a9e42`; no new release commit, push or deployed SHA exists.
- Current source repository: `varshith-dev/HACKWAYS-IN`. Only `.github` is currently visible in `HACKWAYS-OTA`; the product repository decision remains pending with the user/Devara.
- Local deployment review additionally found the public image omitted `public/` assets, including the site's fonts. Added the same asset copies already used by the core/organizer images.
- Excluded ignored `logs/` trees from Docker context, preventing local diagnostic/build copies from being uploaded as image context.
- Targeted whitespace checks pass; actual Docker image construction remains pending off-VM CI. Prior proposal bundles predate these additional changes and must not be represented as the updated release.
- Stable ticket signing configuration, named operational owners, accepted commit/full CI, immutable image delivery, remaining migration/rollback rehearsal and release ingress checks remain required. No replacement signing material or other credentials were generated.

## Immediate handoff

1. Confirm the concrete cleanup list above before deletion.
2. Provide authenticated Azure management access or have the infrastructure owner add the two required extensions to the existing allow-list and provide the resource-group inventory.
3. Confirm the product repository for the actual candidate commit and CI image builds.
4. Complete rehearsal and recovery checks, then perform the authorized cleanup and deploy all six independent services using CI-built images. No production frontend build is permitted.

Do not report this environment as live or fully deployed. The old backends remain configured with the rejected stored password until the deployment/configuration change is activated.
