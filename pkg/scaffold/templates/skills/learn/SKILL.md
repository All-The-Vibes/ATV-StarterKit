---
name: learn
description: Extract reusable patterns from recent work into instincts. Run after completing features, fixing bugs, or at session end to capture what the project learned.
argument-hint: "[recent work summary or blank for session observations]"
---

# /learn — Extract Patterns into Instincts

Analyze recent work (observations, git history, solutions) and extract reusable "instincts" — small learned behaviors with confidence scoring.

## When to Use

- After completing a feature or fixing a bug
- At the end of a coding session
- When you want to capture a pattern you noticed
- Periodically to keep the project's knowledge current

## Execution Flow

### Step 1: Gather Evidence

Run these in parallel to collect data:

1. **Git history** — `git log --oneline -20` for recent commits
2. **Recent diffs** — `git diff HEAD~5..HEAD --stat` for what changed
3. **Observations** — Read `.atv/observations.jsonl` for tool use patterns from hooks
4. **Existing instincts** — Read `.atv/instincts/project.yaml` (create if missing)
5. **Solutions** — Read `docs/solutions/` for documented patterns
6. **Steering memory** — Read any current goal or manifest steering-memory path
   named by the caller, such as a `Live Steering` ledger section or
   `docs/context/operations/steering/<slug>.md`

### Step 2: Analyze Patterns

Look for recurring patterns across the evidence:

**Code style patterns:**
- Error handling conventions (wrapping, custom types, sentinel errors)
- Naming conventions (variable, function, file naming)
- Import organization preferences
- Comment style and documentation patterns

**Workflow patterns:**
- Test-first vs test-after behavior
- Commit granularity preferences
- Branch naming conventions
- Review practices

**Architecture patterns:**
- Package/module organization
- Dependency injection style
- Interface usage patterns
- Configuration management approach

**Tool usage patterns** (from observations.jsonl):
- Frequently used shell commands
- Common file editing sequences
- Preferred build/test commands

**Steering feedback patterns** (from curated steering memory):
- Permanent scope constraints or false-positive areas
- Reviewer preferences that should change future target selection
- Controller-selection guidance for recurring goals
- Durable "do this next time" feedback repeated across runs

Do not learn from raw transcripts, unclassified PR comments, or one-off
instructions. If feedback only applied to the current PR, leave it out of
project instincts.

### Step 2.5: Apply Recency Decay

Before creating or updating instincts, apply time-based decay to all existing entries:

1. For each instinct in `.atv/instincts/project.yaml`:
   - Calculate days since `last_seen`
   - Apply decay: `new_confidence = confidence × 0.5^(days_since_last_seen / 90)`
   - Update the confidence value in place

2. **Archive stale instincts:**
   - If decayed confidence falls below 0.3, move the instinct to `.atv/instincts/archive/YYYY-MM-DD-archived.yaml`
   - Add `archived_reason: confidence decayed below 0.3 (last seen: <date>)`
   - Remove from `project.yaml`

3. Write updated confidence values back to `project.yaml` before proceeding to Step 3.

**Half-life rationale:** 90 days balances stability (project conventions rarely change weekly) with freshness (patterns unused for 6+ months are likely obsolete). At 90 days, an unobserved instinct at 0.85 decays to:
- 30 days: 0.68
- 90 days: 0.43
- 180 days: 0.21 (archived)

### Step 3: Create or Update Instincts

For each new pattern discovered, create an instinct entry.
For patterns that match existing instincts, increase confidence and observation count.

Before creating or updating an instinct from feedback, classify it:

| Route | Use When | Durable Output |
|---|---|---|
| `current-only` | Feedback only changes the active PR/session | manifest or PR note only |
| `steering-memory` | Feedback should steer future target selection but is not yet a broad project instinct | goal ledger or `docs/context/operations/steering/<slug>.md` |
| `observation` | Feedback is an evidence point for later pattern extraction | `.atv/observations.jsonl` |
| `landmine-candidate` | Feedback exposes a concrete repo-specific trap | `docs/context/landmines.md` candidate with strict evidence |
| `instinct-evidence` | Repeated evidence supports a project-wide behavior | `.atv/instincts/project.yaml` |

One verified high-severity trap may become a landmine candidate, but ordinary
preferences need repeated evidence before becoming instincts. Steering memory is
the middle layer: it changes future loop behavior without pretending the pattern
is ready to become a reusable skill.

**Instinct format** (YAML in `.atv/instincts/project.yaml`):

```yaml
instincts:
  - id: kebab-case-unique-id
    trigger: "when [specific situation]"
    behavior: "do [specific action]"
    confidence: 0.5
    domain: code-style|testing|architecture|error-handling|workflow|tooling
    observations: 1
    first_seen: YYYY-MM-DD
    last_seen: YYYY-MM-DD
    evidence:
      - "commit abc123: wrapped all errors with fmt.Errorf"
      - "observed 3 times in observations.jsonl"
```

**Confidence rules:**
- New instinct starts at 0.5
- Each additional observation: +0.1 (capped at 0.95)
- Contradictory evidence: -0.15
- No observations for 30 days: -0.1
- Minimum: 0.1 (below this, remove the instinct)

### Step 3.5: Capture Landmine Candidates

A landmine is not a generic lesson. It is a verified repo-specific trap the
model is likely to miss without an explicit warning.

Only create a landmine candidate when the evidence shows at least one of:

- the model already made the mistake;
- the repo convention conflicts with common defaults;
- a command, sync path, runtime, auth mode, or generated artifact has a specific
  failure mode;
- a workflow gate is likely to be skipped;
- the trap is high-cost, destructive, or hard to notice from code alone.

Landmine candidates must include:

```yaml
landmine:
  severity: low|medium|high|critical
  owner_surface: "<skill, script, doc, generator, fixture, or workflow>"
  failure_mode: "<specific mistake likely without the warning>"
  evidence:
    - "<file, command, review finding, failing test, or observed mistake>"
  fix_condition: "<what change retires the landmine>"
  verification: "<command, eval, test, or review check proving the fix>"
```

Reject candidates that only say to test, read the code, keep things simple, or
follow normal engineering practice. Those are not landmines.

High-severity landmines may be recorded from one verified observation, but skill
promotion still belongs to `/evolve` and requires its promotion gate.

**Important constraints:**
- Maximum 50 active instincts per project
- Each instinct must be atomic — one trigger, one behavior
- Triggers must be specific (not "when writing code")
- Behaviors must be actionable (not "write good code")
- Evidence must cite specific commits or observations

### Step 4: Write Results

1. Write updated `.atv/instincts/project.yaml`
2. Ensure `.atv/instincts/` directory exists

### Step 5: Report

Show a summary:

```
Learning complete!

New instincts:
  + always-wrap-errors (0.5) — wrap errors with fmt.Errorf using %w
  + table-driven-tests (0.5) — use table-driven test pattern for Go tests

Updated instincts:
  ↑ prefer-early-returns (0.6 → 0.7) — 1 new observation
  ↑ run-tests-before-commit (0.7 → 0.8) — 2 new observations

Ready to evolve (confidence > 0.8):
  ★ run-tests-before-commit — consider /evolve to generate a skill

Total: X instincts (Y new, Z updated)
Instinct file: .atv/instincts/project.yaml
```

## Notes

- Instincts are project-scoped and committed to the repo — the whole team benefits
- Run `/instincts` to see all learned patterns
- Run `/evolve` when instincts reach high confidence to generate full skills
- Optional observer hooks such as `.github/hooks/copilot-hooks.json` can capture tool use data when a consuming repo installs them; this portable bundle does not ship that hook file
