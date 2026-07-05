---
name: kb-check
description: Deterministic verification harness for KB workflows. Use when code should be tested, linted, typechecked, built, security-checked, or validated by scripts instead of relying on LLM judgment; also use before kb-complete, kb-ship, or after kb-work slices.
argument-hint: "[optional scope, changed files, or command]"
---

# KB Check

Prefer executable truth over model judgment. If a script can check it, run the script.

## Rule

LLM review can find risks, but it does not prove behavior. A slice is not verified until deterministic checks pass or a clear reason is recorded.

When a slice declares protected oracles, deterministic proof must include the
oracle integrity check: the test, fixture, scorer, snapshot, schema, or contract
file used as the behavior target must still match the recorded SHA unless the
plan explicitly updated the oracle. This prevents the model from moving the
target after implementation starts.

## Check Sources

Discover commands from:

- `package.json`, `pnpm-workspace.yaml`, `turbo.json`, `nx.json`
- `pyproject.toml`, `requirements*.txt`, `pytest.ini`, `tox.ini`
- `.csproj`, `.sln`, `global.json`
- `Makefile`, `justfile`, `Taskfile.yml`
- repo docs: `README.md`, `AGENTS.md`, `docs/context/operations/testing.md`
- existing CI files under `.github/workflows/`

Prefer existing project commands over invented commands.

## Workflow

1. Run `go run ./cmd/kbcheck core --list` when present to inspect discovered commands.
2. Pick the narrowest commands that verify the touched behavior.
3. Run checks in this order when available: format/lint, typecheck/static analysis, unit tests, integration/e2e/browser checks, build/package, security/dependency audit.
4. Capture command, exit code, and relevant output.
5. If a check fails, route to `kb-repair` or `kb-fix`; do not ask the user to test normal app behavior.
6. If a check is missing, add a small reusable script or test when practical, then document it in `docs/context/operations/testing.md`.

In this portable skill bundle, the canonical local gate is:

```powershell
go run ./cmd/kbcheck core
```

`cmd/kbcheck` owns top-level orchestration. Existing PowerShell scripts may
still be individual validators until their behavior has separate Go parity
coverage.

For failure-first proof, use the local proof spine:

```powershell
go run ./cmd/kbcheck sense --check <check.json> --trace .kb/trace.jsonl
go run ./cmd/kbcheck accept --check <check.json> --trace .kb/trace.jsonl
go run ./cmd/kbcheck trace-verify --trace .kb/trace.jsonl
```

`accept` passes only when the same check was observed RED and then GREEN, the
trace chain is intact, and the current sensor run is still GREEN. It rejects
vacuous "already green" proof and tampered traces.

For learning changes that claim measurable improvement, use:

```powershell
go run ./cmd/kbcheck learning-adoption --result-path <results.json>
```

The adoption gate requires at least 20 samples, no right-to-wrong regressions,
no holdout string leakage, and either a two-case net gain or a 10 percentage
point gain before a learning rule may be promoted beyond local/scoped use.

## Functional Checks

Use `kb-functional-test` when a change touches user-visible behavior, API/CLI workflows, persistence, auth, streaming, integrations, or any bug that escaped unit tests.

For UI-reachable changes, the check must exercise the rendered UI. Do not substitute a backend/API call, component-handler invocation, mocked request, or direct state assertion for browser proof. If `.tsx`, `.jsx`, `.vue`, or `.svelte` files changed, expect `test_level: functional-browser` and run or call the UI/browser proof path.

Default timing:

- Slice: narrow functional check for the changed path.
- Manifest complete: broader smoke tests over changed workflows.
- Ship: full functional/e2e suite when practical.

Headless by default. Do not spawn visible browser windows from multiple workers; serialize browser/e2e checks.

## Script Rule

When the same manual verification would be repeated twice, create a script.

Good scripts accept scope arguments, print concise pass/fail output, exit nonzero on failure, avoid network unless needed, run in CI or from an agent session, and are documented in `docs/context/operations/testing.md`.

For protected-oracle work, prefer reusable SHA/manifest checks over manual
inspection. In this repo, `go run ./cmd/kbcheck skill-eval-manifest-selftest`
proves that tampering with a protected fixture/scorer manifest fails
deterministically.

## Output

Report commands run, pass/fail status, failures fixed or parked, checks added, and remaining manual-only verification with why it cannot be automated.

For every check, include machine proof: command or test file path, exit code, timestamp, and log/artifact path when available. Do not summarize as "tests pass" without the executable proof fields.
