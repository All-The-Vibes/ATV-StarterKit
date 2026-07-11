---
name: kb-map
description: Project-memory setup, lookup, and refresh skill for KB workflows. Use when starting in a repo, when another KB skill needs the right architecture docs without broad repo crawling, or when the user says "map", "project context", "where is this", "what should I read", "setup memory", or "bootstrap context". If memory is missing or badly stale, invokes kb-map-bootstrap.
argument-hint: "[lookup|refresh] [optional task or subsystem]"
---

# KB Map

Use local project memory so fresh sessions do not need the user to reteach the app.

This skill owns the project-memory preflight. `kb-start` and other skills should call `kb-map` instead of checking bootstrap rules themselves.

Keep normal lookup cheap. Deep indexing belongs in `kb-map-bootstrap`.

## Project Root Rule

Anchor every lookup to the active project root before reading memory.

1. Determine the project root:
   - Prefer the current working directory's Git root: `git rev-parse --show-toplevel`.
   - If Git is unavailable, use the current working directory only when it is clearly a project directory.
   - Treat drive roots such as `E:\`, home directories such as `~`/`%USERPROFILE%`, `.copilot`, `.codex`, and `.agents` as invalid project roots unless the user explicitly chose them.
   - If the resolved root is invalid or not the user's intended project, ask the user to change into the project directory or provide the project path before searching.
2. Read memory only from that root:
   - `<repo>/todo.md`
   - `<repo>/docs/context/PROJECT.md`
   - `<repo>/docs/context/landmines.md` when it exists
   - `<repo>/docs/handoffs/**`
3. Do not search `~`, `%USERPROFILE%`, `.copilot/handoffs`, the whole drive, or sibling repos for KB memory unless the user explicitly asks for cross-repo/global lookup.
4. If the project root has no KB memory, invoke `kb-map-bootstrap` in that project root. Do not silently substitute a global or unrelated handoff.

This prevents the agent from picking up stale personal handoffs when the user is working inside a specific repo.

Forbidden fallback: do not use glob/search to find `todo.md`, `PROJECT.md`, or handoffs when the project root is unresolved. Resolve the root first or ask the user.

## Contract Check

Before lookup or refresh, check the standard layout.

- Standard memory files must be checked by exact path under the project root, not by grep/glob:
  - `<repo>/todo.md`
  - `<repo>/docs/context/PROJECT.md`
  - `<repo>/docs/handoffs/active/`
  - `<repo>/docs/handoffs/parked/`
  - `<repo>/docs/handoffs/done/`
- If `todo.md` or `docs/context/PROJECT.md` is missing, invoke `kb-map-bootstrap`.
- If only directories are missing, create them during `refresh`.
- Never overwrite non-empty user docs without reading and merging.
- After bootstrap or refresh, continue the original lookup so the caller receives route-ready context.

## Large Repo Recheck

Normal `kb-map lookup` must stay cheap. Do not run graphify every time. Instead,
use a cheap size/staleness guard after the contract check and before trusting a
large or bloated project map.

Check:

- `docs/context/PROJECT.md` size.
- whether `docs/context/memory-maintenance.md` contains a recent
  `graphify-size-check` line.
- repo code-file count only when the last check is missing, older than 30 days,
  or `PROJECT.md` crosses a size threshold.

Thresholds:

- `PROJECT.md` over 40 KB or about 900 lines: treat the project map as too large
  for startup. **Before recommending graphify or a split-refresh, first check
  whether the bloat is ephemeral work-log pollution** — count dated/append-style
  sections (e.g. `## YYYY-MM-DD ...`, "render progress checkpoint", "audio QA
  update", per-run status entries). If most of the file is that churn, the fix is
  EVICTION, not graphify: move the dated log entries to `todo-done.md` or
  `docs/context/archive/<topic>-<yyyy-mm>.md` and keep only the durable route map
  (entry points, main areas, current state, routing). Only recommend a
  child-doc split-refresh when the DURABLE map itself is genuinely large after
  eviction. Graphify does not shrink a log-polluted map and is the wrong remedy
  for it.
- repo over 200 code files with no `graphify-size-check` in the last 30 days:
  run a targeted size recheck and consider graphify-assisted refresh.
- repo over 80 code files and the current lookup requires caller, callee,
  impact, dependency, or subsystem-boundary rediscovery: consider graphify if
  the existing docs do not answer the question.

**`PROJECT.md` is a route map, never a work log.** Do not append (and instruct
other skills not to append) render progress, checkpoints, per-chapter QA updates,
or dated run status into `PROJECT.md`. Those belong in `todo.md` (work log) and
`todo-done.md` (completed summaries). A project map that accumulates dated
checkpoint sections is a bug to evict, not a size that justifies graphify.

Record rechecks in `docs/context/memory-maintenance.md`:

```text
graphify-size-check: 2026-06-03 code_files=402 project_md_bytes=18422 decision=skip|consider|use reason=<short reason>
```

If the check decides `use`, delegate to `kb-map-bootstrap refresh <subsystem>`
or a targeted coverage audit rather than running raw graphify inside ordinary
lookup. After refresh, rerun the original lookup.

Load `references/graph-routing.md` only when the guard says graph routing may
pay for itself. It contains the graphify/TokenMasterX mechanics, graph route
row shape, and evidence rules.

## Modes

| Mode | Use When | Cost |
|---|---|---|
| `preflight` | Another skill needs memory verified before routing | low to high only if bootstrap is needed |
| `lookup` | Memory exists; find the right context for the current request | low |
| `refresh` | Recent work changed project memory or route pointers | medium |
| `setup` | User explicitly wants memory initialized | high; delegates to `kb-map-bootstrap` |

Default to `lookup`.

### Optional Model Routes During Explicit Setup

Only when the user explicitly asks to set up the project—not when lookup
silently bootstraps missing memory—offer one optional question after memory
setup succeeds: `Configure local/private model routes now?`

Normal `preflight`, `lookup`, `refresh`, and bootstrap triggered by missing or
stale memory ask no model or routing questions. Host-native discovery and model
selection remain automatic.

- If no, finish setup without creating model state.
- If yes, invoke `kb-models configure`; do not collect endpoints, credentials,
  transports, or model IDs inside `kb-map`.
- `kb-models` owns supported connection types, actual route details, and the
  optional user-local project source preference: `automatic`,
  `self-hosted-first`, or `native-first`.
- Do not imply that every transport is supported. Host-native and built-in
  OpenAI-compatible/LiteLLM routes may be configured today; generic MCP model
  dispatch requires a versioned adapter before it is offered as executable.

## Standard Layout

```text
todo.md
todo-done.md
docs/context/
  PROJECT.md
  architecture/
    README.md
    <major-subsystem>.md
  research/
    README.md
    <topic>.md
  decisions/
    README.md
  operations/
    README.md
    testing.md
docs/handoffs/
  active/
  parked/
  done/
```

## Lookup Mode

Read in order:

1. `todo.md`.
2. `docs/context/PROJECT.md`.
3. `docs/context/landmines.md` if it exists. Read only `Active Landmines`; do
   not load resolved/archive entries during startup.
4. Active handoff files linked from `todo.md`.
5. Only the subsystem, research, decision, operation, brainstorm, or plan files needed for the request.

Stop reading once you can answer:

- What app/repo is this?
- What is active, blocked, parked, or queued?
- Which subsystem is relevant?
- Which files or commands are likely involved?
- What was already tried or researched?
- Which KB lane should handle the request?
- Are there active repo-specific landmines that apply to this request?

Report route, docs loaded, and any stale-work refresh needed. Do not bulk-load all context docs.

Do not use `rg`, glob, or whole-repo search to find the standard memory files. Use search only after the exact project-root memory files are loaded and only for task-specific context.

### Graph Route Pointers

`PROJECT.md` is the routing surface. For large or structural subsystems, it may
point to the graph instead of listing every relevant source file.

Use one orientation path, not both:

- If `PROJECT.md` points to an architecture doc or a small source file list,
  follow those docs/files and do not run graphify.
- If `PROJECT.md` points to a `graph_route`, use the graph for structural
  orientation, then verify only the cited source edges needed for the task.
- If `PROJECT.md` has both a child doc and a `graph_route`, treat the child doc
  as the product/architecture summary and the graph as the caller/callee/impact
  lookup surface. Do not re-enumerate the whole subsystem manually.

When a graph route exists but `.token-master/graph.json` is missing or stale,
run a targeted `refresh` or delegate to `kb-map-bootstrap` rather than falling
back to broad rediscovery inside lookup.

## Coverage Gap Rule

If lookup cannot get a fresh session meaningfully up to speed on the requested
subsystem from `PROJECT.md` plus pointed docs, treat that as a project-memory
coverage gap.

Load `references/coverage-refresh.md` only when coverage is insufficient. That
reference contains the gap signals, audit mode, shallow-map detection, and
targeted refresh steps.

Example: an installer or release workflow that spans build config, packaging
scripts, CI workflows, release assets, and runtime startup checks needs its own
architecture pointer or child doc. A generic runtime doc is not enough if it
cannot explain how the workflow is built, shipped, and updated.

## Coverage Audit Mode

Use `references/coverage-refresh.md` when the user says the initial map missed
a whole class of things, or when one invisible subsystem suggests there may be
more. Keep ordinary lookup narrow unless the requested work touches the gap.

## Shallow Map Detection

During lookup, warn and recommend `refresh` or `kb-map-bootstrap` when the KB
appears much thinner than the repo. Load `references/coverage-refresh.md` for
the specific signals and response path.

## Missing Memory and Setup

If `todo.md` or `docs/context/PROJECT.md` is missing, invoke `kb-map-bootstrap`.

If handoff directories are missing but the project map exists, create or recommend the missing directories during `refresh`; do not deep-crawl the repo.

Use `setup` when the user explicitly wants to initialize KB memory. It always delegates to `kb-map-bootstrap` unless the standard layout already exists, in which case run `refresh`.

`kb-map-bootstrap` is the expensive first-pass mapper. `kb-map` is the durable entry point that decides whether bootstrap is needed.

## Refresh Mode

Use after meaningful architecture, workflow, or project-memory changes.

Refresh is required when work changes:

- User-visible behavior, feature boundaries, or major workflows.
- API contracts, data models, storage, auth, permissions, routing, streaming, tools, actions, jobs, or integrations.
- Build, run, test, deploy, or QA commands.
- Subsystem ownership, entry points, or first files a fresh session should read.
- Known sharp edges, rejected approaches, or "do not repeat" lessons.
- A lookup exposed a coverage gap: the map could not explain a named subsystem
  without broad rediscovery.

Refresh is usually not required for:

- Pure styling, copy, formatting, lint-only changes, dependency lockfile churn, or isolated tests that do not change behavior.

When unsure, write a one-line manifest or `todo.md` note explaining why refresh was skipped or required.

Workflow:

1. Read `docs/context/PROJECT.md`.
2. Inspect changed files, recent manifests, active handoffs, and `todo.md`.
3. Update only affected subsystem docs and indexes.
4. Add child docs when a parent doc grows too large.
5. Update `todo.md` if active state, blockers, or pointers changed.
6. Update active handoff files if restart instructions changed.
7. If `docs/context/landmines.md` exists, archive resolved landmines whose
   owner surface was fixed and verification passed. Leave unfixed stale entries
   active or mark `stale-review`; do not silently delete high-severity traps.
8. If `todo_rules.md`, `todo-rules.md`, or another separate todo rules file exists, inline the rules into the top `## Rules` section of `todo.md`, move any unique durable content to `docs/context/*`, then delete the separate rules file.
9. Run `document-review` when changes are substantial.

Do not re-bootstrap the whole repo here.

## Contracts

`PROJECT.md` is a route map, not an encyclopedia. Subsystem docs carry durable app truth. `todo.md` carries current operational truth and its own board rules. `todo-done.md` carries completed-work summaries. Handoff files carry resumable work packets.
