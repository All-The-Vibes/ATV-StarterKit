# Known Pitfalls — ATV Contributors

A consolidated, imperative `DO NOT` register of the mistakes that the ATV
codebase already actively guards against. ATV is a **GitHub Copilot harness**:
the skills, agents, prompts, and plugin manifests it ships are loaded as
instructions by an agent at runtime, so a structural slip here is a runtime
correctness bug for every downstream user — not just a style nit.

Most entries below are **machine-enforced**: a Go test under `pkg/scaffold/` or
a workflow under `.github/workflows/` fails the build when the pitfall
regresses. Each enforced entry carries a `enforced-by:` marker naming the guard.
A companion test (`pkg/scaffold/known_pitfalls_test.go`) parses those markers and
fails if any referenced guard has been moved or deleted — so this register can
never quietly drift out of sync with the checks it documents.

When you fix a bug rooted in any item below, add (or point at) a regression
guard and record it here with a `enforced-by:` marker. See
[When You Discover a New Pitfall](#when-you-discover-a-new-pitfall).

---

## Skill & Agent Authoring

### DO NOT ship a definition with a malformed frontmatter block
Every `SKILL.md` and `*.agent.md` must open on line 1 with `---` and close with a
`---` on its own line. No missing closing delimiter, no single-line minified form
(`---description: x user-invocable: true---…`), no duplicate top-level keys. The
agent templates were once stored minified and silently excluded from validation;
they are now held to the same contract as every other definition.

Enforced by `pkg/scaffold/skillspec_test.go`.
<!-- enforced-by: pkg/scaffold/skillspec_test.go -->

### DO NOT let a skill's `name` differ from its folder
`skills/foo/SKILL.md` must declare `name: foo`. A mismatch breaks invocation and
de-duplication. Skill names must also be unique across the tree.

Enforced by `pkg/scaffold/skillspec_test.go`.
<!-- enforced-by: pkg/scaffold/skillspec_test.go -->

### DO NOT leave `description` empty or a placeholder
Every skill and agent needs a real, specific `description`. The unfilled
placeholders `TODO`, `TBD`, `FIXME`, `changeme`, `change me`, and the literal
word `description` are rejected — Copilot surfaces this text in the picker and
uses it to decide when to load the definition.

Enforced by `pkg/scaffold/skillspec_test.go`.
<!-- enforced-by: pkg/scaffold/skillspec_test.go -->

### DO NOT omit `user-invocable` on an agent definition
Every `*.agent.md` must declare `user-invocable:` with a literal `true` or
`false`. Anything else is a structural defect.

Enforced by `pkg/scaffold/skillspec_test.go`.
<!-- enforced-by: pkg/scaffold/skillspec_test.go -->

### DO NOT ship a definition with an empty body
A definition that is all frontmatter and no body carries no instructions. The
body after the closing `---` must be non-empty.

Enforced by `pkg/scaffold/skillspec_test.go`.
<!-- enforced-by: pkg/scaffold/skillspec_test.go -->

---

## Copilot Harness Neutrality

### DO NOT leave Claude Code references in skill or agent instructions
ATV runs under GitHub Copilot. Instructional text under `.github/skills/` and
`pkg/scaffold/templates/skills/` must not contain `Claude Code`, `.claude/`
paths, `CLAUDE.md`, `anthropic`, `CLAUDE_*` environment variables, or
`code`/`platform`.`claude.com`/`.ai` hosts — these resolve to nothing in a
Copilot harness, so they are runtime correctness bugs, not just naming drift.
Genuine multi-provider documentation, security-rule key patterns
(`sk-ant-*`), and real SDK package names (`@anthropic-ai/claude-agent-sdk`) are
legitimate and go on the test's allowlist **with a per-entry justification** —
never as a silent escape hatch.

Enforced by `pkg/scaffold/no_claude_refs_test.go`.
<!-- enforced-by: pkg/scaffold/no_claude_refs_test.go -->

---

## Catalog & Template Parity

### DO NOT add a dogfood skill without mirroring it into the installer templates
Every skill under `.github/skills/<name>/` must also exist under
`pkg/scaffold/templates/skills/<name>/`, or be explicitly recorded as tech debt
in the `pendingMirror` list in the parity test. Skip this and `--guided`
installs silently ship without the skill.

Enforced by `pkg/scaffold/parity_test.go`.
<!-- enforced-by: pkg/scaffold/parity_test.go -->

### DO NOT add a skill template without registering it in the catalog
Every directory under `pkg/scaffold/templates/skills/` must be registered in
**exactly one** catalog slice (`coreSkillDirectories`,
`orchestratorSkillDirectories`, or `easterEggSkillDirectories`) in
`pkg/scaffold/catalog.go`. An unregistered template is silently excluded from
`--guided` installs; a double-registered one is a wiring bug.

Enforced by `pkg/scaffold/parity_test.go`.
<!-- enforced-by: pkg/scaffold/parity_test.go -->

### DO NOT hand-edit the `.github/prompts/*.prompt.md` shims
The VS Code Copilot Chat prompt shims are generated from the skill list by
`promptgen`. Edit the generator or the skill, then regenerate with
`go generate ./pkg/scaffold/...`. The dogfood shims must match `BuildPromptShim`
byte-for-byte, and there must be no orphan shim without a backing skill.

Enforced by `pkg/scaffold/parity_test.go`.
<!-- enforced-by: pkg/scaffold/parity_test.go -->

### DO NOT drop `land` or `takeoff` from the core-skills layer
The session-lifecycle pair must ship whenever the `core-skills` layer is
selected, and must not leak in via any other layer.

Enforced by `pkg/scaffold/parity_test.go`.
<!-- enforced-by: pkg/scaffold/parity_test.go -->

---

## Plugin Marketplace

### DO NOT hand-edit the generated plugin marketplace tree
`plugins/**`, `.github/plugin/marketplace.json`, and the root `marketplace.json`
are generated by `plugingen`. Change the source and regenerate with
`go run ./cmd/plugingen`; CI runs `go run ./cmd/plugingen -check` and fails if
the committed tree differs from what the generator would produce.

Enforced by `cmd/plugingen/main.go` (via `.github/workflows/ci.yml`).
<!-- enforced-by: cmd/plugingen/main.go, .github/workflows/ci.yml -->

---

## Supply Chain

### DO NOT use a floating GitHub Action reference
Pin every third-party Action in `.github/workflows/` to an immutable 40-char
commit SHA with a trailing human-readable version comment
(`uses: actions/checkout@<sha>  # v4.3.1`). Mutable tags (`@v4`, `@main`) can be
re-pointed at a malicious commit — the policy was adopted after the
litellm / Shai-Hulud / tj-actions incidents. Local (`./…`) and `docker://`
references are exempt.

Enforced by `.github/workflows/supply-chain.yml`.
<!-- enforced-by: .github/workflows/supply-chain.yml -->

---

## Tests & CI

### DO NOT write change-detector tests
A test that snapshots current data — a skill count, a version literal, an exact
list membership — fails every time that data legitimately changes. Assert
**invariants and contracts** instead. ATV's own tests model the better pattern:
allowlists assert that every exempted path still exists on disk, so a stale
entry becomes a failure rather than a silent future exemption.

See `pkg/scaffold/parity_test.go` (the stale-entry and graduation checks).

### DO NOT hardcode the repo root or working directory in tests
Derive the repository root from the test file's own location via the
`repoRoot(t)` helper so tests pass regardless of the current directory. Do not
read from `os.Getwd()` or a home-relative path.

See `pkg/scaffold/parity_test.go` (`repoRoot`).

---

## Session Workflow (`/land`, `/takeoff`)

### DO NOT merge a PR while landing
`/land` means **commit → push → open PR**. It does not mean merge. A PR is the
human review surface for agent work; never merge unless the user explicitly says
"merge this PR".

See `.github/skills/land/SKILL.md`.

### DO NOT `git add -A` or `git add .`
Stage specific files deliberately. Blanket staging risks committing `.env`,
credentials, or large binaries.

See `.github/skills/land/SKILL.md`.

### DO NOT bypass hooks to force a push or commit
No `--no-verify` / `--no-gpg-sign` unless the user explicitly asks. If a hook
fails, fix the root cause.

See `.github/skills/land/SKILL.md`.

### DO NOT rely on `backlog task list` as the primary source in `/takeoff`
When `backlog/config.yml` sets a `task_prefix`, the CLI silently returns only
the matching prefix and drops every other task. Scan the filesystem for task
files so the briefing surfaces every task across every prefix.

See `.github/skills/takeoff/SKILL.md`.

---

## Protected Artifacts (ATV Override Rules)

### DO NOT flag protected knowledge artifacts for deletion
Never propose deleting or "cleaning up" `docs/plans/`, `docs/solutions/`,
`docs/brainstorms/`, `compound-engineering.local.md`, or
`.github/skills/gstack/`. These are durable institutional knowledge, not scratch
files. The same rule means design docs go to `docs/brainstorms/` and plans use
the `docs/plans/` naming convention.

See `.github/copilot-instructions.md`.

---

## When You Discover a New Pitfall

1. Add an imperative `DO NOT` entry here with a one-to-two-sentence postmortem
   explaining the bite.
2. If it is machine-checkable, add (or extend) a guard — a Go test under
   `pkg/scaffold/` or a workflow under `.github/workflows/` — and record it with
   a `<!-- enforced-by: <repo-relative-path> -->` marker so
   `known_pitfalls_test.go` keeps the reference honest.
3. If it is judgement-based, link the relevant `SKILL.md` with a `See:` line.
4. Reference the entry from the related `SKILL.md`'s own `Pitfalls` section so
   the lesson is discoverable in context.
