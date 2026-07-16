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

# Cleanup helper referenced by the R1 RETURN trap. Reads the dir from a global
# so the trap body never embeds a raw (possibly quote-containing) path.
_R1_CLEANUP_DIR=""
_r1_cleanup() { [ -n "$_R1_CLEANUP_DIR" ] && rm -rf "$_R1_CLEANUP_DIR"; _R1_CLEANUP_DIR=""; }

# _behavioral_step8_check <file-label> <fence-code>
# Runs the ENTIRE extracted Step 8 bash fence against the four real git states
# and asserts the exact exit code for each. Executing the whole fence (not a
# sliced-out sub-block) means top-level mutations before/after the guard —
# `exit 0`, `if true; then exit 1`, stray commands — are actually run and
# therefore caught. Returns 0 iff all four states produce the required rc.
_behavioral_step8_check() {
  local label="$1" fence="$2"
  local root
  root="$(mktemp -d "${TMPDIR:-/tmp}/step8-behav.XXXXXX")" || {
    fail "$label: could not create scratch dir for behavioral Step 8 check (mktemp failed)"
    return 1
  }
  # Fail-fast if the scratch dir is not a real, non-empty path we can enter.
  if [ -z "$root" ] || [ ! -d "$root" ]; then
    fail "$label: scratch dir invalid for behavioral Step 8 check"
    rm -rf "$root" 2>/dev/null
    return 1
  fi
  # Clean up on any return path. Use a variable-referencing trap body (not the
  # raw path embedded in the string) so a scratch path containing a single quote
  # or other shell metachar can never break the trap syntax.
  _R1_CLEANUP_DIR="$root"
  trap '_r1_cleanup' RETURN

  local rc
  (
    set +e
    cd "$root" || exit 99          # never proceed in the real repo
    [ "$PWD" = "$root" ] || exit 99
    git init -q --bare origin.git || exit 99
    git init -q work || exit 99
    cd work || exit 99
    git config user.email t@t; git config user.name t
    git remote add origin "$root/origin.git" || exit 99
    echo seed > a && git add a && git -c commit.gpgsign=false commit -q -m seed || exit 99

    # Run the whole fence. `git status` etc. in the fence are harmless here.
    run() { ( bash -c "$fence" ) >/dev/null 2>&1; echo $?; }

    git checkout -q --detach
    local rd; rd=$(run)                       # detached -> 0
    git checkout -q -b feature 2>/dev/null
    local rn; rn=$(run)                        # no upstream -> 1
    git push -q -u origin feature
    echo more >> a && git -c commit.gpgsign=false commit -q -am more
    local ru; ru=$(run)                        # unpushed -> 1
    git push -q origin feature
    local rp; rp=$(run)                        # pushed -> 0

    [ "$rd" = 0 ] && [ "$rn" = 1 ] && [ "$ru" = 1 ] && [ "$rp" = 0 ]
    exit $?
  )
  rc=$?
  if [ "$rc" -ne 0 ]; then
    fail "$label: Step 8 fence does not enforce the 4 states behaviorally (detached=0,no-upstream=1,unpushed=1,pushed=0)"
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

    # (b) Positive + BEHAVIORAL: extract the Step 8 ```bash fence and RUN THE
    # WHOLE FENCE against the four real git states, asserting the exact exit code
    # for each. Executing the entire fence (not a sliced-out sub-block) means
    # any top-level mutation before/after the guard — a stray `exit 0`, an
    # `if true; then exit 1`, an extra command — is actually executed and thus
    # caught. Grep/awk structural checks cannot achieve this; execution can.
    # Also assert there is exactly ONE Step 8 bash fence, so a decoy second
    # fence cannot hide behavior.
    local fence_count
    fence_count=$(awk '
      /^### Step 8/ { in8=1 }
      /^### Step 9/ { in8=0 }
      in8 && /^```bash/ { n++ }
      END { print n+0 }
    ' "$f")
    if [ "$fence_count" != 1 ]; then
      fail "$f: Step 8 must contain exactly one \`\`\`bash fence (found $fence_count)"
      hit=1
      continue
    fi
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

    # Behavioral harness: run the WHOLE Step 8 fence in a scratch repo, 4 states.
    if ! _behavioral_step8_check "$f" "$step8fence"; then
      hit=1
      continue
    fi

    ok "$f: Step 8 fence behaves correctly (detached=0, no-upstream=1, unpushed=1, pushed=0)"
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
    cmd_line=$(printf '%s' "$step2backlogfence" | grep -nF 'backlog sequence list --plain' | head -1 | cut -d: -f1 || true)
    guard_line=$(printf '%s' "$step2backlogfence" | grep -nE 'command -v backlog' | head -1 | cut -d: -f1 || true)
    if [ -z "$guard_line" ]; then
      fail "$f: Step 2 'backlog sequence list --plain' is not guarded by 'command -v backlog'"
      hit=1
      continue
    elif [ "$guard_line" -ge "$cmd_line" ]; then
      fail "$f: Step 2 'command -v backlog' guard does not precede the backlog invocation (guard line $guard_line, call line $cmd_line)"
      hit=1
      continue
    fi

    # BEHAVIORAL + HERMETIC, two-sided, redirection-proof.
    #
    # (i) ABSENCE case — prove the call is NOT reached when backlog is missing.
    #     Instead of grepping stderr (which redirection, backgrounding, traps, or
    #     a shadowing function can defeat, and which is locale-dependent), install
    #     a `command_not_found_handle`: bash invokes it at exec-resolution time
    #     for ANY unresolved command, BEFORE stderr is written and regardless of
    #     how the caller redirects or backgrounds. The handle writes a sentinel
    #     file when `backlog` is the unresolved command. If the sentinel appears,
    #     the guard did not protect the call.
    #
    # (ii) PRESENCE case — prove the call IS actually executed when backlog
    #     exists (so a heredoc/printf/eval literal that merely CONTAINS the text
    #     but never runs it cannot pass). A stub `backlog` on PATH records each
    #     invocation; the else-branch fallback must run it exactly once.
    local probe pass=1
    probe="$(mktemp -d "${TMPDIR:-/tmp}/r2-probe.XXXXXX")" || {
      fail "$f: could not create probe dir for behavioral R2 check (mktemp failed)"
      hit=1; continue
    }
    if [ -z "$probe" ] || [ ! -d "$probe" ]; then
      fail "$f: probe dir invalid for behavioral R2 check"
      rm -rf "$probe" 2>/dev/null; hit=1; continue
    fi

    # (i) absence: prove the documented guard does NOT reach the call when
    # `command -v backlog` reports absent. A RECORDING STUB named `backlog` sits
    # on PATH (so it is reachable by any exec path — direct, `env backlog`,
    # `sh -c`, `bash -c`, a wrapper), but the shell's `command` builtin is
    # shadowed to report `backlog` as absent. A correct guard consults
    # `command -v backlog`, sees "absent", and takes the else branch — never
    # touching the stub. Any wrapper that reaches `backlog` regardless of the
    # guard trips the stub, which writes a sentinel. Run under `env -i` +
    # --noprofile --norc so no inherited function/alias/PATH perturbs it.
    printf '#!/usr/bin/env bash\n: > "%s/reached"\nexit 0\n' "$probe" > "$probe/backlog"
    chmod +x "$probe/backlog"
    env -i PATH="$probe:/usr/bin:/bin" bash --noprofile --norc -c '
      command() { if [ "$1" = -v ] && [ "$2" = backlog ]; then return 1; fi; builtin command "$@"; }
      '"$step2backlogfence"'
    ' >/dev/null 2>&1 || true
    if [ -f "$probe/reached" ]; then
      fail "$f: Step 2 reaches backlog even though 'command -v backlog' reports absent — the invocation is not actually guarded (wrapper/direct exec bypass)"
      hit=1; pass=0
    fi
    rm -f "$probe/reached" "$probe/calls"

    # (ii) presence: with `command -v backlog` truthful and a logging stub on
    # PATH, the guarded fallback must invoke it EXACTLY once — proving the
    # `backlog sequence list --plain` text is a real, reached command (a
    # heredoc/printf/eval/variable literal that never executes fails here).
    printf '#!/usr/bin/env bash\necho call >> "%s/calls"\nexit 0\n' "$probe" > "$probe/backlog"
    chmod +x "$probe/backlog"
    env -i PATH="$probe:/usr/bin:/bin" bash --noprofile --norc -c "$step2backlogfence" >/dev/null 2>&1 || true
    local calls=0
    [ -f "$probe/calls" ] && calls=$(wc -l < "$probe/calls")
    if [ "$calls" -lt 1 ]; then
      fail "$f: Step 2 'backlog sequence list --plain' text is present but never actually executed (heredoc/printf/eval literal?) — no real invocation"
      hit=1; pass=0
    elif [ "$calls" -gt 1 ]; then
      fail "$f: Step 2 fence invokes backlog $calls times (expected exactly once via the guarded fallback)"
      hit=1; pass=0
    fi

    rm -rf "$probe"
    if [ "$pass" = 1 ]; then
      ok "$f: Step 2 backlog invocation guarded (behaviorally: not reached when absent, executed once when present)"
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
