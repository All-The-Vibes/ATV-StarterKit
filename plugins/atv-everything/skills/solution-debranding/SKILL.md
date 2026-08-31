---
name: solution-debranding
description: Shared workflow package for the solution-debranding-plan, solution-debranding-apply, and solution-debranding-verify skills. Holds the stage contracts, references, scripts, and evidence rules they all depend on.
user-invocable: false
disable-model-invocation: true
---

# Solution Debranding

Shared workflow package. The `solution-debranding-plan`, `solution-debranding-apply`,
and `solution-debranding-verify` skills are the entry points. Follow the invoking
skill's fixed mode and stage contract. Never widen the mode during a run.

## Purpose

Transform a brand-specific repository into a reusable, brand-agnostic solution without disguising coupling or weakening technical correctness.

The workflow must:

1. Remove direct and indirect brand-identifying information.
2. Preserve behavior and reference consistency.
3. Separate presentation branding from tenant, security, deployment, and immutable identity.
4. Make branding interchangeable through canonical configuration.
5. Model brand-specific services by capability and propose portable alternatives.
6. Produce redacted, auditable evidence and transparent residual risk.

## Inputs

Resolve these values from the request and repository:

* `SOURCE_CUSTOMER`: Identity to remove
* `TARGET_CUSTOMER`: Optional example replacement for content that must render a name, such as brand profiles, sample data, and demo content; default to `Contoso` only when an example is required. Anonymize prose with a neutral role instead
* `MODE`: `audit`, `plan`, `apply`, or `verify`
* `SCOPE`: Entire repository unless explicitly limited
* `BRAND_SCOPE`: Organization only by default, or organization plus explicitly named product, project, program, and code names when the user requests that broader scope
* `MODERNIZATION_SCOPE`: Brand-specific dependencies to assess, recommend alternatives for, or replace, or `none` when another workstream owns portability
* `TRACKS`: Workstreams this run owns, defaulting to `debranding` plus `release` when `RELEASE_SCOPE` is external or public
* `RELEASE_SCOPE`: Internal reuse, external sharing, or public release
* `ARTIFACT_MODE`: `single` by default, `split` when the user asks for separate companion documents, or `chat-only` when the user explicitly forbids file creation
* `ARTIFACT_ROOT`: Repository-relative plan location, defaulting to `docs/debranding`; honor an established repository planning location or a user-supplied override
* `DESIGN_SYSTEM`: Optional URL or name of a target design system (such as Fluent 2, Carbon, or Material) whose tokens define the replacement brand's palette, typography, and spacing. When the reference is a visual inspiration source rather than a structured token system, classify it as visual and require human approval of the derived palette
* `RUN_NAME`: Optional stable slug for the artifact directory

Infer values that repository evidence can establish safely. Preserve existing product names, product abbreviations, and product identity when the user requests only organization debranding. Do not infer broader `BRAND_SCOPE` merely because a product name appears beside the source-customer logo or wordmark. A new product name, short name, acronym, monogram, logo concept, or icon treatment is a design decision: require an explicit user-provided value or record it as a human decision in the plan. Example values in this package illustrate configuration shape only and are never approved replacement values. Ask one focused question only when ambiguity changes ownership, scope, safety, or required remediation.

Narrowing scope narrows the work, never the observation. A run scoped to one track still reports what it noticed elsewhere, because the scan already paid for that evidence and the owner of the other track has no other way to learn it. Route those findings as [the reporting contract](./references/reporting.md) requires instead of dropping them or promoting them into this run's plan.

The invoking skill fixes `MODE`. If a request reaches this package without one:

* Use `plan` for analysis, assessment, recommendations, or review, and whenever the request is ambiguous.
* Use `apply` only when the user names an existing primary plan or explicitly asks to execute approved plan work.
* Use `verify` when asked whether debranding is complete or correct.
* Use `audit` only when the user explicitly forbids artifact creation.

A request to debrand, remove, replace, or convert resolves to `plan`, never to `apply`. Entering `apply` requires an existing primary plan on disk that the user has pointed to. When no such plan exists, run `plan` and stop. Treat the absence of an explicit mode as `plan`.

## Operating Modes

| Mode | Behavior |
|------|----------|
| `audit` | Inspect and report without changing repository content |
| `plan` | Inspect and produce a file-level remediation and modernization plan |
| `apply` | Inspect, plan, change, test, and validate authorized repository content |
| `verify` | Re-scan and validate without remediation unless explicitly requested |

In `audit`, `plan`, and `verify` modes, writing requested report artifacts does not count as changing solution code. Interpret "do not modify the repository" as "do not modify product, configuration, infrastructure, or operational files" unless the user explicitly says not to create files.

## Stage Contracts

Each mode has a fixed contract. Follow the resolved mode's contract and no other.

### Plan contract

1. Resolve the source brand, scope, brand scope, modernization scope, release scope, artifact mode, and run name from the request and repository evidence.
2. Establish a pre-change baseline from the documented setup on the current host. Record the operating system and relevant tool versions, use the declared package manager with its immutable or frozen lockfile mode, then run the cheapest representative install, build, test, startup, and health check that the environment permits. Record pre-existing failures without repairing product code in plan mode.
3. Create or update the canonical plan under `ARTIFACT_ROOT`.
4. Decide planning-record disposition early. The plan must state, per directory, whether planning and history records will be deleted, anonymized, or excluded from the public tree. This decision dominates the finding count and blocks verify if left open.
5. Stop after planning. Do not modify product, configuration, infrastructure, operational, external-system, or Git-history content. Leave no scratch files behind in the repository.
6. Close by naming the primary plan path, presenting the ordered unit ledger, and stating that applying any unit requires explicit approval.

### Apply contract

1. Require one primary plan path. If none is supplied, discover a unique active plan under `ARTIFACT_ROOT`; otherwise stop and request the path.
2. Read the plan before changing anything.
3. Confirm the requested unit is unblocked and that required legal, security, privacy, ownership, or external-system approvals are recorded. Never infer an approval from the plan's existence.
4. When a gated unit needs a human answer, pose the question and continue with unblocked units. Do not stop the run to wait for an answer that can arrive while mechanical work proceeds.
5. Execute only the selected unit, or the next unblocked unit when none is named. Referred findings are not units: decline them and name their track owner.
6. After each term replacement, search the full tree for residual matches of that term across all case variants before proceeding to the next term. Do not rely on path-scoped bulk replacements to cover the entire repository.
7. After the first substantive edit, run the cheapest focused validation. Finish with the relevant tests, builds, scans, or rendering checks for that unit.
8. Update the same plan's checkboxes, status, and `## Validation` section with commands, outcomes, limits, and residual findings.
9. Stop at a failed acceptance check or a completed unit. Do not create a second plan or repeat full discovery unless the plan requires it.

When validation exposes a pre-existing setup or runtime failure outside the selected unit, record it against the baseline and refer it to the appropriate track. Do not silently add dependencies, regenerate a lockfile, change bundlers, add platform launchers, disable readiness or security checks, or substitute feature-degrading stubs merely to make debranding validation pass. Such work requires its own approved unit and acceptance criteria.

After every apply run, the user must run verify. Do not treat a passing apply as sufficient evidence that the unit is clean.

### Verify contract

1. Resolve the primary plan from the supplied path or a unique active plan under `ARTIFACT_ROOT`.
2. From a clean worktree or clone, repeat the documented setup with the declared package manager's immutable or frozen lockfile mode. Run full debranding and readiness scans plus the checks relevant to detected repository surfaces. For runnable applications, include startup and health smoke checks. Render configuration or infrastructure only when present. Validate two fictional brand profiles only when branding is in scope.
3. For external or public release, verify the recorded history-scan decision, clean-clone workflow, and disposition of external surfaces. Do not modify those surfaces. If history files with branded content remain in the tree and the plan has no recorded disposition decision for them, block the verdict.
4. Verify only the tracks the plan owns. Report referred findings as handed over and unverified, and name the tracks the verdict covers so a pass is not read as a whole-repository clearance.
5. Update the plan's `## Validation` section and status with reproducible evidence, exclusions, residual risk, and a clear pass, blocked, or failed verdict.
6. Do not fix failures. Report them and stop.

State the operating system and runtime versions actually exercised. Never claim cross-platform support for hosts that were not tested. A current-host startup failure blocks a runnable or release-ready verdict unless that host is explicitly unsupported and documented as such.

## Artifact-First Output

Read [the artifact workflow](./references/artifacts.md) before discovery. Unless `ARTIFACT_MODE=chat-only`, create one execution-compatible plan:

```text
<artifact-root>/YYYY-MM-DD-NNN-<run-name>-plan.md
```

`ARTIFACT_ROOT` defaults to `docs/debranding`. One run produces one file. Create its section headings at the start, then fill each section progressively as its evidence arrives. Keep detailed findings, modernization analysis, branding design, release readiness, and validation evidence in their sections rather than the chat response. Split into companion documents only when the user asks with `ARTIFACT_MODE=split`.

Chat responses must stay brief: outcome, top recommendation, blockers, plan link, and the next decision or command. Do not paste the full plan, complete finding tables, command transcript, or scanner output into chat.

## Workflow

### 1. Establish safety and scope

Read repository instructions and determine whether public release is requested or plausible. If it is, read [the public-release workflow](./references/public-release.md) and perform its sensitive-history gate before ordinary debranding.

Never expose sensitive values. Never modify external systems, deployed resources, credentials, identity objects, registries, releases, issues, pull requests, or history unless explicitly authorized.

### 2. Discover evidence

Read [the discovery workflow](./references/discovery.md). Establish the identity map, inspect the owning abstractions, and run the scanner from the directory containing this `SKILL.md`:

```bash
python <skill-root>/scripts/scan-debranding.py \
  --root <repository> \
  --source-customer <name> \
  --output debranding-findings.json
```

Use `--alias`, `--domain`, and `--resource-prefix` for confirmed identifiers. Use `--mode delta --base <ref>` only after a full baseline exists.

The scanner sorts every file into one of three surfaces. `product` is the code under review and the coupling denominator. `history` is planning and decision records, which are branded by design and never rewritten in place. `tooling` is agent tooling that ships with the repository, which needs branding but carries no product coupling. Adjust the defaults with `--history-path` and `--tooling-path`, and confirm `coverage.files_by_surface` before trusting any count.

The scanner supplies redacted evidence, stable identifiers, and deterministic ordering. It does not replace semantic review, binary inspection, generated-content review, or a dedicated history scanner.

Add `--summary` to print a rollup of counts by category and surface, the files carrying the most findings, and the suspicious-file list. Use it to read the scan without loading the full JSON into context, and never hand-roll a parser to produce the same view.

Run [the coupling review](./references/coupling-review.md) after the scan when the outcome is reuse, resale, handover, or public release. The scanner cannot detect coupling that survives renaming, and a clean scan is not evidence of portability.

The coupling review runs against a generated batch plan that assigns every reviewable product file to exactly one lens. Use `--split-dir` to write one file per batch, dispatch one subagent per batch, and do not consolidate batches. Subagents cannot write files, so transcribe each returned JSON response to its own file yourself before merging. The merge script exits non-zero when a batch returns fewer files than it was assigned. Re-dispatch the named batches rather than continuing with partial coverage, and record `file_coverage.coverage_percent` in the plan's `## Assessment` section. Read the merged output with `--summary`, and read the blocking couplings with `--blocks` when drafting the plan. Never write the plan from a run that did not reach the merge step.

Summarize discovery in the plan before continuing. Do not defer all document writing until the final response.

### 3. Classify findings

Assign every finding both a disposition and remediation type.

| Disposition | Consequence |
|-------------|-------------|
| `Blocks release` | Prevents the requested reuse or release outcome |
| `Needs your decision` | Requires authorization or specialist judgment |
| `Improvement` | Improves quality, cost, portability, or operations without blocking the outcome |

Use direct replacement, parameterization, generalization, neutral redesign, removal, human review, or modernization candidate as the remediation type. Do not rename a coupled implementation and classify it as portable.

### 4. Design interchangeable branding

Read [the branding contract](./references/branding.md) when branding is in scope. Use one canonical profile or a documented source-of-truth hierarchy, separate display values from infrastructure and identity, and classify values as runtime, build-time, deployment-time, or generated.

When `DESIGN_SYSTEM` is set, identify the concrete consumer before proposing token infrastructure. Read the system's official documentation and prefer its supported theme provider, generation API, or semantic aliases over manual primitive assignment. Record the consumer, mapping method, and gaps in the plan's `## Branding` section. If no consumer exists, ask one focused scope question or record adoption as a separately gated unit.

Validate at least two distinct fictional profiles. Switching profiles must not require repository-wide source edits. Document rebuild, regeneration, redeployment, immutable names, and intentional stable identifiers.

### 5. Modernize brand-specific capabilities

Read [the modernization workflow](./references/modernization.md) for brand-specific gateways, identity mechanisms, platform services, internal APIs, data stores, queues, proxies, observability, or deployment dependencies.

Describe the required capability before comparing products. Use current official sources for changing product claims. Prefer application-owned interfaces and provider adapters. Preserve the current provider until contract compatibility, rollout, and rollback are validated.

### 6. Plan before applying

Read [the reporting contract](./references/reporting.md). Produce a file-level plan with evidence, dependencies, breaking effects, delivery mechanism, human-review owner, and validation strategy. Keep the summary and unit table concise and move supporting detail into the plan's later sections.

In `plan` mode, stop here. In `apply` mode, implement only authorized repository-local changes. After the first substantive edit, immediately run the cheapest focused check that could falsify the change.

### 7. Validate independently

Re-run discovery in full mode after applying changes, and make unresolved blockers fail the command:

```bash
python <skill-root>/scripts/scan-debranding.py \
  --root <repository> \
  --source-customer <name> \
  --fail-on-blocking \
  --summary
```

Pass every confirmed `--alias`, `--domain`, and `--resource-prefix` from the identity map. Do not treat exit code 0 from a scan without `--fail-on-blocking` as completion evidence. Then run the repository's relevant tests, builds, linters, type checks, configuration rendering, infrastructure validation, and two-profile branding checks.

For public release, run:

```bash
python <skill-root>/scripts/check-readiness.py \
  --root <repository> \
  --output release-readiness.json
```

Treat readiness indicators as prompts unless a missing item creates a concrete security, legal, operational, or release blocker. Validate the candidate from a clean clone as an outsider.

### 8. Report the outcome

Follow [the reporting contract](./references/reporting.md). Update the primary plan with the final status and next action. In chat, provide no more than eight short lines of prose unless the user asks for detail. In `plan` mode, follow that prose with the ordered unit ledger so the user can choose what to run first without opening the plan.

Do not claim complete debranding from one literal search, a passing linter, or an unverified agent report.

## Non-Negotiable Rules

1. Do not perform blind global search and replacement. In particular, never substitute inside URLs, hostnames, domain names, email addresses, DNS records, package names, container orchestration namespaces, or other structured identifiers. A replacement that produces a malformed URL leaves the source organization's domain perfectly legible in a broken form. Replace the entire structured value with a placeholder or parameterize it.
2. Do not turn the target example into an architectural dependency.
3. Do not hide brand-specific behavior behind a neutral name.
4. Do not invent production domains, tenant IDs, subscriptions, endpoints, credentials, or security settings.
5. Do not remove legitimate third-party names, package names, licenses, standards, or required attribution.
6. Do not expose secrets, personal information, or sensitive identifiers in output.
7. Do not rewrite history or modify external systems without explicit scope and authorization.
8. Treat authentication, authorization, certificates, network trust, CI/CD, cloud identity, legal ownership, and deployed infrastructure as `Needs your decision` by default.
9. Use current official evidence when product capabilities, support, pricing, availability, or licensing affect modernization advice.
10. Report every unresolved or suspected reference and every surface that was not validated.
11. Do not write scratch files into the target repository. Analysis scripts, lens responses, batch plans, and intermediate JSON belong in a system temporary directory. The only files this skill creates under the repository are the artifact set, plus product edits in `apply` mode.
12. Use the packaged scripts under `scripts/` for scanning, batching, and merging. When one is missing a capability, say so and stop rather than improvising a replacement parser.
13. Read a script's results from its `--summary` output. Never quote a result obtained by hand-indexing the report JSON. A wrong key returns nothing rather than failing, and nothing reads as a clean result, so a typo becomes a false all-clear that no later step catches.
14. Before citing a check as evidence, name what would have made it fail. A command that exits zero whatever the code does is not evidence, and reporting it as evidence is worse than reporting nothing, because it closes the question. When no failing case exists, record the check as run and the claim as unproven.
15. One unit at a time. Validate each unit before starting the next. Do not skip validation to batch edits across units.
16. Re-run every validation after the final edit. A check result from before the last edit is stale and must not be reported as the outcome. Report the literal output of the final run.
17. Preserve dependency and runtime integrity. Do not accept manifest additions without direct imports or documented peer-dependency evidence, broad lockfile or registry churn without explanation, development-only bypasses of production checks, or stubs that remove user-visible capability as incidental debranding changes.
18. After deleting or removing any file, verify that no CI pipeline, configuration, import, documentation link, or license reference still points at it. A deletion that leaves dangling references is a half-done removal.
19. When scrubbing a value from an environment template, trace it to every other representation: code-level defaults, test fixtures, documentation, and CI configuration. A scrubbed env template with an intact code fallback carries the original value into every deployment that does not override it.
20. When replacing a visual asset with a different format, verify that the consuming framework accepts the new format without additional configuration. A PNG-to-SVG swap that breaks a framework image optimizer is a production defect, not a completed debranding.

## Completion Contract

Complete the requested mode only when its evidence is reproducible and its limits are explicit. For an `apply` or `verify` completion claim, require:

* No known source-brand references in reviewed scope except documented exclusions
* Review of indirect identifiers, brand-owned content, and sensitive artifacts
* Centralized or parameterized brand-dependent values where practical
* Two-profile brand validation or a documented blocker
* Portable alternatives, adapters, or explicit blockers for brand-specific dependencies
* Consistent code, configuration, tests, infrastructure, generated content, and documentation
* Passing relevant validation or accurately reported failures
* Explicit status for Git history and every external release surface
* Transparent residual risk, confidence, and human-review requirements