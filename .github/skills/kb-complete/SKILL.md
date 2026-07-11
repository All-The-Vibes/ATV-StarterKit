---
name: kb-complete
description: "Single user-facing KB completion command. Takes a feature, plan, or manifest from its current state through planning, work, post-work finalization, and configured delivery: local completion, pushed PR, or explicitly configured direct integration and sync."
argument-hint: "[feature description, plan path, manifest path, or blank to resume active work]"
---

# KB Complete

Take KB work from its current durable state to the project's configured endpoint.
Users should not need to choose between plan, work, finalize, ship, or land.

```text
source/plan/manifest
  -> kb-plan when needed
  -> kb-work
  -> kb-finalize
  -> delivery policy
  -> local | PR | direct integration
  -> post-integration sync when configured
```

This is the user-facing orchestrator. Internal phases remain separate so they
can enforce narrow gates and resume safely.

## Safety Contract

- Explicit `kb-complete` invocation authorizes safe local planning, execution,
  review, repair, proof, learning, and cleanup.
- Publishing authority comes from project delivery policy or an explicit
  run-scoped user instruction. Absence of policy defaults to `local`.
- Never infer direct-default permission from repository ownership or write
  access. Permissions answer where a push can go; policy answers whether a PR is
  required.
- Never merge, direct-push default, deploy, or propagate external copies unless
  the selected policy explicitly authorizes that action.
- Do not stage, commit, revert, or overwrite unrelated dirty work.

## Input Resolution

1. Run `kb-map lookup <request>` and resolve the active project root.
2. If input is a manifest, use it.
3. If input is a slice plan, requirements doc, brainstorm, or handoff, follow
   its source/manifest pointers before creating duplicate artifacts.
4. If input is a clear feature description with no manifest, invoke
   `kb-plan <input>` with execution intent.
5. If material product or architecture questions remain, route through
   `kb-brainstorm` before planning.
6. If input is blank, resume the single active manifest from `todo.md`. Ask only
   when multiple active manifests are genuinely plausible.

## State-Driven Loop

Re-read the manifest after every delegated phase. Durable state, not chat memory,
chooses the next action.

| Current state | Action |
|---|---|
| no valid manifest | `kb-plan <source>` |
| active with runnable slices | `kb-work <manifest>` |
| completed with `work-to-complete: passed` | `kb-finalize <manifest>` |
| reviewed with `complete-to-ship: passed|quarantined` | apply delivery policy |
| blocked/human-required/parked | persist exact resume condition and stop |
| no state change after one repair | stop with the smallest unblock action |

`kb-work` may invoke `kb-finalize` automatically. Re-read the manifest and skip
already-proven phases rather than repeating review or learning.

## Delivery Policy

Read `docs/context/operations/kb-routing.yaml` when present. `kb-configure`
manages the portable project policy.

```yaml
delivery:
  mode: local        # local | pr | direct
  merge: manual      # manual | auto-after-checks
  post_merge_sync: false
```

Defaults when absent:

- `mode: local`
- `merge: manual`
- `post_merge_sync: false`

### Local

Stop after `kb-finalize` passes. Report the reviewed manifest and exact delivery
command if the user later wants publishing.

### PR

Invoke `kb-ship <manifest>` to audit scope, commit intentional files, push a
topic branch, and create/update a PR.

- With write access, use a same-repository topic branch.
- Without write access, use an authorized fork and upstream PR.
- Write access never bypasses the PR policy.
- With `merge: manual`, stop at the correctly based open PR.
- With `merge: auto-after-checks`, invoke `kb-land <manifest>` only after ship
  proof and required checks/approvals pass.

### Direct

Invoke `kb-land <manifest>` with direct-delivery policy.

- Direct mode must be explicitly stored or stated for this run.
- Branch protection, required reviews, stale default, failed release checks, or
  ambiguous scope force PR fallback or block; never bypass protection.
- Do not use force push or admin bypass.

## Terminal Outcomes

```text
KB complete: local|pr-open|landed|nothing-to-deliver|blocked
Manifest: <path>
Finalization: <complete-to-ship status>
Delivery policy: <local|pr|direct>
Branch: <branch or none>
Commit: <sha or none>
PR: <url or none>
Integration: <not-requested|pending-review|merged|direct>
Sync: <not-configured|done|blocked>
Next: none|<exact resume action>
```

Do not report `landed` unless the remote default branch contains the delivered
commit and any configured post-integration sync has been verified.

## Compatibility

- `kb-finish` is a deprecated alias that delegates here.
- `klfg` may delegate here for its full idea-to-endpoint loop.
- `kb-finalize`, `kb-ship`, and `kb-land` remain internal phase skills.
