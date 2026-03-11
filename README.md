# ATV Starter Kit

**A**gentic **T**ool & **W**orkflow — a one-click installer that scaffolds a complete GitHub Copilot agentic coding environment into any project.

## Installation

### Option 1: Download Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/github/atv-installer/releases):

| Platform | Download |
|----------|----------|
| **Windows** | `atv-installer_windows_amd64.zip` |
| **macOS (Intel)** | `atv-installer_darwin_amd64.tar.gz` |
| **macOS (Apple Silicon)** | `atv-installer_darwin_arm64.tar.gz` |
| **Linux** | `atv-installer_darwin_amd64.tar.gz` |

Extract and move to your PATH:

```bash
# macOS/Linux
tar xzf atv-installer_*.tar.gz
sudo mv atv-installer /usr/local/bin/

# Windows — extract zip, add folder to PATH
```

### Option 2: Build from Source

Requires [Go 1.22+](https://go.dev/dl/):

```bash
git clone https://github.com/github/atv-installer.git
cd atv-installer
go build -o atv-installer .

# Move to PATH
sudo mv atv-installer /usr/local/bin/   # macOS/Linux
# Or on Windows: move atv-installer.exe to a folder in your PATH
```

### Option 3: Go Install

```bash
go install github.com/github/atv-installer@latest
```

## Quick Start

### One-Click Mode (Default)

```bash
cd your-project
atv-installer init
```

That's it. Auto-detects your stack, installs all 6 Copilot lifecycle hooks, done in seconds.

```
  Auto-detected: TypeScript project (tsconfig.json found)

  Created .github/copilot-instructions.md
  Created .github/copilot-setup-steps.yml
  Created .github/copilot-mcp-config.json
  Created .github/skills/ (11 skills)
  Created .github/agents/ (19 agents)
  Created .github/typescript.instructions.md
  Created .vscode/extensions.json
  Created docs/plans/, docs/brainstorms/, docs/solutions/

  ATV Starter Kit ready!

  Next steps:
    1. Open this folder in VS Code
    2. Install recommended extensions when prompted
    3. Try: /ce-brainstorm "your first feature idea"
```

### Guided Mode

Want to customize what gets installed? Use the interactive TUI wizard:

```bash
atv-installer init --guided
```

Select your stack, toggle individual component layers on/off with checkboxes:

```
┃ What's your primary stack?
┃ > TypeScript
┃   Python
┃   Rails
┃   General
┃
┃ Which components do you want?
┃ > [•] Core pipeline skills (brainstorm, plan, work, review, compound)
┃   [•] Orchestrators (lfg, slfg)
┃   [•] Universal agents (security, performance, architecture, ...)
┃   [•] Stack-specific agents (language reviewers)
┃   [•] MCP servers (GitHub, Azure, Terraform, Context7)
┃   [•] VS Code extensions.json
┃   [•] Copilot instructions
┃   [•] Copilot setup steps
┃   [•] File-scoped instructions (applyTo globs)
┃   [•] docs/ structure
┃   [ ] Compound engineering local config
```

**Note:** The guided TUI requires a proper terminal (Windows Terminal, iTerm, etc.). The VS Code integrated terminal may not render Unicode box-drawing characters correctly.

### Idempotent Re-runs

Run it again any time — existing files are skipped, new content is added, JSON configs are merged:

```
  Skipped 37 existing files
  Merged 2 JSON configs
  Created 2 files, 0 directories
```

## What Gets Installed

### All 6 Copilot Lifecycle Hooks

| # | Hook | File | When It Fires |
|---|------|------|---------------|
| 1 | **System Instructions** | `.github/copilot-instructions.md` | Every Copilot chat — injected as system context |
| 2 | **Setup Steps** | `.github/copilot-setup-steps.yml` | When Copilot Coding Agent initializes |
| 3 | **MCP Servers** | `.github/copilot-mcp-config.json` | When Copilot starts — registers tool servers |
| 4 | **Skills** | `.github/skills/*/SKILL.md` | When skill description matches user's request |
| 5 | **Agents** | `.github/agents/*.agent.md` | When invoked by subagent orchestration |
| 6 | **File Instructions** | `.github/*.instructions.md` | Auto-loaded by `applyTo` glob when editing files |

Plus: `.vscode/extensions.json` and `docs/` structure (plans, brainstorms, solutions).

## Supported Stacks

| Stack | Detection | Additional Content |
|-------|-----------|-------------------|
| **TypeScript** | `tsconfig.json` | `kieran-typescript-reviewer` agent, TS file instructions (`applyTo: **/*.ts`) |
| **Python** | `pyproject.toml` / `requirements.txt` | `kieran-python-reviewer` agent, Python file instructions (`applyTo: **/*.py`) |
| **Rails** | `Gemfile` + `config/routes.rb` | 8 additional agents (DHH, data integrity, schema drift, ...), 3 skills, Ruby file instructions |
| **General** | fallback | Universal agents and skills only |

## MCP Servers

Pre-configured in `.github/copilot-mcp-config.json`:

| Server | Type | Package |
|--------|------|---------|
| **Context7** | SSE (remote) | `mcp.context7.com` |
| **GitHub** | stdio (npx) | `@modelcontextprotocol/server-github` (needs PAT) |
| **Azure** | stdio (npx) | `@azure/mcp` |
| **Terraform** | stdio (npx) | `terraform-mcp-server` |

### Extension-Only Servers

These require VS Code extensions (listed in `.vscode/extensions.json` — VS Code prompts to install on open):

| Extension | ID |
|-----------|----|
| Bicep | `ms-azuretools.vscode-bicep` |
| Azure Tools for Copilot | `ms-azuretools.vscode-azure-github-copilot` |
| HashiCorp Terraform | `hashicorp.terraform` |
| GitHub Copilot | `github.copilot` |
| GitHub Copilot Chat | `github.copilot-chat` |

### Prerequisites

- **Node.js** — required for `npx` to run stdio MCP servers
- **GitHub PAT** — prompted on first use (needs `repo`, `read:org` scopes)
- **Azure CLI** — `az login` for Azure MCP authentication

---

## Compound Engineering Pipeline

The ATV Starter Kit includes the **compound-engineering** multi-agent pipeline:

### `/lfg` (sequential) and `/slfg` (swarm/parallel)

These are the "run everything" entry points. The pipeline is:

```
1. /ce-brainstorm  →  Explore WHAT to build (optional, user-driven)
2. /ce-plan        →  Create a structured plan document
3. /deepen-plan    →  Enrich plan sections with parallel research agents
4. /ce-work        →  Execute the plan (write code, tests, commits)
5. /ce-review      →  Multi-agent code review (security, perf, architecture)
6. /resolve-todos  →  Fix findings from review in parallel
7. /test-browser   →  Browser-based testing
8. /feature-video  →  Record a walkthrough, attach to PR
9. /ce-compound    →  Document what you learned for future sessions
```

Each step has a **GATE** — the pipeline won't advance until the prior step produces its artifact (plan file exists, code was actually changed, etc.).

## The 5 Core Skills

### 1. `ce-brainstorm` — "What should we build?"

- Interactive dialogue using `AskUserQuestion` to clarify requirements
- Runs `repo-research-analyst` to understand existing patterns
- Produces a brainstorm document in `docs/brainstorms/`
- Assesses whether brainstorming is even needed (clear requirements skip it)
- Hands off to `/ce-plan`

### 2. `ce-plan` — "How do we build it?"

- Checks for existing brainstorm docs and uses them as primary input
- Runs **parallel research agents**: `repo-research-analyst`, `learnings-researcher`, optionally `best-practices-researcher` + `framework-docs-researcher`
- Makes a **risk-based research decision** — security/payments always get external research; strong local context skips it
- Runs `spec-flow-analyzer` to validate user flows and edge cases
- Outputs a plan in `docs/plans/` with YAML frontmatter, acceptance criteria, and checkboxes

### 3. `ce-work` — "Execute the plan"

- Reads the plan, breaks it into a todo list
- Sets up git branches (supports worktrees for parallel dev)
- Implements task-by-task with a **System-Wide Test Check** table (callbacks, mocks, orphaned state, error alignment)
- Makes **incremental commits** at logical boundaries
- Checks off plan items as they're completed (`[ ]` → `[x]`)
- Optionally runs Figma design sync for UI work
- Quality checks with configurable reviewer agents before finishing

### 4. `ce-review` — "Multi-perspective code review"

- Configurable via `compound-engineering.local.md` (created by `/setup`)
- Launches **parallel review agents** (e.g., security-sentinel, performance-oracle, architecture-strategist, code-simplicity-reviewer, language-specific reviewers like kieran-rails-reviewer)
- Has **conditional agents** that only fire for specific PR types (database migrations trigger schema-drift-detector + data-migration-expert + deployment-verification-agent)
- "Ultra-thinking" deep dive with stakeholder perspective analysis (developer, ops, end user, security, business)
- Supports `--serial` mode for long sessions to avoid context limits

### 5. `ce-compound` — "Document what we learned"

- Captures recently-solved problems into `docs/solutions/` with YAML frontmatter
- Launches 5 parallel sub-agents: Context Analyzer, Solution Extractor, Related Docs Finder, Prevention Strategist
- Has a **context budget check** — warns if the session is too long and offers a compact-safe mode
- Creates searchable institutional knowledge that future sessions consume via `learnings-researcher`

## The Agent Roster (28 Specialized Agents)

The `.github/agents/` directory contains **28 `.agent.md` files**, each a specialized persona:

| Category | Agents |
|----------|--------|
| **Code Review** | `kieran-rails-reviewer`, `kieran-python-reviewer`, `kieran-typescript-reviewer`, `dhh-rails-reviewer`, `code-simplicity-reviewer`, `julik-frontend-races-reviewer` |
| **Security** | `security-sentinel` (OWASP, input validation, secrets) |
| **Architecture** | `architecture-strategist` (SOLID, coupling, boundaries) |
| **Performance** | `performance-oracle` (Big-O, DB queries, memory, scaling) |
| **Data** | `data-integrity-guardian`, `data-migration-expert`, `schema-drift-detector`, `deployment-verification-agent` |
| **Design** | `design-implementation-reviewer`, `design-iterator`, `figma-design-sync` |
| **Research** | `repo-research-analyst`, `best-practices-researcher`, `framework-docs-researcher`, `learnings-researcher`, `git-history-analyzer` |
| **Process** | `pr-comment-resolver`, `spec-flow-analyzer`, `bug-reproduction-validator`, `pattern-recognition-specialist` |
| **Meta** | `agent-native-reviewer` (ensures features are agent-accessible), `ankane-readme-writer` |
| **Ops** | `lint` |

## Key Design Patterns

### 1. Parallel Sub-Agent Orchestration

Nearly every step spawns multiple agents simultaneously. `/ce-plan` runs research agents in parallel. `/ce-review` runs all reviewers in parallel. `/ce-compound` runs 5 extractors in parallel. `/slfg` even parallelizes review + browser testing.

### 2. Living Documents as State

Plans in `docs/plans/` serve as shared state. `/ce-plan` creates them, `/deepen-plan` enriches them, `/ce-work` checks off items, `/ce-review` protects them from deletion. Brainstorms in `docs/brainstorms/` feed into plans. Solutions in `docs/solutions/` feed back into future planning.

### 3. Knowledge Compounding Loop

This is the signature pattern:

```
solve problem → /ce-compound documents it → docs/solutions/
                                                    ↓
future /ce-plan → learnings-researcher searches docs/solutions/ → avoids past mistakes
```

### 4. Configurable Per-Project

The `setup` skill auto-detects your stack (Rails, Python, TypeScript, etc.) and writes `compound-engineering.local.md` with the right reviewer agents. This means `/ce-review` and `/ce-work` adapt to any project.

### 5. Gate-Based Progression

`/lfg` enforces strict ordering with verification gates. No coding until a plan exists. No review until code exists. This prevents the common AI failure mode of jumping straight to implementation.

## Strengths

- **Full-lifecycle coverage** — brainstorm through video demo, nothing is manual
- **Institutional memory** — the compound loop is genuinely novel; AI agents that learn from past sessions
- **Multi-perspective review** — 28 agents cover more angles than any human reviewer
- **Configurable depth** — `/ce-plan` has MINIMAL/STANDARD/COMPREHENSIVE detail levels; agents can run serial or parallel
- **Risk-aware research** — doesn't waste time researching when patterns are clear, but always researches high-risk areas

## Project Structure

```
atv-installer/
├── cmd/
│   ├── root.go              # Cobra root command
│   └── init.go              # `init` subcommand + --guided flag
├── pkg/
│   ├── detect/detect.go     # Stack detection (Rails/Python/TS/General)
│   ├── scaffold/
│   │   ├── scaffold.go      # Idempotent file writer + JSON merge
│   │   ├── catalog.go       # Component registry + go:embed
│   │   ├── hooks.go         # Copilot lifecycle hooks (1-6)
│   │   └── templates/       # All embedded content
│   │       ├── skills/      # 11 SKILL.md files
│   │       ├── agents/      # 28 .agent.md files
│   │       ├── configs/     # MCP config + extensions.json
│   │       ├── instructions/  # copilot-instructions.md per stack
│   │       ├── setup-steps/   # copilot-setup-steps.yml per stack
│   │       └── file-instructions/  # *.instructions.md with applyTo
│   ├── tui/wizard.go        # Interactive guided mode (charmbracelet/huh)
│   └── output/printer.go    # Terminal output with status indicators
├── main.go
├── go.mod / go.sum
├── .goreleaser.yml           # Cross-platform release builds
└── .github/workflows/
    ├── ci.yml                # Build + test + lint
    └── release.yml           # goreleaser on tag push
```

## Development

```bash
# Build
go build -o atv-installer .

# Run locally
./atv-installer init
./atv-installer init --guided

# Test in a sandbox
mkdir /tmp/test-project && cd /tmp/test-project
/path/to/atv-installer init
```

## Limitations & Considerations

- **Token-heavy pipeline** — running 5+ parallel agents in a long session can hit context limits
- **TUI requires proper terminal** — `--guided` mode needs Windows Terminal / iTerm / real TTY (not VS Code integrated terminal)
- **Assumes CLI tools** — MCP servers need Node.js (`npx`), GitHub operations need `gh` CLI
- **Opinionated docs structure** — creates `docs/plans/`, `docs/brainstorms/`, `docs/solutions/`
- **Originally Claude Code** — some skill patterns (Task tool, Bash commands) are Claude Code idioms that map approximately to Copilot

## License

MIT
