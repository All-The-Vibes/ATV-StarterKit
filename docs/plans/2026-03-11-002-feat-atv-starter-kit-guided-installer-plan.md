---
title: "feat: ATV Starter Kit — guided CLI installer for agentic coding"
type: feat
status: active
date: 2026-03-11
origin: docs/brainstorms/2026-03-11-agentic-coding-starter-kit-installer-brainstorm.md
---

# feat: ATV Starter Kit — Guided CLI Installer

## Overview

Build `atv-installer init` — a Go CLI command that scaffolds a complete GitHub Copilot agentic coding environment into any directory. One command, zero questions by default, all 6 Copilot lifecycle hooks configured. (see brainstorm: docs/brainstorms/2026-03-11-agentic-coding-starter-kit-installer-brainstorm.md)

## Problem Statement

Setting up an agentic coding environment today requires manually creating 30+ files across 6 different Copilot hook types, knowing which MCP servers to configure, which agents to install, and how skills work. This is a barrier to adoption — most developers never get past "install Copilot extension."

## Proposed Solution

A single Go binary (`atv-installer`) with an `init` subcommand that:

1. Auto-detects the target environment (existing repo vs empty dir, stack type)
2. Scaffolds all 6 Copilot lifecycle hooks with stack-appropriate content
3. Runs in one-click mode by default, guided mode via `--guided` flag
4. Is idempotent — re-running skips existing files, adds new ones

## Technical Approach

### Architecture

```
atv-installer/
├── cmd/
│   ├── root.go          # cobra root command
│   └── init.go          # `init` subcommand — wizard entry point
├── pkg/
│   ├── detect/
│   │   └── detect.go    # Stack detection (Rails/Python/TS/General)
│   ├── scaffold/
│   │   ├── scaffold.go  # Core file-writing logic (idempotent)
│   │   ├── hooks.go     # Copilot lifecycle hooks registry
│   │   └── catalog.go   # Component catalog (universal + stack-specific)
│   ├── tui/
│   │   └── wizard.go    # Interactive guided mode (charmbracelet/huh)
│   └── output/
│       └── printer.go   # Colored output with ✅/⏭️/→ indicators
├── templates/            # Go embed source directory
│   ├── skills/           # All SKILL.md files
│   ├── agents/           # All .agent.md files
│   ├── configs/          # copilot-mcp-config.json, extensions.json templates
│   ├── instructions/     # copilot-instructions.md templates per stack
│   ├── setup-steps/      # copilot-setup-steps.yml templates per stack
│   └── file-instructions/ # *.instructions.md with applyTo globs per stack
└── go.mod
```

### Implementation Phases

#### Phase 1: Foundation — CLI + Detection + Flat Scaffold

- [x] Create `cmd/init.go` with cobra subcommand registration
- [x] Implement `pkg/detect/detect.go` — stack detection logic:
  - Check for `Gemfile` + `config/routes.rb` → Rails
  - Check for `Gemfile` → Ruby
  - Check for `tsconfig.json` → TypeScript
  - Check for `package.json` → JavaScript/TypeScript
  - Check for `pyproject.toml` or `requirements.txt` → Python
  - Check for `.git` → existing repo flag
  - Fallback → General
- [x] Implement `pkg/scaffold/scaffold.go` — core file writer:
  - `WriteFile(path, content) → (created|skipped|merged)`
  - Check if file exists before writing (idempotent)
  - Create parent directories as needed
  - Return status enum for output display
- [x] Implement `pkg/output/printer.go` — colored terminal output:
  - `✅` for created files
  - `⏭️` for skipped (already exists)
  - `→` for newly added on re-run
  - Summary counts at the end
  - Post-install "Next steps" message

#### Phase 2: Content — Embed Templates + Catalog

- [x] Create `templates/` directory tree with all embedded content:
  - Copy 11 core skill SKILL.md files into `templates/skills/`
  - Copy 12 universal agent .agent.md files into `templates/agents/`
  - Create `templates/configs/copilot-mcp-config.json` (Context7, GitHub, Azure, Terraform)
  - Create `templates/configs/extensions.json` (5 recommended extensions)
  - Create `templates/instructions/rails.md`, `python.md`, `typescript.md`, `general.md`
  - Create `templates/setup-steps/rails.yml`, `python.yml`, `typescript.yml`, `general.yml`
  - Create `templates/file-instructions/rails.instructions.md` (`applyTo: "**/*.rb"`)
  - Create `templates/file-instructions/python.instructions.md` (`applyTo: "**/*.py"`)
  - Create `templates/file-instructions/typescript.instructions.md` (`applyTo: "**/*.ts"`)
- [x] Implement `pkg/scaffold/catalog.go` — component registry:
  - `UniversalComponents()` → list of all always-installed files
  - `StackComponents(stack)` → additional files for detected stack
  - `OptionalComponents()` → items only in guided mode
- [x] Implement `pkg/scaffold/hooks.go` — Copilot hooks mapper:
  - Map each component to its hook type (1-6)
  - Group output by hook type for display
- [x] Add `//go:embed templates/*` directive in scaffold package
- [x] Wire up one-click flow: detect → catalog → write → print summary

#### Phase 3: Guided Mode — Interactive TUI

- [x] Add `charmbracelet/huh` dependency to `go.mod`
- [x] Implement `pkg/tui/wizard.go` — guided wizard:
  - Step 1: Confirm or override detected stack (select input)
  - Step 2: Multi-select checkboxes for component layers (all checked by default)
  - Step 3: Confirmation screen showing what will be written
- [x] Wire `--guided` flag on `init` command
- [x] If `--guided`: run wizard → get selections → scaffold selected components
- [x] If not `--guided`: auto-detect → scaffold all → print summary

#### Phase 4: Idempotent Re-run + JSON Merge

- [x] Implement JSON merge for `copilot-mcp-config.json`:
  - Read existing file, parse JSON
  - Add missing `mcpServers` entries (don't overwrite existing)
  - Write merged result
- [x] Implement JSON merge for `extensions.json`:
  - Read existing `recommendations` array
  - Append missing extension IDs
  - Write merged result
- [x] Add re-run detection to printer:
  - On re-run, show `✓ (already exists)` for existing files
  - Show `→ (new — adding)` for new files
  - Show summary: "Skipped N existing, Added M new"

#### Phase 5: Build + Release

- [x] Add goreleaser config for `atv-installer` binary
- [x] Update README.md with install instructions
- [x] Add CI workflow for build + test
- [ ] Tag v0.1.0 release

## Acceptance Criteria

### Functional Requirements

- [ ] `atv-installer init` runs in one-click mode with zero prompts
- [ ] `atv-installer init --guided` shows interactive TUI wizard
- [ ] Stack auto-detection works for Rails, Python, TypeScript, General
- [ ] All 6 Copilot lifecycle hooks scaffolded correctly
- [ ] Re-running is idempotent — existing files skipped, new files added
- [ ] JSON configs (MCP, extensions) are merged additively on re-run
- [ ] Works on both existing repos and empty directories
- [ ] Binary runs on Windows, macOS, Linux (cross-compiled via goreleaser)

### Non-Functional Requirements

- [ ] Init completes in under 5 seconds (local file I/O only)
- [ ] Binary size under 20MB (embedded content is markdown — small)
- [ ] No network access required (all content embedded)
- [ ] No runtime dependencies (single static binary)

### Quality Gates

- [ ] Unit tests for detect, scaffold, catalog, and merge logic
- [ ] Integration test: run init in temp dir → verify all files created
- [ ] Integration test: run init twice → verify idempotency (no overwrites)
- [ ] Integration test: run init in repo with existing MCP config → verify merge

## Dependencies & Risks

| Risk | Mitigation |
|------|------------|
| Embedded content inflates binary size | Markdown files are tiny — 47 skills + 28 agents ≈ 1-2MB compressed |
| `charmbracelet/huh` adds dependency weight | Only imported for guided mode; core flow works without TUI |
| Content gets outdated | Version-locked to binary release; goreleaser publishes binary + content together |
| Copilot hook file format changes | Monitor GitHub Copilot docs; templates are markdown — easy to update |

## Sources & References

### Origin

- **Brainstorm:** [docs/brainstorms/2026-03-11-agentic-coding-starter-kit-installer-brainstorm.md](docs/brainstorms/2026-03-11-agentic-coding-starter-kit-installer-brainstorm.md)
  - Key decisions: Go CLI with embedded content (Approach A), one-click default + guided opt-in, all 6 Copilot hooks, `atv-installer` as binary name

### Internal References

- Existing cobra CLI pattern: `cmd/root.go`, `cmd/install.go`
- Current MCP config: `.github/copilot-mcp-config.json`
- Current extensions config: `.vscode/extensions.json`
- Agent frontmatter pattern: `.github/agents/*.agent.md` (description-only after fix)
- Skill structure: `.github/skills/*/SKILL.md`

### External References

- [charmbracelet/huh](https://github.com/charmbracelet/huh) — Go TUI form library
- [Go embed package](https://pkg.go.dev/embed) — compile-time file embedding
- [goreleaser](https://goreleaser.com/) — cross-platform Go binary releases
- [Copilot customization docs](https://docs.github.com/en/copilot/customizing-copilot)
