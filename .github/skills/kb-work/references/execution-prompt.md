# KB Work Slice Execution Prompt

Use this as the per-slice execution checklist.

```text
You are executing a single vertical slice. Complete it fully.

KB: <kb_id>
Slice: <slice_id> - <title>
Verification mode: <tdd|integration|functional|verification-only>

Plan contents:
<full slice plan content>

Context packet:
<validated packet, or explicit small legacy/no-packet reason>

Route request:
<planned correction/authority tier, optional explicit runtime attempt_tier,
saved user-local project source priority, allowed fallback, user overrides, and current-model
degraded fallback policy>

Router commands:
<exact discover command, exact select command, and exact dispatch command with slice-unique artifact names; or router-unavailable reason>

Execution policy:
<lower-tier attempts enabled|disabled, attempt eligibility evidence, exact proof,
and escalation triggers>

Instructions:
1. Read the plan completely.
2. Set up on the current branch.
3. If the slice runs Go inside a workspace sandbox, load
   `references/go-sandbox.md` and apply its environment inside every Go shell
   invocation. Never put its temp overrides on the agent launcher.
4. Use the packet's files and deterministic prefetch before broad search.
   Escalate when an escalation trigger fires or the packet is insufficient;
   do not silently expand authority.
5. Treat model routing as live dispatch, not a plan commitment. Request the
   chosen route immediately before execution and keep slice authority bounded.
   Use `attempt_tier` only when the current master explicitly established bounded
   intent, scope, authority, proof, trust, and escalation triggers; never infer
   it from “code” or a file extension.
6. Apply user-local project source priority (`automatic`, `self-hosted-first`, or
   `native-first`) only among eligible live routes. Plans never choose a model,
   alias, adapter, endpoint, or transport. Source preference never grants
   attempt eligibility; only run-scoped `require <model>` hard-pins.
7. Record the actual route and provenance only from dispatcher or host evidence.
   If the host cannot prove the selected model/session, report provenance as
   `unknown`/`unavailable`.
8. Run the exact deterministic proof after an AMR attempt. A passing attempt is
   kept without a stronger-model rewrite. Proof—not model self-review or a
   routing receipt—is the acceptance oracle.
9. If proof fails, prepare the planned-tier-or-higher correction packet with independently accepted
   result, failed criterion and location, smallest allowed change, invariants,
   relevant interfaces, exact proof result, compact current diff, attempt
   ledger, `corrective_diff_only: true`, and focused/regression proof to rerun.
   Preserve only hunks with a machine-verifiable hunk-local acceptance oracle.
   Without that oracle or when failure is not localizable, abort the surgical
   pilot path and record separate ordinary planned-tier execution; do not infer
   broadened authority.
   Treat this as a handoff only. Automatic correction dispatch into the live
   checkout is disabled until an isolated workspace, host-owned proof runner,
   and compare-and-swap apply path exist. Fall back to ordinary planned-tier
   execution and claim no preserved-work savings.
10. For files marked `op: edit` in expected_files:
   - Read the current file content first.
   - Make only the change described in the `scope` field.
   - Preserve all existing behavior not mentioned in scope.
   - Current disk content is authoritative over stale plan text.
11. For files marked `op: create`, create the planned file.
12. Apply the verification mode:
   - tdd: failing test -> implementation -> passing test -> refactor.
   - integration: integration test proves the wired path.
   - functional: workflow/API/CLI/UI path is proven from public surface.
   - verification-only: build/check proves no regression.
13. Run relevant deterministic checks first, then broader checks when practical.
14. Stage only files changed for this slice.
15. Commit only if the user asked for commits.

Do not modify other slices' files unless required for this slice.
Do not add scope beyond what the plan specifies.
Do not stage unrelated dirty or untracked files.
```
