---
name: kb-goal
description: Durable objective governor for KB workflows. Use when the user sets a goal, wants work to run for days across sessions, says continue until done, asks for vDone, or needs a long-lived objective forced through KB routing, planning, work, finalization, delivery, and proof gates. Owns goal state and stop conditions; delegates execution to kb-start, kb-task, kb-epic, kb-work, kb-finalize, and kb-complete.
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

- Before reading `todo.md`, `.kb/runs/`, a goal ledger, or chat/session history,
  resolve and record the current working directory and `git rev-parse
  --show-toplevel`. If the user or launcher names a repo/workspace, verify that
  the resolved root matches that identity by path, repo name, or remote.
- A named repo/workspace is an identity constraint, not conversational context.
  On mismatch, do not resume any active goal. Switch to the named checkout when
  it is unambiguous; otherwise ask only for its path.
- Resume goal state only when `todo.md`, the goal ledger, and `.kb/runs/<goal>/`
  are all under the verified repo root. Never carry an active goal from a prior
  session, sibling checkout, global memory, or chat history into another repo.
- Run `kb-map lookup <goal>` before creating, resuming, or routing a goal.
- Store goal state in the active repo, not in global memory or chat history.
- Route each work unit through `kb-start` unless the ledger already names a
  valid next action such as `kb-work <manifest>`, `kb-finalize <manifest>`, or
  `kb-complete <manifest>`.
- Preserve the smallest correct lane. Do not force every goal through `klfg`.
- When routed work is first about to run, initialize or reuse exactly one
  ephemeral run catalog under `.kb/runs/<goal-slug>/` from live host evidence.
  Do not ask a setup questionnaire for Small/Medium/Large/Planner models.
- Merge only what the active orchestration surface can actually select plus any
  user-local or project-allowed extra routes at work time. Never assume one
  host exposes another host's catalog.
- Require an objective `done_check` before starting a durable goal, or record an
  explicit human-approved exception with the reason no objective check can exist
  yet. Do not treat a vague Done Criteria list as terminal proof.
- Continue across sessions by updating the goal ledger and active handoff before
  stopping.
- Mark complete only after terminal proof matches the goal's done criteria.
- When the objective can be expressed as a check JSON, terminal proof should
  include `go run ./cmd/kbcheck accept --check <check.json> --trace
  .kb/trace.jsonl`.
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

## Done Check

- Type: command_exit|artifact_exists|gate|human_exception
- Check: <exact command, artifact path, gate id, or exception summary>
- Expected result: <exit code, path condition, gate status, or approval source>
- Why sufficient: <which done criterion this proves>

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

## Run State

For autonomous, recurring, or multi-session goals, create ephemeral run state at:

```text
.kb/runs/<goal-slug>/
```

This is not a replacement for `todo.md`, manifests, handoffs, or the goal
ledger. It is git-ignored control-loop state for the active run only.

Required files:

- `goal.md` - pointer to the durable goal ledger and current objective.
- `done-check.json` - optional `kbcheck sense/accept` check spec when the done
  check can be expressed as JSON.
- `catalog.json` - the redacted live run catalog for this goal/run only.
- `catalog-fingerprint.txt` - the last accepted host/config fingerprint used to
  decide whether the run catalog can be reused.
- `backlog.json` - small queue of candidate work units with route, priority,
  blockers, and source artifact.
- `progress.md` - compact current state, last accepted proof, and next allowed
  action.
- `route-history.jsonl` - one JSON object per route decision.

Each `route-history.jsonl` row should include:

```json
{"ts":"<ISO-8601>","route":"kb-work","confidence":0.82,"state_changed":true,"progress_key":"slice-003-done"}
```

Before choosing the next route for an existing run, validate the history:

```powershell
go run ./cmd/kbcheck run-state --history .kb/runs/<goal-slug>/route-history.jsonl
```

If the guard flags `route-oscillation`, `low-confidence-no-progress`, or
`no-progress-loop`, stop the loop and re-plan or ask the smallest human question
instead of bouncing between lanes.

The run catalog stays redacted, ephemeral, and project-local. It records only what this host
and this run can select, plus any project-allowed aliases the user already
configured. It is never a trust source; credentials, approvals, and trust stay
under the OS user's private KB state. Refresh it only when the
surface/provider/configuration or generated agent fingerprint changes.

## Routing

Pick the next smallest useful unit, then delegate:

| Goal State | Route |
|---|---|
| One bounded task can finish the goal | `kb-task` |
| Small known bug or contained fix | `kb-fix` |
| Broken behavior needs diagnosis | `kb-troubleshoot` |
| Clear feature needs slices and configured delivery | `kb-complete` |
| Fuzzy objective or high path dependency | `kb-complete` (routes through brainstorm) |
| Many streams, blockers, or manifests | `kb-epic`, then run each produced manifest |
| Legacy strict pipeline request | `kb-complete` |
| Valid manifest already exists | `kb-complete <manifest>` |
| Work is implemented and needs internal quality gates | `kb-finalize <manifest>` |
| Delivery is the remaining unit | `kb-complete <manifest>` |

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

1. **Bind repo identity** - resolve the git root, verify any user/launcher repo
   name, and refuse cross-repo goal restoration on mismatch.
2. **Restore** - read `todo.md`, `docs/context/PROJECT.md`, and the goal ledger
   only from that verified root.
3. **Check staleness** - if the next artifact is older than 72 hours, run the
   normal stale-work refresh before execution.
4. **Check run state** - if `.kb/runs/<goal>/route-history.jsonl` exists, run
   the `kbcheck run-state` guard before choosing another route.
5. **Choose next unit** - identify the smallest work unit that moves the goal.
6. **Delegate** - invoke the route from the ledger or route through `kb-start`.
7. **Verify unit** - require the delegated lane's gate evidence. For manifests
   with `objective_contract: true`, require `go run ./cmd/kbcheck
   manifest-contract --manifest <manifest-path>` to pass before accepting a
   unit as complete.
8. **Update ledger/run state** - record artifact, status, proof, blocker, next
   action, and a route-history row with confidence and progress signal.
9. **Decide**:
   - if done criteria and terminal proof are satisfied, mark `complete`;
   - if more units remain, continue or write a handoff and resume next session;
   - if blocked, record exact resume criteria and stop honestly.

Do not stop at weaker milestones:

- one work unit passed;
- a manifest says all slices are done but `kb-finalize` has not run;
- tests passed before review/follow-up proof;
- `klfg` emitted DONE for one pipeline but the goal has remaining criteria;
- the model believes the objective is probably satisfied.

## Completion Rules

Complete only when all are true:

- the ledger `Done Criteria` are satisfied;
- the ledger `Done Check` passed, or the recorded human-approved exception is
  still valid and scoped;
- the latest delegated route has terminal proof;
- every active manifest is `complete`, `reviewed`, `parked`, or explicitly
  blocked with resume criteria;
- unresolved P0/P1 findings are absent;
- final verification commands or artifacts are recorded;
- memory/handoff state points to the completed goal or no longer points to it.

If `kb-finalize` creates follow-up work, keep the goal open and route that work
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
