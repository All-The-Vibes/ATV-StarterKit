#!/usr/bin/env bash
# test/skills/test_skill_invariants.sh
#
# Mirror-aware invariant harness for the deferred MEDIUM review threads
# on the merged PR #42 (chore/sync-land-takeoff-ghcp-review-resolve-skills).
#
# Each check pins on the exact failure shape that broke the prior review and
# loops over every mirror copy of the affected SKILL.md so drift across
# scaffold + plugin manifests cannot silently re-introduce the regression.
#
# Usage:
#   bash test/skills/test_skill_invariants.sh
#
# Exit code: 0 = all checks pass, 1 = at least one check failed.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

# ---------- color helpers ----------
if [ -t 1 ] && command -v tput >/dev/null 2>&1; then
  C_RED="$(tput setaf 1)"; C_GREEN="$(tput setaf 2)"; C_YELLOW="$(tput setaf 3)"
  C_BOLD="$(tput bold)"; C_RESET="$(tput sgr0)"
else
  C_RED=""; C_GREEN=""; C_YELLOW=""; C_BOLD=""; C_RESET=""
fi

ok()   { printf "%sOK%s   %s\n"   "$C_GREEN" "$C_RESET" "$1"; }
fail() { printf "%sFAIL%s %s\n"   "$C_RED"   "$C_RESET" "$1"; }
note() { printf "%s...%s  %s\n"   "$C_YELLOW" "$C_RESET" "$1"; }

# ---------- mirror inventories ----------
LAND_MIRRORS=(
  ".github/skills/land/SKILL.md"
  "pkg/scaffold/templates/skills/land/SKILL.md"
  "plugins/atv-skill-land/skills/land/SKILL.md"
  "plugins/atv-pack-shipping/skills/land/SKILL.md"
  "plugins/atv-everything/skills/land/SKILL.md"
)

TAKEOFF_MIRRORS=(
  ".github/skills/takeoff/SKILL.md"
  "pkg/scaffold/templates/skills/takeoff/SKILL.md"
  "plugins/atv-skill-takeoff/skills/takeoff/SKILL.md"
  "plugins/atv-pack-shipping/skills/takeoff/SKILL.md"
  "plugins/atv-everything/skills/takeoff/SKILL.md"
)

GHCP_FILE=".github/skills/ghcp-review-resolve/SKILL.md"

FAILURES=0

# _behavioral_step8_check <file-label> <guard-block>
# Runs the extracted Step 8 push-verification block against the four real git
# states and asserts the exact exit code for each. Returns 0 if all four match,
# 1 (with a fail message) otherwise. Executing the snippet makes the check
# immune to markup tricks that fool grep/awk (decoy fences, `if true; then
# exit 1`, indented top-level exits, single-branch exits).
_behavioral_step8_check() {
  local label="$1" block="$2"
  local root; root="$(mktemp -d "${TMPDIR:-/tmp}/step8-behav.XXXXXX")"
  # shellcheck disable=SC2064
  trap "rm -rf '$root'" RETURN
  ( # subshell so cd/traps don't leak
    set +e
    cd "$root" || exit 99
    git init -q --bare origin.git
    git init -q work && cd work
    git config user.email t@t; git config user.name t
    git remote add origin "$root/origin.git"
    echo seed > a && git add a && git -c commit.gpgsign=false commit -q -m seed

    run() { bash -c "$block" >/dev/null 2>&1; echo $?; }

    # State 1: detached HEAD -> rc 0 (non-fatal skip)
    git checkout -q --detach
    local r_detached; r_detached=$(run)
    git checkout -q -b feature 2>/dev/null

    # State 2: branch, no upstream -> rc 1
    local r_noupstream; r_noupstream=$(run)

    # State 3: branch, upstream exists, unpushed commit -> rc 1
    git push -q -u origin feature
    echo more >> a && git -c commit.gpgsign=false commit -q -am more
    local r_unpushed; r_unpushed=$(run)

    # State 4: fully pushed -> rc 0
    git push -q origin feature
    local r_pushed; r_pushed=$(run)

    [ "$r_detached" = 0 ] && [ "$r_noupstream" = 1 ] && \
      [ "$r_unpushed" = 1 ] && [ "$r_pushed" = 0 ]
    exit $?
  )
  local rc=$?
  if [ "$rc" -ne 0 ]; then
    fail "$label: Step 8 guard does not enforce the 4 states behaviorally (detached=0,no-upstream=1,unpushed=1,pushed=0)"
    return 1
  fi
  return 0
}

# ---------- R1: detached-HEAD guard on land Step 8 ----------
check_land_step8_guard() {
  printf "\n%s[R1] land Step 8 - detached-HEAD guard%s\n" "$C_BOLD" "$C_RESET"
  local hit=0
  for f in "${LAND_MIRRORS[@]}"; do
    if [ ! -f "$f" ]; then
      fail "$f: missing mirror"
      hit=1; continue
    fi

    # (a) Negative: the old unguarded substitution must NOT appear. This is the
    # exact shape that broke on detached HEAD (`origin/..HEAD`).
    if grep -qF 'git log "origin/$(git branch --show-current)..HEAD"' "$f"; then
      fail "$f: unguarded git branch --show-current inside git-log argument (breaks on detached HEAD)"
      hit=1
      continue
    fi

    # (b) Positive + BEHAVIORAL: extract the first executable ```bash fence in
    # Step 8 and actually RUN it against the four real git states, asserting the
    # exact exit code for each. Grep/awk structural checks can be fooled by
    # markup tricks (comments, decoy fences, `if true; then exit 1`, indented
    # top-level exits); executing the snippet cannot. This is the same
    # behavioral approach as detached_head_smoke.sh, folded into the invariant
    # so every mirror is proven, not just pattern-matched.
    local step8fence
    step8fence=$(awk '
      /^### Step 8/ { in8=1 }
      /^### Step 9/ { in8=0 }
      in8 && !seen && /^```bash/ { infence=1; seen=1; next }
      infence && /^```/ { infence=0 }
      infence { print }
    ' "$f")
    if [ -z "$step8fence" ]; then
      fail "$f: Step 8 has no executable \`\`\`bash guard block (heading or fence removed?)"
      hit=1
      continue
    fi
    # Keep only the push-verification `if branch=...; then ... fi` block from the
    # fence (drop `git status` and any surrounding prose lines). The block starts
    # at the `if branch=` line and ends at its matching top-level `fi`.
    local guard_block
    guard_block=$(printf '%s\n' "$step8fence" | awk '
      /^if branch="\$\(git branch --show-current\)"/ { grab=1 }
      grab { print }
      grab && /^fi[[:space:]]*$/ { exit }
    ')
    if [ -z "$guard_block" ]; then
      fail "$f: Step 8 has no top-level 'if branch=...; then ... fi' push-verification block"
      hit=1
      continue
    fi

    # Behavioral harness: run guard_block in a scratch repo across 4 states.
    if ! _behavioral_step8_check "$f" "$guard_block"; then
      hit=1
      continue
    fi

    ok "$f: Step 8 guard behaves correctly (detached=0, no-upstream=1, unpushed=1, pushed=0)"
  done
  return "$hit"
}

# ---------- R2: command -v backlog guard on takeoff Step 2 ----------
check_takeoff_backlog_guard() {
  printf "\n%s[R2] takeoff Step 2 - command -v backlog guard%s\n" "$C_BOLD" "$C_RESET"
  local hit=0
  for f in "${TAKEOFF_MIRRORS[@]}"; do
    if [ ! -f "$f" ]; then
      fail "$f: missing mirror"
      hit=1; continue
    fi
    # Positive, section-anchored and order-aware. Mirrors R1's rigor:
    #   1. Scope to the "### Step 2" section (up to the next "### " heading), so a
    #      guarded fence elsewhere in the file cannot satisfy the check.
    #   2. Within Step 2, find the bash fence that CONTAINS the backlog call
    #      (Step 2 may have more than one fence), stripping #-comment and blank
    #      lines so a commented-out invocation counts as absent.
    #   3. Require the executable `backlog sequence list --plain` line AND the
    #      `command -v backlog` guard to PRECEDE it in that fence.
    local step2backlogfence
    step2backlogfence=$(awk '
      /^### Step 2/ { in2=1; next }
      in2 && /^### / { in2=0 }
      in2 && /^```bash/ { infence=1; buf=""; next }
      infence && /^```/ {
        infence=0
        if (buf ~ /backlog sequence list --plain/) { printf "%s", buf }
        buf=""
        next
      }
      infence {
        line=$0; sub(/^[[:space:]]+/, "", line)
        if (line ~ /^#/ || line == "") next   # drop comment/blank lines
        buf = buf $0 "\n"
      }
    ' "$f")
    if [ -z "$step2backlogfence" ]; then
      fail "$f: no executable 'backlog sequence list --plain' in a Step 2 bash fence (guard removed?)"
      hit=1
      continue
    fi
    local cmd_line guard_line
    cmd_line=$(printf '%s' "$step2backlogfence" | grep -nF 'backlog sequence list --plain' | head -1 | cut -d: -f1)
    guard_line=$(printf '%s' "$step2backlogfence" | grep -nE 'command -v backlog' | head -1 | cut -d: -f1)
    if [ -z "$guard_line" ]; then
      fail "$f: Step 2 'backlog sequence list --plain' is not guarded by 'command -v backlog'"
      hit=1
      continue
    elif [ "$guard_line" -ge "$cmd_line" ]; then
      fail "$f: Step 2 'command -v backlog' guard does not precede the backlog invocation (guard line $guard_line, call line $cmd_line)"
      hit=1
      continue
    fi

    # BEHAVIORAL: run the Step 2 backlog fence with `backlog` ABSENT from PATH.
    # A correctly guarded fence takes the else branch and exits 0 (or at least
    # does NOT hit 127 command-not-found). This proves the guard actually
    # protects the call rather than merely appearing before it — markup that
    # looks ordered but leaves the call reachable will surface here.
    local backlog_rc
    backlog_rc=$(
      # Minimal PATH with just the shell utilities awk/grep/etc. resolve to,
      # but WITHOUT any real `backlog`. Use a scratch dir as the only PATH entry
      # plus /usr/bin:/bin for coreutils; ensure no `backlog` shim exists.
      _bp="$(mktemp -d "${TMPDIR:-/tmp}/nobacklog.XXXXXX")"
      PATH="/usr/bin:/bin" bash -c "command -v backlog >/dev/null 2>&1 && exit 42; $step2backlogfence" >/dev/null 2>&1
      echo $?
      rm -rf "$_bp"
    )
    if [ "$backlog_rc" = 42 ]; then
      # A real `backlog` exists on this runner's /usr/bin:/bin — skip the
      # absence assertion (can't simulate absence), the static guard above stands.
      ok "$f: Step 2 backlog invocation guarded (static; live backlog present, absence not simulated)"
    elif [ "$backlog_rc" = 127 ]; then
      fail "$f: Step 2 fence hits command-not-found (127) when backlog is absent — the guard does not protect the call"
      hit=1
    else
      ok "$f: Step 2 backlog invocation guarded — runs cleanly (rc $backlog_rc) with backlog absent"
    fi
  done
  return "$hit"
}

# ---------- R3: --json state + early-stop guard on section 0c of ghcp-review-resolve ----------
check_ghcp_state_guard() {
  printf "\n%s[R3] ghcp-review-resolve 0c - state field + CLOSED|MERGED guard%s\n" "$C_BOLD" "$C_RESET"
  local hit=0
  if [ ! -f "$GHCP_FILE" ]; then
    fail "$GHCP_FILE: missing file"
    return 1
  fi

  local section_line
  section_line=$(grep -nF 'gh pr view "$PR_NUMBER" --json' "$GHCP_FILE" | head -1 | cut -d: -f1)
  if [ -z "$section_line" ]; then
    fail "$GHCP_FILE: section 0c gh pr view invocation not found"
    hit=1
  else
    local window
    window=$(sed -n "${section_line},$((section_line+5))p" "$GHCP_FILE")
    if echo "$window" | grep -qE '(^|,|[[:space:]])state(,|[[:space:]]|$)'; then
      ok "$GHCP_FILE:$section_line: state present in section 0c --json field list"
    else
      fail "$GHCP_FILE:$section_line: state MISSING from section 0c --json field list"
      hit=1
    fi
  fi

  # 3b: a CLOSED|MERGED early-stop guard must exist in the prose. Accept either:
  #   - both tokens on the same line, OR
  #   - both tokens within 3 lines of each other.
  if grep -nE 'CLOSED|MERGED' "$GHCP_FILE" | head -20 | awk -F: '
    {
      # Single line containing both tokens — invariant satisfied immediately.
      if ($0 ~ /CLOSED/ && $0 ~ /MERGED/) { ok=1; exit }
      seen[$1]++; line[$1]=$0
    }
    END {
      if (ok) exit 0
      for (a in seen) for (b in seen)
        if (a != b && (b - a > 0 && b - a <= 3)) {
          if (line[a] ~ /CLOSED/ && line[b] ~ /MERGED/) exit 0
          if (line[a] ~ /MERGED/ && line[b] ~ /CLOSED/) exit 0
        }
      exit 1
    }'; then
    ok "$GHCP_FILE: CLOSED/MERGED early-stop guard prose present"
  else
    fail "$GHCP_FILE: no co-located CLOSED/MERGED guard prose found"
    hit=1
  fi

  return "$hit"
}

# ---------- R4: /ce: -> /ce- in takeoff mirrors ----------
check_takeoff_ce_command_form() {
  printf "\n%s[R4] takeoff - /ce- (hyphen) command form, no /ce: colon form%s\n" "$C_BOLD" "$C_RESET"
  local hit=0
  for f in "${TAKEOFF_MIRRORS[@]}"; do
    if [ ! -f "$f" ]; then
      fail "$f: missing mirror"
      hit=1; continue
    fi
    local bad
    bad=$(grep -nE '/ce:(work|plan|ideate)\b' "$f" || true)
    if [ -n "$bad" ]; then
      while IFS= read -r line; do
        fail "$f:$line"
      done <<< "$bad"
      hit=1
    else
      ok "$f: no /ce: colon-form command references"
    fi
  done
  return "$hit"
}

# ---------- main ----------
# NOTE: intentionally no absolute-path banner here. Printing $REPO_ROOT (or even
# its basename, which varies by worktree/checkout name) leaked a local path into
# committed screenshot artifacts and made those artifacts non-reproducible.
# Paths in check output are always repo-relative.
printf "%sMirror-aware SKILL.md invariant harness%s\n" "$C_BOLD" "$C_RESET"
note "Each check loops over all mirror copies and returns non-zero on first defect"

if ! check_land_step8_guard; then FAILURES=$((FAILURES+1)); fi
if ! check_takeoff_backlog_guard; then FAILURES=$((FAILURES+1)); fi
if ! check_ghcp_state_guard; then FAILURES=$((FAILURES+1)); fi
if ! check_takeoff_ce_command_form; then FAILURES=$((FAILURES+1)); fi

printf "\n%s---- Summary ----%s\n" "$C_BOLD" "$C_RESET"
if [ "$FAILURES" -eq 0 ]; then
  printf "%sAll 4 invariant checks passed.%s\n" "$C_GREEN" "$C_RESET"
  exit 0
else
  printf "%s%d invariant check(s) failed.%s\n" "$C_RED" "$FAILURES" "$C_RESET"
  exit 1
fi
