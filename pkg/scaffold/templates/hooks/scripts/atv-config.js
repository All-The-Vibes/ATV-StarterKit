#!/usr/bin/env node
// ATV config shim — override-only key/value store backed by ~/.atv/config.json.
//
// The skill owns the real defaults; this shim carries a small DEFAULTS table as
// the fallback source of truth and persists user overrides. Mirrors the style of
// observe.js / lfg-state.js: "use strict", pure exported functions, atomic
// temp-write + rename, and an env override (ATV_CONFIG_HOME) for test isolation
// so tests never touch the real home directory.
//
// Usable two ways:
//   * require('./atv-config') — exported pure + fs-backed functions (tested)
//   * node atv-config.js <command> ...                          — CLI for skills
//
// CLI:
//   atv-config get <key>          print effective value (override else default); exit 0
//   atv-config set <key> <value>  persist override to ~/.atv/config.json; exit 0
//   atv-config list               print all effective values (defaults + overrides) as JSON

"use strict";

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

// The skill owns the real default; this table is the shim's fallback truth.
const DEFAULTS = {
  proactive: true, // controls auto-invoke. false = suggest-only
};

// --- pure helpers ----------------------------------------------------------

function coerce(value) {
  if (value === "true") return true;
  if (value === "false") return false;
  return value;
}

function configPath() {
  const home = process.env.ATV_CONFIG_HOME;
  if (home) return path.join(home, "config.json");
  return path.join(os.homedir(), ".atv", "config.json");
}

// Returns {} on missing/unreadable/corrupt file — never throws.
function readConfig(filePath) {
  let raw;
  try {
    raw = fs.readFileSync(filePath, "utf8");
  } catch (_) {
    return {};
  }
  try {
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
      return parsed;
    }
    return {};
  } catch (_) {
    return {};
  }
}

// Effective value: override in file, else DEFAULTS, else "".
function getValue(cfg, key) {
  if (cfg && Object.prototype.hasOwnProperty.call(cfg, key)) return cfg[key];
  if (Object.prototype.hasOwnProperty.call(DEFAULTS, key)) return DEFAULTS[key];
  return "";
}

// --- fs-backed ops ---------------------------------------------------------

function atomicWriteJson(filePath, obj) {
  const tmp = `${filePath}.${process.pid}.${Date.now()}.tmp`;
  fs.writeFileSync(tmp, JSON.stringify(obj, null, 2) + "\n", { mode: 0o600 });
  try {
    fs.renameSync(tmp, filePath);
  } catch (err) {
    if (err && (err.code === "EEXIST" || err.code === "EPERM")) {
      fs.rmSync(filePath, { force: true });
      fs.renameSync(tmp, filePath);
    } else {
      fs.rmSync(tmp, { force: true });
      throw err;
    }
  }
}

// Merges key/value into existing (parseable) overrides — corrupt content is
// discarded rather than merged — then writes atomically. Returns merged object.
function setValue(filePath, key, value) {
  fs.mkdirSync(path.dirname(filePath), { recursive: true, mode: 0o700 });
  const current = readConfig(filePath);
  const merged = Object.assign({}, current, { [key]: coerce(value) });
  atomicWriteJson(filePath, merged);
  return merged;
}

// --- CLI -------------------------------------------------------------------

function main(argv) {
  const cmd = argv[0];
  const filePath = configPath();

  switch (cmd) {
    case "get": {
      const key = argv[1];
      const val = getValue(readConfig(filePath), key);
      process.stdout.write(String(val === undefined || val === null ? "" : val) + "\n");
      return 0;
    }
    case "set": {
      const key = argv[1];
      const value = argv[2];
      if (key === undefined) {
        process.stderr.write("atv-config set: missing <key>\n");
        return 2;
      }
      setValue(filePath, key, value === undefined ? "" : value);
      return 0;
    }
    case "list": {
      const merged = Object.assign({}, DEFAULTS, readConfig(filePath));
      process.stdout.write(JSON.stringify(merged, null, 2) + "\n");
      return 0;
    }
    default:
      process.stderr.write(
        "usage: atv-config get <key> | set <key> <value> | list\n"
      );
      return 2;
  }
}

module.exports = {
  DEFAULTS,
  coerce,
  configPath,
  readConfig,
  getValue,
  setValue,
};

if (require.main === module) {
  process.exit(main(process.argv.slice(2)));
}
