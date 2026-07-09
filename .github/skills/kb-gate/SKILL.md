---
name: kb-gate
description: Shared phase-gate policy for KB workflows. Use before moving from brainstorm to plan, plan to work, work to complete, or complete to ship when P0/P1/P2/P3/P4 findings, review issues, ambiguity, weak tests, or unresolved risks exist.
argument-hint: "[phase, artifact path, or finding list]"
---

# KB Gate

Do not let known issues drift silently into the next phase.

## Gate Ledger

For multi-phase workflows, expensive tests, benchmark waves, or any work where
phase drift is costly, use `references/gate-ledger.md`. The ledger is the
source of truth for whether the next phase may start.

Hard rule: do not advance a workflow because the agent believes a phase is
"probably done." Advance only when the relevant gate record is `passed` or, for
out-of-scope evidence, explicitly `quarantined` with forbidden claims recorded.

When a phase claims a known failure was fixed and the failure has a runnable
sensor, prefer proof from `go run ./cmd/kbcheck accept --check <check.json>
--trace .kb/trace.jsonl`. Latest-green evidence without a recorded prior RED is
not proof of a repair.

Before answering that a phase is complete:

1. Read or create the manifest's `gate_ledger`.
2. Verify every required evidence item has proof.
3. Classify unresolved findings.
4. Set the gate status and `allowed_next_action`.
5. Report the gate check output.

If no manifest exists, write the gate record into the active plan, packet,
handoff, or `todo.md` section and require the next phase to create a manifest
before execution.

## Severity

| Severity | Meaning | Default |
|---|---|---|
| P0 | Will likely build the wrong thing, break core behavior, create safety/security/data risk, or make the next phase invalid | Block |
| P1 | Important ambiguity, missing verification, serious design/test gap, or likely rework | Block |
| P2 | Non-blocking but fixable quality, clarity, edge-case, or maintainability issue | Offer rectify-all |
| P3 | Minor polish, wording, naming, or cleanup issue | Offer rectify-all when cheap |
| P4 | Tiny wording, formatting, traceability, or optional cleanup note | Defer or fix only when already touching the artifact |

## Who Fixes

Severity is not the same as human-in-loop. Findings are expected. Stop only when an unresolved finding would make the next phase wrong, risky, or dependent on a human-only decision.

P0/P1 block the next phase while unresolved, but they do not automatically require human input.

The agent should rectify P0/P1 without asking when the fix is safe and evidence-backed:

- contradiction or stale wording in a doc;
- missing acceptance criterion derivable from the source material;
- missing verification mode or expected files;
- broken dependency DAG with one clear correction;
- deterministic test/lint/build failure with a local fix;
- review finding with a concrete safe/gated auto-fix.

Ask the human only when resolution requires:

- product intent or priority judgment;
- accepting or rejecting scope;
- credentials, login, external system access, or real-world approval;
- destructive/risky operation;
- choosing between multiple reasonable architecture/product paths;
- changing the user's stated requirements.

## Phase Gates

- **Brainstorm -> plan:** block on unresolved P0/P1 requirements, contradictions, missing core behavior, unsafe assumptions, missing verification inputs, unresolved `ask-now` items, unresolved `research-first` items, or unlabeled material assumptions.
- **Plan -> work:** block on broken DAG, missing acceptance criteria, missing verification mode, missing expected files, weak functional coverage, unsafe HITL, missing objective `done_check` when `objective_contract: true`, missing per-slice `proof_check`/valid `no_check_reason`, invalid `model_route`, or unresolved architecture/security risk.
- **Work -> complete:** block on failing deterministic checks, failed functional flows, scope violations, unresolved durable memory refresh, objective-contract validation failure, missing slice proof-check evidence, or blocked slices not explicitly parked.
- **Complete -> ship:** block on unresolved P0/P1 review findings, failed checks, missing objective `done_check` result, missing proof/demo evidence, release risk, or unrecorded human-only blockers.

P2/P3/P4 do not block by severity alone. Before moving on, fix the cheap/actionable ones that improve the artifact. Defer only when the finding is genuinely non-blocking and logging it will not cause avoidable rework.

## Question Gate Classes

Use these classes whenever a workflow is tempted to assume its way across a
phase boundary:

| Class | Meaning | May Advance? |
|---|---|---|
| `ask-now` | Human answer changes scope, user intent, acceptance criteria, safety, architecture direction, or verification | No |
| `research-first` | Source or external research can answer before user input | No, until researched or reclassified |
| `safe-assumption` | Reversible, low-risk assumption with evidence and a proof hook | Yes |
| `defer-to-planning` | Technical detail better answered during planning or code reading | Yes, only into planning |
| `parked` | Explicitly out of scope | Yes, with forbidden claims recorded |

`safe-assumption` is not a loophole. It must name why the assumption is
reversible, what evidence supports it, and which later proof would catch it if
wrong. If any of those are missing, classify the item as `ask-now` or
`research-first`.

## Rectify Prompt

When findings exist, first classify them:

- `auto_rectify`: agent can fix safely now;
- `needs_human`: requires one of the human-only conditions above;
- `defer_log`: non-blocking and not worth fixing now.

Fix `auto_rectify` items before asking. Then ask only for remaining human/judgment decisions:

```text
I found <count> issues before <next phase>: P0=<n>, P1=<n>, P2=<n>, P3=<n>, P4=<n>.
I can rectify <auto_count> safely now. <human_count> need your decision.
I will rectify safe/actionable issues now and ask only for the decisions I cannot make.
```

After fixing safe/actionable issues, rerun the relevant review/check, then continue if the remaining findings are deferred or non-blocking.

If the user asks not to fix findings:

- P0/P1 still block unless resolved, reclassified with evidence, or converted into a scoped parked item outside this work.
- P2/P3/P4 may be logged in the artifact, `todo.md`, or `todo-done.md` with owner/status.

## Output

Report severity counts, actions taken, remaining blockers, deferred items, and whether the next phase is allowed.
