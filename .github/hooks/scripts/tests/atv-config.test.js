"use strict";

const { test } = require("node:test");
const assert = require("node:assert");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { execFileSync } = require("node:child_process");

const SCRIPT = path.join(__dirname, "..", "atv-config.js");
const MOD = require("../atv-config.js");

function tmpHome() {
  return fs.mkdtempSync(path.join(os.tmpdir(), "atv-config-test-"));
}

function cli(args, home) {
  return execFileSync(process.execPath, [SCRIPT, ...args], {
    env: Object.assign({}, process.env, { ATV_CONFIG_HOME: home }),
    encoding: "utf8",
  });
}

test("1. get returns DEFAULTS value when file absent", () => {
  const home = tmpHome();
  const cfgPath = path.join(home, "config.json");
  assert.strictEqual(fs.existsSync(cfgPath), false);
  assert.strictEqual(MOD.getValue(MOD.readConfig(cfgPath), "proactive"), true);
});

test("2. set persists, then get reads it back", () => {
  const home = tmpHome();
  const cfgPath = path.join(home, "config.json");
  MOD.setValue(cfgPath, "proactive", "false");
  assert.strictEqual(MOD.getValue(MOD.readConfig(cfgPath), "proactive"), false);
});

test("3. set creates ~/.atv dir when absent", () => {
  const home = path.join(tmpHome(), ".atv");
  const cfgPath = path.join(home, "config.json");
  assert.strictEqual(fs.existsSync(home), false);
  MOD.setValue(cfgPath, "proactive", "false");
  assert.strictEqual(fs.existsSync(cfgPath), true);
});

test("4. get unknown key returns '' and exit 0 (no throw)", () => {
  const home = tmpHome();
  const out = cli(["get", "nope"], home);
  assert.strictEqual(out, "\n");
});

test("5. malformed JSON -> readConfig {} and get falls back to default", () => {
  const home = tmpHome();
  const cfgPath = path.join(home, "config.json");
  fs.writeFileSync(cfgPath, "{ this is not json ");
  assert.deepStrictEqual(MOD.readConfig(cfgPath), {});
  assert.strictEqual(MOD.getValue(MOD.readConfig(cfgPath), "proactive"), true);
  // CLI does not crash either
  const out = cli(["get", "proactive"], home);
  assert.strictEqual(out, "true\n");
});

test("6. set against existing file preserves other keys", () => {
  const home = tmpHome();
  const cfgPath = path.join(home, "config.json");
  MOD.setValue(cfgPath, "a", "1");
  const merged = MOD.setValue(cfgPath, "b", "2");
  assert.strictEqual(merged.a, "1");
  assert.strictEqual(merged.b, "2");
  const onDisk = MOD.readConfig(cfgPath);
  assert.strictEqual(onDisk.a, "1");
  assert.strictEqual(onDisk.b, "2");
});

test("7. coerce true/false/string", () => {
  assert.strictEqual(MOD.coerce("true"), true);
  assert.strictEqual(MOD.coerce("false"), false);
  assert.strictEqual(MOD.coerce("hello"), "hello");
});

test("8. list emits merged JSON with defaults + overrides", () => {
  const home = tmpHome();
  const cfgPath = path.join(home, "config.json");
  MOD.setValue(cfgPath, "extra", "x");
  const out = cli(["list"], home);
  const parsed = JSON.parse(out);
  assert.strictEqual(parsed.proactive, true); // default present
  assert.strictEqual(parsed.extra, "x"); // override present
});

test("9. atomic write: no leftover temp files, target present", () => {
  const home = tmpHome();
  const cfgPath = path.join(home, "config.json");
  MOD.setValue(cfgPath, "proactive", "false");
  assert.strictEqual(fs.existsSync(cfgPath), true);
  const leftovers = fs.readdirSync(home).filter((f) => f.includes(".tmp"));
  assert.deepStrictEqual(leftovers, []);
});

test("10. ATV_CONFIG_HOME override honored", () => {
  const home = tmpHome();
  const realHome = path.join(os.homedir(), ".atv", "config.json");
  const before = fs.existsSync(realHome) ? fs.readFileSync(realHome, "utf8") : null;
  cli(["set", "proactive", "false"], home);
  assert.strictEqual(fs.existsSync(path.join(home, "config.json")), true);
  const after = fs.existsSync(realHome) ? fs.readFileSync(realHome, "utf8") : null;
  assert.strictEqual(after, before); // real home untouched
});
