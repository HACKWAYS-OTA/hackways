# Phase 1 Implementation Guide: Discovery & Architecture Alignment

## 1. Prerequisites & Toolchain Setup
- Ensure Go 1.23+ and Node.js 20+ are installed.
- Ensure PowerShell or bash is available.
- Work from repository root (`C:\Users\ADMIN\Desktop\RSVP-P`).

---

## 2. Step-by-Step Execution Sequence

### Step 1: Clean Untracked Artifacts
1. Inspect git status:
   ```bash
   git status
   ```
2. Verify that untracked `LICENSE` contains OpenAI Apps SDK guidelines.
3. Remove the untracked file:
   ```bash
   rm LICENSE
   ```

### Step 2: Establish Git Line Ending Normalization
1. Create `.gitattributes` at root with content:
   ```text
   * text=auto eol=lf
   *.go text eol=lf
   *.ts text eol=lf
   *.tsx text eol=lf
   *.sql text eol=lf
   ```
2. Re-test `gofmt -l .` in `services/api` to verify clean output.

### Step 3: Author and Validate the Execution Plan System
1. Verify that `docs/plans/README.md` and master plans exist.
2. Verify that `docs/plans/phase1/` through `docs/plans/phase8/` each contain all 8 mandatory files:
   - `plan.md`
   - `task.md`
   - `status.md`
   - `context.md`
   - `architecture.md`
   - `implementation-guide.md`
   - `verification.md`
   - `handoff.md`

### Step 4: Verification Checkpoints
1. Run `go test ./... -v -count=1` inside `services/api`.
2. Confirm zero regressions.
3. Update `status.md` with timestamps and results.

---

## 3. What Must NOT Be Changed in Phase 1
- Do not modify any Go logic in `cmd/api/main.go` or `internal/platform/`.
- Do not write new SQL migration files.
- Do not install npm dependencies or build Next.js.
