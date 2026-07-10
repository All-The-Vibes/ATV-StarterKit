---
name: kb-finish
description: "Take a plan, KB manifest, or active planned work all the way to done-done and checked in: plan if needed, execute kb-work, run kb-complete, then kb-ship to commit, push, and open/update a PR. Use when the user says 'finish this plan', 'done-done and checked in', 'take this manifest to PR', or wants plan-to-PR automation."
argument-hint: "[manifest path, plan/requirements path, or feature description]"
---

# KB Finish

Own the explicit plan-to-PR loop:

```text
plan/source -> kb-plan -> kb-work -> kb-complete -> kb-ship -> pushed PR
```

This is the shipping orchestrator. `kb-complete` remains the non-shipping
quality/learning phase because `kb-work` calls it automatically.

## Input Resolution

1. Run `kb-map lookup <request>` and resolve the active project root.
2. If input is a KB manifest, use it.
3. If input is a slice plan or handoff, follow its `kb_id`, `Manifest:`, or
   plan pointer to an existing manifest before creating anything. If no
   manifest exists, invoke `kb-plan <input>` with execution intent.
4. If input is empty, use the single active manifest from `todo.md`. If multiple
   manifests are plausible, ask one blocking selection question.
5. Refresh artifacts older than 72 hours before execution.

## State-Driven Loop

Re-read the manifest after every delegated lane; never rely on chat state.

| Manifest state | Action |
|---|---|
| missing/invalid | `kb-plan <source>` |
| active with runnable/pending/in-progress slices | `kb-work <manifest>` |
| completed with missing/pending/blocked `complete-to-ship` | `kb-complete <manifest>` once; if still blocked, stop with its exact resume action |
| reviewed with `complete-to-ship: passed|quarantined` | audit final completion scope, then `kb-ship <manifest>` |
| reviewed with missing/stale/blocked gate | `kb-complete <manifest>` once; stop if no state change |
| manifest or slice blocked/human-required/parked | persist exact blocker and stop |
| malformed state or no progress after a delegated route | `kb-gate`/`kb-plan` repair once, then stop if unchanged |

`kb-work` may invoke `kb-complete` itself. After it returns, re-read the
manifest and skip phases already proven by their gates.

Before `kb-ship`, reconcile completion-generated manifest, todo, solution,
instinct, memory, and cleanup files into a final audited shipping scope.

## Safety and Completion

- Do not skip plan, work, review, proof, learning, or memory gates.
- Do not ship parked or human-required work as complete.
- Do not stage unrelated changes.
- Do not merge. Checked in means committed, pushed, and represented by a PR.
- If `kb-ship` reports a push or PR blocker, remain blocked until resolved.
- Do not report success at "tests passed", "reviewed", "committed", or
  "ready to push".

Terminal success requires:

- manifest `status: reviewed`;
- `complete-to-ship` passed or is explicitly quarantined with risks preserved;
- `kb-ship` release checks passed;
- branch committed and pushed;
- PR URL recorded, or honest `nothing-to-ship`.

## Output

```text
KB finish: shipped|nothing-to-ship|blocked
Manifest: <path>
Review gate: <complete-to-ship status>
Branch: <branch>
Commit: <sha or none>
PR: <url or none>
Next: none|<exact unblock action>
```
