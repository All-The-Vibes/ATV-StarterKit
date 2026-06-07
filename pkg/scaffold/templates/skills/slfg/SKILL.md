---
name: slfg
description: Full autonomous engineering workflow using swarm mode for parallel execution
argument-hint: "[feature description]"
disable-model-invocation: true
---

Swarm-enabled LFG. Run these steps in order, parallelizing where indicated. Do not stop between steps — complete every step through to the end.

## Run State (resumability + artifact passing)

This workflow is **resumable**. A tiny helper tracks which phases are `done` and where each phase's output lives, so re-invoking `/slfg` continues from the first unfinished phase instead of restarting.

- **Helper:** `node .github/hooks/scripts/lfg-state.js` — commands `init`, `bind-plan`, `done <phase> --run-id <id> [--artifact <repo-relative-path>]`, `status --run-id <id>`, `run-id-from-plan --plan <path>`. It writes `.atv/runs/<run-id>/` (gitignored, local-only).
- **On start:** derive `RUN_ID` from an existing `docs/plans/` plan via `run-id-from-plan`, else `init` a provisional id. Run `status --run-id <RUN_ID>` and **skip phases already `done`**.
- **Pass `run:<RUN_ID>`** to every sub-skill so artifacts co-locate and downstream phases read paths, not full content.
- **PARENT-ONLY WRITES (concurrency safety):** during the Parallel Phase, swarm subagents MUST NOT write run state. Each subagent only returns its artifact path; the **parent** orchestrator records each `done <phase> --run-id <RUN_ID> --artifact <path>` after the parallel agents join. This avoids races on the state directory.
- **Re-entry safety:** `ce-work` MUST run with `mode:orchestrated` so resume reconciles existing work instead of duplicating commits/PRs. If `.atv/runs/` is absent, infer progress from the plan, branch, and any open PR.

## Sequential Phase

1. **Optional:** If the `ralph-loop` skill is available, run `/ralph-loop-ralph-loop "finish all slash commands" --completion-promise "DONE"`. If not available or it fails, skip and continue to step 2 immediately.
2. `/ce-plan $ARGUMENTS run:<RUN_ID>` — **Record the plan file path** from `docs/plans/` for steps 4 and 6. Then (parent) `lfg-state.js bind-plan --run-id <RUN_ID> --plan <plan-path>` and `done ce-plan --run-id <RUN_ID> --artifact <plan-path>`.
3. `/ce-work mode:orchestrated plan:<plan-path-from-step-2> run:<RUN_ID>` — **Use swarm mode**: Make a Task list and launch an army of agent swarm subagents to build the plan. Then (parent) `done ce-work --run-id <RUN_ID>`.

## Parallel Phase

After work completes, launch steps 4, 5, and 6 as **parallel swarm agents** (all read-only, safe to run concurrently). Subagents return artifact paths only — the parent records state after they join:

4. `/ce-review mode:report-only plan:<plan-path-from-step-2> run:<RUN_ID>` — spawn as background Task agent
5. `/compound-engineering-test-browser run:<RUN_ID>` — spawn as background Task agent
6. `/unslop run:<RUN_ID>` — spawn as background Task agent (read-only de-slop report)

Wait for all three to complete before continuing. Then (parent) `node .github/hooks/scripts/lfg-state.js done test-browser --run-id <RUN_ID> --artifact <report-path>`. 

## Autofix Phase

7. `/ce-review mode:autofix plan:<plan-path-from-step-2> run:<RUN_ID>` — run sequentially after the parallel phase so it can safely mutate the checkout, apply `safe_auto` fixes, and emit residual todos for step 8. Then (parent) `done ce-review --run-id <RUN_ID> --artifact <review-artifact-path>`.
8. `/unslop fix` — run sequentially after ce-review autofix to strip AI slop (commented-out code, filler comments, stale TODOs). Then `done unslop --run-id <RUN_ID>`.

## Learning Phase

9. `/observe` on the areas of code that were changed — analyze patterns in the modified files to capture what was done and how.
10. `/learn` — extract reusable patterns from this session into project instincts.

## Finalize Phase

11. `/compound-engineering-todo-resolve` — resolve findings, compound on learnings, clean up completed todos. Then `done todo-resolve --run-id <RUN_ID>`.
12. `/compound-engineering-feature-video` — record the final walkthrough and add to PR. Then `done feature-video --run-id <RUN_ID>`.
13. Output `<promise>DONE</promise>` when video is in PR

Start with step 1 now.
