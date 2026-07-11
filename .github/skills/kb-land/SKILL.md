---
name: kb-land
description: "Internal KB integration phase. Applies an explicit direct or merge delivery policy after kb-finalize, verifies the remote default branch contains the delivered commit, and runs configured post-integration synchronization. Normally invoked by kb-complete."
argument-hint: "[reviewed manifest path]"
---

# KB Land

Integrate reviewed work into the remote default branch and synchronize approved
post-integration targets. This is an internal delivery phase, not the default
user entrypoint.

## Authorization

Proceed only when one of these is true:

- `kb-complete` delegated with project `delivery.mode: direct`; or
- `kb-complete` delegated after PR delivery with
  `delivery.merge: auto-after-checks`; or
- the user explicitly requested merge/direct integration in this run.

Repository ownership, collaborator status, or write permission alone is not
authorization.

## Preconditions

1. Require reviewed manifest and `complete-to-ship: passed|quarantined`.
2. Require audited final scope and release checks from `kb-ship`, or run the
   equivalent audit before direct integration.
3. Fetch the selected remote and resolve its actual default branch.
4. Require no unrelated staged files, unmerged entries, secrets, unexplained
   binaries, or scope drift.
5. Inspect branch protection, required checks, approvals, PR state, remote
   ancestry, and local/remote SHAs.
6. Never use force push, admin bypass, hook bypass, or destructive history
   rewriting.

## PR Integration

For `delivery.mode: pr` with authorized auto-merge:

1. Require the exact open PR produced or verified by `kb-ship`.
2. Require correct base repository/default branch and required checks/approvals.
3. Merge using repository-supported policy without bypass flags.
4. Fetch and prove the remote default contains the PR head or resulting merge
   commit.

If checks are pending, leave the PR open or enable repository-native auto-merge
only when policy explicitly allows it. Report `pending-review`, not landed.

## Direct Integration

For explicit `delivery.mode: direct`:

1. Require write access and a policy that permits direct default delivery.
2. Require the candidate commit to be based on the current fetched default.
3. Run final release proof against the exact tree being delivered.
4. Push without force to the resolved default ref.
5. If protection or non-fast-forward rejection occurs, fall back to
   `kb-ship`/PR when policy permits; otherwise block.
6. Fetch again and prove the remote default contains the delivered commit.

## Post-Integration Sync

Run only when `post_merge_sync: true` and remote-default integration is proven.

1. Treat merged Git as source of truth.
2. Never copy unreviewed global/installed drift back into Git automatically.
3. Use the repository's documented sync/install workflow and approved targets.
4. For skill bundles, inspect drift first, sync source outward, and verify
   hashes. Thin packaging variants require their own documented propagation
   policy; do not overwrite them as mirrors.
5. Record command, targets, hashes, and failures. A sync failure means
   integration succeeded but terminal `landed` remains blocked until policy is
   satisfied or the user explicitly waives that target.

## Cleanup

- Fetch and update local knowledge of the remote default without discarding
  unrelated work.
- Delete topic branches/worktrees only when merged, clean, and safe.
- Do not switch branches in a dirty shared checkout merely for cosmetic cleanup.

## Output

```text
KB land: landed|pending-review|blocked|nothing-to-land
Mode: pr|direct
Remote default: <remote>/<branch>
Integrated commit: <sha or none>
PR: <url or none>
Sync: <done|not-configured|blocked>
Blocker: <none or exact reason>
```
