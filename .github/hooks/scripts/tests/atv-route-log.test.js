"use strict";

// Tests for atv-route-log.js — the /atv router's route telemetry writer.
// Real temp-dir fs via ATV_CONFIG_HOME; no mocks. Node built-in test runner.

const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { execFileSync } = require("node:child_process");

const R = require("../atv-route-log");

const SCRIPT = path.join(__dirname, "..", "atv-route-log.js");

function tmpHome() {
  return fs.mkdtempSync(path.join(os.tmpdir(), "atv-route-"));
}

// --- buildRecord (T9.2) ----------------------------------------------------

test("buildRecord returns an OTel-shaped object with injected timestamp", () => {
  const now = "2026-07-16T06:12:00.000Z";
  const rec = R.buildRecord(
    { intentCategory: "code-review", routedTo: "ce-review", outcome: "invoked" },
    now
  );
  assert.equal(rec.name, "atv.route");
  assert.equal(rec.timestamp, now);
  assert.deepEqual(Object.keys(rec.attributes).sort(), [
    "intent_category",
    "outcome",
    "routed_to",
  ]);
  assert.equal(rec.attributes.intent_category, "code-review");
  assert.equal(rec.attributes.routed_to, "ce-review");
  assert.equal(rec.attributes.outcome, "invoked");
});

// --- sanitizeToken + PII cap (T9.3) ----------------------------------------

test("sanitizeToken caps at 64 chars, strips newlines, coerces non-strings", () => {
  assert.equal(R.sanitizeToken("code-review"), "code-review");
  const long = "a".repeat(200);
  assert.equal(R.sanitizeToken(long).length, 64);
  assert.equal(R.sanitizeToken("line1\nline2"), "line1line2");
  assert.equal(R.sanitizeToken(42), "42");
  assert.equal(R.sanitizeToken(undefined), "");
  assert.equal(R.sanitizeToken(null), "");
});

test("PII guarantee: a request-sentence intentCategory cannot pass through intact", () => {
  const rawRequest =
    "please review my authentication changes because I think the login flow is broken and leaks the session token to the client";
  const rec = R.buildRecord(
    { intentCategory: rawRequest, routedTo: "ce-review", outcome: "invoked" },
    "2026-07-16T06:12:00.000Z"
  );
  assert.ok(
    rec.attributes.intent_category.length <= 64,
    "intent_category must be truncated to <= 64 chars"
  );
  assert.ok(
    !rec.attributes.intent_category.includes("\n"),
    "intent_category must not contain newlines"
  );
  // The full raw sentence must NOT survive.
  assert.ok(
    !rec.attributes.intent_category.includes("leaks the session token"),
    "raw request text must not be recorded intact"
  );
});

// --- normalizeOutcome (T9.4) -----------------------------------------------

test("normalizeOutcome passes valid values, maps invalid to invoked", () => {
  for (const ok of ["invoked", "emitted", "suggested", "no-match", "control"]) {
    assert.equal(R.normalizeOutcome(ok), ok);
  }
  assert.equal(R.normalizeOutcome("garbage"), "invoked");
  assert.equal(R.normalizeOutcome(""), "invoked");
  assert.equal(R.normalizeOutcome(undefined), "invoked");
});

// --- logPath + appendRecord (T9.5) -----------------------------------------

test("logPath honors ATV_CONFIG_HOME, falls back to ~/.atv when unset", () => {
  const home = tmpHome();
  try {
    const prev = process.env.ATV_CONFIG_HOME;
    process.env.ATV_CONFIG_HOME = home;
    assert.equal(
      R.logPath(),
      path.join(home, "analytics", "routes.jsonl")
    );
    delete process.env.ATV_CONFIG_HOME;
    assert.equal(
      R.logPath(),
      path.join(os.homedir(), ".atv", "analytics", "routes.jsonl")
    );
    if (prev !== undefined) process.env.ATV_CONFIG_HOME = prev;
  } finally {
    fs.rmSync(home, { recursive: true, force: true });
  }
});

test("appendRecord creates the dir and writes one parseable line", () => {
  const home = tmpHome();
  try {
    const file = path.join(home, "analytics", "routes.jsonl");
    const rec = R.buildRecord(
      { intentCategory: "build", routedTo: "emit:lfg", outcome: "emitted" },
      "2026-07-16T06:12:00.000Z"
    );
    const ok = R.appendRecord(rec, file);
    assert.equal(ok, true);
    const lines = fs.readFileSync(file, "utf8").trim().split("\n");
    assert.equal(lines.length, 1);
    const parsed = JSON.parse(lines[0]);
    assert.equal(parsed.attributes.routed_to, "emit:lfg");
  } finally {
    fs.rmSync(home, { recursive: true, force: true });
  }
});

test("appendRecord appends a second line without clobbering the first", () => {
  const home = tmpHome();
  try {
    const file = path.join(home, "analytics", "routes.jsonl");
    R.appendRecord(
      R.buildRecord({ intentCategory: "a", routedTo: "ce-plan", outcome: "invoked" }, "t1"),
      file
    );
    R.appendRecord(
      R.buildRecord({ intentCategory: "b", routedTo: "ce-review", outcome: "invoked" }, "t2"),
      file
    );
    const lines = fs.readFileSync(file, "utf8").trim().split("\n");
    assert.equal(lines.length, 2);
    assert.equal(JSON.parse(lines[0]).attributes.routed_to, "ce-plan");
    assert.equal(JSON.parse(lines[1]).attributes.routed_to, "ce-review");
  } finally {
    fs.rmSync(home, { recursive: true, force: true });
  }
});

test("appendRecord never throws on an unwritable path (returns false)", () => {
  // Point the file at a path whose parent is a FILE, so mkdir fails.
  const home = tmpHome();
  try {
    const blocker = path.join(home, "blocker");
    fs.writeFileSync(blocker, "x");
    const file = path.join(blocker, "analytics", "routes.jsonl"); // parent is a file
    let result;
    assert.doesNotThrow(() => {
      result = R.appendRecord(
        R.buildRecord({ intentCategory: "x", routedTo: "y", outcome: "invoked" }, "t"),
        file
      );
    });
    assert.equal(result, false);
  } finally {
    fs.rmSync(home, { recursive: true, force: true });
  }
});

// --- CLI (T9.6) ------------------------------------------------------------

function runCli(args, home) {
  return execFileSync("node", [SCRIPT, ...args], {
    env: { ...process.env, ATV_CONFIG_HOME: home },
    encoding: "utf8",
  });
}

test("CLI: full call writes exactly one line with correct attributes", () => {
  const home = tmpHome();
  try {
    runCli(
      ["--intent", "security", "--routed-to", "atv-security", "--outcome", "invoked"],
      home
    );
    const file = path.join(home, "analytics", "routes.jsonl");
    const lines = fs.readFileSync(file, "utf8").trim().split("\n");
    assert.equal(lines.length, 1);
    const rec = JSON.parse(lines[0]);
    assert.equal(rec.name, "atv.route");
    assert.equal(rec.attributes.intent_category, "security");
    assert.equal(rec.attributes.routed_to, "atv-security");
    assert.equal(rec.attributes.outcome, "invoked");
    assert.ok(typeof rec.timestamp === "string" && rec.timestamp.length > 0);
  } finally {
    fs.rmSync(home, { recursive: true, force: true });
  }
});

test("CLI: missing --routed-to writes nothing and exits 0", () => {
  const home = tmpHome();
  try {
    // Should not throw (exit 0) and should not create the file.
    assert.doesNotThrow(() => runCli(["--intent", "security"], home));
    const file = path.join(home, "analytics", "routes.jsonl");
    assert.equal(fs.existsSync(file), false);
  } finally {
    fs.rmSync(home, { recursive: true, force: true });
  }
});

test("CLI: default outcome is invoked when --outcome omitted", () => {
  const home = tmpHome();
  try {
    runCli(["--intent", "planning", "--routed-to", "ce-plan"], home);
    const file = path.join(home, "analytics", "routes.jsonl");
    const rec = JSON.parse(fs.readFileSync(file, "utf8").trim());
    assert.equal(rec.attributes.outcome, "invoked");
  } finally {
    fs.rmSync(home, { recursive: true, force: true });
  }
});

// --- template parity (T9.7) ------------------------------------------------

test("dogfood and template copies are byte-identical", () => {
  const dogfood = fs.readFileSync(SCRIPT);
  const template = fs.readFileSync(
    path.join(
      __dirname,
      "..", "..", "..", "..",
      "pkg", "scaffold", "templates", "hooks", "scripts", "atv-route-log.js"
    )
  );
  assert.ok(dogfood.equals(template), "atv-route-log.js copies must be byte-identical");
});
