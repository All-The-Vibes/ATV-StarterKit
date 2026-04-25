---
title: "feat: Add session-state capture step to /land skills (Claude + Copilot)"
type: feat
status: active
date: 2026-04-25
---

# feat: Add session-state capture step to /land skills (Claude + Copilot)

## Overview

The `/land` skill currently ends a session with commit → push → PR → verbal handoff. The verbal handoff is ephemeral — when the next session starts, the new agent has no written record of what was accomplished, what's next, or what tripped us up. Claude Code has a `remember` skill that writes `.remember/now.md` (read by the `SessionStart:clear` hook), but `/land` doesn't invoke it. GitHub Copilot CLI has no `remember` skill at all, but it does read `AGENTS.md` and `.github/copilot-instructions.md` at session start, which means a `.remember/now.md` file becomes durable Copilot context if instructions point Copilot at it.

This plan adds a "Capture session state" step to both the Copilot land skill (`.github/skills/land/SKILL.md`) and the scaffold template (`pkg/scaffold/templates/skills/land/SKILL.md`), and adds a small instruction snippet to `.github/copilot-instructions.md` so Copilot reads `.remember/now.md` on startup. The Claude side has already been updated in `~/.claude/skills/land/SKILL.md` (out of repo) — this plan covers only the in-repo Copilot surfaces.

## Problem Frame

- `/land` ships PRs cleanly but leaves no machine-readable handoff for the next session.
- Verbal summaries vanish into the transcript; the next agent has to re-derive context from `git log` and PR descriptions.
- Claude has `remember` + `SessionStart` hook to bridge sessions; Copilot has neither, but does read AGENTS.md / copilot-instructions.md unconditionally on startup.
- The atv-starterkit ships both Claude and Copilot skills as first-class citizens (see `.github/skills/land/SKILL.md` and `pkg/scaffold/templates/skills/land/SKILL.md`), so the fix needs to land in both.

## Requirements Trace

- R1. The `.github/skills/land/SKILL.md` checklist captures session state to `.remember/now.md` before the verbal handoff, on every land regardless of whether code changed.
- R2. The scaffold template (`pkg/scaffold/templates/skills/land/SKILL.md`) carries the same step so newly scaffolded repos inherit the behavior.
- R3. Step renumbering is internally consistent — every cross-reference inside the file points to the right step number after insertion.
- R4. `.github/copilot-instructions.md` instructs Copilot to read `.remember/now.md` at session start when it exists.
- R5. The session-state file uses a stable, scannable format (branch, PR, accomplished, next, blockers) so any agent can parse it without an SDK.
- R6. The step is shell-only — no skill dependency — so it works whether Copilot, Claude, or a human runs the checklist.
- R7. The final-banner step ("PLANE LANDED") remains the last line of output; the new step sits before the handoff.

## Scope Boundaries

- **Out of scope:** Adding a Copilot-side `remember` skill or chatmode/prompt file. The native `--name`/`--resume` flow is already documented elsewhere; this plan does not wire it into `/land`.
- **Out of scope:** Modifying the Claude `~/.claude/skills/land/SKILL.md` (already done out-of-repo).
- **Out of scope:** Adding Stop hooks, SessionStart hooks, or settings.json changes for either tool.
- **Out of scope:** Changing the `/takeoff` skill to read `.remember/now.md` (Copilot will pick it up via the AGENTS.md instruction; Claude already reads it via the existing SessionStart hook).
- **Out of scope:** Backfilling existing scaffolded repos.
- **Out of scope:** Per-session `--share=<path>` transcript export (different concern; tracked separately if pursued).

## Context & Research

### Relevant Code and Patterns

- `.github/skills/land/SKILL.md` — Active Copilot land skill in this repo (10 steps, currently no session-state capture). Mirrors structure of the Claude skill.
- `pkg/scaffold/templates/skills/land/SKILL.md` — Same content shipped as a scaffold template via the `atv` CLI.
- `.github/copilot-instructions.md` — Repo's instruction file Copilot reads on every session start. Already documents `/land`, `/takeoff`, etc. New "session bookends" guidance fits naturally here.
- `~/.claude/skills/land/SKILL.md` (out of repo, reference only) — Already patched with the Step 9 `remember:remember` invocation; this plan mirrors its shape, but for Copilot which has no skill equivalent.
- `.remember/` directory — Already exists in the working tree (untracked). Created by the Claude `remember` skill earlier in the session. The format (`now.md`, `today-*.md`, `recent.md`) is the de facto standard.

### Institutional Learnings

- `/land` has a documented stale-state-file failure mode (ralph-loop replay) — Step 7 already does belt-and-braces cleanup. The new session-state step should follow the same posture: unconditional, idempotent, cheap.
- The "verbal handoff" pattern was insufficient in practice — this session itself proves the gap (the user had to ask whether session state is captured during land). Writing a file makes the handoff durable.

### External References

- [GitHub Copilot CLI docs — `--continue`, `--resume`, AGENTS.md / copilot-instructions.md auto-loading](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/use-copilot-cli)
- [Copilot CLI slash commands — `/compact`, `/context`, `/model`; no `/remember`](https://docs.github.com/en/copilot/concepts/agents/about-copilot-cli)

## Key Technical Decisions

- **Shell-only implementation, no skill dependency.** Copilot has no `remember` skill, and we want the step to work identically for any agent (or a human) running the checklist. A heredoc into `.remember/now.md` is universal.
- **Use `.remember/now.md` as the canonical filename**, matching the Claude `remember` skill's primary buffer. Both ecosystems converge on the same file.
- **Step 9 placement**, before "Hand off" (Step 10) and "Final banner" (Step 11). The verbal handoff in Step 10 should reference the file. The banner stays the last line of output.
- **Run unconditionally, including on docs-only / no-commit lands.** A buffer is cheap; a missing handoff is expensive.
- **Idempotent overwrite.** `.remember/now.md` is replaced on every land (it's a "now" buffer, not an append-only log). Older state lives in `today-*.md` / `recent.md` if/when the Claude `remember` skill runs.
- **Don't gate on `.remember/` existing** — `mkdir -p .remember` first.
- **Copilot instruction is opt-in-ish** — phrased as "if `.remember/now.md` exists, read it for prior-session context" so it's a no-op in repos that don't use the workflow.
- **No frontmatter changes** — both files already have YAML frontmatter; only the body changes.

## Open Questions

### Resolved During Planning

- **Q: Add a Copilot prompt file (`.github/prompts/land.prompt.md`) too?** — No. Out of scope; current `.github/skills/land/SKILL.md` is the canonical surface.
- **Q: Wire `copilot --name "<branch>-land"` into the step for native Copilot resume?** — No. Adds a Copilot-only line to a tool-agnostic checklist. Track as future work if the resume pattern proves valuable.
- **Q: Should the step also call `/remember` on Claude?** — Out of scope (already handled in `~/.claude/skills/land/SKILL.md`). The in-repo Copilot skill has no equivalent skill to call, and shell-writing the file works for both tools, so we keep the in-repo step pure shell.

### Deferred to Implementation

- **Exact wording of the file template** — happy-path content is "branch / PR / accomplished / next / blockers", but the precise heredoc structure is implementation detail. Mirror the Claude `remember` skill's `now.md` shape if it's already on disk in this working tree.

## Implementation Units

- [ ] **Unit 1: Add Step 9 "Capture session state" to `.github/skills/land/SKILL.md`**

  **Goal:** Insert a new Step 9 that writes `.remember/now.md`, renumber Hand off → 10 and Final banner → 11, and update all internal cross-references.

  **Requirements:** R1, R3, R5, R6, R7

  **Dependencies:** None

  **Files:**
  - Modify: `.github/skills/land/SKILL.md`

  **Approach:**
  - Insert new `### Step 9: Capture session state` section before the existing `### Step 9: Hand off`.
  - Body: short rationale + a heredoc shell snippet that creates `.remember/` and writes `now.md` with branch, PR URL, date, accomplished, next-up, blockers.
  - Renumber existing Step 9 → Step 10, existing Step 10 → Step 11.
  - Update any in-file references to "Step 9" / "Step 10" that meant the old numbers (Step 7's ralph-loop sweep references "Step 9 handoff"; Step 1's "surface it to the user at Step 9"; Final banner's "Step 5 is skipped" — search and patch).
  - Keep Step 11 (Final banner) as the last user-visible content; do not add anything after it.

  **Patterns to follow:**
  - The Claude version `~/.claude/skills/land/SKILL.md` (out-of-repo, reference only) for shape, but use shell instead of `Skill: remember:remember`.
  - Existing voice/tone of the Copilot `.github/skills/land/SKILL.md` — terse, imperative, no fluff.

  **Test scenarios:**
  - Edge case: Run `grep -nE "^### Step" .github/skills/land/SKILL.md` and confirm sequential numbering 1–11 with no gaps or duplicates.
  - Edge case: Run `grep -n "Step [0-9]" .github/skills/land/SKILL.md` and confirm every cross-reference points to a step that still exists with the correct meaning (handoff = 10, banner = 11).
  - Happy path: Run the embedded shell snippet by hand from a worktree; confirm `.remember/now.md` is created with the expected sections populated.
  - Happy path: Final banner is still on the last meaningful line of the document body.
  - Test expectation: none — no executable code, doc-only change. Verification is structural/visual.

  **Verification:**
  - All step headings renumber cleanly.
  - All in-body "Step N" references point to the correct (post-renumber) step.
  - Heredoc snippet is valid bash (no unmatched quotes, no `<<EOF` typos).
  - Final banner remains the last instruction.

- [ ] **Unit 2: Mirror Step 9 into `pkg/scaffold/templates/skills/land/SKILL.md`**

  **Goal:** Apply the identical change to the scaffold template so `atv`-scaffolded repos inherit the behavior.

  **Requirements:** R2, R3, R5, R6, R7

  **Dependencies:** Unit 1 (use the same wording for consistency)

  **Files:**
  - Modify: `pkg/scaffold/templates/skills/land/SKILL.md`

  **Approach:**
  - Diff Unit 1's edit against the template — both files start from the same lineage. Apply the same patch.
  - If the two files have drifted (which they may have), sync the template to match `.github/skills/land/SKILL.md` only for the new step + renumbering. Don't sweep up unrelated drift in this unit.

  **Patterns to follow:**
  - Whatever shape Unit 1 ships.

  **Test scenarios:**
  - Edge case: `diff .github/skills/land/SKILL.md pkg/scaffold/templates/skills/land/SKILL.md` shows only intended divergence (or zero divergence if they were already identical for the affected sections).
  - Happy path: Step numbering 1–11 sequential and consistent.
  - Test expectation: none — doc-only change.

  **Verification:**
  - Same renumbering and cross-reference checks as Unit 1.
  - The template still scaffolds cleanly (no Go template syntax broken — though this file has no `{{ }}` directives, confirm visually).

- [ ] **Unit 3: Add `.remember/now.md` reading instruction to `.github/copilot-instructions.md`**

  **Goal:** Tell Copilot to read `.remember/now.md` at session start when it exists, so the file written by Step 9 actually gets used.

  **Requirements:** R4

  **Dependencies:** None (independent of Units 1/2 — file consumption can land alongside or after producer)

  **Files:**
  - Modify: `.github/copilot-instructions.md`

  **Approach:**
  - Add a small "Session Continuity" section (or extend the existing "Session Bookends" section) with a one-paragraph instruction:
    > At session start, if `.remember/now.md` exists in the repo, read it before responding — it contains a handoff from the previous session (branch, PR, accomplished, next steps, blockers). Treat it as authoritative context for continuing work.
  - Keep wording short and durable; don't reference Claude's hook semantics.

  **Patterns to follow:**
  - Existing voice in `.github/copilot-instructions.md` — short bullets, single source of truth.

  **Test scenarios:**
  - Happy path: Open the file; the new instruction is present, scannable, and not duplicated.
  - Edge case: No-op when `.remember/now.md` doesn't exist — the instruction explicitly conditions on existence.
  - Test expectation: none — doc-only.

  **Verification:**
  - Instruction is present in `.github/copilot-instructions.md`.
  - Wording is conditional (`if exists`) so it doesn't break in repos that haven't run `/land` yet.

## System-Wide Impact

- **Interaction graph:** None. No code paths, callbacks, middleware. Doc-only changes.
- **Error propagation:** N/A.
- **State lifecycle risks:** `.remember/now.md` is overwritten on every land. If a session crashes mid-land before Step 9, no state is captured — same risk profile as the existing verbal handoff. Acceptable.
- **API surface parity:** The Claude `~/.claude/skills/land/SKILL.md` already has equivalent Step 9 (out-of-repo). The two surfaces now agree. The `pkg/scaffold/templates/skills/land/SKILL.md` template propagates to new repos.
- **Integration coverage:** N/A — no automated tests for skill markdown content.
- **Unchanged invariants:** Step 5 (push to remote) is still mandatory and gates the banner. Step 11 (banner) is still the last line. PR-creation behavior unchanged. No new dependencies introduced.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Renumbering misses a cross-reference | Run `grep -n "Step [0-9]" .github/skills/land/SKILL.md` after edit; eyeball every match. |
| Template drift between `.github/skills/...` and `pkg/scaffold/templates/...` | Apply the same diff to both in the same PR. Don't expand Unit 2 to fix unrelated drift. |
| `.remember/` is in `.gitignore` somewhere upstream and `.remember/now.md` never persists for the next session | Check `.gitignore`; if `.remember/` is excluded, that's actually fine — it's session-local state, not source. Document this if it's a gotcha. |
| Copilot ignores the new instruction | Low — Copilot reads `.github/copilot-instructions.md` reliably per docs. Worst case: instruction is silently no-op, which is the current behavior anyway. |
| Heredoc shell snippet has subtle bash incompatibility (e.g., user has `sh` not `bash`) | Use POSIX-compatible heredoc syntax; avoid bashisms. The existing land skill already assumes bash, so consistency is fine. |

## Documentation / Operational Notes

- No release notes needed — this is a quality-of-life change to the `/land` workflow itself.
- The change is self-documenting: the new Step 9 explains what it does and why.
- Future scaffolds (`atv` CLI) inherit the behavior automatically once Unit 2 lands.

## Sources & References

- This plan: `docs/plans/2026-04-25-001-feat-land-skill-session-state-step-plan.md`
- Active Copilot skill: `.github/skills/land/SKILL.md`
- Scaffold template: `pkg/scaffold/templates/skills/land/SKILL.md`
- Repo-wide Copilot instructions: `.github/copilot-instructions.md`
- Reference (out of repo): `~/.claude/skills/land/SKILL.md` already has equivalent Step 9.
- [GitHub Copilot CLI docs](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/use-copilot-cli)
