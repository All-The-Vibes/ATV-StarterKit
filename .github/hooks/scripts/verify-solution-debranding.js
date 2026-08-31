#!/usr/bin/env node
"use strict";

const crypto = require("node:crypto");
const fs = require("node:fs");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

const repoRoot = path.resolve(__dirname, "..", "..", "..");
const lockPath = path.join(repoRoot, "solution-debranding.lock.json");
const modes = new Set(["vendor", "distribution", "lifecycle"]);
const REQUIRED_SKILLS = [
  "solution-debranding",
  "solution-debranding-plan",
  "solution-debranding-apply",
  "solution-debranding-verify",
];

function fail(message) {
  throw new Error(message);
}

function sha256(file) {
  return crypto.createHash("sha256").update(fs.readFileSync(file)).digest("hex");
}

function listFiles(root) {
  const files = [];
  function walk(current) {
    if (!fs.existsSync(current)) return;
    for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
      const full = path.join(current, entry.name);
      if (entry.isDirectory()) walk(full);
      else if (entry.isFile()) files.push(full);
    }
  }
  walk(root);
  return files.sort();
}

function verifyTree(root, lock) {
  const expected = new Map(lock.files.map((file) => [file.path, file.sha256]));
  const actualFamily = fs
    .readdirSync(root, { withFileTypes: true })
    .filter((entry) => entry.isDirectory() && entry.name.startsWith("solution-debranding"))
    .map((entry) => entry.name)
    .sort();
  const requiredFamily = [...REQUIRED_SKILLS].sort();
  if (JSON.stringify(actualFamily) !== JSON.stringify(requiredFamily)) {
    fail(`${root} must contain exactly ${requiredFamily.join(", ")}`);
  }
  const actualFiles = REQUIRED_SKILLS
    .flatMap((skill) => listFiles(path.join(root, skill)))
    .map((file) => path.relative(root, file).split(path.sep).join("/"));
  const actual = new Set(actualFiles);
  for (const [relative, digest] of expected) {
    const file = path.join(root, ...relative.split("/"));
    if (!actual.has(relative)) fail(`${root} is missing ${relative}`);
    const found = sha256(file);
    if (found !== digest) fail(`${root}/${relative} hash ${found} does not match ${digest}`);
  }
  for (const relative of actual) {
    if (!expected.has(relative)) fail(`${root} has unexpected vendored file ${relative}`);
  }
}

function runUpstreamTests(templateRoot) {
  const python = process.platform === "win32" ? "python" : "python3";
  const version = spawnSync(python, ["--version"], { encoding: "utf8" });
  if (version.status !== 0) {
    process.stdout.write("Python unavailable; skipped upstream script tests\n");
    return;
  }
  const match = `${version.stdout}${version.stderr}`.match(/Python (\d+)\.(\d+)/);
  if (!match || Number(match[1]) < 3 || (Number(match[1]) === 3 && Number(match[2]) < 11)) {
    process.stdout.write("Python 3.11+ unavailable; skipped upstream script tests\n");
    return;
  }
  const tests = path.join(templateRoot, "solution-debranding", "tests", "test_scripts.py");
  const result = spawnSync(python, ["-m", "unittest", tests], {
    cwd: repoRoot,
    encoding: "utf8",
    env: { ...process.env, PYTHONDONTWRITEBYTECODE: "1" },
  });
  if (result.status !== 0) {
    process.stderr.write(result.stdout || "");
    process.stderr.write(result.stderr || "");
    fail("upstream solution-debranding tests failed");
  }
}

function run(command, args) {
  const result = spawnSync(command, args, { cwd: repoRoot, encoding: "utf8" });
  if (result.status !== 0) {
    process.stderr.write(result.stdout || "");
    process.stderr.write(result.stderr || "");
    fail(`${command} ${args.join(" ")} failed`);
  }
}

function verifyVendor() {
  if (!fs.existsSync(lockPath)) fail("solution-debranding.lock.json is missing");
  const lock = JSON.parse(fs.readFileSync(lockPath, "utf8"));
  if (lock.repository !== "https://github.com/lshade/solution-debranding") {
    fail(`unexpected upstream repository ${lock.repository}`);
  }
  if (!/^[0-9a-f]{40}$/.test(lock.commit)) fail("lock commit must be a full SHA");
  if (!Array.isArray(lock.files) || lock.files.length === 0) fail("lock files must be non-empty");
  const lockedFamily = [...new Set(lock.files.map((file) => file.path.split("/")[0]))].sort();
  const requiredFamily = [...REQUIRED_SKILLS].sort();
  if (JSON.stringify(lockedFamily) !== JSON.stringify(requiredFamily)) {
    fail(`lock must contain exactly ${requiredFamily.join(", ")}`);
  }
  if (new Set(lock.files.map((file) => file.path)).size !== lock.files.length) {
    fail("lock contains duplicate file paths");
  }
  const templateRoot = path.join(repoRoot, "pkg", "scaffold", "templates", "skills");
  const dogfoodRoot = path.join(repoRoot, ".github", "skills");
  verifyTree(templateRoot, lock);
  verifyTree(dogfoodRoot, lock);
  const notice = fs.readFileSync(
    path.join(repoRoot, "THIRD_PARTY_NOTICES", "solution-debranding-LICENSE"),
    "utf8"
  );
  if (!notice.includes("Copyright (c) 2026 Lisa Shade") || !notice.includes("MIT License")) {
    fail("solution-debranding attribution is incomplete");
  }
  runUpstreamTests(templateRoot);
  run("node", ["--test", ".github/hooks/scripts/tests/sync-solution-debranding.test.js"]);
  process.stdout.write(`vendor verified at ${lock.commit} (${lock.files.length} files)\n`);
}

function verifyDistribution() {
  verifyVendor();
  const lock = JSON.parse(fs.readFileSync(lockPath, "utf8"));
  const pluginsRoot = path.join(repoRoot, "plugins");
  const pluginNames = [
    "atv-everything",
    "atv-pack-quality",
    "atv-skill-solution-debranding",
    "atv-skill-solution-debranding-apply",
    "atv-skill-solution-debranding-plan",
    "atv-skill-solution-debranding-verify",
  ];
  for (const plugin of pluginNames) {
    verifyTree(path.join(pluginsRoot, plugin, "skills"), lock);
  }

  const catalog = fs.readFileSync(path.join(repoRoot, "pkg", "scaffold", "catalog.go"), "utf8");
  for (const skill of [
    "solution-debranding",
    "solution-debranding-apply",
    "solution-debranding-plan",
    "solution-debranding-verify",
  ]) {
    if (!catalog.includes(`"${skill}"`)) fail(`scaffold catalog does not register ${skill}`);
  }

  for (const stage of ["apply", "plan", "verify"]) {
    const prompt = path.join(
      repoRoot,
      ".github",
      "prompts",
      `solution-debranding-${stage}.prompt.md`
    );
    if (!fs.existsSync(prompt)) fail(`missing prompt shim ${path.relative(repoRoot, prompt)}`);
  }
  if (fs.existsSync(path.join(repoRoot, ".github", "prompts", "solution-debranding.prompt.md"))) {
    fail("shared solution-debranding package must not have a prompt shim");
  }

  const qualityManifest = JSON.parse(
    fs.readFileSync(path.join(pluginsRoot, "atv-pack-quality", "plugin.json"), "utf8")
  );
  if (!qualityManifest.description.includes("Solution Debranding")) {
    fail("quality pack does not describe Solution Debranding");
  }
  run("go", ["run", "./cmd/plugingen", "-check"]);
  process.stdout.write("distribution verified across scaffold, prompts, and plugins\n");
}

function verifyLifecycle() {
  const templates = path.join(repoRoot, "pkg", "scaffold", "templates", "skills");
  const required = [
    "quality-release-readiness",
    "get-decision quality-release-readiness",
    "--choice <choice>",
    "/unslop fix",
    "/solution-debranding-plan",
    "/solution-debranding-apply <debranding-plan-path>",
    "/solution-debranding-verify <debranding-plan-path>",
    "solution-debranding-apply",
    "solution-debranding-verify",
    "Resolve both new and resumed decisions",
    "stored debranding plan artifact",
    "declined",
    "not-applicable",
    "failed verify result blocks completion",
  ];

  for (const orchestrator of ["lfg", "slfg"]) {
    const canonicalPath = path.join(templates, orchestrator, "SKILL.md");
    const canonical = fs.readFileSync(canonicalPath, "utf8");
    for (const text of required) {
      if (!canonical.includes(text)) fail(`${orchestrator} lifecycle is missing ${text}`);
    }
    const reviewIndex = canonical.indexOf("/ce-review mode:autofix");
    const unslopIndex = canonical.indexOf("/unslop fix", reviewIndex);
    const gateIndex = canonical.indexOf("quality-release-readiness", reviewIndex);
    const planIndex = canonical.indexOf("/solution-debranding-plan", gateIndex);
    const applyIndex = canonical.indexOf("/solution-debranding-apply", planIndex);
    const verifyIndex = canonical.indexOf("/solution-debranding-verify", applyIndex);
    if (!(reviewIndex < gateIndex && gateIndex < unslopIndex)) {
      fail(`${orchestrator} must place unslop inside the post-review release-readiness stage`);
    }
    if (!(gateIndex < planIndex && planIndex < applyIndex && applyIndex < verifyIndex)) {
      fail(`${orchestrator} must preserve debranding plan, apply, verify order`);
    }

    const copies = [
      path.join(repoRoot, ".github", "skills", orchestrator, "SKILL.md"),
      ...listFiles(path.join(repoRoot, "plugins"))
        .filter((file) => file.endsWith(path.join("skills", orchestrator, "SKILL.md"))),
    ];
    for (const copy of copies) {
      if (fs.readFileSync(copy, "utf8").replace(/\r\n/g, "\n") !== canonical.replace(/\r\n/g, "\n")) {
        fail(`${path.relative(repoRoot, copy)} drifted from canonical ${orchestrator}`);
      }
    }
  }

  const helperTemplate = fs.readFileSync(
    path.join(repoRoot, "pkg", "scaffold", "templates", "hooks", "scripts", "lfg-state.js"),
    "utf8"
  );
  const helperDogfood = fs.readFileSync(
    path.join(repoRoot, ".github", "hooks", "scripts", "lfg-state.js"),
    "utf8"
  );
  if (helperTemplate.replace(/\r\n/g, "\n") !== helperDogfood.replace(/\r\n/g, "\n")) {
    fail("scaffolded lfg-state.js drifted from dogfood helper");
  }
  for (const orchestrator of ["lfg", "slfg"]) {
    const packageHelper = path.join(templates, orchestrator, "scripts", "lfg-state.js");
    if (
      !fs.existsSync(packageHelper) ||
      fs.readFileSync(packageHelper, "utf8").replace(/\r\n/g, "\n") !==
        helperTemplate.replace(/\r\n/g, "\n")
    ) {
      fail(`${orchestrator} package-local lfg-state.js is missing or stale`);
    }
    const pluginHelpers = listFiles(path.join(repoRoot, "plugins")).filter((file) =>
      file.endsWith(path.join("skills", orchestrator, "scripts", "lfg-state.js"))
    );
    if (pluginHelpers.length === 0) fail(`plugins do not package the ${orchestrator} state helper`);
    for (const pluginHelper of pluginHelpers) {
      if (
        fs.readFileSync(pluginHelper, "utf8").replace(/\r\n/g, "\n") !==
        helperTemplate.replace(/\r\n/g, "\n")
      ) {
        fail(`${path.relative(repoRoot, pluginHelper)} drifted from canonical state helper`);
      }
    }
  }
  for (const text of ["recordDecision", "getDecision", "accepted", "declined", "not-applicable"]) {
    if (!helperTemplate.includes(text)) fail(`lfg-state helper is missing ${text}`);
  }

  const readme = fs.readFileSync(path.join(repoRoot, "README.md"), "utf8");
  const docs = fs.readFileSync(path.join(repoRoot, "DOCS.md"), "utf8");
  if (!readme.includes("unslop + debranding proposal")) fail("README lifecycle diagram is stale");
  if (!docs.includes("quality-release-readiness")) fail("DOCS lifecycle contract is stale");

  run("node", ["--test", ".github/hooks/scripts/tests/lfg-state.test.js"]);
  run("go", ["run", "./cmd/plugingen", "-check"]);
  process.stdout.write("lifecycle verified for fresh, resumed, LFG, and SLFG routing\n");
}

try {
  const mode = process.argv[2];
  if (!modes.has(mode)) fail("usage: verify-solution-debranding.js <vendor|distribution|lifecycle>");
  if (mode === "vendor") verifyVendor();
  else if (mode === "distribution") verifyDistribution();
  else verifyLifecycle();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
