---
title: "feat: Add Meme Generation Skill and Creative Agent"
type: feat
status: completed
date: 2026-04-23
origin: docs/brainstorms/2026-04-23-meme-generation-brainstorm.md
---

# feat: Add Meme Generation Skill and Creative Agent

## Enhancement Summary

**Deepened on:** 2026-04-23
**Research agents used:** skill-spec-research, template-validation, agent-native-review, multi-lens-plan-review

### Key Improvements
1. **Fixed 8 incorrect template IDs** — validated all 16 curated templates against live API; corrected IDs and line counts
2. **Redesigned agent as outcome-oriented** — replaced choreographed 7-step pipeline with autonomous creative agent that self-evaluates before presenting
3. **Added dynamic template discovery** — curated list is starting point; agent can query API for more
4. **Added composability** — agent accepts structured input from other agents (changelog, PR review)
5. **Added `argument-hint`** to skill frontmatter for better autocomplete

### New Considerations Discovered
- memegen.link is single-maintainer (jacebrowning); no SLA but stable for years
- Encoding rules (`~s`=`/`, etc.) should be verified empirically during implementation
- 200-char limit is undocumented in OpenAPI spec; verify empirically
- Custom fonts, style overlays, and text colors are API capabilities explicitly out of scope

## Overview

Add a meme generation capability to the ATV Starter Kit through two new files:
1. **`meme-generation` skill** — Domain knowledge for the memegen.link API (URL construction, encoding, templates, safety)
2. **`meme-creator` agent** — Creative workflow agent for context-aware meme generation with iterative refinement

Uses memegen.link as the sole backend — zero config, no auth, 200+ templates. (see brainstorm: `docs/brainstorms/2026-04-23-meme-generation-brainstorm.md`)

## Problem Statement / Motivation

Developers benefit from humor in PR descriptions, changelogs, and team communication. Currently, creating memes requires leaving the development environment and using external tools. An integrated meme skill lets AI agents generate contextually appropriate memes inline during development workflows.

## Proposed Solution

Create two files following established ATV patterns:

| Component | File | Pattern Source |
|-----------|------|----------------|
| Skill | `.github/skills/meme-generation/SKILL.md` | `gemini-imagegen` skill pattern |
| Agent | `.github/agents/meme-creator.agent.md` | `design-iterator` agent pattern |

**Placement:** Repo-only (`.github/`) for now. Scaffold promotion (`pkg/scaffold/templates/`) deferred until proven. This means no changes to `catalog.go`, TUI, or guided install — by design.

## Technical Considerations

### memegen.link API (from brainstorm research)

- **URL pattern:** `https://api.memegen.link/images/{template_id}/{top}/{bottom}.png`
- **Special chars:** `_`=space, `__`=underscore, `--`=dash, `~q`=?, `~a`=&, `~p`=%, `~h`=#, `~s`=/,  `~n`=newline, `''`=double-quote
- **Formats:** .png, .jpg, .gif (animated), .webp (animated)
- **Query params:** `width`, `height`, `font`, `style`, `layout`
- **Templates API:** `GET https://api.memegen.link/templates` — returns JSON array with `id`, `name`, `lines` (count), `keywords`
- **Limits:** 200 chars max per text line (HTTP 414 if exceeded); no rate limits documented
- **Line counts vary by template:** Some templates support 1 line, most support 2, some support 3+. The `lines` field in template metadata indicates the count.

### Architecture Decisions

- **No env vars required** — memegen.link is fully public, no API key needed
- **No helper scripts** — URL construction is simple string formatting; YAGNI (see brainstorm)
- **No MCP server** — direct URL construction, no tool-calling overhead needed
- **No reference files initially** — keep everything in SKILL.md (< 500 lines); extract to `references/` only if it grows

## Acceptance Criteria

### Functional Requirements

- [x] Skill file `.github/skills/meme-generation/SKILL.md` exists with valid frontmatter
- [x] Skill includes memegen.link URL construction pattern with encoding rules
- [x] Skill includes curated dev template table (12-16 templates with line counts)
- [x] Skill includes template discovery instructions (GET /templates endpoint)
- [x] Skill includes output formatting guidance (URL, markdown embed, download)
- [x] Skill includes content safety guidelines
- [x] Skill includes line-count handling rules (1-line vs 2-line templates, long text fallback)
- [x] Agent file `.github/agents/meme-creator.agent.md` exists with valid frontmatter
- [x] Agent includes `<examples>` section with 3-4 use cases
- [x] Agent includes context detection logic (PR vs changelog vs freeform)
- [x] Agent includes template selection intelligence
- [x] Agent includes self-evaluation loop (checks quality before presenting)
- [x] Agent includes composability (accepts structured input from other agents)
- [x] Agent references meme-generation skill for domain knowledge

### Quality Gates

- [x] `go build ./...` still passes (no Go changes, but verify no breakage)
- [x] `go test ./...` still passes
- [x] Skill frontmatter name matches directory name (`meme-generation`)
- [x] Skill description includes trigger keywords for auto-discovery
- [x] Agent description clearly states when to use it
- [x] Both files render correctly as markdown

## Implementation Phases

### Phase 1: Create `meme-generation` Skill

**File:** `.github/skills/meme-generation/SKILL.md`

**Frontmatter:**
```yaml
---
name: meme-generation
description: This skill should be used when generating memes using the memegen.link API. It applies when creating memes from templates, adding text to meme images, or generating humor for PR descriptions, changelogs, and team communication. Triggers on "create a meme", "make a meme", "meme", "generate meme", "funny image for PR".
argument-hint: "[topic or context for the meme]"
---
```

**Sections to implement:**

1. **Title + intro** — "Generate memes using memegen.link. No API key required."
2. **Quick Reference** — base URL, formats, limits in compact table
3. **URL Construction Pattern** — the core: `https://api.memegen.link/images/{id}/{top}/{bottom}.{ext}`
   - Include 2 canonical examples with real URLs
   - Include 1 markdown embed example: `![meme](url)`
4. **Special Character Encoding** — complete table of all encodings
5. **Curated Dev Template List** — compact table with columns: Template ID, Name, Lines, Best For
   - Include the `lines` count for each template (critical for text placement)
   - ~16 templates, all **validated against live API** (2026-04-23):

   | API ID | Name | Lines | Best For |
   |--------|------|-------|----------|
   | `drake` | Drakeposting | 2 | Preferring one thing over another |
   | `db` | Distracted Boyfriend | 3 | Temptation / switching technologies |
   | `cmm` | Change My Mind | 1 | Hot takes / unpopular opinions |
   | `both` | Why Not Both? | 2 | Having it all |
   | `fine` | This Is Fine | 2 | Production incidents / ignoring problems |
   | `mordor` | One Does Not Simply | 2 | Difficulty of a task |
   | `astronaut` | Always Has Been | 4 | Realizations |
   | `exit` | Left Exit 12 Off Ramp | 3 | Choosing the wrong path |
   | `gb` | Galaxy Brain | 4 | Escalating ideas |
   | `disastergirl` | Disaster Girl | 2 | Watching something burn |
   | `rollsafe` | Roll Safe | 2 | "Can't fail if..." logic |
   | `kermit` | But That's None of My Business | 2 | Clever workarounds |
   | `buzz` | X, X Everywhere | 2 | Something ubiquitous |
   | `success` | Success Kid | 2 | Celebrating wins |
   | `fry` | Futurama Fry | 2 | "Not sure if..." |
   | `gru` | Gru's Plan | 4 | Step-by-step plans that backfire |
6. **Template Discovery** — how to fetch full list from `GET /templates` and filter by keywords
7. **Line Count & Text Handling** — rules for:
   - Matching text to template line count
   - Long text fallback (shorten, split, or switch template)
   - When to use `layout=top` for single-line placement
   - 200-char limit warning
8. **Output Formatting** — three modes:
   - URL only (default): clickable link
   - Markdown embed: `![description](url)` for PRs/docs
   - Download: `curl -o meme.png "url"` when user requests local file
9. **Content Safety Guidelines** — workplace-appropriate, no copyright violations, no targeting individuals
10. **Important Notes** — gotchas (414 on long text, template ID must be valid, format affects animation support)

### Phase 2: Create `meme-creator` Agent

**File:** `.github/agents/meme-creator.agent.md`

**Frontmatter:**
```yaml
---
description: Creative meme generation agent. Detects context (PR, changelog, freeform), selects templates, constructs memegen.link URLs with proper encoding, and offers iterative refinement. Use when generating memes, adding humor to PRs, or creating visual jokes.
user-invocable: true
---
```

**Sections to implement:**

1. **`<examples>` section** (XML format, 4 examples):
   - PR description meme (context-aware)
   - Changelog/release humor
   - Freeform "make me a meme about X"
   - Refinement: "try a different template"

2. **Preamble** — "You are a creative meme generation specialist who combines humor with developer culture. Your job: generate memes that land."

3. **Outcome Definition** (outcome-oriented, NOT a choreographed pipeline):
   ```
   ## Your Job
   Generate a meme that lands. Given a request (explicit or contextual),
   produce a memegen.link URL with text that's concise, punchy, and
   contextually appropriate. Use your judgment about template selection,
   tone, and delivery format.
   ```

4. **Self-Evaluation Loop** — agent checks quality before presenting:
   - Does the text fit the template's line count?
   - Is it actually funny, or just technically correct?
   - Does it match the tone (snarky for PRs, celebratory for releases)?
   - If not, try a different template or rewrite. Don't present mediocre output.

5. **Template Selection Intelligence** — mapping from context/sentiment to templates:
   - "frustration/difficulty" → `mordor`, `fine`
   - "comparison/preference" → `drake`, `db`
   - "celebration/success" → `success`, `both`
   - "realization/surprise" → `astronaut`, `gb`
   - "plans gone wrong" → `gru`, `exit`
   - Start with curated list for speed; query `GET /templates` for unusual requests

6. **Context Awareness** — the agent uses available signals:
   - PR diff summary, branch name, recent commits (if PR context)
   - Changelog entries, version numbers (if release context)
   - Freeform topic (default)
   - Previous memes in session (avoid repeats)

7. **Composability** — accept input from other agents:
   ```
   ## Input Modes
   - Freeform: "Make a meme about our deploy"
   - Structured: { context: "pr_merged", title: "Fix auth bug", sentiment: "relief" }
   - Embedded: Another agent asks for a meme URL to include in its output
   ```

8. **Skill Loading Check** — "Check if the meme-generation skill is loaded in your context. Apply its URL construction rules and encoding table."

9. **Output Format Template**:
   ```
   🎭 **Meme: [template name]**
   
   [clickable URL]
   
   📋 **Markdown embed** (copy for PR/docs):
   `![description](url)`
   
   🔄 Want changes? Try: "different template", "change the text", "make it about X instead"
   ```

10. **Content Safety** — agent-level enforcement: check text before generating, refuse inappropriate content

11. **Important Guidelines** — keep text short and punchy, always return the final URL, never fabricate template IDs, proactively suggest one alternative angle alongside primary meme

### Phase 3: Verify & Test

1. **Syntax check** — verify both files have valid YAML frontmatter
2. **Name consistency** — skill frontmatter `name: meme-generation` matches directory name
3. **Build verification** — `go build ./...` and `go test ./...` still pass
4. **URL validation** — test 2-3 constructed URLs to confirm they render correctly on memegen.link
5. **Template ID validation** — verify all curated template IDs exist by fetching `https://api.memegen.link/templates/{id}`
6. **Encoding verification** — empirically test each special char encoding (`_`, `~q`, `~a`, etc.) against live API
7. **Line limit verification** — test 200+ char text to confirm 414 response
8. **Cross-reference check** — agent mentions skill by name; skill is self-contained
9. **Markdown render check** — both files render correctly as GitHub-flavored markdown

## Dependencies & Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| memegen.link goes down or changes API | Low | Single-maintainer project but stable for years; URL pattern is simple enough to document alternatives |
| Template IDs change | Low | Verify IDs during implementation; template list is advisory, not hard-coded logic |
| Content safety edge cases | Medium | Guidelines are advisory; AI agent judgment handles edge cases |
| Skill too long (>500 lines) | Low | Extract curated list to `references/templates.md` if needed |

## Success Metrics

- Skill and agent are discoverable via trigger keywords ("create a meme", "meme for PR")
- Generated memegen.link URLs resolve to valid images
- Agent produces contextually appropriate meme suggestions
- Markdown embeds render correctly in GitHub PR descriptions

## Out of Scope

Per brainstorm: batch generation, custom image uploads, AI-powered `/images/automatic` endpoint, local gallery, ImgFlip integration, MCP server config, TUI/scaffold changes, custom fonts, style overlays, text color customization.

## Integration Opportunities (Future)

Other ATV skills could optionally invoke or suggest the meme-creator agent:
- **`changelog` skill** — offer a celebratory meme when generating release notes
- **PR review agents** — suggest a meme for PR descriptions
- **`ce-compound`** — embed a meme in solution documentation for memorability

These integrations are out of scope for v1 but the agent's composability (structured input mode) is designed to enable them.

## Sources & References

### Origin

- **Brainstorm document:** [docs/brainstorms/2026-04-23-meme-generation-brainstorm.md](docs/brainstorms/2026-04-23-meme-generation-brainstorm.md) — Key decisions: memegen.link only backend, repo-only placement, skill+agent approach, URL-primary delivery

### Internal References

- Skill pattern: `.github/skills/gemini-imagegen/SKILL.md` (closest analog — image generation skill)
- Agent pattern: `.github/agents/design-iterator.agent.md` (agent-skill complementarity, `<examples>` format)
- Skill spec: `.github/skills/create-agent-skills/references/official-spec.md` (authoritative SKILL.md format)
- Agent-skill loading: `.github/agents/design-iterator.agent.md:183-189` (skill auto-loading pattern)

### External References

- memegen.link API: `https://api.memegen.link/docs/`
- memegen.link OpenAPI spec: `https://api.memegen.link/docs/openapi.json`
- memegen.link templates: `https://api.memegen.link/templates`
- Source: `https://github.com/jacebrowning/memegen`
