# Feature rollouts and configuration versions

## Deployment model

Build each of the six services in CI and deploy one compatible, digest-pinned release. A release may contain both a compact and an expanded UI. Gradually expose the implemented variant through feature flags; do not run duplicate Next.js stacks on the 4 GB VM just to divide traffic.

This module versions **feature configuration**, not executable source code. It does not deploy arbitrary commits, execute shell commands, download UI code, or roll database schemas backward. Application image deployment and rollback retain the existing approved CI pipeline. The release-candidate build passes `SOURCE_COMMIT` into all six runtime images; the admin version page reports its own and the global engine's actual configured source. Older/local images explicitly report an unavailable source instead of inventing a platform version.

## Authoritative control plane

Migration 17 adds an append-only application API for `runtime_rollout_versions`, with one locked `runtime_rollout_control` pointer. PostgreSQL holds configuration, the private cohort salt and history. Administrators publish a named version with a reason, current revision, confirmation and idempotency key. A restore copies a prior configuration into a new version; it never rewrites history. Audit records include the old/new version references. Unknown, unimplemented flags and invalid percentages/windows are rejected by Go.

The initial configuration preserves the currently implemented behavior at 100%. Shipped flags:

| Key | Enabled presentation | Disabled baseline |
|---|---|---|
| `discover.promotions` | Admin-configured scheduled Discover banner | Regular event catalogue |
| `manage.overview.v2` | Expanded overview with recent activity and event details | Compact overview retaining attendance counts and essential event actions |

Future features require a real implementation and an entry in the Go registry. Adding a database key cannot manufacture a working feature.

## Evaluation without a database lookup per feature

The global engine refreshes one small configuration snapshot every ten seconds, with a three-second query deadline. Publication attempts an immediate refresh after commit. A mutex serializes refreshes; an atomic pointer publishes a complete, immutable snapshot to readers. An older response cannot overwrite a later refresh. HTTP evaluation performs no database call and takes no database connection.

Rules are evaluated against the clock in memory, including start-inclusive/end-exclusive windows. HMAC-based buckets are stable for each feature and cohort, so increasing a percentage retains previously eligible members. Sign-in binds a pseudonymous cohort to the account; anonymous visitors receive an opaque random browser cohort. The cookie is Secure, HttpOnly, host-scoped and grants no access. Logout clears it. A first-visit Next middleware transports the cookie issued by Go into the current SSR request and response; it performs no flag decisions or authorization. If that transport fails, it does not block the page.

`GET /api/v1/auth/runtime` returns only booleans, version/revision, label and freshness, with an ETag. It does not return rules, identifiers or the cohort salt. Missing or more-than-two-minute-old snapshots disable presentation flags and retain their usable baseline. The admin page exposes the evaluator's loaded version, age and freshness separately from the saved configuration.

Authentication, authorization, ownership, database constraints, input validation and capacity locking remain live Go responsibilities. **A feature flag is never a permission check.** This module deliberately does not gate backend security or migration compatibility on a browser cookie.

## Adoption and rollback

Management resolves the UI variant on server navigation. It does not replace an active editor midway through work. Discover revalidates while visible at schedule boundaries or at most once per minute; hidden tabs pause and revalidate on return. The seven-second carousel rotation is local and makes no API request.

An emergency disable is a reviewed configuration publication. Healthy evaluators adopt within the ten-second refresh interval; already open Discover pages adopt on their next revalidation, and management on navigation. This is bounded eventual adoption, not a claim of instantaneous remote removal of every open page. A failed database refresh uses the last known snapshot for at most two minutes, then the safe baseline.

Roll out at 1/5/10/25/50/100 percent as appropriate. Inspect actual errors and workflow outcomes before increasing; this module does not invent automatic health judgments or claim measured capacity. Restore a prior configuration to stop exposure. Keep schema 17 on application rollback so configuration history remains available.

## Verification

Mandatory Go suites cover stable/nested cohorts, exact time boundaries, disabled/stale/missing snapshots, concurrent HTTP reads without a database dependency, admin/origin enforcement, unknown keys, revisions, idempotent publication, restore-as-new-version and revoked grants. Local PostgreSQL 18 checks passed, and Release Candidate run 39795523731 passed PostgreSQL 16/race CI and all six image builds. All six production runtimes now identify source `dab8d1cd545b843bf20c2d8cdac16bb7feb058d8`; schema 19 is clean and the public evaluator reports fresh configuration version 1. No Discover promotions are configured, so no banner is rendered. Local Windows race execution is unavailable without a C compiler. Browser presentation review is separate from those API checks.
