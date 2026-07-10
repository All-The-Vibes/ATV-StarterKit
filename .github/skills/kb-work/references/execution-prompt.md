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
<requested capability floor, allowed fallback, user overrides, and current-model degraded fallback policy>

Router commands:
<exact discover command, exact select command, and exact dispatch command with slice-unique artifact names; or router-unavailable reason>

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
   chosen route immediately before execution, keep the slice authority bounded,
   and do not silently move to a lower capability tier.
6. Record receipt/provenance only from dispatcher or host evidence. If the host
   cannot prove the selected model/session, report provenance as
   `unknown`/`unavailable`.
7. For files marked `op: edit` in expected_files:
   - Read the current file content first.
   - Make only the change described in the `scope` field.
   - Preserve all existing behavior not mentioned in scope.
   - Current disk content is authoritative over stale plan text.
8. For files marked `op: create`, create the planned file.
9. Apply the verification mode:
   - tdd: failing test -> implementation -> passing test -> refactor.
   - integration: integration test proves the wired path.
   - functional: workflow/API/CLI/UI path is proven from public surface.
   - verification-only: build/check proves no regression.
10. Run relevant deterministic checks first, then broader checks when practical.
11. Stage only files changed for this slice.
12. Commit only if the user asked for commits.

Do not modify other slices' files unless required for this slice.
Do not add scope beyond what the plan specifies.
Do not stage unrelated dirty or untracked files.
```
