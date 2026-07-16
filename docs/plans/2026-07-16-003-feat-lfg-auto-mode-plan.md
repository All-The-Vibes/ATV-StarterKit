# Plan: T8 — `/lfg auto` mode (autoplan 6 Decision Principles)

- **Status:** Ready to implement (TDD)
- **Type:** feat
- **Date:** 2026-07-16
- **Branch:** fresh `feat/lfg-auto-mode` (independent of T7)
- **Origin:** Follow-up **T8** from `docs/research/gstack-router-atv-plan.md` (§5
  "Filed separately", §8 NOT-in-scope). Port `/autoplan`'s **6 Decision
  Principles** into `/lfg` as an opt-in `auto` mode so the pipeline can run
  gate-to-gate without stopping to ask, while the default (gated) behavior is
  unchanged.
- **Port reference:** `~/.claude/skills/gstack/autoplan/SKILL.md` — the 6 Decision
  Principles + the Decision Classification (Mechanical / Taste / User-Challenge).
  Methodology only; not the file/execute plumbing.

---

## Problem

`/lfg` is a sequential pipeline (`ce-plan → ce-work → ce-review → unslop → observe
→ learn → todo-resolve → test-browser → feature-video`) with a hard **STOP GATE**
between the early phases (`SKILL.md` steps 2–3). Those gates force a human round-
trip. For a user who wants true "just build it, don't ask me", the gates defeat
the one-command promise.

`/autoplan` already solved this for the *review* pipeline: it auto-answers every
intermediate question via **6 Decision Principles** + a **Decision
Classification** that decides when it's safe to auto-decide vs. when it must
surface. This plan brings that engine to `/lfg` as an **opt-in mode**, so the safe
default (ask at gates) is preserved.

```
/lfg <feature>            → gated pipeline (unchanged): STOP at each gate, ask
/lfg auto <feature>       → autonomous: auto-decide via 6 principles, surface only
                            Taste calls at the final gate, never auto-decide
                            User-Challenge (destructive/irreversible) calls
```

---

## Decisions (locked)

- **D1 — Opt-in `auto` keyword, default unchanged.** `/lfg auto <feature>` (or
  `/lfg <feature> --auto`) enables autonomous mode. Bare `/lfg <feature>` keeps
  today's gated behavior byte-for-byte. Zero regression risk to existing users.
- **D2 — Parse the mode in the skill, persist run-mode in lfg-state.** The `auto`
  flag is recorded in the run-state file (`lfg-state.js`) so a **resumed** `/lfg`
  continues in the same mode it started (resumability already exists — commit
  `855a214` lineage). This is the one code change: `lfg-state.js` gains an
  `--auto` flag on `init` and surfaces it in `status`.
- **D3 — 6 Decision Principles + Classification live in the skill body.** They are
  prose methodology (like the rest of `/lfg`), not code. Port from autoplan:
  (1) choose completeness, (2) boil lakes / fix blast radius, (3) pragmatic,
  (4) DRY, (5) explicit over clever, (6) bias toward action. Plus Classification:
  Mechanical → silent; Taste → decide + surface at final gate; User-Challenge →
  never auto-decide (stop and ask even in auto mode).
- **D4 — Irreversible gates are NEVER auto-passed.** Even in `auto` mode, the
  `ce-work`→PR push, `/land`, and any deploy/ship step remain User-Challenge
  class → they still stop. Auto mode removes the *planning/decision* round-trips,
  not the *destructive-action* confirmations. (Mirrors the router's GATE-2
  irreversible-target rule.)
- **D5 — Auto mode is surfaced, not silent.** At start, auto mode announces
  "Running /lfg in AUTO mode — I'll auto-decide plan/design/impl choices via the
  6 principles and only stop for destructive actions." At the final gate it emits
  a **decision log** (every Taste call it made + rationale) so the user can audit.

---

## Architecture & data flow

```
/lfg auto <feature>
     │  parse mode (auto|gated)
     ▼
lfg-state.js init --auto  ──▶  .atv/runs/<run-id>/state  (mode: auto persisted)
     │
     ▼  each phase:
   ┌─ Mechanical decision  → decide silently, proceed
   ├─ Taste decision       → decide via 6 principles, LOG it, proceed
   └─ User-Challenge       → STOP + ask (even in auto)  ← destructive/irreversible
     │
     ▼  final gate
   emit AUTO DECISION LOG (all Taste calls + rationale)  → user audits
```

The only executable change is `lfg-state.js` (+ its test). The decision engine is
skill-body prose the model follows.

---

## Acceptance criteria

| # | Criterion | How verified |
|---|-----------|--------------|
| A1 | `/lfg auto <feature>` and `/lfg <feature> --auto` both enable auto mode; bare `/lfg <feature>` is unchanged. | Skill-body parse instructions + Node test on `lfg-state.js` mode persistence. |
| A2 | `lfg-state.js init --auto` persists `mode: "auto"`; without it, `mode: "gated"` (default). | Node: `node --test` on `lfg-state.test.js` new cases. |
| A3 | `lfg-state.js status` reports the mode so a resumed run continues in the same mode. | Node test: init --auto, then status shows auto. |
| A4 | Skill body contains the 6 Decision Principles and the 3-way Decision Classification. | Go/grep content test on both `/lfg` copies. |
| A5 | Skill body states irreversible/destructive steps (PR push, `/land`, deploy) STOP even in auto mode. | Content test greps the User-Challenge rule. |
| A6 | Auto mode announces at start and emits a decision log at the final gate. | Content test greps both directives. |
| A7 | Both `/lfg` copies (template + dogfood) stay in sync; `plugingen -check` clean. | `plugingen -check` + parity test. |
| A8 | All Go + Node suites green. | CI commands below. |
| A9 | **Screenshot evidence:** passing `node --test` on `lfg-state.test.js` and a live `/lfg auto` run showing the AUTO-mode announcement + decision log. | Manual, attached at validation. |

---

## Work breakdown (atomic, TDD — RED before GREEN each step)

- [ ] **T8.1 — RED: `lfg-state` mode persistence tests.**
  In `.github/hooks/scripts/tests/lfg-state.test.js`, add cases: `init --auto`
  writes `mode: "auto"`; `init` (no flag) writes `mode: "gated"`; `status`
  returns the persisted mode. Run `node --test …/lfg-state.test.js` → **RED**
  (flag unhandled, `mode` absent).
  - Desired output: assertion failures on missing `mode`.

- [ ] **T8.2 — GREEN: `lfg-state.js` `--auto` flag + `mode` field.**
  Add `--auto` parsing to the `init` command; default `mode: "gated"`. Include
  `mode` in the `status` output and the written state JSON. Keep the pure-fn +
  `require.main` guard convention (matches `atv-config.js`). Run tests → GREEN.
  - Ship the byte-identical template copy: `pkg/scaffold/templates/hooks/
    scripts/lfg-state.js`. Diff-guard test asserts identical (existing pattern
    from `atv-route-log` T9.7).

- [ ] **T8.3 — GREEN: skill-body mode parsing + resume.**
  Edit `templates/skills/lfg/SKILL.md` **and** `.github/skills/lfg/SKILL.md`:
  - Add a "Step 0 — Mode" section: parse `auto` keyword / `--auto` flag; pass
    `--auto` to `lfg-state.js init`; on resume, read `mode` from `status` and
    continue in that mode.
  - This is additive; the existing numbered steps 1–11 are unchanged for gated
    runs.

- [ ] **T8.4 — GREEN: port the 6 Decision Principles + Classification.**
  Add an "## Auto mode (when mode == auto)" section to both `/lfg` copies with:
  the 6 principles (verbatim intent from autoplan), the 3-way Classification, the
  D4 irreversible-STOP rule, the D5 announce + decision-log directives. Keep it
  tight (~40–55 lines) — it's the only substantive body growth.

- [ ] **T8.5 — GREEN: content guard tests.**
  Add a Go content test (or extend an existing skill-content test) asserting both
  `/lfg` copies contain: "6" principles markers, "Mechanical"/"Taste"/"User-
  Challenge", the irreversible-STOP line, and the decision-log directive. Run
  `go test ./pkg/scaffold/...` → GREEN.

- [ ] **T8.6 — GREEN: regenerate + parity.**
  `go run ./cmd/plugingen` (regenerates the `plugins/*` `/lfg` copies — lfg lives
  in `atv-pack-shipping` + `atv-everything` + `atv-skill-lfg`). `plugingen
  -check` clean. `parity_test.go` green (both trees already present).

- [ ] **T8.7 — Docs.**
  README `/lfg` section: document `auto` mode, what it auto-decides, and what it
  still stops for. CHANGELOG line under Unreleased.

---

## Test scenarios (delivered)

**Node (`lfg-state.test.js`, real temp-dir fs via existing env override):**

1. `init --auto` → state file has `mode: "auto"`.
2. `init` (no flag) → `mode: "gated"`.
3. `status` after `init --auto` → reports `auto`.
4. `status` after plain `init` → reports `gated`.
5. `--auto` combined with `--feature/--repo/--branch` → mode set AND other fields
   intact (no arg-parse regression).
6. Template/dogfood `lfg-state.js` copies byte-identical.

**Go (content + generation):**

7. Both `/lfg` copies contain the 6 principles + 3-way classification markers.
8. Both contain the irreversible-STOP rule and the decision-log directive.
9. `plugingen -check` clean (generated `/lfg` copies match).
10. `parity_test.go` green.

## How to run the tests

```
node --test .github/hooks/scripts/tests/lfg-state.test.js
diff .github/hooks/scripts/lfg-state.js pkg/scaffold/templates/hooks/scripts/lfg-state.js
go test ./pkg/scaffold/... ./pkg/plugingen/...
go run ./cmd/plugingen -check
```

## Validation & screenshots (goal requirement)

1. `node --test …/lfg-state.test.js` → screenshot green run.
2. `go test ./pkg/scaffold/... ./pkg/plugingen/...` + `plugingen -check` →
   screenshot green.
3. Live: `/lfg auto add a health-check endpoint` → screenshot the **AUTO mode
   announcement** (start) and, if run far enough, the **decision log** at a gate.
   If a full pipeline run is too heavy for validation, screenshot at minimum the
   Step-0 mode parse + announcement and the `lfg-state.js status` showing
   `mode: auto`.

Attach to the PR as A9 evidence.

---

## What already exists (reuse, do not rebuild)

- **`lfg-state.js` + resumability** — run-state persistence already exists; we add
  ONE field (`mode`) and ONE flag (`--auto`). Do not build a parallel state store.
- **autoplan 6 Decision Principles** — port the prose; do not reinvent the
  decision framework.
- **Router GATE-2 irreversible-target rule** — the "never auto-pass destructive
  actions" posture is already established in the `/atv` router; reuse the same
  boundary definition (PR push, `/land`, deploy/ship).
- **`plugingen` + `parity_test.go`** — regen + two-tree contract already enforced.
- **`atv-route-log` T9.7 diff-guard pattern** — reuse for the `lfg-state.js`
  template-copy identity test.

## NOT in scope (deferred, with rationale)

- **T7 `/investigate`** — separate plan (`…-002-…`).
- **Making `auto` the default** — explicitly rejected; gated is the safe default,
  auto is opt-in. Flipping the default is a product decision for later, backed by
  T9 route telemetry usage data.
- **Auto mode for `/slfg`** — the swarm variant has different concurrency/isolation
  concerns (AGENTS.md §4). Port to `/slfg` only after `/lfg auto` proves out.
- **A decision-log artifact file** — the decision log is emitted inline at the
  gate (D5); persisting it to disk is a follow-up only if audit-after-the-fact is
  needed.
- **Porting autoplan's phase-skip logic** — `/lfg`'s phases are fixed; no
  conditional skipping needed.

## Failure modes (new codepath)

| Failure | Test? | Handling | User sees |
|---|---|---|---|
| `--auto` flag unparsed → mode lost on resume | ✅ T8.1/T8.3 | `status` reads persisted mode | resumes gated (safe fallback) not silent-auto |
| Auto mode auto-passes a destructive step | ✅ T8.5 content guard | D4 User-Challenge rule STOPs it | user is asked before any push/land |
| Template/dogfood `lfg-state.js` drift | ✅ T8.2 diff-guard | CI fails | build-time only |
| `mode` field absent in old state file (pre-upgrade resume) | ✅ T8.1 (default case) | missing → treated as `gated` | safe default, no surprise auto-run |
| Auto decisions made silently with no audit trail | ✅ T8.5 (decision-log directive present) | D5 decision log at final gate | user can audit every Taste call |

Critical-gap check: the dangerous mode — auto-passing a destructive action — is
covered by BOTH a content guard (D4 rule present) AND the safe-default fallback
(missing/unknown mode → gated). No silent unhandled path.

## Dependency / parallelization

Single lane, sequential (T8.1→T8.7); all changes cluster on `/lfg` + `lfg-state`.
**No parallelization.** Fully independent of T7 (different skill, different files)
— T7 and T8 could run in parallel worktrees with zero conflict (T7 touches
`investigate/` + `packs.go` + bug fixtures; T8 touches `lfg/` + `lfg-state.js`).

## GSTACK REVIEW REPORT

| Review | Trigger | Why | Runs | Status | Findings |
|--------|---------|-----|------|--------|----------|
| CEO Review | `/plan-ceo-review` | Scope & strategy | 0 | — | inherited: parent plan filed T8 as fast-follow, not launch scope |
| Codex Review | `/codex review` | Independent 2nd opinion | 0 | — | pending (offer at implementation) |
| Eng Review | `/plan-eng-review` | Architecture & tests (required) | 1 | clean | One code change (`lfg-state.js` +1 field/flag), rest is skill-body prose; opt-in preserves default; destructive-STOP double-guarded. |
| Design Review | `/plan-design-review` | UI/UX gaps | 0 | n/a | no UI |
| DX Review | `/plan-devex-review` | Developer experience gaps | 0 | n/a | not run |

- **UNRESOLVED:** 0 — D1–D5 locked.
- **VERDICT:** ENG CLEARED — test-first (6 Node + 4 Go scenarios), opt-in with safe
  gated default, destructive actions double-guarded. Ready to implement. Start at
  T8.1 (RED). Independent of T7 — parallel-worktree safe.
