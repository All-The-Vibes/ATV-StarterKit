---
name: kb-start
description: Default KB start/router. Use when the user says "kb", gives an idea or ambiguous work request, starts a fresh session, asks what to do next, or wants the workflow to choose the right lane without making the user pick ceremony. Delegates project-memory setup and lookup to kb-map before choosing a lane.
argument-hint: "[user request or blank for session startup]"
---

# KB Start

Pick the right KB lane for the user's idea/request. The user should be able to ask normally; do not make them choose ceremony.

`kb-start` is not the memory bootstrapper. `kb-map` owns project-memory setup, lookup, and refresh.

## Map First

On every fresh session or ambiguous work request:

1. Invoke `kb-map lookup <user request>`.
2. Let `kb-map` decide whether lookup, refresh, or bootstrap is required.
3. After `kb-map` returns project context, classify the user request and route it.
4. If `kb-map` reports stale work or missing memory, honor that before executing work.

If `kb-map` cannot identify a valid active project root, ask the user to change into the project directory or provide the project path before routing. Drive roots such as `E:\`, home directories, and global skill/config folders are not valid project roots unless the user explicitly chose them. Do not route from global handoffs or home-directory memory when the user is working in a repo.

## Run-State Guard

When a durable goal or autonomous loop names `.kb/runs/<goal-slug>/`, validate
its route history before choosing another lane:

```powershell
go run ./cmd/kbcheck run-state --history .kb/runs/<goal-slug>/route-history.jsonl
```

If the guard reports `route-oscillation`, `low-confidence-no-progress`, or
`no-progress-loop`, do not keep routing. Stop and choose the smallest repair:
refresh stale context, re-plan the work unit, or ask the one human question that
would break the loop.

After choosing a lane for an active run-state goal, append one route-history row
with at least `ts`, `route`, `confidence`, and either `state_changed` or
`progress_key`. Use confidence as a routing-confidence signal, not a completion
claim.

## Session Hygiene Check

Run this check only when `kb-start` begins a request. Do not interrupt an active brainstorm, plan, work slice, review, or test loop just to suggest a restart.

Goal: decide whether the user is better served by staying in the current session, compacting, or creating/updating a handoff and restarting fresh.

Use exact context telemetry when the platform exposes it. In GitHub Copilot CLI, `/context` shows context usage. If the agent cannot read telemetry directly, do not guess a percentage; use the evidence-based fallback below.

Context thresholds when exact telemetry is available:

| Context Used | Default |
|---|---|
| `<60%` | Stay in session. Do not mention restart unless the user asks. |
| `60-80%` | Mention restart only if the user is switching tasks or lanes. |
| `80-90%` | Recommend handoff/restart before starting substantial new work. |
| `>90%` | Strongly recommend handoff/restart, or compact if the user must continue here. |

Evidence-based fallback when telemetry is unavailable:

- Suggest restart when the session is long, tool output has been heavy, compaction likely happened, the user is switching tasks, or the agent is relying on chat history instead of local files.
- Do not suggest restart merely because the session feels long.

Before recommending restart, estimate rebuild cost:

| Rebuild Cost | Signals | Recommendation |
|---|---|---|
| Low | current handoff exists; `todo.md`, `PROJECT.md`, and manifest/plan pointers are current | Recommend fresh session when context pressure exists. |
| Medium | project memory exists but handoff needs updating | Offer to update/create a handoff, then restart. |
| High | important nuance is only in chat; mid-debug observations matter; no current handoff/map | Stay, or compact first, then write durable memory before restarting. |

Restart rule:

> Do not recommend a fresh session merely because the session is long. Recommend it only when durable local memory can replace the live chat at lower total context cost or lower drift risk.

When restart is advisable, ask once:

```text
This looks like a good reset point. I can create/update a handoff so the next session starts cleanly, or we can keep going here.

1. Create/update handoff and restart
2. Compact current context and continue
3. Continue in this session
4. Other / let me explain
```

If the user chooses handoff/restart, create or update the active handoff under `docs/handoffs/active/`, ensure `todo.md` points to it, and include the exact `kb-start <handoff/task>` prompt for the next session. Do not run the next workflow in the old session unless the user asks.

## Read Order

Read only what `kb-map` points to, then only what is needed to route:

1. `kb-map` result.
2. Relevant active handoff files or manifest paths named by `kb-map`.
3. Specific subsystem, research, brainstorm, or plan files pointed to by `kb-map`.

## Ranked Routing Decision List

Choose the first matching route. This list is the only routing taxonomy; do not
reconcile it with a second shape or complexity table.

| Rank | Request Signal | Route | Proof/Gate |
|---|---|---|---|
| 1 | Project memory missing, partial, stale, or root invalid | `kb-map` | `kb-map` decides lookup/refresh/bootstrap |
| 2 | User explicitly says `kb-goal`, sets a durable goal, wants work to run for days, asks for vDone, or needs cross-session objective tracking | `kb-goal` | goal ledger plus delegated KB gates |
| 3 | User explicitly says `kb-task`, asks for first-principles execution, or wants one bounded task carried until verified/blocked | `kb-task` | task runner owns verification |
| 4 | Direct explanation, tradeoff discussion, or pushback with no file changes requested | answer directly; use `kb-first-principles` behavior when challenged | no work gate |
| 5 | User wants a feature/plan/manifest taken from its current state through configured local, PR, or direct delivery | `kb-complete` | plan/work/finalize gates plus delivery policy |
| 6 | Existing valid manifest should be executed without check-in intent | `kb-work` | manifest plus slice verification |
| 7 | All runnable slices are done and only internal review/learning/cleanup is needed | `kb-finalize` | `kb-check`, `kb-review`, learning gates |
| 8 | Already reviewed work needs configured delivery | `kb-complete` | delivery policy plus release/ship/land gates |
| 9 | Broken behavior needs logs, browser checks, test iteration, or self-correction | `kb-troubleshoot` | reproduce and regression proof |
| 10 | Architecture/module-depth exploration before implementation | `kb-architecture-deepening` | source/docs evidence and tradeoff table |
| 11 | External docs, prior art, framework behavior, or market/product research materially affects the answer | `kb-research` | cited source notes |
| 12 | Multiple independent streams, many blockers, deletion policy, migration scale, or several brainstorms/plans needed | `kb-epic` | brainstorms and plans complete before work |
| 13 | Scripts/evals/proof harness plus skills/docs must change together, or cross-runtime propagation is part of the change | `kb-epic` or coded pipeline manifest | eval/proof/sync gate |
| 14 | Clear feature/refactor needs slices, or user wants execution but no valid manifest exists | `kb-plan` | vertical-slice manifest |
| 15 | Skill-bundle change with sync/docs/eval/standard gate implications | `kb-plan` | `kb-check -All` and sync report |
| 16 | Fuzzy idea, product direction, or high path dependency | `kb-brainstorm` | answered questions before planning |
| 17 | Small known bug, typo, narrow cleanup, or one skill/doc edit with no sync/eval/proof-harness implications | `kb-fix` or bounded direct edit | targeted proof plus `kb-check` when relevant |
| 18 | Memory/docs/responses are too verbose | `kb-compact` | preserve commands, paths, dates, blockers |
| 19 | Legacy `klfg` or `kb-finish` invocation | compatibility alias to `kb-complete` | full state-aware pipeline |

Pipeline-worthy changes have at least one of these signals: multiple owning
surfaces, cross-runtime behavior, scorer/fixture/baseline changes,
propagation/sync rules, several independent workstreams, or deletion/loaded
surface measurement.

If none of those signals are present, keep the route small. Do not build a
pipeline just because the request mentions a skill.

## Current Truth

`todo.md` may hold short-lived operational truth: current focus, active manifest, parked slices, blockers, and handoff pointers.

Durable app truth belongs in `docs/context/architecture/*`. If an operational fact becomes durable architecture knowledge, ask `kb-map refresh` to update it.

## Stale Work Rule

Before running a handoff, brainstorm, plan, or parked todo older than 72 hours, perform a refresh check:

- What changed since it was created or last refreshed?
- Did touched files/subsystems change?
- Does the route still make sense?
- Does the artifact need updating before execution?

Do not run stale work blindly.

## Handoff Routing Rule

Handoffs are restart packets, not automatically executable plans.

Before resuming any `docs/handoffs/active/*` file, classify it:

| Handoff Shape | Route |
|---|---|
| Contains or links a `docs/plans/*-kb-*-manifest.md` with slice plans | `kb-work <manifest>` |
| Contains vertical slices with `expected_files`, verification, blockers, and status | `kb-plan` to normalize into a manifest, then `kb-work` |
| Contains phases, workstreams, bullets, open decisions, or broad next steps | `kb-plan` |
| Contains unclear product/technical intent | `kb-brainstorm` |
| Contains multiple child initiatives or a migration/rewrite scale objective | `kb-epic` |

Do not route a phase-shaped handoff directly to `kb-work`. `kb-work` requires a manifest and per-slice plans with `expected_files`.

Route by complexity, not by file count or guessed duration. The useful signals are uncertainty, blast radius, coupling, reversibility, verification burden, and user/product path dependency.

Record `workflow_shape` in generated manifests when planning follows the ranked
list. Use the closest rank label, such as `single-skill-edit`,
`skill-bundle-change`, `pipeline-change`, or `multi-stream-epic`.

When in doubt, prefer the lane that prevents rework. Do not pick a 20-minute shortcut when the decision creates path dependency.

Execution intent is not permission to skip planning. If the user wants fewer questions or wants the agent to continue directly to implementation, reduce Q&A and carry an `execute_after_plan` intent forward, but still create or reuse a KB manifest before `kb-work`.

## Ceremony Rule

Minimize visible ceremony:

- Do not ask "which KB skill should I use?"
- State the chosen lane in one line, then proceed when safe.
- Ask only when the choice changes risk, cost, or user intent.
- If the wrong lane becomes obvious, switch lanes and record why.

## Token Budget

Every token must pay rent. Keep startup output short and load only pointed-to files.

Route to `kb-compact` when:

- `todo.md`, handoffs, research notes, or architecture docs carry repeated history instead of current signal.
- A skill draft repeats rules already in `AGENTS.md` or `.github/copilot-instructions.md`.
- The user asks for fewer words, terser output, or token reduction.

Do not compact away exact commands, paths, dates, IDs, acceptance criteria, blockers, HITL reasons, or safety warnings.

## Output

Report briefly:

- Map status.
- Route chosen.
- Why that route fits.
- Any stale-work refresh needed.
- Exact next action, including the skill command and artifact path when known.

If the route is obvious and safe, proceed into the chosen skill workflow. Also
name the exact command before or as you invoke it, so users on hosts that do not
auto-chain skills can continue manually. If the host cannot invoke the target
skill, stop with: `Next command: <skill> <artifact-or-request>`.
