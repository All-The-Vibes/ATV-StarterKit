# Skill Source of Truth

ATV has one canonical local source for installable product skills:

```text
pkg/scaffold/templates/skills
```

That tree feeds both `atv init` scaffold output and the generated plugin marketplace artifacts. Generated plugin copies under `plugins/` are artifacts and should not be edited by hand. The repository dogfood surface under `.github/skills` is explicit compatibility/dogfood inventory, not the installable product source of truth.

The source inventory lives in `pkg/scaffold/skill_sources.json`:

- `origin: compound-engineering` means the skill has a Compound Engineering upstream base.
- `tracking: upstream` means ATV intends to track the same upstream skill name.
- `tracking: alias` means ATV uses a local name for an upstream CE skill, such as `ce-review` tracking upstream `ce-code-review`.
- `tracking: overlay` means ATV intentionally differs from upstream, such as `lfg`.
- `dogfood: mirror` means `.github/skills/<name>/SKILL.md` must match the canonical template body.
- `dogfood: divergent` means the same-name dogfood copy intentionally differs and must remain explicit.
- `dogfood: omitted` means the product skill is not dogfooded in this repo.

## Compare Against Compound Engineering

Run the read-only comparison before refreshing CE-derived skills:

```bash
go run ./cmd/skillsync -upstream ../compound-engineering-plugin/plugins/compound-engineering/skills
```

The command does not modify files. It reports:

- `in_sync`: local ATV template matches the tracked upstream CE skill body.
- `stale`: local ATV template differs from the tracked upstream CE skill body and needs manual review.
- `atv_overlay`: ATV intentionally differs from CE and should not be overwritten mechanically.
- `missing_local`: the manifest tracks an upstream CE skill that is absent locally.
- `missing_upstream`: the manifest points to an upstream skill that no longer exists.
- `untracked_upstream`: CE has a skill that ATV does not currently track.

## 2026-04-29 CE Baseline Report

Using the local reference checkout at `../compound-engineering-plugin/plugins/compound-engineering/skills`, the tracked CE-derived ATV entries currently report:

```text
ce-brainstorm        ce-brainstorm        stale              upstream
ce-compound          ce-compound          stale              upstream
ce-compound-refresh  ce-compound-refresh  stale              upstream
ce-ideate            ce-ideate            stale              upstream
ce-plan              ce-plan              stale              upstream
ce-review            ce-code-review       stale              alias
ce-work              ce-work              stale              upstream
document-review      ce-doc-review        stale              alias
lfg                  lfg                  atv_overlay        overlay
setup                ce-setup             stale              alias
test-browser         ce-test-browser      stale              alias

stale=10
atv_overlay=1
untracked_upstream=24
```

The untracked upstream CE skills are:

```text
ce-agent-native-architecture
ce-agent-native-audit
ce-clean-gone-branches
ce-commit
ce-commit-push-pr
ce-debug
ce-demo-reel
ce-dhh-rails-style
ce-frontend-design
ce-gemini-imagegen
ce-optimize
ce-polish-beta
ce-proof
ce-release-notes
ce-report-bug
ce-resolve-pr-feedback
ce-session-extract
ce-session-inventory
ce-sessions
ce-slack-research
ce-test-xcode
ce-update
ce-work-beta
ce-worktree
```

This report means CE is the upstream base for those tracked skills, but refreshes must be reviewed skill by skill. Do not bulk-copy CE over ATV: the point is to understand what changed upstream, decide whether ATV needs an overlay, then update the canonical template and manifest deliberately.

Notable current interpretation:

- `lfg` is an ATV overlay. It has a CE upstream skill with the same name, but ATV's local pipeline includes ATV-specific stages and should not be mechanically replaced.
- `ce-review`, `document-review`, `setup`, and `test-browser` are local aliases for CE skills with different upstream names.
- CE currently has 24 skills ATV does not track. They are candidates for future curation, not automatic imports.

## Refresh Workflow

1. Run `go run ./cmd/skillsync -upstream <ce-skills-dir>`.
2. For each `stale` skill, inspect the upstream diff and classify the difference as safe refresh, ATV overlay, or deferred.
3. Edit only `pkg/scaffold/templates/skills/<name>/` for installable product content.
4. Update `pkg/scaffold/skill_sources.json` if the origin, alias, overlay, or dogfood policy changes.
5. Run `go run ./cmd/plugingen` to regenerate `plugins/` and marketplace files.
6. Run `go test ./pkg/scaffold ./pkg/skillsync ./pkg/plugingen` and `go run ./cmd/plugingen -check`.
