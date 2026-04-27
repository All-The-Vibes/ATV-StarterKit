---
name: unslop
description: "Unified de-slop pass: code simplification + comment rot detection + design slop check. Run after completing features or before PRs to strip AI-generated generic patterns."
argument-hint: "[blank for recent changes, or file/directory path]"
---

# /unslop — Strip AI Slop from Your Codebase

Three parallel analysis passes that detect and report AI-generated "slop" — generic, template-looking, over-enthusiastic output that makes code, comments, and UI feel artificial.

## When to Use

- After completing a feature — check for slop before PR
- Before code review — pre-clean your changes
- When code "feels AI-generated" but you can't pinpoint why
- Periodic codebase hygiene — run on a directory
- After a long AI-assisted session to audit quality

## Argument Parsing

Parse `$ARGUMENTS` for these tokens:

| Token | Example | Effect |
|-------|---------|--------|
| `fix` | `/unslop fix` | Auto-apply safe fixes after reporting |
| `<path>` | `/unslop src/components/` | Scope to specific file or directory |
| (none) | `/unslop` | Analyze all files changed since the base branch |

## Execution Flow

### Stage 1: Determine Scope

**If a file or directory path is provided:**

Scope to that path. Use `find` or glob to list all code files under it.

**If no argument (default):**

Determine changed files since the base branch:

```bash
BASE=$(git merge-base HEAD origin/main 2>/dev/null || git merge-base HEAD origin/master 2>/dev/null || echo "HEAD~10")
echo "FILES:" && git diff --name-only $BASE
echo "DIFF:" && git diff -U5 $BASE
```

**Classify files in scope:**

| File extensions | Passes to run |
|----------------|---------------|
| `.ts`, `.tsx`, `.js`, `.jsx`, `.py`, `.go`, `.rb`, `.rs`, `.java`, `.cs`, `.swift`, `.kt` | Code Slop + Comment Rot |
| `.css`, `.scss`, `.less`, `.tsx`, `.jsx`, `.html`, `.vue`, `.svelte` | + Design Slop |
| `.json`, `.yaml`, `.yml`, `.toml`, `.md` | Code Slop + Comment Rot only |

If no UI/style files are in scope, skip the Design Slop pass entirely.

Announce the scope:

```
De-slop scope: 14 files changed since origin/main
Passes: Code Slop | Comment Rot | Design Slop (3 UI files detected)
```

### Stage 2: Parallel Analysis

Launch three sub-agents IN PARALLEL. Each receives the file list and diff content, and returns structured findings as text.

<parallel_tasks>

#### Pass 1: Code Slop Detector

You are a code simplification expert. Analyze the provided files for AI-generated slop patterns. Focus ONLY on slop — not correctness, security, or architecture.

**What to flag:**

1. **Unnecessary complexity**
   - Deep nesting (>3 levels) that could use early returns
   - Nested ternary operators — use if/else or switch
   - Dense one-liners that sacrifice readability for brevity

2. **Redundant abstractions**
   - Interfaces/types used only once — inline them
   - Wrapper functions that add no logic — call the wrapped function directly
   - Abstract base classes with a single implementation
   - Premature generalization ("just in case" extensibility points)

3. **YAGNI violations**
   - Features not required by current use cases
   - Configuration options nobody uses
   - Generic solutions for specific problems
   - "Future-proofing" code that adds complexity now

4. **Dead weight**
   - Commented-out code blocks (>3 lines)
   - Unused imports, variables, or functions
   - Duplicate error checks (caller already validates)
   - Defensive code that can never trigger (type system prevents it)

5. **Over-engineering**
   - Factory patterns for creating a single type
   - Strategy patterns with one strategy
   - Event systems for synchronous single-consumer flows
   - Dependency injection where direct instantiation is clearer

**Return format:**
```json
{
  "pass": "code-slop",
  "findings": [
    {
      "file": "src/auth.ts",
      "line": 42,
      "issue": "Nested ternary — use if/else for readability",
      "severity": "medium",
      "fix_safe": true,
      "suggested_fix": "Replace ternary chain with if/else block"
    }
  ]
}
```

#### Pass 2: Comment Rot Detector

You are a technical documentation expert specializing in comment quality. Analyze the provided files for comment rot — inaccurate, redundant, or AI-generated filler comments.

**What to flag:**

1. **Obvious restatements**
   - `// increment counter` above `counter++`
   - `// return the result` above `return result`
   - `// set the name` above `this.name = name`
   - Comments that repeat the function/variable name in prose

2. **AI-generated filler phrases** (hard bans — flag these immediately)
   - "This function is responsible for handling..."
   - "The following code implements..."
   - "This is a comprehensive solution that..."
   - "This method provides a robust and scalable..."
   - "leverages" or "utilizes" (when "uses" works fine)
   - "seamlessly integrates"
   - "In today's rapidly evolving..."
   - "game-changer", "revolutionary", "cutting-edge"
   - "This class encapsulates the logic for..."
   - "Ensures proper handling of..."

3. **Factual inaccuracy**
   - Documented parameters that don't match the function signature
   - Return type descriptions that don't match the actual return
   - Described behavior that doesn't match the code logic
   - Edge case documentation for cases not actually handled

4. **Stale comments**
   - TODOs/FIXMEs for work that's already been done
   - References to removed/renamed functions, classes, or files
   - Version-specific notes for versions no longer supported
   - "Temporary" markers on permanent code

5. **Over-documentation**
   - JSDoc/docstrings on trivial getters/setters
   - Multi-line comments on self-explanatory one-liners
   - Repeating type information already in the signature
   - Section divider comments (`// ==================`)

**Return format:**
```json
{
  "pass": "comment-rot",
  "findings": [
    {
      "file": "src/auth.ts",
      "line": 10,
      "issue": "\"This function handles authentication\" — restates the function name",
      "severity": "low",
      "fix_safe": true,
      "suggested_fix": "Remove comment — function name is self-documenting"
    }
  ]
}
```

#### Pass 3: Design Slop Detector

**Only run this pass when UI/style files are in scope.**

You are a design quality expert. Analyze the provided UI files for AI-generated visual slop — generic, template-looking patterns that signal "an AI made this, not a designer."

**What to flag:**

1. **Generic color patterns**
   - Purple-to-blue gradients (the AI default palette)
   - Gratuitous gradients on everything (buttons, cards, backgrounds)
   - Safe gray-on-white with one decorative accent color
   - Unintentional color usage (decorative, not semantic)

2. **Template layouts**
   - Default card grids with uniform spacing and no hierarchy
   - Generic hero: centered headline + subtitle + gradient blob + CTA
   - Dashboard-by-numbers: sidebar + cards + charts with no point of view
   - Uniform radius, spacing, and shadows across every component

3. **Missing interaction states**
   - No hover states on interactive elements
   - No focus states (accessibility gap)
   - No active/pressed states
   - No loading/empty/error states

4. **Lazy defaults**
   - Unmodified library defaults (default shadcn/Tailwind without customization)
   - Default font stacks with no intentional pairing
   - Glassmorphism cards that serve no UI purpose
   - Rounded corners on elements that shouldn't be rounded (tables, code blocks)
   - Excessive scroll-triggered animations

5. **No visual hierarchy**
   - Flat layouts with no layering, depth, or motion
   - Uniform emphasis on everything (nothing stands out)
   - No intentional rhythm in spacing

**Return format:**
```json
{
  "pass": "design-slop",
  "findings": [
    {
      "file": "src/Hero.tsx",
      "line": 22,
      "issue": "Generic purple-to-blue gradient — replace with brand palette",
      "severity": "medium",
      "fix_safe": false,
      "suggested_fix": "Define an intentional brand gradient in design tokens"
    }
  ]
}
```

</parallel_tasks>

### Stage 3: Merge & Deduplicate

**WAIT for all Stage 2 sub-agents to complete.**

1. Collect findings from all passes that ran
2. Deduplicate: if two passes flag the same file+line (within 3 lines), keep the more specific finding and note both passes
3. Sort by severity: High → Medium → Low
4. Group by pass for the report

### Stage 4: Present Report

Format the consolidated report using pipe-delimited markdown tables:

```
De-slop Report
==============
Scope: [N] files changed since [base]
Passes: Code Slop [✓|✗] | Comment Rot [✓|✗] | Design Slop [✓|skipped]

## Code Slop ([N] findings)

| # | File | Line | Issue | Severity |
|---|------|------|-------|----------|
| 1 | path | line | description | High/Medium/Low |

## Comment Rot ([N] findings)

| # | File | Line | Issue | Severity |
|---|------|------|-------|----------|
| 1 | path | line | description | High/Medium/Low |

## Design Slop ([N] findings)

| # | File | Line | Issue | Severity |
|---|------|------|-------|----------|
| 1 | path | line | description | High/Medium/Low |

─────────────────────────────
Summary: [N] findings ([H] High, [M] Medium, [L] Low)
Potential LOC reduction: ~[N] lines
```

Omit any pass section with zero findings. If all passes return zero findings:

```
De-slop Report: Clean! No slop detected in [N] files.
```

### Stage 5: Auto-Fix (only if `fix` argument was provided)

If the user invoked `/unslop fix`:

1. Collect all findings where `fix_safe: true`
2. Apply fixes in file order:
   - **Code Slop safe fixes:** Remove commented-out code blocks, remove unused imports
   - **Comment Rot safe fixes:** Delete obvious restatement comments, remove stale TODOs
3. Do NOT auto-fix:
   - Design slop (requires design judgment)
   - Factually inaccurate comments (requires understanding intent)
   - YAGNI violations (requires knowing the roadmap)
   - Abstractions (requires understanding the broader architecture)
4. Report what was fixed:

```
Auto-fix applied:
  ✓ Removed 3 commented-out code blocks (42 lines)
  ✓ Deleted 5 obvious-restatement comments
  ✓ Removed 2 stale TODOs

  Remaining (manual review needed): 4 findings
```

If `fix` was NOT provided but fixable findings exist, suggest it:

```
Tip: Run /unslop fix to auto-apply [N] safe fixes
```

## Severity Guide

| Level | Meaning | Examples |
|-------|---------|---------|
| **High** | Actively misleading or creates maintenance burden | Inaccurate comment, missing hover states, large dead code block |
| **Medium** | Noticeable slop that reduces quality | Unnecessary abstraction, AI filler phrase, generic gradient |
| **Low** | Minor quality improvement | Restatement comment, unused import, over-documentation |

Slop is never "Critical" — it's a quality concern, not a correctness or security issue.

## Quality Gates

Before presenting findings:

1. **Every finding must be actionable.** Don't say "could be simpler" — say what to change and where.
2. **No false positives from skimming.** Verify the abstraction isn't used elsewhere before flagging it. Verify the comment is truly redundant.
3. **Line numbers must be accurate.** Check each cited line against file content.
4. **Respect project conventions.** If the project uses JSDoc everywhere, don't flag JSDoc as over-documentation.
5. **Don't flag generated code.** Skip files in `dist/`, `build/`, `node_modules/`, `.next/`, `vendor/`, or similar generated directories.

## Notes

- This skill is read-only by default. It reports but does not edit files unless `fix` is specified.
- Design Slop pass is automatically skipped for backend-only projects.
- Works on any language/framework — the slop patterns are universal.
- Pairs well with `/ce-review` (which checks correctness) — `/unslop` checks aesthetics and quality.
