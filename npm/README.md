<p align="center">
	<img src="https://raw.githubusercontent.com/All-The-Vibes/ATV-StarterKit/main/assets/hero-retro.svg" alt="ATV — All The Vibes 2.0 Starter Kit" width="100%" />
</p>

# ATV — All The Vibes 2.0 Starter Kit

<p align="center"><strong>Install a complete GitHub Copilot agentic coding stack into any repo with one command.</strong></p>

<p align="center">Compound Engineering for planning and review. gstack for QA, shipping, safety, and browser-based testing. Vercel <code>agent-browser</code> for direct browser automation.</p>

<p align="center">
	<a href="https://www.npmjs.com/package/atv-starterkit"><img alt="npm version" src="https://img.shields.io/npm/v/atv-starterkit?style=flat-square&logo=npm&logoColor=white&color=cb3837"></a>
	<a href="https://opensource.org/licenses/MIT"><img alt="MIT License" src="https://img.shields.io/badge/License-MIT-ffd700?style=flat-square"></a>
	<a href="https://github.com/features/copilot"><img alt="GitHub Copilot Ready" src="https://img.shields.io/badge/GitHub%20Copilot-Ready-8957e5?style=flat-square&logo=github"></a>
</p>

`ATV` is the short name. The brand is **All The Vibes 2.0 Starter Kit**. The package name stays `atv-starterkit`.

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

## Why people install it

- **One command setup** for Copilot instructions, skills, agents, MCP config, and stack-specific guidance
- **Better planning before coding** with Compound Engineering workflows
- **Real browser QA** through gstack runtime support
- **Direct browser automation** through optional `agent-browser` installation in guided mode
- **File-based knowledge compounding** via `docs/brainstorms/`, `docs/plans/`, and `docs/solutions/`

## `agent-browser` support

ATV can also install [Vercel `agent-browser`](https://github.com/vercel-labs/agent-browser), a browser automation CLI for AI agents.

Use it when you want direct control over the browser for:

- screenshots
- filling forms
- clicking through flows
- scraping content
- checking console or network issues

Best workflow:

```bash
agent-browser open https://example.com
agent-browser snapshot -i --json
agent-browser click @e1
agent-browser fill @e2 "hello"
```

Use gstack's browser skills when you want a higher-level QA workflow; use `agent-browser` when you want precise, one-off browser control.

## What Gets Installed

| # | Hook | File |
|---|------|------|
| 1 | System Instructions | `.github/copilot-instructions.md` |
| 2 | Setup Steps | `.github/copilot-setup-steps.yml` |
| 3 | MCP Servers | `.github/copilot-mcp-config.json` |
| 4 | Skills | `.github/skills/*/SKILL.md` |
| 5 | Agents | `.github/agents/*.agent.md` |
| 6 | File Instructions | `.github/*.instructions.md` |
| 7 | Prompt Shims | `.github/prompts/*.prompt.md` (VS Code Copilot Chat slash commands) |

Plus: `.vscode/extensions.json` and `docs/` structure.

## Supported Stacks

| Stack | Detection |
|-------|-----------|
| **TypeScript** | `tsconfig.json` |
| **Python** | `pyproject.toml` / `requirements.txt` |
| **Rails** | `Gemfile` + `config/routes.rb` |
| **General** | fallback |

## How It Works

This npm package downloads the pre-built `atv-inst-tool` binary for your platform (macOS, Linux, or Windows) from [GitHub Releases](https://github.com/All-The-Vibes/ATV-StarterKit/releases) during `npm install`. The `atv-starterkit` command then delegates to that binary.

## Alternative Installation

If you prefer not to use npm, you can also:

- Download binaries directly from [GitHub Releases](https://github.com/All-The-Vibes/ATV-StarterKit/releases)
- Build from source with Go 1.22+

See the [main repository](https://github.com/All-The-Vibes/ATV-StarterKit) for details.

## License

MIT
