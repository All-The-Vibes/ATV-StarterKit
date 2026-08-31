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
- **Helper resolution:** Prefer `.github/hooks/scripts/lfg-state.js` in the target repository. If it is absent, use `scripts/lfg-state.js` beside this loaded `SKILL.md` (the plugin-packaged fallback). In every helper command below, replace `.github/hooks/scripts/lfg-state.js` with the resolved path.
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

## Quality and Release-Readiness Phase

5. Enter the `quality-release-readiness` phase after review:

   a. Run `/unslop fix` to strip AI slop after review fixes land. Then run `node .github/hooks/scripts/lfg-state.js done unslop --run-id <RUN_ID>`.

   b. Run `node .github/hooks/scripts/lfg-state.js get-decision quality-release-readiness --run-id <RUN_ID>`. If a decision already exists, reuse it and do not prompt again. Keep its `choice` and `artifact` as the current decision.

   c. If no decision exists, propose the adjacent release-readiness option to the user:

   > `/unslop fix` cleaned implementation slop. `/solution-debranding-plan` can now prepare this solution for reuse across brands, white-labeling, ownership transfer, demo use, or public release. Run solution debranding now?

   If no decision exists, record exactly one choice:

   - `declined` or `not-applicable`: `node .github/hooks/scripts/lfg-state.js decision quality-release-readiness --run-id <RUN_ID> --choice <choice>`.
   - `accepted`: immediately record acceptance with `node .github/hooks/scripts/lfg-state.js decision quality-release-readiness --run-id <RUN_ID> --choice accepted`. Then invoke `/solution-debranding-plan` with the user's source brand and scope. Update the accepted decision with the returned plan path: `node .github/hooks/scripts/lfg-state.js decision quality-release-readiness --run-id <RUN_ID> --choice accepted --artifact <debranding-plan-path>`.

   d. Resolve both new and resumed decisions:

   - For `declined` or `not-applicable`, continue without debranding.
   - For `accepted`, require the stored debranding plan artifact. If acceptance was recorded before planning completed and the artifact is absent, resume `/solution-debranding-plan` without prompting again, then update the decision with its returned plan path. Run `node .github/hooks/scripts/lfg-state.js is-done solution-debranding-apply --run-id <RUN_ID>`. If it is incomplete, run `/solution-debranding-apply <debranding-plan-path>` only after required approvals are recorded. If approval or another human decision is pending, stop and resume this same step later. Only when every approved, unblocked unit is complete, run `node .github/hooks/scripts/lfg-state.js done solution-debranding-apply --run-id <RUN_ID> --artifact <debranding-plan-path>`.
   - Then run `node .github/hooks/scripts/lfg-state.js is-done solution-debranding-verify --run-id <RUN_ID>`. If it is incomplete, run `/solution-debranding-verify <debranding-plan-path>`. Only on a passing verdict, run `node .github/hooks/scripts/lfg-state.js done solution-debranding-verify --run-id <RUN_ID> --artifact <debranding-plan-path>`.

   Never start with apply or verify. A missing plan artifact, missing approval, human-gated legal/security decision, or failed verify result blocks completion rather than being reported as success. When the decision is resolved and any accepted workflow verifies successfully, run `node .github/hooks/scripts/lfg-state.js done quality-release-readiness --run-id <RUN_ID> --artifact <debranding-plan-path-if-any>`.

6. `/observe` on the areas of code that were changed — analyze patterns in the modified files to capture what was done and how.

7. `/learn` — extract reusable patterns from this session into project instincts.

8. `/compound-engineering-todo-resolve` — then `node .github/hooks/scripts/lfg-state.js done todo-resolve --run-id <RUN_ID>`

9. `/compound-engineering-test-browser` — then `node .github/hooks/scripts/lfg-state.js done test-browser --run-id <RUN_ID> --artifact <report-path>`

10. `/compound-engineering-feature-video` — then `node .github/hooks/scripts/lfg-state.js done feature-video --run-id <RUN_ID>`

11. Output `<promise>DONE</promise>` when video is in PR

Start with step 2 now (or step 1 if ralph-loop is available). Remember: plan FIRST, then work. Never skip the plan.
