---
name: solution-debranding-apply
description: "Execute approved units from an existing solution-debranding plan, validate each, and update plan evidence. Processes all unblocked units in sequence, pausing only when a human decision is needed. Use only when the user supplies or uniquely identifies a primary plan and asks to apply, execute, or ship that planned work. Requires an existing plan. When no plan exists, use solution-debranding-plan instead."
argument-hint: "[plan path] [unit number or 'all' for batch]"
---

# Solution Debranding Apply

Execute approved work from an existing solution-debranding plan.

## Execution Mode

When a specific unit number is given, execute that unit only. When `all` is given or no unit is specified, execute all unblocked units in dependency order, one at a time, validating each before moving to the next. **Pause and ask the user** when:

- A unit requires a human decision (legal, security, privacy, ownership, or external-system approval)
- A validation fails and the fix is ambiguous
- A unit is gated on work outside the debranding track
- The next unit's scope is unclear or the plan says "Your call"

Do not pause for mechanical work that has a clear correct answer. Resume automatically after each completed unit.

## Steps (per unit)

1. Require one primary plan path. If none is supplied, discover a unique active plan under the configured `ARTIFACT_ROOT`; otherwise stop and request the path. Never create a replacement plan.
2. Read the plan's summary and unit table, then read only the section its reference column names for the unit you are executing. Read the whole plan only when the unit has no reference. Then read [the shared workflow package](../solution-debranding/SKILL.md) and follow its Apply contract with `MODE=apply`.
3. Confirm the requested unit is unblocked and that required approvals are recorded. If not, skip to the next unblocked unit rather than stopping entirely.
4. Execute the unit. Referred findings are not units: decline them and name their track owner.
5. After the first substantive edit, run the cheapest focused validation. Finish with the relevant tests, builds, scans, or rendering checks for that unit.
6. Update the same plan's checkboxes, status, and `## Validation` section with commands, outcomes, limits, and residual findings.
7. After completing a unit, proceed to the next unblocked unit unless a pause condition (above) is hit.

## Response Format

End every message with a TLDR block:

```
---
TLDR: <one sentence summary of what was done and what's next>
```

Before the TLDR, provide the unit outcome, validations run, any limits, and the plan link. When pausing to ask the user a question, state the question clearly before the TLDR.

When all executable units are complete, list the remaining human-gated units and what each is waiting on.
