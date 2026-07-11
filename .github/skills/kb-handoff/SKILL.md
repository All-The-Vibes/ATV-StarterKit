---
name: kb-handoff
description: Compact the current session into a KB handoff restart packet for a fresh agent. Use when the user asks for a handoff, restart packet, fresh-session note, session transfer, pause/resume note, or wants the next session to continue from local project memory.
argument-hint: "[what the next session should focus on]"
---

# KB Handoff

Create a compact restart packet so a fresh session can continue without relying on chat history.

This is the KB adaptation of Matt Pocock's `handoff` skill:

- compact the current conversation for another agent;
- suggest the skills the next session should invoke;
- reference existing artifacts instead of duplicating them;
- redact secrets, tokens, passwords, credentials, PII, and private keys.

KB changes the default storage target. Matt's original writes to the OS temp directory. KB should write to the active work repo when there is one, because `kb-map` and `kb-start` recover state from repo-local memory.

## Non-Goals

- Do not create a plan. Classify the suggested route instead.
- Do not bootstrap project memory just to create a handoff.
- Do not write project-work handoffs into a portable skill-bundle repo by accident.

## Root Rule

Resolve the active work root first:

```powershell
git rev-parse --show-toplevel
```

Valid targets:

- a consuming project repo where the current work is happening;
- a portable skill-bundle repo only when the handoff is explicitly about maintaining that skill bundle.

Invalid targets for project-work handoffs:

- drive roots such as `E:\`;
- home folders;
- global skill folders such as `.copilot`, `.codex`, or `.agents`;
- a portable skill-bundle repo when the resumed work belongs to another project;
- `C:\Users\marowe\.copilot\handoffs` or any other global handoff folder.

If the current repo is not the active work repo, ask for the project path before writing. Do not search sibling repos or the whole drive to guess.

If there is no valid project root and the user only needs a conversation transfer, write a temporary handoff in the OS temp directory and print the path. Do not update `todo.md`, create `docs/handoffs/`, or run `kb-map-bootstrap` in that fallback mode.

## Output Path

Preferred repo-local path:

```text
<project-root>/docs/handoffs/active/YYYY-MM-DD-<short-topic>.md
```

Temporary fallback path:

```text
<os-temp>/handoff-<short-topic>.md
```

Create `docs/handoffs/active/`, `docs/handoffs/parked/`, and `docs/handoffs/done/` only inside a validated project root.

When updating an existing handoff, read it before writing and merge current truth instead of overwriting useful details.

## What To Include

Keep the handoff under 1200 words unless the user explicitly asks for more.

1. **Purpose** — what the next session should do.
2. **Suggested route** — `kb-start`, `kb-map`, `kb-fix`, `kb-brainstorm`, `kb-plan`, `kb-work`, `kb-finalize`, `kb-complete`, or another exact skill.
3. **Suggested skills** — skills the next session should invoke and why.
4. **Current state** — branch, repo root, relevant manifest/plan/todo status.
5. **Files to read first** — exact repo-local paths or temp artifact path.
6. **Existing artifacts** — PRDs, plans, ADRs, issues, commits, diffs, screenshots, traces, or logs to reference instead of duplicating.
7. **Work completed** — compact facts only.
8. **Next action** — the recommended `kb-start <task or handoff>` prompt.
9. **Blockers / HITL** — exact missing input, access, decision, or failing command.
10. **Verification state** — tests, QA, snapshots, or proof already run, including command and exit status when known.

## Todo Pointer

Update `todo.md` only when all are true:

- the handoff is in a validated project root;
- `todo.md` already exists or project memory has already been intentionally bootstrapped;
- the handoff represents active or blocked work.

- add or update a compact row pointing to `docs/handoffs/active/<file>.md`;
- use `🔒 blocked` for dependency/tool/access waits;
- use `🛑 human-required` for decisions or inputs only the user can provide;
- do not paste the whole handoff into `todo.md`.

Do not create `todo.md` solely for a handoff. If the project has no KB memory yet, write the handoff and set the next action to `kb-map setup <project focus>`.

## Rules

- A handoff is a restart packet, not an executable plan.
- If the handoff only has phases or broad next steps, set `Suggested route: kb-plan`.
- If it links a valid `docs/plans/*-kb-*-manifest.md`, set `Suggested route: kb-work`.
- If intent is unclear, set `Suggested route: kb-brainstorm`.
- If the next session only needs orientation, set `Suggested route: kb-start <handoff path>`.
- If the target repo has no `todo.md` or `docs/context/PROJECT.md`, say that `kb-map setup` is required before normal KB startup.
- Prefer exact file paths, commands, errors, and proof artifacts over prose.
