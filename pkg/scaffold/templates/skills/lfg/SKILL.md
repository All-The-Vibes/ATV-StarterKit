---
name: lfg
description: Full autonomous engineering workflow
argument-hint: "[feature description]"
disable-model-invocation: true
---

CRITICAL: You MUST execute every step below IN ORDER. Do NOT skip any required step. Do NOT jump ahead to coding or implementation. The plan phase (step 2) MUST be completed and verified BEFORE any work begins. Violating this order produces bad output.

## Run State (resumability + artifact passing)

This workflow is **resumable**. A tiny helper tracks which phases are `done` and where each phase's output lives, so re-invoking `/lfg` continues from the first unfinished phase instead of restarting.

- **Helper:** `node .github/hooks/scripts/lfg-state.js` — commands `init`, `bind-plan`, `done <phase> --run-id <id> [--artifact <repo-relative-path>]`, `status --run-id <id>`, `run-id-from-plan --plan <path>`. It writes `.atv/runs/<run-id>/` (gitignored, local-only).
- **On start (resume check):**
  1. If a recent plan for this feature already exists in `docs/plans/`, derive `RUN_ID` from it: `node .github/hooks/scripts/lfg-state.js run-id-from-plan --plan <plan-path>`. Otherwise create a provisional id: run `node .github/hooks/scripts/lfg-state.js init --skill lfg --feature "$ARGUMENTS" --repo <repo> --branch <branch>` and read `run_id`.
  2. Run `node .github/hooks/scripts/lfg-state.js status --run-id <RUN_ID>` and **skip every phase whose sentinel is already `done`**; resume at the first not-done phase.
- **After each phase passes its gate:** record it with `node .github/hooks/scripts/lfg-state.js done <phase> --run-id <RUN_ID> [--artifact <path>]` (pass the artifact path the phase produced; omit when it produced none).
- **Pass `run:<RUN_ID>`** to every sub-skill so artifacts co-locate and downstream phases read paths, not full content.
- **Re-entry safety:** local state is only valid in this worktree; if `.atv/runs/` is absent, infer progress from the plan, branch, and any open PR. Non-`done` phases are re-entered from the start, so `ce-work` MUST be called with `mode:orchestrated` to reconcile existing work instead of duplicating commits/PRs.

1. **Optional:** If the `ralph-loop` skill is available, run `/ralph-loop-ralph-loop "finish all slash commands" --completion-promise "DONE"`. If not available or it fails, skip and continue to step 2 immediately.

2. `/ce-plan $ARGUMENTS run:<RUN_ID>`

   GATE: STOP. Verify that the `ce-plan` workflow produced a plan file in `docs/plans/`. If no plan file was created, run `/ce-plan $ARGUMENTS run:<RUN_ID>` again. Do NOT proceed to step 3 until a written plan exists. **Record the plan file path**, then bind it and mark the phase done: `node .github/hooks/scripts/lfg-state.js bind-plan --run-id <RUN_ID> --plan <plan-path>` and `node .github/hooks/scripts/lfg-state.js done ce-plan --run-id <RUN_ID> --artifact <plan-path>`.

3. `/ce-work mode:orchestrated plan:<plan-path-from-step-2> run:<RUN_ID>`

   GATE: STOP. Verify that implementation work was performed - files were created or modified beyond the plan. Do NOT proceed to step 4 if no code changes were made. Then `node .github/hooks/scripts/lfg-state.js done ce-work --run-id <RUN_ID>`.

4. `/ce-review mode:autofix plan:<plan-path-from-step-2> run:<RUN_ID>`

   Pass the plan file path from step 2 so ce-review can verify requirements completeness. Then `node .github/hooks/scripts/lfg-state.js done ce-review --run-id <RUN_ID> --artifact <review-artifact-path>`.

5. `/unslop fix`

   Strip AI slop after review fixes land — removes commented-out code, filler comments, stale TODOs. Then `node .github/hooks/scripts/lfg-state.js done unslop --run-id <RUN_ID>`.

6. `/observe` on the areas of code that were changed — analyze patterns in the modified files to capture what was done and how.

7. `/learn` — extract reusable patterns from this session into project instincts.

8. `/compound-engineering-todo-resolve` — then `node .github/hooks/scripts/lfg-state.js done todo-resolve --run-id <RUN_ID>`

9. `/compound-engineering-test-browser` — then `node .github/hooks/scripts/lfg-state.js done test-browser --run-id <RUN_ID> --artifact <report-path>`

10. `/compound-engineering-feature-video` — then `node .github/hooks/scripts/lfg-state.js done feature-video --run-id <RUN_ID>`

11. Output `<promise>DONE</promise>` when video is in PR

Start with step 2 now (or step 1 if ralph-loop is available). Remember: plan FIRST, then work. Never skip the plan.
