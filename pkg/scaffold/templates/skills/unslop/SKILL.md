---
name: unslop
description: "Full-surface de-slop workflow: code hygiene, comments/docs, frontend/design, and architecture. Reports everything by default; fixes safe hygiene by default; `fix all` applies high-priority auto-fix-eligible fixes across all lanes."
argument-hint: "[fix] [all|hygiene|comments|frontend|architecture] [high|medium|low] [candidate IDs or file/directory paths]"
---

# /unslop - Full-Surface AI Slop Cleanup

`/unslop` is the broad X-ray for AI-generated slop. It reports every surface where AI-assisted work tends to leave cheapness behind: **code hygiene**, **comments/docs**, **frontend/design**, and **architecture**.

Default behavior is safe:

- `/unslop` reports across all four lanes and does not edit.
- `/unslop fix` edits only safe code hygiene and comments/docs issues.
- `/unslop fix all` and `/unslop fix --all` apply **high-priority, auto-fix-eligible fixes across all four lanes**.

Architecture and frontend fixes are allowed only when they pass explicit eligibility gates. Priority selects importance. Eligibility selects safety. Scope selects where edits are allowed.

## When to Use

Use `/unslop` when:

- a feature works but feels bloated, repetitive, generic, over-commented, or over-abstracted
- recent AI-assisted work left dead code, filler comments, stale docs, wrapper layers, weak seams, or generic UI
- a PR needs a quality pass before review
- you want a ranked cleanup report with priority, risk, effort, and next commands
- you want high-priority eligible cleanup via `/unslop fix all`

## When Not to Use

Do not use `/unslop` when:

- the main task is a new feature or behavior change
- desired behavior is unclear and cannot be protected by tests or a concrete verification plan
- the user wants a broad redesign rather than an incremental cleanup pass
- the request is only formatting or lint trivia
- the cleanup would require product decisions the user has not made

## Command Contract

The parser should be forgiving. This is a slash-command skill, not a strict Unix binary. Accept flag forms and natural token forms.

### Report Commands

| Command | Behavior |
|---------|----------|
| `/unslop` | Full read-only report across all lanes |
| `/unslop hygiene` or `/unslop --hygiene` | Code hygiene report only |
| `/unslop comments` or `/unslop --comments` | Comments/docs report only |
| `/unslop frontend` or `/unslop --frontend` | Frontend/design report only |
| `/unslop architecture` or `/unslop --architecture` | Architecture report only |
| `/unslop src/auth` | Full report scoped to a path |

### Fix Commands

| Command | Behavior |
|---------|----------|
| `/unslop fix` | Safe default: code hygiene + comments/docs only |
| `/unslop fix all` or `/unslop fix --all` | High-priority, auto-fix-eligible fixes across all four lanes |
| `/unslop fix hygiene` or `/unslop fix --hygiene` | High-priority eligible code hygiene fixes only |
| `/unslop fix comments` or `/unslop fix --comments` | High-priority eligible comments/docs fixes only |
| `/unslop fix frontend` or `/unslop fix --frontend` | High-priority eligible frontend/design fixes only |
| `/unslop fix architecture` or `/unslop fix --architecture` | High-priority eligible architecture fixes only |
| `/unslop fix architecture medium` | Medium-priority eligible architecture fixes only |
| `/unslop fix frontend low` | Low-priority eligible frontend/design fixes only |
| `/unslop fix architecture A1` | Candidate-scoped architecture fix |
| `/unslop fix all app/components` | High-priority eligible fixes across all lanes inside a path |
| `/unslop fix architecture --dry-run` | Preview selected fixes without editing |

### Aliases

| Canonical token | Accepted aliases |
|-----------------|------------------|
| `all` | `--all` |
| `hygiene` | `--hygiene`, `code-hygiene` |
| `comments` | `--comments`, `comment`, `docs`, `documentation` |
| `frontend` | `--frontend`, `front-end`, `design` |
| `architecture` | `--architecture`, `arch` |
| `high` | `--high` |
| `medium` | `--medium` |
| `low` | `--low` |
| `dry-run` | `--dry-run`, `preview` |

### Priority Selectors

**Priority selectors are exact**, not cumulative.

| Selector | Meaning |
|----------|---------|
| no priority token | high priority only |
| `high` | high priority only |
| `medium` | medium priority only |
| `low` | low priority only |

`all` does not mean all priorities. In v2, `all` means all lanes with high-priority eligible fixes.

### Candidate and Path Scope

Candidate IDs narrow the selected fixes:

- `H1`, `H2` for hygiene
- `C1`, `C2` for comments/docs
- `F1`, `F2` for frontend/design
- `A1`, `A2` for architecture

Examples:

```text
/unslop fix architecture A1
/unslop fix frontend F1,F4
/unslop fix comments C2 src/api
```

Paths narrow analysis and edits:

```text
/unslop src/auth
/unslop fix architecture src/auth
/unslop fix all app/components
```

Do not silently expand a path-scoped run into unrelated files except to verify references/usages. If a safe fix requires editing outside scope, skip it and report why.

### Conflict Rule

**Multiple lane selectors are rejected** unless the user chose `all`.

Reject:

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

## Lanes

### Lane 1: Code Hygiene

Find and optionally fix mechanical code slop:

- commented-out code blocks
- unused imports, variables, functions, or helpers
- duplicate tiny helpers
- needless defensive branches that cannot trigger
- pointless wrappers that add no behavior
- deep nesting that can be safely flattened with early returns
- copy-paste branches with identical behavior

Safe fixes are deletion-first. Prefer deleting code over rewriting code. Prefer inlining over a new abstraction. Do not add dependencies during `/unslop` unless the user explicitly asks.

### Lane 2: Comments/docs

Find and optionally fix comment and documentation slop:

- comments that restate the next line of code
- filler phrases such as "this function is responsible for"
- stale TODOs or FIXMEs that are already done
- inaccurate parameter or return docs
- redundant JSDoc/docstrings on trivial getters or obvious one-liners
- section divider comments that add noise but no navigation value
- docs that describe implementation details no caller should know

Safe fixes include deleting obvious restatement comments and stale filler. Do not delete a comment that explains non-obvious intent, edge cases, product constraints, security rationale, or external system behavior.

### Lane 3: Frontend/design

Find and optionally fix UI slop:

- generic AI color palettes, especially default blue/purple gradients without brand rationale
- uniform 3-column card grids with no hierarchy
- weak empty/loading/error/success states
- missing hover/focus/active states on interactive elements
- design-system bypasses and one-off class soup
- repetitive eyebrow/title/description stuffing
- emoji badges or decorative icons with no product voice
- uniform radius/shadow treatment on every surface
- overly perfect layouts where rhythm or emphasis would help
- components whose props expose implementation detail instead of product concepts

Frontend fixes require extra care. Auto-fix only when the change is eligible:

- the fix is mechanical accessibility or interaction-state cleanup, or
- an existing design-system convention clearly dictates the change, or
- screenshot/browser verification is possible and included in the report

If visual verification is not possible for a visual/layout change, skip the fix and report the exact command or manual review needed.

### Lane 4: Architecture

Add an **Architecture Slop Detector** pass. It should use Matt-style deepening language and report architecture findings as candidates, not vague refactor opinions.

Look for:

- shallow modules where the interface is almost as complex as the implementation
- pass-through wrappers that hide nothing
- fake seams or adapters with one implementation and no real substitutability
- caller knowledge leakage, where call sites know ordering, invariants, retries, parsing, or lifecycle details
- duplicated orchestration spread across multiple files
- modules that fail the deletion test: deleting the module removes complexity instead of moving useful complexity behind a deeper interface
- frontend component boundaries where props mirror implementation details instead of product concepts
- tests that mock internals instead of testing through a stable public interface
- interfaces that make easy things verbose and hard things impossible

Use these terms in findings where relevant:

- module
- interface
- implementation
- depth
- seam
- adapter
- leverage
- locality
- test surface

Architecture fixes are eligible only when they are scoped, testable, and low enough risk. High-risk architecture findings must remain report-only unless the user selects a specific candidate and the verification plan is explicit.

## Execution Flow

### Stage 1: Parse Arguments

Determine:

1. Mode: report or fix
2. Lane selector: all lanes, one lane, or safe default
3. Priority selector: high by default for lane/all fixes, exact `high`/`medium`/`low` when provided
4. Candidate IDs, if present
5. Path scope, if present
6. Dry-run/preview mode, if present
7. Conflicts, especially multiple lanes without `all`

### Stage 2: Determine Scope

If path arguments are provided, scope to those files/directories.

If no path is provided, use changed files since the base branch:

```bash
BASE=$(git merge-base HEAD origin/main 2>/dev/null || git merge-base HEAD origin/master 2>/dev/null || echo "HEAD~10")
echo "FILES:" && git diff --name-only "$BASE"
echo "DIFF:" && git diff -U5 "$BASE"
```

Skip generated/vendor output such as `dist/`, `build/`, `node_modules/`, `.next/`, `vendor/`, checked-in minified bundles, and generated lockfile chunks unless the user explicitly scoped to them.

### Stage 3: Run Analysis Passes

Run all selected lanes. For default `/unslop`, run every lane that applies to the scope.

If the harness supports safe read-only delegated analysis, the passes may run in parallel. If not, run them sequentially. The output contract is the same either way.

#### Pass 1: Hygiene Detector

Return:

```json
{
  "pass": "hygiene",
  "findings": [
    {
      "id": "H1",
      "file": "src/auth.ts",
      "line": 42,
      "finding": "Commented-out fallback branch is dead weight.",
      "priority": "high",
      "risk": "low",
      "effort": "small",
      "auto_fix_eligible": true,
      "suggested_fix": "Delete the commented block.",
      "verification": ["go test ./..."]
    }
  ]
}
```

#### Pass 2: Comments/docs Detector

Return:

```json
{
  "pass": "comments-docs",
  "findings": [
    {
      "id": "C1",
      "file": "src/auth.ts",
      "line": 10,
      "finding": "Comment restates the function name instead of explaining intent.",
      "priority": "low",
      "risk": "low",
      "effort": "small",
      "auto_fix_eligible": true,
      "suggested_fix": "Remove the comment.",
      "verification": ["go test ./..."]
    }
  ]
}
```

#### Pass 3: Frontend/design Detector

Return:

```json
{
  "pass": "frontend-design",
  "findings": [
    {
      "id": "F1",
      "file": "src/Hero.tsx",
      "line": 22,
      "finding": "Generic card grid gives every item equal weight and weakens hierarchy.",
      "priority": "high",
      "risk": "medium",
      "effort": "medium",
      "auto_fix_eligible": true,
      "suggested_fix": "Promote the primary card and add a focused empty/error state treatment using existing design tokens.",
      "verification": ["take before/after screenshots", "run affected frontend tests"]
    }
  ]
}
```

#### Pass 4: Architecture Slop Detector

Return Matt-style deepening candidates:

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
      "verification": ["targeted session tests", "existing auth tests"]
    }
  ]
}
```

### Stage 4: Merge, Score, and Deduplicate

1. Collect findings from all selected passes.
2. Deduplicate overlapping findings by file + line or shared candidate description.
3. Assign stable IDs if the detector did not.
4. Verify priority/risk/effort are present for every finding.
5. Verify every auto-fix-eligible finding has a concrete verification plan.
6. Sort by lane, then priority high to low, then risk low to high.

### Stage 5: Present Report

Use this report format:

```markdown
# /unslop Report

Scope: [N] files changed since [base] or [explicit path]
Mode: report
Fix default: safe hygiene + comments/docs only
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
| C1 | Comments/docs | src/foo.ts:12 | Comment restates function name | Low | Low | Small | Yes | /unslop fix comments C1 |
| F1 | Frontend/design | app/card.tsx:70 | Generic card grid lacks hierarchy | High | Medium | Medium | Yes | /unslop fix frontend F1 |
| A1 | Architecture | src/session.ts | Session rules leak into 4 callers | High | Medium | Medium | Yes | /unslop fix architecture A1 |
```

For architecture findings, include a short candidate detail block after the table:

```markdown
### Architecture Candidate A1

Problem: [problem]
Proposed deeper interface: [proposed_fix]
Depth: [depth]
Locality: [locality]
Leverage: [leverage]
Test surface: [test_surface]
Verification: [commands/checks]
```

If no findings are present:

```markdown
/unslop Report: Clean. No slop detected in [N] files.
```

### Stage 6: Fix Selection

For `/unslop fix`, select only auto-fix-eligible code hygiene and comments/docs findings.

For `/unslop fix all` or `/unslop fix --all`, select high-priority, auto-fix-eligible findings across all four lanes.

For lane-specific fixes, select only that lane and the exact priority requested. No priority token means high priority.

Candidate IDs narrow the selected findings after lane and priority selection.

Never apply a finding unless:

- it matches the requested lane/all selection
- it matches the requested priority selector
- it is auto-fix eligible
- it stays within path scope
- it has a verification plan

### Stage 7: Cleanup Plan Before Edits

Before editing, print a short cleanup plan:

```markdown
Cleanup plan:
1. [Lane] [ID] [specific change]
2. [Lane] [ID] [specific change]

Behavior lock / verification before editing:
- [command or manual verification]
```

If tests cannot run before editing, record why and name the post-edit verification plan. Do not pretend unverified architecture/frontend work is safe.

### Stage 8: Apply Fixes

Work lane by lane. Prefer deletion before rewriting.

Recommended order:

1. Comments/docs
2. Hygiene
3. Frontend/design
4. Architecture

Run targeted verification after each lane when practical. If verification fails, fix the failure or back out the risky cleanup.

### Stage 9: Final Fix Report

Always report:

```markdown
## /unslop Fix Report

Changed files:
- [file]

Simplifications:
- [ID] [what changed]

Verification run:
- [command] -> [result]

Skipped findings:
- [ID] [reason, such as high risk or outside scope]

Remaining risks:
- [risk or "none known"]
```

## Auto-Fix Eligibility Guide

| Lane | Eligible examples | Not eligible by default |
|------|-------------------|-------------------------|
| Hygiene | dead comments, unused imports, trivially unused helpers | broad refactors, behavior-changing rewrites |
| Comments/docs | obvious restatements, stale filler, done TODOs | intent comments, security rationale, external API notes |
| Frontend/design | missing focus state following existing convention, mechanical token alignment | visual redesign without screenshot verification |
| Architecture | shallow wrapper deletion, fake seam removal, small deeper interface with tests | high-risk boundary moves, large module migrations |

## Severity and Scoring

Use **priority** for importance, **risk** for regression chance, and **effort** for implementation size.

Priority:

- High: fixing this materially improves quality or removes real maintenance drag
- Medium: noticeable quality improvement, but not blocking
- Low: small cleanup or polish

Risk:

- Low: mechanical or deletion-only with clear verification
- Medium: modest behavior or visual surface, protected by tests or screenshots
- High: could change behavior, product semantics, layout intent, or public interfaces

Effort:

- Small: one file or a tiny local edit
- Medium: several related files, targeted tests
- Large: cross-module migration or broad redesign

## Quality Gates

Before presenting findings:

1. Every finding must be actionable.
2. Line numbers must be accurate when line numbers are cited.
3. Do not flag generated code unless explicitly scoped.
4. Verify unused/dead code before flagging it as safe.
5. Respect project conventions.
6. Do not mark architecture/frontend fixes auto-fix eligible without a verification plan.
7. Report skipped fixes instead of forcing them.

## Relationship to Other ATV Skills

- Pair with `/ce-review`: correctness and safety review.
- Pair with `/frontend-design` or design review workflows when frontend findings require real visual judgment.
- Keep `/lfg` and `/slfg` conservative by default. They should continue to use `/unslop fix` unless the user explicitly asks for `/unslop fix all`.

## Notes

- `/unslop` is read-only by default.
- `/unslop fix` is conservative by default.
- `/unslop fix all` is intentionally stronger, but still only applies high-priority auto-fix-eligible findings.
- Architecture reports should use Matt-style deepening candidates, but this skill does not import another skill verbatim.
- Frontend visual changes need evidence, not just class-name vibes.