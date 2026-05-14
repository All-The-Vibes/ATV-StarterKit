---
name: atv-update
description: "Update ATV Starter Kit to the latest version. Handles Copilot CLI marketplace plugins, VS Code source-installed AgentPlugins, and project scaffold advisory status. Marketplace plugins use `copilot plugin update` with confirmation. Clean source AgentPlugins can fast-forward with confirmation. Project scaffold remains advisory because today's installer is additive-only. Triggers on 'atv update', 'update atv', 'upgrade atv', 'atv upgrade', 'refresh atv', 'atv latest'."
argument-hint: "[mode: dry-run | apply (default)]"
---

# /atv-update — Update ATV Starter Kit

Bring your ATV install up to the latest version. Handles three install paths:

- **Marketplace plugins** — auto-updated via `copilot plugin update` (with per-plugin confirmation).
- **VS Code source-installed AgentPlugins** — updated only when the checkout is clean and can fast-forward safely (with confirmation and reload guidance).
- **Project scaffold** — advisory only. Today's installer is `additive-only` and `npx atv-starterkit@latest init` will NOT refresh existing scaffold files. This skill prints the version delta and the exact commands you can run yourself.

> **Future:** Once the installer gains a `--refresh` flag (read manifest → overwrite checksum-clean files → preserve drifted files), this skill will gain auto-update for project scaffold too. Tracked as a known gap.

## Arguments

<mode> #$ARGUMENTS </mode>

**Mode detection:** Check if arguments contain `dry-run` or `dry` (case-insensitive) → dry-run mode. Otherwise → apply mode (default).

## Execution Flow

```
Phase 1: Detect install scope         → same as /atv-doctor Phase 1
Phase 2: Read installed versions      → project + marketplace + source AgentPlugins
Phase 3: Fetch latest npm version     → npm view
Phase 4: Show changelog (best-effort) → fetch + parse, fall back to link
Phase 5: Plan update                  → structured table of components + commands
Phase 6: Apply marketplace updates    → copilot plugin update with confirmation
Phase 7: Apply source plugin updates  → clean fast-forward only, with confirmation
Phase 8: Verify and reload guidance   → suggest /atv-doctor + VS Code reload/restart
```

---

## Phase 1: Detect install scope

Same detection logic as `/atv-doctor` Phase 1 — repo-artifact-based, not manifest-only:

| Flag | Detection |
|------|-----------|
| `hasProject` | any of `.github/skills/`, `.github/copilot-instructions.md`, `.github/copilot-mcp-config.json` exists |
| `hasManifest` | `.atv/install-manifest.json` exists |
| `hasMarketplace` | `~/.copilot/installed-plugins/atv-starter-kit/` exists |
| `hasSourceAgentPlugins` | any git checkout exists under a VS Code AgentPlugin root such as `$HOME/.vscode/agent-plugins/github.com/<owner>/<repo>` or `$HOME/.vscode-insiders/agent-plugins/github.com/<owner>/<repo>` |

Probe both Stable and Insiders roots. On Windows, prefer profile-aware paths such as `%USERPROFILE%\.vscode-insiders\agent-plugins\github.com` in PowerShell and `$HOME/.vscode-insiders/agent-plugins/github.com` in Bash; never hardcode a machine-specific user directory.

If `!hasProject && !hasMarketplace && !hasSourceAgentPlugins`: print "No ATV install detected. Nothing to update. Run `npx atv-starterkit init` to scaffold, `copilot plugin marketplace add All-The-Vibes/ATV-StarterKit` to register the marketplace, or VS Code `Chat: Install Plugin from source` with `All-The-Vibes/ATV-StarterKit`." and stop.

---

## Phase 2: Read installed versions

### Project (when `hasProject && hasManifest`)

```bash
node -e "console.log(JSON.parse(require('fs').readFileSync('.atv/install-manifest.json','utf8')).catalogVersion || 'unknown')"
```

When `hasProject && !hasManifest`: project version is unknown (auto-mode doesn't write a manifest). Note this and continue — `/atv-update` will still recommend running the installer.

### Marketplace plugins

Walk `~/.copilot/installed-plugins/atv-starter-kit/` directly. For each subdirectory with a `plugin.json`, read `name` and `version`:

```bash
for dir in ~/.copilot/installed-plugins/atv-starter-kit/*/; do
  if [ -f "$dir/plugin.json" ]; then
    node -e "const p = require('$dir/plugin.json'); console.log(p.name + '\t' + (p.version || 'unknown'))"
  fi
done
```

Build a list `{name, currentVersion}` for each ATV plugin.

### VS Code source-installed AgentPlugins

Walk VS Code Stable and Insiders AgentPlugin roots under `agent-plugins/github.com/<owner>/<repo>`. By default, target only `All-The-Vibes/ATV-StarterKit` and `datorresb/ATV-StarterKit` source installs. If the user explicitly named another owner/repo or exact plugin path in the arguments, inspect that target too.

For each target checkout, collect:

- exact path
- owner/repo from the path
- version metadata from `package.json`, `VERSION`, `.claude-plugin/plugin.json`, or `plugin.json`
- current branch or detached HEAD
- current commit and `git describe --tags --always` when available
- origin URL
- dirty state from `git status --short`
- upstream tracking branch and ahead/behind counts after best-effort fetch

Classify source plugin update state:

| State | Update action |
|-------|---------------|
| clean, tracking branch, behind remote only | eligible for fast-forward after confirmation |
| clean and aligned | no update needed |
| dirty worktree | stop; explain local edits must be reviewed first |
| ahead of remote | stop; explain local commits must be reviewed first |
| diverged | stop; explain both local and remote changed |
| detached HEAD or no upstream | stop; manual update path only |
| fetch failed | stop; remote state unknown |

Do not run `git reset`, `git stash`, branch checkout, folder deletion, or reclone as part of normal update.

---

## Phase 3: Fetch latest npm version

```bash
LATEST=$(npm view atv-starterkit version 2>/dev/null)
```

All ATV plugins share the kit's release cadence — the latest plugin version equals the latest npm version. If the call fails: print "Could not reach the npm registry. Skipping update plan." and stop.

---

## Phase 4: Show changelog (best-effort)

Always print the version numbers regardless of network state:

```
ATV Starter Kit
  Latest:   <LATEST>
  Project:  <PROJECT_VERSION or "unknown">
  Plugins:  N installed (mixed versions: <list>)
```

Then attempt to fetch the changelog snippet:

```bash
curl -fsSL https://raw.githubusercontent.com/All-The-Vibes/ATV-StarterKit/main/CHANGELOG.md 2>/dev/null
```

If the fetch succeeds, attempt to extract the section between the heading for the current version and the heading for the latest version (heading format `## [X.Y.Z] — YYYY-MM-DD`). Print up to ~80 lines of that excerpt.

If the fetch or parse fails, just print:

> See https://github.com/All-The-Vibes/ATV-StarterKit/blob/main/CHANGELOG.md for what's new.

This is best-effort — never block the rest of the flow on changelog availability.

---

## Phase 5: Plan update

Build a structured plan and print it. Two sections:

### Project scaffold (advisory only)

If `hasProject` and project version is behind the latest, print:

```
📋 Project scaffold update (manual)

The installer is currently additive-only. `npx atv-starterkit@latest init`
will NOT refresh existing scaffold files. To update your project scaffold,
choose one approach:

Option A — Review and apply selectively (preserves customizations):
  • Visit https://github.com/All-The-Vibes/ATV-StarterKit
  • Diff the templates against your .github/ files
  • Apply changes that matter to you

Option B — Clean reinstall (loses local edits to ATV files):
  npx atv-starterkit@latest uninstall
  npx atv-starterkit@latest init  # or `init --guided`

Option C — Wait for installer refresh support (tracked as a known gap)
```

If `hasProject` but project version equals the latest: print "Project scaffold is up to date." and skip Options.

### Marketplace plugins (auto)

For each plugin from Phase 2 that's behind `LATEST`, list it as:

```
🔄 Marketplace updates available (Phase 6 will run these):
  • copilot plugin update <name>   (current: vX.Y.Z → latest: vA.B.C)
  • ...
```

If all plugins are up to date: print "All marketplace plugins are up to date."

### VS Code source-installed AgentPlugins (auto only when safe)

For each detected ATV source plugin:

```
🔄 Source AgentPlugin updates
  • All-The-Vibes/ATV-StarterKit
    Path: <exact path>
    State: clean, main behind origin/main by N commits
    Action: eligible for fast-forward in Phase 7
```

For blocked states, print the exact reason and do not include an automatic command in the apply plan:

```
⚠️ Source AgentPlugin requires manual review
  • EveryInc/compound-engineering-plugin
    Path: <exact path>
    State: dirty worktree
    Action: skipped — local edits would be overwritten by an update
```

If a reinstall/removal is the only practical remediation, present it as a separate destructive option, not as the default update path. It must show the exact folder that would be removed and require explicit confirmation before removal.

---

## Phase 6: Apply marketplace updates

**Skip this phase entirely in `dry-run` mode.**

Project scaffold updates are NEVER auto-applied — see Phase 5 rationale. Source AgentPlugin updates are handled in Phase 7.

For each marketplace plugin from Phase 5 that's behind:

1. Use `AskUserQuestion`:
   > Update `<plugin-name>` from v<current> to v<latest> now? (y/n)
2. On `y`: run `copilot plugin update <plugin-name>`. Capture output.
3. On `n`: skip and continue.
4. On error: report the failure clearly and continue with the next plugin.

After the loop, print a summary:

```
Updated N marketplace plugins.
Skipped M.
Errored on X (see above for details).
```

---

## Phase 7: Apply source plugin updates

**Skip this phase entirely in `dry-run` mode.**

For each source AgentPlugin classified as clean and behind-only:

1. Show the exact owner/repo, installed path, current commit, upstream branch, and behind count.
2. Use `AskUserQuestion`:
  > Fast-forward source AgentPlugin `<owner>/<repo>` at `<exact path>` now? (y/n)
3. On `y`: run a fast-forward-only update in that plugin directory.
4. On `n`: skip and continue.
5. On error: report the failure clearly and do not try destructive remediation automatically.

For blocked states:

- **Dirty worktree:** print the changed-file summary and stop. Do not stash, reset, or overwrite.
- **Ahead or diverged:** print ahead/behind counts and stop. The user must review local commits before update.
- **Detached HEAD or no upstream:** print that no safe automatic update target exists.
- **Fetch failed:** print that remote state is unknown and stop.

### Destructive reinstall/removal fallback

Only offer this after a clean in-place update is impossible or has failed, and only when the user asks to remediate.

Before removal:

1. Print exactly what would be lost: the full plugin folder path and all files inside it.
2. Ask for explicit confirmation naming that exact folder.
3. Remove only that plugin folder after confirmation.
4. Verify the folder and its metadata file are gone.
5. Tell the user to reinstall through VS Code `Chat: Install Plugin from source`.

Never remove an entire `agent-plugins`, `github.com`, `.vscode`, or `.vscode-insiders` directory.

---

## Phase 8: Verify and reload guidance

After applying any changes, suggest:

> Run `/atv-doctor` to verify the post-update install state.

If any source AgentPlugin was updated or reinstalled, also print:

> Reload or restart VS Code / VS Code Insiders so the AgentPlugin content is reloaded. If the old skills still appear, close all VS Code windows and reopen the workspace before re-running `/atv-doctor`.

Don't run `/atv-doctor` or reload VS Code automatically — let the user choose.

---

## What this skill does NOT do

- Auto-refresh project scaffold files (installer limitation; tracked as future work).
- Force-update dirty, ahead, diverged, detached, or untracked source AgentPlugin checkouts.
- Silently delete or reinstall VS Code AgentPlugin folders.
- Update gstack — that's `gstack`'s own responsibility (gstack syncs from its source repo).
- Update agent-browser — `npm install -g agent-browser` is manual.
- Pin specific versions — `copilot plugin install plugin@marketplace` doesn't accept a version. All ATV plugins share the kit version.
- Roll back to a previous version — `copilot plugin uninstall` then `install` to a specific tag is the manual path.
