---
name: ghcp-review-resolve
description: Request a GitHub Copilot review AND a pr-review-toolkit review on the current PR, wait for both, adjudicate their findings with an independent subagent, post inline PR comments for verified bugs/fixes only, then run a tight inline fix-and-reply loop per comment (test, commit, reply on thread). After each verified fix lands, auto-resolves the matching Copilot review thread via GraphQL and ticks the corresponding PR-body task-list item. Surfaces merge conflicts and prior-review state as explicit preflight output so the skill stops cleanly instead of fighting reality. Use whenever the user invokes /ghcp-review-resolve, asks to "run copilot review and resolve", asks to "review and fix my PR with copilot", asks for a "dual review and fix pass", or wants automated bot-review triage and remediation on a pull request they just opened. Does NOT close, approve, or merge the PR.
---

# ghcp-review-resolve

Orchestrates a dual-review-and-fix pipeline on an open PR. The workflow:

0. **Preflight** — detect PR, fetch size + merge state + head SHA, probe Copilot availability, check for prior resolved reviews. Emit a preflight table. If there's a blocker (merge conflict, nothing useful to do), report it and stop cleanly — no side effects.
1. Request Copilot + pr-review-toolkit reviews in parallel (skip Copilot if unavailable or already resolved)
2. Poll until expected reviewers are done (30s interval, 10-minute cap)
3. Independently adjudicate findings via a subagent that inspects the actual code
4. Post inline PR comments only for verified bugs/fixes
5. Run a tight inline fix loop per comment: edit → test → commit → reply on the thread
6. Summarize — never close, approve, or merge

## Why this exists

Copilot reviews and the repo's own review toolkit both produce lots of findings. Some are real bugs. Some are stylistic noise. Some contradict each other. Blindly "fix everything the bots said" is how you ship regressions or waste a day on non-issues.

This skill's job is to be the adult in the room: collect both reviews (when both are available), have an independent reviewer verify each finding against the actual code, and only act on what's real. Overlapping findings are high-confidence. Unique findings are kept only when high-severity and verifiable.

It also knows when **not** to run. A PR with an unresolved merge conflict, or a PR whose prior Copilot review is already fully resolved, shouldn't trigger another round of bot noise — the skill reports that state and gets out of the way.

## Step 0 — Preflight

The preflight is the first and only place allowed to abort the run. If it passes, every later step trusts its flags. If a blocker is reported, no reviewer is contacted, no comment is posted, no fix is attempted.

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

`small` → Step 4 uses the full-diff path (`gh pr diff`).
`large` → Step 4 uses the per-file paginated path (`gh api .../pulls/{n}/files --paginate`). This avoids `gh pr diff`'s 20k-line API cap.

### 0e. Check merge state

If `MERGE_STATE_STATUS == "UNKNOWN"`, GitHub hasn't finished computing mergeability (common right after a push). Wait up to 30 seconds, re-fetching every 10s, then proceed with whatever state is reported.

If the final `MERGE_STATE_STATUS == "DIRTY"` (has conflicts with base), this is a **blocker**. Emit the preflight table (see 0h) and the following recommended-action block, then stop:

```
Blocker: PR has merge conflicts with base (mergeStateStatus=DIRTY).

Recommended next action — resolve conflicts before re-running this skill:

  Option A (manual):
    git fetch origin && git rebase origin/<base>
    # resolve conflicts, then: git push --force-with-lease

  Option B (delegated):
    Skill(skill="compound-engineering:ce-work", args="resolve the merge conflicts on PR #<N>")
```

No reviewers are contacted. No comments are posted. No commits are made. The skill exits cleanly.

### 0f. Probe Copilot availability (non-mutating)

Determine whether Copilot code review is available on this repo **without** actually requesting a review — preflight must have no side effects so that short-circuiting at 0i is truly free.

Use the repo's suggested reviewers and/or the Copilot code-review config to probe. Two practical non-mutating checks, in order of preference:

```bash
# 1. Ask GitHub which reviewers can be requested for this PR. Copilot, when
#    available, shows up in the suggested/possible reviewer list.
gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/requested_reviewers" >/dev/null 2>&1

# 2. Inspect any prior Copilot review on the PR (from /reviews). If the bot has
#    ever posted here, it is available.
gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/reviews" --paginate \
  | jq -e '.[] | select(.user.login | test("copilot.*\\[bot\\]|github-copilot\\[bot\\]"))' \
  >/dev/null 2>&1
```

- Either check succeeding (prior Copilot review found, or suggested-reviewers query succeeds and the repo is known to have Copilot code review enabled) → `COPILOT_AVAILABLE=true`.
- Otherwise → `COPILOT_AVAILABLE=false`. The **actual** `--add-reviewer @copilot` call happens in Step 1, where its failure with `422`/`403`/`not a collaborator` is the authoritative signal and causes the skill to demote to single-reviewer mode for the rest of the run.

This split matters: 0f must not mutate PR state. Running `gh pr edit --add-reviewer @copilot` here would request a Copilot review, and if 0i then short-circuits (or 0e flags a merge conflict blocker), the user is left with a spurious review request. All mutations live in Step 1.

When `COPILOT_AVAILABLE=false`, the skill continues in **single-reviewer mode** with pr-review-toolkit only. Log this clearly:

```
Copilot unavailable on this repo (reviewer assignment returned 422 / not a collaborator).
Continuing in single-reviewer mode with pr-review-toolkit only.
To enable dual review, configure Copilot code review on the repo settings.
```

### 0g. Check for prior resolved reviews (idempotency)

Use GraphQL to fetch review threads and resolution state. Paginate through **all** review threads before deciding whether a prior Copilot review was already resolved — a 100-thread cap will silently miss findings on large PRs. Align the query shape with the repo's existing working reviewer script (`.github/skills/resolve-pr-parallel/scripts/get-pr-comments`), which is the source of truth for field names that actually exist on `PullRequestReviewThread`:

```bash
all_threads='[]'
after=null

while :; do
  page="$(gh api graphql \
    -F owner=<owner> \
    -F repo=<repo> \
    -F number=$PR_NUMBER \
    -F after="$after" \
    -f query='
      query($owner:String!, $repo:String!, $number:Int!, $after:String) {
        repository(owner:$owner, name:$repo) {
          pullRequest(number:$number) {
            reviewThreads(first: 100, after: $after) {
              pageInfo { hasNextPage endCursor }
              nodes {
                id
                isResolved
                isOutdated
                viewerCanResolve
                path
                line
                comments(last: 1) {
                  nodes {
                    databaseId
                    author { login }
                    updatedAt
                  }
                }
              }
            }
          }
        }
      }')"

  all_threads="$(jq -c \
    --argjson existing "$all_threads" \
    '$existing + (.data.repository.pullRequest.reviewThreads.nodes // [])' \
    <<<"$page")"

  has_next="$(jq -r '.data.repository.pullRequest.reviewThreads.pageInfo.hasNextPage' <<<"$page")"
  [ "$has_next" = "true" ] || break
  after="$(jq -r '.data.repository.pullRequest.reviewThreads.pageInfo.endCursor' <<<"$page")"
done
# Evaluate PRIOR_RESOLVED from "$all_threads" only after pagination completes.
```

Notes on the shape:

- `comments(last: 1)` returns the most recent comment in the thread, not the oldest (GraphQL default order on this connection is oldest-first, so `first: 1` would give the wrong record for any freshness check).
- `commit { oid }` is intentionally **not** requested — it isn't reliably available on the review-comment node across schema versions. Freshness is derived from `isOutdated` (GitHub's own signal that HEAD has moved past the comment's anchor) instead of comparing commit SHAs by hand.
- `id` (the thread-level node ID) and `viewerCanResolve` are captured on every thread so **Step 6.7** can invoke `resolveReviewThread` without re-querying. Persist these alongside the comment `databaseId` in the per-thread record (e.g., `comments_by_id.json`: `{ comment_databaseId: { threadId, viewerCanResolve, path, line } }`).

Classify each Copilot-authored thread (author login matching `github-copilot[bot]` or `copilot-pull-request-reviewer[bot]`):

- **resolved-and-fresh** — `isResolved=true` AND `isOutdated=false`. The resolution still applies to current HEAD.
- **resolved-but-stale** — `isResolved=true` AND `isOutdated=true`. The code around the fix has moved; the resolution may no longer apply.
- **open** — `isResolved=false`.

Derive:

```
PRIOR_RESOLVED = (there is at least one Copilot thread) AND
                 (every Copilot thread is resolved-and-fresh)
```

When `PRIOR_RESOLVED=true`:
- Skip Copilot re-request in Step 1 (do not assign `@copilot` again — the skill already did its work).
- Log: `Prior Copilot review found with all threads resolved at current HEAD (<sha>). Skipping Copilot re-request.`
- pr-review-toolkit still runs by default — it's an independent reviewer and may find new things.

**Escape hatch:** if the user passed `--force` (or explicitly asked to re-request Copilot), ignore `PRIOR_RESOLVED` and request anyway. Exact argument syntax is up to the skill's arg-handling layer; honor the user's explicit intent.

**REST fallback:** if the GraphQL query fails (API change, auth scope, older `gh` version), fall back to reading `/pulls/{n}/reviews` and `/pulls/{n}/comments`. Heuristically mark a Copilot review thread as resolved if each top-level Copilot comment has a subsequent reply on the same thread from the PR author or a maintainer. This is less precise than the GraphQL path, so when in doubt, set `PRIOR_RESOLVED=false` and proceed with re-review.

### 0h. Emit the preflight table

Before doing anything with side effects, print a compact preflight report. The shape (modeled on what the user actually got when the skill failed on PR #9):

```
Preflight — PR #<N>

  Check                      Status
  ────────────────────────── ──────────────────────────────────────────────────
  gh auth                    ok
  PR detected                #<N> (<branch>, head <short-sha>)
  Size                       <files> files / +<add> −<del>  [<size_class>]
  Merge state                <CLEAN | DIRTY | BLOCKED | UNKNOWN>
  Copilot available          <yes | no — 422/not a collaborator>
  Prior Copilot resolved     <yes — skipping re-request | no | n/a>
  pr-review-toolkit          will run

Decision: <proceed to Step 1 | STOP — blocker: ...>
```

### 0i. Short-circuit: nothing useful to do

If, after 0a–0g, all of the following hold:

- `COPILOT_AVAILABLE=false` OR `PRIOR_RESOLVED=true`
- `SIZE_CLASS == "large"`
- No `--force`

... then this run has low expected value: there's no new Copilot reviewer to fire, pr-review-toolkit alone on a 10k+ line PR is expensive for ambiguous signal, and the user is probably better served by merging or by narrower review. Emit the preflight table, log a one-paragraph rationale, and stop cleanly.

Otherwise proceed to Step 1.

### 0j. Record preflight flags

Later steps read these flags; they are the contract between preflight and the rest of the skill:

- `PR_NUMBER`, `PR_HEAD_SHA`, `BASE_REF`
- `SIZE_CLASS` ∈ {small, large}
- `COPILOT_AVAILABLE` ∈ {true, false}
- `PRIOR_RESOLVED` ∈ {true, false}
- `EXPECTED_REVIEWERS` — the set of reviewers the skill will actually wait for. Derived:

```
EXPECTED_REVIEWERS = {"pr-review-toolkit"}
if COPILOT_AVAILABLE and not PRIOR_RESOLVED: EXPECTED_REVIEWERS.add("copilot")
```

If `EXPECTED_REVIEWERS == {}`, the skill has nothing to do — report and stop.

**Any later step that references preflight flags but finds them unset must refuse to run** (defense against the preflight ever being accidentally skipped).

## Step 1 — Request reviews in parallel

**Fire the expected reviewers in the same turn.** Don't serialize — that wastes wall-clock time.

### 1a. Request Copilot review (conditional)

Only if `"copilot" ∈ EXPECTED_REVIEWERS`:

```bash
gh pr edit "$PR_NUMBER" --add-reviewer @copilot
```

Note: Step 0f already probed this, so assignment should succeed here. If it still fails at this point (transient error), log it and remove "copilot" from `EXPECTED_REVIEWERS` — do not abort the pipeline.

If `"copilot" ∉ EXPECTED_REVIEWERS` (unavailable or prior-resolved), skip this step silently; preflight already logged the reason.

### 1b. Invoke pr-review-toolkit

Always runs (pr-review-toolkit is always in `EXPECTED_REVIEWERS`):

```
Skill(skill="pr-review-toolkit", args="<PR URL or #PR_NUMBER>")
```

The pr-review-toolkit review typically posts its findings as PR review comments. Capture any IDs/markers it returns so you can distinguish its comments later.

## Step 2 — Poll until expected reviewers complete

Poll every **30 seconds**, cap at **10 minutes** (20 attempts). The poll waits **only for reviewers in `EXPECTED_REVIEWERS`** — don't wait on Copilot if it wasn't requested.

On each tick, check results:

```bash
# All review comments on this PR
gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/comments" \
  --paginate > /tmp/ghcp-review-comments.json

# Top-level reviews (includes Copilot's "review submitted" events)
gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/reviews" \
  --paginate > /tmp/ghcp-review-reviews.json
```

**Signals that a review is "done":**

- **Copilot** (only if in `EXPECTED_REVIEWERS`): A review from `github-copilot[bot]` or `copilot-pull-request-reviewer[bot]` exists in `/reviews` with `state` set, OR inline comments from that bot are visible. Capture both line comments and the top-level review body.
- **pr-review-toolkit**: Either the `Skill()` invocation returned, or its posted comments are visible on the PR. Prefer the skill call's return as the completion signal.

If after 10 minutes any expected reviewer is still missing:

- If **at least one** expected reviewer completed, proceed with what's available and note the gap in the final summary.
- If **none** completed, stop and report. Do not fabricate findings.

Between polls, use `Bash` with `sleep 30`. Don't use `ScheduleWakeup` — the 30s cadence is too tight and the user is actively waiting.

## Step 3 — Collect and normalize findings

Build a single list of findings from whichever sources produced results. For each:

```
{
  "source": "copilot" | "pr-toolkit",
  "file": "path/to/file.go",
  "line": 42,              // nullable — some findings are file-level
  "severity": "...",       // pr-toolkit provides this; copilot usually doesn't
  "body": "...",           // the raw review text
  "comment_id": 12345      // GitHub review comment ID for reply/resolve
}
```

Deduplicate near-identical findings (same file + overlapping line range + substantively similar body) and mark them as `overlap: true`. Overlap is your strongest positive signal.

**Single-reviewer mode:** when only one source produced findings, there is no overlap to detect. That's fine — the adjudicator (Step 4) still operates on unique findings, it just loses the "both bots flagged it" signal and must rely entirely on severity + verifiability.

## Step 4 — Adjudicate findings with a subagent

Spawn a fresh subagent (general-purpose or code-reviewer) to independently verify each finding against the actual code. This subagent did NOT write the code and did NOT write the reviews — that independence is the whole point.

**Diff-fetch strategy depends on `SIZE_CLASS`:**

- `small` → include full PR diff in the adjudicator's context:
  ```bash
  gh pr diff "$PR_NUMBER" > /tmp/ghcp-pr-diff.patch
  ```
- `large` → fetch per-file patches and only give the adjudicator the files referenced by findings, plus file list:
  ```bash
  gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/files" --paginate \
    > /tmp/ghcp-pr-files.json
  ```
  The adjudicator reads per-file `patch` fields as needed and can request additional file context via the Read tool. Cap adjudicator per-file reads at ~30 files per run; if findings span more, process in batches.

Prompt shape (adapt to the chosen mode):

> You are adjudicating a set of automated PR review findings on PR #`<N>` at SHA `<sha>` in `<repo>`. Diff-fetch mode: `<small: full diff attached | large: per-file patches for files referenced by findings>`.
>
> For each finding, read the referenced file/line, decide whether the finding is a real bug, logic error, security issue, or concrete correctness problem that warrants a code change. Reject style preferences, speculative refactors, "consider adding a comment" suggestions, and anything not grounded in code you can actually see.
>
> Keep a finding if:
> - Both reviewers flagged it (overlap), AND it is a real issue you can verify in the code, OR
> - Only one reviewer flagged it, AND it is high-severity (bug, security, data loss, incorrect logic, broken contract) AND verifiable.
>
> **Drop** any finding that references a file or line not present in the PR diff — such findings are not grounded in the changes.
>
> Return JSON: `[{file, line, severity, rationale, proposed_fix, source_comment_ids: [...]}, ...]`.

Run tests/build before adjudication if cheap — a failing test is ground-truth evidence. The subagent is allowed (and encouraged) to actually run the test suite if it helps verify a finding.

## Step 5 — Post inline PR comments for accepted findings only

Post **one** PR review (`event: COMMENT`) that batches every accepted finding as a `comments[]` entry. One review, many inline comments — not one review per finding. This keeps the PR timeline readable and makes it obvious which comments were produced by this skill vs. the original bots.

Use `gh api` with a single call, repeating the `-F "comments[][...]=..."` flags for each finding (`gh`/`curl` build an array from repeated keys):

```bash
gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/reviews" \
  -X POST \
  -f commit_id="$PR_HEAD_SHA" \
  -f event="COMMENT" \
  -f body="ghcp-review-resolve: verified findings to address" \
  -F "comments[][path]=path/to/file.go" \
  -F "comments[][line]=42" \
  -F "comments[][body]=**Verified finding** (from: copilot, pr-toolkit)\n\n<rationale>\n\n**Proposed fix:** <proposed_fix>" \
  -F "comments[][path]=path/to/other.go" \
  -F "comments[][line]=17" \
  -F "comments[][body]=**Verified finding** (from: copilot)\n\n<rationale>\n\n**Proposed fix:** <proposed_fix>"
```

Skip Step 5 entirely when every accepted finding is a verbatim confirmation of an existing bot comment already anchored at the correct file/line — re-posting the same finding as a `ghcp-review-resolve` comment just duplicates noise. In that case, move straight to Step 6 and reply on each original thread. Log the decision ("accepted findings already anchored as bot comments; skipping duplicate post").

Prefix each comment body with `ghcp-review-resolve:` so later steps can identify comments this skill created vs. comments from the bots themselves.

Before posting, re-check `PR_HEAD_SHA` against the current HEAD. If it changed (someone else pushed), stop — see Guardrails.

Do NOT:
- Submit `event: APPROVE`
- Submit `event: REQUEST_CHANGES`
- Close, merge, or mark-ready any PR

## Step 6 — Inline fix loop, sequentially per comment

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

5. **Push** after each commit (so the reply comment can point at a real pushed SHA). Use `--force-with-lease` so a parallel push from a teammate or CI is detected rather than silently clobbered:
   ```bash
   git push --force-with-lease
   ```
   If the push is rejected because the remote moved, do **not** retry blindly. Re-fetch `PR_HEAD_SHA` and stop the fix loop — see Guardrails. The user decides whether to rebase and restart.

6. **Reply** on the specific review comment thread with what changed and why:
   ```bash
   gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER/comments" \
     -X POST \
     -f body="Fixed in <commit_sha>: <one-line description>. Verification: <tests run / build status>." \
     -F in_reply_to=<comment_id>
   ```
   The `in_reply_to` field is the GitHub-supported way to thread under an existing review comment. If that call fails (some older API versions), fall back to posting a new top-level PR comment that references the original comment URL.

   **Body formatting gotcha:** if the reply body contains shell-special characters (`&&`, `$`, backticks, etc.), pass it via a variable with `-F body="$REPLY_BODY"` rather than `-f body='…'` — `-f` interprets the string and has bitten this skill before (see earlier session note on comment 3139049925). When in doubt, write the body to a temp file and use `-F body=@reply.txt`.

7. **Resolve the thread (Step 6.7).** See the dedicated section below — this is a new guardrailed step that must run before moving on.

8. **Move on** to the next finding. Do not pause for user input between findings — the whole point is one-shot resolution. If something truly blocks progress (repo credentials, missing dependency), stop the whole pipeline and report.

## Step 6.7 — Resolve the review thread (guardrailed)

After replying on a thread, mark it resolved on GitHub so the PR page reflects reality. This is gated by **five** checks; if **any** gate fails, skip resolution and record why. Never resolve a thread that hasn't been properly addressed.

**Gate checklist** (all must be `true` to proceed):

| Gate | Check |
|------|-------|
| G1. Adjudicated real | The finding was accepted in Step 4 as a real issue (not rejected / not a false positive). |
| G2. Fix landed | A commit addressing the finding was pushed to the PR branch in this session (Steps 6.4–6.5). |
| G3. Verified | Step 6.3 verification passed, OR the reply body explicitly documents the limitation ("not independently verified — review reply documents intent"). |
| G4. Reply posted | Step 6.6 successfully posted a reply on the thread (non-error HTTP 201 and the reply body is what was intended). |
| G5. Permission | `viewerCanResolve == true` on the thread node captured in Step 0g. If `false`, the auth token lacks `Pull requests: write` for this repo — do not attempt. |

If any gate fails, log `"thread <threadId> not auto-resolved: gate <G#> failed — <reason>"` and skip to the next finding. **Manual human review decides those.**

**Resolution call** (when all gates pass):

Call the GraphQL `resolveReviewThread` mutation directly via `gh api graphql`. `$THREAD_ID` comes from the Step 0g map (`reviewThreads.nodes[].id`):

```graphql
mutation($threadId: ID!) {
  resolveReviewThread(input: { threadId: $threadId }) {
    thread { id isResolved viewerCanResolve }
  }
}
```

**Post-conditions and error handling:**

- On success, verify `isResolved == true` in the response. If the mutation returned `200 OK` but `isResolved == false`, treat it as a failure (don't assume).
- **Idempotency:** re-resolving an already-resolved thread is treated as a no-op success by GitHub's API (community-observed; not formally documented). If you encounter a `GraphQL error: already resolved`, log it and continue — do not retry.
- **On transient error** (HTTP 5xx, network timeout), retry **once** with 2s backoff. On repeated failure, leave the thread unresolved and record `"resolve failed for <threadId>: <error>"` in the Step 7 summary.
- **Never** resolve a thread authored by someone other than the Copilot reviewer — this skill only adjudicates Copilot (and, when enabled, pr-review-toolkit) findings.

**Why this is guardrailed:** auto-resolving threads changes reviewer-visible state. A false positive that gets fixed + resolved + replied-to is fine; a finding that was *rejected* but accidentally resolved hides a disagreement the PR author should see on the page. The five gates exist so resolution only happens when the record shows an end-to-end honest fix.

## Step 6.8 — Tick off matching PR task-list items

If the PR body contains a GitHub task list (`- [ ] …` / `- [x] …`), attempt to tick off items that a landed fix just completed. This closes the loop between "reviewer raised X" and "PR description says X is done."

**Extract and match:**

1. **Fetch the current body** (PR bodies have no optimistic-concurrency API, so re-fetch *immediately* before writing to minimize the clobber window):
   ```bash
   BODY="$(gh api "repos/{owner}/{repo}/pulls/$PR_NUMBER" --jq .body)"
   ```
2. **Locate task items.** Regex-extract `^(\s*)-\s+\[( |x|X)\]\s+(.+)$` grouped by line number. Only unticked items (`[ ]`) are candidates.
3. **Match fixes to items** using a deterministic pipeline — **do not** jump to the LLM for every fix:
   - **Pass A (normalized substring):** lowercase, strip punctuation, collapse whitespace. If the fix summary (commit title or reply body first sentence) normalized-contains the task item text normalized, it's a match.
   - **Pass B (fuzzy):** tokenize both strings (lowercase, strip punctuation, collapse whitespace), compute a token-overlap / token-set ratio, and accept the match at ≥0.85. Implementation may use any available fuzzy-matching library (e.g., Python `rapidfuzz.fuzz.token_set_ratio`, JS `fast-levenshtein`, Ruby `string-similarity`) or a hand-rolled equivalent — the specific library is not part of the contract.
   - **Pass C (LLM fallback, only for ambiguous cases):** if Passes A/B produced no match **and** the fix summary looks topically related (shared nouns), call an LLM with `temperature=0`, strict JSON output, and require `confidence >= 0.8` before accepting. Log every LLM call.
   - **Never guess.** Unmatched fixes → log `"fix <commit_sha> did not match any PR task item — manual triage needed"` and continue. **Silent drops are forbidden.**

**Update the body with a sentinel fence:**

Wrap the skill's updates in an HTML-comment fence so re-runs can idempotently find and refresh the block without corrupting surrounding prose:

```markdown
<!-- BEGIN:ghcp-audit v1 -->
<!-- machine-maintained by ghcp-review-resolve; do not edit between fences -->
Ticked by ghcp-review-resolve on <ISO-8601 UTC>:
- Copilot finding at `src/foo.ts:42` → commit <sha>
- …
<!-- END:ghcp-audit v1 -->
```

**Write safely:**

1. **Size guard:** reject the write if the new body would exceed **60 KB** (safety margin below GitHub's ~64 KiB de-facto limit). Log and fall back to posting a top-level PR comment listing the ticks instead.
2. **Normalize line endings** to `\n` before writing (GitHub normalizes CRLF→LF server-side, but be explicit).
3. **Write via `gh pr edit`:**
   ```bash
   gh pr edit "$PR_NUMBER" --body "$NEW_BODY"
   ```
4. **Verify round-trip:** re-fetch the body and confirm the ticks + sentinel block are present. If verification fails, restore the original body and post a top-level comment instead.

**Gates that skip Step 7.0:**

- No task list in PR body → skip silently.
- Body size >= 60 KB after update → skip and post a comment instead.
- `gh pr edit` returns non-zero → skip and log.
- Auth token lacks PR write permission → skip and log.

Step 6.8 failures never block Step 7 — the final summary to the user still runs.

## Step 6.9 — Appendix: batching and per-language verification

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

## Step 7 — Final summary to the user

Produce a single summary message covering:

- PR number and URL
- Preflight outcome (size class, merge state, Copilot availability, prior-resolved state)
- Reviewers expected vs. completed
- Findings: total raised, total accepted, total rejected (with top reasons), any dropped as "not grounded in diff"
- Fixes: what was changed, commit SHAs, any fixes that failed or were skipped
- **Thread resolution (Step 6.7):** count of threads auto-resolved, plus any that failed a gate — list each with the gate that failed (e.g., `"thread T_abc not resolved: G5 viewerCanResolve=false"`).
- **PR task-list ticking (Step 7.0):** count of task items ticked, the match method per tick (substring / fuzzy / LLM), and any unmatched fixes flagged for manual triage. If the body was too large or the write failed, say so and link to the fallback comment.
- Explicit confirmation: **PR was not closed, approved, or merged.**
- Remaining reviewer comments that were intentionally left unresolved, with a one-line reason each

Format as Markdown. Keep it scannable.

## Guardrails — do not cross these

- **If preflight reports a blocker, stop cleanly.** Do not proceed into Steps 1–7. No reviewer requests, no comments, no commits.
- **Never** run `gh pr merge`, `gh pr close`, `gh pr ready` (if it would change state unexpectedly), or submit an `APPROVE` review.
- **Never** act on a finding the adjudicator subagent rejected, even if both bots flagged it — the adjudicator is the tiebreaker.
- **Never** fabricate finding text. If a bot's comment is ambiguous, include the verbatim original in the inline comment so a human can sanity-check.
- **Never** silently drop all findings from one reviewer because of a parsing issue. If you can't parse, stop and report.
- If the PR head SHA changes mid-run (someone else pushed), stop fixing, report state, and let the user decide whether to restart.
- Treat missing preflight flags in any later step as a bug — refuse to run rather than assume defaults.

## Example runs

### Example 1 — happy path (small, clean, first-time PR)

```
User: /ghcp-review-resolve
→ Preflight:
    PR #42 (feat/add-payments, head abc123)
    Size: 7 files / +212 −41  [small]
    Merge state: CLEAN
    Copilot available: yes
    Prior Copilot resolved: no
    Decision: proceed
→ Requesting Copilot review... ok
→ Invoking pr-review-toolkit on PR #42... ok
→ Polling (30s, max 10min)...
  t=30s: copilot pending, pr-toolkit done
  t=60s: copilot done
→ 11 raw findings (6 copilot, 5 pr-toolkit; 3 overlap)
→ Adjudicator subagent verifying against src/ at abc123 [diff mode: full]...
→ 4 accepted (3 overlap + 1 unique high-severity), 7 rejected
→ Posting 4 inline comments on PR #42
→ Fix 1/4: null check in PaymentProcessor.go:88 → edit, go test ./payment/ ok → commit def456 → reply posted
→ Fix 2/4: off-by-one in pagination → edit, go test ./api/ ok → commit ghi789 → reply posted
→ Fix 3/4: missing error wrap → edit, go vet ok → commit jkl012 → reply posted
→ Fix 4/4: race in cache update → edit, go test ./cache/ -race FAIL on retry → reverted, skipped with note
→ Summary: PR #42 — 4 findings verified, 3 fixed, 1 skipped. Not closed/approved/merged.
```

### Example 2 — degraded path (merge conflict + prior resolved)

```
User: /ghcp-review-resolve
→ Preflight:
    PR #9 (fix/pr7-lint-fixes, head 025bb1f)
    Size: 143 files / +15044 −3262  [large]
    Merge state: DIRTY         ← blocker
    Copilot available: no — 422/not a collaborator
    Prior Copilot resolved: yes — 8 threads, all resolved at current HEAD
    Decision: STOP

Blocker: PR has merge conflicts with base (mergeStateStatus=DIRTY).

Recommended next action — resolve conflicts before re-running this skill:

  Option A (manual):
    git fetch origin && git rebase origin/main
    # resolve conflicts, then: git push --force-with-lease

  Option B (delegated):
    Skill(skill="compound-engineering:ce-work", args="resolve the merge conflicts on PR #9")

No reviewers contacted. No comments posted. No commits made. PR not closed/approved/merged.
```

### Example 3 — single-reviewer mode on a large PR without prior reviews

```
User: /ghcp-review-resolve
→ Preflight:
    PR #17 (feat/big-refactor, head 9ab12c3)
    Size: 42 files / +3100 −900  [large]
    Merge state: CLEAN
    Copilot available: no — 422/not a collaborator
    Prior Copilot resolved: n/a
    Decision: proceed (single-reviewer mode, pr-review-toolkit only)
→ Invoking pr-review-toolkit on PR #17... ok
→ Polling (30s, max 10min)... t=45s: pr-toolkit done
→ 8 findings (pr-toolkit only; no overlap signal)
→ Adjudicator subagent verifying [diff mode: per-file]... read 11/42 files
→ 3 accepted (high-severity + verifiable), 5 rejected (style / not grounded)
→ Posting 3 inline comments, running fix loop...
→ Summary: PR #17 — 3 findings verified, 3 fixed. Copilot unavailable on this repo; ran pr-review-toolkit alone. Not closed/approved/merged.
```
