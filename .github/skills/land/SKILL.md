---
name: land
description: Session completion protocol - the opposite of "takeoff". Use this skill whenever the user says "land", "/land", "land the plane", "land it", "let's land", "land this", "bring it in", "wrap it up", "land the plan", "land plane", "time to land", "ok land", "go ahead and land", or any variation that signals they want to finish, close out, ship, or wrap up the current session's work. Executes the full commit → push → PR → handoff checklist without asking. If the user's message contains "land" in the context of finishing work, invoke this skill — it is a keyword trigger, not an exact match.
argument-hint: "[optional: extra context, e.g. 'skip PR' or 'no handoff notes']"
---

# Land the Plane Protocol

The session completion counterpart to `/takeoff`. Where takeoff *starts* work by surfacing the backlog, `/land` *finishes* work by running the full commit → push → PR → handoff checklist.

## Trigger

Any variation of: `land`, `/land`, `land the plane`, `land it`, `let's land`, `land this`, `bring it in`, `wrap it up`, `land the plan`, `land plane`, `time to land`, `ok land`, `go ahead and land`.

**This is a keyword trigger, not an exact match.** If the user's message contains "land" in the context of finishing/wrapping up work, invoke this protocol. When in doubt, invoke it.

## Core principle

"Landing" means **commit, push, and create a PR** — it does **not** mean merge. A PR is how humans review agent work; no PR means no review means no trust. Never merge unless the user explicitly says "merge this PR".

Work is **not complete until `git push` succeeds**. If push fails, resolve and retry until it works. Do not stop at "ready to push when you are" — you must push.

## Execution

Run the checklist **in order, completely, without asking**. Each step is non-negotiable.

### Step 1: File remaining work

Review what was worked on this session. Capture anything that's unfinished, deferred, or follow-up so it doesn't vanish when the session closes.

- If the project has Backlog.md (a `backlog/` directory at repo root), create tasks for unfinished/follow-up work via the backlog CLI or `mcp__backlog__task_create`.
- Otherwise, gather remaining work into a short handoff list and surface it to the user at Step 10.

Skip silently if nothing remains.

### Step 2: Run quality gates (only if code changed this session)

Run the project's build and test commands. Detect the stack from the repo root rather than assuming:

```bash
# Node / Next.js (npm)
npm run build && npm run lint 2>&1 || true

# pnpm projects
pnpm build && pnpm lint 2>&1 || true

# Python
pytest && ruff check . 2>&1 || true

# Go
go build ./... && go vet ./... 2>&1 || true

# Rust
cargo build && cargo test 2>&1 || true
```

If build or tests fail, **fix them before proceeding**. Do not skip this step. A broken build does not ship.

If no code changed (docs-only, config-only, planning-only sessions), skip quality gates and note that in the handoff.

### Step 3: Update task status (if project has task tracking)

Mirror Step 1's conditional + literal-command pattern so a fresh agent — one cold-loading the skill on a session where the user said "land it" — can mechanically follow the CLI shape rather than guessing it from prose:

```bash
if command -v backlog >/dev/null 2>&1 && [ -d backlog ]; then
  # Mark completed tasks as Done
  backlog task edit TASK-XYZ --status Done

  # Update in-progress tasks with implementation notes for the next session
  backlog task edit TASK-XYZ \
    --status "In Progress" \
    --notes "<one-line implementation context, commit refs, what's blocking>"
fi
```

Substitute the actual task IDs from this session. Skip silently when the `backlog` CLI is absent — the conditional guard covers projects without backlog tooling.

If the project uses `backlog_task_id` in plan frontmatter, ensure status reflects reality.

### Step 4: Commit all changes

Stage specific files — **never** `git add -A` or `git add .`. That risks pulling in `.env`, credentials, or large binaries.

```bash
git status                         # see what's outstanding
git add <specific-files>           # stage deliberately
git commit -m "<type>: <summary>"  # conventional commits (feat, fix, refactor, docs, test, chore, perf, ci)
```

If there are no changes, skip. Do not create empty commits.

### Step 5: PUSH TO REMOTE (MANDATORY)

Work is not complete until this step succeeds.

```bash
# confirm branch is not main/master — project hooks may block push to main anyway
branch=$(git branch --show-current)
# only rebase if this branch already tracks a remote — new branches have no upstream yet
if git rev-parse --verify "origin/$branch" >/dev/null 2>&1; then
  git pull --rebase origin "$branch"
fi
git push -u origin "$branch"
git status                         # must show "up to date with origin"
```

If push fails (conflicts, hook rejection, branch protection), **resolve and retry** until it works. Do not hand off with unpushed commits.

### Step 6: Create or update the PR

A PR is the review artifact. Agent work without a PR has no trust surface.

```bash
gh pr view                         # check if PR already exists for this branch
# if not:
gh pr create --fill --web=false    # or craft a proper title + body
```

When creating the PR body, summarize the **full branch** diff (not just the latest commit). Resolve the default branch dynamically — it's not always `main`:

```bash
default_branch=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@')
# fallback if remote HEAD isn't set
if [ -z "$default_branch" ]; then
  default_branch=$(git rev-parse --verify origin/main >/dev/null 2>&1 && echo "main" || echo "master")
fi
git log "origin/$default_branch..HEAD" --oneline
git diff "origin/$default_branch...HEAD" --stat
```

Include a test plan checklist in the body. Share the PR URL at handoff.

**Never merge the PR** unless the user explicitly says "merge this PR". Landing ≠ merging.

### Step 7: Clean up

```bash
git stash list                     # check for session-era stashes
# drop only stashes from this session; leave older ones alone
```

If working in a worktree:
- PR open/pending review: `ExitWorktree(action: "keep")`
- Work merged or abandoned: `ExitWorktree(action: "remove")`

**Sweep stale ralph-loop state files.** If a previous session ran `ralph-loop` and the state file was orphaned (session crash, cwd drift into a worktree, completion promise emitted but not as the trailing tokens of the assistant message), the plugin's stop-hook will replay the loop's prompt verbatim in the next session as "Stop hook feedback." Run unconditionally — the cost is microseconds and detecting "did this session use ralph-loop?" is unreliable:

```bash
DELETED_RL=$(find . -name ralph-loop.local.md -path '*/.claude/*' -print -delete 2>/dev/null)
if [ -n "$DELETED_RL" ]; then
  echo "ralph-loop state files removed (surface in Step 10 handoff under blockers/gotchas):"
  echo "$DELETED_RL"
fi
```

The path glob `*/.claude/*` ensures only state files inside `.claude/` directories are deleted — a legitimately named `ralph-loop.local.md` document elsewhere is left alone.

### Step 8: Verify

Confirm a clean state:

```bash
git status                                # working tree clean (or only untracked)
git log "origin/$(git branch --show-current)..HEAD"  # must be empty — all pushed
```

If either check fails, loop back and fix. Do not hand off a dirty or unpushed tree.

### Step 9: Capture session state

Persist a written record of the session so the next agent (or human) can resume with full context, not just the verbal handoff in Step 10.

Invoke the `remember:remember` skill — it writes to the project's `.remember/` buffer (`now.md`, `today-*.md`, `recent.md`) and is picked up automatically by the `SessionStart:clear` hook on the next session.

```
Skill: remember:remember
```

Run unconditionally — even on docs-only or planning-only sessions. The buffer is cheap and the recovery value is high. If the `remember` skill is unavailable in this environment, fall back to writing a brief `.remember/now.md` (or `~/.claude/session-data/<date>.md` outside a repo) by hand with: branch, PR URL, accomplished, next up, blockers.

### Step 10: Hand off

Provide a concise summary for the next session:

- **Accomplished** — what shipped (with task IDs if applicable)
- **Next up** — what's queued (with task IDs if applicable)
- **Blockers / gotchas** — anything that tripped you up or is waiting on a decision
- **Branch** — current branch name
- **PR** — PR URL from Step 6

Keep it scannable. The next session (human or agent) should be able to take off from this handoff without re-reading the whole transcript.

### Step 11: Final banner

After the handoff summary, emit a single final line:

```
🛬 PLANE LANDED — NICE WORK
```

This **must** be the last line of output — no content after it, no code fence, no trailing heading. The banner is a **completion** marker, not a "we tried" marker: emit it only when the routine completes successfully (including the clean-tree / nothing-to-commit path where Step 5 is skipped because there's nothing to push). If `git push` never succeeds, a quality gate fails and halts the routine, or the PR step errors out and cannot be resolved, do **not** emit the banner.

## Critical rules

These are non-negotiable when `/land` is invoked:

- **NEVER stop before pushing.** Unpushed work is stranded work.
- **NEVER say "ready to push when you are."** You push. That is the job.
- **NEVER skip quality gates.** Broken code does not ship.
- **NEVER merge the PR** unless the user explicitly says "merge this PR".
- **NEVER use `git add -A` or `git add .`.** Stage specific files.
- **NEVER bypass hooks** (`--no-verify`, `--no-gpg-sign`) unless the user explicitly asks. If a hook fails, investigate and fix the root cause.
- If push fails, **resolve and retry** until it succeeds.
- **Always end successful landing output with the `🛬 PLANE LANDED — NICE WORK` banner line.** Do not emit the banner on failure paths (push failed, quality gate halted the routine, PR step errored out with no resolution).

## Project-specific considerations

Some repos have local conventions layered on top of this protocol — read `CLAUDE.md` and `AGENTS.md` at the repo root for project-specific rules (e.g., branch protection, PR comment workflows, backlog linkage requirements). Project rules override these defaults where they conflict.
