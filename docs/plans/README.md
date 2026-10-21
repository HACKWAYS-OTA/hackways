# Hackways engineering plans

[Repository overview](../../README.md) · [Current architecture](../ARCHITECTURE.md) · [Backend behavior](../backend/README.md) · [Dated worklog](../planning/WORKLOGS.md)

## Read this first

These phase documents preserve planning and implementation history. They are not current production runbooks, and a historical VERIFIED label does not certify the present release. References to services/api or apps/web describe superseded layouts; the current product has four Next.js apps and two Go backends. Reconcile every proposed action with current source, AGENTS.md, and dated evidence before execution.

Start with [MASTER_PLAN.md](MASTER_PLAN.md) for the roadmap, [RISKS_AND_BLOCKERS.md](RISKS_AND_BLOCKERS.md) for recorded risks, and [DEPLOYMENT.md](../DEPLOYMENT.md) for operational prerequisites. Production readiness remains gated on the reviewed release, database/migration recovery, and runtime evidence.

## 1. Purpose & Overview

This directory houses the **canonical, repository-backed engineering execution system** for the **Hackways** platform. It provides persistent, dependency-ordered technical specifications, task registers, architecture context, and verification protocols across the original eight phases and later extensions.

Every engineering task—whether executed by human contributors or autonomous AI agents—must be recorded, tracked, and verified through this system. **No phase or task may be declared complete without verifiable evidence executed against the live repository.**

---

## 2. Directory Structure & Architecture

```text
docs/plans/
├── README.md                      # This overview and operating manual
├── MASTER_PLAN.md                 # End-to-end execution sequence, milestones, and exit gates
├── ARCHITECTURE_CONTEXT.md        # Historical architectural planning context
├── DEPENDENCY_MAP.md              # Cross-phase dependency graph and blocking decision registers
├── DECISIONS.md                   # Architectural Decision Record (ADR) index and status
├── RISKS_AND_BLOCKERS.md          # Prioritized register of technical risks, conflicts, and blockers
│
├── phase1/                        # Phase 1: Repository Discovery and Architecture Alignment
│   ├── plan.md                    # Authoritative execution plan
│   ├── task.md                    # Executable task register with stable IDs
│   ├── status.md                  # Real-time progress record and verified outcomes
│   ├── context.md                 # Autonomous agent pre-context package
│   ├── architecture.md            # Phase-specific technical architecture
│   ├── implementation-guide.md    # Step-by-step engineering instructions
│   ├── verification.md            # Evidence requirements and test commands
│   └── handoff.md                 # Next-agent entry point and resumption protocol
│
├── phase2/                        # Phase 2: Dependency and Toolchain Baseline
│   └── [8 mandatory phase files]
├── phase3/                        # Phase 3: Database and SQL Engineering
│   └── [8 mandatory phase files]
├── phase4/                        # Phase 4: Security Foundations
│   └── [8 mandatory phase files]
├── phase5/                        # Phase 5: Azure Integration Foundations
│   └── [8 mandatory phase files]
├── phase6/                        # Phase 6: Frontend Foundations
│   └── [8 mandatory phase files]
├── phase7/                        # Phase 7: Testing and CI Enforcement
│   └── [8 mandatory phase files]
├── phase8/                        # Phase 8: Production Deployment and Operational Readiness
│   └── [8 phase files]
├── phase9/                        # Dual Go backend separation
│   └── [8 phase files]
└── phase10/                       # Managed cloud architecture extension
    └── plan.md, task.md, status.md, architecture.md
```

---

## 3. Authoritative Source Documents

All planning files in this directory derive their authority from and must strictly adhere to the project's root canonical specifications:

1. **[`AGENTS.md`](../../AGENTS.md)**: Canonical, permanent engineering specification and quality enforcement rules. Rules in `AGENTS.md` take absolute precedence over any plan document.
2. **[`docs/ARCHITECTURE_PLAN.md`](../ARCHITECTURE_PLAN.md)**: Target production architecture for the 6-service decoupled monorepo on a single Azure VM (2 vCPU, 4 GB RAM, 50 GB NVMe SSD) with Azure Blob Storage and Azure Communication Services.
3. **[`docs/planning/CONTEXT.md`](../planning/CONTEXT.md)**: Current repository and engineering milestones baseline.
4. **[`docs/planning/MEMORY.md`](../planning/MEMORY.md)**: Durable architectural decisions log (AD-001 through AD-006).

---

## 4. Phase Status & Task State Machine

### Phase States
- `PLANNED`: Requirements defined; prerequisites not yet complete.
- `IN_PROGRESS`: Active execution underway; task checklist active.
- `BLOCKED`: Blocked on an explicit decision, dependency, or upstream phase.
- `VERIFIED`: All acceptance criteria and verification commands passed with logged evidence.

### Task Status Vocabulary
Every task in `task.md` must strictly use one of the following states:
- `NOT_STARTED`: Ready for implementation once prerequisites are met.
- `IN_PROGRESS`: Currently being drafted, coded, or configured.
- `BLOCKED`: Awaiting prerequisite task completion or architectural decision.
- `IN_REVIEW`: Implementation completed; undergoing linting, testing, or review.
- `VERIFIED`: Confirmed passing via executed test/build command with recorded output.
- `CANCELLED`: Formally superseded or rejected by documented architectural decision.

---

## 5. Agent Resumption Protocol

When a new AI agent or developer initiates a session:
1. **Never guess the state:** Read this index, MASTER_PLAN.md, the latest worklog entries, and the relevant current source. Do not infer the active release solely from phase labels.
2. **Inspect the active phase:** Open `phaseX/handoff.md` and `phaseX/status.md` first.
3. **Verify current repo facts:** Never assume files or packages exist based on text descriptions alone. Execute safe read-only inspection commands (`git status`, directory listings, toolchain versions).
4. **Select next action:** Pick the highest-priority `NOT_STARTED` task in `phaseX/task.md` whose prerequisites are `VERIFIED`.
5. **Update state on pause:** Update `status.md` and `handoff.md` before concluding the session.
