---
name: kb-epic
description: Large-initiative coordinator for KB workflows. Use for app migrations, framework rewrites, major architecture changes, multi-subsystem initiatives, multiple brainstorms/manifests, long backlogs, or queued execution across related workstreams.
argument-hint: "[initiative description or epic path]"
---

# KB Epic

Coordinate large work without turning it into one huge plan. This is the lane for "migrate this app to a different framework", "rewrite the architecture", or "run a whole program of related work."

## Goal

Create an epic map that points to multiple brainstorms, plans, manifests, handoffs, research notes, and subsystem docs. Keep each execution unit small enough for `kb-plan` and `kb-work`.

The default `kb-epic` outcome is planning complete: every workstream has a
current manifest, is explicitly parked, or is blocked by a named human decision.
Do not stop after creating brainstorms unless human input is required before
planning can continue.

Record workflow shape during planning. Use `multi-stream-epic` when independent
streams, blockers, brainstorms, and manifests exist; use `pipeline-change` when
skills, scripts, evals, docs, and proof gates must change together.

Assume the user invoked `kb-epic` while available for critical decisions and
wants the planning blockers surfaced early so they can leave while execution
runs later.

## When To Use

Use `kb-epic` when:

- The request is bigger than one brainstorm or one manifest.
- Example: language/runtime migration, app-shell rewrite, auth overhaul, full interaction architecture replacement.
- One plan would be too large to execute or review.
- Work spans multiple subsystems.
- Architecture direction affects many later slices.
- Several brainstorms should feed one release.
- The user wants a queued multi-workstream execution plan.

## Epic Location

Create:

```text
docs/context/epics/<epic-name>.md
```

If `docs/context/epics/` does not exist, create it and link it from `docs/context/PROJECT.md`.

## Epic Template

```markdown
# <Epic Name>

Status: draft|active|parked|complete
Created: YYYY-MM-DD
Last refreshed: YYYY-MM-DD

## Intent

## Success Criteria

## Architecture Decisions

## Research

## Workstreams

| Workstream | Brainstorm | Manifest | Status | Notes |
|---|---|---|---|---|

## Dependency Map

## Execution Queue

## Human Checkpoints

## Parked / Blocked

## Completion Criteria
```

## Scheduling

Default to serial execution when workstreams share files, schemas, prompts, auth, routing, generated artifacts, or architecture decisions.

Use read-only swarms for research/review. Use coding swarms only when file ownership is declared and non-overlapping.

## Planning Completion Contract

`kb-epic` is the coordinator that gets an initiative to a fully planned state.
Its job is not done when brainstorms exist. It is done only when the epic map
shows one of these outcomes for every workstream:

- `planned` or `queued` with a manifest path;
- `blocked` or `human-required` with the exact question that blocks planning;
- `parked` with a rationale;
- `done` with proof or completion notes.

After creating or refreshing brainstorm docs, immediately extract their
`Resolve Before Planning` items. Convert those items into the smallest possible
human checkpoint list, ask only the questions that block planning, then continue
to `kb-plan` for every workstream that is unblocked.

Use the shared Question Gate classes from `kb-gate` for every workstream:
`ask-now`, `research-first`, `safe-assumption`, `defer-to-planning`, and
`parked`. Only `ask-now` questions go to the human by default. Resolve
`research-first` with source/external research before asking, unless research
cannot answer it. Carry `safe-assumption`, `defer-to-planning`, and `parked`
items into the epic map with rationale and forbidden claims.

When a question can be handled as an explicit assumption without changing
safety, architecture, acceptance criteria, or user intent, record the assumption
and keep planning. Do not stop the epic just because a brainstorm contains nice
to-have questions.

HITL-first rule:

- Surface planning-blocking questions before generating manifests.
- Ask the smallest number of questions needed to unblock planning, but collect
  `ask-now` questions across all brainstorm-needed workstreams before asking.
- Ask brainstorm questions in one ordered batch, grouped by workstream, instead
  of bouncing between question, answer, plan, question, answer, plan.
- Do not batch unrelated questions inside one prompt unless the user explicitly
  asked for a single consolidated checkpoint. Preserve question order and keep
  each answer mappable back to one workstream.
- After the last planning-blocking question is answered, continue until all
  unblocked workstreams have manifests or explicit parked/blocked status.
- Complete or update all brainstorm docs with the answers before creating
  manifests.
- Then generate plans/slices for every unblocked workstream: both those that
  came from completed brainstorms and those that were clear enough to skip
  brainstorming.
- Once planning is complete, ask:

```text
Planning is complete. Do you want me to continue until all planned work is
completed and tested? If yes, I will run kb-work, then kb-finalize, across the
runnable manifests from this epic.
```

If the user says yes, execute runnable manifests serially unless the epic has
declared safe parallel ownership. After each manifest, run `kb-finalize` or an
equivalent milestone completion gate before moving on. Do not execute blocked,
parked, or human-required manifests.

## Workstream Routing

An epic may have one umbrella brainstorm, but execution should not depend on one
giant brainstorm. Use the smallest useful artifact per workstream:

| Workstream State | Route | Output |
|---|---|---|
| Goal, user value, constraints, or success criteria are unclear | `kb-brainstorm` for that workstream | `docs/brainstorms/<date>-<workstream>-requirements.md` |
| Behavior is clear enough to slice, but execution shape is not | `kb-plan` for that workstream | `docs/plans/<date>-000-kb-<workstream>-manifest.md` plus slice plans |
| A valid manifest already exists and is current | `kb-work <manifest>` | executed slices |
| Workstream is blocked by architecture/research decision | `kb-research` or a small decision note first | research/decision linked from epic |

Default phase order:

- Brainstorm all workstreams whose goals, boundaries, or acceptance criteria are
  unclear before creating plans for them.
- Do not create one large plan that embeds unresolved brainstorm questions.
- After all required brainstorm docs exist, extract their blocking questions
  together, ask them in workstream/dependency order, then update every affected
  brainstorm before planning.
- Batch planning only after brainstorm questions are answered or parked. If
  some workstreams are blocked, continue planning the unblocked workstreams.

Brainstorm granularity:

- Use one umbrella brainstorm when the epic has one coherent decision surface.
  Its slice candidates may seed the Workstreams table.
- Use one brainstorm per workstream when the workstream has its own product
  tradeoffs, unclear acceptance criteria, or human checkpoint.
- Multiple brainstorms are right when the workstreams could make different
  product or architecture decisions without invalidating each other.
- Do not run a brainstorm for a workstream whose behavior is already clear; send
  it straight to `kb-plan`.
- Before sending a brainstormed workstream to `kb-plan`, resolve or explicitly
  park any `Resolve Before Planning` items. If those items affect scope,
  acceptance criteria, architecture direction, data contracts, or human risk,
  planning is premature.

Manifest convention:

- A manifest is the `kb-plan` output file under `docs/plans/` whose frontmatter
  has `type: kb-manifest` and whose filename normally looks like
  `<date>-000-kb-<topic>-manifest.md`.
- Record that manifest path in the epic Workstreams table and in `todo.md`.
- Queue only runnable or deliberately parked manifests. Do not queue a raw
  brainstorm as runnable work.

Todo target:

- Default target is the active project root's `todo.md` loaded by `kb-map`.
- If the environment has a separate session/private todo store, do not use it as
  the durable queue unless repo instructions explicitly say so. The epic must be
  resumable from repo-local memory.

`todo.md` queue convention:

```markdown
| <Workstream name> | ⬜ pending | P0|P1|P2 | `docs/plans/<manifest>.md` |
```

Use `🔧 in_progress` only while an agent is actively executing that manifest.
Use `🔒 blocked` or `🛑 human-required` when the next action cannot run without a
specific decision. Keep completed manifests out of active rows; move summaries
to `todo-done.md`.

## Flow

1. Read `todo.md`, `docs/context/PROJECT.md`, and relevant subsystem docs.
2. Create or refresh the epic map.
3. Split the initiative into workstreams small enough for normal KB flow.
4. Fill the Workstreams table as you route:
   - `Brainstorm`: blank, `skipped-clear`, or the brainstorm path.
   - `Manifest`: blank until `kb-plan` creates it, then the manifest path.
   - `Status`: `draft`, `planned`, `queued`, `running`, `blocked`, `done`, or
     `parked`.
   - `Notes`: blocker, dependency, human checkpoint, or proof pointer.
5. Run or refresh required brainstorms.
6. Extract all Question Gate items from every brainstorm into `Human
   Checkpoints`, preserving workstream/dependency order.
7. Resolve `research-first` items with research where possible, then ask only
   remaining `ask-now` human questions. If a question is not blocking, record a
   `safe-assumption`, `defer-to-planning`, or `parked` entry and continue.
8. Apply the answers to complete or update all brainstorm docs.
9. Route every unblocked workstream to `kb-plan`, including workstreams that
   skipped brainstorming.
10. Record each created manifest in the epic Workstreams table and queue it in
   `todo.md` using this repo's queue convention.
11. Continue until every workstream is planned, queued, blocked, parked, or done.
12. Ask whether to continue into execution for all runnable manifests.
13. If yes, use `kb-work` for runnable manifests.
14. Use `kb-finalize` after each manifest or at epic milestones.
15. Use `kb-map refresh` after durable architecture changes.
16. Use `kb-complete` once at the intended epic delivery boundary.

Refresh cold epic work older than 72 hours before execution.
