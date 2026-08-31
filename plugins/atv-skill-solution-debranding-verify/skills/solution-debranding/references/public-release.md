---
title: Public-Release Readiness
description: Safety and operational checks before sharing a formerly customer-specific repository
---

## Audit sensitive history first

Before normal debranding, determine whether the repository ever contained credentials, customer data, production-derived fixtures, internal endpoints, private certificates, confidential architecture, or personal information.

Run a dedicated scanner against complete Git history. `check-readiness.py` does this for you: it invokes `gitleaks` (or `trufflehog`) when one is on `PATH`, redacts as it goes, and reports the verdict under `history_scanning`. Do not hand-roll the scanner invocation, and do not treat the scan as optional. An installed scanner that was never invoked leaves history uncertified for no reason, and the plan then recommends a fresh repository on an absence of evidence it could have had.

Read the verdict from `history_scanning.status`:

| Status | Meaning | What the plan should say |
| --- | --- | --- |
| `clean` | Scanner ran and found nothing | History may be certified; a fresh repository is optional |
| `findings` | Scanner ran and found matches | History is contaminated; treat a fresh repository as the default |
| `unavailable-no-scanner-installed` | Nothing on `PATH` | History is uncertified; recommend installing a scanner and re-running |
| `skipped-by-request` | `--skip-history-scan` was passed | History is uncertified; say the operator chose to skip |
| `error`, `timed-out` | Scanner failed | History is uncertified; report the failure rather than assuming either outcome |

`--skip-history-scan` exists for repositories where the scan is prohibitively slow. Passing it is a decision to publish on an uncertified history, so name it in the artifact.

Do not print discovered values. Record only category, redacted evidence, path or commit location, and required action.

Record the scanner name, its version, and its verdict in the plan's `## Release readiness` section. Never report a clean history when no scanner ran.

If sensitive history cannot be certified clean, recommend creating a fresh repository from the reviewed working tree and retaining the old repository privately. Rewriting history does not reach forks, clones, caches, downloaded artifacts, or copied secrets.

## Review surfaces outside the tree

Explicitly review or mark out of scope:

* Branches, tags, GitHub Actions logs, caches, and artifacts
* Releases and attached binaries
* Issues, pull requests, discussions, and wiki pages
* Package and container registries
* CI/CD variables, environments, service connections, and secrets
* External documentation, dashboards, monitoring, and telemetry
* Deployed cloud resources, DNS, certificates, and identity objects

Working-tree debranding never certifies these surfaces.

## Decide the planning records

Planning and history artifacts block public release whenever they name the source customer. They leak the customer relationship, internal architecture decisions, incident history, and review discussion well beyond the brand string itself.

Ask which tree the team keeps working in before deciding anything, because that answer sets the default:

| Tree that continues | Default for planning records | Why |
| --- | --- | --- |
| The internal repository, and the public tree is a snapshot | Remove from the published tree | The records stay where the work happens, so nothing compounds away |
| The public repository, because development moves into the open | Anonymize and publish | Removal ships a codebase that has lost its decision history |
| Both, kept in sync | Anonymize and publish, then hold the line | Every future sync reintroduces branded records unless the pipeline anonymizes them |
| The working repository itself is anonymized in place, and the public repository is created from it later | Anonymize in place | One tree carries the records forward, so the decision is made once instead of at every sync |

Anonymizing in place is the strongest option for continuity and the easiest to get wrong at the final step. The tree looks clean while the commit history still holds every pre-anonymization revision, so the public repository must be created as a fresh repository from the reviewed working tree. Cloning, forking, or pushing the anonymized branch carries the original names into the new remote.

## Anonymizing in place breaks the repository the team still runs

Planning records are prose, so anonymizing them in place costs nothing operationally. Deployment identity is not prose. Reusable workflow references, `CODEOWNERS` entries, chart values, registry paths, and service connections are live wiring, and replacing them with placeholders stops the repository from building, deploying, or merging.

The merge itself can become impossible. If branch protection requires review from code owners, the pull request that neutralizes `CODEOWNERS` is blocked by the `CODEOWNERS` it is trying to replace, because the placeholder teams do not exist and nobody can satisfy the rule. If the neutralized workflows are required checks, no pull request merges at all, including that one. Landing the change then needs an administrator override to get past a gate the change itself created.

So do not merge anonymized deployment identity into the branch the team develops on. Keep this work on a long-lived branch that is never merged back, and create the public repository from that branch's tree at cutover. The internal branch keeps working CI, real values, and real owners while internal development continues.

This does not weaken the continuity table. The records still carry forward, because the fresh public repository is built from the reviewed tree either way. It only says which branch holds the reviewed tree in the meantime.

Do not treat removal as the safe default. In a plan-driven repository the planning records are the asset that makes the next change cheap, and a public tree stripped of them asks its next contributor, human or agent, to start from nothing.

Decide one of the following for each record:

* Keep when the record is already neutral and carries no confidential detail
* Anonymize when the substance is publishable but the identities are not
* Relocate to an internal repository that is not part of the published tree
* Remove from the published tree and preserve the original internally

Anonymizing is not the same as falsifying. Replacing a customer name with a neutral role preserves the decision the record documents; rewriting the decision, the outcome, or the reasoning does not. Anonymize identities, never conclusions, and never resolve a record by find and replace alone, because these files carry confidential substance that no term list detects.

Use a neutral role in prose, not a substitute brand. `TARGET_CUSTOMER` exists for places where something has to render as a name: brand profile values, sample data, screenshots, and demo content. A historical record is not one of those places. A sprint summary that reads `Contoso approved the migration` asserts a relationship with a company that was never involved, which is a worse record than the one it replaced. Write the role the sentence actually needs, such as the customer, the project team, or the product team.

A keep or anonymize decision reopens the review. These records were tagged `surface=history` on the assumption they would leave, so no lens has read them for coupling. Move the retained paths out of the history prefixes, re-run the planner, and review them as product documentation.

These decisions govern which records reach the published tree. Never delete a planning record from the working repository, because plan-driven workflows treat them as living artifacts and protect them from removal. Anonymizing one in place is a different act and is allowed: it preserves the record while removing the identity, and it is the whole point of the anonymize-in-place pattern.

## Anonymize the repository slug, not only the brand

The organization and repository names are identifiers in their own right, and they appear inside the tree as well as on it. Issue and pull request links, clone URLs, badge targets, package names, container image paths, and platform descriptors all carry the slug, and the count is usually dominated by planning records that cite issues by full path.

Add the organization name, the repository name, and the combined slug to the term list as resource prefixes so the scanner reports them alongside the brand. Renaming the repository does not reach any of these, and a tree that no longer says the customer's name while still linking to `customer-org/product` is not anonymized.

`scan-debranding.py` derives these from the Git remote and the working directory name by default and lists them under `scan.derived_terms`. Confirm that list rather than trusting it. Derivation reports what the remote says, and the remote is frequently a personal fork or a migrated host whose owner is not the organization the tree actually cites. It also emits shortened forms of the repository name, because a team that works in `customer-platform-widget-svc` writes `widget-svc` in its own prose, and that abbreviation reaches no term list anybody thinks to supply. Pass `--no-derive-terms` only when the repository's own identity is genuinely out of scope.

Links into a private repository do not survive anonymization as links. Reduce them to a bare issue reference or mark them as an internal record rather than repointing them somewhere that does not exist.

## Keep this workflow's own artifacts internal

The records this workflow writes to `ARTIFACT_ROOT` (default: `docs/debranding`) are the one planning record the continuity table never applies to. Whatever the team decides for every other record, these do not publish.

They are a different kind of document. Other planning records explain why the product is the way it is, which is why publishing them compounds. A debranding artifact explains how the customer was removed, which compounds for nobody downstream and discloses three things that should not leave the internal tree:

* A reversal key, because the artifact enumerates every location the customer appeared and what replaced it. Anonymizing is not available here: identity is the subject of the document, not an incidental detail in it
* A security map of a repository that still exists, because `## Release readiness` records the history-scan verdict, the credential material found, and the datasets holding identity columns
* Internal review discussion, including referrals that name owners and tracks

`check-readiness.py` lists these paths under `workflow_artifacts`. The list is expected to be non-empty in the working repository, where the artifacts belong. It exists so the published tree can be checked against a specific set of paths rather than a memory of what the run produced, because these files are created during the run and are the ones a discovery-time review never saw.

Publishing a debranding methodology is a separate decision. Write a new document for that purpose and review it as product documentation. Do not republish the run artifact.

## Check what depends on the planning records

Agent tooling ships on public release, and in a plan-driven repository that tooling reads the planning records. Removing history while publishing the tooling that indexes it produces a workflow pointing at directories that no longer exist.

`check-readiness.py` reports this coupling as `files_referencing_history`. Read it before acting on any removal decision:

* A high count means removal is a breaking change to the shipped workflow, not a deletion
* Each referring file needs its own decision: update the reference, relocate alongside the records, or drop from the published tree
* A zero count in a repository that clearly uses plan-driven engineering means the history prefixes are wrong, so confirm them with `--history-path`

Name the count in the plan's `## Release readiness` section alongside the record decision it constrains.

Confirm before issuing a release-ready verdict that no planning record and no planning file name carries the source customer.

## Run the readiness indicators

```bash
python scripts/check-readiness.py --root <repository> --output release-readiness.json
```

Missing indicators are discussion prompts unless they create a concrete blocker. Review README, license, contribution guidance, security policy, code of conduct, ignore rules, editor settings, CI, tests, dependency updates, release tags, and suspicious tracked files.

## Decide what the committed data is

The same report lists `committed_datasets`: every tracked spreadsheet, delimited file, and record dump, with its size, its column count, and the subset of column names that look like identity fields. Column names are reported because they are enough to classify the file; values are never read into the report, and must not be quoted into the plan either.

A dataset marked `readable: false` is a spreadsheet or columnar binary that nothing in this workflow can open. Absence of evidence is not evidence of absence, so each one needs a named human owner before release.

Decide one of the following for every dataset, and record the decision in the plan's `## Release readiness` section:

* Keep when the data is already synthetic and carries no customer detail
* Synthesize a replacement that preserves shape and volume so demos and tests still work
* Remove from the published tree and document how a consumer supplies their own

Treat personal data and customer-confidential records as a release blocker in their own right. A ledger export, an HR extract, or a customer list is damaging to publish whether or not it mentions the brand, so it does not become acceptable because the debranding passed.

## Verify licensing and ownership

Confirm:

* The organization has authority to release the source and non-code assets
* Customer-owned code, data, fixtures, screenshots, diagrams, certificates, and documentation are removed or approved
* The repository license, package metadata, notices, and README agree
* Third-party licenses and required attribution remain intact
* Copyright and ownership changes have human approval

Treat uncertainty as `Needs your decision`; treat confirmed lack of release rights as `Blocks release`.

## Harden release operations

Review CI permissions, action pinning, branch protection, required checks, dependency updates, release automation, package provenance, and publishing identities. Replace organization-owned reusable workflows only after equivalent behavior and credentials are available.

## Validate as an outsider

Clone the candidate repository into a clean directory and follow the README verbatim. Confirm that an outsider can configure, build, test, run, and understand the project without internal knowledge or customer infrastructure.

Use the repository's declared package manager and immutable or frozen lockfile mode before building. Treat undeclared dependencies, lockfile drift, registry changes, generated local launchers, and required untracked stubs as reproducibility failures. Run a startup and health smoke check on the current host, record the operating system and runtime versions, and state which supported platforms remain untested.

Re-run discovery and readiness checks from the clean clone. Compare failures with the recorded pre-change baseline so debranding regressions are distinguished from existing portability defects. State every deliberate gap and scope exclusion before issuing a release-ready verdict.

