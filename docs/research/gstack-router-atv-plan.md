# Research: G-Stack `/gstack` Router → ATV Starter Kit `/atv` Router

**Date:** 2026-07-15
**Source studied:** G-Stack plugin installed at `~/.claude/skills/gstack/` (Garry Tan's "Garry's Stack", repo `github.com/garrytan/gstack`). Studied the actual on-disk skill files, not web content.
**Goal:** Understand how `/gstack` routes between skills, then plan an ATV equivalent (`/atv` / "ATV Starter Kit") so users can invoke one command and be routed to the best workflow without naming skills.

---

## 1. What `/gstack` actually is

`/gstack` is a **pure router skill**, not a pipeline. Its whole job (its own words): *"This is the gstack router. Its one job is to send the request to the right skill."*

Key facts from `~/.claude/skills/gstack/SKILL.md` (generated from `SKILL.md.tmpl`):

- **Frontmatter is minimal:** `name: gstack`, `allowed-tools: [Bash, Read, AskUserQuestion]`, and `triggers: [gstack, which gstack skill, route this with gstack]`. It deliberately holds *no* Edit/Write — a router routes, it doesn't build.
- **`preamble-tier: 1`** — it's the lightest, first-loaded skill.
- The routing itself is **LLM-driven intent classification against a static table**, not code. The "## Route first" section is a flat list of `pattern → invoke /skill` rules. The model reads the user's request, matches a rule, and calls the **Skill tool** to invoke the target skill.

### The routing contract (the important part)

```
1. Browser/QA/screenshots/inspect-a-page  → invoke /browse   (special-cased first)
2. Otherwise match a routing rule below    → invoke that skill via Skill tool
3. Nothing matches                          → answer directly
```

The rule table is ~35 lines of `User asks X / describes Y → invoke /skill`. Examples:
- new idea / "is this worth building" → `/office-hours`
- "spec this out" / "file an issue" → `/spec`
- bug / "why is this broken" / "wtf" → `/investigate`
- "does this work" / "check the deploy" → `/qa`
- "look at my changes" → `/review`
- "ship it" / "send it" → `/ship`
- "review everything automatically" → `/autoplan`

The router's governing heuristic: **"When in doubt, invoke the skill. A false positive is cheaper than a false negative."** It biases toward routing into a structured workflow rather than answering ad-hoc.

### The one behavioral gate: `PROACTIVE`

A config flag (`gstack-config get proactive`, default `true`) decides routing aggressiveness:
- `PROACTIVE=true` → **invoke** the matched skill directly via the Skill tool. Do not answer inline when a skill exists.
- `PROACTIVE=false` → do not auto-invoke; at most ask *"I think /x might help — want me to run it?"*

This is the single most important design lever: it lets the same router be either an auto-dispatcher or a suggest-only assistant, per user preference, persisted across sessions.

### Discovery index: `llms.txt`

`~/.claude/skills/gstack/gstack/llms.txt` is a machine-readable catalog of every skill (one line each: `[/skill](path): one-line description`) plus every `browse`/`design` sub-command. Its stated purpose: *"index every capability so agents can discover and invoke them without crawling individual SKILL.md files."* It's auto-generated (`bun run gen:skill-docs`). This is how the router (and any agent) knows the menu without reading 50 files.

### The rest of the router file is preamble, not routing

~530 of the 600 lines are shared boilerplate every gstack skill carries: update-check, session tracking, telemetry opt-in, first-run onboarding, artifacts/gbrain sync, model-behavior patch, voice rules, completion-status protocol. **None of this is required for routing.** For an ATV port, the routing logic is ~65 lines; the rest is gstack's operational envelope that ATV already handles its own way.

---

## 2. The complementary pattern: `/autoplan` (auto-pipeline)

The user's request mixes two ideas ("route to the right skill" **and** "route through all processes in order to the best implementation"). G-Stack implements these as **two separate skills**:

| | `/gstack` (router) | `/autoplan` (pipeline) |
|---|---|---|
| Job | classify intent → invoke **one** skill | run **many** skills **in sequence** with auto-decisions |
| Mechanism | LLM matches a rule table, calls Skill tool | Reads sub-skill SKILL.md files from disk, executes each phase's methodology inline |
| Decisions | user still drives the invoked skill | auto-answers every intermediate question via **6 Decision Principles** |
| Order | n/a (single hop) | strict `CEO → Design → Eng → DX`, no parallelism, phase gates between |

`/autoplan`'s mechanism (from `autoplan/SKILL.md`, 1852 lines):
1. **Load skill files from disk** with the Read tool (`plan-ceo-review`, `plan-design-review`, `plan-eng-review`, `plan-devex-review`) — conditionally skipping design/DX if no UI/DX scope.
2. Follow each loaded skill's methodology, but **skip a defined "section skip list"** (preamble, telemetry, AskUserQuestion format, etc. — already handled by the orchestrator).
3. **Auto-decide** every question using **6 Decision Principles**: (1) choose completeness, (2) boil lakes / fix blast radius, (3) pragmatic, (4) DRY, (5) explicit over clever, (6) bias toward action. Plus a **Decision Classification** (Mechanical → silent; Taste → decide + surface at final gate; User Challenge → never auto-decide).
4. **Sequential execution is mandatory** — each phase fully completes and writes its outputs before the next starts, with a phase-transition summary between.

**This is the key insight for ATV:** "route through all processes to the best implementation" = the `/autoplan` pattern (sequential skill chaining with auto-decisions), while "route to the right skill without specifying it" = the `/gstack` pattern (intent router). They are different skills and ATV likely wants both faces behind one entry point.

---

## 3. What ATV already has (gap analysis)

ATV Starter Kit ships skills as `plugins/atv-skill-<name>/skills/<name>/SKILL.md`, each its own plugin with a `plugin.json` manifest. Relevant existing skills:

- **Pipelines already exist:** `/lfg` and `/slfg` are exactly the `/autoplan` pattern — hardcoded sequential chains (`ce-plan → ce-work → ce-review → unslop → observe → learn → todo-resolve → test-browser → feature-video`), with STOP gates between steps and a resumable run-state helper (`lfg-state.js`). `/slfg` is the swarm-parallel variant.
- **Lifecycle bookends:** `/takeoff` (session kickoff) and `/land` (session close).
- **Rich skill catalog:** `ce-plan`, `ce-work`, `ce-review`, `ce-brainstorm`, `ce-ideate`, `brainstorming`, `deepen-plan`, `document-review`, `unslop`, `observe`, `learn`, `instincts`, `evolve`, `autoresearch`, `test-browser`, `feature-video`, `atv-security`, `meme-iq`, `ralph-loop`, `resolve_todo_parallel`, `atv-doctor`, `atv-update`, `setup`.

**The gap:** ATV has **pipelines** (`/lfg`, `/slfg`) but **no intent router**. There's no `/atv` that a user hits with a vague request and gets dispatched to the right skill (or pipeline). That's exactly the `/gstack` role.

---

## 4. Recommended design for `/atv`

Build **one router skill** modeled on `/gstack`, that dispatches to ATV's existing skills — and, critically, treats the `/lfg` and `/slfg` pipelines as first-class routing targets for "build me X" requests.

### 4.1 Structure
- New plugin: `plugins/atv-skill-atv/skills/atv/SKILL.md` (mirrors existing ATV plugin layout + `plugin.json`).
- Frontmatter: `name: atv`, `allowed-tools: [Bash, Read, AskUserQuestion]`, `disable-model-invocation: false`, triggers: `atv`, `atv starter kit`, `which atv skill`, `route this`.
- Keep it **thin** — routing table + one proactivity gate. Do not copy gstack's 500-line operational preamble; ATV has its own conventions (Backlog.md, `.atv/runs`, remember).

### 4.2 The routing table (intent → target)
Map ATV's catalog the way gstack maps its own. Draft:

| User intent | Route to |
|---|---|
| "build/implement feature X", "just do it", full autonomous work | `/lfg` (or `/slfg` if they say swarm/parallel/fast) |
| new idea, "is this worth building", pitch, explore | `/ce-brainstorm` or `/brainstorming` |
| generate + critique improvement ideas | `/ce-ideate` |
| "plan this", turn requirements into a plan | `/ce-plan` |
| deepen/enrich an existing plan | `/deepen-plan` |
| review a plan/requirements doc | `/document-review` |
| do the implementation work on an existing plan | `/ce-work` |
| review code / "look at my diff" | `/ce-review` |
| strip AI slop | `/unslop` |
| bug / "why is this broken" / "this doesn't work" | **GAP** — no dedicated debug skill yet. Interim: measurable/repro bug → `/autoresearch`; vague/no-repro → `/ce-work`. Follow-up: build custom `/investigate`-style skill (see D4). |
| browser QA / "does this page work" | `/test-browser` |
| research/experiment loop | `/autoresearch` |
| security / "is this secure" | `/atv-security` |
| capture a solved problem / learnings | `/ce-compound`, `/learn`, `/observe` |
| "what should I work on" / session start | `/takeoff` |
| "land the plane" / wrap up | `/land` |
| record a feature demo | `/feature-video` |
| meme | `/meme-iq` |
| health / install check | `/atv-doctor` |

### 4.3 The proactivity gate
Copy gstack's `PROACTIVE` idea. Two options for storage:
- Reuse an existing ATV config surface if one exists (check `setup` skill / `.atv/`), or
- Add a small `atv-config get/set proactive` shim.
Default `true` = auto-invoke; `false` = suggest-only ("I think /x helps — run it?").

### 4.4 A discovery index
Generate an ATV `llms.txt` (or reuse `marketplace.json`/`DOCS.md`) so the router — and any agent — has a one-line-per-skill menu to match against and to keep in sync as skills are added. gstack auto-generates this; ATV should too, from the `plugin.json` descriptions it already has.

### 4.5 Router vs pipeline: keep them distinct
- `/atv` = **router** (single-hop intent dispatch). This is the new build.
- `/lfg` `/slfg` = **pipelines** (already exist). The router *routes into* them for "build X" requests.
- Optionally, later, add an `autoplan`-style **auto-decision layer** to `/lfg` (the 6 Decision Principles) so it can run gate-to-gate without stopping to ask — but that's an enhancement to `/lfg`, not part of the router.

---

## 5. Implementation plan (phased) — REVISED by /plan-eng-review 2026-07-15

**Phase 0 — De-risk spike (BLOCKING, do first)**
0. **Verify the invocation model.** Confirm whether a Skill-tool call can invoke a skill with `disable-model-invocation: true` (which `/lfg` and `/slfg` both have). Test in Claude Code AND Copilot CLI. Outcome decides the handoff design:
   - Works → router invokes `/lfg` directly (current design holds).
   - Blocked → router emits "Run /lfg <feature>" and stops (fallback; the "one command all the way" promise is reduced to a one-hop suggestion for build requests). Do NOT remove the flag from `/lfg`/`/slfg` — it's there to stop accidental full-pipeline auto-fires.

**Phase 1 — Routing data + generator (was Phase 2; now precedes the skill)**
1. Add a `routing_triggers` array to each skill's `plugin.json` (~20 files). Each entry: intent phrases that should route to that skill.
2. Write the generator: emits the routing table + an `llms.txt`-style catalog from every `plugins/*/plugin.json` (`name`, `description`, `routing_triggers`). Skips malformed manifests without failing the whole run. Wire into the release/update flow.

**Phase 2 — Router skill (core deliverable)**
3. Scaffold `plugins/atv-skill-atv/` with `plugin.json` + `skills/atv/SKILL.md` (layout from `atv-skill-takeoff`).
4. Write the `## Route first` section: browser/QA special-case, the **generated** routing table, the "when in doubt, invoke" bias, and the D2 one-line route announcement.
5. Router reads `PROACTIVE` (hardcoded default `true` in the skill; override from `~/.atv/config.json` if present). `PROACTIVE=false` → suggest-only, never auto-invoke.
6. Build handoff per Phase 0 outcome (invoke `/lfg`/`/slfg`, or emit-and-stop).

**Phase 3 — Config shim**
7. `atv-config get/set` backed by `~/.atv/config.json` (HOME dir, gstack-style — NOT in-repo `.atv/`, which is gitignored runtime state and less portable across Copilot/VS Code). File is override-only; skill owns the default.

**Phase 4 — Tests (first-class, not deferred)**
8. Routing fixtures: ~20 prompts → expected skill; harness runs each through the router and asserts the pick. Guards against routing regressions when skills are added.
9. Unit tests: `atv-config` get/set (incl. default-when-absent, set-persists, PROACTIVE=false suppresses invoke); generator (emits line per manifest, skips malformed).

**Phase 5 — Register + document**
10. Add `atv-skill-atv` to `marketplace.json` (currently ships only `atv-everything`) and the relevant pack(s).
11. Document `/atv` in `README.md` / `DOCS.md` as the one-command entry point.

**Filed separately (NOT in this plan):** porting `/autoplan`'s 6 Decision Principles into `/lfg` as an `auto` mode. It's a feature on a different skill; out of scope here.

---

## 6. Open questions — RESOLVED (office-hours, 2026-07-15)

Worked through as a proper one-question-at-a-time office-hours diagnostic (Builder mode: open-source tooling, not a startup). Four forcing decisions:

**D1 — Scope of the router: FULL PARITY with gstack.** `/atv` routes the whole ATV catalog (~20 skills) from day one, not a narrow wedge. Guardrail added in dialogue: the routing table must be **generated/maintained from each skill's own `plugin.json` description** (like gstack regenerates from `llms.txt`), not hand-tuned prose that rots. Breadth is cheap (dispatch table + "when in doubt, invoke"); the real cost is *mapping quality*, so keep the table a living, generated file.

**D2 — Route confidence: ALWAYS INVOKE, NEVER CONFIRM** (gstack's literal posture). No blocking confirmation prompt on every call — that would train users that `/atv` is slow and push them back to typing skills directly. Mitigation for the newcomer/mis-route risk: **announce the route in one line as it invokes** ("Routing to /X — say 'no, I meant …' to redirect"). Zero round-trip, but transparent and correctable. Net: invoke + announce, no gate.

**D3 — Config home: NEW `.atv/config.json` + `atv-config get/set` shim.** Load-bearing, not speculative: the router needs a *persisted* preference so "stop auto-routing" survives across sessions; hardcoding would force users to edit skill files to turn routing off. The `PROACTIVE` flag (default `true`) lives here; future router prefs get a home too.

**D4 — Debugging: DO NOT retrofit `/autoresearch`. Build a custom debug skill later; leave as an identified GAP.** Reversal of the earlier retrofit idea, and the right call. `/autoresearch`'s engine is "improve a measurable metric" — a clean bug (`test X fails`) has a metric, but a vague bug ("the page feels broken") has none, and the loop spins aimlessly. Retrofitting would dilute autoresearch's sharp identity and misbehave on no-repro bugs.
   - **Interim routing** (until the custom skill exists): `/atv` sends clear/measurable/repro-backed bugs to `/autoresearch`, vague/no-repro bugs to `/ce-work`. Both are stopgaps, explicitly noted as not the real answer.
   - **KNOWN GAP → follow-up task:** build a dedicated ATV `/investigate`-style debugging skill (systematic root-cause, repro-first, no metric requirement). Port gstack's `/investigate` methodology as the reference.

**D5 (carried, unchanged) — Router vs `/lfg` overlap: route INTO `/lfg`.** `/atv` stays single-responsibility; "build feature X" → `/lfg` (or `/slfg` for swarm/parallel). `/atv` does not absorb the pipeline.

---

## 7. Sources (all local, on-disk)
- `~/.claude/skills/gstack/SKILL.md` — the router (generated from `SKILL.md.tmpl`)
- `~/.claude/skills/gstack/SKILL.md.tmpl` — router template (routing table lives here)
- `~/.claude/skills/gstack/gstack/llms.txt` — capability discovery index
- `~/.claude/skills/gstack/autoplan/SKILL.md` — the sequential auto-pipeline + 6 Decision Principles
- ATV repo: `plugins/atv-skill-*/skills/*/SKILL.md`, `plugins/atv-everything/skills/lfg/SKILL.md`, `plugins/atv-skill-slfg/skills/slfg/SKILL.md`, `plugin.json` manifests, `marketplace.json`

---

## 8. Eng-review outputs (2026-07-15)

### NOT in scope (deferred, with rationale)
- **Autoplan 6-principles ported into `/lfg`** — separate feature on a different skill; file as its own task.
- **Custom `/investigate`-style debug skill** — the D4 gap; real answer to bug routing, but its own build. Interim routing stands.
- **Removing `disable-model-invocation` from `/lfg`/`/slfg`** — rejected; the flag prevents accidental full-pipeline auto-fires.
- **Per-project router prefs** — config is a user preference (`~/.atv/`), not per-project. Revisit only if a real per-repo need appears.

### What already exists (reuse, don't rebuild)
- `/lfg`, `/slfg` pipelines — router routes INTO them, does not reimplement.
- `plugin.json` per skill — source for the generated routing table (after adding `routing_triggers`).
- `marketplace.json` — extend, don't replace (today it ships only `atv-everything`).
- `atv-skill-takeoff` layout — copy as the scaffold for `atv-skill-atv`.

### Dependency / parallelization
| Phase | Touches | Depends on |
|---|---|---|
| P0 spike | none (investigation) | — |
| P1 triggers+generator | all `plugin.json`, new generator script | P0 (design certainty helps but not strict) |
| P2 router skill | `plugins/atv-skill-atv/` | P1 (needs generated table), P0 (handoff design) |
| P3 config shim | new `atv-config` script | — (independent) |
| P4 tests | fixtures + unit tests | P2, P3 |
| P5 register+docs | `marketplace.json`, README/DOCS | P2 |

**Parallel lanes:** Lane A: P0 → P1 → P2 (sequential, core spine). Lane B: P3 config shim (independent, run anytime before P4). Merge, then P4 tests, then P5. Conflict flag: none — P3 touches only new files.

### Failure modes (new codepaths)
- **Router invokes a `disable-model-invocation` skill and it silently no-ops** → build fails the headline path. Covered by P0 spike + a fixtures test. NOT silent if the spike is done first.
- **`~/.atv/config.json` unreadable/corrupt** → config shim must fall back to hardcoded default, not crash. Needs a unit test (malformed-file case).
- **Generator hits a malformed `plugin.json`** → must skip + warn, not abort the whole table. Needs a unit test. **Critical gap if unhandled** (one bad manifest breaks all routing).
- **Ambiguous intent ("the review looks off")** routes to the wrong review skill → mitigated by D2 announce-and-redirect; fixtures should include collision cases.

### Implementation Tasks
Synthesized from findings above. P1 blocks ship.

- [ ] **T1 (P1, human: ~30min / CC: ~10min)** — spike — Verify Skill-tool can invoke `disable-model-invocation:true` skills (Claude Code + Copilot). Decides handoff design.
  - Surfaced by: Architecture Finding 1
  - Verify: attempt `/atv` → `/lfg` handoff in both hosts
- [ ] **T2 (P1, human: ~3h / CC: ~20min)** — routing — Add `routing_triggers` to ~20 `plugin.json` + write the table/catalog generator (skips malformed).
  - Surfaced by: Code Quality Finding 3 (DRY)
  - Files: `plugins/*/plugin.json`, new generator script
  - Verify: generator unit test (emits per manifest, skips malformed)
- [ ] **T3 (P1, human: ~2h / CC: ~20min)** — router — Scaffold `atv-skill-atv` + `## Route first` (generated table, invoke+announce, PROACTIVE gate, handoff per T1).
  - Files: `plugins/atv-skill-atv/**`
- [ ] **T4 (P1, human: ~1h / CC: ~15min)** — config — `atv-config get/set` over `~/.atv/config.json`; default-when-absent; PROACTIVE=false suppresses invoke.
  - Surfaced by: Architecture Finding 2
- [ ] **T5 (P1, human: ~4h / CC: ~30min)** — tests — Routing fixtures + harness (~20 prompts → expected skill, incl. collision cases) + unit tests for config shim and generator.
  - Surfaced by: Test Review (0/10 behaviors covered)
- [ ] **T6 (P2, human: ~1h / CC: ~15min)** — release — Register `atv-skill-atv` in `marketplace.json` + relevant pack; document `/atv` in README/DOCS.
- [ ] **T7 (P3, follow-up)** — debug skill — Build the custom `/investigate`-style skill (D4 gap).
- [ ] **T8 (P3, follow-up)** — pipeline — Port autoplan 6 principles into `/lfg` `auto` mode.
- [ ] **T9 (P2, human: ~1h / CC: ~10min)** — telemetry — Router appends one line per route to `~/.atv/analytics/routes.jsonl` in an **OTel-shaped** schema (`name`, `attributes`, `timestamp`). Logs **intent category + routed-to skill ONLY, never raw request text** (secret/PII safety). No otel dependency; forwardable by anyone with a collector.
  - Surfaced by: CEO E1 + Security §3
- [ ] **T10 (P2, human: ~15min / CC: ~5min)** — copy — Bug-intent route to `/ce-work` carries a one-line "debugging skill is provisional / coming" note until T7 lands.
  - Surfaced by: CEO E3

## 9. CEO-review outputs (2026-07-15)

**Mode:** SELECTIVE EXPANSION. Approach A (pure router) confirmed over B (router+auto-pipeline, re-adds cut scope) and C (thin dispatcher, undoes eng-review quality).

**Premise (held, with a flag):** the router solves a real discovery tax (~20 skills, no newcomer front door), but the size of that pain is unmeasured — hence E1 telemetry, to replace the proxy-metric guess with data.

### Scope decisions (cherry-pick ceremony)
| # | Proposal | Effort | Decision | Reasoning |
|---|----------|--------|----------|-----------|
| E1 | Route telemetry, OTel-shaped local JSONL | CC ~10min | **ACCEPTED** (→ T9) | Cheapest way to learn real usage + mis-routes; attacks proxy-metric risk. No otel dep. |
| E2 | `--explain`/dry-run mode | CC ~10min | **SKIPPED** | Redundant with D2 announce-as-it-routes. → NOT in scope. |
| E3 | Ship router now vs build debug skill first | — | **ACCEPTED: ship now** (→ T10) | Bias to action; gap stays honest/visible via provisional label. |

### Security (§3)
- **Telemetry PII/secret leak** → resolved: log intent category + routed-to skill only, never raw request text. Baked into T9.
- Config/telemetry files live in `~/.atv/` (user home, 0600-appropriate); no new network surface.

### Trajectory (§10)
- Reversibility **5/5** — additive skill, zero blast radius to delete.
- The generated table (T2) + telemetry (T9) are exactly the foundation that makes a future evolution to Approach B (router-as-auto-pipeline, T8) cheap. Good positioning without paying for it now.

### NOT in scope (CEO additions)
- `--explain`/dry-run mode (E2) — redundant with route announcement.
- Full OpenTelemetry SDK — spends an innovation token on plumbing; local OTel-shaped JSONL gets 90% of value at 0 dependencies.
- Approach B auto-pipeline (T8) and custom debug skill (T7) — real, but fast-follows, not launch scope.

## 10. /autoplan outputs (2026-07-15) — independent dual-voice review

Ran /autoplan (CEO + Eng + DX phases; Design skipped, no UI). Dispatched independent Claude subagents (fresh context) + Codex outside voice per phase. Mode: SELECTIVE EXPANSION. Premise gate: confirmed. This pass caught **three verified critical bugs the interactive CEO+Eng reviews missed**, plus DX safety gaps.

### Consensus — verified critical findings (fix before build)
All three independently surfaced AND verified against the repo:

**C1 — T2 reads the wrong data source.** `plugins/*/plugin.json` `description` is install boilerplate ("single-skill plugin (granular install)…") with zero intent signal. The intent-rich text is in each skill's **`SKILL.md` frontmatter `description`** ("Transform feature descriptions… plan this"). Confirmed: eng-subagent + Codex + repo grep.

**C2 — T2 edits a generated, reset-on-build layer.** `plugins/` is generated from `pkg/scaffold/templates/skills/` by `pkg/plugingen/generate.go` (verified: `generate.go:41-61`). Hand-editing `plugins/*/plugin.json` gets **wiped on next `Generate`**. Also `routing_triggers` isn't in the `PluginManifest` struct — a custom field risks CLI/source-install compat. Confirmed: Codex + repo.

**C3 — Skills are triplicated.** Every skill exists 3× (`atv-skill-*/`, `atv-pack-*/`, `atv-everything/`). A generator walking `plugins/*` emits 3 conflicting rows per skill. Confirmed: eng-subagent + Codex + repo (`find` shows lfg×3, ce-plan×3).

### GATE decisions (final approval)
- **GATE 1 → Fix T2 (accepted).** Rewrite T2: generator reads `pkg/scaffold/templates/skills/*/SKILL.md` frontmatter descriptions (authoritative, intent-rich), **dedupes by skill name**, emits a single canonical catalog, and does NOT touch generated `plugins/*`. **Drop the `routing_triggers` plugin.json idea entirely.** Router skill also lands in `templates/skills/atv/` so it flows into `atv-everything` (matches the 3× layout).
- **GATE 2 → Add DX safety + escape hatches (accepted).** Add to the router: (a) **no-match floor** — below-confidence → answer directly + show catalog, never force-route (restores gstack's rule-3 that D2 dropped); (b) **confirm-gate for irreversible targets** (`/lfg`, `/slfg`, `/land`, `/ship`-likes) — announce-then-redirect is too late once a pipeline acts; (c) **force-skill syntax** `/atv @<skill> <args>` (bypass classification + mis-route recovery); (d) **in-chat toggle** `/atv off|on|suggest` persisting to `~/.atv/config.json`.
- **GATE 3 → Two-layer test split (accepted).** T5 becomes: (1) deterministic unit tests for the generator + config shim (real coverage, no LLM); (2) a small live-model smoke set accepting nondeterminism (top-2 OK), including **no-match, collision, and PROACTIVE=false** cases. Stop asserting exact LLM routing picks.
- **CHALLENGE → Full build, premise stands (accepted).** Independent CEO voice argued build-before-evidence; user confirmed premise and elected full build now. Recorded, not gated.

### Also-raised (folded into existing tasks, not new decisions)
- Spike (T1) likely returns "blocked" (both eng voices) → design emit-and-stop as the **primary** handoff, upgrade only if spike surprises.
- Bare `/atv` (no args) → print catalog menu (folds into no-match floor + T6 docs).
- `/atv` self-advertisement via `/takeoff` + README first line (T6 docs).
- Install topology: router must ship in the **default/full** install or detect installed skills, else it routes to skills the user doesn't have (T6 + GATE1 canonical-source fix).

### Revised/added tasks
- **T2 (REWRITTEN, P1)** — generator reads `templates/skills/*/SKILL.md` frontmatter, dedupes by name, single canonical catalog; no plugin.json edits, no `routing_triggers`.
- **T3 (AMENDED, P1)** — router adds no-match floor, irreversible-target confirm gate, `/atv @skill` force syntax, `/atv off|on` in-chat toggle. Scaffold into `templates/skills/atv/` (not only `atv-skill-atv/`).
- **T5 (REWRITTEN, P1)** — two-layer tests: deterministic (generator+config) + live smoke (no-match, collision, PROACTIVE=false).
- **T6 (EXPANDED, P2)** — docs spec: 0-to-route quickstart, generated route table (human + llms.txt from one source), how-to-turn-off, how-to-force-skill, `/atv` vs `atv` binary distinction; ensure router in default install.
- **T11 (NEW, P2)** — install-topology rule: router ships in `atv-everything`/full scaffold OR does installed-skill detection before routing.

## GSTACK REVIEW REPORT

| Review | Trigger | Why | Runs | Status | Findings |
|--------|---------|-----|------|--------|----------|
| CEO Review | `/plan-ceo-review` | Scope & strategy | 2 | clean | Interactive: Approach A, 3 scope proposals. Autoplan dual-voice: premise challenge (held), build-before-evidence flagged |
| Codex Review | `/codex review` | Independent 2nd opinion | 1 (via autoplan) | issues_found | 3 verified critical T2 bugs (wrong field, generated layer, triplication) + test-flakiness + install-topology |
| Eng Review | `/plan-eng-review` | Architecture & tests (required) | 2 | issues_resolved | Interactive: 2 P1 arch + 1 DRY. Autoplan: 3 critical T2 data bugs, test strategy, spike-likely-blocked — all fixed at gate |
| Design Review | `/plan-design-review` | UI/UX gaps | 0 | n/a | no UI |
| DX Review | `/plan-devex-review` | Developer experience gaps | 1 (via autoplan) | issues_resolved | Dropped no-match floor, unsafe redirect on irreversible targets, missing off/force hatches — all added at GATE 2 |

- **CROSS-MODEL:** Eng-subagent + Codex independently confirmed the T2 data-source and triplication bugs; both verified against the repo. High-confidence signal.
- **UNRESOLVED:** 0 — all gate decisions answered.
- **VERDICT:** CEO + ENG + DX CLEARED (autoplan, via dual-voice). Plan revised: T2 rewritten (correct source), T3 gains DX safety + escape hatches, T5 two-layer tests, T6 expanded docs, T11 install-topology added. Ready to implement — start with T1 spike, then rewritten T2.
