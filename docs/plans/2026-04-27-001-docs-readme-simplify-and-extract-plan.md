---
title: "docs: Simplify README, extract deep docs to DOCS.md"
type: docs
status: active
date: 2026-04-27
---

# docs: Simplify README, extract deep docs to DOCS.md

## Overview

The current `README.md` is 628 lines and mixes critical onboarding (what ATV is, how to install, the sprint workflow) with deep reference material. This plan reduces `README.md` to **only the critical first-touch material** and moves everything else to a single `DOCS.md` at the repo root.

## Problem Frame

- **Critical (stays in README):** what ATV 2.0 is + the four pillars, OS-specific install instructions (macOS/Linux vs. Windows), and the full sprint table.
- **Reference (moves to `DOCS.md`):** Behavioral Foundation + Four Pillars deep dive, Guided Experience walkthrough, How Learning Works, De-Slop, Memory Architecture, Agents inventory, What Gets Installed, How It Works Under the Hood, Development, Limitations, the collapsed full skill reference table.
- **Updates:** Full Sprint reflects weekend additions — `/atv-security` (folds in former `/cso`, OWASP + STRIDE + AgentShield config rules), `/autoresearch` (autonomous metric-driven loop), `/atv-doctor`, `/atv-update`.

## Requirements Trace

- R1. README contains an ATV 2.0 explainer with links to each pillar.
- R2. README contains OS-specific install instructions for macOS/Linux and Windows.
- R3. README contains the **Full Sprint** map including `/atv-security`, `/autoresearch`, `/atv-doctor`, `/atv-update`.
- R4. All non-critical content is preserved verbatim inside `DOCS.md` with same headings/anchors.
- R5. README links to `DOCS.md` from nav and a "Deeper docs" pointer.
- R6. Existing intra-repo links continue to resolve.

## Scope Boundaries

- Out of scope: rewriting `docs/marketplace.md`, changing any SKILL.md, restructuring `docs/`, hero image, badges, video embed, footer, code under `pkg/`/`plugins/`/`npm/`.
- Non-goal: docs site or splitting reference across multiple files. One root `DOCS.md` only.

## Context & Research

### Relevant files

- `README.md` — current 628-line file.
- `pkg/scaffold/templates/skills/atv-security/SKILL.md` — `/atv-security` absorbs `/cso` (OWASP + STRIDE + 33 AgentShield rules).
- `pkg/scaffold/templates/skills/autoresearch/SKILL.md` — autonomous metric-driven loop.
- `pkg/scaffold/templates/skills/atv-doctor/SKILL.md`, `atv-update/SKILL.md` — v2.6.3 additions.

### Weekend additions to surface

| Skill | Phase placement |
|---|---|
| `/atv-security` | Review (replaces `/gstack-cso`) |
| `/autoresearch` | Build (autonomous variant) |
| `/atv-doctor` | Reflect (diagnose install drift) |
| `/atv-update` | Reflect (self-update install) |

## Key Technical Decisions

- **Single `DOCS.md` at repo root** — matches user's explicit instruction.
- **Verbatim move, not rewrite** — keeps diff reviewable and preserves external deep-links.
- **Top-of-README nav trimmed** to surviving sections plus `DOCS.md` link.
- **No anchor breaking changes** in moved content.

## Open Questions

### Resolved During Planning

- `/atv-doctor`, `/atv-update` → Reflect column (project-health skills).
- Three Pillars stays in README? → No, the four pillars are already linked from "What is ATV 2.0?"; deeper section (now restructured as Behavioral Foundation + Four Pillars) moves.
- 45-skill reference table stays? → No, moves to `DOCS.md`.

### Deferred to Implementation

- Whether to keep Quick Start block as-is or fold into Installation. Default: keep 6-line Quick Start after Installation.

## Implementation Units

- [ ] **Unit 1: Create `DOCS.md` with all non-critical sections moved verbatim**

**Goal:** Single root reference doc containing every section that leaves `README.md`.

**Files:**
- Create: `DOCS.md`

**Approach:**
- H1 `# ATV 2.0 — Deeper Documentation` + orientation paragraph + "← Back to README" link.
- Sections moved (in order, headings preserved verbatim except for the restructured Foundation/Four Pillars split): Three Pillars, Guided Experience, How Learning Works, De-Slop, Memory Architecture, Agents, What Gets Installed, How It Works Under the Hood, Development, Limitations, full skill reference (un-collapsed).
- Update Full skill reference inside `DOCS.md` to add `/atv-security`, `/autoresearch`, `/atv-doctor`, `/atv-update`; remove `/gstack-cso`.

**Test scenarios:**
- Happy path: rendered `DOCS.md` shows every moved section with intact tables.
- Edge case: relative links resolve from repo root.
- Edge case: anchor `#how-learning-works` resolves inside `DOCS.md`.

**Verification:** `grep -c "^## " DOCS.md` ≥ 11.

- [ ] **Unit 2: Slim `README.md` to critical sections + OS-specific install**

**Goal:** Compact README — what ATV is, install, full sprint.

**Dependencies:** Unit 1.

**Files:**
- Modify: `README.md`

**Approach:**
- Keep: hero, title, tagline, video, footer.
- Trim nav to: Quick start · Install (macOS/Linux) · Install (Windows) · Marketplace · Full sprint · Deeper docs · Training Quest.
- Keep `## What is ATV 2.0?`.
- Restructure `## Installation` into `### macOS / Linux` and `### Windows` subsections; surface Windows gstack caveat inline.
- Update `## The Full Sprint` per Unit 3.
- Add `## Deeper Docs` pointer to `DOCS.md`.
- Remove: Three Pillars, Guided Experience, How Learning Works, De-Slop, Memory Architecture, Agents, What Gets Installed, How It Works Under the Hood, Development, Limitations, full-skill `<details>`.

**Test scenarios:**
- Happy path: `wc -l README.md` ~180–230 lines.
- Edge case: Windows reader gets PowerShell-labelled block; macOS/Linux gets bash block.
- Negative: no link points to a moved section without redirect.

**Verification:** `grep -E "^## " README.md` returns only the surviving H2s.

- [ ] **Unit 3: Refresh Full Sprint table for weekend additions**

**Goal:** Sprint map reflects current skill surface.

**Dependencies:** Unit 2.

**Files:**
- Modify: `README.md` (Full Sprint section).

**Approach:**
- Replace `/gstack-cso` → `/atv-security` under Review with caption "Config + OWASP + STRIDE in one pass".
- Add `/autoresearch` to Build with caption "Autonomous metric-driven loop".
- Add `/atv-doctor` and `/atv-update` to Reflect.
- Keep `/lfg` and `/slfg` mini-diagrams.

**Test scenarios:**
- Happy path: table renders; new skills visible.
- Edge case: `grep -n "gstack-cso" README.md` returns nothing.
- Negative: every listed skill exists under `pkg/scaffold/templates/skills/`.

- [ ] **Unit 4: Sanity sweep — links, anchors, nav consistency**

**Files:** Modify `README.md`, `DOCS.md` if needed.

**Approach:**
- Walk every relative link in new README; confirm targets exist.
- Confirm deleted README headings have anchors in `DOCS.md`.
- Update top nav so every entry corresponds to a surviving H2.
- Verify footer links resolve.

## System-Wide Impact

- Pure docs change. No code paths, skills, or scaffold templates touched.
- README is the marketplace landing page — keep What is ATV 2.0?, Installation, Full Sprint.
- Untouched: `pkg/`, `plugins/`, `npm/`, `.github/skills/`, `docs/`.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| External deep-links to README anchors break | Preserve same anchors inside `DOCS.md`. |
| Marketplace consumers read README as landing page | Keep What is ATV 2.0? + Installation + Full Sprint. |
| Listing a skill in Sprint that doesn't exist | Cross-reference Unit 3 against `pkg/scaffold/templates/skills/`. |
| Windows install diverges from macOS/Linux | Commands are identical; only shell context + gstack caveat differ. |

## Documentation / Operational Notes

- No CHANGELOG entry — pure docs reshuffle.
- No version bump.

## Sources & References

- Current `README.md` (lines 1–628).
- `pkg/scaffold/templates/skills/atv-security/SKILL.md`
- `pkg/scaffold/templates/skills/autoresearch/SKILL.md`
- `pkg/scaffold/templates/skills/atv-doctor/SKILL.md`, `atv-update/SKILL.md`
- `git log --since=2026-04-24` — weekend feature surface.
