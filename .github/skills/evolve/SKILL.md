---
name: evolve
description: Promote mature instincts (confidence > 0.8) into full Copilot skills that get auto-discovered. Clusters related instincts and generates SKILL.md files in .github/skills/.
argument-hint: "[instinct domain or blank for all mature instincts]"
---

# /evolve — Promote Instincts to Skills

Transform mature instincts into full Copilot skills. When an instinct reaches high confidence through repeated observation, it has proven valuable enough to become a permanent part of the project's skill set.

## When to Use

- When `/instincts` shows patterns with ★ (confidence > 0.8)
- When you want to formalize a recurring pattern into a discoverable skill
- Periodically to graduate well-established project conventions

## Execution Flow

### Step 1: Identify Candidates

Read `docs/context/kb/instincts/project.yaml` and all `docs/context/kb/instincts/scoped/<scope-path>.yaml` files. For each file, filter candidates by:
- Confidence > 0.85
- Observations > 5
- `last_seen` within the last 90 days (rejects stale instincts even if confidence is high)
- Not already evolved (check `docs/context/kb/instincts/archive/`)

If no candidates found:
```
No instincts ready to evolve yet.
Run /learn to build confidence, or check /instincts for current status.
Instincts need confidence > 0.85, 5+ observations, and activity within 90 days to evolve.
```

### Step 2: Cluster Related Instincts

Group candidates by domain. Each cluster becomes one skill:

```
Evolution candidates:

  Cluster 1: "Go Error Handling" (error-handling domain)
    ★ always-wrap-errors      0.9  15 obs
    ★ sentinel-errors         0.85  8 obs
    → Will generate: .github/skills/learned-go-error-handling/SKILL.md

  Cluster 2: "Testing Conventions" (testing domain)
    ★ table-driven-tests      0.85  12 obs
    → Will generate: .github/skills/learned-testing-conventions/SKILL.md

Proceed with evolution? (Copilot will generate the skills)
```

### Step 3: Human Approval Gate

Numeric maturity is necessary but not sufficient in this portable bundle.
Before generating, committing, or syncing any `learned-*` skill, ask:

```
Promote these generated learned-* skills and allow them to be committed or
synced from this portable bundle? yes/no
```

If the answer is not an explicit yes:

- Do not create or modify `.github/skills/learned-*`.
- Do not archive the source instincts as evolved.
- Report the candidates as ready but unapproved.
- Leave the instincts in their source file (`docs/context/kb/instincts/project.yaml` or the relevant scoped file) so evidence can continue
  accumulating or decay naturally.

If approved, continue. Generated skills are still drafts and must be reviewed
before commit.

### Step 3.5: Measured Adoption Gate

If the candidate changes agent behavior, scoring, routing, decomposition, or
promotion rules, require a measured adoption result before generating or syncing
shared/project/global behavior:

```bash
go run ./cmd/kbcheck learning-adoption --result-path <results.json>
```

Only `ADOPT_ELIGIBLE` candidates may auto-promote into shared/project/global
surfaces. A rejected candidate may stay local/scoped or experimental, but it
must not be sold as a learned improvement.

### Step 4: Generate SKILL.md

For each cluster, generate a `SKILL.md` file:

```yaml
---
name: learned-[domain-name]
description: "[Auto-generated] Project conventions for [domain] learned from [N] observations across [M] sessions."
---
```

The skill content should:
1. Describe when to apply these conventions
2. List each instinct as a concrete guideline with examples
3. Reference the evidence (commit hashes, observation counts)
4. Mark as auto-generated so humans know to review

**Naming convention:** `learned-` prefix so generated skills are visually distinct from hand-written ones.

**Output path — promotion target (scope-aware):**

- **Global / project-tier instinct** (`scope: global` or `scope: project` in `project.yaml`):
  promote to a new global skill at `.github/skills/learned-<domain>/SKILL.md`.
- **Scoped (workflow / component-tier) instinct** (`scope: <workflow>` or `scope: <workflow/component>` in a `scoped/*.yaml` file):
  promote into that component's **owned surface** — its existing `SKILL.md` (as an appended convention section) or its `config/calibration.yaml`/doc — NOT a new global skill.
  A workflow-tier instinct that matured via promotion-on-recurrence evolves at that workflow's surface.

**Decision rule:** global/project-tier instinct → global skill; scoped instinct → component-owned surface.

### Step 5: Archive Evolved Instincts

Move evolved instincts from their source file to `docs/context/kb/instincts/archive/evolved-YYYY-MM-DD.yaml` with metadata:
```yaml
evolved_to: .github/skills/learned-go-error-handling/SKILL.md
evolved_at: 2026-04-06
```

### Step 6: Report

```
Evolution complete!

Generated skills:
  ✅ .github/skills/learned-go-error-handling/SKILL.md (from 2 instincts)
  ✅ .github/skills/learned-testing-conventions/SKILL.md (from 1 instinct)

Archived 3 instincts to docs/context/kb/instincts/archive/

These skills will be auto-discovered by Copilot in the next session.
Review the generated files and adjust as needed — they're a starting point.

Remaining instincts: X active (run /instincts to see)
```

## Important Notes

- Generated skills use `learned-` prefix — easy to spot and edit
- Always review generated skills before committing — they're a draft
- In this portable bundle, generated `learned-*` skills require explicit human
  approval before commit or sync
- Evolved instincts are archived, not deleted, so history is preserved
- If a generated skill doesn't feel right, delete it and the instincts will re-accumulate
