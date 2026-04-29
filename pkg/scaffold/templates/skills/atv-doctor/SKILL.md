---
name: atv-doctor
description: "Diagnose ATV Starter Kit installation health across project scaffold, Copilot CLI marketplace plugins, and VS Code source-installed AgentPlugins. Detects install scope, version drift, AgentPlugin git state, file integrity, hook validity, MCP prereqs, and optional dependency status. Triggers on 'atv doctor', 'atv health', 'check atv', 'diagnose atv', 'atv status', 'atv check', 'atv healthcheck', 'is atv ok'."
argument-hint: "[mode: report (default) | fix]"
---

# /atv-doctor — ATV installation health check

Diagnose your ATV Starter Kit installation. Checks project scaffold via `atv init`, Copilot CLI marketplace plugins via `copilot plugin install`, and VS Code source-installed AgentPlugins via `Chat: Install Plugin from source`. Reports version drift, AgentPlugin git state, file integrity, hook validity, MCP prereqs, and optional dependency status.

## Arguments

<mode> #$ARGUMENTS </mode>

**Mode detection:** Check if arguments contain `fix` (case-insensitive) → fix mode. Otherwise → report mode (default).

## Execution Flow

```
Phase 1: Detect install scope        → repo-artifact + marketplace + VS Code AgentPlugin roots
Phase 2: Version check               → installed vs latest npm / source git state
Phase 3: Source AgentPlugin health   → owner/repo, path, commit, version metadata, divergence
Phase 4: File integrity (project)    → manifest checksum verification (when manifest exists)
Phase 5: Hook validity (project)     → JSON parse + script existence
Phase 6: MCP prereqs (project)       → parse inputs[] from MCP config
Phase 7: Optional dep gating         → only warn for deps the user opted into
Phase 8: Output graded report
Phase 9: Fix mode (opt-in, marketplace plugins only)
```

---

## Phase 1: Detect install scope

Use `Bash` and `read_file` to determine which install paths are present. **Detection is repo-artifact-based, not manifest-based** — auto-mode `atv init` does not write `.atv/install-manifest.json`, so the manifest is treated as optional metadata.

Set these flags from the working directory's repo root (or `~/.copilot/` for marketplace):

| Flag | Detection |
|------|-----------|
| `hasProject` | true if **any** of these exist: `.github/skills/` (directory), `.github/copilot-instructions.md`, `.github/copilot-mcp-config.json`, `.github/hooks/copilot-hooks.json` |
| `hasManifest` | true if `.atv/install-manifest.json` exists (used to enable Phases 3 & 5; not gating) |
| `hasMarketplace` | true if `~/.copilot/installed-plugins/atv-starter-kit/` exists OR `copilot plugin marketplace list` mentions `atv-starter-kit` |
| `hasSourceAgentPlugins` | true if any git checkout exists under a VS Code AgentPlugin root such as `$HOME/.vscode/agent-plugins/github.com/<owner>/<repo>` or `$HOME/.vscode-insiders/agent-plugins/github.com/<owner>/<repo>` |

On Windows, `$HOME` may resolve through Git Bash and `%USERPROFILE%` may be clearer in PowerShell. Probe both Stable and Insiders roots without hardcoding a machine-specific username:

| Product | Root examples |
|---------|---------------|
| VS Code Stable | `$HOME/.vscode/agent-plugins/github.com`, `%USERPROFILE%\.vscode\agent-plugins\github.com` |
| VS Code Insiders | `$HOME/.vscode-insiders/agent-plugins/github.com`, `%USERPROFILE%\.vscode-insiders\agent-plugins\github.com` |

If `!hasProject && !hasMarketplace && !hasSourceAgentPlugins`: print "No ATV install detected. To install: `npx atv-starterkit init` (project scaffold), `copilot plugin marketplace add All-The-Vibes/ATV-StarterKit && copilot plugin install atv-everything@atv-starter-kit` (marketplace), or VS Code `Chat: Install Plugin from source` with `All-The-Vibes/ATV-StarterKit`." and stop.

Record what was detected for the report header.

---

## Phase 2: Version check

For each detected scope, determine the installed version and compare to the latest npm release.

### 2a. Latest version (always)

```bash
npm view atv-starterkit version 2>/dev/null
```

If the call fails (offline, registry down), record "latest: unknown" and proceed — version comparison becomes informational.

### 2b. Project scaffold version

If `hasManifest`: read `.atv/install-manifest.json` and extract the `catalogVersion` string field:

```bash
node -e "console.log(JSON.parse(require('fs').readFileSync('.atv/install-manifest.json','utf8')).catalogVersion || 'unknown')"
```

If `hasProject && !hasManifest`: report "project install detected but no manifest. Auto-mode `atv init` does not write a manifest. Re-run `atv init --guided` to opt into manifest-tracked state." Don't fail — this is informational.

### 2c. Marketplace plugin versions

Walk `~/.copilot/installed-plugins/atv-starter-kit/` directly (more robust than parsing `copilot plugin list` text output). For each subdirectory containing a `plugin.json`, read its `name` and `version` fields:

```bash
for dir in ~/.copilot/installed-plugins/atv-starter-kit/*/; do
  if [ -f "$dir/plugin.json" ]; then
    node -e "const p = require('$dir/plugin.json'); console.log(p.name + '@' + (p.version || 'unknown'))"
  fi
done
```

### 2d. Compare

For each installed version, compare to latest. Use semver-aware comparison if possible; otherwise string compare and flag any difference as 🟡 with the suggestion to run `/atv-update`.

### 2e. Source-installed AgentPlugin versions

If `hasSourceAgentPlugins`, inspect each discovered checkout under the VS Code AgentPlugin roots. Include ATV and other GitHub-source AgentPlugins such as Compound Engineering in the report, because stale source installs are easiest to diagnose when the actual installed directory is inspected.

For each plugin directory:

1. Derive `owner/repo` from the `github.com/<owner>/<repo>` path segments.
2. Read version metadata from the first available source:
   - `package.json` `version`
   - `VERSION`
   - `.claude-plugin/plugin.json` `version`
   - `plugin.json` `version`
3. Read git state from inside that directory:
   - current branch or detached HEAD
   - current commit
   - `git describe --tags --always` when available
   - `origin` URL when available
   - dirty state from `git status --short`
   - upstream tracking branch and ahead/behind counts when available
4. Best-effort fetch remote state before ahead/behind comparison. If fetch fails because the user is offline or the remote is unavailable, report remote drift as unknown and keep going.

Do not mutate source AgentPlugin checkouts in `/atv-doctor`. This phase is read-only except for best-effort git remote metadata fetch.

---

## Phase 3: Source AgentPlugin health

Summarize source-installed AgentPlugins separately from Copilot CLI marketplace plugins.

Suggested report shape:

```markdown
**VS Code source-installed AgentPlugins:**
- All-The-Vibes/ATV-StarterKit
  - Path: <exact installed path>
  - Version: 2.6.3 (from VERSION)
  - Git: main @ b2a2d11 (v2.6.3)
  - Remote: origin/main, clean, behind by 0 / ahead by 0 🟢
- EveryInc/compound-engineering-plugin
  - Path: <exact installed path>
  - Version: 3.2.0 (from package.json)
  - Git: main @ <commit> (<describe>)
  - Remote: origin/main, behind by N 🟡 update available
```

Severity rules:

| State | Severity | Meaning |
|-------|----------|---------|
| clean and aligned with upstream | 🟢 ok | Source plugin checkout appears current |
| clean and behind upstream | 🟡 warn | Update available; run `/atv-update` |
| dirty worktree | 🟡 warn | Local edits present; update must not overwrite silently |
| ahead of upstream | 🟡 warn | Local commits present; update needs manual review |
| diverged from upstream | 🔴 critical | Both local and remote changed; do not auto-update |
| detached HEAD or no upstream | ⚪ info | Version can be reported, but update path is manual |
| missing version metadata | ⚪ info | Git state is still useful; version is unknown |

---

## Phase 4: File integrity (project scaffold, manifest required)

Skip this phase entirely when `!hasManifest` and emit:

> ⚪ Integrity check requires a manifest. Auto-mode `atv init` does not write one. Run `atv init --guided` next time to enable checksum-based integrity verification.

When `hasManifest`, read `.atv/install-manifest.json` and iterate the `fileChecksums` map (key = repo-relative path, value = SHA-256). For each entry:

```bash
# Compute current SHA-256 (cross-platform: prefer node)
node -e "const c=require('crypto'),fs=require('fs'); console.log(c.createHash('sha256').update(fs.readFileSync(process.argv[1])).digest('hex'))" <path>
```

| Comparison | Severity | Meaning |
|------------|----------|---------|
| File missing | 🔴 critical | Scaffold file expected but absent |
| Checksum matches | 🟢 ok | Unmodified — would be safely overwritten by a future installer refresh |
| Checksum differs | ⚪ info | User-modified — `atv init` would preserve this file |

Group findings by severity in the final report (don't print one line per file unless asked).

---

## Phase 5: Hook validity (project scaffold only)

Skip when `!hasProject` or `.github/hooks/copilot-hooks.json` is absent.

1. Parse `.github/hooks/copilot-hooks.json` as JSON. If parse fails: 🔴 with the parse error and the file path.
2. For each hook entry, walk its referenced script paths under `.github/hooks/scripts/`. If any referenced script is missing: 🔴 with the missing path and the hook key that references it.
3. If all scripts present: 🟢 with a count of hooks verified.

---

## Phase 6: MCP prereqs (project scaffold only)

Skip when `!hasProject` or `.github/copilot-mcp-config.json` is absent.

Parse the MCP config:

1. **Inputs** — read the top-level `inputs` array (if present). Each entry typically looks like `{ "id": "github_pat", "type": "promptString", "description": "..." }`. List the input IDs that any server depends on.
2. **Servers** — for each entry under `servers`, scan its `env` block for `${input:foo}` references. Build a list of which inputs are required by which servers.
3. **Document, don't enforce** — Copilot CLI prompts for missing inputs at runtime. Just produce a friendly summary:

```
MCP servers configured:
  - github (requires input: github_pat)
  - azure  (uses Azure CLI auth)
  - terraform (no auth required)
  - context7 (uses default API key)
```

**Do NOT hardcode env var names.** Always parse the actual config file — variable names drift between releases (`github_pat` vs `github_token`, etc.).

---

## Phase 7: Optional dep gating

For each optional tool, only warn when there's evidence the user wants it. **Never warn unconditionally.**

| Tool | Probe | Warn only when |
|------|-------|----------------|
| `bun` | `bun --version` | `requested.gstackRuntime == true` in manifest, OR any gstack runtime-requiring skill exists in `.github/skills/` |
| `agent-browser` | `agent-browser --version` | `requested.includeAgentBrowser == true` in manifest, OR `.github/skills/agent-browser/` exists |
| `gh` | `gh --version` | MCP config references the `github` server |
| `az` | `az --version` | MCP config references the `azure` server |
| `node` | `node --version` | always check (required for ATV usage) |

When `!hasManifest`, fall back to ⚪ informational status for everything optional ("might be needed depending on which features you use") rather than 🟡 warnings.

---

## Phase 8: Output

Print a graded report. Use this skeleton:

```markdown
## 🩺 ATV Doctor Report

**Detected:**
- Project scaffold: ✓ (manifest: yes/no)
- Marketplace plugins: ✓ (N installed)
- Source AgentPlugins: ✓ (N detected)

**Versions:**
- Latest on npm: 2.6.3
- Project (catalogVersion): 2.6.2 🟡 update available
- atv-skill-autoresearch: 2.6.2 🟡 update available
- atv-pack-planning: 2.6.3 🟢 up to date

### 🔴 Critical
- ...

### 🟡 Warn
- ...

### 🟢 OK
- ...

### ⚪ Info
- ...

**Next steps:**
- Run `/atv-update` to update marketplace plugins or clean source-installed ATV checkouts.
- Project scaffold updates require manual review (today's installer is additive-only).
```

If zero non-info findings: "Your ATV install looks healthy! 🩺"

---

## Phase 9: Fix mode (opt-in)

Only runs when `mode=fix`. **Limited scope today** — only Copilot CLI marketplace plugin updates are auto-fixable from doctor mode. Project scaffold "fixes" require the user to manually re-run the installer because today's `atv init` is additive-only and would not refresh existing files. Source AgentPlugin updates belong in `/atv-update`, where the workflow can inspect git state and ask for explicit confirmation before changing a checkout.

For each stale marketplace plugin:

1. Use `AskUserQuestion`:
   > Update `<plugin-name>` from v<current> to v<latest>? (y/n)
2. On confirmation: `copilot plugin update <plugin-name>`
3. On error: report the failure and continue with the next plugin.

After all fixes, print a summary: "Updated N plugins, skipped M."

**Constraints:**
- Never auto-rerun `atv init` (today's installer is additive-only — would not actually refresh files).
- Never modify scaffold files directly.
- Never update, reset, delete, or reinstall VS Code AgentPlugin folders from `/atv-doctor`.
- Always confirm before running anything.

---

## What this skill does NOT do

- Refresh project scaffold files (installer is additive-only today; tracked as future work).
- Update VS Code source-installed AgentPlugins directly (use `/atv-update`).
- Validate MCP server connectivity (would require network calls to each server).
- Run JSON-schema validation on configs (just parse + key presence).
- Auto-install missing optional dependencies (Bun, agent-browser, gh, az).
- Replace `npx atv-starterkit uninstall` for full removal.
