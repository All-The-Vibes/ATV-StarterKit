# Plan: T7 — `/investigate` debug skill (systematic root-cause, repro-first)

- **Status:** Ready to implement (TDD)
- **Type:** feat
- **Date:** 2026-07-16
- **Branch:** `feat/atv-router` (or a fresh `feat/atv-investigate` — see Scope note)
- **Origin:** Follow-up **T7** / **D4 GAP** from `docs/research/gstack-router-atv-plan.md`.
  The `/atv` router currently sends **all** bug intents to `/ce-work` with a
  provisional "dedicated debug skill coming" label (router `SKILL.md`, commit
  `6bd34fb`). This plan builds that dedicated skill and flips the router's bug
  route to it.
- **Port reference:** `~/.claude/skills/gstack/investigate/SKILL.md` — methodology
  only (Phases 1–5 + Iron Law + Scope Lock). We do **not** port gstack's ~800-line
  operational envelope (telemetry, gbrain, artifacts sync, AskUserQuestion
  boilerplate); ATV skills are thin and carry their own conventions.

---

## Problem

ATV has no dedicated debugging skill. Bug intents ("why is this broken", "the test
suite fails", "this page doesn't work") route to `/ce-work`, whose engine is
"execute a plan" — it has no repro-first, root-cause-before-fix discipline. The
D4 decision (research doc §6) explicitly rejected retrofitting `/autoresearch`
(its engine is "improve a measurable metric" — a vague bug has no metric and the
loop spins). The real answer is a purpose-built skill.

### Why this is the right shape (not a retrofit)

```
 bug intent ──▶ /atv router ──▶ (today) /ce-work  ← "execute", no root-cause gate
                              └▶ (this plan) /investigate ← repro → root cause → fix → verify
```

`/investigate` enforces an **Iron Law**: no fix is written until the root cause is
named and a reproduction is confirmed. This is the discipline `/ce-work` lacks and
`/autoresearch` can't provide without a metric.

---

## Decisions (locked)

- **D1 — Methodology port, not envelope port.** Copy gstack `/investigate`'s
  Phases 1–5, Iron Law, and Scope Lock. Drop everything above `# Systematic
  Debugging` (line 822 in the reference). Target skill length: **~120–170 lines**,
  in line with ATV's other workflow skills (`brainstorming` 190, `observe` 133).
- **D2 — Skill lives in `templates/skills/investigate/` (authoritative) + dogfood
  mirror `.github/skills/investigate/`.** This is the standard ATV two-tree layout
  (`parity_test.go` enforces presence). `plugingen` then generates the
  `plugins/*` copies (triplication handled by the generator, not by hand).
- **D3 — Pack membership: `atv-pack-quality`.** Debugging is a code-quality
  workflow; it joins `ralph-loop` + `unslop` there (`packs.go`). Also flows into
  `atv-everything` automatically.
- **D4 — Router bug route flips to `/investigate`.** The provisional `/ce-work`
  route (research doc T10) is replaced. Bug fixtures in
  `pkg/plugingen/testdata/routing-fixtures.txt` change from `ce-work` →
  `investigate`. The "provisional / coming" note is removed from the router
  `SKILL.md`.
- **D5 — No new Node/Go helper.** `/investigate` is a pure methodology skill (like
  `brainstorming`). It writes no state file, needs no shim. The only *code* changes
  are the Go catalog/pack/fixtures wiring, which have existing test harnesses.

---

## Architecture & data flow

```
templates/skills/investigate/SKILL.md   ← authoritative skill body (NEW)
        │  (parity presence check: parity_test.go)
        ▼
.github/skills/investigate/SKILL.md      ← dogfood mirror (NEW, byte-copy at author time)
        │
        │  go run ./cmd/plugingen
        ▼
plugins/atv-skill-investigate/…          ← single-skill plugin (GENERATED)
plugins/atv-pack-quality/…               ← pack copy       (GENERATED)
plugins/atv-everything/…                 ← everything copy (GENERATED)
        │
        │  BuildRoutingCatalog(templates/skills) reads investigate/SKILL.md frontmatter
        ▼
templates/skills/atv/llms.txt            ← catalog gains an `investigate` line (REGEN)
        │
        ▼
/atv router  ──bug intent──▶  /investigate   (fixtures assert this)
```

### Skill frontmatter (the routing signal)

The `description:` field is the **intent-rich trigger text** the router matches
against (`routing.go` reads exactly this). It must carry the bug vocabulary:

```yaml
---
name: investigate
description: Systematic root-cause debugging — repro-first, fix the cause not the symptom. Use when something is broken, a test fails, "why is this broken", "wtf is happening", a stack trace, a regression, or any bug that needs diagnosis before a fix. Names the root cause and confirms a reproduction BEFORE writing any fix.
argument-hint: "[what's broken / the error / the failing test]"
---
```

---

## Acceptance criteria

| # | Criterion | How verified |
|---|-----------|--------------|
| A1 | `templates/skills/investigate/SKILL.md` exists, has valid frontmatter with `name: investigate` and a bug-intent-rich `description`. | Go: `parseSkillDescription` succeeds; unit assert on key phrases. |
| A2 | Skill body contains the Iron Law (no fix before root cause + repro) and 5 phases (Investigation → Pattern → Hypothesis → Implementation → Verification). | Go content test greps required section headers. |
| A3 | `.github/skills/investigate/SKILL.md` exists (dogfood mirror). | `parity_test.go` (presence) passes with `investigate` removed from any exclusion list. |
| A4 | `BuildRoutingCatalog` includes an `investigate` entry; committed `llms.txt` is regenerated and fresh. | `routing_sync_test.go` (`CommittedLLMsTxtIsFresh`, `IncludesKeyTargets` extended with `investigate`). |
| A5 | `investigate` is a member of `atv-pack-quality` and appears in `atv-everything`. | `generate_test.go` pack-membership assertions; `plugingen -check` clean. |
| A6 | Bug-intent fixtures route to `investigate`, not `ce-work`. | `routing_fixtures_test.go` (targets exist in catalog, no dangling). |
| A7 | Router `SKILL.md` bug row = `/investigate`; provisional note removed. | Go/grep test on both router copies; `plugingen -check` clean. |
| A8 | `go run ./cmd/plugingen -check` is clean; all Go + Node suites green. | CI commands below. |
| A9 | **Screenshot evidence:** passing `go test ./pkg/plugingen/... ./pkg/scaffold/...` and `plugingen -check` captured to a terminal screenshot. | Manual, attached at validation. |

---

## Work breakdown (atomic, TDD — RED before GREEN each step)

Ordering matters: tests that assert the skill exists must fail first (RED), then
the skill file makes them pass (GREEN).

- [ ] **T7.1 — RED: catalog + pack tests expect `investigate`.**
  Extend `routing_sync_test.go` `IncludesKeyTargets` to require `"investigate"`.
  Add a pack-membership assertion to `generate_test.go` that
  `atv-pack-quality.SkillNames` contains `investigate`. Run `go test
  ./pkg/plugingen/...` → **RED** (target missing, pack missing).
  - Desired output: failures naming `investigate` as absent.

- [ ] **T7.2 — GREEN: author the skill (both trees).**
  Write `templates/skills/investigate/SKILL.md` (port methodology per D1). Copy
  byte-identical to `.github/skills/investigate/SKILL.md`. Add `investigate` to
  `atv-pack-quality.SkillNames` in `packs.go` (after `ralph-loop`, before
  `unslop` — alpha order within pack is not required but keep it tidy).
  - Run `go test ./pkg/plugingen/...` → pack test GREEN; catalog test still RED
    until regen (T7.3).

- [ ] **T7.3 — GREEN: regenerate catalog + plugins.**
  `go run ./cmd/plugingen` (regenerates `plugins/`, marketplace, and the
  `llms.txt` catalog). Verify `templates/skills/atv/llms.txt` gained an
  `investigate` line. Run `go test ./pkg/plugingen/...` →
  `CommittedLLMsTxtIsFresh` GREEN.
  - Verify: `go run ./cmd/plugingen -check` exits 0.

- [ ] **T7.4 — RED→GREEN: flip bug fixtures.**
  In `pkg/plugingen/testdata/routing-fixtures.txt`, change the three bug lines
  (`the test suite fails…`, `why is this broken`, `this page doesn't work…`)
  from `ce-work` → `investigate`. Add 2 more bug fixtures (`wtf is happening
  with this stack trace | investigate`, `this used to work and now it 500s |
  investigate`). Run `routing_fixtures_test.go` → GREEN (targets now exist in
  catalog after T7.3). Before T7.3 this would be RED (dangling target).

- [ ] **T7.5 — GREEN: flip the router skill.**
  In `templates/skills/atv/SKILL.md` **and** `.github/skills/atv/SKILL.md`:
  change the routing-map bug row to `| bug / "why is this broken" / "this
  doesn't work" | /investigate |` and delete the "Bug routing is provisional…"
  paragraph. Regenerate plugins (`go run ./cmd/plugingen`). Run `plugingen
  -check` → clean.
  - Verify: grep both router copies for `provisional` → no match; for
    `investigate` → present.

- [ ] **T7.6 — GREEN: parity + exclusion cleanup.**
  Confirm `parity_test.go` passes with `investigate` present in both trees (it
  should need **no** exclusion entry since both copies exist). If a
  `pendingMirror`/`templateOnly` entry is needed, that's a smell — both trees
  are authored, so neither list should mention `investigate`.
  - Run `go test ./pkg/scaffold/...` → GREEN.

- [ ] **T7.7 — Docs.**
  Add `/investigate` to the README skill list and the `/atv` router section
  (bug route now goes to a real skill, not provisional). CHANGELOG line under
  Unreleased.

---

## Test scenarios (delivered)

**Go (deterministic, real fs — no mocks):**

1. `parseSkillDescription(investigate/SKILL.md)` returns a description containing
   `"root-cause"` / `"broken"` / `"test fails"` (intent signal present).
2. `BuildRoutingCatalog` output includes `{Name: "investigate"}` exactly once
   (dedupe holds despite triplication).
3. Committed `llms.txt` matches `RenderRoutingCatalog` (freshness).
4. `atv-pack-quality` plugin.json skills array contains `investigate`
   (`generate_test.go`).
5. `atv-everything` contains `investigate` (generated).
6. Every routing fixture target exists in the catalog — the flipped bug lines
   now resolve (no dangling `investigate`).
7. Skill body contains all 5 phase headers + the Iron Law line (content guard).
8. Both router copies contain `investigate` and NOT `provisional` (bug route
   flip guard).
9. `plugingen -check` clean (generated tree == committed).

**No Node tests** — `/investigate` ships no JS helper (D5). The Node suite is
unaffected; run it to confirm no regression.

## How to run the tests

```
go test ./pkg/plugingen/... ./pkg/scaffold/...
go run ./cmd/plugingen -check
node --test .github/hooks/scripts/tests/   # confirm no regression (unaffected)
```

## Validation & screenshots (goal requirement)

1. Run the full Go suite + `plugingen -check` → screenshot the green output.
2. Live smoke (Layer 2, manual, nondeterministic): invoke `/atv why is this
   broken, the login test fails` in Claude Code → confirm it announces
   `Routing to /investigate` → screenshot the routing announcement.
3. Invoke `/investigate the login test throws a null ref` → confirm it starts
   with root-cause investigation (asks for repro / reads the failing test)
   BEFORE proposing a fix → screenshot the first response showing the Iron Law
   in action.

Attach all three screenshots to the PR as acceptance evidence for A9.

---

## What already exists (reuse, do not rebuild)

- **`routing.go` `BuildRoutingCatalog`** — already reads `SKILL.md` frontmatter,
  dedupes by name, skips malformed. Adding a skill needs **zero** generator code
  changes; just author the file and regen.
- **`packs.go`** — declarative pack membership; add one string.
- **`routing_fixtures_test.go` + `testdata/routing-fixtures.txt`** — the fixture
  harness already asserts targets exist; just flip the bug lines.
- **`parity_test.go`** — enforces the two-tree contract; authoring both copies
  satisfies it with no new exclusion.
- **gstack `/investigate` SKILL.md** — the methodology source. Port Phases 1–5.

## NOT in scope (deferred, with rationale)

- **T8 `/lfg` auto mode** — separate plan (`…-003-…`). Different skill, different
  concern.
- **A `/investigate` state/run-artifact file** — the methodology is stateless;
  add a helper only if a real resumability need appears (YAGNI).
- **Brain/gbrain context load** (in gstack's version) — ATV has no gbrain
  dependency; dropped from the port.
- **Auto-invoking `/investigate` from a failing-test hook** — a nice future
  trigger, but a separate feature; the router route is enough for launch.
- **Porting gstack's telemetry/AskUserQuestion envelope** — ATV skills are thin.

## Failure modes (new codepath)

| Failure | Test? | Handling | User sees |
|---|---|---|---|
| Skill authored in only one tree | ✅ parity_test | CI fails until both exist | build-time only |
| `llms.txt` not regenerated after adding skill | ✅ CommittedLLMsTxtIsFresh | CI fails | build-time only |
| Bug fixture flipped before skill exists (dangling target) | ✅ routing_fixtures_test | RED until T7.3 regen | build-time only |
| Malformed frontmatter in new skill | ✅ parseSkillDescription | generator skips (never aborts whole catalog) | skill silently absent from menu — caught by IncludesKeyTargets |
| Router still says "provisional" after flip | ✅ T7.8 grep guard | CI fails | build-time only |

No failure mode is both silent AND unhandled: every wiring gap is caught by an
existing Go test harness at build time.

## Dependency / parallelization

Single lane, sequential: T7.1 (RED) → T7.2 (skill) → T7.3 (regen) → T7.4 (fixtures)
→ T7.5 (router flip) → T7.6 (parity) → T7.7 (docs). All touch the same skill/
catalog surface — **no parallelization opportunity**. Independent of T8.

## GSTACK REVIEW REPORT

| Review | Trigger | Why | Runs | Status | Findings |
|--------|---------|-----|------|--------|----------|
| CEO Review | `/plan-ceo-review` | Scope & strategy | 0 | — | inherited from parent plan D4 (build custom skill, don't retrofit) |
| Codex Review | `/codex review` | Independent 2nd opinion | 0 | — | pending (offer at implementation) |
| Eng Review | `/plan-eng-review` | Architecture & tests (required) | 1 | clean | Reuses existing catalog/pack/fixture/parity harnesses; only new artifact is the skill body. Zero new generator code. |
| Design Review | `/plan-design-review` | UI/UX gaps | 0 | n/a | no UI |
| DX Review | `/plan-devex-review` | Developer experience gaps | 0 | n/a | not run |

- **UNRESOLVED:** 0 — D1–D5 locked from parent research doc.
- **VERDICT:** ENG CLEARED — test-first (9 Go scenarios), reuses proven harnesses,
  bug route flip is fixture-guarded. Ready to implement. Start at T7.1 (RED).
