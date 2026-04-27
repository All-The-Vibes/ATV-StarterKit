# Project Conventions

This project uses the ATV (Agentic Tool & Workflow) Starter Kit.

## Available Workflows

- `/ce-brainstorm` — Explore what to build through collaborative dialogue
- `/ce-plan` — Create a structured implementation plan
- `/ce-work` — Execute the plan with quality checks
- `/ce-review` — Multi-agent code review
- `/ce-compound` — Document solutions for future reference
- `/atv-security` — Unified security audit: agentic config (secrets, permissions, hooks, MCP, agents) + OWASP Top 10 + STRIDE threat model for application source code
- `/autoresearch` — Autonomous iterative experiment loop: define a measurable metric and the agent edits → tests → measures → keeps/reverts on a dedicated branch until interrupted
- `/lfg` — Full autonomous pipeline (plan → work → review)

## Documentation Structure

- `docs/plans/` — Implementation plans (living documents with checkboxes)
- `docs/brainstorms/` — Brainstorm documents (what to build decisions)
- `docs/solutions/` — Documented solutions (institutional knowledge)

## gstack Skills (if installed)

- `/office-hours` — YC-style forcing questions to reframe your product
- `/plan-ceo-review` — Rethink the problem; find the 10-star product
- `/plan-eng-review` — Lock architecture, data flow, edge cases
- `/review` — Staff-level code review; auto-fix obvious issues
- `/qa` — Test app in real browser, find and fix bugs (requires Bun)
- `/ship` — Sync main, run tests, push, open PR
- `/careful` — Warn before destructive commands
- `/investigate` — Systematic root-cause debugging
- `/retro` — Weekly retrospective with trends

## Browser Automation (if installed)

- `agent-browser` — Vercel's headless browser CLI for AI agents (Rust native, fast)
- Core workflow: `agent-browser open <url>` → `snapshot -i` → interact with `@refs` → re-snapshot
- Install: `npm install -g agent-browser && agent-browser install`
- Use for QA testing, form filling, screenshots, data extraction, and web automation

## ATV Override Rules

When both ATV and gstack provide similar functionality, ATV takes priority:

- **Design docs**: Write to `docs/brainstorms/` (ATV), not `DESIGN.md` (gstack)
- **Solutions**: Document via `/ce-compound` into `docs/solutions/` (ATV), not gstack's `/retro`
- **Plans**: Use `docs/plans/` with ATV naming (`YYYY-MM-DD-NNN-type-name-plan.md`)
- **Reviews**: ATV's `/ce-review` agent selection governs; gstack's `/review` runs alongside
- **Protected artifacts**: Never flag `docs/plans/`, `docs/solutions/`, `docs/brainstorms/`, `compound-engineering.local.md`, or `.github/skills/gstack/` for deletion

## Coding Conventions

- Follow existing patterns in the codebase
- Write tests for new functionality
- Use conventional commit messages (`feat:`, `fix:`, `refactor:`)

## Continuous Learning Pipeline

This project uses ATV's continuous learning system to capture and evolve patterns.

### Learning Commands

- `/learn` — Extract patterns from recent work into instincts
- `/instincts` — View all learned patterns with confidence scores
- `/evolve` — Promote mature instincts (confidence > 0.8) into full skills
- `/observe` — Run a focused observation session on a specific domain
- `/unslop` — Unified de-slop pass: strip AI-generated code slop, comment rot, and design slop

### How It Works

1. **Observer hooks** automatically capture tool use data to `.atv/observations.jsonl`
2. **`/learn`** analyzes recent work and creates instincts in `.atv/instincts/project.yaml`
3. **Instincts** build confidence over time through repeated observation
4. **`/evolve`** promotes mature instincts into discoverable skills in `.github/skills/learned-*/`

### Key Files

- `.atv/observations.jsonl` — Raw tool use log (gitignored, ephemeral)
- `.atv/instincts/project.yaml` — Learned patterns (committed, shared with team)
- `.github/hooks/copilot-hooks.json` — Observer hook configuration
- `.github/skills/learned-*/` — Auto-generated skills from evolved instincts

### Best Practices

- Run `/learn` after completing features or at session end
- Run `/instincts` to check what patterns the project has learned
- Run `/evolve` periodically to graduate well-established conventions
- Run `/unslop` before PRs to strip AI-generated slop from code, comments, and UI
- Review generated skills before committing — they're a starting point
