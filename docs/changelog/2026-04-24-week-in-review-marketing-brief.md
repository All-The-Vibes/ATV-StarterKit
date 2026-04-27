# 🚀 ATV Starter Kit — Week in Review

**2026-04-24 · Marketing Brief**

Seven days. Eight drops. One Starter Kit that just leveled up from "useful scaffolding" to **"opinionated AI co-pilot for shipping software."** Dual-reviewer PR pipelines. Session bookends that actually land the plane. A security audit skill that knows the OWASP Top 10 by heart. And yes — a memeIQ Easter Egg, because shipping should be fun.

If you installed before 2026-04-17, you are missing real horsepower. Read on.

---

## 🎯 The Headliners

### 🤖 `/ghcp-review-resolve` — Dual-reviewer PR triage on autopilot

> **One command. Two reviewers. Zero noise.**

**What it is.** A skill that requests a Copilot review *and* runs `pr-review-toolkit` in parallel, adjudicates findings against the actual code, posts only verified bugs as inline comments, then runs a tight fix → verify → commit → reply loop on each thread. Now with **guardrailed thread resolution** and **PR task-list ticking** so the audit trail closes itself when fixes land.

**Why it slaps.**
- Copilot suggestions and toolkit findings overlap on real bugs and disagree on noise. The skill lets the *agreement* through and rejects the rest.
- It never approves, never merges, never closes the PR. Humans stay in control of ship decisions.
- Aborts cleanly on merge conflicts, moving HEAD shas, or already-resolved threads — no fighting reality.

**How to use it.** Open a PR. Run `/ghcp-review-resolve`. Walk away.

**Shipped by.** [@stephschofield](https://github.com/stephschofield) — PR [#23](https://github.com/All-The-Vibes/ATV-starterkit/pull/23) (foundation) + [#26](https://github.com/All-The-Vibes/ATV-starterkit/pull/26) (toolkit port + thread resolution + task ticking)

---

### 🛬 `/land` — Land the plane. Every time.

> **Commit. Push. PR. Handoff. Never strand work locally again.**

**What it is.** A session-completion skill ported from Claude Code into Copilot. Runs adaptive quality gates (Go build + vet at the repo root, npm subproject when relevant), commits the right files (never `git add -A`), pushes to remote, opens a PR with a value-first description, and surfaces a handoff summary. Refuses to merge for you — that's still a human call.

**Why it slaps.**
- A failing gate halts the routine. No green banners on broken builds.
- PR body craft is delegated to `git-commit-push-pr` so the two skills stay in sync.
- Worktree-aware. Multi-project-aware. Quietly portable.

**How to use it.** When you're done coding, type `/land`. That's it.

**Shipped by.** [@stephschofield](https://github.com/stephschofield) — PR [#25](https://github.com/All-The-Vibes/ATV-starterkit/pull/25)

---

### ✈️ `/takeoff` — Know what to work on before your coffee finishes brewing

> **Session kickoff. Prioritized backlog. Zero scrolling.**

**What it is.** The opposite of `/land`. Surfaces top-priority backlog tasks at session start with status, blockers, and dependencies. Prefers `backlog/` (CLI + MCP) when available, falls back gracefully to `docs/plans/*.md` filtered by `status: active` frontmatter — which is exactly what this repo runs on today.

**Why it slaps.**
- Day one, no `backlog/` directory? Still works. Reads your active plans like a project manager.
- The 30,000-foot banner (`✈️ TAKE OFF — NOW AT 30,000 FEET`) is non-negotiable and you will love it.

**How to use it.** Start a session. Type `/takeoff`. Pick your fight.

**Shipped by.** [@stephschofield](https://github.com/stephschofield) — PR [#25](https://github.com/All-The-Vibes/ATV-starterkit/pull/25)

---

### 🛡️ `/atv-security` — 30 rules. 5 categories. Your repo, audited.

> **AgentShield-grade audit for your `.github/` and `.vscode/` configs.**

**What it is.** Rajesh's new agentic-config security skill. Audits **30 rules across 5 categories** — Secrets, Permissions, Hooks, MCP Servers, and Agents/Skills — adapted from the [AgentShield](https://github.com/affaan-m/agentshield) project. Three modes: `report` (read-only), `fix` (interactive auto-fix for safe rules), and `deep` (delegates to AgentShield CLI and merges results). Auto-persists to `docs/security/YYYY-MM-DD-security-report.md`.

**Why it slaps.**
- Catches RCE in hooks, exfiltration patterns, container escape primitives, prompt injection, hidden Unicode chars in agents, oversized prompts, unpinned MCP versions, wildcard tool grants, and `autoApprove` foot-guns.
- The report auto-upserts via HTML-comment markers so it plays nicely with `/cso`.

**How to use it.** Run `/atv-security` for a read-only scan. `/atv-security fix` to apply safe fixes interactively. `/atv-security deep` for the full AgentShield engagement.

**Shipped by.** [@rajesh-ms](https://github.com/rajesh-ms) — PR [#24](https://github.com/All-The-Vibes/ATV-starterkit/pull/24)

---

### 🔒 `/cso` — The Chief Security Officer your repo deserves

> **OWASP Top 10 + STRIDE threat matrix on demand.**

**What it is.** The application-source counterpart to `/atv-security`. Reviews application code against the **OWASP Top 10 (2021)** plus a **STRIDE threat matrix**. Auto-detects Node/TS, Python, Ruby, Go, .NET, and Java. Writes to the same shared `docs/security/<date>-security-report.md` so config + code findings live side-by-side.

**Why it slaps.**
- Two skills, one report file, no clobbering — each section is upserted between HTML-comment sentinels.
- Multi-stack autodetection means you don't tell it what language you're in.

**How to use it.** Run `/cso` from any project root.

**Shipped by.** [@rajesh-ms](https://github.com/rajesh-ms) — PR [#24](https://github.com/All-The-Vibes/ATV-starterkit/pull/24)

---

### 🧠 `/karpathy-guidelines` — Behavioral guardrails from the master himself

> **Andrej Karpathy's LLM-coding observations, ported to Copilot.**

**What it is.** Behavioral coding guardrails derived from [Andrej Karpathy's tweet](https://x.com/karpathy/status/2015883857489522876) on common LLM coding pitfalls. Originally built as a Claude Code plugin by [forrestchang/andrej-karpathy-skills](https://github.com/forrestchang/andrej-karpathy-skills) (h/t Kevin's import lineage), now ported to Copilot's instruction system. Four principles: **Think Before Coding**, **Simplicity First**, **Surgical Changes**, **Goal-Driven Execution**.

**Why it slaps.**
- Stops the "rewrite everything" reflex.
- Bakes test-first sanity into every prompt.
- Included in all three installer presets — Starter, Pro, Full — so the guardrails are on by default.

**How to use it.** Already installed if you ran the guided installer at v2.5.7+. Otherwise, `/karpathy-guidelines` to invoke directly.

**Shipped by.** Karpathy Guidelines — landed in v2.5.7 via commit `f47e6e0` ([release notes](../../CHANGELOG.md#257--2026-04-15))

---

### 🥚 memeIQ — Because shipping should be fun

> **An opt-in Easter Egg that turns your installer into a meme factory.**

**What it is.** Guided installs now expose a `🥚 Easter Eggs` category with an opt-in `memeIQ` entry that scaffolds `.github/skills/meme-iq/SKILL.md` and `.github/agents/meme-iq.agent.md`. Powered by [memegen.link](https://memegen.link). Non-default. Discoverable. Removable.

**Why it slaps.**
- Sandbox installs scaffold the skill + agent in one motion — no manual wiring.
- Local planning artifacts (`PRD.md`, `PROGRESS.md`, `.omc/`, `atv-installer`, `banner-block.txt`) are now properly `.gitignore`'d so your demos don't leak into PRs.

**How to use it.** Run the guided installer. Hit `🥚 Easter Eggs`. Pick `memeIQ`. Make memes.

**Shipped by.** [@shyamsridhar123](https://github.com/shyamsridhar123) — PR [#22](https://github.com/All-The-Vibes/ATV-starterkit/pull/22)

---

## 🔧 Under the Hood

These aren't headliners but they make everything above feel sharper:

- **🪟 Windows installer hardening** — postinstall extraction now uses `tar.exe` with .NET fallbacks, eliminating the "extract failed on Windows" support tickets. ([@dc995](https://github.com/dc995), PR [#20](https://github.com/All-The-Vibes/ATV-starterkit/pull/20))
- **🩺 Agent files repaired (Opus 4.6)** — a previous regression flattened agent files into single-line YAML soup, breaking discovery. Brandon ran them through Opus 4.6 and put the newlines back where God intended. ([@brandonh-msft](https://github.com/brandonh-msft), PR [#21](https://github.com/All-The-Vibes/ATV-starterkit/pull/21))

---

## 🚀 Try it Now

```bash
# Fresh install
npx atv-installer

# Already installed? Reinstall to pull the new skills.
npx atv-installer
```

Once installed, give the new skills a spin in any project:

```text
/takeoff                  # See your top-priority work
/atv-security             # Audit your agentic config
/cso                      # Review application code
/ghcp-review-resolve      # Dual-reviewer triage on an open PR
/land                     # Wrap the session, push, open the PR
```

Browsing? The full version history lives in [`CHANGELOG.md`](../../CHANGELOG.md). The plans behind each drop live in [`docs/plans/`](../plans/).

---

## 🙌 Credit Roll

Massive thanks to this week's shippers — **@stephschofield, @rajesh-ms, @shyamsridhar123, @brandonh-msft, @dc995** — and the upstream lineage we ported from: [forrestchang/andrej-karpathy-skills](https://github.com/forrestchang/andrej-karpathy-skills), [Anthropic's pr-review-toolkit](https://github.com/anthropics/claude-plugins-official/tree/main/plugins/pr-review-toolkit), and [AgentShield](https://github.com/affaan-m/agentshield).

PRs welcome. Memes encouraged. 🥚

— *The ATV Starter Kit team*

---

## ⚡ TL;DR — Skills at a Glance

This week's drops, distilled:

| Skill | Command | What it does | PR |
|-------|---------|--------------|----|
| 🤖 GHCP Review Resolve | `/ghcp-review-resolve` | Dual-reviewer (Copilot + pr-review-toolkit) PR triage with verified-only inline fixes, thread resolution, and PR task ticking | [#23](https://github.com/All-The-Vibes/ATV-starterkit/pull/23), [#26](https://github.com/All-The-Vibes/ATV-starterkit/pull/26) |
| 🛬 Land | `/land` | End-of-session: commit → push → PR → handoff (never merges) | [#25](https://github.com/All-The-Vibes/ATV-starterkit/pull/25) |
| ✈️ Takeoff | `/takeoff` | Start-of-session prioritized backlog briefing with `docs/plans/` fallback | [#25](https://github.com/All-The-Vibes/ATV-starterkit/pull/25) |
| 🛡️ ATV Security | `/atv-security` | 30-rule audit of `.github/` + `.vscode/` configs (secrets, hooks, MCP, agents) | [#24](https://github.com/All-The-Vibes/ATV-starterkit/pull/24) |
| 🔒 CSO | `/cso` | OWASP Top 10 + STRIDE source-code review across Node/TS, Python, Ruby, Go, .NET, Java | [#24](https://github.com/All-The-Vibes/ATV-starterkit/pull/24) |
| 🧠 Karpathy Guidelines | `/karpathy-guidelines` | Behavioral guardrails (Think Before Coding, Simplicity First, Surgical Changes, Goal-Driven) | v2.5.7 |
| 🥚 memeIQ | (Easter Egg installer) | Opt-in meme-generation skill + agent via memegen.link | [#22](https://github.com/All-The-Vibes/ATV-starterkit/pull/22) |

**Plus:** Windows installer hardened ([#20](https://github.com/All-The-Vibes/ATV-starterkit/pull/20)), agent files repaired with Opus 4.6 ([#21](https://github.com/All-The-Vibes/ATV-starterkit/pull/21)).

**High-level take:** Two new session bookends (`/takeoff` → `/land`), two new security skills (`/atv-security` config + `/cso` code), one dual-reviewer PR pipeline (`/ghcp-review-resolve`), one behavioral guardrail (`/karpathy-guidelines`), and one Easter Egg (`memeIQ`). Install or upgrade — everything above is a single `npx atv-installer` away.
