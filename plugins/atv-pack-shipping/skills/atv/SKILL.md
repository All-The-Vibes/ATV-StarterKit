---
name: atv
description: ATV Starter Kit router - one command that routes your request to the right ATV skill or pipeline. Use when you have a request but don't know which skill to name, when you type "/atv", "atv", "atv starter kit", "route this", "which atv skill", or when you want ATV to pick the best workflow. Dispatches to planning, review, security, research, browser-QA, and the /lfg /slfg build pipelines. Say "/atv off" to stop auto-routing, "/atv @<skill> <args>" to force a specific skill.
argument-hint: "[your request, or 'off'|'on'|'suggest', or '@<skill> <args>' to force]"
---

# ATV Router

One job: send the request to the right ATV skill. This is a **router**, not a
pipeline — it classifies intent and hands off. It holds no Edit/Write; a router
routes, it does not build.

The capability menu is `llms.txt` in this skill's directory — one line per skill,
generated from each skill's own `SKILL.md` frontmatter description (which carries
the trigger phrases). **Read `llms.txt` first**, then match the user's request
against it.

## Step 0 — Handle control commands first (before any routing)

Parse `$ARGUMENTS`. If it is one of these, act and STOP (do not route):

- **`off`** → `node .github/hooks/scripts/atv-config.js set proactive false`, then say:
  "ATV auto-routing off. I'll suggest skills but not invoke them. Turn back on with `/atv on`." STOP.
- **`on`** → `atv-config.js set proactive true`, then say: "ATV auto-routing on." STOP.
- **`suggest`** → `atv-config.js set proactive suggest`, then say:
  "ATV will suggest a skill and ask before invoking." STOP.
- **empty (bare `/atv`)** → print the catalog menu (read `llms.txt`, show it as a
  readable list grouped sensibly) and one line: "Type your request after `/atv`,
  or `/atv @<skill>` to jump straight to one." STOP.
- **`@<skill> <args>`** (force syntax) → skip classification entirely. If `<skill>`
  exists in `llms.txt`, invoke it directly with `<args>` (honoring the
  irreversible-target confirm gate below). If it does not exist, say so and print
  the closest matches from the menu. This is the mis-route escape hatch.

## Step 1 — Read the proactivity setting

`node .github/hooks/scripts/atv-config.js get proactive` (default `true` when
unset or the config is absent/corrupt — the shim falls back safely).

- `true` → **invoke** the matched skill via the Skill tool (with the gates below).
- `false` → do NOT invoke. At most say: "I think `/<skill>` fits — want me to run it?"
- `suggest` → name the skill you'd route to and ask for a yes before invoking.

## Step 2 — Route

1. **Browser / QA / screenshots / inspect-a-page** (open a site, test a deploy,
   take a screenshot, check a flow visually) → route to `/test-browser`. Special-cased first.
2. **"build / implement / just do it" (full autonomous work)** → this is a
   **pipeline**, not a single skill. See Handoff below — emit `/lfg` (or `/slfg`
   for swarm/parallel/fast). Do NOT try to auto-invoke it.
3. **Otherwise** match `$ARGUMENTS` against `llms.txt` and route to the best skill.
4. **No confident match** (below-confidence floor) → do NOT force a route.
   Answer directly if you can, and show the catalog menu so the user can pick.
   A wrong route is worse than an honest "here are your options."

**Governing heuristic:** when a skill clearly fits, invoke it — a structured
workflow beats an ad-hoc answer. But respect the no-match floor: never jam an
ambiguous request into the nearest skill just to route something.

### Announce as you route (D2)

When you invoke, say it in one line first:
`Routing to /<skill> — say "no, I meant ..." to redirect.`
Zero round-trip, but transparent and correctable.

### Confirm gate for irreversible targets (GATE 2)

Before invoking or emitting any of these, **ask for confirmation first** — announce-
then-redirect is too late once a pipeline or ship acts:

- `/lfg`, `/slfg` (full autonomous pipelines)
- `/land` (commit + push + PR)
- anything that ships, deploys, or pushes

Say: "This runs `/<skill>`, which <what it does irreversibly>. Proceed?" and wait.

## Build handoff (the `/lfg` `/slfg` pipelines)

`/lfg` and `/slfg` carry `disable-model-invocation: true` — a deliberate guard so
the full pipeline never auto-fires. The router does **not** remove or bypass that
flag. For "build X" intents, the router uses **emit-and-stop**:

> "This is a full build. Run `/lfg <feature>` (or `/slfg <feature>` for parallel
> swarm mode) to start the autonomous pipeline."

Then STOP. The user fires the pipeline themselves. This keeps the one-command
promise honest (one hop to the right pipeline) without defeating the guard.

## Routing map (intent → target)

Match against `llms.txt` for the authoritative, always-current list. Common intents:

| User intent | Route to |
|---|---|
| build / implement feature X, "just do it" | **emit** `/lfg` (or `/slfg` for swarm) |
| new idea, "is this worth building", explore, pitch | `/ce-brainstorm` or `/brainstorming` |
| "what should I improve", generate ideas | `/ce-ideate` |
| "plan this", turn requirements into a plan | `/ce-plan` |
| deepen / enrich an existing plan | `/deepen-plan` |
| review a plan / requirements doc | `/document-review` |
| do implementation work on an existing plan | `/ce-work` |
| review code / "look at my diff" | `/ce-review` |
| strip AI slop | `/unslop` |
| bug / "why is this broken" / "this doesn't work" | `/ce-work` *(provisional — dedicated debug skill coming)* |
| browser QA / "does this page work" | `/test-browser` |
| research / experiment loop | `/autoresearch` |
| security / "is this secure" | `/atv-security` |
| capture a solved problem / learnings | `/ce-compound`, `/learn`, `/observe` |
| "what should I work on" / session start | `/takeoff` |
| "land the plane" / wrap up | `/land` *(confirm gate)* |
| record a feature demo | `/feature-video` |
| meme | `/meme-iq` |
| health / install check | `/atv-doctor` |

Bug routing is **provisional** — when routing a bug to `/ce-work`, add one line:
"Debugging routing is provisional; a dedicated `/investigate`-style skill is
coming." (Plan D4.)

## Telemetry (route logging)

After routing, best-effort append ONE line to `~/.atv/analytics/routes.jsonl`
(OTel-shaped: `name`, `attributes`, `timestamp`). Log the **intent category and
the routed-to skill ONLY — never the raw request text** (secret/PII safety, plan
Security §3). Never block on it.

## Not a router job

The router does not implement, plan, or review anything itself. It classifies and
hands off. If nothing fits, it says so and shows the menu.
