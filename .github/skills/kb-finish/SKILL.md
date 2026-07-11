---
name: kb-finish
description: "Deprecated compatibility alias for kb-complete. Use when an older prompt or workflow says 'kb finish'; delegate the same input to kb-complete and report the kb-complete result."
argument-hint: "[feature description, plan path, manifest path, or blank]"
---

# KB Finish Compatibility Alias

`kb-complete` is the single user-facing command for plan-to-endpoint work.

Immediately invoke:

```text
kb-complete <same arguments>
```

Do not run a second plan/work/finalize/ship loop. Re-read durable manifest state
after delegation and return the `kb-complete` outcome.
