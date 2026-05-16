# `/atv-security` — Unified Security Auditor

> One slash command. Two surfaces. 33 config rules + OWASP Top 10 + STRIDE.

`/atv-security` is ATV's chief security officer skill. It scans your repository in one pass across both surfaces an agentic codebase exposes:

1. **Agentic configuration** — `.github/` and `.vscode/` files that drive AI behaviour (skills, agents, hooks, MCP servers, instruction prompts).
2. **Application source code** — your actual product code, evaluated against OWASP Top 10 (2021) and the STRIDE threat model.

It produces a combined report with per-surface grades, persists it to `docs/security/YYYY-MM-DD-security-report.md`, and optionally applies safe fixes for the deterministic config rules.

---

## Contents

- [Why this skill exists](#why-this-skill-exists)
- [Quick start](#quick-start)
- [Argument grammar](#argument-grammar)
- [What gets scanned](#what-gets-scanned)
- [The 33 config rules](#the-33-config-rules)
  - [Secrets (5 rules)](#secrets-5-rules)
  - [Permissions (2 rules)](#permissions-2-rules)
  - [Hooks (11 rules)](#hooks-11-rules)
  - [MCP Servers (4 rules)](#mcp-servers-4-rules)
  - [Agents & Skills (11 rules)](#agents--skills-11-rules)
- [OWASP Top 10 (2021) coverage](#owasp-top-10-2021-coverage)
- [STRIDE threat model](#stride-threat-model)
- [Scoring & grading](#scoring--grading)
- [Sample report](#sample-report)
- [`mode=fix` — what gets auto-fixed](#modefix--what-gets-auto-fixed)
- [Report persistence and the `/cso` heritage block](#report-persistence-and-the-cso-heritage-block)
- [Heritage: `/cso` → `/atv-security`](#heritage-cso--atv-security)
- [Limitations and honest disclosure](#limitations-and-honest-disclosure)
- [FAQ](#faq)

---

## Why this skill exists

A typical security tool scans one of three things: source code, dependencies, or infrastructure. None scan the new attack surface that comes with agentic coding: the prompts, skills, agents, hooks, and MCP servers that tell an AI what to do. That surface can contain:

- Hardcoded API keys checked into instruction files
- Prompt-injection payloads embedded in skill descriptions
- Hooks that `curl | bash` arbitrary URLs at session start
- MCP servers configured with `autoApprove: true` and `tools: ["*"]`
- Zero-width Unicode characters hiding instructions inside agent definitions

At the same time, the application code itself still needs the standard OWASP/STRIDE review. `/atv-security` does both in a single pass so security posture is evaluated holistically — config drift can't hide behind clean source code, and clean configs can't excuse a vulnerable API.

The 33 config rules are adapted from [AgentShield](https://github.com/affaan-m/agentshield) (the leading taxonomy for agentic-config threats). The OWASP and STRIDE phases follow the official 2021 OWASP Top 10 and Microsoft's STRIDE methodology.

---

## Quick start

After installing ATV (`npx atv-starterkit init` or via the marketplace plugin `atv-skill-atv-security`), run:

```text
/atv-security
```

That's it. The skill auto-detects what surfaces exist in your project, runs every applicable phase, and prints a report. The report is also written to `docs/security/YYYY-MM-DD-security-report.md` so you can diff it across runs.

To apply safe auto-fixes (replacing secrets with environment-variable references, etc.):

```text
/atv-security fix
```

To scope the scan:

```text
/atv-security config        # configuration audit only
/atv-security owasp         # OWASP scan only
/atv-security stride        # STRIDE threat model only
/atv-security src/api       # OWASP + STRIDE narrowed to src/api/
/atv-security fix config    # apply auto-fixes to config issues only
```

---

## Argument grammar

The skill parses `$ARGUMENTS` along two independent axes:

| Axis | Values | Default | Notes |
|------|--------|---------|-------|
| **Mode** | `report`, `fix` | `report` | Token order doesn't matter. Case-insensitive. |
| **Scope** | `full`, `config`, `owasp`, `stride`, `<path>` | `full` | A token containing `/` or `\`, or matching an existing path, is treated as a path scope. |

A path scope (e.g., `src/api/`) narrows OWASP and STRIDE to that directory and skips the config phase entirely. The `config` scope narrows to the configuration surfaces and skips OWASP/STRIDE.

Examples (all valid):

| Invocation | Mode | Scope |
|------------|------|-------|
| `/atv-security` | report | full |
| `/atv-security fix` | fix | full |
| `/atv-security config` | report | config |
| `/atv-security config fix` | fix | config |
| `/atv-security fix config` | fix | config |
| `/atv-security owasp` | report | owasp |
| `/atv-security src/api/` | report | `src/api/` |
| `/atv-security fix src/api/` | fix | `src/api/` |

> **Heritage triggers:** Phrases like `cso`, `owasp scan`, `stride analysis`, `threat model`, `application security`, and `security review code` all route to `/atv-security` for migration discoverability from the deprecated `/cso` skill.

---

## What gets scanned

### Configuration surfaces (Phase 1a)

| Surface | File pattern | Category |
|---------|--------------|----------|
| Instructions | `.github/copilot-instructions.md` | Agents & Skills |
| MCP config | `.github/copilot-mcp-config.json` | MCP Servers |
| Skills | `.github/skills/**/*.md` | Agents & Skills |
| Agents | `.github/agents/**/*.agent.md` | Agents & Skills |
| Hooks | `.github/hooks/copilot-hooks.json` + `.github/hooks/scripts/**` | Hooks |
| Setup steps | `.github/copilot-setup-steps.yml` | Hooks |
| VS Code | `.vscode/settings.json`, `.vscode/extensions.json` | Permissions |

### Application source (Phase 1b)

Stack detection is signal-based:

| Detection signal | Stack | Default scan paths |
|------------------|-------|--------------------|
| `package.json`, `*.ts`, `*.js` | Node.js / TypeScript | `src/**`, `routes/**`, `api/**`, `pages/**` |
| `requirements.txt`, `*.py` | Python | `app/**`, `src/**`, `views/**`, `api/**` |
| `Gemfile`, `*.rb` | Ruby / Rails | `app/**`, `config/**`, `db/**` |
| `go.mod`, `*.go` | Go | `**/*.go` |
| `*.cs`, `*.csproj` | .NET | `**/*.cs`, `Controllers/**` |
| `pom.xml`, `*.java` | Java | `src/**/*.java` |

### Bail rules

The skill stops cleanly with an explanatory message if:

- Neither config surfaces nor source code exist in scope.
- `scope=config` is requested but there's no `.github/` directory.
- `scope ∈ {owasp, stride, <path>}` is requested but no source files are found.

---

## The 33 config rules

Rules are organized into 5 categories with two evaluation tiers:

- **Tier 1 — Deterministic regex.** Run via `grep_search`. Fast, no false-positive tolerance, eligible for `mode=fix`.
- **Tier 2 — LLM-assessed.** The skill reads the file and reasons about it. Catches things regex can't (prompt injection semantics, business logic, missing controls).

Severities throughout: 🔴 critical · 🟡 high · 🟢 medium · 🔵 low · ⚪ info

### Secrets (5 rules)

All Tier 1 (deterministic), all scoped to `.github/**` and `.vscode/**`.

| Rule | What it catches | Severity | Auto-fix |
|------|-----------------|----------|----------|
| **SEC-01** | Anthropic API keys (`sk-ant-…`) | 🔴 critical | Replace with `${ANTHROPIC_API_KEY}` |
| **SEC-02** | OpenAI project keys (`sk-proj-…`) | 🔴 critical | Replace with `${OPENAI_API_KEY}` |
| **SEC-03** | AWS access keys (`AKIA…`) | 🔴 critical | Replace with `${AWS_ACCESS_KEY_ID}` |
| **SEC-04** | GitHub tokens (`ghp_…`, `github_pat_…`) | 🔴 critical | Replace with `${GITHUB_TOKEN}` |
| **SEC-05** | Bearer tokens and DB connection strings (Mongo, Postgres, MySQL, Redis) | 🟡 high | Replace with appropriate env-var reference |

> SEC-* rules are the only ones that fire on **any** match. There are no benign exceptions for hardcoded credentials in agentic config files.

### Permissions (2 rules)

| Rule | Tier | What it catches | Severity |
|------|------|-----------------|----------|
| **PERM-01** | T1 regex | `security.workspace.trust.enabled: false` or `chat.tools.autoApprove: true` in `.vscode/settings.json` | 🟢 medium |
| **VSCODE-01** | T2 LLM | Extension recommendations from untrusted/unknown publishers in `.vscode/extensions.json` without justification | 🔵 low |

### Hooks (11 rules)

Hooks are the most dangerous category — they execute code at session lifecycle events. Treat hook findings as load-bearing.

#### Tier 1 — grep patterns over `.github/hooks/scripts/**`

| Rule | Pattern catches | Severity | Auto-fix |
|------|-----------------|----------|----------|
| **HOOK-01** | Unsanitized variables in `curl`/`wget`/`eval` (`curl …${var}…`) | 🟡 high | Manual — review & sanitize |
| **HOOK-02** | Data exfil patterns (`curl -X POST …$`, `wget --post`) | 🔴 critical | Manual — review intent |
| **HOOK-03a** | Silent error suppression: `2>/dev/null` | 🟢 medium | Replace with logging |
| **HOOK-03b** | Silent error suppression: `\|\| true` at end of line | 🟢 medium | Replace with logging |
| **HOOK-03c** | Silent error suppression: `\|\| exit 0` at end of line | 🟢 medium | Replace with logging |

#### Tier 2 — LLM-assessed (`.github/hooks/copilot-hooks.json` + scripts)

| Rule | What it catches | Severity |
|------|-----------------|----------|
| **EXEC-01** | Hook downloads and executes remote code (`curl \| sh`, `wget` + execute) | 🔴 critical |
| **EXEC-02** | Global package installs (`npm install -g`, `pip install`, `gem install`, `cargo install`) | 🟢 medium |
| **EXEC-03** | Container escape patterns (`docker --privileged`, `--pid=host`, `--network=host`, root volume mounts) | 🔴 critical |
| **EXEC-04** | Credential access (keychain reads, `/etc/shadow`, `.aws/credentials`) | 🔴 critical |

#### Tier 2 — Setup steps (`.github/copilot-setup-steps.yml`)

| Rule | What it catches | Severity |
|------|-----------------|----------|
| **SETUP-01** | Remote script execution in setup (`curl \| bash`, `wget \| sh`, download + run) | 🔴 critical |
| **SETUP-02** | Privileged operations (`sudo` without justification, `chmod 777`, `chown root`) | 🟡 high |

### MCP Servers (4 rules)

MCP (Model Context Protocol) servers expose tools to the agent. Misconfiguration here grants the AI capabilities you didn't intend.

#### Tier 1 — grep patterns over `.github/copilot-mcp-config.json`

| Rule | Pattern catches | Severity | Auto-fix |
|------|-----------------|----------|----------|
| **MCP-02** | Wildcard tool grant: `"tools": ["*"]` | 🟡 high | Scope to specific tools |
| **MCP-03** | `autoApprove` flag present | 🟢 medium | Remove or restrict to read-only tools |
| **MCP-04** | Hardcoded credentials in MCP env sections (`sk-ant-…`, `sk-proj-…`, `AKIA…`, `ghp_…`, `Bearer …`) | 🔴 critical | Replace with `${input:VAR}` or `${ENV_VAR}` |

#### Tier 2 — LLM assessment

| Rule | What it catches | Severity |
|------|-----------------|----------|
| **MCP-01** | MCP servers using `npx -y` without a pinned version (`@package` instead of `@package@1.2.3`) — a supply-chain pinning gap | 🟡 high |

### Agents & Skills (11 rules)

This category covers the AI's instructional surface — the prose that defines what the agent does. It's the prime target for prompt injection.

#### Tier 1 — grep patterns

| Rule | Pattern catches | Severity | Notes |
|------|-----------------|----------|-------|
| **AGENT-01** | Zero-width Unicode characters (`U+200B`, `U+200C`, `U+200D`, `U+FEFF`) in instruction files | 🔴 critical | Hidden-instruction injection vector |
| **AGENT-02** | Base64-ish blobs (≥80 chars, `[A-Za-z0-9+/]{80,}={0,2}`) in skills/agents | 🟢 medium | Decode and inspect. Ignore if preceded by `sha256:`, `data:`, or `http` |

#### Tier 2 — Prompt injection (`.github/copilot-instructions.md`, all agents, all skills)

| Rule | What it catches | Severity | Benign exceptions |
|------|-----------------|----------|-------------------|
| **INJ-01** | "always run", "without asking", "automatically install", "execute without confirmation" | 🟡 high | "always run tests" / "always check" are quality patterns, not injections |
| **INJ-02** | "fetch and execute", "download and run", `curl \| bash`, "eval remote" | 🔴 critical | None — always flag |
| **INJ-03** | System-prompt overrides: "ignore previous instructions", "you are now", "DAN", "jailbreak", fake system messages | 🔴 critical | None — always flag |
| **INJ-04** | Output manipulation: "always report ok", "suppress warnings", "remove security findings", "hide errors" | 🟡 high | Legitimate error handling is benign |
| **INJ-05** | Time-delayed execution: "after 5 minutes", "when user is away", "at 3am", conditional-on-absence triggers | 🟡 high | Scheduled CI/CD references are benign |

#### Tier 2 — Agent access control (all `.github/agents/*.agent.md`)

| Rule | What it catches | Severity |
|------|-----------------|----------|
| **ACC-01** | Unrestricted Bash/shell access granted to an agent without scoping | 🟡 high |
| **ACC-02** | Agent with tool access but no `allowedTools` restriction | 🟢 medium |
| **ACC-03** | Escalation chains: agent can spawn sub-agents with elevated permissions | 🟡 high |

#### Tier 2 — Oversized prompts

| Rule | What it catches | Severity | Self-exemption |
|------|-----------------|----------|----------------|
| **AGENT-03** | Skill or agent files with >8,000 chars of effective prose (excluding YAML frontmatter, fenced code blocks, and markdown tables) | 🟢 medium | First-party ATV security skills (e.g., `atv-security` itself) are exempt — they intentionally bundle rule definitions and exceed the limit by design. User-authored skills are **not** exempt. |

---

## OWASP Top 10 (2021) coverage

Phase 4 runs against the application source. Every OWASP category mixes Tier 1 regex (fast, low-noise patterns) with Tier 2 LLM assessment (semantic checks). A grep alone can't tell you whether a route is protected by auth middleware — but a 30-second read of the file can.

| Category | Tier 1 patterns | Tier 2 assessment |
|----------|-----------------|---------------------|
| **A01: Broken Access Control** | Hardcoded role strings, ad-hoc auth (`req.user &&` without middleware) | Auth coverage on state-changing endpoints; ownership validation on `/resource/:id`; admin endpoint protection |
| **A02: Cryptographic Failures** | Weak algorithms (`md5`, `sha1`, `DES`, `RC4`), hardcoded passwords, plaintext `http://` API URLs | Password hashing (bcrypt/scrypt/argon2 vs MD5/SHA1); encryption at rest; TLS enforcement |
| **A03: Injection** | SQL string concat, `eval`/`exec`/`Function`, XSS (`innerHTML`, `dangerouslySetInnerHTML`, `\|safe`, `\|raw`), OS command injection, NoSQL injection | Parameterized queries; output escaping; LDAP/XPath/header injection |
| **A04: Insecure Design** | _(Tier 2 only)_ | Rate limiting on auth endpoints; business-logic flaws (negative quantities, price manipulation); account enumeration via differential errors |
| **A05: Security Misconfiguration** | `DEBUG = True`, unrestricted CORS, Django wildcard hosts | Default credentials; custom error pages (no stack traces); security headers (helmet etc.); disabled-in-prod features |
| **A06: Vulnerable & Outdated Components** | Dependency declarations in `package.json` | Recommend `npm audit`, `pip-audit`, `bundle audit`, `govulncheck` — does not duplicate them |
| **A07: ID & Auth Failures** | Excessive JWT TTLs (≥30d), long session `maxAge`, weak bcrypt rounds (<6) | Brute-force protection; password complexity; session invalidation on logout |
| **A08: Software/Data Integrity** | Unsafe deserialization (`deserialize`, `unserialize`, `pickle.load`, `yaml.load`) | CI/CD tampering protection; signed updates; SRI (Subresource Integrity) on third-party CDN scripts |
| **A09: Logging & Monitoring** | _(Tier 2 only)_ | Login failures / access-denied / validation failures logged; structured logging (no log injection); alerting on suspicious patterns |
| **A10: SSRF** | User-controllable URL passed to `fetch`, `axios`, `requests`, `http.Get`, `urllib.request.urlopen` | URL allowlisting; blocked internal ranges (`169.254.x.x`, `10.x.x.x`, `127.x.x.x`) |

---

## STRIDE threat model

Phase 5 reads architectural signals (entry points, data flows, trust boundaries, asset locations) and produces a threat matrix:

| Threat | Concern | Typical questions |
|--------|---------|--------------------|
| **S**poofing | Identity | Can an attacker impersonate a user or service? Are auth tokens forgeable? |
| **T**ampering | Data integrity | Can payloads be modified in transit or at rest? Are signatures verified? |
| **R**epudiation | Accountability | Can actions be performed without an audit trail? Are logs append-only? |
| **I**nformation Disclosure | Confidentiality | Can sensitive data leak via error pages, logs, or overly-verbose API responses? |
| **D**enial of Service | Availability | Can endpoints be overwhelmed? Is there rate limiting? Are bulk operations bounded? |
| **E**levation of Privilege | Authorization | Can a user gain unauthorized access? Are admin endpoints properly gated? |

For each row, the skill identifies whether a mitigation **already exists** in the codebase, rates the risk High/Medium/Low, and (if missing) suggests a concrete control.

STRIDE is the slowest phase. If you want a fast read, use `/atv-security config` or `/atv-security owasp`.

---

## Scoring & grading

Three independent grades are computed. Surfaces that were not scanned (absent or out of scope) render as **N/A** and are excluded from the aggregate — the skill never invents a 0 or 100 for a surface it didn't actually evaluate.

### Per-finding deductions

| Severity | Config penalty | OWASP penalty |
|----------|-----------------|----------------|
| 🔴 critical | −15 | −15 |
| 🟡 high | −10 | −10 |
| 🟢 medium | −5 | −5 |
| 🔵 low | −2 | −2 |
| ⚪ info | 0 | — |

Each category (and OWASP score) starts at 100 and floors at 0 — never negative.

### Config aggregate weighting

```text
ConfigScore = Secrets×0.20 + Permissions×0.15 + Hooks×0.25 + MCP×0.25 + Agents×0.15
```

Hooks and MCP carry the heaviest weight because they're the most directly exploitable.

### STRIDE posture

| Unmitigated threats | Posture |
|---------------------|---------|
| 0 | 🟢 Strong |
| 1–2 | 🟡 Moderate |
| 3+ | 🔴 Weak |

### Letter grades

| Score | Grade |
|-------|-------|
| 90–100 | A |
| 80–89 | B |
| 65–79 | C |
| 50–64 | D |
| 0–49 | F |

### Overall

```text
OverallScore = mean(ConfigScore, OWASPScore)       # surfaces that ran only
              − 5 × min(STRIDE_unmitigated, 4)      # capped at −20
```

If only one surface ran, the overall score is simply that surface's score; the other surface is marked N/A.

### Simplified pass/fail mode

When exact arithmetic is hard (e.g., many findings across categories), the skill may fall back to per-category pass/fail:

- ≥1 critical → 🔴 → 40/F
- ≥1 high, 0 critical → 🟡 → 70/C
- otherwise → 🟢 → 95/A

Overall config status = worst category status.

---

## Sample report

The skill prints a report shaped like this (illustrative — real output will reflect your findings):

````markdown
## 🛡️ ATV Security Report

**Date:** 2026-05-16
**Scope:** full
**Surfaces scanned:** config: yes · source: yes — stack: Node.js / TypeScript

### Summary

| Surface | Grade | Score |
|---------|-------|-------|
| Configuration | B | 82 |
| OWASP Top 10 | A | 91 |
| STRIDE posture | 🟡 Moderate (2 unmitigated) |
| **Overall** | **B** | **81** |

### Configuration findings (5)

#### 🔴 SEC-04 — GitHub token committed
- **File:** `.github/copilot-mcp-config.json:14`
- **Evidence:** `"GITHUB_TOKEN": "ghp_AbCdEf…"`
- **Fix:** Replace with `${input:GITHUB_TOKEN}` reference

#### 🟡 MCP-02 — Wildcard tool grant
- **File:** `.github/copilot-mcp-config.json:22`
- **Evidence:** `"tools": ["*"]`
- **Fix:** Scope to the specific tools the server actually needs.

(… 3 more findings …)

### OWASP findings (2)

#### 🟡 A07 — Excessive JWT lifetime
- **File:** `src/auth/tokens.ts:34`
- **Evidence:** `jwt.sign(payload, secret, { expiresIn: '30d' })`
- **Recommendation:** Reduce to 15m–1h with a refresh-token flow.

(… 1 more …)

### STRIDE matrix

| Threat | Risk | Mitigation status |
|--------|------|---------------------|
| **S**poofing | M | ✅ JWT auth in place |
| **T**ampering | L | ✅ HTTPS + payload schemas |
| **R**epudiation | H | ❌ No audit log for state-changing endpoints |
| **I**nformation Disclosure | L | ✅ Custom error pages |
| **D**enial of Service | H | ❌ No rate limiting on `/api/auth/login` |
| **E**levation of Privilege | L | ✅ RBAC middleware on admin routes |

### Recommended next steps

1. Rotate the leaked GitHub token immediately (it was in `.github/copilot-mcp-config.json`).
2. Run `/atv-security fix config` to apply the safe auto-fixes.
3. Add an audit log for the 3 state-changing endpoints listed under Repudiation.
4. Add rate limiting to `/api/auth/login`.
````

---

## `mode=fix` — what gets auto-fixed

`mode=fix` only modifies **deterministic Tier 1 config rules** where the remediation is mechanical and reversible. The skill never modifies application source code — OWASP/STRIDE findings always come back as recommendations the human reviews.

| Auto-fixable | Always-manual |
|--------------|----------------|
| SEC-01 through SEC-05 (replace secret with env-var reference) | All INJ-* prompt-injection findings (semantic — needs human judgment) |
| MCP-04 (env-section secret replacement) | All ACC-* agent access-control findings |
| MCP-02 (wildcard tools → typically suggested, not applied unless an obvious safe set is detectable) | All EXEC-* / SETUP-* hook execution findings |
| HOOK-03a/b/c (error suppression → suggested replacement with logging, applied only with confirmation) | AGENT-01 (zero-width chars — may be intentional in i18n contexts) |
|   | AGENT-02 (base64 blobs — needs decoding & inspection) |
|   | All OWASP / STRIDE findings |

The skill always prints the full report **before** applying fixes, so you can abort if the proposed fix list looks wrong.

---

## Report persistence and the `/cso` heritage block

After each run the skill upserts `docs/security/YYYY-MM-DD-security-report.md` with two HTML marker blocks so old `/cso` reports and new `/atv-security` reports compose cleanly without overwriting each other:

```markdown
<!-- atv-security -->
…config audit findings…
<!-- /atv-security -->

<!-- cso -->
## /cso Scan
…OWASP + STRIDE findings…
<!-- /cso -->
```

- The `<!-- atv-security -->` block holds the **config audit** half.
- The `<!-- cso -->` block holds the **OWASP + STRIDE** half (with the legacy `## /cso Scan` heading shape preserved for tools that parse the older format).
- Re-running `/atv-security` on the same day **upserts in place** — your manual notes in surrounding markdown survive.
- Running on a new date creates a new file; old reports stay around for historical diffing.

---

## Heritage: `/cso` → `/atv-security`

`/atv-security` is the result of folding the former standalone `/cso` skill into a single unified auditor in v2.5.9 (commit `4f9e219`, 2026-04-26). The merger fixed two problems:

1. **Name collision** with gstack's separate `/cso` skill. The ATV `/cso` was renamed to `/atv-security`. Gstack's `/cso` is untouched and continues to ship with gstack.
2. **Surface fragmentation.** Running two skills back-to-back to audit one repo was awkward. The unified skill now produces a combined report.

What changed in the merger:

- ✅ All `/cso` triggers (`cso`, `owasp scan`, `stride analysis`, `threat model`, `application security`, `security review code`) still route to `/atv-security`.
- ✅ Report persistence keeps both marker blocks, so historical reports parse without changes.
- ✅ The guided installer's `🔒 Security` category now shows a single entry instead of two.
- ✅ Instruction templates (`general.md`, `python.md`, `rails.md`, `typescript.md`) reference `/atv-security` only.
- ❌ `pkg/scaffold/templates/skills/cso/` was deleted from the installer. Gstack's `/cso` is unaffected and continues to ship via gstack.

See `CHANGELOG.md` lines 62–86 for the full v2.5.9 release notes.

---

## Limitations and honest disclosure

`/atv-security` is a thorough first pass, **not a replacement for professional security review**. Specifically:

- **Tier 2 rules are LLM-assessed**, which means non-deterministic. Two runs may surface slightly different findings for borderline cases. Critical findings (🔴) tend to be stable; medium findings (🟢) drift more.
- **No dependency CVE scanning.** A06 recommends `npm audit` / `pip-audit` / `bundle audit` / `govulncheck` rather than duplicating them.
- **No dynamic analysis.** No DAST, no fuzzing, no penetration testing. It's a static review.
- **AGENT-03 self-exemption.** The `atv-security` skill itself is intentionally >8,000 chars and is exempted from the oversized-prompt rule. User-authored skills are **not** exempt — long custom skills will still be flagged.
- **Stack coverage is biased toward web stacks.** Mobile, embedded, and ML-pipeline-specific patterns are not first-class. PRs welcome.
- **OWASP A06 is a stub.** It recommends ecosystem tooling rather than reimplementing it. Run `npm audit` etc. separately.
- **Heritage `/cso` triggers can route unexpectedly.** If you have a gstack installation, both gstack's `/cso` and ATV's `/atv-security` exist. The trigger `cso` is shared semantics; the skills are distinct.

---

## FAQ

### Q: I ran `/atv-security` and got "No ATV configuration or application source detected."

Run it from inside a project directory. The skill checks for `.github/`, `.vscode/`, or detectable source files. An empty workspace returns this bail message.

### Q: Why is my OverallScore lower than both surface scores?

STRIDE applies a penalty (−5 per unmitigated threat, capped at −20) on top of the mean of Config and OWASP. If STRIDE found 3 unmitigated threats, you'll see a −15 applied to the overall.

### Q: How do I exclude a directory?

Use a path scope: `/atv-security src/` will run OWASP/STRIDE only against `src/`. To exclude `vendor/` etc., use a more specific scope.

### Q: Can I run this in CI?

The skill is invoked through Copilot Chat — it's an LLM-driven workflow, not a CLI tool. For CI, use deterministic scanners (`gitleaks`, `semgrep`, `bandit`, etc.) and reserve `/atv-security` for human-in-the-loop review.

### Q: Where can I add custom rules?

The rule set lives in `pkg/scaffold/templates/skills/atv-security/SKILL.md` (the canonical source). Edit there and re-run the installer (`atv install`) to push your customizations into your project's `.github/skills/atv-security/SKILL.md`. To contribute upstream, open a PR against [All-The-Vibes/ATV-StarterKit](https://github.com/All-The-Vibes/ATV-StarterKit).

### Q: Why doesn't this skill modify my application code in `mode=fix`?

By design. OWASP and STRIDE findings often have multiple valid remediations (parameterize this query, swap algorithms, add rate limiting where exactly?). Auto-applying one of those would be presumptuous. Config secrets and MCP env vars, by contrast, have one correct fix shape (replace literal with reference), so they're safe to automate.

### Q: I get findings I disagree with. How do I suppress them?

There's no suppression file. Document the false positive in your security report under a heading like `## Accepted Risks` — the surrounding markdown survives upserts, so your notes persist across runs.

### Q: How does this compare to AgentShield?

The 33 config rules are adapted from [AgentShield](https://github.com/affaan-m/agentshield). `/atv-security` wraps them in Copilot's skill format, adds OWASP/STRIDE coverage, and persists reports. If you want AgentShield's standalone CLI (no LLM, no OWASP), use it directly — they're complementary.

---

## See also

- **Source of truth:** `pkg/scaffold/templates/skills/atv-security/SKILL.md` (the canonical skill definition, ~517 lines, shipped to every install).
- **Plugin packages:** `atv-pack-security`, `atv-everything`, `atv-skill-atv-security` in [`docs/marketplace.md`](marketplace.md).
- **v2.5.9 release notes:** `CHANGELOG.md` lines 62–86.
- **AgentShield rule taxonomy:** <https://github.com/affaan-m/agentshield>
- **OWASP Top 10 (2021):** <https://owasp.org/Top10/>
- **STRIDE:** <https://learn.microsoft.com/en-us/azure/security/develop/threat-modeling-tool-threats>
