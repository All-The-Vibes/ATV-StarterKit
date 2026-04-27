---
title: "refactor: Centralize GitHub skills + reconcile dogfood/template drift + dispose orphan plans"
type: refactor
status: completed
date: 2026-04-25
---

# refactor: Centralize GitHub skills + reconcile dogfood/template drift + dispose orphan plans

## Overview

Three loosely-related housekeeping items, batched into one plan because they
share a single review surface and a single PR's worth of risk:

1. **Reconcile skill drift** between `.github/skills/` (69 entries — the
   repo's dogfood / source-of-truth) and `pkg/scaffold/templates/skills/`
   (29 entries — the embedded payload `atv-starterkit init` actually ships).
   The `TestDogfoodTemplateParity` test currently freezes the gap as
   `pendingMirror` tech debt with 44 entries.
2. **Centralize the installable skill catalog** so users who run
   `npx atv-starterkit@latest init` get every skill we want them to have,
   and so contributors have one obvious place ("how do I ship a skill?")
   instead of two.
3. **Dispose of three untracked plan docs** that have been floating in
   `docs/plans/` across multiple sessions:
   - `2026-04-22-001-feat-cross-tool-session-skills-and-recipes-plan.md`
   - `2026-04-24-003-docs-marketing-brief-changelog-plan.md`
   - `2026-04-24-004-fix-pr26-review-fixes-plan.md`

## Problem Frame

**Drift signal.** `pkg/scaffold/parity_test.go` currently encodes a
44-entry `pendingMirror` allow-list. Every entry is, by the test's own
docstring, "real tech debt: a skill that this repo dogfoods but that
--guided users don't get." The list has been growing, not shrinking.

**Installer signal.** `pkg/scaffold/catalog.go` exposes three layer
groups (`coreSkillDirectories`, `orchestratorSkillDirectories`,
`easterEggSkillDirectories`) totaling 29 skills. The other 40 skills
under `.github/skills/` are simply unreachable from the install pipeline.
A user running `--guided` cannot opt into `frontend-design`,
`ghcp-review-resolve`, `gemini-imagegen`, `git-worktree`, `proof`,
`rclone`, `skill-creator`, `onboarding`, `reproduce-bug`,
`andrew-kane-gem-writer`, the `todo-*` suite, the `workflows-*` suite,
or the `git-*` family even though those skills are part of the project's
identity.

**Untracked-plan signal.** Three plan docs have been left untracked in
`docs/plans/` for three sessions running. Each describes work that has
since either landed (PR #26 review fixes), has a different scope owner
(cross-tool recipes targeted a different repo), or is documentation work
that was never picked up (marketing-brief changelog). Leaving them
untracked traps every future session into re-classifying them.

## Requirements Trace

- **R1.** Every skill currently under `.github/skills/` is either:
  (a) mirrored under `pkg/scaffold/templates/skills/` and registered in
  exactly one catalog layer, or (b) explicitly excluded with a one-line
  rationale in `parity_test.go`'s `templateOnly` / repo-only docs.
  No silent drift.
- **R2.** Layer taxonomy is intelligible to a contributor reading
  `catalog.go` for the first time. New layer(s) introduced for groups
  that don't fit `core` / `orchestrators` / `easter-eggs` (the existing
  three are about to absorb 30+ entries — they need a refactor, not a
  dump).
- **R3.** `--guided` install lists every newly-mirrored skill as an
  opt-in choice with a one-line description.
- **R4.** `pendingMirror` entries are removed for any skill we mirror.
  Entries that remain (because we deliberately decline to ship them)
  move to a documented `dogfoodOnly` allow-list with per-entry rationale.
- **R5.** Each of the three orphan plan docs is resolved exactly once:
  committed (with status updated), moved to an archive location, or
  deleted — with rationale captured in this plan.
- **R6.** `go build ./...` and `go test ./...` pass after the change.
  `TestDogfoodTemplateParity` continues to enforce drift.
- **R7.** README + `CHANGELOG.md` reflect the expanded skill catalog so
  users discover what's now available.

## Scope Boundaries

- **Out of scope:** rewriting any individual SKILL.md content. The mirror
  is a presence-and-wiring exercise, not a content audit. Drift in
  *content* between dogfood and template copies is explicitly accepted
  by the existing parity test ("This is a presence check only. Content
  drift between the two copies is accepted").
- **Out of scope:** changing the `npm` wrapper, GoReleaser config, or
  binary distribution path.
- **Out of scope:** reworking the `gstack` / `agent-browser` / Compound
  Engineering upstream-clone flows. Those have their own catalog seams.
- **Out of scope:** the `.remember/` and `docs/changelog/` untracked
  directories — they are session-state, not orphan plans, and have been
  intentionally left alone in prior sessions.

## Context & Research

### Relevant Code and Patterns

- `pkg/scaffold/catalog.go` — the install catalog. The three
  `*SkillDirectories` slices and `BuildFilteredCatalogForPacks`
  layer-gating are the seam to extend.
- `pkg/scaffold/parity_test.go` — the drift enforcement test. Already
  contains the canonical list of all 44 drifted skills under
  `pendingMirror`. This plan's primary effect on this file is to move
  entries from `pendingMirror` to either nothing (mirrored) or to a new
  `dogfoodOnly` map (deliberately repo-local).
- `pkg/scaffold/templates/skills/` — the embedded payload. Adding a
  skill = `cp -R` the directory + register in catalog.
- `pkg/tui/` — the `--guided` TUI that reads layer descriptions. Any new
  layer needs a label here.
- `cmd/init.go` — the install entrypoint, where layer flags are wired.

### Institutional Learnings

- `docs/solutions/workflow-issues/ralph-loop-stop-hook-blocking-session-exit-2026-04-17.md`
  — confirms the LFG pipeline expects plans to write under `docs/plans/`
  with the `YYYY-MM-DD-NNN-` convention.
- The repo's own `feat/land-ralph-loop-cleanup` PR #28 (just merged
  upstream cleanup of ralph-loop sweeps in `/land`) sets the precedent
  that "find ... -delete" cleanups must be `set -e` safe.
- The `pendingMirror` docstring already lists the two valid resolutions:
  "(1) Copy the skill into `pkg/scaffold/templates/skills/<name>/` and
  register it in catalog.go (then remove the entry here), or (2) Remove
  the unused `.github/skills/<name>/` directory entirely." This plan
  follows that guidance.

### External References

None — this is fully internal restructuring; no upstream API surface or
third-party docs are relevant.

## Key Technical Decisions

- **Triage by intent, not bulk-mirror.** Of the 44 `pendingMirror`
  skills, several are stale aliases (`workflows-*` are deprecated
  shims for the `ce-*` skills; `ce-work-beta` is a beta of `ce-work`;
  `resolve-pr-feedback` / `resolve-pr-parallel` / `resolve_parallel` /
  `resolve_todo_parallel` overlap; `report-bug` and `report-bug-ce`
  overlap; `create-agent-skill` and `create-agent-skills` are duplicates
  of `skill-creator`). Each gets a triage decision: **MIRROR**,
  **DELETE-DOGFOOD**, or **KEEP-DOGFOOD-ONLY** (with rationale). We do
  not mass-`cp -R`.

  *Rationale:* mass-mirror would ship deprecated and duplicate skills to
  every user, polluting the slash-command namespace and contradicting
  the `unslop` / Karpathy guidelines we already ship.

- **New layer taxonomy.** `core-skills` is already overloaded (planning
  + lifecycle + learning + quality + behavioral + security all in one).
  Introduce three new layer groups so the `--guided` TUI surfaces
  meaningful choices:
  - `dev-tools` — git/worktree/commit/PR helpers, onboarding,
    skill-creator, reproduce-bug.
  - `style-skills` — dhh-rails-style, andrew-kane-gem-writer,
    every-style-editor, frontend-design.
  - `media-skills` — gemini-imagegen, proof, rclone, feature-video
    (already in orchestrators — leave it there).

  *Rationale:* keeps `core-skills` focused on the planning/lifecycle
  pipeline, gives users meaningful per-layer toggles, and matches the
  README's existing "three pillars" framing without breaking it.

- **Orphan-plan triage.**
  - `2026-04-22-001-feat-cross-tool-session-skills-and-recipes-plan.md`
    → **commit as `status: superseded`** with a forward-pointer to this
    plan and the relevant `pkg/scaffold/templates/` skills that
    landed via PRs #25/#27 (`/land`, `/takeoff`). Reason: the plan's
    intent (cross-tool session skills) was partially executed via a
    different route; preserving the doc captures the design history.
  - `2026-04-24-003-docs-marketing-brief-changelog-plan.md` →
    **commit as `status: deferred`** with rationale that the marketing
    brief is a one-off doc artifact, not blocking work. The plan can
    be executed later via `/ce:work` against the file directly.
  - `2026-04-24-004-fix-pr26-review-fixes-plan.md` → **commit as
    `status: completed`** with a pointer to PR #26 (merged
    `2026-04-24T17:01:47Z`). Reason: the work *did* land; the plan was
    just never status-updated.

  *Rationale:* deletion loses history; archival inside `docs/plans/`
  with status frontmatter is exactly the pattern already used by the
  other 13 tracked plans in that directory.

- **Drift enforcement stays.** `TestDogfoodTemplateParity` is the
  contract that prevents this from happening again. The plan keeps it
  green by trimming `pendingMirror` to zero entries and adding a
  `dogfoodOnly` map for any skill we deliberately keep repo-local.

## Open Questions

### Resolved During Planning

- **Q: Should we mirror `ghcp-review-resolve`?** A: Yes — it's actively
  dogfooded (PRs #23/#26) and explicitly listed in
  `feat/land-ralph-loop-cleanup`'s allow-list. MIRROR. Belongs in
  `dev-tools` layer.
- **Q: Should we mirror the `workflows-*` aliases?** A: No —
  `workflows-work`'s description literally says
  `"[DEPRECATED] Use /ce:work instead"`. DELETE-DOGFOOD for all five
  (`workflows-brainstorm`, `workflows-compound`, `workflows-plan`,
  `workflows-review`, `workflows-work`).
- **Q: Should we mirror `ce-work-beta`?** A: No — it's a beta variant
  of an already-shipped skill. KEEP-DOGFOOD-ONLY for now, with a
  deletion follow-up once external-delegate mode lands in `ce-work`
  proper. Add to `dogfoodOnly`.
- **Q: Should we mirror `agent-browser`?** A: No — it's installed
  separately by `cmd/init.go` via the upstream Vercel project, not as
  a markdown skill. DELETE-DOGFOOD (the `.github/skills/agent-browser/`
  copy is stale documentation duplication).
- **Q: Should we mirror `agent-native-architecture` and
  `agent-native-audit`?** A: Yes for `agent-native-architecture` (it's
  an upstream Compound Engineering skill that pairs with our agentic
  positioning). MIRROR to `core-skills`. No for `agent-native-audit`
  (CE-internal). DELETE-DOGFOOD.

### Deferred to Implementation

- Exact one-line description text for each new layer choice in the TUI
  — pull from each SKILL.md's frontmatter `description:` field at
  implementation time.
- Whether `compound-docs` and `deploy-docs` are still relevant or are
  superseded by `ce-compound` / `ce-compound-refresh` — verify by
  reading the SKILL.md content during the triage unit; route to MIRROR
  or DELETE-DOGFOOD accordingly.

## High-Level Technical Design

> *This illustrates the intended approach and is directional guidance
> for review, not implementation specification. The implementing agent
> should treat it as context, not code to reproduce.*

**Triage decision matrix** (applied per skill in Unit 1):

| Skill class | Examples | Decision |
|---|---|---|
| Active dogfood, broadly useful | `ghcp-review-resolve`, `git-worktree`, `git-commit`, `git-commit-push-pr`, `frontend-design`, `onboarding`, `reproduce-bug`, `skill-creator`, `agent-native-architecture` | **MIRROR** → new layer |
| Deprecated alias | `workflows-*`, `create-agent-skill`, `create-agent-skills`, `resolve_parallel` | **DELETE-DOGFOOD** |
| Beta / internal-only | `ce-work-beta`, `agent-native-audit`, `heal-skill`, `report-bug`, `report-bug-ce` | **KEEP-DOGFOOD-ONLY** (`dogfoodOnly` allow-list) |
| Stale duplicate | `agent-browser` (already installed via separate flow) | **DELETE-DOGFOOD** |
| Style/media skills | `dhh-rails-style`, `andrew-kane-gem-writer`, `every-style-editor`, `gemini-imagegen`, `proof`, `rclone` | **MIRROR** → `style-skills` / `media-skills` layer |

**Layer wiring shape:**

```
catalog.go
  coreSkillDirectories         // unchanged: planning, lifecycle, learning, quality, behavioral, security
  orchestratorSkillDirectories // unchanged: lfg, ralph-loop, slfg, ...
  easterEggSkillDirectories    // unchanged: meme-iq
  devToolsSkillDirectories     // NEW: git-*, worktree, onboarding, ghcp-review-resolve, skill-creator, reproduce-bug
  styleSkillDirectories        // NEW: dhh-rails-style, andrew-kane-gem-writer, every-style-editor, frontend-design
  mediaSkillDirectories        // NEW: gemini-imagegen, proof, rclone

BuildFilteredCatalogForPacks
  + layer "dev-tools"   → devToolsSkillDirectories
  + layer "style-skills" → styleSkillDirectories
  + layer "media-skills" → mediaSkillDirectories
```

**Orphan-plan workflow** (Unit 4):

```
for each orphan plan:
  edit frontmatter status: active → (completed | superseded | deferred)
  add a one-paragraph "Resolution" section at top of body
  git add docs/plans/<file>
single commit:  docs(plans): triage three orphan planning docs
```

## Implementation Units

- [ ] **Unit 1: Triage every drifted skill**

**Goal:** Produce a single triage table (markdown comment in
`parity_test.go` or a short doc under `docs/changelog/`) that classifies
each of the 44 `pendingMirror` entries as MIRROR / DELETE-DOGFOOD /
KEEP-DOGFOOD-ONLY with one-line rationale per entry.

**Requirements:** R1, R2, R4

**Dependencies:** none

**Files:**
- Modify: `pkg/scaffold/parity_test.go` (move entries off `pendingMirror`,
  introduce `dogfoodOnly` map with rationale comments)
- Create: `docs/changelog/2026-04-25-skill-triage.md` (decision log;
  one row per skill, MIRROR/DELETE/KEEP-ONLY decision, one-line reason)

**Approach:**
- Read each SKILL.md's frontmatter `description:` field as the
  primary signal for "is this useful to ship?"
- Apply the decision matrix from High-Level Technical Design.
- Capture rationale per skill in the changelog doc; the parity test
  carries only short comments.

**Patterns to follow:**
- The existing `templateOnly` map's per-entry comment style in
  `parity_test.go`.
- `docs/changelog/` already exists as untracked — committing this
  decision log establishes its purpose.

**Test scenarios:**
- *Test expectation: none — this unit produces a decision artifact (a
  markdown doc + comment edits in a test file). Behavior verification
  for the resulting wiring lives in Units 2 and 3.*

**Verification:**
- The triage doc covers all 44 skills with decisions.
- A reviewer can answer "what happens to skill X?" by grepping the doc.

---

- [ ] **Unit 2: Mirror MIRROR-class skills + register catalog layers**

**Goal:** For every skill marked MIRROR in Unit 1, copy
`.github/skills/<name>/` → `pkg/scaffold/templates/skills/<name>/` and
register it in the appropriate `*SkillDirectories` slice.

**Requirements:** R1, R2, R3

**Dependencies:** Unit 1

**Files:**
- Create: `pkg/scaffold/templates/skills/<name>/SKILL.md` for each
  MIRROR skill (filenames driven by Unit 1 output; expect ~15-20 dirs)
- Modify: `pkg/scaffold/catalog.go` (add `devToolsSkillDirectories`,
  `styleSkillDirectories`, `mediaSkillDirectories`; wire each into
  `BuildFilteredCatalogForPacks` under matching layer keys)
- Modify: `pkg/scaffold/parity_test.go` (remove mirrored entries from
  `pendingMirror`)
- Test: `pkg/scaffold/parity_test.go` (TestDogfoodTemplateParity,
  TestSkillDirectoryParity already cover the wiring; add a
  TestNewLayersExposeSkills that exercises each new layer key)

**Approach:**
- Copy SKILL.md verbatim (no content edits — that's the parity test's
  contract).
- Use `cp -R` semantics in implementation; preserve any sub-files in
  the dogfood skill dir (most SKILLs are SKILL.md only, but some have
  references/ subdirs).
- Register exactly once per skill — `TestSkillDirectoryParity` enforces
  no double-registration.

**Patterns to follow:**
- `coreSkillDirectories` for the slice declaration shape.
- `BuildFilteredCatalogForPacks` for layer-flag handling — mirror the
  existing `if layerSet["core-skills"]` pattern.

**Test scenarios:**
- *Happy path:* `TestDogfoodTemplateParity` passes with a smaller
  `pendingMirror` (only DELETE-DOGFOOD and KEEP-DOGFOOD-ONLY entries
  remain temporarily; cleared further in Unit 3).
- *Happy path:* `TestSkillDirectoryParity` passes — every new template
  dir is registered exactly once.
- *Happy path (new test):* `TestNewLayersExposeSkills` —
  `BuildFilteredCatalog(StackGeneral, []string{"dev-tools"})` returns
  components for each MIRROR skill assigned to `dev-tools`. Same for
  `style-skills` and `media-skills`.
- *Edge case (new test):* selecting a layer that does not exist
  produces no skill components (no panic, no false positives).
- *Integration:* `BuildFilteredCatalog(StackGeneral, []string{"core-skills","dev-tools"})`
  returns the union without duplicates.

**Verification:**
- `go test ./pkg/scaffold/...` passes.
- `go build ./...` passes.
- `pendingMirror` shrinks by the count of MIRROR skills.

---

- [ ] **Unit 3: Remove DELETE-DOGFOOD skills + freeze KEEP-DOGFOOD-ONLY**

**Goal:** Delete the `.github/skills/<name>/` directory for every
DELETE-DOGFOOD entry in Unit 1's triage. Move every KEEP-DOGFOOD-ONLY
entry from `pendingMirror` into a new `dogfoodOnly` map in
`parity_test.go`, with a per-entry rationale comment. After this unit,
`pendingMirror` is empty (and removed from the test file).

**Requirements:** R1, R4, R6

**Dependencies:** Unit 2

**Files:**
- Delete: `.github/skills/<name>/` for each DELETE-DOGFOOD entry
- Modify: `pkg/scaffold/parity_test.go` (introduce `dogfoodOnly` map,
  remove `pendingMirror` block, update logic to assert
  `dogfoodSkill ∉ templateSkills` for each `dogfoodOnly` entry)
- Modify: `.github/copilot-instructions.md` if any deleted skill is
  referenced there (greppable)

**Approach:**
- Use `git rm -r .github/skills/<name>/` per delete (one commit per
  ~5 deletes for review tractability is fine; final PR squashes
  optionally).
- For `dogfoodOnly`, mirror the comment style of `templateOnly`. Each
  entry needs a one-line `// reason: ...` comment.

**Patterns to follow:**
- `templateOnly` block in `parity_test.go` is the model for
  `dogfoodOnly`.

**Test scenarios:**
- *Happy path:* `TestDogfoodTemplateParity` passes with no
  `pendingMirror` and a populated `dogfoodOnly`.
- *Edge case:* a skill listed in `dogfoodOnly` but accidentally
  re-mirrored under `templates/skills/` causes test failure with a
  clear message: "skill X is in dogfoodOnly but also exists as a
  template — pick one."
- *Edge case:* a `dogfoodOnly` entry with no corresponding
  `.github/skills/<name>/` directory causes test failure (stale
  allow-list).
- *Integration:* `go test ./...` passes.

**Verification:**
- `pendingMirror` no longer exists in the test file.
- `dogfoodOnly` documents every deliberately repo-local skill.
- No stale references to deleted skills in
  `.github/copilot-instructions.md` or README.

---

- [ ] **Unit 4: Triage the three orphan plan docs**

**Goal:** Resolve the three currently-untracked plan docs in
`docs/plans/` per the Key Technical Decisions table. Each gets a status
update, a one-paragraph Resolution section at the top of the body, and
a single commit landing all three.

**Requirements:** R5

**Dependencies:** none (independent of skill-mirror work)

**Files:**
- Modify: `docs/plans/2026-04-22-001-feat-cross-tool-session-skills-and-recipes-plan.md`
  — frontmatter `status: active` → `status: superseded`; add Resolution
  section pointing at PRs #25/#27 and the new plan in `2026-04-25-003`.
- Modify: `docs/plans/2026-04-24-003-docs-marketing-brief-changelog-plan.md`
  — frontmatter `status: active` → `status: deferred`; Resolution
  section noting the doc is unblocked but not currently scheduled.
- Modify: `docs/plans/2026-04-24-004-fix-pr26-review-fixes-plan.md`
  — frontmatter `status: active` → `status: completed`; Resolution
  section pointing at merged PR #26.

**Approach:**
- Edit frontmatter only; do not rewrite plan body.
- Resolution section is 2-4 sentences max, prefixed with
  `> **Resolution (2026-04-25):**` so it visually separates from the
  original plan body.

**Patterns to follow:**
- The existing `2026-04-24-006-feat-land-ralph-loop-cleanup-plan.md`
  was set to `status: completed` last session — same shape.

**Test scenarios:**
- *Test expectation: none — these are documentation edits with no
  behavioral change. Verification is reviewer-readable: the three plans
  are tracked, each carries an explicit status, and a future session
  cannot re-classify them as orphans.*

**Verification:**
- `git status` shows no untracked files under `docs/plans/`.
- Each plan's frontmatter `status:` matches its triage decision.

---

- [ ] **Unit 5: Update README + CHANGELOG to surface new layer choices**

**Goal:** README's installation section advertises the new layer choices.
`CHANGELOG.md` records the catalog expansion in the next-version entry.

**Requirements:** R3, R7

**Dependencies:** Units 2 and 3

**Files:**
- Modify: `README.md` — extend the "Pro" / `--guided` section to list
  `dev-tools`, `style-skills`, `media-skills` as opt-in layers.
- Modify: `CHANGELOG.md` — add an entry under the next unreleased
  version that lists the newly-installable skills (group by new layer)
  and notes the deletion of deprecated dogfood-only skills.

**Approach:**
- Skill-list table in CHANGELOG should match the form already used
  for v2.5.7's Karpathy entry.
- README addition is a single subsection under the existing
  "Installation" / TUI section — no full rewrite.

**Patterns to follow:**
- v2.5.7 CHANGELOG entry style (Karpathy Guidelines).
- README's existing layered-install description (currently calls out
  `core-skills`, `orchestrators`, `easter-eggs`).

**Test scenarios:**
- *Test expectation: none — pure docs change. Reviewer-verified.*

**Verification:**
- README mentions every new layer name.
- CHANGELOG entry lists every newly-shipping skill.

---

- [ ] **Unit 6: Final wiring smoke test + parity green**

**Goal:** End-to-end smoke that the installer actually emits the new
files when `--layers dev-tools,style-skills,media-skills` is requested.

**Requirements:** R6

**Dependencies:** Units 1-5

**Files:**
- Test: `pkg/scaffold/parity_test.go` — extend
  `TestNewLayersExposeSkills` (added in Unit 2) with assertions on the
  exact set of skill paths each layer emits.
- Optional smoke: `test/sandbox/` — if there's an existing sandbox
  install fixture, exercise it; otherwise document the manual
  verification command in the PR body.

**Approach:**
- Compute expected skill paths from the catalog slices declared in
  Unit 2; assert presence in `BuildFilteredCatalog` output.
- Lightweight; no external process spawned.

**Patterns to follow:**
- `TestCoreLayerShipsLandAndTakeoff` — same shape, different layer keys.

**Test scenarios:**
- *Happy path:* each new layer key returns the exact expected set of
  `.github/skills/<name>/SKILL.md` paths.
- *Edge case:* selecting all layers simultaneously returns the union
  without duplicates.
- *Edge case:* selecting an empty layer list returns no skill
  components.
- *Integration:* `go build ./... && go vet ./... && go test ./...` all
  pass.

**Verification:**
- All `pkg/scaffold` tests pass.
- `pendingMirror` is gone, `dogfoodOnly` is honest, parity is enforced.

## System-Wide Impact

- **Interaction graph:** `cmd/init.go` reads `BuildFilteredCatalogForPacks`,
  the TUI in `pkg/tui/` reads layer descriptions, and
  `pkg/installstate/manifest_test.go` snapshots install output. Each of
  these surfaces will see the new layer keys and new skill paths.
- **Error propagation:** `skillComponents` panics on FS walk error
  (existing behavior); adding more dirs does not change this contract.
- **State lifecycle risks:** `pkg/installstate/snapshot.go` checksums
  installed files; users who previously installed and customized a now-
  mirrored skill might see new "untouched template" entries appear in
  their manifest. This is benign — uninstall already preserves
  customized files via checksum diff.
- **API surface parity:** the npm wrapper (`npm/`) does not need
  changes; it downloads the GoReleaser binary which embeds the new
  templates automatically.
- **Integration coverage:** Unit 6's
  `TestNewLayersExposeSkills` covers the new wiring; existing
  `TestDogfoodTemplateParity` and `TestSkillDirectoryParity` continue to
  enforce drift.
- **Unchanged invariants:** the three existing layer keys (`core-skills`,
  `orchestrators`, `easter-eggs`) keep their current skill lists. No
  user who is currently relying on `--layers core-skills` will see a
  smaller install. The existing default install (no `--layers` flag,
  `BuildCatalog`) already includes every template skill and will
  automatically pick up the new ones.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Mass-mirror ships stale or duplicate skills | Unit 1 triage with explicit DELETE-DOGFOOD verdict per stale entry; Unit 3 actually deletes them. |
| Deleting `.github/skills/<name>/` breaks an internal workflow that secretly depended on it | Grep `.github/copilot-instructions.md`, `.github/workflows/`, and `pkg/scaffold/templates/instructions/` for each deleted skill name in Unit 3. |
| Layer-key change breaks an external consumer of `--layers` flag | `--layers` is a 2.x-era flag; existing keys are preserved, only new keys are added. No breaking change. |
| Untracked plan docs contain content we still need | Unit 4 preserves bodies verbatim — only frontmatter and a Resolution preamble change. |
| `dogfoodOnly` allow-list silently grows like `pendingMirror` did | Stale-entry check (modeled on existing `staleTemplateOnly` / `stalePendingMirror` checks) fails CI when an entry is unused. |

## Documentation / Operational Notes

- Update `README.md` to advertise new layer choices (Unit 5).
- Update `CHANGELOG.md` under the next unreleased section (Unit 5).
- No migration runbook needed; users get the new skills on next
  `npx atv-starterkit@latest init` or `--upgrade`.
- No monitoring change.

## Sources & References

- Working tree state: branch `feat/land-ralph-loop-cleanup`, version
  `VERSION` = `2.5.7`, last commit `c7fcf81`.
- Drift evidence: `pkg/scaffold/parity_test.go` `pendingMirror` block.
- Catalog seam: `pkg/scaffold/catalog.go`, `BuildFilteredCatalogForPacks`.
- Related merged PRs: #23, #24, #25, #26, #27, #28.
- Orphan plans:
  - `docs/plans/2026-04-22-001-feat-cross-tool-session-skills-and-recipes-plan.md`
  - `docs/plans/2026-04-24-003-docs-marketing-brief-changelog-plan.md`
  - `docs/plans/2026-04-24-004-fix-pr26-review-fixes-plan.md`
