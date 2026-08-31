---
title: "feat: integrate solution debranding into the ATV lifecycle"
type: feat
status: implemented
date: 2026-08-28
source_repository: https://github.com/lshade/solution-debranding
source_commit: 0a1f069020a949cdd4d647b82ea58a15a48ca330
intent_manifest: .phoenix-intent/intent.json
---

# Integrate solution debranding into the ATV lifecycle

## Intent

Incorporate `lshade/solution-debranding` as a faithfully vendored, maintainable
four-skill family and have ATV propose it to users beside `/unslop` at the
post-review quality and release-readiness stage of the enforced lifecycle.

The integration must preserve the upstream command names, arguments, safety
constraints, plan/apply/verify handoff, scripts, references, tests, and generated
artifact contract. ATV-specific routing and packaging logic must live outside
the vendored package.

## Current-state findings

- Upstream HEAD is pinned for this plan at
  `0a1f069020a949cdd4d647b82ea58a15a48ca330`.
- Upstream is an MIT-licensed family of four sibling skill directories:
  `solution-debranding`, `solution-debranding-plan`,
  `solution-debranding-apply`, and `solution-debranding-verify`.
- The shared `solution-debranding` directory contains the operating contract,
  seven references, four Python scripts, script tests, and evals. Copying only
  `SKILL.md` would not preserve usage.
- The three stage skills resolve the shared package by relative path. They
  cannot be flattened or shipped independently from the shared sibling.
- ATV project scaffolding already embeds nested skill files recursively.
- ATV plugin generation currently reads and emits only each skill's `SKILL.md`.
  It must learn to copy complete skill package trees before this family can work
  from marketplace plugins.
- ATV has no separate component named "auto router." The enforced routing
  surfaces are the `/lfg` and `/slfg` orchestrator skills, their duplicated
  installer/plugin copies, the run-state helper, prompt shims, and lifecycle
  documentation.
- The checked-in orchestrator copies have drift. Some LFG copies include
  `/unslop`; the dogfood `.github/skills/lfg` and `.github/skills/slfg` copies
  do not. This integration must establish one tested lifecycle contract rather
  than add another hand-maintained divergence.

## Product decision

Debranding is proposed, not silently executed.

At the post-review gate, Copilot must present two adjacent quality and
release-readiness actions:

1. `/unslop fix` for code and design cleanup.
2. `/solution-debranding-plan` for portability, white-labeling, handoff, reuse,
   or public-release preparation.

`/unslop fix` keeps its existing workflow semantics. Debranding starts with
plan mode only after the user opts in and supplies or confirms the source brand
and scope. ATV must never jump directly to apply or verify, because upstream
requires an approved plan and explicit authorization for sensitive actions.
Declining debranding records the decision and does not block the rest of the
pipeline.

For `/slfg`, the proposal is emitted after review and concurrent checks join.
Debranding is not placed in the parallel mutation phase. If accepted, its own
plan, apply, and verify sequence runs in order.

## Goal 1: vendor and maintain the upstream package

### Implementation

- Add all files from the four upstream skill directories under
  `pkg/scaffold/templates/skills/` without editing their contents.
- Mirror the same four directories under `.github/skills/` for dogfooding.
- Preserve relative sibling layout, executable script behavior, and LF line
  endings.
- Add a vendor lock containing the upstream repository, pinned commit, complete
  file inventory, and SHA-256 for every vendored file.
- Preserve Lisa Shade's upstream MIT license and copyright notice in a
  third-party notice associated with the vendored package.
- Add a deterministic sync command that accepts an explicit upstream commit,
  stages the full family, refreshes hashes, and refuses a dirty or partial
  result.
- Add a scheduled upstream-drift workflow that compares the pinned commit with
  upstream HEAD and reports an update is available. It must not silently modify
  vendored code.
- Keep upstream content changes separate from ATV adapters so future updates can
  be reviewed as a clean upstream diff.

### Acceptance

`node .github/hooks/scripts/verify-solution-debranding.js vendor` must:

- compare both vendored roots against the lock file;
- fail on a missing, extra, renamed, or content-modified upstream file;
- verify all four sibling directories and all nested assets are present;
- verify the pinned repository, commit, and license notice; and
- run upstream's Python script tests when Python 3.11 or later is available.

## Goal 2: distribute the family through every ATV surface

### Implementation

- Change plugin generation to copy complete skill directory trees instead of
  only `SKILL.md`, while continuing to audit each entry point.
- Introduce a skill-family packaging rule for the four debranding directories.
  Installing any debranding stage through a granular plugin must include all
  four siblings so relative references always resolve.
- Register the four directories in the scaffold catalog as core quality skills.
- Add all four to `atv-pack-quality`, `atv-everything`, the source-install
  bundle, guided/full installs, and generated marketplace metadata.
- Generate prompt shims for the three user-invocable stage skills. Do not expose
  the shared `solution-debranding` package as a user command.
- Update plugin descriptions and counts from generated data rather than
  hand-editing generated output.
- Add parity tests for project scaffold, quality pack, everything bundle,
  source install, granular family plugins, prompt shims, and marketplace
  manifests.

### Acceptance

`node .github/hooks/scripts/verify-solution-debranding.js distribution` must
build or inspect every generated surface and fail unless:

- all four sibling trees are complete and byte-identical to the vendored source;
- every supported installation path yields a runnable family;
- no installation path emits a stage skill without its shared package;
- the three stage commands are discoverable; and
- generated output is clean according to the existing plugin parity check.

## Goal 3: add the post-review proposal to lifecycle routing

### Implementation

- Define a named `quality-release-readiness` phase in the resumable LFG state
  contract immediately after review and before final resolution, browser/video,
  compound, or handoff work.
- In that phase, present `/unslop fix` and `/solution-debranding-plan` together.
- Make the debranding prompt explain when it fits: reuse across brands,
  white-labeling, ownership transfer, demo preparation, or public release.
- Record `accepted`, `declined`, or `not-applicable` plus any debranding plan
  artifact in `.atv/runs/<run-id>/` so resume does not prompt twice.
- When accepted, invoke plan first and pass its artifact through apply and
  verify according to the unchanged upstream contract.
- Never treat an unapproved apply step, a failed verify result, or a
  human-gated legal/security decision as successful pipeline completion.
- Update canonical LFG and SLFG templates, dogfood copies, generated plugin
  copies, README lifecycle diagrams, `DOCS.md`, and marketplace documentation.
- Add parity tests that fail when orchestrator copies disagree on phase order or
  proposal wording.

### Acceptance

`node .github/hooks/scripts/verify-solution-debranding.js lifecycle` must model
fresh and resumed LFG and SLFG runs and fail unless:

- the proposal occurs after review at the same named stage as unslop;
- debranding never starts with apply or verify;
- decline and not-applicable decisions continue without re-prompting;
- acceptance preserves plan, then apply, then verify ordering;
- unresolved human approval or failed verification blocks a false success; and
- all checked-in orchestrator copies expose the same lifecycle contract.

## Dependency order

| Goal | Depends on | Reason |
|---|---|---|
| Vendor upstream package | None | Establishes the immutable source and maintenance contract. |
| Distribute skill family | Vendor upstream package | Packaging tests need the complete source tree and lock. |
| Route release readiness | Distribute skill family | The router must not propose commands that an install does not provide. |

## Failure-first baseline

All three isolated acceptance checks were run on 2026-08-28 and returned red
because `.github/hooks/scripts/verify-solution-debranding.js` does not exist.
This is the intended pre-implementation baseline, not a missing planning step.

| Goal | Baseline | Trace workspace |
|---|---|---|
| Vendor upstream package | RED | `.phoenix-intent/vendor-upstream-package/.phoenix/trace.jsonl` |
| Distribute skill family | RED | `.phoenix-intent/distribute-skill-family/.phoenix/trace.jsonl` |
| Route release readiness | RED | `.phoenix-intent/route-release-readiness/.phoenix/trace.jsonl` |

Implementation is complete only after each same check turns green in its own
trace and composite intent acceptance returns `ok=true`.

## Expected file surfaces

- `.github/skills/solution-debranding*/**`
- `pkg/scaffold/templates/skills/solution-debranding*/**`
- `.github/prompts/solution-debranding-*.prompt.md`
- `.github/hooks/scripts/lfg-state.js`
- `.github/hooks/scripts/verify-solution-debranding.js`
- lifecycle copies of `lfg/SKILL.md` and `slfg/SKILL.md`
- `pkg/scaffold/catalog.go`
- `pkg/scaffold/*solution_debranding*_test.go`
- `pkg/plugingen/*.go`
- generated `plugins/**`, `marketplace.json`,
  `.github/plugin/marketplace.json`, and `.claude-plugin/marketplace.json`
- third-party provenance and license notice
- `README.md`, `DOCS.md`, and `docs/marketplace.md`

## Non-goals

- Changing upstream debranding semantics, command names, arguments, scans,
  approval rules, or artifact format.
- Automatically debranding every project.
- Combining `/unslop` and solution debranding into one skill.
- Removing legitimate attribution or license notices.
- Rewriting Git history or touching external systems without explicit user
  authorization.

## Verification sequence

1. Run the three intent checks and require each trace to show red, then green.
2. Run upstream Python tests from both vendored roots.
3. Run targeted scaffold, plugin generation, lifecycle state, prompt, and parity
   tests.
4. Run `go test ./...`.
5. Run the plugin clean-generation check.
6. Install into a temporary project via guided/full scaffold and each relevant
   plugin route, then invoke plan discovery to confirm relative references.
7. Run composite acceptance against `.phoenix-intent/intent.json`.

## Completion

Implemented on 2026-08-28. All three isolated Phoenix goals satisfy their
failure-first gates, the full Go suite passes, generated plugin output is clean,
and composite intent acceptance reports `goals_ok: 3`.
