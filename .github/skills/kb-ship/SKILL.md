---
name: kb-ship
description: "Internal checked-in PR delivery for reviewed KB work. Validates complete-to-ship, audits final scope, commits intentional files, pushes a non-default topic branch, and creates or updates a correctly based PR without merging. Normally invoked by kb-complete."
argument-hint: "[reviewed manifest path]"
---

# KB Ship

Deliver reviewed KB work as a pushed topic branch and PR. Explicit invocation,
including delegation from `kb-complete`, authorizes commit, push, and PR creation.
It never authorizes merge, force-push, hook bypass, or default-branch delivery.

## Preconditions

1. Read repo instructions, `todo.md`, `todo-done.md`, and the manifest.
2. Require `status: reviewed` and `complete-to-ship: passed|quarantined`.
   Validate through `kb-gate`; when repo `cmd/kbcheck` exists, also run:

   ```powershell
   go run ./cmd/kbcheck gate-ledger --manifest <manifest> --gate complete-to-ship --allow-quarantine
   ```

   Without either validator, inspect required evidence, proof, blockers,
   `passed_at`, and allowed next action directly and record
   `ship-gate-validator: unavailable`.
   For a quarantined gate, also require owner, evidence, forbidden claims, and
   an explicit proof that quarantined paths do not overlap the final shipping
   scope. Any overlap blocks shipping.
3. Resolve the current branch, remotes, remote default branch, upstream, local
   and remote SHAs, worktrees, index, unstaged/untracked files, and existing PR.
4. Before mutation, require `gh`, authentication for the push remote's host,
   repository access, and permission to create or update a PR. Missing
   prerequisites block before commit or push.

## Final Shipping Scope

`kb-finalize` may add manifest, todo, solution, instinct, memory, and cleanup
artifacts after `kb-work` first recorded scope.

1. Reconcile `scope-verified-files` with the actual final diff.
2. Classify every additional completion artifact against `kb-finalize` output.
3. Record the final audited shipping scope in the manifest.
4. Block on unexplained files, secrets, credentials, large binaries, generated
   bulk output, or unrelated edits.
5. Inspect the existing index. If any pre-staged path is outside final shipping
   scope, stop; never silently commit or unstage another user's work.

After deliberate staging, require the staged path set to be a subset of the
audited shipping scope and inspect the staged diff.

## Branch Safety

1. Fetch the selected push remote and resolve its default branch.
2. If currently on the default branch:
   - with uncommitted work and no unpublished default-branch commits, create a
     descriptive topic branch before staging;
   - with unpublished commits already attached to the default branch, stop with
     exact recovery context. Do not rewrite the local default ref automatically.
3. Refuse any topic branch whose configured upstream targets the remote default
   branch.
4. Select a non-default remote topic ref matching the local topic branch. Push
   explicitly as `HEAD:refs/heads/<topic>`; do not rely on a dangerous existing
   upstream.

## Verification and Commit

Run verification against the tree that will be delivered.

- When repo `cmd/kbcheck` exists, run `go run ./cmd/kbcheck local-release`.
- Otherwise invoke `kb-check` and run the consuming repo's native release
  checks; record the fallback commands and results.
- Run both unstaged and staged whitespace checks:
  `git diff --check` and `git diff --cached --check`.
- Run required browser/API/CLI release proof from the manifest.

Stage only explicit paths; never use `git add .` or `git add -A`. Commit with a
concise conventional message and repo-required trailers. Do not create an empty
commit.

## Remote Integration and Push

- If the remote topic branch exists, fetch it and integrate without force or
  shared-history rewriting.
- After any merge, rebase, or conflict resolution changes `HEAD`, rerun all
  release checks and re-audit the full branch diff against the remote default
  branch before push. Remote commits do not bypass shipping scope review.
- Require no unmerged entries and a clean audited index.
- Push explicitly to the selected non-default remote topic ref and set that ref
  as upstream.
- Fetch again and require local `HEAD` and the upstream SHA to match exactly.
- A failed push is a blocker to resolve, not success.

## Pull Request

Inspect PRs by exact head repository/ref.

- Existing PR success requires open state and the resolved remote default branch
  as base. Block before retargeting a PR with a different base or repository.
- Otherwise create a PR against the resolved default branch.
- Summarize the full branch diff: Summary, Verification, Slices Completed,
  Parked/Deferred, Risks, Follow-Up.
- Put quarantined gate details in Risks.
- Record the PR URL. Never merge without a separate explicit merge request.

## Terminal State

`shipped` requires release checks, audited staged scope, committed changes,
matching local/upstream SHAs, a correctly based open PR, and a clean worktree
except explicitly documented unrelated files.

If no local/remote branch delta exists, still create a missing PR when the
topic branch has deliverable commits. Use `nothing-to-ship` only when no
deliverable commit or PR work exists.

```text
KB ship: shipped|nothing-to-ship|blocked
Branch: <branch>
Commit: <sha or none>
PR: <url or none>
Blocker: <none or exact reason>
Next: <none or exact recovery action>
```
