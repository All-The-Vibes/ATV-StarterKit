---
name: kb-qa
description: "Quality assurance gate for all slices. Runs lint checks on every slice, browser verification on frontend slices. On any failure, invokes kb-repair for surgical fixes. Hard gate — the browser reports what rendered, the linter reports what's dirty, the model does not self-report."
argument-hint: "[slice plan path, or blank to verify the current slice]"
---

# KB QA — Quality Assurance Gate

Quality verification for all slices. Lint for every slice. Browser checks for frontend and UI-reachable behavior. On any failure, hands off to `kb-repair` for surgical fixes.

Prefer deterministic checks over model judgment. Use `kb-check` for lint, typecheck, tests, builds, browser checks, and audits before relying on visual or textual inspection.

Use `kb-functional-test` for user-visible workflows, API/CLI journeys, auth/session/persistence paths, and bugs that escaped unit tests. If the changed behavior is reachable through the UI and a browser transport exists, verification must drive it through the UI. Backend/API/unit checks may supplement but cannot replace that UI proof. Headless browser checks are preferred; visible browser sessions must be serialized.

## When to Run

Called from `kb-work` at Step 3.8, after all safety gates pass.

- **All slices:** lint check (Step 7)
- **UI-reachable slices:** browser verification (Steps 0–6) — triggered when `expected_files` includes frontend file extensions (`.tsx`, `.jsx`, `.html`, `.css`, `.scss`, `.vue`, `.svelte`, `.ejs`, `.hbs`) or when backend/API/state changes alter behavior a user reaches through the UI.

If the slice is truly backend-only with no affected UI path, skip browser checks with a one-line note: `qa-browser: skipped — no UI-reachable behavior changed`. Lint still runs. Do not skip browser checks merely because the implementation file is backend-side.

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Read the current slice context from the active KB manifest.

**If input is a path:** Read that slice plan directly.

## Step 0: Transport Selection

Pick the transport based on what the slice needs, not a fixed priority list.

### Decision Logic

1. **Is this an internal/corporate site?** (SSO, Conditional Access, company-owned domains, session cookies from a real login)

   - **YES** → CDP required. Connect to the user's existing browser session via `ws://localhost:9222` (or `CDP_ENDPOINT` env var). The real browser already has cookies, tokens, and session state from the user's login. No way to fake this.
   - If CDP is unavailable → **STOP.** Do not attempt Playwright or Agent Browser on internal sites — they cannot pass SSO/Conditional Access. Log in `todo.md`: `qa: skipped - internal site, no CDP session. Start browser with --remote-debugging-port=9222`

2. **Is this a regular site or local dev server?** (localhost, public URLs, no corporate auth)

   - Agent Browser if installed (`agent-browser` on PATH) — structured element targeting, fast
   - Playwright if available — headless, clean viewport control, required default for local/public UI functional checks
   - CDP as fallback

3. **Does the slice need responsive/viewport testing?** (deep tier, or slice touches layout/grid/responsive components)

   - Playwright preferred for multi-viewport. Headless, spawns 375px/768px/1440px cleanly.
   - Fallback: CDP with device emulation.

4. **None available** → Log: `qa: blocked - no browser transport available (checked: CDP, Playwright, Agent Browser)`. For UI-reachable changes this is a blocking verification gap, not a pass. Add a `🔒 blocked` or `🛑 human-required` entry in `todo.md` with the missing transport/setup.

## Step 1: Connect and Navigate

1. Connect via the selected transport.
2. **Enable continuous console monitoring.** From this point through all subsequent steps, capture console errors, warnings, and failed network requests after every action — not as a one-time check at the end. Every click, navigation, and form fill gets a before/after console snapshot.
3. **Scope pages using the diff.** Use the verified file list from kb-work Step 3.6 to determine which pages need testing:
   - Map changed files to testable URLs: `pages/dashboard.tsx` → `/dashboard`, `components/Header.tsx` → any page rendering that component.
   - Cross-reference with the slice plan's acceptance criteria for explicit URLs.
   - Dev server URL (default `http://localhost:3000`, respect `DEV_SERVER_URL` env var).
   - Test every changed UI path and every UI-visible acceptance criterion. Do not test the entire app, but do not collapse multiple affected routes into one convenient page.
4. Wait for the page to reach a stable state (network idle or DOM content loaded).

## Step 2: Capture Evidence

1. **Screenshot** the rendered page. Save to `.kb/qa-screenshots/<slice-id>-<timestamp>.png`.
2. **Capture console output** — errors, warnings, and failed network requests (4xx/5xx).
3. **Check for render failures** — blank pages, loading spinners that never resolve, error boundaries.

Create `.kb/qa-screenshots/` if it doesn't exist.

## Step 3: Verify Against Slice Requirements

Read the slice plan's acceptance criteria. For each criterion that describes visible behavior, verify it through the rendered UI itself:

| Check | Method |
|-------|--------|
| Element exists | Write and run a locator assertion against the rendered UI |
| Text content matches | Write and run a rendered text assertion |
| Layout is correct | Write and run a deterministic viewport/locator/screenshot assertion when possible |
| No regressions | Console has no new errors or warnings |

For every visible-behavior acceptance criterion, create an ephemeral assertion file for this QA pass and execute it. Prefer Playwright for browser apps; use the project's equivalent deterministic UI test runner when Playwright is not the stack. The proof must be executable code such as:

```ts
await expect(page.locator('.margin-value')).toHaveText('42%')
```

When writing ephemeral assertion files, use template literals for selectors and avoid locator strings built through concatenation. If a selector needs dynamic values, parameterize it:

- Wrong: `await page.locator('.deal-row[data-name="' + dealName + '"]').click()`
- Right: ``await page.locator(`.deal-row[data-name="${dealName}"]`).click()``

Do not pass a criterion by looking at a screenshot and deciding it seems right. Screenshots are evidence for debugging and review, not the pass/fail oracle. The pass/fail oracle is the executed assertion file, its command, exit code, timestamp, and any trace/log artifact.

Delete the ephemeral assertion file after the QA pass unless keeping it as a reusable regression test is clearly valuable. Preserve the command output, trace path, screenshot path, or log path needed for manifest proof.

If a visible criterion cannot be expressed as a programmatic assertion, flag it as `🛑 human-required` with the reason. Do not substitute model visual inspection for deterministic proof.

**Never read source code during QA.** Test as a user — if you can't verify it from the browser, flag it as needing manual verification.

Calling backend endpoints, dispatching component events directly, invoking button handlers, mocking network requests, inspecting component state, or reading logs does not satisfy visible-behavior criteria. Those are useful supporting evidence only after the rendered UI path has been exercised.

## Step 4: Interaction Checks (standard + deep tiers)

If the tier is `standard` or `deep` and the slice added interactive elements:

1. Click buttons, links, and interactive elements added by the slice.
2. Fill form fields if the slice added forms.
3. After each interaction, check for:
   - New console errors
   - Failed network requests
   - Unexpected navigation
   - UI elements disappearing or breaking
4. Screenshot after each significant interaction.

## Step 5: Responsive Checks (deep tier only)

If the tier is `deep`:

1. Resize viewport to **375px** (mobile) — screenshot + console check.
2. Resize viewport to **768px** (tablet) — screenshot + console check.
3. Resize viewport to **1440px** (desktop) — screenshot + console check.
4. Flag any layout breakage, overflow, or elements hidden at specific breakpoints.

## Step 6: Report

**On pass:**

```text
qa: PASS — <transport used>
  checks: N/N passed
  console: clean (0 errors, 0 warnings)
  screenshots: .kb/qa-screenshots/<slice-id>-*.png
  tier: quick|standard|deep
```

**On fail:**

For each failure:
1. Verify it's reproducible — retry once before reporting.
2. Screenshot the failure state (mandatory).
3. Log the specific check that failed and what was expected vs. observed.

```text
qa: FAIL — <transport used>
  failed: "<check description>"
  expected: "<what the slice said>"
  observed: "<what the browser showed>"
  screenshot: .kb/qa-screenshots/<slice-id>-fail-<n>.png
  console_errors: [list if any]
```

**FAIL on any critical check (browser or lint) invokes `kb-repair`.** The agent does not proceed to the next slice until all checks pass or the repair loop exhausts.

Log all results in `todo.md` under the slice's status or notes. Also update the manifest `notes` field.

## Tiers

Set per-feature or per-slice in `todo.md` or the manifest. Default is `quick`.

| Tier | What it does |
|------|-------------|
| `quick` | Screenshot + console check only |
| `standard` | + interaction checks on new/modified elements |
| `deep` | + responsive checks at 3 breakpoints (375px, 768px, 1440px) |

## Step 7: Lint Check (all slices)

Runs for every slice, not just frontend.

1. **Detect the project linter** from project instructions, `package.json` scripts, `.eslintrc`, `Makefile`, or project conventions.
2. **Scope to slice files** — pass `expected_files` as arguments when the linter supports file-level targeting. Avoid linting the entire repo for one slice.
3. **Capture output** — errors and warnings, with file paths and line numbers.
4. **If no linter configured**, skip with note: `lint: skipped — no linter configured`. Not fatal.

```text
lint: PASS — <linter used>
  files checked: N
  errors: 0, warnings: 0
```

```text
lint: FAIL — <linter used>
  errors: N in M files
  <file:line — error message>
```

Lint failures are fixable — they feed into the failure handoff (Step 8).

## Step 8: Failure Handoff

After Steps 0–7 complete, collect all failures (browser and lint).

- **All checks passed** → QA passes. Return to `kb-work`.
- **Any check failed** → Invoke `kb-repair` with the full failure report:
  - Which checks failed (browser, lint, or both)
  - Failure details (expected vs observed, lint errors with file:line)
  - Slice context (`expected_files`, slice plan path)
  - Screenshots for any browser failures
- `kb-repair` handles surgical fixes and re-verification.
- If repair succeeds → QA passes.
- If repair exhausts (stuck or ceiling hit) → QA fails. Slice stays `in_progress`. The agent MUST NOT proceed to the next slice.

## Transport Reference

Actions are transport-agnostic above. Here's the mapping:

| Action | CDP | Agent Browser | Playwright |
|--------|-----|---------------|------------|
| Connect | `ws://localhost:9222/json` → get `webSocketDebuggerUrl` | `agent-browser open <url>` | `playwright.chromium.launch()` |
| Navigate | `Page.navigate` | `agent-browser open <url>` | `page.goto(url)` |
| Screenshot | `Page.captureScreenshot` | `agent-browser screenshot <file>` | `page.screenshot()` |
| Console errors | `Runtime.consoleAPICalled` event listener | `agent-browser snapshot -i` (inspect output) | `page.on('console')` |
| Click element | `DOM.querySelector` + `Input.dispatchMouseEvent` | `agent-browser click @<ref>` | `page.click(selector)` |
| Get text | `DOM.querySelector` + `DOM.getOuterHTML` | `agent-browser snapshot -i` + parse | `page.textContent(selector)` |
| Resize viewport | `Emulation.setDeviceMetricsOverride` | Not supported — use Playwright | `page.setViewportSize()` |
| Network errors | `Network.responseReceived` event listener | Not directly — check console | `page.on('response')` |

## Principles

- Never read source code during QA. Test as a user, not a developer.
- Verify before documenting — retry once to confirm reproducible.
- Screenshot evidence is mandatory for any failure.
- Write incrementally, don't batch findings.
- Auth is sacred — never attempt authenticated routes without a real session.
- If a UI-reachable check can't be verified from the browser, block or flag human-required setup rather than guessing or substituting backend-only verification.

## Integration

- **Called from:** `kb-work` (Step 3.8, all slices)
- **Hands off to:** `kb-repair` (on any failure)
- **Results feed into:** `kb-review` (Step 5.4) as additional context
- **Screenshots persist:** `.kb/qa-screenshots/` (gitignored, ephemeral)
- **Logs persist:** `todo.md` notes + manifest notes
