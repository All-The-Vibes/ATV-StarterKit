---
title: "refactor: Close PR #28 and strip Claude-Code-specific dependencies from ATV-starterkit"
type: refactor
status: active
date: 2026-04-25
---

# refactor: Close PR #28 and strip Claude-Code-specific dependencies from ATV-starterkit

## Overview

ATV-starterkit is a **GitHub Copilot harness** — a one-command installer that scaffolds `.github/skills/`, `.github/agents/`, `.github/copilot-instructions.md`, and `.github/hooks/copilot-hooks.json` into a target repo so the human can drive Copilot Chat with skills like `/ce-plan`, `/lfg`, `/land`. It is not a Claude Code distribution.

Three categories of Claude-Code-specific surface area have crept in:

1. **PR #28 itself** — adds an unconditional `find . -name ralph-loop.local.md -path '*/.claude/*' -delete` sweep to the `land` skill in both `.github/skills/land/SKILL.md` and `pkg/scaffold/templates/skills/land/SKILL.md`. The `.claude/` directory only exists when the user is also running Claude Code locally; a Copilot-only user has no such directory and the sweep is dead behavior. The PR's stated motivation (a stop-hook replay loop in ralph-loop) is itself a Claude-Code-only failure mode.
2. **The `claude-permissions-optimizer` skill** — ships in both `.github/skills/` and `pkg/scaffold/templates/skills/`. Its scripts read `~/.claude/settings.json` (`extract-commands.mjs:46,74`) and write Claude Code permission allowlists. There is no Copilot equivalent of this surface — Copilot does not have a `settings.json` permissions array. The skill is non-functional for the harness's target audience.
3. **Functional `.claude/` path handling elsewhere** — confined to the two artifacts above. Repo-wide grep for `.claude` in code (`*.go`, `*.json`, `*.yml`, `*.js`, `*.mjs`, `*.sh`) excluding the developer-local gitignored `.claude/settings.local.json` returns only the four files that constitute (1) and (2). No other code paths read or write `.claude/` directories.

This plan closes PR #28 without merging, removes the `claude-permissions-optimizer` skill from both the source-of-truth `.github/skills/` and the scaffold templates, deregisters it from `pkg/scaffold/catalog.go` and `pkg/tui/categories.go`, and updates parity tests and CHANGELOG. Documentation mentions of "Claude Code" as a peer platform (e.g., `AskUserQuestion in Claude Code, request_user_input in Codex, ask_user in Gemini`) are **out of scope** — those are platform-portability annotations inherited verbatim from upstream Compound Engineering content and are not Claude Code dependencies.

## Problem Frame

The harness's installation contract is: "given a target repo, install Copilot-only artifacts into `.github/`." Anything that requires `.claude/` to exist, or a Claude Code runtime, violates that contract. Two artifacts violate it today (the land-skill sweep and the permissions optimizer skill); PR #28 actively expands the violation. We close PR #28 and remove the optimizer skill so the harness's installed footprint is Copilot-only.

## Requirements Trace

- R1. Close PR #28 (https://github.com/All-The-Vibes/ATV-StarterKit/pull/28) without merging, on a clean main.
- R2. Remove the ralph-loop `.claude/`-state-file sweep from both copies of the `land` SKILL (the scaffold template and the repo-vendored copy), reverting Step 7 to the pre-PR-#28 cleanup behavior.
- R3. Remove the `claude-permissions-optimizer` skill directory and its scripts from `.github/skills/` and `pkg/scaffold/templates/skills/`.
- R4. Remove `"claude-permissions-optimizer"` from `orchestratorSkillDirectories` in `pkg/scaffold/catalog.go` and from `atvCategoryMapping` in `pkg/tui/categories.go`.
- R5. Update `pkg/scaffold/parity_test.go` and `pkg/tui/categories_test.go` so they pass with the optimizer removed.
- R6. `go build ./...`, `go vet ./...`, and `go test ./...` all pass.
- R7. CHANGELOG.md records the removal under the unreleased section, attributing the cleanup to "ATV is a Copilot harness; Claude-Code-only tooling is out of scope."
- R8. After the cleanup, repo-wide grep for `.claude` (excluding gitignored developer-local files, brainstorms, and prior plan docs) returns zero functional dependencies.

## Scope Boundaries

**In scope:**
- Closing PR #28 with a comment explaining the harness's Copilot-only contract.
- Deleting the `.claude/` sweep block from both `land/SKILL.md` files.
- Deleting the `claude-permissions-optimizer` skill directory tree (SKILL.md + `scripts/extract-commands.mjs` + `scripts/normalize.mjs`) from both skill homes.
- Removing the skill's catalog and category registrations.
- Fixing parity and category tests.
- Updating CHANGELOG.md.

**Out of scope (explicit non-goals):**
- Rewriting documentation that *mentions* Claude Code as a peer platform (`AskUserQuestion in Claude Code, request_user_input in Codex, ask_user in Gemini` and similar). These are upstream Compound Engineering portability annotations; they document multi-tool support and do not introduce a runtime dependency on Claude Code. Touching them creates large diffs against an upstream source-of-truth and produces no harness behavior change.
- The brainstorm/plan documents in `docs/brainstorms/` and `docs/plans/` that reference Claude Code historically. These are immutable session artifacts.
- The developer-local `.claude/settings.local.json` (gitignored).
- Prior plan documents like `2026-04-24-006-feat-land-ralph-loop-cleanup-plan.md` — leaving the historical plan in place as a record of the abandoned PR is the right move. We do not retroactively rewrite plan history.
- `~/.remember/` instruction file references to Claude Code. Out of repo scope.
- New replacement functionality for the optimizer. Copilot has no equivalent surface. Removing the skill is the change.

## Context & Research

### Relevant Code and Patterns

- `pkg/scaffold/catalog.go:194-202` — `orchestratorSkillDirectories` slice. `"claude-permissions-optimizer"` is the first entry. Removing it follows the established slice-deletion pattern (compare commit history for `meme-iq` additions).
- `pkg/tui/categories.go:49` — `atvCategoryMapping` "shipping" entry: `{Label: "Claude Permissions Optimizer — optimize tool permissions", Key: "orchestrators:claude-permissions-optimizer", Source: "atv"}`. Surrounding entries in the shipping bucket (lfg, slfg, ce-compound, ce-compound-refresh) provide the deletion pattern.
- `pkg/scaffold/parity_test.go` — enforces parity between `.github/skills/` and `pkg/scaffold/templates/skills/`. Will need updating for both removed paths.
- `pkg/tui/categories_test.go` — likely asserts the optimizer entry's presence by key. Inspect and adjust.
- `.github/skills/land/SKILL.md:151-167` and `pkg/scaffold/templates/skills/land/SKILL.md:151-167` — the ralph-loop sweep block introduced by PR #28 that must be removed.

### Institutional Learnings

- The plan `docs/plans/2026-04-02-002-feat-compound-engineering-latest-update-plan.md` documented why `claude-permissions-optimizer` was added to `orchestratorSkillDirectories` rather than a new `utility-skills` layer (YAGNI). The same YAGNI argument now justifies removal: it's a single-purpose skill with no Copilot analogue.
- Prior plans (`2026-04-24-002-feat-port-land-takeoff-skills-copilot-plan.md`) explicitly call out that ports from Claude Code must "remove Claude-Code-specific surface area" and replace with Copilot-equivalent behavior. The optimizer was imported without that translation step.
- `2026-04-24-006-feat-land-ralph-loop-cleanup-plan.md` is the planning document that produced PR #28. Its motivation was real (a stop-hook trap), but the failure mode is Claude-Code-specific and does not belong in the Copilot harness.

### External References

None required — the change is internal removal, not new behavior.

## Key Technical Decisions

- **Close PR #28, do not merge.** Comment on the PR explaining the harness's Copilot-only contract, then `gh pr close 28`. Do not delete the branch immediately — the user can reclaim the work for their personal `~/.claude/` config out-of-tree if they want it.
- **Remove the optimizer skill outright; do not replace.** Copilot's `.github/copilot-instructions.md` permission model is prose-based, not a structured allowlist. There is no equivalent surface to optimize. A "Copilot permissions optimizer" would be vapor.
- **Revert the land-skill Step 7 ralph-loop sweep entirely.** Do not replace it with a Copilot-equivalent because there is no Copilot equivalent — ralph-loop's stop-hook replay is a Claude Code mechanism. The land skill should not pretend to clean up state it cannot create.
- **Treat doc mentions of Claude Code as portability annotations, not dependencies.** Skills like `ce-plan`, `ce-compound`, `report-bug-ce` mention Claude Code in lists like "the platform's blocking question tool when available (`AskUserQuestion` in Claude Code, `request_user_input` in Codex, `ask_user` in Gemini)". These document multi-tool fallbacks and do not bind the harness to Claude Code; rewriting them creates upstream drift and breaks the "Compound Engineering source-of-truth" trace.
- **Sequence: delete files first, then catalog, then tests.** Forces the test failures to surface real callsites rather than letting stale parity tests pass against deletes.

## Open Questions

### Resolved During Planning

- *Should we also strip "Claude Code" doc mentions?* — No. User confirmed the cleanup target is functional dependencies and `.claude/` path handling, not branding scrubs across upstream-imported skill docs.
- *Should the land skill keep a Copilot-flavored cleanup step?* — No. The PR #28 sweep was specifically motivated by a Claude Code stop-hook quirk. Reverting Step 7 to its pre-PR-#28 form is the right outcome.
- *What about prior plan docs that reference Claude Code?* — Leave them. Plans are historical session artifacts.

### Deferred to Implementation

- The exact diff of `parity_test.go` depends on whether the test enumerates skill directories explicitly or uses an embedded-FS walk. Inspect first; choose the minimal edit.
- Whether `categories_test.go` asserts the optimizer's presence by index or by key. The fix shape depends on this.

## Implementation Units

- [ ] **Unit 1: Close PR #28**

**Goal:** Close the open PR without merging, leaving a comment that explains why.

**Requirements:** R1

**Dependencies:** None

**Files:**
- No repo file changes.

**Approach:**
- `gh pr comment 28 --body "..."` with a short note: "Closing — ATV-starterkit is a GitHub Copilot harness. The `.claude/` ralph-loop state-file sweep is Claude-Code-specific and out of scope for the harness's installed footprint. The branch is preserved if anyone wants to reuse the snippet in their personal `~/.claude/` config."
- `gh pr close 28` (do not pass `--delete-branch`).

**Verification:**
- `gh pr view 28 --json state` returns `"state": "CLOSED"`.
- The remote branch `feat/land-ralph-loop-cleanup` still exists.

- [ ] **Unit 2: Remove the ralph-loop `.claude/` sweep from both `land` SKILLs**

**Goal:** Restore Step 7 of the `land` skill to its pre-PR-#28 form.

**Requirements:** R2

**Dependencies:** Unit 1 (close the PR before reverting its content so the PR's history stays coherent)

**Files:**
- Modify: `.github/skills/land/SKILL.md`
- Modify: `pkg/scaffold/templates/skills/land/SKILL.md`

**Approach:**
- In each file, locate the block introduced by PR #28 (around line 151-167 — the `DELETED_RL=$(find ...)` snippet plus surrounding prose). Delete the entire ralph-loop sweep block including its prose justification. The Step 7 cleanup should look as it did before PR #28 (git stash drop + worktree cleanup only).
- Cross-check the two files render byte-identical for the modified region (parity test will enforce this).

**Patterns to follow:**
- The pre-PR-#28 Step 7 form, recoverable via `git show main:.github/skills/land/SKILL.md`.

**Test scenarios:**
- Happy path: After the edit, `grep -n "ralph-loop.local.md" .github/skills/land/SKILL.md pkg/scaffold/templates/skills/land/SKILL.md` returns nothing.
- Happy path: `grep -n '\\.claude' .github/skills/land/SKILL.md pkg/scaffold/templates/skills/land/SKILL.md` returns nothing.
- Integration: `diff <(sed -n '/^### Step 7/,/^### Step 8/p' .github/skills/land/SKILL.md) <(sed -n '/^### Step 7/,/^### Step 8/p' pkg/scaffold/templates/skills/land/SKILL.md)` is empty.

**Verification:**
- The two land SKILLs render identically and contain no ralph-loop or `.claude/` references in Step 7.

- [ ] **Unit 3: Remove the `claude-permissions-optimizer` skill from both skill homes**

**Goal:** Delete the skill directory tree from `.github/skills/` and `pkg/scaffold/templates/skills/`.

**Requirements:** R3

**Dependencies:** None (independent of Units 1-2)

**Execution note:** Execution target: external-delegate (pure file deletion).

**Files:**
- Delete: `.github/skills/claude-permissions-optimizer/SKILL.md`
- Delete: `.github/skills/claude-permissions-optimizer/scripts/extract-commands.mjs`
- Delete: `.github/skills/claude-permissions-optimizer/scripts/normalize.mjs`
- Delete: `.github/skills/claude-permissions-optimizer/` (directory itself)
- Delete: `pkg/scaffold/templates/skills/claude-permissions-optimizer/SKILL.md`
- Delete: `pkg/scaffold/templates/skills/claude-permissions-optimizer/scripts/extract-commands.mjs`
- Delete: `pkg/scaffold/templates/skills/claude-permissions-optimizer/scripts/normalize.mjs`
- Delete: `pkg/scaffold/templates/skills/claude-permissions-optimizer/` (directory itself)

**Approach:**
- `git rm -r .github/skills/claude-permissions-optimizer pkg/scaffold/templates/skills/claude-permissions-optimizer`.
- Confirm both directories no longer exist.

**Test scenarios:**
- Happy path: `ls .github/skills/claude-permissions-optimizer 2>&1` returns "No such file or directory".
- Happy path: `ls pkg/scaffold/templates/skills/claude-permissions-optimizer 2>&1` returns "No such file or directory".
- Edge case: `grep -rn "extract-commands.mjs\|normalize.mjs" .github pkg/scaffold` returns nothing (no orphaned references).

**Verification:**
- Both skill directories are deleted; no other file references the removed scripts by name.

- [ ] **Unit 4: Deregister the optimizer from the scaffold catalog and TUI categories**

**Goal:** Remove the skill from the orchestrator slice and the shipping category mapping.

**Requirements:** R4

**Dependencies:** Unit 3 (remove the underlying files first so the embed.FS walk does not surface a stale entry between commits)

**Files:**
- Modify: `pkg/scaffold/catalog.go`
- Modify: `pkg/tui/categories.go`

**Approach:**
- In `pkg/scaffold/catalog.go`, delete the `"claude-permissions-optimizer",` line from `orchestratorSkillDirectories`. Preserve alphabetical/grouping conventions of the surrounding entries (`feature-video`, `lfg`, `ralph-loop`, `resolve_todo_parallel`, `slfg`, `test-browser`).
- In `pkg/tui/categories.go`, delete the `{Label: "Claude Permissions Optimizer — optimize tool permissions", Key: "orchestrators:claude-permissions-optimizer", Source: "atv"}` entry from the shipping category. Verify category ordering (timeline-readability — start-of-session → end-of-session) is preserved across surrounding entries.

**Patterns to follow:**
- The entry-deletion pattern shown when prior plans removed legacy items from these slices.

**Test scenarios:**
- Happy path: `grep -n "claude-permissions-optimizer" pkg/` returns nothing.
- Integration: `go build ./...` succeeds.
- Integration: `go vet ./...` succeeds.

**Verification:**
- Both files build cleanly with no references to the removed skill.

- [ ] **Unit 5: Update parity and category tests**

**Goal:** Make `pkg/scaffold/parity_test.go` and `pkg/tui/categories_test.go` pass with the optimizer removed.

**Requirements:** R5, R6

**Dependencies:** Units 3 and 4

**Files:**
- Modify: `pkg/scaffold/parity_test.go`
- Modify: `pkg/tui/categories_test.go`

**Approach:**
- Run `go test ./pkg/scaffold/... ./pkg/tui/...` first to see the actual failure shape.
- If `parity_test.go` walks the embed.FS and asserts a hard-coded list, drop `claude-permissions-optimizer` from the expected list. If it walks both trees and asserts equality, no change is needed (the deletes already match).
- If `categories_test.go` asserts a count or specific Key, adjust the expected count or remove the assertion for the optimizer key. Preserve all other assertions verbatim.
- Re-run tests until green.

**Test scenarios:**
- Happy path: `go test ./pkg/scaffold/...` passes.
- Happy path: `go test ./pkg/tui/...` passes.
- Edge case: No remaining test in the repo references the string `"claude-permissions-optimizer"` — `grep -rn "claude-permissions-optimizer" pkg/ test/` returns nothing.
- Integration: `go test ./...` passes top-to-bottom (catches any indirect coverage in `test/sandbox/` or other suites).

**Verification:**
- All Go tests pass; no test references the removed skill.

- [ ] **Unit 6: Update CHANGELOG.md**

**Goal:** Record the removal in the unreleased section with rationale.

**Requirements:** R7

**Dependencies:** Units 1-5

**Files:**
- Modify: `CHANGELOG.md`

**Approach:**
- Add an entry under the unreleased section (or a new unreleased section if none exists) with two sub-bullets:
  - "Removed `claude-permissions-optimizer` skill — ATV-starterkit is a GitHub Copilot harness; Claude Code permission management is out of scope. Users who want this skill can install it directly from upstream Compound Engineering for their personal `~/.claude/` config."
  - "Reverted the `land` skill's Step 7 ralph-loop `.claude/` state-file sweep (closed PR #28 unmerged). The behavior was Claude-Code-specific and produced no value in a Copilot-only installation."
- Match the surrounding CHANGELOG voice (terse, imperative, links where prior entries link).

**Test scenarios:**
- Happy path: CHANGELOG renders without markdown errors (`mdformat --check CHANGELOG.md` if available, else manual visual check).

**Verification:**
- CHANGELOG.md has a clear entry describing both removals and the harness rationale.

- [ ] **Unit 7: Final repo-wide audit**

**Goal:** Confirm zero remaining functional `.claude/`-path or Claude-Code-runtime dependencies in tracked code.

**Requirements:** R8

**Dependencies:** Units 1-6

**Files:**
- No file changes — verification only.

**Approach:**
- Run the audit grep:
  ```
  grep -rn "\\.claude" --include="*.go" --include="*.json" --include="*.yml" --include="*.js" --include="*.mjs" --include="*.sh" \
    --exclude-dir=.git --exclude-dir=.remember --exclude=settings.local.json
  ```
- Expected output: empty.
- Run a second audit for `claude-permissions-optimizer`:
  ```
  grep -rn "claude-permissions-optimizer" --exclude-dir=.git --exclude-dir=.remember --exclude-dir=docs/plans --exclude-dir=docs/brainstorms
  ```
- Expected output: empty.
- Document mentions of "Claude Code" as a peer platform in skill docs are *expected* to remain and are not failures.

**Test scenarios:**
- Happy path: Both audit greps return empty across tracked code (excluding historical brainstorm/plan/changelog narrative).
- Edge case: If anything surfaces, treat it as a missed file — fix in place rather than handwaving.

**Verification:**
- `go build ./... && go vet ./... && go test ./...` all pass on a clean checkout post-audit.
- The two audit greps return empty.

## System-Wide Impact

- **Interaction graph:** The scaffold pipeline (`pkg/scaffold/scaffold.go` → `BuildCatalog` → `skills()`) walks the embed.FS. Removing files from `pkg/scaffold/templates/skills/claude-permissions-optimizer/` removes them from the embedded tree at compile time; the catalog slice change makes the bucket assignment match. The TUI multi-select (`pkg/tui/wizard.go` → `pkg/tui/categories.go`) renders categories from `atvCategoryMapping`; removing the entry removes the choice without changing surrounding ordering.
- **Error propagation:** None expected — file deletion + slice-entry removal are mechanical.
- **State lifecycle risks:** Existing installations have a `.github/skills/claude-permissions-optimizer/` directory in target repos. The harness's uninstaller (`cmd/uninstall.go` / `pkg/scaffold/uninstall.go`) walks tracked skill directories — verify no hard-coded entry for the removed skill exists there. If one does, removing it cleans up; if not, prior installs leave the directory orphan in target repos until the user runs uninstall, which is acceptable.
- **API surface parity:** Two skill homes (`.github/skills/` and `pkg/scaffold/templates/skills/`) must stay in sync. The parity test enforces this; Unit 5 fixes the test to match the new reality.
- **Integration coverage:** The sandbox test suite (`test/sandbox/sandbox_test.go`) installs into a temp dir and asserts the resulting tree. If it asserts the optimizer's installed presence, it must be updated (covered by Unit 5's "go test ./..." gate).
- **Unchanged invariants:** The harness's overall installation contract — `.github/skills/<name>/SKILL.md`, `.github/agents/<name>.agent.md`, `.github/copilot-*.{md,json,yml}`, `.github/hooks/copilot-hooks.json` — is unchanged. Only the optimizer skill exits the catalog. Doc mentions of Claude Code as a peer platform are explicitly preserved.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| `parity_test.go` may use a fixture that hard-codes skill counts in multiple places. | Run `go test` first, fix all surfaced counts in one pass, re-run until green (Unit 5). |
| Sandbox tests in `test/sandbox/` may assert the optimizer's installed footprint in a temp project. | Final `go test ./...` gate in Unit 7 catches this; fix surfaced assertions inline. |
| Closing PR #28 invalidates an outstanding planning document (`2026-04-24-006-feat-land-ralph-loop-cleanup-plan.md`). | Leave the plan in place as a historical record. The CHANGELOG entry from Unit 6 explicitly explains the abandonment. |
| The branch `feat/land-ralph-loop-cleanup` is the user's current working branch. | Cleanup work happens **on the current branch** (not main). The branch will then carry both the original PR #28 commits and the revert/cleanup commits — that's fine, the PR is closed unmerged anyway. The next PR off this branch (or a fresh branch off main) ships the cleanup. |
| User may have local `.claude/settings.local.json` that the optimizer was managing. | Out of scope — the developer-local file is gitignored and unaffected by repo-side changes. |

## Documentation / Operational Notes

- Update CHANGELOG.md (Unit 6).
- No README changes required — the README does not advertise the optimizer skill by name.
- After landing, surface the change in the next week-in-review marketing brief if one is generated. Out of plan scope.

## Sources & References

- PR being closed: https://github.com/All-The-Vibes/ATV-StarterKit/pull/28
- Related code: `pkg/scaffold/catalog.go`, `pkg/tui/categories.go`, `pkg/scaffold/parity_test.go`, `pkg/tui/categories_test.go`
- Prior plans: `docs/plans/2026-04-24-006-feat-land-ralph-loop-cleanup-plan.md` (the plan being abandoned), `docs/plans/2026-04-02-002-feat-compound-engineering-latest-update-plan.md` (where the optimizer was added), `docs/plans/2026-04-24-002-feat-port-land-takeoff-skills-copilot-plan.md` (the "remove Claude-Code-specific surface area" precedent)
- Removed scripts: `.github/skills/claude-permissions-optimizer/scripts/extract-commands.mjs:46` (`process.env.CLAUDE_CONFIG_DIR || join(homedir(), ".claude")`)
