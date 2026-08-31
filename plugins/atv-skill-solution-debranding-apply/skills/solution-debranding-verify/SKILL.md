---
name: solution-debranding-verify
description: "Independently verify a solution-debranding plan with full scans, relevant tests, two-brand checks, and explicit residual risk. Use when the user asks whether debranding is complete, correct, or safe to release. Reports failures and never remediates them."
argument-hint: "[plan path] [additional verification scope]"
---

# Solution Debranding Verify

Verify completion without silently remediating findings.

## Steps

1. Resolve the primary solution-debranding plan from the supplied path or a unique active plan under the configured `ARTIFACT_ROOT`.
2. Read the plan, then read [the shared workflow package](../solution-debranding/SKILL.md) and follow its Verify contract with `MODE=verify`.
3. Run the full debranding scan with `--fail-on-blocking`, every confirmed identity-map term, and `--summary`. Any nonzero result blocks verification. Confirm the scan reports no unreviewed UUID-shaped literals in any text format and no tenant-specific Microsoft authority. Run readiness scans plus the tests and builds relevant to detected repository surfaces. Render configuration or infrastructure only when present. When branding is in scope, validate both fictional brand profiles, including their palette, typeface, and logo assets when those visual surfaces exist. When a design system is declared in the plan, confirm rendered output uses only tokens from that system and flag raw values that bypass the token layer. Where authentication is in scope, verify source precedence, redacted failure states, provider routing, and credential refresh or rotation without restart when hot reload is promised. Include a negative test proving no provider credential is sent to another destination.
4. For public release, verify the recorded history-scan decision, clean-clone workflow, and disposition of external surfaces. Do not modify those surfaces.
5. Verify only the tracks the plan owns. Report referred findings as handed over and unverified rather than as failures of this run.
6. Update the plan's `## Validation` section and status with reproducible evidence, exclusions, residual risk, and a clear pass, blocked, or failed verdict.
7. Do not fix failures. Report them and stop. Remediation belongs to `solution-debranding-apply` on a planned unit.

## Response Format

Return no more than eight short lines of prose containing the verdict, the tracks it covers, failed or excluded checks, plan link, and the exact next step when remediation remains. A pass verdict must name its tracks, because a reader who sees only "pass" will take it for a whole-repository clearance.
