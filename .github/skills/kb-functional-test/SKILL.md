---
name: kb-functional-test
description: Functional-test strategy and test-quality audit for KB workflows. Use when deciding whether a slice needs functional/e2e/browser/API workflow tests, when existing tests may be mocked theater, or when user-visible behavior must be verified without manual QA.
argument-hint: "[slice plan, feature, changed files, or test file]"
---

# KB Functional Test

Functional tests prove the real workflow works. Unit tests prove parts. Both matter, but mocked tests that never exercise the behavior do not count.

## Ownership

This skill owns the test-level decision for KB work. `kb-plan` records the initial decision, `kb-work` re-checks it before marking a slice done, and `kb-finalize` uses it for manifest-level smoke coverage.

This is a lazy helper lane. Do not load it for every slice. Load it when the
test level is unclear, UI/API/CLI behavior could be faked by low-level tests, or
the slice touches user-visible behavior.

`verification` describes the workflow mode (`tdd`, `integration`, `functional`, `verification-only`, `hitl`). `test_level` describes the concrete proof required:

| `test_level` | Use When | Minimum Proof |
|---|---|---|
| `none` | copy/style/docs-only, dead-code removal, generated/config-only with build/lint coverage | lint/build/check only |
| `unit` | isolated pure logic, parser/formatter/helper, no public workflow boundary | focused unit test that fails before fix |
| `integration` | wiring between modules, persistence, callbacks, service boundaries, API contract internals | integration test using real collaborating code where practical |
| `functional-api` | HTTP/API workflow, action/tool endpoint, command handler, data contract visible to callers | API smoke/test through public surface |
| `functional-cli` | CLI command or script behavior users/operators invoke | CLI smoke script/test with real arguments and observable output |
| `functional-browser` | UI flow, DOM state, navigation, visual interaction, browser-only behavior, or major functionality exposed through the UI | headless Playwright/Cypress/CDP/agent-browser probe that drives the behavior through the UI with observable assertions |
| `full` | release/high-risk flow touching auth, persistence, integration, UI, or multiple critical paths | targeted functional checks plus broader suite/smoke |

`functional_risk` is the execution breadth: `none`, `narrow`, `broad`, or `full`.

## Mandatory Auto-Classification

Run this before any judgment-based classification.

If `expected_files` contains `.tsx`, `.jsx`, `.vue`, or `.svelte` files, set `test_level: functional-browser` automatically. The agent must not downgrade this classification to `unit`, `integration`, `functional-api`, `verification-only`, or `none`.

Also set `test_level: functional-browser` when non-UI files change behavior whose primary user path is through the rendered app UI.

`functional-browser` means:

1. Start or connect to the running app.
2. Use Playwright to navigate to the actual feature route/screen in the rendered UI.
3. Exercise the happy path with real clicks, keyboard input, form input, navigation, or other user-visible controls.
4. Assert observable rendered outcomes after the interaction; do not assert only backend calls, component handlers, mocked requests, or internal state.
5. Capture screenshots of key pass/fail states as evidence.
6. Clean up any artifacts, test data, screenshots, traces, temp files, or browser state created during testing according to the repo's QA cleanup rules.

If Playwright cannot access the target because the route requires an existing authenticated corporate browser session, use the repo/platform's authenticated browser transport (for example CDP) and record why Playwright was not viable. This is still `functional-browser`; it is not backend/API verification.

## Behavioral Assertions

Functional assertions must test behavior the user would verify, not implementation structure.

Behavioral assertions are correct:

- "The margin value is visible and shows a number" -> `locator('.margin-value').toBeVisible()` plus text matching a number pattern.
- "Clicking submit shows a success state" -> click submit, wait for the success element, assert it is visible.
- "The deal list shows at least one entry" -> assert `.deal-row` count is greater than 0.

Structural assertions are wrong:

- "There is a div with class `margin-value` containing 42%" -> breaks on harmless CSS or markup refactors.
- "The form POSTs to `/api/deals/submit`" -> tests implementation, not user behavior.
- "Component `state.isLoading` becomes false" -> tests internals, not rendered outcome.

Use stable selectors when needed, but make the asserted condition behavioral. Before writing an assertion, ask: "Would a user verify this the same way?" If not, it is structural and should be rewritten.

## How To Decide

Classify from concrete evidence, not vibes:

1. Read the slice acceptance criteria and `expected_files`.
2. Identify the public surface touched: UI, API, CLI, tool/action, job, event, persistence, auth/session, streaming, external integration.
3. Ask: "Could a unit test pass while the user-visible workflow is broken?"
   - If yes, require integration or functional proof.
   - If no, unit-level proof may be enough.
4. Ask: "Does the changed behavior cross a boundary?"
   - Process/module boundary -> `integration`.
   - API/CLI/tool public boundary -> `functional-api` or `functional-cli`.
   - Browser/DOM/user interaction boundary, or backend behavior whose primary user path is UI-driven -> `functional-browser`.
5. Ask: "Did a bug escape lower-level tests before?"
   - If yes, add a functional regression check at the level where the bug escaped.

Default upward when uncertain. A narrow functional check is cheaper than shipping a workflow lie.

## When Functional Tests Are Required

Require at least one functional or integration-style test when a change touches:

- user-visible UI flow;
- frontend route/component/state, browser-only behavior, or a major feature path the user reaches through the UI;
- API route, command, tool/action, or workflow orchestration;
- auth, permissions, session, persistence, streaming, external integration, or background job behavior;
- wiring between two or more subsystems;
- a bug that escaped unit tests.

If the changed behavior has a runnable UI path, the functional proof must go through the rendered UI. That means opening the page/screen, interacting with the visible control the user would use, and asserting the rendered result or visible state change. Backend/API/unit tests may supplement that proof, but they do not satisfy it. Do not downgrade to API/backend-only verification because it is faster.

Skip or defer functional tests only when the change is dead-code removal, local refactor with existing coverage and no user-visible path, or a generated/config-only change covered by build/lint. Pure copy/style still needs rendered UI evidence when it changes what a user sees.

## Proof Classification Only

This skill classifies the proof required. It does not lower a slice's planned
correction tier, select an implementation attempt, or decide whether a worker's
result passes.

The classification task itself is eligible for small/mini models when the
platform supports model-tiered subagents.

Good mini-model tasks:

- classify `test_level` and `functional_risk` from a slice plan;
- audit whether existing tests are meaningful or mocked theater;
- suggest the narrowest deterministic command or Playwright/API/CLI probe;
- summarize required test inputs.

Do not use a mini model as the final proof of behavior. The proof is the command, test, browser probe, screenshot, or failing/passing output. Escalate to a stronger model when classification depends on architecture, auth/security, flaky async behavior, complex UI state, or repeated failures.

When a slice declares `model_tier`, preserve it as the planned
correction/authority tier. `kb-work` alone decides whether the already-bounded
slice and this objective proof make one lower-tier implementation attempt safe.
“This is code” is not evidence. Subjective UX/design, philosophy/policy,
unresolved architecture, security/auth/data boundaries, and proof that cannot
reject a bad implementation remain ineligible for a lower-tier attempt.

## Test Quality Audit

An existing test is meaningful only if it:

- would fail against the broken or pre-change behavior;
- exercises the public surface, not private implementation details;
- asserts observable output, persisted state, emitted event, response contract, DOM state, or side effect;
- mocks only external boundaries, not the behavior under test;
- can fail for the bug it claims to cover.

If a test mostly asserts mocks were called, snapshots noise, or duplicates implementation logic, mark it weak and add a better functional probe.

## Execution Timing

- **During a slice:** run the narrowest functional check that proves the changed path. For UI-reachable behavior, drive the changed workflow through the UI itself. Prefer headless browser/API/CLI checks only after choosing the correct public surface.
- **After a manifest:** run broader workflow smoke tests across changed areas.
- **Before ship:** run full functional/e2e suite when practical, plus targeted high-risk flows.
- **During parallel work:** only one worker owns browser/e2e execution at a time. Other workers may run unit/lint/typecheck. Queue UI functional checks to avoid spawning many visible sessions.

## Browser / UI Rules

- Headless by default.
- Visible browser only when debugging visual behavior, SSO/CDP is required, or the user explicitly asks.
- Prefer programmatic probes: Playwright locators, API checks, DOM text extraction, screenshot assertions, or CLI commands.
- UI-reachable behavior must be exercised through the rendered UI route/control that a user uses. API calls, backend logs, unit tests, direct state inspection, calling a React/Vue/Svelte component method, invoking a button handler directly, or mocking the request the UI would send are supporting evidence only.
- A passing UI functional test must include at least one real navigation/open, one real user interaction when the feature is interactive, and one observable rendered assertion after the action. For display-only changes, assert the rendered screen state directly.
- Map every changed UI file and every UI-visible acceptance criterion to at least one route, screen, or interaction. If a changed component appears on multiple important routes, test the affected route set, not just one convenient page.
- Save screenshots as evidence for UI functional checks:
  - one baseline/pass screenshot per tested page or major workflow state;
  - mandatory screenshot for each failure state;
  - responsive screenshots only for deep tier or layout-sensitive changes.
- Store evidence under `.kb/qa-screenshots/` or the repo's configured QA artifact path.
- Keep screenshots until `kb-finalize` cleanup. Do not keep unlimited traces/videos unless they explain a failure.

## Script Rule

If a functional check will be repeated, turn it into a script or test:

- Playwright/Cypress test for UI flows.
- API smoke script for endpoints.
- CLI smoke script for commands.
- Small HTML/DOM extractor only when a full browser test is overkill.

Add the command to `docs/context/operations/testing.md` and make `kb-check` able to discover or call it.

## Output

Report:

- selected `test_level`;
- functional risk level: none, narrow, broad, full;
- tests audited;
- weak tests found;
- functional checks added or run;
- screenshot evidence path for UI checks;
- command/results with command or test file path, exit code, timestamp, and log/trace/artifact path;
- remaining manual-only verification and why.

For `functional-browser`, proof must include the executed browser assertion file or project test path plus exit code. A screenshot-only or prose-only claim is not functional proof.
