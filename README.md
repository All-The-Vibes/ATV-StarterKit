<p align="center">
       <img src="./assets/hero-retro.svg" alt="ATV — All The Vibes 2.0 Starter Kit" width="100%" />
</p>

---

> **Fork of [All-The-Vibes/ATV-StarterKit](https://github.com/All-The-Vibes/ATV-StarterKit)** — built on ATV's learning system, 45+ skills, 51 agents. Adds enforcement-first execution.

<h1 align="center">The KB Pipeline</h1>

<p align="center"><strong>Research, slice, build, review, and learn with voice-friendly <code>kb-</code> commands.</strong></p>

<p align="center">
       <code>/klfg "your feature"</code> — /kb-brainstorm → /kb-plan → /kb-work → /kb-complete.
</p>

---

## What It Does

KB means **Kanban-Based**: the workflow still uses vertical slices, a shared board, and manifest files, but every user-facing workflow command uses the voice-friendly `kb-` prefix so you do not have to say "kanban".

Most KB skills are augmentations on top of the ATV StarterKit and CE
review/learning workflow. KB adds the voice-friendly routing, project-memory
map, fresh-session handoff loop, proportional planning, and execution gates; it
still depends on selected ATV skills and reviewer agents.

## Token-Minimizing Design

The core purpose of the KB skill set is to reduce wasted context without
lowering the engineering bar.

- Fresh sessions are expected. Handoffs, `todo.md`,
  `docs/context/PROJECT.md`, plans, and architecture notes let a new session
  recover the project instead of carrying days of chat history.
- `kb-map` builds or refreshes project memory once, then future sessions load
  exact pointers instead of crawling the repo or making the user reteach the
  app.
- `kb-start` chooses the smallest correct lane: small fix, brainstorm, plan,
  work, complete, ship, or epic. It should not turn a small fix into a large
  ceremony.
- Vertical slicing and functional verification cost tokens up front, but they
  are cheaper than redoing broken or under-tested work later. The target is the
  fewest wasted tokens per finished, verified change.

## Work-Time Model Selection

`kb-plan` records slice difficulty, risk, tools, context, and proof without
freezing a provider/model. `kb-work` discovers the active host's selectable
catalog, merges optional user-local routes, and chooses a bounded worker right
before execution. Fallback is same-tier, then higher-tier, then the current
model only as an explicit degraded path; routing receipts never replace normal
tests and proof. Private endpoints and credentials stay out of shared skills.

![KB model selection workflow](docs/assets/kb-model-selection.png)

## Fresh Session Loop

The KB workflow is meant to make every new task safe to start in a fresh
session:

1. Finish or pause the current task with a handoff.
2. Close the old session.
3. Start a new session in the project repo.
4. Run `kb-start <next task or handoff>`.

`kb-start` calls `kb-map`, which reads local project memory and points the new
session to the specific files it needs. The handoff tells the model what work is
being resumed; `docs/context/PROJECT.md` tells it what the app is and where the
relevant architecture docs live.

## 2026-05-23 KB Workflow Split

The voice-friendly KB workflow now has a smaller standalone home:

**[Irtechie/working-skill-repo](https://github.com/Irtechie/working-skill-repo)**

Use that repo when you want the current working KB skill bundle installed into
GitHub Copilot or Codex. The preferred install is now personal/global
(`~/.copilot/skills`, `~/.copilot/agents`, `~/.agents/skills`, and
`~/.codex/skills`) instead of copying the bundle into every project. Repo-local
installs are still useful for pinned or project-specific overrides. This ATV
fork keeps the broader ATV StarterKit, CE skills, agents, plugin experiments,
historical docs, and upstream lineage. The new repo is the trimmed day-to-day
bundle.

What changed:

- `kanban-*` user-facing workflows became `kb-*`.
- `kb-start` is now the default entry point for ambiguous work.
- `kb-start` replaces the older `kb-route` name; the workflow still maps context
  first, then chooses the right lane.
- `kb-start` now delegates project-memory setup to `kb-map`; route chooses the
  lane for the idea/request, while map decides lookup, refresh, or bootstrap.
- `kb-start` now has a startup-only session hygiene check. It recommends
  handoff/restart only when context pressure exists and durable local memory can
  replace the live chat at lower total context cost or lower drift risk. It does
  not interrupt active work just because a session is long.
- `kb-goal` is the durable objective governor for work that may run for days
  across sessions. It keeps the goal ledger, terminal proof, blockers, and next
  action while routing each unit through normal KB lanes. `klfg` remains one
  strict pipeline run; `kb-goal` may run many pipelines or smaller lanes before
  the larger goal can be called complete. Under a goal, brainstorming is
  low-interruption: the agent picks the best path from evidence and asks only
  for true planning blockers.
- `kb-map` is now project-root anchored: it reads memory from the active repo,
  not `~`, `.copilot/handoffs`, the whole drive, or sibling repos.
- `kb-map` checks standard memory files by exact path under the repo root, not
  by broad grep/glob.
- Drive roots such as `E:\` are not valid project roots unless explicitly
  chosen; `kb-map` should ask for the project path instead of searching the
  drive.
- `kb-map-bootstrap` and `kb-map` create/update project memory so new sessions
  can recover context without a long chat history.
- `kb-eval-map` is now part of bootstrap-owned setup. It detects the target
  repo's native eval surface, writes `docs/context/eval-map.md`, and scaffolds
  one real smoke eval only when the primary workflow is known and safe to run.
  It maps/scaffolds proof; it is not the full live skill-eval suite.
- `todo.md` and `todo-done.md` replace `docs/kanban.md` and
  `docs/kanban-done.md` for the current KB workflow.
- `todo.md` now carries its own rules at the top. Completed feature, slice,
  handoff, and fix summaries move to `todo-done.md`; routine completion logs
  should not stay in `todo.md`. Do not create or depend on `todo_rules.md`,
  `todo-rules.md`, or any separate rules file.
- Board row markers are part of the inline `todo.md` contract: `⬜ pending`,
  `🔧 in_progress`, `✅ done`, `🔒 blocked`, `⊘ skipped`, and
  `🛑 human-required`. Section icons are also standardized: `💡 Feature Ideas`,
  `📋 Queued Improvements`, `🧊 Parked / Cold Storage`, `🛑 Human Required`,
  and `📝 Work Log`. `🔒 blocked` is for dependency, tool, or another-agent
  waits that can resume when the blocker clears. `🧊 Parked / Cold Storage` is
  intentionally out of bounds today and only a human promotes it back to active.
- `kb-task`, `kb-fix`, `kb-troubleshoot`, `kb-handoff`, `kb-functional-test`,
  `kb-regression-snapshot`, `kb-gate`, `kb-check`, `kb-research`, `kb-epic`,
  `kb-compact`, and `kb-ship` were added to cover first-principles autonomous
  task execution, small fixes, autonomous troubleshooting, repo-local restart
  packets, deterministic testing, P0-P4 gates, research, large initiatives, token trimming, and
  release readiness.
- `kb-functional-test` owns test-level classification for slices. Plans record
  `test_level` (`none`, `unit`, `integration`, `functional-api`,
  `functional-cli`, `functional-browser`, or `full`) and `functional_risk`
  (`none`, `narrow`, `broad`, or `full`). Small/mini models may classify or
  audit test quality when available, but executable checks remain the proof.
- UI-reachable work is tested through the rendered UI. `.tsx`, `.jsx`, `.vue`,
  and `.svelte` changes auto-classify as `functional-browser`; backend/API/unit
  checks can support that proof but cannot replace real navigation, clicks or
  inputs, rendered assertions, screenshots, and cleanup.
- `kb-qa` must convert visible acceptance criteria into executable browser
  assertions or the project stack equivalent. Screenshots support the result,
  but they are not the pass/fail oracle.
- Generated commands and assertions must avoid nested-quote traps. If shell
  commands, file operations, JSON, SQL, HTML, config blocks, or Playwright
  selectors require quotes inside quotes or escaped escapes, write the content
  to a temp file, heredoc, template literal, or parameterized locator helper
  instead of constructing it inline.
- `kb-regression-snapshot` captures deterministic state after each passed slice
  in `.atv/snapshots/<slice-id>.json` and verifies prior snapshots before the
  next slice starts. The LLM writes the compact snapshot spec; the bundled
  runner verifies DOM/API/CLI/file checks mechanically. This keeps old slice
  behavior machine-checkable across long runs and fresh sessions.
- `kb-brainstorm` now proceeds to `kb-plan` when the requirements artifact is
  gate-clean. It pauses only for unresolved blockers, required human decisions,
  required research, or an explicit user stop.
- "Don't ask many questions", "go straight to work", and similar phrasing is
  execution intent, not permission to skip slices. The workflow still goes
  requirements/assumptions -> `kb-plan` -> `kb-work` -> `kb-complete`.
- `kb-brainstorm` multiple-choice questions now always include an escape hatch
  such as `Other / let me explain` or `None of these`. If the answer may need an
  image, screenshot, file, pasted output, diagram, or longer explanation, the
  skill should ask in normal chat instead of the blocking question UI.
- `kb-work` owns slice execution from a valid KB manifest and calls
  `kb-complete` only after all slices are done or intentionally skipped. Raw
  brainstorm notes, phase lists, and free-form feature asks route to `kb-plan`
  first. A slice's `expected_files` are a forecast, not a hard allowlist; files
  discovered during implementation are allowed when required by the slice and
  recorded in the scope ledger. Material slices may also carry a context packet
  with exact files, deterministic prefetch, constraints, proof targets, search
  policy, and escalation triggers so bounded workers do not rediscover the repo.
- `kb-finish` is the explicit plan-to-PR lane: it recovers missing
  plan/work/complete phases, then `kb-ship` commits intentional files, pushes a
  topic branch, and creates or updates a PR without merging. Ordinary
  `kb-complete` calls remain non-shipping.
- `kb-complete` now records memory-maintenance signals in
  `docs/context/memory-maintenance.md`: contradictions, overlaps, stale docs,
  bloat, repeated rediscovery, durable refreshes, and closed handoffs. It stores
  pointers and the actual issue so later deep memory review can be targeted
  instead of a blind full scan.
- `kb-memory-review` is the explicit high-cost pass that consumes those signals,
  reconciles/compacts/consolidates targeted memory docs, invokes narrower helper
  skills when useful, and updates the maintenance index. It is recommended by
  thresholds but does not run automatically.
- Once `kb-work` starts execution, runnable slices continue without per-slice
  confirmation. It pauses only for HITL, blocked/manual work, destructive
  approval, scope failures, QA/repair exhaustion, dependency deadlock, or an
  explicit user stop.
- Active handoffs no longer jump straight to `kb-work` unless they link a valid
  KB manifest. Phase-shaped handoffs route through `kb-plan` first.
- Before planning from a handoff, `kb-plan` checks for existing brainstorm,
  requirements, manifest, or plan files and uses the best existing source of
  truth instead of duplicating work.
- `klfg` remains the full hands-off orchestrator for one brainstorm -> plan ->
  work -> complete pipeline run.
- The day-to-day working bundle still carries required ATV/CE dependencies:
  `document-review`, `kb-review`, `ce-review`, `ce-compound`,
  `ce-compound-refresh`, `learn`, `evolve`, `tdd`, `todo-create`,
  `todo-triage`, and the reviewer/specialist agents in
  `.github/agents/*.agent.md`.
- Testing showed the agents are required runtime dependencies, not optional
  docs. In particular, `document-review` needs its document personas
  (`coherence-reviewer`, `feasibility-reviewer`, `product-lens-reviewer`,
  `design-lens-reviewer`, `security-lens-reviewer`, `scope-guardian-reviewer`,
  `adversarial-document-reviewer`), and `kb-review`/`ce-review` need their
  code-review personas (`correctness-reviewer`, `testing-reviewer`,
  `thermo-nuclear-code-quality-reviewer`, security/performance/API reviewers,
  language reviewers, schema/deployment reviewers, and learning agents).
- Heavy inherited ATV/CE skills now use a token diet in the working bundle:
  `ce-review` and `ce-compound-refresh` keep routing, gates, and safety rules in
  `SKILL.md`, while detailed phase mechanics live in lazy `references/` files.
  The goal is lower startup load without losing review or learning behavior.
- `kb-review` is the KB-specific review orchestrator. It keeps the CE review
  pipeline shape but replaces the standard maintainability persona with
  `thermo-nuclear-code-quality-reviewer`.
- The 2026-05-24 token-diet pass reduced `ce-review` to 235 lines and
  `ce-compound-refresh` to 289 lines in the working bundle by moving execution
  details into lazy references, not by deleting behavior.
- The working bundle should stay portable: skills, agents, scripts, templates,
  and durable references only. Project-generated brainstorms, plans, research,
  handoffs, and context maps belong in the project that created them or in this
  broader starter-kit history, not in the trimmed global skill repo.
- KB skill changes are authored in `E:\working-skill-repo`, compared against
  any global or ATV drift before overwrite, then synced to Codex, Copilot,
  shared agents, this ATV fork, and the scaffold/plugin copies that ship that
  skill. Update both READMEs when the visible workflow or shipped skill surface
  changes.
- The working bundle now separates contributor checks from release/sync checks:
  `go run ./cmd/kbcheck core` stays repo-local and fresh-clone safe, while
  `go run ./cmd/kbcheck local-release` and `go run ./cmd/kbcheck
  skill-sync-report` own global/ATV propagation drift. Optional ATV scaffold and
  plugin skill drift is warning-only unless that surface is intentionally being
  shipped.
- `learn` is intentionally shipped in the ATV GitHub, scaffold, and plugin skill
  surfaces. Observer hooks such as `.github/hooks/copilot-hooks.json` remain an
  ATV integration layer; the portable KB skill bundle treats them as optional,
  not as files guaranteed by the skill itself.
- Blocking question pickers are used only when the answer is truly one short
  choice. For voice dictation, paste, images, screenshots, files, or long
  corrections, skills should ask in normal chat or include `Other / let me
  explain` and return to chat. This is part of the token-minimizing design:
  one good dictated answer is cheaper than several picker turns.

### Skill Runtime Surface

The KB bundle ships the full `.github/agents/*.agent.md` surface for now because
real testing showed missing ATV agents break `document-review`, `kb-review`,
and `ce-review` dispatch. Treat agents in three tiers:

- **Required dispatch agents** are called by skills and must stay installed.
- **Conditional specialists** are used only when the diff/task warrants their
  lens.
- **Optional direct-use agents** can be trimmed later only after benchmark runs
  prove no workflow depends on them.

### Skill Benchmarking

Line count is only a proxy. A better benchmark runs fixed prompts in a scratch
repo and records whether `kb-start` chooses the right lane, `kb-map` loads only
repo-local memory, slices include `expected_files` and verification, runnable
slices continue without unnecessary user prompts, review agents dispatch, tests
run deterministically, and completed work lands in the right lifecycle files.

Shorter skills win only when those behaviors stay intact.

### Why KB Start Exists

`kb-start` is the workflow router. Its job is to choose the right lane for the
actual work, not blindly run the ceremony implied by the user's wording.

Every request starts by calling `kb-map lookup <request>` so the session has the
current project memory before it decides what to do. Then `kb-start` classifies
the work by task shape, risk, and available artifacts:

- Use `kb-fix` for small, bounded bugs or narrow changes where the likely fix is
  obvious and verification can prove it.
- Use `kb-troubleshoot` when broken behavior needs evidence gathering and
  self-correction. It must inspect local logs/tests/browser behavior and, when
  framework/tool/dependency behavior may matter, research current external docs,
  issues, changelogs, or known fixes before editing.
- Use `kb-brainstorm` when product behavior, technical framing, success
  criteria, or tradeoffs are still unclear.
- Use `kb-plan` when requirements or a handoff already explain the work and the
  next useful output is vertical slices.
- Use `kb-work` when a valid manifest and slice plans already exist.
- Use `kb-epic` when the request is too large for one brainstorm or manifest:
  migrations, framework rewrites, multi-subsystem initiatives, or long backlogs.
- Use `kb-research` only when external docs, prior art, framework behavior, or
  known failure modes could change the decision.

The goal is proportional ceremony. A typo fix should not become a brainstorm. A
framework migration should not become a quick fix. A clear handoff should not
rerun discovery just because the user said "brainstorm" casually. The user's
words are input; the route should be based on the actual task, the repo memory,
and the cost of being wrong.

### Why KB Map Exists

`kb-map` is the context router for fresh sessions. The workflow assumes sessions
are disposable: instead of keeping one expensive chat open for days, a new
session should enter a repo, resolve the active project root, and load only the
local memory needed for the current handoff, bug, feature, or plan.

It points the model to `todo.md`, `docs/context/PROJECT.md`, the relevant
subsystem architecture docs, operations/testing notes, and active handoffs. It
does not crawl the whole drive, search unrelated repos, or load every
architecture file by default. The goal is scoped orientation: get the model to
the project truth that matters now so tokens are spent on execution instead of
rediscovery.

`docs/context/PROJECT.md` is the entry map. It explains what the app is, how to
run and test it, what major subsystems exist, and which subsystem documents to
read next. `docs/context/architecture/*.md` files are the deeper subsystem
notes. `kb-map` should read `PROJECT.md` first, then follow its pointers to the
smallest relevant architecture file for the current task.

Coverage matters. If a fresh session asks about a named high-risk workflow such
as installer, release, auth, workflows, actions, tools, runtime, or deployment,
`kb-map` must point to the exact subsystem doc and source-of-truth files without
broad rediscovery. If it cannot, targeted refresh should create or refine the
missing child architecture doc and record a memory-maintenance signal.

Bootstrap owns the first coverage pass: inventory the repo, reconcile discovered
systems against `PROJECT.md` and `docs/context/architecture/README.md`, and
route-test every mapped major area. One invisible subsystem is evidence to run a
coverage audit, not to keep fixing one doc at a time.

Bootstrap must also validate chains, not just describe files. For high-risk
systems like installers, releases, auth, data, integrations, and embedded
runtimes, it should connect build config, shipped artifacts, install locations,
first-launch downloads, runtime consumers, version pins, architecture-specific
paths, and smoke tests. A subsystem doc is not good enough if a smaller fresh
session still has to rediscover what must exist on disk or what gets used at
runtime.

Bootstrap must discover concepts, not just folders. It descends into substantial
child directories, clusters cross-cutting concerns, mines repo memories and
AGENTS/README files for subsystem hints, checks route/page and filename-prefix
patterns, and records known-unknowns. `kb-map` also warns when lookup sees a
thin map compared with the actual repo shape.

When memory is missing, `kb-map` invokes `kb-map-bootstrap` to build the project
map once. After that, normal startup is cheap: `kb-start` calls `kb-map lookup
<request>`, `kb-map` returns the relevant docs and likely route, and the next
skill can work without the user reteaching the app.

This repo can still carry the full ATV and CE ecosystem. The point of the split
is discoverability: active projects should find the smaller KB bundle first,
while this repo remains the larger starter kit and historical source.

You can run the stages directly, or let `/klfg` orchestrate them. The pipeline:

1. **Researches** the landscape before asking you product questions (not after)
2. **Decomposes** your feature into vertical slices — each one cuts through all layers end-to-end
3. **Requires each slice to declare which files it will touch** — before execution starts
4. **Executes** each slice through 7 mandatory safety gates
5. **Reviews** the full diff with multi-agent code review (scope already verified — no redundant discovery)
6. **Documents** patterns worth remembering, extracts instincts, promotes mature ones to skills
7. **Cleans up** after itself

You're interactive during brainstorm Q&A and when safety gates fire. Everything else is autonomous.

---

## The Pipeline

```
/klfg "your feature"
       │
       ▼
 ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌──────────────┐     ┌──────┐
 │  BRAINSTORM  │────▶│    PLAN     │────▶│    WORK     │────▶│   COMPLETE   │────▶│ SHIP │
 │              │     │             │     │             │     │              │     │      │
 │ research     │     │ vertical    │     │ per-slice   │     │ kb-review    │     │/land │
 │ then ask     │     │ slices +    │     │ execution   │     │ compound     │     │      │
 │ questions    │     │ expected_   │     │ through 7   │     │ learn        │     │      │
 │              │     │ files       │     │ hard gates  │     │ evolve       │     │      │
 └─────────────┘     └─────────────┘     └─────────────┘     └──────────────┘     └──────┘
       │                    │                   │                    │
  requirements.md      manifest +          tested code         reviewed &
                     slice plans         + atomic commits      documented
```

---

## The Full Pipeline, Step by Step

### Step 1: Brainstorm (`kb-brainstorm`)

Research runs **before** questions so decisions are grounded in real prior art — not the user's first framing.

| Phase | What happens |
|-------|-------------|
| Topic intake | Restate the feature, confirm understanding. No questions yet. |
| Repo context | Search codebase for related patterns, existing features, constraints. |
| External research | Parallel agents: market landscape, prior art from `docs/solutions/`, applicable skills, risk & failure modes. |
| Research brief | Synthesize findings, show the user before any questions. |
| Product pressure test | Challenge the request: Right problem? Better framing? Do nothing? |
| Targeted Q&A | Ask sharp product questions informed by research. One at a time. |
| Approaches | Propose 2–3 approaches with tradeoffs. User picks. |
| Capture | Write requirements to `docs/brainstorms/*-requirements.md`. |
| Document review | Multi-persona review (PM, engineer, security). |
| Handoff | Proceed to planning. |

### Step 2: Plan (`kb-plan`)

Break the brainstorm into independently-executable vertical slices — not horizontal phases.

- ❌ WRONG: "Create schema" → "Build service" → "Add API" → "Build UI"
- ✅ RIGHT: "Award points + show on dashboard" → "Track streaks" → "Add level display"

Every slice declares:
- **`expected_files`** — which files it will create or modify, with `op` and `scope`
- **`verification`** — `tdd`, `integration`, `verification-only`, or `hitl`
- **`blockers`** — dependency DAG between slices

```yaml
expected_files:
  - path: src/services/streaks.py
    op: create
  - path: src/models/user.py
    op: edit
    scope: "add current_streak and longest_streak fields"
  - path: tests/test_streaks.py
    op: create
```

This isn't documentation — it's a machine-enforced contract. The pipeline checks it before AND after execution.

### Step 3: Work (`kb-work`)

Executes all slices in dependency order. Resumable — re-running picks up where it left off.

`kb-work` requires a KB manifest and per-slice plans. A handoff with phases,
workstreams, bullets, or broad next steps is not executable yet; route it through
`kb-plan` first so the work becomes vertical slices with `expected_files`,
verification, blockers, and status.

**Per-slice, every slice, no exceptions:**

| Gate | Type | What it does |
|------|------|-------------|
| **3.0 Scope Lock** | Proactive | Load `expected_files`. Block writes to any undeclared file. No `expected_files` = cannot start. Convention-matched test files auto-allowed (`src/foo.py` permits `tests/test_foo.py`). |
| **3 Execute** | — | Implement the slice. TDD, integration, or verification-only per the plan. |
| **3.5 System Tests** | Analytical | What fires when this runs? Callbacks, middleware, observers 2 levels out. |
| **3.6 Diff-Scope** | Reactive | `git diff --name-only` vs declared `expected_files`. Out-of-scope files = **STOP**. Missing expected files = flag incomplete. |
| **3.7 Destructive Guard** | Preventive | Block `rm -rf`, `git push --force`, `DROP TABLE`, etc. Requires human approval. Cannot be overridden programmatically. |
| **3.8 QA** | Hard gate | Lint on all slices. Browser verification on frontend or UI-reachable slices via Playwright, CDP, or agent-browser. **Slice cannot advance until all checks pass.** On failure → `kb-repair` autonomous fix loop. |
| **3.9 Figma Sync** | Visual | Compare rendered UI to Figma designs (UI slices only). |

**Why the scope gates matter:** An agent reporting "I only modified `src/foo.py`" is generating that statement from its context window — same source as everything else, same hallucination probability. `git diff --name-only` has zero hallucination probability. The scope lock prevents out-of-scope writes before they happen. The diff-scope gate catches anything that slipped through after.

**Gate 3.8 is a hard acceptance gate, not advisory.** The slice stays `in_progress` until every lint check and every browser check passes. The agent cannot mark it done, cannot move to the next slice, and cannot proceed to Step 4 while any check is failing.

**Automated browser testing — what runs:**

`kb-qa` picks the best available transport for the environment:

| Transport | When used |
|-----------|-----------|
| **CDP** (Chrome DevTools Protocol) | Internal/corporate sites — connects to your existing authenticated browser session. Required for SSO/Conditional Access. |
| **Playwright** | Local dev servers and public URLs — headless, clean viewport, best for responsive testing at 375px/768px/1440px. |
| **agent-browser** | Structured element targeting via snapshot refs (`@e1`, `@e2`) — ~100ms latency, no CSS selectors needed. |

For each changed frontend file or UI-reachable behavior change, it maps the change to the affected page, navigates there, exercises the real UI controls, screenshots key states, checks the console, and verifies rendered acceptance criteria from the slice plan. Every click and form fill gets a before/after console snapshot. Backend/API calls, component handler invocation, mocked requests, and state inspection are supporting evidence only.

**When QA fails — the autonomous repair loop:**

`kb-repair` runs immediately, without losing context (same agent, no handoff):
- Each fix is an **atomic commit** — one commit = one revert if it causes regression
- **Progress-based:** fewer failures = continue. Same failures = stuck, stop.
- **Stuck detection:** same failure twice, fix reverted twice, 3+ files touched in one fix, same file edited→reverted→re-edited
- **5-iteration hard ceiling**, no exceptions — prevents infinite loops on flaky rendering or cascading lint
- On exhaustion: slice stays `in_progress`, agent STOPS, user decides

After each fix, kb-repair re-runs the full QA check — not just the failing check. A fix for one failure can introduce another; the loop catches it immediately.

This is not "retry 3 times and give up." The agent keeps working autonomously until tests pass — or until it hits a wall it can't climb, at which point it stops and hands the problem to you with screenshots and a full failure log.

**Board sync:** `todo.md` is the live multi-agent board. Agents claim slices before working and release after completing. Completed summaries move to `todo-done.md` so `todo.md` stays small and current.

### Step 4: Complete (`kb-complete`)

After all slices pass, the quality and learning pipeline runs automatically:

| Step | What happens |
|------|-------------|
| **Code review** | `kb-review` with scope-verified file list pre-loaded. Multiple persona agents including thermonuclear structural quality, security, performance, and correctness. |
| **Resolution gate** | Safe/actionable P0-P4 findings are fixed by the agent. Human input is required only for product intent, access, risky operations, competing reasonable paths, or genuine ambiguity. |
| **Follow-up resolution** | Review/TODO fallout is resolved or explicitly logged before completion. Parallel resolution is allowed only when file scopes are disjoint. |
| **Proof/demo evidence** | Final checks rerun after review fixes. Browser, CLI, API, desktop, service, or snapshot proof is captured with available repo/platform tools. Every slice needs machine-verifiable evidence in the manifest: command/test path, exit code, timestamp, trace/log/API artifact, or snapshot result. Prose-only proof fails completion. |
| **Compound** | `ce-compound` documents surprising patterns to `docs/solutions/`. Skips boilerplate. |
| **Learn** | `/learn` extracts instincts from resolved findings + recent work. |
| **Evolve** | Every 5th completion, `/evolve` checks for instincts ready to become full skills. |
| **Memory refresh + compact** | `kb-map refresh` updates durable project memory when behavior changed. `kb-compact` trims bloat that hurts fresh-session startup. |
| **Cleanup + alerts** | Prune QA screenshots, trim observations log to 90 days, and alert on unresolved memory/review/tooling issues with evidence paths. |

### Step 5: Ship

Run `/land` when you're ready to push and open a PR. Shipping is a separate, deliberate act — not buried in the pipeline.

---

## What's New vs. What Existed

This fork doesn't replace the upstream tools — it adds an execution engine with enforcement. Here's what came from where:

| Capability | Existed In | What This Fork Changed |
|------------|-----------|----------------------|
| **QA (lint + browser)** | gstack `/qa` | Moved from post-batch to **per-slice**. Failures caught before they compound. |
| **Code review** | ATV `ce-review` | KB now calls `kb-review`, a KB-specific fork that keeps the orchestrator and swaps in the thermonuclear structural-quality reviewer. General `ce-review` remains available. |
| **Learning pipeline** | ATV observations → instincts → evolve | Unchanged — runs automatically after review findings are resolved. |
| **Compound docs** | ATV `ce-compound` | Unchanged — fed by per-slice micro-learnings instead of just the final diff. |
| **Destructive guards** | gstack `/careful` | Changed from **overridable warning** to **hard block**. Cannot be bypassed. |
| **Vertical slices** | Pocock `/to-issues` | Added `expected_files` contract (advisory → enforceable), DAG execution order. |
| **Browser automation** | Vercel `agent-browser` | Added diff-aware page scoping and continuous console capture. |

**Genuinely new in this fork (not from upstream):**
- `expected_files` scope contract — slices declare files during planning, enforced before and after execution
- Scope Lock (Step 3.0) — proactive write blocking
- Diff-Scope Verification (Step 3.6) — reactive git-diff verification
- `kb-repair` — progress-based autonomous fix loop with stuck detection
- Board sync protocol — multi-agent mutex via `todo.md`, with completed work archived to `todo-done.md`
- Convention-matched test auto-allow — `src/foo.py` automatically permits `tests/test_foo.py`

---

## Skills Reference

| Skill | Role |
|-------|------|
| `/klfg` | Full KB orchestrator — `/kb-brainstorm` → `/kb-plan` → `/kb-work` → `/kb-complete` |
| `/kb-task` | First-principles task runner that chooses the KB route and continues until verified or blocked |
| `/kb-troubleshoot` | Autonomous debug loop: inspect logs/browser/tests, research uncertain assumptions, fix, and verify |
| `/kb-brainstorm` | Research-first requirements gathering; auto-starts planning when gate-clean |
| `/kb-plan` | Vertical-slice decomposition with `expected_files` contracts |
| `/kb-work` | Execute slices through 7 mandatory gates |
| `/kb-complete` | Post-work: kb-review → compound → learn → evolve → cleanup |
| `/kb-qa` | Lint + browser verification (called by kb-work) |
| `/kb-repair` | Surgical fix loop (called by kb-qa on failure) |

---

## Credits

| Project | What We Built On |
|---------|-----------------|
| **[ATV StarterKit](https://github.com/All-The-Vibes/ATV-StarterKit)** | The entire foundation: learning system, ce-review, ce-compound, 45+ skills, 51 agents, observer hooks |
| **[gstack](https://github.com/garrytan/gstack)** (Garry Tan / YC) | QA philosophy, continuous console monitoring, atomic commit pattern, stuck detection |
| **[Matt Pocock](https://github.com/mattpocock/skills)** | Vertical-slice-as-primitive, TDD anti-pattern identification, hard-gate pattern (`git-guardrails`), [`/handoff`](https://github.com/mattpocock/skills/blob/main/skills/productivity/handoff/SKILL.md) session handoff, workflow skill patterns. The KB pipeline is a Kanban-Based synthesis of Matt's skills + ATV StarterKit foundations |
| **[Shyam Sridhar's kevin-copilot](https://github.com/shyamsridhar123/kevin-copilot)** | Copilot-first token-saver / terse-response instruction surface |
| **[Shyam Sridhar's TokenMasterX](https://github.com/shyamsridhar123/TokenMasterX)** | Graph/token-aware repo orientation ideas that informed the graphify/TokenMasterX map-bootstrap path |
| **[agent-browser](https://github.com/vercel-labs/agent-browser)** (Vercel Labs) | Native Rust CDP automation, snapshot refs, ~100ms latency |
| **[Compound Engineering](https://github.com/EveryInc/compound-engineering-plugin)** (Every, Inc.) | Knowledge-compounds-over-time: plan → work → review → document → learn |
| **[Karpathy](https://x.com/karpathy/status/2015883857489522876)** | "Models make wrong assumptions and run with them" — the observation that motivated structural enforcement |

📖 **[Full technical documentation →](docs/KANBAN-SKILLS.md)**

---

<h1 align="center">ATV — All The Vibes 2.0 Starter Kit</h1>

<p align="center"><strong>One command. Full agentic coding setup. Maximum tasteful chaos.</strong></p>

<p align="center">
       <a href="https://blazingbeard.github.io/quests/atv-starterkit.html"><img src="https://img.shields.io/badge/🎮_New%3F_Start_the_Guided_Training_Quest-ff8c00?style=for-the-badge" alt="Start ATV Quest"></a>
</p>

<p align="center">
       <a href="#quick-start">Quick start</a> ·
       <a href="#installation">Installation</a> ·
       <a href="docs/marketplace.md">Marketplace</a> ·
       <a href="#uninstalling">Uninstalling</a> ·
       <a href="#the-three-pillars">Three pillars</a> ·
       <a href="#the-full-sprint">Full sprint</a> ·
       <a href="#how-learning-works">Learning</a> ·
       <a href="https://blazingbeard.github.io/quests/atv-starterkit.html">🎮 Training Quest</a> ·
       <a href="#development">Development</a>
</p>

<video src="https://github.com/user-attachments/assets/7b6bf18a-2bab-482b-a72d-fac9ab7436c2" width="100%" autoplay loop muted playsinline controls></video>

---

## What is ATV 2.0?

ATV 2.0 is a one-command installer that wires together three open-source systems into a single coherent agentic coding environment for GitHub Copilot — grounded in the behavioral principles from [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876) on LLM coding pitfalls:

- **[Compound Engineering](https://github.com/EveryInc/compound-engineering-plugin)** — planning-to-knowledge pipeline
- **[gstack](https://github.com/garrytan/gstack)** — sprint execution engine (by Garry Tan / Y Combinator)
- **[agent-browser](https://github.com/vercel-labs/agent-browser)** — browser automation layer (by Vercel)
- **[Karpathy Guidelines](https://github.com/forrestchang/andrej-karpathy-skills)** — behavioral guardrails: think before coding, simplicity first, surgical changes, goal-driven execution

Together they cover the full software lifecycle — from "what should I build?" through "is it healthy in production?" — with 45+ skills, 51 agents, and a learning system that makes your repo smarter with every session.

---

## Quick Start

**Project install** (scaffolds files into your repo, team-shared):

```bash
cd your-project
npx atv-starterkit@latest init           # auto-detect stack, install everything
npx atv-starterkit@latest init --guided  # interactive TUI with multi-stack selection
npx atv-starterkit@latest uninstall      # cleanly remove everything ATV installed
```

**Personal install** (VS Code source install or Copilot CLI marketplace, follows you across projects):

VS Code / VS Code Insiders:

1. Open the Command Palette.
2. Run `Chat: Install Plugin from source`.
3. Enter `All-The-Vibes/ATV-StarterKit`.
4. Choose `atv-starter-kit`.

Copilot CLI:

```bash
copilot plugin marketplace add All-The-Vibes/ATV-StarterKit
copilot plugin install atv-starter-kit@atv-starter-kit
```

The VS Code source-install path gives one complete ATV option. The Copilot CLI marketplace keeps category bundles and per-skill plugins for CLI users. Both personal paths can coexist with the project scaffold. See [Installation](#installation) for the decision matrix and [docs/marketplace.md](docs/marketplace.md) for CLI bundles and per-skill plugins.

Then open **Copilot Chat** (⌃⌘I / Ctrl+Shift+I) and go:

```text
/kb-brainstorm   →  Explore the problem, produce a design doc
/kb-task         →  Reason from first principles, choose the KB route, continue until verified or blocked
/kb-troubleshoot →  Reproduce, inspect logs/browser evidence, research known fixes, fix, and verify
/kb-plan         →  Generate an implementation plan with acceptance criteria
/kb-work         →  Build against the plan with incremental commits
/kb-review       →  KB code review with thermonuclear structural-quality lens
/ce-review       →  Multi-agent code review (security, architecture, performance)
/ce-compound     →  Document what you learned for future sessions

/klfg             →  Run the full KB pipeline

/atv-doctor      →  Diagnose ATV install health
/atv-update      →  Update ATV marketplace plugins and safe source-installed AgentPlugins
```

---

## Installation

ATV ships in **three flavours** — pick whichever matches your need:

| | `npx atv-starterkit init` | VS Code source install | Copilot CLI marketplace |
|---|---|---|---|
| **Files land in** | Your project's `.github/`, `.vscode/`, `docs/` | VS Code AgentPlugin directory | `~/.copilot/installed-plugins/` |
| **Scope** | Project-level, committed, team-shared | Personal/editor-level | Personal, follows you across CLI projects |
| **What ships** | Skills + agents + MCP + hooks + instructions + setup-steps + docs | One complete ATV skills + agents bundle | Skills + agents only |
| **Best for** | Bootstrapping a new repo, codifying team workflow | VS Code Copilot users who want one obvious install choice | CLI users who want bundles or granular skills |

### Path 1 — npm scaffold (project-level, recommended for teams)

```bash
npx atv-starterkit@latest init       # quick run — downloads binary automatically
npm install -g atv-starterkit        # or global install
atv-starterkit init                  # then run from anywhere
```

The npm package downloads the correct platform binary from [GitHub Releases](https://github.com/All-The-Vibes/ATV-StarterKit/releases) — no Go toolchain needed.

#### Binary (direct download)

Grab a pre-built binary from [GitHub Releases](https://github.com/All-The-Vibes/ATV-StarterKit/releases/latest) for your platform (macOS, Linux, Windows — amd64/arm64).

#### From source

```bash
git clone https://github.com/All-The-Vibes/ATV-StarterKit.git
cd ATV-StarterKit && go build -o atv-installer .
```

### Path 2 — VS Code source install (personal, editor-level)

In VS Code or VS Code Insiders:

1. Open the Command Palette.
2. Run `Chat: Install Plugin from source`.
3. Enter `All-The-Vibes/ATV-StarterKit`.
4. Choose `atv-starter-kit`.

This installs the complete recommended ATV personal bundle: all ATV skills and all reviewer/specialist agents. It does not install MCP config, hooks, instructions templates, setup steps, or project docs; use Path 1 for those.

### Path 3 — Copilot CLI marketplace (personal, cross-project)

```bash
copilot plugin marketplace add All-The-Vibes/ATV-StarterKit
copilot plugin install atv-starter-kit@atv-starter-kit       # all skills + all agents
```

Or pick a category bundle / single skill — full tier breakdown in **[docs/marketplace.md](docs/marketplace.md)**:

```bash
copilot plugin install atv-pack-planning@atv-starter-kit    # one category
copilot plugin install atv-skill-autoresearch@atv-starter-kit  # one skill
```

The CLI marketplace ships skills + agents only. For MCP config, hooks, instructions templates, and docs scaffolding use Path 1. For the cleanest VS Code picker, use Path 2.

### Prerequisites

**Required:** Git, Node.js 16+ (for Path 1).

**Optional:**
- **Bun** — for gstack browser skills (`/gstack-qa`, `/gstack-browse`, `/gstack-benchmark`)
- **GitHub PAT** — for GitHub MCP server
- **Azure CLI** — for Azure MCP server
- **Copilot CLI** — for Path 3 (`copilot` command)

Without Bun, text-based gstack skills still work. `agent-browser` works independently of Bun.

### Uninstalling

```bash
npx atv-starterkit@latest uninstall          # remove ATV files, preserve user-modified configs
npx atv-starterkit@latest uninstall --force  # remove everything including modified files
```

Removes `.github/skills/`, `.github/agents/`, `.github/hooks/`, `.github/copilot-*` config files, `.gstack/`, `.atv/`, and empty doc directories. Files you've customized since installation are preserved by default (checksum comparison against the install manifest). `.vscode/` is never touched.

---

## The Three Pillars

### Karpathy Guidelines — the behavioral foundation

Every skill and agent in ATV operates under four principles derived from [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876) on how LLMs fail at coding. These are installed as a skill (`.github/skills/karpathy-guidelines/SKILL.md`) and shape how Copilot approaches all work:

| Principle | What it prevents |
|---|---|
| **Think Before Coding** | Wrong assumptions, hidden confusion, silently picking one interpretation |
| **Simplicity First** | Overcomplication, bloated abstractions, speculative features |
| **Surgical Changes** | Drive-by refactoring, touching code you shouldn't, cosmetic "improvements" |
| **Goal-Driven Execution** | Vague success criteria, no verification loop, "make it work" without checking |

These aren't just instructions — they're the operating contract between you and the AI. Without them, Copilot tends toward the exact pitfalls Karpathy described: "The models make wrong assumptions on your behalf and just run along with them."

### Autoresearch — autonomous experimentation loop

For tasks with a measurable metric — performance tuning, test pass rate, bundle size, latency, build time — `/autoresearch` runs an autonomous loop: define the goal + metric + scope, the agent works on a dedicated `autoresearch/<tag>` branch, committing each experiment, running the metric command, and keeping or reverting based on the result. Every experiment is logged to `results.tsv` so you can audit the research trail when the loop ends.

Installed as a skill (`.github/skills/autoresearch/SKILL.md`). Sourced from [github/awesome-copilot](https://github.com/github/awesome-copilot/blob/main/skills/autoresearch/SKILL.md) (MIT) by [@luiscantero](https://github.com/luiscantero), inspired by [Karpathy's autoresearch](https://github.com/karpathy/autoresearch).

**Use when** you have a measurable outcome and want the agent to hill-climb autonomously. **Skip for** one-shot tasks, simple bug fixes, or anything without a clear metric.

### Compound Engineering — knowledge compounds

A gated pipeline where each step produces an artifact the next step consumes:

```text
/kb-brainstorm → /kb-plan → /kb-work → /kb-review → /ce-compound
```

Every time you run `/ce-compound`, solved problems get saved to `docs/solutions/`. Next time `/kb-plan` runs, the `learnings-researcher` agent searches those files first. Your repo gets smarter with every PR.

### gstack — the AI sprint process

30 slash-command skills covering office hours, engineering review, browser QA, shipping, deploy verification, security audits, and weekly retros. gstack doesn't just give the AI more tools — it gives it a *role*. `/gstack-review` acts as a staff engineer. `/gstack-cso` acts as a chief security officer. The skills are opinionated engineering processes encoded as markdown.

Includes safety guardrails (`/gstack-careful`, `/gstack-freeze`, `/gstack-guard`) that prevent destructive commands like `rm -rf` or force-pushes.

### agent-browser — the eyes of the agent

A native Rust CLI that controls Chrome via CDP with ~100ms latency. Uses snapshot refs (`@e1`, `@e2`) for deterministic element selection — no CSS selectors or XPath needed. The `open → snapshot → interact → re-snapshot` workflow fits cleanly into an LLM's tool-calling loop.

---

## The Guided Experience

The guided installer (`--guided`) walks you through four screens:

**1. Stack Packs** — Multi-select your stacks (TypeScript, Python, Rails). Auto-detected packs are pre-selected.

**2. Preset** — Choose your depth:

| Preset | What you get |
|---|---|
| **Starter** | Core KB workflow (13 skills). No network calls, instant install. |
| **Pro** | + gstack sprint skills (35+ skills total) |
| **Full** | + browser QA, benchmarks, agent-browser, Chrome (45+ skills). Requires Bun. |

**3. Customize** — Power users can drill into category-grouped multi-select. Beginners skip straight to install.

The customize screen exposes opt-in skill layers grouped by intent:

| Layer | Contents |
|---|---|
| **`core-skills`** | Planning, lifecycle, learning, quality, security, behavioral guidelines |
| **`orchestrators`** | KLFG, ralph-loop, feature-video, test-browser |
| **`dev-tools`** | git-worktree, git-commit / git-commit-push-pr, ghcp-review-resolve, onboarding, reproduce-bug, skill-creator, todo-create / -resolve / -triage, changelog, git-clean-gone-branches |
| **`style-skills`** | dhh-rails-style, andrew-kane-gem-writer, dspy-ruby, every-style-editor, frontend-design |
| **`media-skills`** | gemini-imagegen, proof, rclone |
| **`easter-eggs`** | memeIQ |

**4. Install + Summary** — Real-time progress with structured telemetry, then actionable next steps.

```text
  ✅ Scaffolding ATV files (24 files created, 8 directories) · 340ms
  ⚠️  Syncing gstack skills — fell back to markdown-only · 2.1s
  ✅ Installing agent-browser (CLI ready, skill copied) · 1.8s

  🎉 ATV Starter Kit ready!
  Install state saved to .atv/install-manifest.json
```

---

## The Full Sprint

Every skill maps to a phase of the development lifecycle:

<table>
       <tr>
              <td width="25%" valign="top">
                     <strong>💭 Think</strong><br />
                     <sub>Frame the problem</sub><br /><br />
                     <code>/kb-brainstorm</code><br />
                     <code>/gstack-office-hours</code>
              </td>
              <td width="25%" valign="top">
                     <strong>📋 Plan</strong><br />
                     <sub>Pressure-test the approach</sub><br /><br />
                     <code>/kb-plan</code><br />
                     <code>/gstack-plan-ceo-review</code><br />
                     <code>/gstack-plan-eng-review</code><br />
                     <code>/gstack-plan-design-review</code><br />
                     <code>/gstack-autoplan</code>
              </td>
              <td width="25%" valign="top">
                     <strong>🔨 Build</strong><br />
                     <sub>Execute with momentum</sub><br /><br />
                     <code>/kb-work</code><br />
                     <code>/klfg</code>
              </td>
              <td width="25%" valign="top">
                     <strong>👀 Review</strong><br />
                     <sub>Find what you missed</sub><br /><br />
                     <code>/ce-review</code><br />
                     <code>/gstack-review</code><br />
                     <code>/gstack-cso</code><br />
                     <code>/gstack-codex</code>
              </td>
       </tr>
       <tr>
              <td width="33.33%" valign="top">
                     <strong>🧪 Test</strong><br />
                     <sub>Use real browser eyes</sub><br /><br />
                     <code>agent-browser</code><br />
                     <code>/gstack-qa</code><br />
                     <code>/gstack-benchmark</code><br />
                     <code>/gstack-browse</code>
              </td>
              <td width="33.33%" valign="top">
                     <strong>🚀 Ship</strong><br />
                     <sub>Land without chaos</sub><br /><br />
                     <code>/gstack-ship</code><br />
                     <code>/gstack-land-and-deploy</code><br />
                     <code>/gstack-canary</code><br />
                     <code>/gstack-document-release</code>
              </td>
              <td width="33.33%" valign="top">
                     <strong>📊 Reflect</strong><br />
                     <sub>Compound what you learned</sub><br /><br />
                     <code>/ce-compound</code><br />
                     <code>/learn</code><br />
                     <code>/evolve</code><br />
                     <code>/unslop</code><br />
                     <code>/gstack-retro</code>
              </td>
       </tr>
</table>

### `/klfg` — full KB pipeline

Each step must produce output before the next starts: requirements exist, the vertical-slice manifest exists, every slice is complete, and review/learning has finished. Retries resume from the first missing gate.

"Don't ask many questions", "go straight to work", and similar requests shorten Q&A; they do not skip planning. ATV must still produce or reuse a KB manifest before `/kb-work`, so `/kb-complete` has slice scope and verification evidence.

The manifest's `expected_files` are a planning forecast, not a literal file prison. `/kb-work` records justified file discovery in the scope ledger and stops only for real boundary expansion or unrelated edits.

```
brainstorm → plan → work → complete
     ✓        ✓      ✓        ✓
```

`compound` saves learnings for future `kb-plan` runs. Run `/unslop` separately when you want an explicit cleanup pass.

<details>
<summary><strong>Full skill reference (45 skills)</strong></summary>

### Think

| Skill | What it does |
|---|---|
| `/kb-brainstorm` | Interactive dialogue to clarify requirements; produces design docs in `docs/brainstorms/` and invokes `/kb-plan` when gate-clean |
| `/gstack-office-hours` | YC-style forcing questions that challenge your framing before you write code |
| `/gstack-plan-ceo-review` | CEO-level review: find the 10-star product hiding in the request |

### Plan

| Skill | What it does |
|---|---|
| `/kb-plan` | Produces the KB manifest and vertical slice plans; if execution intent is present, hands off to `/kb-work` |
| `/deepen-plan` | Enriches each plan section with best practices and performance guidance |
| `/gstack-plan-eng-review` | Forces hidden assumptions into the open: architecture, data flow, edge cases |
| `/gstack-plan-design-review` | Scores design quality 0-10 per dimension; rewrites plan to hit 10 |
| `/gstack-autoplan` | Runs CEO → design → eng review in one command |

### Build

| Skill | What it does |
|---|---|
| `/kb-task` | First-principles task runner: choose the KB route and continue until verified or blocked |
| `/kb-troubleshoot` | Autonomous troubleshooting: reproduce, inspect logs/browser/tests, research uncertain assumptions and known fixes, fix, and verify |
| `/kb-work` | Implements an existing KB manifest with incremental commits and system-wide sanity checks; routes free-form work back through `/kb-plan` first |
| `/klfg` | Full KB pipeline: brainstorm → plan → work → complete |

### Review

| Skill | What it does |
|---|---|
| `/kb-review` | KB review gate with thermonuclear structural-quality reviewer plus security, performance, correctness, testing, and standards |
| `/ce-review` | Parallel review agents: security, performance, architecture, language-specific |
| `/gstack-review` | Staff-level code review with auto-fix and completeness checks |
| `/gstack-design-review` | Design audit with atomic fix commits |
| `/gstack-cso` | OWASP Top 10 + STRIDE threat model |
| `/gstack-codex` | Cross-model review via OpenAI Codex CLI |

### Test

| Skill | What it does |
|---|---|
| `agent-browser` | Direct browser automation: open, snapshot, click, fill, screenshot, inspect |
| `/gstack-qa` | Full QA loop: find bugs in real browser, fix them, write regressions, re-verify |
| `/gstack-qa-only` | Report-only QA (no fixes) |
| `/gstack-benchmark` | Page load baselines, Core Web Vitals, resource sizes |
| `/gstack-browse` | Persistent browser runtime for deeper sessions |

### Ship

| Skill | What it does |
|---|---|
| `/gstack-ship` | Sync main, run tests, audit coverage, push, open PR |
| `/gstack-land-and-deploy` | Merge → CI → deploy → verify production |
| `/gstack-canary` | Post-deploy monitoring for errors and regressions |
| `/gstack-document-release` | Auto-update project docs to match what shipped |

### Reflect

| Skill | What it does |
|---|---|
| `/ce-compound` | Documents solved problems in `docs/solutions/` for future sessions |
| `/learn` | Extracts coding patterns from recent work into instincts with confidence scoring |
| `/instincts` | Dashboard showing all learned patterns grouped by domain |
| `/evolve` | Promotes mature instincts (confidence >0.8) into permanent Copilot skills |
| `/observe` | Focused pattern analysis on a specific domain or file pattern |
| `/unslop` | De-slop pass: code simplification + comment rot + design slop detection |
| `/gstack-retro` | Team-aware weekly retro with per-person breakdowns |
| `/gstack-learn` | Per-project self-learning infrastructure |

### Safety Guardrails

| Skill | What it does |
|---|---|
| `/gstack-careful` | Warns before `rm -rf`, `DROP TABLE`, force-push |
| `/gstack-freeze` | Restricts edits to one directory while debugging |
| `/gstack-guard` | Careful + Freeze combined |
| `/gstack-investigate` | No fixes without systematic investigation first |

</details>

---

## How Learning Works

Most AI coding tools treat every session as day one. ATV remembers.

Every time you start a Copilot session, the AI has no memory of how *your team* writes code — that you wrap errors with `%w`, prefer table-driven tests, or use constructor injection. ATV fixes this with a **continuous learning pipeline** that observes how you code, extracts reusable patterns, and graduates proven ones into permanent Copilot skills.

### The Loop

```text
You code normally
     ↓
Observer hooks silently capture tool use → .atv/observations.jsonl
     ↓
/learn analyzes observations + git history → instincts with confidence scores
     ↓
Confidence grows with each session (0.5 → 0.6 → 0.7 → 0.8)
     ↓
/evolve promotes mature instincts → .github/skills/learned-*/SKILL.md
     ↓
Next session: Copilot already knows your patterns
```

### Observer Hooks

ATV installs hooks for all 6 Copilot lifecycle events (`sessionStart`, `sessionEnd`, `preToolUse`, `postToolUse`, `userPromptSubmitted`, `errorOccurred`). A lightweight Node.js script captures every tool interaction to `.atv/observations.jsonl` — silently, with zero impact on your workflow.

### Instincts

`/learn` analyzes git history, diffs, observations, and existing solutions to find recurring patterns. Each becomes an "instinct" with a confidence score:

```yaml
# .atv/instincts/project.yaml
instincts:
  - id: always-wrap-errors
    trigger: "when returning an error from a function"
    behavior: "wrap with fmt.Errorf using %w"
    confidence: 0.85
    observations: 12
```

Run `/instincts` to see the dashboard:

```text
  Error Handling (2 instincts)
    ★ always-wrap-errors        0.9  "wrap errors with fmt.Errorf %w"    15 obs
    ● sentinel-errors           0.6  "use sentinel errors for expected"   5 obs

  Testing (1 instinct)
    ★ table-driven-tests        0.85 "use table-driven test pattern"     12 obs

  Legend: ★ ready to evolve (>0.8)  ● active  ○ tentative (<0.5)
```

When an instinct reaches >0.8 confidence, `/evolve` promotes it into a full SKILL.md at `.github/skills/learned-*/`. Copilot auto-discovers these — your AI assistant now *permanently knows* your team's conventions.

### Design Decisions

- **Instincts are committed to git** — the whole team benefits, not just one developer
- **Observations are gitignored** — raw data is ephemeral, instincts are permanent
- **Generated skills use `learned-` prefix** — visually distinct from hand-written skills
- **Confidence scoring prevents noise** — only well-established patterns get promoted

---

## De-Slop

AI coding assistants have a tell: over-abstraction, `// This function handles the logic for...` comments, purple-to-blue gradients. Code review catches bugs — but nobody catches *slop*.

`/unslop` runs three parallel analysis passes on your recent changes:

```text
/unslop                          →  Report slop in changed files
/unslop src/components/          →  Scope to a directory
/unslop fix                      →  Auto-apply safe fixes
```

| Pass | What it catches | Example |
|------|----------------|---------|
| **Code Slop** | Over-abstraction, YAGNI violations, nested ternaries | Interface used once → inline it |
| **Comment Rot** | Obvious restatements, AI filler phrases, stale TODOs | `// This function handles auth` → delete |
| **Design Slop** | Generic gradients, template layouts, missing hover states | Purple-to-blue default → use brand palette |

`/unslop` is available as a deliberate cleanup pass after review when you want it.

`/ce-review` asks "is this correct?" — `/unslop` asks "does this look human-written?" Run both.

---

## Memory Architecture

ATV builds seven layers of memory across three reinforcing cycles:

| Layer | Where | Timescale |
|---|---|---|
| **Observations** | `.atv/observations.jsonl` | Per-session (gitignored) |
| **Instincts** | `.atv/instincts/project.yaml` | Grows every session |
| **Evolved skills** | `.github/skills/learned-*/` | Permanent |
| **Institutional knowledge** | `docs/solutions/*.md` | Permanent |
| **Design decisions** | `docs/brainstorms/*.md` | Permanent |
| **Implementation plans** | `docs/plans/*.md` | Per-feature |
| **Install manifest** | `.atv/install-manifest.json` | Per-install |

**How they reinforce each other:**

- **Knowledge compounding** (per-PR): `/ce-compound` saves solved problems → future `/kb-plan` finds them via `learnings-researcher` → fewer repeated mistakes
- **Pattern learning** (per-session): observer hooks → `/learn` → instincts → `/evolve` → permanent skills → Copilot knows your conventions
- **Team propagation** (per-commit): instincts are committed to git → the whole team inherits learned patterns without a style guide

Over weeks, your repo develops a memory that makes every Copilot session more effective than the last.

---

## Agents

51 specialized agents in `.github/agents/`, invoked by skills during review, planning, learning, and debugging:

| Category | Agents |
|---|---|
| **Code Review** | `kieran-rails-reviewer`, `kieran-python-reviewer`, `kieran-typescript-reviewer`, `dhh-rails-reviewer`, `code-simplicity-reviewer`, `julik-frontend-races-reviewer` |
| **Security** | `security-sentinel` |
| **Architecture** | `architecture-strategist` |
| **Performance** | `performance-oracle` |
| **Data** | `data-integrity-guardian`, `data-migration-expert`, `schema-drift-detector`, `deployment-verification-agent` |
| **Design** | `design-implementation-reviewer`, `design-iterator`, `figma-design-sync` |
| **Research** | `repo-research-analyst`, `best-practices-researcher`, `framework-docs-researcher`, `learnings-researcher`, `git-history-analyzer` |
| **Process** | `pr-comment-resolver`, `spec-flow-analyzer`, `bug-reproduction-validator`, `pattern-recognition-specialist` |
| **Learning** | `pattern-observer` |
| **Meta** | `agent-native-reviewer`, `ankane-readme-writer` |
| **Ops** | `lint` |

---

## What Gets Installed

### Copilot Integration Points

| File | Purpose |
|---|---|
| `.github/copilot-instructions.md` | System instructions loaded into every chat |
| `.github/copilot-setup-steps.yml` | Coding Agent initialization steps |
| `.github/copilot-mcp-config.json` | MCP server configuration |
| `.github/skills/*/SKILL.md` | Skills auto-discovered by description match |
| `.github/agents/*.agent.md` | Agents for subagent orchestration |
| `.github/*.instructions.md` | File-scoped instructions via `applyTo` globs |
| `.github/hooks/copilot-hooks.json` | Observer hooks (silent, every tool use) |

### Supported Stacks

| Stack | Detection | Additions |
|---|---|---|
| **TypeScript** | `tsconfig.json` | TypeScript reviewer, TS file instructions |
| **Python** | `pyproject.toml` / `requirements.txt` | Python reviewer, Python file instructions |
| **Rails** | `Gemfile` + `config/routes.rb` | 8 Rails-specific agents, Ruby file instructions |
| **General** | fallback | Universal agents and skills |

### MCP Servers

| Server | Type | Package |
|---|---|---|
| **Context7** | SSE | `mcp.context7.com` |
| **GitHub** | stdio | `@modelcontextprotocol/server-github` |
| **Azure** | stdio | `@azure/mcp` |
| **Terraform** | stdio | `terraform-mcp-server` |

---

## How It Works Under the Hood

```text
atv-installer init --guided
        │
        ▼
 Detect stack + prerequisites (git, bun, node)
        │
        ▼
 Stack Packs → Preset → Customize?
        │
        ▼
 Install with structured telemetry:
        │
        ├── ATV scaffold ──► Embedded templates → .github/skills/*/SKILL.md
        │
        ├── Learning pipeline ──► Observer hooks + skills + instinct storage
        │
        ├── gstack ──► git clone → .gstack/ (staging, gitignored)
        │               └── Copy SKILL.md → .github/skills/gstack-*/
        │
        └── agent-browser ──► npm install -g → agent-browser install (Chrome)
                              └── .github/skills/agent-browser/SKILL.md
        │
        ▼
 Write manifest to .atv/install-manifest.json
```

All templates are embedded at compile time — no runtime network calls for the core scaffold. gstack requires a network clone (~22MB). Re-running is idempotent: existing files are skipped, JSON configs are merged.

---

## Development

```bash
go build -o atv-installer .             # build
go test ./...                            # all tests
go test ./pkg/installstate/ -v           # manifest + recommendations tests
go test ./pkg/monitor/ -v                # watcher + drift detection tests
go test ./test/sandbox/ -v               # integration tests (E2E scenarios)
```

## Limitations

- **Bun required for browser skills** — `/gstack-qa`, `/gstack-browse`, `/gstack-benchmark`
- **Network required for gstack** — clones ~22MB at install time
- **gstack setup on Windows** — falls back to `bun run gen:skill-docs` (bash path issues)
- **Token-heavy pipelines** — long multi-agent sessions can hit context limits

---

<div align="center">

MIT — Built by [All The Vibes](https://github.com/All-The-Vibes)

Powered by [Compound Engineering](https://github.com/EveryInc/compound-engineering-plugin) · [gstack](https://github.com/garrytan/gstack) · [agent-browser](https://github.com/vercel-labs/agent-browser) · [Karpathy Guidelines](https://github.com/forrestchang/andrej-karpathy-skills)

Special thanks to [blazingbeard](https://github.com/blazingbeard) for building out the [guided training quest](https://blazingbeard.github.io/quests/atv-starterkit.html).

</div>
