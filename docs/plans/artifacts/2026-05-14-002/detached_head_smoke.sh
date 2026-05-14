#!/usr/bin/env bash
# Smoke test for U2's detached-HEAD guard.
# Verifies that the guarded snippet from land/SKILL.md Step 8 no-ops cleanly
# on detached HEAD, while the unguarded original errors out.

set -e
SMOKE_DIR="/tmp/detached-smoke"
rm -rf "$SMOKE_DIR"
mkdir -p "$SMOKE_DIR"
cd "$SMOKE_DIR"

git init -q
git config user.email "smoke@test"
git config user.name "smoke"
echo "test" > a.txt
git add a.txt
git -c commit.gpgsign=false commit -q -m "seed"
git checkout -q --detach

echo "=== detached HEAD state ==="
git status --short --branch
echo ""

echo "=== Running GUARDED snippet (post-fix, from land/SKILL.md Step 8) ==="
set +e
if branch="$(git branch --show-current)" && [ -n "$branch" ]; then
  if git rev-parse --verify --quiet "refs/remotes/origin/$branch" >/dev/null; then
    git log "origin/$branch..HEAD"
  else
    echo "(no origin/\$branch yet -- push the branch before landing)"
  fi
else
  echo "(detached HEAD -- skipping unpushed-commits check)"
fi
guarded_rc=$?
echo "guarded snippet exit code: $guarded_rc"
echo ""

echo "=== Running UNGUARDED snippet (pre-fix, original) for contrast ==="
git log "origin/$(git branch --show-current)..HEAD" 2>&1
unguarded_rc=$?
echo "unguarded snippet exit code: $unguarded_rc"
set -e

echo ""
echo "=== Verdict ==="
if [ "$guarded_rc" -eq 0 ] && [ "$unguarded_rc" -ne 0 ]; then
  echo "PASS: guarded snippet succeeds on detached HEAD; unguarded original fails as expected."
  exit 0
else
  echo "FAIL: guarded=$guarded_rc unguarded=$unguarded_rc"
  exit 1
fi
