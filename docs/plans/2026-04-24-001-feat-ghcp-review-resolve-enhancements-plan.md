---
title: "feat: ghcp-review-resolve — port pr-review-toolkit locally, resolve Copilot threads, and tick PR task-list items"
type: feat
status: active
date: 2026-04-24
---

# feat: ghcp-review-resolve — port pr-review-toolkit locally, resolve Copilot threads, and tick PR task-list items

## Enhancement Summary

**Deepened on:** 2026-04-24
**Sections enhanced:** A (port), B (resolve), C (tick), Technical Considerations, Dependencies & Risks
**Research agents used:** `best-practices-researcher` (external: GitHub GraphQL, `gh pr edit`, Apache-2.0, fuzzy-match tradeoffs), `repo-research-analyst` (local: 57 SKILL.md files, 51 agent files, invocation conventions)

### Key Improvements
1. **Resolve guardrail upgraded** — check `thread.viewerCanResolve` before calling the mutation; inspect the GraphQL `errors[]` array (not just HTTP status) for failure detection.
2. **PR body edit made safe** — no optimistic concurrency exists on the GitHub API, so plan C switches from "hash the body" to a `<!-- BEGIN:ghcp-audit v1 -->` ... `<!-- END:ghcp-audit v1 -->` sentinel fence + strict re-fetch-before-write. 64 KiB body ceiling flagged.
3. **Fuzzy-match downsized** — at this scale (<20 items × <10 findings) LLM matching is overkill. Switch to normalized substring + rapidfuzz `token_set_ratio ≥ 85`, LLM only as fallback for ambiguity. Unmatched findings become "unlinked — needs manual triage" rather than being silently dropped.
4. **Port conventions locked down** — confirmed only `name` + `description` are required in SKILL frontmatter, only `description` in agents; scripts invoked via `bash ${CLAUDE_PLUGIN_ROOT}/skills/<name>/scripts/...` with no exec bit required.
5. **Attribution scheme chosen** — upstream has no `NOTICE`, so we don't need one either. Per-file HTML-comment header (not a "Source:" line) in each ported file + `LICENSE` inside `.github/skills/pr-review-toolkit/`. Establishes the first attribution pattern in the repo.
6. **Agent collision avoided** — renaming `pr-code-simplifier` → `pr-simplification-analyzer` to avoid semantic overlap with existing `code-simplicity-reviewer.agent.md`. Adding a disambiguating `description:` on `pr-comment-analyzer` to distinguish from sibling `pr-comment-resolver.agent.md`.

### New Considerations Discovered
- **`resolveReviewThread` idempotency is not formally documented** — community reports indicate re-resolving is a no-op success, but the plan should treat a second call as "verify, not rely on". Ambiguity flagged, not blocking.
- **PR body 64 KiB limit** is de-facto (not in REST reference). Plan 7.0 now refuses to write if the audited body would exceed 60 KiB, leaving margin.
- **CRLF → LF normalization** happens server-side, so `--body-file` is safe cross-platform.
- **`${CLAUDE_PLUGIN_ROOT}` convention** is used by existing `resolve-pr-parallel` scripts — ported scripts should follow suit if any are added.

## Overview

The existing `ghcp-review-resolve` skill orchestrates a dual-reviewer pipeline (GitHub Copilot + Anthropic's `pr-review-toolkit`), then fixes verified findings. In practice today:

1. `pr-review-toolkit` is referenced as if installed, but **it is not vendored in this repo** — Step 1b calls a skill that doesn't exist, so the pipeline silently degrades to single-reviewer mode (as observed on PR #23).
2. After fixing a finding, the skill commits + replies on the review thread but **never marks the Copilot thread `resolved`**, so the PR stays visually cluttered with open "unresolved" threads even after fixes land.
3. PR descriptions frequently contain GitHub task-list items (`- [ ] ...`) that correspond to the work the skill just did, but **the skill never edits the PR body to tick them off**, leaving the PR looking incomplete.

This plan ports `pr-review-toolkit` into `.github/skills/` + `.github/agents/` following this repo's conventions, and extends `ghcp-review-resolve` with two new post-fix actions: (a) resolve each addressed Copilot review thread via GraphQL mutation, and (b) update the PR body to check off task-list items that map to the work done this run.

## Problem Statement / Motivation

- **Missing dependency.** The `pr-review-toolkit` skill is the second reviewer in the pipeline. Without it present locally, `ghcp-review-resolve` permanently runs in single-reviewer mode — losing the whole "overlap = high confidence" signal the skill was designed around. Vendoring it once here removes the external-install friction and makes the dual-review path actually reachable.
- **Thread hygiene.** After a real fix has been pushed, a reviewer (human or bot) reading the PR sees N unresolved threads and has to manually resolve each one to know what's still outstanding. The skill already knows which threads it addressed — it should resolve them.
- **Task-list completion.** PR descriptions in this repo commonly include acceptance-criteria checklists (see the plan templates in `.github/skills/ce-plan/SKILL.md`). When `ghcp-review-resolve` satisfies one of those items via a code change, leaving the box unchecked hides real progress. Automating the tick keeps the PR body honest and readable.

## Proposed Solution

Three coordinated changes, each landable independently:

### A. Port `pr-review-toolkit` into this repo

Vendor Anthropic's `pr-review-toolkit` plugin (`https://github.com/anthropics/claude-plugins-official/tree/main/plugins/pr-review-toolkit`) as a native ATV skill + agents.

Layout:

```
.github/
├── skills/
│   └── pr-review-toolkit/
│       ├── SKILL.md             # Adapted from upstream commands/review-pr.md + README.md
│       └── LICENSE              # Apache-2.0 text, copied verbatim from upstream
└── agents/
    ├── pr-comment-analyzer.agent.md         # sibling to pr-comment-resolver — description disambiguates
    ├── pr-test-analyzer.agent.md
    ├── pr-silent-failure-hunter.agent.md
    ├── pr-type-design-analyzer.agent.md
    ├── pr-code-reviewer.agent.md
    └── pr-simplification-analyzer.agent.md  # renamed from upstream code-simplifier to avoid overlap with existing code-simplicity-reviewer
```

Agents get the `pr-` prefix to make it obvious they belong to the toolkit. Only `pr-comment-resolver.agent.md` currently uses `pr-`, so the namespace is clear. Two repo-specific renames/disambiguations (from research findings):

- **`pr-code-simplifier` → `pr-simplification-analyzer`** to avoid semantic overlap with existing `.github/agents/code-simplicity-reviewer.agent.md`. The `-analyzer` suffix also aligns with the other ported siblings (`pr-test-analyzer`, `pr-type-design-analyzer`).
- **`pr-comment-analyzer`** keeps its upstream-derived name but its `description:` explicitly distinguishes "analyze comment accuracy in diffs" from `pr-comment-resolver`'s "reply-and-resolve review comment threads".

### Research Insights

**Local conventions (verified against 57 SKILL.md + 51 `.agent.md` files):**
- Skill frontmatter: only `name` + `description` are universal (57/57). Optional fields used elsewhere: `license`, `allowed-tools`, `argument-hint`, `disable-model-invocation`. The port uses `name`, `description`, and `license: Apache-2.0 — see LICENSE`.
- Agent frontmatter: only `description` is universal. No `name` field. Add `user-invocable: true` only if the agent should be directly triggerable.
- Agent discovery is **filename-based** — `<name>.agent.md` where `<name>` is the identifier. No manifest.
- Script invocation (if any are added): `bash ${CLAUDE_PLUGIN_ROOT}/skills/pr-review-toolkit/scripts/<script>` with shebang `#!/usr/bin/env bash`. No exec bit required. Pattern verified in `.github/skills/resolve-pr-parallel/SKILL.md:27,64,74`.
- Cross-skill invocation: `Skill(skill="pr-review-toolkit:review-pr", args="...")` — the `:` prefix is the plugin/bundle namespace. Verified in `.github/skills/ghcp-review-resolve/SKILL.md:97,281,528`.

**Attribution (research-validated):**
- Upstream `anthropics/claude-plugins-official/plugins/pr-review-toolkit/` has a `LICENSE` (Apache-2.0) but **no `NOTICE` file**. Under §4(c), we only need to propagate a `NOTICE` if upstream has one — so we don't need one either.
- Apache-2.0 §4(a) requires a copy of the license → `.github/skills/pr-review-toolkit/LICENSE` (verbatim upstream file).
- Apache-2.0 §4(b) requires a "prominent notice" in modified files → per-file HTML-comment header (below). Not needed on files copied verbatim, but every ported file in this plan will have at least frontmatter edits, so all get the header.
- Apache-2.0 does **not** require per-file copyright headers on unmodified files. Ours are modified, so they get the header.

**Per-file attribution header** (top of each ported `.md` file, inside HTML comment so it doesn't render):

```markdown
<!--
Portions of this file are derived from anthropics/claude-plugins-official
(https://github.com/anthropics/claude-plugins-official), licensed under
the Apache License, Version 2.0. See .github/skills/pr-review-toolkit/LICENSE.

Modifications: ported to All-The-Vibes/ATV-StarterKit; frontmatter adapted
to repo conventions; agent renamed (see plan for rename rationale); 2026-04-24.
-->
```

**This port establishes the first attribution precedent in the repo** — no existing `LICENSE`/`NOTICE` files, no `SPDX-License-Identifier` headers anywhere in-tree. Keep the pattern minimal so it's easy for future ports to follow.

**References:**
- https://www.apache.org/licenses/LICENSE-2.0 (§4)
- https://infra.apache.org/licensing-howto.html
- Verified: `plugins/pr-review-toolkit/LICENSE` exists, `plugins/pr-review-toolkit/NOTICE` does not.

The skill exposes a single callable entry point:

```
Skill(skill="pr-review-toolkit:review-pr", args="<PR URL or #PR_NUMBER>")
```

which fans out to the 6 agents, posts findings as inline PR review comments on the target PR, and returns a structured list of the findings it posted (so `ghcp-review-resolve` Step 2 can detect completion without polling for minutes).

### B. Resolve Copilot threads after successful fixes

Extend Step 6 ("Inline fix loop") of `ghcp-review-resolve` with a new sub-step **6.5 — Resolve thread** that runs only after the reply-posting sub-step (6.6) succeeds.

New behavior per addressed finding:

```bash
# Resolve the thread using the existing in-repo helper, which wraps
# the resolveReviewThread GraphQL mutation.
.github/skills/resolve-pr-parallel/scripts/resolve-pr-thread "$THREAD_ID"
```

Where `$THREAD_ID` is the GraphQL node ID of the review thread (fetched in Step 0g's already-existing `reviewThreads` query — we thread it through the pipeline's per-finding record so this resolution step doesn't need an extra API call).

Guardrails on resolution:

- **Only resolve a thread if** (all five must hold):
  1. Its finding was accepted by the adjudicator (Step 4), AND
  2. The fix commit was pushed successfully (Step 6.5, the existing push step), AND
  3. The reply was posted successfully (Step 6.6), AND
  4. Verification in Step 6.3 succeeded (tests/lint passed OR the finding was a pure-prose / docs finding), AND
  5. **`thread.viewerCanResolve == true`** — query this in the 0g `reviewThreads` fetch so we can skip the mutation (and log the reason) when the running identity lacks resolve permissions. This avoids a noisy `FORBIDDEN` GraphQL error on every run when the pipeline is invoked by a user without PR-write access.
- **Do not resolve** threads for findings marked "skipped — left for human", "needs larger change", or "not independently verified".
- If the mutation fails (network flake, permissions), log the failure and continue — resolution is best-effort; the reply + commit are the source of truth.

Resolution happens per-finding, inline, not in a batch at the end, so a mid-run abort still leaves the PR in a consistent state (fixes that landed are marked resolved; fixes that didn't, aren't).

### Research Insights

**GraphQL shape (verified against GitHub docs):**

```graphql
mutation ResolveReviewThread($threadId: ID!) {
  resolveReviewThread(input: {threadId: $threadId}) {
    thread {
      id
      isResolved
      viewerCanResolve
    }
  }
}
```

**Failure detection:** Inspect the top-level `errors[]` array, not just HTTP status. `data.resolveReviewThread == null` with `errors[0].type` indicating the failure mode:
- `NOT_FOUND` — bad threadId or no read access (shouldn't happen post-0g).
- `FORBIDDEN` / `"Resource not accessible by integration"` — token lacks PR write. The `viewerCanResolve` pre-check (above) prevents this in the expected path.
- `"Could not resolve to a node with the global id of …"` — wrong ID type; indicates a bug in how 0g passed the ID through.

**Permissions needed:**
- Classic PAT: `repo` (private) or `public_repo` (public only).
- Fine-grained PAT / GitHub App: `Pull requests: write` on the target repo.
- The skill already has push access and posts reviews, so these are satisfied in all expected call contexts.

**Rate-limit cost:** 1 GraphQL point per resolve call against the 5000 pts/hr pool (10000 for Enterprise-owned apps, 1000 for `GITHUB_TOKEN` in Actions). Negligible even on a PR with 100 resolved threads.

**Idempotency (⚠️ not formally documented):** Community reports indicate that resolving an already-resolved thread returns `thread.isResolved: true` with no error — effectively a no-op success. The plan treats this as expected behavior but flags it as non-authoritative. Implementation should verify `thread.isResolved` in the response rather than assume success from a 200 status.

**References:**
- https://docs.github.com/en/graphql/reference/mutations#resolvereviewthread
- https://docs.github.com/en/graphql/reference/input-objects#resolvereviewthreadinput
- https://docs.github.com/en/graphql/overview/resource-limitations

### C. Tick off PR body task-list items

Extend Step 7 (final summary) of `ghcp-review-resolve` with a new sub-step **7.0 — Update PR body** that runs before the user-facing summary.

Algorithm:

1. **Fetch** the current PR body and snapshot it:
   ```bash
   gh pr view "$PR_NUMBER" --repo "$OWNER/$REPO" --json body -q .body > /tmp/ghcp-pr-body-original.md
   ```
2. **Size-guard:** if `wc -c /tmp/ghcp-pr-body-original.md` ≥ 60000 bytes, skip the rewrite (GitHub's de-facto body limit is 65,536 bytes — leave 5 KB margin for the audit block and finding entries).
3. **Extract unchecked task-list items** via regex `^\s*- \[ \] (.+)$` (per-line, anchored). Preserve line numbers so step 5 can edit in place without line-shifting the rest of the body.
4. For each accepted-and-fixed finding (from the per-finding record built during Step 6), attempt to match it to an unchecked item using a **deterministic-first** strategy:
   - **Pass 1 — normalized substring match.** Lowercase + collapse whitespace + strip markdown syntax on both sides. If the finding's file path (e.g. `.github/skills/ghcp-review-resolve/SKILL.md`) or the first 40 chars of its rationale appears as a substring in any task-list item, that's a match.
   - **Pass 2 — rapidfuzz token_set_ratio ≥ 85.** Use `token_set_ratio` (handles reordered words and subset overlap) rather than Levenshtein (poor for reordered tokens). Threshold 85 verified against rapidfuzz docs as the standard high-confidence cutoff for short-to-medium strings.
   - **Pass 3 — LLM fallback, only for ambiguous cases.** If passes 1 and 2 yield zero matches OR multiple items above threshold, invoke a small classification prompt (temperature 0, strict JSON output with `matched_id`, `confidence ∈ [0.0, 1.0]`, `reason`, `alternates[]`). Require `confidence ≥ 0.8` and validate `matched_id ∈ input_ids` before accepting. Expected invocation rate: 0–3 per run at this project's scale.
   - **Unmatched findings** become a trailing "unlinked — needs manual triage: `<finding-title>`" line inside the audit block. Never silently drop.
5. **Tick** each matched item by rewriting `- [ ]` → `- [x]` on that specific line number only. Preserve the rest of the body byte-for-byte.
6. **Manage the audit block.** Use sentinel fences so the block is idempotent across re-runs:
   ```markdown
   <!-- BEGIN:ghcp-audit v1 -->
   <!-- Managed by ghcp-review-resolve. Do not edit by hand. -->
   Last run: 2026-04-24T11:48:37Z (commits abc123..def456)
   Ticked: 3 of 7 task-list items
   - Item "Fix Step 5 reference" ← comment 3139049827 (fixed in abc123)
   - Item "Add pagination to 0g query" ← comment 3139049895 (fixed in def456)
   - Item "Resolve threads after fix" ← comment 3139049941 (fixed in def456)
   Unlinked (needs manual triage):
   - "Race condition in cache update" (skipped — left for human)
   <!-- END:ghcp-audit v1 -->
   ```
   Parse out the existing block with regex `(?s)<!-- BEGIN:ghcp-audit v1 -->.*?<!-- END:ghcp-audit v1 -->`, rebuild, splice back. Never nest `-->` inside the payload (HTML comments don't nest); if any rationale contains `--`, base64-encode or strip it before embedding.
7. **Write back with a sentinel-based concurrency check** (there is no native optimistic-concurrency API on PR bodies — no `If-Match`, no ETag):
   ```bash
   # Re-fetch right before write to detect mid-run edits
   CURRENT_BODY=$(gh pr view "$PR_NUMBER" --json body -q .body)
   if [ "$CURRENT_BODY" != "$(cat /tmp/ghcp-pr-body-original.md)" ]; then
     echo "PR body changed during run — skipping auto-tick to avoid overwriting user edits."
   else
     gh pr edit "$PR_NUMBER" --body-file /tmp/ghcp-pr-body-updated.md
   fi
   ```
   `gh pr edit --body-file` reads bytes as UTF-8 and sends them as the REST `body` field. Server-side normalizes CRLF → LF; no client-side trim. Trailing newline in the file is preserved.
8. If no task-list items exist in the body, or no findings matched, skip the write and log "no task-list items to update" — the audit block is not created on empty runs.

Guardrails on body edits:

- **Never** introduce, delete, or reorder non-task-list content outside the audit fence. Only flip `[ ]` → `[x]` on matched lines and manage the fenced block.
- **Never** tick an item the skill can't confidently map to work it did. Unmatched goes to "needs manual triage"; false positives are worse than false negatives here.
- **Never** exceed the 60 KB body ceiling. If adding the audit block would push the body over 60 KB, truncate finding entries inside the block (keep ticks, summarize unmatched).
- **Abort the whole sub-step** on sentinel mismatch (user pushed a body edit mid-run); log it, continue to the summary without updating.

### Research Insights

**Why deterministic-first instead of LLM-first:** At this project's scale (<20 task-list items × <10 findings = ~200 pairs), normalized substring + `token_set_ratio` catches almost all real matches at zero marginal cost, and false-positive rate is lower than an LLM's (LLMs rarely emit confidence < 0.7 even on wrong matches — "0.8" typically means "maybe"). LLMs are worth it only at >50×50 scale or when matching requires reasoning across paraphrases.

**Why no embeddings pipeline:** At N=200 pairs, embeddings add a dependency (`text-embedding-3-small` API or local model) with no accuracy win over rapidfuzz. Revisit at 10× scale.

**LLM prompt shape** (for Pass 3 fallback, temperature=0, strict JSON):
```
You are matching review findings to a PR checklist.

Checklist items (id → text):
1: <text>
2: <text>
...

Finding: "<finding rationale + file:line>"

Return strict JSON:
{
  "matched_id": <int or null>,
  "confidence": <0.0–1.0>,
  "reason": "<one sentence>",
  "alternates": [<int>, ...]
}

Rules:
- matched_id = null if no item meaningfully addresses the finding.
- confidence < 0.8 → treat as unmatched.
- Do not invent ids outside the list.
```

**Known LLM failure modes to guard against:** stylistic lexical overlap (both mention "error handling" but refer to different paths — mitigate by including file:line in both sides), confidence inflation (calibrate by sampling), order bias (shuffle checklist per call or request `alternates`), negation anchoring ("this is NOT about X" still matches X — strip negations).

**References:**
- https://cli.github.com/manual/gh_pr_edit
- https://docs.github.com/en/rest/pulls/pulls#update-a-pull-request
- 64 KiB body limit: sourced from community discussions (github/orgs/community#31295, #34879) — not in formal REST reference; treated as reliable de-facto limit.
- https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html#token-set-ratio

## Technical Considerations

- **GraphQL ID acquisition for thread resolution.** The existing `0g` query already fetches `reviewThreads` but doesn't currently select the `id` field. Update the query in 0g to select `id` **and `viewerCanResolve`** on each thread node and keep both in the in-memory finding record. No extra API call is needed. The `viewerCanResolve` flag is what Step 6.7 uses to skip the mutation cleanly when permissions are lacking.
- **Agent discovery.** This repo loads agents from `.github/agents/*.agent.md` — filename-based, no manifest. The port must use that naming (not upstream's `<name>.md`) or the agents won't be invocable. Frontmatter contract: only `description:` is universal across the 51 existing agents; add `user-invocable: true` only if the agent should be user-triggerable (not required for the toolkit's internal use).
- **License.** `pr-review-toolkit` is Apache-2.0. Include the upstream `LICENSE` verbatim under `.github/skills/pr-review-toolkit/LICENSE`. Per-file HTML-comment attribution header (see section A) on every ported `.md`. No `NOTICE` file needed — upstream has none. The port introduces the repo's first attribution precedent; keep it minimal.
- **Single-reviewer mode stays supported.** Even after porting, `COPILOT_AVAILABLE=false` + `pr-review-toolkit` still constitutes a valid run. Don't delete the single-reviewer branch in the SKILL — the port just makes the dual-reviewer branch actually reachable.
- **GraphQL permissions.** `resolveReviewThread` requires `repo` (classic) or `Pull requests: write` (fine-grained/App). The skill already has that (it pushes commits and posts reviews). No new auth scope needed.
- **PR body edit noise.** Using `gh pr edit --body-file` rewrites the whole body, which shows up as a PR-body edit event in the timeline. That's acceptable and matches how humans edit PR bodies; the audit-fence marker (`<!-- BEGIN:ghcp-audit v1 -->`) makes the skill's edits self-explanatory and keeps them replaceable.
- **No optimistic concurrency on PR bodies.** Confirmed there's no `If-Match` / ETag / SHA guard on `PATCH /repos/{o}/{r}/pulls/{n}`. Plan 7.0's sentinel-based guard (snapshot-on-read, diff-on-write) is the accepted workaround.
- **Script invocation convention.** If any scripts get added under `.github/skills/pr-review-toolkit/scripts/`, invoke them via `bash ${CLAUDE_PLUGIN_ROOT}/skills/pr-review-toolkit/scripts/<name>` (matches existing `resolve-pr-parallel` usage). Exec bit is not required.

## System-Wide Impact

- **Interaction graph:** `ghcp-review-resolve` Step 1b → `Skill(pr-review-toolkit:review-pr)` → 6 sub-agents → PR review comments. Step 6 → `resolve-pr-thread` script → `resolveReviewThread` GraphQL mutation. Step 7.0 → `gh pr edit --body-file` → PR body update event.
- **Error propagation:** Thread-resolution failures are logged and ignored (best-effort). Body-update failures abort 7.0 only; summary still runs. Agent failures inside `pr-review-toolkit` surface as "reviewer missing" in Step 2's completion check and degrade gracefully to whichever reviewer did complete.
- **State lifecycle risks:** A mid-run abort between commit-push and thread-resolve leaves an unresolved thread with a posted fix reply — human-visible as "code pushed, human can resolve". Not a regression; matches today's behavior.
- **API surface parity:** `resolve-pr-parallel` already ships `resolve-pr-thread` and `get-pr-comments`. Reuse both; don't duplicate the GraphQL logic inside `ghcp-review-resolve`.
- **Integration test scenarios:**
  1. Dual-reviewer happy path: both reviewers produce findings, overlap is preserved, all accepted findings produce commits + resolved threads + ticked boxes.
  2. Copilot unavailable: only `pr-review-toolkit` runs, Step C still resolves its own threads.
  3. PR has no task-list items: 7.0 exits silently.
  4. User edits PR body mid-run: 7.0 aborts with an explanatory log line.
  5. Thread-resolve mutation fails for one finding but succeeds for others: the failing one stays open, others resolve.

## Acceptance Criteria

### A — Port pr-review-toolkit

- [ ] `.github/skills/pr-review-toolkit/SKILL.md` exists with required frontmatter (`name`, `description`) and adapted content from upstream's `commands/review-pr.md` + `README.md`.
- [ ] Six agent files exist at `.github/agents/pr-*.agent.md` — `pr-comment-analyzer`, `pr-test-analyzer`, `pr-silent-failure-hunter`, `pr-type-design-analyzer`, `pr-code-reviewer`, `pr-simplification-analyzer` (renamed from upstream `code-simplifier`) — each with a `description:` frontmatter line.
- [ ] `.github/skills/pr-review-toolkit/LICENSE` contains the upstream Apache-2.0 license text verbatim.
- [ ] Each ported file starts with the HTML-comment attribution header (upstream repo URL + Apache-2.0 + LICENSE path + modifications note with date).
- [ ] `pr-comment-analyzer.agent.md` description explicitly distinguishes its scope ("analyze accuracy and rot of code comments") from sibling `pr-comment-resolver.agent.md` ("reply-and-resolve review comment threads").
- [ ] The skill can be invoked via `Skill(skill="pr-review-toolkit:review-pr", args="<PR URL>")` and posts at least one inline review comment on a test PR (verified by manual smoke test).
- [ ] `ghcp-review-resolve` Step 1b's existing `Skill(...)` call succeeds (no "skill not found" error) and its review comments are picked up by Step 3's normalization.

### B — Resolve Copilot threads

- [ ] The 0g GraphQL query selects thread `id` **and `viewerCanResolve`**, and both are carried through to per-finding records.
- [ ] `ghcp-review-resolve` SKILL.md has a new Step 6.7 titled "Resolve thread" describing the five-gate guardrail (accepted + pushed + replied + verified + viewerCanResolve) before calling `resolve-pr-thread`.
- [ ] After a successful fix loop on a test PR, the addressed threads show `isResolved=true` in `gh api graphql` output.
- [ ] Findings that are skipped, reverted, or unverifiable leave their threads **open**.
- [ ] When `viewerCanResolve` is false, the skill logs a permission-skip without attempting the mutation (no `FORBIDDEN` error in logs).
- [ ] Thread-resolve failures (any `errors[]` in GraphQL response) are logged but do not abort the pipeline.

### C — Tick PR task-list items

- [ ] `ghcp-review-resolve` SKILL.md has a new Step 7.0 "Update PR body" before the existing Step 7 summary.
- [ ] Running the skill on a test PR whose body contains `- [ ] Fix <foo>` lines flips matched items to `- [x] Fix <foo>` and leaves the rest of the body unchanged.
- [ ] Matching uses the deterministic-first strategy: normalized substring → rapidfuzz `token_set_ratio ≥ 85` → LLM fallback for ambiguous cases only.
- [ ] Unmatched findings appear in the audit block under "Unlinked (needs manual triage)" — never silently dropped.
- [ ] A `<!-- BEGIN:ghcp-audit v1 --> ... <!-- END:ghcp-audit v1 -->` audit block lists ticked items, commit SHAs, and unlinked findings.
- [ ] Re-running the skill on the same PR does not duplicate the audit block (it replaces the previous one using the sentinel-fence regex).
- [ ] If the PR body changes between the initial fetch and the write-back, the update is skipped and the summary logs the skip (sentinel-based concurrency check).
- [ ] If the post-audit body would exceed 60 KB, the write is aborted with a logged warning.
- [ ] No false-positive tick: an unrelated `- [ ]` item elsewhere in the body stays unchecked.

## Success Metrics

- On a PR with Copilot findings, running `/ghcp-review-resolve` end-to-end produces: N addressed findings, N commits, N thread replies, **N resolved threads**, and all corresponding PR-body task-list items ticked — with no manual follow-up required.
- The dual-reviewer pipeline is actually exercised (not silently degraded to single-reviewer) in at least one real PR run.
- Zero incorrect ticks or incorrect resolutions observed across the first 3 real runs.

## Dependencies & Risks

- **Upstream license compliance.** Apache-2.0 requires attribution (§4(a)) and a modifications notice (§4(b)). Handled by the `LICENSE` copy + per-file HTML-comment header (see §A Research Insights). `NOTICE` file is optional — upstream has none, so we don't propagate one.
- **Agent collision.** Soft semantic overlap identified: `pr-code-simplifier` → renamed to `pr-simplification-analyzer` to avoid colliding in purpose with `.github/agents/code-simplicity-reviewer.agent.md`. `pr-comment-analyzer` sits next to `pr-comment-resolver`; disambiguated via `description:` copy ("analyze accuracy" vs. "reply-and-resolve").
- **Body-edit flakiness.** GitHub occasionally rejects `pr edit --body-file` with transient 5xx. 7.0 treats this as a warning, not a pipeline-abort. The sentinel-fence design means a retry on the next run picks up where this one left off without duplicating work.
- **`resolveReviewThread` idempotency assumption.** Behavior on already-resolved threads is **not formally documented** (community reports say no-op success). Mitigation: always verify `thread.isResolved` in the mutation response rather than assuming success from a 200 status. Do not log re-resolve as a failure.
- **GraphQL schema drift.** `resolveReviewThread` and `reviewThreads` are stable, but the `id` + `viewerCanResolve` field selectors must be added to the 0g query and tested. A schema-change smoke test (call the query once against a real PR, confirm both fields are present) is part of implementation.
- **PR body 64 KiB ceiling.** De-facto, not in docs. 7.0's 60 KB safety threshold protects against runaway audit blocks; exceeding the limit aborts the body edit with a logged warning.
- **`pr-review-toolkit` wall-clock.** The 6-agent review is slower than Copilot's. Step 2's 10-minute cap may need raising for very large PRs, but that's out of scope for this plan — raise it only if we observe timeouts in practice.
- **Attribution precedent.** This port introduces the first `LICENSE` and HTML-comment attribution headers in the repo. Future third-party ports should follow the same pattern. Document the convention in `CONTRIBUTING.md` as a follow-up (out of scope).

## Sources & References

### Internal

- Skill to modify: `.github/skills/ghcp-review-resolve/SKILL.md`
- Existing resolve helper (reuse): `.github/skills/resolve-pr-parallel/scripts/resolve-pr-thread`
- Existing comment-fetch helper (pattern reference): `.github/skills/resolve-pr-parallel/scripts/get-pr-comments`
- Agent file convention: `.github/agents/architecture-strategist.agent.md` (shape example)
- Copilot instructions: `.github/copilot-instructions.md`

### External

- Anthropic `pr-review-toolkit` source: https://github.com/anthropics/claude-plugins-official/tree/main/plugins/pr-review-toolkit
- GraphQL `resolveReviewThread` mutation: https://docs.github.com/en/graphql/reference/mutations#resolvereviewthread
- `gh pr edit --body-file` docs: https://cli.github.com/manual/gh_pr_edit

### Related Work

- PR #23 (`feat/ghcp-review-resolve-skill`) — the PR that added `ghcp-review-resolve` and on which this plan's enhancements will land (or a follow-up PR).
