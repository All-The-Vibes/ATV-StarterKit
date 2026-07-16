---
title: "fix: Resolve 5 deferred SKILL.md issues from PR #42 review"
type: fix
status: active
date: 2026-05-14
---

# fix: Resolve 5 deferred SKILL.md issues from PR #42 review

## Summary

Address the deferred MEDIUM review findings left open on the merged PR #42 (`chore/sync-land-takeoff-ghcp-review-resolve-skills`). Pre-investigation against current `main` shows 5 of the 6 original items remain live; one (`remember:remember` invocation in `land/SKILL.md` Step 9) was already removed before merge. Fixes are applied to the canonical `.github/skills/` copies and every mirror under `pkg/scaffold/templates/skills/` and `plugins/atv-*/skills/` so drift does not re-introduce them. TDD: a bash test harness in `test/skills/test_skill_invariants.sh` asserts each invariant across all mirror copies; tests are written first and fail before fixes are applied.

---

## Problem Frame

PR #42 shipped `/land`, `/takeoff`, and `/ghcp-review-resolve` skills. The ghcp-review-resolve pass on PR #42 surfaced 6 MEDIUM issues that were acknowledged but deferred at the time so the parent PR could merge. Those threads still need fixes — they are real shell-snippet bugs (detached-HEAD breakage, unguarded CLI invocation, missing `state` field on `gh pr view`) and broken command-name references (`/ce:work` vs `/ce-work`). Because every skill is also installed via plugin manifests, the same defects exist in multiple mirrored copies of each SKILL.md — fixes have to land everywhere or drift re-emerges.

---

## Requirements

- R1. Step 8 of every `land/SKILL.md` mirror does not error when invoked outside a branch (detached HEAD) or without an upstream-tracking branch.
- R2. Step 2 of every `takeoff/SKILL.md` mirror guards `backlog sequence list --plain` behind `command -v backlog` so the skill works in repos without `backlog` installed.
- R3. §0c of `.github/skills/ghcp-review-resolve/SKILL.md` requests `state` in the `gh pr view --json` call and stops early when the PR is `CLOSED` or `MERGED`.
- R4. Every `takeoff/SKILL.md` mirror references `/ce-work`, `/ce-ideate`, `/ce-plan` (hyphen form) — never `/ce:work`, `/ce:plan`, `/ce:ideate`.
- R5. A bash test harness exists that fails when any of R1–R4 regress, and passes after fixes are applied.
- R6. Screenshots (terminal captures) of failing-then-passing test runs are recorded under `docs/plans/artifacts/2026-05-14-002/` and referenced from the PR description.

**Out of original scope:** Original Issue 2 (`remember:remember`) — already fixed on `main`; Step 9 of `land/SKILL.md` no longer invokes that skill. Verified during planning; no work needed.

---

## Scope Boundaries

- No behavioral changes to the skills themselves beyond the 4 invariants above. Wording cleanup unrelated to the deferred threads is out of scope.
- No new tests for skill semantics beyond the bash invariant harness — full skill-execution coverage is a much larger effort and not what these review threads asked for.
- The `agents/*.agent.md` sibling-pass (the PR #43 deferred follow-up) is explicitly not in this plan.

### Deferred to Follow-Up Work

- A future `make sync-skills` target (or CI check) that detects drift between `.github/skills/` and the mirror copies — flagged in the prior session as worth considering; not part of this PR.

---

## Context & Research

### Relevant Code and Patterns

**Mirror-copy fan-out** (confirmed via `find . -name SKILL.md`):

| Skill | Mirror locations |
|---|---|
| `land/SKILL.md` | `.github/skills/land/`, `pkg/scaffold/templates/skills/land/`, `plugins/atv-skill-land/skills/land/`, `plugins/atv-pack-shipping/skills/land/`, `plugins/atv-everything/skills/land/` (5 copies) |
| `takeoff/SKILL.md` | `.github/skills/takeoff/`, `pkg/scaffold/templates/skills/takeoff/`, `plugins/atv-skill-takeoff/skills/takeoff/`, `plugins/atv-pack-shipping/skills/takeoff/`, `plugins/atv-everything/skills/takeoff/` (5 copies) |
| `ghcp-review-resolve/SKILL.md` | `.github/skills/ghcp-review-resolve/` (1 copy; not mirrored — single canonical location) |

**Invariant survey on current `main`:**
- All 5 `land/SKILL.md` copies still contain the unguarded `git log "origin/$(git branch --show-current)..HEAD"` (line 168).
- All 5 `takeoff/SKILL.md` copies contain unguarded `backlog sequence list --plain` (line 77) AND 2 occurrences of the `/ce:` colon form (lines 159, 200).
- `.github/skills/ghcp-review-resolve/SKILL.md` §0c (lines 50–56) still omits `state` from the `--json` list.

### Institutional Learnings

- Prior session learning logged in `.remember/now.md`: the 5 mirror copies drift naturally — fixes must be applied to each. This plan operationalizes that by making the test harness fan-out aware: it loops over every mirror and asserts the invariant in each.

### External References

None needed — these are localized shell-snippet fixes with no external API surface.

---

## Key Technical Decisions

- **Bash test harness, not Go test, for an invariant check.** The skill files are documentation; the assertions are about file contents. A 60-line bash script with grep is the right tool. Adding a Go test for a docs invariant would obscure the signal.
- **One test file, multiple invariants, one mirror loop.** The harness in `test/skills/test_skill_invariants.sh` defines 4 invariant checks and runs each over the full mirror list. Adding a new mirror later requires updating one array, not editing per-invariant test files.
- **`grep -E` patterns must be tight enough to catch the regression and loose enough to survive incidental rewording.** Each invariant pins on the exact failure shape (e.g., the literal `git branch --show-current` substitution inside a `git log` argument), not on surrounding prose.
- **Fix scope: just-enough wrapping.** For the detached-HEAD guard, wrap the existing one-liner in `if branch=$(git branch --show-current) && [ -n "$branch" ]; then ... fi`. Don't redesign Step 8. Same for the `backlog` guard.
- **`gh pr view --json state` + early-stop.** Add `state` to the `--json` field list in §0c and add a 1-line guard block after the extraction that aborts with a clear message if `state` is `CLOSED` or `MERGED`.

---

## Open Questions

### Resolved During Planning

- **Q: Is issue #2 (`remember:remember`) still live?** A: No — Step 9 of `land/SKILL.md` on `main` no longer references that skill. Scope reduced to 5 issues.
- **Q: Should we add a sync-skills CI target now?** A: No — flagged as follow-up. Out of scope for this PR; the test harness in this plan catches *recurrence* of these specific bugs, which is the immediate need.

### Deferred to Implementation

- **Exact phrasing of the early-stop message in §0c.** Will be decided when writing the fix — must be informative but short.

---

## Implementation Units

- U1. **Mirror-aware bash test harness**

**Goal:** Add `test/skills/test_skill_invariants.sh` with one function per invariant (R1–R4), each looping over the relevant mirror list. The harness fails loudly with file:line of the offending pattern; passes silently. Make it executable, with a `set -euo pipefail` preamble. Add a tiny `test/skills/README.md` documenting how to run it.

**Requirements:** R5

**Dependencies:** None

**Files:**
- Create: `test/skills/test_skill_invariants.sh`
- Create: `test/skills/README.md`

**Approach:**
- Define top-of-file arrays: `LAND_MIRRORS=(...)`, `TAKEOFF_MIRRORS=(...)`, `GHCP_FILE=...`
- Four functions: `check_land_step8_guard`, `check_takeoff_backlog_guard`, `check_ghcp_state_guard`, `check_takeoff_ce_command_form`
- Each function iterates mirrors, runs a targeted `grep -nE` pattern, reports `FAIL: <file:line> – <reason>` and returns non-zero on hit
- `main` runs all four, accumulates failures, prints a summary, exits non-zero if any failed
- README documents `bash test/skills/test_skill_invariants.sh` invocation and what each check enforces

**Execution note:** Test-first. This unit lands the test harness as a separate commit and **must fail** when run against the unfixed tree. Capture the failing-run terminal output as screenshot #1 before proceeding to U2.

**Patterns to follow:**
- Repo has no existing bash test convention — use plain bash with `set -euo pipefail`, color-coded `OK:` / `FAIL:` output via tput, and exit-code aggregation. Keep it under 150 lines.

**Test scenarios:**
- Happy path: run against current `main` (pre-fix) → harness reports 4 failure categories with file:line list, exits non-zero. Captured as `failing-run.png`.
- Happy path: run against post-fix tree → harness reports all checks OK, exits zero. Captured as `passing-run.png`.
- Edge case: pattern must not match incidental occurrences of similar wording — e.g., the `/ce:` regex must allow `/ce-work` text in surrounding prose without flagging.

**Verification:**
- `bash test/skills/test_skill_invariants.sh; echo $?` returns non-zero before U2–U5 land and zero after.

---

- U2. **Fix R1 — detached-HEAD guard in land/SKILL.md Step 8 (×5 mirrors)**

**Goal:** Replace the unguarded `git log "origin/$(git branch --show-current)..HEAD"` line in Step 8 of all 5 `land/SKILL.md` mirrors with a guarded variant that no-ops when not on a branch or when no upstream tracking exists.

**Requirements:** R1

**Dependencies:** U1

**Files:**
- Modify: `.github/skills/land/SKILL.md`
- Modify: `pkg/scaffold/templates/skills/land/SKILL.md`
- Modify: `plugins/atv-skill-land/skills/land/SKILL.md`
- Modify: `plugins/atv-pack-shipping/skills/land/SKILL.md`
- Modify: `plugins/atv-everything/skills/land/SKILL.md`

**Approach:**
- Replace the single bash line with a small conditional block that:
  1. Captures the current branch into a variable; if empty (detached HEAD), prints a notice and skips the unpushed-commits check
  2. Verifies `origin/<branch>` ref exists with `git rev-parse --verify --quiet "refs/remotes/origin/$branch"`; if missing, prints "no upstream yet — push first" and skips
  3. Otherwise runs the original `git log` and notes the expected-empty contract
- Keep the surrounding prose ("If either check fails, loop back and fix.") intact.

**Patterns to follow:**
- Use the same `if-then-else-fi` shell shape already used elsewhere in the skill where conditional steps appear.

**Test scenarios:**
- Happy path: `check_land_step8_guard` passes on all 5 mirrors post-fix.
- Edge case (manual smoke test, captured as screenshot #3): run the new snippet in a detached-HEAD state on a scratch worktree — verify it prints the notice and exits 0 instead of erroring.

**Verification:**
- `grep -nF 'git log "origin/$(git branch --show-current)..HEAD"' .github/skills/land/SKILL.md` → no matches.
- Test harness `check_land_step8_guard` reports OK.

---

- U3. **Fix R2 — `command -v backlog` guard in takeoff/SKILL.md Step 2 (×5 mirrors)**

**Goal:** In all 5 `takeoff/SKILL.md` mirrors, wrap the `backlog sequence list --plain` invocation in `command -v backlog >/dev/null 2>&1` so missing-CLI doesn't break the skill, and document the fallback (rely on the filesystem scan already declared as source of truth on line 72).

**Requirements:** R2

**Dependencies:** U1

**Files:**
- Modify: `.github/skills/takeoff/SKILL.md`
- Modify: `pkg/scaffold/templates/skills/takeoff/SKILL.md`
- Modify: `plugins/atv-skill-takeoff/skills/takeoff/SKILL.md`
- Modify: `plugins/atv-pack-shipping/skills/takeoff/SKILL.md`
- Modify: `plugins/atv-everything/skills/takeoff/SKILL.md`

**Approach:**
- Replace the bare `backlog sequence list --plain` snippet with:
  ```
  if command -v backlog >/dev/null 2>&1; then
    backlog sequence list --plain
  else
    echo "backlog CLI not installed — skipping sequence dependency scan; filesystem scan above is authoritative"
  fi
  ```
- Adjust the surrounding sentence ("Also pull sequence info…") to read "If `backlog` is installed, also pull sequence info — otherwise the filesystem scan covers it."

**Patterns to follow:**
- Mirrors the `command -v` guard pattern already present in `ghcp-review-resolve/SKILL.md` and elsewhere in the repo.

**Test scenarios:**
- Happy path: `check_takeoff_backlog_guard` passes on all 5 mirrors post-fix.
- Edge case: pattern must not match `backlog sequence list --plain` appearing inside a code-block comment elsewhere — the test pin includes context (no leading whitespace + no leading `if` on the previous non-blank line).

**Verification:**
- `grep -nE '^backlog sequence list --plain$' takeoff-mirror` returns no matches; `grep -nE 'command -v backlog' takeoff-mirror` returns the expected match in each mirror.

---

- U4. **Fix R3 — `state` field + early-stop guard in §0c of ghcp-review-resolve**

**Goal:** Add `state` to the `gh pr view --json` field list in §0c of `.github/skills/ghcp-review-resolve/SKILL.md`, document the new extracted variable `PR_STATE`, and add a 1-paragraph guard after extraction: if `PR_STATE` is `CLOSED` or `MERGED`, emit the preflight table with a "PR is `<state>` — no review needed; bail out" blocker note and stop.

**Requirements:** R3

**Dependencies:** U1

**Files:**
- Modify: `.github/skills/ghcp-review-resolve/SKILL.md`

**Approach:**
- Edit the §0c code block at lines 53–55 to include `state` in the `--json` argument list (keep current order for diff minimization).
- Add a new bullet to the "Extract into local variables" list: `PR_STATE — OPEN | CLOSED | MERGED`.
- Add a 2-line guard subsection between 0c and 0d titled "0c.1 Bail on non-OPEN PRs":
  ```
  If `PR_STATE` is `CLOSED` or `MERGED`, emit the preflight table with a blocker note ("PR is <state> — no review needed") and stop. Closed and merged PRs aren't review targets.
  ```

**Patterns to follow:**
- Same shape as the existing §0e "Check merge state" guard block — short, prose-led, ends with "stop."

**Test scenarios:**
- Happy path: `check_ghcp_state_guard` passes — pattern asserts both that `state` appears in the §0c `--json` field list AND that a guard for `CLOSED|MERGED` exists in nearby prose.

**Verification:**
- `grep -nE 'headRefOid.*state|state.*headRefOid|--json.*state' .github/skills/ghcp-review-resolve/SKILL.md` finds the field in §0c.
- `grep -nE 'CLOSED.*MERGED|MERGED.*CLOSED' .github/skills/ghcp-review-resolve/SKILL.md` finds the new guard.

---

- U5. **Fix R4 — `/ce:` → `/ce-` in takeoff/SKILL.md (×5 mirrors, 2 occurrences each)**

**Goal:** Replace `/ce:work`, `/ce:plan`, `/ce:ideate` (colon form) with their hyphen-form equivalents `/ce-work`, `/ce-plan`, `/ce-ideate` in every `takeoff/SKILL.md` mirror.

**Requirements:** R4

**Dependencies:** U1

**Files:**
- Modify: `.github/skills/takeoff/SKILL.md`
- Modify: `pkg/scaffold/templates/skills/takeoff/SKILL.md`
- Modify: `plugins/atv-skill-takeoff/skills/takeoff/SKILL.md`
- Modify: `plugins/atv-pack-shipping/skills/takeoff/SKILL.md`
- Modify: `plugins/atv-everything/skills/takeoff/SKILL.md`

**Approach:**
- Each mirror has 2 occurrences (lines ~159 and ~200). Replace exactly: `/ce:work` → `/ce-work`, `/ce:plan` → `/ce-plan`, `/ce:ideate` → `/ce-ideate`.
- Use literal string replacement, not regex — keeps the diff trivial to review.

**Patterns to follow:**
- The rest of the repo already standardizes on hyphen-form skill invocations (`/ce-debug`, `/ce-work`, etc.).

**Test scenarios:**
- Happy path: `check_takeoff_ce_command_form` passes on all 5 mirrors post-fix; total `/ce:` colon-form occurrences across the takeoff mirrors equals zero.
- Edge case: legitimate prose containing the literal substring `/ce:` for unrelated purposes is non-existent in the corpus (verified during planning) — but the regex still pins on the three specific command names to avoid false positives.

**Verification:**
- `grep -rn '/ce:work\|/ce:plan\|/ce:ideate' .github/skills/takeoff plugins/*/skills/takeoff pkg/scaffold/templates/skills/takeoff` → no matches.

---

- U6. **Capture screenshots and update PR description**

**Goal:** Record the failing pre-fix and passing post-fix terminal runs of the test harness as PNG screenshots, store under `docs/plans/artifacts/2026-05-14-002/`, and reference them from the eventual PR description.

**Requirements:** R6

**Dependencies:** U1, U2, U3, U4, U5

**Files:**
- Create: `docs/plans/artifacts/2026-05-14-002/failing-run.png`
- Create: `docs/plans/artifacts/2026-05-14-002/passing-run.png`
- Optional: `docs/plans/artifacts/2026-05-14-002/detached-head-smoke.png` (manual smoke of U2's guard)

**Approach:**
- Capture screenshots via the headless-friendly mechanism available in this environment. Prefer imagemagick (`convert -background black -fill white -font Courier label:"@output.txt" out.png`).
- If image-conversion tooling isn't available in the runtime, fall back to checking in plaintext `.txt` artifacts alongside a brief `README.md` noting "PNG conversion tooling not available in this environment — text captures provided instead."
- Embed paths in the PR description so reviewers can verify the TDD flow visually.

**Test scenarios:**
- Test expectation: none — this is an artifact-capture unit, not a feature-bearing one.

**Verification:**
- Files exist at expected paths and are non-empty.
- PR description in U7's commit-and-push step references them.

---

- U7. **Commit, push, open follow-up PR**

**Goal:** Land U1–U6 as a single coherent PR titled `fix: resolve 5 deferred SKILL.md issues from PR #42 review` against `main`. PR description traces each fix to its original review thread and embeds the test-harness screenshots.

**Requirements:** R1–R6 (verification)

**Dependencies:** U1, U2, U3, U4, U5, U6

**Files:**
- No new files beyond those created in U1–U6.

**Approach:**
- One feature-shaped PR rather than 5 micro-PRs — the fixes are tightly related and share one test harness.
- PR body sections: Summary, Scope (5 issues with links to original review threads on PR #42), Test plan (link to harness + screenshots), Mirror impact table.
- Pipeline-driven (LFG): commit/push/PR happens in step 4 of the LFG flow via `/ce-review` autofix or `/ce-commit-push-pr`, not as a manual U-level action.

**Test scenarios:**
- Test expectation: none — this is a delivery unit. The test harness from U1 is what verifies the fix.

**Verification:**
- `gh pr view` on the new PR shows status `OPEN`, title matches, body contains the test-plan and screenshot links.
- CI green on the new PR's head.

---

## System-Wide Impact

- **Interaction graph:** None — SKILL.md files are documentation read by agents at skill-invocation time, not at repo-build time. There is no caller graph to update.
- **Error propagation:** The R1 and R2 guards convert previously-erroring snippets into informative no-ops. No existing caller depends on the erroring behavior.
- **API surface parity:** R3 adds `state` to the `gh pr view --json` field list — additive, no breaking change to any downstream parser of `/tmp/ghcp-pr-meta.json` (the file gets a new key; existing keys unchanged).
- **Mirror parity:** The whole point of this plan. Each invariant is enforced across every mirror via U1's harness.
- **Unchanged invariants:** The skill contracts (what `/land`, `/takeoff`, `/ghcp-review-resolve` are *for*) are unchanged. This is targeted hardening, not a redesign.

---

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Image-conversion tooling not present in runtime — screenshots can't be generated | U6 fallback: commit `.txt` captures with a README explaining the limitation. The TDD evidence is preserved either way. |
| Test harness regex too tight, breaks on benign rewording later | Patterns pin on the exact failure shape (literal substrings inside code fences), not surrounding prose. Re-run harness post-edit to confirm. |
| Edits to one mirror but not all 5 silently re-introduce drift | The harness loops every mirror; a partial fix fails the post-fix run. Drift is caught before commit. |
| `/ce-review` autofix in step 4 of LFG flow misinterprets the plan's deferred-question section as a blocker | The deferred question (exact phrasing of the §0c early-stop message) is intentionally trivial and will be resolved inline during U4 — autofix should not flag it. |

---

## Documentation / Operational Notes

- The new test harness (`test/skills/test_skill_invariants.sh`) is invokable manually but is not yet wired into CI. The follow-up `make sync-skills` work (already deferred above) is the natural home for CI wiring.
- No user-facing docs change — these are internal skill files.

---

## Sources & References

- **Prior session handoff:** `.remember/now.md` (Next list, items 1–6)
- **Original PR:** #42 — `chore/sync-land-takeoff-ghcp-review-resolve-skills` (MERGED)
- **Related code:** `.github/skills/land/SKILL.md`, `.github/skills/takeoff/SKILL.md`, `.github/skills/ghcp-review-resolve/SKILL.md`, and the 9 mirror copies under `pkg/scaffold/templates/skills/` and `plugins/atv-*/skills/`
