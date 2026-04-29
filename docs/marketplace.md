# ATV Starter Kit — source install and Copilot CLI marketplace

The ATV Starter Kit is available through VS Code source install and as a [GitHub Copilot CLI plugin marketplace](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/plugins-marketplace) in addition to the project-scaffolding `npx atv-starterkit init` flow.

## Three install paths — pick whichever matches your need

| | `npx atv-starterkit init` | VS Code source install | `copilot plugin install …@atv-starter-kit` |
|---|---|---|---|
| **Where files land** | Your project's `.github/`, `.vscode/`, `docs/` | VS Code AgentPlugin directory | `~/.copilot/installed-plugins/` |
| **Scope** | Project-level, committed to git, shared with the team | Personal/editor-level | Personal, follows you across CLI projects |
| **What ships** | Skills + agents + MCP config + hooks + instructions + setup-steps + docs scaffolding | One complete ATV skills + agents bundle | Skills + agents only |
| **Stack-aware** | Yes — Python/Rails/TypeScript instructions and reviewer agents | No — complete personal bundle | No — install plugins manually per project |
| **Best for** | Bootstrapping a new repo, codifying team-wide AI workflow | VS Code Copilot users who want one obvious install choice | CLI users who want bundles or granular skills |

You can use these paths together. They write to different locations and Copilot CLI's loading precedence (project skills > user skills > plugin skills) means project-level ATV files always win when there's overlap.

## VS Code source install

In VS Code or VS Code Insiders:

1. Open the Command Palette.
2. Run `Chat: Install Plugin from source`.
3. Enter `All-The-Vibes/ATV-StarterKit`.
4. Choose `atv-starter-kit`.

The source-install catalog is intentionally curated to one option. That option points at the complete `atv-everything` bundle so VS Code users get all ATV skills and all reviewer/specialist agents without choosing from category packs or single-skill plugins.

Maintainer validation for releases: the VS Code source-install picker should show one ATV option with a concise description. If it shows `atv-pack-*`, `atv-skill-*`, or `atv-agents`, the source-install catalog has regressed.

## Add the Copilot CLI marketplace

```bash
copilot plugin marketplace add All-The-Vibes/ATV-StarterKit
```

Browse the CLI catalog:

```bash
copilot plugin marketplace browse atv-starter-kit
```

## Copilot CLI install tiers

### 1. Flagship — install everything

```bash
copilot plugin install atv-everything@atv-starter-kit
```

Bundles **every** ATV skill (31) + **every** reviewer/specialist agent (51). Equivalent in coverage to the Full preset of `atv init` (scoped to skills + agents — no MCP servers, hooks, or instructions templates).

Advanced standalone agents-only install:

```bash
copilot plugin install atv-agents@atv-starter-kit   # agents for category-pack or single-skill installs
```

### 2. Category packs — bundle related skills

| Pack | Skills | Use when |
|---|---|---|
| `atv-pack-planning` | brainstorming, ce-brainstorm, ce-ideate, ce-plan, deepen-plan | Shape work before coding |
| `atv-pack-review` | ce-review, document-review | Multi-agent review passes |
| `atv-pack-shipping` | takeoff, ce-work, ce-compound, ce-compound-refresh, land, lfg, slfg | Execute and ship |
| `atv-pack-security` | atv-security | Config audit + OWASP/STRIDE |
| `atv-pack-quality` | unslop, ralph-loop | Tighten up code, iterate |
| `atv-pack-guidelines` | karpathy-guidelines, autoresearch | Behavioral guardrails + autonomous experiment loop |
| `atv-pack-easter-eggs` | meme-iq | Fun extras |
| `atv-pack-learning` | learn, instincts, evolve, observe | Compounding institutional knowledge |

```bash
copilot plugin install atv-pack-planning@atv-starter-kit
copilot plugin install atv-pack-shipping@atv-starter-kit
# ...etc
```

### 3. Granular — single-skill plugins

For each skill listed above (and a few utility skills like `setup`, `feature-video`, `resolve_todo_parallel`, `test-browser`), there is an `atv-skill-<name>` plugin:

```bash
copilot plugin install atv-skill-autoresearch@atv-starter-kit
copilot plugin install atv-skill-atv-security@atv-starter-kit
copilot plugin install atv-skill-ce-plan@atv-starter-kit
```

> **Heads up:** category-pack and per-skill plugins include skills only. Several skills (`ce-plan`, `ce-ideate`, `deepen-plan`, `ce-review`, `document-review`) dispatch reviewer/research agents that are bundled separately in `atv-agents`. For the most predictable experience, install `atv-everything`. If you choose a category pack or single skill that dispatches agents, also install `atv-agents`.

## Skill dependencies (subset)

Skills that depend on agents bundled in `atv-agents`:

| Skill | Why |
|---|---|
| `ce-review` | Dispatches the full compound-engineering reviewer fleet. Falls back to bundled agents when missing. |
| `document-review` | Dispatches document-review specialist agents. |
| `ce-plan`, `ce-ideate`, `deepen-plan` | Dispatch research agents during discovery/ideation. Degrade gracefully if missing. |

Skills that reference the [compound-engineering plugin](https://github.com/EveryInc/compound-engineering-plugin) (optional, separate install):

- `ce-plan`, `ce-ideate`, `deepen-plan`, `ce-review`, `document-review` — all degrade gracefully when compound-engineering is not installed.

## What the marketplace does NOT install

These remain available only through `npx atv-starterkit init` (project-level scaffolding):

- **MCP server config** (`.github/copilot-mcp-config.json`) — generic server names like `github`, `azure`, `terraform`, `context7` would silently override existing user MCP setups under last-wins precedence. Will ship as `atv-mcp` in a future release with a namespacing strategy.
- **Hooks** (`.github/hooks/`) — wired into project paths.
- **Copilot instructions templates** (`.github/copilot-instructions.md`) — stack-specific, need substitution.
- **Setup steps** (`.github/copilot-setup-steps.yml`) — environment bootstrapping.
- **Docs structure scaffolding** (`docs/plans/`, `docs/brainstorms/`, `docs/solutions/`).

## Plugin overlap & precedence

Copilot CLI uses **first-found-wins** for skills/agents and **last-wins** for MCP servers (project beats user beats plugin). If you install both `atv-everything` and `atv-pack-shipping`, the plugin loaded first determines which version of `ce-work` is used. **Install one ATV scope at a time** to keep behaviour predictable.

## Versioning

All ATV plugins share the kit's release cadence — e.g. `atv-skill-autoresearch@2.6.1`, `atv-pack-planning@2.6.1`, etc. Note that `copilot plugin install plugin@marketplace` does not accept a version (the version field in `plugin.json` is metadata). To pin to a specific kit version, install from a tagged commit:

```bash
copilot plugin install ./   # from a local clone checked out at the desired tag
```

## How the catalogs are generated

`plugins/`, root `marketplace.json`, `.github/plugin/marketplace.json`, and `.claude-plugin/marketplace.json` are generated from `pkg/scaffold/templates/` by `pkg/plugingen/` and the `cmd/plugingen` CLI. Run:

```bash
go run ./cmd/plugingen        # regenerate after editing any template
go run ./cmd/plugingen -check # CI dry-run mode (exits 1 on drift)
```

CI runs the `-check` mode automatically. If templates or generated catalog metadata change without a regeneration, CI fails with a clear "plugin tree out of sync" error.

Catalog intent:

- `marketplace.json` — VS Code source install, one curated `atv-starter-kit` option. VS Code checks this before `.github/plugin/marketplace.json`.
- `.claude-plugin/marketplace.json` — mirrored curated source-install catalog for Claude-format compatibility.
- `.github/plugin/marketplace.json` — Copilot CLI marketplace, granular bundles and per-skill plugins.

If VS Code source install ever ignores root `marketplace.json` and reads `.github/plugin/marketplace.json` instead, prioritize the clean one-option VS Code experience and make the CLI granularity tradeoff explicit before release.

## Further reading

- [Creating a plugin marketplace for GitHub Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/plugins-marketplace)
- [GitHub Copilot CLI plugin reference](https://docs.github.com/en/copilot/reference/cli-plugin-reference)
