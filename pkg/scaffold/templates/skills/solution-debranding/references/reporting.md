---
title: Debranding Reporting Contract
description: Finding taxonomy, artifact routing, concise handoff, and completion criteria
---

## Assign dispositions

Use one of these dispositions before remediation type:

| Disposition | Meaning |
|-------------|---------|
| `Blocks release` | Confirmed exposure, release-rights problem, broken portability, or required migration prevents the requested outcome |
| `Needs your decision` | Identity, authorization, certificates, networking, external systems, ownership, legal judgment, or deployed infrastructure requires approval |
| `Improvement` | Optional quality, cost, portability, or operational improvement that does not block the requested outcome |

Disposition describes consequence. Remediation type describes the action. Keep both fields.

## Route work to its track

Debranding, deployment, and modernization are often three people working in parallel, and a plan that mixes them gives each of them someone else's homework. Assign every finding one track:

| Track | Owns |
|-------|------|
| `debranding` | Literal identifiers, indirect attribution, brand parameterization, visual identity |
| `release` | Publication rights, licensing, history certification, external surfaces |
| `deployment` | Getting the solution to build, run, and deploy as it stands |
| `modernization` | Portability, dependency replacement, platform migration |
| `security` | Vulnerabilities and exposure that are wrong regardless of who publishes it |

Only tracks named in `TRACKS` produce implementation units. Everything else is recorded once in the plan's `## Referrals` section and named in the handoff:

| ID | Finding | Track | Urgency on that track | Suggested owner |
|----|---------|-------|-----------------------|-----------------|
| SEC-004 | Tokens decoded without verifying signature, issuer, or audience | `security` | High | Platform security |

Urgency in that table belongs to the receiving track, not to this run. A disposition answers "does this block the outcome I was asked for", so a severe authentication defect that does not prevent publication is `Improvement` for debranding, and reporting only that disposition tells the reader it is minor. State the referred urgency separately, and never downgrade a finding because it landed outside the current scope.

Do not open referred findings as units, do not remediate them in `apply`, and do not let them block debranding completion. Hand them over.

## Open the assessment with an exposure map

A finding table answers what is wrong; it does not show where the exposure concentrates. Lead `## Assessment` with the annotated tree the merge script renders:

```bash
python scripts/merge-lens-findings.py lens-results/*.json --report scan-report.json --batch-plan lens-batches.json --tree
```

```text
exposure map (depth 3; B=blocks G=gated T=tune)
  (root)/                                     B1
  sample-data/                                B6
    imports/                                  B6 unread=4
  src/                                        B4 G2 T18
    api/                                      B1 T3
    frontend/                                 G1 T15 unread=12

largest unexamined directories
  docs/images/                                unread=8
```

The tree carries only directories that hold findings, so it stays the shape of the problem rather than a second file listing. The two annotations a finding table cannot show are `unread`, which counts files no lens opened inside a directory that does have findings, and the unexamined list, which names directories with no findings and no coverage. A directory with no findings and a high unread count is not clean, it is unexamined, and the reader needs that distinction before trusting a coverage percentage.

Keep the depth at three. Depth exists to show where exposure concentrates, and a tree that runs past one screen has become a second finding table.

## Plan each finding

For every finding, include:

* Stable identifier, file path, and line or location
* Description, redacted evidence, confidence, and risk level
* Disposition and remediation type
* Proposed change, dependencies, and possible breaking effects
* Brand delivery mechanism when applicable
* Required capability and modernization recommendation when applicable
* Human-review requirement and suggested owner

Group the plan by literal references, indirect identifiers, sensitive material, neutral redesign, parameterization, generalization, removal, renames, binary content, generated content, breaking changes, human review, modernization, and brand-profile work.

## Apply safely

In `apply` mode, implement low-risk repository-local changes after planning. Validate the touched behavior immediately. Do not modify external systems, deployed resources, CI/CD secrets, registries, issues, pull requests, or releases without explicit authorization.

Keep the current provider available during staged modernization when compatibility requires it. Avoid editing generated or vendored content directly when regeneration is the correct mechanism.

## Produce navigable artifacts

Follow [the artifact workflow](./artifacts.md). Treat the single execution plan as the durable output, and chat as a concise handoff.

Keep the plan's summary and units focused on decisions, ordered implementation units, dependencies, owners, acceptance checks, and next actions. Route supporting material to sections of the same file:

* Literal, indirect, sensitive, and behavioral evidence to `## Assessment`
* Findings owned by another track to `## Referrals`
* Brand profile and two-profile design to `## Branding`
* Capability options, recommendation, migration, and rollback to `## Modernization`
* History, external surfaces, licensing, and outsider checks to `## Release readiness`
* Executed validation, failures, and residual findings to `## Validation`

Do not duplicate the same finding table across sections. Give every fact one owning section as [the artifact workflow](./artifacts.md) requires, and cite it from elsewhere by identifier and anchor. Restating the release blockers in the summary, the assessment, and the readiness section is the most common form of this, and it is the one that ages worst. Keep the plan's opening summary under 40 lines and the final chat response under eight short lines of prose unless the user asks for detail.

## Close the plan with a unit ledger

A `plan` run ends by asking the user what to do first, so it must show them enough to answer without opening a file. After the prose, print every implementation unit in recommended order:

| # | Unit | Disposition | Depends on | Size | Your call |
|---|------|-------------|------------|------|-----------|
| 1 | Rotate and remove committed credentials | `Blocks release` | — | S | No |
| 2 | Confirm outbound licence and publication rights | `Needs your decision` | — | S | Yes |
| 3 | Parameterize brand profile | `Improvement` | 1 | M | No |

Add a `Track` column when the run owns more than one track, and leave it out when every unit shares one, because a column with a single repeated value costs width and tells the reader nothing. Follow the ledger with the referral count so the other owners' work is visible without competing for position in the list.

Use the unit numbers `solution-debranding-apply` accepts, so a row is directly runnable as `unit=N`. Size is blast radius, not duration: `S` touches a handful of files with no behavioral change, `M` spans a subsystem or changes configuration shape, `L` crosses subsystems or risks breaking consumers. Mark `Your call` as `Yes` only when the unit cannot start until a human decides something the run cannot decide for them, such as licensing, data ownership, or accepting a residual risk.

Order by what unblocks the most work, not by severity alone. Put `Needs your decision` units that need an outside answer near the front even when their size is small, because their lead time runs in parallel with mechanical work, and a run that sequences them last stalls at the end for an answer it could have requested at the start.

Close by inviting a sequence rather than announcing one: the recommended order is a default the user can overrule, and they may know about constraints the repository does not record. Offer `unit=N` to start a specific unit, and accept a reordered list.

When Git history is excluded, state:

> Working-tree content was reviewed. Git history was not rewritten or certified as debranded.

## Completion criteria

Complete the task only when:

* No known source-customer references remain in reviewed scope except documented exclusions
* Indirect identifiers and customer-owned content have been reviewed
* Customer-dependent values are centralized or parameterized where practical
* Customer-specific concepts have neutral alternatives or explicit review items
* Code, configuration, tests, infrastructure, and documentation remain consistent
* Available validation passes or failures are reported accurately
* Future branding changes use the documented profile rather than repository-wide edits
* Two fictional brand profiles have been validated or the limitation is explicit
* Customer-specific dependencies have portable alternatives, compatibility adapters, or blockers
* Operation without source-customer infrastructure is verified or every remaining dependency is documented
* Scope boundaries, external surfaces, residual risk, and confidence are transparent

Referred findings do not block completion of the tracks this run owns. Report them as handed over, with their track and suggested owner, and do not describe the run as incomplete because another workstream has open work.

Never claim completion from one literal search, a passing linter, or an agent report without independent evidence.

