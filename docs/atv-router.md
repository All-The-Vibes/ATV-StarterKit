# `/atv` router — reference & live-smoke procedure

The `/atv` router is a one-command front door for the ATV Starter Kit. You
describe a request in plain language; the router classifies intent and dispatches
to the right skill or build pipeline, so you don't have to memorize the catalog.
It **routes** — it does not build (it holds no Edit/Write).

- Skill source: `pkg/scaffold/templates/skills/atv/` (installer template),
  mirrored to `.github/skills/atv/` (dogfood) and into the plugin copies.
- Capability menu: `llms.txt` next to `SKILL.md`, **generated from each skill's
  own `SKILL.md` frontmatter description** by the `pkg/plugingen` catalog builder,
  deduped by name.

## How routing works

1. **Control commands first** — `off` / `on` / `suggest` (toggle proactivity,
   persisted in `~/.atv/config.json` via the `atv-config` shim), bare `/atv`
   (print the menu), `@<skill> <args>` (force a skill, skip classification).
2. **Proactivity setting** — `true` invokes the matched skill, `false` only
   suggests, `suggest` names the target and asks first.
3. **Route** — browser/QA is special-cased to `/test-browser`; "build/implement"
   intents **emit** `/lfg` (or `/slfg` for swarm) rather than auto-firing them;
   everything else matches against `llms.txt`.
4. **No-match floor** — a below-confidence request is answered directly with the
   menu shown, never jammed into the nearest skill.

### Safety gates

- **Confirm gate (GATE 2)** — irreversible targets (`/lfg`, `/slfg`, `/land`,
  anything that ships/deploys/pushes) require an explicit confirmation first.
- **Emit-and-stop** — `/lfg` and `/slfg` carry `disable-model-invocation: true`.
  The router never removes or bypasses that guard; it emits the command for the
  user to run.

### Telemetry & privacy

Each routing decision is logged by `atv-route-log.js` as one OTel-shaped line to
`~/.atv/analytics/routes.jsonl`. The writer accepts a **fixed schema** only —
`--intent`, `--routed-to`, `--outcome` — with **no free-form field**, so the
user's raw request sentence is not recorded. The two classifier tokens are
additionally newline-stripped and capped at 64 characters. Writing is
best-effort: it never throws and never blocks a route.

> Note the honest bound: because `--intent`/`--routed-to` are caller-supplied
> tokens, up to 64 characters of caller text *can* be recorded. The guarantee is
> "no free-form request text + 64-char cap," not "nothing sensitive is ever
> representable." The router is instructed to pass a short classifier label
> (e.g. `code-review`, `build`, `security`), never raw request text.

### Graceful degradation

Some install surfaces (the `atv-everything` and `atv-pack-shipping` plugins, the
`atv-skill-atv` plugin) ship the skill **without** the `hooks/scripts/` tree. In
those installs `atv-config.js` and `atv-route-log.js` are absent. The router
treats both as best-effort: if the config shim is missing it uses the default
(`proactive = true`); if the telemetry writer is missing it skips logging. A
missing hook script never blocks a route.

## Layer-2 live-model smoke procedure

Routing quality has two test layers:

- **Layer 1 (deterministic, in CI):** `TestRoutingFixtures_*` in
  `pkg/plugingen` assert every fixture's expected target exists in the generated
  catalog, required edge cases are present, and no fixture points at a dangling
  skill. This does **not** exercise a live model — it validates the catalog and
  the fixture set structurally.

- **Layer 2 (live smoke, manual — this procedure):** feed each prompt in
  `pkg/plugingen/testdata/routing-fixtures.txt` to a live `/atv` router and check
  the routing decision. This is nondeterministic and is **not** run in CI.

### Fixture format

Each line in `pkg/plugingen/testdata/routing-fixtures.txt` is:

```
<prompt> | <expected-target>
```

where `<expected-target>` is one of: a skill name (must exist in the catalog),
`emit:lfg`, `emit:slfg`, `no-match`, `control:off|on|suggest`, or
`force:<skill>`.

### Running the smoke pass

1. Install the kit (or work in this repo's dogfood `.github/skills/atv/`).
2. For each non-comment line, split on `|` into `prompt` and `expected`.
3. Send `prompt` to the `/atv` router and observe the single route it chooses
   (the router routes to one best skill; it does not expose a ranked list).
4. **Adjudicate manually.** The decision passes if the router's chosen route
   equals `expected`. Because live classification is nondeterministic, a single
   run is only a sample — for a fixture you care about, run it a few times and
   judge whether `expected` is the router's stable, dominant choice rather than
   requiring every run to match.
5. Record systematic misses. A fixture that stably routes somewhere other than
   `expected` means either the target skill's frontmatter description needs
   sharper trigger phrases, or the fixture's `expected` target is wrong.

Because Layer 2 is nondeterministic and model-dependent, treat it as a
qualitative signal for tuning skill descriptions — not a pass/fail CI gate.
