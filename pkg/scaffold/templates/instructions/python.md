# Project Conventions

This is a Python project using the ATV (Agentic Tool & Workflow) Starter Kit.

## Python Conventions

- Follow PEP 8 style guide
- Use type hints for all function signatures
- Use `pathlib.Path` over `os.path`
- Prefer dataclasses or Pydantic models over dicts
- Use `ruff` for linting and formatting

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
