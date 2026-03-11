# Project Conventions

This is a TypeScript project using the ATV (Agentic Tool & Workflow) Starter Kit.

## TypeScript Conventions

- Use strict TypeScript (`strict: true` in tsconfig)
- Prefer `interface` over `type` for object shapes
- Use `readonly` for immutable data
- Avoid `any` — use `unknown` with type guards
- Use named exports over default exports

## Available Workflows

- `/ce-brainstorm` — Explore what to build through collaborative dialogue
- `/ce-plan` — Create a structured implementation plan
- `/ce-work` — Execute the plan with quality checks
- `/ce-review` — Multi-agent code review
- `/ce-compound` — Document solutions for future reference
- `/lfg` — Full autonomous pipeline (plan → work → review)

## Documentation Structure

- `docs/plans/` — Implementation plans
- `docs/brainstorms/` — Brainstorm documents
- `docs/solutions/` — Documented solutions
