#!/usr/bin/env node

"use strict";

const path = require("path");
const os = require("os");
const { execFileSync } = require("child_process");
const fs = require("fs");

const BINARY_NAME = "atv-installer";
const binaryExt = os.platform() === "win32" ? ".exe" : "";
const binaryPath = path.join(__dirname, `${BINARY_NAME}${binaryExt}`);

if (!fs.existsSync(binaryPath)) {
  console.error(
    `Error: ${BINARY_NAME} binary not found at ${binaryPath}\n\n` +
      `The binary may not have been downloaded during installation.\n` +
      `Try reinstalling: npm install -g atv-starterkit\n\n` +
      `Or download manually from:\n` +
      `https://github.com/All-The-Vibes/ATV-StarterKit/releases`
  );
  process.exit(1);
}

// Forward all arguments to the Go binary
const args = process.argv.slice(2);

try {
  execFileSync(binaryPath, args, { stdio: "inherit" });
} catch (err) {
  // execFileSync throws if the child process exits with non-zero.
  // The child's stderr/stdout were already inherited, so just propagate the exit code.
  process.exit(err.status || 1);
}
