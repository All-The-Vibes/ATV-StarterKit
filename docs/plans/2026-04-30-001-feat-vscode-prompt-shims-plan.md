---
title: "feat: Ship .github/prompts/*.prompt.md shims so VS Code Copilot Chat surfaces ATV slash commands"
type: feat
status: completed
date: 2026-04-30
issue: 36
---

# feat: Ship `.github/prompts/*.prompt.md` shims so VS Code Copilot Chat surfaces ATV slash commands

## Overview

The ATV installer ships skills as `.github/skills/<name>/SKILL.md`. Claude Code and Copilot CLI discover those natively, but **VS Code Copilot Chat** discovers slash commands from `.github/prompts/*.prompt.md`. Today the installer ships zero prompt files, so `/ce-brainstorm`, `/ce-plan`, `/lfg`, `/learn`, etc. don't appear in the VS Code chat picker even though `copilot-instructions.md` advertises them.

Fix: generate a thin `.prompt.md` shim per user-facing skill at install time. Each shim delegates to the canonical `SKILL.md` so there is one source of truth. Wire it through the existing skill-catalog/parity-test machinery so a new user-facing skill cannot land without a corresponding shim.

## Problem Frame

Reported in [#36](https://github.com/All-The-Vibes/ATV-StarterKit/issues/36) by @stephschofield (2026-04-30). After `npx atv-starterkit init` (or VS Code source install), users opening the project in VS Code Copilot Chat see no completions for `/ce-` or any documented workflow. They either rediscover workflows by typing prose or conclude the install is broken. Verified workaround: hand-rolling 15 `.prompt.md` shims into `.github/prompts/` immediately surfaces every command.

The installer already *recognizes* prompt files exist — `pkg/installstate/recommendations.go` counts them, `pkg/monitor/watcher.go` watches the directory, and there is even an `add-prompts` recommendation that fires when the count is zero — but no template is shipped. The recommendation reproduces the issue: the installer tells the user to create prompt files instead of just shipping them.

## Requirements Trace

- **R1.** After a fresh `--guided` install (or full install) on a VS Code-targeted project, every user-facing ATV slash command listed in `copilot-instructions.md` appears in the VS Code Copilot Chat slash picker.
- **R2.** Each generated `.prompt.md` is a thin delegation shim — the `SKILL.md` remains the single source of truth for behavior. No skill instructions are duplicated.
- **R3.** Sub-skills that are not meant to be invoked directly (e.g., `document-review`, `deepen-plan`, `setup`, `karpathy-guidelines`) do **not** receive shims and do not pollute the chat picker.
- **R4.** Adding a new user-facing skill template without also wiring its prompt shim fails a parity test, mirroring the existing `TestSkillDirectoryParity` guard.
- **R5.** The install summary acknowledges shipped prompts (so users know slash commands are wired up), and the stale `add-prompts` recommendation no longer fires when the installer's own shims are present.
- **R6.** The agentic-primitives table in `README.md`, `npm/README.md`, and `docs/brainstorms/2026-03-11-agentic-coding-starter-kit-installer-brainstorm.md` documents prompt files as a primitive (currently lists 1–6; this adds the 7th).

## Scope Boundaries

- **In scope:** generating shim `.prompt.md` files for user-facing skills, wiring them into the catalog, parity tests, install summary, removing the now-misleading `add-prompts` recommendation.
- **Out of scope:** rewriting skill content, changing how Claude Code or Copilot CLI discover skills, building richer prompt-file templates with per-skill front-matter beyond what's needed for VS Code discovery, optional `/atv-doctor` drift check (issue suggests it; defer to a follow-up).
- **Out of scope:** shimming sub-skills like `document-review` or auto-generating shims from skill front-matter — explicit allow-list is intentional to keep the chat picker clean.
- **Out of scope:** changes to gstack-prefixed skills under `.github/skills/gstack-*` — gstack manages its own surface (see existing memory: gstack copies generated skills with `gstack-` prefix).

## Context & Research

### Relevant Code and Patterns

- `pkg/scaffold/catalog.go` — entry point. Contains `BuildCatalog`, `BuildFilteredCatalog`, the three skill allow-lists (`coreSkillDirectories`, `orchestratorSkillDirectories`, `easterEggSkillDirectories`), and `skillComponents()` which walks `templates/skills/`. The new prompt generation will mirror this pattern with its own selector function and component builder.
- `pkg/scaffold/templates/skills/` — the embedded source of truth. New sibling directory: `pkg/scaffold/templates/prompts/` (or generated content from a single template constant — see Key Technical Decisions).
- `pkg/scaffold/parity_test.go` — `TestSkillDirectoryParity` and `TestDogfoodTemplateParity` show exactly how to enforce that catalog allow-lists stay aligned with the embedded filesystem and with `.github/skills/`. New parity test will follow the same shape for prompts.
- `pkg/installstate/recommendations.go:81,246-251,457-465` — counts `.prompt.md`, recommends `add-prompts`, exposes `ListPromptFiles`. The `add-prompts` recommendation needs to be either deleted or reframed (e.g., suggest user-authored prompts only, gate on a directory of files we didn't ship).
- `pkg/monitor/watcher.go:124,240,320,330,357,365` — already watches `.github/prompts/` and surfaces `PromptCount` to repo state. No change needed; will Just Work once we ship files.
- `pkg/scaffold/templates/instructions/{rails,python,typescript,general}.md` — the four canonical instruction templates list every user-facing slash command. These are the authoritative reference for the allow-list of skills that need shims.
- `test/sandbox/sandbox_test.go` — existing pattern for asserting installed file presence in a temp repo. New sandbox assertion can confirm `.github/prompts/<skill>.prompt.md` is created for the expected set after a guided/full install.
- Existing parity test (`TestDogfoodTemplateParity`) implies templates must be mirrored from `.github/skills/` — for prompts, since they are pure generated shims and the dogfooded repo does not currently have `.github/prompts/`, we explicitly dogfood the same shims into `.github/prompts/` of this repo to preserve the "what we ship is what we use" invariant.

### Institutional Learnings

- `docs/solutions/` — no existing solution covers VS Code Copilot Chat discovery semantics. This plan implicitly creates one (recommend `/ce-compound` after merge).
- Memory: *"Installer skills are guarded by parity tests: every templates/skills dir must be registered in exactly one catalog slice, and .github/skills must be mirrored into templates/skills"* — directly applicable. We extend the same posture to prompts.
- Memory: *"Guided installs only include skills listed in coreSkillDirectories or orchestratorSkillDirectories"* — confirms layer-gating model; prompt shims should follow the same layer system so a user who excludes the `orchestrators` layer doesn't get `/lfg` in their picker.

### External References

- VS Code Copilot Chat docs (well-established public behavior): slash commands are discovered from `.github/prompts/*.prompt.md` and `~/.config/.../prompts/`. No external research needed beyond the contract documented in the issue and the workaround that the issue reporter verified end-to-end.

## Key Technical Decisions

| Decision | Rationale |
|---|---|
| **Static, generated shims with a fixed template body** rather than per-skill custom prompt files | Each shim is ~10 lines and identical except for the skill name. A single Go template + an allow-list keeps maintenance to one place and removes the risk of skill/prompt drift. Custom per-skill prompt content would force two sources of truth. |
| **Explicit allow-list (`promptShimSkillDirectories`)** rather than auto-shimming every skill | Sub-skills (`document-review`, `deepen-plan`, `setup`, `karpathy-guidelines`, `brainstorming`) are not meant to surface as user commands. Shimming them would clutter the picker and invite misuse. Explicit allow-list also gives a single auditable source for "what is a user-facing slash command." |
| **Generate at scaffold time from a Go template, not as embedded `.prompt.md` template files** | Shim content is derived purely from the skill name. Embedding 18 nearly-identical files would create churn for any future shim shape change. A Go template + allow-list is more honest. (Open-question fork — see below — covers the alternative where we embed files for parity-test simplicity.) |
| **Tie shims to existing layer selections** (`core-skills`, `orchestrators`) rather than introducing a new `prompts` layer | Users who skip orchestrators expect no `/lfg` anywhere — including the picker. Re-using layer logic preserves that intent and avoids a new TUI category. |
| **Dogfood the shims into this repo's `.github/prompts/`** | The repo already dogfoods skills under `.github/skills/`. A new `TestDogfoodPromptParity` test enforces the mirror. Without dogfooding, the maintainers don't get the same VS Code picker experience as users. |
| **Remove (not just suppress) the `add-prompts` recommendation** | The recommendation no longer makes sense once shims are shipped by default — its presence after install would be a red herring. Reframing it ("create *additional* prompt files for your own workflows") is possible but adds noise; deletion is cleaner. |
| **No new dependency on skill front-matter parsing** | The shim's `description:` line can hardcode a generic delegation message ("Run the <name> skill"). Reading every `SKILL.md` front-matter at scaffold time would couple the installer to skill internals and invite parser brittleness. |

## Open Questions

### Resolved During Planning

- **Which skills are user-facing?** Resolved: take the union of slash commands listed in `pkg/scaffold/templates/instructions/general.md` (the most complete reference) plus the orchestrator entry `/lfg`. Concretely: `ce-brainstorm`, `ce-plan`, `ce-work`, `ce-review`, `ce-compound`, `ce-ideate`, `takeoff`, `land`, `learn`, `instincts`, `evolve`, `observe`, `unslop`, `autoresearch`, `atv-security`, `atv-doctor`, `atv-update`, `lfg`. Excludes: `document-review`, `deepen-plan`, `setup`, `karpathy-guidelines`, `brainstorming`, `meme-iq`, `ralph-loop`, `resolve_todo_parallel`, `slfg`, `feature-video`, `test-browser`, `ce-compound-refresh`. Implementation may add comments justifying each exclusion to satisfy parity-test diagnostics.
- **Where do shim files live in the embed FS?** Resolved: generate at scaffold time from a Go template constant in `pkg/scaffold/prompts.go`. No embedded `.prompt.md` files. The dogfooded copies under `.github/prompts/` are written from the same template via a small generator (or `go:generate` directive — see deferred items).
- **Layer gating?** Resolved: prompt shims for core skills are gated on `core-skills`; the `lfg` shim (and any future orchestrator shims) is gated on `orchestrators`.

### Deferred to Implementation

- **Exact shim template wording** — the issue proposes a 10-line Markdown body; final wording (especially the `description:` line that VS Code surfaces in the picker) should be tuned during implementation by previewing against the actual VS Code chat picker. The plan fixes the *shape*, not the *prose*.
- **`go:generate` vs runtime generation for the dogfooded `.github/prompts/` files** — implementer should pick whichever keeps the parity test trivially green. If runtime generation, the dogfood files are written by a small `make` target or `go run`; if `go:generate`, they're committed and verified for parity. Decision can be made when the test is wired up.
- **Whether `meme-iq` deserves a shim** — currently treated as easter-egg; if user feedback wants `/meme` discoverable, add it to the allow-list later. Not in scope here.
- **`/atv-doctor` drift check** — the issue suggests warning when a skill exists with no shim. Useful but separable; tracked as a follow-up.

## Implementation Units

- [x] **Unit 1: Define the prompt-shim allow-list and template**

**Goal:** Establish the single source of truth for which skills get prompt shims and what each shim looks like.

**Requirements:** R1, R2, R3.

**Dependencies:** None.

**Files:**
- Create: `pkg/scaffold/prompts.go`
- Test: `pkg/scaffold/prompts_test.go`

**Approach:**
- Add a `promptShimSkillDirectories []string` slice grouped by category (core vs orchestrator) with inline comments explaining each entry's user-facing role.
- Define a Go text template that produces a `.prompt.md` body taking only the skill directory name as input. Body delegates to `.github/skills/<name>/SKILL.md` and forwards user arguments verbatim, per the issue's verified workaround.
- Export `BuildPromptShim(skillName string) []byte` so other packages (and the dogfood generator) reuse the same template.

**Patterns to follow:**
- `coreSkillDirectories` / `orchestratorSkillDirectories` slices in `pkg/scaffold/catalog.go:165-209`.

**Test scenarios:**
- Happy path: `BuildPromptShim("ce-plan")` returns bytes containing `mode: agent`, a `description:` line referencing `ce-plan`, and a Markdown body referencing `.github/skills/ce-plan/SKILL.md`.
- Edge case: skill name with hyphens (`ce-brainstorm`) round-trips into both the YAML front-matter and the link target without escaping.
- Happy path: every entry in `promptShimSkillDirectories` is a substring of the union of `coreSkillDirectories` and `orchestratorSkillDirectories` — i.e., we never claim to shim a skill that isn't shipped.

**Verification:**
- Unit test passes; running `go vet ./...` is clean.

- [x] **Unit 2: Wire shims into the install catalog (full + filtered)**

**Goal:** Make the installer actually emit `.github/prompts/*.prompt.md` for selected skills.

**Requirements:** R1, R3.

**Dependencies:** Unit 1.

**Files:**
- Modify: `pkg/scaffold/catalog.go` (add `promptShims()` / `promptShimsForLayers()` and call from `BuildCatalog` and `BuildFilteredCatalogForPacks`)
- Modify: `pkg/scaffold/prompts.go` (add component builder)
- Test: `pkg/scaffold/catalog_test.go` (or new `pkg/scaffold/prompts_test.go`)

**Approach:**
- New helper builds a `[]Component` keyed off `promptShimSkillDirectories`, gated by which skill layers were selected so a user who deselects orchestrators gets no `/lfg.prompt.md`.
- Add the `.github/prompts` directory to the `directories()` list (so the install creates it even if no shims are emitted, mirroring `.github/skills`).
- Components carry a sensible `HookType` (likely `1` for system instructions surface or a new constant; see deferred item below for ergonomics).

**Patterns to follow:**
- `skillComponents(selected map[string]bool)` in `pkg/scaffold/catalog.go:242-266`.
- Layer gating in `BuildFilteredCatalogForPacks` at `pkg/scaffold/catalog.go:98-132`.

**Test scenarios:**
- Happy path: `BuildCatalog(StackGeneral)` produces a component for `.github/prompts/ce-plan.prompt.md` whose content matches `BuildPromptShim("ce-plan")`.
- Happy path: `BuildFilteredCatalog(StackGeneral, []string{"core-skills"})` includes `.github/prompts/ce-plan.prompt.md` but **not** `.github/prompts/lfg.prompt.md`.
- Happy path: `BuildFilteredCatalog(StackGeneral, []string{"orchestrators"})` includes `.github/prompts/lfg.prompt.md` but no core-skill prompts.
- Edge case: `BuildFilteredCatalog(StackGeneral, []string{})` (no layers) emits zero prompt components.
- Integration: snapshot the set of prompt-component paths and assert it equals the allow-list filtered by selected layers — guards against silent additions.

**Verification:**
- Existing scaffold and parity tests still pass; new tests pass; running `BuildCatalog` end-to-end produces the expected prompt files when written by `pkg/scaffold/scaffold.go` (covered in Unit 4).

- [x] **Unit 3: Add parity tests**

**Goal:** Lock the invariants so future drift is caught at test time, not by users on Discord.

**Requirements:** R3, R4.

**Dependencies:** Units 1 and 2.

**Files:**
- Modify: `pkg/scaffold/parity_test.go`

**Approach:**
- `TestPromptShimAllowListSubsetOfSkills` — every entry in `promptShimSkillDirectories` is present in `coreSkillDirectories ∪ orchestratorSkillDirectories ∪ easterEggSkillDirectories`. Failure message names the offending entry and points to `pkg/scaffold/prompts.go`.
- `TestPromptShimExclusionsAreIntentional` — for each skill in the union not in the allow-list, assert it is in a small declared `nonUserFacingSkills` set (also in `pkg/scaffold/prompts.go`). Forces a deliberate decision when adding a new skill: either add to the shim allow-list or add to the exclusion list with a reason comment.
- `TestDogfoodPromptParity` — every shim emitted by `BuildPromptShim` for entries in the allow-list exists at `.github/prompts/<name>.prompt.md` in the repo, with content equal to the generator output. Mirrors `TestDogfoodTemplateParity`.

**Patterns to follow:**
- `TestSkillDirectoryParity` and `TestDogfoodTemplateParity` in `pkg/scaffold/parity_test.go:55-...`.

**Test scenarios:**
- Happy path: tests pass with the implemented allow-list and the dogfooded files.
- Error path (synthetic): adding a fake skill directory to the embedded FS without registering it in the allow-list or exclusion list causes `TestPromptShimExclusionsAreIntentional` to fail with a clear message.
- Error path (synthetic): mutating one of the dogfooded `.github/prompts/*.prompt.md` files causes `TestDogfoodPromptParity` to fail.

**Verification:**
- All parity tests green on the new branch; intentional regressions (e.g., temporarily removing a shim) flip the relevant test red.

- [x] **Unit 4: Dogfood shims into this repo's `.github/prompts/`**

**Goal:** This repo gets the same VS Code Copilot Chat experience users do, and the parity test from Unit 3 has files to compare against.

**Requirements:** R1 (for maintainers), R4.

**Dependencies:** Units 1 and 3.

**Files:**
- Create: `.github/prompts/<skill>.prompt.md` for every skill in `promptShimSkillDirectories` (~18 files)
- Optionally modify: `Makefile` or add `go:generate` directive in `pkg/scaffold/prompts.go` for regenerate-on-change

**Approach:**
- Each file is generated by `BuildPromptShim(<skill>)` to guarantee byte-for-byte parity with what the installer ships. Decide between `go:generate` (commit results) or a `make prompts` target (run before commit) based on which keeps parity test trivially green.
- Add a one-line note to `CONTRIBUTING.md` (or the closest equivalent) explaining how to regenerate when the allow-list changes.

**Patterns to follow:**
- The dogfood mirror under `.github/skills/` and `templates/skills/` validated by `TestDogfoodTemplateParity` in `pkg/scaffold/parity_test.go`.

**Test scenarios:**
- Test expectation: covered by `TestDogfoodPromptParity` from Unit 3 — no additional tests needed.

**Verification:**
- `ls .github/prompts/` shows every allow-listed skill exactly once; parity test green.
- Open this repo in VS Code Copilot Chat; type `/ce-` and see completions for every shimmed command.

- [x] **Unit 5: Update install summary, recommendations, and primitive docs**

**Goal:** Tell users the prompts shipped, stop telling them to create their own, and document prompts as a primitive.

**Requirements:** R5, R6.

**Dependencies:** Unit 2 (catalog must actually emit the files for the summary count to be honest).

**Files:**
- Modify: `pkg/installstate/recommendations.go` (delete or rescope the `add-prompts` recommendation at lines ~244-252)
- Modify: `pkg/scaffold/scaffold.go` (or wherever the install summary is printed) — add a "Created N prompt-file shims under `.github/prompts/`" line
- Modify: `README.md`, `npm/README.md`, `docs/brainstorms/2026-03-11-agentic-coding-starter-kit-installer-brainstorm.md` — add a 7th primitive row for prompt files with a one-sentence note about VS Code discovery semantics
- Modify: `pkg/scaffold/templates/instructions/{rails,python,typescript,general}.md` — optional one-line note that slash commands are wired via `.github/prompts/`
- Test: `pkg/installstate/recommendations_test.go` (existing tests likely cover `add-prompts`; update to assert it no longer fires when `.github/prompts/` matches the shipped allow-list)

**Approach:**
- The `add-prompts` recommendation today fires whenever `PromptFileCount == 0`. After this change, a fresh install always yields `PromptFileCount > 0`, so the recommendation becomes dead code. Either delete it (preferred) or change the trigger to "no *user-authored* prompts beyond ATV's shims" (deferred — adds complexity).
- Install summary line is purely cosmetic but closes the loop the issue calls out: "no error, no warning ... users either rediscover ... or conclude the install is broken."

**Patterns to follow:**
- `Recommendation` struct usage in `pkg/installstate/recommendations.go:246-251`.
- Existing summary formatting in `pkg/scaffold/scaffold.go` (whichever function prints "Created N components").

**Test scenarios:**
- Happy path: after a fresh install, `BuildRecommendations(state)` does not include an `add-prompts` entry.
- Edge case: after a user manually deletes their prompts directory, recommendations behave reasonably (likely: nothing fires, since this isn't an error condition any more — confirm with existing test expectations).
- Happy path: install summary text contains the expected count and path string.
- Documentation diff: the agentic-primitives table in all three docs has 7 rows including "Prompt files".

**Verification:**
- Recommendation test green; install summary regression test (if any) green; manual review of the three docs confirms the new row is present.

## System-Wide Impact

- **Interaction graph:** scaffold pipeline (`pkg/scaffold/catalog.go` → `pkg/scaffold/scaffold.go`) gains a new component category. `pkg/installstate/recommendations.go` and `pkg/monitor/watcher.go` already wire `.github/prompts/` and need no functional changes — but `recommendations.go` does need the dead `add-prompts` rule retired.
- **Error propagation:** failures generating shims must surface like other scaffold errors (e.g., panics in `mustRead` for missing templates). Use the same posture in `BuildPromptShim`: if the template fails to execute, panic at scaffold-build time, not silently emit empty files.
- **State lifecycle risks:** none — these are pure file outputs, idempotent, no migrations or persistent state.
- **API surface parity:** the four instruction templates already enumerate the same slash-command list. Whatever lands in the allow-list must match those bullet lists, and a future addition to either side should ripple to the other. The exclusion-list parity test (Unit 3) provides the alarm; the four instruction docs are still hand-edited but small enough to keep in sync. Consider a follow-up that makes those bullet lists themselves generated.
- **Integration coverage:** end-to-end sandbox test in `test/sandbox/sandbox_test.go` should assert the expected prompt files exist after a guided install with `core-skills` and `orchestrators` selected. Unit-level catalog tests don't prove the file-write path actually fires.
- **Unchanged invariants:** `SKILL.md` remains the source of truth for skill behavior. Shims do not duplicate skill content. Existing skills, agents, MCP config, hooks, and docs structure are untouched. gstack-prefixed skills under `.github/skills/gstack-*` are not shimmed — gstack manages its own surface.

## Risks & Dependencies

| Risk | Mitigation |
|---|---|
| Allow-list drifts from the bullet lists in `instructions/*.md` | Unit 3's parity tests catch new skills missing from the allow-list. The four instruction docs are still hand-edited; flag in the implementer's PR description that they were checked. Long-term follow-up: generate the bullet lists from the same allow-list. |
| VS Code chat picker UX surprise — too many slash commands clutters the picker | Allow-list explicitly excludes sub-skills. If 18 commands is still too many, the allow-list is the single knob to dial down. No reshipping of skills required. |
| `description:` text in shim front-matter is what VS Code surfaces; bad copy looks unprofessional | Tune wording during implementation by previewing in VS Code; deferred per Open Questions. Keep a single template so updates are one-place. |
| Shims become out of date if a skill is renamed | Rename in `coreSkillDirectories` triggers `TestPromptShimAllowListSubsetOfSkills` to fail, forcing the rename to propagate. Dogfood test then forces the matching `.github/prompts/<oldname>.prompt.md` to be regenerated. |
| Users who already hand-rolled their own `.github/prompts/<name>.prompt.md` (issue reporter included) get clobbered | The installer's normal collision behavior applies — existing files are preserved unless `--force`. Confirm during implementation that prompt components inherit the same collision policy as skill components. |
| `add-prompts` recommendation removal breaks an external workflow that depends on it | Search the repo (and gstack templates) for references; the recommendation is internal to ATV. Changelog note at release. |

## Documentation / Operational Notes

- Add a short section to `README.md` and `npm/README.md` agentic-primitives tables explaining prompt files as a 7th primitive and their role in VS Code Copilot Chat discovery.
- Update `docs/brainstorms/2026-03-11-agentic-coding-starter-kit-installer-brainstorm.md` similarly so the foundational design doc reflects current reality.
- After merge, recommend the user runs `/ce-compound` to capture a `docs/solutions/` entry: "VS Code Copilot Chat needs `.github/prompts/*.prompt.md` shims; ATV ships them by default" — this is exactly the kind of lesson the institutional-knowledge channel is for.
- No rollout, monitoring, or migration concerns — pure additive change to scaffold output.

## Sources & References

- **Origin issue:** [#36](https://github.com/All-The-Vibes/ATV-StarterKit/issues/36) — slash commands don't appear in VS Code Copilot Chat (filed 2026-04-30 by @stephschofield)
- Related code: `pkg/scaffold/catalog.go:165-209` (skill allow-lists), `pkg/scaffold/parity_test.go:55-...` (parity test patterns), `pkg/installstate/recommendations.go:81,246-251` (existing prompt awareness), `pkg/monitor/watcher.go:124,240,320,330` (existing prompt directory watch)
- Related docs: `pkg/scaffold/templates/instructions/general.md` (canonical slash-command list), `docs/brainstorms/2026-04-28-vscode-source-install-clean-plugin-requirements.md` (recent VS Code install work — does not cover this gap), `docs/brainstorms/2026-03-11-agentic-coding-starter-kit-installer-brainstorm.md` (primitives table)
- Branch: `feat/vscode-prompt-shims` (off `origin/main`, created 2026-04-29)
