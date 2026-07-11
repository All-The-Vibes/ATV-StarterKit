---
name: kb-task
description: "First-principles autonomous KB task runner. Use when the user says /kb-task, asks the app/agent to do a task from first principles, wants the agent to figure out the route and continue until done, or wants one bounded app/workflow change handled end-to-end without choosing between kb-fix, kb-brainstorm, kb-plan, kb-work, klfg, or kb-complete."
argument-hint: "[bounded task or outcome to complete]"
---

# KB Task

Reason from first principles, choose the smallest correct KB route, and continue until the task is verified, completed, or honestly blocked.

This is a task runner, not a separate implementation lane. It uses `kb-map` for repo memory, applies `kb-first-principles` reasoning directly, then delegates to the right KB skill.

## Contract

- Keep moving without asking the user to choose ceremony.
- Challenge weak assumptions before building on them.
- Use tools to verify checkable facts.
- Ask only for information the agent cannot safely infer or access.
- Stop only for HITL, unsafe/destructive action, missing credentials/access, unresolved product choice, blocked verification, or explicit user stop.
- Do not mark done until deterministic proof exists or the remaining gap is clearly human-required.

## Workflow

1. **Map first**
   - Run `kb-map lookup <task>`.
   - If repo memory is missing, let `kb-map` invoke setup/bootstrap in the active project root.
   - If the current root is a portable skill bundle but the task belongs to another app, ask for the app project path before writing project memory or handoffs.

2. **Frame from first principles**
   - State the real goal in one sentence.
   - Identify the core assumptions, constraints, and success criteria.
   - Verify assumptions that are factual, recent, high-impact, or cheaply checkable.
   - Name tradeoffs when there are multiple reasonable paths.
   - Push back if the request implies a route that would skip required evidence or create rework.

3. **Choose the route**

   | Task shape | Route |
   |---|---|
   | Narrow known bug, failing test, or obvious contained fix | `kb-fix` |
   | Broken behavior needs autonomous diagnosis from logs, browser evidence, commands, or unclear symptoms | `kb-troubleshoot` |
   | Clear bounded task that needs slices before implementation | `kb-complete` |
   | Unclear product/technical behavior or high path dependency | `kb-complete` (routes through brainstorm first) |
   | Long-lived objective that should persist across sessions or run for days | `kb-goal` |
   | User wants full feature from idea to configured endpoint | `kb-complete` |
   | Valid manifest already exists and should reach configured endpoint | `kb-complete <manifest>` |
   | Work is implemented and needs only post-work quality gates | `kb-finalize` |
   | Large multi-subsystem initiative | `kb-epic` |
   | Release, PR, direct integration, or final readiness | `kb-complete` |
   | External docs or current ecosystem behavior could change the answer | `kb-research`, then resume route selection |

4. **Execute until done**
   - Run the selected route immediately when safe.
   - If the first route produces a required artifact, continue to the next phase without asking unless a gate blocks.
   - Preserve phase order: brainstorm before plan, plan before work, work before finalize, finalize before delivery.
   - Treat "don't ask many questions" or "go straight to work" as execution intent, not permission to skip slices. If no valid manifest exists, run `kb-plan` before `kb-work`.
   - If the route changes because evidence contradicts the initial classification, record why and switch.

5. **Verify**
   - Run the repo's deterministic checks for touched behavior.
   - For UI/user-visible changes, verify through the rendered UI with real navigation, clicks, inputs, and DOM assertions when available.
   - Re-run checks after review or repair changes.
   - Record command, exit status, artifact path, or exact blocker.

6. **Close**
   - Update project memory when behavior, routes, commands, architecture, or active state changed.
   - Create a `kb-handoff` only if work must continue in a fresh session or is blocked.
   - Finish with a concise summary, files changed, verification run, and remaining blockers if any.

## Autonomy Boundaries

Continue without asking for:

- choosing between KB skills when the route is evident;
- running safe local reads, tests, builds, linters, and browser checks;
- fixing safe/actionable review findings;
- updating repo-local memory after meaningful workflow changes.

Stop and ask for:

- credentials, MFA, paid/external access, or private data not present locally;
- destructive operations, force pushes, production writes, deploys, or irreversible data changes;
- product decisions with multiple reasonable outcomes;
- subjective design/taste calls after tradeoffs are explained;
- action outside the active repo or approved sync targets.

## Output

During execution, keep updates short:

- route chosen and why;
- evidence checked;
- current gate or blocker;
- next action.

Final output must include:

- route actually run;
- important files changed;
- verification proof;
- unresolved human-required items, if any.
