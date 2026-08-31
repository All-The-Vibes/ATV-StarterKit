---
name: solution-debranding-plan
description: "Assess a brand-specific repository and create a debranding, interchangeable-branding, modernization, and release-readiness plan. Use when the user says 'debrand this repo', 'remove the brand', 'make this reusable', 'prepare this for reuse or resale or handover', 'prepare this for public release', 'genericize this codebase', or 'white-label this solution'. Plans only and never changes product code. Prefer this over solution-debranding-apply whenever no approved plan exists yet."
argument-hint: "[source brand] [scope] [release-scope: internal|external|public] [artifact-root=path] [design-system=url-or-name]"
---

# Solution Debranding Plan

Create the durable assessment and implementation plan that `solution-debranding-apply` consumes. `solution-debranding-verify` checks the result independently.

## Steps

1. Read [the shared workflow package](../solution-debranding/SKILL.md) and follow its Plan contract with `MODE=plan`.
2. Resolve the source brand, scope, brand scope, modernization scope, release scope, artifact mode, artifact root, and run name from the arguments and repository evidence.
3. Create or update the canonical plan under `ARTIFACT_ROOT`, which defaults to `docs/debranding`.
4. Stop after planning. Do not modify product, configuration, infrastructure, operational, external-system, or Git-history content.

## Response Format

Return no more than eight short lines of prose containing the outcome, top recommendation, blocking decisions, primary plan link, and the exact next step naming `solution-debranding-apply` with the plan path.

Then print the unit ledger the reporting contract defines: every implementation unit in recommended order, with its number, disposition, dependencies, size, and whether it needs a human decision. Invite the user to confirm or reorder, and offer `unit=N` to start one. Name how many findings were referred to other tracks. The ledger does not count against the eight lines.
