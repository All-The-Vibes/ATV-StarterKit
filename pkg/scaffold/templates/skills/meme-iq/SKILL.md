---
name: meme-iq
description: "memeIQ — your AI-powered meme generation toolkit. Use when generating memes using the memegen.link API. It applies when creating memes from templates, adding text to meme images, or generating humor for PR descriptions, changelogs, and team communication. Triggers on \"create a meme\", \"make a meme\", \"meme\", \"generate meme\", \"funny image for PR\", \"memeIQ\"."
argument-hint: "[topic or context for the meme]"
---

# memeIQ — Meme Generation with memegen.link

> **memeIQ** is your AI-powered meme generation toolkit. It wraps the free memegen.link API so you can create developer memes in seconds — no API key, no auth, no config required.

## Quick Reference

| Setting | Value |
|---------|-------|
| Base URL | `https://api.memegen.link/images` |
| Formats | `.png`, `.jpg`, `.gif` (animated), `.webp` (animated) |
| Max text | ~200 chars per line (HTTP 414 if exceeded) |
| Templates | 200+ available; 16 curated below |
| Auth | None required |

## URL Construction

The URL pattern uses **variable-length path segments** — one segment per text line:

```
https://api.memegen.link/images/{template_id}/{line1}/{line2}/.../{lineN}.{format}
```

### Examples

**2-line template** (`drake` — Drakeposting):
```
https://api.memegen.link/images/drake/Writing_tests/Writing_memes.png
```

**1-line template** (`cmm` — Change My Mind):
```
https://api.memegen.link/images/cmm/Tabs_are_better_than_spaces.png
```

**3-line template** (`db` — Distracted Boyfriend):
```
https://api.memegen.link/images/db/New_framework/My_current_project/Proven_technology.png
```

**4-line template** (`gru` — Gru's Plan):
```
https://api.memegen.link/images/gru/Write_a_migration/Deploy_to_prod/Drop_the_table/Drop_the_table.png
```

**Blank line** — use a single underscore `_` for an empty text segment:
```
https://api.memegen.link/images/fry/_/Not_sure_if_bug_or_feature.png
```

**Markdown embed** for PRs and docs:
```markdown
![Meme](https://api.memegen.link/images/drake/Manual_testing/CI~s~CD_pipeline.png)
```

### Optional Query Parameters

| Param | Effect | Example |
|-------|--------|---------|
| `width` | Image width in px | `?width=400` |
| `height` | Image height in px | `?height=300` |
| `font` | Font name | `?font=impact` |
| `layout` | `default` or `top` | `?layout=top` |
| `style` | Template-specific styles | `?style=maga` |

## Special Character Encoding

Text in the URL path must encode special characters:

| Character | Encoding | Example |
|-----------|----------|---------|
| space | `_` | `Hello_world` |
| underscore | `__` | `my__var` |
| dash | `--` | `well--known` |
| `?` | `~q` | `Why~q` |
| `&` | `~a` | `this~a_that` |
| `%` | `~p` | `100~p` |
| `#` | `~h` | `issue_~h42` |
| `/` | `~s` | `CI~sCD` |
| newline | `~n` | `line1~nline2` |
| `"` | `''` | `''quoted''` |

## Curated Dev Templates

Known-good as of 2026-04-23. If a template ID fails, query `/templates` (see Template Discovery below).

| API ID | Name | Lines | Best For |
|--------|------|-------|----------|
| `drake` | Drakeposting | 2 | Preferring one thing over another |
| `db` | Distracted Boyfriend | 3 | Temptation / switching technologies |
| `cmm` | Change My Mind | 1 | Hot takes / unpopular opinions |
| `both` | Why Not Both? | 2 | Having it all |
| `fine` | This Is Fine | 2 | Production incidents / ignoring problems |
| `mordor` | One Does Not Simply | 2 | Difficulty of a task |
| `astronaut` | Always Has Been | 4 | Realizations / "wait, it always was?" |
| `exit` | Left Exit 12 Off Ramp | 3 | Choosing the wrong path |
| `gb` | Galaxy Brain | 4 | Escalating ideas from normal to absurd |
| `disastergirl` | Disaster Girl | 2 | Watching something burn |
| `rollsafe` | Roll Safe | 2 | "Can't fail if..." logic |
| `kermit` | But That's None of My Business | 2 | Clever workarounds / passive observation |
| `buzz` | X, X Everywhere | 2 | Something ubiquitous |
| `success` | Success Kid | 2 | Celebrating wins |
| `fry` | Futurama Fry | 2 | "Not sure if..." ambiguity |
| `gru` | Gru's Plan | 4 | Step-by-step plans that backfire |

## Template Discovery

To find templates beyond the curated list, query the API:

```
GET https://api.memegen.link/templates
```

Returns JSON array. Each entry has:
- `id` — template ID for URL construction
- `name` — human-readable name
- `lines` — number of text lines supported
- `keywords` — searchable tags

**Check a specific template:**
```
GET https://api.memegen.link/templates/{id}
```

**Search by keyword:** Filter the full list client-side by matching `keywords` or `name` fields. If a template ID you want to use isn't in the curated list, verify it exists with a GET request before constructing the URL.

## Line Count and Text Handling

### Match text to the template's line count

Each template supports a specific number of text lines (see `Lines` column in the curated table). **You must provide the correct number of path segments.**

- **1-line template** (e.g., `cmm`): `/cmm/{line1}.png`
- **2-line template** (e.g., `drake`): `/drake/{line1}/{line2}.png`
- **3-line template** (e.g., `db`): `/db/{line1}/{line2}/{line3}.png`
- **4-line template** (e.g., `gru`): `/gru/{line1}/{line2}/{line3}/{line4}.png`

### Blank lines

Use a single underscore `_` for any line that should be empty:
```
/drake/_/Only_bottom_text.png
```

### Long text strategies

If text exceeds ~200 characters per line:
1. **Shorten it** — meme text should be punchy, not a paragraph
2. **Split across lines** — use `~n` for newlines within a single line segment
3. **Switch template** — choose one with more lines (e.g., `gru` has 4)
4. **Use `layout=top`** — places all text at the top for better readability on long text

### When to use `layout=top`

Add `?layout=top` when a template's default text placement doesn't work well:
- Single-line templates where you want text at the top
- Long text that wraps poorly at default position

## Output Formatting

### URL only (default)
Return the clickable link:
```
https://api.memegen.link/images/drake/Fixing_bugs/Creating_features.png
```

### Markdown embed (for PRs and docs)
```markdown
![Fixing bugs vs creating features](https://api.memegen.link/images/drake/Fixing_bugs/Creating_features.png)
```

### Download command (when user requests local file)
```bash
curl -sL -o meme.png "https://api.memegen.link/images/drake/Fixing_bugs/Creating_features.png"
```

## Content Safety Guidelines

- **Workplace appropriate** — memes should be safe for professional contexts
- **No targeting individuals** — never create memes that mock specific people
- **No hate speech or harassment** — refuse requests for discriminatory content
- **Use existing public templates** — avoid branded or logo-heavy custom concepts
- **No NSFW content** — keep all generated memes clean
- When in doubt, choose a lighter tone or decline the request

## Important Notes

- **Template IDs must be valid** — never fabricate an ID. If unsure, check `/templates/{id}` first.
- **HTTP 414** occurs when URL path is too long (~200+ chars per line). Keep meme text short.
- **Animation** — `.gif` and `.webp` formats support animation on templates that have it.
- **Blank text** — use `_` not an empty string. Empty segments produce malformed URLs.
- **Case insensitive** — template IDs are case-insensitive but conventionally lowercase.
- **GitHub rendering** — markdown image embeds render in PR descriptions, comments, and README files.
