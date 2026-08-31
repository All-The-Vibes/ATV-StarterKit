const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

const { replaceTransactionally } = require("../sync-solution-debranding");

test("transactional replacement restores every live destination after a failure", () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "solution-debranding-sync-test-"));
  try {
    const destinationA = path.join(root, "live", "a.txt");
    const destinationB = path.join(root, "live", "b.txt");
    const stagedA = path.join(root, "staged", "a.txt");
    const missingStagedB = path.join(root, "staged", "missing.txt");
    fs.mkdirSync(path.dirname(destinationA), { recursive: true });
    fs.mkdirSync(path.dirname(stagedA), { recursive: true });
    fs.writeFileSync(destinationA, "original-a");
    fs.writeFileSync(destinationB, "original-b");
    fs.writeFileSync(stagedA, "replacement-a");

    assert.throws(
      () =>
        replaceTransactionally(
          [
            { staged: stagedA, destination: destinationA },
            { staged: missingStagedB, destination: destinationB },
          ],
          path.join(root, "backups")
        ),
      /ENOENT/
    );

    assert.equal(fs.readFileSync(destinationA, "utf8"), "original-a");
    assert.equal(fs.readFileSync(destinationB, "utf8"), "original-b");
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
});
