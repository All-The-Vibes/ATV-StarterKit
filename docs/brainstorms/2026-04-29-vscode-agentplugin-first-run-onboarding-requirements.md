---
date: 2026-04-29
topic: agentplugin-onboarding-metadata-hygiene
---

# AgentPlugin Onboarding and Skill Metadata Hygiene

## Summary

After ATV installs cleanly as an AgentPlugin, the next action should be obvious and the generated plugin artifacts should not become a maintenance trap. The core v1 should focus on product metadata, command identity, picker hygiene, and reducing skill-content drift; rendered README support is useful when the host shows it, but it should not be the foundation of the experience.

This follows the clean source-install picker requirements in `docs/brainstorms/2026-04-28-vscode-source-install-clean-plugin-requirements.md` and adds the next layer: first-run clarity, metadata coherence, and a practical answer to repeated skill copies.

---

## Problem Frame

The clean source-install picker removes the first layer of noise by giving users one ATV install option. A second layer remains after install: users can see metadata or commands that expose internal bundle names such as `atv-everything`, and wonder whether the next step is a chat command, `atv init`, a Copilot CLI command, or project scaffolding.

Protocol install links and rendered plugin README pages are not reliable enough to be v1 foundations. They can stay as opportunistic polish after validation, but the product should work even when the host only shows plugin metadata and slash-command suggestions.

There is also a separate maintainability problem: ATV skills now exist in multiple places. The current generator treats `pkg/scaffold/templates/skills` as the source for plugin outputs, then projects those skills into `plugins/atv-everything`, pack plugins, and per-skill plugins. The repository also has `.github/skills`, which overlaps with the template set but is not the same inventory. The product goal is to keep the install surfaces clean while making the source-of-truth story explicit enough that maintainers know where to edit and validation catches drift.

---

## Actors

- A1. First-time AgentPlugin user: Installs ATV and needs one obvious next action.
- A2. Returning ATV user: Wants to update, check health, or add ATV to a workspace after install.
- A3. ATV maintainer: Maintains plugin metadata and docs across VS Code source install, Claude-format source install, Copilot CLI marketplace, bundle manifests, and granular plugin manifests.

---

## Key Flows

- F1. Clean picker stays parallel
  - **Trigger:** Plugin metadata or generated marketplaces change.
  - **Actors:** A1
  - **Steps:** The source-install picker continues to show one flagship ATV option, while CLI users can still reach granular packs and single-skill plugins through the CLI catalog.
  - **Outcome:** Onboarding work does not regress the clean picker solved by the previous requirements doc.
  - **Covered by:** R1, R2

- F2. Metadata-based handoff
  - **Trigger:** A user finishes install and sees the plugin details/card metadata, or opens chat command suggestions.
  - **Actors:** A1, A2
  - **Steps:** ATV exposes product-facing name, description, tags, and first command guidance that explain the next step even if a README is not rendered.
  - **Outcome:** The user can start using ATV without relying on host-specific README rendering.
  - **Covered by:** R3, R4, R5

- F3. Productized first chat command
  - **Trigger:** A user opens Copilot Chat after installing ATV.
  - **Actors:** A1, A2
  - **Steps:** The user types or selects an ATV front-door command, chooses the relevant path, and is routed to the right existing skill or setup flow.
  - **Outcome:** The normal user path does not require understanding `atv-everything` as the internal bundle name.
  - **Covered by:** R7, R8

- F4. Metadata coherence validation
  - **Trigger:** ATV plugin manifests, generated plugin metadata, or install docs change.
  - **Actors:** A3
  - **Steps:** Validation checks the product-facing names, source paths, descriptions, and default install surfaces across the relevant metadata files.
  - **Outcome:** Consumer-specific metadata can coexist without drifting into contradictory product names or noisy install surfaces.
  - **Covered by:** R8, R9, R10

- F5. Skill source-of-truth hygiene
  - **Trigger:** A maintainer edits or adds an ATV skill.
  - **Actors:** A3
  - **Steps:** The maintainer edits the canonical source, regeneration updates all required plugin projections, and checks fail if generated copies or compatibility snapshots drift.
  - **Outcome:** ATV can keep self-contained plugin artifacts without asking humans to maintain the same skill body by hand in three or more places.
  - **Covered by:** R11, R12, R13, R14, R15

---

## Requirements

**Parallel picker track**

- R1. The clean one-option source-install picker from the earlier requirements doc must remain intact while adding onboarding and metadata improvements.
- R2. Protocol install links may be documented only after they are validated in the target host; they are optional polish, not a requirement for the v1 onboarding solution.

**Metadata handoff**

- R3. Product-facing plugin metadata should make the next step clear through name, description, tags, and command identity even when the host does not render a README.
- R4. If a host renders plugin README/details content, ATV should provide a useful start-here page, but rendered README support remains nice-to-have.
- R5. Metadata and optional details content should not make `npx atv-starterkit init`, Copilot CLI marketplace syntax, or plugin internals appear to be required before a user can try ATV personally.

**Chat command identity**

- R6. The primary chat command identity should be productized around `atv` or the closest platform-supported equivalent.
- R7. ATV should provide a front-door chat command that routes users between common first actions: use ATV personally, add ATV to this workspace, check install health, and update ATV.

**Metadata coherence**

- R8. Consumer-specific plugin metadata should be generated or validated from one coherent product model so product-facing names, source paths, and descriptions do not drift.
- R9. The validation surface should include root source-install metadata, Claude-format source-install metadata, Copilot CLI marketplace metadata, bundle plugin manifests, and granular plugin manifests.
- R10. Metadata repetition is allowed when each consumer needs a distinct shape, but repeated values must not leak conflicting product names into the primary user journey.

**Skill content source of truth**

- R11. ATV must define which directory is canonical for skill content edits. The current best candidate is `pkg/scaffold/templates/skills` because `pkg/plugingen` already projects it into plugin outputs.
- R12. Generated plugin copies under `plugins/` must be treated as artifacts, not hand-edited source. Validation must fail when they differ from the canonical templates after normalization.
- R13. The role of `.github/skills` must be clarified: either generate the overlapping ATV skills from the same canonical source, explicitly mark it as a separate compatibility/demo inventory, or remove stale overlap from the product path.
- R14. The solution should avoid symlinks, hardlinks, or local-only shortcuts that are fragile on Windows, GitHub source install, or packaged plugin consumers.
- R15. Planning should research whether plugin manifests can safely reference canonical repo paths instead of copied skill folders. If supported, reduce copies; if not supported, keep generated self-contained plugin directories and make generation/drift checks the official answer.

---

## Acceptance Examples

- AE1. **Covers R1, R2.** Given onboarding metadata changes, when validation runs, then the clean source-install picker still exposes one flagship ATV option and protocol-link docs are absent unless the link has been validated.
- AE2. **Covers R3, R4, R5.** Given a host does not render plugin README content, when a user sees ATV metadata and command suggestions, then the product name, description, and first command still explain what to do next.
- AE3. **Covers R5.** Given a user reads any ATV install handoff, when they choose personal use, then the handoff does not imply that project scaffolding or Copilot CLI setup is required first.
- AE4. **Covers R6.** Given a user opens chat after installing ATV, when they begin typing an ATV command, then the primary suggestion reads like an ATV product action rather than exposing `atv-everything` as the main mental model.
- AE5. **Covers R7.** Given a user invokes the front-door command, when they choose health check or update, then the flow routes to the existing ATV maintenance path instead of inventing a separate update surface.
- AE6. **Covers R8, R9, R10.** Given generated metadata changes, when validation runs, then it catches drift where consumer-specific catalogs disagree about the flagship plugin name, source path, description, or primary install surface.
- AE7. **Covers R11, R12.** Given a maintainer edits a canonical skill template, when plugin generation runs, then the corresponding full-bundle, pack, and per-skill plugin copies update deterministically.
- AE8. **Covers R13.** Given a skill exists in both `.github/skills` and `pkg/scaffold/templates/skills`, when validation runs, then the project either proves the overlap is intentionally generated/equivalent or reports the drift as a maintainability issue.
- AE9. **Covers R14, R15.** Given planning evaluates copy reduction, when a proposed approach depends on symlinks, hardlinks, or unsupported manifest references, then it is rejected or deferred until proven compatible with the target plugin consumers.

---

## Success Criteria

- A user who successfully installs ATV can tell what to do next without reading repository internals, learning Copilot CLI plugin syntax, or guessing why commands are prefixed with `atv-everything`.
- ATV metadata and command identity are strong enough that rendered README support is a bonus, not a dependency.
- The primary chat command feels like ATV, not an internal bundle artifact.
- Maintainers can update generated plugin metadata without reintroducing contradictory product names or noisy VS Code install choices.
- Maintainers know exactly where to edit skill content, and generated copies cannot drift silently.

---

## Scope Boundaries

- Do not make protocol install links or rendered README behavior required for v1 success.
- Do not build a native button-based VS Code wizard in this AgentPlugin-only follow-up; real buttons inside editor chrome remain a separate extension or VSIX path.
- Do not suppress, bypass, or hide host trust prompts for source-installed plugins.
- Do not remove consumer-specific metadata files solely to reduce apparent repetition if doing so would reintroduce noisy install pickers or break Copilot CLI granularity.
- Do not rename the Copilot CLI full bundle from `atv-everything` unless planning confirms that a shared rename is safe across CLI and VS Code consumers.
- Do not redesign `npx atv-starterkit init`; only clarify when a VS Code user needs it.
- Do not change the clean one-option source-install picker requirements from the earlier brainstorm.
- Do not ask maintainers to manually keep generated skill copies in sync.

---

## Key Decisions

- Picker and onboarding are parallel tracks: the clean picker remains a requirement, but this brainstorm should not depend on protocol links or README rendering.
- Metadata plus front-door command as v1 onboarding: Without a native extension, the practical handoff is productized metadata plus a productized first chat command.
- Metadata repetition is a generation problem, not inherently a product problem: multiple manifests can coexist if each has a clear consumer and validation prevents drift.
- Skill repetition should be solved through canonical source plus generated projections, unless planning proves plugin consumers can safely reference canonical paths directly.

---

## Approach Direction

- Recommended v1: Treat `pkg/scaffold/templates/skills` as canonical for starter-kit skill content, keep `plugins/` as generated self-contained artifacts, and strengthen validation so humans never edit copied skill bodies by hand.
- In parallel: Clarify `.github/skills`. It currently has 67 skills, while `pkg/scaffold/templates/skills` has 30; 24 overlap, 43 exist only in `.github/skills`, and 6 exist only in templates. That means `.github/skills` is not just another generated copy of the starter kit and needs an explicit product role.
- Research-only reduction path: Investigate whether plugin manifests can point at canonical skill directories outside each plugin folder. If that works across source install and Copilot CLI, it could reduce copies. If it does not, generated duplication is acceptable as long as the generator owns it.

---

## Dependencies / Assumptions

- Protocol install links and README rendering have shown inconsistent local behavior, so they should stay optional until validated in the actual target host.
- Current generated metadata has multiple consumer surfaces: root `marketplace.json`, `.claude-plugin/marketplace.json`, `.github/plugin/marketplace.json`, `plugins/atv-everything/plugin.json`, `plugins/atv-everything/.claude-plugin/plugin.json`, and one manifest per granular plugin. This is currently intentional but creates naming and drift risk.
- Current skill content surfaces include `.github/skills` (67 skills), `pkg/scaffold/templates/skills` (30 skills), `plugins/atv-everything/skills` (30 generated skills), pack plugin skill folders, and per-skill plugin folders. The generated plugin copies are expected; the `.github/skills` overlap needs a documented role.
- `pkg/plugingen` already loads skill bodies from `pkg/scaffold/templates/skills`, writes full-bundle, pack, and per-skill plugin copies, and checks generated output through `go run ./cmd/plugingen -check`.

---

## Outstanding Questions

### Resolve Before Planning

- [Affects R6, R7][Technical] Confirm whether the visible slash-command prefix can be productized to `atv` for the source-installed bundle without breaking Copilot CLI marketplace naming or installed-plugin update identity.
- [Affects R11-R15][Technical] Confirm whether plugin manifests may safely reference canonical skill directories outside each plugin folder across source install and Copilot CLI.
- [Affects R13][Product/Technical] Decide the role of `.github/skills`: generated compatibility output, separate historical inventory, or removable/stale surface.

### Deferred to Planning

- [Affects R7][Product] Decide the exact front-door command wording and menu options after confirming platform naming constraints.
- [Affects R8-R15][Technical] Decide whether metadata and skill-copy validation belongs in `cmd/plugingen`, CI, or focused smoke tests.
- [Affects R2, R4][Nice-to-have] Revisit protocol links and rendered README behavior only after the core metadata/command path works.

---

## Next Steps

-> /ce-plan for structured implementation planning