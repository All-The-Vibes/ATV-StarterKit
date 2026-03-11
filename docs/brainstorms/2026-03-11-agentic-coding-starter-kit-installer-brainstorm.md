---
date: 2026-03-11
topic: atv-starter-kit
---

# ATV Starter Kit

## What We're Building

**ATV** = **A**gentic **T**ool & **W**orkflow

A new `atv-installer init` command that transforms any directory into a fully-configured ATV starter kit. The command runs an interactive guided wizard that:

1. **Detects environment** — existing repo vs empty directory, current stack
2. **Asks stack preference** — Rails, Python, TypeScript, or General
3. **Selects layers** — which components to install (skills, agents, MCP config, extensions, docs structure)
4. **Writes files** — scaffolds everything into the target directory, skipping files that already exist

### UX Philosophy: One-Click by Default, Guided When Needed

The wizard must be **frictionless**. The default path should require exactly **one command and zero questions** for most users:

```
$ atv-installer init
```

Auto-detects stack, selects all universal + stack-specific components, writes everything. Done in under 5 seconds.

**Guided mode** is opt-in for users who want to customize:

```
$ atv-installer init --guided
```

This enters the interactive TUI wizard with checkboxes, stack selection, and layer toggling.

**UX Principles:**
- **Zero questions by default** — detect everything, install everything, show a summary at the end
- **Smart defaults** — if `tsconfig.json` exists, TypeScript stack is auto-selected without asking
- **Progress feedback** — real-time file creation with checkmarks, not a blank screen
- **Idempotent re-runs** — running `init` again shows what's already installed (✓) vs what would be added (→) vs what was skipped (⏭️)
- **No destructive actions** — never overwrite, only skip or merge
- **Post-install guidance** — clear "what to do next" message with the first command to try

### Target Experience

**One-click mode (default):**

```
$ atv-installer init

🔍 Auto-detected: TypeScript project (tsconfig.json found)

  ✅ .github/copilot-instructions.md
  ✅ .github/copilot-setup-steps.yml
  ✅ .github/copilot-mcp-config.json (Context7, GitHub, Azure, Terraform)
  ✅ .github/skills/ (14 skills)
  ✅ .github/agents/ (13 agents + kieran-typescript-reviewer)
  ✅ .github/typescript.instructions.md (applyTo: **/*.ts, **/*.tsx)
  ✅ .vscode/extensions.json (5 extensions)
  ✅ docs/plans/, docs/brainstorms/, docs/solutions/

🎉 ATV Starter Kit ready!

Next steps:
  1. Open this folder in VS Code
  2. Install recommended extensions when prompted
  3. Try: /ce-brainstorm "your first feature idea"
```

**Guided mode (`--guided`):**

```
$ atv-installer init --guided
🔍 Detected: existing git repo, TypeScript project

? What's your primary stack? [TypeScript]
? Which components do you want?
  ✓ Core pipeline skills (ce-brainstorm, ce-plan, ce-work, ce-review, ce-compound)
  ✓ Orchestrators (lfg, slfg)
  ✓ Universal agents (security-sentinel, performance-oracle, code-simplicity-reviewer, architecture-strategist)
  ✓ Stack-specific agents (kieran-typescript-reviewer)
  ✓ MCP servers (GitHub, Azure, Terraform, Context7)
  ✓ VS Code extensions.json
  ✓ Copilot instructions (.github/copilot-instructions.md)
  ✓ Copilot setup steps (.github/copilot-setup-steps.yml)
  ✓ File-scoped instructions (.github/*.instructions.md)
  ✓ docs/ structure (plans, brainstorms, solutions)
  ○ Compound engineering local config

  ... (same output as above)
```

**Re-run (idempotent):**

```
$ atv-installer init

🔍 Auto-detected: TypeScript project

  ✓ .github/copilot-instructions.md (already exists)
  ✓ .github/copilot-mcp-config.json (already exists)
  → .github/skills/changelog/ (new — adding)
  ✓ .github/agents/ (13 of 14 exist)
  ...

⏭️  Skipped 18 existing files
✅ Added 2 new files

🎉 ATV Starter Kit updated!
```

## Why This Approach

**Extend existing Go CLI (Approach A):**

- Builds on the existing cobra CLI, goreleaser, CI pipeline — no new project to maintain
- Single binary, zero runtime deps, works offline
- Content embedded via Go `embed` package — version-locked to the binary release
- Guided interactive wizard via `charmbracelet/huh` or `charmbracelet/bubbletea`

**Rejected alternatives:**
- *Node.js CLI* — adds Node.js runtime dependency, separate project to maintain
- *Remote content registry* — adds network dependency, complexity for caching/versioning

## Key Decisions

- **Name**: ATV Starter Kit (Agentic Tool & Workflow)
- **Wizard name**: `atv-installer`
- **Delivery**: New `init` subcommand on the `atv-installer` Go CLI
- **Content storage**: Go `embed` package — all skills, agents, configs embedded in binary
- **Interaction model**: Guided interactive TUI wizard with sensible defaults
- **Stack support**: Rails, Python, TypeScript, General — determines which agents/skills are included
- **Idempotency**: Skip files that already exist (don't overwrite), merge JSON configs additively
- **Target repos**: Both existing repos (additive merge) and fresh/empty directories (full scaffold)
- **Content is modular**: User picks layers — not all-or-nothing

## Copilot Lifecycle Hooks

The ATV Starter Kit scaffolds **all 6 Copilot lifecycle hook types**. Each hook fires at a different moment in the Copilot experience:

| # | Hook Type | File(s) | When It Fires | ATV Generates |
|---|-----------|---------|---------------|---------------|
| 1 | **System Instructions** | `.github/copilot-instructions.md` | Every Copilot chat interaction — injected as system-level context | Stack-aware starter template (coding conventions, project structure, preferred patterns) |
| 2 | **Setup Steps** | `.github/copilot-setup-steps.yml` | When Copilot Coding Agent initializes an environment | Install deps, build project, run migrations — detected from stack |
| 3 | **MCP Servers** | `.github/copilot-mcp-config.json` | When Copilot starts — registers external tool servers | Context7, GitHub, Azure, Terraform servers pre-configured |
| 4 | **Skills** | `.github/skills/*/SKILL.md` | When a skill's description matches the user's chat request | Core pipeline + stack-specific skills |
| 5 | **Agents** | `.github/agents/*.agent.md` | When invoked by name via subagent orchestration | Universal + stack-specific reviewer agents |
| 6 | **File Instructions** | `.github/*.instructions.md` | Auto-loaded based on `applyTo` glob patterns when editing matching files | Stack-specific file instructions (e.g., `applyTo: "**/*.ts"`) |

## Component Catalog

### Universal (always installed)

| Hook Type | Files |
|-----------|-------|
| **System Instructions** | `copilot-instructions.md` — project conventions, stack info, ATV workflow overview |
| **Setup Steps** | `copilot-setup-steps.yml` — auto-detected install/build/test commands |
| **MCP Servers** | `copilot-mcp-config.json` — Context7 (SSE), GitHub (npx), Azure (npx), Terraform (npx) |
| **Core Skills** | ce-brainstorm, ce-plan, ce-work, ce-review, ce-compound, lfg, slfg, deepen-plan, setup, brainstorming, document-review |
| **Universal Agents** | security-sentinel, performance-oracle, code-simplicity-reviewer, architecture-strategist, repo-research-analyst, best-practices-researcher, framework-docs-researcher, learnings-researcher, pattern-recognition-specialist, spec-flow-analyzer, pr-comment-resolver, agent-native-reviewer |
| **VS Code** | `extensions.json` — Copilot, Copilot Chat, Bicep, Azure Tools, Terraform |
| **Docs Structure** | `docs/plans/`, `docs/brainstorms/`, `docs/solutions/` |

### Stack-Specific (conditional)

| Stack | Additional Agents | Additional Skills | File Instructions |
|-------|-------------------|-------------------|-------------------|
| **Rails** | kieran-rails-reviewer, dhh-rails-reviewer, julik-frontend-races-reviewer, data-integrity-guardian, schema-drift-detector, data-migration-expert, deployment-verification-agent, lint | dhh-rails-style, andrew-kane-gem-writer, dspy-ruby | `applyTo: "**/*.rb"`, `applyTo: "**/*.erb"` |
| **Python** | kieran-python-reviewer | (none specific) | `applyTo: "**/*.py"` |
| **TypeScript** | kieran-typescript-reviewer | (none specific) | `applyTo: "**/*.ts"`, `applyTo: "**/*.tsx"` |
| **General** | (universal only) | (universal only) | (none) |

### Optional Layers

| Layer | Description |
|-------|-------------|
| **compound-engineering.local.md** | Per-project config for review agents |
| **Browser testing** | test-browser, agent-browser, reproduce-bug skills |
| **Design sync** | design-implementation-reviewer, design-iterator, figma-design-sync agents; frontend-design skill |

## Open Questions

*None — all questions resolved during brainstorming.*

## Resolved Questions

- **Goal**: Self-contained installer anyone can take and use
- **Wizard name**: `atv-installer` (binary name)
- **Delivery method**: CLI command (`atv-installer init`)
- **Content depth**: Modular/guided — interactive prompts let user pick stack + layers
- **Target repos**: Both existing and fresh repos, auto-detect and adapt
- **Implementation approach**: Extend existing Go CLI with embedded content (Approach A)
- **Copilot hooks**: All 6 lifecycle hook types scaffolded (instructions, setup steps, MCP config, skills, agents, file instructions)
- **UX**: One-click by default (`atv-installer init` — zero questions, auto-detect everything). Guided mode opt-in via `--guided` flag
- **No duplication**: Skills, agents, and MCP config are listed once under Component Catalog (they ARE hooks 3-5)
