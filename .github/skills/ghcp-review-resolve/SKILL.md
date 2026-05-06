---
name: ghcp-review-resolve
description: Produce two independent agent reviews on the current PR (an existing Copilot review when present, otherwise two distinct subagent reviewers), adjudicate findings with a third independent subagent, post inline PR comments for verified issues, run a tight inline fix-and-reply-and-resolve loop, then deliver a verdict — APPROVE if clean, APPROVE-WITH-CHANGES if findings were verified (and fixed or noted), or CLOSE if the PR is fundamentally unsound. The skill is read-only with respect to GitHub reviewer assignment — it never adds reviewers. Use whenever the user invokes /ghcp-review-resolve, asks to "run copilot review and resolve", asks to "review and fix my PR with copilot", asks for a "dual review and fix pass", or wants automated review triage, remediation, and verdict on a pull request they just opened. Submits APPROVE or CLOSE; never submits REQUEST_CHANGES and never merges.
---

# ghcp-review-resolve

Orchestrates a dual-review-and-fix pipeline on an open PR. The workflow:

0. **Preflight** — detect PR, fetch size + merge state + head SHA, passively detect any existing Copilot review, check for prior resolved reviews. Emit a preflight table. If the PR has merge conflicts with base, run the **conflict-resolution subroutine** (study recent merged PRs in the repo for resolution conventions, then rebase and resolve) before continuing — merge conflicts are no longer a blocker. The only remaining blocker is "nothing useful to do" (e.g., prior Copilot review fully resolved at HEAD).
1. **Collect two reviews** synchronously:
   - If the PR already has a fresh Copilot review, read it (Review A) and spawn one subagent reviewer (Review B).
   - Otherwise, spawn two distinct subagent reviewers in parallel (Review A + Review B).
2. Independently adjudicate findings via a third subagent that inspects the actual code.
3. Post inline PR comments only for verified bugs/fixes.
4. Run a tight inline fix loop per comment: edit → test → commit → push → reply on the thread → resolve the thread.
5. **Render a verdict and act on it** — `APPROVE` (clean), `APPROVE-WITH-CHANGES` (findings verified, fixed-or-noted), or `CLOSE` (PR is fundamentally unsound). Submit an APPROVE review or close the PR accordingly.
6. Summarize — never submit REQUEST_CHANGES, never merge, never assign reviewers.

## Why this exists

Automated reviewers produce lots of findings. Some are real bugs. Some are stylistic noise. Some contradict each other. Blindly "fix everything the bots said" is how you ship regressions or waste a day on non-issues.

This skill's job is to be the adult in the room: produce two independent reviews, have a fresh adjudicator verify each finding against the actual code, and only act on what's real. Overlapping findings are high-confidence. Unique findings are kept only when high-severity and verifiable.

It also knows when **not** to run. A PR whose prior Copilot review is already fully resolved at the current HEAD shouldn't trigger another round of bot noise — the skill reports that state and gets out of the way. Merge conflicts, by contrast, are not a reason to stop: the skill resolves them as a preflight subroutine before reviewing.

**The skill never assigns reviewers.** Asking GitHub to add `@copilot` (or anyone else) as a reviewer is silently dropped on most repos and 422s on others, and chasing those failure modes was a recurring source of wasted runs. Instead the skill reads any Copilot review that already exists on the PR and pairs it with a subagent reviewer, or spawns two subagents itself when no Copilot review is present.

## Step 0 — Preflight

The preflight is the first and only place allowed to abort the run. If it passes, every later step trusts its flags. If a blocker is reported, no review is produced, no comment is posted, no fix is attempted.

### 0a. Basic environment

```bash
gh auth status
```

If this fails, stop with a clear error. Do not proceed.

### 0b. Detect the PR

Auto-detect the PR for the current branch:

```bash
PR_NUMBER=$(gh pr view --json number -q .number 2>/dev/null)
```

If the user passed an argument, prefer that. If `PR_NUMBER` is still empty, ask the user and stop.

### 0c. Fetch PR metadata

```bash
gh pr view "$PR_NUMBER" --json \
  headRefOid,changedFiles,additions,deletions,mergeStateStatus,mergeable,baseRefName \
  > /tmp/ghcp-pr-meta.json
```

Extract into local variables:

- `PR_HEAD_SHA` — head SHA (later mutations re-check this to detect mid-run pushes)
- `CHANGED_FILES` — file count
- `LINES_CHANGED` = additions + deletions
- `MERGE_STATE_STATUS` — `CLEAN`, `DIRTY`, `BLOCKED`, `BEHIND`, `UNKNOWN`, etc.
- `BASE_REF` — base branch name

### 0d. Classify PR size

Size thresholds (named so they're easy to tune later):

- `SIZE_THRESHOLD_FILES = 20`
- `SIZE_THRESHOLD_LINES = 2000`

```
if CHANGED_FILES <= SIZE_THRESHOLD_FILES and LINES_CHANGED <= SIZE_THRESHOLD_LINES:
    SIZE_CLASS = "small"
else:
    SIZE_CLASS = "large"
```

`small` → adjudicator and reviewer subagents use the full-diff path (`gh pr diff`).
`large` → they use the per-file paginated path (`gh api .../pulls/{n}/files --paginate`). This avoids `gh pr diff`'s 20k-line API cap.

### 0e. Check merge state

If `MERGE_STATE_STATUS == "UNKNOWN"`, GitHub hasn't finished computing mergeability (common right after a push). Wait up to 30 seconds, re-fetching every 10s, then proceed with whatever state is reported.

If the final `MERGE_STATE_STATUS == "DIRTY"` (has conflicts with base), **do NOT stop by default** — run the conflict-resolution subroutine in §0e.1 to bring the PR up to date, then re-fetch metadata and continue preflight. Merge conflicts are work the skill is expected to do, not a reason to bail out.

If the resolution subroutine fails after a reasonable effort (see exit criteria below), only then emit the preflight table with a blocker note and stop.

**Opt-out:** if the user passed `--no-auto-resolve` (or the skill config sets `auto_resolve_conflicts: false`), skip the subroutine entirely on `DIRTY` and behave the way the skill did before this capability landed: emit the preflight table, log "auto-resolve disabled by flag/config", and stop with the recommended-action block:

```
Blocker: PR has merge conflicts with base (mergeStateStatus=DIRTY); auto-resolve disabled.

Recommended next action — resolve conflicts before re-running this skill:

  Option A (manual):
    git fetch origin && git rebase origin/<base>
    # resolve conflicts, then: git push --force-with-lease

  Option B (delegated):
    Skill(skill="compound-engineering:ce-work", args="resolve the merge conflicts on PR #<N>")
```

`--no-auto-resolve` is orthogonal to `--force` (the existing escape hatch for `PRIOR_RESOLVED=true`); both can be passed independently and either may be active without affecting the other.

#### 0e.1 Conflict-resolution subroutine

The goal: produce a clean rebase of the PR branch onto `origin/<BASE_REF>` and force-push it, so the rest of the skill can review against an up-to-date head. The subroutine is **convention-aware** — it studies how this repo has handled similar conflicts before, and follows that style rather than guessing.

1. **Checkout the PR branch locally** (if not already):
   ```bash
   gh pr checkout "$PR_NUMBER"
   git fetch origin "$BASE_REF"
   ```
   If checkout fails because the branch is on a fork the runner can't push to, abort the subroutine — emit a blocker explaining the fork ownership issue. Don't try to resolve conflicts you can't push back.

2. **Study repo conflict-resolution conventions.** Before attempting the rebase, gather context so the resolutions follow repo norms instead of arbitrary picks. Look at the most recent merged PRs that touched the same files as this PR:
   ```bash
   # Files this PR changes
   gh pr view "$PR_NUMBER" --json files -q '.files[].path' > /tmp/ghcp-pr-files.txt

   # Recently merged PRs (last 30) and their changed files
   gh pr list --state merged --limit 30 --json number,title,mergedAt,files \
     > /tmp/ghcp-recent-merged.json

   # Recently closed-without-merge PRs (last 10) — often informative about
   # rejected resolution approaches
   gh pr list --state closed --limit 10 --json number,title,closedAt,body \
     > /tmp/ghcp-recent-closed.json
   ```
   Cross-reference: which of the last 30 merged PRs touched files that overlap with this PR's files? Read the merge commits and any "fixup conflict" / "rebase onto main" commits in those PRs to see how the maintainer resolved similar collisions. Note any pattern (e.g., "always keep the PR side for `AGENTS.md` table rows", "always take main for generated lockfiles").

   Also check the PR body and review comments for any explicit instruction the author or reviewer left about the conflict (e.g., "this needs to merge after #251 because both touch X — keep my version"). Surface those in your subroutine plan.

3. **Attempt the rebase**:
   ```bash
   git rebase "origin/$BASE_REF"
   ```
   If it succeeds with no conflicts, jump to step 6.

4. **Resolve conflicts file by file.** For each conflicted file:
   - Read the conflict markers and both sides.
   - Apply repo conventions from step 2.
   - Default heuristics when no convention is obvious:
     - **Lockfiles** (`pnpm-lock.yaml`, `package-lock.json`, `yarn.lock`, `Cargo.lock`, `Gemfile.lock`, `go.sum`, `poetry.lock`): take main's version, then re-run the package manager's install/lock command in the working tree to integrate the PR's dependency changes; commit the regenerated lockfile.
     - **Auto-generated files** (`schema.rb`, `next-env.d.ts`, anything marked "DO NOT EDIT"): take main's version and let regeneration sort it out.
     - **Documentation tables / lists** (e.g., `AGENTS.md` gate tables, README service lists): keep both sides' additions where possible — append the PR's new rows after main's rows.
     - **Code with overlapping additions**: prefer a manual merge that preserves both intents; never silently drop functionality from either side.
   - After editing, `git add <file>` it.
   - If a conflict is genuinely ambiguous and no repo convention applies, write a short note describing both options and **stop the subroutine** — surface the ambiguity to the user as a blocker rather than guessing.

5. **Continue the rebase** until clean:
   ```bash
   git rebase --continue
   ```
   Repeat step 4 for each commit that conflicts. If the same file conflicts on multiple commits, suspect that the resolution at the earlier commit was wrong — reconsider before continuing.

6. **Verify the rebase locally** before pushing:
   - Run a fast sanity check that the working tree at least parses/builds. Use whatever the repo's cheapest verification command is (e.g., `pnpm install --frozen-lockfile`, `cargo check`, `go build ./...`, `python -m py_compile $(git diff --name-only origin/$BASE_REF -- '*.py')`).
   - If verification fails, attempt one repair pass. If that also fails, abort the subroutine — do not push a broken rebase.

7. **Force-push with lease**:
   ```bash
   git push --force-with-lease
   ```
   `--force-with-lease` (not `--force`) — refuses if someone else pushed in the meantime, preventing clobbered work.

8. **Re-fetch PR metadata** so later preflight steps see the new head:
   ```bash
   gh pr view "$PR_NUMBER" --json \
     headRefOid,changedFiles,additions,deletions,mergeStateStatus,mergeable,baseRefName \
     > /tmp/ghcp-pr-meta.json
   ```
   Update `PR_HEAD_SHA`, `CHANGED_FILES`, `LINES_CHANGED`, `MERGE_STATE_STATUS`. Recompute `SIZE_CLASS` if it changed. The new `MERGE_STATE_STATUS` should now be `CLEAN`, `BLOCKED`, or `BEHIND` — if it's somehow still `DIRTY`, the subroutine failed; treat as a blocker.

**Exit criteria — when to give up and escalate to a blocker:**
- A conflict exists in a file the runner can't reasonably resolve without product-level decisions (e.g., two divergent feature implementations of the same function with no clear convention).
- Verification (step 6) fails twice in a row.
- The PR branch is on a fork the runner can't push to.
- The user's repo has a hook or branch protection that rejects the force-push.

When any exit criterion fires, abort the rebase (`git rebase --abort`), emit the preflight table with the conflict-resolution subroutine's findings, and stop with a clear note:

```
Blocker: conflict-resolution subroutine could not complete.

Reason: <one-line summary>
Files attempted: <list>
Conventions studied: <one-line summary of step 2 findings>
Recommended next action:
  - Review the PR's conflicting files manually, OR
  - Rerun: Skill(skill="compound-engineering:ce-work", args="resolve the merge conflicts on PR #<N> using the repo conventions documented in <link>")
```

No inline comments are posted, no review is submitted, no fixes are attempted on the still-conflicted PR. The skill exits cleanly.

### 0f. Detect existing Copilot review (read-only)

The skill **never** assigns reviewers. It only reads what's already on the PR. Use GraphQL to fetch reviews and review threads in one query:

```bash
gh api graphql -f query='
  query($owner:String!, $repo:String!, $number:Int!) {
    repository(owner:$owner, name:$repo) {
      pullRequest(number:$number) {
        reviews(first: 50) {
          nodes {
            author { login }
            state
            commit { oid }
            submittedAt
            body
          }
        }
        reviewThreads(first: 100) {
          nodes {
            id
            isResolved
            comments(first: 20) {
              nodes {
                id
                author { login }
                path
                line
                body
                commit { oid }
              }
            }
          }
        }
      }
    }
  }' -F owner=<owner> -F repo=<repo> -F number=$PR_NUMBER
```

A review is "Copilot-authored" if its author login matches `github-copilot[bot]` or `copilot-pull-request-reviewer[bot]` (case-insensitive). Compute `EXISTING_COPILOT_REVIEW`:

- **`fresh`** — Copilot has at least one review whose `commit.oid` equals or is an ancestor of `PR_HEAD_SHA`, AND at least one Copilot-authored thread is unresolved (so there is something to feed the adjudicator). Use this review as Review A in Step 1.
- **`stale`** — Copilot reviewed at some point, but `PR_HEAD_SHA` has moved past every Copilot review's commit. Logged in preflight; treated as `none` for sourcing.
- **`none`** — no Copilot review on this PR at all.

A fully-resolved fresh Copilot review (every Copilot thread `isResolved=true` at current head) does NOT count as `fresh` for sourcing — the work is already done. It feeds the `PRIOR_RESOLVED` check below instead.

**REST fallback:** if the GraphQL query fails (auth scope, schema drift, older `gh`), fall back to `/pulls/{n}/reviews` and `/pulls/{n}/comments`, filter by author login the same way, and approximate the same classification. When in doubt, set `EXISTING_COPILOT_REVIEW=none` and proceed — the two-subagent path always works.

### 0g. Check for prior fully-resolved Copilot review (idempotency)

Using the threads fetched in 0f, classify each Copilot-authored thread:

- **resolved-and-fresh** — `isResolved=true` AND the thread's most recent comment SHA is an ancestor of (or equals) `PR_HEAD_SHA`.
- **resolved-but-stale** — `isResolved=true` BUT `PR_HEAD_SHA` is newer than the thread's resolution SHA.
- **open** — `isResolved=false`.

Derive:

```
PRIOR_RESOLVED = (there is at least one Copilot thread) AND
                 (every Copilot thread is resolved-and-fresh)
```

When `PRIOR_RESOLVED=true`, the skill has nothing useful to do on the Copilot side — and re-running pr review is low-value too. Emit the preflight table with a clear note ("Prior Copilot review found with all threads resolved at current HEAD; skipping run") and stop, unless the user passed `--force`.

**Escape hatch:** `--force` (or explicit user instruction to re-run anyway) bypasses this short-circuit and proceeds with the normal two-source flow in Step 1, ignoring the existing Copilot review for sourcing purposes.

### 0g.1 Build the regression watch list

Resolved Copilot findings are signal, not noise. They are issues Copilot once flagged that an author claimed to fix — and over many commits a "fix" can be reverted, refactored away, or recreated as the same bug in a new file. Carry every resolved Copilot finding forward into a structured **regression watch list** that the adjudicator (Step 3) checks against the current diff.

From the Copilot threads classified in §0g, build the watch list from threads that are **resolved-and-fresh** OR **resolved-but-stale** (both buckets represent "Copilot caught it, author claimed to fix it"). Skip open threads (those feed Review C in §1b) and dismissed threads (author-rejected; carrying them forward re-litigates that judgment).

For each qualifying thread, capture:

```
{
  "watch_id": "<thread id>",
  "file": "<path>",
  "line": <int|null>,
  "body": "<verbatim original Copilot comment>",
  "resolution_sha": "<sha at which the thread was resolved>",
  "comment_ids": [<original GitHub comment IDs>],
  "comment_url": "<URL of the original Copilot comment for human reference>"
}
```

Record `REGRESSION_WATCH_COUNT = len(watch list)` and surface it in the preflight table so the user sees how much carry-forward signal will be checked. The watch list is bounded (one entry per resolved Copilot thread on this PR), the lookups are cheap, and the cost scales with the actual review history of the PR.

**Cross-PR carry-forward is out of scope.** This watch list is per-PR only; it does not consult resolved findings from other PRs.

**Rebase implication:** if the conflict-resolution subroutine in §0e.1 just rewrote the head SHA, original `comment_ids` and `resolution_sha` values still resolve correctly via the GitHub API (they're immutable), but file/line coordinates may now refer to lines that no longer exist. The adjudicator's regression-check pass (§3) is responsible for diffing against current HEAD; stale coordinates are filtered there, not here.

### 0j. Preflight CI / GitHub Actions status check

Read the current GitHub Actions status for `PR_HEAD_SHA` and emit a `CI_STATUS` flag that downstream verdict logic can gate on. This step happens **after** the regression watch list (§0g.1) is built and **before** the preflight table (§0h) is emitted.

The actual classifier is a pure function over `gh run list` JSON; reference implementation ships alongside this skill at `lib/ci-classifier.js` with unit tests at `tests/ci-classifier.test.js`.

```bash
# Fetch all workflow runs for this branch, filter to PR_HEAD_SHA in classifier.
gh run list \
  --branch "$BRANCH" \
  --json status,conclusion,name,databaseId,headSha,event \
  --limit 50 \
  > /tmp/ghcp-pr-runs.json
```

Pass the JSON plus `PR_HEAD_SHA` to the classifier. The classifier filters stale runs (any `headSha != PR_HEAD_SHA` is dropped — they belong to older pushes) and produces:

`CI_STATUS ∈ {green, red-code, red-flake-suspected, pending, none, unknown}`

Classification rules:

| Bucket | Condition |
|---|---|
| `green` | All considered runs concluded `success`, `skipped`, or `neutral` |
| `red-code` | At least one `failure` / `timed_out` / `cancelled` AND failing-job logs do NOT all match the infra-flake allowlist |
| `red-flake-suspected` | Failing runs ALL match infra-flake patterns (`ECONNRESET`, `TLS handshake timeout`, `pull access denied`, `runner has received a shutdown signal`, `The operation was canceled.`, etc.) |
| `pending` | At least one run is `in_progress`, `queued`, `waiting`, or `action_required` AND no failures yet |
| `none` | No workflow runs exist at the current `PR_HEAD_SHA` (repo has no Actions, or runs haven't been triggered) |
| `unknown` | `gh run list` failed (rate limit, auth scope, API error). Non-blocking — surfaced in summary |

Capture failing-job metadata (databaseId, name) for the resolution subroutine in §0j.2:

```
CI_FAILING_JOBS = [{name, databaseId, run_url}, ...]   # populated when CI_STATUS in {red-code, red-flake-suspected}
```

If the gh call itself fails, set `CI_STATUS=unknown` and continue — `unknown` does not block preflight or the verdict (the user has no signal in either direction; surface the failure in §7).

### 0j.1. Pending-CI policy at preflight

When `CI_STATUS=pending` at preflight, the skill **does not stall**. It proceeds to Step 1 with a note in the preflight table: `CI status: pending — proceeding without preflight CI gate`. The authoritative gate is §5.5 (post-fix-loop), which runs against a SHA the skill itself produced.

Courtesy wait (optional): if every workflow is in `queued` state and the youngest run is < 30s old, wait up to 60s for at least one run to enter `in_progress` so §5.5 has a baseline to compare against. Do not extend this wait — the post-fix-loop check is the real gate, and stalling preflight on slow runners costs more than it saves.

Mixed `success` + `in_progress` is `pending` (any incomplete demotes from green). Same proceed-with-note behavior.

### 0j.2. CI failure resolution subroutine

When `CI_STATUS=red-code` or `red-flake-suspected` at preflight, run this subroutine to bring CI green before reviews are spawned. Mirrors §0e.1 in shape: study conventions → attempt fix → verify → push → re-check → bounded retries → blocker on exhaustion.

**Opt-out:** if the user passed `--no-auto-fix-ci` (or the skill config sets `auto_fix_ci: false`), skip the subroutine entirely on red CI: emit the preflight table, log the opt-out, and stop with a blocker (red CI + APPROVE would contradict the verdict — see §6 Guardrails).

1. **Checkout the PR branch locally** (reuse `gh pr checkout` if §0e.1 didn't already do it).

2. **Identify failing jobs.** Use `CI_FAILING_JOBS` from §0j. For each failing run, fetch only the failing-step logs:
   ```bash
   gh run view <run-id> --log-failed > /tmp/ghcp-ci-fail-<run-id>.log
   ```

3. **Flake triage.** If `CI_STATUS=red-flake-suspected` (every failing job's log matched the infra allowlist), run a single rerun and re-poll:
   ```bash
   for run_id in $CI_FAILING_JOBS; do
     gh run rerun "$run_id" --failed
   done
   # Bounded poll: 30s interval, 5min ceiling. Re-run §0j classifier on completion.
   ```
   If the rerun lands `green`, exit subroutine successfully. If still red, treat as code-caused and continue. **Single rerun only** — flake-then-fail-again is treated as a real failure, not perpetual flake.

4. **Study repo conventions.** Before guessing at a fix, gather context:
   - Recent merged PRs (last 30) whose changed files overlap with the failing workflow's surface area (e.g., the workflow file itself, or files the failing step exercises).
   - Commits with messages matching `fix:.*ci|fix:.*lint|fix:.*test|fix:.*build` and their diffs.
   - Any `.github/workflows/` README or comments documenting how this repo handles common failures.

   Note any pattern (e.g., "this repo always pins node version when actions/setup-node fails", "lockfile mismatches → `pnpm install --no-frozen-lockfile && pnpm install --frozen-lockfile`").

5. **Spawn a fix-attempt subagent** (`correctness-reviewer` or `general-purpose`) with the failing logs, the repo conventions, and the diff. Prompt:
   > Identify the root cause of the failing job. Propose the minimal fix as a code edit. Do NOT refactor unrelated code, do NOT rename, do NOT reformat. If the failure is not bounded and obvious, return `{"unresolvable": "<one-line reason>"}` instead.

   Apply the proposed edit. Skip any subagent reply that includes a wide refactor.

6. **Verify locally.** Run the closest local equivalent of the failing job (e.g., `pnpm lint`, `pytest <path>`, `cargo check`, `go test ./<package>/...`). If verification fails, attempt one repair pass; if still failing, abort the subroutine.

7. **Commit and push** with `--force-with-lease`:
   ```bash
   git add <specific-files>
   git commit -m "fix(ci): <one-line> (ghcp-review-resolve PR#<N>)"
   git push --force-with-lease
   ```

8. **Wait for re-run and re-classify.** Bounded poll (30s interval, 5min ceiling). Re-run §0j classifier. If now `green`, exit subroutine successfully. If still red, recurse once more with a fresh subagent. **Max 2 total fix attempts at preflight.**

**Exit criteria → blocker:**
- 2 fix attempts exhausted at preflight.
- Local verification fails twice in a row.
- Subagent returns `unresolvable` (root cause unclear, needs product decision).
- Failing job is in a workflow controlled by branch protection / required checks the runner can't modify.
- `--no-auto-fix-ci` flag is set.

When any exit criterion fires, abort, emit the preflight table with subroutine findings, and stop:

```
Blocker: CI gate could not complete.

Reason: <one-line>
Failing jobs: <list with run URLs>
Conventions studied: <one-line summary>
Recommended next action:
  - Review the failing jobs manually, OR
  - Rerun: Skill(skill="compound-engineering:ce-work", args="fix the failing CI jobs on PR #<N> using the failure logs at <gh run view URL>")
```

After the subroutine completes successfully (CI green), refetch PR metadata so later steps see the new head:

```bash
gh pr view "$PR_NUMBER" --json headRefOid,... > /tmp/ghcp-pr-meta.json
# update PR_HEAD_SHA, recompute SIZE_CLASS if needed
```

Then continue preflight with the updated `CI_STATUS=green`.

### 0h. Emit the preflight table

Before doing anything with side effects, print a compact preflight report:

```
Preflight — PR #<N>

  Check                       Status
  ─────────────────────────── ──────────────────────────────────────────────────
  gh auth                     ok
  PR detected                 #<N> (<branch>, head <short-sha>)
  Size                        <files> files / +<add> −<del>  [<size_class>]
  Merge state                 <CLEAN | DIRTY | BLOCKED | UNKNOWN>
  Existing Copilot review     <fresh | stale | none>
  Prior Copilot resolved      <yes — skipping run | no | n/a>
  Regression watch list       <N entries | none>
  CI status                   <green | red-code | red-flake-suspected | pending | none | unknown>
  CI auto-fix                 <attempted: ok | attempted: failed | skipped (flag) | not needed | flake rerun: ok | n/a>
  Review sourcing             <2 subagents + Copilot context (Review C) | 2 subagents>
  Auto-resolve conflicts      <attempted: ok | attempted: failed | skipped (flag) | not needed | n/a>

Decision: <proceed to Step 1 | run conflict-resolution subroutine then proceed | run CI resolution subroutine then proceed | STOP — blocker: ...>
```

The "Review sourcing" row is determined by `EXISTING_COPILOT_REVIEW`:

- `fresh` → `2 subagents + Copilot context (Review C)`
- `stale` or `none` → `2 subagents`

The "Regression watch list" row reports `REGRESSION_WATCH_COUNT` from §0g.1 — how many resolved Copilot findings will be re-checked by the adjudicator against the current diff.

### 0i. Record preflight flags

Later steps read these flags; they are the contract between preflight and the rest of the skill:

- `PR_NUMBER`, `PR_HEAD_SHA`, `BASE_REF`
- `SIZE_CLASS` ∈ {small, large}
- `EXISTING_COPILOT_REVIEW` ∈ {fresh, stale, none}
- `PRIOR_RESOLVED` ∈ {true, false}
- `REGRESSION_WATCH_COUNT` ∈ ℕ (number of resolved Copilot threads carried into the adjudicator's regression-check pass)
- `REVIEW_SOURCING` ∈ {`two-subagents`, `two-subagents+copilot-context`}
- `CI_STATUS` ∈ {green, red-code, red-flake-suspected, pending, none, unknown}
- `CI_FAILING_JOBS` — array of `{name, databaseId, run_url}`, populated when `CI_STATUS` is red-*
- `CI_AUTO_FIX_OPT_OUT` ∈ {true, false} — set when `--no-auto-fix-ci` is passed (orthogonal to `--no-auto-resolve` and `--force`)
- `CI_POST_STATUS` ∈ {green, red-code, red-flake-suspected, pending, none, unknown} — set in §5.5 after the fix loop completes

If `PRIOR_RESOLVED=true` and no `--force`, the skill stops here.

**Any later step that references preflight flags but finds them unset must refuse to run** (defense against the preflight ever being accidentally skipped).

## Step 1 — Collect two reviews

This step **always** produces two fresh subagent review streams, "Review A" and "Review B," that flow into the adjudicator. There is no polling and no reviewer assignment. Both streams come from sources the skill controls. Any existing Copilot review on the PR is treated as **supplementary context** — it becomes a third input to the adjudicator (Review C) when fresh, but it never replaces one of the two subagent reviews.

Rationale: relying on Copilot's pre-existing review as one of the two streams produces correlated coverage when Copilot is fresh and silently degrades to single-reviewer mode when it isn't. Always running two fresh subagents keeps independence high and makes the skill behave the same way regardless of whether the org has Copilot enabled.

### Diff-fetch shared context

Both subagent reviewers (and later the adjudicator) need the diff in `SIZE_CLASS`-aware form:

- `small` → fetch full PR diff once and pass it to each subagent:
  ```bash
  gh pr diff "$PR_NUMBER" > /tmp/ghcp-pr-diff.patch
  ```
- `large` → fetch the per-file index once; each subagent reads only the files referenced by its own analysis:
  ```bash
  gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/files" --paginate \
    > /tmp/ghcp-pr-files.json
  ```

Cap subagent per-file reads at ~30 files per run; if the PR is bigger than that, subagents prioritize files with the largest churn.

### 1a. Spawn two subagent reviewers (always)

Spawn two distinct subagent reviewers via two `Task` calls in the same message so they run concurrently. They produce Review A and Review B regardless of `EXISTING_COPILOT_REVIEW`. Each prompt:

> You are reviewing PR #`<N>` at SHA `<sha>` in `<repo>`. Diff-fetch mode: `<small | large>`. Diff is at `/tmp/ghcp-pr-diff.patch` (small) or per-file via `/tmp/ghcp-pr-files.json` (large).
>
> Identify real bugs, logic errors, security issues, and concrete correctness problems. Skip style preferences, speculative refactors, and cosmetic suggestions.
>
> Return JSON only: `[{file, line, severity, body}, ...]`. Do NOT post comments to the PR. Do NOT edit any files.

Personas come from the pair below.

### 1b. Read existing Copilot review as supplementary input (when fresh)

When `EXISTING_COPILOT_REVIEW == "fresh"`, also read the Copilot review's line comments and top-level body from the GraphQL result captured in Step 0f. Normalize each Copilot comment into the finding shape used downstream and label it `Review C (copilot-existing)`:

```
{
  "review": "C",
  "source": "copilot-existing",
  "file": "<path>",
  "line": <int|null>,
  "severity": null,           // Copilot rarely emits a structured severity
  "body": "<verbatim>",
  "comment_id": "<github comment id, for thread reply later>"
}
```

Filter Review C to **unresolved** Copilot threads only — resolved threads at fresh head feed the regression watch list (§0g.1) and the `PRIOR_RESOLVED` short-circuit instead, not the live review stream. Stale Copilot reviews are not used as Review C; they fall through to "no Review C" semantics.

Review C is supplementary context for the adjudicator, not a substitute for either subagent. The adjudicator evaluates A, B, and (when present) C using the same overlap heuristic — a finding flagged by ≥2 of the available streams is high-confidence.

### Persona pair

The two subagent reviewers must have meaningfully different focus areas — correlated reviewers defeat the point of independent dual review. Default pair (chosen from the global agent registry; final pick verified at edit time):

- **Reviewer A:** `correctness-reviewer` (logic errors, edge cases, state bugs, intent-vs-implementation mismatches)
- **Reviewer B:** `adversarial-reviewer` (actively constructs failure scenarios; high-risk diffs)

For Branch 1a (existing Copilot + 1 subagent), use only Reviewer B by default — Copilot already covers the correctness lens reasonably well, so adversarial coverage is the highest-leverage complement. The user can override the pair via skill args if exposed.

The persona pair is documented here so the user can see what's running. Subagent reviewers must NOT touch the PR — only the fix loop in Step 4 writes to the PR.

### 1c. Failure handling

- If one subagent (Review A or B) errors out: log the failure, fall back to single-subagent mode for this run, note in the final summary that only one fresh review was collected. Adjudicator still runs.
- If both subagents error out: stop. No fix loop, no verdict, no state change on the PR.
- If reading the existing Copilot review for Review C fails (rate limit, transient API): retry once; on second failure, treat Review C as absent and continue with A + B only.

## Step 2 — Collect and normalize findings

Build a single list of findings from Review A and Review B. For each:

```
{
  "review": "A" | "B",
  "source": "copilot-existing" | "subagent:<persona>",
  "file": "path/to/file.go",
  "line": 42,              // nullable — some findings are file-level
  "severity": "...",       // subagents provide this; copilot usually doesn't
  "body": "...",           // the raw review text
  "comment_id": 12345      // GitHub review comment ID for reply/resolve (only set for copilot-existing)
}
```

Deduplicate near-identical findings (same file + overlapping line range + substantively similar body) and mark them `overlap: true`. Overlap is the strongest positive signal.

When Branch 1c forced a single-subagent fallback, there is no overlap to detect. That's fine — the adjudicator (Step 3) still operates on unique findings, it just loses the "both reviewers flagged it" signal and must rely entirely on severity + verifiability.

## Step 3 — Adjudicate findings with an independent subagent

Spawn a fresh subagent (general-purpose or `code-reviewer`) to independently verify each finding against the actual code. **This subagent did NOT write the code and did NOT produce either review** — that independence is the whole point.

Adjudicator prompt:

> You are adjudicating a set of automated PR review findings on PR #`<N>` at SHA `<sha>` in `<repo>`. Diff-fetch mode: `<small: full diff at /tmp/ghcp-pr-diff.patch | large: per-file patches at /tmp/ghcp-pr-files.json>`. The findings come from two independent reviewers labeled "Review A" and "Review B" (and optionally "Review C" — supplementary context from an existing Copilot review on the PR).
>
> For each finding, read the referenced file/line, decide whether the finding is a real bug, logic error, security issue, or concrete correctness problem that warrants a code change. Reject style preferences, speculative refactors, "consider adding a comment" suggestions, and anything not grounded in code you can actually see.
>
> Keep a finding if:
> - Two or more of the available streams flagged it (overlap), AND it is a real issue you can verify in the code, OR
> - Only one stream flagged it, AND it is high-severity (bug, security, data loss, incorrect logic, broken contract) AND verifiable.
>
> **Drop** any finding that references a file or line not present in the PR diff — such findings are not grounded in the changes.
>
> **Regression-check pass.** You will also receive a `REGRESSION_WATCH_LIST` of resolved Copilot findings from earlier on this PR (per §0g.1). For each watch-list entry, read the original `body` and check whether the same problem class reappears anywhere in the **current** diff (the file/line in the entry may be stale post-rebase — search by symptom, not coordinates). If you can verify a regression in the current code, emit it as an additional finding flagged `regression: true` with `prior_comment_id` and `prior_comment_url` populated from the watch-list entry. Skip watch-list entries whose subject matter is no longer present in the diff.
>
> Return JSON: `[{file, line, severity, rationale, proposed_fix, source_review_labels: ["A"|"B"|"C"], source_comment_ids: [...], regression: <bool>, prior_comment_id?: <id>, prior_comment_url?: <url>}, ...]`.

Run tests/build before adjudication if cheap — a failing test is ground-truth evidence. The adjudicator is allowed (and encouraged) to actually run the test suite if it helps verify a finding.

## Step 4 — Post inline PR comments for accepted findings only

For each accepted finding, post an inline review comment on the PR at the exact file/line. Use `gh api` with a review that has `event: COMMENT` (not APPROVE, not REQUEST_CHANGES):

```bash
gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/reviews" \
  -X POST \
  -f commit_id="$PR_HEAD_SHA" \
  -f event="COMMENT" \
  -f body="ghcp-review-resolve: verified findings to address" \
  -F "comments[][path]=path/to/file.go" \
  -F "comments[][line]=42" \
  -F "comments[][body]=**Verified finding** (from: Review A [<source>], Review B [<source>])\n\n<rationale>\n\n**Proposed fix:** <proposed_fix>"
```

Name the actual sources in the body (e.g., `existing Copilot review`, `correctness-reviewer subagent`, `adversarial-reviewer subagent`) so a human PR author can trace each finding. Prefix the comment body with `ghcp-review-resolve:` so later steps can identify comments this skill created.

**Regression findings.** When a finding has `regression: true`, the inline comment body must additionally include a "**Regression of:**" line linking to `prior_comment_url` from the watch-list entry. Example: `**Regression of:** [original Copilot finding](https://github.com/.../pull/N#discussion_rXXX) — that thread was resolved at commit `<resolution_sha>` but the same issue is present in the current diff.` This makes the carry-forward chain auditable.

Before posting, re-check `PR_HEAD_SHA` against current HEAD. If it changed (someone else pushed), stop — see Guardrails.

Do NOT (in this step):
- Submit `event: APPROVE` here — APPROVE is reserved for Step 6 after the fix loop and verdict
- Submit `event: REQUEST_CHANGES` ever — see Guardrails
- Close, merge, or mark-ready any PR here — closing is reserved for Step 6

## Step 5 — Inline fix loop, sequentially per comment

Sequential per comment (not parallel) — multiple findings can touch the same file, and serial edits avoid merge conflicts and let each fix be verified independently before moving on.

`/lfg` is intentionally **not** used here. `/lfg` is the full autonomous pipeline (plan → work → review → todo-resolve → test → video) and that's overkill for a single reviewer comment. Instead, run this tight inline loop in the current session:

For each accepted finding, in order by file then line:

1. **Read** the referenced file and surrounding context (±30 lines) so you understand what the fix needs to preserve.

2. **Edit** minimally. The goal is the smallest change that addresses the finding. Don't refactor adjacent code, don't rename things, don't reformat. If the finding needs a larger change than a localized edit can deliver, skip it and note "needs larger change — left for human" in the final summary rather than snowballing scope.

3. **Verify** the fix:
   - Run the narrowest relevant tests (prefer targeted test file/package over the full suite — faster feedback).
   - If the project's build is cheap (< 30s), run it.
   - If no tests exist for the area, at minimum run the linter/type-checker on the touched files.
   - If verification fails: attempt one repair. If that also fails, revert the edit for this finding, record it as skipped, and move on. Do not spend more than one retry per finding — fix fatigue is real and one stubborn item shouldn't block the others.

4. **Commit** with a focused message referencing the finding:
   ```bash
   git add <specific-files>
   git commit -m "fix: <one-line summary> (ghcp-review-resolve PR#<N> comment <comment_id>)"
   ```
   One commit per finding. Small commits are easier to revert if the reviewer disagrees with the fix.

5. **Push** after each commit (so the reply comment can point at a real pushed SHA):
   ```bash
   git push
   ```

6. **Reply** on the specific review comment thread with what changed and why:
   ```bash
   gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/comments" \
     -X POST \
     -f body="Fixed in <commit_sha>: <one-line description>. Verification: <tests run / build status>." \
     -F in_reply_to=<comment_id>
   ```
   The `in_reply_to` field is the GitHub-supported way to thread under an existing review comment. If that call fails (some older API versions), fall back to posting a new top-level PR comment that references the original comment URL.

7. **Resolve the thread.** After the fix is pushed and the reply posted, mark the review thread resolved so the PR's "unresolved conversations" count drops to zero. Use the GraphQL `resolveReviewThread` mutation:
   ```bash
   gh api graphql -f query='mutation($threadId:ID!) { resolveReviewThread(input:{threadId:$threadId}) { thread { isResolved } } }' -F threadId=<thread_id>
   ```
   The `<thread_id>` is the `id` from the `reviewThreads` GraphQL query (you may need to refetch since new threads were created in Step 4). If a fix was skipped (reverted after retry, or "needs larger change"), do NOT resolve — leave the thread open so a human notices it.

8. **Move on** to the next finding. Do not pause for user input between findings — the whole point is one-shot resolution. If something truly blocks progress (repo credentials, missing dependency), stop the whole pipeline and report.

### Batching by file (optional optimization)

If multiple accepted findings touch the same file and are close in line numbers, it's fine to address them in a single edit pass and a single commit — just make the reply post on each original comment thread. This keeps git history clean without losing reviewer-facing traceability. Don't batch across files.

### What "verification" means per language (quick guide)

The skill doesn't need to be language-aware, but orient the verification step to the repo you're in:

- **Go**: `go test ./<package>/...` for the touched package, then `go vet ./...` and `go build ./...` if fast.
- **TypeScript/JavaScript**: the project's test command on the affected file pattern, then `tsc --noEmit` on the touched file.
- **Python**: `pytest <test_file>` or `pytest <dir>`, then the project's configured linter (`ruff`, `flake8`, etc.) on the touched file.
- **Rust**: `cargo test <test_name>` or the relevant module, then `cargo check`.
- **Other**: look for scripts in `package.json` / `Makefile` / `justfile` that name-match "test" or "check".

If you can't identify a verification command in ~30 seconds of looking, commit the change and note "not independently verified — review reply documents intent" in the reply. Honesty about uncertainty is better than silent hand-waving.

## Step 5.5 — Post-fix-loop CI re-check and regression handling

After the §5 fix loop pushes its last commit (and before §6 renders a verdict), re-check CI on the new HEAD. The skill's own commits may have introduced regressions on workflows that were green at preflight — the §1–§5 path runs targeted local tests per finding, but CI runs the full matrix, and parity gaps surface here.

**Skip §5.5 entirely** when the fix loop made zero commits (no findings accepted, or all were skipped without code changes). `CI_POST_STATUS` defaults to `CI_STATUS` in that case — there is no new SHA to re-check against.

1. **Re-fetch `PR_HEAD_SHA`.** It moved during the fix loop. Use the latest `gh pr view --json headRefOid` value for all subsequent steps.

2. **Wait for CI runs on the new SHA.** Bounded poll: 30s interval, 5min ceiling. Workflows triggered by the fix-loop pushes are the runs whose outcome the skill actually owns, so a short wait here is justified (unlike preflight, where the skill didn't trigger CI).

   If runs are still pending after the ceiling, surface in summary and do **not** block — proceed to §6 with `CI_POST_STATUS=pending` and a body note. The user can re-run or wait.

3. **Classify CI on the new SHA** using the same §0j logic, producing `CI_POST_STATUS ∈ {green, red-code, red-flake-suspected, pending, none, unknown}`.

4. **Diff against `CI_STATUS` from preflight to detect regressions.** Match by **workflow name**, not run-id (run-ids change on every push). For each workflow that was `success` at preflight and is now `failure`, build a regression record:

   ```
   {
     workflow_name: "lint",
     prior_status: "success",
     current_status: "failure",
     run_url: "https://github.com/.../actions/runs/<id>",
     commits_since_preflight: ["<sha1>", "<sha2>", ...]
   }
   ```

5. **If regressions exist:**
   - Treat each regression as a high-severity adjudicator finding (analogous to §0g.1 watch-list regressions, but for CI status — not Copilot threads).
   - Run §0j.2 resolution subroutine **once** on the regression(s). **Max 1 round** of post-fix-loop fixing — this prevents infinite loops where one fix breaks a different job, then a new fix breaks a third, etc.
   - If the subroutine succeeds and CI is green: proceed to §6 with `CI_POST_STATUS=green` and a note in the §7 summary about the regression and its fix.
   - If the subroutine fails: emit a blocker, do **not** submit APPROVE in §6. The verdict is `WITHHELD` — the fix loop introduced a regression we couldn't resolve; left for human. Inline review threads remain in whatever state they're in; the PR is not approved or closed.

6. **If pre-existing failures persist** (red at preflight, still red): preflight should have already either (a) resolved them via §0j.2, or (b) blocked the run. We should not reach §5.5 with unaddressed pre-existing red CI. If somehow we do, treat as a blocker and emit `CI_POST_STATUS=red-code` with a note that preflight resolution did not stick.

7. **Edge case: cascading rewrites.** If §0e.1 conflict-resolution AND the fix loop both rewrote head, CI status from preflight may compare against an obsolete SHA. The workflow-name-keyed diff (step 4) handles this: workflows that didn't exist at preflight are not regressions; workflows that existed but at a different SHA are still comparable by name.

8. **Edge case: regression-fix triggered new CI.** When the regression subroutine succeeds, its own fix is a new commit that itself triggers CI. Treat the latest CI as authoritative — re-run step 2's bounded poll once more to wait for that new CI. The skill caps at 1 regression-fix round, so this can fire at most once.

After §5.5 completes, set `CI_POST_STATUS` and proceed to §6.

## Step 6 — Render the verdict and act on it

After the fix loop, the skill owes the user (and the PR author) a decision. There are exactly three verdicts. **REQUEST_CHANGES is intentionally not one of them** — by this point either you fixed the issue and approved, or you concluded the PR shouldn't merge and closed it. Asking a human to "go fix this" defeats the purpose of an automated review-and-resolve.

### Verdict rules

Compute the verdict from adjudicator output and fix-loop results:

| State after fix loop | Verdict |
|---|---|
| 0 verified findings (nothing the adjudicator accepted) AND `CI_POST_STATUS in {green, none}` | **APPROVE** (clean) |
| ≥1 verified findings, all fixed-or-noted, PR is salvageable, AND `CI_POST_STATUS in {green, none, pending}` | **APPROVE-WITH-CHANGES** |
| `CI_POST_STATUS=red-code` AND §5.5 regression subroutine could not resolve | **WITHHELD** (no APPROVE, no CLOSE) |
| PR is fundamentally unsound (see CLOSE criteria below) | **CLOSE** |

**CI gate on verdict:**

- `CI_POST_STATUS=green` or `none` → no gate; verdict follows finding-state.
- `CI_POST_STATUS=pending` → APPROVE-WITH-CHANGES is allowed only with an explicit body note linking to in-flight runs; APPROVE (clean, zero findings) is also allowed with the same note. The user can re-run if they want a definitive gate.
- `CI_POST_STATUS=red-flake-suspected` → treated as `pending` for verdict purposes (the rerun in §0j.2 either flipped it to green or escalated to red-code; if it's still flake-suspected here, the rerun budget was exhausted and the human should look).
- `CI_POST_STATUS=red-code` → **WITHHELD**. The skill exits cleanly with a blocker summary; no APPROVE submitted, no CLOSE submitted.
- `CI_POST_STATUS=unknown` → proceed with finding-state verdict but include a body note that CI status could not be determined (gh API failure, etc.).

**WITHHELD criteria** — all of:
- `CI_POST_STATUS=red-code`
- §5.5 regression-fix subroutine ran and failed (or §5.5 detected a regression that §0j.2 declined as out-of-scope)
- The CLOSE criteria below do **not** apply (the PR is salvageable in principle, but the skill's own commits broke something it can't fix)

When WITHHELD, the inline review threads remain in whatever state they're in from §5 — fixed threads stay resolved, skipped threads stay open. The PR is **not** approved, **not** closed, **not** merged. The §7 summary names the failing job and the regression chain so a human can pick up.

**CLOSE criteria** — only when at least one is true:
- Adjudicator concluded the PR's premise is wrong (wrong approach, duplicates existing functionality, contradicts an architectural decision, addresses a non-problem).
- The PR introduces a critical security or data-loss risk that cannot be patched without a rewrite.
- The PR is abandoned/superseded (e.g., references a closed issue, supersedes a merged PR, branch is months stale and rebasing is non-trivial).
- The PR is empty, broken, or spam after Dependabot/automation rebase failures with no recoverable changes.

If none of those apply, the verdict is APPROVE or APPROVE-WITH-CHANGES, even if there are open low-severity suggestions. CLOSE is reserved for "this PR should not exist in its current form" — not "I'd merge it but it could be nicer."

### How to act on each verdict

**APPROVE** — submit a formal approval review:
```bash
gh pr review "$PR_NUMBER" --approve --body "ghcp-review-resolve consensus: **APPROVE**.

<one-paragraph rationale: which review sources ran (Review A: ..., Review B: ...), what they found, why no findings survived adjudication>

<optional: low-priority suggestions that did not block approval>"
```

**APPROVE-WITH-CHANGES** — same `--approve` flag, but the body explicitly names what was fixed and what (if anything) is deferred:
```bash
gh pr review "$PR_NUMBER" --approve --body "ghcp-review-resolve consensus: **APPROVE-WITH-CHANGES**.

Review A: <source>. Review B: <source>.

Verified findings: <N>. Fixed in this run: <M> (commit SHAs: ...). Deferred: <K> (with one-line reason each).

<one-paragraph rationale referencing the inline review threads, all of which should now be resolved>"
```

The "with-changes" label is a signal to the human reviewer: the code that was reviewed is not exactly the code originally pushed — the skill made commits in this branch. They should see those commits in the diff before merging.

**WITHHELD** — emit a blocker summary and stop. Do **not** call `gh pr review` and do **not** call `gh pr close`:
```
Verdict: WITHHELD — CI red after fix loop, regression resolution failed.

Preflight CI: <CI_STATUS>. Post-fix CI: <CI_POST_STATUS>.
Failing jobs: <list with run URLs>
Regressions detected: <workflow names that flipped success → failure>
Skill commits since preflight: <list of SHAs>

Recommended next action — review the failing jobs manually:
  gh run view <run-id> --log-failed
  Or rerun: Skill(skill="compound-engineering:ce-work", args="fix the failing CI jobs on PR #<N> introduced by ghcp-review-resolve commits <sha-range>")
```

**CLOSE** — close the PR with a rationale:
```bash
gh pr close "$PR_NUMBER" --comment "ghcp-review-resolve consensus: **CLOSE**.

<one-paragraph rationale citing the specific CLOSE criterion that applies>

<what the PR author could do next, if anything — e.g., 'open a fresh PR off current main', 'see issue #X for the alternate approach', or 'this work is superseded by #Y'>"
```

Only close when the criteria above are clearly met. Closing somebody else's work is high-blast-radius — when in doubt, prefer APPROVE-WITH-CHANGES with detailed inline notes and let the human decide.

### Re-check head SHA before each state change

Right before submitting the approval or the close, refetch `headRefOid`. If it differs from `PR_HEAD_SHA`, someone pushed during your run — stop, do not approve or close, and report. The user can re-run the skill against the new head if they want.

### Don't double-approve

If a `gh pr view --json reviews` lookup shows you (the running user) already submitted an APPROVE on this exact `headRefOid`, don't submit a second one. Just log "already approved at this SHA" and move to Step 7.

## Step 7 — Final summary to the user

Produce a single summary message covering:

- PR number and URL
- Preflight outcome (size class, merge state, existing-Copilot-review status, prior-resolved state, review sourcing)
- **Preflight CI status** (`green` / `red-code` / `red-flake-suspected` / `pending` / `none` / `unknown`)
- **Preflight CI auto-fix outcome** (when applicable: resolved / flaked-and-passed / blocker / skipped via `--no-auto-fix-ci`)
- Review A source and Review B source (e.g., "existing Copilot review" + "correctness-reviewer subagent", or "correctness-reviewer subagent" + "adversarial-reviewer subagent")
- Findings: total raised, total accepted, total rejected (with top reasons), any dropped as "not grounded in diff"
- Fixes: what was changed, commit SHAs, any fixes that failed or were skipped
- **Post-fix-loop CI status** (`green` / `red-code` / `pending` / etc.) and any regressions detected
- **Post-fix-loop CI auto-fix outcome** (when §5.5 regression subroutine fired: resolved / blocker)
- **Verdict**: `APPROVE` / `APPROVE-WITH-CHANGES` / `CLOSE` / `WITHHELD`, with the rationale
- **State change**: which API call was made (`gh pr review --approve`, `gh pr close`, or none for `WITHHELD`) and whether it succeeded
- Resolved threads vs. left-open threads (left-open = a fix was skipped, so a human needs to look)
- Explicit confirmation: **PR was not merged. REQUEST_CHANGES was not submitted. No reviewers were assigned.**

Format as Markdown. Keep it scannable.

## Guardrails — do not cross these

- **If preflight reports a blocker, stop cleanly.** Do not proceed into Steps 1–7. No reviews produced, no comments, no commits, no verdict.
- **Never** call `gh pr edit --add-reviewer @copilot` or any other reviewer-assignment API. The skill is read-only with respect to GitHub reviewer assignment. If a future contributor is tempted to "just try requesting Copilot," the answer is no — that path produces silent drops on most repos and is the reason this skill was rewritten.
- **Never** run `gh pr merge` — merging is the maintainer's call, full stop.
- **Never** submit a `REQUEST_CHANGES` review — by Step 6 you've either fixed the issue (APPROVE / APPROVE-WITH-CHANGES), concluded the PR shouldn't merge (CLOSE), or withheld the verdict because CI broke (WITHHELD). REQUEST_CHANGES asks a human to do work the skill exists to do.
- **APPROVE and CLOSE are allowed in Step 6 only**, and only after the verdict logic in Step 6 is followed. Don't approve in Step 4 mid-fix-loop. Don't close on a hunch — the CLOSE criteria are narrow.
- **Never** APPROVE on `CI_POST_STATUS=red-code`. The verdict is `WITHHELD` — emit a blocker summary and stop. APPROVE on `CI_POST_STATUS=pending` is allowed only with an explicit body note linking to the in-flight runs.
- **Never** silently retry CI fixes more than the bounded limits — max 2 preflight rounds in §0j.2, max 1 post-fix-loop round in §5.5. Bounds are non-negotiable; exhaustion emits a blocker.
- **Never** treat real failures as flake without log evidence. The flake heuristic is an allowlist of infrastructure log signatures (network, runner, image-pull). Anything else is code-caused, even if `conclusion=cancelled`.
- **Never** act on a finding the adjudicator subagent rejected, even if both reviews flagged it — the adjudicator is the tiebreaker.
- **Never** fabricate finding text. If a review's text is ambiguous, include the verbatim original in the inline comment so a human can sanity-check.
- **Never** silently drop all findings from one review because of a parsing issue. If you can't parse, stop and report.
- **Subagent reviewers must NOT touch the PR.** Only the fix loop in Step 5 writes to the PR. If a reviewer subagent posts comments or edits files, that's a bug — stop and report.
- If the PR head SHA changes mid-run (someone else pushed), stop fixing, do not approve or close, report state, and let the user decide whether to restart.
- Treat missing preflight flags in any later step as a bug — refuse to run rather than assume defaults. This applies to `CI_STATUS`, `CI_POST_STATUS`, and `CI_AUTO_FIX_OPT_OUT` the same way it applies to the original preflight flags. §6 entered with `CI_POST_STATUS` unset must abort with an internal-error message.

## Flags reference (escape hatches)

The skill accepts three orthogonal opt-out flags. Any combination is valid; each is independently toggleable.

| Flag | Purpose | Default behavior without flag |
|---|---|---|
| `--force` | Bypass `PRIOR_RESOLVED=true` short-circuit and re-run the dual-review flow | Skip when prior Copilot review fully resolved at HEAD |
| `--no-auto-resolve` | Skip the §0e.1 conflict-resolution subroutine on `DIRTY` | Auto-rebase and resolve conflicts before reviewing |
| `--no-auto-fix-ci` | Skip the §0j.2 CI fix subroutine and the §5.5 regression-fix subroutine; emit a blocker on red CI instead | Auto-fix red CI before reviewing; auto-fix regressions after fix loop |

## Example runs

### Example 1 — happy path (small clean PR, no existing Copilot review)

```
User: /ghcp-review-resolve
→ Preflight:
    PR #42 (feat/add-payments, head abc123)
    Size: 7 files / +212 −41  [small]
    Merge state: CLEAN
    Existing Copilot review: none
    Prior Copilot resolved: n/a
    CI status: green
    Review sourcing: 2 subagents
    Decision: proceed
→ Spawning two subagent reviewers in parallel:
    Review A: correctness-reviewer
    Review B: adversarial-reviewer
→ Both returned: 9 findings total (3 overlap)
→ Adjudicator subagent verifying against src/ at abc123 [diff mode: full]...
→ 4 accepted (3 overlap + 1 unique high-severity), 5 rejected
→ Posting 4 inline comments on PR #42
→ Fix 1/4: null check in PaymentProcessor.go:88 → edit, go test ./payment/ ok → commit def456 → push → reply → thread resolved
→ Fix 2/4: off-by-one in pagination → edit, go test ./api/ ok → commit ghi789 → push → reply → thread resolved
→ Fix 3/4: missing error wrap → edit, go vet ok → commit jkl012 → push → reply → thread resolved
→ Fix 4/4: race in cache update → edit, go test ./cache/ -race FAIL on retry → reverted, skipped with note, thread left open
→ Verdict: APPROVE-WITH-CHANGES (4 verified, 3 fixed, 1 left for human)
→ §5.5 post-fix CI re-check: green (no regressions detected on the 3 push commits)
→ Submitting `gh pr review --approve` with rationale referencing fixes...
→ Summary: PR #42 — verdict APPROVE-WITH-CHANGES. 3 threads resolved, 1 left open. No reviewers assigned. Not merged.
```

### Example 2 — auto-resolve path (merge conflict, resolvable)

```
User: /ghcp-review-resolve
→ Preflight:
    PR #9 (fix/pr7-lint-fixes, head 025bb1f)
    Size: 143 files / +15044 −3262  [large]
    Merge state: DIRTY
    CI status: pending (proceeding without preflight CI gate; post-fix-loop check is authoritative)
    Decision: run conflict-resolution subroutine
→ gh pr checkout 9 → ok
→ Studying repo conventions:
    Last 30 merged PRs scanned. 4 touched overlapping files (#251, #244, #239, #228).
    Pattern: AGENTS.md gate-table conflicts → preserve both rows; lockfile conflicts → take main, regenerate.
→ git rebase origin/main → 6 conflicts:
    pnpm-lock.yaml          → take main + pnpm install --lockfile-only → staged
    AGENTS.md               → preserve both gate-table rows → staged
    .github/copilot-instructions.md → preserve both sections → staged
    README.md               → preserve both lists → staged
    package.json            → manual merge (both deps preserved) → staged
    apps/foo/index.ts       → manual merge → staged
→ git rebase --continue × 2 (one more commit conflicted on AGENTS.md, same resolution)
→ Verification: pnpm install --frozen-lockfile → ok
→ git push --force-with-lease → ok (new head 8a1f2d3)
→ Re-fetched PR metadata: mergeStateStatus=CLEAN, head 8a1f2d3
→ Continuing preflight...
    Existing Copilot review: stale (prior reviews predate the rebase)
    Prior Copilot resolved: no
    Review sourcing: 2 subagents
→ Spawning correctness-reviewer + adversarial-reviewer in parallel...
→ [normal review/fix flow continues]
```

### Example 2b — escalation path (conflict-resolution subroutine gives up)

```
User: /ghcp-review-resolve
→ Preflight:
    PR #88 (feat/payment-rewrite, head 7c11d4e)
    Merge state: DIRTY
    Decision: run conflict-resolution subroutine
→ gh pr checkout 88 → ok
→ Studying repo conventions: no recent PRs touched overlapping files; no convention available
→ git rebase origin/main → conflict in src/payment/processor.ts
    Both sides reimplement the same function with different signatures.
    No repo convention applies. Resolution requires product-level decision.
→ git rebase --abort → ok
→ Emitting blocker:

Blocker: conflict-resolution subroutine could not complete.

Reason: src/payment/processor.ts has divergent feature implementations on both sides; no repo convention applies.
Files attempted: src/payment/processor.ts
Conventions studied: no overlap with last 30 merged PRs.
Recommended next action:
  - Review the conflicting file manually with the PR author, OR
  - Rerun: Skill(skill="compound-engineering:ce-work", args="resolve the merge conflicts on PR #88, deciding between the two processor.ts implementations")

No reviews produced. No comments posted. PR not approved/closed/merged.
```

### Example 3 — Copilot already reviewed (Branch 1a)

```
User: /ghcp-review-resolve
→ Preflight:
    PR #17 (feat/big-refactor, head 9ab12c3)
    Size: 42 files / +3100 −900  [large]
    Merge state: CLEAN
    Existing Copilot review: fresh (5 open threads at HEAD)
    Prior Copilot resolved: no
    CI status: green
    Review sourcing: Copilot + 1 subagent
    Decision: proceed
→ Reading existing Copilot review (5 inline findings + body) → Review A
→ Spawning adversarial-reviewer subagent → Review B (returned 4 findings)
→ Adjudicator subagent verifying [diff mode: per-file]... read 8/42 files
→ 3 accepted (2 overlap with Copilot + 1 unique high-severity from Review B), 6 rejected
→ Posting 3 inline comments, running fix loop...
→ Fixes 1-3 applied and verified, threads resolved (including the 2 original Copilot threads).
→ Verdict: APPROVE-WITH-CHANGES (3 verified, 3 fixed).
→ Submitting `gh pr review --approve` ... ok.
→ Summary: PR #17 — verdict APPROVE-WITH-CHANGES. Used existing Copilot review + adversarial-reviewer subagent. No reviewers assigned. Not merged.
```

### Example 4 — CLOSE verdict (PR is fundamentally unsound)

```
User: /ghcp-review-resolve
→ Preflight:
    PR #88 (feat/dedupe-helper, head 7c11d4e)
    Size: 6 files / +180 −12  [small]
    Merge state: CLEAN
    Existing Copilot review: none
    Prior Copilot resolved: n/a
    CI status: green
    Review sourcing: 2 subagents
    Decision: proceed
→ Spawning correctness-reviewer + adversarial-reviewer in parallel...
→ 7 raw findings (overlap on 2). Adjudicator subagent reads code at 7c11d4e [diff mode: full]...
→ Adjudicator concludes: PR reimplements a helper that already exists in `internal/dedup/` (last touched 3 weeks ago) and contradicts the architectural decision in docs/adr/0014-dedup-strategy.md to centralize this logic. PR's helper is not a drop-in replacement — call sites would diverge.
→ Verdict: CLOSE (criterion: "duplicates existing functionality / contradicts an architectural decision")
→ Refetched headRefOid → unchanged at 7c11d4e.
→ Submitting `gh pr close --comment "..."` with rationale + pointer to ADR-0014 and `internal/dedup/` ... ok.
→ Summary: PR #88 — verdict CLOSE. No fixes attempted (closing supersedes the fix loop). No reviewers assigned. Not merged. REQUEST_CHANGES not submitted.
```

### Example 5 — CI gate fires and resolves (preflight red → fix → post-fix green → APPROVE-WITH-CHANGES)

```
User: /ghcp-review-resolve
→ Preflight:
    PR #103 (feat/api-cache-layer, head 4d7a91c)
    Size: 12 files / +480 −58  [small]
    Merge state: CLEAN
    Existing Copilot review: none
    Prior Copilot resolved: n/a
    CI status: red-code (1 failing job: lint)
    Decision: run CI resolution subroutine then proceed
→ §0j.2 fetching failing logs:
    gh run view 12345 --log-failed → 'src/api/cache.ts:42 — Missing semicolon'
→ Studying repo conventions: 8 of last 30 merged PRs touched lint config; pattern: minimal targeted fix, no auto-format.
→ Spawning fix-attempt subagent (correctness-reviewer)...
→ Subagent: "Root cause is a missing semicolon at line 42. Minimal fix: add `;`."
→ Applied edit, verified locally: pnpm lint → ok.
→ git commit "fix(ci): add missing semicolon (ghcp-review-resolve PR#103)" → push --force-with-lease → ok.
→ Bounded poll for re-run: 90s elapsed, lint job → success.
→ Re-classify: CI status: green. Re-fetched PR_HEAD_SHA → 5e8b0a2.
→ Continuing preflight... CI status: green
→ Spawning correctness-reviewer + adversarial-reviewer in parallel.
→ 6 findings, 2 overlap. Adjudicator: 2 accepted, 4 rejected.
→ Posting 2 inline comments. Fix loop: 2/2 fixed and verified. Threads resolved.
→ §5.5 post-fix CI re-check: green (no regressions on the 2 fix-loop commits).
→ Verdict: APPROVE-WITH-CHANGES (2 verified, 2 fixed; preflight CI was red but resolved before review).
→ Submitting `gh pr review --approve` ... ok.
→ Summary: PR #103 — verdict APPROVE-WITH-CHANGES. Preflight CI auto-fix: resolved (1 commit). Post-fix CI: green. 2 threads resolved. No reviewers assigned. Not merged.
```

### Example 6 — CI gate withholds verdict (post-fix-loop regression unresolvable)

```
User: /ghcp-review-resolve
→ Preflight:
    PR #117 (feat/payment-refactor, head a1b2c3d)
    Size: 18 files / +890 −340  [small]
    Merge state: CLEAN
    Existing Copilot review: none
    Prior Copilot resolved: n/a
    CI status: green
    Review sourcing: 2 subagents
    Decision: proceed
→ Spawning correctness-reviewer + adversarial-reviewer.
→ 5 findings, 1 overlap. Adjudicator: 3 accepted, 2 rejected.
→ Posting 3 inline comments. Fix loop: 3/3 fixed and verified locally. Threads resolved.
→ §5.5 re-fetching PR_HEAD_SHA → e8f9a01 (3 fix-loop commits since preflight).
→ Bounded poll for CI on e8f9a01 → 4min elapsed, all jobs done.
→ Re-classify: CI status: red-code. Failing: integration-tests.
→ Diff vs preflight: integration-tests was success → failure. REGRESSION detected.
→ Running §0j.2 regression resolution subroutine (max 1 round)...
→ Studying conventions, fetching failing logs: 'TypeError: Cannot read property "amount" of undefined at PaymentProcessor.process'
→ Spawning fix-attempt subagent: "Cannot determine root cause without product decision — fix #2 changed the shape of the payment payload, but tests in integration-tests/ assert the old shape. Fixing requires either reverting fix #2 or updating the integration tests, which depends on intended schema." → unresolvable.
→ Subroutine aborts. Emit blocker.

Verdict: WITHHELD — CI red after fix loop, regression resolution failed.

Preflight CI: green. Post-fix CI: red-code.
Failing jobs: integration-tests (https://github.com/.../actions/runs/22334)
Regressions detected: integration-tests (success → failure)
Skill commits since preflight: <sha1>, <sha2>, <sha3>

Recommended next action — review the failing jobs manually:
  gh run view 22334 --log-failed
  Or rerun: Skill(skill="compound-engineering:ce-work", args="fix the failing CI jobs on PR #117 introduced by ghcp-review-resolve commits <sha1>..<sha3>")

→ NOT submitting APPROVE. NOT closing. PR threads remain in current state (3 resolved by fix loop, 0 left open).
→ Summary: PR #117 — verdict WITHHELD. Fix loop introduced an integration-test regression we couldn't resolve. Left for human. No reviewers assigned. Not merged. REQUEST_CHANGES not submitted.
```
