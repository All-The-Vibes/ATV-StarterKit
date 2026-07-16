// Unit tests for the LFG/SLFG run-state helper.
// Run with: node --test .github/hooks/scripts/tests/lfg-state.test.js
//
// The helper backs the resumability + artifact-by-reference features of the
// `lfg` and `slfg` skills. Tests use the real filesystem in a temp dir — no
// mocks — per the project's testing conventions.

const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

const S = require('../lfg-state');

function tmpRuns() {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'lfg-runs-'));
  return path.join(dir, '.atv', 'runs');
}

// ---------------------------------------------------------------------------
// slugify — pure
// ---------------------------------------------------------------------------

test('slugify lowercases and hyphenates', () => {
  assert.equal(S.slugify('Add Resume Support'), 'add-resume-support');
});

test('slugify collapses runs of non-alphanumerics and trims edges', () => {
  assert.equal(S.slugify('  Foo___Bar !! baz  '), 'foo-bar-baz');
});

test('slugify caps length to keep ids small', () => {
  const out = S.slugify('x'.repeat(200));
  assert.ok(out.length <= 50, `expected <=50, got ${out.length}`);
});

// ---------------------------------------------------------------------------
// run-id derivation — pure & deterministic
// ---------------------------------------------------------------------------

test('runIdFromPlan strips directory and -plan.md suffix', () => {
  assert.equal(
    S.runIdFromPlan('docs/plans/2026-06-01-001-feat-foo-plan.md'),
    '2026-06-01-001-feat-foo'
  );
});

test('runIdFromPlan strips a plain .md suffix when no -plan suffix present', () => {
  assert.equal(S.runIdFromPlan('docs/plans/some-notes.md'), 'some-notes');
});

test('provisionalRunId is deterministic for the same inputs', () => {
  const a = S.provisionalRunId({ feature: 'Add resume', repo: 'atv', branch: 'main' });
  const b = S.provisionalRunId({ feature: 'Add resume', repo: 'atv', branch: 'main' });
  assert.equal(a, b);
});

test('provisionalRunId differs when the branch differs', () => {
  const a = S.provisionalRunId({ feature: 'Add resume', repo: 'atv', branch: 'main' });
  const b = S.provisionalRunId({ feature: 'Add resume', repo: 'atv', branch: 'feature-x' });
  assert.notEqual(a, b);
});

test('provisionalRunId is filesystem-safe (no slashes or spaces)', () => {
  const id = S.provisionalRunId({ feature: 'A/B test: spaces', repo: 'x', branch: 'y' });
  assert.doesNotMatch(id, /[^a-z0-9-]/);
});

// ---------------------------------------------------------------------------
// init — creates immutable meta + phases dir, idempotent
// ---------------------------------------------------------------------------

test('init creates meta.json and a phases directory', () => {
  const runs = tmpRuns();
  const meta = S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo' });
  assert.equal(meta.run_id, 'r1');
  assert.equal(meta.skill, 'lfg');
  assert.equal(meta.feature, 'demo');
  assert.ok(meta.created_at, 'created_at should be set');
  assert.ok(fs.existsSync(path.join(runs, 'r1', 'meta.json')));
  assert.ok(fs.statSync(path.join(runs, 'r1', 'phases')).isDirectory());
});

test('init is idempotent — re-init does not overwrite meta (resume safety)', () => {
  const runs = tmpRuns();
  const first = S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo' });
  const second = S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'CHANGED' });
  assert.equal(second.created_at, first.created_at);
  assert.equal(second.feature, 'demo', 'feature must not be overwritten on resume');
});

// ---------------------------------------------------------------------------
// mode (auto | gated) — T8 auto-mode persistence
// ---------------------------------------------------------------------------

test('init defaults mode to gated when --auto is absent', () => {
  const runs = tmpRuns();
  const meta = S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo' });
  assert.equal(meta.mode, 'gated');
});

test('init records mode auto when auto:true is passed', () => {
  const runs = tmpRuns();
  const meta = S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo', auto: true });
  assert.equal(meta.mode, 'auto');
});

test('status surfaces the persisted mode so a resume continues in it', () => {
  const runs = tmpRuns();
  S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo', auto: true });
  const st = S.status(runs, 'r1');
  assert.equal(st.meta.mode, 'auto');
});

test('mode is not overwritten on re-init (resume keeps its mode)', () => {
  const runs = tmpRuns();
  S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo', auto: true });
  const second = S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo' });
  assert.equal(second.mode, 'auto', 'auto mode must survive a gated re-init on resume');
});

test('auto flag does not disturb the other init fields', () => {
  const runs = tmpRuns();
  const meta = S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo', plan: 'p.md', auto: true });
  assert.equal(meta.skill, 'lfg');
  assert.equal(meta.feature, 'demo');
  assert.equal(meta.plan_path, 'p.md');
  assert.equal(meta.mode, 'auto');
});

// ---------------------------------------------------------------------------
// bindPlan — two-stage identity (parent-only single writer)
// ---------------------------------------------------------------------------

test('bindPlan records the plan path on the run meta after planning', () => {
  const runs = tmpRuns();
  S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo' });
  const meta = S.bindPlan(runs, 'r1', 'docs/plans/2026-06-01-001-feat-demo-plan.md');
  assert.equal(meta.plan_path, 'docs/plans/2026-06-01-001-feat-demo-plan.md');
  const onDisk = JSON.parse(fs.readFileSync(path.join(runs, 'r1', 'meta.json'), 'utf8'));
  assert.equal(onDisk.plan_path, 'docs/plans/2026-06-01-001-feat-demo-plan.md');
});

// ---------------------------------------------------------------------------
// markDone / isDone — atomic per-phase sentinels (SLFG concurrency safety)
// ---------------------------------------------------------------------------

test('isDone is false before a phase completes, true after markDone', () => {
  const runs = tmpRuns();
  S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo' });
  assert.equal(S.isDone(runs, 'r1', 'ce-work'), false);
  S.markDone(runs, 'r1', 'ce-work', { artifact: null });
  assert.equal(S.isDone(runs, 'r1', 'ce-work'), true);
});

test('markDone records the artifact path by reference', () => {
  const runs = tmpRuns();
  S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo' });
  S.markDone(runs, 'r1', 'ce-review', {
    artifact: '.context/compound-engineering/ce-review/abc/report.md',
  });
  const sentinel = JSON.parse(
    fs.readFileSync(path.join(runs, 'r1', 'phases', 'ce-review.done.json'), 'utf8')
  );
  assert.equal(sentinel.phase, 'ce-review');
  assert.equal(sentinel.status, 'done');
  assert.equal(sentinel.artifact, '.context/compound-engineering/ce-review/abc/report.md');
  assert.ok(sentinel.ended_at, 'ended_at should be set');
});

test('markDone rejects phase names containing path traversal', () => {
  const runs = tmpRuns();
  S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo' });
  assert.throws(() => S.markDone(runs, 'r1', '../escape', {}), /invalid phase/i);
});

// ---------------------------------------------------------------------------
// status — compact view for the orchestrator (paths, not content)
// ---------------------------------------------------------------------------

test('status returns done phases with their artifact paths', () => {
  const runs = tmpRuns();
  S.init(runs, { runId: 'r1', skill: 'lfg', feature: 'demo' });
  S.markDone(runs, 'r1', 'ce-plan', { artifact: 'docs/plans/p.md' });
  S.markDone(runs, 'r1', 'ce-work', { artifact: null });
  const st = S.status(runs, 'r1');
  assert.equal(st.run_id, 'r1');
  const byPhase = Object.fromEntries(st.phases.map((p) => [p.phase, p]));
  assert.equal(byPhase['ce-plan'].status, 'done');
  assert.equal(byPhase['ce-plan'].artifact, 'docs/plans/p.md');
  assert.equal(byPhase['ce-work'].status, 'done');
});

test('status on a non-existent run returns an empty phase list, not an error', () => {
  const runs = tmpRuns();
  const st = S.status(runs, 'ghost');
  assert.equal(st.run_id, 'ghost');
  assert.deepEqual(st.phases, []);
  assert.equal(st.meta, null);
});

// ---------------------------------------------------------------------------
// parseRunToken — shared `run:<id>` argument convention for sub-skills
// ---------------------------------------------------------------------------

test('parseRunToken extracts run:<id> and returns remaining args', () => {
  const { runId, rest } = S.parseRunToken('mode:autofix run:2026-06-01-001-feat-foo plan:docs/p.md');
  assert.equal(runId, '2026-06-01-001-feat-foo');
  assert.equal(rest, 'mode:autofix plan:docs/p.md');
});

test('parseRunToken returns null runId when no token present', () => {
  const { runId, rest } = S.parseRunToken('mode:autofix plan:docs/p.md');
  assert.equal(runId, null);
  assert.equal(rest, 'mode:autofix plan:docs/p.md');
});
