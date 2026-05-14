// Fixtures for `gh run list --json status,conclusion,name,databaseId,headSha,event` output.
// Each fixture pins headSha so the classifier can filter stale runs.

const HEAD = 'abc123def4567890abc123def4567890abc12345';
const STALE = '0000000000000000000000000000000000000001';

module.exports = {
  HEAD,
  STALE,

  allGreen: [
    { databaseId: 1, name: 'lint',  status: 'completed', conclusion: 'success', headSha: HEAD, event: 'pull_request' },
    { databaseId: 2, name: 'test',  status: 'completed', conclusion: 'success', headSha: HEAD, event: 'pull_request' },
    { databaseId: 3, name: 'build', status: 'completed', conclusion: 'success', headSha: HEAD, event: 'pull_request' },
  ],

  mixedPassFail: [
    { databaseId: 1, name: 'lint',  status: 'completed', conclusion: 'success', headSha: HEAD, event: 'pull_request' },
    { databaseId: 2, name: 'test',  status: 'completed', conclusion: 'failure', headSha: HEAD, event: 'pull_request' },
  ],

  cancelledFlakeLog: [
    {
      databaseId: 7, name: 'integration', status: 'completed', conclusion: 'cancelled',
      headSha: HEAD, event: 'pull_request',
      // log signature embedded for log-pattern flake matcher in tests
      _failingLog: 'Error: connect ECONNRESET 10.0.0.5:443\nThe operation was canceled.',
    },
  ],

  inProgress: [
    { databaseId: 1, name: 'lint',  status: 'completed',   conclusion: 'success', headSha: HEAD, event: 'pull_request' },
    { databaseId: 2, name: 'test',  status: 'in_progress', conclusion: null,      headSha: HEAD, event: 'pull_request' },
    { databaseId: 3, name: 'build', status: 'queued',      conclusion: null,      headSha: HEAD, event: 'pull_request' },
  ],

  noWorkflows: [],

  allSkipped: [
    { databaseId: 1, name: 'optional-lint', status: 'completed', conclusion: 'skipped', headSha: HEAD, event: 'pull_request' },
  ],

  staleRuns: [
    // these are from a prior SHA and must be filtered out
    { databaseId: 99, name: 'test', status: 'completed', conclusion: 'failure', headSha: STALE, event: 'pull_request' },
    // current SHA, all green
    { databaseId: 1, name: 'test', status: 'completed', conclusion: 'success', headSha: HEAD, event: 'pull_request' },
  ],

  newerCancelledByPush: [
    // run for current head was cancelled because a newer push superseded it; but no newer runs exist for HEAD
    // we expect this to NOT be flake (cancelled with no infra log signature) - treated as red
    {
      databaseId: 5, name: 'test', status: 'completed', conclusion: 'cancelled',
      headSha: HEAD, event: 'pull_request',
      _failingLog: 'Run was canceled by user',
    },
  ],
};
