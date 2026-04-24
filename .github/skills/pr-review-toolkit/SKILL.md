<!--
Portions of this file are derived from anthropics/claude-plugins-official
(https://github.com/anthropics/claude-plugins-official/tree/main/plugins/pr-review-toolkit),
licensed under the Apache License, Version 2.0. See ./LICENSE for the full
license text.

Modifications: ported to All-The-Vibes/ATV-StarterKit on 2026-04-24. Frontmatter
adapted to this repo's skill convention (`name` + `description`, kebab-case dir,
SKILL.md filename). Agent references updated to this repo's renamed agents
(`pr-code-reviewer`, `pr-simplification-analyzer`, `pr-comment-analyzer`,
`pr-test-analyzer`, `pr-silent-failure-hunter`, `pr-type-design-analyzer`).
Invocation paths updated to match Copilot CLI's `Skill(...)` syntax.
-->

---
name: pr-review-toolkit
description: Comprehensive multi-agent PR review covering code quality, simplicity, comments, tests, silent failures, and type design. Use when reviewing a PR beyond what a single-pass review catches, or when invoked by `ghcp-review-resolve` as the second reviewer alongside GitHub Copilot.
argument-hint: "[review-aspects]"
license: Apache-2.0 — see ./LICENSE
---

# Comprehensive PR Review

Run a comprehensive pull request review using multiple specialized agents, each focusing on a different aspect of code quality.

**Review Aspects (optional):** "$ARGUMENTS"

## Review Workflow

1. **Determine Review Scope**
   - Run `git status` / `git diff --name-only` to identify changed files.
   - Parse arguments to see if the user requested specific review aspects.
   - Default: run all applicable reviews.

2. **Available Review Aspects**

   | Aspect | Agent | Purpose |
   |--------|-------|---------|
   | `comments` | `pr-comment-analyzer` | Analyze code comment accuracy and maintainability |
   | `tests` | `pr-test-analyzer` | Review test coverage quality and completeness |
   | `errors` | `pr-silent-failure-hunter` | Check error handling for silent failures |
   | `types` | `pr-type-design-analyzer` | Analyze type design and invariants |
   | `code` | `pr-code-reviewer` | General code review for project guidelines |
   | `simplify` | `pr-simplification-analyzer` | Flag simplification opportunities |
   | `all` | (all of the above) | Default — run every applicable review |

3. **Identify Changed Files**
   - `git diff --name-only` for unstaged changes, or use the PR diff if one already exists (`gh pr view --json files`).
   - Map file types to applicable reviews.

4. **Determine Applicable Reviews**

   Based on changes:
   - **Always applicable:** `pr-code-reviewer`
   - **If test files changed:** `pr-test-analyzer`
   - **If comments/docs added or changed:** `pr-comment-analyzer`
   - **If error handling changed:** `pr-silent-failure-hunter`
   - **If types added/modified:** `pr-type-design-analyzer`
   - **After passing review:** `pr-simplification-analyzer` (polish pass)

5. **Launch Review Agents**

   Prefer **parallel** invocation when running `all` — launch every applicable agent in a single turn via the `Task` tool and wait for results to return together. Use **sequential** when the user requested a specific short list and wants each report resolved before the next.

   When invoked by the `ghcp-review-resolve` skill, always run in parallel — the outer skill is time-sensitive.

6. **Aggregate Results**

   After agents complete, summarize:
   - **Critical Issues** (must fix before merge)
   - **Important Issues** (should fix)
   - **Suggestions** (nice to have)
   - **Positive Observations** (what's good)

7. **Provide Action Plan**

   ```markdown
   # PR Review Summary

   ## Critical Issues (X found)
   - [agent-name]: Issue description [file:line]

   ## Important Issues (X found)
   - [agent-name]: Issue description [file:line]

   ## Suggestions (X found)
   - [agent-name]: Suggestion [file:line]

   ## Strengths
   - What's well-done in this PR

   ## Recommended Action
   1. Fix critical issues first
   2. Address important issues
   3. Consider suggestions
   4. Re-run review after fixes
   ```

## Usage Examples

**Full review (default):**
```
Skill(skill="pr-review-toolkit")
```

**Specific aspects:**
```
Skill(skill="pr-review-toolkit", args="tests errors")
# Reviews only test coverage and error handling

Skill(skill="pr-review-toolkit", args="comments")

Skill(skill="pr-review-toolkit", args="simplify")
```

**Parallel (explicit):**
```
Skill(skill="pr-review-toolkit", args="all parallel")
```

## Agent Descriptions

**pr-comment-analyzer** — Verifies comment accuracy vs. code; flags comment rot; checks documentation completeness.

**pr-test-analyzer** — Reviews behavioral test coverage; identifies critical gaps; evaluates test quality.

**pr-silent-failure-hunter** — Finds silent failures; reviews catch blocks; checks error logging.

**pr-type-design-analyzer** — Analyzes type encapsulation; reviews invariant expression; rates type-design quality.

**pr-code-reviewer** — Checks project-guideline compliance (e.g., `CLAUDE.md`, `.github/copilot-instructions.md`); detects bugs; reviews general code quality.

**pr-simplification-analyzer** — Flags unnecessary complexity; improves clarity and readability; preserves functionality.

## Tips

- **Run early:** Before creating a PR, not after.
- **Focus on changes:** Agents analyze the diff by default.
- **Address critical first:** Fix high-priority issues before lower priority.
- **Re-run after fixes:** Verify issues are resolved.
- **Use specific reviews:** Target specific aspects when you know the concern.

## Workflow Integration

**Before committing:**
```
1. Write code
2. Skill(skill="pr-review-toolkit", args="code errors")
3. Fix any critical issues
4. Commit
```

**Before creating a PR:**
```
1. Stage all changes
2. Skill(skill="pr-review-toolkit", args="all")
3. Address all critical and important issues
4. Re-run specific reviews to verify
5. Create PR
```

**After PR feedback:**
```
1. Make requested changes
2. Run targeted reviews based on feedback
3. Verify issues are resolved
4. Push updates
```

## Notes

- Agents run autonomously and return detailed reports.
- Each agent focuses on its specialty for deep analysis.
- Results are actionable with specific `file:line` references.
- This skill is consumed by `ghcp-review-resolve` as the second-reviewer lane alongside GitHub Copilot; when invoked in that context, results are adjudicated before being posted as inline PR comments.
