---
name: kb-goal
description: Durable objective governor for KB workflows. Use when the user sets a goal, wants work to run for days across sessions, says continue until done, asks for vDone, or needs a long-lived objective forced through KB routing, planning, work, completion, and proof gates. Owns goal state and stop conditions; delegates execution to kb-start, kb-task, kb-epic, klfg, kb-work, and kb-complete.
argument-hint: "[goal objective, goal ledger path, or blank to resume active goal]"
---

# KB Goal

Own the durable objective, not the implementation lane.

`kb-goal` is the durable-objective lane for work that can span days. It keeps
the objective, terminal proof, blockers, next action, and restart state honest
while the normal KB lanes do the actual work.

Do not use chat confidence as a completion signal. A goal is complete only when
the routed KB lane reaches its proof gate and that proof satisfies the original
goal contract.

## Contract

- Run `kb-map lookup <goal>` before creating, resuming, or routing a goal.
- Store goal state in the active repo, not in global memory or chat history.
- Route each work unit through `kb-start` unless the ledger already names a
  valid next action such as `kb-work <manifest>` or `kb-complete <manifest>`.
- Preserve the smallest correct lane. Do not force every goal through `klfg`.
- Continue across sessions by updating the goal ledger and active handoff before
  stopping.
- Mark complete only after terminal proof matches the goal's done criteria.
- Mark blocked only with exact blocker, attempted route, and resume condition.

## Goal Ledger

Create or update:

```text
docs/context/goals/<goal-slug>.md
```

Also add a compact pointer in `todo.md` while the goal is active.

Use this shape:

```markdown
# <Goal Name>

Status: active|blocked|complete|parked
Created: YYYY-MM-DD
Last updated: YYYY-MM-DD

## Objective

One sentence.

## Done Criteria

- <observable condition>

## Terminal Proof

- <command, gate, artifact, or review condition required before completion>

## Current State

- Current artifact: <manifest/epic/handoff/path or none>
- Next allowed action: <exact KB command>
- Last proof: <command/artifact/status or none>

## Live Steering (optional)

Use this block only for recurring, scheduled, or trend-improvement goals where
future runs should be steered by measurements and durable feedback. Omit it for
ordinary one-shot goals.

- Set point: <desired invariant, threshold, or direction>
- Sensor: <command, query, test, or review signal that measures the gap>
- Controller: <how the next reviewable increment is selected>
- Actuator: <KB lane, coding agent, or workflow that applies the increment>
- Disturbances: <outside changes the loop must tolerate>
- Dampener: <optional check that prevents the measured issue getting worse>
- Scope gate: <paths or systems the loop may change/read>
- Batch size: <maximum targets per run>
- WIP bound: <maximum active manifests/PRs/work items for this loop>
- Steering memory: <goal-ledger section or docs/context/operations/steering/<slug>.md>

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|

## Notes
```

Keep the ledger compact. Move routine history into `todo-done.md` when the goal
closes.

## Routing

Pick the next smallest useful unit, then delegate:

| Goal State | Route |
|---|---|
| One bounded task can finish the goal | `kb-task` |
| Small known bug or contained fix | `kb-fix` |
| Broken behavior needs diagnosis | `kb-troubleshoot` |
| Clear feature needs slices | `kb-plan` -> `kb-work` -> `kb-complete` |
| Fuzzy objective or high path dependency | `kb-brainstorm` -> `kb-plan` -> `kb-work` -> `kb-complete` |
| Many streams, blockers, or manifests | `kb-epic`, then run each produced manifest |
| User wants one strict idea-to-done pipeline | `klfg` |
| Valid manifest already exists | `kb-work <manifest>` -> `kb-complete <manifest>` |
| Work is implemented and needs terminal gates | `kb-complete <manifest>` |
| Release or deploy is the remaining unit | `kb-ship` |

`klfg` is one strict pipeline run. `kb-goal` may run many pipeline runs.

### Goal Brainstorm Rule

Inside a goal, brainstorming should minimize human stops. The agent should pick
the best path from repo evidence, prior requirements, safe assumptions, and
research whenever that is enough to move forward.

Ask the user only for `ask-now` blockers: product choices, safety approvals,
credentials/access, irreversible tradeoffs, or ambiguity that would make the
plan wrong. Resolve `research-first` with research. Carry `safe-assumption`,
`defer-to-planning`, and `parked` items in the ledger with rationale instead of
turning them into questions.

### Live Steering Rule

Use live steering only when the goal benefits from repeated feedback-driven
runs. The goal ledger should name the set point, sensor, controller, actuator,
scope gate, batch size, WIP bound, and steering-memory path. Do not manufacture
separate sensor/controller/actuator steps when one repo tool or prompt naturally
fuses them; record the fusion instead.

The steering memory is durable guidance loaded into future runs after the next
increment is selected and before execution begins. It is for permanent scope
constraints, known false positives, reviewer preferences, or feedback that
should change future selections. Keep it concise and human-readable. Do not
store raw transcripts, one-off PR instructions, or single-run logs there.

Default flow control for scheduled or repeated loops is one active manifest or
PR per loop unless the ledger records a different WIP bound and proof strategy.
This prevents a loop from producing work faster than it can be reviewed.

## Loop

1. **Restore** - read `todo.md`, `docs/context/PROJECT.md`, and the goal ledger.
2. **Check staleness** - if the next artifact is older than 72 hours, run the
   normal stale-work refresh before execution.
3. **Choose next unit** - identify the smallest work unit that moves the goal.
4. **Delegate** - invoke the route from the ledger or route through `kb-start`.
5. **Verify unit** - require the delegated lane's gate evidence.
6. **Update ledger** - record artifact, status, proof, blocker, and next action.
7. **Decide**:
   - if done criteria and terminal proof are satisfied, mark `complete`;
   - if more units remain, continue or write a handoff and resume next session;
   - if blocked, record exact resume criteria and stop honestly.

Do not stop at weaker milestones:

- one work unit passed;
- a manifest says all slices are done but `kb-complete` has not run;
- tests passed before review/follow-up proof;
- `klfg` emitted DONE for one pipeline but the goal has remaining criteria;
- the model believes the objective is probably satisfied.

## Completion Rules

Complete only when all are true:

- the ledger `Done Criteria` are satisfied;
- the latest delegated route has terminal proof;
- every active manifest is `complete`, `reviewed`, `parked`, or explicitly
  blocked with resume criteria;
- unresolved P0/P1 findings are absent;
- final verification commands or artifacts are recorded;
- memory/handoff state points to the completed goal or no longer points to it.

If `kb-complete` creates follow-up work, keep the goal open and route that work
through the smallest valid KB lane.

## Blocked Rules

A goal is blocked only when further agent work would be fake progress.

Valid blockers include:

- missing credentials, MFA, paid access, hardware, or private data;
- product decision with multiple reasonable outcomes;
- unsafe/destructive action awaiting approval;
- external service outage or unavailable dependency;
- verification cannot run and no safe substitute exists;
- repeated gate failure with no new evidence path.

When blocked, write:

- exact blocker;
- what was attempted;
- current artifact;
- next allowed action after unblock;
- whether unrelated units can continue.

## Output

During work, report only:

- active goal;
- next route;
- current gate/proof;
- blocker or next action.

Final output:

```text
Goal: <name>
Status: complete|blocked|active
Route(s): <routes actually run>
Proof: <commands/artifacts/gates>
Next: <exact next action or none>
```
