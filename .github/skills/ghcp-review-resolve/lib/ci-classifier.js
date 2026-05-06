// CI status classifier — pure function over `gh run list` JSON.
//
// Input shape (from `gh run list --branch <b> --json status,conclusion,name,databaseId,headSha,event`):
//   runs: Array<{
//     databaseId: number,
//     name: string,
//     status: 'completed'|'in_progress'|'queued'|'waiting'|'action_required'|...,
//     conclusion: 'success'|'failure'|'cancelled'|'skipped'|'neutral'|'timed_out'|'action_required'|null,
//     headSha: string,
//     event: string,
//   }>
//
// Output: { status, considered, failing, blocking }
//   status ∈ 'green' | 'red-code' | 'red-flake-suspected' | 'pending' | 'stale-only' | 'none' | 'unknown'
//   considered: runs that were not filtered out by stale-SHA pruning
//   failing: failed runs (with optional flake hints)
//   blocking: whether this status should gate later phases (true except for 'unknown')
//
// 'stale-only' vs 'none':
//   - 'none': the API returned zero runs at all (repo has no Actions, or no runs ever triggered for this branch).
//   - 'stale-only': runs exist for this branch but every one belongs to an older SHA — typically the gap between
//     a fresh push and CI starting on it. Treated as effectively pending: no signal yet for the current HEAD,
//     but CI is wired up. The verdict layer treats 'stale-only' the same way it treats 'pending'.

'use strict';

// Log signatures that indicate infrastructure / runner / network flake rather than code bug.
// Conservative allowlist — anything else is treated as code-caused.
const FLAKE_SIGNATURES = [
  /ECONN(RESET|REFUSED|ABORTED)/i,
  /ETIMEDOUT/i,
  /TLS handshake timeout/i,
  /net\/http: TLS handshake timeout/i,
  /pull access denied/i,
  /failed to pull image/i,
  /Error response from daemon/i,
  /runner has received a shutdown signal/i,
  /lost communication with the server/i,
  /The operation was canceled\./i,
  /The runner was lost/i,
  /No space left on device/i,
];

function looksLikeFlake(logText) {
  if (!logText) return false;
  return FLAKE_SIGNATURES.some((re) => re.test(logText));
}

function isCompleted(run) { return run.status === 'completed'; }
function isPending(run)   { return run.status === 'in_progress' || run.status === 'queued' || run.status === 'waiting' || run.status === 'action_required'; }

function classifyCi({ runs, headSha, logFor } = {}) {
  // gh API failure (null/undefined runs) — non-blocking unknown.
  if (runs == null) {
    return { status: 'unknown', considered: [], failing: [], blocking: false };
  }
  if (!Array.isArray(runs)) {
    return { status: 'unknown', considered: [], failing: [], blocking: false };
  }

  // Filter out stale runs from older SHAs.
  const considered = headSha
    ? runs.filter((r) => r.headSha === headSha)
    : runs.slice();

  if (considered.length === 0) {
    // Distinguish "no runs at all" (repo has no Actions / nothing triggered) from
    // "runs exist but only for older SHAs" (fresh push, CI hasn't started yet).
    // The first is a non-gate; the second must be treated as pending so the verdict
    // layer doesn't APPROVE a SHA that has never been seen by CI.
    if (headSha && runs.length > 0) {
      return { status: 'stale-only', considered: [], failing: [], blocking: true };
    }
    return { status: 'none', considered: [], failing: [], blocking: true };
  }

  // Bucket runs by terminal state.
  const failing = [];
  let pending = false;

  for (const run of considered) {
    if (isCompleted(run)) {
      const c = run.conclusion;
      if (c === 'failure' || c === 'timed_out' || c === 'cancelled') {
        failing.push(run);
      } else if (c === 'action_required') {
        // pending-equivalent — needs human action, not failed.
        pending = true;
      }
      // success | skipped | neutral → no-op (counted toward green).
    } else if (isPending(run)) {
      pending = true;
    } else {
      // unknown status string — be conservative and treat as pending.
      pending = true;
    }
  }

  if (failing.length > 0) {
    // Flake heuristic: ALL failing runs match an infra signature => red-flake-suspected.
    // Mixed bag (some flake + some code) => red-code (don't paper over real failures).
    const adapter = typeof logFor === 'function' ? logFor : () => '';
    const allFlake = failing.every((r) => looksLikeFlake(adapter(r)));
    return {
      status: allFlake ? 'red-flake-suspected' : 'red-code',
      considered,
      failing,
      blocking: true,
    };
  }

  if (pending) {
    return { status: 'pending', considered, failing: [], blocking: true };
  }

  return { status: 'green', considered, failing: [], blocking: true };
}

module.exports = { classifyCi, looksLikeFlake, FLAKE_SIGNATURES };
