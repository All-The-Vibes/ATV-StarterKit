#!/usr/bin/env node
"use strict";

const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

const REPOSITORY = "https://github.com/lshade/solution-debranding";
const SKILLS = [
  "solution-debranding",
  "solution-debranding-plan",
  "solution-debranding-apply",
  "solution-debranding-verify",
];

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd,
    encoding: "utf8",
    env: options.env,
    stdio: options.capture ? "pipe" : "inherit",
  });
  if (result.status !== 0) {
    throw new Error(`${command} ${args.join(" ")} failed with exit code ${result.status}`);
  }
  return String(result.stdout || "").trim();
}

function listFiles(root) {
  const files = [];
  function walk(current) {
    for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
      const full = path.join(current, entry.name);
      if (entry.isDirectory()) walk(full);
      else if (entry.isFile()) files.push(full);
    }
  }
  walk(root);
  return files.sort();
}

function sha256(file) {
  return crypto.createHash("sha256").update(fs.readFileSync(file)).digest("hex");
}

function manifest(root) {
  return SKILLS.flatMap((skill) => listFiles(path.join(root, skill))).map((file) => ({
    path: path.relative(root, file).split(path.sep).join("/"),
    sha256: sha256(file),
  }));
}

function runUpstreamTests(templateRoot) {
  const python = process.platform === "win32" ? "python" : "python3";
  const version = spawnSync(python, ["--version"], { encoding: "utf8" });
  if (version.status !== 0) {
    throw new Error("Python 3.11+ is required to validate staged upstream scripts");
  }
  const match = `${version.stdout}${version.stderr}`.match(/Python (\d+)\.(\d+)/);
  if (!match || Number(match[1]) < 3 || (Number(match[1]) === 3 && Number(match[2]) < 11)) {
    throw new Error("Python 3.11+ is required to validate staged upstream scripts");
  }
  run(
    python,
    [
      "-m",
      "unittest",
      path.join(templateRoot, "solution-debranding", "tests", "test_scripts.py"),
    ],
    {
      cwd: templateRoot,
      env: { ...process.env, PYTHONDONTWRITEBYTECODE: "1" },
    }
  );
}

function replaceTransactionally(operations, backupRoot) {
  const applied = [];
  fs.mkdirSync(backupRoot, { recursive: true });
  try {
    operations.forEach(({ staged, destination }, index) => {
      const backup = path.join(backupRoot, String(index));
      fs.mkdirSync(path.dirname(destination), { recursive: true });
      const hadDestination = fs.existsSync(destination);
      if (hadDestination) fs.renameSync(destination, backup);
      try {
        fs.renameSync(staged, destination);
      } catch (error) {
        if (hadDestination) fs.renameSync(backup, destination);
        throw error;
      }
      applied.push({ destination, backup, hadDestination });
    });
  } catch (error) {
    const rollbackErrors = [];
    for (const item of applied.reverse()) {
      try {
        fs.rmSync(item.destination, { recursive: true, force: true });
        if (item.hadDestination) fs.renameSync(item.backup, item.destination);
      } catch (rollbackError) {
        rollbackErrors.push(rollbackError.message);
      }
    }
    if (rollbackErrors.length) {
      throw new Error(`${error.message}; rollback failed: ${rollbackErrors.join("; ")}`);
    }
    throw error;
  }
}

function main() {
  const commit = process.argv[2];
  if (!commit || !/^[0-9a-f]{40}$/i.test(commit)) {
    throw new Error("usage: sync-solution-debranding.js <40-character-upstream-commit>");
  }

  const repoRoot = path.resolve(__dirname, "..", "..", "..");
  const dirty = run("git", ["status", "--porcelain"], { cwd: repoRoot, capture: true });
  if (dirty) {
    throw new Error("refusing to sync into a dirty worktree");
  }

  const temp = fs.mkdtempSync(path.join(os.tmpdir(), "solution-debranding-"));
  const staging = path.join(
    repoRoot,
    `.solution-debranding-sync-${process.pid}-${Date.now()}`
  );
  try {
    run("git", ["clone", "--quiet", REPOSITORY, temp]);
    run("git", ["checkout", "--quiet", commit], { cwd: temp });
    const resolved = run("git", ["rev-parse", "HEAD"], { cwd: temp, capture: true });
    if (resolved.toLowerCase() !== commit.toLowerCase()) {
      throw new Error(`requested ${commit}, checked out ${resolved}`);
    }

    const upstreamSkills = path.join(temp, "skills");
    for (const skill of SKILLS) {
      const source = path.join(upstreamSkills, skill);
      if (
        !fs.existsSync(source) ||
        !fs.statSync(source).isDirectory() ||
        !fs.existsSync(path.join(source, "SKILL.md"))
      ) {
        throw new Error(`upstream checkout is missing complete skill ${skill}`);
      }
    }
    if (!fs.existsSync(path.join(temp, "LICENSE"))) {
      throw new Error("upstream checkout is missing LICENSE");
    }

    const stagedRoots = [
      path.join(staging, "template-skills"),
      path.join(staging, "dogfood-skills"),
    ];
    for (const skill of SKILLS) {
      const source = path.join(upstreamSkills, skill);
      for (const root of stagedRoots) {
        fs.cpSync(source, path.join(root, skill), { recursive: true });
      }
    }

    const files = manifest(stagedRoots[0]);
    if (JSON.stringify(files) !== JSON.stringify(manifest(stagedRoots[1]))) {
      throw new Error("staged template and dogfood trees differ");
    }
    if (files.length === 0 || new Set(files.map((file) => file.path)).size !== files.length) {
      throw new Error("staged package inventory is empty or contains duplicate paths");
    }
    runUpstreamTests(stagedRoots[0]);
    const lock = { repository: REPOSITORY, commit: resolved, files };
    const stagedLock = path.join(staging, "solution-debranding.lock.json");
    fs.writeFileSync(stagedLock, `${JSON.stringify(lock, null, 2)}\n`);
    const stagedNotice = path.join(staging, "solution-debranding-LICENSE");
    fs.copyFileSync(path.join(temp, "LICENSE"), stagedNotice);
    const notice = fs.readFileSync(stagedNotice, "utf8");
    if (!notice.includes("Copyright (c) 2026 Lisa Shade") || !notice.includes("MIT License")) {
      throw new Error("staged upstream attribution is incomplete");
    }

    const destinationRoots = [
      path.join(repoRoot, "pkg", "scaffold", "templates", "skills"),
      path.join(repoRoot, ".github", "skills"),
    ];
    const operations = [];
    stagedRoots.forEach((root, rootIndex) => {
      for (const skill of SKILLS) {
        operations.push({
          staged: path.join(root, skill),
          destination: path.join(destinationRoots[rootIndex], skill),
        });
      }
    });
    operations.push(
      {
        staged: stagedLock,
        destination: path.join(repoRoot, "solution-debranding.lock.json"),
      },
      {
        staged: stagedNotice,
        destination: path.join(
          repoRoot,
          "THIRD_PARTY_NOTICES",
          "solution-debranding-LICENSE"
        ),
      }
    );
    replaceTransactionally(operations, path.join(staging, "backups"));
    process.stdout.write(`Synced solution-debranding at ${resolved}\n`);
  } finally {
    fs.rmSync(temp, { recursive: true, force: true });
    fs.rmSync(staging, { recursive: true, force: true });
  }
}

if (require.main === module) {
  try {
    main();
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exit(1);
  }
}

module.exports = { replaceTransactionally };
