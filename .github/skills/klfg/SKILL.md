---
name: klfg
description: "Full KB pipeline orchestrator. Chains /kb-brainstorm → /kb-plan → /kb-work → kb-complete → DONE. kb-work handles the per-slice gauntlet (scope lock, execution, tests, diff-scope, destructive guard, QA, repair, Figma sync). kb-complete handles post-work quality, follow-up resolution, proof/demo evidence, learning, memory refresh, compaction, and alerts. Use when the user says 'klfg', 'run the full KB pipeline', 'go from brainstorm to done', or wants the hands-off KB vertical-slice workflow."
argument-hint: "[feature description]"
disable-model-invocation: true
---

CRITICAL: You MUST execute every step below IN ORDER. Do NOT skip any required step. Do NOT jump ahead to coding or implementation. The brainstorm (step 1), plan (step 2), work (step 3), and complete (step 4) phases each have a GATE that must verify required evidence before the next step begins. Violating this order produces bad output.

Gate evidence lives in the KB manifest `gate_ledger` once a manifest exists.
Before advancing phases, read `kb-gate/references/gate-ledger.md` when needed
and verify the relevant gate status and `allowed_next_action`. Artifact
existence is necessary but not sufficient.

`klfg` is the strict workflow governor. It must not let the agent assume across
phase boundaries:

- Brainstorm cannot advance while `ask-now` or `research-first` items remain.
- Plan cannot advance without a passed `plan-to-work` gate.
- Work cannot advance while runnable slices remain or slice proof is missing.
- Complete cannot emit DONE without a passed or quarantined `complete-to-ship`
  gate.

This pipeline is interactive in **two specific places** and autonomous everywhere else:

1. **Step 1 (brainstorm)** stops for product Q&A. That is the design. Under `klfg`, once the requirements doc is complete and unblocked, the orchestrator continues to planning.
2. **Step 3 (work)** stops only on slices the manifest flagged `hitl: true` and when safety gates fire (scope violations, destructive commands, QA failures that exhaust repair). `kb-work` handles them and resumes automatically once the user answers.

Everything else — including kb-plan, kb-work, and kb-complete — proceeds without prompting unless a gate blocks. This auto-chaining belongs to `klfg`, not to standalone `kb-brainstorm` or `kb-plan`.

## Pipeline

1. `/kb-brainstorm $ARGUMENTS`

   GATE: STOP. Verify the brainstorm produced a requirements document at `docs/brainstorms/*-requirements.md`. If no requirements doc exists, re-run `/kb-brainstorm $ARGUMENTS` and resume the conversation. Do NOT proceed to step 2 until a written requirements doc exists.

   **Record the requirements doc path.** Refer to it as `<reqs-path>` for the rest of the pipeline.

   Also check the requirements doc for `## Outstanding Questions` → `### Resolve Before Planning`, unresolved `ask-now`, and unresolved `research-first` entries. If any exist, do NOT proceed — return to step 1 and resolve or reclassify them first. `kb-brainstorm` is responsible for not handing off until that section is empty, but verify here as a safety check.

   GATE: run `kb-gate` if brainstorm/document-review surfaced P0/P1/P2/P3 issues. Safe/actionable P0/P1 are rectified by the agent; human-only P0/P1 block planning. P2/P3 get the rectify-all prompt. Record `brainstorm-to-plan` as passed, blocked, or needs-human before invoking `kb-plan`.

2. `/kb-plan <reqs-path>`

   GATE: STOP. Verify `/kb-plan` produced a manifest file at `docs/plans/*-kb-*-manifest.md` and one plan file per slice. If no manifest was created, re-run `/kb-plan <reqs-path>`. Do NOT proceed to step 3 until both the manifest and per-slice plans exist.

   **Record the manifest path.** Refer to it as `<manifest-path>` for the rest of the pipeline.

   GATE: read the manifest `gate_ledger`. `plan-to-work` must be `passed` and `allowed_next_action` must be `kb-work <manifest-path>`. Run `kb-gate/scripts/check_gate_ledger.py <manifest-path> --gate plan-to-work --allowed-next "kb-work <manifest-path>"`. If absent, pending, blocked, stale, or the checker fails, run `kb-gate`/repair planning and do not start work. Safe/actionable P0/P1 are rectified by the agent; human-only P0/P1 block work. P2/P3 get the rectify-all prompt.

3. `/kb-work <manifest-path>`

   `kb-work` executes every pending slice in dependency order, running the full gauntlet per slice:

   **Per-slice gates (all mandatory):**
   - 3.0 Scope Lock — block writes outside `expected_files`
   - 3 Execute — TDD/integration/verification-only
   - 3.5 System-Wide Test Check — trace side effects
   - 3.6 Diff-Scope Verification — git diff vs declared scope
   - 3.7 Destructive Command Guard — block rm -rf, force push, etc.
   - 3.8 QA — lint (all slices) + browser checks (frontend) → kb-repair on failure
   - 3.9 Figma Design Sync — UI slices only

   After all slices: persists scope-verified file list in manifest for kb-complete.

   HITL pauses: slices flagged `hitl: true`, scope violations, destructive commands, QA failures that exhaust repair (5-iteration cap, stuck detection).

   GATE: STOP. After `kb-work` returns, re-read the manifest. Every slice must be `status: done` or `status: skipped`, every completed slice must have a passing `slice-<id>-to-done` gate, and `work-to-complete` must be `passed` with `allowed_next_action` set to `kb-complete <manifest-path>`. Run `kb-gate/scripts/check_gate_ledger.py <manifest-path> --gate work-to-complete --allowed-next "kb-complete <manifest-path>"`. If any slice is `pending`, `in_progress`, or `blocked`, re-run `/kb-work <manifest-path>` to resume.

   If a slice is genuinely stuck (e.g., `blocked` for an external reason), surface that to the user and stop the pipeline. Do not paper over a blocked slice.

4. `/kb-complete <manifest-path>`

   `kb-complete` runs automatically — no pause, no prompt. The whole point of `klfg` is hands-off.

   `kb-complete` runs the post-work quality and learning pipeline:

   - kb-review — full multi-agent code review with scope passthrough from kb-work's gates
   - Resolution Gate — P0/P1 must be fixed before proceeding
   - Follow-up Resolution — resolve or record review/TODO fallout
   - Proof/Demo Evidence — re-run final checks and capture demo evidence when useful
   - Compound + Learn + Evolve — document patterns, extract instincts, promote mature ones
   - Memory Refresh + Compact + Alerts — keep fresh-session memory usable
   - Cleanup — prune ephemeral artifacts (screenshots, old observations)

   GATE: STOP. After `kb-complete` returns, verify the manifest status is `reviewed` and `complete-to-ship` is `passed` or explicitly quarantined with forbidden claims recorded. Run `kb-gate/scripts/check_gate_ledger.py <manifest-path> --gate complete-to-ship --allow-quarantine`. If kb-review found unresolved P0/P1s, or the final proof/cleanup/learning evidence is missing, `kb-complete` must stop — re-run it after fixes.

5. Output `<promise>DONE</promise>` once steps 1–4 are complete.

## Notes

- **Why no `/unslop`:** intentionally omitted. Risk of flagging parallel agent WIP as false positives. Run manually if needed.
- **Why a separate `kb-complete`:** the finish pipeline (review, follow-up resolution, proof/demo evidence, compound, learn, evolve, memory refresh, compact, cleanup, alerts) is a separate skill. `kb-work` invokes it automatically only after all slices are done or intentionally skipped; `klfg` verifies that happened.
- **Why no separate `/kb-review`:** kb-complete runs kb-review at Step 1 with full scope context from kb-work's gates. A second pass would be redundant.
- **Why no separate `/learn` or `/observe`:** kb-complete feeds resolved P0/P1 findings to `.kb/observations.jsonl` (Step 2), runs `/learn` (Step 3), and auto-triggers `/evolve` every 5th KB completion.
- **Why no separate `/ce-compound`:** kb-complete invokes ce-compound at Step 3 for features with novel patterns. Skips automatically for boilerplate.
- **Why no `/land`:** committing, pushing, and opening a PR is a separate, deliberate act. Run `/land` after `klfg` finishes when you're ready to ship.
- **Resuming after interruption:** `klfg` is idempotent across restarts because each step's GATE checks for the produced artifact. If the session is interrupted between steps, re-invoke `klfg` with the same arguments and it will pick up at the first failing GATE.

Start with step 1 now. Remember: brainstorm FIRST, plan SECOND, work THIRD. Never skip a phase.
