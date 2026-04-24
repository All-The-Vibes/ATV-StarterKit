## ATV 2.0 — All The Vibes Starter Kit

One command. Full agentic coding setup. 45 skills, 29 agents, and a memory system that makes your repo smarter with every PR.

### Recent additions

- **memeIQ joins the guided installer** — the customize flow now includes a `🥚 Easter Eggs` category with an opt-in meme generation toolkit backed by memegen.link.
- **Installer-ready memeIQ scaffolding** — guided installs can now write both `.github/skills/meme-iq/SKILL.md` and `.github/agents/meme-iq.agent.md` into the target repo.
- **Cleaner releases and PRs** — local planning, session, and build artifacts are now ignored so only intentional product changes ship.

### What's new in 2.0

**Three pillars, one installer:**
- **Compound Engineering** — gated planning pipeline with institutional knowledge that compounds across sessions
- **gstack** — Garry Tan's AI sprint process: 30 skills covering review, QA, shipping, safety, and retros
- **agent-browser** — Vercel's native Rust browser CLI for AI agents with snapshot-ref workflow

**Guided experience redesigned:**
- Preset-based wizard: Starter (13 skills, instant) / Pro (35+ skills) / Full (45+ skills)
- Animated install progress with per-step spinners
- "ALL THE VIBES 2.0" retro terminal banner

**Memory as a first-class feature:**
- `docs/solutions/` — git-tracked institutional knowledge searched by future planning sessions
- `docs/brainstorms/` + `docs/plans/` — living documents that flow through the pipeline
- gstack session tracking + agent-browser session persistence

### Install

```bash
npx atv-starterkit init              # quick: ATV core only
atv-installer init --guided          # choose: Starter / Pro / Full
```

### Links

- [README](https://github.com/All-The-Vibes/ATV-StarterKit#readme)
- [Compound Engineering](https://github.com/EveryInc/compound-engineering-plugin)
- [gstack](https://github.com/garrytan/gstack)
- [agent-browser](https://github.com/vercel-labs/agent-browser)
