---
title: Debranding Discovery Workflow
description: Evidence-driven discovery and classification for organization-specific repository content
---

## Establish the identity map

Record confirmed and suspected identifiers before deciding what to change:

* Source organization names, aliases, abbreviations, domains, and email patterns
* Product, project, program, and internal code names, including initiative codenames, phase names, dataset labels, geographic qualifiers, and regional business-unit names that narrow which entity built the solution — even when the organization name is absent. When a country-level qualifier is confirmed, expand the term list to include its sub-national identifiers: state or province names, major cities, currency codes, regulatory bodies, and other jurisdiction-specific defaults that the solution hardcodes
* Owner-specific terminology, taxonomies, workflows, and file formats
* Resource prefixes, namespaces, registries, ingress hosts, and deployment names
* Container orchestration namespaces, cluster-internal DNS, and service mesh identifiers
* Tenant, subscription, organization, identity, role, and service-connection identifiers
* Personal names, ownership metadata, copyright, and attribution
* Legitimate third-party products, standards, package names, and licenses that must remain

Pass every confirmed codename, namespace, and cluster identifier to the scanner as `--alias` or `--resource-prefix`. A codename that names a directory, a CI matrix entry, a test fixture, or a data file survives every rename of the organization and identifies the source on its own.

To discover terms the user did not supply, read the README, eval and test configuration files, fixture and sample-data directories, and CI pipeline configuration before the first scan. Project codenames, entity names, and geographic identifiers concentrate in these locations and may not be mentioned in the initial request.

Ask one focused question only when ownership or scope cannot be inferred safely and the answer changes remediation. Record inferred scope otherwise.

## Select the scan mode

Use a full scan on the first run, after identity-map changes, or before a release verdict. Use a delta scan only after a full baseline exists and the user wants to evaluate subsequent changes.

Run the deterministic scanner:

```bash
python scripts/scan-debranding.py \
  --root <repository> \
  --source-customer <name> \
  --alias <abbreviation> \
  --domain <domain> \
  --resource-prefix <prefix> \
  --output debranding-findings.json
```

For a delta scan:

```bash
python scripts/scan-debranding.py \
  --root <repository> \
  --source-customer <name> \
  --mode delta \
  --base origin/main \
  --output debranding-findings.json
```

The script is an evidence collector, not a complete verdict. Supplement it with semantic inspection, binary and image review, generated-content analysis, and dedicated secret scanning.

Discovery scans exit successfully when they find blockers because findings are their output. For the final full scan in `apply` or `verify` mode, add `--fail-on-blocking`; the command then exits 1 when any `Blocks release` finding remains. A successful scan without this flag proves execution, not absence of blockers.

## Inspect repository surfaces

Inspect all relevant tracked, untracked, generated, ignored, archived, minified, cached, binary, and Git LFS content when tools and scope permit. Cover:

* Source, comments, tests, fixtures, mocks, samples, API payloads, and logs
* File and directory names, imports, namespaces, package metadata, and public APIs
* Documentation, diagrams, screenshots, icons, logos, manifests, titles, and email templates
* Environment templates, settings, feature flags, telemetry labels, and build outputs
* Docker, Compose, Helm, Kubernetes, Terraform, Bicep, ARM, Pulumi, and deployment scripts
* GitHub Actions, Azure DevOps, service connections, secret references, and release automation
* Registries, image names, DNS, ingress, certificates, proxies, managed identities, and role names
* UUID-shaped literals in every text format, including source, configuration, Markdown, examples, and tests
* Tenant-specific authentication authorities such as `login.microsoftonline.com/<tenant>`

Do not treat zero literal matches as proof that indirect identifiers or customer-specific behavior are absent.
The deterministic scanner treats every UUID-shaped literal as an opaque identifier
blocker and redacts its value. Classify each match, then parameterize customer-bound
identity values and verify the resulting login flow with a fictional tenant profile.
Do not replace matches with invented production IDs. A benign UUID may be retained only
through an explicit, documented review decision rather than an implicit scanner exclusion.

Run [the coupling review](coupling-review.md) to cover the behavior that literal matching cannot reach.

## Review environment templates holistically

Environment templates (`.env.example`, `.env.sample`, `.env.template`, and similar) carry real infrastructure values by design: server FQDNs, database names, managed-identity names, client IDs, cluster URLs, and registry paths. A brand-string scan catches only the values that happen to contain the customer name. Every other value in the file is also customer-specific and also ships on release.

Review each value in every environment template for source-specific infrastructure, not just for brand-string matches. Scrubbing one value but leaving an equally identifying value directly below it is a partial fix that exposes the same origin.

## Trace scrubbed values to all representations

A configuration value rarely lives in one place. An infrastructure URL, a namespace, or a managed-identity name typically appears in the environment template, a code-level default or fallback, test fixtures, documentation, and CI configuration. Scrubbing one representation without the others leaves the value exposed.

After scrubbing any value, search the full tree for every other occurrence of the same literal. Pay particular attention to:

* Code-level defaults and fallback values in settings or configuration classes
* Test fixtures, mocks, and assertion strings that embed the same value
* Documentation, runbooks, and planning records that quote it
* CI/CD pipeline configuration that passes it as an input or matrix entry

## Check reference integrity after deletions

When discovery or apply removes files, immediately check that nothing still references them. Search for each deleted filename and its path fragments across the full tree. Record each orphaned reference as a finding and plan its cleanup in the same unit.

## Inventory customer-specific capabilities

For each external dependency or internal platform service, record:

* Capability supplied and dependent workflows
* Call sites, protocols, SDKs, schemas, identity flows, and network assumptions
* Availability, reliability, latency, compliance, observability, and cost constraints
* Required, replaceable, optional, or obsolete status
* Existing interfaces, adapters, factories, dependency injection, and feature flags
* Failure behavior when the dependency is unavailable

Load [the modernization workflow](modernization.md) when the dependency affects portability.

## Use evidence-based findings

Each finding must include:

* Stable finding identifier
* Path and line or location when available
* Identifier category and remediation type
* Confidence level: high, medium, or low
* Redacted evidence and the detection method
* Disposition: `Blocks release`, `Needs your decision`, or `Improvement`
* Proposed action, dependencies, and possible breaking effects
* Human-review owner when automation is unsafe

Use high confidence for exact confirmed identifiers, medium confidence for contextual or structural evidence, and low confidence for naming heuristics that require review. A low-confidence finding is not a low-risk finding.

## Separate planning and history artifacts

Repositories that use plan-driven engineering accumulate planning records under `docs/`, in folders such as `plans`, `pm`, `analysis`, `brainstorms`, `ideation`, `solutions`, and `verification`, alongside the records this workflow writes to `ARTIFACT_ROOT` (default: `docs/debranding`). These records name the customer by design, and they are the wrong target for debranding edits.

Treat them as a distinct surface:

* Scan them, because they carry customer identity into any published tree
* Tag them `surface=history` and keep them out of product remediation counts
* Never resolve them by direct replacement, parameterization, or generalization alone, because a mechanical edit gives false confidence over records whose confidential substance no term list detects
* Route every one of them to human review with a keep, anonymize, relocate, or remove decision
* Record the decision as a documented exclusion instead of a completed remediation

Tagging a record `surface=history` says it is reviewed separately, not that it is disposable. Where the repository practices plan-driven engineering, these records are the compounding asset and the tooling that ships with the repository reads them, so removal is a decision with a cost on the other side. [The public-release workflow](./public-release.md) owns that decision.

Not all history records carry the same reuse value. Reusable planning records (requirements, implementation plans, design decisions, compounding solutions) are usually worth anonymizing and keeping. Project-specific records such as validation reports, project-management artifacts, dashboards, and verification logs rarely transfer — recommend deletion when the release scope is external or public.

The scanner tags `docs/debranding`, `docs/plans`, `docs/pm`, `docs/analysis`, `docs/brainstorms`, `docs/dashboard`, `docs/ideation`, `docs/solutions`, and `docs/verification` by default. Use `--history-path` when the repository stores these records elsewhere. Confirm the default set matches the repository before trusting the product count, because a planning folder scanned as product inflates remediation scope.

Plan-driven repositories usually keep sibling folders for solutions, brainstorms, ideation, analysis, and project management. Inspect them during discovery and pass each one that holds engineering or decision records to `--history-path`, because they carry the same exposure as plans.

A debranding run also produces its own artifacts, and its file names embed `RUN_NAME`. Choose a neutral `RUN_NAME` whenever `RELEASE_SCOPE=public` so the plan path does not carry the source customer.

## Separate agent tooling

Repositories that carry agent tooling ship it on release, so it needs branding review, but it is not the product under review. Skill packages, agent definitions, prompt libraries, and vendored context directories describe how the repository is worked on, not what it does, so they cannot hold rename-surviving coupling in the product.

Treat them as a third surface:

* Scan them, because a published tree exposes every branded string they contain
* Tag them `surface=tooling` and remediate them in the branding pass like product text
* Keep them out of the coupling-review denominator, because coupling lenses read product code

The scanner tags `.agents`, `.atv`, `.claude`, `.context`, `.github/agents`, `.github/chatmodes`, `.github/instructions`, `.github/prompts`, and `.github/skills` by default. Use `--tooling-path` when the repository stores tooling elsewhere. `.github/workflows` stays product on purpose: CI is real infrastructure and carries registry hosts, org names, and identity coupling that survive a rename.

Confirm the tooling set before dispatching coupling batches. Tooling scanned as product consumes review batches on files no lens can act on, and it inflates both the denominator and the finding count.

## Classify remediation

Use these remediation types:

1. Direct replacement for display text or simple examples with no hidden customer assumption
2. Parameterization for values that vary by brand, deployment, tenant, or environment
3. Generalization for customer-specific business language that has a neutral equivalent
4. Neutral redesign for workflows, authorization, architecture, schemas, or integrations whose assumptions must change
5. Removal for sensitive, obsolete, confidential, customer-owned, or unsafe-to-generalize content
6. Human review for legal, licensing, contractual, intellectual-property, security, ownership, or production decisions
7. Modernization candidate for customer-specific infrastructure or unsupported operational mechanisms

Never rename a coupled implementation and classify it as a neutral redesign.

Renaming labels in test fixtures and E2E scripts is cosmetic when the data structure itself identifies the brand. A hierarchical org chart with renamed divisions still reveals the original entity's shape to anyone familiar with the source. When the release scope is external or public, recommend removal or replacement with synthetic data rather than find-and-replace anonymization of structurally identifying fixtures.
