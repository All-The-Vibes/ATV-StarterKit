---
name: kb-brainstorm
description: 'Proportional brainstorming for vertical-slice work. Orients from repo memory before asking product questions, runs external research only when it is likely to change framing, and produces requirements for `/kb-plan`. Use when the user says ''kb brainstorm'', ''brainstorm'', ''brainstorm with research'', or ''brainstorm before kb-plan''.'
argument-hint: "[feature idea or problem to explore]"
---

# KB Brainstorm - Proportional Requirements

**Note: The current year is 2026.** Use this when dating requirements documents.

`kb-brainstorm` answers **WHAT** to build. It reads the minimum local context needed before asking product questions. It runs external research only when research can change framing, reduce rework, or settle a material uncertainty.

This pairs naturally with `/kb-plan` (vertical-slice decomposition).

This skill does not implement code. It explores, validates, clarifies, and documents decisions for later planning or execution.

**IMPORTANT: All file references in generated documents must use repo-relative paths (e.g., `src/models/user.rb`), never absolute paths. Absolute paths break portability across machines, worktrees, and teammates.**

## When to Pick `kb-brainstorm`

| Situation | Pick |
|---|---|
| Design space is well known; conversation is the bottleneck | `brainstorming` or lightweight chat |
| Prior art / competitive landscape is **likely to change framing** | `kb-brainstorm` with research |
| Output will feed `/kb-plan` (vertical slices) | `kb-brainstorm` |
| Existing brainstorm doc just needs research enrichment | `kb-research` plus update the doc |

`kb-brainstorm` does **bounded orientation before product decisions**. `kb-research` does reusable research notes. Do not run external research just because research feels good.

## Core Principles

1. **Orientation before questions** — Do not interview the user before checking the local memory and obvious repo patterns. External research is conditional, not automatic.
2. **Evidence beats intuition** — Every decision in the requirements doc should have either a research citation or an explicit "no evidence — assumption" tag.
3. **Be a thinking partner** — Bring alternatives, challenge assumptions, and surface what-ifs. Don't just extract requirements.
4. **Resolve product decisions here** — User-facing behavior, scope boundaries, and success criteria belong in this workflow. Detailed implementation belongs in planning.
5. **Right-size the artifact** — Match ceremony to scope. Lightweight work gets a compact doc; deep work gets a fuller one. Do not pad sections that add no value.
6. **Apply YAGNI to carrying cost, not coding effort** — Prefer the simplest approach that delivers meaningful value. Avoid speculative complexity, but include low-cost polish that compounds.
7. **Keep implementation out of the requirements doc by default** — Do not include libraries, schemas, endpoints, file layouts, or code-level design unless the brainstorm itself is inherently technical.

## Interaction Rules

1. **Ask one question at a time** — Do not batch several unrelated questions into one message.
2. **Prefer single-select multiple choice** — Use single-select when choosing one direction, one priority, or one next step.
3. **Use multi-select rarely and intentionally** — Only for compatible sets such as goals, constraints, or success criteria that can all coexist. If prioritization matters, follow up by asking which selected item is primary.
4. **Every multiple-choice question needs an escape hatch** — Always include an option such as `Other / let me explain` or `None of these`. The listed choices are suggestions, not a cage. If the user selects the escape hatch, return to the normal conversation and let them answer freely.
5. **Use the platform's question tool when available** — `ask_user` in Copilot CLI, equivalent blocking tools elsewhere. Otherwise present numbered options in chat and wait.
6. **Do not trap rich answers in the question UI** — If the answer may need an image, screenshot, file, diagram, long explanation, pasted error, or nuanced correction, ask in the normal conversation instead of the blocking question tool. If a question-tool exchange reveals that richer context is needed, stop asking choices and bring it back to chat.
7. **Questions must earn their keep** — Ask only when the answer changes scope, behavior, priority, acceptance criteria, risk, or verification. Do not ask quota questions.

## Question Gate

`kb-brainstorm` owns the assumption boundary before planning. It must classify
material unknowns before handoff:

| Class | Meaning | Planning Allowed? |
|---|---|---|
| `ask-now` | Human answer changes scope, user intent, acceptance criteria, safety, architecture direction, or verification | No |
| `research-first` | External/source research can answer before asking the user | No, until researched or reclassified |
| `safe-assumption` | Reversible, low-risk assumption with evidence or explicit confidence | Yes, record it |
| `defer-to-planning` | Technical detail better answered while slicing or reading code | Yes, record owner and trigger |
| `parked` | Out of current scope by explicit decision | Yes, record forbidden claims |

Do not hand off with unresolved `ask-now` or `research-first` items. Ask
blocking questions one at a time until resolved, or convert each item into a
recorded decision, `safe-assumption`, `defer-to-planning`, or `parked` entry.

Safe assumptions must include why they are reversible, what evidence supports
them, and what proof will catch a wrong assumption later. If that cannot be
stated compactly, the item is not safe.

## Token Budget

Every token must pay rent.

- Default to the shortest artifact that will let `/kb-plan` avoid inventing product behavior.
- Prefer local memory over new external research.
- Use `kb-research` for reusable research that should survive this brainstorm.
- Use `kb-compact` if the draft grows without adding decision value.
- Keep exact requirements, decisions, assumptions, paths, sources, and open questions.

## Intellectual Honesty

Apply `kb-first-principles` behavior during this workflow.

- Accept user-owned intent/context corrections without debate.
- Verify factual claims when they affect the requirements.
- Defend recommendations only while evidence or reasoning supports them.
- Concede only the specific premise that changed; do not pendulum-swing.
- Challenge assumptions when an unchallenged flaw would cause rework.

Response shape under pushback:

```text
I still think X because Y.
You're right about Z; that changes A.
I would revise to B, not the opposite extreme.
```

## Output Guidance

- **Keep outputs concise** — Short sections, brief bullets, only enough detail to support the next decision.
- **Use repo-relative paths** — When referencing files, use paths relative to the repo root (e.g., `src/models/user.rb`), never absolute paths.
- **Mark evidence** — When the requirements doc cites research, link or quote the source. When a claim has no source, label it as an assumption.
- **Verify before claiming** — When the brainstorm touches checkable infrastructure (database tables, routes, config files, dependencies, model definitions), read the relevant source files to confirm what actually exists. Any claim that something is absent must be verified or labelled as an unverified assumption.

## Feature Description

<feature_description> #$ARGUMENTS </feature_description>

**If the feature description above is empty, ask the user:** "What would you like to explore? Describe the feature, problem, or improvement. I'll orient from local context first, then research only if it changes the decision."

Do not proceed until you have a feature description from the user.

## Execution Flow

### Phase 0: Resume, Assess, and Route

#### 0.1 Resume Existing Work When Appropriate

If the user references an existing brainstorm topic or document, or there is an obvious recent matching `*-requirements.md` file in `docs/brainstorms/`:

- Read the document.
- Confirm with the user before resuming: "Found an existing requirements doc for [topic]. Should I continue from this, or start fresh?"
- If resuming, summarize the current state briefly, continue from existing decisions and outstanding questions, and update the existing document instead of creating a duplicate. Skip Phase 1 (intake). Decide in Phase 3 whether new external research is needed; the existing doc may already contain enough.

#### 0.2 Assess Whether Brainstorming Is Needed

**Clear-requirements indicators:**

- Specific acceptance criteria provided
- Referenced existing patterns to follow
- Described exact expected behavior
- Constrained, well-defined scope
- No framing risk — the user clearly knows the right shape and the landscape isn't going to change it

**If all of the above are true:** keep the interaction brief. Do local orientation only (Phase 2 plus Phase 3.2 prior art if useful), then go to Phase 8 (capture) -> Phase 9 (review) -> Phase 10 (handoff). Skip external landscape research unless a concrete risk surfaces. Skip Phases 5-7 unless they would prevent rework.

Do not skip orientation entirely. Do skip external research when it will not change the decision.

#### 0.3 Assess Scope

Use the feature description plus a **very light pre-scan** (one or two ripgrep queries at most) to classify the work:

- **Lightweight** — small, well-bounded, low ambiguity
- **Standard** — normal feature or bounded refactor with some decisions to make
- **Deep** — cross-cutting, strategic, or highly ambiguous

Match research depth and Q&A depth to scope. Lightweight scopes get local orientation and 0-2 questions. Deep scopes may get full landscape research and a longer dialogue.

If the scope is unclear, ask **one topic-identity question** to disambiguate (e.g., "is this about the API or the UI?"), then proceed. Do not ask scope, user, success-criteria, constraint, or prioritization questions yet — those come after research in Phase 6.

**Grilling mode:** activate for Deep scope, explicit "interview me" / "don't assume" / "think first" requests, or when the feature description is too vague to safely proceed. Record: `Grilling mode: ON`. This changes Phase 6 behavior — see below. Grilling mode does not change the Question Gate; it only changes how questions are surfaced.

### Phase 1: Topic Intake

Restate the user's feature in your own words in 1–3 sentences and confirm:

- "Here's what I heard: [restated topic]. Did I get the core right?"

If the user corrects you, accept the correction silently and proceed.

**Strict rule for this phase:** Only **topic-identity confirmation** is allowed here. Do **not** ask scope, users, success criteria, constraints, prioritization, or trade-off questions. Those come in Phase 6 after research has run. The point of this phase is to make sure research targets the right thing, not to start product discovery.

### Phase 2: Repo Context Scan

Scan the repo before research. Match depth to scope.

**Lightweight** — Search for the topic, check if something similar already exists, and move on.

**Standard and Deep** — Two passes:

- *Constraint Check* — Check `AGENTS.md` and adjacent project instruction files for workflow, product, or scope constraints that affect the brainstorm. If these add nothing, move on.
- *Topic Scan* — Search for relevant terms. Read the most relevant existing artifact (brainstorm, plan, spec, skill, feature doc). Skim adjacent examples covering similar behavior.

If nothing obvious appears after a short scan, say so and continue.

Two rules govern technical depth during the scan:

1. **Verify before claiming** — When the topic touches checkable infrastructure, read the relevant source files. Claims of absence must be verified or labelled as unverified.
2. **Defer design decisions to planning** — Schemas, migration strategies, endpoint structure, and deployment topology belong in planning unless the brainstorm is itself about a technical decision.

### Phase 3: Research Decision

Decide whether external research is worth the tokens before running it.

Run external research only when at least one is true:

- Prior art can change product framing, protocol, architecture direction, or UX.
- The user explicitly asks for research.
- The brainstorm concerns security, auth, streaming, persistence, external APIs, pricing, regulations, or fast-moving libraries.
- Local memory is stale or thin and the decision has path dependency.

If none are true, record: `External research skipped: low expected decision value.`

When research runs, time-box it. Prefer 3-7 targeted questions over exhaustive landscape scanning.

#### 3.1 Market and Landscape

For the problem space:

- How do other tools or products solve this?
- What is the current state of the art?
- Are there open-source solutions worth studying?
- What user-experience or implementation patterns are considered best practice?
- What scale or complexity thresholds make common approaches break down?

Aim for **3–5 concrete examples** when external research is available. If browsing or network access is unavailable, mark those facts as unverified rather than guessing.

#### 3.2 Prior Art and Learnings

Search for relevant institutional knowledge and similar code:

```bash
rg --files docs/solutions
rg -n "[key terms from the topic]"
```

If `rg` is unavailable, use the platform's native file search.

For each potentially relevant learning:

- Does it apply to this brainstorm?
- What specific insight should carry forward?
- If not applicable, why not?

#### 3.3 Applicable Skills

Check for skills that could provide domain-specific perspective:

```bash
rg --files .github/skills -g "SKILL.md"
```

Also check global / plugin skill roots exposed by the current platform (e.g., `~/.copilot/skills`, `~/.codex/skills`, or plugin skill directories).

For each matching skill, apply only the relevant perspective: does it suggest a framing, a constraint, or a known failure mode?

#### 3.4 Risk and Failure-Mode Survey

For the candidate approaches the topic implies, list known failure modes from prior art:

- What goes wrong at scale?
- What goes wrong on day 30 or 90, not day 1?
- What integrations or operational concerns are commonly underestimated?

### Phase 4: Synthesize Research Brief

Before any product question, produce a short **orientation brief**. If external research ran, include the research findings. This is conversational scaffolding, not part of the requirements doc.

**Distinction:**
- **Orientation Brief (Phase 4)** — local context plus any material research, used to align the user before Q&A. Lives in chat.
- **Research Summary (Phase 8 doc)** — only the findings that materially affected the requirements or decisions. Lives in the requirements doc.

Do not paste the brief verbatim into the doc later — distill.

```markdown
## Orientation Brief

**Landscape (only if researched):**
- [Tool / approach] — [what they do] — [why notable]

**Established patterns:**
- [Pattern] — [where it shows up] — [when it fits]

**Known failure modes:**
- [Failure] — [conditions]

**Repo prior art:**
- [Existing capability or pattern] (`path/to/file`) — [relevance]

**Applicable learnings:**
- [Learning title from docs/solutions or N/A]

**Open uncertainty:**
- [Question research could not resolve cleanly]
```

Display the brief to the user. Then ask one alignment question:

> "Does any of this change the framing before we go deeper? (a) Yes, here's what shifts; (b) No, my framing still holds; (c) Show me more on [topic]; (d) Other / let me explain."

If the user selects (c), do another targeted research pass on that topic only and re-show the relevant section.

### Phase 5: Product Pressure Test

Now challenge the request using local context and any research that was worth running. Match depth to scope.

**Lightweight:**

- Does the research reveal a simpler off-the-shelf path?
- Is this duplicating something that already covers it?
- Is there a clearly better framing with near-zero extra cost?

**Standard:**

- Is this the right problem, or a proxy for a more important one?
- What user or business outcome actually matters here?
- What happens if we do nothing?
- Given the landscape we just surveyed, is there a nearby framing that compounds value at low extra cost?
- Is the highest-leverage move the request as framed, a reframing, an adjacent addition, a simplification, or doing nothing?

**Deep** — Standard questions plus:

- What durable capability should this create in 6–12 months?
- Does this move the product toward that, or is it only a local patch?
- Which of the failure modes from Phase 3.4 are we accepting?

Use the result to sharpen the conversation, not to bulldoze the user's intent.

### Phase 6: Targeted Q&A

Now run the conversation. The questions are sharper because they reference the research brief. Use the platform's blocking question tool when available.

**Guidelines:**

- Ask questions **one at a time**.
- Prefer **single-select** when choosing one direction, one priority, or one next step.
- Use **multi-select** only for compatible sets that can all coexist; if prioritization matters, ask which selected item is primary.
- Include **Other / let me explain** or **None of these** in every multiple-choice question.
- Ask in normal chat, not the blocking question tool, when the user may need to attach an image, screenshot, file, diagram, pasted output, or longer explanation.
- Anchor each question to the research where appropriate: "Tools like X do A, others do B — which fits our users?"
- Start broad (problem, users, value) then narrow (constraints, exclusions, edge cases).
- Validate assumptions explicitly: "I'm assuming Y based on research finding Z — is that right?"
- Resolve product decisions here; leave technical implementation choices for planning.
- Make requirements concrete enough that planning will not need to invent behavior.

**Every question must pass the Question Gate before being asked.** Classify each candidate question — if it is not `ask-now` or would not change scope, behavior, priority, acceptance criteria, risk, or verification, convert it to a `safe-assumption` or `defer-to-planning` and skip asking. Do not ask quota questions or questions whose answer you can reasonably infer.

**Grilling mode (when active from Phase 0.3):**

When grilling mode is ON, apply this additional discipline to every question:

- **Surface your own recommendation first.** Before asking, state the answer you would choose and why: "I'd go with X because Y — does that fit, or is there a constraint I'm missing?" The user corrects or confirms rather than thinking from scratch. This keeps questions productive instead of open-ended interrogations.
- **Walk the decision tree, don't roam it.** Each question should resolve a branch that is genuinely open. Once a branch is resolved, move to the next unresolved one — do not revisit settled decisions or ask follow-ups that the answer already answered.
- **Stop when shared understanding is reached**, not when a question quota is filled. The signal is: no unresolved `ask-now` items remain and you could write the requirements doc without inventing behavior.
- **Grilling does not suspend the Question Gate.** Every grilling question must still earn its place — if it would not change anything material, skip it even in grilling mode.

**Exit condition:** Continue until the idea is clear OR the user explicitly wants to proceed.

### Phase 7: Approaches

If multiple plausible directions remain, propose **2–3 concrete approaches** based on research and conversation. Otherwise state the recommended direction directly.

For each approach, provide:

- Brief description (2–3 sentences)
- Pros and cons
- Key risks or unknowns (from Phase 3.4)
- When it's best suited
- Closest analogue from research (e.g., "this is how X solves it")

When useful, include one deliberately higher-upside alternative — an adjacent reframing or addition that the landscape suggests would compound value, presented as a challenger option, not the default. Omit it when the work is already obviously over-scoped.

Lead with your recommendation and explain why. Prefer simpler solutions when added complexity creates real carrying cost, but do not reject low-cost, high-value polish.

If relevant, call out whether the choice is:

- Reuse an existing pattern
- Extend an existing capability
- Build something net new

### Phase 8: Capture the Requirements

Write or update a requirements document only when the conversation produced durable decisions worth preserving.

This document behaves like a lightweight PRD without PRD ceremony. Include what planning needs to execute well, and skip sections that add no value for the scope. Do **not** include implementation details such as libraries, schemas, endpoints, file layouts, or code structure unless the brainstorm is inherently technical.

**Required content for non-trivial work:**

- Problem frame
- Concrete requirements or intended behavior with stable IDs
- Scope boundaries
- Success criteria
- Research summary (top 3–5 findings with sources)

**Include when materially useful:**

- Key decisions and rationale (with research citations where applicable)
- Dependencies or assumptions
- Outstanding questions
- Alternatives considered (with research citations)
- Slice candidates — when handing off to `/kb-plan`, list 3–7 candidate **user-visible increments** the research and conversation suggest. Keep these advisory and high-level — describe what each increment delivers, not blockers, ordering, or dependency design. `/kb-plan` owns sequencing.

**Document structure:** Use `references/requirements-template.md`. Load it only when writing or updating the requirements document.

Use visual aids only when they materially reduce ambiguity. See `references/requirements-template.md` for the visual-aid rules.
For **Standard** and **Deep** brainstorms, a requirements document is usually warranted.

For **Lightweight** brainstorms, keep the document compact. Skip document creation when only brief alignment is needed and no durable decisions need to be preserved.

For very small requirements docs with only 1–3 simple requirements, plain bullet requirements are acceptable. For **Standard** and **Deep** docs, use stable IDs like `R1`, `R2`, `R3` so planning and review can refer to them unambiguously.

When requirements span multiple distinct concerns, group them under bold topic headers within the Requirements section. Group by logical theme, not discussion order. Requirements keep their original IDs — numbering does not restart per group.

When the work is simple, combine sections rather than padding them. A short requirements document is better than a bloated one.

Before finalizing, check:

- What would `/kb-plan` or `/kb-plan` still have to invent if this brainstorm ended now?
- Do any requirements depend on something claimed to be out of scope?
- Are any unresolved items actually product decisions rather than planning questions?
- Did implementation details leak in when they shouldn't have?
- Do any requirements claim that infrastructure is absent without verification?
- Is the research summary honest about confidence and gaps?
- Would a visual aid (flow diagram, comparison table, relationship diagram) help a reader grasp the requirements faster than prose alone?

If planning would need to invent product behavior, scope boundaries, or success criteria, the brainstorm is not complete yet.

Ensure `docs/brainstorms/` directory exists before writing.

If the document contains outstanding questions:

- Use `Resolve Before Planning` only for questions that truly block planning.
- If `Resolve Before Planning` is non-empty, keep working those questions during the brainstorm by default.
- If the user explicitly wants to proceed anyway, convert each remaining item into an explicit decision, assumption, or `Deferred to Planning` question first.
- Put technical or research-needing questions under `Deferred to Planning` when they are better answered there.
- Use tags like `[Needs research]` when the planner should likely investigate the question rather than answer from repo context alone.
- Before Phase 10, every material unknown must be classified through the
  Question Gate. `Resolve Before Planning` may be empty only when no unresolved
  `ask-now` or `research-first` items remain.

### Phase 9: Document Review

When a requirements document was created or updated, run the `document-review` skill on it before presenting handoff options. Pass the document path as the argument.

If document-review returns findings that were auto-applied, note them briefly when presenting handoff options. If residual P0-P4 findings were surfaced, classify them through `kb-gate` before proceeding.

When document-review returns "Review complete", proceed to Phase 10.

Run `kb-gate` before Phase 10 when document-review or your own checks surfaced P0/P1/P2/P3/P4 issues. Severity is not human-in-loop:

- Fix `auto_rectify` findings yourself before asking the user.
- Ask only for `needs_human` findings: product intent, scope acceptance, credentials/access, risky/destructive operations, or choosing between multiple reasonable architecture/product paths.
- P0/P1 block planning while unresolved.
- P2/P3/P4 do not block by severity alone, but fix cheap/actionable findings before planning when they improve requirements clarity.
- Do not stop just because findings existed. Stop only when unresolved findings would make `/kb-plan` invent behavior, accept risk blindly, or encode a decision the agent should not make.

### Phase 10: Handoff

#### 10.1 Phase Boundary

`kb-brainstorm` is complete when the requirements artifact is reviewed,
gate-clean, and ready to become slices. The next phase is `kb-plan`, but do not
assume every host will auto-chain skills.

Do not jump from brainstorm directly to `kb-work` or `kb-complete`.

If the user says "don't ask many questions", "go straight to work", "just build it", or similar, treat that as a request to compress Q&A and continue after planning. It does not authorize skipping the requirements artifact, `kb-plan`, or the manifest. Convert safe unknowns into explicit assumptions or `Deferred to Planning`, then invoke `kb-plan <requirements-doc>` with execution intent so planning can continue to `kb-work`.

If `Resolve Before Planning` contains any items:

- Ask the blocking questions now, one at a time, by default.
- If the user explicitly wants to proceed anyway, first convert each remaining item into an explicit decision, assumption, or `Deferred to Planning` question.
- If the user chooses to pause instead, present the handoff as paused or blocked rather than complete.
- Do not proceed to planning while `Resolve Before Planning` remains non-empty.

Before invoking `kb-plan`, write or expect a `brainstorm-to-plan` gate record
when the downstream artifact supports a gate ledger. Required evidence:

- requirements path;
- Question Gate classification completed;
- `Resolve Before Planning` empty;
- safe assumptions, deferred planning questions, and parked items recorded with
  rationale;
- document review findings resolved, deferred with rationale, or human-blocked.

If no blocking questions remain:

- If the user asked to continue straight to implementation, or an orchestrator
  such as `klfg`, `kb-epic`, or `kb-goal` called this brainstorm with execution
  intent, invoke `kb-plan <requirements-doc>` with that execution intent so
  planning can continue to `kb-work` after writing the manifest.
- Otherwise ask once: "Brainstorm is gate-clean. Continue with `kb-plan
  <requirements-doc>` now?"
- If the user says yes, invoke `kb-plan <requirements-doc>`.
- If the user says yes and asks to implement too, invoke `kb-plan
  <requirements-doc>` with execution intent so `kb-plan` continues to
  `kb-work <manifest-path>`.
- If the user says no, or the host cannot invoke the next skill, print the
  closing summary with the exact next command.
- If additional research is needed before planning, route to `kb-research`,
  update the requirements doc, then stop or return to the orchestrator.

#### 10.2 Handle the Selected Option

**Default:** Ask whether to start `/kb-plan <requirements-doc>` after the
requirements artifact is gate-clean, unless execution intent or an orchestrator
already authorized continuing.

**Stop conditions:** Pause instead of planning only when:

- unresolved `Resolve Before Planning` items remain;
- unresolved P0/P1 findings remain;
- unresolved P2/P3/P4 findings would change scope, acceptance criteria, risk, verification, or architecture direction;
- a human-only decision is required;
- targeted research must happen before planning;
- the user explicitly said to stop after brainstorm.

Never jump directly from brainstorm to `kb-work`.

**More Q&A:** Return to Phase 6 only when the user asks for more questions or the doc still lacks behavior, scope, acceptance criteria, or verification inputs.

When the user is satisfied with the additional Q&A, **do not jump straight back to Phase 10**. If the new conversation produced any change to requirements, scope, decisions, or success criteria, re-run Phase 8 (capture / update the requirements doc) → Phase 9 (document review) → Phase 10. Only short-circuit straight back to Phase 10 if the conversation purely confirmed existing decisions and added nothing new to the doc.

#### 10.3 Closing Summary

Use the closing summary when this brainstorm run is ending, pausing, handing off
without immediately invoking `kb-plan`, or the host cannot auto-chain the next
skill.

When complete and ready for planning, display:

```text
KB brainstorm complete!

Requirements doc: docs/brainstorms/YYYY-MM-DD-<topic>-requirements.md  # if one was created

Top research findings:
- [Finding 1]
- [Finding 2]

Key decisions:
- [Decision 1]
- [Decision 2]

Slice candidates: [count]
Confidence: [High/Medium/Low]

Next command: `/kb-plan docs/brainstorms/YYYY-MM-DD-<topic>-requirements.md`
To implement after planning: say "yes, then work" or run `kb-work <manifest-path>` after `kb-plan` creates the manifest.
```

If the user pauses with `Resolve Before Planning` still populated, display:

```text
KB brainstorm paused.

Requirements doc: docs/brainstorms/YYYY-MM-DD-<topic>-requirements.md  # if one was created

Planning is blocked by:
- [Blocking question 1]
- [Blocking question 2]

Resume with `/kb-brainstorm` when ready to resolve these before planning.
```

## Quality Checks

- [ ] Research happened **before** the first product question.
- [ ] The research brief was shown to the user before targeted Q&A.
- [ ] Every requirements claim about absent infrastructure was verified or labelled as an assumption.
- [ ] Decisions cite either a research source or are explicitly tagged as assumptions.
- [ ] The Slice Candidates section has 3–7 entries when recommending `/kb-plan`, or is omitted with reason.
- [ ] Confidence level in the research summary is honest about gaps.
- [ ] No implementation details leaked into the requirements doc (unless inherently technical).
- [ ] Document-review pass completed.
- [ ] All P0-P4 findings are resolved, deferred with rationale, or classified as human-only blockers before planning.

## Integration with Other Skills

- **Input from:** idea exploration or a fresh feature description from the user.
- **Default next step:** ask whether to continue with `/kb-plan <requirements-doc>` once the brainstorm is gate-clean, unless execution intent or an orchestrator already authorized continuing.
- **Stop instead:** only for unresolved blockers, required human decisions, required research, or explicit user instruction.
- **Optional follow-up:** `kb-research` for another targeted research pass that should become reusable local memory.
- **Document review:** Always run `document-review` before handoff (Phase 9).
- **Peer skill:** `/kb-research` — reusable research notes when the research itself should survive beyond this brainstorm.
