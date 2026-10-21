# HACKWAYS-OTA: GitHub manager setup guide

Prepared 20 October 2026. Purpose: organize the existing Hackways product in GitHub. This is a repository, access and delivery setup guide, not a request to rebuild the product.

## Start here

The recommended setup is **one private product monorepo, six independently deployed applications/services, and one organization Project**. Keep the existing folder structure, history and workflows. Separate deployments do not require separate source repositories.

The manager can configure GitHub. The engineering and infrastructure owners must confirm the source release, production configuration and deployment connection before enabling production automation.

### What is already there

Read-only GitHub inspection on the date above found:

| Item | Observed state |
| --- | --- |
| Destination organization | [HACKWAYS-OTA](https://github.com/HACKWAYS-OTA) exists |
| Organization plan | Free |
| Organization base repository permission | Read |
| Organization requirement for two-factor authentication | Not enabled |
| Visible destination repository | [HACKWAYS-OTA/.github](https://github.com/HACKWAYS-OTA/.github), public, default branch `main` |
| Organization teams | None returned by the authenticated API request |
| Existing product repository | [varshith-dev/HACKWAYS-IN](https://github.com/varshith-dev/HACKWAYS-IN), private |
| Product default branch | `main` |
| Current local branch | `develop` |
| Current local HEAD | `95388304079f32a3c5d205e6ac7699e7747a9e42` |
| Latest inspected successful CI | [CI run 38863256258](https://github.com/varshith-dev/HACKWAYS-IN/actions/runs/38863256258), `develop`, same SHA |
| Existing product collaborator returned | `varshith-dev`, repository admin; component ownership still needs assignment |
| Existing product branch protection | `main` and `develop`: one approval and only Repository Hygiene required; administrator enforcement and Code Owner reviews disabled |
| Product repository Actions variables/secrets | Both lists returned zero repository-level entries; this does not establish host or organization-level configuration |
| Product repository webhooks / GitHub deployment records | Both lists returned zero entries |
| GitHub Projects | Not inspected: current CLI authorization lacks `read:project`; do not assume there are no Projects |
| Configured deployment branch | Checked-in deployer defaults and systemd unit use `develop` |
| Exact branch/SHA currently serving production | **Unverified. Successful CI and local configuration are not proof of a live deployment.** |

The working directory contains substantial modified and untracked application code beyond this HEAD. A transfer or fresh clone of the remote repository will not contain those local changes. The engineering owner must review, verify and commit the intended product snapshot before the manager treats GitHub as the complete handover.

## 1. Agree who controls the organization

- [ ] Assign two named organization owners for continuity. These people manage membership, billing and recovery; do not make every developer an owner.
- [ ] Confirm the business owner controls the organization recovery details and billing contact.
- [ ] Require two-factor authentication after checking members and automation can meet the requirement. Keep recovery codes in an approved password manager, outside GitHub issues.
- [ ] Change base repository permission to **No permission**, then grant access through teams. Confirm the effect on existing members first.
- [ ] Restrict repository creation, deletion, transfers and visibility changes to the appropriate administrators.
- [ ] Review installed GitHub Apps, OAuth access, outside collaborators and Actions policies. Give integrations access only to required repositories.
- [ ] Decide who owns security reports and offboarding. Record a primary and backup by name.

### Plan decision before moving the private repository

The destination currently uses GitHub Free. Check the plan before relying on private-repository branch protection, rulesets, Code Owner enforcement or deployment environment features. A private repository transferred to a Free organization can lose protected-branch capabilities. Choose a suitable plan or document the unavailable controls before transfer. Some environment reviewer gates require a higher plan even when private environments are available. [GitHub transfer documentation](https://docs.github.com/en/repositories/creating-and-managing-repositories/transferring-a-repository), [environment availability](https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments).

## 2. Set up repositories without splitting the product

| Repository | Purpose | Visibility / action |
| --- | --- | --- |
| `HACKWAYS-IN` | Existing product, four frontends, two Go engines, shared packages, infrastructure and private engineering docs | Recommended private destination name under HACKWAYS-OTA; not created or transferred by this guide |
| `.github` | Public organization profile and general community defaults | Already exists and is public. Do not put private runbooks, credentials, customer information or internal incident details here |

Additional repositories are not needed to match the existing architecture. Create a separate infrastructure or documentation repository only if there is an agreed ownership or access need; do not duplicate deployment files across repositories.

### Component map for assigning work

| Folder | Responsibility | Intended public destination |
| --- | --- | --- |
| `apps/public-web` | Marketing, Help and policy pages | `hackways.in` |
| `apps/core-web` | Attendee discovery, RSVP, My events, Spaces and event creation | `app.hackways.in` |
| `apps/organizer-web` | Management of existing events, guests and check-in | `organizer.hackways.in` |
| `apps/admin-web` | Restricted administration surface; current page shows service health | `admin.hackways.in`, VPN/IP restricted |
| `services/core-backend` | Go event lifecycle, registration, capacity, tickets and owner-scoped management | Internal; Nginx routes approved API paths |
| `services/global-backend` | Go authentication, sessions, upload signing/completion and email worker | Internal; Nginx routes approved API paths |
| `packages/ui`, `packages/api-client` | Shared frontend controls and TypeScript API client | Used by the separate apps |
| `deploy/`, `.github/`, `docs/` | Runtime configuration, CI and engineering documentation | Owned alongside the product |

Source version baseline: root lockfile resolves Next.js **14.2.35**, React **18.3.1**, TypeScript **5.9.3** and Tailwind **3.4.19**. CI/Docker select Node **22** and Go **1.23**; Go modules specify chi **5.1.0** and pgx **5.7.1**. PostgreSQL **16** is the CI/production target, not a verified live database version. Do not upgrade this stack as part of organization setup.

### Preferred migration: transfer the existing repository

1. Have the current repository owner and destination organization owner agree the transfer and final repository name.
2. Have engineering reconcile the local working tree with the intended release. Do not blanket-commit generated files, `.env`, logs, binaries or unrelated edits.
3. Record the latest source SHA, branch protections, open issues/PRs, Actions runs, package access and current release evidence.
4. Arrange a deployment pause with the infrastructure owner during the ownership change. Do not enable a new deployment path while another one can deploy the same VM.
5. The authorized owner transfers the repository using GitHub repository settings. Preserve visibility, history, issues and PRs. Do not delete the source or use a force push as a substitute.
6. Recheck teams, branch rules, Actions policies, secrets/variables, webhook configuration and package access after transfer. GitHub says secrets, deploy keys and webhooks remain associated with a transferred repository; their suitability and permissions still need review. [Transfer behavior](https://docs.github.com/en/repositories/creating-and-managing-repositories/transferring-a-repository).
7. Update the remote URL in developer clones and the deployment checkout. The proposed destination URL is `https://github.com/HACKWAYS-OTA/HACKWAYS-IN`; verify the transfer before using it as an existing link.
8. Review references to `varshith-dev/HACKWAYS-IN` in README, issue-template contact links, security policy and other documentation. Preserve historical evidence links where useful.
9. Review the GHCR registry change described in section 6. A repository transfer is not proof that container packages or host pull access moved correctly.

If transfer is unavailable, agree a history-preserving import with the source owner. A Git mirror alone does not carry issues, PR discussions, settings, secrets or Projects. Keep the old repository until the new source and operational access are accepted. Do not create two actively maintained product copies.

## 3. Create teams and assign access

The team names below are proposals, not claims that these teams already exist. Put people in teams based on responsibility, then add a backup for every critical area.

| Proposed team | Repository role | Responsibility | Production/cloud access |
| --- | --- | --- | --- |
| `engineering-leads` | Maintain | Cross-service decisions, review ownership, releases | Only named release owners as needed |
| `frontend` | Write | Four Next.js apps and shared UI | No routine database or deployment credentials |
| `backend` | Write | Go engines, SQL, contracts, migrations | Restricted nonproduction access; production through approved operations |
| `platform` | Write or Maintain | CI, containers, ingress, deployment and recovery | Named VM/registry/Azure operators with least privilege |
| `security` | Read, with explicit extra access when needed | Security review, vulnerability coordination, access audits | Restricted incident access |
| `support` | Triage | Issue classification and follow-up | No production secrets, private guest exports or general DB access |

Repository Write access applies across the repository. CODEOWNERS requests reviews; it does not create folder-level write authorization.

Copy the proposed [CODEOWNERS template](CODEOWNERS.example) into `.github/CODEOWNERS` only after creating the teams and granting eligible Code Owners explicit Write access. Validate with a real PR and enable required Code Owner review where the plan supports it. The template is deliberately inactive until then.

Multiple teams on one CODEOWNERS rule do not require approval from every listed team. If a security or database change needs two distinct sign-offs, define and verify that additional review policy explicitly.

## 4. Configure branches and pull requests

### Preserve existing branch meaning first

- `main`: current default branch. Do not declare it the production branch without production evidence.
- `develop`: current development branch and the branch the checked-in deployer accepts.
- `feature/<short-name>`, `fix/<short-name>`, `docs/<short-name>`: proposed short-lived branches merged by PR.

Changing deployment to `main` is a separate coordinated release-policy change. It requires updating the deployer configuration and verifying the trigger; changing the default branch alone does not do it.

### Rules to apply to main and develop, subject to plan support

- Require a PR and at least one approval; require the appropriate Code Owner for sensitive changes.
- Dismiss stale approvals when new commits arrive; protect against unreviewed final pushes where supported.
- Require resolved review conversations and successful checks on the current PR revision.
- Block force pushes and branch deletion; keep bypass access narrow and record emergency use.
- Require these actual PR check names from the current workflow:
  - `Repository Hygiene & Security Guard`
  - `Go Backend CI`
  - `Next.js Frontend CI`
- Do **not** require container image publish jobs on PRs: they run only on pushes and would leave PRs waiting for checks that never run.
- Confirm the exact check names after the first destination PR. Branch rules must match the names GitHub actually reports.
- Prefer squash merging for task PRs if the team agrees; preserve release tags and historical commits. Delete merged task branches.

Keep the existing PR template and bug/feature templates. They live in `.github/`. Update their repository links after transfer. Add release, incident and operational request templates if useful; avoid collecting credentials or personal guest data in issues.

## 5. Create one private organization Project

First check whether an appropriate Project already exists. The inspection account could not list Projects because it lacks `read:project`; the manager can check in the GitHub UI with their own authorized account. No token expansion is necessary just to use this guide.

Recommended Project name: **Hackways Delivery & Operations**. Link the product repository and keep operational/security-sensitive work private.

### Fields

| Field | Suggested choices / use |
| --- | --- |
| Status | Backlog, Ready, In progress, In review, Blocked, Done |
| Priority | P0: active outage/security; P1: release blocker; P2: normal work; P3: improvement |
| Area | Public web, Core web, Organizer web, Admin web, Core backend, Global backend, Shared UI/API, Platform, Documentation |
| Work type | Bug, Feature, Infrastructure, Security, Documentation, Operations |
| Owner | Actual GitHub assignee, plus a backup in the issue when operational |
| Target release | Agreed release identifier, not a guessed production version |
| Due date | Only for a real external deadline |

### Views and automation

- **Board:** group by Status for daily work.
- **Release readiness:** filter the target release, group by Area, show blockers and acceptance evidence.
- **Operations:** filter Infrastructure, Security and Operations work.
- **Backlog:** table sorted by Priority, with unassigned work visible.
- Auto-add new product issues/PRs when supported and configured. Confirm the repository filter so unrelated organization work is excluded.
- Auto-set opened items to Backlog and closed items to Done where appropriate. Do not auto-complete a bug merely because a linked PR exists.
- Use an issue, rather than only a draft item, for actionable work that needs comments, ownership and links to PRs.

Use [the setup backlog](SETUP_BACKLOG.csv) as the initial issue list. It is a planning CSV, not an executed import. Create issues manually or through an approved import process, then add them to the Project. Each row includes acceptance criteria.

### Suggested labels

`type:bug`, `type:feature`, `type:docs`, `type:infra`, `type:security`, `type:ops`; `priority:P0` through `priority:P3`; `area:public-web`, `area:core-web`, `area:organizer-web`, `area:admin-web`, `area:core-backend`, `area:global-backend`, `area:shared`, `area:platform`; `blocked`, `needs-triage`, `needs-production-verification`.

Use Project Status as the workflow state rather than maintaining a second competing set of status labels. Keep security vulnerability details in the restricted reporting workflow, not in ordinary public issues.

## 6. Configure Actions and container packages

Existing file: `.github/workflows/ci.yml`. Keep its implementation as the starting point.

This file is locally modified. The steps below describe the inspected working copy; the linked successful GitHub run proves only that its committed revision passed, not that all current local changes passed remotely.

### Current pipeline

1. Runs on pushes and PRs targeting `main` and `develop`.
2. Checks basic repository hygiene.
3. Starts isolated PostgreSQL 16 for backend CI, applies numbered migrations, checks formatting, runs `go vet` and race-instrumented Go tests.
4. Uses Node 22 and `npm ci` at the monorepo root. Checks shared UI, builds the API client and runs frontend type checks/lint.
5. On successful pushes to main/develop, builds six independent container images in GitHub-hosted runners and publishes them to GHCR.

The frontend `npm test` scripts currently run type checking and lint; do not describe them as browser/end-to-end suites. Next.js production compilation happens in the image build. No Next.js production builds may run on the constrained production VM.

### Registry names after an organization transfer

The workflow computes the lowercase repository owner. Under HACKWAYS-OTA, new image names would be:

```text
ghcr.io/hackways-ota/hackways-public-web
ghcr.io/hackways-ota/hackways-core-web
ghcr.io/hackways-ota/hackways-organizer-web
ghcr.io/hackways-ota/hackways-admin-web
ghcr.io/hackways-ota/hackways-core-backend
ghcr.io/hackways-ota/hackways-global-backend
```

These are expected names, not verified published packages. The current template points to `ghcr.io/varshith-dev`. After the first successful destination image build, confirm package ownership, repository linkage, Actions access and host pull permissions. Then coordinate any `DOCKER_REGISTRY` change with the deployment owner.

Images are tagged both `latest` and the full commit SHA. **Use the tested SHA/digest for a release.** Main and develop both publish `latest`, so that mutable tag is not a reliable production branch identifier. Keep the previous known-good images available for rollback and review package cleanup policies accordingly.

### Actions settings

- Restrict allowed Actions to the approved set; review third-party actions and pin reviewed versions/SHAs through a separate engineering change.
- Default workflow token permissions should be read-only; grant `contents: read` and `packages: write` only where required by the image publishing job.
- Use GitHub-hosted CI as configured. Do not place a general untrusted PR runner on the production VM.
- Set a budget/usage alert for private-repository Actions minutes and container storage.
- Add dependency/security maintenance work deliberately. The current hygiene job is not a full history secret scan or dependency vulnerability audit.

## 7. Add variables and arrange secrets with the infrastructure owner

Read [the sanitized environment inventory](.env.example). All credential fields are blank. It is an inventory, not a ready-to-run production environment file.

### Repository Actions variables used by the current workflow

| Name | Intended public production value | Purpose |
| --- | --- | --- |
| `NEXT_PUBLIC_SITE_URL` | `https://hackways.in` | Public site origin; also feeds the app's public origin build argument |
| `NEXT_PUBLIC_APP_URL` | `https://app.hackways.in` | Attendee app origin |
| `NEXT_PUBLIC_ORGANIZER_URL` | `https://organizer.hackways.in` | Management app origin |

The workflow explicitly supplies `/api/v1` for both browser API build arguments. Public values are bundled at build time. Changing a runtime container environment variable does not rewrite an already built Next.js browser bundle.

`NEXT_PUBLIC_SUPPORT_EMAIL` is supported by the public site with confirmed fallback `SUPPORT@hackota.com`; the current Dockerfile/workflow does not pass an override build argument. Adding an Actions variable alone will not wire a new override into that image.

### Secrets and access

- `GITHUB_TOKEN` is automatically issued for Actions; do not create or paste a personal token with that name.
- The current workflow does not require Azure credentials to build images. Do not copy database credentials or storage keys into repository secrets just to make CI green.
- Production runtime secrets belong in the approved Azure Key Vault/runtime injection flow. The exact live vault, identities and access assignments are not confirmed by this guide.
- The deployment webhook requires a `WEBHOOK_SECRET` matching the host's protected deployer configuration. Exchange and rotate it privately; never put it in an issue, Project field or guide.
- Give the host an approved identity/credential to pull private GHCR images and read the private Git checkout. Record owner and expiry, not token values.
- Record secret purpose, location, owner, rotation procedure and last verification date. Nobody needs to disclose a secret value to the GitHub manager for the manager to document it.

Use organization secrets only when multiple approved repositories need them; restrict repository access. A GitHub environment is effective only when the relevant job references it. [GitHub secret scoping](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets).

## 8. Connect deployment only after engineering preflight

The existing repository describes Azure VM + Docker Compose + Nginx, not a Vercel deployment. The configured host deployer listens for an HMAC-signed GitHub **workflow_run** event, with workflow name `CI Pipeline`, action `completed`, conclusion `success`, branch `develop` and a full commit SHA. It runs the deploy script with that SHA as `DEPLOY_IMAGE_TAG`.

The checked-in public webhook path is `/api/deploy/webhook`. The live endpoint and routing must be verified by the infrastructure owner. Current GitHub inspection found no repository webhook, so do not assume this link is installed or active. Do not substitute a push webhook; the handler intentionally ignores it.

### Release blockers observed in this working copy

1. `deploy/docker/docker-compose.prod.yml` fails YAML parsing at line 183: server API URL entries are indented incorrectly under `core-web`; the organizer block has the same pattern. Engineering must fix and run `docker compose ... config --quiet` before deployment. Do not deploy this file as-is.
2. Local uncommitted/untracked code is not represented by the current GitHub SHA. Agree the actual product source snapshot.
3. Exact live production SHA, image digests, migration version, infrastructure owner and service readiness are not supplied.
4. Current source CI expects origin variables, but none were returned from repository-level Actions variables. Confirm the intended build configuration and any organization-level values.
5. Database topology differs across documents: optional local PostgreSQL container versus managed Azure PostgreSQL. Backup/restore scripts reference `hackways-postgres`; do not assume they back up a managed server.
6. Azure Blob/ACS integrations exist in code, but local external uploads/email delivery are unverified without credentials. Do not mark those integrations operational from CI alone.

Track these in the Project and have engineering close them with evidence. This guide does not modify the production stack or enable deployment.

### Production evidence to request from the operator

- Running image reference and immutable digest for each of the six services.
- Corresponding Git SHA, source branch, CI run URL, release date and release owner.
- Database provider/version and migration version, with no connection strings.
- Health/readiness result, TLS certificate coverage/renewal, private GHCR pull check, and whether the webhook actually triggered the rollout.
- Last successful backup, isolated restore evidence and the previous schema-compatible rollback release.

Do not label a release Deployed because a container image was published. Do not run migration downs as routine rollback; at least one workflow migration is deliberately forward-only. Engineering must approve schema-compatible image rollback or a forward repair.

If the team later adds staging/production GitHub environments or manual approvals, verify plan support and wire them into the actual deploy path. An environment badge alone will not gate the separate host webhook.

## 9. Put the essential documentation in the product repository

| Document / source | Owner to assign | Manager's task |
| --- | --- | --- |
| `README.md` | Engineering lead | Verified setup and links to the right apps |
| `AGENTS.md`, `docs/ARCHITECTURE.md` | Architecture lead | Preserve service boundaries and production constraints |
| `docs/backend/README.md`, `workflow.md`, `scenarios.md` | Backend lead | Current API behavior, failures and integration limitations |
| `docs/API.md`, `docs/DATABASE.md`, migrations and sqlc configs | Backend/database owner | Versioned contract and migration ownership |
| `docs/DEPLOYMENT.md`, `deploy/` | Platform owner | Validated deployment, rollback, DNS/TLS and recovery runbooks |
| `docs/SECURITY.md`, `.github/SECURITY.md` | Security owner | Private reporting, incident contact and secret handling |
| `docs/TESTING.md`, `.github/workflows/ci.yml` | Engineering/QA lead | Actual CI checks and verification limits |
| `packages/ui/README.md`, `docs/PUBLIC_CONTENT.md` | Frontend/content owner | Shared controls, clean routes, support and policy facts |
| `docs/planning/WORKLOGS.md` | All contributors | Evidence of what changed and commands actually run |

Some older plans describe target capabilities rather than implemented features. Backend current-state documentation and actual handlers take precedence when preparing the product inventory. Paid checkout, full admin moderation, password recovery, self-service account deletion/export and offline scanning must not be advertised as delivered without implementation evidence. A legal operator/address, retention schedule and minimum-age policy still require business confirmation. No root LICENSE file was present during inspection; ask the owner to choose the intended licensing position rather than copying an open-source license without approval.

## 10. Record access beyond GitHub

Keep an access register with provider, resource, named primary/backup, role, recovery owner and verification date. Store credentials separately.

| Resource | Access the handover must establish |
| --- | --- |
| GitHub organization, repo and GHCR | Owners, maintainers, security contact, package read/write principals |
| Azure subscription/resource group and VM | Business billing owner, restricted platform operators, recovery route |
| PostgreSQL | Actual hosting choice, DBA/migration operator, least-privilege app accounts, backup owner |
| Key Vault / runtime secret injection | Vault administrator versus application secret-read identity |
| Blob Storage | `media-uploads` and private `backups`, CORS/public read policy, lifecycle/cleanup ownership |
| ACS email | Verified sender/domain, sending identity and delivery monitoring owner |
| Domain registrar and Cloudflare | Domain renewal/billing plus DNS administration; these may be different accounts |
| Support mailbox | Owner and backup for `SUPPORT@hackota.com`; email address confirmation is not proof of mailbox access |
| TLS, monitoring and incidents | Certificate renewal, alert routing, incident coordinator, escalation channel |

Authoritative NS lookup on 20 October 2026 returned `norah.ns.cloudflare.com` and `hassan.ns.cloudflare.com` for `hackways.in`. This confirms Cloudflare DNS delegation at that time; registrar ownership, proxy mode, live A/CNAME records, Azure resource IDs and production access remain to be supplied by their owners.

## 11. Acceptance checklist for the manager

- [ ] Approved org plan and access policy; two named owners and recovery custody established.
- [ ] Intended product source, including local-only work, reviewed and safely recorded in Git.
- [ ] Product repository transferred/imported once, private visibility preserved and source history retained.
- [ ] Teams, named owners/backups and inactive-to-active CODEOWNERS rollout verified.
- [ ] Branch protection/rules actually enforce all three PR checks and reviews, subject to the purchased plan.
- [ ] Project linked to the repo; labels, views, real issues, owners and blockers visible.
- [ ] Destination PR passes real checks; destination push builds all six images.
- [ ] Public origin variables and private package pull access verified.
- [ ] Secrets inventoried with owners and rotation procedure, without putting values in GitHub content.
- [ ] Deployment configuration validated; webhook and release branch agreed with the operator.
- [ ] Live release SHA/images/migration version and rollback evidence recorded.
- [ ] DNS/TLS, database, storage, email, backups and incident access have named owners.
- [ ] Offboarding procedure tested: remove team/org access, review outside collaborators/apps/deploy keys and revoke affected external credentials.
- [ ] Product lead, engineering lead, platform owner and GitHub manager sign off with date and evidence links.

**Guide preparation did not create or transfer repositories, create Projects/teams, change remote permissions, reveal secret values or deploy the application.** Those setup actions belong to the authorized manager and infrastructure owners following this checklist.
