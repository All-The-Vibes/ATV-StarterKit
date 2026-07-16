// Contract tests for the /atv router fixes (PR #62 review findings).
// Run with: node --test .github/hooks/scripts/tests/atv-router-contract.test.js
//
// These lock in the four doc/skill-level review findings so they cannot
// regress across any mirror copy:
//   Fix1 — every atv-config.js invocation in the skill is runnable
//          (`node .github/hooks/scripts/atv-config.js ...`), never a bare
//          `atv-config.js` (not on PATH → command not found).
//   Fix2 — the dogfood .github/skills/atv/llms.txt is byte-identical to the
//          template copy (no stale catalog missing /investigate).
//   Fix4 — the skill documents graceful degradation when the hook scripts are
//          absent (marketplace plugins that ship no hooks).
//   Fix5 — no copy claims raw request text is "structurally impossible to
//          log"; the honest, bounded framing is used instead.

const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const ROOT = path.resolve(__dirname, '..', '..', '..', '..');

function read(rel) {
  return fs.readFileSync(path.join(ROOT, rel), 'utf8');
}
function exists(rel) {
  return fs.existsSync(path.join(ROOT, rel));
}

// Every mirror copy of the /atv router skill.
const ATV_SKILL_COPIES = [
  '.github/skills/atv/SKILL.md',
  'pkg/scaffold/templates/skills/atv/SKILL.md',
  'plugins/atv-everything/skills/atv/SKILL.md',
  'plugins/atv-pack-shipping/skills/atv/SKILL.md',
  'plugins/atv-skill-atv/skills/atv/SKILL.md',
];

// Docs/copy that describe the telemetry privacy posture.
const PII_CLAIM_COPIES = [
  'README.md',
  'CHANGELOG.md',
  '.github/skills/atv/SKILL.md',
  'pkg/scaffold/templates/skills/atv/SKILL.md',
  '.github/hooks/scripts/atv-route-log.js',
  'pkg/scaffold/templates/hooks/scripts/atv-route-log.js',
  'docs/plans/2026-07-16-001-feat-atv-route-telemetry-plan.md',
];

// ---------------------------------------------------------------------------
// Fix1: no bare `atv-config.js` invocation — it is not on PATH.
// Any line that invokes the shim must call it via `node <path>/atv-config.js`.
// ---------------------------------------------------------------------------

for (const rel of ATV_SKILL_COPIES) {
  test(`Fix1: every atv-config.js invocation is node-prefixed: ${rel}`, () => {
    const src = read(rel);
    const lines = src.split('\n');
    lines.forEach((line, i) => {
      // Only inspect lines that actually *invoke* the shim: a `set`/`get`
      // subcommand follows the filename. A bare prose mention (e.g. listing the
      // script names in the degradation note) is not an invocation.
      const invokes = /atv-config\.js\s+(set|get)\b/;
      if (!invokes.test(line)) return;
      const runnable = /node\s+\S*hooks\/scripts\/atv-config\.js\s+(set|get)\b/;
      if (!runnable.test(line)) {
        assert.fail(
          `${rel}:${i + 1} invokes bare \`atv-config.js\` (not on PATH). ` +
            `Use \`node .github/hooks/scripts/atv-config.js ...\`.\n  ${line.trim()}`,
        );
      }
    });
  });
}

// ---------------------------------------------------------------------------
// Fix2: dogfood llms.txt catalog is byte-identical to the template copy.
// The template copy is the generator output guarded by
// TestRoutingCatalog_CommittedLLMsTxtIsFresh; the dogfood copy had no guard
// and silently drifted (missing /investigate). Lock them together.
// ---------------------------------------------------------------------------

test('Fix2: dogfood llms.txt matches template llms.txt (no drift)', () => {
  const dogfood = read('.github/skills/atv/llms.txt');
  const template = read('pkg/scaffold/templates/skills/atv/llms.txt');
  assert.equal(
    dogfood,
    template,
    '.github/skills/atv/llms.txt has drifted from the template catalog. ' +
      'Regenerate the dogfood copy so both include the same skills (e.g. /investigate).',
  );
});

test('Fix2: llms.txt catalog includes /investigate', () => {
  for (const rel of [
    '.github/skills/atv/llms.txt',
    'pkg/scaffold/templates/skills/atv/llms.txt',
  ]) {
    assert.match(
      read(rel),
      /investigate/,
      `${rel} should list the /investigate route (bug intent target)`,
    );
  }
});

// ---------------------------------------------------------------------------
// Fix4: the skill documents graceful degradation when hook scripts are absent
// (atv-everything / atv-pack-shipping / atv-skill-atv ship no hooks/ tree).
// ---------------------------------------------------------------------------

for (const rel of ATV_SKILL_COPIES) {
  test(`Fix4: documents behavior when hook scripts are absent: ${rel}`, () => {
    const src = read(rel);
    // (a) acknowledge the hook scripts can be absent, and (b) state the route
    // still proceeds (best-effort / never blocks).
    assert.match(
      src,
      /hook scripts are absent|do not exist|not present|missing/i,
      `${rel} should acknowledge the hook scripts may be absent`,
    );
    assert.match(
      src,
      /never (block|error|stop)|best-effort|skip (logging|the command)|default to `?true`?/i,
      `${rel} should state that /atv still routes when the hook scripts are absent`,
    );
  });

  test(`Fix4: hook-script invocations are conditional at point of use: ${rel}`, () => {
    const src = read(rel);
    const lines = src.split('\n');
    // For each executable invocation of a hook script, require guard language
    // (exist/absent/present/skip/if) within a small window around the call —
    // so a guard relocated to a distant trailing note does NOT satisfy this.
    const WINDOW = 8;
    const guard = /absent|not present|if present|exists?|missing|skip/i;
    const invocationRe = /node\s+\S*hooks\/scripts\/atv-(config|route-log)\.js/;
    lines.forEach((line, i) => {
      if (!invocationRe.test(line)) return;
      const lo = Math.max(0, i - WINDOW);
      const hi = Math.min(lines.length, i + WINDOW + 1);
      const window = lines.slice(lo, hi).join('\n');
      assert.match(
        window,
        guard,
        `${rel}:${i + 1} invokes a hook script with no existence/skip guard ` +
          `within ${WINDOW} lines — a hookless install would execute a missing ` +
          `script. Guard it at the point of use.\n  ${line.trim()}`,
      );
    });
  });
}

// ---------------------------------------------------------------------------
// Fix3: docs/atv-router.md exists (referenced by routing-fixtures.txt Layer 2).
// ---------------------------------------------------------------------------

test('Fix3: docs/atv-router.md referenced by routing-fixtures exists', () => {
  const fixtures = read('pkg/plugingen/testdata/routing-fixtures.txt');
  if (fixtures.includes('docs/atv-router.md')) {
    assert.ok(
      exists('docs/atv-router.md'),
      'routing-fixtures.txt points reviewers at docs/atv-router.md but the file is missing',
    );
  }
});

test('Fix3: docs/atv-router.md documents the Layer-2 live smoke procedure', () => {
  if (!exists('docs/atv-router.md')) return;
  const src = read('docs/atv-router.md');
  assert.match(src, /layer 2|live smoke|live-model/i,
    'docs/atv-router.md should describe the Layer-2 live-model smoke procedure');
  assert.match(src, /routing-fixtures\.txt/,
    'docs/atv-router.md should point at the fixtures file');
});

// Fix3 (consistency): the router routes to ONE best skill — no ranked "top-2"
// acceptance rule may survive anywhere (doc, fixtures header, or plan), or the
// smoke procedure contradicts real behavior.
const TOP2_SCAN = [
  'docs/atv-router.md',
  'pkg/plugingen/testdata/routing-fixtures.txt',
  'docs/research/gstack-router-atv-plan.md',
];
for (const rel of TOP2_SCAN) {
  test(`Fix3: no stale "top-2" acceptance rule: ${rel}`, () => {
    if (!exists(rel)) return;
    assert.doesNotMatch(
      read(rel),
      /top-2|top 2 |first or second choice/i,
      `${rel} still describes a "top-2" match, but the router exposes one best ` +
        `route (no ranked list). Use single-route stable-choice adjudication.`,
    );
  });
}

// ---------------------------------------------------------------------------
// Fix5: no "structurally impossible to log/pass" overclaim anywhere.
// The bounded, honest framing is required instead.
// ---------------------------------------------------------------------------

for (const rel of PII_CLAIM_COPIES) {
  test(`Fix5: no "structurally impossible" PII overclaim: ${rel}`, () => {
    if (!exists(rel)) return; // plan doc may be absent in some checkouts
    const src = read(rel);
    assert.doesNotMatch(
      src,
      /structurally impossible/i,
      `${rel} still claims raw text is "structurally impossible" to log/pass. ` +
        `Use bounded framing: no free-form field + 64-char cap, so raw request ` +
        `text is not recorded, but --intent/--routed-to are length-bounded tokens.`,
    );
  });
}

// Fix5 (positive): the user-facing privacy copy must carry the honest bound —
// naming that the logged tokens are caller-supplied and only length-capped, not
// merely that raw text is "never recorded". Guards against a partial fix that
// deletes the old phrase but leaves an unqualified guarantee.
const FIX5_BOUND_COPIES = [
  'README.md',
  'CHANGELOG.md',
  '.github/skills/atv/SKILL.md',
];
for (const rel of FIX5_BOUND_COPIES) {
  test(`Fix5: privacy copy states the caller-supplied / bounded caveat: ${rel}`, () => {
    const src = read(rel);
    assert.match(
      src,
      /caller-supplied|a bound, not|not a blanket|contract the bound relies on/i,
      `${rel} must qualify the privacy claim (caller-supplied, length-bounded tokens), ` +
        `not just assert raw text is "never recorded"`,
    );
  });
}

// The two atv-route-log.js copies and all five SKILL.md copies must stay
// byte-identical — a partial fix that touches one mirror is a defect.
test('mirror byte-identity: all 5 atv/SKILL.md copies match', () => {
  const canonical = read(ATV_SKILL_COPIES[0]);
  for (const rel of ATV_SKILL_COPIES.slice(1)) {
    assert.equal(read(rel), canonical, `${rel} has drifted from ${ATV_SKILL_COPIES[0]}`);
  }
});

test('mirror byte-identity: both atv-route-log.js copies match', () => {
  const a = read('.github/hooks/scripts/atv-route-log.js');
  const b = read('pkg/scaffold/templates/hooks/scripts/atv-route-log.js');
  assert.equal(a, b, 'atv-route-log.js dogfood and template copies have drifted');
});
