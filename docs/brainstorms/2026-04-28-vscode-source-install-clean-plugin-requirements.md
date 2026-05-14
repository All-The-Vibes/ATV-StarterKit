---
date: 2026-04-28
topic: vscode-source-install-clean-plugin
---

# VS Code Source Install Clean Plugin

## Summary

ATV should present one clear default install choice on each personal-plugin surface: `atv-starter-kit` for VS Code source install and `atv-everything` for Copilot CLI. Granular CLI packs, single-skill plugins, and agents-only installs remain available as advanced options, but they must not leak into the default VS Code picker.

---

## Problem Frame

VS Code Insiders can now install agent plugins directly from a GitHub repository through `Chat: Install Plugin from source`. When users install `All-The-Vibes/ATV-StarterKit`, the selector shows the full ATV Copilot CLI marketplace catalog: category packs, agents-only bundles, and many individual skill plugins. This makes ATV feel noisy and harder to trust at the exact moment a user is deciding what to install.

The Copilot CLI catalog already has a general full-bundle plugin, `atv-everything`. The product problem is not that granular CLI plugins exist; it is that the default user journey is not clearly hierarchical. VS Code should show one productized choice backed by the complete bundle, while CLI users should be guided first to `atv-everything` and only then to advanced pack or single-skill installs.

The same lifecycle should also be easy after install: users should be able to see which source-installed AgentPlugin version is active, update it, and reload VS Code without manually deleting plugin folders.

---

## Actors

- A1. VS Code Copilot user: Installs ATV from a GitHub repository inside VS Code and expects one clear decision.
- A2. ATV maintainer: Publishes and validates plugin metadata without breaking the existing starter kit experience.
- A3. Copilot CLI user: Uses `atv-everything` as the recommended full install, with packs, single skills, and agents-only installs available as advanced granular options.
- A4. Source-installed plugin user: Has ATV, CE, or another AgentPlugin installed under VS Code's agent plugin directory and needs a reliable update path.

---

## Key Flows

- F1. Clean VS Code source install
  - **Trigger:** A VS Code user runs `Chat: Install Plugin from source` and enters `All-The-Vibes/ATV-StarterKit` or a fork such as `datorresb/ATV-StarterKit`.
  - **Actors:** A1
  - **Steps:** VS Code resolves the repository, reads the plugin metadata, shows a single ATV install option, the user selects it, and ATV skills/agents become available.
  - **Outcome:** The user sees ATV as one coherent product rather than 40+ install fragments.
  - **Covered by:** R1, R2, R3, R4

- F2. Maintainer validation
  - **Trigger:** ATV plugin metadata, templates, or generated plugin folders change.
  - **Actors:** A2
  - **Steps:** The maintainer regenerates or updates plugin metadata, validates the generated output, confirms the source-install catalog remains one entry, confirms the CLI catalog still has a clear flagship default, and performs a VS Code source-install smoke test.
  - **Outcome:** The clean selector remains stable across releases and CLI users still have a coherent recommended default.
  - **Covered by:** R5, R6, R7, R10

- F3. Source-installed plugin update
  - **Trigger:** A user suspects their installed AgentPlugin is stale after a VS Code uninstall, reload, reinstall, or upstream release.
  - **Actors:** A1, A4
  - **Steps:** The user runs an ATV health/update workflow, sees each installed source plugin's current version and git state, confirms any update, and receives clear reload guidance after the update.
  - **Outcome:** The user can update source-installed plugins without manually finding, deleting, or recloning plugin folders.
  - **Covered by:** R11, R12, R13, R14, R15

---

## Requirements

**User-facing install surface**

- R1. VS Code source install must show one primary ATV option, named clearly enough that a first-time user can select it without reading ATV's marketplace docs.
- R2. The primary option description must be short, user-facing, and fit comfortably in VS Code's Quick Pick row without relying on long warnings, URLs, or implementation notes.
- R3. The option should represent the complete recommended ATV personal plugin experience for VS Code, not a partial skill, hidden dependency, or expert-only bundle.
- R4. Individual skill plugins and category packs must not appear in the VS Code source-install picker for the default ATV repository install flow.

**Compatibility and positioning**

- R5. The project-level `npx atv-starterkit init` flow remains separate and continues to be documented as the team/shared repo bootstrap path.
- R6. Copilot CLI marketplace support may remain granular, but it must position `atv-everything` as the recommended full install and must not force the VS Code source-install picker to expose every granular ATV package.
- R7. If VS Code and Copilot CLI cannot use separate catalogs cleanly, prioritize the one-option VS Code source-install experience for this change and defer granular CLI preservation to a follow-up.

**Documentation and validation**

- R8. ATV documentation must include the VS Code source-install flow using the current command palette wording: `Chat: Install Plugin from source`.
- R9. Documentation must explain the difference between VS Code source install, Copilot CLI plugin install, and `npx atv-starterkit init` without making users choose from a large matrix before the quick start; Copilot CLI docs must lead with `atv-everything` before advanced granular installs.
- R10. A release or validation check must catch drift where the VS Code source-install surface grows beyond the intended single primary option.

**VS Code AgentPlugin lifecycle**

- R11. ATV health/update workflows must detect source-installed AgentPlugins under the VS Code and VS Code Insiders agent plugin roots, including repositories installed from GitHub source.
- R12. The workflow must report a source-installed plugin's repository owner/name, current branch, current commit, version metadata when available, and whether the local checkout differs from its remote tracking branch.
- R13. Updating a source-installed plugin must be opt-in and must avoid overwriting local changes silently; dirty or diverged worktrees require an explicit remediation choice.
- R14. If a clean in-place update is not possible, the workflow may offer a reinstall/removal path only after showing the exact plugin folder that would be removed and receiving explicit confirmation.
- R15. After any successful update, the workflow must tell the user how to make VS Code load the new plugin content, including reload/restart guidance when needed.

---

## Acceptance Examples

- AE1. **Covers R1, R2, R4.** Given the ATV repository is installed through `Chat: Install Plugin from source`, when VS Code shows the plugin picker, then the picker contains one ATV install option with a concise description instead of the current 42-entry list.
- AE2. **Covers R3.** Given a user installs the single ATV option, when they open Copilot Chat, then the recommended ATV skills and agents needed for the normal personal workflow are available without separately installing `atv-agents` or a category pack.
- AE3. **Covers R5, R6, R9.** Given a user reads the README, when they compare install paths, then they can quickly tell that VS Code source install is personal/editor-level, `npx atv-starterkit init` is project/team-level, and Copilot CLI's recommended full install is `atv-everything`.
- AE4. **Covers R10.** Given generated plugin metadata changes, when validation runs, then it fails or warns if the VS Code source-install catalog would expose granular skill or pack entries again, or if regeneration would replace the curated root source-install catalog with the granular CLI catalog.
- AE5. **Covers R11, R12.** Given CE is installed under VS Code Insiders AgentPlugins, when the health workflow runs, then it reports the installed path, current package version, current commit, git tag/description when available, and whether it is behind `origin/main`.
- AE6. **Covers R13, R14, R15.** Given a source-installed plugin is stale and its worktree is clean, when the user confirms an update, then ATV updates it safely and tells the user to reload VS Code; if the worktree is dirty or update fails, ATV explains the exact state and does not delete or overwrite anything without explicit confirmation.

---

## Success Criteria

- A first-time VS Code user sees ATV as one polished product, not a long list of internal packages.
- The install picker for ATV visually resembles the clean `EveryInc/compound-engineering-plugin` experience: one obvious flagship option rather than many granular options.
- A Copilot CLI user can identify `atv-everything` as the default full install before encountering pack-level or single-skill advanced choices.
- Source-installed plugin updates no longer require users to manually inspect `.vscode-insiders/agent-plugins`, delete folders, and reinstall by hand.
- Planning can proceed without inventing the product shape: the chosen v1 UX is a single source-install entry.
- Maintainers have a repeatable way to verify that ATV's source-install surface has not regressed.

---

## Scope Boundaries

- Do not build a full VS Code extension, VSIX package, chat participant, or Visual Studio Marketplace listing in this v1.
- Do not change the behavior of ATV skills or agents as part of this cleanup unless required for packaging correctness.
- Do not remove or redesign the `npx atv-starterkit init` project bootstrap flow.
- Do not remove Copilot CLI pack, single-skill, or agents-only plugins in this v1; preserve them as advanced/granular options.
- Do not rename the Copilot CLI full bundle from `atv-everything` as part of this change.
- Do not expose single-skill plugins, category packs, or agent-only bundles in the VS Code source-install picker.
- Do not require users to understand Copilot CLI marketplace mechanics before installing ATV in VS Code.
- Do not silently update, reset, delete, or reinstall source-installed plugin folders.
- Do not attempt to solve generic VS Code extension auto-update or Visual Studio Marketplace publishing in this v1.

---

## Key Decisions

- One-option source install: The desired VS Code experience is a single clean ATV option.
- CLI flagship plus advanced granularity: `atv-everything` is the Copilot CLI default full install; granular packs, single skills, and agents-only installs remain available for advanced users.
- VS Code UX over catalog granularity: Granular packs are useful for CLI/power users, but they should not appear in the default VS Code source-install picker.
- Documentation should follow the current VS Code flow: The relevant user path is `Chat: Install Plugin from source`, not manual copying into `.github/skills/`.
- Update belongs in the ATV maintenance workflow: Source-installed AgentPlugin health/update should extend `/atv-doctor` and `/atv-update` rather than becoming another visible install option.

---

## Dependencies / Assumptions

- ATV's Copilot CLI metadata exposes 42 plugin entries from `.github/plugin/marketplace.json`. That catalog includes `atv-everything` as the full-bundle default plus advanced pack, single-skill, and agents-only options.
- `EveryInc/compound-engineering-plugin` exposes only `compound-engineering` and `coding-tutor` in its plugin marketplace metadata, which matches the clean picker shown in VS Code.
- VS Code's source-install resolver consumes root `marketplace.json` before `.github/plugin/marketplace.json`; generated metadata must preserve that split so regeneration does not replace the curated root catalog with the 42-entry CLI catalog.
- GitHub issue `EveryInc/compound-engineering-plugin#637` confirms the current recommended source-install path for VS Code users.
- On Windows with VS Code Insiders, source-installed AgentPlugins are present under `%USERPROFILE%/.vscode-insiders/agent-plugins/github.com/<owner>/<repo>` and are git worktrees.
- Locally observed examples include `EveryInc/compound-engineering-plugin` with version metadata in `package.json` and `All-The-Vibes/ATV-StarterKit` with version metadata in `VERSION`.

---

## Outstanding Questions

### Resolve Before Planning

- None.

### Deferred to Planning

- [Affects R6, R9][Technical] Confirm whether Copilot CLI marketplace browsing honors catalog order; if it does, keep `atv-everything` first, and if it does not, make documentation and install commands lead with `atv-everything`.
- [Affects R10][Technical] Decide whether validation belongs in `cmd/plugingen`, CI, or a focused metadata smoke test.
- [Affects R11-R15][Needs research] Confirm whether VS Code provides an official AgentPlugin update/reload command, or whether ATV should treat source-installed plugins as git checkouts and guide users to reload VS Code after update.
- [Affects R11, R12][Technical] Decide whether ATV should update only `All-The-Vibes/ATV-StarterKit` by default or also offer generic diagnostics for other source-installed plugins such as `EveryInc/compound-engineering-plugin`.

---

## Next Steps

-> /ce-plan for structured implementation planning