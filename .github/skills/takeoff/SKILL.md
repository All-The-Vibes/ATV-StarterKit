---
name: takeoff
description: Session kickoff protocol - the opposite of "land the plane". Use this skill whenever the user says "takeoff", "take off", "/takeoff", "starting a new session", "what should I work on", "kickoff", "what's next", or wants a prioritized view of the backlog at the start of work. Surfaces the top-priority backlog tasks as a table with status, dependencies, and blockers pulled from Backlog.md (CLI + MCP). Use this at the start of a coding session to orient on what to work on.
argument-hint: "[optional: filter by tag, assignee, or count e.g. '--top 5' or '--mine']"
---

# Take Off Protocol

The session kickoff counterpart to "land the plane". Where `land the plane` closes out work (commit, push, PR, handoff), `/takeoff` *starts* work by giving a crisp, prioritized picture of the backlog so the user knows what to pick up.

## Trigger

Any variation of:
- `/takeoff`, `takeoff`, `take off`, `let's take off`, `time to take off`, `ready to take off`
- `what should I work on`, `what's next`, `kickoff`, `start a new session`, `show me the backlog`

When in doubt and the user is at the *start* of work (not finishing), invoke this protocol.

## What this skill produces

A short, scannable briefing with four sections:

1. **Top Priority Table** — the N highest-priority actionable tasks
2. **Blocked / Dependent** — tasks that exist but cannot be started
3. **In Progress** — tasks already underway (not to be double-claimed)
4. **Recommendation** — one sentence: "Start with `X` because …"

Keep it terse. The user wants to pick a task and go, not read a report.

## Execution

### Step 1: Pick the source of work

Decide where the actionable items come from. Two paths, mutually exclusive — never both:

```bash
if [ -d backlog ]; then
  echo "backlog present — using Backlog.md"
  source="backlog"
elif [ -d docs/plans ] && ls docs/plans/*.md >/dev/null 2>&1; then
  echo "no backlog/, falling back to docs/plans/"
  source="plans"
else
  echo "no backlog/ and no docs/plans/ — nothing to surface"
  source="none"
fi
```

- **`source=backlog`** → continue to Step 2.
- **`source=plans`** → skip Steps 2–3 (they are backlog-specific) and go straight to Step 4, summarizing the plan files in priority order. Tell the user you're using the docs/plans/ fallback because no backlog exists.
- **`source=none`** → tell the user there's nothing to surface, then jump to Step 6 to emit the banner. The successful-no-work path still counts as a successful completion.

### Step 2: Pull the backlog

**Important:** do **not** rely on `backlog task list` as the primary source. When `backlog/config.yml` sets a `task_prefix`, the CLI filters output to that prefix only and silently drops every other task in the backlog. Read the filesystem instead so `/takeoff` always shows **every** task regardless of prefix.

Enumerate every task file directly:

```bash
# All task markdown files across the backlog (tasks/, drafts/, nested epic dirs, etc.)
find backlog -type f -name '*.md' \
  -not -path 'backlog/archive/*' \
  -not -path 'backlog/completed/*' \
  -not -path 'backlog/docs/*' \
  -not -path 'backlog/decisions/*' \
  -not -path 'backlog/milestones/*' \
  -not -name 'README.md'
```

For each file, extract YAML frontmatter (the block between the first two `---` lines) and pull: `id`, `title`, `status`, `priority`, `assignee`, `dependencies`, `parent_task_id` (if present), and any `blocked_by` field. Treat missing `priority` as MEDIUM. Treat missing `status` as `To Do`.

If the MCP tool `mcp__backlog__task_list` is available **and** is known to honor every prefix (not the CLI behavior), it can supplement — but the filesystem scan is the source of truth.

Also pull sequence info to detect cross-task dependencies (this command does not filter by prefix):

```bash
backlog sequence list --plain
```

Parse both. You want: id, title, priority, status, assignee, dependencies, blocked-by, parent.

After parsing, group tasks by ID prefix (the segment before the first `-`, e.g. `ACANEBULA`, `NEBULA`, `CLEANUP`) so Step 4 can render related clusters together. **Every** prefix must be represented in the output — never narrow to a single prefix unless `--tag` or `--mine` was passed.

### Step 3: Classify each task

For every task returned, bucket it:

| Bucket | Criteria |
|---|---|
| **Actionable** | status = "To Do", no unresolved dependencies, not assigned to someone else |
| **Blocked** | status = "To Do" but has a `blocked_by` / unresolved dependency |
| **In Progress** | status = "In Progress" or "Doing" |
| **Done** | status = "Done" — skip, don't surface |

If a task has a dependency, note *which* task blocks it by ID.

### Step 4: Render the backlog as bulleted groups

Sort actionable tasks by priority (HIGH → MEDIUM → LOW), then by ID. By default show the top 5 in the top-priority group. If the user passed `--top N`, honor that. If they passed `--mine`, filter to their assignee.

**Format: flat bullet lists grouped by category, with an emoji header on each secondary group.** Tables wrap poorly in narrow terminals and compete with the numbered ID column. A plain `- ID — Title` bullet reads cleanly at any width and matches how users already scan backlogs.

Use `—` (em-dash) as the separator between ID and title. For tasks with dependencies, append `(blocked by X, Y)` or `(depends on X)` at the end of the line.

Group headers use these emojis:
- 🛫 Top-priority actionable group (the main recommendation pool)
- 🔵 Epic subtasks, or clusters of related follow-ups (e.g., `🔵 DOCREVIEW epic subtasks (children of DOCREVIEW-1)`)
- 🟢 In Progress
- ⚪ LOW / Housekeeping
- 🔴 Blocked / Dependent (only when all blockers are hard blockers)

Emit the output directly as markdown in your final response — no code fences, no custom rendering scripts.

**Shape:**

```markdown
## 🛫 Takeoff — <repo-name>

### Top Priority (actionable)

- AGENTSAPI-1 — Orchestrator admin-agents client + manifest types
- AGENTSAPI-2 — Admin Agents list page + row actions (depends on AGENTSAPI-1)
- NEBULA-42 — PR comment three-step enforcement hook
- NEBULA-53 — agent-browser coverage audit for Nebula admin agents UI

🔵 DOCREVIEW epic subtasks (children of DOCREVIEW-1)

- DOCREVIEW-1.1 — strip Recommendations / Answered-by / suggestion strip from final chain message
- DOCREVIEW-1.2 — generalize Review Document button via ChainConfig.producesReviewableDocument

🔵 PIPELINEPANE follow-ups (related to in-progress)

- PIPELINEPANE-3 — stop resetting orchestrationActive after animation completes (blocked by PIPELINEPANE-1, PIPELINEPANE-2)
- PIPELINEPANE-4 — regression sweep + solutions note (blocked by PIPELINEPANE-3)

🟢 In Progress

- AGENTSAPI-7 — something currently being worked on (@sam)

⚪ LOW / Housekeeping

- NEBULA-35 — Enforce backlog.md usage via MCP server and hooks
- NEBULA-39 — Document backlog-first workflow in AGENTS.md

### Recommendation
Start with **AGENTSAPI-1**. It's the next unblocked HIGH-priority item and clears the path for 2 more.
```

**Formatting rules for this step:**
- One task per line. Do not wrap a task across multiple lines — long titles stay on one line and let the terminal soft-wrap them.
- Group headers use H3 (`###`) for top-priority only; secondary groups use a bold-emoji line (e.g., `🔵 DOCREVIEW epic subtasks`), not an H-level heading. This matches the shape the user has already endorsed.
- Omit groups that have zero tasks. If In Progress is empty, write `🟢 In Progress — _None_` or just skip it.
- Never emit a table. No pipes, no box-drawing characters.

**Escape any pipes inside task titles** with `\|` as a defensive measure, even though bullets don't depend on pipe syntax. Do not truncate titles.

### Step 5: Offer a next step

End with a single question: "Want me to open the plan for `<ID>` and start `/ce:work` on it?" — do not auto-start. Takeoff is for orientation, not execution.

### Step 6: Final banner

After the recommendation question, emit a single final line:

```
✈️ TAKE OFF — NOW AT 30,000 FEET
```

This **must** be the last line of output — no content after it, no code fence, no trailing heading, no follow-up prose. It fires on every successful completion path, including the no-backlog fallback (Step 1) and the empty-task-list edge case (see Edge cases). Do not emit the banner if the routine aborts before producing a briefing.

## Formatting rules

- **Bullets, not tables.** Tables break at narrow widths and force column padding; bullets flow cleanly and match how the user already reads a backlog.
- **Group with emoji headers.** 🛫 top-priority, 🔵 related clusters / epics, 🟢 in progress, ⚪ housekeeping, 🔴 hard-blocked.
- **Never hide blockers.** Annotate dependencies inline with `(blocked by X)` or `(depends on X)` — don't silently drop the task.
- **Show dependency IDs explicitly.** `(blocked by AGENTSAPI-1)` is useful; `(has dependencies)` is not.
- **Be honest about emptiness.** If there are zero actionable tasks, say so and suggest creating one or picking up in-progress work.
- **Always end with the takeoff banner.** The final line of every successful invocation must be `✈️ TAKE OFF — NOW AT 30,000 FEET` — including fallback and empty-list paths.

## Why this shape

The user already has a "land the plane" routine that handles closing work. Takeoff's job is the mirror image: one concentrated briefing that answers *"what am I about to do and why"* in under 15 seconds of reading. A table + a one-line recommendation beats a paragraph because the user is scanning, not reading.

Dependencies matter more than priority on their own — a HIGH task that's blocked is worse than a MEDIUM task that's ready. That's why the table shows `Deps` and why the recommendation should factor in unblocking power, not just raw priority.

## Arguments

- `--top N` — show N tasks in the top-priority table (default 5)
- `--mine` — filter to tasks assigned to the current user (resolve via `git config user.email` or `git config user.name`)
- `--all` — also include LOW priority actionable tasks
- `--tag <tag>` — filter by a backlog tag/label

If no arguments are given, use defaults.

## Edge cases

- **No backlog directory** → fall back to `docs/plans/` summaries; tell the user.
- **CLI missing but MCP present** → use MCP only.
- **Everything is blocked** → surface the root blockers and ask whether to create a task to unblock them.
- **Task list is empty** → congratulate the user and suggest `/ce:ideate` or `/ce:plan` to generate new work.
- **Tasks without priority set** → treat as MEDIUM.
- **`backlog/config.yml` sets `task_prefix`** → the CLI's `backlog task list` will only return tasks matching that prefix. Do not use the CLI as the primary source — Step 2 already mandates a filesystem scan to ensure every task across every prefix is surfaced. If the briefing only shows tasks from one prefix, that's the bug — re-scan the filesystem.
