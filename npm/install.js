#!/usr/bin/env node

"use strict";

const https = require("https");
const fs = require("fs");
const path = require("path");
const os = require("os");
const { execSync } = require("child_process");
const zlib = require("zlib");

const REPO_OWNER = "All-The-Vibes";
const REPO_NAME = "ATV-StarterKit";
const BINARY_NAME = "atv-installer";

/**
 * Resolve the platform and architecture to match goreleaser naming.
 * Returns { platform, arch, ext } or throws if unsupported.
 */
function resolvePlatform() {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap = {
    darwin: "darwin",
    linux: "linux",
    win32: "windows",
  };

  const archMap = {
    x64: "amd64",
    arm64: "arm64",
  };

  const resolvedPlatform = platformMap[platform];
  const resolvedArch = archMap[arch];

  if (!resolvedPlatform || !resolvedArch) {
    throw new Error(
      `Unsupported platform: ${platform}/${arch}. ` +
        `Supported: darwin/linux/windows on amd64/arm64.`
    );
  }

  const ext = platform === "win32" ? "zip" : "tar.gz";

  return { platform: resolvedPlatform, arch: resolvedArch, ext };
}

/**
 * Build the expected archive filename from goreleaser naming convention.
 */
function buildArchiveName(version, platform, arch, ext) {
  return `${BINARY_NAME}_${version}_${platform}_${arch}.${ext}`;
}

/**
 * Follow redirects and return the final response (up to 5 redirects).
 */
function httpsGet(url, headers = {}) {
  return new Promise((resolve, reject) => {
    const options = {
      headers: {
        "User-Agent": "atv-starterkit-npm-installer",
        ...headers,
      },
    };

    https
      .get(url, options, (res) => {
        if (
          res.statusCode >= 300 &&
          res.statusCode < 400 &&
          res.headers.location
        ) {
          return httpsGet(res.headers.location, headers).then(resolve, reject);
        }
        resolve(res);
      })
      .on("error", reject);
  });
}

/**
 * Fetch the latest release tag from GitHub API.
 */
async function getLatestVersion() {
  const url = `https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest`;
  const res = await httpsGet(url, { Accept: "application/vnd.github.v3+json" });

  return new Promise((resolve, reject) => {
    let data = "";
    res.on("data", (chunk) => (data += chunk));
    res.on("end", () => {
      if (res.statusCode !== 200) {
        reject(
          new Error(
            `GitHub API returned ${res.statusCode}. ` +
              `No releases found for ${REPO_OWNER}/${REPO_NAME}. ` +
              `Ensure a release exists before installing via npm.`
          )
        );
        return;
      }
      try {
        const release = JSON.parse(data);
        // tag_name is typically "v1.0.0", strip the "v" prefix for archive naming
        resolve(release.tag_name);
      } catch (e) {
        reject(new Error(`Failed to parse GitHub release response: ${e.message}`));
      }
    });
  });
}

/**
 * Download a file from a URL to a local path.
 */
async function downloadFile(url, destPath) {
  const res = await httpsGet(url);

  if (res.statusCode !== 200) {
    throw new Error(`Download failed with status ${res.statusCode}: ${url}`);
  }

  return new Promise((resolve, reject) => {
    const fileStream = fs.createWriteStream(destPath);
    res.pipe(fileStream);
    fileStream.on("finish", () => {
      fileStream.close();
      resolve();
    });
    fileStream.on("error", (err) => {
      fs.unlinkSync(destPath);
      reject(err);
    });
  });
}

/**
 * Extract a .tar.gz archive to a destination directory.
 */
function extractTarGz(archivePath, destDir) {
  // Use tar command which is available on macOS and Linux
  execSync(`tar xzf "${archivePath}" -C "${destDir}"`, { stdio: "pipe" });
}

/**
 * Extract a .zip archive to a destination directory.
 */
function extractZip(archivePath, destDir) {
  // Use PowerShell on Windows for zip extraction
  if (os.platform() === "win32") {
    execSync(
      `powershell -Command "Expand-Archive -Path '${archivePath}' -DestinationPath '${destDir}' -Force"`,
      { stdio: "pipe" }
    );
  } else {
    execSync(`unzip -o "${archivePath}" -d "${destDir}"`, { stdio: "pipe" });
  }
}

/**
 * Main installation logic.
 */
async function install() {
  const binDir = path.join(__dirname, "bin");

  // Check if binary already exists
  const binaryExt = os.platform() === "win32" ? ".exe" : "";
  const binaryPath = path.join(binDir, `${BINARY_NAME}${binaryExt}`);

  if (fs.existsSync(binaryPath)) {
    console.log(`  ✓ ${BINARY_NAME} binary already exists, skipping download.`);
    return;
  }

  console.log("  ATV Starter Kit — downloading binary for your platform...\n");

  // Step 1: Resolve platform
  const { platform, arch, ext } = resolvePlatform();
  console.log(`  Platform: ${platform}/${arch}`);

  // Step 2: Get latest version
  let version;
  try {
    version = await getLatestVersion();
  } catch (err) {
    console.error(
      `\n  ⚠ Could not fetch latest release: ${err.message}\n` +
        `  You can manually download the binary from:\n` +
        `  https://github.com/${REPO_OWNER}/${REPO_NAME}/releases\n`
    );
    process.exit(0); // Don't fail npm install, just warn
  }

  const versionNumber = version.replace(/^v/, "");
  console.log(`  Version:  ${version}`);

  // Step 3: Download archive
  const archiveName = buildArchiveName(versionNumber, platform, arch, ext);
  const downloadUrl = `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${archiveName}`;

  console.log(`  Downloading: ${archiveName}\n`);

  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "atv-"));
  const archivePath = path.join(tmpDir, archiveName);

  try {
    await downloadFile(downloadUrl, archivePath);
  } catch (err) {
    console.error(
      `\n  ⚠ Download failed: ${err.message}\n` +
        `  You can manually download from:\n` +
        `  ${downloadUrl}\n`
    );
    // Cleanup
    fs.rmSync(tmpDir, { recursive: true, force: true });
    process.exit(0); // Don't fail npm install
  }

  // Step 4: Extract
  try {
    if (ext === "tar.gz") {
      extractTarGz(archivePath, tmpDir);
    } else {
      extractZip(archivePath, tmpDir);
    }
  } catch (err) {
    console.error(`\n  ⚠ Extraction failed: ${err.message}`);
    fs.rmSync(tmpDir, { recursive: true, force: true });
    process.exit(0);
  }

  // Step 5: Move binary to bin/
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  const extractedBinary = path.join(tmpDir, `${BINARY_NAME}${binaryExt}`);
  if (!fs.existsSync(extractedBinary)) {
    console.error(
      `\n  ⚠ Binary not found in archive. Expected: ${BINARY_NAME}${binaryExt}`
    );
    fs.rmSync(tmpDir, { recursive: true, force: true });
    process.exit(0);
  }

  fs.copyFileSync(extractedBinary, binaryPath);
  fs.chmodSync(binaryPath, 0o755);

  // Cleanup
  fs.rmSync(tmpDir, { recursive: true, force: true });

  console.log(`  ✓ Installed ${BINARY_NAME} ${version} to ${binaryPath}\n`);
}

install().catch((err) => {
  console.error(`  ⚠ Installation error: ${err.message}`);
  process.exit(0); // Don't fail npm install
});
