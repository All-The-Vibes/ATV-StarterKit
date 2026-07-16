# test/skills/

Tests for invariants in `SKILL.md` files (the skills consumed by `/land`, `/takeoff`, `/ghcp-review-resolve`, and their mirror copies under `pkg/scaffold/templates/skills/` and `plugins/atv-*/skills/`).

## Run

```bash
bash test/skills/test_skill_invariants.sh
```

Exit code: `0` if all checks pass, `1` if any invariant is broken.

## What's checked

| ID | Invariant | Mirrors |
|----|-----------|---------|
| R1 | `land/SKILL.md` Step 8 guards `git log "origin/$(git branch --show-current)..HEAD"` so detached HEAD or missing-upstream does not error. | 5 (`land/`) |
| R2 | `takeoff/SKILL.md` Step 2 guards `backlog sequence list --plain` behind `command -v backlog`. | 5 (`takeoff/`) |
| R3 | `.github/skills/ghcp-review-resolve/SKILL.md` §0c includes `state` in the `gh pr view --json` field list **and** has a co-located `CLOSED`/`MERGED` early-stop guard. | 1 (`ghcp-review-resolve/`) |
| R4 | `takeoff/SKILL.md` references `/ce-work`, `/ce-plan`, `/ce-ideate` (hyphen form). No `/ce:work`, `/ce:plan`, `/ce:ideate` (colon form). | 5 (`takeoff/`) |

Each check loops the full mirror list. Adding a new mirror later means updating one array in `test_skill_invariants.sh`.

## Origin

Plan: `docs/plans/2026-05-14-002-fix-pr42-deferred-skill-fixes-plan.md`
Original deferred review threads: PR #42 (merged).
