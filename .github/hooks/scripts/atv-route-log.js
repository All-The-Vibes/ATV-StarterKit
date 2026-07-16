#!/usr/bin/env node
// ATV route telemetry writer — records one OTel-shaped line per /atv routing
// decision to ~/.atv/analytics/routes.jsonl.
//
// Fixed schema by design (plan D1): the CLI accepts ONLY --intent, --routed-to,
// and --outcome. There is no free-form field, so the user's raw request text is
// structurally impossible to pass. Tokens are additionally length-capped and
// newline-stripped as defense in depth. Best-effort: never throws, never blocks
// a route.
//
// Mirrors the atv-config.js / observe.js hook-helper conventions: zero deps,
// pure exported functions + require.main CLI guard, ATV_CONFIG_HOME test
// override.

"use strict";

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

const EVENT_NAME = "atv.route";
const MAX_TOKEN_LEN = 64;
const VALID_OUTCOMES = ["invoked", "emitted", "suggested", "no-match", "control"];
const DEFAULT_OUTCOME = "invoked";

// --- pure helpers ----------------------------------------------------------

// Coerce to string, trim, strip newlines, cap length. This is the structural
// PII defense: a classifier token is short; a request sentence is not, so a cap
// + newline strip prevents raw text from being recorded intact or breaking the
// one-record-per-line invariant.
function sanitizeToken(value) {
  if (value === undefined || value === null) return "";
  let s = String(value).replace(/[\r\n]+/g, "").trim();
  if (s.length > MAX_TOKEN_LEN) s = s.slice(0, MAX_TOKEN_LEN);
  return s;
}

// Map an outcome to the fixed enum, defaulting to "invoked".
function normalizeOutcome(outcome) {
  return VALID_OUTCOMES.includes(outcome) ? outcome : DEFAULT_OUTCOME;
}

// Build the OTel-shaped record. Timestamp is injected so the function stays
// pure and testable.
function buildRecord({ intentCategory, routedTo, outcome }, nowIso) {
  return {
    name: EVENT_NAME,
    timestamp: nowIso,
    attributes: {
      intent_category: sanitizeToken(intentCategory),
      routed_to: sanitizeToken(routedTo),
      outcome: normalizeOutcome(outcome),
    },
  };
}

// --- fs-backed ops ---------------------------------------------------------

function logPath() {
  const home = process.env.ATV_CONFIG_HOME;
  if (home) return path.join(home, "analytics", "routes.jsonl");
  return path.join(os.homedir(), ".atv", "analytics", "routes.jsonl");
}

// Append one JSON line. Best-effort: returns true on success, false on any
// failure (unwritable dir, etc.) — NEVER throws, so telemetry can never block a
// route.
function appendRecord(record, filePath) {
  try {
    fs.mkdirSync(path.dirname(filePath), { recursive: true, mode: 0o700 });
    fs.appendFileSync(filePath, JSON.stringify(record) + "\n", { mode: 0o600 });
    return true;
  } catch (_) {
    return false;
  }
}

// --- CLI -------------------------------------------------------------------

function parseArgs(argv) {
  const out = {};
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === "--intent") out.intent = argv[++i];
    else if (a === "--routed-to") out.routedTo = argv[++i];
    else if (a === "--outcome") out.outcome = argv[++i];
  }
  return out;
}

function main(argv) {
  const args = parseArgs(argv);
  // Both classifier tokens are required. A malformed call writes nothing and
  // exits 0 — the router must never be blocked by a telemetry mistake.
  if (!args.intent || !args.routedTo) return;
  const record = buildRecord(
    {
      intentCategory: args.intent,
      routedTo: args.routedTo,
      outcome: args.outcome,
    },
    new Date().toISOString()
  );
  appendRecord(record, logPath());
}

if (require.main === module) {
  try {
    main(process.argv.slice(2));
  } catch (_) {
    // Best-effort: never surface a telemetry failure to the caller.
  }
  process.exit(0);
}

module.exports = {
  EVENT_NAME,
  MAX_TOKEN_LEN,
  VALID_OUTCOMES,
  sanitizeToken,
  normalizeOutcome,
  buildRecord,
  logPath,
  appendRecord,
};
