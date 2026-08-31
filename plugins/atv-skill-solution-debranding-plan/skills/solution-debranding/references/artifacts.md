---
title: Debranding Artifact Workflow
description: Progressive Markdown artifacts for navigable assessments, plans, recommendations, and validation
---

## Create the primary plan first

Unless the user explicitly requests chat-only output, create this path before discovery:

```text
<artifact-root>/YYYY-MM-DD-NNN-<run-name>-plan.md
```

`ARTIFACT_ROOT` defaults to `docs/debranding`. Honor a user-supplied repository-relative location when the repository has an established planning convention. One run produces one file. Follow the repository's sequence convention for `NNN`; use `001` when no earlier plan exists that day. Derive `<run-name>` from the repository or user-supplied scope and use lowercase kebab-case. If a matching active plan already exists, update it instead of creating a duplicate. Use repo-relative paths throughout.

Plan mode permits writing this artifact while prohibiting product, configuration, infrastructure, and operational changes. If the user explicitly says not to create files, set `ARTIFACT_MODE=chat-only` and state that no navigable artifact was created.

Set `ARTIFACT_MODE=split` only when the user asks for separate documents. It writes the same sections as `01-assessment.md` through `05-validation.md` inside `<artifact-root>/YYYY-MM-DD-NNN-<run-name>/`. Splitting is a filing preference, not a scope change: it never adds, removes, or reorders content, and the section contract below governs both shapes.

## Write progressively

Create the file with its section headings at the start, then fill each section as its evidence arrives. Do not wait until the final response.

| Stage | Section | Purpose |
|-------|---------|---------|
| Start | `## Summary` | Status, outcome, recommendation, blockers, and next action |
| Discovery | `## Assessment` | Scope, method, finding groups, exclusions, and evidence limitations |
| Discovery | `## Referrals` | Findings owned by another track, with urgency and suggested owner |
| Planning | `## Implementation units` | Ordered units, dependencies, acceptance checks, and owners |
| Branding | `## Branding` when in scope | Canonical profile, delivery mechanisms, constraints, and two-profile validation |
| Modernization | `## Modernization` when in scope | Capability requirements, options, recommendation, migration, and rollback |
| Public release | `## Release readiness` when in scope | Blockers, history status, external surfaces, licensing, and clean-clone checks |
| Apply or verify | `## Validation` | Commands, results, residual findings, failures, and verdict |

Omit a section entirely when it is out of scope. An empty heading reads as unfinished work rather than as work that was never required.

## Give every fact one owning section

A fact stated in three places is three things to keep in sync and one thing the reader stops trusting. One file does not relax this rule, it sharpens it: repetition that was merely annoying across documents is visible padding within one. Each of the following has exactly one owner, and everywhere else cites it by identifier and anchor.

| Fact | Owning section | Everywhere else |
|------|----------------|-----------------|
| Release blockers and their evidence | `## Release readiness` when public release is in scope, otherwise `## Summary` | Cite the blocker identifier |
| Coverage, counts, method, and evidence limits | `## Assessment` | Cite the number once in the summary if it drives the recommendation |
| Brand profile keys and validation design | `## Branding` | Link from the units that consume the profile |
| Capability options, migration, and rollback | `## Modernization` | Link from the units that replace the dependency |
| Findings referred to another track | `## Referrals` | Name the count in the handoff |

The summary carries the decision, the identifier, and the reason it blocks. It does not restate the evidence. The assessment carries the evidence and the grouping. It does not re-narrate the blockers as a second list.

## Keep the plan decisive

Keep the opening summary under 40 lines. Front-load:

1. Status and one-sentence outcome
2. Recommended next action
3. Blocking decisions or approvals
4. A section index with anchors
5. At most three next-step options, with one marked as recommended

Follow the summary with the ordered implementation units, then the supporting sections. A reader who stops after the summary should still know what to do next.

## Keep the plan executable

The plan should contain:

* Problem frame and explicit scope boundary
* Decisions and rationale
* Ordered implementation units
* Repo-relative files or owning areas
* Dependencies and gates
* A reference column naming the one section anchor that unit needs
* Specific acceptance and validation checks
* Human-review owner where automation is unsafe

The reference column exists so `solution-debranding-apply` can execute a unit after reading one section instead of the whole plan. Point at a heading: `#modernization--capability-1-model-inference`. Leave it empty when the unit row and its acceptance check are genuinely self-contained.

Prefer one table of implementation units over a long file-by-file narrative. Keep detailed evidence in `## Assessment` and specialized analysis in its own section. Keep task checkboxes compatible with automated execution. Do not include terminal transcripts, tool narration, or raw scanner output.

## Handle machine evidence separately

Write scanner and readiness JSON to a temporary directory by default. Record the command, schema version, counts, and limitations in the plan. Persist raw JSON beside the plan only when the user requests it or another automated workflow consumes it.

Every intermediate file follows the same rule. Batch plans, per-batch lens responses, merged coupling output, and any helper script belong in a system temporary directory, never in the repository being debranded. A run that leaves a `.tmp-*` file or a one-off parser in the working tree has modified the repository it promised not to touch, and the user has to clean up after it.

Never reproduce sensitive matched values in Markdown.

## Use standard frontmatter

Every generated Markdown artifact must begin with:

```yaml
---
title: <descriptive title>
description: <one-sentence purpose>
status: draft
date: YYYY-MM-DD
mode: audit | plan | apply | verify
---
```

Use `status: active`, `blocked`, or `complete` as the run progresses. Do not add an H1 when `title` is present.

## Keep chat brief

The final chat response must be no more than eight short lines of prose unless the user asks for details. Include only:

* Outcome
* Primary recommendation
* Blocker count or most important blockers
* Link to the plan
* One explicit next action

Do not paste document sections, finding tables, command logs, or generic offers such as "say the word." The artifacts hold the detail.

In `plan` mode, close with the unit ledger that [the reporting contract](./reporting.md) defines. The ledger does not count against the eight lines: it is the handoff itself, not commentary on it. A response that links five documents and names no units asks the user to read the plan before they can answer the only question the plan exists to raise, which is what to do first.
Name the referral count in the same breath. A reader who owns one track needs to know the run saw the others and wrote them down, or they will assume the scan missed them and pay for a second one.
