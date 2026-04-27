---
name: land
description: Close out a session by committing, pushing, and opening a PR — then handing off. Use when the user says "land", "/land", "land the plane", "land plane", "land it", "let's land", "land this", "bring it in", "wrap it up", "land the plan", "time to land", "ok land", "go ahead and land", or any variation that signals they want to finish, close out, ship, or wrap up the current session's work. Executes the full checklist without asking. Never merges the PR — landing ≠ merging.
---

# Land the Plane

The session completion counterpart to `/takeoff`. Where `/takeoff` *starts* work by surfacing the backlog, `/land` *finishes* work by running the full commit → push → PR → handoff checklist.

## Trigger

Any variation of: `land`, `/land`, `land the plane`, `land it`, `let's land`, `land this`, `bring it in`, `wrap it up`, `land the plan`, `land plane`, `time to land`, `ok land`, `go ahead and land`.

When the user's message contains "land" in the context of finishing/wrapping up work, invoke this protocol. When in doubt, invoke it.

## Core principle

"Landing" means **commit, push, and create a PR** — it does **not** mean merge. A PR is how humans review agent work; no PR means no review means no trust. **Never merge unless the user explicitly says "merge this PR".**

Work is **not complete until `git push` succeeds**. If push fails, resolve and retry until it works. Do not stop at "ready to push when you are" — you must push.

## Execution

Run the checklist **in order, completely, without asking**. Each step is non-negotiable.

### Step 1: File remaining work

Review what was worked on this session. Capture anything that's unfinished, deferred, or follow-up so it doesn't vanish when the session closes.

- If the repo has Backlog.md tooling (`backlog/` directory at repo root and the `backlog` CLI available), create tasks for unfinished/follow-up work:
  ```bash
  if command -v backlog >/dev/null 2>&1 && [ -d backlog ]; then
    backlog task create "<title>" --description "<context>"
  fi
  ```
- Otherwise, gather remaining work into a short handoff list and surface it at Step 9.

Skip silently if nothing remains.

### Step 2: Run quality gates (only if code changed this session)

Detect the stack from the repo root and run the matching build + test/lint commands. For ATV-starterkit specifically, this means Go at the root plus an optional Node subproject under `npm/`.

```bash
# Go (repo root) — run when go.mod is present AND Go files changed this session.
# Quality gates run BEFORE commit, so we cannot rely on HEAD~1; check working tree.
if [ -f go.mod ]; then
  if { git diff --name-only;
       git diff --name-only --cached;
       git ls-files --others --exclude-standard; } | grep -qE '\.go$|^go\.(mod|sum)$'; then
    go build ./... && go vet ./...
  fi
fi

# Node subproject — run when npm/ files have changed (staged, unstaged, or untracked).
if [ -f npm/package.json ]; then
  if { git diff --name-only;
       git diff --name-only --cached;
       git ls-files --others --exclude-standard; } | grep -q '^npm/'; then
    (cd npm && npm run build)
  fi
fi

# Generic fallbacks (portable to other repos)
# pnpm:    [ -f pnpm-workspace.yaml ] && pnpm build && pnpm lint
# npm:     [ -f package.json ] && npm run build && npm run lint
# Python:  [ -f pyproject.toml ] && pytest && ruff check .
# Rust:    [ -f Cargo.toml ] && cargo build && cargo test
```

If any gate fails (non-zero exit), **halt the routine, fix the failure, and re-run from Step 2 before proceeding to Step 4**. Do not append `|| true` to swallow failures — a broken build does not ship, and a swallowed failure would falsely emit the success banner at Step 10.

If no code changed (docs-only, config-only, planning-only sessions), skip quality gates and note that in the handoff.

### Step 3: Update task status (if the repo has task tracking)

- Mark completed tasks as `Done`.
- Update in-progress tasks with short implementation notes so the next session has context.
- If the repo uses `backlog_task_id` in plan frontmatter, ensure status reflects reality.

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
branch=$(git branch --show-current)
if [ -z "$branch" ]; then
  echo "ERROR: detached HEAD — refusing to push. Check out a branch first." >&2
  exit 1
fi
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

Check first whether an **open** PR already exists on this branch. `gh pr view` returns success for closed and merged PRs too, so inspect `state` — not just the exit code — otherwise a stale CLOSED PR will trick the skill into skipping creation and leave the user with nothing to ship:

```bash
if PR_VIEW_OUTPUT=$(gh pr view --json url,title,state 2>&1); then
  PR_VIEW_EXIT=0
  PR_VIEW_STATE=$(printf '%s' "$PR_VIEW_OUTPUT" | jq -r '.state // empty' 2>/dev/null)
else
  PR_VIEW_EXIT=$?
  PR_VIEW_STATE=""
fi
printf '%s\n__GH_PR_VIEW_EXIT__=%s\n__GH_PR_VIEW_STATE__=%s\n' \
  "$PR_VIEW_OUTPUT" "$PR_VIEW_EXIT" "$PR_VIEW_STATE"
```

Treat the branch as having no usable PR when **either** `PR_VIEW_EXIT != 0` **or** `PR_VIEW_STATE != "OPEN"`. In that case, create a fresh PR. If a CLOSED or MERGED PR is found, mention it in the new PR body so reviewers have the breadcrumb. **For PR body construction, follow the conventions in [`git-commit-push-pr`](../git-commit-push-pr/SKILL.md)** — value-first, intent-forward, scaled to the complexity of the change. Do not re-implement that logic here.

Resolve the default branch dynamically — it's not always `main`:

```bash
default_branch=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@')
if [ -z "$default_branch" ]; then
  default_branch=$(git rev-parse --verify origin/main >/dev/null 2>&1 && echo "main" || echo "master")
fi
git log "origin/$default_branch..HEAD" --oneline
git diff "origin/$default_branch...HEAD" --stat
```

Summarize the **full branch** diff (not just the latest commit) when authoring the body. Include a test plan checklist. Share the PR URL at handoff.

**Never merge the PR** unless the user explicitly says "merge this PR". Landing ≠ merging.

### Step 7: Clean up

```bash
git stash list                     # check for session-era stashes
# drop only stashes from this session; leave older ones alone
```

**If working inside a git worktree:** leave the worktree in place while the PR is open. Remove it manually with `git worktree remove <path>` only after the PR is merged or abandoned. Do not attempt to tear down worktrees from this skill.

### Step 8: Verify

Confirm a clean state:

```bash
git status                                          # working tree clean (or only untracked)
branch=$(git branch --show-current)
if [ -n "$branch" ] && git rev-parse --verify "origin/$branch" >/dev/null 2>&1; then
  git log "origin/$branch..HEAD"                    # must be empty — all pushed
else
  echo "WARNING: no upstream tracking ref for '$branch' — cannot verify pushed state"
fi
```

If either check fails, loop back and fix. Do not hand off a dirty or unpushed tree.

### Step 9: Hand off

Provide a concise summary for the next session:

- **Accomplished** — what shipped (with task IDs if applicable)
- **Next up** — what's queued (with task IDs if applicable)
- **Blockers / gotchas** — anything that tripped you up or is waiting on a decision
- **Branch** — current branch name
- **PR** — PR URL from Step 6

Keep it scannable. The next session (human or agent) should be able to take off from this handoff without re-reading the whole transcript.

### Step 10: Final banner

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

Some repos have local conventions layered on top of this protocol — read `.github/copilot-instructions.md` (and any `AGENTS.md` at the repo root, when present) for project-specific rules (e.g., branch protection, PR comment workflows, backlog linkage requirements). Project rules override these defaults where they conflict.
