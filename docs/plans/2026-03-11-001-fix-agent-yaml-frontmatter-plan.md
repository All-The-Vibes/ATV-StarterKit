---
title: "fix: Agent YAML frontmatter — remove deprecated/invalid attributes"
type: fix
status: completed
date: 2026-03-11
---

# fix: Agent YAML frontmatter — remove deprecated/invalid attributes

## Overview

All 28 `.agent.md` files in `.github/agents/` have 3 YAML frontmatter errors that VS Code reports as compile errors:

1. **`infer: true`** — deprecated in favour of `user-invocable` and `disable-model-invocation`
2. **`tools: ["*"]`** — unknown tool `*`
3. **`model: inherit`** — unknown model `inherit` (2 files use `model: haiku` instead)

These are Claude Code-specific attributes that don't exist in the GitHub Copilot agent schema. They need to be removed or replaced.

## Affected Files (28 total)

### Pattern A: `tools: ["*"]`, `infer: true`, `model: inherit` (26 files)

- [x] agent-native-reviewer.agent.md
- [x] ankane-readme-writer.agent.md
- [x] architecture-strategist.agent.md
- [x] best-practices-researcher.agent.md
- [x] bug-reproduction-validator.agent.md
- [x] code-simplicity-reviewer.agent.md
- [x] data-integrity-guardian.agent.md
- [x] data-migration-expert.agent.md
- [x] deployment-verification-agent.agent.md
- [x] design-implementation-reviewer.agent.md
- [x] design-iterator.agent.md
- [x] dhh-rails-reviewer.agent.md
- [x] figma-design-sync.agent.md
- [x] framework-docs-researcher.agent.md
- [x] git-history-analyzer.agent.md
- [x] julik-frontend-races-reviewer.agent.md
- [x] kieran-python-reviewer.agent.md
- [x] kieran-rails-reviewer.agent.md
- [x] kieran-typescript-reviewer.agent.md
- [x] pattern-recognition-specialist.agent.md
- [x] performance-oracle.agent.md
- [x] pr-comment-resolver.agent.md
- [x] repo-research-analyst.agent.md
- [x] schema-drift-detector.agent.md
- [x] security-sentinel.agent.md
- [x] spec-flow-analyzer.agent.md

### Pattern B: `tools: ["*"]`, `infer: true`, `model: haiku` (2 files)

- [x] learnings-researcher.agent.md
- [x] lint.agent.md

## Fix

For each file, remove the three invalid lines from YAML frontmatter:

**Before:**
```yaml
---
description: ...
tools:
  - "*"
infer: true
model: inherit
---
```

**After:**
```yaml
---
description: ...
---
```

The `description` field is the only valid/required frontmatter attribute for Copilot `.agent.md` files.

## Acceptance Criteria

- [x] All 28 agent files have only `description` in their YAML frontmatter
- [x] `tools`, `infer`, and `model` lines removed from all files
- [x] Zero compile errors reported by VS Code for the agents folder (1 pre-existing broken link in learnings-researcher remains — unrelated)
- [x] Agent descriptions preserved exactly (no content changes)
- [x] Agent body content preserved exactly (no content changes)
