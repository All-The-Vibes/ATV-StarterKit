# ATV 2.0 — Deeper Documentation

This file holds the deep-dive material that used to live in `README.md`. The README now focuses on what ATV is, how to install it on macOS/Linux/Windows, and the full sprint of where each skill fits. Everything else — pillar internals, learning pipeline mechanics, agent inventory, install architecture, and the full skill reference — lives here.

[← Back to README](README.md)

---

## The Behavioral Foundation

### Karpathy Guidelines — the behavioral foundation

Every skill and agent in ATV operates under four principles derived from [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876) on how LLMs fail at coding. These are installed as a skill (`.github/skills/karpathy-guidelines/SKILL.md`) and shape how Copilot approaches all work:

| Principle | What it prevents |
|---|---|
| **Think Before Coding** | Wrong assumptions, hidden confusion, silently picking one interpretation |
| **Simplicity First** | Overcomplication, bloated abstractions, speculative features |
| **Surgical Changes** | Drive-by refactoring, touching code you shouldn't, cosmetic "improvements" |
| **Goal-Driven Execution** | Vague success criteria, no verification loop, "make it work" without checking |

These aren't just instructions — they're the operating contract between you and the AI. Without them, Copilot tends toward the exact pitfalls Karpathy described: "The models make wrong assumptions on your behalf and just run along with them."

## The Four Pillars

### Autoresearch — autonomous experimentation loop

For tasks with a measurable metric — performance tuning, test pass rate, bundle size, latency, build time — `/autoresearch` runs an autonomous loop: define the goal + metric + scope, the agent works on a dedicated `autoresearch/<tag>` branch, committing each experiment, running the metric command, and keeping or reverting based on the result. Every experiment is logged to `results.tsv` so you can audit the research trail when the loop ends.

Installed as a skill (`.github/skills/autoresearch/SKILL.md`). Sourced from [github/awesome-copilot](https://github.com/github/awesome-copilot/blob/main/skills/autoresearch/SKILL.md) (MIT) by [@luiscantero](https://github.com/luiscantero), inspired by [Karpathy's autoresearch](https://github.com/karpathy/autoresearch).

**Use when** you have a measurable outcome and want the agent to hill-climb autonomously. **Skip for** one-shot tasks, simple bug fixes, or anything without a clear metric.

### Compound Engineering — knowledge compounds

A gated pipeline where each step produces an artifact the next step consumes:

```text
/ce-brainstorm → /ce-plan → /ce-work → /ce-review → /ce-compound
```

Every time you run `/ce-compound`, solved problems get saved to `docs/solutions/`. Next time `/ce-plan` runs, the `learnings-researcher` agent searches those files first. Your repo gets smarter with every PR.

### gstack — the AI sprint process

gstack adds sprint execution workflows for planning, review, QA, shipping, safety, debugging, and retros. It doesn't just give the AI more tools — it gives it a *role*. `/gstack-review` acts as a staff engineer, while ATV's own `/atv-security` skill carries the Review/security slot where `/gstack-cso` used to live.

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
| **Starter** | Repo-local ATV scaffold: core skills, orchestrators, agents, MCP config, instructions, setup steps, and docs. No gstack clone or browser runtime. |
| **Pro** | Starter plus text-first gstack workflows for planning, review, shipping, safety, debugging, and retros. |
| **Full** | + all 30 ATV skills, 18 prompt shims, 51 agents, browser tooling, and optional integrations. Bun recommended for gstack generation/browser workflows. |

**3. Customize** — Power users can drill into category-grouped multi-select. Beginners skip straight to install.

The customize screen exposes opt-in skill layers grouped by intent:

| Layer | Contents |
|---|---|
| **`core-skills`** | Planning, lifecycle, learning, quality, security, behavioral guidelines |
| **`orchestrators`** | LFG, SLFG, ralph-loop, feature-video, test-browser |
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

## Full Skill Reference

This is the same 30-skill surface shipped by `plugins/atv-everything` and `pkg/scaffold/templates/skills`. Prompt-shimmed skills appear in VS Code Copilot Chat via `.github/prompts/*.prompt.md`; helper skills are still installed, but intentionally kept out of the picker.

| Sprint phase | Skill | Pack | What it does |
|---|---|---|---|
| Think | `/takeoff` | `atv-pack-shipping` | Session-start briefing: open PRs, in-flight branches, failed CI, todos, blockers, and next move. |
| Think | `/brainstorming` | `atv-pack-planning` | Helper for exploring intent, approaches, and design decisions before planning. |
| Think | `/ce-brainstorm` | `atv-pack-planning` | Turns a vague feature/problem into a focused requirements or brainstorm doc. |
| Think | `/ce-ideate` | `atv-pack-planning` | Generates and critiques grounded improvement ideas. |
| Think | `/karpathy-guidelines` | `atv-pack-guidelines` | Behavioral guardrails for simpler, more surgical LLM-assisted code changes. |
| Plan | `/ce-plan` | `atv-pack-planning` | Implementation plans with repo research, acceptance criteria, and test strategy. |
| Plan | `/deepen-plan` | `atv-pack-planning` | Helper that deepens an existing plan with parallel research. |
| Plan | `/document-review` | `atv-pack-review` | Requirements/plan review with specialist reviewers before implementation. |
| Build | `/ce-work` | `atv-pack-shipping` | Executes planned work while preserving repo patterns and quality gates. |
| Build | `/autoresearch` | `atv-pack-guidelines` | Autonomous experiment loops for measurable goals. |
| Build | `/ralph-loop` | `atv-pack-quality` | Iterative autonomous task loop with fresh context, filesystem memory, and git versioning. |
| Build | `/resolve_todo_parallel` | `atv-everything` / single skill | Parallel resolution of pending CLI todos. |
| Review | `/ce-review` | `atv-pack-review` | Multi-agent code review across security, performance, architecture, and language concerns. |
| Review | `/atv-security` | `atv-pack-security` | Agentic config audit + OWASP Top 10 + STRIDE threat modeling. Occupies the same Review/security slot where `/gstack-cso` was listed. |
| Review | `/unslop` | `atv-pack-quality` | Code simplification, comment rot detection, and design slop check. |
| Test / demo | `/test-browser` | `atv-everything` / single skill | Browser tests for pages affected by the current PR or branch. |
| Test / demo | `/feature-video` | `atv-everything` / single skill | Visual walkthrough capture for PRs. |
| Ship | `/land` | `atv-pack-shipping` | Session-end handoff: quality gates, commit, push, PR, and notes. Never merges by default. |
| Ship | `/lfg` | `atv-pack-shipping` | Full pipeline: plan → deepen → work → review → unslop → resolve → test → video → compound. |
| Ship | `/slfg` | `atv-pack-shipping` | Swarm variant of `/lfg` with parallel review/test/unslop. |
| Reflect | `/ce-compound` | `atv-pack-shipping` | Writes solved-problem docs in `docs/solutions/`. |
| Reflect | `/ce-compound-refresh` | `atv-pack-shipping` | Refreshes stale or drifted learnings against the current codebase. |
| Reflect | `/learn` | `atv-pack-learning` | Extracts reusable patterns from recent work into instincts. |
| Reflect | `/instincts` | `atv-pack-learning` | Shows learned instincts with confidence scores, grouped by domain. |
| Reflect | `/observe` | `atv-pack-learning` | Focused observation over a domain or file pattern. |
| Reflect | `/evolve` | `atv-pack-learning` | Promotes mature instincts into full Copilot skills. |
| Maintain | `/atv-doctor` | `atv-pack-maintenance` | Diagnoses project scaffold, marketplace plugin, and VS Code source-install drift. |
| Maintain | `/atv-update` | `atv-pack-maintenance` | Updates marketplace plugins and safe source-installed AgentPlugins. Project scaffold refresh is advisory. |
| Maintain | `/setup` | `atv-everything` / single skill | Project setup helper for compound-engineering workflow configuration. |
| Optional | `/meme-iq` | `atv-pack-easter-eggs` | Meme generation via memegen.link. |

### Prompt-shimmed commands

These 18 skills get `.github/prompts/*.prompt.md` shims for VS Code Copilot Chat discovery:

```text
/atv-doctor  /atv-security  /atv-update  /autoresearch
/ce-brainstorm  /ce-compound  /ce-ideate  /ce-plan  /ce-review  /ce-work
/evolve  /instincts  /land  /learn  /lfg  /observe  /takeoff  /unslop
```

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

`/unslop` is wired into both autonomous pipelines — `/lfg` runs `/unslop fix` after review, and `/slfg` runs the report pass in parallel with `ce-review` and browser testing for zero added wall-clock time.

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

- **Knowledge compounding** (per-PR): `/ce-compound` saves solved problems → future `/ce-plan` finds them via `learnings-researcher` → fewer repeated mistakes
- **Pattern learning** (per-session): observer hooks → `/learn` → instincts → `/evolve` → permanent skills → Copilot knows your conventions
- **Team propagation** (per-commit): instincts are committed to git → the whole team inherits learned patterns without a style guide

Over weeks, your repo develops a memory that makes every Copilot session more effective than the last.

---

## Agents

51 agents ship in `.github/agents/`. The 29 featured agents below are invoked by skills during review, planning, learning, and debugging:

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
| `.github/prompts/*.prompt.md` | VS Code Copilot Chat slash-command shims (one per user-facing skill) |

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
- **gstack setup on Windows** — if the gstack sub-installer's `./setup` fails (e.g., Git Bash unavailable or other bash/runtime build issue), it falls back to `bun run gen:skill-docs`. Markdown-only gstack skills install fine either way.
- **Token-heavy pipelines** — long multi-agent sessions can hit context limits
