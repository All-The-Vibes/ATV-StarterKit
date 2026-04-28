# ATV Starter Kit — Copilot CLI plugin marketplace

The ATV Starter Kit is available as a [GitHub Copilot CLI plugin marketplace](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/plugins-marketplace) in addition to the project-scaffolding `npx atv-starterkit init` flow.

## Two install paths — pick whichever matches your need

| | `npx atv-starterkit init` | `copilot plugin install …@atv-starter-kit` |
|---|---|---|
| **Where files land** | Your project's `.github/`, `.vscode/`, `docs/` | `~/.copilot/installed-plugins/` |
| **Scope** | Project-level, committed to git, shared with the team | Personal, follows you across projects |
| **What ships** | Skills + agents + MCP config + hooks + instructions + setup-steps + docs scaffolding | Skills + agents only |
| **Stack-aware** | Yes — Python/Rails/TypeScript instructions and reviewer agents | No — install plugins manually per project |
| **Best for** | Bootstrapping a new repo, codifying team-wide AI workflow | Personal/cross-project skills you want everywhere |

You can use both. They write to different locations and Copilot CLI's loading precedence (project skills > user skills > plugin skills) means project-level ATV files always win when there's overlap.

## Add the marketplace

```bash
copilot plugin marketplace add All-The-Vibes/ATV-StarterKit
```

Browse the catalog:

```bash
copilot plugin marketplace browse atv-starter-kit
```

## Three install tiers

### 1. Flagship — install everything

```bash
copilot plugin install atv-everything@atv-starter-kit
```

Bundles **every** ATV skill (29) + **every** reviewer/specialist agent (51). Equivalent in coverage to the Full preset of `atv init` (scoped to skills + agents — no MCP servers, hooks, or instructions templates).

You probably also want:

```bash
copilot plugin install atv-agents@atv-starter-kit   # already included in atv-everything; install standalone if you only want the agents
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

> **Heads up:** per-skill plugins are NOT standalone for all skills. Several skills (`ce-plan`, `ce-ideate`, `deepen-plan`, `ce-review`, `document-review`) dispatch reviewer/research agents that are bundled separately in `atv-agents`. For the most predictable experience, install `atv-pack-*` (which includes the relevant agents through the marketplace's recommended bundling) or `atv-everything`.

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

## How the marketplace is generated

`plugins/` and `.github/plugin/marketplace.json` are generated from `pkg/scaffold/templates/` by `pkg/plugingen/` and the `cmd/plugingen` CLI. Run:

```bash
go run ./cmd/plugingen        # regenerate after editing any template
go run ./cmd/plugingen -check # CI dry-run mode (exits 1 on drift)
```

CI runs the `-check` mode automatically. If templates change without a regeneration, CI fails with a clear "plugin tree out of sync" error.

## Further reading

- [Creating a plugin marketplace for GitHub Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/plugins-marketplace)
- [GitHub Copilot CLI plugin reference](https://docs.github.com/en/copilot/reference/cli-plugin-reference)
