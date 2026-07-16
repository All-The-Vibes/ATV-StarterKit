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

    # (a) Negative: the old unguarded substitution must NOT appear. This is the
    # exact shape that broke on detached HEAD (`origin/..HEAD`).
    if grep -qF 'git log "origin/$(git branch --show-current)..HEAD"' "$f"; then
      fail "$f: unguarded git branch --show-current inside git-log argument (breaks on detached HEAD)"
      hit=1
      continue
    fi

    # (b) Positive: the REPLACEMENT guard must be present in the executable
    # ```bash fence INSIDE Step 8 — not merely somewhere in the file, and not in
    # prose/comments. Extract the first ```bash ... ``` fence that falls between
    # the "### Step 8" and "### Step 9" headings, then require all structural
    # markers of the push-verification guard within that fenced snippet:
    #   - the branch capture guard,
    #   - the upstream existence check,
    #   - a BLOCKED failure that stops the routine (exit 1) on unpushed/no-upstream.
    # Scoping to the fenced block means a stray marker in prose, a comment, or a
    # different step cannot mask a deleted guard.
    local step8fence
    step8fence=$(awk '
      /^### Step 8/ { in8=1 }
      /^### Step 9/ { in8=0 }
      in8 && /^```bash/ { infence=1; next }
      in8 && infence && /^```/ { infence=0 }
      in8 && infence { print }
    ' "$f")
    if [ -z "$step8fence" ]; then
      fail "$f: Step 8 has no executable \`\`\`bash guard block (heading or fence removed?)"
      hit=1
      continue
    fi
    local missing=""
    printf '%s\n' "$step8fence" | grep -qF 'branch="$(git branch --show-current)"' || missing="$missing branch-guard"
    printf '%s\n' "$step8fence" | grep -qF 'git rev-parse --verify --quiet "refs/remotes/origin/$branch"' || missing="$missing upstream-check"
    printf '%s\n' "$step8fence" | grep -qE 'BLOCKED:.*(push before landing|push the branch before landing)' || missing="$missing blocked-notice"
    printf '%s\n' "$step8fence" | grep -qE '^[[:space:]]*exit 1' || missing="$missing hard-exit"
    if [ -n "$missing" ]; then
      fail "$f: Step 8 push-verification guard missing marker(s):$missing (guard removed, not just old string absent)"
      hit=1
      continue
    fi

    ok "$f: Step 8 guard present (branch guard + upstream check + BLOCKED hard-exit)"
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
    # Positive: the guarded `backlog sequence list --plain` invocation must be
    # present inside an executable ```bash fence, guarded by `command -v backlog`
    # in the SAME fence. Absence is a FAIL (deleting the snippet must not pass),
    # and a match in prose or an unrelated `command -v backlog` elsewhere cannot
    # satisfy the guard.
    if ! grep -qF 'backlog sequence list --plain' "$f"; then
      fail "$f: guarded 'backlog sequence list --plain' invocation missing (guard removed, not just old string absent)"
      hit=1
      continue
    fi
    # Extract each ```bash ... ``` fence and check whether ANY fence contains
    # both the invocation and its command -v backlog guard together.
    local guarded
    guarded=$(awk '
      /^```bash/ { infence=1; buf=""; next }
      infence && /^```/ {
        infence=0
        if (buf ~ /backlog sequence list --plain/ && buf ~ /command -v backlog/) { print "GUARDED" }
        else if (buf ~ /backlog sequence list --plain/) { print "UNGUARDED" }
        buf=""
        next
      }
      infence { buf = buf "\n" $0 }
    ' "$f")
    if printf '%s\n' "$guarded" | grep -q 'UNGUARDED'; then
      fail "$f: 'backlog sequence list --plain' appears in a \`\`\`bash fence without a command -v backlog guard in the same fence"
      hit=1
    elif printf '%s\n' "$guarded" | grep -q 'GUARDED'; then
      ok "$f: backlog invocation guarded by command -v backlog (same fence)"
    else
      fail "$f: 'backlog sequence list --plain' not found inside an executable \`\`\`bash fence"
      hit=1
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
printf "Repo: %s\n" "$(basename "$REPO_ROOT")"
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
