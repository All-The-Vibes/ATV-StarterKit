---
name: kb-plan
description: "Break a brainstorm or feature into vertical-slice task plans with dependency DAG, verification strategy, and HITL flags. Default planning workflow for end-to-end vertical slices instead of horizontal phases. Use when the user says 'kb plan', 'plan', 'create a plan', 'plan this', 'slice this', 'break into vertical slices', or wants independently-grabbable tasks."
argument-hint: "[brainstorm path, feature description, or PRD]"
---

# KB Plan - Vertical Slice Decomposition

<!-- Inspired by mattpocock/skills to-issues - credit: github.com/mattpocock/skills -->

Break work into independently executable **vertical slices** (tracer bullets). Each slice cuts through all relevant layers end-to-end. Avoid horizontal phases.

## Quick Start

1. Read the brainstorm, PRD, or feature description.
2. Draft thin end-to-end slices with dependencies and verification modes.
3. Review the breakdown yourself against the source material; ask the user only for blocking decisions.
4. Write one KB manifest plus one plan file per slice.
5. Create or update the manifest `gate_ledger`; `plan-to-work` must be
   `passed` before `kb-work` may execute.
6. After writing the manifest, continue to `kb-work <manifest-path>` only when
   execution was requested or an orchestrator called this plan. Otherwise ask
   once and print the exact next command.
7. Stage or commit only the generated files when the user explicitly asked for a commit.

## Interaction Method

Default to non-interactive planning when the source material is clear. Use the platform's blocking question tool only when an answer changes behavior, scope, acceptance criteria, risk, or verification.

When assumptions are safe and reversible, record them in the manifest instead of stopping. Ask one concise question only for material uncertainty.

Planning cannot launder brainstorm ambiguity. If the source contains unresolved
`ask-now` or `research-first` items, a non-empty `Resolve Before Planning`
section, or unlabeled material assumptions that affect scope, acceptance
criteria, architecture direction, safety, or verification, stop and route back
to `kb-brainstorm`/`kb-gate`. Only `safe-assumption`,
`defer-to-planning`, and `parked` items may cross into planning, and each must
be recorded in the manifest.

Phase boundary: `kb-plan` produces a manifest and slice plans. It does not
automatically invoke `kb-work` unless the user explicitly asked for execution or
an orchestrator such as `klfg`, `kb-epic`, or `kb-goal` called it.

Execution intent includes phrases such as "go straight to work", "just build it", "don't ask many questions", "continue until done", "run it", or a handoff from `kb-task`, `kb-brainstorm`, or `klfg` that says to continue. In those cases, write the manifest and slice plans first, then immediately invoke `kb-work <manifest-path>`. Never skip manifest creation.

Without execution intent, ask once after the manifest is valid:

```text
Plan is ready: <manifest-path>
Continue with `kb-work <manifest-path>` now?
```

If the user says yes, invoke `kb-work <manifest-path>`. If the user says no, or
the host cannot invoke the next skill, stop and print:

```text
Next command: `kb-work <manifest-path>`
```

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Check `todo.md` and `docs/brainstorms/` for the active or most recent brainstorm. If exactly one likely source exists, use it and record the assumption. If multiple plausible sources exist, ask which one to use. If none exist, ask: "What feature or work should I slice?"

**If input is a brainstorm path:** Read it thoroughly. This is the source of truth for what to build. Carry forward all decisions, scope boundaries, and requirements.

**If input is a handoff path:** Do source discovery before planning:

1. Read the handoff.
2. Check the handoff for explicit `Brainstorm:`, `Requirements:`, `Source:`, `Manifest:`, or `Plan:` pointers.
3. Check `todo.md` for a source pointer tied to that handoff or feature.
4. Look for matching existing source artifacts under project-root paths only:
   - `docs/brainstorms/*<topic>*`
   - `docs/requirements/*<topic>*` if that folder exists
   - `docs/plans/*<topic>*`
5. If a matching brainstorm or requirements doc exists, read it and use it as the planning source of truth. The handoff becomes restart context, not the primary source.
6. If a matching manifest already exists, ask whether to resume with `kb-work` instead of creating a duplicate plan.
7. If no source exists and the handoff is concrete enough, plan from the handoff and record `source: handoff`.
8. If no source exists and the handoff leaves material product or architecture decisions open, stop and route to `kb-brainstorm`.

**If input is a feature description:** Proceed directly to decomposition.

## Core Rules

### Vertical Slices Only

Each slice must deliver a narrow but complete path through every relevant layer: schema, service, API, UI, tests, docs, or ops as applicable. A completed slice is demoable or verifiable on its own.

```text
WRONG (horizontal phases):
  Task 1: Create database schema
  Task 2: Build service layer
  Task 3: Add API routes
  Task 4: Build frontend

RIGHT (vertical slices):
  Task 1: Award points on lesson completion + show on dashboard
  Task 2: Track streaks (builds on task 1)
  Task 3: Add level progression display
```

### Enabling Slices Are Acceptable

Some work is legitimately enabling infrastructure: migrations, auth plumbing, shared config, repo setup. Allow enabling slices only when:

- They unlock a named downstream slice
- They are the smallest viable prerequisite
- The slice names its immediate consumer(s)

### Live-Steering Slices

For recurring, scheduled, or trend-improvement work routed from `kb-goal`,
include the control-loop fields in the manifest or the first slice plan:

- set point: the invariant, threshold, or direction being driven;
- sensor: the command, query, test, or review signal that measures the gap;
- controller: how the next small reviewable increment is selected;
- actuator: the KB lane, coding agent, or workflow that applies the change;
- disturbances: outside changes the loop must tolerate;
- dampener: optional regression gate that keeps the measured problem from
  getting worse while the loop improves it;
- scope gate, batch size, WIP bound, and steering-memory path.

Do not force this framing onto one-shot feature work. Do not invent separate
sensor, controller, and actuator artifacts when the repo's real toolchain fuses
them; record the fused component and the selection policy. HumanLayer-style CI,
Bun, CodeLayer, or GitHub Actions runners are examples, not KB defaults.

### Every Slice Has a Verification Strategy

| Mode | When | Gate |
|------|------|------|
| `tdd` | Behavior changes, business logic | Define protected oracle first when practical -> prove RED -> implement -> unchanged oracle passes |
| `integration` | Wiring, cross-boundary, API contracts | Integration test proves path works |
| `functional` | User-visible workflow, UI/API/CLI journey, escaped bug | Workflow-level check proves the user path |
| `verification-only` | Config, scaffolding, ops | Builds pass, no regression |
| `hitl` | UX taste, design judgment | Human confirms acceptable |

Also record `test_level` and `functional_risk` for each slice. `kb-functional-test` owns this classification:

- `test_level`: `none`, `unit`, `integration`, `functional-api`, `functional-cli`, `functional-browser`, or `full`
- `functional_risk`: `none`, `narrow`, `broad`, or `full`

Use `unit` only when a unit test can genuinely prove the changed behavior. Use functional levels when a unit test could pass while the user-visible, API, CLI, browser, persistence, auth/session, streaming, or integration path is broken.

## Process

### 1. Understand the Source Material

Read the brainstorm/PRD/description. Extract:

- What behaviors need to exist
- What the user-visible outcomes are
- What constraints/dependencies exist
- What's explicitly out of scope
- Question Gate state: unresolved `ask-now`/`research-first` blockers, safe
  assumptions, deferred planning questions, and parked forbidden claims.

If the source has unresolved `ask-now` or `research-first` items, stop before
decomposition. Write or update the `brainstorm-to-plan` gate as `blocked` or
`needs-human` and set `allowed_next_action` to the smallest repair action, such
as `kb-brainstorm <requirements-path>`.

### 1.5. Research (Parallel)

Run lightweight research to ground slice design in reality:

- `repo-research-analyst`: similar features, routes, components, tests, commands, conventions.
- `learnings-researcher`: relevant `docs/solutions/`, prior fixes, and known failures.
- Local memory check: `docs/context/PROJECT.md`, `docs/context/landmines.md` when present, relevant `docs/context/architecture/*`, and `docs/context/research/*`.

If `docs/context/landmines.md` exists, read only `Active Landmines`. Any relevant active landmine must become an explicit planning constraint, risk, guardrail, or verification requirement in the slice plan. Do not copy resolved/archive landmines into new plans unless the current work reopens the same failure mode.

Run independent agents/reads/searches in parallel when the platform supports it. If named agents are unavailable, do the same work with native search/read tools.

**Research decision:** Based on findings, decide if external research is needed.

- **High-risk topics** (security, payments, external APIs, data privacy) → always research externally
- **Strong local patterns exist** → skip external research
- **Unfamiliar territory** → research externally

**If external research is warranted**, use `kb-research` and write a reusable research note before finalizing slices.

Optional specialist checks before finalizing slices:

- `spec-flow-analyzer` when the feature has multi-step user flows or unclear edge cases.
- `security-sentinel` or `security-reviewer` when auth, permissions, secrets, public endpoints, payments, PII, external callbacks, or user input are involved.
- `adversarial-reviewer` when the plan is architecture-shaping, high-risk, or large enough that assumption/cascade failures are likely.
- `architecture-strategist` when subsystem boundaries, framework migration, or long-term architecture direction are being set.

Carry research findings forward into slice plans — each slice should reference relevant patterns, gotchas, and file paths discovered here.

### 2. Draft Vertical Slices

Break the work into thin end-to-end slices. For each slice, determine:

- **Title** - short descriptive name
- **What it delivers** - end-to-end behavior description
- **Verification mode** - tdd / integration / verification-only / hitl. For `tdd`, record the oracle path/command before implementation whenever practical.
- **Test level** - none / unit / integration / functional-api / functional-cli / functional-browser / full
- **Functional risk** - none / narrow / broad / full
- **Blocked by** - which other slices must complete first, or none
- **HITL flag** - does this need human judgment? Most should be `false` if the brainstorm was thorough.
- **Expected files** - best current forecast of files this slice may create or modify, with operation type. Used by `kb-work` as an orientation and review-scope seed, not as a literal allowlist.

Each entry in `expected_files` should specify:
  - `path` — the file path
  - `op` — `create`, `edit`, or `delete`
  - `scope` — one-line description of what specifically changes (for `edit` operations)

This helps agents start surgically instead of rediscovering the whole repo. It cannot perfectly predict implementation reality; `kb-work` records discovered files in the scope ledger when current code requires touching files not forecast here.

### 3. Validate the Breakdown

Check the proposed breakdown against:

- Granularity: each slice is independently executable and reviewable.
- Dependencies: blockers are necessary, not accidental.
- Verification: each slice has agent-runnable tests/checks where possible.
- Functional coverage: user-visible or cross-boundary slices include a narrow functional check unless explicitly justified.
- Test-level classification: each slice says whether unit, integration, API/CLI/browser functional, or full-suite proof is required.
- HITL: human flags are limited to credentials, external systems, subjective approval, or true decisions.
- Expected files: each slice declares likely touched files and scope, with enough specificity to guide the first edit. Do not pretend the list is exhaustive when current code may reveal adjacent files.

Ask the user only when a material decision remains. Otherwise proceed and record assumptions.

Run `kb-gate` before writing final plans when validation surfaces P0/P1/P2/P3 issues. P0/P1 block work, but the agent should rectify safe/actionable blockers before asking the user. For P2/P3, ask whether to rectify all fixable issues before moving on.

Before handing off to `kb-work`, write a `plan-to-work` gate in the manifest.
Load `kb-gate/references/gate-ledger.md` if needed. The gate must include proof
for: manifest path, every slice plan path, dependency DAG validation, acceptance
criteria, `expected_files`, verification mode, `test_level`, `functional_risk`,
HITL classification, and any protected oracle policy. If any proof is missing,
set `status: blocked` and do not invoke `kb-work`.

### 4. Generate Plan Files

Create a manifest and individual slice plans.

#### Manifest: `docs/plans/YYYY-MM-DD-000-kb-<name>-manifest.md`

```yaml
---
type: kb-manifest
kb_id: kb-YYYY-MM-DD-<name>
brainstorm_path: docs/brainstorms/<source-file>.md
created: YYYY-MM-DD
status: active
workflow_shape: "<direct-chat|single-skill-edit|skill-bundle-change|pipeline-change|multi-stream-epic>"
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-brainstorm
    status: passed
    required_evidence:
      - "<requirements path exists>"
      - "Question Gate classification exists"
      - "Resolve Before Planning is empty"
      - "no unresolved ask-now or research-first items remain"
      - "safe assumptions, deferred planning questions, and parked items are recorded"
    proof:
      - docs/brainstorms/<source-file>.md
    blockers: []
    passed_at: "<timestamp>"
    allowed_next_action: "kb-plan <requirements-path>"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "<manifest path exists>"
      - "<all slice plan paths exist>"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk"
    proof:
      - docs/plans/YYYY-MM-DD-000-kb-<name>-manifest.md
      - docs/plans/YYYY-MM-DD-001-<type>-<name>-plan.md
    blockers: []
    passed_at: "<timestamp>"
    allowed_next_action: "kb-work <manifest-path>"
slices:
  - id: slice-001
    title: "<title>"
    path: docs/plans/YYYY-MM-DD-001-<type>-<name>-plan.md
    blockers: []
    verification: tdd
    test_level: unit
    functional_risk: none
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles:
      - path: "tests/<behavior>.test.<ext>"
        role: "behavior oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan update"
  - id: slice-002
    title: "<title>"
    path: docs/plans/YYYY-MM-DD-002-<type>-<name>-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-browser
    functional_risk: narrow
    hitl: false
    status: pending
    notes: ""
---

# KB: <Feature Name>

## Origin
Brainstorm: `<brainstorm_path>`

## Workflow Shape

`<workflow_shape>` - why this shape fits.

## Slice Overview
| # | Slice | Blocked By | Verification | HITL | Status |
|---|-------|------------|--------------|------|--------|
| 1 | <title> | - | tdd | no | pending |
| 2 | <title> | slice-001 | tdd | no | pending |
| 3 | <title> | - | integration | no | pending |
```

#### Individual Slice Plans: `docs/plans/YYYY-MM-DD-NNN-<type>-<name>-plan.md`

Each slice plan uses standard ATV plan format with additional frontmatter:

```yaml
---
kb_id: kb-YYYY-MM-DD-<name>
slice_id: slice-NNN
title: "<title>"
blockers: []
verification: tdd
test_level: unit
functional_risk: none
hitl: false
expected_files:
  - path: ""
    op: edit
    scope: "what specifically changes"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---
```

The plan body should include:

- What to build, expressed as end-to-end behavior
- Acceptance criteria
- Expected files (must match `expected_files` in frontmatter as the initial forecast; actual touched files may expand during `kb-work` when justified by the acceptance criteria and recorded in the scope ledger)
- Test scenarios specific enough for TDD or integration verification
- Protected oracle candidates when expected behavior is known before implementation: tests, fixtures, scorers, snapshots, or contract files that should be written or selected first, proven RED when practical, and protected from mutation with SHA before implementation continues
- Test inputs needed to run those scenarios without asking the user to manually test later
- Scope boundary: what this slice does not include
- Dependencies and why they are needed
- HITL question if `hitl: true`

If verification needs realistic input values, include them in frontmatter:

```yaml
test_inputs:
  - name: "<input name>"
    source: user|fixture|env|generated
    required_for: "<acceptance criterion or QA step>"
    value: "<literal value, fixture path, env var name, or TODO-human>"
```

Only mark `hitl: true` when the human step is truly required. Do not use HITL for checks the agent can run with provided inputs.

Use `protected_oracles` when a slice has a known behavior target before
implementation. Each entry should name the oracle file, its role, and the update
policy. `kb-work` fills or verifies the SHA after RED/protection. If the correct
oracle cannot be known until implementation reveals the interface, leave
`protected_oracles: []` and explain the verification strategy in the plan body.

### 5. Update Todo and Handoffs

After generating plan files, update `todo.md` — the human-visible live execution board.
Create or update a compact handoff file under `docs/handoffs/active/` only when a future session needs a restart packet.

**If `todo.md` doesn't exist**, create it with this template:

```markdown
# Todo

## Rules

**Conventions:** these match the KB skill spec. Keep them inline here; do not split into `todo_rules.md`, `todo-rules.md`, or any separate rules file.

**This file is the single source of truth for active work** — not chat history, not session SQL, and not stale manifests. Any agent should be able to claim a row from here cold.

**Status markers** (applied to individual rows):

| Marker | Meaning |
|--------|---------|
| ⬜ pending | Ready when blockers clear |
| 🔧 in_progress | Agent claimed and actively working |
| ✅ done | Complete and verified — move summary to `todo-done.md` promptly |
| 🔒 blocked | Cannot proceed — explain in `## Blocked` with `Depends on:` |
| ⊘ skipped | Intentionally skipped with reason |
| 🛑 human-required | Needs human action (HITL) — also surface under `## Human Required` |

**Section icons** (section headers, not row markers):

- 💡 Feature Ideas — not yet brainstormed; a human promotes to active
- 📋 Queued Improvements — approved but not yet planned
- 🧊 Parked / Cold Storage — intentionally out of bounds today; never auto-runs, human-promote only
- 🛑 Human Required — items only a person can complete (logins, approvals, decisions)
- 📝 Work Log — short dated entries for cross-session visibility

**Task metadata** lines under a row when relevant: `Task ID:`, `Ready: yes|no`, `Depends on:`, `Discovered from:`, `Validation:`.

**Promotion rules:**
- Newly discovered work goes to 🧊 Parked / Cold Storage first. Never auto-execute from there.
- Items stalled because another agent, dependency, tool failure, or missing input must finish first go to 🔒 Blocked, not Parked.
- Human-required work must not be silently folded into generic blocked notes.
- Detailed handoffs live under `docs/handoffs/`; link them here instead of pasting content.
- Refresh cold or parked work older than 72 hours before execution.
- Keep this file current and small. When all active todos are done, check the handoff queue.

## Objective

## Current Focus

## Current Truth

## Active Work

## Handoff Queue

| Handoff | Status | Route | Created | Stale Check | Link |
|---|---|---|---|---|---|

## Human Required

## Parked / Cold Storage

## Blocked

## Work Log
```

**If `todo-done.md` doesn't exist**, create it with this template:

```markdown
# Completed Work

> Archive of completed items from `todo.md`. Most recent at top.

## YYYY-MM-DD
- <feature or slice group> — <compact outcome, important proof, commit/link if available>
```

**Add an active work section** for the new KB workflow:

```markdown
### <Feature Name> (kb-YYYY-MM-DD-name)

Source: `docs/brainstorms/<file>.md`
Manifest: `docs/plans/<manifest>.md`

| # | Slice | Blocked By | Verification | Status |
|---|-------|------------|--------------|--------|
| 1 | <title> | - | tdd | ⬜ pending |
| 2 | <title> | slice-001 | tdd | ⬜ pending |

Done criteria: All N slices done or skipped with reason. Archive a compact summary to `todo-done.md`, then remove this feature section and routine work-log entries from `todo.md`.
```

**If a restart packet is needed**, create `docs/handoffs/active/YYYY-MM-DD-<feature>.md`:

```markdown
# <Feature Handoff>

Created: YYYY-MM-DD
Last refreshed: YYYY-MM-DD
Status: active
Suggested route: kb-work

## Intent

## Current State

## Next Agent Action

## Human Required

## Pointers

- Project map: docs/context/PROJECT.md
- Manifest: docs/plans/<manifest>.md
- Todo: todo.md

## Staleness Check

Refresh before execution if older than 72 hours.

## Completion Criteria
```

Use `Suggested route: kb-work` only when the handoff links the generated KB manifest. If a handoff is created before slice planning exists, set `Suggested route: kb-plan` and state that execution must wait for a manifest.

**Board status markers** (superset of manifest statuses):

| Marker | Meaning |
|--------|---------|
| ⬜ pending | Ready when blockers clear |
| 🔧 in_progress | Agent claimed and actively working |
| ✅ done | Complete and verified |
| 🔒 blocked | Cannot proceed — reason noted |
| ⊘ skipped | Intentionally skipped with reason |
| 🛑 human-required | Needs human action (HITL) |

**Standing sections** (add once, keep across features):

- **💡 Feature Ideas** — not yet brainstormed, human promotes to active
- **📋 Queued Improvements** — approved but not yet planned
- **🧊 Parked / Cold Storage** — intentionally out of bounds today; only a human promotes back to active
- **🛑 Human Required** — items only a person can complete (logins, approvals, decisions)
- **📝 Work Log** — short dated entries for cross-session visibility

Omit empty sections. These conventions are defined inline in the top `## Rules` section of `todo.md`; do not create or depend on `todo_rules.md` or `todo-rules.md`.

### 6. Validate Output

- Confirm every `blockers` entry references an existing slice ID.
- Confirm no dependency cycles exist.
- Confirm every slice has a verification mode and acceptance criteria.
- Confirm the manifest has a `plan-to-work` gate with `status: passed` or `status: blocked`; never leave it absent or pending.
- Confirm every generated plan path is listed in the manifest.
- Confirm the manifest body table matches the YAML frontmatter.

### 6. Optional Commit

Commit only when the user explicitly asked for it. Stage only the generated manifest and slice plan files, never the whole `docs/plans/` directory:

```bash
git add docs/plans/YYYY-MM-DD-000-kb-<name>-manifest.md docs/plans/YYYY-MM-DD-001-<type>-<name>-plan.md todo.md docs/handoffs/active/YYYY-MM-DD-<feature>.md
git commit -m "kb-plan: decompose <feature> into N vertical slices"
```

## Success Criteria

- The manifest is a valid DAG with no missing blockers or cycles.
- Each slice is independently grabbable and has a clear verification gate.
- Enabling slices name their immediate downstream consumers.
- Generated paths are precise enough for `kb-work` to resume without rediscovery.
- No unrelated existing plans are staged or changed.

## Integration with Other Skills

- **Input from:** `kb-brainstorm` or a clear feature description.
- **Deepening:** Use `kb-research` only for individual slices with material unresolved uncertainty.
- **Execution:** `kb-work` runs all slices in order when invoked, or can pick up one slice at a time
- **Verification:** Each `tdd` slice carries protected-oracle proof in the manifest; load the standalone `tdd` skill only for explicit test-first coaching.
- **Protected oracles:** Known behavior targets can be frozen before implementation so tests, fixtures, scorers, snapshots, or contracts cannot be rewritten silently
