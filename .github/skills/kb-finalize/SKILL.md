---
name: kb-finalize
description: "Internal post-work quality and learning phase. Runs kb-review -> resolution gate -> follow-up resolution -> proof/demo evidence -> compound -> learn -> evolve -> memory refresh/compact -> cleanup after kb-work finishes all slices. Normally invoked by kb-work or kb-complete; not the primary user-facing completion command."
argument-hint: "[path to KB manifest, or blank to find latest]"
---

# KB Finalize - Post-Work Quality & Learning Pipeline

After `kb-work` finishes executing all slices, this skill runs the quality review, follow-up resolution, proof/demo evidence, knowledge capture, memory-health, and cleanup steps. Separated from kb-work so the user can choose when to run it — `klfg` prompts automatically, standalone users invoke it manually.

## Terminal Completion Contract

When invoked by `kb-work`, this skill is the terminal half of the execution
loop. Continue until Step 5 reports Done, or until a real blocker is recorded
with exact resume criteria.

Do not stop at weaker milestones:

- deterministic checks passed before review;
- reviewers returned findings;
- P0/P1 findings were fixed but proof has not rerun;
- proof passed but memory, learn/evolve cadence, or cleanup is unfinished;
- a summary was written without the final `KB <name> complete` report.

If a repo has a project-specific `done.md` contract such as "can't stop til its
done", interpret "done" as this skill's Step 5 terminal report. KB state still
lives in `todo.md`, `todo-done.md`, manifests, and handoffs unless the repo
already opted into a `done.md` workflow.

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Scan `docs/plans/` for the most recent `*-kb-*-manifest.md` file with `status: completed`. If none is found but an active KB manifest exists with pending or in-progress slices, stop and invoke `kb-work <manifest-path>` first. If no manifest exists, do not run completion from a raw diff; route back to `kb-plan`/`kb-work` so completion has slice scope and verification evidence.

**If input is a path:** Read the manifest at that path.

## Pre-Flight

1. **Read the manifest** — confirm `status: completed` (all slices done/skipped). If slices are still `pending` or `in_progress`, stop: "This manifest has unfinished slices. Run `kb-work` first."
2. **Validate gate ledger** — the manifest must contain `gate_ledger` with `work-to-complete` set to `passed`, and every completed slice must have a passing `slice-<id>-to-done` gate. New manifests use `allowed_next_action: kb-finalize <manifest-path>`; accept legacy `kb-complete <manifest-path>` while old plans remain active. Run the checker against the recorded allowed action. If missing, blocked, or the checker fails, stop and invoke `kb-work <manifest-path>` or `kb-gate` to repair the ledger. Do not run finalization from a manifest that merely says `status: completed`.
3. **Validate objective contract** — if the manifest has `objective_contract: true`, run `go run ./cmd/kbcheck manifest-contract --manifest <manifest-path>`. Confirm the top-level `done_check` is present and each completed slice has `proof_check` evidence or an accepted `no_check_reason` recorded in slice notes/gates. If missing, stop and return to `kb-work` for proof repair.
4. **Collect scope context** — scan each slice's `notes` field for `scope-check:` and `scope-discovery:` entries. Build the combined list of actually changed, scope-verified files across all slices. This becomes the review scope.
5. **Collect memory impact** — scan slice notes for `memory-impact:` and `kb-map-refresh:` entries.
6. **Review routing evidence honestly** — collect any routing receipts or host evidence linked from the run state, but treat them as telemetry rather than proof of correctness. Missing, unknown, or mismatched provenance does not invalidate already-proven work.
7. **Identify the branch baseline** — run `git merge-base HEAD main` to establish the diff range.
8. **Run final snapshot sweep** — invoke `kb-regression-snapshot verify` for all snapshots under `.kb/snapshots/`. If any snapshot fails, STOP before review; later work regressed earlier passing behavior.

If the manifest has no scope-check notes (older format), fall back to `git diff --name-only $(git merge-base HEAD main)..HEAD` for the file list.

## Step 1: Code Review

Before code review, run `kb-check` against the completed manifest scope. If deterministic checks fail, route to `kb-repair` or `kb-fix` before `kb-review`. LLM review does not replace executable verification.

Do not choose, rerun, or require a different model solely to improve routing telemetry. `kb-finalize` reviews receipts and notes useful observations for future selection, but proof and review gates remain model-independent.

If the manifest contains slices with `test_level` of `integration`, `functional-api`, `functional-cli`, `functional-browser`, or `full`, run `kb-functional-test` before `kb-review` to confirm the functional coverage is real and not mock-only. Also run it when the diff shows user-visible, API/CLI, persistence, auth, streaming, or integration changes without an adequate recorded test level.

If the final diff includes `.tsx`, `.jsx`, `.vue`, or `.svelte` files, or any changed behavior primarily reached through the rendered UI, require `functional-browser` evidence before completion. Backend/API/unit-only proof is insufficient for those changes.

**Invoke the `kb-review` skill** — structured code review on the feature diff.

`kb-review` is a skill/orchestrator, not an Agent tool type. Do not call the Agent tool with `agent_type: kb-review`. Load/run the `kb-review` skill, and let that skill spawn valid reviewer agent types such as `code-review`, `correctness-reviewer`, `security-reviewer`, or `adversarial-reviewer`.

This is mandatory. Do not skip, defer, or make it optional. Record the review
mode from `kb-review`: `review-mode: multi-agent` when reviewer agents actually
ran, or `review-mode: local-fallback` when the runtime could not or did not
authorize reviewer subagents. Do not claim multi-agent review happened when
fallback was used.

- **Pass scope from prior gates:** use the collected scope-verified actual file list from Pre-Flight. Pass this as the scoped file list so kb-review skips its own scope discovery (Stage 1). The scope gates already explained planned and discovered files — no need to re-derive from git diff unless the manifest lacks this data.
- Pass context: the full `git diff` of the feature branch against baseline, scoped to the verified file list
- Capture the output: each finding has a severity (P0/P1/P2/P3) and confidence score
- Capture review mode and any fallback residual risk.
- Store findings for the resolution gate (Step 2)
- **Note:** if scope-verified files are unavailable (older manifest, standalone run), let kb-review do its own scope discovery.

## Step 2: Resolution Gate

Review findings from `kb-review` determine whether completion is allowed:

| Severity | Action |
|----------|--------|
| P0 (critical) | STOP. Fix before proceeding. Re-run affected tests after fix. |
| P1 (important) | STOP. Fix before proceeding. |
| P2 (suggestion) | Log in manifest `notes`. Do not block. |
| P3 (nit) | Log in manifest `notes`. Do not block. |

This gate is mandatory. The agent MUST NOT proceed to Step 3 while unresolved P0/P1 findings exist.

For any P2/P3 findings, invoke `kb-gate` with the rectify prompt. Do not silently leave fixable P2/P3 issues when the user would prefer a clean finish.

After resolving all P0/P1s, update the manifest notes with a summary:
`review: P0=0 P1=2(resolved) P2=3(logged) P3=1(logged)`

**Feed learnings to the observation log:**

For each resolved P0/P1 finding, append one line to `.kb/observations.jsonl`:

```json
{"ts":"<ISO-8601>","hook":"kb-review","tool":"kb-finalize","args":{"finding_type":"<category>","severity":"P0|P1","resolution":"<what was fixed>"},"cwd":"<repo-root>","result":"resolved"}
```

This connects the review → learn pipeline. Only P0/P1 findings are worth learning from — P2/P3 are style preferences, not systemic patterns.

Create `.kb/observations.jsonl` if it doesn't exist. Append, never overwrite.

## Step 2.5: Follow-Up Resolution Gate

Mirror the useful part of the original LFG finish pattern: do not leave known, fixable follow-up work unresolved just because the main implementation passed.

1. Collect unresolved review findings, TODO files, checklist items, and manifest
   notes produced by `kb-review`, `kb-gate`, `kb-work`, `kb-qa`, or
   `kb-functional-test`.
2. Resolve all safe/actionable P0/P1 findings before continuing.
3. For P2/P3 and todo-style follow-ups, run the smallest suitable resolver:
   `kb-gate` rectify prompt, `todo-triage`, `todo-create`, or a repo-local
   todo/PR-comment resolver if one is installed.
4. Do not parallelize follow-ups that touch the same files or depend on the same
   decision. Parallel resolution is allowed only when file scopes are disjoint.
5. Record: `follow-up-resolution: resolved N, logged M, blocked K`.

Blocked/human-required items stay visible in `todo.md` or the manifest with evidence paths. They must not disappear into chat history.

## Step 2.6: Proof and Demo Evidence Gate

Run a final evidence pass after review fixes, because review changes can alter
behavior.

1. Re-run `kb-check` or the narrow deterministic commands affected by review
   fixes.
2. If user-visible, API/CLI, browser, persistence, auth, streaming, or
   integration behavior changed, run `kb-functional-test` again on the final
   diff. For browser/UI work, verify through the rendered UI with Playwright
   where viable, or the repo/platform authenticated browser transport when
   Playwright cannot access the route.
3. If the change is visual, workflow-heavy, reviewer-facing, or the user/PR
   expects a demo, capture demo evidence. Use a repo-local `feature-video` or
   equivalent demo skill if installed; otherwise capture screenshots, logs, or a
   concise manual demo checklist and add an alert with the missing tool.
4. Store proof in the manifest notes: commands run, routes/screens/workflows
   checked, artifacts created, and any skipped proof with reason.

Every slice must have machine-verifiable proof recorded in the manifest before completion can continue.

Acceptable proof formats:

- test file path + exit code + timestamp;
- Playwright/Cypress/CDP trace path or browser assertion artifact;
- API response log path with status/schema assertion result;
- CLI output log path with command and exit code;
- `go run ./cmd/kbcheck accept --check <check.json> --trace .kb/trace.jsonl` result for repaired failures with a RED-before-GREEN trace;
- regression snapshot result from `.kb/snapshots/<slice-id>.json` with all previous snapshots passing.

Not acceptable:

- "I checked and it works";
- "Page loaded successfully";
- "Tests pass" without the actual command or test file, exit code, and timestamp;
- latest passing check with no recorded prior RED when the work claims to fix a known failure;
- screenshots or prose-only notes without an executable assertion/result;
- any model-only visual inspection.

If any slice has only prose proof, this gate fails. Return to `kb-work`,
`kb-check`, `kb-qa`, `kb-functional-test`, or `kb-regression-snapshot` to produce
executable evidence before proceeding.

For objective-contract manifests, summarize the top-level `done_check` and each slice's `proof_check` result in the manifest notes. If the final objective check cannot run yet, record the blocker and do not declare `KB <name> complete`.

If routing evidence is partial, record the strongest honest attribution such as `exact`, `explained-external`, `unknown`, or `unavailable`, then continue with ordinary proof. Never rerun already-correct work only to upgrade provenance.

## Step 3: Compound & Learn

After the resolution gate passes, document what this feature taught the system:

0. **Classify steering feedback** before `/learn` runs.
   - Sources: resolved P0/P1 review findings, manifest notes named
     `steering-feedback:`, `/iterate` or PR feedback summaries, goal-ledger
     feedback, and maintainer comments copied into the completion artifact.
   - Classify each item as exactly one primary route:
     `current-only`, `steering-memory`, `observation`, `landmine-candidate`, or
     `instinct-evidence`.
   - `current-only`: record in manifest notes only; do not update durable memory.
   - `steering-memory`: update the steering memory path named by the goal or
     manifest, usually the goal ledger's `Live Steering` section or
     `docs/context/operations/steering/<slug>.md`. Keep entries concise:
     durable scope constraints, known false positives, reviewer preferences, or
     selection guidance. Do not append raw transcripts or one-off PR details.
   - `observation`: append one JSONL entry to `.kb/observations.jsonl`.
     Do not duplicate the resolved P0/P1 entries already written by Step 2;
     reference those existing entries when they are the evidence source.
   - `landmine-candidate`: apply `/learn` landmine criteria; record only with
     owner surface, concrete evidence, fix condition, and verification.
   - `instinct-evidence`: leave the evidence visible to `/learn`; do not
     promote it directly to a skill.
   - If no steering memory path is named, do not create one automatically. Add a
     manifest note: `steering-feedback: no durable steering memory path`.
   - Record the result in manifest notes:
     `steering-feedback: current=<N> memory=<N> observations=<N> landmine-candidates=<N> instinct-evidence=<N>`.
1. **Invoke `ce-compound`** with context: a one-sentence summary of what was built and any surprising patterns discovered during implementation.
2. ce-compound writes to `docs/solutions/` with YAML frontmatter — let it run without modification.
3. If the implementation was pure boilerplate (no novel patterns, no gotchas, no decisions worth preserving), skip with a manifest note: `compound: skipped — standard implementation, no novel patterns`
4. Per-slice micro-learnings from slice notes feed into the compound context. Reference them when invoking ce-compound.
5. **Invoke `/learn`** — Extract instincts from this session's work.
   - Run after compound completes (observations from Step 2 are now available)
   - `/learn` reads: observations.jsonl, recent git history, docs/solutions/,
     existing instincts, and any steering-memory updates classified above
   - Record result in manifest notes: `learn: N new instincts, M updated` or `learn: no new patterns`
   - This is automatic — do not ask the user whether to run it
   - If the work exposed a repo-specific landmine, record it only with owner
     surface, concrete evidence, severity, fix condition, and verification.
     Reject generic "remember to test/read code" lessons.
   - If the work fixed an active landmine, move it to resolved/archive only
     after the verification command passes.
6. **Check evolution cadence:**
   - Read `docs/context/kb/kb-completions.txt` (create with `0` if missing)
   - Increment by 1
   - Write the new value back
   - If the new value is divisible by 5:
     - Invoke `/evolve` to check for promotable instincts
     - Log result in manifest notes: `evolve: promoted N instincts` or `evolve: no candidates ready`
   - If not divisible by 5: skip silently
   - Commit the counter file with the manifest update

## Step 3.5: Project Memory Refresh Gate

Before cleanup or final "complete", make sure a fresh session can resume without a lesson from the user.

Run `kb-map refresh` when any of these are true:

- The manifest contains `memory-impact: durable`.
- `kb-work` left any `refresh=pending` note.
- Review fixes changed behavior, architecture, run/test commands, integrations, or known sharp edges.
- `docs/context/PROJECT.md` points to stale subsystem docs after the feature diff.

Skip with a manifest note only when changes are clearly cosmetic, copy-only, formatting-only, lint-only, or isolated tests with no durable behavior change:

```text
kb-map-refresh: skipped - cosmetic/no durable architecture change
```

When refresh runs, update affected docs only:

- `docs/context/PROJECT.md` for route-map, command, or subsystem index changes.
- `docs/context/architecture/*` for durable subsystem behavior.
- `docs/context/operations/*` for run/test/deploy/QA changes.
- `docs/context/landmines.md` for active/resolved landmine changes.
- `docs/context/research/*` for reusable research outcomes.
- `docs/context/decisions/*` for accepted/rejected approaches.
- `todo.md` and handoff files for current operational state.

Then add:

```text
kb-map-refresh: done - <docs updated>
```

## Step 3.75: Memory Maintenance Signals

Update `docs/context/memory-maintenance.md` before cleanup. This file is a targeted signal index for future deep memory review. It must contain pointers, not just counters.

Create the file if it does not exist:

```markdown
# Memory Maintenance

Last deep review: never

## Counters Since Last Review

- Completed KB cycles: 0
- Durable memory refreshes: 0
- Closed handoffs: 0
- Contradiction signals: 0
- Overlap signals: 0
- Stale-doc signals: 0
- Bloat signals: 0
- Repeated-rediscovery signals: 0

## Signals Since Last Review
```

For every `kb-finalize` run:

1. Increment `Completed KB cycles` by 1.
2. Increment `Durable memory refreshes` if Step 3.5 ran `kb-map refresh`.
3. Increment `Closed handoffs` by the number of handoffs moved to `docs/handoffs/done/` during this completion.
4. Append exact signals found during review, compound, learn, evolve, memory refresh, and cleanup.

Signal types:

| Type | Record When |
|---|---|
| `contradiction` | two memory docs, handoffs, plans, requirements, or architecture notes disagree |
| `overlap` | two docs cover the same topic and may need consolidation/cross-linking |
| `stale-doc` | a memory doc references old paths, old behavior, old commands, or removed systems |
| `bloat` | a memory doc, handoff folder, research index, or `todo-done.md` is growing past useful startup size |
| `repeated-rediscovery` | this work re-learned something already discovered in another brainstorm, plan, research note, or solution |

Signal entry format:

```markdown
### YYYY-MM-DD - <type> - <short topic>
- Source: `<repo-relative path or manifest note>`
- Found during: `kb-finalize` / `<step or skill>`
- Signal: <one or two sentences with the actual issue>
- Suggested pass: <refresh, compact, consolidate, replace, cross-link, or promote>
```

Generic examples:

```markdown
### 2026-05-24 - contradiction - workflow source of truth
- Source: `docs/context/architecture/workflows.md`
- Found during: `kb-map refresh`
- Signal: requirements and architecture describe different canonical workflow definitions.
- Suggested pass: reconcile architecture and decision docs.

### 2026-05-24 - overlap - release deployment research
- Source: `docs/context/research/release-packaging.md`
- Found during: `kb-finalize`
- Signal: overlaps older deployment research on the same update-delivery path.
- Suggested pass: consolidate or cross-link research notes.
```

Do not invent signals to satisfy a quota. If no signal exists, only increment counters and add a manifest note: `memory-maintenance: no new signals`.

If any signal counter crosses a conservative threshold, add a recommendation to the final report, but do not run the deep pass automatically:

- Completed KB cycles >= 5
- Durable memory refreshes >= 10
- Closed handoffs >= 10
- Any contradiction signals >= 2
- Any combined signals >= 8

Recommendation format:

```text
Memory review recommended: <reason>. Run `kb-memory-review` against docs/context/memory-maintenance.md before the next large feature.
```

## Step 3.7: Compact and Alert Gate

Run `kb-compact` when review, learning, memory refresh, or maintenance signals show that durable memory is getting too large for fresh-session startup.

Use targeted compaction only:

- Compact a specific bloated architecture doc, handoff, research note, `todo-done.md`, or memory-maintenance section.
- Preserve exact paths, commands, current truth, known sharp edges, and unresolved decisions.
- Do not compact away active work, blockers, HITL items, open handoffs, or evidence needed for review.
- Record the result in manifest notes: `compact: <path> compacted` or `compact: skipped - no startup bloat`.

Alert the user in the final report when any condition needs deliberate follow-up:

- `kb-memory-review` threshold crossed.
- P2/P3 findings remain logged after the rectify gate.
- A memory contradiction, stale-doc signal, or repeated-rediscovery signal was recorded.
- Compacting was skipped because the doc was too risky to summarize safely.
- A required tool, reviewer, compound, learn, evolve, refresh, or compact step failed.

Alerts are concise status lines, not extra ceremony. They should tell the user exactly what needs attention and which file contains the evidence.

## Step 4: Cleanup

Prune ephemeral artifacts. Heavy KB usage generates file sprawl — clean it up per-feature, not manually.

1. **QA screenshots** — delete `.kb/qa-screenshots/` contents for this feature's slices. Screenshots should already be referenced in commits or PR bodies. Safe to remove.

2. **Observations log** — trim `.kb/observations.jsonl` entries older than 90 days. Matches the recency decay half-life in `/learn`. Use any available local scripting runtime; if no suitable runtime is available or the file does not exist, skip with a note.

3. **Plan files** — leave manifests and slice plans in `docs/plans/`. Lightweight reference material, useful for tracing decisions.

4. **Log cleanup** in the manifest notes:

   ```text
   cleanup: screenshots pruned, observations trimmed to 90d
   ```

5. **Todo hygiene** — verify `todo.md` contains only active rows, `🔒 blocked` rows, `🛑 human-required` rows, the `🧊 Parked / Cold Storage` section, or handoff-pointer work. If the completed feature or routine slice completion logs remain there, move a compact summary to `todo-done.md` and remove those entries from `todo.md`. `todo.md` must keep its `## Rules` section at the top; do not depend on `todo_rules.md`, `todo-rules.md`, or any separate rules file.

## Step 5: Done

Before updating the manifest to `status: reviewed`, write `complete-to-ship` in the manifest `gate_ledger`.

Required proof:

- `kb-check` final command/result;
- objective `done_check` result, or explicit scoped human exception;
- per-slice `proof_check` result or accepted `no_check_reason`;
- `kb-functional-test` or explicit skip reason for every functional/API/CLI/UI slice;
- `kb-review` mode and finding counts;
- P0/P1 resolved or human/quarantine blocker recorded;
- follow-up-resolution summary;
- proof/demo evidence paths or skip reason;
- compound/learn/evolve result or non-blocking failure note;
- project-memory refresh/skip proof;
- memory-maintenance update;
- cleanup result;
- alerts list.

If any required proof is missing, set `complete-to-ship` to `blocked` and do not report `KB <name> finalized`.

Update the manifest `status: reviewed` only after `complete-to-ship` is `passed` or explicitly `quarantined` for out-of-scope issues, and after `kb-gate/scripts/check_gate_ledger.py <manifest-path> --gate complete-to-ship --allow-quarantine` passes. Then report:

```text
KB <name> finalized.
- Review: P0=N P1=N(resolved) P2=N P3=N
- Follow-up resolution: <resolved N | logged M | blocked K>
- Proof/demo evidence: <commands/artifacts | skipped with reason>
- Compound: <written | skipped>
- Learn: <N new, M updated | no new patterns>
- Evolve: <promoted N | skipped | no candidates>
- Project memory: <refreshed | skipped with reason>
- Memory maintenance: <N signals recorded | no new signals | review recommended>
- Compact: <ran on paths | skipped with reason>
- Alerts: <none | concise follow-up lines with evidence paths>
- Cleanup: done

Ready for configured delivery. Return control to `kb-complete <manifest>`.
```

## Failure Handling

| Situation | Action |
|-----------|--------|
| kb-review fails to run | Log error, ask user whether to retry or skip review |
| P0/P1 fix breaks tests | Re-run tests, treat as new failure, fix before proceeding |
| proof/demo evidence cannot run | Log the missing tool/server/credential and alert user with the closest deterministic proof completed |
| compound/learn/evolve fails | Log error, continue — these are non-blocking |
| kb-compact fails | Log error and alert user with the bloated file path; do not block completion |
| Manifest not found | Ask user for path |
| Manifest has unfinished slices | Stop, tell user to run kb-work first |

## Integration

- **Input from:** `kb-work` or `kb-complete` (completed manifest)
- **Review engine:** `kb-review` with scope passthrough
- **Follow-up resolution:** `kb-gate`, `todo-triage`, `todo-create`, or repo-local resolvers
- **Proof/demo evidence:** `kb-check`, `kb-functional-test`, browser/CLI/API probes, optional repo-local demo skills
- **Documentation:** `ce-compound` → `docs/solutions/`
- **Project memory:** `kb-map refresh` → `docs/context/*`, `todo.md`, handoffs
- **Memory maintenance:** `docs/context/memory-maintenance.md` signal index
- **Compaction:** `kb-compact` for targeted memory bloat
- **Learning:** `/learn` → `docs/context/kb/instincts/project.yaml`
- **Evolution:** `/evolve` → `.github/skills/learned-*/`
- **Delivery handoff:** `kb-complete` selects local, PR, or direct policy
