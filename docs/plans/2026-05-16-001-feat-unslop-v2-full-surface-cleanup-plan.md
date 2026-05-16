---
title: "feat: unslop v2 full-surface cleanup workflow"
type: feat
status: active
date: 2026-05-16
---

# feat: unslop v2 full-surface cleanup workflow

## Summary

Upgrade `/unslop` from a three-pass slop detector into a full-surface cleanup workflow. The default command reports every slop lane: code hygiene, comments/docs, frontend/design, and codebase architecture. Fix mode stays safe by default, while `fix all` applies high-priority auto-fix-eligible fixes across all lanes.

The architecture lane should borrow Matt Pocock's "improve codebase architecture" framing: find deepening candidates, evaluate module depth, seam placement, locality, leverage, interface shape, fake adapters, and testability. ATV keeps its existing strengths: comment cleanup, code hygiene cleanup, and frontend design slop detection.

---

## Problem Frame

Current `/unslop` is useful, but narrow:

- It reports code slop, comment rot, and design slop.
- `fix` only handles low-risk code/comment cleanup.
- It does not report architecture slop as a first-class lane.
- It does not expose priority/risk/effort in the report.
- It does not have a complete forgiving command contract for slash-command usage.

The new product direction:

- `/unslop` should be the broad X-ray across all slop surfaces.
- `/unslop fix` should stay conservative.
- `/unslop fix all` and `/unslop fix --all` should apply high-priority auto-fix-eligible fixes across code hygiene, comments/docs, architecture, and frontend/design.
- Users can scope fixes by lane, priority, path, and candidate ID.

---

## Requirements

- R1. `/unslop` produces a full read-only report across four lanes:
  - code hygiene
  - comments/docs
  - frontend/design
  - architecture
- R2. Every finding includes:
  - stable ID
  - lane
  - file and line where possible
  - finding
  - priority: high, medium, low
  - risk: low, medium, high
  - effort: small, medium, large
  - auto-fix eligibility
  - suggested next command
- R3. `/unslop fix` applies only safe code hygiene and comments/docs fixes.
- R4. `/unslop fix all` and `/unslop fix --all` apply high-priority auto-fix-eligible fixes across all four lanes.
- R5. Lane-specific fixes work with both flag and natural slash-command token forms:
  - `/unslop fix architecture`
  - `/unslop fix --architecture`
  - `/unslop fix frontend medium`
  - `/unslop fix --frontend --medium`
- R6. Priority selectors are exact, not cumulative:
  - no priority token means high
  - `medium` means medium only
  - `low` means low only
  - `high` means high only
- R7. Multiple lane selectors are rejected unless the user chose `all`.
- R8. Path scope is supported for reports and fixes:
  - `/unslop src/auth`
  - `/unslop fix architecture src/auth`
  - `/unslop fix all app/components`
- R9. Candidate ID scope is supported:
  - `/unslop fix architecture A1`
  - `/unslop fix frontend F2,F4`
- R10. Architecture reporting uses Matt-style deepening language:
  - module
  - interface
  - implementation
  - depth
  - seam
  - adapter
  - leverage
  - locality
  - test surface
- R11. Architecture fixes are only auto-fix eligible when scoped, testable, and low enough risk.
- R12. Frontend fixes are only auto-fix eligible when visual behavior can be verified or the change is clearly mechanical, such as missing focus/hover state additions that follow existing conventions.
- R13. The skill documentation and duplicated plugin/template copies remain in sync.
- R14. Existing Go tests and scaffold parity tests continue to pass.

---

## Current Repo Surfaces

### Skill copies that must stay in sync

- `pkg/scaffold/templates/skills/unslop/SKILL.md`
- `plugins/atv-skill-unslop/skills/unslop/SKILL.md`
- `plugins/atv-pack-quality/skills/unslop/SKILL.md`
- `plugins/atv-everything/skills/unslop/SKILL.md`

These copies currently have the same SHA-256 hash. Edit the scaffold template first, then mirror the exact content to plugin copies.

### Prompt shim

- `.github/prompts/unslop.prompt.md`

Current issue noticed during planning: this prompt says to invoke `.github/skills/unslop/SKILL.md`, but this checkout does not contain `.github/skills/unslop/SKILL.md`. Fix or clarify the prompt shim as part of this work so it points at the actual installed skill surface or uses harness-neutral "invoke the unslop skill" wording without a missing path.

### Pipeline references

- `plugins/atv-skill-lfg/skills/lfg/SKILL.md`
- `plugins/atv-pack-shipping/skills/lfg/SKILL.md`
- `plugins/atv-pack-shipping/skills/slfg/SKILL.md`
- `README.md`
- `docs/marketplace.md`
- `.github/plugin/marketplace.json`

These references should be updated only where the command semantics changed.

---

## Command Contract

### Reports

```text
/unslop
```

Full report across all four lanes.

```text
/unslop frontend
/unslop --frontend
/unslop architecture
/unslop --architecture
/unslop comments
/unslop --comments
/unslop hygiene
/unslop --hygiene
```

Filtered report for one lane.

### Fixes

```text
/unslop fix
```

Safe default. Applies safe code hygiene and comments/docs fixes only.

```text
/unslop fix all
/unslop fix --all
```

High-priority, auto-fix-eligible fixes across all four lanes:

- code hygiene
- comments/docs
- frontend/design
- architecture

```text
/unslop fix architecture
/unslop fix --architecture
```

High-priority, auto-fix-eligible architecture fixes only.

```text
/unslop fix architecture medium
/unslop fix --architecture --medium
```

Medium-priority, auto-fix-eligible architecture fixes only.

```text
/unslop fix architecture low
/unslop fix --architecture --low
```

Low-priority, auto-fix-eligible architecture fixes only.

```text
/unslop fix frontend
/unslop fix --frontend
/unslop fix frontend medium
/unslop fix --frontend --medium
/unslop fix frontend low
/unslop fix --frontend --low
```

Frontend/design equivalents.

```text
/unslop fix comments
/unslop fix --comments
/unslop fix hygiene
/unslop fix --hygiene
```

Narrow safe cleanup lanes.

### Aliases

Treat these as equivalent:

| Canonical | Aliases |
|-----------|---------|
| `all` | `--all` |
| `architecture` | `--architecture`, `arch` |
| `frontend` | `--frontend`, `front-end`, `design` |
| `comments` | `--comments`, `comment`, `docs`, `documentation` |
| `hygiene` | `--hygiene`, `code-hygiene` |
| `high` | `--high` |
| `medium` | `--medium` |
| `low` | `--low` |
| `dry-run` | `--dry-run`, `preview` |

### Rejected combinations

Reject multiple judgment lanes in one command:

```text
/unslop fix frontend architecture
/unslop fix --frontend --architecture
```

Return:

```text
Choose one lane, or use all.

Examples:
- /unslop fix frontend
- /unslop fix architecture
- /unslop fix all
```

Do not support a lane-local `--all` meaning "all priorities" in v2. `all` means all lanes with high-priority eligible fixes. If all-priority lane cleanup is needed later, add a new unambiguous token such as `all-priorities`.

---

## Report Format

```markdown
# /unslop Report

Scope: changed files since origin/main
Fix default: safe hygiene + comments only
Fix all: high-priority eligible fixes across all lanes

## Executive Summary

| Lane | Findings | High | Medium | Low | Auto-fix eligible | Recommended next |
|------|----------|------|--------|-----|-------------------|------------------|
| Hygiene | 4 | 1 | 2 | 1 | 3 | /unslop fix hygiene |
| Comments/docs | 8 | 2 | 4 | 2 | 7 | /unslop fix comments |
| Frontend/design | 3 | 1 | 2 | 0 | 1 | /unslop fix frontend |
| Architecture | 3 | 1 | 1 | 1 | 1 | /unslop fix architecture |

## Findings

| ID | Lane | File | Finding | Priority | Risk | Effort | Auto-fix | Next command |
|----|------|------|---------|----------|------|--------|----------|--------------|
| H1 | Hygiene | src/foo.ts:42 | Dead helper never used | High | Low | Small | Yes | /unslop fix hygiene H1 |
| C1 | Comments | src/foo.ts:12 | Comment restates function name | Low | Low | Small | Yes | /unslop fix comments C1 |
| F1 | Frontend | app/card.tsx:70 | Generic card grid lacks hierarchy | High | Medium | Medium | Yes | /unslop fix frontend F1 |
| A1 | Architecture | src/session.ts | Session rules leak into 4 callers | High | Medium | Medium | Yes | /unslop fix architecture A1 |
```

### ID prefixes

- `H` for hygiene
- `C` for comments/docs
- `F` for frontend/design
- `A` for architecture

---

## Architecture Lane Design

### Architecture detector prompts

Add a fourth analysis pass: Architecture Slop Detector.

It should report deepening candidates, not generic refactor complaints.

Detect:

- shallow modules where the interface is nearly as complex as the implementation
- pass-through wrappers that hide nothing
- fake adapters with one implementation and no real seam
- caller knowledge leakage
- duplicated orchestration across call sites
- interfaces that expose implementation detail
- frontend components whose props mirror internals instead of product concepts
- tests that mock internals instead of exercising a stable public interface
- modules that fail the deletion test

### Matt-style candidate output

```json
{
  "pass": "architecture-slop",
  "findings": [
    {
      "id": "A1",
      "files": ["src/auth/session.ts", "src/auth/cookies.ts"],
      "problem": "Callers must know cookie parsing order and session fallback behavior.",
      "proposed_fix": "Create a deeper Session interface that owns load, refresh, and clear behavior.",
      "deepening": {
        "depth": "Callers stop coordinating cookie and session details.",
        "locality": "Session rules move into one module.",
        "leverage": "Four callers shrink to one stable interface.",
        "test_surface": "Tests can exercise session behavior through the public interface."
      },
      "priority": "high",
      "risk": "medium",
      "effort": "medium",
      "auto_fix_eligible": true,
      "verification": ["targeted unit tests for session load/refresh/clear", "existing auth tests"]
    }
  ]
}
```

### Architecture auto-fix eligibility

An architecture finding is auto-fix eligible only when all are true:

- priority matches the requested priority selector
- risk is low or medium
- effort is small or medium
- affected files are inside the requested scope
- behavior can be locked by existing tests or a narrow new test
- the fix reduces caller knowledge or removes a fake seam without changing product behavior
- the skill can name the exact verification commands before editing

High-risk or large architecture findings must stay report-only and point users to a candidate workflow.

---

## Frontend Lane Design

Frontend/design findings remain part of default `/unslop`.

Detect:

- generic card grids with no hierarchy
- purple/blue default gradients without brand rationale
- missing hover/focus/active states
- generic emoji badges and redundant eyebrow/title/description stacks
- design-system bypasses
- overly uniform radius/shadow treatment
- layout rhythm that feels template-generated
- component props that expose internals rather than product concepts

Frontend auto-fix eligibility requires one of:

- purely mechanical accessibility/design-system cleanup
- existing design-system convention clearly dictates the change
- screenshot/browser verification is possible and included in the fix report

If visual verification is not possible, skip the finding and report the reason.

---

## Fix Workflow

```text
parse arguments
  -> determine report or fix mode
  -> determine lane selector or all lanes
  -> determine priority selector, default high for lane/all fixes
  -> determine path scope
  -> determine candidate IDs
  -> reject conflicting lanes

determine scope
  -> default changed files since merge-base
  -> or explicit files/directories

run report passes
  -> hygiene
  -> comments/docs
  -> frontend/design when UI files are present
  -> architecture

merge findings
  -> assign stable IDs
  -> score priority/risk/effort
  -> determine auto-fix eligibility

if report mode
  -> print full report

if fix mode
  -> select eligible findings
  -> produce cleanup plan
  -> run behavior lock checks or name manual verification
  -> apply fixes one lane at a time
  -> run targeted verification after each lane
  -> report changed files, simplifications, verification, skipped findings, remaining risks
```

ASCII flow:

```text
          /unslop args
               |
          parse tokens
               |
      +--------+---------+
      |                  |
   report              fix
      |                  |
 run all selected    run report first
 analysis passes        |
      |             select eligible
 print ranked report    |
                    behavior lock
                         |
                    apply lane fixes
                         |
                    verify and report
```

---

## Implementation Units

### U1. Rewrite argument parsing and command contract in the skill doc

**Goal:** Make the command grammar explicit and forgiving.

**Files:**

- Modify: `pkg/scaffold/templates/skills/unslop/SKILL.md`
- Mirror to:
  - `plugins/atv-skill-unslop/skills/unslop/SKILL.md`
  - `plugins/atv-pack-quality/skills/unslop/SKILL.md`
  - `plugins/atv-everything/skills/unslop/SKILL.md`

**Approach:**

- Replace the current three-token table with the v2 command contract.
- Add alias table.
- Add conflict and rejection rules.
- Add priority selector semantics.
- Add path and candidate ID rules.

**Verification:**

- Manual read-through confirms every example command resolves to one parse outcome.
- No command uses `all` with two meanings.

### U2. Add the architecture lane

**Goal:** Add a fourth analysis pass using Matt-style deepening candidates.

**Files:**

- Modify all `unslop/SKILL.md` copies listed in U1.

**Approach:**

- Add "Architecture Slop Detector" alongside existing code/comment/design passes.
- Define candidate JSON schema.
- Add architecture smell checklist.
- Add auto-fix eligibility gate.
- Add skip rules for high-risk or large architecture candidates.

**Verification:**

- A reader can distinguish architecture findings from hygiene findings.
- Every architecture finding has priority/risk/effort and a verification path.

### U3. Upgrade the report format

**Goal:** Make `/unslop` a decision surface, not a lint dump.

**Files:**

- Modify all `unslop/SKILL.md` copies.

**Approach:**

- Add executive summary table.
- Add unified finding table.
- Add lane-specific detail sections.
- Require stable IDs.
- Require "skipped, not eligible" sections for fix mode.

**Verification:**

- Report format shows impact/risk/effort at a glance.
- Report includes next command for every finding.

### U4. Define fix selection and execution rules

**Goal:** Make mutation safe and predictable.

**Files:**

- Modify all `unslop/SKILL.md` copies.

**Approach:**

- Define `/unslop fix` safe default.
- Define `/unslop fix all` and `/unslop fix --all`.
- Define lane-specific priority behavior.
- Add cleanup plan before edits.
- Add verification after each lane.
- Add dry-run/preview support.

**Verification:**

- `/unslop fix` cannot change frontend or architecture.
- `/unslop fix all` can change frontend and architecture, but only high-priority auto-fix-eligible findings.
- Multiple lanes without `all` are rejected.

### U5. Fix the prompt shim

**Goal:** Remove the missing `.github/skills/unslop/SKILL.md` path reference.

**Files:**

- Modify: `.github/prompts/unslop.prompt.md`

**Approach:**

- Replace the brittle path instruction with harness-neutral wording:
  - "Invoke the installed `unslop` skill and forward arguments verbatim."
- If the harness requires a path, point to a real repo path or document that the prompt assumes installed skill availability.

**Verification:**

- No prompt references a missing `.github/skills/unslop/SKILL.md` path.

### U6. Update pipeline docs and README references

**Goal:** Keep user-facing docs aligned with new semantics.

**Files:**

- Modify: `README.md`
- Modify: `docs/marketplace.md` if needed
- Modify: `plugins/atv-skill-lfg/skills/lfg/SKILL.md` if the intended LFG behavior changes
- Modify: `plugins/atv-pack-shipping/skills/lfg/SKILL.md` if needed
- Modify: `plugins/atv-pack-shipping/skills/slfg/SKILL.md` if needed

**Approach:**

- README should explain:
  - `/unslop` reports all lanes.
  - `/unslop fix` safe cleanup.
  - `/unslop fix all` high-priority eligible cleanup across all lanes.
  - lane filters and priority examples.
- Leave LFG using `/unslop fix` unless we explicitly want the autonomous pipeline to mutate frontend and architecture by default.

**Recommendation:** keep LFG on `/unslop fix` for v2. Do not silently expand autonomous shipping pipelines to architecture/frontend changes. Let users opt into `/unslop fix all` manually.

**Verification:**

- README examples match the skill's command contract.
- LFG semantics are intentionally documented.

### U7. Add or update tests

**Goal:** Protect skill sync and command documentation.

**Files:**

- Modify or add Go tests under `pkg/scaffold/` only if existing tests do not catch drift.

**Approach:**

- Run existing tests first:
  - `go test ./...`
- If parity already ensures skill copies, no new test required.
- If not, add a narrow parity test that asserts all unslop copies match the template.
- Add a prompt-shim path test only if prompt paths are already tested elsewhere. Otherwise keep verification manual.

**Verification:**

- `go test ./...`
- `git diff --check`
- `grep -RIn ".github/skills/unslop/SKILL.md" .github/prompts README.md docs plugins pkg || true`

---

## Test Plan

### Static verification

```bash
go test ./...
git diff --check
```

### Skill copy parity

```bash
sha256sum \
  pkg/scaffold/templates/skills/unslop/SKILL.md \
  plugins/atv-skill-unslop/skills/unslop/SKILL.md \
  plugins/atv-pack-quality/skills/unslop/SKILL.md \
  plugins/atv-everything/skills/unslop/SKILL.md
```

All four hashes should match.

### Command contract smoke matrix

Manually inspect the skill logic against these cases:

| Command | Expected behavior |
|---------|-------------------|
| `/unslop` | full report, no edits |
| `/unslop frontend` | frontend report only |
| `/unslop architecture` | architecture report only |
| `/unslop fix` | safe hygiene + comments only |
| `/unslop fix all` | high-priority eligible fixes across all lanes |
| `/unslop fix --all` | same as `fix all` |
| `/unslop fix architecture` | high-priority eligible architecture only |
| `/unslop fix architecture medium` | medium-priority eligible architecture only |
| `/unslop fix frontend low` | low-priority eligible frontend only |
| `/unslop fix frontend architecture` | reject with choose-one-lane message |
| `/unslop fix all src/components` | high-priority eligible fixes across all lanes inside path scope |
| `/unslop fix architecture A1` | only architecture candidate A1 |

### Regression checks

- Verify `/lfg` still references `/unslop fix`.
- Verify `/slfg` read-only parallel report still uses `/unslop`.
- Verify `/slfg` sequential cleanup still uses `/unslop fix` unless intentionally changed.
- Verify README command examples do not imply `/unslop fix` mutates architecture/frontend.

---

## Failure Modes

| Failure mode | Impact | Mitigation |
|--------------|--------|------------|
| `all` means two different things | User surprises and unsafe cleanup | Reserve `all` for all lanes high-priority eligible fixes only. |
| Architecture auto-fix changes behavior | Broken app after cleanup | Require auto-fix eligibility, behavior lock, verification, and skip high-risk candidates. |
| Frontend auto-fix changes visual intent | UI gets worse while "cleaning" | Require visual verification or clear design-system convention. |
| Prompt shim points to missing path | `/unslop` prompt fails or confuses agent | Fix `.github/prompts/unslop.prompt.md`. |
| Skill copies drift | Installed plugin behaves differently than scaffold | Edit template first, mirror copies, verify hashes and tests. |
| LFG becomes too aggressive | Autonomous shipping mutates design/architecture unexpectedly | Keep LFG on `/unslop fix`, not `/unslop fix all`, for v2. |

---

## NOT in Scope

- Building a separate `architecture` skill. This plan keeps architecture inside `/unslop` as a lane.
- Importing Matt Pocock's skill verbatim. This plan adapts the framework into ATV's slash-command workflow.
- Adding a runtime parser implementation outside `SKILL.md`. The current skill is instruction-driven, so v2 updates the skill contract and examples.
- Changing all autonomous pipelines to run `/unslop fix all`. That is too aggressive for default shipping.
- Supporting lane-local "all priorities" in v2. Add `all-priorities` later if users need it.

---

## What Already Exists

- Existing `/unslop` skill already has code slop, comment rot, and design slop detector prompts.
- Existing `/unslop fix` already has safe auto-fix language for comments and code hygiene.
- Existing LFG and SLFG pipelines already call `/unslop` and `/unslop fix`.
- Existing scaffold parity tests likely catch broad skill-surface drift.
- Existing README already positions `/unslop` as part of the quality workflow.

---

## Open Decisions

### OD1. Should `/lfg` ever use `/unslop fix all`?

Recommendation: no for v2. Keep autonomous pipelines conservative. Users can manually run `/unslop fix all` after seeing the report.

### OD2. Should skipped architecture findings route to a future `/unslop architecture design A1` command?

Recommendation: not in v2. Use next-command copy in the report, but do not add a new command branch until we see demand.

### OD3. Should frontend auto-fix require browser access every time?

Recommendation: no. Require browser/screenshot verification only when the change affects visual layout, hierarchy, or styling beyond a mechanical accessibility/design-system fix.

---

## Worktree Parallelization Strategy

Sequential implementation is safer. Most units touch the same primary file, `unslop/SKILL.md`, across multiple mirrored copies. Parallel work would create avoidable merge conflicts.

Suggested order:

1. U1 through U4 in the scaffold template.
2. Mirror template to plugin copies.
3. U5 prompt shim.
4. U6 docs and pipeline references.
5. U7 tests and verification.

---

## Completion Criteria

- `/unslop` docs describe a full four-lane report.
- `/unslop fix` remains conservative.
- `/unslop fix all` and `/unslop fix --all` are documented aliases.
- Architecture lane uses deepening candidate language and eligibility gates.
- Frontend lane has explicit verification rules.
- Prompt shim no longer points to a missing path.
- All unslop copies match.
- `go test ./...` passes.
- README examples match the final command contract.
