---
name: kb-troubleshoot
description: "Autonomous troubleshooting and self-correction loop. Use when the user says troubleshoot, debug this, figure out what's broken, self-correct, run the logs, use Playwright/browser, iterate until fixed, or asks the agent to diagnose failing app behavior without hand-holding."
argument-hint: "[symptom, failing command, URL, log, or broken behavior]"
---

# KB Troubleshoot

Diagnose broken behavior, reproduce it with agent-runnable evidence, fix it, and verify it without asking the user to drive normal testing.

This is the autonomous debugging lane. Use `kb-fix` for tiny known bugs; use `kb-troubleshoot` when the agent must discover the failure mode from logs, UI behavior, commands, browser evidence, or unclear symptoms.

## Contract

- Own the loop: observe -> reproduce -> localize -> fix -> verify -> repeat.
- Look before reasoning. Do not guess from model memory when the repo, running app, logs, browser, tests, docs, or external sources can answer the question.
- Treat model knowledge as stale by default for dependencies, frameworks, browser/runtime behavior, build tools, package managers, platform APIs, and third-party services. Go read current docs, issue trackers, changelogs, source, or known fixes when those facts affect the fix.
- Ask the user only for credentials/MFA, unavailable private input, risky/destructive action, production writes, subjective product intent, or a genuine choice with multiple reasonable outcomes.
- Prefer deterministic evidence over prose: commands, failing tests, browser assertions, console/network logs, traces, screenshots, API responses, service logs, and exit codes.
- When a bug has a repeatable sensor, prefer `kbcheck sense` + `kbcheck accept` so the fix proves RED-before-GREEN instead of latest-green only.
- For UI-reachable bugs, drive the rendered UI with Playwright, CDP, Agent Browser, or the repo's browser transport. Screenshots are evidence, not the pass/fail oracle.
- Do not stop after "I found the issue." Stop only when fixed and verified, honestly blocked, or explicitly told to stop.

## Workflow

1. **Map and frame**
   - Run `kb-map lookup <symptom>` unless recent repo context is already loaded.
   - State the symptom, expected behavior, affected surface, and first verification target in one compact note.
   - Check `docs/context/PROJECT.md`, `todo.md`, recent handoffs, and relevant `docs/solutions/` notes for known run/test/debug commands.

2. **Collect live evidence**
   - Run the narrowest relevant app command first: failing test, lint/typecheck/build, dev server, CLI command, API probe, or reproduction script.
   - For UI behavior, start or reuse the dev server, then use the best browser transport available:
     - CDP for authenticated/internal sessions.
     - Playwright for local/public UI checks and repeatable assertions.
     - Agent Browser when installed and suitable.
   - Capture console errors, failed network requests, page errors, terminal logs, server logs, screenshots/traces, and exact commands.
   - If an observable source exists, inspect it before forming the fix: DOM, console, network, server logs, stack traces, request/response payloads, database state, queue/job state, config, generated files, or test output.
   - If the error has a distinctive message, code, stack frame, package name, browser warning, or platform status, run targeted external research before editing. Prefer primary docs, upstream issues/PRs, changelogs, release notes, and source over blog posts.

3. **Reproduce before editing**
   - Build at least one agent-runnable reproduction signal before changing code.
   - If practical, write or select a compact `<check.json>` and record the failing state with `go run ./cmd/kbcheck sense --check <check.json> --trace .kb/trace.jsonl`.
   - If the bug is intermittent, run the reproduction enough times to identify frequency or conditions.
   - If no reproduction is possible after three strategies, record the attempted signals and switch to the closest deterministic probe instead of guessing.

4. **Diagnostic plan**
   - Before editing, write a compact diagnostic plan: reproduced signal, current evidence, 1-3 falsifiable hypotheses, likely files/boundaries to inspect, protected test/oracle files, and the exact verification command/browser/API/CLI probe that must pass.
   - This is not a `kb-plan` manifest. Escalate to `kb-plan` only when the fix becomes multi-slice, crosses several owning surfaces, or needs dependency ordering.
   - If the plan cannot name an executable verification target, gather more evidence before editing.

5. **Localize**
   - Form 1-3 falsifiable hypotheses.
   - Inspect the smallest code path that explains the evidence.
   - Prefer reading actual runtime boundaries over guessing: routes, event handlers, network calls, state transitions, persistence, worker/job logs, auth/session checks, and browser console/network output.
   - If a hypothesis depends on framework/library/runtime behavior and the agent is relying on memory, run `kb-research` or consult primary docs/source before editing. Do this especially for changed APIs, browser behavior, async timing, auth/session behavior, build tooling, package manager quirks, database semantics, or third-party integrations.
   - Search externally when a known fix is likely to exist: exact error text, dependency version mismatch, migration failure, browser console warning, failing Playwright locator behavior, build error, test runner failure, auth/session issue, deployment/runtime error, or package manager conflict.

6. **Fix**
   - Make the smallest coherent fix that addresses the reproduced failure.
   - If the fix expands into a feature or multi-slice change, route to `kb-plan` with execution intent instead of continuing as ad hoc debugging.
   - If QA/lint/browser checks fail from the attempted fix, invoke `kb-repair` with the failure report.
   - If the problem is a narrow known bug, `kb-fix` may own the inner fix loop.

7. **Verify and iterate**
   - Re-run the original reproduction.
   - If a proof-spine check was recorded, run `sense` again and require `go run ./cmd/kbcheck accept --check <check.json> --trace .kb/trace.jsonl` before closing.
   - Run the relevant regression checks from `kb-check`.
   - For UI bugs, run browser assertions through the rendered UI and capture console/network cleanliness after each interaction.
   - Compare failure signatures after each attempt:
     - same failure: hypothesis likely wrong;
     - fewer failures: progress, continue;
     - different failure: side effect or next layer, continue with evidence;
     - all checks pass: close.
   - At each failed iteration, identify what new evidence would distinguish the next hypothesis. Gather that evidence before editing again. If the next step would rely mainly on model knowledge, do a targeted `kb-research` pass first.
   - If two fix attempts fail or the failure signature changes into unfamiliar framework/tooling behavior, research the current external behavior before the next edit.
   - Iterate up to five fix attempts. Continue past five only when every attempt produced measurable progress and no risky scope expansion.

## No-Guessing Rule

Before every edit, the agent must be able to complete this sentence:

```text
I am changing <file/behavior> because <observed evidence> shows <specific cause>.
```

If the sentence depends on "probably", "usually", "I think", or remembered framework behavior, pause the edit path and gather evidence:

- inspect the live app or failing command output;
- read the local implementation and tests;
- check runtime logs, console output, network traces, stack traces, or persisted state;
- run a smaller probe that isolates the assumption;
- use `kb-research` for external/framework behavior and prefer primary docs or source.
- search exact error signatures and dependency versions for known fixes before inventing one.

8. **Close the loop**
   - Record the final proof: command/test/browser assertion, `kbcheck accept` result when used, exit status, artifact path, and timestamp when available.
   - Update `todo.md` or handoff only if work remains blocked or follow-up is real.
   - Run `kb-map refresh` when the fix changes durable behavior, commands, architecture, integration knowledge, or a sharp edge worth remembering.
   - If the fix teaches a reusable lesson, let `kb-complete`/`ce-compound` capture it when part of a manifest flow; otherwise note the learning target for later compounding.

## Stop Conditions

Stop and ask only when:

- login, MFA, credentials, private data, paid access, or external hardware is required;
- the next action is destructive, production-affecting, or irreversible;
- a product decision has multiple reasonable outcomes;
- verification needs subjective judgment the agent cannot measure;
- all available local/browser/test probes are blocked and the exact missing setup is known.

When blocked, return a compact troubleshooting packet:

```markdown
## Troubleshooting Blocker

Symptom:
Reproduction attempted:
Evidence gathered:
Hypotheses tested:
Changes attempted:
Current best theory:
Blocked on:
Exact human input needed:
```

## Integration

- **Uses:** `kb-map`, `kb-check`, `kb-functional-test`, `kb-qa`, `kb-repair`, `kb-fix`, browser transports, repo run/test commands.
- **Escalates to:** `kb-plan` when the fix becomes multi-slice; `kb-research` when external/framework behavior is uncertain; `kb-handoff` only when work must resume later.
- **Does not replace:** `kb-work`/`kb-finalize` for planned feature execution. It is for diagnosing and self-correcting broken behavior.
