# Project Conventions

This project uses the ATV (Agentic Tool & Workflow) Starter Kit.

## Available Workflows

- `/ce-brainstorm` — Explore what to build through collaborative dialogue
- `/ce-plan` — Create a structured implementation plan
- `/ce-work` — Execute the plan with quality checks
- `/ce-review` — Multi-agent code review
- `/ce-compound` — Document solutions for future reference
- `/lfg` — Full autonomous pipeline (plan → work → review)

## Session Bookends

- `/takeoff` — Prioritized backlog briefing to start a session
- `/land` — Commit → push → PR → handoff to finish a session (never merges)

### Session continuity

At session start, if `.remember/now.md` exists in the repo, read it before responding — it contains a handoff from the previous session (branch, PR, accomplished, next steps, blockers). Treat it as authoritative context for continuing work. The `/land` skill writes this file as part of its checklist.

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
- `/cso` — OWASP Top 10 + STRIDE threat model
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
- Review generated skills before committing — they're a starting point
