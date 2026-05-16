<p align="center">
       <img src="./assets/hero-retro.svg" alt="ATV - All The Vibes 2.0 Starter Kit" width="100%" />
</p>

<h1 align="center">ATV - All The Vibes 2.0 Starter Kit</h1>

<p align="center"><strong>One command. Full agentic coding setup. Maximum tasteful chaos.</strong></p>

<p align="center">
       <a href="https://blazingbeard.github.io/quests/atv-starterkit.html"><img src="https://img.shields.io/badge/🎮_New%3F_Start_the_Guided_Training_Quest-ff8c00?style=for-the-badge" alt="Start ATV Quest"></a>
</p>

<p align="center">
       <a href="#installation">Installation</a> ·
       <a href="#quick-start">Quick start</a> ·
       <a href="#what-is-atv-20">What is ATV?</a> ·
       <a href="#the-sprint-skill-map">Sprint skill map</a> ·
       <a href="docs/marketplace.md">Marketplace</a> ·
       <a href="DOCS.md">Deeper docs</a> ·
       <a href="https://blazingbeard.github.io/quests/atv-starterkit.html">🎮 Training Quest</a>
</p>

https://github.com/user-attachments/assets/7b6bf18a-2bab-482b-a72d-fac9ab7436c2

---

## Installation

ATV has three install paths. Use the project install when you want the workflow committed into a repo. Use the VS Code or Copilot CLI paths when you want a personal install that follows you across projects.

| | Project install | VS Code source install | Copilot CLI marketplace |
|---|---|---|---|
| **Command** | `npx atv-starterkit@latest init` | `Chat: Install Plugin from source` | `copilot plugin install ...@atv-starter-kit` |
| **Files land in** | Your project's `.github/`, `.vscode/`, `docs/`, `.atv/` | VS Code AgentPlugin directory | `~/.copilot/installed-plugins/` |
| **Scope** | Team-shared and committed | Personal editor install | Personal CLI install |
| **Ships** | Skills, prompt shims, agents, MCP config, hooks, instructions, setup steps, docs; guided Full can add gstack + agent-browser | 30 ATV skills + 51 agents from `atv-everything` | ATV skills + agents only, no gstack/runtime scaffold |
| **Best for** | New repo bootstrap or team workflow | VS Code Copilot Chat users | CLI users who want bundles or single skills |

### macOS / Linux

```bash
cd your-project

# Project install, recommended for teams
npx atv-starterkit@latest init
npx atv-starterkit@latest init --guided

# Or install globally
npm install -g atv-starterkit
atv-starterkit init

# Personal install via Copilot CLI marketplace
copilot plugin marketplace add All-The-Vibes/ATV-StarterKit
copilot plugin install atv-everything@atv-starter-kit
```

For the VS Code source-install path: open the Command Palette, run `Chat: Install Plugin from source`, enter `All-The-Vibes/ATV-StarterKit`, then choose `atv-starter-kit`.

The npm package downloads the correct platform binary from [GitHub Releases](https://github.com/All-The-Vibes/ATV-StarterKit/releases), so you do not need Go. You can also build from source:

```bash
git clone https://github.com/All-The-Vibes/ATV-StarterKit.git
cd ATV-StarterKit
go build -o atv-starterkit .
```

### Windows

Run from PowerShell or `cmd.exe`. The project and marketplace commands are the same:

```powershell
cd your-project

# Project install, recommended for teams
npx atv-starterkit@latest init
npx atv-starterkit@latest init --guided

# Or install globally
npm install -g atv-starterkit
atv-starterkit init

# Personal install via Copilot CLI marketplace
copilot plugin marketplace add All-The-Vibes/ATV-StarterKit
copilot plugin install atv-everything@atv-starter-kit
```

For VS Code: Command Palette → `Chat: Install Plugin from source` → `All-The-Vibes/ATV-StarterKit` → `atv-starter-kit`.

Pre-built Windows binaries are on the [latest release](https://github.com/All-The-Vibes/ATV-StarterKit/releases/latest) page.

> **Windows caveat:** if the gstack sub-installer's `./setup` step fails because Git Bash or another bash/runtime dependency is missing, ATV falls back to `bun run gen:skill-docs`. Markdown-only skills still install. The bash-based gstack skill-doc generator is the only degraded piece.

### Marketplace bundles and single-skill installs

Install everything:

```bash
copilot plugin marketplace add All-The-Vibes/ATV-StarterKit
copilot plugin install atv-everything@atv-starter-kit
```

Or install one pack or one skill:

```bash
copilot plugin install atv-pack-planning@atv-starter-kit
copilot plugin install atv-pack-security@atv-starter-kit
copilot plugin install atv-skill-atv-security@atv-starter-kit
copilot plugin install atv-skill-autoresearch@atv-starter-kit
```

> **Heads up:** `atv-pack-*` and `atv-skill-*` installs include skills only. If you install a pack or single skill that dispatches reviewer/research agents, also install `atv-agents@atv-starter-kit`. Full tier breakdown: **[docs/marketplace.md](docs/marketplace.md)**.

### Prerequisites

**Required:** Git, Node.js 16+ for the npm path.

**Optional:**

- **Bun** for gstack browser skills and skill-doc generation.
- **GitHub PAT** for GitHub MCP server workflows.
- **Azure CLI** for Azure MCP server workflows.
- **Copilot CLI** for marketplace installs.

### Uninstall

```bash
npx atv-starterkit@latest uninstall
npx atv-starterkit@latest uninstall --force
```

The default uninstall preserves user-modified files. `--force` removes ATV files even if they drifted.

---

## Quick Start

After install, open **Copilot Chat** (`⌃⌘I` / `Ctrl+Shift+I`) and run the workflow like a sprint:

```text
/ce-brainstorm   →  Explore the problem and write a design doc
/ce-plan         →  Turn the design into an implementation plan
/ce-work         →  Build against the plan
/ce-review       →  Run multi-agent code review
/atv-security    →  Run config + OWASP + STRIDE security review
/ce-compound     →  Save what you learned for next time

/lfg             →  Run the full pipeline in one shot
/autoresearch    →  Hill-climb autonomously against a measurable metric
```

Project installs write **30 embedded ATV skills**, **18 VS Code prompt shims** for the main slash commands, and **51 reviewer/specialist agents** into the repo. VS Code source installs and Copilot CLI marketplace installs provide the ATV skills + agents bundle without repo-local `.github/prompts/` shims. Guided **Full** project installs can also add **25 gstack sprint skills** plus `agent-browser`.

---

## What is ATV 2.0?

ATV 2.0 is a one-command installer that wires together four open-source pillars on a behavioral foundation, into a single agentic coding environment for GitHub Copilot.

**Behavioral foundation:**

- **[Karpathy Guidelines](https://github.com/forrestchang/andrej-karpathy-skills)** - behavioral guardrails grounded in [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876) on LLM coding pitfalls: think before coding, simplicity first, surgical changes, goal-driven execution.

**Four pillars:**

- **[Autoresearch](https://github.com/github/awesome-copilot/blob/main/skills/autoresearch/SKILL.md)** - autonomous metric-driven experiment loops on dedicated branches.
- **[Compound Engineering](https://github.com/EveryInc/compound-engineering-plugin)** - brainstorm, plan, work, review, and compound knowledge.
- **[gstack](https://github.com/garrytan/gstack)** - sprint execution patterns, browser QA, shipping discipline, and review posture.
- **[agent-browser](https://github.com/vercel-labs/agent-browser)** - browser automation layer.

For pillar deep dives, learning mechanics, agent inventory, MCP details, and architecture diagrams, see **[`DOCS.md`](DOCS.md)**.

---

## The Sprint Skill Map

This first table is the **30 embedded ATV skills** from `plugins/atv-everything` and `pkg/scaffold/templates/skills`. Every skill also has a granular marketplace plugin named `atv-skill-<kebab-cased-skill-name>`; for example, `/resolve_todo_parallel` installs as `atv-skill-resolve-todo-parallel`. Category packs install the groups shown below. If you choose the guided **Full** project install, ATV can also sync the gstack add-on skills in the next table.

| Sprint phase | Skill | Pack | What it does |
|---|---|---|---|
| Think | `/takeoff` | `atv-pack-shipping` | Starts a session with prioritized work, blockers, active branches, and the best next move. |
| Think | `/brainstorming` | `atv-pack-planning` | Helper for exploring user intent, design decisions, and approaches before planning. |
| Think | `/ce-brainstorm` | `atv-pack-planning` | Turns a vague feature/problem into a right-sized requirements or brainstorm doc. |
| Think | `/ce-ideate` | `atv-pack-planning` | Generates and critiques grounded improvement ideas before you commit to a direction. |
| Think | `/karpathy-guidelines` | `atv-pack-guidelines` | Behavioral guardrails for simpler, more surgical LLM-assisted code changes. |
| Plan | `/ce-plan` | `atv-pack-planning` | Writes implementation plans with repo research, acceptance criteria, and test strategy. |
| Plan | `/deepen-plan` | `atv-pack-planning` | Helper that deepens an existing plan with parallel research and implementation detail. |
| Plan | `/document-review` | `atv-pack-review` | Reviews requirements or plan docs with specialist reviewers before code starts. |
| Build | `/ce-work` | `atv-pack-shipping` | Executes planned work while preserving repo patterns and quality gates. |
| Build | `/autoresearch` | `atv-pack-guidelines` | Runs autonomous experiment loops when success has a measurable metric. |
| Build | `/ralph-loop` | `atv-pack-quality` | Iterative autonomous task loop with fresh context, filesystem memory, and git versioning. |
| Build | `/resolve_todo_parallel` | `atv-everything` / single skill | Resolves pending CLI todos in parallel when a branch has many small follow-ups. |
| Review | `/ce-review` | `atv-pack-review` | Multi-agent code review with security, performance, architecture, and language reviewers. |
| Review | `/atv-security` | `atv-pack-security` | Default ATV security audit for agentic config, OWASP Top 10 checks, and STRIDE threat modeling. Guided Full can also add gstack `/gstack-cso`. |
| Review | `/unslop` | `atv-pack-quality` | De-slops code: simplifies, detects comment rot, and catches design slop before PR. |
| Test / demo | `/test-browser` | `atv-everything` / single skill | Browser test pass for pages affected by the current PR or branch. |
| Test / demo | `/feature-video` | `atv-everything` / single skill | Captures a visual walkthrough and adds it to PR context. |
| Ship | `/land` | `atv-pack-shipping` | Ends a session with commit, push, PR, and handoff. Landing does not mean merging. |
| Ship | `/lfg` | `atv-pack-shipping` | Full autonomous pipeline: plan → deepen → work → review → unslop → resolve → test → video → compound. |
| Ship | `/slfg` | `atv-pack-shipping` | Swarm variant of `/lfg`, with parallel review, test, and unslop stages. |
| Reflect | `/ce-compound` | `atv-pack-shipping` | Documents solved problems into `docs/solutions/` so future sessions reuse the learning. |
| Reflect | `/ce-compound-refresh` | `atv-pack-shipping` | Refreshes stale or drifted learnings against the current codebase. |
| Reflect | `/learn` | `atv-pack-learning` | Extracts reusable patterns from recent work into project instincts. |
| Reflect | `/instincts` | `atv-pack-learning` | Shows learned instincts with confidence scores, grouped by domain. |
| Reflect | `/observe` | `atv-pack-learning` | Runs a focused observation session over a domain or file pattern. |
| Reflect | `/evolve` | `atv-pack-learning` | Promotes mature instincts into full Copilot skills. |
| Maintain | `/atv-doctor` | `atv-pack-maintenance` | Diagnoses project scaffold, marketplace plugin, and VS Code source-install drift. |
| Maintain | `/atv-update` | `atv-pack-maintenance` | Updates marketplace plugins and safe source-installed AgentPlugins. Project scaffold refresh is advisory. |
| Maintain | `/setup` | `atv-everything` / single skill | Placeholder/project setup helper for compound-engineering workflow configuration. |
| Optional | `/meme-iq` | `atv-pack-easter-eggs` | AI meme generation via memegen.link. Not on the critical path. |

### Prompt-shimmed slash commands

These 18 top-level commands are surfaced in VS Code Copilot Chat by `.github/prompts/*.prompt.md`:

```text
/atv-doctor  /atv-security  /atv-update  /autoresearch
/ce-brainstorm  /ce-compound  /ce-ideate  /ce-plan  /ce-review  /ce-work
/evolve  /instincts  /land  /learn  /lfg  /observe  /takeoff  /unslop
```

The remaining embedded ATV skills are still installed. They are helpers, internal orchestrator steps, behavioral references, or optional extras that should not clutter the slash-command picker.

### Guided Full gstack add-ons

Guided Full project installs can also clone and sync gstack. ATV writes these into `gstack-*` skill directories to avoid collisions; some clients may show the unprefixed skill name from the skill frontmatter.

| Sprint phase | gstack skill | What it adds | Runtime |
|---|---|---|---|
| Think / Plan | `/gstack-office-hours` | YC-style forcing questions before you commit to a direction. | Markdown |
| Think / Plan | `/gstack-plan-ceo-review` | Founder/CEO review for scope and product shape. | Markdown |
| Think / Plan | `/gstack-plan-eng-review` | Engineering review for architecture, data flow, edge cases, and tests. | Markdown |
| Think / Plan | `/gstack-plan-design-review` | Design-quality review before UI work starts. | Markdown |
| Think / Plan | `/gstack-design-consultation` | Design-system consultation for brand, typography, color, and layout. | Markdown |
| Think / Plan | `/gstack-autoplan` | CEO → design → engineering review sequence in one pass. | Markdown |
| Review | `/gstack-review` | Staff-level pre-landing code review. | Markdown |
| Review | `/gstack-design-review` | Visual/design audit with iterative fixes. | Markdown |
| Review | `/gstack-design-shotgun` | Multiple design variants for comparison and feedback. | Markdown |
| Review | `/gstack-codex` | Independent OpenAI Codex review. | Markdown |
| Review | `/gstack-cso` | gstack security review. Use `/atv-security` as ATV's default config + OWASP + STRIDE audit. | Markdown |
| Test / QA | `/gstack-qa` | Browser QA loop that finds, fixes, and verifies bugs. | Bun/browser |
| Test / QA | `/gstack-qa-only` | Browser QA report without fixes. | Bun/browser |
| Test / QA | `/gstack-benchmark` | Page speed, Core Web Vitals, and resource-size baselines. | Bun/browser |
| Test / QA | `/gstack-browse` | Persistent browser runtime for deeper dogfooding. | Bun/browser |
| Ship / Deploy | `/gstack-ship` | Sync, test, review, push, and open PR. | Markdown |
| Ship / Deploy | `/gstack-land-and-deploy` | Merge, wait for CI/deploy, and verify production. | Markdown |
| Ship / Deploy | `/gstack-canary` | Post-deploy monitoring for errors and regressions. | Markdown |
| Ship / Deploy | `/gstack-document-release` | Update docs to match what shipped. | Markdown |
| Safety | `/gstack-careful` | Warn before destructive commands. | Markdown |
| Safety | `/gstack-freeze` | Restrict edits to one directory. | Markdown |
| Safety | `/gstack-guard` | Careful + Freeze together. | Markdown |
| Safety | `/gstack-unfreeze` | Remove the freeze boundary. | Markdown |
| Debug | `/gstack-investigate` | Root-cause investigation before fixes. | Markdown |
| Reflect | `/gstack-retro` | Weekly/team retrospective with trends and follow-ups. | Markdown |

---

## Key workflows

### `/lfg` - full pipeline, one command

Each step must produce output before the next starts. Retries on failure.

```text
plan → deepen → work → review → unslop → resolve → test → video → compound
```

### `/slfg` - parallel swarm variant

Planning is sequential. Review, test, and unslop run in parallel.

```text
plan → deepen → work (swarm) ──→ review    ⎤
                                  test      ⎥  parallel → resolve → unslop fix → video → compound
                                  unslop    ⎦
```

### `/autoresearch` - hill-climb against a metric

For tasks with a measurable outcome, `/autoresearch` runs a loop on a dedicated `autoresearch/<tag>` branch, commits each experiment, runs the metric command, and keeps or reverts based on the result. Every experiment is logged to `results.tsv`.

### `/atv-security` - security audit in one pass

```bash
/atv-security                  # full audit, report mode
/atv-security fix              # full audit, safe opt-in fixes
/atv-security config           # config-only scan
/atv-security owasp src/api    # OWASP scan scoped to src/api
```

`/atv-security` scans `.github/` and `.vscode/` config with AgentShield-style rules, then scans application source for OWASP Top 10 and STRIDE risks. Use it as ATV's default security pass; guided Full can also add gstack `/gstack-cso` for teams that want that extra role.

### `/takeoff` and `/land` - session bookends

- **`/takeoff`** starts work with a prioritized briefing: open PRs, in-flight branches, failed CI, todos, blockers, and the recommended next move.
- **`/land`** closes work with quality gates, commit, push, PR, and handoff. It never merges unless you explicitly ask.

---

## Deeper Docs

For pillar deep dives, learning pipeline mechanics, agent inventory, MCP server reference, install architecture, full skill details, and limitations, see **[`DOCS.md`](DOCS.md)**.

---

<div align="center">

MIT - Built by [All The Vibes](https://github.com/All-The-Vibes)

Powered by [Autoresearch](https://github.com/github/awesome-copilot/blob/main/skills/autoresearch/SKILL.md) · [Compound Engineering](https://github.com/EveryInc/compound-engineering-plugin) · [gstack](https://github.com/garrytan/gstack) · [agent-browser](https://github.com/vercel-labs/agent-browser) - grounded in the [Karpathy Guidelines](https://github.com/forrestchang/andrej-karpathy-skills)

Special thanks to [blazingbeard](https://github.com/blazingbeard) for building out the [guided training quest](https://blazingbeard.github.io/quests/atv-starterkit.html).

</div>
