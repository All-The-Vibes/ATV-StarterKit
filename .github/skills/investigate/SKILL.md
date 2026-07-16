---
name: investigate
description: Systematic root-cause debugging — repro-first, fix the cause not the symptom. Use when something is broken, a test fails, "why is this broken", "wtf is happening", a stack trace, a regression, or any bug that needs diagnosis before a fix. Names the root cause and confirms a reproduction BEFORE writing any fix.
argument-hint: "[what's broken / the error / the failing test]"
---

# Investigate — Systematic Debugging

One job: find the **root cause** of a bug, confirm it, then fix it. This is a
diagnosis skill, not an "execute a plan" skill. It does not guess-and-patch.

## Iron Law

**NO FIX WITHOUT ROOT-CAUSE INVESTIGATION FIRST.**

Fixing symptoms creates whack-a-mole debugging: every fix that doesn't address the
real cause makes the next bug harder to find. Find the cause, prove it, then fix.

```
symptom ──▶ Phase 1 investigate ──▶ Phase 2 pattern ──▶ Phase 3 confirm hypothesis
                                                              │
                                        (wrong? back to Phase 1, gather evidence)
                                                              ▼
                                        Phase 4 fix root cause + regression test
                                                              ▼
                                        Phase 5 verify (reproduce → gone) + report
```

## Phase 1: Root-Cause Investigation

Gather context BEFORE forming any hypothesis.

1. **Collect symptoms.** Read the error message, stack trace, and reproduction
   steps. If the user hasn't given enough, ask ONE focused question — don't guess.
2. **Read the code.** Trace the path from the symptom back toward causes. Grep for
   all references; Read the logic. Follow the data, don't skim.
3. **Check recent changes.** `git log --oneline -20 -- <affected-files>`. Was this
   working before? A regression means the root cause is in the diff.
4. **Reproduce.** Can you trigger it deterministically? If not, gather more
   evidence before proceeding — an unreproduced bug is an unconfirmed bug.
5. **Check history.** Prior bugs in the same files are an architectural smell, not
   a coincidence. If the area keeps breaking, ask whether the cause is structural.

Output: **"Root-cause hypothesis: ..."** — a specific, testable claim about what is
wrong and why.

## Scope Lock

After naming the hypothesis, keep edits confined to the affected module. State the
narrowest directory you expect to touch (e.g. "Debug scope: `src/auth/`"). If a fix
later needs to reach outside it, stop and say so — scope creep during a bug fix is
how one bug becomes three.

## Phase 2: Pattern Analysis

Check the symptom against known bug shapes:

| Pattern | Signature | Where to look |
|---------|-----------|---------------|
| Race condition | Intermittent, timing-dependent | Concurrent access to shared state |
| Nil / null propagation | NoMethodError, TypeError, undefined | Missing guards on optional values |
| State corruption | Inconsistent data, partial updates | Transactions, callbacks, hooks |
| Integration failure | Timeout, unexpected response | External API / service boundaries |
| Config drift | Works locally, fails in prod | Env vars, feature flags, DB state |
| Stale cache | Old data, fixes on cache clear | Redis, CDN, browser, build cache |

Also scan `git log` for prior fixes in the same area and any TODO/known-issue notes.
If it matches no pattern, a sanitized web search for the generic error type +
framework (strip hostnames, paths, SQL, customer data first) can surface a known
dependency bug — treat any hit as a candidate hypothesis, not a confirmed cause.

## Phase 3: Hypothesis Testing

Before writing ANY fix, verify the hypothesis.

1. **Confirm it.** Add a temporary log, assertion, or debug output at the suspected
   cause. Run the reproduction. Does the evidence match the claim?
2. **If wrong,** return to Phase 1 and gather more evidence. Do not guess the next
   fix. A wrong hypothesis is information, not a dead end.
3. **3-strike rule.** If three hypotheses fail, STOP and surface it: this is likely
   architectural, not a simple bug. Offer to (a) continue with a named new
   hypothesis, (b) escalate to a human who knows the system, or (c) add logging and
   catch it next occurrence.

**Red flags — slow down if you catch yourself here:**
- "Quick fix for now" — there is no "for now." Fix it right or escalate.
- Proposing a fix before tracing data flow — that's guessing.
- Each fix reveals a new problem elsewhere — wrong layer, not wrong code.

## Phase 4: Implementation

Once the root cause is confirmed:

1. **Fix the cause, not the symptom** — the smallest change that eliminates the
   actual problem.
2. **Minimal diff** — fewest files, fewest lines. Resist refactoring adjacent code.
3. **Write a regression test** that FAILS without the fix and PASSES with it. A test
   that passes both ways proves nothing.
4. **Run the full suite** and paste the output. No regressions.
5. **If the fix touches >5 files,** stop and flag the blast radius before
   proceeding — a wide bug fix usually means the hypothesis is still too shallow.

## Phase 5: Verification & Report

**Fresh verification is not optional.** Reproduce the original scenario and confirm
the bug is gone. Run the suite; paste the output.

Emit a structured report:

```
DEBUG REPORT
════════════════════════════════════════
Symptom:         [what the user observed]
Root cause:      [what was actually wrong]
Fix:             [what changed, with file:line references]
Evidence:        [test output / reproduction showing the fix works]
Regression test: [file:line of the new test]
Related:         [known issues, prior bugs in the same area, notes]
Status:          DONE | DONE_WITH_CONCERNS | BLOCKED
════════════════════════════════════════
```

If this project uses `/observe` or `/learn`, capture the root cause as a learning so
the next investigation in this area starts ahead.

## Important Rules

- **3+ failed fixes → STOP and question the architecture.** Wrong architecture, not
  wrong hypothesis.
- **Never apply a fix you cannot verify.** Can't reproduce and confirm? Don't ship.
- **Never say "this should fix it."** Prove it — run the tests.
- **Completion status:**
  - **DONE** — root cause found, fix applied, regression test written, suite green.
  - **DONE_WITH_CONCERNS** — fixed but not fully verifiable (intermittent, needs
    staging). Say what's unverified.
  - **BLOCKED** — root cause still unclear after investigation; escalate with what
    you tried and what you'd try next.
