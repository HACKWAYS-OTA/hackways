# ACS global-backend replacement proposal

## Approved scope

The user authorized a minimal first-send fix, explicit regression tests, and a single global-backend image build in GitHub Actions. Production changes and container replacement are **not authorized in this step**. Keep this PR unmerged until reviewed. No new credentials, database migration, other application change, or six-image candidate tag is required.

## Failure and change

Candidate `39c8b4487a557c7dfb3a1d46aaddecfa848fbbd3` asks ACS for a new operation's status before sending the email. The live signed request timed out at 20.09 seconds; a longer probe returned HTTP 504. An invalid-signature control returned 401 immediately. No email was sent during diagnosis.

The existing PostgreSQL outbox claim already returns `first_sent_at`. The worker now forwards whether that persisted value was absent **before** `BeginOutboxProviderAttempt`, which still persists the timestamp under the current lease before any provider call. A first attempt POSTs directly with the stable operation ID, repeatability request ID and first-sent timestamp, then polls acceptance. No new persistence mechanism or migration is introduced.

Reclaimed attempts, including a process crash with `retry_count=0`, reconcile the persisted operation first. A status timeout, 504, malformed result, authentication failure or other ambiguous state never triggers another POST. A confirmed 404 permits a repeatable POST only inside the existing conservative four-minute replay cutoff, using the original identity/timestamp. Expired ambiguous outcomes require operator reconciliation. An accepted or uncertain send is never assigned a new operation ID. Provider acceptance is not inbox delivery.

## Verification

Explicit test-only transports cover first-send ordering despite an unavailable unknown-operation GET, accepted-operation polling, status timeout, POST timeout followed by successful reconciliation, confirmed-404 repeatability, expired replay windows, 504/401/429/unreadable/failed/unknown results, and missing persisted identity. The existing real PostgreSQL outbox suite additionally checks first-attempt identity, regular retries and restart recovery with retry count zero.

Run in `services/global-backend`:

```text
go test ./internal/notifications -run TestACS -v -count=1
go vet ./...
```

Mandatory Linux CI applies numbered migrations to its disposable PostgreSQL 16 database and runs formatting, `go vet`, `go test ./... -json -race -count=1` and `sqlc compile`. Local Windows lacks the C compiler/CGO setup required by the race detector; do not claim a local race pass. Database tests must not target production. Passing controlled transport tests and CI do not verify actual Azure delivery after deployment.

## Single-image publication

After the exact PR head SHA has passing mandatory CI, an authorized operator adds `build-global-backend` to the open, non-draft PR targeting main. `Global Backend Candidate` runs only on that label event for a same-repository PR. It verifies the current head, label, restricted file scope and latest successful PR CI for that SHA before publication. It builds only `deploy/docker/Dockerfile.global-backend` on Linux/amd64, reuses the existing service build cache, publishes only the full source-SHA tag, and records the immutable digest in a one-day artifact.

Do not create an `rc-*` tag for this replacement: the existing candidate workflow would rebuild all six images. No workflow activates Azure, deletes images or changes paid settings. Reuse the first successful image build and its recorded digest; do not relabel/rerun it to generate another image unnecessarily. The exact source SHA, CI link, single-image run and digest belong in the PR review evidence.

The pre-build billing API reported 113 Linux minutes for October 2026, fully discounted, net amount zero; usage reporting can lag. Repository cache usage was 3,633,984,036 bytes. Estimated additional integration/publication use is 15–25 Linux runner minutes, with one image and a small short-lived digest artifact. No paid overage, plan upgrade or package deletion is enabled.

## Images retained

The replacement applies only to global-backend. Preserve its previous approved image `ghcr.io/hackways-ota/hackways-global-backend@sha256:c8a6bbe794077e3b651a07e86ce7d2795d526f522ca44795c8b5ea455d56282a` and the existing VM rollback image `sha256:97904385653604e0addaa2e854645a856f931b46aef8a1e05e701f289da5bed8`.

| Unchanged component | Approved immutable image |
| --- | --- |
| admin-web | `ghcr.io/hackways-ota/hackways-admin-web@sha256:3f3dce927289bd4c45b93900cd48100f538d80bdf624a54a1d7dc9034f587ff7` |
| core-backend | `ghcr.io/hackways-ota/hackways-core-backend@sha256:0860e11530d612441b0fd496d3d6fb052d21b6717b2a7c615d78a3c2b7952db7` |
| core-web | `ghcr.io/hackways-ota/hackways-core-web@sha256:751fb2e34982d14055dabfc901d0f7807be7752d173137701f83a204ad33d67f` |
| organizer-web | `ghcr.io/hackways-ota/hackways-organizer-web@sha256:b36a2412ce1f64bda8125f74b057d66bb19c8060f4b1d730f7a6b962f07f2c79` |
| public-web | `ghcr.io/hackways-ota/hackways-public-web@sha256:4184513e0cf2386a694a4504f492b2b08dd85d0fc34e4c89cb7ff75b3947f81b` |

This step does not resolve the existing public 502 or claim that Hackways is live. Review the replacement before authorizing any production migration or activation.
