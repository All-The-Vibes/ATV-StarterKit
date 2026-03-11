# atv-starterkit

**A**gentic **T**ool & **W**orkflow — a one-click installer that scaffolds a complete GitHub Copilot agentic coding environment into any project.

## Prerequisites

- **Node.js 16+** must be installed on your machine.

## Installation

### Quick Run (no install)

```bash
npx atv-starterkit init
```

### Global Install

```bash
npm install -g atv-starterkit
atv-starterkit init
```

## Usage

### One-Click Mode (Default)

```bash
cd your-project
atv-starterkit init
```

Auto-detects your stack, installs all 6 Copilot lifecycle hooks, done in seconds.

### Guided Mode

```bash
atv-starterkit init --guided
```

Interactive TUI wizard — select your stack, toggle individual components on/off.

## What Gets Installed

| # | Hook | File |
|---|------|------|
| 1 | System Instructions | `.github/copilot-instructions.md` |
| 2 | Setup Steps | `.github/copilot-setup-steps.yml` |
| 3 | MCP Servers | `.github/copilot-mcp-config.json` |
| 4 | Skills | `.github/skills/*/SKILL.md` |
| 5 | Agents | `.github/agents/*.agent.md` |
| 6 | File Instructions | `.github/*.instructions.md` |

Plus: `.vscode/extensions.json` and `docs/` structure.

## Supported Stacks

| Stack | Detection |
|-------|-----------|
| **TypeScript** | `tsconfig.json` |
| **Python** | `pyproject.toml` / `requirements.txt` |
| **Rails** | `Gemfile` + `config/routes.rb` |
| **General** | fallback |

## How It Works

This npm package downloads the pre-built `atv-installer` binary for your platform (macOS, Linux, or Windows) from [GitHub Releases](https://github.com/All-The-Vibes/ATV-StarterKit/releases) during `npm install`. The `atv-starterkit` command then delegates to that binary.

## Alternative Installation

If you prefer not to use npm, you can also:

- Download binaries directly from [GitHub Releases](https://github.com/All-The-Vibes/ATV-StarterKit/releases)
- Build from source with Go 1.22+

See the [main repository](https://github.com/All-The-Vibes/ATV-StarterKit) for details.

## License

MIT
