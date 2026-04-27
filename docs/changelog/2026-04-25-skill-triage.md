# Skill Triage: 2026-04-25

Decision log for reconciling the drift between `.github/skills/` (dogfood)
and `pkg/scaffold/templates/skills/` (installer payload). See plan
`docs/plans/2026-04-25-003-refactor-centralize-skills-and-reconcile-drift-plan.md`.

Three verdicts:
- **MIRROR** — copy SKILL.md into `pkg/scaffold/templates/skills/<name>/`,
  register in a catalog layer, remove from `pendingMirror`.
- **DELETE-DOGFOOD** — remove `.github/skills/<name>/`. Skill is
  deprecated, an internal CE alias, or duplicated by another skill we
  already ship.
- **KEEP-DOGFOOD-ONLY** — leave `.github/skills/<name>/` in place,
  move from `pendingMirror` to a new `dogfoodOnly` allow-list with
  rationale. Skill is intentionally repo-local (beta, internal,
  CE-internal review tooling, etc.).

| Skill | Verdict | Layer (if MIRROR) | Rationale |
|---|---|---|---|
| agent-browser | DELETE-DOGFOOD | — | Installed separately by `cmd/init.go` via Vercel upstream; markdown copy is stale documentation duplication. |
| agent-native-architecture | MIRROR | `core-skills` | Pairs with our agentic positioning; upstream CE skill that strengthens the core pipeline. |
| agent-native-audit | KEEP-DOGFOOD-ONLY | — | Internal CE audit tool; not user-facing. |
| andrew-kane-gem-writer | MIRROR | `style-skills` | Specialized Ruby gem-writing style; useful when users opt in. |
| ce-work-beta | KEEP-DOGFOOD-ONLY | — | Beta variant of `ce-work`. Will be deleted once external-delegate mode lands in `ce-work` proper. |
| changelog | MIRROR | `dev-tools` | Generates engaging changelogs from merged PRs; broadly useful. |
| compound-docs | KEEP-DOGFOOD-ONLY | — | Internal CE documentation skill superseded for users by `ce-compound` / `ce-compound-refresh`; kept for repo-internal use. |
| create-agent-skill | DELETE-DOGFOOD | — | Duplicated by `skill-creator`; older one-off CE alias. |
| create-agent-skills | DELETE-DOGFOOD | — | Duplicated by `skill-creator`; older one-off CE alias. |
| deploy-docs | KEEP-DOGFOOD-ONLY | — | Internal CE deployment tool for plugin docs site; not relevant to ATV users. |
| dhh-rails-style | MIRROR | `style-skills` | Strong opinionated Rails style guide; matches the "style-skills" theme. |
| dspy-ruby | MIRROR | `style-skills` | DSPy.rb framework guide; useful Ruby AI library. |
| every-style-editor | MIRROR | `style-skills` | Every.to copy-editing style guide. |
| file-todos | KEEP-DOGFOOD-ONLY | — | Repo-internal todo tracking; superseded for users by the CLI todo skills. |
| frontend-design | MIRROR | `style-skills` | Distinctive frontend design skill; the headline anti-AI-slop tool. |
| gemini-imagegen | MIRROR | `media-skills` | Gemini Nano Banana Pro image generation. |
| generate_command | DELETE-DOGFOOD | — | Older slash-command generator; superseded by `skill-creator`. |
| ghcp-review-resolve | MIRROR | `dev-tools` | Actively dogfooded (PRs #23/#26/#27/#28); already on the install allow-list per the merged work in `feat/land-ralph-loop-cleanup`. |
| git-clean-gone-branches | MIRROR | `dev-tools` | Branch hygiene utility. |
| git-commit | MIRROR | `dev-tools` | Conventional-commit creator. |
| git-commit-push-pr | MIRROR | `dev-tools` | One-shot commit + push + PR. |
| git-worktree | MIRROR | `dev-tools` | Worktree management for parallel dev. |
| heal-skill | KEEP-DOGFOOD-ONLY | — | CE-internal skill repair tool; meta-skill not for end users. |
| onboarding | MIRROR | `dev-tools` | Generates `ONBOARDING.md`; useful for any repo. |
| orchestrating-swarms | KEEP-DOGFOOD-ONLY | — | CE-internal swarm orchestration; superseded for users by `slfg` / `lfg`. |
| proof | MIRROR | `media-skills` | Proof.editor markdown collaboration. |
| rclone | MIRROR | `media-skills` | Cloud storage uploads. |
| report-bug | DELETE-DOGFOOD | — | "Report a bug in the compound-engineering plugin" — not relevant to ATV. |
| report-bug-ce | KEEP-DOGFOOD-ONLY | — | Same scope as `report-bug`; CE-internal. Keep one for repo-internal use. |
| reproduce-bug | MIRROR | `dev-tools` | Reproduces bugs from GitHub issues; broadly useful. |
| resolve-pr-feedback | KEEP-DOGFOOD-ONLY | — | Superseded for users by `ghcp-review-resolve`; kept for repo-internal use. |
| resolve-pr-parallel | DELETE-DOGFOOD | — | Older parallel-resolve variant; superseded by `ghcp-review-resolve`. |
| resolve_parallel | DELETE-DOGFOOD | — | Older TODO-comment resolver; underscore-naming inconsistent with rest of catalog; superseded. |
| skill-creator | MIRROR | `dev-tools` | Canonical skill creator/editor. |
| test-xcode | KEEP-DOGFOOD-ONLY | — | iOS-specific build/test skill; not relevant to current ATV stack-detection set. Keep dogfood-only; revisit if/when iOS pack ships. |
| todo-create | MIRROR | `dev-tools` | CLI todo system creator. |
| todo-resolve | MIRROR | `dev-tools` | Batch-resolves approved todos. |
| todo-triage | MIRROR | `dev-tools` | Reviews and approves pending todos. |
| triage | KEEP-DOGFOOD-ONLY | — | Lower-level CLI todo triage primitive; user-facing equivalent is `todo-triage`. |
| workflows-brainstorm | DELETE-DOGFOOD | — | Marked `[DEPRECATED] Use /ce:brainstorm instead`. |
| workflows-compound | DELETE-DOGFOOD | — | Marked `[DEPRECATED] Use /ce:compound instead`. |
| workflows-plan | DELETE-DOGFOOD | — | Marked `[DEPRECATED] Use /ce:plan instead`. |
| workflows-review | DELETE-DOGFOOD | — | Marked `[DEPRECATED] Use /ce:review instead`. |
| workflows-work | DELETE-DOGFOOD | — | Marked `[DEPRECATED] Use /ce:work instead`. |

## Counts

- MIRROR: 21
- DELETE-DOGFOOD: 12
- KEEP-DOGFOOD-ONLY: 11
- **Total reconciled:** 44 (matches the original `pendingMirror` size)

## New layer groups in `pkg/scaffold/catalog.go`

- `dev-tools` (12): changelog, ghcp-review-resolve, git-clean-gone-branches,
  git-commit, git-commit-push-pr, git-worktree, onboarding, reproduce-bug,
  skill-creator, todo-create, todo-resolve, todo-triage
- `style-skills` (5): andrew-kane-gem-writer, dhh-rails-style, dspy-ruby,
  every-style-editor, frontend-design
- `media-skills` (3): gemini-imagegen, proof, rclone
- `core-skills` (existing) gains: agent-native-architecture
