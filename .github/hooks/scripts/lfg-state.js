#!/usr/bin/env node
// ATV LFG/SLFG run-state helper.
//
// Backs the resumability + artifact-by-reference features of the `lfg` and
// `slfg` skills. State lives under `.atv/runs/<run-id>/`:
//
//   meta.json                 written once at init, then only the parent binds
//                             the plan path (single writer) — two-stage identity
//   phases/<phase>.done.json  one atomic sentinel per completed phase, written
//                             via temp-file + rename so parallel SLFG agents
//                             never corrupt or lose each other's updates
//
// Design notes:
//   * The consumer is an LLM measuring cost in tokens + tool round-trips, not
//     disk microseconds, so plain JSON files (readable, git-diffable, zero
//     dependencies) beat a SQLite/binary store. Mirrors observe.js.
//   * Per-phase sentinels avoid read-modify-write races entirely; "is this
//     phase done?" is just "does the sentinel exist?".
//
// Usable two ways:
//   * require('./lfg-state') — exported pure + fs-backed functions (tested)
//   * node lfg-state.js <command> ...                          — CLI for skills

"use strict";

const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");

const MAX_SLUG = 50;

// --- pure helpers ----------------------------------------------------------

function slugify(text) {
  return String(text == null ? "" : text)
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, MAX_SLUG)
    .replace(/-+$/g, "");
}

function runIdFromPlan(planPath) {
  const base = path.basename(String(planPath || ""));
  return base.replace(/\.md$/i, "").replace(/-plan$/i, "");
}

function provisionalRunId({ feature, repo, branch }) {
  const hash = crypto
    .createHash("sha1")
    .update(`${repo}|${branch}|${feature}`)
    .digest("hex")
    .slice(0, 8);
  const slug = slugify(feature) || "run";
  return `prov-${slug}-${hash}`.slice(0, MAX_SLUG + 16);
}

function sanitizePhase(name) {
  const safe = String(name || "");
  if (!/^[a-z0-9][a-z0-9-]*$/.test(safe)) {
    throw new Error(`invalid phase name: ${JSON.stringify(name)}`);
  }
  return safe;
}

function parseRunToken(args) {
  const text = String(args || "");
  const match = text.match(/(?:^|\s)run:(\S+)/);
  if (!match) return { runId: null, rest: text.trim() };
  const rest = (text.slice(0, match.index) + text.slice(match.index + match[0].length))
    .replace(/\s+/g, " ")
    .trim();
  return { runId: match[1], rest };
}

// --- filesystem helpers ----------------------------------------------------

function runDir(runsDir, runId) {
  return path.join(runsDir, runId);
}

function phasesDir(runsDir, runId) {
  return path.join(runDir(runsDir, runId), "phases");
}

function atomicWriteJson(filePath, obj) {
  const tmp = `${filePath}.${process.pid}.${Date.now()}.tmp`;
  fs.writeFileSync(tmp, JSON.stringify(obj, null, 2) + "\n");
  fs.renameSync(tmp, filePath);
}

function init(runsDir, { runId, skill, feature, plan }) {
  fs.mkdirSync(phasesDir(runsDir, runId), { recursive: true });
  const metaPath = path.join(runDir(runsDir, runId), "meta.json");
  if (fs.existsSync(metaPath)) {
    return JSON.parse(fs.readFileSync(metaPath, "utf8"));
  }
  const meta = {
    run_id: runId,
    skill: skill || null,
    feature: feature || null,
    plan_path: plan || null,
    created_at: new Date().toISOString(),
  };
  atomicWriteJson(metaPath, meta);
  return meta;
}

function bindPlan(runsDir, runId, planPath) {
  const metaPath = path.join(runDir(runsDir, runId), "meta.json");
  const meta = JSON.parse(fs.readFileSync(metaPath, "utf8"));
  meta.plan_path = planPath;
  atomicWriteJson(metaPath, meta);
  return meta;
}

function markDone(runsDir, runId, phase, opts) {
  const safe = sanitizePhase(phase);
  const options = opts || {};
  fs.mkdirSync(phasesDir(runsDir, runId), { recursive: true });
  const sentinel = {
    phase: safe,
    status: "done",
    artifact: options.artifact == null ? null : options.artifact,
    ended_at: new Date().toISOString(),
  };
  atomicWriteJson(path.join(phasesDir(runsDir, runId), `${safe}.done.json`), sentinel);
  return sentinel;
}

function isDone(runsDir, runId, phase) {
  const safe = sanitizePhase(phase);
  return fs.existsSync(path.join(phasesDir(runsDir, runId), `${safe}.done.json`));
}

function status(runsDir, runId) {
  const metaPath = path.join(runDir(runsDir, runId), "meta.json");
  let meta = null;
  if (fs.existsSync(metaPath)) {
    meta = JSON.parse(fs.readFileSync(metaPath, "utf8"));
  }
  const pdir = phasesDir(runsDir, runId);
  let phases = [];
  if (fs.existsSync(pdir)) {
    phases = fs
      .readdirSync(pdir)
      .filter((f) => f.endsWith(".done.json"))
      .map((f) => JSON.parse(fs.readFileSync(path.join(pdir, f), "utf8")))
      .sort((a, b) => String(a.ended_at).localeCompare(String(b.ended_at)));
  }
  return { run_id: runId, meta, phases };
}

// --- CLI -------------------------------------------------------------------

function defaultRunsDir() {
  return process.env.LFG_RUNS_DIR || path.join(process.cwd(), ".atv", "runs");
}

function parseFlags(argv) {
  const flags = {};
  const positional = [];
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a.startsWith("--")) {
      const key = a.slice(2);
      const val = argv[i + 1] && !argv[i + 1].startsWith("--") ? argv[++i] : "true";
      flags[key] = val;
    } else {
      positional.push(a);
    }
  }
  return { flags, positional };
}

function main(argv) {
  const [command, ...rest] = argv;
  const { flags, positional } = parseFlags(rest);
  const runsDir = flags["runs-dir"] || defaultRunsDir();
  let out;
  switch (command) {
    case "init": {
      const runId =
        flags["run-id"] ||
        (flags.plan
          ? runIdFromPlan(flags.plan)
          : provisionalRunId({
              feature: flags.feature || "",
              repo: flags.repo || "",
              branch: flags.branch || "",
            }));
      out = init(runsDir, {
        runId,
        skill: flags.skill,
        feature: flags.feature,
        plan: flags.plan,
      });
      break;
    }
    case "bind-plan":
      out = bindPlan(runsDir, flags["run-id"], flags.plan);
      break;
    case "done":
      out = markDone(runsDir, flags["run-id"], positional[0] || flags.phase, {
        artifact: flags.artifact,
      });
      break;
    case "is-done":
      out = { phase: positional[0], done: isDone(runsDir, flags["run-id"], positional[0]) };
      break;
    case "status":
      out = status(runsDir, flags["run-id"]);
      break;
    case "run-id-from-plan":
      out = { run_id: runIdFromPlan(flags.plan || positional[0]) };
      break;
    default:
      process.stderr.write(
        "usage: lfg-state.js <init|bind-plan|done|is-done|status|run-id-from-plan> [--run-id id] [--plan path] [--skill s] [--feature f] [--repo r] [--branch b] [--artifact path]\n"
      );
      process.exit(2);
  }
  process.stdout.write(JSON.stringify(out, null, 2) + "\n");
}

module.exports = {
  slugify,
  runIdFromPlan,
  provisionalRunId,
  sanitizePhase,
  parseRunToken,
  init,
  bindPlan,
  markDone,
  isDone,
  status,
};

if (require.main === module) {
  main(process.argv.slice(2));
}
