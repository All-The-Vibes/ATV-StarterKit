# Project Conventions

This is a Ruby on Rails project using the ATV (Agentic Tool & Workflow) Starter Kit.

## Rails Conventions

- Follow Rails conventions: fat models, thin controllers
- Use service objects for complex business logic
- Use concerns for shared model behavior
- Prefer Hotwire (Turbo + Stimulus) over heavy JS frameworks
- Use `bin/rails test` for testing

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
