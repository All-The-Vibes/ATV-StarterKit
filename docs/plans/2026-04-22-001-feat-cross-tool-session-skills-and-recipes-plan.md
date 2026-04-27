---
title: "feat: Cross-tool session skills + opt-in recipes for ATV Starter Kit"
type: feat
status: superseded
date: 2026-04-22
origin: docs/brainstorms/2026-04-21-cross-tool-harness-starterkit-requirements.md
---

# feat: Cross-Tool Session Skills + Opt-In Recipes for ATV Starter Kit

> **Resolution (2026-04-25):** Superseded. The session-skills half of this
> plan (`/takeoff`, `/land`) shipped as Copilot skills via PRs #25 and #27
> and the templates now live in `pkg/scaffold/templates/skills/takeoff/`
> and `pkg/scaffold/templates/skills/land/`. The recipes/install-recipe
> half was not pursued; the layered `--guided` install (now extended in
> `docs/plans/2026-04-25-003-refactor-centralize-skills-and-reconcile-drift-plan.md`)
> covers the same opt-in goals with less surface area. Preserved here for
> design history.

> **Target repo:** `ATV-starterkit` (this repo). Repo-relative paths throughout.

## Overview

Teach the `atv-installer` Go CLI to scaffold four cross-tool session skills
(`/takeoff`, `/land`, `/solutions`, `/pr-threads`) and an opt-in `recipes/`
catalog into any project it initializes. Each skill ships with a
tool-neutral Node script and paired Claude + Copilot entry points, so both
agents behave identically. Enforcement hooks (block-push-to-main, harness
exclusion, backlog gating) live in `recipes/` and are never installed by
default — adopters opt in via a new `install-recipe` companion script.

The original requirements were written against a different repo
(`ai-platform-accelerator`). This plan keeps the intent but re-homes the
work: the **templates live in `pkg/scaffold/templates/`** in this repo, and
the **installer scaffolds them** into target directories. No runtime
behavior is added to the installer — only new template files and two new
catalog entries in `pkg/scaffold/catalog.go`.

## Problem Frame

ATV-starterkit today scaffolds skills, agents, MCP config, setup-steps,
and file-instructions — but it does not scaffold the session-workflow
primitives (start-of-session digest, end-of-session land, knowledge-base
writer, PR-thread audit) that the origin brainstorm identifies as the
current gap. Worse, the enforcement behaviors the starterkit author uses
personally (push-to-main blocking, harness-path exclusion) are friction
for adopters. Shipping everything through the scaffolder without the
cross-tool + opt-in framing would repeat the problem.

See origin: `docs/brainstorms/2026-04-21-cross-tool-harness-starterkit-requirements.md`.

## Requirements Trace

- R1. A project scaffolded by `atv-installer init` can invoke `/takeoff`,
  `/land`, `/solutions`, `/pr-threads` from both Claude Code (skill
  trigger) and GitHub Copilot (prompt picker) and get identical behavior.
- R2. No hook ships enabled by default. A fresh `atv-installer init` does
  not create any husky hook or modify any Claude settings hook.
- R3. Adopters can opt into enforcement recipes via a single idempotent
  command.
- R4. Skill logic lives in Node scripts under `scripts/` (in the
  scaffolded project); Claude `.claude/skills/*/SKILL.md` and Copilot
  `.github/prompts/*.prompt.md` are thin narration wrappers (≤ ~30 lines
  each) that consume a shared `--json` contract from the script.
- R5. Backlog.md-dependent sections of `/takeoff` and `/land` degrade
  gracefully when the scaffolded project has no `backlog/` directory.
- R6. The `/land` script is advisory (never mutates git state); the
  narrator is responsible for staging/committing/pushing.
- R7. Re-running `atv-installer init` on a project that already has the
  new templates is idempotent (existing `scaffold.WriteAll` already
  skips existing files — this plan does not change that).
- R8. Installer tests cover the new catalog entries (scripts, recipes,
  new skills, new prompts) the same way existing entries are covered.
- R9. The four skill trigger keywords (`takeoff`, `land`, `solutions`,
  `pr-threads`) are surfaced in the post-install "next steps" output so
  new users know they exist.

## Scope Boundaries

- Does NOT change the installer's public CLI surface (`atv-installer init`
  still runs, still auto-detects, still guided-mode-optional).
- Does NOT alter existing hooks 1–6 in `pkg/scaffold/catalog.go` — this
  plan only adds new component categories (scripts, recipes, prompts)
  and new skill/prompt template files.
- Does NOT install any husky hook, Claude PreToolUse hook, or hookify
  rule by default.
- Does NOT build Copilot `.agent.md` personas for the four skills (they
  are task-style prompts, not persona agents).
- Does NOT replace or require Backlog.md; skills degrade when absent.
- Does NOT remove or modify `.github/skills/land/` or any existing
  template if present — new skills use their own names.

## Context & Research

### Relevant Code and Patterns

- `pkg/scaffold/catalog.go`:
  - `BuildCatalog(stack)` (line 26) — orchestrates all 6 hook categories.
  - `skills()` (line 132) — `fs.WalkDir` over `templates/skills/` picks
    up new skills automatically. Dropping a new skill directory under
    `pkg/scaffold/templates/skills/<name>/SKILL.md` is sufficient for
    Hook 4 pickup.
  - `agents(stack)` (line 152) — same walk pattern for agents.
  - `directories()` (line 93) — where to add new always-created
    directories (e.g., `scripts/`, `recipes/`, `.claude/skills/`,
    `.github/prompts/`).
- `pkg/scaffold/scaffold.go` — `WriteAll` / `WriteFile` idempotency
  semantics (create | skip | merge-JSON). New files participate in the
  same contract.
- `pkg/scaffold/templates/skills/` — 11 existing skill templates. New
  skill directories follow the same shape (`SKILL.md` plus optional
  sub-files).
- `pkg/output/printer.go` — `PrintNextSteps(stack)`. R9 wiring happens
  here.
- `cmd/init.go` — no change required; new catalog entries flow through
  existing `scaffold.BuildCatalog(env.Stack)` path.
- `.github/skills/ce-plan/SKILL.md` — format reference for new SKILL.md
  frontmatter.
- `.github/prompts/*.prompt.md` — not present in this repo yet; new
  category, but the scaffolded format follows the documented Copilot
  `description` / `argument-hint` / optional `agent` frontmatter.
- Prior plan `docs/plans/2026-03-11-002-feat-atv-starter-kit-guided-installer-plan.md`
  documents the catalog/idempotency model; this plan extends it.

### Institutional Learnings

- From the prior guided-installer plan: writing directly to the target
  dir, never overwriting, always reporting status. All additions here
  follow that rule.
- From `CLAUDE.md` convention across the ecosystem: `docs/solutions/`
  frontmatter is `module`, `tags`, `problem_type`. `/solutions new` must
  write that shape.

### External References

None — skipping external research. Go `embed.FS` patterns and the four
script behaviors are well-covered by local precedent.

## Key Technical Decisions

- **Templates live under `pkg/scaffold/templates/`.** Four skills go
  under `templates/skills/<name>/`; four Copilot prompts go under a new
  `templates/prompts/` tree; four Node scripts go under a new
  `templates/scripts/`; six recipes and a `README.md` go under a new
  `templates/recipes/` tree; one `install-recipe.mjs` companion script
  under `templates/scripts/`. Each tree gets its own catalog function
  mirroring the existing `skills()` walk.
- **Add three new catalog categories** to `pkg/scaffold/catalog.go`:
  `scripts()`, `recipes()`, `copilotPrompts()`. Wire them into
  `BuildCatalog` and `BuildFilteredCatalog`. New layers in guided mode:
  `session-scripts`, `recipes`. Default (one-click) mode includes
  session-scripts; `recipes` stays off by default in BOTH modes (opt-in
  only via `install-recipe.mjs` after the fact).
- **Scripts use `.mjs` ESM + `node:test`, no new npm deps.** Consistent
  with the broader ecosystem pattern. Tests for scripts live alongside
  the scripts in the scaffolded project; inside THIS repo, the tests are
  just embedded-file-presence tests (the installer does not execute the
  scripts).
- **Every skill script emits `--json`.** Both narrators consume the same
  contract. Human-readable fallback when run directly in a terminal.
- **`/land` is advisory.** Never runs `git add`/`commit`/`push`. Narrator
  does mutations.
- **Recipes are documentation-plus-manifest files.** Each
  `recipes/<category>/<name>.{sh,md}` starts with a YAML frontmatter
  block declaring `install_to`, `mode`, and `action` (`copy` |
  `merge-json` | `append`). `install-recipe.mjs` reads the manifest and
  performs the install. Uninstall = manual `rm`.
- **New skills and prompts use distinct names.** `land-starterkit` is
  NOT used here — this repo has no collision with a personal override.
  The scaffolded skill is simply `land`. Adopters who later want the
  author's harness-exclusion behavior install
  `recipes/land/harness-exclusion.md`.
- **Stack-agnostic.** All four skills work for Rails, Python,
  TypeScript, and General stacks. No `isStackSpecific` filtering.

## Open Questions

### Resolved During Planning

- Where does `scripts/` live in the scaffolded project? → Repo root of
  the scaffolded project (matches prior art in the ecosystem).
- Does the starterkit `/land` collide with any existing template? → No
  (this repo has no `land` skill today; the collision concern was
  specific to `ai-platform-accelerator` and does not apply here).
- Does `/land` refuse on `main`? → Warns only (no default coercion).
- Is `/solutions` Backlog-aware? → No; `docs/solutions/` only.
- Does `install-recipe` support uninstall? → No; manual delete.
- Do recipes get installed in guided mode? → No. Recipes are always
  opt-in post-install.

### Deferred to Implementation

- Exact ranking algorithm for `/solutions search`. Start with
  frontmatter-field + substring scoring; revisit if relevance is poor.
- Recently-closed-task heuristic for `/takeoff` (git log walk vs.
  Backlog.md `updated` field). Decide against live repo state.
- `/pr-threads` pagination. Defer until observed need >30 open PRs.
- Whether to add Haiku-friendly compact mode to `/takeoff` narrator.
  Out of scope.

## High-Level Technical Design

> *This illustrates the intended structure and is directional guidance
> for review, not implementation specification. The implementing agent
> should treat it as context, not code to reproduce.*

Template tree (added to `pkg/scaffold/templates/`):

```
templates/
├── skills/
│   ├── takeoff/SKILL.md              (new — wraps scripts/takeoff.mjs)
│   ├── land/SKILL.md                 (new — wraps scripts/land.mjs)
│   ├── solutions/SKILL.md            (new — wraps scripts/solutions.mjs)
│   └── pr-threads/SKILL.md           (new — wraps scripts/pr-threads.mjs)
├── prompts/                          (NEW top-level dir)
│   ├── takeoff.prompt.md
│   ├── land.prompt.md
│   ├── solutions.prompt.md
│   └── pr-threads.prompt.md
├── scripts/                          (NEW top-level dir)
│   ├── README.md
│   ├── lib/harness.mjs
│   ├── takeoff.mjs
│   ├── land.mjs
│   ├── solutions.mjs
│   ├── pr-threads.mjs
│   └── install-recipe.mjs
└── recipes/                          (NEW top-level dir)
    ├── README.md
    ├── husky/
    │   ├── block-push-to-main.sh
    │   ├── harness-leak-canary.sh
    │   └── update-backlog-on-fix.sh
    ├── claude/
    │   ├── require-backlog-task.md
    │   └── link-plan-backlog.md
    └── land/
        └── harness-exclusion.md
```

Catalog wiring (`pkg/scaffold/catalog.go`):

```
BuildCatalog(stack)
├── directories()            (+ scripts/, scripts/lib/, recipes/,
│                              .claude/skills/, .github/prompts/)
├── systemInstructions(stack)
├── setupSteps(stack)
├── mcpConfig()
├── skills()                 (unchanged — picks up 4 new subdirs)
├── agents(stack)
├── fileInstructions(stack)
├── vscodeConfig()
├── sessionScripts()         ← NEW (scripts/ tree, HookType=7)
└── copilotPrompts()         ← NEW (.github/prompts/, HookType=8)
```

`recipes()` is registered but ONLY included when the guided-mode user
ticks `recipes` — it is never in the one-click default. Runtime
installation of a recipe is deliberately pushed to
`scripts/install-recipe.mjs` (post-init, adopter-initiated).

## Implementation Units

- [ ] **Unit 1: Template scaffolding — directory + shared helper**

**Goal:** Create the new template directory structure and the shared
`scripts/lib/harness.mjs` helper that the four skill scripts depend on.

**Requirements:** R1, R4, R5

**Dependencies:** None.

**Files:**
- Create: `pkg/scaffold/templates/scripts/README.md`
- Create: `pkg/scaffold/templates/scripts/lib/harness.mjs`
- Create: `pkg/scaffold/templates/recipes/README.md`

**Approach:**
- `scripts/README.md` (scaffolded into target project) documents the
  exit-code contract (`0` ok, `1` expected signal, `2` unexpected error),
  the `--json` convention, and the test command (`node --test
  scripts/**/*.test.mjs`).
- `recipes/README.md` (scaffolded into target project) states clearly
  that nothing in `recipes/` is active until installed, and shows the
  install one-liner.
- `lib/harness.mjs` exposes pure helper functions used by later units:
  - `readGitState()` — branch, is-main, uncommitted file list, unpushed
    commits. Works outside git (returns empty fields).
  - `hasBacklog()` — presence of `backlog/`.
  - `readBacklogTasks(status?)` — parses frontmatter from
    `backlog/tasks/*.md`; `[]` on missing dir.
  - `readSolutionDocs()` — parses `docs/solutions/*.md` frontmatter; `[]`
    on missing dir.
  - `parseFrontmatter(text)` — minimal YAML frontmatter reader (regex,
    no deps).

**Patterns to follow:**
- File layout: mirror the `templates/skills/<name>/SKILL.md` shape for
  depth.
- Content discipline: these are TEMPLATE files — the installer embeds
  them verbatim. Do not reference absolute paths.

**Test scenarios:**
- Integration: `go test ./pkg/scaffold/...` continues to pass after
  these files are added (no regression in existing embedded walks).
- Happy path: a new unit test confirms `templates/scripts/lib/harness.mjs`
  is embedded (non-empty bytes returned from `templateFS.ReadFile`).
- Happy path: `templates/scripts/README.md` and
  `templates/recipes/README.md` are embedded.

**Verification:**
- `go test ./pkg/scaffold/...` green.
- `grep -r "scripts/lib/harness.mjs" pkg/scaffold/templates` finds the
  file.

---

- [ ] **Unit 2: Four skill scripts — takeoff, land, solutions, pr-threads**

**Goal:** Author the four Node scripts that the Claude and Copilot
narrators will invoke.

**Requirements:** R1, R4, R5, R6

**Dependencies:** Unit 1 (shared helper).

**Files:**
- Create: `pkg/scaffold/templates/scripts/takeoff.mjs`
- Create: `pkg/scaffold/templates/scripts/land.mjs`
- Create: `pkg/scaffold/templates/scripts/solutions.mjs`
- Create: `pkg/scaffold/templates/scripts/pr-threads.mjs`

**Approach:**
- `takeoff.mjs`: emits digest sections (branch+main-flag, uncommitted,
  unpushed, epic-prefix, in-progress tasks, recently-closed
  divergence). `--json` flag for narrator; human-readable default.
  Read-only; exit `0` unless internal error (`2`).
- `land.mjs`: advisory only. Emits branch status, quality-gate
  candidate commands from `package.json` scripts, file categorization
  (user-authored vs untracked, no harness filtering), proposed
  conventional-commit type heuristic (test-only → `test:`, docs-only →
  `docs:`, else `feat:`/`fix:` by diff summary), push readiness (remote,
  unpushed count, open PR via non-fatal `gh pr view --json state`),
  backlog task IDs referenced from commits. Never mutates git.
- `solutions.mjs`: `new` subcommand creates
  `docs/solutions/YYYY-MM-DD-<slug>.md` with frontmatter (`module`,
  `tags`, `problem_type`, `title`, `date`); refuses to overwrite.
  `search <query>` reads all solution docs, ranks by frontmatter-field
  exact match > frontmatter-value partial match > body substring.
  Both support `--json`.
- `pr-threads.mjs`: `gh pr list --author @me --state open` plus `gh api
  graphql` for per-PR `PullRequestReviewThread.isResolved`. Computes
  `unresolved_threads` and `needs_reply` (author commit newer than last
  reviewer comment on an unresolved thread, no subsequent reply).
  Supports `--offline --fixture <path>` for deterministic tests.
  Read-only, exit `2` if `gh` unavailable.
- All scripts import from `./lib/harness.mjs`.

**Patterns to follow:**
- ESM (`.mjs`), `process.argv` parsing (no deps), clear exit codes
  documented at top of each file.

**Test scenarios:**
- Happy path: embedded-file test — each of the four scripts appears in
  `templateFS` and is non-empty.
- Integration: a new installer test invokes `BuildCatalog(stack)` and
  asserts the four script components are present with
  `destPath` values of `scripts/takeoff.mjs` etc. and `HookType` set to
  the new session-scripts hook value (see Unit 4).
- (Script behavior is covered by fixture-driven `*.test.mjs` files that
  are authored AS TEMPLATES and written into the scaffolded project —
  see Unit 3 note about test templates.)

**Execution note:** Scripts are template content. The installer does
not execute them. Author them with inline usage docstrings so a user
who runs `node scripts/takeoff.mjs --help` sees a summary.

**Verification:**
- Embedded-presence test green.
- Reading each script as text shows: shebang-style header or ESM
  `#!/usr/bin/env node`, exit-code comment, inline `--help` text.

---

- [ ] **Unit 3: Script test templates (scaffolded alongside scripts)**

**Goal:** Ship `*.test.mjs` files as templates so scaffolded projects
get runnable `node --test` coverage for the four scripts and the
installer.

**Requirements:** R4

**Dependencies:** Unit 2.

**Files:**
- Create: `pkg/scaffold/templates/scripts/lib/harness.test.mjs`
- Create: `pkg/scaffold/templates/scripts/takeoff.test.mjs`
- Create: `pkg/scaffold/templates/scripts/land.test.mjs`
- Create: `pkg/scaffold/templates/scripts/solutions.test.mjs`
- Create: `pkg/scaffold/templates/scripts/pr-threads.test.mjs`
- Create: `pkg/scaffold/templates/scripts/install-recipe.test.mjs`
- Create: `pkg/scaffold/templates/scripts/fixtures/pr-threads-sample.json`
- Create: `pkg/scaffold/templates/scripts/fixtures/backlog-sample/tasks/.gitkeep`

**Approach:**
- Each test file uses only `node:test` and `node:assert`. No new deps.
- Tests are authored as templates that WILL BE RUN in the scaffolded
  project — NOT in this repo. In this repo, they are inert embedded
  bytes.
- Script test contents cover the scenarios enumerated in the origin
  requirements: happy path, missing-backlog degradation, empty
  working-tree, `--json` shape, commit-type heuristic buckets (tests-only
  / docs-only / else), `gh`-unavailable branch for pr-threads, etc.

**Patterns to follow:**
- Mirror the existing test style in the ecosystem; keep fixtures
  deterministic (static JSON, static fake backlog tasks).

**Test scenarios:**
- Embedded-presence test confirms all eight new template files are in
  `templateFS`.
- Integration: an installer-level test walks `BuildCatalog(stack)` and
  confirms every `*.test.mjs` is in the output with its `destPath` set
  under `scripts/`.

**Verification:**
- `go test ./pkg/scaffold/...` green.
- `ls pkg/scaffold/templates/scripts/fixtures/` shows the fixture
  files.

---

- [ ] **Unit 4: Catalog wiring — `sessionScripts()` + directories**

**Goal:** Extend `pkg/scaffold/catalog.go` so the installer actually
scaffolds the scripts tree into target projects. Introduces the new
session-scripts category.

**Requirements:** R1, R4, R7, R8

**Dependencies:** Units 2, 3.

**Files:**
- Modify: `pkg/scaffold/catalog.go`
- Modify: `pkg/scaffold/catalog_test.go` (if absent, create it)
- Modify: `pkg/scaffold/hooks.go` (to register a new HookType constant
  if that file centralizes them)

**Approach:**
- Add `HookType = 7` constant (or whatever the next free number is in
  `hooks.go`) named "Session Scripts".
- Add `sessionScripts()` function mirroring the `skills()` walk:
  walks `templates/scripts/` (recursively), maps each file to a
  `Component{Path: filepath.Join("scripts", relPath), HookType: 7}`.
- Wire `sessionScripts()` into `BuildCatalog` after `vscodeConfig()`.
- Add a `session-scripts` layer to `BuildFilteredCatalog`. In the
  guided UI (see Unit 7), it defaults ON.
- Update `directories()` to include `scripts/`, `scripts/lib/`,
  `scripts/fixtures/`, `scripts/fixtures/backlog-sample/tasks/`.

**Patterns to follow:**
- `skills()` (lines 132–150 of `catalog.go`) — exact walk shape,
  including the `err != nil || d.IsDir() || path == "templates/scripts"`
  guard.

**Test scenarios:**
- Happy path: `BuildCatalog(stack.TypeScript)` returns components
  whose paths include `scripts/takeoff.mjs`, `scripts/land.mjs`,
  `scripts/solutions.mjs`, `scripts/pr-threads.mjs`,
  `scripts/lib/harness.mjs`, `scripts/install-recipe.mjs`,
  `scripts/README.md`, and the six test files.
- Happy path: all returned components for scripts have `HookType: 7`.
- Edge case: `BuildFilteredCatalog(stack, [])` (no layers) does NOT
  include any `scripts/` components.
- Edge case: `BuildFilteredCatalog(stack, ["session-scripts"])` DOES
  include them.
- Happy path: all four stacks (Rails, Python, TypeScript, General) get
  identical scripts output (stack-agnostic).
- Regression: existing hook categories (1–6 + vscode) still produce the
  same set of components they did before.

**Verification:**
- `go test ./pkg/scaffold/...` green, including new assertions.
- `go build ./...` green.

---

- [ ] **Unit 5: Claude skill templates — takeoff / land / solutions / pr-threads**

**Goal:** Author the four Claude skill wrappers that narrate the scripts.
Picked up automatically by the existing `skills()` walk.

**Requirements:** R1, R4

**Dependencies:** Unit 2 (scripts) for the `--json` contract.

**Files:**
- Create: `pkg/scaffold/templates/skills/takeoff/SKILL.md`
- Create: `pkg/scaffold/templates/skills/land/SKILL.md`
- Create: `pkg/scaffold/templates/skills/solutions/SKILL.md`
- Create: `pkg/scaffold/templates/skills/pr-threads/SKILL.md`

**Approach:**
- Each SKILL.md ≤ ~30 lines.
- Frontmatter uses the established shape (`name`, `description`,
  `triggers`).
- Body: invoke `node scripts/<name>.mjs --json`, then narrate each JSON
  field in human language. Do not duplicate logic.
- `/land` SKILL.md explicitly instructs the narrator to stage / commit /
  push on the user's behalf AFTER surfacing the advisory, consistent
  with the global `land-the-plane.md` protocol. The narrator is where
  mutation happens; the script stays read-only.
- `/solutions` SKILL.md handles the two subcommands (`new`, `search`)
  with argument routing.
- `/pr-threads` SKILL.md flags the "needs_reply" PRs and suggests
  `/resolve-pr` as the follow-up.

**Patterns to follow:**
- Frontmatter reference: `.github/skills/ce-plan/SKILL.md`.

**Test scenarios:**
- Embedded-presence: the four SKILL.md files appear in `templateFS`.
- Integration: `BuildCatalog(stack)` via the existing `skills()` walk
  includes all four, targeting `.github/skills/<name>/SKILL.md`.
- Regression: other skills still show up (count check).

**Verification:**
- `go test ./pkg/scaffold/...` green.
- Manual read: each SKILL.md's triggers list includes the expected
  keyword variants (`takeoff`/`take off`, `land`/`land the plane`/
  `wrap it up`, etc.).

---

- [ ] **Unit 6: Copilot prompt templates + `copilotPrompts()` catalog wiring**

**Goal:** Author four `*.prompt.md` files and wire a new
`copilotPrompts()` catalog category.

**Requirements:** R1, R4, R7, R8

**Dependencies:** Unit 2 (shared `--json` contract).

**Files:**
- Create: `pkg/scaffold/templates/prompts/takeoff.prompt.md`
- Create: `pkg/scaffold/templates/prompts/land.prompt.md`
- Create: `pkg/scaffold/templates/prompts/solutions.prompt.md`
- Create: `pkg/scaffold/templates/prompts/pr-threads.prompt.md`
- Modify: `pkg/scaffold/catalog.go`
- Modify: `pkg/scaffold/hooks.go` (HookType=8 "Copilot Prompts" if
  centralized)
- Modify: `pkg/scaffold/catalog_test.go`
- Modify: `pkg/scaffold/catalog.go` `directories()` to include
  `.github/prompts/`.

**Approach:**
- Prompt frontmatter: `description`, `argument-hint` (optional), `agent`
  (optional — omit for task-style prompts).
- Body: same narration logic as the paired Claude skill, but written
  for the Copilot prompt picker surface.
- `copilotPrompts()` walks `templates/prompts/`, maps each to
  `Component{Path: filepath.Join(".github", "prompts", relPath),
  HookType: 8}`.
- Wired into both `BuildCatalog` and
  `BuildFilteredCatalog` (layer: `copilot-prompts`, default ON).

**Patterns to follow:**
- Walk mirrors `skills()` exactly.
- Frontmatter convention mirrors the documented Copilot prompt format
  (even though this repo has no prior `.github/prompts/` templates —
  new category).

**Test scenarios:**
- Happy path: `BuildCatalog(stack)` includes four
  `.github/prompts/*.prompt.md` components with `HookType: 8`.
- Edge case: `BuildFilteredCatalog(stack, [])` excludes them.
- Edge case: `BuildFilteredCatalog(stack, ["copilot-prompts"])`
  includes them.
- Integration: each prompt body references the paired script
  (`scripts/<name>.mjs`) — grep assertion in the installer test
  confirms the pairing so drift is caught early.

**Verification:**
- `go test ./pkg/scaffold/...` green.
- `go build ./...` green.

---

- [ ] **Unit 7: Install-recipe companion script + recipe catalog entries + `recipes()` wiring**

**Goal:** Author `install-recipe.mjs` and the six opt-in recipe files,
plus register a `recipes()` catalog category that is OFF by default in
BOTH one-click and guided modes — installer only scaffolds `recipes/`
into the target project when the guided layer is explicitly chosen;
activation of any individual recipe always requires the post-init
script.

**Requirements:** R2, R3, R7, R8

**Dependencies:** Unit 4 (for the hooks.go HookType pattern).

**Files:**
- Create: `pkg/scaffold/templates/scripts/install-recipe.mjs`
- Create: `pkg/scaffold/templates/recipes/husky/block-push-to-main.sh`
- Create: `pkg/scaffold/templates/recipes/husky/harness-leak-canary.sh`
- Create: `pkg/scaffold/templates/recipes/husky/update-backlog-on-fix.sh`
- Create: `pkg/scaffold/templates/recipes/claude/require-backlog-task.md`
- Create: `pkg/scaffold/templates/recipes/claude/link-plan-backlog.md`
- Create: `pkg/scaffold/templates/recipes/land/harness-exclusion.md`
- Modify: `pkg/scaffold/catalog.go` (new `recipes()` function; add
  `recipes` layer to `BuildFilteredCatalog` only — NOT `BuildCatalog`)
- Modify: `pkg/scaffold/hooks.go` (HookType=9 "Opt-In Recipes")
- Modify: `pkg/scaffold/catalog_test.go`
- Modify: `pkg/tui/wizard.go` (add `recipes` and `session-scripts`
  layers to the guided checkbox list; `recipes` starts UNCHECKED,
  `session-scripts` starts CHECKED)

**Approach:**
- `install-recipe.mjs` reads the recipe file's frontmatter manifest
  (`install_to`, `mode`, `action`), performs the action:
  - `copy` → byte-for-byte copy to `install_to`.
  - `merge-json` → read destination JSON, add-only merge, write back.
  - `append` → idempotent line-append.
  Supports `--dry-run`.
- Recipes themselves follow the shape documented in the original
  requirements:
  - `husky/block-push-to-main.sh` — pre-push hook rejecting `main`/
    `master`.
  - `husky/harness-leak-canary.sh` — pre-commit blocking staged paths
    matching `.harness-exclusions`.
  - `husky/update-backlog-on-fix.sh` — commit-msg hook verifying
    referenced Backlog task is `Done`; no-op when `backlog/` absent.
  - `claude/require-backlog-task.md` — doc recipe (action: `merge-json`
    into `.claude/settings.local.json`) re-enabling the
    `require-backlog-task` PreToolUse hook.
  - `claude/link-plan-backlog.md` — same pattern for the
    `link-plan-backlog` PostToolUse hook.
  - `land/harness-exclusion.md` — documentation-only; explains how to
    write a project-local `/land` override with harness-path filtering.
- **Default behavior:** `BuildCatalog(stack)` does NOT include
  `recipes()`. One-click `atv-installer init` never puts `recipes/` in
  the target project.
- **Guided behavior:** a user who ticks the `recipes` layer receives
  `recipes/` in their project, but NOTHING is activated. The only way
  any recipe becomes an active hook is running
  `node scripts/install-recipe.mjs <path>`.

**Patterns to follow:**
- Walk mirror of `skills()`.
- `hooks.go` HookType registration pattern.
- TUI layer pattern already present in `tui/wizard.go`.

**Test scenarios:**
- Happy path: `BuildCatalog(stack)` does NOT include any `recipes/*`
  component. (This is the core R2 guarantee.)
- Happy path: `BuildFilteredCatalog(stack, ["recipes"])` includes all
  six recipe files AND the `install-recipe.mjs` script (install-recipe
  always ships with session-scripts; recipes toggle only controls the
  catalog tree).
- Edge case: `BuildFilteredCatalog(stack,
  ["session-scripts", "recipes"])` includes both sets.
- Happy path: each `.sh` recipe passes `bash -n` (syntax check) — test
  shells out if `bash` is available; skips on Windows CI where shell
  absent.
- Integration: a manifest-presence test parses the frontmatter of each
  recipe file and asserts `install_to` + `action` fields are present
  and match an allowed action.

**Verification:**
- `go test ./pkg/scaffold/...` green.
- Running `./atv-installer init` on a temp dir produces NO `recipes/`
  directory.
- Running `./atv-installer init --guided` and ticking the `recipes`
  layer DOES produce `recipes/` with the six files and
  `scripts/install-recipe.mjs`.

---

- [ ] **Unit 8: Next-steps output + README update**

**Goal:** Surface the four new skills in the post-install "next steps"
message so scaffolded-project users know they exist. Update this repo's
`README.md` to advertise the new catalog additions.

**Requirements:** R9

**Dependencies:** Units 4, 5, 6, 7.

**Files:**
- Modify: `pkg/output/printer.go` (`PrintNextSteps`)
- Modify: `README.md`
- Create: `docs/session-skills.md` (short reference)

**Approach:**
- Append a "Session skills" block to `PrintNextSteps` that lists the
  four slash commands plus their one-line purpose. Phrased to work for
  both Claude Code users ("type `/takeoff`") and Copilot users
  ("open the prompt picker → `takeoff`").
- `docs/session-skills.md` is a short reference page describing each
  skill, the paired script, and the `--json` contract.
- `README.md` gains a short bullet under "What gets installed" pointing
  at the new categories.

**Test scenarios:**
- Integration: `go test ./pkg/output/...` (or a new test) captures
  stdout of `PrintNextSteps(stack.TypeScript)` and asserts the four
  skill names appear.
- Regression: existing next-steps content still present.

**Test expectation:** behavioral — see above.

**Verification:**
- `go test ./...` green.
- `./atv-installer init` on a temp dir prints the new session-skills
  hint at the end.

## System-Wide Impact

- **Interaction graph:** `BuildCatalog` gains two new category functions
  (`sessionScripts`, `copilotPrompts`); `BuildFilteredCatalog` gains
  three new layer names (`session-scripts`, `copilot-prompts`,
  `recipes`). All four skill wrappers (Claude + Copilot) consume the
  same per-script `--json` contract; a change to any script's output
  schema must update both wrappers.
- **Error propagation:** Scripts use the documented exit-code contract
  (`0` normal, `1` expected failure signal, `2` unexpected error).
  Narrators treat `2` as a hard failure surface.
- **State lifecycle risks:** `/land` is advisory — it never mutates git
  state. The narrator is where real mutation happens. This preserves
  safety across both Claude and Copilot surfaces.
- **API surface parity:** Every new skill has both a Claude entry point
  (`.claude/skills/<name>/SKILL.md`) and a Copilot entry point
  (`.github/prompts/<name>.prompt.md`). Future additions must maintain
  parity.
- **Integration coverage:** Installer-side tests confirm catalog
  composition; script-side tests (scaffolded templates) confirm
  behavior in the scaffolded project — two separate test surfaces, no
  overlap expected.
- **Unchanged invariants:**
  - Existing Hook categories 1–6 unchanged.
  - `atv-installer init` CLI surface unchanged (no new flags).
  - `scaffold.WriteAll` idempotency semantics unchanged.
  - Stack detection unchanged.
  - Existing agent list unchanged.
  - No husky hook, Claude settings hook, or hookify rule is installed
    by default. R2 is the headline invariant.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Recipes accidentally included in one-click mode | `BuildCatalog` explicitly does NOT call `recipes()`; regression test asserts no `recipes/*` paths in default catalog |
| Claude and Copilot narrators drift out of sync | Shared `--json` contract; grep assertion in Unit 6 tests verifies each prompt references its paired script |
| Script behavior tests live in scaffolded projects, not this repo, making install-time regressions invisible | Embedded-presence tests in this repo at least ensure the files ship; a follow-up "smoke test harness" that runs the scripts against a fixture repo can land as a separate plan if drift becomes a problem |
| `gh` CLI absence breaks `/pr-threads` in CI | Script exits `2` with clear guidance; `--offline --fixture` flag supports deterministic testing |
| Shell recipe portability (macOS vs Linux, and Windows) | Recipes use POSIX sh only; `bash -n` test runs where available; Windows documented as "install via WSL" in recipe README |
| Adopter re-runs `install-recipe.mjs` and clobbers customizations | `merge-json` is strictly add-only; `--dry-run` surfaces planned writes first; idempotent byte-for-byte skip on `copy` |
| New catalog category numbers collide with future hook types | Register in a central `hooks.go` constant block; tests assert uniqueness |

## Documentation / Operational Notes

- `docs/session-skills.md` is the canonical user-facing reference for
  what the four skills do and how to invoke them under each agent.
- `README.md` "What gets installed" section names the new categories.
- No rollout flag — these are template additions, not runtime
  features.
- No monitoring.
- Release notes (via goreleaser) should mention the new scaffolded
  skills so existing adopters know to re-run `atv-installer init` to
  pick them up.

## Sources & References

- **Origin document:** [docs/brainstorms/2026-04-21-cross-tool-harness-starterkit-requirements.md](../brainstorms/2026-04-21-cross-tool-harness-starterkit-requirements.md)
- Related code:
  - `pkg/scaffold/catalog.go` (existing `skills()` walk pattern)
  - `pkg/scaffold/scaffold.go` (`WriteAll` idempotency)
  - `pkg/scaffold/hooks.go` (HookType registration)
  - `pkg/output/printer.go` (`PrintNextSteps`)
  - `pkg/tui/wizard.go` (guided-mode layer list)
  - `.github/skills/ce-plan/SKILL.md` (Claude frontmatter reference)
- Related docs:
  - `docs/plans/2026-03-11-002-feat-atv-starter-kit-guided-installer-plan.md` (installer architecture baseline)
  - `docs/brainstorms/2026-03-11-agentic-coding-starter-kit-installer-brainstorm.md` (installer UX philosophy)
