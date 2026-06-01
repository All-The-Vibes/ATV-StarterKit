# Plan: Resumable Phase State + Artifact-by-Reference for LFG / SLFG

- **Status:** Implemented (this plan is the durable record; checkboxes reflect delivered work)
- **Type:** feat
- **Date:** 2026-06-01
- **Owner:** ATV maintainers
- **Origin:** Two workflow-runtime features were identified as borrowable into ATV's `lfg`/`slfg`
  markdown-skill system: (A) phase-completion markers for **resumability**, (B) passing phase
  outputs **by reference**.

## Problem

`lfg` and `slfg` are markdown skills that an LLM executes turn-by-turn, orchestrating a pipeline of
sub-skills (`ce-plan → ce-work → ce-review → todo-resolve → test-browser → feature-video`). Unlike
workflow engines that execute JS in a runtime, they have:

1. **No resumability** — an interruption restarts the whole turn; completed phases re-run.
2. **No artifact-by-reference discipline** — phase outputs flow through the agent's context.

The fix must be **extremely lightweight** because it runs on *every* `/lfg` and `/slfg` session.

## Decision: stack & architecture

**Reframe:** the state consumer is an LLM measuring cost in tokens + tool round-trips, not disk
microseconds. This settles the storage question.

- **Rejected SQLite** — needs a `sqlite3` binary (dependency, often absent); a binary file is opaque
  to the agent (a bash round-trip per peek); not git-diffable; `*.db` is already gitignored;
  "fast CRUD" is irrelevant for ~6–9 rows whose bottleneck is the LLM.
- **Rejected plan-checkbox reuse** — fuzzy to parse; mixes a durable artifact with ephemeral run
  state; multiple runs of one plan collide.
- **Chosen: small JSON files mutated by a tiny Node CLI helper**, mirroring the existing
  `observe.js` convention. Node is already a baseline hook dependency → zero new dependencies.

**Layout** (`.atv/runs/<run-id>/`, gitignored, local-only):

```
meta.json                 # written once at init; only the parent binds plan_path (single writer)
phases/<phase>.done.json  # one atomic sentinel per completed phase (temp-write + rename)
```

Per-phase sentinels eliminate read-modify-write races (critical for SLFG's parallel phase): "is this
phase done?" == "does the sentinel exist?". Each sentinel's `artifact` field is the by-reference
pointer — one structure solves **both** (A) and (B).

### Rubber-duck-driven refinements (folded in)

1. **Two-stage run identity** — deriving the id from the plan filename can't work *during* planning
   (chicken-and-egg). A provisional id (`slug(feature)+hash(repo|branch|feature)`) is created at
   init; after `ce-plan` succeeds the run binds `plan_path`. On resume, the orchestrator searches
   `docs/plans/` before re-running `ce-plan`.
2. **`ce-work mode:orchestrated`** — `ce-work` mutates state, so it is *not* blindly restarted; the
   orchestrated mode reconciles existing `git diff` / plan checkboxes / todos and skips PR + shipping.
3. **`run:<run-id>` parsed before threading** — sub-skills recognize-and-strip the token so an
   unknown token is never mis-read as a PR/branch arg.
4. **Parent-only state writes in SLFG** — parallel swarm subagents return artifact paths only; the
   parent records `done` after they join.

## Work breakdown

- [x] **Step 1 — Helper** `.github/hooks/scripts/lfg-state.js` (+ shipped copy
  `pkg/scaffold/templates/hooks/scripts/lfg-state.js`). Commands: `init`, `bind-plan`,
  `done <phase> [--artifact]`, `is-done`, `status`, `run-id-from-plan`. Pure functions exported for
  tests; CLI guarded by `require.main`. Atomic writes via temp-file + rename; phase-name sanitization.
  - Tests: `.github/hooks/scripts/tests/lfg-state.test.js` (18 cases, real temp-dir fs, no mocks).
- [x] **Step 2 — `.gitignore`** add `.atv/runs/` (ephemeral, local resume only).
- [x] **Step 3 — Sub-skill arg parsing**: `ce-plan`, `ce-work`, `ce-review` recognize-and-strip
  `run:<run-id>`; `ce-work` gains `mode:orchestrated` (resume-safe, no PR/ship). Both dogfood
  (`.github/skills/...`) and shipped (`pkg/scaffold/templates/skills/...`) copies.
- [x] **Step 4 — Orchestrator wiring**: `lfg` and `slfg` (both copies) get a *Run State* section —
  derive/resume run-id, skip `done` phases, mark each phase done with its artifact path, thread
  `run:<RUN_ID>`, and (SLFG) parent-only writes.
- [x] **Contract tests** `.github/hooks/scripts/tests/skill-contract.test.js` (22 cases) assert the
  skills actually adopt the protocol and that dogfood/template copies cannot drift.
- [x] **Regenerate** `plugins/` from templates (`go run ./cmd/plugingen`); parity check clean.

## Test scenarios (delivered)

- `lfg-state.test.js`: slugify bounds; deterministic provisional id; plan→run-id derivation;
  idempotent `init` (resume safety); `bindPlan`; atomic `markDone`/`isDone`; artifact recorded by
  reference; path-traversal rejection; compact `status`; missing-run returns empty (no error);
  `parseRunToken` extraction + passthrough.
- `skill-contract.test.js`: helper exported API present in both copies; LFG references helper,
  documents resume, threads `run:<...>`, marks done with `--artifact`; SLFG documents parent-only
  writes; every sub-skill recognizes `run:<run-id>`; `ce-work` exposes `mode:orchestrated`.
- End-to-end resume simulation: interrupt after `ce-plan`+`ce-work` → `status` resumes at
  `ce-review`; re-`init` preserves original meta.

## How to run the tests

```
node --test .github/hooks/scripts/tests/lfg-state.test.js .github/hooks/scripts/tests/skill-contract.test.js
go test ./pkg/scaffold/... ./pkg/plugingen/...
go run ./cmd/plugingen -check
```

## Scope boundaries (non-goals)

- No new runtime, daemon, or DB engine; no new dependencies.
- Cross-machine/fresh-checkout resume is intentionally out of scope — state is local-only; recovery
  there falls back to durable signals (plan checkboxes, branch, open PR), as the skills now state.
- Pre-existing content drift between dogfood and template skill copies (e.g. unslop/observe/learn
  steps) is preserved as-is; only the new resume wiring is kept in sync.
