# Brainstorm: Meme Generation Skill & Agent

**Date:** 2026-04-23
**Status:** Ready for planning
**Approach:** Skill + Creative Agent (Approach B)

## What We're Building

A meme generation capability for the ATV Starter Kit consisting of two components:

1. **`meme-generation` skill** (`.github/skills/meme-generation/SKILL.md`) — Domain knowledge for memegen.link API: URL construction, special character encoding, curated dev-friendly template list, content safety guidelines.

2. **`meme-creator` agent** (`.github/agents/meme-creator.agent.md`) — Creative workflow agent that detects context (PR description, changelog, or freeform), selects appropriate templates, generates text, presents URL/markdown embeds, and offers iterative refinement.

**Backend:** memegen.link only (free, no auth, 200+ templates, stateless URL construction). No ImgFlip/meme-mcp dependency.

## Why This Approach

- **Zero-config:** memegen.link requires no API keys, accounts, or setup — works immediately
- **Agent-skill complementarity:** Follows established ATV pattern (like design-iterator + design skills) where skill provides domain knowledge and agent provides workflow loop
- **Context-aware:** Agent knows when it's being used for PR descriptions vs general meme creation, and auto-formats output accordingly
- **Iterative:** Agent supports refinement loop — user can ask for different template, adjust text, change style
- **YAGNI-compliant:** No helper scripts, no batch processing, no local download by default — just URL-based delivery with optional download

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Placement** | `.github/skills/` now, scaffold later | Prove the concept before shipping to all ATV users |
| **Backend** | memegen.link only | Zero config, no auth, free, 200+ templates |
| **Agent location** | `.github/agents/meme-creator.agent.md` | Follows ATV convention; mirrors scaffold agent path |
| **Delivery** | URL primary + optional download | URLs are lightweight; markdown embeds for PRs |
| **PR integration** | Auto-generate `![meme](url)` embeds | Key use case: humor in PR descriptions & changelogs |
| **Template curation** | Include curated dev-friendly list | Help users find relevant templates quickly |
| **Content safety** | Include guidelines | Workplace-appropriate memes, no copyright violations |
| **MCP server** | Not included | memegen.link works via direct URL construction, no MCP needed |

## Technical Shape

### memegen.link API Essentials

- **URL pattern:** `https://api.memegen.link/images/{template_id}/{top_text}/{bottom_text}.png`
- **Special char encoding:** `_`=space, `__`=underscore, `--`=dash, `~q`=?, `~a`=&, `~p`=%, `~h`=#, `~s`=/,  `~n`=newline, `''`=double-quote
- **Formats:** .png, .jpg, .gif, .webp
- **Query params:** `width`, `height`, `font`, `style`
- **Templates API:** `GET https://api.memegen.link/templates` returns full list with metadata
- **Limit:** 200 chars max per text line (414 error if exceeded)

### Skill Structure

```
.github/skills/meme-generation/
├── SKILL.md          # Domain knowledge, API docs, curated templates, safety guidelines
```

### Agent Structure

```
.github/agents/
└── meme-creator.agent.md   # Creative workflow: context → template → text → URL → refine
```

Agent lives in `.github/agents/` following the ATV convention for standalone agents (mirrors scaffolded agents at `pkg/scaffold/templates/agents/`). The skill provides domain knowledge; the agent provides the workflow loop.

### Curated Dev Template List (Initial)

Templates to include with dev-friendly descriptions:

| Template ID | Common Name | Best For |
|-------------|-------------|----------|
| `drake` | Drake Hotline Bling | Preferring one thing over another |
| `distracted` | Distracted Boyfriend | Temptation / switching technologies |
| `change-mind` | Change My Mind | Hot takes / opinions |
| `both` | Why Not Both | Having it all |
| `fine` | This Is Fine | Production incidents / ignoring problems |
| `one-does-not` | One Does Not Simply | Difficulty of a task |
| `always-has-been` | Always Has Been | Realizations |
| `exit` | Left Exit 12 Off Ramp | Choosing the wrong path |
| `expanding-brain` | Expanding Brain | Escalating ideas |
| `surprised-pikachu` | Surprised Pikachu | Predictable outcomes |
| `disaster-girl` | Disaster Girl | Watching something burn |
| `rollsafe` | Roll Safe | "Can't fail if..." logic |
| `think-about-it` | Think About It | Clever workarounds |
| `buzz` | Buzz Lightyear | "X, X Everywhere" |
| `success` | Success Kid | Celebrating wins |
| `fry` | Futurama Fry | "Not sure if..." |

### Content Safety Guidelines

- Keep memes workplace-appropriate
- No offensive, discriminatory, or NSFW content
- No copyrighted characters or logos beyond fair-use meme templates
- Avoid targeting individuals or teams negatively
- When in doubt, err on the side of wholesome humor

### Agent Workflow

```
1. User asks for a meme (with context or freeform)
2. Agent analyzes context (PR? changelog? general?)
3. Agent suggests template + text based on context
4. Agent constructs memegen.link URL with proper encoding
5. Agent presents:
   - Clickable URL to view the meme
   - Markdown embed if PR/doc context: ![meme description](url)
6. User can refine: "try a different template", "change the text", "make it about X instead"
7. Agent iterates until user is satisfied
```

## Resolved Questions

- **Should we use ImgFlip/meme-mcp?** No — memegen.link alone provides everything needed with zero config.
- **Should this be scaffold-embedded from day one?** No — start in `.github/skills/`, promote to scaffold once proven.
- **Should we include helper scripts?** No — YAGNI. URL construction is simple enough to do inline.
- **Should meme-mcp be in default MCP config?** No — not using meme-mcp at all.

## Open Questions

_None — all questions resolved through dialogue._

## Out of Scope

- Batch meme generation
- Custom image uploads (memegen.link `/images/custom` endpoint)
- AI-powered automatic meme generation (memegen.link `/images/automatic` endpoint)
- Local image storage/gallery
- ImgFlip integration
- MCP server configuration
- TUI category changes (not scaffolded yet)

These can be added in future iterations if demand exists.
