---
name: kb-regression-snapshot
description: Capture and replay deterministic regression snapshots between KB slices. Use after a slice passes QA to freeze passing state, and before future slices start to prove earlier slice behavior has not regressed.
argument-hint: "[capture|verify plus slice id/spec path]"
---

# KB Regression Snapshot

Freeze what passed so later slices cannot quietly break it.

This is not a test-design skill. The LLM defines the smallest useful snapshot spec. The runner executes it mechanically.

Use this for cross-slice application state snapshots. Use the skill-eval
baseline path for skill-harness fixture regressions. The two mechanisms solve
different problems: app/workflow replay between slices versus scorer output
comparison across eval runs.

## Runner

Use:

```powershell
.github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1 capture -SliceId <id> -SpecPath <spec.json>
.github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1 verify
```

Store snapshots at:

```text
.kb/snapshots/<slice-id>.json
```

## Snapshot Shape

```json
{
  "slice_id": "JE3",
  "captured_at": "ISO-8601",
  "checks": [
    {"type": "dom-element", "url": "/dashboard", "selector": ".margin-value", "expected_text": "42%"},
    {"type": "route-status", "url": "/api/deals/AIG", "expected_status": 200},
    {"type": "file-checksum", "path": "src/config.ts", "sha256": "abc123..."}
  ]
}
```

## Capture

After a slice passes `kb-check`, `kb-functional-test`, and `kb-qa`, build a compact spec from what changed:

| Change | Snapshot checks |
|---|---|
| Frontend/UI | route URL, key DOM selector, expected text or text pattern, console error count `0` |
| API | endpoint URL, expected status, required response fields or schema shape |
| CLI | command, expected exit code, expected output substring |
| Files | path and SHA-256 checksum for generated/config/runtime files |

Use behavioral checks for UI snapshots. Prefer "the margin value is visible and numeric" over "a div has class X." A class selector is acceptable only as a stable locator, not as the behavior being proven.

Do not store secrets, cookies, tokens, credentials, or large response bodies. Store only deterministic assertions and small metadata.

## Verify

Before the next slice starts execution, run the runner in `verify` mode against all previous snapshots.

The runner must:

- verify DOM checks with Playwright/CDP or the repo browser transport;
- verify API/route status with `fetch`, `curl`, or platform equivalent;
- verify CLI checks by executing the command and checking exit code/output;
- verify file checksums with SHA-256;
- exit nonzero on the first failed snapshot.

If any snapshot fails, STOP. Mark the current slice `🔒 blocked` with the failing snapshot path, check type, expected value, observed value, and log/trace path. Do not edit implementation files until the regression is resolved, parked by the human, or explicitly skipped.

## Output

Capture:

```text
snapshot-capture: PASS JE3 -> .kb/snapshots/JE3.json
```

Verify:

```text
snapshot-verify: PASS 7/7 snapshots
```

or:

```text
snapshot-verify: FAIL .kb/snapshots/JE3.json
failed: dom-element /dashboard .margin-value
expected: 42%
observed: <missing>
```

Record the result in the manifest notes. Snapshot verification is acceptable machine proof for `kb-complete`.
