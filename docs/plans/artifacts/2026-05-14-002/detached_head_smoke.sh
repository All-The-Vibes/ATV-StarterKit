#!/usr/bin/env bash
# Smoke test for the land/SKILL.md Step 8 push-verification guard.
#
# Exercises the CURRENT Step 8 snippet across all four states it must handle:
#   1. detached HEAD          -> non-fatal skip, rc 0
#   2. branch, no upstream    -> BLOCKED, rc 1
#   3. branch, unpushed commit -> BLOCKED, rc 1
#   4. branch, fully pushed    -> OK, rc 0
# and contrasts state 1 against the pre-fix UNGUARDED one-liner, which errors
# (rc 128) on detached HEAD.

set -e
SMOKE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/land-step8-smoke.XXXXXX")"
trap 'rm -rf "$SMOKE_DIR"' EXIT
cd "$SMOKE_DIR"

# The guard, verbatim from land/SKILL.md Step 8 (kept in sync by
# test/skills/test_skill_invariants.sh R1, which asserts these same markers).
run_guard() {
  if branch="$(git branch --show-current)" && [ -n "$branch" ]; then
    if git rev-parse --verify --quiet "refs/remotes/origin/$branch" >/dev/null; then
      if [ -n "$(git log "origin/$branch..HEAD" --oneline)" ]; then
        echo "BLOCKED: unpushed commits on $branch -- push before landing." >&2
        # Normalize the leading short-SHA to <sha> so captured output is
        # byte-reproducible (the commit hash is nondeterministic).
        git log "origin/$branch..HEAD" --oneline | sed -E 's/^[0-9a-f]{7,} /<sha> /'
        return 1
      fi
      echo "OK: all commits pushed to origin/$branch."
    else
      echo "BLOCKED: $branch has no origin/$branch upstream -- push the branch before landing." >&2
      return 1
    fi
  else
    echo "(detached HEAD -- not on a branch; unpushed-commits check skipped)"
  fi
}

# --- shared origin so we can create real upstream states -------------------
git init -q --bare origin.git
git init -q work
cd work
git config user.email "smoke@test"
git config user.name "smoke"
git remote add origin "$SMOKE_DIR/origin.git"
echo seed > a.txt
git add a.txt
git -c commit.gpgsign=false commit -q -m "seed"

pass=0; fail=0
check() { # <label> <expected-rc> <actual-rc>
  if [ "$2" -eq "$3" ]; then echo "PASS: $1 (rc $3)"; pass=$((pass+1))
  else echo "FAIL: $1 (expected rc $2, got $3)"; fail=$((fail+1)); fi
}

echo "=== State 1: detached HEAD -> non-fatal skip (rc 0) ==="
git checkout -q --detach
set +e; run_guard; rc=$?; set -e
check "detached HEAD skips" 0 "$rc"
echo ""

echo "=== Contrast: pre-fix UNGUARDED one-liner on detached HEAD (errors) ==="
set +e; git log "origin/$(git branch --show-current)..HEAD" >/dev/null 2>&1; unrc=$?; set -e
check "unguarded original errors" 128 "$unrc"
echo ""

echo "=== State 2: branch with NO upstream -> BLOCKED (rc 1) ==="
git checkout -q -b feature
set +e; run_guard; rc=$?; set -e
check "no-upstream blocks" 1 "$rc"
echo ""

echo "=== State 3: branch with UNPUSHED commit -> BLOCKED (rc 1) ==="
git push -q -u origin feature          # establish upstream at current commit
echo more >> a.txt
git -c commit.gpgsign=false commit -q -am "unpushed work"
set +e; run_guard; rc=$?; set -e
check "unpushed commit blocks" 1 "$rc"
echo ""

echo "=== State 4: branch fully pushed -> OK (rc 0) ==="
git push -q origin feature
set +e; run_guard; rc=$?; set -e
check "fully pushed passes" 0 "$rc"
echo ""

echo "=== Verdict ==="
if [ "$fail" -eq 0 ]; then
  echo "PASS: all $pass Step 8 states behave correctly (detached skip, no-upstream/unpushed BLOCK, pushed OK)."
  exit 0
else
  echo "FAIL: $fail of $((pass+fail)) states wrong."
  exit 1
fi
