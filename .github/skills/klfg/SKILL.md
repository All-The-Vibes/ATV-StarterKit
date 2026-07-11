---
name: klfg
description: "Deprecated compatibility alias for kb-complete. Existing 'klfg' and full-pipeline requests delegate to the single state-aware kb-complete command."
argument-hint: "[feature description, plan path, manifest path, or blank]"
disable-model-invocation: true
---

# KLFG Compatibility Alias

The strict brainstorm/plan/work/finalize gates remain enforced by their owning
skills. `kb-complete` now orchestrates those phases and applies project delivery
policy.

Immediately invoke:

```text
kb-complete <same arguments>
```

Do not duplicate phase execution. Re-read durable manifest state after
delegation and return the `kb-complete` outcome.
