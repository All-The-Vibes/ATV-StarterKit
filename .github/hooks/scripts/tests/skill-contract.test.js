// Contract tests for the LFG/SLFG resumability + artifact-by-reference wiring.
// Run with: node --test .github/hooks/scripts/tests/skill-contract.test.js
//
// These assert that the *skills themselves* adopt the new features — not just
// that the helper exists. They cover both the dogfood install (.github/skills)
// and the shipped templates (pkg/scaffold/templates/skills) so the two copies
// cannot silently drift on the resume protocol.

const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const ROOT = path.resolve(__dirname, '..', '..', '..', '..');

function read(rel) {
  return fs.readFileSync(path.join(ROOT, rel), 'utf8');
}

const LFG_COPIES = [
  '.github/skills/lfg/SKILL.md',
  'pkg/scaffold/templates/skills/lfg/SKILL.md',
];
const SLFG_COPIES = [
  '.github/skills/slfg/SKILL.md',
  'pkg/scaffold/templates/skills/slfg/SKILL.md',
];
const HELPER_COPIES = [
  '.github/hooks/scripts/lfg-state.js',
  'pkg/scaffold/templates/hooks/scripts/lfg-state.js',
];

// ---------------------------------------------------------------------------
// helper ships in both the dogfood and template locations
// ---------------------------------------------------------------------------

for (const rel of HELPER_COPIES) {
  test(`helper exists and exports run-state API: ${rel}`, () => {
    const src = read(rel);
    for (const fn of ['markDone', 'isDone', 'status', 'runIdFromPlan', 'parseRunToken']) {
      assert.ok(src.includes(fn), `${rel} should define ${fn}`);
    }
  });
}

// ---------------------------------------------------------------------------
// LFG wires in the resume protocol
// ---------------------------------------------------------------------------

for (const rel of LFG_COPIES) {
  test(`LFG references the run-state helper: ${rel}`, () => {
    assert.match(read(rel), /lfg-state\.js/);
  });

  test(`LFG documents resume / skip-completed-phases: ${rel}`, () => {
    assert.match(read(rel), /resume/i);
    assert.match(read(rel), /status|is-done/i);
  });

  test(`LFG threads a run id through sub-skills: ${rel}`, () => {
    assert.match(read(rel), /run:</);
  });

  test(`LFG marks phases done by reference: ${rel}`, () => {
    assert.match(read(rel), /done .*--artifact|--artifact/);
  });
}

// ---------------------------------------------------------------------------
// SLFG wires in the resume protocol AND parent-only writes for the parallel
// phase (the concurrency-safety requirement)
// ---------------------------------------------------------------------------

for (const rel of SLFG_COPIES) {
  test(`SLFG references the run-state helper: ${rel}`, () => {
    assert.match(read(rel), /lfg-state\.js/);
  });

  test(`SLFG documents parent-only state writes for the parallel phase: ${rel}`, () => {
    assert.match(read(rel), /parent/i);
    assert.match(read(rel), /run:</);
  });
}

// ---------------------------------------------------------------------------
// Sub-skills must recognize-and-strip `run:<id>` before it is threaded to
// them (otherwise an unknown token is mis-read as a PR/branch arg), and
// ce-work must expose an orchestrated mode so resume does not duplicate work.
// ---------------------------------------------------------------------------

const SUBSKILL_COPIES = [
  '.github/skills/ce-plan/SKILL.md',
  '.github/skills/ce-work/SKILL.md',
  '.github/skills/ce-review/SKILL.md',
  'pkg/scaffold/templates/skills/ce-plan/SKILL.md',
  'pkg/scaffold/templates/skills/ce-work/SKILL.md',
  'pkg/scaffold/templates/skills/ce-review/SKILL.md',
];

for (const rel of SUBSKILL_COPIES) {
  test(`sub-skill recognizes the run: orchestration token: ${rel}`, () => {
    assert.match(read(rel), /run:<[^>]*id[^>]*>/i);
  });
}

for (const rel of SUBSKILL_COPIES.filter((r) => r.includes('ce-work'))) {
  test(`ce-work exposes mode:orchestrated for resume safety: ${rel}`, () => {
    assert.match(read(rel), /mode:orchestrated/);
  });
}
