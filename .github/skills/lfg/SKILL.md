---
name: lfg
description: Full autonomous engineering workflow
argument-hint: "[feature description]"
disable-model-invocation: true
---

CRITICAL: You MUST execute every step below IN ORDER. Do NOT skip any required step. Do NOT jump ahead to coding or implementation. The plan phase (step 2) MUST be completed and verified BEFORE any work begins. Violating this order produces bad output.

## Step 0 — Mode (gated default, opt-in `auto`)

Parse `$ARGUMENTS` for a mode selector **before** anything else:

- Leading `auto` keyword or an `--auto` flag → **AUTO mode**. Strip the token from
  `$ARGUMENTS` (it is not part of the feature description) and pass `--auto` to the
  `lfg-state.js init` call below so the mode is persisted for resumes.
- Otherwise → **GATED mode** (the default, unchanged behavior).

On **resume**, do not re-parse the keyword — read the persisted mode from
`node .github/hooks/scripts/lfg-state.js status --run-id <RUN_ID>` (`meta.mode`) and
continue in that mode. A missing or unknown `mode` means **gated** (safe default):
a pre-upgrade run never silently becomes autonomous.

If mode is `auto`, announce it once before step 1:

> Running `/lfg` in **AUTO mode** — I'll auto-decide plan/design/impl choices via
> the 6 principles below and stop only for destructive actions (PR push, `/land`,
> deploy). Say "switch to gated" to get the per-gate STOPs back.

Then follow the numbered steps. In auto mode, apply the "## Auto mode" rules below
at every GATE.


## Run State (resumability + artifact passing)

This workflow is **resumable**. A tiny helper tracks which phases are `done` and where each phase's output lives, so re-invoking `/lfg` continues from the first unfinished phase instead of restarting.

- **Helper:** `node .github/hooks/scripts/lfg-state.js` — commands `init`, `bind-plan`, `done <phase> --run-id <id> [--artifact <repo-relative-path>]`, `status --run-id <id>`, `run-id-from-plan --plan <path>`. It writes `.atv/runs/<run-id>/` (gitignored, local-only).
- **On start (resume check):**
  1. If a recent plan for this feature already exists in `docs/plans/`, derive `RUN_ID` from it: `node .github/hooks/scripts/lfg-state.js run-id-from-plan --plan <plan-path>`. Otherwise create a provisional id: run `node .github/hooks/scripts/lfg-state.js init --skill lfg --feature "$ARGUMENTS" --repo <repo> --branch <branch>` (append `--auto` when Step 0 selected AUTO mode) and read `run_id`.
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

5. `/compound-engineering-todo-resolve` — then `node .github/hooks/scripts/lfg-state.js done todo-resolve --run-id <RUN_ID>`

6. `/compound-engineering-test-browser` — then `node .github/hooks/scripts/lfg-state.js done test-browser --run-id <RUN_ID> --artifact <report-path>`

7. `/compound-engineering-feature-video` — then `node .github/hooks/scripts/lfg-state.js done feature-video --run-id <RUN_ID>`

8. Output `<promise>DONE</promise>` when video is in PR

Start with step 2 now (or step 1 if ralph-loop is available). Remember: plan FIRST, then work. Never skip the plan.

## Auto mode (applies only when mode == `auto`)

In AUTO mode the pipeline runs gate-to-gate without stopping to ask, **except** for
destructive actions. Auto-decide every intermediate choice using the 6 Decision
Principles, classify each decision, and keep an auditable log.

### The 6 Decision Principles

1. **Choose completeness** — prefer the complete version (full tests, edge cases,
   error paths); AI makes completeness cheap.
2. **Fix the blast radius** — address the whole failure class, not just the demo path.
3. **Pragmatic** — the simplest thing that fully works; no speculative generality.
4. **DRY** — reuse existing code/flows; don't rebuild what already exists.
5. **Explicit over clever** — readable and obvious beats compact and surprising.
6. **Bias toward action** — when the choice is reversible and low-stakes, decide and
   move; don't stall the pipeline on a two-way door.

### Decision classification (decides whether to auto-decide)

- **Mechanical** (naming, formatting, obvious wiring) → decide **silently**, proceed.
- **Taste** (design trade-offs, structure, scope-within-plan) → decide via the 6
  principles, **log it** (choice + one-line rationale), proceed. Surface the full
  decision log at the final gate.
- **User-Challenge** (destructive, irreversible, or contradicts the plan's intent) →
  **STOP and ask, even in auto mode.** Never auto-decide these.

### Never auto-pass a destructive action

Even in AUTO mode, these STOP for explicit confirmation — auto mode removes the
*decision* round-trips, not the *destructive-action* confirmations:

- pushing a branch / opening or updating a PR
- `/land` (commit + push + PR)
- anything that ships, deploys, or force-updates remote state

### Decision log (surfaced, not silent)

At the final gate (before `<promise>DONE</promise>`), emit an **AUTO DECISION LOG**:
every Taste-class decision you made, with its one-line rationale, so the user can
audit what autonomy chose. Auto mode is fast, not opaque.

