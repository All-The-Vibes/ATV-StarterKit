---
title: Strip Claude Code References from Skills for GitHub Copilot Harness
type: refactor
status: active
date: 2026-05-06
---

# Strip Claude Code References from Skills for GitHub Copilot Harness

## Summary

Conservatively normalize Claude Code-specific references (literal "Claude Code", `~/.claude/` paths, Anthropic API examples, Claude Code-only affordances like `AskUserQuestion`) across both the deployed skill surface (`.github/skills/`) and the installer source-of-truth templates (`pkg/scaffold/templates/skills/`). Preserve agent-type names (`general-purpose`, `code-reviewer`, "subagent") and existing semantics — only remove or rephrase content that presumes a Claude Code harness, so the same skill files work cleanly under GitHub Copilot.

---

## Problem Frame

ATV is a GitHub Copilot harness. Recent merges (PR #23–#27) brought in skills derived from the upstream `compound-engineering` plugin and other Claude-Code-shaped sources. Several of those skills still carry literal "Claude Code" / "Anthropic" / `~/.claude/...` references in instructional text and examples. When a Copilot agent loads these skills, it sees instructions describing a harness it isn't running on (e.g., "use `AskUserQuestion` in Claude Code") and either follows them incorrectly, fabricates a Copilot equivalent, or surfaces confusing user-facing copy ("This skill targets Claude Code…"). The user has explicitly flagged `ghcp-review-resolve` and asked for a sweep across all skills.

The scope per user confirmation covers all skill surfaces in this repo, including `agent-native-architecture/`, `create-agent-skills/`, and related meta-docs. Normalization is conservative: replace explicit Claude Code references and `~/.claude/` paths; keep agent-type names where they read as harness-neutral terminology.

---

## Requirements

- R1. No literal "Claude Code" or "Anthropic" string remains in any `SKILL.md`, reference, or workflow file under `.github/skills/` or `pkg/scaffold/templates/skills/` unless it is part of a harness-comparison list (e.g., a multi-platform compatibility table) where its presence is explicitly informative.
- R2. No `~/.claude/` filesystem path remains in skill instructions; references are either removed, rephrased to a harness-neutral location, or replaced with `.github/skills/` (project) and a generic "your agent's user config directory" (personal).
- R3. Anthropic API code examples (`ENV['ANTHROPIC_API_KEY']`, `LM.new('anthropic/...')`) inside skill bodies are removed or replaced with provider-neutral phrasing unless the skill is explicitly a multi-provider how-to where Anthropic is one of N options.
- R4. Claude Code-specific affordances (e.g., "use `AskUserQuestion` in Claude Code", `TodoWrite` tool name) are rephrased to harness-neutral guidance ("ask the user via the platform's blocking question primitive", "track progress in your harness's task list") OR retained inside a clearly multi-platform compatibility note.
- R5. `pkg/scaffold/templates/skills/` and `.github/skills/` stay in sync after the refactor — the same file content lands in both locations for any skill that exists in both. Drift detection in `pkg/scaffold/parity_test.go` continues to pass.
- R6. The `ghcp-review-resolve` skill is verified manually to operate end-to-end after the refactor: preflight, dual review, adjudication, fix loop, summary — no Claude-Code-only assumption blocks any step.
- R7. Existing tests (`go test ./...`) continue to pass; no Go code is changed except where parity tests reference removed/renamed paths.

---

## Scope Boundaries

- Will not redesign skills, change their names, or change the workflows they orchestrate. This is a string/phrasing pass plus targeted rephrases.
- Will not remove agent-type names (`general-purpose`, `code-reviewer`, `subagent`) — these are semantically meaningful and Copilot supports analogous agent invocation patterns.
- Will not touch the upstream `compound-engineering` plugin cache (`/home/sschofield/.claude/plugins/cache/compound-engineering-plugin/...`) — that is read-only vendor content.
- Will not modify `~/.claude/skills/` (user's personal Claude Code skills outside this repo).
- Will not change `pkg/scaffold/templates/agents/*.agent.md` files unless they contain the same Claude-Code-specific patterns; agent personas are out of scope unless flagged during U2.

### Deferred to Follow-Up Work

- A broader "harness-neutral skill style guide" document under `docs/solutions/`: deferred to a follow-up `ce-compound` pass once this refactor lands and we know which patterns recurred.
- Auditing `pkg/scaffold/templates/agents/` for the same drift: deferred to a separate plan unless U2 surfaces blocking issues.

---

## Context & Research

### Relevant Code and Patterns

- `/home/sschofield/repos/ATV-starterkit/.github/skills/` — currently deployed skills in this repo (64 files match the search regex).
- `/home/sschofield/repos/ATV-starterkit/pkg/scaffold/templates/skills/` — installer source-of-truth templates (17 files match). Catalog wiring lives in `pkg/scaffold/catalog.go`; parity is enforced by `pkg/scaffold/parity_test.go`.
- `pkg/scaffold/skills.go` (and `pkg/gstack/skills.go`) — existing skill-handling logic, useful for understanding install layout but not modified by this plan.
- `.github/copilot-instructions.md` — the project's harness-instruction file (does not currently contain Claude refs; serves as the canonical "what harness this is" anchor).

### Institutional Learnings

- No existing `docs/solutions/` entry covers this exact migration. Past plans `2026-04-24-001-fix-ghcp-review-resolve-skill-robustness-plan.md` and `2026-04-24-002-feat-ghcp-review-resolve-dual-subagent-and-resolve-plan.md` already converged the skill on Copilot semantics, so this refactor builds on that foundation rather than rewriting the skill.

### External References

- The `compound-engineering` upstream repo is the source of much of the Claude-Code-flavored phrasing. Their convention "AskUserQuestion in Claude Code, request_user_input in Codex, ask_user in Gemini" is a useful template for harness-neutral phrasing — collapse to "the platform's blocking question primitive" in our copy.

---

## Key Technical Decisions

- **Normalization layer is conservative, per user direction.** Replace explicit "Claude Code" / "Anthropic" / `~/.claude/` references and Claude-Code-only affordances. Keep agent-type names. Rationale: minimize churn and review surface; the goal is a clean Copilot experience, not a full upstream divergence.
- **Source of truth is `pkg/scaffold/templates/skills/`.** When a skill exists in both `.github/skills/` and `pkg/scaffold/templates/skills/`, edit the template first, then mirror to the deployed copy in `.github/skills/`. Rationale: future `atv` installs ship the template; the deployed copy is a shipped artifact for *this* repo. This ordering avoids re-introducing drift.
- **Multi-platform compatibility notes are preserved but trimmed.** Where existing copy lists every harness ("AskUserQuestion in Claude Code, request_user_input in Codex, ask_user in Gemini"), collapse to one neutral phrase ("ask via your harness's blocking question primitive") rather than an exhaustive list. Rationale: ATV is a Copilot harness; we don't need to instruct Copilot's agent on Codex affordances.
- **Drift verification is manual + parity tests.** No new automated regex check is added in this pass. Rationale: a regex check belongs in a separate hardening pass once we've seen which patterns recur. The parity test already enforces filename-level sync.

---

## Open Questions

### Resolved During Planning

- **Scope of refactor:** All skill surfaces, including `agent-native-architecture/` and `create-agent-skills/` references — confirmed by user.
- **Normalization aggressiveness:** Conservative — confirmed by user.
- **Should the upstream `~/.claude/plugins/cache/compound-engineering-plugin/` content be touched?** No — that's vendored read-only content.

### Deferred to Implementation

- Whether any `pkg/scaffold/templates/agents/*.agent.md` files contain the same patterns and need a sibling refactor. U2 will sample a handful; if material drift is found, log a follow-up task instead of expanding scope.
- Whether `claude-permissions-optimizer/` should be removed entirely (its purpose is Claude Code permission tuning). Decision deferred to U4 — likely keep but rename the user-facing guidance, since power-users may still toggle into Claude Code.

---

## Implementation Units

- U1. **Inventory and triage Claude Code references**

**Goal:** Produce a single, reviewable list of every file containing the patterns of interest, classified by treatment (replace literal, rephrase affordance, retain in compatibility note, remove example).

**Requirements:** R1, R2, R3, R4

**Dependencies:** None

**Files:**
- Create: `docs/plans/2026-05-06-001-inventory.md` (working scratch — can be deleted after U5)
- Read: `.github/skills/**`, `pkg/scaffold/templates/skills/**`

**Approach:**
- Run a single grep pass across both surfaces (regex: `(claude code|claude\.md|\.claude/|anthropic|AskUserQuestion|TodoWrite tool)`).
- For each hit, classify: (a) string-literal replace, (b) rephrase affordance, (c) keep in compatibility list (trimmed), (d) drop example block.
- Group hits by skill so U3/U4 can be parallel-friendly.

**Patterns to follow:**
- Output format mirrors `docs/solutions/` table-of-hits style for easy review.

**Test scenarios:**
- Test expectation: none — this unit produces a triage doc, not behavior.

**Verification:**
- The inventory doc lists every file matching the regex with a one-line classification, and the count matches `grep -rli ... | wc -l` (currently 64 deployed + 17 templates).

---

- U2. **Refactor `ghcp-review-resolve` (priority skill flagged by user)**

**Goal:** Bring `ghcp-review-resolve/SKILL.md` to a clean Copilot-neutral state in both surfaces.

**Requirements:** R1, R2, R4, R5, R6

**Dependencies:** U1

**Files:**
- Modify: `.github/skills/ghcp-review-resolve/SKILL.md`
- Modify (if exists): `pkg/scaffold/templates/skills/ghcp-review-resolve/SKILL.md` — verify presence; if absent, log whether installer should ship it (out of scope to add).
- Read: `/home/sschofield/.claude/skills/ghcp-review-resolve/SKILL.md` (upstream variant, reference only)

**Approach:**
- Replace the single `~/.claude/skills/ghcp-review-resolve/lib/ci-classifier.js` reference (line 299 of upstream) with a project-local note or remove if the lib is not shipped here.
- Verify any remaining "subagent" references map to Copilot's agent invocation. The current text already uses harness-neutral "fresh subagent (general-purpose or code-reviewer)" — keep.
- Manually walk every step (Preflight → Step 5) to confirm tool invocations (`gh`, `git`, `Read`, `Edit`) are framed as harness-neutral capabilities.

**Patterns to follow:**
- Existing copilot-instructions.md tone — terse, capability-named, no harness branding.

**Test scenarios:**
- Manual: open the skill in Copilot, invoke `/ghcp-review-resolve` on a small test PR (a stub PR in a branch with one file change). Verify Step 0 preflight runs, Step 4 spawns a subagent, Step 5 fix loop posts a reply. (Integration scenario — no automated coverage; this is documentation that is interpreted at runtime.)
- Automated: existing `pkg/scaffold/parity_test.go` continues to pass after edits.

**Verification:**
- `grep -niE "claude|anthropic" .github/skills/ghcp-review-resolve/` returns zero hits.
- A manual `/ghcp-review-resolve` run on a test PR completes Steps 0–6 without referencing a non-existent harness primitive.

---

- U3. **Refactor remaining skills in `.github/skills/`**

**Goal:** Apply the conservative normalization to the remaining 63 deployed skill files identified in U1.

**Requirements:** R1, R2, R3, R4, R5

**Dependencies:** U1, U2

**Files:**
- Modify: every file from the U1 inventory, scoped to `.github/skills/` (the deployed copy).
- Notable groups:
  - `agent-native-architecture/SKILL.md` and `references/*.md`
  - `create-agent-skills/SKILL.md`, `references/api-security.md`, `references/official-spec.md`, `workflows/*.md`
  - `claude-permissions-optimizer/SKILL.md`, `scripts/extract-commands.mjs`
  - `ce-*` skills (ce-plan, ce-work, ce-review, ce-brainstorm, ce-compound, ce-compound-refresh, ce-ideate)
  - `dspy-ruby/SKILL.md` (Anthropic API examples)
  - `setup/SKILL.md`, `land/SKILL.md`, `report-bug/SKILL.md`, `report-bug-ce/SKILL.md`
  - `feature-video/SKILL.md`, `git-commit*/SKILL.md`, `git-clean-gone-branches/SKILL.md`, `git-worktree/SKILL.md`
  - `document-review/SKILL.md`, `frontend-design/SKILL.md`, `onboarding/SKILL.md`
  - `orchestrating-swarms/SKILL.md`, `resolve-pr-feedback/SKILL.md`, `resolve-pr-parallel/SKILL.md`
  - `test-browser/SKILL.md`, `test-xcode/SKILL.md`
  - `todo-create/SKILL.md`, `todo-resolve/SKILL.md`, `file-todos/*`
  - `ce-review/references/persona-catalog.md`

**Approach:**
- Apply the U1 classification per file. Use Edit (not Write) for surgical changes.
- Preserve frontmatter exactly. Edit only body content.
- For `dspy-ruby`: keep multi-provider tables that legitimately list Anthropic alongside OpenAI/Gemini; remove standalone Anthropic-only examples that imply Anthropic is the default.
- For `claude-permissions-optimizer`: rename the in-body claim "Claude Code permissions" to "your harness's permission allowlist (Claude Code-style settings.json)" and gate the heaviest Claude-only sections behind a clearly labeled "If targeting Claude Code:" header. Keep the script intact (it's still useful when a user does target Claude Code).
- For `agent-native-architecture/references/*.md`: replace `~/.claude/skills/` examples with `.github/skills/` and Copilot-flavored agent-type identifiers where they appear in code blocks.

**Patterns to follow:**
- The U2 result for `ghcp-review-resolve/SKILL.md` is the canonical example.
- Frontmatter preservation per `pkg/scaffold/templates/skills/*/SKILL.md` conventions.

**Test scenarios:**
- Automated: `cd /home/sschofield/repos/ATV-starterkit && go test ./pkg/scaffold/...` continues to pass.
- Spot-check: `grep -rniE "claude code|\.claude/|anthropic" .github/skills/` returns zero hits OR only hits inside explicitly-flagged compatibility sections.
- Spot-check: `git diff .github/skills/` shows only body edits, no frontmatter changes (`grep -P '^[+-]' diff | grep -E '^[+-](name|description):' | wc -l` returns 0).

**Verification:**
- The grep sweep returns zero unexpected hits.
- Manual review of 5 randomly-picked skills confirms the prose still reads naturally and instructions are followable by a Copilot agent.

---

- U4. **Refactor templates in `pkg/scaffold/templates/skills/`**

**Goal:** Mirror the U3 changes into the installer source-of-truth so future `atv` installs ship clean templates.

**Requirements:** R1, R2, R3, R4, R5, R7

**Dependencies:** U3

**Files:**
- Modify: every file from the U1 inventory scoped to `pkg/scaffold/templates/skills/` (17 files).
- Specifically:
  - `ce-work/SKILL.md`, `document-review/SKILL.md`, `land/SKILL.md`
  - `claude-permissions-optimizer/SKILL.md`, `claude-permissions-optimizer/scripts/extract-commands.mjs`
  - `test-browser/SKILL.md`, `ce-review/SKILL.md`, `ce-review/references/persona-catalog.md`
  - `ce-ideate/SKILL.md`, `feature-video/SKILL.md`, `ce-compound/SKILL.md`
  - `setup/SKILL.md`, `ce-plan/SKILL.md`, `ce-brainstorm/SKILL.md`
  - `deepen-plan/SKILL.md`, `atv-security/SKILL.md`, `ce-compound-refresh/SKILL.md`

**Approach:**
- For each template that has a sibling in `.github/skills/`, copy the body changes verbatim, preserving any template-specific markers (e.g., `{{` placeholders, `__INSERT_HERE__`).
- For templates without a deployed sibling (atv-security, deepen-plan), apply the same classification rules from U1.
- Do **not** add or remove files in this unit; that's a structural change requiring catalog wiring updates.

**Patterns to follow:**
- The relationship between `pkg/scaffold/templates/skills/foo/SKILL.md` and `.github/skills/foo/SKILL.md` is "template ships to deployed location"; keeping them character-identical (modulo template markers) is the simplest invariant.

**Test scenarios:**
- Automated: `go test ./pkg/scaffold/...` passes (parity test, install-flow tests).
- Automated: `go vet ./...` passes.
- Spot-check: `diff -ru pkg/scaffold/templates/skills/ce-work/SKILL.md .github/skills/ce-work/SKILL.md` shows no semantic divergence (only template-marker differences if any).

**Verification:**
- `grep -rniE "claude code|\.claude/|anthropic" pkg/scaffold/templates/skills/` returns zero unexpected hits.
- Existing parity test passes.

---

- U5. **Final sweep, confirmation grep, and inventory cleanup**

**Goal:** Confirm zero unexpected hits across both surfaces, run the full test suite, and remove the working inventory file.

**Requirements:** R1, R2, R3, R4, R5, R6, R7

**Dependencies:** U2, U3, U4

**Files:**
- Read: `.github/skills/`, `pkg/scaffold/templates/skills/`
- Delete: `docs/plans/2026-05-06-001-inventory.md` (the U1 working file)
- Run: `go test ./...`, `go vet ./...`, `go build ./...`

**Approach:**
- Final regex sweep: `grep -rniE "(claude code|claude\.md|\.claude/|anthropic|AskUserQuestion)" .github/skills/ pkg/scaffold/templates/skills/`. Any remaining hits must be inside an explicitly-flagged compatibility list, otherwise treat as a regression.
- Run `go test ./...` and ensure green.
- Delete the U1 inventory artifact.

**Test scenarios:**
- Happy path: `go test ./...` passes with zero failures.
- Edge case: a stale Claude reference in a fresh file added between U1 and U5 (e.g., a merge during the refactor) — the final grep catches it.
- Integration: `go build ./...` produces a clean binary; `./atv --help` runs without referencing Claude Code.

**Verification:**
- The final grep returns zero unexpected hits.
- `go test ./...` is green.
- The plan document moves to `status: completed` (handled by `ce-work` at land time).

---

## System-Wide Impact

- **Interaction graph:** None — these are documentation/instruction files. No runtime callbacks change.
- **Error propagation:** N/A.
- **State lifecycle risks:** Only risk is partial application — a half-refactored skill that mixes Copilot-neutral and Claude-Code-flavored copy is more confusing than the current state. U5's final grep is the safety net.
- **API surface parity:** The catalog's filename invariants are unchanged. No skill is renamed or removed.
- **Integration coverage:** Manual verification of `ghcp-review-resolve` end-to-end on a test PR is the only cross-layer integration check; everything else is documentation linting.
- **Unchanged invariants:** The set of skills shipped (`coreSkillDirectories`, `orchestratorSkillDirectories`, `easterEggSkillDirectories` in `pkg/scaffold/catalog.go`) is unchanged. Frontmatter (`name`, `description`) on every skill is unchanged. Agent persona files in `pkg/scaffold/templates/agents/` are unchanged unless U2 surfaces a blocker.

---

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Hidden upstream sync from `compound-engineering` plugin re-introduces Claude refs after this refactor lands | Add a follow-up todo to add a CI grep check; out of scope here. |
| Touching `claude-permissions-optimizer/scripts/extract-commands.mjs` breaks the script for users who *do* target Claude Code | Edit only string-literal user-facing text; leave runtime behavior unchanged. Verify by running the script with `node --check`. |
| Frontmatter edits accidentally change the skill's `name` field, breaking `pkg/scaffold/parity_test.go` | Use `Edit` with explicit `old_string`/`new_string` scoped to body content; never replace frontmatter blocks. U3's diff spot-check confirms zero frontmatter changes. |
| `pkg/scaffold/templates/skills/` and `.github/skills/` drift between U3 and U4 | U4 immediately follows U3 with identical edits; U5 final grep catches drift on either surface. |
| The "inventory doc" U1 produces becomes stale during U2–U4 if iteration uncovers new files | Re-run the U1 grep at the start of U5 as the authoritative final check, not the inventory doc. |

---

## Documentation / Operational Notes

- No CHANGELOG entry needed — this is internal hygiene.
- Update `.github/copilot-instructions.md` only if U1–U5 surfaces a Claude reference there (currently none found).
- Mention the refactor in the next release notes under "Maintenance" if a release ships this work.
- Once the refactor lands, consider opening a follow-up issue: "Add CI grep guard against Claude Code references in skill files."

---

## Sources & References

- User request (this session): "I am still seeing references to claude code in ghcp-review-resolve, and potentially other skills from latest merges. please find and remove those references and make a plan for refactoring to ensure the skills work in github copilot. this is a github copilot harness."
- Related code:
  - `.github/skills/ghcp-review-resolve/SKILL.md`
  - `pkg/scaffold/templates/skills/`
  - `pkg/scaffold/catalog.go`, `pkg/scaffold/parity_test.go`
- Related prior plans:
  - `docs/plans/2026-04-24-001-fix-ghcp-review-resolve-skill-robustness-plan.md`
  - `docs/plans/2026-04-24-002-feat-ghcp-review-resolve-dual-subagent-and-resolve-plan.md`
- Recent merges that brought drift:
  - PR #23 (atv-security skill), PR #24 (atv-security MD updates), PR #25 (land + takeoff Copilot port), PR #27 (guided-install land+takeoff)
