// Unit tests for the CI classifier (pure function over `gh run list` JSON).
// Run with: node --test tests/ci-classifier.test.js

const { test } = require('node:test');
const assert = require('node:assert/strict');

const { classifyCi } = require('../lib/ci-classifier');
const F = require('./fixtures/ci-runs');

test('all-success runs at HEAD => green', () => {
  const r = classifyCi({ runs: F.allGreen, headSha: F.HEAD });
  assert.equal(r.status, 'green');
  assert.equal(r.failing.length, 0);
});

test('mixed pass+fail at HEAD => red-code', () => {
  const r = classifyCi({ runs: F.mixedPassFail, headSha: F.HEAD });
  assert.equal(r.status, 'red-code');
  assert.deepEqual(r.failing.map(j => j.name), ['test']);
});

test('cancelled with infra log signature => red-flake-suspected', () => {
  // Classifier is pure over JSON; pass `logFor` adapter to expose log text per-run.
  const logFor = (run) => run._failingLog || '';
  const r = classifyCi({ runs: F.cancelledFlakeLog, headSha: F.HEAD, logFor });
  assert.equal(r.status, 'red-flake-suspected');
  assert.equal(r.failing.length, 1);
});

test('cancelled WITHOUT infra signature => red-code (not flake)', () => {
  const logFor = (run) => run._failingLog || '';
  const r = classifyCi({ runs: F.newerCancelledByPush, headSha: F.HEAD, logFor });
  assert.equal(r.status, 'red-code');
});

test('in_progress / queued with no failures yet => pending', () => {
  const r = classifyCi({ runs: F.inProgress, headSha: F.HEAD });
  assert.equal(r.status, 'pending');
});

test('no workflows at all => none', () => {
  const r = classifyCi({ runs: F.noWorkflows, headSha: F.HEAD });
  assert.equal(r.status, 'none');
});

test('runs exist but all stale (only older SHAs) => stale-only, blocking', () => {
  // Common state right after a fresh push: runs exist for the branch but every one
  // belongs to a previous SHA. Must NOT collapse to 'none' (which the verdict layer
  // treats as a non-gate) — the verdict layer would otherwise APPROVE a SHA CI has
  // never seen.
  const onlyStale = [
    { databaseId: 10, name: 'lint',  status: 'completed', conclusion: 'success', headSha: 'older-sha', event: 'pull_request' },
    { databaseId: 11, name: 'test',  status: 'completed', conclusion: 'success', headSha: 'older-sha', event: 'pull_request' },
  ];
  const r = classifyCi({ runs: onlyStale, headSha: F.HEAD });
  assert.equal(r.status, 'stale-only');
  assert.equal(r.blocking, true);
  assert.equal(r.considered.length, 0);
});

test('truly empty runs array => none (no Actions configured)', () => {
  const r = classifyCi({ runs: [], headSha: F.HEAD });
  assert.equal(r.status, 'none');
});

test('all-skipped => green (no gating signal)', () => {
  const r = classifyCi({ runs: F.allSkipped, headSha: F.HEAD });
  assert.equal(r.status, 'green');
});

test('stale runs from older SHAs are excluded', () => {
  const r = classifyCi({ runs: F.staleRuns, headSha: F.HEAD });
  // STALE failure should be filtered out; only HEAD success remains => green.
  assert.equal(r.status, 'green');
  assert.equal(r.considered.length, 1);
  assert.equal(r.considered[0].headSha, F.HEAD);
});

test('gh API failure (null runs) => unknown, does not block', () => {
  const r = classifyCi({ runs: null, headSha: F.HEAD });
  assert.equal(r.status, 'unknown');
  assert.equal(r.blocking, false);
});

test('classifier exposes failing-job metadata for downstream subroutine', () => {
  const r = classifyCi({ runs: F.mixedPassFail, headSha: F.HEAD });
  assert.equal(r.failing[0].name, 'test');
  assert.equal(r.failing[0].databaseId, 2);
});

test('flake heuristic recognises common infra signatures', () => {
  const cases = [
    'Error: connect ECONNRESET',
    'failed to pull image: net/http: TLS handshake timeout',
    'The runner has received a shutdown signal',
    'The operation was canceled.',
    'Error response from daemon: pull access denied',
  ];
  for (const log of cases) {
    const runs = [{
      databaseId: 1, name: 'job', status: 'completed', conclusion: 'failure',
      headSha: F.HEAD, event: 'pull_request', _failingLog: log,
    }];
    const r = classifyCi({ runs, headSha: F.HEAD, logFor: (x) => x._failingLog });
    assert.equal(r.status, 'red-flake-suspected', `expected flake for: ${log}`);
  }
});

test('unknown statuses (e.g. action_required) do not crash classifier', () => {
  const runs = [
    { databaseId: 1, name: 'job', status: 'action_required', conclusion: null, headSha: F.HEAD, event: 'pull_request' },
  ];
  const r = classifyCi({ runs, headSha: F.HEAD });
  // action_required is neither pending nor failed; classifier should treat as pending (incomplete).
  assert.equal(r.status, 'pending');
});
