# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [2.6.2] — 2026-04-26

### Fixed

- **Marketplace registration was rejected by Copilot CLI** — the v2.6.1 marketplace contained a plugin entry named `atv-skill-resolve_todo_parallel` (sourced from the `resolve_todo_parallel` skill template, which uses an underscore). Copilot CLI's marketplace validator requires plugin names to be strict kebab-case (letters, numbers, hyphens only) and refused to load the marketplace with `Plugin name must be kebab-case`. The generator now sanitizes plugin names by converting underscores to hyphens (`atv-skill-resolve-todo-parallel`). The skill's slash command and SKILL.md `name:` field are unchanged — only the marketplace-facing plugin name is sanitized. Regression test added in `pkg/plugingen/generate_test.go`.

## [2.6.1] — 2026-04-26

### Added

- **Copilot CLI plugin marketplace** — ATV Starter Kit is now installable via `copilot plugin marketplace add All-The-Vibes/ATV-StarterKit`. The marketplace ships three install tiers:
  - **`atv-skill-<name>`** — 29 granular per-skill plugins for cherry-pick installs (note: not standalone — see `docs/marketplace.md`).
  - **`atv-pack-<category>`** — 8 category bundles (`atv-pack-planning`, `atv-pack-review`, `atv-pack-shipping`, `atv-pack-security`, `atv-pack-quality`, `atv-pack-guidelines`, `atv-pack-easter-eggs`, `atv-pack-learning`).
  - **`atv-everything`** + **`atv-agents`** — flagship bundle and standalone agents plugin.
- **Plugin generator** (`pkg/plugingen/` and `cmd/plugingen/`) — single source of truth: `go run ./cmd/plugingen` regenerates the entire `plugins/` tree and `.github/plugin/marketplace.json` from `pkg/scaffold/templates/`. Deterministic output (sorted lists, slash-normalized paths, LF line endings) for reliable CI drift checks via `go run ./cmd/plugingen -check`.
- **Two install paths documented** — new `docs/marketplace.md` plus a top-level README section comparing project-level (`atv init`) vs personal (marketplace) installs with a decision matrix.

### Changed

- **`atv init` is unchanged** — the marketplace is purely additive. Project scaffolding continues to read from the same templates.

### Notes

- **MCP server config plugin (`atv-mcp`) deferred** — generic server names (`github`, `azure`, `terraform`, `context7`) would silently override existing user MCP configs under last-wins precedence. Will ship in a future release once a namespacing strategy is designed.
- **Hooks, instructions templates, and stack-specific reviewer agent splits** are also deferred — they require more design (substitution, runtime, etc.) and are out of MVP scope.

## [2.6.0] — 2026-04-26

### Added

- **Autoresearch skill** — autonomous iterative experimentation loop for any programming task with a measurable outcome. Define a goal + metric command + scope, and the agent works on a dedicated `autoresearch/<tag>` branch, committing each experiment, running the metric, and keeping or reverting based on the result. Every experiment is tracked in `results.tsv`. Sourced verbatim from [github/awesome-copilot](https://github.com/github/awesome-copilot/blob/main/skills/autoresearch/SKILL.md) (MIT, by [@luiscantero](https://github.com/luiscantero), inspired by [Karpathy's autoresearch](https://github.com/karpathy/autoresearch)). Installs as `.github/skills/autoresearch/SKILL.md`.
- **Coding Guidelines TUI category now lists Autoresearch** — `📐 Coding Guidelines` shows both `Karpathy Guidelines` and `Autoresearch` as toggleable, both pre-selected. Autoresearch is included in all three presets (Starter, Pro, Full) as a core skill.
- **`/autoresearch` discoverable in instruction templates** — added to the Available Workflows section of `general.md`, `python.md`, `rails.md`, and `typescript.md` so Copilot routes appropriate prompts (performance tuning, hill climbing, automated experimentation) to the skill.

## [2.5.9] — 2026-04-26

### Changed

- **`/cso` folded into `/atv-security`** — the installer's standalone `/cso` skill (OWASP Top 10 + STRIDE for application source code) has been merged into `/atv-security`. The unified skill now scans both agentic configuration (`.github/`, `.vscode/`) AND application source code in a single run. This eliminates a name collision with gstack's separate `/cso` skill (which remains untouched and continues to ship via gstack). Old `/cso` triggers (`cso`, `owasp scan`, `stride analysis`, `threat model`, `application security`, `security review code`) still route to `/atv-security` for migration discoverability.
- **Argument grammar for `/atv-security`** — now accepts two axes: `[mode: report|fix] [scope: full|config|owasp|stride|<path>]`. Defaults remain `report` + `full`. Examples: `/atv-security`, `/atv-security fix`, `/atv-security owasp`, `/atv-security src/api/`.
- **N/A scoring semantics** — when only one surface (configs OR source) is present, the absent half renders as N/A in the report and is excluded from the aggregate grade rather than scored as 0 or 100.
- **Backwards-compatible report persistence** — `docs/security/YYYY-MM-DD-security-report.md` retains both `<!-- atv-security -->` and `<!-- cso -->` marker blocks. `/atv-security` writes the config audit into the `atv-security` block and the OWASP/STRIDE results into the `cso` block, preserving the legacy `## /cso Scan` heading shape so existing reports continue to parse.
- **AGENT-03 (oversized prompt) self-exemption** — the merged `/atv-security` skill file intentionally exceeds 8,000 chars. The AGENT-03 rule now exempts first-party ATV security skills so the auditor doesn't flag itself.
- **Guided installer Security category** — the customize wizard's `🔒 Security` category now shows a single `ATV Security — agentic config audit + OWASP Top 10 + STRIDE source-code review` entry. Users with gstack installed still see gstack's `/cso` separately listed under `[gstack]`.
- **Instruction templates broadened** — `general.md`, `python.md`, `rails.md`, and `typescript.md` now describe `/atv-security` as covering both config and application code; the `/cso` line was removed.

### Removed

- **`/cso` template skill** — `pkg/scaffold/templates/skills/cso/SKILL.md` and its directory have been deleted from the installer. Gstack's `/cso` is unaffected.

### Added

- **memeIQ Easter Egg installer option** — guided installs now expose a `🥚 Easter Eggs` category with an opt-in `memeIQ` entry that scaffolds `.github/skills/meme-iq/SKILL.md` and `.github/agents/meme-iq.agent.md`.

### Notes

- **Migration:** users who previously typed `/cso` will land in `/atv-security` thanks to preserved triggers. To explicitly invoke the OWASP/STRIDE phase only, use `/atv-security owasp` or `/atv-security stride`.
- **Regression guard:** `pkg/tui/categories_test.go` now asserts that `core-skills:cso` does not reappear in the Security category, preventing accidental re-introduction of the name collision.

### Removed

- **`claude-permissions-optimizer` skill** — ATV-starterkit is a GitHub Copilot harness; the optimizer's scripts read `~/.claude/settings.json` and manage Claude Code's permission allowlist, which has no Copilot equivalent. Removed from `.github/skills/`, `pkg/scaffold/templates/skills/`, `orchestratorSkillDirectories` in `pkg/scaffold/catalog.go`, and the shipping category in `pkg/tui/categories.go`. Users who want this skill can install it directly from upstream Compound Engineering for their personal `~/.claude/` config.
- **PR #28 (`feat(land): sweep stale ralph-loop state files`) closed unmerged** — the proposed `find . -name ralph-loop.local.md -path '*/.claude/*' -delete` block in the `land` skill's Step 7 cleanup was Claude-Code-specific (the `.claude/` directory and ralph-loop's stop-hook replay are both Claude Code mechanisms) and produced no value in a Copilot-only installation.

## [2.5.7] — 2026-04-15

### Added

- **Karpathy Guidelines skill** — behavioral guardrails derived from [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876) on LLM coding pitfalls, ported from [forrestchang/andrej-karpathy-skills](https://github.com/forrestchang/andrej-karpathy-skills) (Claude Code plugin) to GitHub Copilot's instruction system. Installs as `.github/skills/karpathy-guidelines/SKILL.md` with four principles: Think Before Coding, Simplicity First, Surgical Changes, and Goal-Driven Execution.
- **Coding Guidelines TUI category** — new `📐 Coding Guidelines` category in the guided installer's customization screen. Karpathy Guidelines are included in all three presets (Starter, Pro, Full) as a core skill and can be toggled in the customize step.

## [2.5.6] — 2026-04-12

### Added

- **Training quest link in README** — added link to the [ATV Starter Kit Quest](https://blazingbeard.github.io/quests/atv-starterkit.html), a guided and gamified training experience by [blazingbeard](https://github.com/blazingbeard).

### Fixed

- **Suppressed noisy gstack output during guided install** — gstack's setup generates skills for every supported host (Cursor, Slate, OpenClaw, Kiro, Factory, OpenCode) then ATV prunes them. Previously all that per-file generation output and token budget tables leaked to stdout. Now subprocess output is captured silently — users see only the TUI spinner and final summary.
- **Copilot hooks hardened against missing node** — observer hook commands now suppress errors (`2>/dev/null || true` on bash, `try/catch` on PowerShell) so projects without Node.js don't get hook failures on every session.
- **Removed excessive observer hooks** — stripped `userPromptSubmitted`, `preToolUse`, `postToolUse`, and `errorOccurred` hooks that fired on every interaction. Only `sessionStart` and `sessionEnd` remain, reducing hook overhead.
- **Prune ordering fixed** — non-GitHub platform dirs are now pruned before copying skills (was after), preventing any chance of non-GitHub artifacts leaking into `.github/skills/`.

## [2.5.5] — 2026-04-09

### Added

- **`atv-installer uninstall` command** — cleanly removes all ATV-installed files from a project. Removes `.github/skills/`, `.github/agents/`, `.github/hooks/`, `.github/copilot-*` config files, `.gstack/`, `.atv/`, and empty doc directories. Preserves user-modified files by default (checksum comparison against install manifest). Use `--force` to remove everything.

## [2.5.3] — 2026-04-09

### Fixed

- **Prune non-GitHub platform dirs from gstack staging** — after cloning gstack, the installer now removes `.cursor/`, `.factory/`, `.kiro/`, `.openclaw/`, `.opencode/`, `.slate/`, `codex/`, `openclaw/`, `node_modules/`, `.git/`, `.github/`, `extension/`, `hosts/`, `contrib/`, `supabase/`, `test/`, `scripts/`, and `docs/` from `.gstack/`. These are gstack's multi-platform outputs (OpenClaw, OpenCode, Cursor, Kiro, Slate, Factory) and build artifacts that are irrelevant to GitHub Copilot users.

## [2.5.1] — 2026-04-07

### Changed

- **README overhauled** — eliminated repetitive sections (continuous learning explained 3x, installation covered 2x, guided installer described 3x), consolidated into a single-pass flow. Same details, no redundancy.
- **`/lfg` and `/slfg` pipeline diagrams added** — visual pipeline flows showing step order, parallel execution in `/slfg`, and where `/unslop` and `/ce-compound` fit
- **De-slop and memory sections tightened** — cut verbose pipeline diagrams and filler phrase tables; kept the core pitch and usage

## [2.5.0] — 2026-04-07

### Fixed

- **ce-brainstorm and brainstorming templates restored** — the compound-engineering update (v2.4.0) accidentally flattened both SKILL.md files into single lines, breaking YAML frontmatter parsing and making `/ce-brainstorm` undiscoverable. Both templates now have proper multi-line content with valid frontmatter.

### Changed

- **Learning pipeline skills renamed** — removed `atv-` prefix from all learning pipeline skills for cleaner slash commands:
  - `atv-learn` → `learn` (`/learn`)
  - `atv-instincts` → `instincts` (`/instincts`)
  - `atv-evolve` → `evolve` (`/evolve`)
  - `atv-observe` → `observe` (`/observe`)
  - `atv-unslop` → `unslop` (`/unslop`)
- **`/lfg` workflow updated** — now includes `/observe` and `/learn` steps after `/unslop fix` to capture patterns from the completed work
- **`/slfg` workflow updated** — added a new Learning Phase with `/observe` and `/learn` between the Autofix Phase and Finalize Phase

## [2.0.1] — 2026-04-01

Patch release to fix npm distribution. The v2.0.0 npm package was published before any corresponding GitHub release with goreleaser binaries existed, so the postinstall binary download would fail. This release publishes a GitHub release for v2.0.1 with pre-built binaries and a matching npm version that resolves correctly via `releases/latest`.

### Fixed

- **npm install works end-to-end** — the latest GitHub release (v2.0.1) now ships goreleaser binaries for all platforms (macOS, Linux, Windows on amd64/arm64), and the npm package points installers at this release
- **Version alignment** — VERSION file, npm package, and GitHub release tag are now all in sync at 2.0.1

## [2.0.0] — 2026-03-29

ATV 2.0 — **All The Vibes** becomes a unified installer combining Compound Engineering, gstack, and agent-browser into one agentic coding setup for GitHub Copilot.

### Added

- **gstack integration** — 30 gstack skills from [garrytan/gstack](https://github.com/garrytan/gstack) installable via the guided wizard. Clone, generate docs, copy skills, and create runtime sidecar — all handled by the Go installer.
- **agent-browser integration** — [Vercel agent-browser](https://github.com/vercel-labs/agent-browser) installable via guided mode. Auto-installs the CLI via npm and downloads Chrome for Testing.
- **Preset-based guided wizard** — Three presets (Starter / Pro / Full) replace the flat checkbox list. Beginners pick a preset; power users drill into category-grouped customization.
- **Animated install progress** — Bubbletea-powered progress display with per-step spinners (pending → running → done/failed) replacing the silent line-by-line output.
- **Retro terminal banner** — "ALL THE VIBES 2.0" in block letters with gold gradient and retro boot messages, matching the hero SVG.
- **Runtime sidecar** — `.github/skills/gstack/` directory with binaries (`bin/`), browser runtime (`browse/dist/`), review checklists, and `ETHOS.md` so gstack skills can find their runtime assets.
- **Memory system documentation** — copilot-instructions templates now include ATV Override Rules and conflict resolution between ATV and gstack memory systems.
- **Function-based TUI categories** — Skills organized by what they do (Planning, Review, QA, Security, Shipping, Safety) instead of where they come from.
- **`pkg/gstack/`** — New Go package for gstack installation: prerequisite detection, git clone, skill doc generation, sidecar creation.
- **`pkg/agentbrowser/`** — New Go package for agent-browser: npm install, Chrome download, SKILL.md fetch.
- **`pkg/tui/presets.go`** — Preset definitions (Starter, Pro, Full) with skill mappings.
- **`pkg/tui/progress.go`** — Bubbletea progress model with animated spinners.
- **Sandbox integration tests** — `test/sandbox/` with tests for auto mode, guided mode, gstack install, idempotency, and instructions content.

### Changed

- **README** rewritten as "The Three Pillars" — Compound Engineering (knowledge compounds), gstack (AI sprint process), agent-browser (eyes of the agent). Memory highlighted as a first-class differentiator.
- **Wizard flow** — Stack → Preset → Customize? → (optional) multi-select instead of Stack → flat checkbox wall.
- **copilot-instructions templates** — all 4 stacks (general, typescript, python, rails) now include gstack skill listing, agent-browser section, and ATV Override Rules.
- **`.gitignore`** — added `.gstack/`, `.env`, `*.db`, `.claude/` entries.
- **`cmd/init.go`** — guided mode now runs gstack and agent-browser install steps with progress display; auto mode unchanged.
- **`pkg/output/printer.go`** — new status indicators (🔗 Cloned, 🔨 Built), gstack/agent-browser progress methods, dynamic next-steps with step numbering.

### Fixed

- **gstack skill discovery** — skills placed at `.github/skills/gstack-*/SKILL.md` (one level deep) instead of nested `.github/skills/gstack/{name}/SKILL.md` which Copilot couldn't discover.
- **Windows CRLF** — bash scripts in gstack clone are auto-fixed (`\r\n` → `\n`) before execution.
- **Preset cursor** — default to Starter so all three options are visible in shorter terminals.

## [1.0.0] — 2026-03-11

Initial release of ATV Starter Kit.

### Added

- One-click installer (`atv-installer init`) for GitHub Copilot agentic coding environment
- All 6 Copilot lifecycle hooks: system instructions, setup steps, MCP config, skills, agents, file instructions
- 13 workflow skills (brainstorm, plan, work, review, compound, lfg, slfg, etc.)
- 28 specialized agents (security, performance, architecture, data, design, research)
- Stack detection: TypeScript, Python, Rails, General
- Interactive guided mode (`--guided`) with charmbracelet/huh TUI
- MCP servers: Context7, GitHub, Azure, Terraform
- npm distribution via `npx atv-starterkit`
- Cross-platform binary releases via goreleaser
