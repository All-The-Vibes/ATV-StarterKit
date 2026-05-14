## ATV 2.0 — All The Vibes Starter Kit

One command. Full agentic coding setup. 45 skills, 29 agents, and a memory system that makes your repo smarter with every PR.

### Recent additions

- **Copilot CLI plugin marketplace** — ATV now ships as a marketplace at `.github/plugin/marketplace.json`. Install personally with `copilot plugin marketplace add All-The-Vibes/ATV-StarterKit` then `copilot plugin install atv-everything@atv-starter-kit` (or pick category packs / single skills). See [docs/marketplace.md](https://github.com/All-The-Vibes/ATV-StarterKit/blob/main/docs/marketplace.md).
- **Autoresearch joins the core skills** — new `/autoresearch` skill scaffolds an autonomous experimentation loop: define a goal + metric, the agent iterates (edit → commit → run → measure → keep/revert) on a dedicated branch until interrupted. Sourced from [github/awesome-copilot](https://github.com/github/awesome-copilot/blob/main/skills/autoresearch/SKILL.md) (MIT). Listed under `📐 Coding Guidelines` in guided installs and shipped in all three presets.
- **`/cso` folded into `/atv-security`** — the standalone `/cso` skill has been merged into `/atv-security`, which now scans both agentic config (`.github/`, `.vscode/`) AND application source code (OWASP Top 10 + STRIDE) in a single run. Old `/cso` triggers still route to `/atv-security`. Eliminates a long-standing name collision with gstack's separate `/cso` skill.
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
