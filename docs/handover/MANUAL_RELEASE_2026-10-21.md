# Manual release checkpoint — 2026-10-21

## Status: blocked before production migration or activation

The user authorized a controlled manual release of candidate `39c8b4487a557c7dfb3a1d46aaddecfa848fbbd3`. Production is **not deployed or verified**. The remaining observed launch blocker is the ACS status request performed before the candidate's first email send. Do not bypass this failure or describe staged containers/images as a live deployment.

No production data was deleted, no production migration was applied, no running production container was replaced, and no existing credentials were reset. No paid settings, automatic deployment, package deletion, or VM image builds were enabled.

## Git and CI

- [PR 1](https://github.com/HACKWAYS-OTA/hackways/pull/1): reviewed and merged into `develop` as `62198641009c4e717982f07b9b5fd7b6b2259472`.
- [PR 3](https://github.com/HACKWAYS-OTA/hackways/pull/3): reviewed production promotion, merged into `main` as `d8a3e3e592ae634a9eca33214429d14563f86483`.
- The main merge tree equals the candidate tree exactly: `a8dcdf65b108010f5cf75eff99c17c33eda946b5`.
- [PR 2](https://github.com/HACKWAYS-OTA/hackways/pull/2) was merged separately into `develop` by the manager. Its documentation is excluded from production main and the staged release.
- Reviews were transparent agent technical review comments under the available author account, not independent human approvals. Private GitHub Free protection/required-review enforcement is not claimed.
- Candidate run [39540190665](https://github.com/HACKWAYS-OTA/hackways/actions/runs/39540190665): all 13 jobs passed, including six image builds/publications.
- Promotion PR run [39544699162](https://github.com/HACKWAYS-OTA/hackways/actions/runs/39544699162), main run [39545152929](https://github.com/HACKWAYS-OTA/hackways/actions/runs/39545152929), and develop documentation run [39545342003](https://github.com/HACKWAYS-OTA/hackways/actions/runs/39545342003): passed. These did not rebuild images.
- Repository hooks were empty. The optional deployer is not installed/enabled; source automatic deployment remains disabled.

## Six immutable images

All six were pulled on the VM by digest, and their local repository digest and OCI revision label matched candidate `39c8b4487a557c7dfb3a1d46aaddecfa848fbbd3`. Existing rollback images were retained.

| Service | Immutable image |
| --- | --- |
| admin-web | `ghcr.io/hackways-ota/hackways-admin-web@sha256:3f3dce927289bd4c45b93900cd48100f538d80bdf624a54a1d7dc9034f587ff7` |
| core-backend | `ghcr.io/hackways-ota/hackways-core-backend@sha256:0860e11530d612441b0fd496d3d6fb052d21b6717b2a7c615d78a3c2b7952db7` |
| core-web | `ghcr.io/hackways-ota/hackways-core-web@sha256:751fb2e34982d14055dabfc901d0f7807be7752d173137701f83a204ad33d67f` |
| global-backend | `ghcr.io/hackways-ota/hackways-global-backend@sha256:c8a6bbe794077e3b651a07e86ce7d2795d526f522ca44795c8b5ea455d56282a` |
| organizer-web | `ghcr.io/hackways-ota/hackways-organizer-web@sha256:b36a2412ce1f64bda8125f74b057d66bb19c8060f4b1d730f7a6b962f07f2c79` |
| public-web | `ghcr.io/hackways-ota/hackways-public-web@sha256:4184513e0cf2386a694a4504f492b2b08dd85d0fc34e4c89cb7ff75b3947f81b` |

Private registry access succeeded. The supplied token was more privileged than a pull token, so Docker authentication was held only in a root-only temporary `/run` directory and removed after the operation. No token was committed, printed, or installed as persistent Docker credentials. Future pulls need a suitably scoped read-only credential; package visibility was not changed.

## Staged VM configuration

- Preserved the existing non-Git `/opt/hackways` tree, three original running containers, SSH/network configuration, certificates, images, and prior backups.
- New checkout: `/opt/hackways-checkout/releases/39c8b4487a557c7dfb3a1d46aaddecfa848fbbd3`, on local `main` at `d8a3e3e592ae634a9eca33214429d14563f86483`.
- Transferred a verified complete-history Git bundle over SSH; the checkout is clean and contains no credentials. Origin points to the private organization repository. Repository policy rejected a proposed read-only deploy key, so no repository deploy key was installed. No broad GitHub token was persisted for fetching.
- Protected configuration: `/etc/hackways/releases/39c8b4487a557c7dfb3a1d46aaddecfa848fbbd3/runtime.env`, mode `0600`, enclosing directory `0700`.
- The approved current DB credential authenticates to **hackways_prod** with TLS; it is staged in both engine targets. The original running containers still use their old runtime configuration. No server password was changed.
- A new stable base64 Ed25519 seed is staged for core-only ticket signing and included in the verified private recovery backup. Existing session/storage/ACS credentials were preserved.
- `compose.release.json` pins all six recorded digests with pull policy `never`, plus the preserved Nginx digest `sha256:516475cc129da42866742567714ddc681e5eed7b9ee0b9e9c015e464b4221a00`.
- Production Compose validation passed: six apps plus Nginx, no local PostgreSQL profile, no backend host ports, explicit production environment and HTTPS CORS origins, configured container memory limits totaling **1,148,190,720 bytes**. These are limits, not workload measurements.
- Installed official prebuilt golang-migrate **4.20.1**, after checking the release archive SHA-256 `b1e13121d4178b7804a9fb29dca62d57b732390f7e5ae33f2cd65e9b192e1bd4`. No application compilation ran on the VM.

## Current verified backups

Both objects are in the private Azure `backups` container, with create-only upload, server encryption, independent download, and SHA-256 verification. Never expose public/SAS links to these recovery objects.

| Backup | Object | SHA-256 | Bytes |
| --- | --- | --- | --- |
| Original application/configuration/certificates, three-container recovery manifest, staged runtime/signing environment | `predeploy/20261021T051350Z-a7e6cfe2/hackways-vm-config-and-signing.tar.gz` | `a8973b4437f5358010b6350342da1f6b75403ee64538b27cb0ec7aac4cb8b17e` | 454,283 |
| Managed application DB | `predeploy/20261021T051353Z-66214246/hackways_prod.dump` | `0cbfa0e7b50c00fb61b5836307db0d52e25e21c81e93120c1cc37dd0a2e6eb8f` | 24,149 |

The config archive passed safe-member and in-memory round-trip checks, including old and staged protected environment files. This is not a bootable VM snapshot.

The database backup restored successfully into `hackways_restore_20261021t051353z_66214246_test`. All seven source table counts matched: one user, one session, four reference categories, zero organizers, events, RSVPs and outbox rows. Existing data was preserved; it was not classified as disposable seed data.

## Migration and rollback evidence

- Production has no `schema_migrations` table and only `plpgsql` installed. Azure allows `PGCRYPTO,UUID-OSSP`; production extension installation remains pending.
- The current production schema equals the restored backup schema. Normalized SHA-256: `9a5e07d69864c9517064842226674e567c5c33a8e44affac7b65c2b1dfff9dd6`.
- Executed the exact initial migration in an isolated canonical database, installed the two extensions in the restored copy, and compared schemas. They matched exactly. Normalized initial schema with extensions: `054299df4a7773adc0a74ec7c1079d86d42a44d2295a687549ae8af55081c30f`.
- Only after this explicit comparison, registered baseline version 1 **in the isolated restore** and ran the approved migration directory through the actual CLI. Result: version **7**, `dirty=false`, original counts preserved.
- Both candidate engine images passed production-mode database readiness against that migrated restore. No rehearsal worker targeted production data or sent email.
- Earlier, both preserved legacy backend images passed readiness against a migrated current backup. Core image ID: `sha256:401f0d8b5c9908e4cb8d108878a47c76d90b3905ac1534a85bf127962aa8e7f4`; global image ID: `sha256:97904385653604e0addaa2e854645a856f931b46aef8a1e05e701f289da5bed8`.
- Temporary rehearsal containers were stopped/removed; original production containers were untouched. Isolated databases are retained for reviewed recovery evidence; do not connect production workers to them.
- Migration 000002 is forward-only. Never run initial/destructive downs, erase migration history, or restore a backup over a live database as an automatic application rollback.
- Before activation, recheck freshness and exact production baseline, install only approved extensions, then register the verified baseline and apply pending migrations. Stop on any mismatch/dirty state.
- If activation fails, stop the rollout and restore the preserved application images/configuration using the approved current DB credential while retaining the additive schema. Restrict changes to this release's services; preserve networking, SSH, backups and unrelated databases. Full authenticated rollback journeys and live frontend rollback remain unverified. The original three-container state had no frontends and was already unhealthy, so it is recovery material, not a claim of a previously healthy complete release.
- Managed server `max_connections` reads 859; candidate application pools are bounded to 28 + 22. No server-wide parameter or unrelated database was modified. Reconcile the canonical 60-connection DB limit before claiming complete resource-policy compliance.

## ACS blocker and public checks

- Blob `media-uploads` authenticated container access: **HTTP 200**.
- ACS DNS, IPv4 TCP/443 and TLS certificate validation: passed.
- Signed empty email request with no sender or recipients: **HTTP 400 BadRequest**. No valid message was submitted and no email was sent.
- Signed GET of a new operation ID, exactly as the sender does before first send: timed out at **20.09 seconds**, matching the candidate HTTP client's 20-second timeout. A longer probe returned **HTTP 504**.
- Deliberately invalid signature control: **HTTP 401 in 0.29 seconds**. Existing credentials were not changed. The failure is not the previous GHCR permission problem.
- `services/global-backend/internal/notifications/acs.go` performs the status GET before its POST and returns immediately on a request error. Therefore the observed provider behavior blocks this candidate's first-send path. A runtime timeout increase alone would still receive 504; do not silently patch source or reuse the same candidate identity for changed code.
- The user authorized deployment only. Request authorization for a minimal launch fix and a newly reviewed image, or obtain confirmed corrected ACS resource configuration and repeat the same preflight successfully. Do not activate the known failing candidate in the meantime.
- Public HTTPS reaches Cloudflare. Last pre-activation roots: apex, app and organizer **502**; public admin **403**. These are the old runtime's results, not release success.
- Intended URLs: `https://hackways.in`, `https://app.hackways.in`, `https://organizer.hackways.in`; admin stays restricted at `https://admin.hackways.in`.
- The user chose to verify attendee/organizer authenticated flows in their browser after activation. No production demo account, fabricated session or seeded event was created for verification.

## Resume gate

Resolve the ACS first-send prerequisite, review any changed candidate and image identity, recheck backup freshness/baseline and resource-policy compliance, prepare the exact rollback invocation, then perform the approved manual migration/activation and actual six-service/domain/authenticated verification. Keep automatic deployment disabled. Record the deployed source/merge SHAs, image digests and live results only after they pass.
