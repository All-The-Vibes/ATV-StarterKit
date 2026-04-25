---
title: "docs: hyped marketing-brief changelog for last-week starter kit drops"
type: docs
status: deferred
date: 2026-04-24
---

# docs: Hyped Marketing-Brief Changelog for Last-Week Starter Kit Drops

> **Resolution (2026-04-25):** Deferred. The marketing-brief is a one-off
> doc artifact, not blocking work. The plan stands as-is and can be
> executed later via `/ce:work` against this file directly when there's
> appetite to ship the doc. Tracked here so it doesn't keep re-appearing
> as an untracked orphan.

## Overview

Produce a single new markdown file: a **marketing-brief / hype changelog** that summarizes the last seven days of merged + pending PRs in the ATV Starter Kit. Headline acts: `/ghcp-review-resolve` (PRs #23, #26), `/land` + `/takeoff` (PR #25), Rajesh's `/atv-security` + `/cso` security skills (PR #24), Karpathy Guidelines (Kevin/forrestchang port, v2.5.7 + commit f47e6e0), memeIQ Easter Egg (PR #22), Windows installer hardening (PR #20), and the Opus-4.6 agent file repair (PR #21).

This is documentation only — no source files change.

## Problem Frame

The team has shipped a flurry of high-leverage skills in the last week and there is no single "go read this" artifact that sells what changed. `CHANGELOG.md` is tonally flat and stops at v2.5.7 (Karpathy). The audience for the new doc is Copilot/Claude users browsing the repo, plus internal hype channels — they want to know what's new, why it matters, and how to invoke it, in a tone that matches the Starter Kit's playful brand.

## Requirements Trace

- R1. Single new markdown file under repo, dated 2026-04-24, marketing-brief tone (hyped, energetic, scannable).
- R2. Covers every named drop: `/ghcp-review-resolve`, `/land`, `/takeoff`, `/atv-security`, `/cso`, Karpathy Guidelines, memeIQ, plus the supporting fixes (Windows installer, agent-file repair).
- R3. Each entry explains: what it is, why it matters, how to invoke it (slash command or install path), and credits the author/PR.
- R4. Closing call-to-action that points readers to install, upgrade, or try the new skills.
- R5. Does NOT modify `CHANGELOG.md` (Keep-a-Changelog discipline preserved separately).

## Scope Boundaries

- No edits to `CHANGELOG.md`, `README.md`, or any source/template files.
- No version bump, no release notes regeneration.
- No PR creation as part of this plan — `/land` will handle that downstream.

## Context & Research

### Source PRs (last 7 days, ordered newest-first)

| PR | Title | Author | Status |
|----|-------|--------|--------|
| #26 | feat(ghcp-review-resolve): port pr-review-toolkit + add thread resolution + PR task-ticking | @stephschofield | open |
| #25 | feat(skills): port /land and /takeoff to Copilot skills | @stephschofield | open |
| #24 | feat: add /atv-security and /cso security skills | @rajesh-ms | merged 2026-04-24 |
| #23 | feat(skills): add ghcp-review-resolve for dual PR review | @stephschofield | merged 2026-04-24 |
| #22 | feat(installer): add memeIQ Easter Egg scaffolding | @shyamsridhar123 | merged 2026-04-24 |
| #21 | fix(agents): Opus-4.6 fix of botched agent files | @brandonh-msft | merged 2026-04-24 |
| #20 | fix: harden Windows postinstall extraction | @dc995 | merged 2026-04-21 |
| commit f47e6e0 | feat: Karpathy Guidelines skill (v2.5.7) | (Kevin/forrestchang upstream port) | shipped |

### Where the file goes

`docs/changelog/2026-04-24-week-in-review-marketing-brief.md` — `docs/changelog/` does not exist yet but `docs/` does; create the subdir. Naming mirrors the `YYYY-MM-DD-<slug>.md` pattern used in `docs/plans/` and `docs/brainstorms/`.

## Key Technical Decisions

- **New file, not a CHANGELOG edit.** The marketing voice would clash with Keep-a-Changelog discipline. The hype doc lives alongside it as `docs/changelog/<date>-week-in-review-marketing-brief.md`.
- **Section per drop, not chronological log.** Each new skill gets its own H2 with a one-line tagline, an "anti-template" bulletized "what / why / how" block, and PR + author credit.
- **Voice: hyped but credible.** Hooks like "ship faster," "pair with Copilot," "no more babysitting," "dual reviewers, zero noise" — declarative and concrete. Avoids "revolutionary," "game-changing," and other AI-slop superlatives.
- **Emoji used as section anchors, not decoration** — one per drop (🛬 for /land, ✈️ for /takeoff, 🛡️ for security, 🎯 for Karpathy, etc.).

## Implementation Units

- [ ] **Unit 1: Create the marketing-brief changelog file**

**Goal:** Produce `docs/changelog/2026-04-24-week-in-review-marketing-brief.md` with all named drops.

**Requirements:** R1, R2, R3, R4, R5

**Dependencies:** none

**Files:**
- Create: `docs/changelog/2026-04-24-week-in-review-marketing-brief.md`

**Approach:**
1. Open with a hyped 2–3 sentence lede ("Seven days. Eight skills. One Starter Kit that just got dangerous.").
2. "The Headliners" section — `/ghcp-review-resolve`, `/land`, `/takeoff`, `/atv-security`, `/cso`, Karpathy Guidelines, memeIQ. One H3 per skill. Each H3 carries:
   - Tagline (one line, declarative)
   - **What** — one sentence
   - **Why it slaps** — one or two bullets on the user value
   - **How to use it** — the slash command + minimal invocation
   - **Shipped by** — author handle + PR link
3. "Under the hood" section — Windows installer hardening (#20), agent-file repair (#21). Compact bullets; these are quality-of-life wins, not headliners.
4. "Try it now" CTA — install command (`npx atv-installer` or equivalent), upgrade nudge, link to README.
5. Close with credit roll and a wink ("PRs welcome, memes encouraged").

**Test scenarios:**
- Happy path: file renders correctly in GitHub markdown preview (headings, links, tables); all PR links resolve to real PRs (#20–#26); voice is consistent across sections; word count ~600–900 (skimmable, not bloated).
- Edge case: `docs/changelog/` directory does not yet exist — create it as part of the write.

**Verification:**
- File exists at the planned path.
- Every named drop in R2 has its own section with author + PR credit.
- No edits to `CHANGELOG.md`, `README.md`, or any code/template file.
- Tone reads as marketing-brief (declarative, energetic) without slipping into AI slop.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Tone reads as AI-slop hype | Use concrete verbs and specific outcomes; ban superlatives like "revolutionary" |
| Missing a recent drop | Use the PR table above as the authoritative list; cross-check against `git log --since="1 week ago"` |
| File location confuses contributors | Place under `docs/changelog/` and link from CTA back to `CHANGELOG.md` for canonical version log |

## Sources & References

- PRs #20–#26 (see table above)
- `CHANGELOG.md` (existing, untouched)
- Recent commits: `dfce627`, `f6a0c3d`, `92b49bd`, `f47e6e0`, `1f30760`
