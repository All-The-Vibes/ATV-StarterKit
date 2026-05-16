<p align="center">
       <img src="./assets/hero-retro.svg" alt="ATV — All The Vibes 2.0 Starter Kit" width="100%" />
</p>

<h1 align="center">ATV — All The Vibes 2.0 Starter Kit</h1>

<p align="center"><strong>One command. Full agentic coding setup. Maximum tasteful chaos.</strong></p>

<p align="center">
       <a href="https://blazingbeard.github.io/quests/atv-starterkit.html"><img src="https://img.shields.io/badge/🎮_New%3F_Start_the_Guided_Training_Quest-ff8c00?style=for-the-badge" alt="Start ATV Quest"></a>
</p>

<p align="center">
       <a href="#quick-start">Quick start</a> ·
       <a href="#install--macos--linux">Install (macOS/Linux)</a> ·
       <a href="#install--windows">Install (Windows)</a> ·
       <a href="docs/marketplace.md">Marketplace</a> ·
       <a href="#the-full-sprint">Full sprint</a> ·
       <a href="DOCS.md">Deeper docs</a> ·
       <a href="https://blazingbeard.github.io/quests/atv-starterkit.html">🎮 Training Quest</a>
</p>

https://github.com/user-attachments/assets/7b6bf18a-2bab-482b-a72d-fac9ab7436c2

---

## What is ATV 2.0?

ATV 2.0 is a one-command installer that wires together four open-source pillars on a behavioral foundation, into a single coherent agentic coding environment for GitHub Copilot:

**Behavioral foundation:**

- **[Karpathy Guidelines](https://github.com/forrestchang/andrej-karpathy-skills)** — behavioral guardrails grounded in [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876) on LLM coding pitfalls: think before coding, simplicity first, surgical changes, goal-driven execution

**Four pillars:**

- **[Autoresearch](https://github.com/github/awesome-copilot/blob/main/skills/autoresearch/SKILL.md)** — autonomous metric-driven experiment loop on a dedicated branch
- **[Compound Engineering](https://github.com/EveryInc/compound-engineering-plugin)** — planning-to-knowledge pipeline
- **[gstack](https://github.com/garrytan/gstack)** — sprint execution engine (by Garry Tan / Y Combinator)
- **[agent-browser](https://github.com/vercel-labs/agent-browser)** — browser automation layer (by Vercel)

Together they cover the full software lifecycle — from "what should I build?" through "is it healthy in production?" — with 45+ skills, 51 agents (29 featured below), and a learning system that makes your repo smarter with every session. **For pillar deep dives, the learning pipeline, agent inventory, and architecture diagrams, see [`DOCS.md`](DOCS.md).**

---

## Quick Start

Install for your project (see the OS-specific sections below for the exact commands), then open **Copilot Chat** (⌃⌘I / Ctrl+Shift+I) and run:

```text
/ce-brainstorm   →  Explore the problem, produce a design doc
/ce-plan         →  Generate an implementation plan with acceptance criteria
/ce-work         →  Build against the plan with incremental commits
/ce-review       →  Multi-agent code review (security, architecture, performance)
/atv-security    →  Config + OWASP + STRIDE security audit in one pass
/ce-compound     →  Document what you learned for future sessions

/lfg             →  Run the full pipeline in one shot
/autoresearch    →  Hill-climb autonomously against a measurable metric
```

---

## Installation

ATV ships in **three flavours** — pick whichever matches your need (or combine them):

| | `npx atv-starterkit init` | VS Code source install | Copilot CLI marketplace |
|---|---|---|---|
| **Files land in** | Your project's `.github/`, `.vscode/`, `docs/` | VS Code AgentPlugin directory | `~/.copilot/installed-plugins/` |
| **Scope** | Project-level, committed, team-shared | Personal, editor-level | Personal, follows you across CLI projects |
| **What ships** | Skills + agents + MCP + hooks + instructions + setup-steps + docs | One complete ATV skills + agents bundle | Skills + agents only |
| **Best for** | Bootstrapping a new repo, codifying team workflow | VS Code Copilot users who want one obvious install choice | CLI users who want bundles or granular skills |

The VS Code source-install path gives one complete ATV option. The Copilot CLI marketplace keeps category bundles and per-skill plugins for CLI users. Both personal paths can coexist with the project scaffold. For MCP config, hooks, instructions templates, and docs scaffolding use the npm/binary path.

### Install — macOS / Linux

Run from your shell of choice (bash, zsh, fish):

```bash
cd your-project

# Project install (team-shared, recommended)
npx atv-starterkit@latest init           # auto-detect stack, install everything
npx atv-starterkit@latest init --guided  # interactive TUI with multi-stack selection

# Or install globally so `atv-starterkit` is on your PATH
npm install -g atv-starterkit
atv-starterkit init

# Personal install via Copilot CLI marketplace (cross-project)
copilot plugin marketplace add All-The-Vibes/ATV-StarterKit
copilot plugin install atv-starter-kit@atv-starter-kit
```

For the VS Code source-install path: open the Command Palette → `Chat: Install Plugin from source` → enter `All-The-Vibes/ATV-StarterKit` → choose `atv-starter-kit`.

The npm package downloads the correct platform binary from [GitHub Releases](https://github.com/All-The-Vibes/ATV-StarterKit/releases) — no Go toolchain needed. You can also grab a pre-built binary directly for macOS or Linux (amd64/arm64) from the [latest release](https://github.com/All-The-Vibes/ATV-StarterKit/releases/latest), or build from source:

```bash
git clone https://github.com/All-The-Vibes/ATV-StarterKit.git
cd ATV-StarterKit && go build -o atv-installer .
```

### Install — Windows

Run from PowerShell or `cmd.exe`. The commands are the same as macOS/Linux — only the shell context differs:

```powershell
cd your-project

# Project install (team-shared, recommended)
npx atv-starterkit@latest init
npx atv-starterkit@latest init --guided

# Or install globally
npm install -g atv-starterkit
atv-starterkit init

# Personal install via Copilot CLI marketplace
copilot plugin marketplace add All-The-Vibes/ATV-StarterKit
copilot plugin install atv-starter-kit@atv-starter-kit
```

For the VS Code source-install path: open the Command Palette → `Chat: Install Plugin from source` → enter `All-The-Vibes/ATV-StarterKit` → choose `atv-starter-kit`.

Pre-built Windows binaries (amd64/arm64) are on the [latest release](https://github.com/All-The-Vibes/ATV-StarterKit/releases/latest) page.

> **Windows caveat:** if the gstack sub-installer's `./setup` step fails on Windows — for example, if Git Bash isn't available or another bash/runtime build issue occurs — it falls back to `bun run gen:skill-docs`. Markdown-only gstack skills still install fine; the bash-based skill-doc generator is the only thing that degrades. Everything else — ATV scaffold, Copilot CLI marketplace, agent-browser — works the same as on macOS/Linux.

### Marketplace bundles and per-skill plugins

Or pick a category bundle / single skill — full tier breakdown in **[docs/marketplace.md](docs/marketplace.md)**:

```bash
copilot plugin install atv-pack-planning@atv-starter-kit       # one category
copilot plugin install atv-skill-autoresearch@atv-starter-kit  # one skill
copilot plugin install atv-pack-security@atv-starter-kit       # security pack
```

### Prerequisites

**Required:** Git, Node.js 16+ for the npm path.

**Optional:**
- **Bun** — for gstack browser skills (`/gstack-qa`, `/gstack-browse`, `/gstack-benchmark`)
- **GitHub PAT** — for GitHub MCP server
- **Azure CLI** — for Azure MCP server
- **Copilot CLI** — for the marketplace path (`copilot` command)

Without Bun, text-based gstack skills still work. `agent-browser` works independently of Bun.

### Uninstalling

```bash
npx atv-starterkit@latest uninstall          # remove ATV files, preserve user-modified configs
npx atv-starterkit@latest uninstall --force  # remove everything including modified files
```

Removes `.github/skills/`, `.github/agents/`, `.github/hooks/`, `.github/copilot-*` config files, `.gstack/`, `.atv/`, and empty doc directories. Files you've customized since installation are preserved by default (checksum comparison against the install manifest). `.vscode/` is never touched.

---

## The Full Sprint

Each phase has skills for it; the table shows where each lives. Slash commands run inside Copilot Chat.

<table>
       <tr>
              <td width="25%" valign="top">
                     <strong>💭 Think</strong><br />
                     <sub>Frame the problem</sub><br /><br />
                     <code>/takeoff</code> <sub>— prioritized backlog briefing to start a session</sub><br />
                     <code>/ce-brainstorm</code><br />
                     <code>/gstack-office-hours</code>
              </td>
              <td width="25%" valign="top">
                     <strong>📋 Plan</strong><br />
                     <sub>Pressure-test the approach</sub><br /><br />
                     <code>/ce-plan</code><br />
                     <code>/gstack-plan-ceo-review</code><br />
                     <code>/gstack-plan-eng-review</code><br />
                     <code>/gstack-plan-design-review</code><br />
                     <code>/gstack-autoplan</code>
              </td>
              <td width="25%" valign="top">
                     <strong>🔨 Build</strong><br />
                     <sub>Execute with momentum</sub><br /><br />
                     <code>/ce-work</code><br />
                     <code>/lfg</code><br />
                     <code>/slfg</code><br />
                     <code>/autoresearch</code> <sub>— autonomous metric loop</sub>
              </td>
              <td width="25%" valign="top">
                     <strong>👀 Review</strong><br />
                     <sub>Find what you missed</sub><br /><br />
                     <code>/ce-review</code><br />
                     <code>/gstack-review</code><br />
                     <code>/atv-security</code> <sub>— config + OWASP + STRIDE (absorbs <code>/cso</code>)</sub><br />
                     <code>/ghcp-review-resolve</code> <sub>— dual Copilot + pr-review-toolkit review, inline fix loop</sub><br />
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
                     <code>/land</code> <sub>— commit → push → PR → handoff (closes out a session, never merges)</sub><br />
                     <code>/gstack-ship</code><br />
                     <code>/gstack-land-and-deploy</code><br />
                     <code>/gstack-canary</code><br />
                     <code>/gstack-document-release</code>
              </td>
              <td width="33.33%" valign="top">
                     <strong>📊 Reflect</strong><br />
                     <sub>Compound what you learned</sub><br /><br />
                     <code>/ce-compound</code><br />
                     <code>/learn</code> · <code>/instincts</code> · <code>/evolve</code><br />
                     <code>/unslop</code><br />
                     <code>/gstack-retro</code><br />
                     <code>/atv-doctor</code> <sub>— diagnose install drift</sub><br />
                     <code>/atv-update</code> <sub>— update marketplace plugins + safe source-installed AgentPlugins</sub>
              </td>
       </tr>
</table>

### `/lfg` — full pipeline, one command

Each step must produce output before the next starts (plan file exists, plan was deepened, code was changed). Retries on failure.

```
plan → deepen → work → review → unslop → resolve → test → video → compound
  ✓       ✓       ✓
```

### `/slfg` — parallel swarm variant

Same steps. Planning is sequential; review + test + unslop run in parallel.

```
plan → deepen → work (swarm) ──→ review    ⎤              resolve → unslop fix → video → compound
                                  test     ⎥ (parallel) →
                                  unslop   ⎦
```

`unslop fix` removes AI slop after review. `compound` saves learnings for future `ce-plan` runs.

### `/autoresearch` — hill-climb against a metric

For tasks with a measurable outcome (perf, bundle size, test pass rate), `/autoresearch` runs an autonomous loop on a dedicated `autoresearch/<tag>` branch — committing each experiment, running the metric command, and keeping or reverting based on the result. Every experiment is logged to `results.tsv`.

### `/atv-security` — config + OWASP + STRIDE in one pass

A single agentic security audit covering three surfaces most other tools only do one of:

- **33 config-security rules** across 5 categories — Secrets (5), Permissions (2), Hooks (11), MCP (4), Agents & Skills (11). Taxonomy adapted from [AgentShield](https://github.com/affaan-m/agentshield).
- **OWASP Top 10 (2021)** scanned against your application source — A01-A10 with file:line evidence per finding.
- **STRIDE threat modeling** — six-class threat enumeration (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege) folded into the overall grade.

Two-tier detection: deterministic grep regex (Tier 1) plus LLM-assessed semantic checks (Tier 2, agent reasoning over the actual code). Reports persist to `docs/security/YYYY-MM-DD-security-report.md` with idempotent upserts via `<!-- atv-security -->` / `<!-- cso -->` marker blocks — re-runnable without losing the surrounding human-authored context, backwards-compatible with the former `/cso` skill folded in v2.5.9. `fix` mode auto-remediates the mechanically-safe rules (literal secrets → env references, missing hook timeouts, overbroad permissions) and advises on the rest.

```bash
/atv-security                  # full audit (config + OWASP + STRIDE), report mode (default)
/atv-security fix              # apply mechanical auto-fixes
/atv-security config           # config-only scan (skip application code)
/atv-security owasp src/api    # OWASP scan scoped to src/api/
```

Full reference: [`docs/atv-security.md`](docs/atv-security.md) — rule-by-rule detail for all 33 config rules, OWASP coverage matrix, STRIDE mapping, scoring formula, sample report layout, and fix-mode semantics.

### Session bookends — `/takeoff` and `/land`

Two skills frame every Copilot session:

- **`/takeoff`** at the start — prioritized backlog briefing: open PRs, in-flight branches, recent failed CI, todos, and the highest-value next move.
- **`/land`** at the end — commit → push → PR → handoff in one step. Never merges; landing ≠ merging.

### `/ghcp-review-resolve` — dual PR review with inline fix loop

Requests a GitHub Copilot review *and* a `pr-review-toolkit` review on the current PR, adjudicates findings with an independent subagent, posts inline comments only for verified bugs, then runs a tight fix-and-reply loop per comment (test → commit → reply on thread). Resolves threads via the GraphQL `resolveReviewThread` mutation, not just `in_reply_to`.

---

## Deeper Docs

For pillar deep dives, the learning pipeline mechanics, agent inventory, MCP server reference, install architecture, full skill reference, and limitations, see **[`DOCS.md`](DOCS.md)**.

---

<div align="center">

MIT — Built by [All The Vibes](https://github.com/All-The-Vibes)

Powered by [Autoresearch](https://github.com/github/awesome-copilot/blob/main/skills/autoresearch/SKILL.md) · [Compound Engineering](https://github.com/EveryInc/compound-engineering-plugin) · [gstack](https://github.com/garrytan/gstack) · [agent-browser](https://github.com/vercel-labs/agent-browser) — grounded in the [Karpathy Guidelines](https://github.com/forrestchang/andrej-karpathy-skills)

Special thanks to [blazingbeard](https://github.com/blazingbeard) for building out the [guided training quest](https://blazingbeard.github.io/quests/atv-starterkit.html).

</div>
