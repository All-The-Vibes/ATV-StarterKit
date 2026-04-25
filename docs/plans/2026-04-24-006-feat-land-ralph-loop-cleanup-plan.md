---
title: "feat: Add ralph-loop state cleanup step to land skill"
type: feat
status: completed
date: 2026-04-24
---

# feat: Add ralph-loop state cleanup step to land skill

## Overview

Add an explicit, idempotent cleanup step to the `land` skill that removes any stale `ralph-loop.local.md` state files anywhere under the repo before the final handoff banner. This prevents the next session from being silently trapped in a replayed ralph-loop prompt when state files are orphaned by session crashes, cwd drift into worktrees, or incomplete completion-promise emission.

## Problem Frame

The ralph-loop skill writes a state file at `.claude/ralph-loop.local.md` (relative to cwd at the time it starts). The plugin's stop-hook reads this file and replays the loop's original prompt verbatim as "Stop hook feedback" until the completion promise is emitted as the final tokens of an assistant turn.

Failure modes observed in this session (PR #26 work):

1. **cwd drift** — work moved into `/tmp/ghcp-wt` (a worktree). If the loop started in the main repo, the hook can read the main-repo state file even from the worktree.
2. **Promise placement** — `<promise>DONE</promise>` was emitted but not as the trailing content of the assistant message; the hook treated the loop as still running.
3. **Belt-and-braces cleanup is conditional** — the LFG pipeline's step 10 cleanup only runs *after* DONE is emitted. If the user invokes `/land` without going through LFG (the common path), no cleanup runs and the next session inherits the orphaned state file.

The `land` skill is the natural place to add this cleanup: it is the canonical "I'm finished — close it out" entry point, and Step 7 already cleans up worktrees and stashes. Adding ralph-loop state to that cleanup costs nothing on the happy path (no file → no-op) and prevents the trap on the failure path.

## Requirements Trace

- **R1.** When `land` runs, any `ralph-loop.local.md` files under the current repo (recursively, in any `.claude/` directory) are deleted.
- **R2.** The cleanup is idempotent — running it on a repo with no ralph-loop state files completes silently and does not fail.
- **R3.** The cleanup runs in Step 7 (alongside existing worktree/stash cleanup), before Step 8 verification and Step 9 handoff. It must run **regardless of whether the prior session used ralph-loop** — there is no reliable way to detect that retroactively.
- **R4.** All three copies of the `land` SKILL.md stay in sync: global (`~/.claude/skills/land/SKILL.md`), repo-vendored (`.github/skills/land/SKILL.md`), and scaffold template (`pkg/scaffold/templates/skills/land/SKILL.md`).
- **R5.** Cleanup output is non-noisy: if files were deleted, mention them in the handoff so the user knows the loop was actively running; if nothing existed, stay silent.

## Scope Boundaries

- Not changing the ralph-loop skill itself or the stop-hook behavior — that's a separate concern.
- Not adding ralph-loop cleanup to `takeoff` — takeoff is for starting work, and a stale state file there should still trip the safety net (the user should know if a previous session left a loop running).
- Not adding cleanup to other "completion" skills (`/lfg`, `/ce:work`, etc.) — `land` is the user-facing terminal step. LFG's own step 10 already handles this for the autonomous-pipeline path.
- Not changing the cleanup pattern used by LFG step 10 — keep the same `find ... -delete` idiom for consistency.

## Context & Research

### Relevant Code and Patterns

- `~/.claude/skills/land/SKILL.md` (Step 7 "Clean up" section) — the insertion point. Currently handles `git stash list` and worktree exit.
- `~/.claude/rules/common/land-the-plane.md` Step 6 — the user's private global rules echo the same checklist; if updated globally, this should also be updated for parity. (Note: this file is in `~/.claude/rules/common/`, not the skill directory — confirm whether it's authoritative or descriptive before editing.)
- LFG pipeline step 10 (in user instruction text from this session): `find . -name ralph-loop.local.md -path '*/.claude/*' -delete 2>/dev/null || true` — this is the exact pattern to mirror.
- `docs/solutions/workflow-issues/ralph-loop-stop-hook-blocking-session-exit-2026-04-17.md` — referenced in user global instructions as the recurrence record. Read it for any prior-art constraints before implementing.
- `.github/skills/land/SKILL.md` and `pkg/scaffold/templates/skills/land/SKILL.md` — repo-vendored copies; the repo's `templates` flow propagates the scaffold version to new projects.

### Institutional Learnings

- The user's global instinct (95%+ confidence): "Always run a separate cleanup agent that did NOT write the original code" — for this work, the cleanup is a single-line `find` command, so no agent is needed; this instinct mostly informs that the cleanup must be unambiguous and self-contained.
- The recurrence-history doc explicitly notes: "if the ralph-loop state file is ever orphaned (session crash, cwd drift into a subdirectory with a stale `.claude/ralph-loop.local.md`), the hook replays this string verbatim as Stop hook feedback." This is the failure mode this plan prevents.

### External References

None needed — this is a local convention update.

## Key Technical Decisions

- **Place the cleanup in Step 7, not Step 4 (commit) or Step 5 (push).** Cleanup is a workspace concern, not a code-or-history concern. Step 7 already groups cleanup actions (stashes, worktrees), so ralph-loop fits the section's existing intent.
- **Run it unconditionally.** Detecting "did this session use ralph-loop?" is unreliable (it might have been started in a prior session, by a different agent, or via a hook). The cost of running the find on every `/land` is microseconds; the cost of a missed cleanup is a trapped next session. Run unconditionally.
- **Glob restricted to `*/.claude/*`.** Match the LFG step 10 pattern exactly. Prevents accidentally deleting any unrelated `ralph-loop.local.md` someone might have authored as a real document elsewhere.
- **Capture deleted-file output for the handoff.** If files were deleted, surface them in Step 9 handoff under "blockers / gotchas" so the user knows a loop was active. If none, say nothing.
- **Update all three skill copies in one commit.** The scaffold template propagates to new projects via `pkg/scaffold/templates/skills/land/`, the repo-vendored copy is used for self-hosting, and the global copy is what runs for the current user. Drift between them creates confusion.

## Open Questions

### Resolved During Planning

- **Should the global rules file (`~/.claude/rules/common/land-the-plane.md`) also be updated?** Yes — the global rules file is the user's private ground-truth for what `land` does. Update it for parity. If the rules file is descriptive (a copy of skill behavior) rather than authoritative, the update is cheap insurance. The plan treats it as in-scope.
- **Should the cleanup also cover the home directory (`~/.claude/`)?** No — the LFG step 10 pattern is repo-relative (`find .`). A user-level `ralph-loop.local.md` would be the user's own concern, not something `/land` for a project should touch.

### Deferred to Implementation

- **Exact wording of the new step**: implementer chooses concise phrasing consistent with the existing Step 7 voice ("Clean up" → bulleted commands). Do not add a new top-level step.
- **Whether to print the deleted file path or just a count**: implementer's choice; lean toward path so the user can audit.

## Implementation Units

- [ ] **Unit 1: Add ralph-loop cleanup to global `land` SKILL.md**

**Goal:** Append a ralph-loop state cleanup action inside Step 7 of the global skill.

**Requirements:** R1, R2, R3, R5

**Files:**
- Modify: `~/.claude/skills/land/SKILL.md` (Step 7 — Clean up section)

**Approach:**
- After the existing stash/worktree cleanup bullets, add a new sub-bullet describing the ralph-loop state-file cleanup.
- The shell snippet should match the LFG step 10 pattern: `find . -name ralph-loop.local.md -path '*/.claude/*' -delete 2>/dev/null || true`.
- Add a one-line note: "If files were deleted, surface them in the Step 9 handoff under blockers/gotchas so the user knows a ralph-loop was active."
- Include a brief rationale comment in the section describing *why* (cwd-drift orphans, replayed Stop hook feedback) so future maintainers don't strip it out as redundant.

**Patterns to follow:**
- Existing Step 7 bullet format (terse imperative, short shell snippet, one-line context).

**Test scenarios:**
- Happy path: running `/land` in a repo with no `ralph-loop.local.md` produces no extra handoff content and no errors.
- Edge case: running `/land` in a repo with `.claude/ralph-loop.local.md` deletes it and surfaces the deletion in handoff.
- Edge case: nested ralph-loop file at `subdir/.claude/ralph-loop.local.md` is also deleted.
- Edge case: a real document literally named `ralph-loop.local.md` outside any `.claude/` directory is **not** deleted (path-glob constraint).
- Error path: file system permission denied on the find — the `2>/dev/null || true` swallows the error and `/land` continues to Step 8.

**Verification:**
- `grep -c "ralph-loop" ~/.claude/skills/land/SKILL.md` returns ≥1.
- The new bullet sits inside the Step 7 section, before Step 8.

- [ ] **Unit 2: Sync repo-vendored `land` SKILL.md**

**Goal:** Apply the same change to the repo's vendored copy at `.github/skills/land/SKILL.md`.

**Requirements:** R4

**Dependencies:** Unit 1 (so the canonical text is settled).

**Files:**
- Modify: `.github/skills/land/SKILL.md`

**Approach:**
- Mirror Unit 1's edit verbatim. Avoid drift.

**Test scenarios:**
- Test expectation: none — pure documentation sync. Verified by `diff <(sed -n '/^### Step 7/,/^### Step 8/p' ~/.claude/skills/land/SKILL.md) <(sed -n '/^### Step 7/,/^### Step 8/p' .github/skills/land/SKILL.md)` showing no functional difference.

**Verification:**
- The two Step 7 sections are textually equivalent (modulo any intentional repo-specific notes).

- [ ] **Unit 3: Sync scaffold template**

**Goal:** Apply the same change to `pkg/scaffold/templates/skills/land/SKILL.md` so new projects scaffolded after this change inherit the cleanup step.

**Requirements:** R4

**Dependencies:** Unit 1.

**Files:**
- Modify: `pkg/scaffold/templates/skills/land/SKILL.md`

**Approach:**
- Mirror Unit 1's edit verbatim.

**Test scenarios:**
- Test expectation: none — scaffold template documentation. Verified by diff against Unit 1's text.

**Verification:**
- A new project scaffolded via the repo's templating flow would get the updated Step 7.

- [ ] **Unit 4: Update global rules file for parity**

**Goal:** Reflect the same step in `~/.claude/rules/common/land-the-plane.md` Step 6 ("Clean up") so the user's private global rules stay aligned with the skill.

**Requirements:** R4

**Dependencies:** Unit 1.

**Files:**
- Modify: `~/.claude/rules/common/land-the-plane.md` (Step 6 — Clean up section)

**Approach:**
- Add a parallel bullet matching Unit 1's wording. Keep the rules-file voice (user-instruction tone) consistent with surrounding bullets.

**Test scenarios:**
- Test expectation: none — pure documentation. Verified by reading the section.

**Verification:**
- Section now mentions ralph-loop cleanup alongside stash/worktree cleanup.

## System-Wide Impact

- **Interaction graph:** `/land` invocation → Step 7 cleanup → new `find` invocation. No new cross-skill calls. Independent of `ralph-loop`, `lfg`, or hook code.
- **Error propagation:** `find ... -delete 2>/dev/null || true` is intentionally tolerant. If file system errors prevent deletion, `/land` continues; the next session may still hit the trap, but `/land` itself is not blocked.
- **State lifecycle risks:** None — the cleanup deletes only files matching a narrow glob. No data is mutated outside `.claude/` directories.
- **Unchanged invariants:** The "no merge unless explicitly asked" guardrail, the "push must succeed" rule, and the final banner remain exactly as written.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| User has a legitimate `ralph-loop.local.md` document at a non-`.claude/` path | Path glob `*/.claude/*` excludes such cases by design |
| `find` invocation fails on Windows / WSL edge cases | The `2>/dev/null \|\| true` swallows errors; documented behavior matches LFG step 10 which has been in use without issue |
| Three-file sync drifts over time (one copy updated, others not) | Plan calls for all three in one commit; future maintenance should grep for `ralph-loop` across all three before editing any |
| Scaffold template change breaks downstream projects already using older land template | The change is additive (new bullet); existing scaffolded projects continue to work and benefit from the next pull/regen |

## Documentation / Operational Notes

- No external docs or runbooks reference Step 7 cleanup specifically; no doc updates needed beyond the four files in scope.
- Optionally, add a note to `docs/solutions/workflow-issues/ralph-loop-stop-hook-blocking-session-exit-2026-04-17.md` (if it exists) recording that the `/land` cleanup was added as a defense-in-depth measure. Treat this as a follow-up, not a blocker.

## Sources & References

- Skill file: `~/.claude/skills/land/SKILL.md`
- Repo-vendored copy: `.github/skills/land/SKILL.md`
- Scaffold template: `pkg/scaffold/templates/skills/land/SKILL.md`
- Global rules echo: `~/.claude/rules/common/land-the-plane.md`
- Recurrence record (per user global instructions): `docs/solutions/workflow-issues/ralph-loop-stop-hook-blocking-session-exit-2026-04-17.md`
- Pattern source: LFG pipeline step 10 (user-instruction text, this session)
