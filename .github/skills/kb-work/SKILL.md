---
name: kb-work
description: "Bounded swarm executor for kb-plan output. Runs all ready vertical slices from the dependency DAG, parallelizing only independent/isolated work, with TDD enforcement, scope gates, and HITL pauses. Use when the user says 'kb work', 'work the plan', 'execute the plan', 'run the KB pipeline', 'execute all slices', or wants guided execution of a planned feature."
argument-hint: "[path to KB manifest, or blank to find latest]"
---

# KB Work - Bounded Swarm Slice Executor

Run all vertical slices from a `kb-plan` manifest by pulling the safe ready set
from the dependency DAG. Keep each slice tied to its acceptance criteria,
enforce the requested verification mode, and pause on HITL tasks.

## Quick Start

1. Read the KB manifest.
2. Validate the dependency DAG and statuses.
3. Verify the manifest `gate_ledger` allows `kb-work`; repair or block if `plan-to-work` is not passed.
4. Confirm execution once unless the user already asked to run/execute/work the manifest.
5. Execute the safe ready set without asking between slices.
6. Update the manifest after each slice so the workflow is resumable.
7. After all runnable slices are terminal, write the `work-to-complete` gate and immediately invoke `kb-complete <manifest-path>` unless the user explicitly said to stop before completion.

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Scan `docs/plans/` for the most recent `*-kb-*-manifest.md` file. If found, use it. If no manifest exists, scan `todo.md`, `docs/brainstorms/`, and `docs/requirements/` for one active unplanned source. If exactly one exists, invoke `kb-plan <source>` with execution intent, then return to `kb-work <manifest-path>`. If none or multiple plausible sources exist, ask for the feature/source to slice; do not execute unplanned work.

**If input is a path:** Read the manifest at that path.

**If input is a handoff:** Do not execute the handoff directly. If it links a `docs/plans/*-kb-*-manifest.md`, use that manifest. If it contains only phases, workstreams, bullets, or broad next steps, stop and invoke `kb-plan` to create vertical slices first.

**If input is a feature description, broad task, or "go straight to work" request instead of a manifest path:** invoke `kb-plan <input>` with execution intent, then return to `kb-work <manifest-path>`. `kb-work` only executes KB manifests with per-slice plans and an initial `expected_files` forecast.

## Continuous Completion Loop

When the user invokes `kb-work` with execution intent, this skill owns the loop
until the work is truly terminal.

Terminal means one of:

- every slice is `done` or intentionally `skipped`, then `kb-complete` has run
  through review, follow-up resolution, proof, memory, and cleanup;
- the only remaining slices are `blocked`, `human-required`, or `parked`, with
  exact resume criteria recorded in `todo.md` and the manifest;
- the user explicitly says to pause or stop.

Default WIP is the safe ready set, not one slice. A slice is ready when its
blockers are `done` or `skipped`, its status is `pending`, and it is not marked
as a serial-only gate. Dispatch ready slices together only when the runtime gives
each active slice an isolated checkout/context or an equivalent write-isolation
guarantee. On a shared checkout, WIP is one mutating slice at a time.

`expected_files` is a forecast, not proof of disjointness. If active slices
observe or claim writes to the same path, serialize or requeue one slice before
continuing. Observed overlap beats planned disjointness.

Do not stop at weaker milestones:

- "the current slice passed";
- "all slices are done";
- "tests passed";
- "review started";
- "I wrote the summary."

Those are progress states. The next action is still to continue the loop, either
to the next runnable slice or to `kb-complete`.

If a repo has a project-specific `done.md` contract such as "can't stop til its
done", treat it as this same terminal rule. Do not create a new `done.md` from
the global skill; use `todo.md`, `todo-done.md`, manifests, and handoffs as the
KB state system unless the repo already opted into `done.md`.

## Pre-Flight

1. **Read the manifest** - parse the YAML frontmatter to get the ordered slice list.
2. **Validate DAG** - confirm no cycles in blockers, all referenced slice IDs exist, and all slice files exist.
3. **Validate gate ledger** - read `gate_ledger`. The `plan-to-work` gate must be `passed` and its `allowed_next_action` must name this manifest. Run `kb-gate/scripts/check_gate_ledger.py <manifest-path> --gate plan-to-work --allowed-next "kb-work <manifest-path>"` before execution. If the gate is absent, pending, blocked, stale, or the checker fails, stop and route to `kb-plan`/`kb-gate` to repair it before execution. Do not execute from a manifest that lacks a passing gate.
4. **Validate objective contract** - when the manifest opts into `objective_contract: true`, it must have a top-level `done_check`, and every runnable slice must have `proof_check` or a valid `no_check_reason`. Do not require a durable `model_route`: new plans record only capability difficulty/risk, while route selection happens immediately before dispatch from live evidence. Legacy `model_route` values may remain readable as hints only. When `cmd/kbcheck` exists, run `go run ./cmd/kbcheck manifest-contract --manifest <manifest-path>` before execution. Otherwise verify those fields and terminal slice gates directly, record `manifest-validator: unavailable`, and block on any missing/false `proof_check`, absent `done_check`, or unproved terminal status. The portable skills do not require the Go maintainer harness. If validation fails, route back to `kb-plan`/`kb-gate`; do not continue from a manifest that can self-report completion without objective proof.
5. **Validate slice contracts** - each slice plan must have `expected_files`, `verification`, `blockers`, `status`, and acceptance criteria. New slice plans should also have `test_level`, `functional_risk`, `model_tier`, and `proof_check` or `no_check_reason`. Treat legacy `tiny` or `model_route` fields as compatibility hints only; do not require them on new plans and do not freeze a provider/model in durable plan state. If core fields are missing, stop and route to `kb-plan`; do not infer a manifest from a phase list. If only `test_level` or `functional_risk` is missing on an older plan, invoke `kb-functional-test` to classify them before execution.
6. **Load the context packet** - for a non-trivial slice, read its packet before broad repo search or delegation. When `cmd/kbcheck` exists, validate JSON packets with `go run ./cmd/kbcheck context-packet --packet <packet.json>`. Otherwise verify the required fields directly and record `packet-validator: unavailable`; the portable skills do not require the Go maintainer harness. If required source, constraint, proof, search-policy, or escalation data is missing, route back to `kb-plan` instead of making a cheap worker rediscover the repo. Legacy tiny/mechanical slices may use the plan itself when it records why no packet is needed.
7. **Check status** - skip any slices already marked `done`. Resume from the first safe ready set.
8. **Check worktree** - note dirty or untracked files before executing so unrelated user changes are not staged or reverted.
9. **Read active landmines** — if `docs/context/landmines.md` exists, read only `Active Landmines` and carry any relevant failure modes into slice execution and verification. If a slice touches an `owner_surface`, treat that landmine as a hard guardrail until the slice proves the `verification` condition or explicitly leaves it active.
10. **Sync with board** — read `todo.md` and confirm its status table matches the manifest. If they diverge, the board wins — another agent may have updated it. Reconcile the manifest from the board before proceeding.
11. **Confirm once only when needed:** If the user did not explicitly ask to run/execute/work the manifest, ask: "Ready to execute N remaining slices in order. Proceed?" If the user already asked to execute, continue without this prompt.

After initial execution starts, do not ask before moving from one safe ready set
to the next.

Treat statuses as:

| Status | Action |
|--------|--------|
| `pending` | Eligible once blockers are `done` or `skipped` |
| `done` | Skip |
| `blocked` | Stop and ask whether to retry, skip, or abort |
| `human-required` | Waiting on human action; continue unrelated runnable slices if possible |
| `parked` | Intentionally out of bounds today; only a human promotes back to active |
| `skipped` | Skip but keep visible in the summary |

## Board Sync Protocol

`todo.md` is the live execution board. Update it at every status transition:

| Event | Board Update |
|-------|-------------|
| Starting a slice | Set status to 🔧 in_progress |
| Slice completes | Set status to ✅ done |
| Slice blocked | Set status to 🔒 blocked + reason in notes |
| Slice needs human action | Set status to 🛑 human-required + exact ask |
| Slice parked by human | Move to 🧊 Parked / Cold Storage with reason |
| Slice skipped | Set status to ⊘ skipped |
| All slices done | Prepend compact summary to `todo-done.md`, then remove completed feature section and routine completion logs from `todo.md` |

Active handoff files under `docs/handoffs/active/` are restart packets. Create or update one whenever work stops, blocks, or changes the next recommended action. Move completed handoffs to `docs/handoffs/done/`.

**Multi-agent rules:**
- Before claiming a slice, re-read `todo.md`. If another agent set it to 🔧, do not claim it.
- The board is the source of truth — not chat history, not the manifest. If the board says done, it's done.
- Update the board BEFORE starting work (claim) and AFTER completing work (release). This prevents two agents from working the same slice.
- Also update the manifest to stay in sync, but if they conflict, the board wins.
- Do not use root **Work Log** as a permanent archive. During execution, add notes only when they help a later agent resume: blockers, verification commands, durable memory impacts, or non-obvious decisions. Routine "slice complete" and verification-success notes belong in `todo-done.md` at feature completion, not in `todo.md`.
- Blocked is not parked. Use `🔒 blocked` for dependencies, another-agent waits, tool failures, or missing inputs. Use `🧊 Parked / Cold Storage` only for work a human intentionally deferred out of scope.

## Run-State Events

If the manifest or goal ledger points to `.kb/runs/<goal-slug>/`, append a row
to `.kb/runs/<goal-slug>/route-history.jsonl` when `kb-work` starts, completes,
blocks, or requeues a slice.

Minimum row fields:

```json
{"ts":"<ISO-8601>","route":"kb-work","confidence":0.8,"state_changed":true,"progress_key":"slice-003-done"}
```

Use `state_changed: true` or `progress_key` only after a manifest status, gate,
proof artifact, or board pointer actually changed. Before choosing another KB
lane for the same run, validate the history with:

```powershell
go run ./cmd/kbcheck run-state --history .kb/runs/<goal-slug>/route-history.jsonl
```

If the guard fails, stop routing and repair the loop by refreshing context,
re-planning the unit, or asking the smallest human question.

## Live Route Selection

Routing is a work-time decision, not a plan rewrite.

For each run/ready slice, use the executable sequence:

1. `kbrouter models discover --run-root <run-root> --current-model <id> --json`
2. `kbrouter models select --run-root <run-root> --run-id <run-id> --tier <tier> --task-family <family> --tool <tool> --context-size <n> --risk <risk> [--override use|require|ignore --alias <alias>] --json`
3. For a routed decision, call `kbrouter dispatch` with the selected primary
   alias, optional next ordered alias, and unique direct-child packet/output/
   receipt/handoff names containing the slice ID.

If discovery/select is missing, incompatible, or returns unavailable, record
`router-unavailable: <reason>`. Use the returned `degraded-current` decision or
the current model only when policy permits; `require` pauses that slice instead.

- Discover or reuse exactly one run-scoped catalog before the first routed
  slice. Reuse it while the host/configuration fingerprint is unchanged;
  refresh it only when that fingerprint changes.
- Do not ask a setup questionnaire for worker tiers. Merge only what the active
  host can actually select plus any user-local/private aliases or project
  preferences already allowed at this project.
- Treat `model_tier` as a capability floor. New worker tiers are `small`,
  `medium`, and `large`. `planner` is an orchestration/review role, not a
  cheaper worker tier. Legacy `tiny` maps to the `small` lane as a compatibility
  hint only.
- Select immediately before dispatch from live evidence in this order:
  exact tier first, another qualified same-tier route, then a higher tier, then
  the current model as an explicit degraded fallback when policy permits.
  Never move downward automatically.
- A higher or planner-grade model may execute lower-tier work when it is
  independently eligible or explicitly requested by the user.
- Security, auth, data-boundary, or process-boundary work must not be silently
  down-routed or sent to a less trusted destination.
- `use <model>` prefers that route first with safe upward fallback.
- `require <model>` pauses only that slice if the exact route is unavailable.
- `prefer local|hosted` is a preference inside trust and risk constraints.
- `ignore model routing` uses the current model and ordinary proof gates only.
- A generic or named subagent spawn without an exact selector cannot claim the
  requested model. Record actual route/model/session only from dispatcher or
  host evidence; otherwise provenance is `unknown` or `unavailable`.
- Existing correct work with missing or mismatched provenance is not blocked or
  redone for telemetry. Investigate once, record what is knowable, then let
  independent proof govern completion.

Show a compact preview only when routed slices exist. Group by selected live
route and include the bounded fallback path, for example
`Medium — Terra -> Sonnet -> current(degraded): 004, 007`.
This preview is dispatch intent for the current ready set only. If nothing is
routed, say nothing about models.

## Ready-Set Ordering

Execute by repeatedly pulling the safe ready set from the dependency DAG:

1. Build a map of `slice_id -> slice`.
2. For each pending slice, check all `blockers`.
3. The candidate ready set is every pending slice whose blockers are complete.
4. Exclude serial-only slices from co-dispatch when other ready slices exist:
   `can_continue_other_slices: false`, HITL-critical gates, destructive
   approvals, browser/e2e contention without isolated sessions, and any slice
   with an active write lease collision.
5. Dispatch the remaining safe ready set in isolated contexts when available.
   If no isolation is available, run the same ready set one mutating slice at a
   time while preserving the ready-set order.
6. If pending slices remain but none are runnable, mark the manifest blocked and
   report the dependency problem.

## Execution Loop

For each slice in dependency order:

### Continuous Execution Rule

Ready sets should run continuously once execution has started.

Do **not** ask "Proceed to execute slice-N?" between slices. Move to the next
safe ready set automatically after:

- slice status is updated;
- board and manifest are synced;
- required deterministic checks pass;
- QA/repair gates pass or are not applicable.

Pause only when a real gate requires it:

- HITL decision or missing value that cannot be generated safely;
- blocked/human-required/parked slice with no unrelated runnable work;
- destructive command approval;
- out-of-scope file edit or diff-scope failure;
- QA/repair exhaustion or stuck loop;
- dependency deadlock;
- observed write overlap that cannot be serialized or requeued safely;
- user explicitly asked to pause or stop.

### Step 1: Check HITL Flag

If `hitl: true`:

- Present the slice title, description, and the specific question/decision needed.
- Classify the HITL item before stopping:
  - `critical-path` — later slices depend on this decision/access/input.
  - `parallel-blocker` — this slice is blocked, but unrelated slices can continue.
  - `final-validation` — human judgment is useful before release, but not needed for development.
  - `agent-runnable-with-inputs` — human only needs to provide values; the agent can run the check.
- Stop only the dependent path. If unrelated slices are runnable, mark this slice `blocked` or `human-required`, update `todo.md` and the manifest, then continue those slices.
- When marking a slice `human-required`, `parked`, or `blocked`, persist: `owner`, `blocked_reason`, `resume_when`, `next_agent_action`, `human_action`, `can_continue_other_slices`, and `parked_at`.
- Record the user's decision in the slice plan.
- Update manifest status to `done` for this slice only if the decision completes the slice.
- Continue to the next runnable slice.

Missing test inputs are not a reason to ask the user to manually test. If `test_inputs` are missing:

- Ask for the specific missing value.
- Use safe generated or fixture values when acceptable.
- If the input blocks only this slice, mark this slice `human-required` or `blocked` and continue unrelated runnable slices.
- Resume the slice after the value is available and run the verification yourself.

### Step 2: Deepen If Thin

If the slice plan has fewer than 3 acceptance criteria or no test scenarios:

- Run a lightweight deepening pass on this single slice.
- Add concrete test scenarios and likely file paths.
- Keep the pass bounded; do not re-plan the whole feature.

### Step 2.5: Test-Level Classification

Before editing, ensure the slice has a recorded test obligation:

- `test_level`: `none`, `unit`, `integration`, `functional-api`, `functional-cli`, `functional-browser`, or `full`
- `functional_risk`: `none`, `narrow`, `broad`, or `full`
- `model_tier`: `small`, `medium`, or `large` (`tiny` remains readable only as a legacy hint that maps to `small`)

If `test_level` or `functional_risk` is missing, stale, or contradicted by the
acceptance criteria or `expected_files`, invoke `kb-functional-test` with the
slice plan. Record the result in the slice frontmatter or notes before
implementation. If `model_tier` is missing in a manifest that opted into
`model_tier_contract`, route back to `kb-plan` to repair the slice contract.

Use smaller worker routes for this classification only when the active host can
actually select them. Keep the task bounded: classify the slice, audit existing
tests for mocked theater, and suggest the narrowest deterministic proof.
Escalate to the main model for complex architecture/auth/security/flaky async
decisions or repeated test failures.

Model-tier boundaries:

| Tier | May own | Escalate when |
|---|---|---|
| `small` | narrow mechanical edits and straightforward tests with clear acceptance criteria | cross-boundary behavior, UI/API workflows, security/auth/data decisions |
| `medium` | ordinary vertical slices and focused integration work | architecture/security migrations, multi-slice replanning, unclear product intent |
| `large` | hard decomposition, architecture, broad debugging, final synthesis/review | proof is impossible or human-only input is required |

The tier does not change the proof bar. It only says what kind of agent can do
the first pass before `kb-work` verifies the result.

Record actual runtime/model, turns, input/output/cache tokens, proof result, and
packet sufficiency when the host exposes them. Missing host telemetry is
`unavailable`, not zero. Raw values remain authoritative; weighted cost scores
must be versioned and reported beside proof outcomes.

Do not use `unit` just because it is cheaper. Use `unit` only when unit-level proof can fail for the user-facing bug or behavior. If a unit test could pass while the workflow is broken, require integration or functional proof.

Hard gate: when `kb-functional-test` auto-classifies a test level, the agent must not downgrade it. If `expected_files` includes `.tsx`, `.jsx`, `.vue`, or `.svelte`, or if non-UI files change behavior primarily reached through the rendered app UI, the slice is `test_level: functional-browser`.

When `test_level` is `functional-browser`, these steps are mandatory:

1. Start or connect to the running app.
2. Use Playwright to navigate to the actual feature route/screen in the rendered UI. Use CDP or the repo/platform authenticated browser transport only when Playwright cannot access an authenticated/corporate route.
3. Exercise the happy path with real clicks, keyboard input, form input, navigation, or other visible controls.
4. Capture screenshots of key pass/fail states and assert observable rendered outcomes after the action.
5. Clean up artifacts created during testing: test data, screenshots/traces when no longer needed, temp files, and browser state per repo QA cleanup rules.

Backend/API/unit checks may supplement this proof, but they cannot replace it. This gate cannot be skipped, overridden, or deferred.

### Step 2.9: Regression Snapshot Gate

Before starting a new slice, invoke `kb-regression-snapshot verify` before Scope Lock and before editing implementation files.

- Verify all previous snapshots under `.kb/snapshots/`.
- If any previous snapshot fails, STOP before new slice execution.
- Mark the current slice `🔒 blocked` with the failing snapshot path, target, expected vs observed result, and artifact/log path.
- Do not continue to implementation, QA, or the next slice until the regression is resolved, parked by the human, or explicitly skipped.

This gate catches entropy between slices. It cannot be skipped, overridden, or deferred.

### Step 3.0: Scope Forecast and Ledger

Before executing the slice, load the declared scope forecast and keep a live ledger of actual files touched. `expected_files` guides the first pass; it is not a literal allowlist.

1. **Read `expected_files`** from the slice plan's frontmatter.
2. **If `expected_files` is empty or missing**, route back to `kb-plan` to repair the slice plan before execution. Do not execute from a phase list or raw task with no file forecast.
3. **Expand the forecast with convention-matched test files.** For each entry in `expected_files`, automatically include its corresponding test file based on project naming conventions:

   | Source file | Auto-allowed test file(s) |
   |-------------|--------------------------|
   | `src/foo.py` | `tests/test_foo.py`, `test/test_foo.py` |
   | `src/Foo.tsx` | `src/Foo.test.tsx`, `src/__tests__/Foo.tsx` |
   | `lib/foo.rb` | `spec/foo_spec.rb`, `test/foo_test.rb` |
   | `pkg/foo.go` | `pkg/foo_test.go` |

   Test files that do not correspond to any `expected_files` entry may still be valid when current code or acceptance criteria require them; record them as discovered files.

4. **Before opening any file for writing**, classify it against the slice intent:

   | Finding | Action |
   |---------|--------|
   | File is listed in `expected_files` or is a convention-matched test | Proceed with the edit. |
   | File is not listed, but current code shows it is directly required for this slice's acceptance criteria | Proceed, and add a manifest note: `scope-discovery: <file> - <why required>`. |
   | File is generated by the repo's normal tooling, formatter, snapshot, lockfile, or test convention | Proceed, and add a manifest note: `scope-discovery: <file> - generated/tooling`. |
   | File would change product scope, architecture direction, dependencies, migrations, auth/security boundaries, destructive behavior, or another slice's promised behavior | STOP for HITL or route back to `kb-plan` to amend the manifest before editing. |
   | File is opportunistic cleanup or unrelated improvement | Do not edit. Park it in `todo.md` or a follow-up note. |

5. **Log the forecast** in the manifest notes: `scope-forecast: loaded N expected files + M convention-matched tests`.

This gate pairs with Step 3.6 (Diff-Scope Verification). The point is traceability, not pretending the planner knew every file in advance.

### Step 3: Execute

Use a fresh sub-agent when the platform supports delegated execution and the user has permitted it. Otherwise execute the slice locally while keeping the scope limited to this slice.

Immediately before dispatch, choose the route for this ready slice from the
live run catalog. Try the exact capability floor first, then another qualified
same-tier route, then a higher tier. Use the current model only as an explicit
degraded fallback when policy permits. Never move downward automatically.

Quoting sanity rule: when shell commands, file operations, or test assertions involve nested quotes, escaped quotes, or more than one quoting context, write the content to a temp file and execute/read that file instead of constructing the command inline.

- If you are escaping an escape, you are doing it wrong. Write to a file, execute the file.
- For multi-line JSON, SQL, HTML, scripts, or config blocks, use heredoc syntax or a temp file rather than inline quoting.
- Do not build JSON strings inside shell commands inside assertion code. Write JSON to a temp file and read it.
- Do not construct CSS selectors through mixed-quote string concatenation. Use template literals or parameterized locator helpers.

Use `references/execution-prompt.md` as the per-slice execution prompt/checklist. Load it only when starting a slice. When that slice runs Go inside a workspace sandbox, also load `references/go-sandbox.md`; its environment applies to Go shell invocations, never the agent launcher.

### Step 3.1: Protected Oracle Gate

If the slice plan or manifest declares `protected_oracles`, enforce the
anti-cheat contract before implementation changes:

1. Identify every oracle file: tests, fixtures, scorers, snapshots, schemas, or
   contracts that define expected behavior.
2. If the oracle is new or intentionally changed for this slice, create/update it
   before implementation and prove RED when practical.
3. Record the oracle SHA256 in the slice plan or manifest after the oracle is
   accepted.
4. After the SHA is recorded, do not edit that oracle unless the plan is
   explicitly amended with a new reason and a new SHA.
5. Before marking the slice done, recompute oracle hashes. Any unexpected hash
   change blocks the slice.

If `protected_oracles` is empty, continue with the declared verification mode.
Do not invent a protected oracle when expected behavior cannot be known before
implementation.

### Step 3.5: System-Wide Test Check

Before marking a slice done, pause and ask these questions — vertical slices cut through all layers, so side-effects matter:

| Question | What to do |
|----------|------------|
| **What fires when this runs?** Callbacks, middleware, observers, event handlers — trace two levels out from your change. | Read the actual code for callbacks on models you touch, middleware in the request chain, `after_*` hooks. |
| **Do my tests exercise the real chain?** If every dependency is mocked, the test proves logic in isolation — not interaction. | Write at least one integration test that uses real objects through the full callback/middleware chain. |
| **Can failure leave orphaned state?** If your code persists state before calling an external service, what happens when the service fails? | Trace the failure path. If state is created before the risky call, test that failure cleans up or that retry is idempotent. |
| **What other interfaces expose this?** Mixins, DSLs, alternative entry points. | Grep for the method/behavior in related classes. If parity is needed, add it now. |

**When to skip:** Leaf-node changes with no callbacks, no state persistence, no parallel interfaces. Purely additive changes (new helper, new partial) take 10 seconds to confirm "nothing fires, skip."

### Step 3.6: Diff-Scope Verification

After a slice completes, verify that the files actually changed are explainable by the slice's acceptance criteria. The agent does not self-report; the actual git diff is checked and recorded.

1. **Get the actual diff:**

   ```bash
   git diff --name-only $(git merge-base HEAD main)..HEAD
   ```

   This produces the list of files modified by this slice relative to the branch baseline.

2. **Load the forecast scope** from the slice plan's `expected_files` frontmatter field plus any `scope-discovery:` notes recorded during execution. Also load any `protected_oracles` and their recorded hashes.

3. **Compare and enforce:**

   Apply the same convention-matched test file expansion as Step 3.0.

   | Finding | Action |
   |---------|--------|
   | Changed file is forecast, convention-matched, generated/tooling output, or recorded `scope-discovery` | Proceed. |
   | Changed file is unforecast but directly required by the acceptance criteria and was not noticed before editing | Record `scope-discovery: <file> - <why required>` before proceeding. |
   | Changed file expands product scope, architecture direction, dependencies, migrations, auth/security boundaries, destructive behavior, or another slice's promised behavior | STOP. Amend the manifest through `kb-plan` or get HITL before proceeding. |
   | Changed file is unrelated cleanup or opportunistic improvement | Revert or park as follow-up before proceeding. |
   | Forecast files were not changed | Treat as a completeness signal, not a failure. If the slice still satisfies acceptance criteria, record `scope-forecast-unused: <file> - <why not needed>`. |

   If a changed file is a protected oracle and its SHA changed after protection,
   STOP unless the manifest or slice plan explicitly records an oracle update
   reason and the new SHA.

4. **Log results** in the KB manifest under the slice's `notes` field:

   ```text
   notes: "scope-check: forecast=5 changed=7 discovered=2 unexplained=0"
   ```

5. **If the slice plan has no `expected_files` field**, route back to `kb-plan` to repair the plan before continuing. Do not execute a slice with no forecast at all.

This gate is mandatory. It cannot be skipped, overridden, or deferred, but it records justified discovery instead of blocking every unforecast file.

### Step 3.7: Destructive Command Guard

Before executing any shell command during a slice, check it against this blocklist:

| Blocked Pattern | Why |
|-----------------|-----|
| `rm -rf`, `rm` with recursive/force flags | Irreversible file deletion |
| `git push --force` / `git push -f` | Rewrites remote history |
| `git reset --hard` | Discards uncommitted work |
| `DROP TABLE` / `DROP DATABASE` / `TRUNCATE` | Irreversible data loss |
| `git clean -fd` | Deletes untracked files permanently |
| Bulk delete operations on files or data | Mass irreversible removal |
| Overwriting production config files | Environment-breaking changes |

**When a blocked command is detected:**

1. **STOP.** Do not execute.
2. Show the user the exact command and explain why it's blocked.
3. Wait for explicit HITL approval before proceeding.
4. If running in autonomous mode (no HITL available), skip the command and log in the manifest notes: `destructive-guard: blocked <command> — no HITL available`

This is enforcement, not a warning. The agent MUST NOT execute destructive commands without explicit human confirmation. This gate cannot be skipped, overridden, or deferred.

### Step 3.8: KB QA (all slices)

Invoke `kb-qa` with the current slice context. QA runs:

- **Lint check** on forecast and discovered files for the slice
- **Browser verification** against acceptance criteria for frontend slices and any backend/API/state slice whose changed behavior is reachable through the UI

If any check fails, `kb-qa` invokes `kb-repair` for surgical fixes (progress-based, 5-iteration cap). If repair exhausts or gets stuck, STOP — do not proceed to the next slice.

Truly backend-only slices skip browser checks but still run lint. Do not classify a slice as backend-only when the behavior being changed is primarily proven by using the app UI.

This gate is mandatory. It cannot be skipped or deferred.

### Step 3.9: Figma Design Sync (UI slices only)

If the slice involves UI changes and Figma designs exist:

1. Implement components following design specs
2. Use the **figma-design-sync** agent iteratively to compare
3. Fix visual differences identified
4. Repeat until implementation matches design

Skip this step entirely for non-UI slices.

### Step 4: Verify and Update

After the slice completes:

1. **Check result**
   - If yes: update manifest `status: done` for this slice and update the body table.
   - If no and repair/progress is still possible: run `kb-repair` or a bounded fix loop, then retry verification.
   - If no progress remains: update manifest `status: blocked` or `parked`, add failure notes and resume criteria, then continue unrelated runnable slices.

   Before setting `status: done`, write or update a gate record
   `slice-<slice_id>-to-done`. This gate must prove: implementation finished,
   scope check passed, protected oracles were preserved or explicitly amended,
   deterministic checks ran, functional/browser checks ran when required,
   regression snapshot captured, memory impact was classified, and the slice's
   `proof_check` passed or its `no_check_reason` was accepted by the manifest
   contract. If any proof is missing, leave the slice `blocked` and set
   `allowed_next_action` to the missing proof step.

   For manifests with `objective_contract: true`, run
   `go run ./cmd/kbcheck manifest-contract --manifest <manifest-path>` after
   adding the slice gate and before changing the slice status to `done`. A
   failing manifest contract blocks completion even when local tests pass.

2. **Sync board** — update `todo.md` status for this slice (done or blocked). Append validation note.

3. **Run verification**
   - Invoke `kb-check` for deterministic verification.
   - Prefer existing scripts, lint, typecheck, tests, browser checks, builds, and CI-equivalent commands over LLM inspection.
   - If a full suite is too expensive or unavailable, run the narrowest deterministic check that proves the slice and record why.
   - Invoke `kb-functional-test` whenever `test_level` is `integration`, `functional-api`, `functional-cli`, `functional-browser`, or `full`, or when user-visible/cross-boundary changes appear despite a lower test level.
   - If the slice fixed a known failure with a runnable sensor, record RED-before-GREEN proof with `go run ./cmd/kbcheck accept --check <check.json> --trace .kb/trace.jsonl`.
   - For UI-reachable changes, record UI proof: route/screen exercised, interaction performed, assertion made, browser transport used, and screenshot path when applicable. Do not mark the slice done with backend-only proof if a UI path exists.
   - After Step 3.8 QA passes, invoke `kb-regression-snapshot capture <slice-id>` with a compact spec for what changed. Store `.kb/snapshots/<slice-id>.json`.
   - Record `test-level: <value>; functional-risk: <value>; proof: <command/artifact>; snapshot: <path/result>` in the manifest notes.
   - Record the slice's `proof_check` command/artifact/result or accepted
     `no_check_reason` in the manifest notes.

4. **Assess memory impact**
   - Classify the slice as `memory-impact: none`, `operational`, or `durable`.
   - `none`: cosmetic, copy, formatting, lint-only, or isolated tests with no behavior/architecture change.
   - `operational`: active state, blockers, verification commands, or handoff instructions changed. Update `todo.md` or the active handoff.
   - `durable`: user-visible behavior, API/data/storage/auth/routing/streaming/tool/action/job/integration behavior, run/test commands, subsystem boundaries, sharp edges, or rejected approaches changed.
   - For durable changes, add a manifest note: `memory-impact: durable; areas=<areas>; docs=<candidate docs>; refresh=pending`.
   - If the affected doc is obvious and small, update it now. Otherwise leave `refresh=pending` for Step 5.

5. **Optional commit**
   - If the user asked for commits, stage only the manifest file for status updates and commit it separately.

6. Continue to the next runnable slice.

### Step 5: Completion

When all slices are `done` or intentionally `skipped`:

1. Update manifest `status: completed`.
2. Run final verification.
3. Run `kb-gate` if verification, QA, repair, or functional-test checks surfaced P0/P1/P2/P3/P4 issues. P0/P1 block completion while unresolved, but safe/actionable blockers should be rectified before asking the user. P2/P3/P4 do not block by severity alone.
4. Write `work-to-complete` in the manifest `gate_ledger`. Required proof: every non-skipped slice has a passing `slice-<id>-to-done` gate, skipped slices have explicit reason, final verification command/result is recorded, no unresolved P0/P1 exists, board/manifest are synced, and `scope-verified-files` is populated. Run `kb-gate/scripts/check_gate_ledger.py <manifest-path> --gate work-to-complete --allowed-next "kb-complete <manifest-path>"`. If the gate is not passed or the checker fails, do not invoke `kb-complete`.
5. **Refresh project memory** — if any slice has `memory-impact: durable` or `refresh=pending`, run `kb-map refresh` before archiving. Update affected architecture, operation, decision, research, `todo.md`, and handoff pointers. Add manifest note: `kb-map-refresh: done` or `kb-map-refresh: skipped - <reason>`.
6. **Archive to board** — move the feature summary from `todo.md` to `todo-done.md`. Prepend at the top of the archive file with a completion date header.
7. **Prune active board** — remove the completed feature section from `todo.md`. Also remove routine work-log entries for the completed feature from `todo.md`; keep only still-active rows, `🔒 blocked` rows, `🛑 human-required` rows, the `🧊 Parked / Cold Storage` section, or handoff-pointer items.
8. Report summary:

```text
KB <name> — all slices complete.
- N slices executed
- S slices skipped
- M tests added
- K files changed
Verification: <command/result>

Next: kb-complete runs automatically for review, documentation, and learning.
```

9. **Persist scope context** — collect the forecast and discovered file lists from each slice's `notes` field (the `scope-check:` and `scope-discovery:` entries from Step 3.6). Write the combined actual changed file list to the manifest frontmatter as `scope-verified-files` so `kb-complete` can pass it to kb-review without re-deriving.

**Post-work steps (kb-review, compound, learn, evolve, cleanup) are handled by `kb-complete`.** After all slices are `done` or intentionally `skipped`, invoke `kb-complete <manifest-path>` automatically unless the user explicitly asked to stop after work execution. The `kb-work` run is not complete until `kb-complete` reaches its Done section or records a real blocker.

## Failure Handling

| Situation | Action |
|-----------|--------|
| Slice execution fails with progress possible | Run `kb-repair` or a bounded fix loop, then retry verification |
| Slice execution fails with no progress | Mark only that slice blocked/parked, write resume packet, continue unrelated runnable slices |
| Test suite fails after a slice | Run `kb-repair`; if stuck, mark affected slice blocked/parked and continue unrelated runnable slices |
| HITL critical-path pause | Present context, wait for user, record decision |
| HITL not on critical path | Park the slice and continue unrelated runnable slices |
| User says "abort" | Mark remaining slices as `pending`, stop |
| User says "skip" | Mark slice `skipped`, continue to next runnable slice |

## Resume Support

KB work is resumable:

- Manifest tracks status per slice.
- Re-running `kb-work` on the same manifest picks up where it left off.
- Already done or skipped slices are not rerun.

## Success Criteria

- No slice runs before its blockers are complete.
- Manifest frontmatter and body table reflect actual slice status.
- Each completed slice has verification evidence recorded in the final response or failure notes.
- No unrelated files are staged, committed, reverted, or overwritten.

## Integration

- **Input from:** `kb-plan`
- **Deepening:** Use `kb-research` only when a slice has a material unresolved uncertainty before execution.
- **Execution engine:** Fresh sub-agents when available, local execution otherwise
- **Verification:** Preserves and reruns protected test oracles for `tdd` slices; load standalone `tdd` only for explicit test-first coaching.
- **Protected oracles:** When declared by `kb-plan`, freezes behavior tests, fixtures, scorers, snapshots, or contracts before implementation so the target cannot move silently
- **Deterministic checks:** Invokes `kb-check` before a slice is marked done
- **Functional checks:** Invokes `kb-functional-test` for user-visible and cross-boundary behavior
- **Post-completion:** Automatically invoke `kb-complete` after all slices are done or intentionally skipped
