---
name: kb-configure
description: "Configure optional portable per-project lower-tier-attempt and delivery policy."
argument-hint: "[show|attempts-on|attempts-off|delivery-local|delivery-pr|delivery-direct|reset]"
---

# KB Configure

Configure optional project execution policy without making ordinary KB startup
interactive.

Adaptive Model Routing (AMR) is automatic and needs no project setup. This skill
owns the lower-tier-attempt opt-out and delivery policy. Personal source
preference belongs to user-local `kb-models` state keyed by project identity.

## Config Path

`docs/context/operations/kb-routing.yaml`

The file contains portable project policy only. Never write model endpoints,
auth environment-variable names, trust approvals, commands, or credentials
there. `kbrouter` continues to own host and user-local model configuration.

## Behavior

1. If the file exists, read it and show a compact AMR/delivery summary. Ask only
   which setting the user wants to change when their request is ambiguous.
2. If the file is absent and an argument was supplied:
   - `attempts-on` enables one bounded next-lower-tier AMR attempt.
   - `attempts-off` starts every slice at its planned correction tier.
   - `delivery-local` keeps reviewed work local.
   - `delivery-pr` commits, pushes a topic/fork branch, and opens/updates a PR.
   - `delivery-direct` permits verified direct-default integration; protection
     or policy rejection falls back to PR or blocks.
   - `show` reports next-lower attempts disabled pending pilot promotion and
     local delivery without creating a file.
   - `reset` removes only this project policy after explicit confirmation.
3. If the file is absent and no mode was supplied, show the defaults and the
   exact commands above. Do not start a setup questionnaire.
4. Do not ask model-by-model questions. AMR discovers the active host catalog at
   work time; `kb-models` configures optional user-local extras only.
5. Preserve unrelated project policy when updating an existing file.

## Canonical Schema

```yaml
schema_version: 1

amr:
  mode: automatic
  lower_tier_attempts: disabled

delivery:
  mode: pr
  merge: manual
  post_merge_sync: false
```

These safety rules are fixed rather than configurable:

- `model_tier` is the planned correction/authority tier, not the validator.
- A lower-tier attempt is allowed only after the current driver establishes
  settled intent, bounded scope and authority, objective proof, trust, and
  escalation triggers. The selector never infers eligibility.
- AMR makes at most one next-lower-tier attempt before deterministic proof.
- Passing attempts are preserved without a stronger-model rewrite.
- Failing attempts produce a surgical planned-tier correction handoff packet.
  Current automatic correction execution is disabled because no isolated
  workspace plus compare-and-swap apply runner exists; fall back to separate
  ordinary planned-tier execution and claim no preserved-work savings.
- Ordinary proof remains authoritative. AMR receipts are telemetry.
- Repository ownership/write access never selects direct delivery by itself.
- Direct delivery, automatic merge, and post-merge sync require explicit policy.

## Defaults

When no config exists:

- AMR: automatic.
- Bounded lower-tier attempts: disabled until explicit pilot/opt-in or promotion.
- Delivery: local.
- `kb-start`, `kb-plan`, and `kb-work` do not ask configuration questions.

## Output

After writing, report the path and a one-line summary:

```text
KB configured: AMR next-lower attempts disabled; delivery PR/manual.
```
