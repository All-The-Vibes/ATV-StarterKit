---
title: "ghcp-review-resolve must call resolveReviewThread; in_reply_to does not auto-resolve"
date: 2026-04-27
module: skills/ghcp-review-resolve
component: pr-review-pipeline
problem_type: logic_error
category: logic-errors
tags:
  - github-api
  - graphql
  - pr-review
  - skill-design
  - copilot-review
symptoms:
  - "After ghcp-review-resolve runs, the PR still shows N unresolved conversations in the GitHub UI even though every finding was fixed and replied to"
  - "GraphQL `reviewThreads.nodes[].isResolved` returns false for every Copilot-authored thread the skill 'handled'"
  - "Skill summary reports 'all findings fixed' but reviewers are blocked because conversations are still open"
root_cause: "The skill replied to each review comment with `gh api .../comments -F in_reply_to=<id>` but never called the `resolveReviewThread` GraphQL mutation. `in_reply_to` posts a reply under an existing thread; it does not change the thread's `isResolved` state. Resolving a thread requires a separate mutation that takes the thread's GraphQL node ID — not the REST comment ID."
resolution_type: code-fix
---

## Problem

`ghcp-review-resolve` documented its own loop as "edit → test → commit → push → reply on the thread" and claimed Step 6 closed out the finding. It didn't. Replying via `in_reply_to` adds a comment under the bot's review thread but leaves `isResolved=false`. Result: the PR keeps accumulating "unresolved conversation" indicators even after every finding is fixed in code and the skill reports success.

Observed on PR #31 (this repo): 18 Copilot findings → 18 fixes pushed → 18 replies posted → **18 threads still unresolved** in the GitHub UI. Required a manual GraphQL sweep (`resolveReviewThread` mutation, 18 calls) to clear them.

## Symptoms

- `gh api graphql ... reviewThreads { isResolved }` returns `false` for every thread after the skill claims success.
- The PR's "X unresolved conversations" badge stays at the original count.
- `gh pr view --json reviewDecision` and any branch protection rule that requires "all conversations resolved" continues to block merge.
- The skill's final summary says "fixes shipped, replies posted" with no mention of resolution state.

## What Didn't Work

1. **Re-running the skill on the same HEAD.** Idempotency check passed (`PRIOR_RESOLVED` derived from `isResolved`), but since the threads were never resolved, every re-run treated them as fresh and re-requested Copilot — which then declined to re-emit (no new code to flag). Bot was removed from `requested_reviewers` without a new review. 10-minute poll timed out reporting "no new findings". The threads stayed unresolved.

2. **Forcing Copilot re-request.** Same outcome — Copilot won't re-emit the same findings on unchanged code, so even a forced re-review produces nothing actionable, and the skill has no codepath that resolves stale-but-fixed threads.

3. **Trusting `in_reply_to` to resolve.** The GitHub REST docs describe `in_reply_to` as posting a comment in the same thread. They do not say it resolves anything. The skill's prior wording ("the GitHub-supported way to thread under an existing review comment") was true but quietly conflated "reply on the thread" with "close out the thread".

## Solution

Two coordinated edits to `.github/skills/ghcp-review-resolve/SKILL.md` (mirrored to `pkg/scaffold/templates/skills/ghcp-review-resolve/SKILL.md`):

### 1. Capture `thread_id` during normalization (Step 3)

`resolveReviewThread` takes the thread's GraphQL node ID, not the REST comment ID. Build a `comment_id → thread_id` lookup once, paginating through all `reviewThreads`:

```bash
gh api graphql -F owner=<owner> -F repo=<repo> -F number=$PR_NUMBER -f query='
  query($owner:String!,$repo:String!,$number:Int!,$after:String) {
    repository(owner:$owner,name:$repo) {
      pullRequest(number:$number) {
        reviewThreads(first:100, after:$after) {
          pageInfo { hasNextPage endCursor }
          nodes {
            id                               # <-- thread node ID
            comments(first:100) { nodes { databaseId } }
          }
        }
      }
    }
  }'
```

Each finding now carries both `comment_id` (for the reply) and `thread_id` (for the resolve).

### 2. Add Step 6.7 — Resolve the thread (and Step 7 verification sweep)

Immediately after the `in_reply_to` reply, call the mutation:

```bash
gh api graphql -F threadId="$THREAD_ID" -f query='
  mutation($threadId: ID!) {
    resolveReviewThread(input: { threadId: $threadId }) {
      thread { isResolved }
    }
  }'
```

Verify `.data.resolveReviewThread.thread.isResolved == true`. Best-effort: if the mutation fails (permissions, already-resolved, transient error), log and continue — don't block the rest of the loop.

Step 7 (final summary) now runs a verification sweep that re-queries `reviewThreads` and lists any still-unresolved threads in the user-facing summary, so the skill can never again silently leave conversations open.

The repo's existing helper `.github/skills/resolve-pr-feedback/scripts/resolve-pr-thread` runs the same mutation if a future implementation prefers shelling out.

## Why This Works

- `resolveReviewThread` is the only GraphQL mutation that flips `PullRequestReviewThread.isResolved`. There is no REST equivalent.
- The mutation is idempotent on already-resolved threads (returns `isResolved=true` either way), so retries and double-runs are safe.
- Capturing `thread_id` once during normalization avoids a per-finding round trip; the lookup map is built from a single paginated query.
- The Step 7 verification sweep guarantees the skill's success summary matches reality — if any threads remain open, the user sees them named, not silently shipped as "all done".

## Prevention

1. **Default suspicion on REST reply endpoints.** Posting in a thread ≠ resolving a thread on GitHub. Whenever a skill's contract includes "close out a review", confirm the operation that flips `isResolved` is actually being called. The same pattern applies to issue comments (replying ≠ closing the issue) and to draft PRs (commenting ≠ marking ready).

2. **Verification sweep at the end of any "do N things" skill.** Before printing the success summary, re-query the system for the state the skill claims to have produced. The cost is one query; the benefit is never gaslighting the user about completion.

3. **Reviewer-script reuse.** This repo already had a working `.github/skills/resolve-pr-feedback/scripts/resolve-pr-thread` shell helper that ran the right mutation. Step 6 should have linked to it from day one. When designing a multi-step skill, audit existing scripts in sibling skills for primitives you'd otherwise rebuild from scratch.

4. **Test against branch-protection requirements.** Any "PR review automation" skill should be tested on a repo with a "must resolve all conversations" rule active. That catches the silent-unresolved class of bug immediately, since the PR refuses to merge.

### Quick verification snippet

Drop into any future PR-review skill to assert all threads it touched are resolved:

```bash
unresolved=$(gh api graphql -F owner=<o> -F repo=<r> -F number=<n> -f query='
  query($owner:String!,$repo:String!,$number:Int!) {
    repository(owner:$owner,name:$repo) { pullRequest(number:$number) {
      reviewThreads(first:100) { nodes { isResolved } } } } }' \
  | jq '[.data.repository.pullRequest.reviewThreads.nodes[] | select(.isResolved|not)] | length')
[ "$unresolved" = "0" ] || echo "WARN: $unresolved threads still unresolved"
```
