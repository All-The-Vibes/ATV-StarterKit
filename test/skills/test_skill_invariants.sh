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

# ---------- R1: detached-HEAD guard on land Step 8 ----------
check_land_step8_guard() {
  printf "\n%s[R1] land Step 8 - detached-HEAD guard%s\n" "$C_BOLD" "$C_RESET"
  local hit=0
  for f in "${LAND_MIRRORS[@]}"; do
    if [ ! -f "$f" ]; then
      fail "$f: missing mirror"
      hit=1; continue
    fi
    local match
    match=$(grep -nF 'git log "origin/$(git branch --show-current)..HEAD"' "$f" || true)
    if [ -z "$match" ]; then
      ok "$f: no unguarded git-log substitution"
      continue
    fi
    local lineno
    lineno=$(echo "$match" | head -1 | cut -d: -f1)
    # Walk upward until we hit a blank line or a section boundary, then check
    # whether an `if ...` appears inside that block. Avoids hard-coding the
    # wrapper size; the guard can grow to N lines without false-failing.
    local upper
    upper=$(awk -v n="$lineno" '
      NR < n { buf[NR]=$0 }
      NR == n {
        for (i = NR-1; i >= 1; i--) {
          if (buf[i] ~ /^[[:space:]]*$/) { print i+1; exit }
        }
        print 1; exit
      }' "$f")
    if sed -n "${upper},$((lineno-1))p" "$f" | grep -qE '^[[:space:]]*if '; then
      ok "$f:$lineno: substitution present but appears guarded by an if above"
    else
      fail "$f:$lineno: unguarded git branch --show-current inside git-log argument (breaks on detached HEAD)"
      hit=1
    fi
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
    if grep -qF 'backlog sequence list --plain' "$f"; then
      local lineno
      lineno=$(grep -nF 'backlog sequence list --plain' "$f" | head -1 | cut -d: -f1)
      # Scope the guard search to a ±6-line window around the matched line so
      # an unrelated `command -v backlog` elsewhere in the file cannot mask a
      # nearby unguarded invocation.
      local win_lower=$((lineno - 6)); [ "$win_lower" -lt 1 ] && win_lower=1
      local win_upper=$((lineno + 6))
      if sed -n "${win_lower},${win_upper}p" "$f" | grep -qF 'command -v backlog'; then
        ok "$f:$lineno: backlog invocation guarded by command -v backlog"
      else
        fail "$f:$lineno: backlog sequence list --plain invoked without command -v backlog guard"
        hit=1
      fi
    else
      ok "$f: backlog invocation either absent or guarded"
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
printf "%sMirror-aware SKILL.md invariant harness%s\n" "$C_BOLD" "$C_RESET"
printf "Repo: %s\n" "$REPO_ROOT"
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
