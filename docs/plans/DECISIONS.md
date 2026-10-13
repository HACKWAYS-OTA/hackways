# Architectural Decisions Index (ADR Register)

## 1. Overview

This document records the architectural decisions governing Hackways. Decisions are classified as **APPROVED** (formally binding across all agents and contributors), **PROPOSED** (under review pending project owner approval), or **REJECTED**.

---

## 2. Architectural Decisions Register

| Decision ID | Title | Status | Date | Affected Phases | Authoritative Source |
|:---:|:---|:---:|:---:|:---:|:---|
| **AD-001** | 6-Service Decoupled Monorepo Architecture | `APPROVED` | 2026-10-10 | Phase 1, Phase 2, Phase 6, Phase 8 | [`docs/planning/MEMORY.md`](../planning/MEMORY.md#L7-L10), [`AGENTS.md`](../../AGENTS.md#L94-L117) |
| **AD-002** | Go 1.23+ Engine with Chi Router and `log/slog` | `APPROVED` | 2026-10-10 | Phase 2, Phase 3, Phase 4 | [`docs/planning/MEMORY.md`](../planning/MEMORY.md#L11-L14) |
| **AD-003** | PostgreSQL Connection Pooling Budget (Max 60 Conns) | `APPROVED` | 2026-10-10 | Phase 3, Phase 8 | [`docs/planning/MEMORY.md`](../planning/MEMORY.md#L15-L18), [`AGENTS.md`](../../AGENTS.md#L202-L215) |
| **AD-004** | Cryptographic Host-Scoped Cookie Sessions (`__Host-`) | `APPROVED` | 2026-10-10 | Phase 4 | [`docs/planning/MEMORY.md`](../planning/MEMORY.md#L19-L22), [`AGENTS.md`](../../AGENTS.md#L224-L232) |
| **AD-005** | Pessimistic Row-Level Locking (`FOR UPDATE` 2000ms timeout) | `APPROVED` | 2026-10-10 | Phase 3, Phase 7 | [`docs/planning/MEMORY.md`](../planning/MEMORY.md#L23-L26), [`AGENTS.md`](../../AGENTS.md#L182-L200) |
| **AD-006** | Off-Node Next.js CI Builds & Standalone Output | `APPROVED` | 2026-10-10 | Phase 2, Phase 6, Phase 8 | [`docs/planning/MEMORY.md`](../planning/MEMORY.md#L27-L30), [`AGENTS.md`](../../AGENTS.md#L125-L135) |
| **AD-007** | Strict Prohibition of `lucide-react` across Hackways | `APPROVED` | 2026-10-10 | Phase 2, Phase 6 | [`AGENTS.md` Rule 1.2 & 3.12](../../AGENTS.md#L30) |
| **AD-008** | Zero-Disk Media Architecture via Azure Blob SAS Tokens | `APPROVED` | 2026-10-10 | Phase 5, Phase 8 | [`docs/ARCHITECTURE_PLAN.md`](../ARCHITECTURE_PLAN.md#L113-L158), [`AGENTS.md`](../../AGENTS.md#L125-L140) |
| **AD-009** | Azure Communication Services via PostgreSQL Outbox | `APPROVED` | 2026-10-10 | Phase 3, Phase 5 | [`docs/ARCHITECTURE_PLAN.md`](../ARCHITECTURE_PLAN.md#L161-L203), [`AGENTS.md`](../../AGENTS.md#L143-L158) |
| **AD-010** | Schema Table Naming Reconciliation (`rsvps` & `email_outbox`) | `APPROVED` | 2026-10-13 | Phase 1, Phase 3 | Formally approved to reconcile with AGENTS.md |
| **AD-011** | Directory Transition Strategy (Scaffold 6 Services in Phase 2) | `APPROVED` | 2026-10-13 | Phase 1, Phase 2 | Formally approved to enforce 6-service decoupled monorepo |
| **AD-012** | Removal of Untracked Misplaced `LICENSE` File | `APPROVED` | 2026-10-13 | Phase 1 | Purged accidental OpenAI documentation (verified) |

---

## 3. Decision Summaries & Rationale

### AD-011: Directory Transition Strategy (Scaffold 6 Services in Phase 2)
- **Decision:** Establish `apps/public-web`, `apps/core-web`, `apps/organizer-web`, `apps/admin-web`, `services/core-backend`, and `services/global-backend` as independent directories in Phase 2 before authoring functional code.
- **Rationale:** Strictly enforces [`AGENTS.md` Section 2.1](../../AGENTS.md#L94-L117) and avoids building monolithic code in `apps/web` or `services/api` that would require later refactoring.

### AD-012: Removal of Untracked `LICENSE`
- **Decision:** Delete untracked `LICENSE` file containing OpenAI Apps SDK guidelines.
- **Rationale:** The repository was formally marked proprietary and confidential in commit `620d316`. The untracked file was an accidental artifact.
