---
name: kb-memory-review
description: High-cost KB memory maintenance pass for repo-local project memory. Use when docs/context/memory-maintenance.md says review is due, after large epics, after repeated contradictions/stale-doc signals, or when the user asks to deep review, consolidate, compact, or reconcile KB memory. Starts from recorded signals and pointers instead of blind repo crawling.
argument-hint: "[optional scope, signal type, or file path]"
---

# KB Memory Review

Review and repair the repo's local memory system so fresh sessions stay cheap and accurate.

This is not startup lookup. `kb-map` handles normal context routing. This skill is an explicit, higher-cost maintenance pass driven by `docs/context/memory-maintenance.md`.

Do not call this "Dreams" in user-facing output.

## Use When

- `kb-complete` recommended a memory review.
- `docs/context/memory-maintenance.md` has contradiction, overlap, stale-doc, bloat, or repeated-rediscovery signals.
- A large epic/migration finished.
- Fresh sessions keep rediscovering the same facts.
- Project memory feels noisy, contradictory, bloated, or stale.

Do not run automatically from `kb-start`, `kb-work`, `kb-finalize`, or `kb-complete`. Those skills may recommend it; this skill runs only when explicitly invoked or accepted by the user.

## Core Rule

Start from signals and pointers. Do not crawl the whole repo unless the user explicitly asks for a broad memory audit.

## Preflight

1. Resolve the active project root using the `kb-map` project-root rule.
2. Read:
   - `todo.md`
   - `docs/context/PROJECT.md`
   - `docs/context/landmines.md` when present
   - `docs/context/memory-maintenance.md`
3. If `docs/context/memory-maintenance.md` is missing, create it only if the user asked for memory review. Otherwise recommend running `kb-complete` after normal work so signals accumulate first.
4. If no signals exist and no explicit broad audit was requested, stop with: `No memory-review signals found. Normal kb-map refresh is enough.`

## Scope Modes

| Mode | Use When | Reads |
|---|---|---|
| `targeted` | one or more signals exist, or user names a subsystem/file | signal sources, direct neighbors, `PROJECT.md` |
| `area` | many signals cluster around one subsystem, research area, or docs folder | relevant architecture/research/decision docs plus linked handoffs/plans |
| `broad` | user explicitly asks for full memory audit, or epic/migration produced widespread drift | indexes first, then sampled subsystem docs; avoid reading all code unless needed |

Default to `targeted`.

## Signal Handling

For each signal in `docs/context/memory-maintenance.md`, classify the action:

| Signal Type | Likely Action |
|---|---|
| `contradiction` | reconcile docs; add/update decision note when both sides had a reason |
| `overlap` | consolidate, cross-link, or clarify scope boundaries |
| `stale-doc` | run `kb-map refresh` or edit affected docs directly |
| `bloat` | run `kb-compact` on the specific doc/section |
| `repeated-rediscovery` | promote the fact into `PROJECT.md`, architecture, research, decisions, `docs/solutions/`, or a learned skill candidate |
| `landmine-stale` | verify whether the owner surface is fixed; archive if verified, otherwise keep active or mark stale-review |

Do not fix by deleting useful context blindly. Preserve exact paths, commands, dates, IDs, decisions, rejected approaches, and verification commands.

## Workflow

1. **Triage signals**
   - Group signals by subsystem, doc area, and action type.
   - Identify the smallest useful batch.
   - If scope is broad, present the batch order and proceed with the first batch unless the user redirects.

2. **Investigate**
   - Read each signal source.
   - Read only directly linked docs needed to verify the conflict, overlap, stale claim, or bloat.
   - Use code search only to verify current truth for disputed claims.

3. **Choose action**
   - Context map or architecture drift: update `docs/context/PROJECT.md` or `docs/context/architecture/*`.
   - Operations drift: update `docs/context/operations/*`.
   - Research drift/overlap: update or consolidate `docs/context/research/*`.
   - Decision drift: update or create `docs/context/decisions/*`.
   - Landmine drift: update `docs/context/landmines.md`; resolved entries need
     proof, while unfixed stale entries stay visible for future sessions.
   - Solution-learning drift: invoke `ce-compound-refresh` with the narrowest scope.
   - Bloat only: invoke `kb-compact` on the specific artifact.
   - Repeated rediscovery: add one durable pointer where future sessions will actually find it.

4. **Edit**
   - Make the smallest edits that restore current truth.
   - Prefer cross-links over duplication when two docs remain distinct.
   - Prefer consolidation when two docs answer the same question for the same audience.

5. **Review**
   - Run `document-review` when requirements, architecture, decision, research, or PROJECT docs changed substantially.
   - Run `kb-map lookup <reviewed area>` afterward to verify a fresh session would find the right pointer.

6. **Update maintenance index**
   - Set `Last deep review: YYYY-MM-DD` if a meaningful review happened.
   - Move resolved signal entries from `## Signals Since Last Review` to a `## Reviewed Signals` section with a one-line result.
   - Leave unresolved signals in place with `Status: open` and the blocker.
   - Reset counters only for resolved signal types. Unresolved signals still count.

## Output

Report:

- Scope reviewed.
- Signals resolved, left open, or converted to follow-up.
- Files edited.
- Skills invoked (`kb-map refresh`, `kb-compact`, `ce-compound-refresh`, `document-review`).
- Whether fresh-session lookup now points to the right docs.

End with one of:

- `Memory review complete.`
- `Memory review partially complete: <open blockers>.`
- `No memory review needed: <reason>.`
