---
title: Debranding Coupling Review
description: Parallel lens review that finds customer coupling surviving removal of every literal identifier
---

## Purpose

The scanner finds literal identifiers. This review finds the coupling that remains after every literal identifier is gone.

Run it during discovery whenever the outcome is reuse, resale, handover, or public release. Skip it in delta mode unless the changed files touch business rules, schemas, authorization, or integrations.

## Divide the work

The scanner owns literal detection. Lenses own semantic detection.

A lens must never report a finding whose only evidence is a literal source-customer string. Those findings already exist with stable identifiers and redacted evidence, and repeating them floods the artifact set.

A lens reports code that still encodes one customer's business, structure, or environment after renaming. Neutral naming is not evidence of portability.

## Run the lenses

Do not let a lens choose what to read. Plan the batches first, then dispatch them:

```bash
python scripts/plan-lens-batches.py --report scan-report.json --output lens-batches.json --split-dir batches/
```

The planner routes every product file to exactly one lens by path and file type, drops files that cannot carry coupling with a recorded reason such as `lockfile`, `vendored`, `minified`, or `generated-output`, and splits the rest into fixed batches. The `adversarial-attribution` lens additionally receives an evenly spaced sample across the whole reviewable set, because its job is noticing what the routed lenses miss.

Raster images and font binaries are excluded because a lens cannot read them, not because they are clean. Inventory them by path for human review, and treat a logo, favicon, or licensed typeface as a finding the review must still account for. Screenshots deserve their own pass: a UI capture carries the palette, the logo, real record values, and often the tenant name in the browser chrome, all at once.

Spreadsheets and columnar data files are excluded with the distinct reason `tabular-binary`, which means unread data rather than absent data. Every such path goes to a human, and `scripts/check-readiness.py` lists the readable ones with their column names so the plan can say what kind of records the repository is carrying.

The `generated-output` exclusion covers build-tool directories only, because the manifest holds tracked files and anything tracked ships. A committed `output/`, `results/`, or `evaluations/` tree is a published record of real runs against real inputs, so it is routed and reviewed like any other file rather than dismissed as reproducible.

Its input is `coverage.product_files`, so files the scanner tagged `surface=history` or `surface=tooling` never reach a lens. Agent tooling still needs branding, but it is not product code, and reviewing it consumes batches that buy nothing. Check the scanner's `files_by_surface` before planning: tooling counted as product silently spends batches and inflates the denominator.

The default history and tooling prefixes encode one repository's filing habits, not a universal layout. Confirm them against the actual tree before planning and override with `--history-path` and `--tooling-path` where they do not fit, because a prefix that matches the wrong directory removes those files from the denominator without appearing anywhere as an exclusion. A design or architecture document filed under a planning prefix is the usual casualty.

Classifying a file as `surface=history` presumes it leaves the published tree. When the user decides to keep one instead, that presumption is spent and nothing has coupling-reviewed it. Reclassify those paths and re-run the planner rather than treating the keep decision as the end of the matter, because a kept planning record is product documentation now and carries the same attribution as any other file.

`--split-dir` writes each batch to `<dir>/<batch-id>.json`. Use it, and hand each subagent the path to its own file. Do not write a script to explode the plan yourself.

Dispatch one parallel read-only subagent per batch. Give each the identity map, its own `batch.id`, its own `batch.files` list, and the lens definition below.

A subagent cannot write files. It returns its JSON to you, and you write that response to `<scratch>/lens-results/<batch-id>.json` exactly as returned. Transcribing a returned payload is not improvised parsing, so it does not conflict with the rule against hand-rolled tooling. Do not summarize, re-key, or merge responses while transcribing them: the merge script is the only thing that reads them.

**Never re-dispatch a batch over its formatting.** The merge script pulls the object out of a code fence or a surrounding sentence, so a fenced or prose-wrapped response is already mergeable. Save it verbatim and move on. The only reason to re-dispatch is a batch that left assigned files unopened, and the merge script names those for you. Chasing bare-JSON compliance burns the review budget on work that was already complete.

Write each response to disk in the same step that receives it. A response held in conversation to be transcribed later can disappear before you get there, and then the review is genuinely lost rather than merely untidy.

Copying the response's backing resource file into `lens-results/` is an acceptable way to transcribe it, and is preferable to retyping a large payload, but those resource files are ephemeral. Whichever way you transcribe, confirm the landing immediately in the same step:

```pwsh
foreach ($f in Get-ChildItem $dst -Filter *.json) {
  $t = Get-Content $f.FullName -Raw
  if ($t -match '"batch_id"\s*:\s*"([^"]+)"') { "$($f.Name) -> $($Matches[1])" } else { "$($f.Name) -> NO batch_id" }
}
```

Every file must report a `batch_id` matching its own name. A missing file, an empty file, or a mismatched `batch_id` means that response is gone, and gone responses are the one case that genuinely requires re-dispatch. Finding out at merge time instead means re-running lens work whose results you already paid for.

One dispatch covers one batch and produces one file. A response carrying two batches, or an array of batch objects, is a dispatch error: re-dispatch those batches separately rather than splitting the array yourself.

Give each subagent the absolute path to its batch file. An abbreviated or repository-relative path resolves differently in a subagent's working directory and the dispatch fails.

The file list is a contract, not a suggestion. A subagent opens every file it was handed, returns that exact list in `files_reviewed`, and does not roam outside it. Batches are already sized to fit a subagent context, so there is no budget justification for sampling within one. A subagent that judges a file uninteresting still opened it, and still lists it. The merge step diffs `files_reviewed` against the assignment and fails the run on any shortfall, so a short list stops the review rather than quietly shrinking it.

Do not consolidate batches. Thirty-five assigned batches means thirty-five subagents. Merging them into fewer, larger dispatches is the failure this contract exists to prevent.

| Lens | Looks for |
|------|-----------|
| `business-logic` | Rules, thresholds, categories, state machines, and vocabulary drawn from one customer's operating model |
| `data-model` | Schemas, enums, taxonomies, and field semantics that encode one customer's structure |
| `tenancy-authorization` | Org hierarchy, role names, permission models, and single-tenant assumptions |
| `integration` | Customer-specific services, endpoints, identity flows, network placement, and gateway assumptions |
| `visual-identity` | Palette hexes, typography, logo geometry, iconography, and layout that reproduce a corporate design system |
| `embedded-data` | Committed records: sample data, fixtures, seeds, and exports that describe real people, entities, or transactions |
| `narrative-documentation` | Prose that names people, meetings, decisions, projects, and internal systems |
| `adversarial-attribution` | Whatever still identifies the customer once every brand string is removed |

Give the visual lens this framing: a brand colour is an identifier that no rename touches. A configured primary colour such as `#2563EB` stays recognisable after every occurrence of the customer's name is gone, and so do a licensed corporate typeface, a logo's silhouette in an inline SVG, and a bespoke iconography set. Ask whether each value is a design decision the solution needs or a corporate palette it inherited, and record the token that should replace it. Treat bundled font and logo files as a redistribution question as well as an attribution one, because corporate typefaces are usually licensed to one organisation and cannot ship in a public repository at all.

Give the data lens this framing: a debranding review asks what identifies the customer, but committed records answer a worse question, which is who the customer's own customers and staff are. A customer record export can name legal entities, contacts, account holders, and internal owners. Fixtures drawn from production keep their real values. Report column names, row counts, and paths; do not quote values into the plan, because that copies the exposure into a second file. Judge each dataset as keep, synthesize, or remove, and treat anything holding personal or customer-confidential records as a release blocker rather than a branding cleanup.

Give the integration lens this framing for credential material: a corporate root certificate or proxy trust bundle in the build couples the solution to one company's network, and no rename undoes it. Removing it is not cosmetic, because the build stops working outside that network, so the finding is a portability defect with an attribution side effect. The filename, the certificate subject, and the `Dockerfile` line that installs it all name the customer independently.

Give the integration lens this additional framing for infrastructure defaults: a configuration value that appears in an environment template and again as a code-level default or fallback is two exposures, not one. Scrubbing the template while leaving the code default intact carries the source-brand value into every deployment that does not override it. For each source-specific endpoint, namespace, FQDN, or managed-identity name found in the code, check whether the same value also appears as a hardcoded default in a settings class, a dataclass field, or an `os.getenv` fallback. Report each code default that duplicates a templated value, even when the template has already been scrubbed or is planned for scrubbing.

Give the documentation lens this framing: prose is where debranding usually fails, because a rename fixes the product name and leaves the story around it intact. Architecture decision records name the people who made the call, design notes recount meetings and customer conversations, runbooks list internal hostnames and ticket queues, and screenshots embedded in a page carry the palette and real record values at once. Read for the narrative, not the vocabulary: a document that never says the customer's name can still describe their org chart, their approval process, and their vendor relationships.

Give the business-logic lens this extra framing for model prompts and agent instructions: a prompt is a specification written in prose, so it concentrates the customer's methodology, role names, review workflow, tone, and regulatory framing more densely than the code that calls it. Treat prompt files, agent definitions, and few-shot examples as product logic under review, and check the examples in particular, because sample rows in a prompt are often lifted from real data.

Give the adversarial lens this framing: every literal identifier has already been replaced, and the reader is an outsider trying to name the customer from sample data, locale, regulatory regime, domain vocabulary, comments, fixtures, and screenshots.

Program and project codenames belong to that lens. A directory named for a codename survives every rename of the customer, reads as neutral to the person who chose it, and identifies the customer to anyone who has seen a proposal, an invoice, or a press release. Treat a codename as an identifier, and check whether removing it also breaks a path, a test baseline, or a result directory.

Platform and service-catalog descriptors deserve a direct read rather than a scan. A `catalog-info.yaml` or equivalent typically carries a named business owner's email address, an internal billing or cost code, a product registry identifier, an internal hostname, and directory group names. None of that is code, none of it is reachable by searching for the brand string, and each field names the organisation on its own.

No lens edits files. No lens runs a command that mutates the repository.

## Return findings as JSON

Each lens returns one JSON object. Prefer a bare object with no code fence and no surrounding prose, but the merge script extracts the object from a fence or a sentence, so formatting alone never invalidates a response:

```json
{
  "lens": "tenancy-authorization",
  "batch_id": "tenancy-authorization-001",
  "findings": [
    {
      "id": "CPL-tenancy-authorization-src/auth/roles.py-RoleHierarchy",
      "path": "src/auth/roles.py",
      "line": 42,
      "anchor": "RoleHierarchy",
      "coupling_type": "tenancy-authorization",
      "confidence": 75,
      "evidence": ["Role tree assumes a fixed four-level reporting structure"],
      "survives_rename": "Renaming the roles preserves the assumed hierarchy depth",
      "why_it_matters": "A reusing customer with flatter reporting cannot express its structure",
      "disposition": "Needs your decision",
      "remediation": "neutral-redesign",
      "owner": "platform-architecture"
    }
  ],
  "residual_risks": [],
  "deferred_questions": [],
  "coverage": {
    "files_reviewed": [],
    "not_reviewed": []
  }
}
```

Build `id` as `CPL-<coupling_type>-<path>-<anchor>` so the same coupling keeps the same identifier across runs. Redact identifiers in `evidence` exactly as the scanner does, and always return `evidence` as an array holding at least one quoted item.

Set `batch_id` to the `id` of the batch you were assigned, and set `files_reviewed` to every file in that batch, as repository-relative paths. This is not a summary count and not a highlight reel. The merge step holds the returned list against the assignment, so omitting a file you opened understates coverage, and listing a file you did not open is a false coverage claim. Use `not_reviewed` only for findings dropped by the 25-finding cap, never for files you chose to skip.

Set `confidence` to exactly one of `0`, `25`, `50`, `75`, or `100`. Never emit a value between anchors, and never emit `high`, `medium`, or `low`. Choose the anchor whose behavior you actually performed:

| Value | Criterion |
|-------|-----------|
| `100` | Verified, and the evidence leaves no room for interpretation |
| `75` | Double-checked, and a reusing customer will hit this in practice |
| `50` | Verified as real, but it may not block reuse on its own |
| `25` | Suspected, and could not be verified from the code available |
| `0` | Does not survive light scrutiny |

Put anything you cannot anchor to a path and symbol in `residual_risks`, and anything needing a human decision before it can be judged in `deferred_questions`. Neither is a finding, and neither may be dropped silently.

Cap each batch at 25 findings ranked by impact, and record everything dropped by that cap in `not_reviewed`.

## Synthesize the results

Merge the transcribed lens files with the packaged script rather than an improvised parser:

```bash
python scripts/merge-lens-findings.py <lens-files> --report scan-report.json --batch-plan lens-batches.json --output merged-couplings.json
```

The script discards confidence `0`, routes confidence `25` into `residual_risks`, merges duplicates by path and anchor, keeps the highest anchor, unions the coupling types, and reports every lens that is missing or that broke the schema. Passing `--report` and `--batch-plan` reconciles the union of `files_reviewed` against the product manifest minus the rule-excluded files, and emits `file_coverage` with the denominator, the percentage, and the exact list of files no lens opened.

The script exits `2` when any batch reports fewer files than it was assigned, and writes the shortfall into `batch_contract.breaches`. That is a hard stop. Re-dispatch the named batches with their unread files and merge again. Do not proceed to the artifacts on a non-zero exit, and do not paper over the gap with a coverage caveat. Treat a non-empty `problems` array as a defect in the run, not as noise.

`notes` is separate and is not a defect. A batch that returns one or two findings over the cap lands there, because the overrun is worth knowing and worth nothing else. Keeping it out of `problems` is what lets `problems` stay a stop signal rather than a list you learn to scroll past.

Add `--summary` to print the contract result, the coverage percentage, the per-lens counts, the highest-confidence findings, and the problems list. Use it to read the merged output without loading the JSON into context, and never hand-roll a reader to produce the same view.

Add `--blocks` to print the blocking couplings grouped by area, remediation, and owner, followed by every high-confidence blocker with its location and stake. That is the view the plan's assessment and modernization sections need, so use it instead of writing a probe against `merged-couplings.json`.

Record `file_coverage.coverage_percent` and the unreviewed count in the plan's `## Assessment` section. A coupling review that reports findings without reporting its coverage is not evidence of anything.

The merge step is mandatory once lenses are dispatched. A run that reviews batches but never produces `merged-couplings.json` has no coupling evidence at all, whatever its subagents reported in prose, and it may not write the plan artifacts. Stop and report the blockage instead.

Apply these judgement rules, which the script cannot enforce, before merging into the artifact set:

* Discard any finding without a concrete path and anchor
* Discard any finding whose evidence is only a literal identifier match
* Discard any finding whose `survives_rename` field does not describe a behavioral assumption
* Record every lens that failed or returned nothing, because a silent lens is not a clean result

Anchors map onto the finding schema as `100` and `75` for high confidence, `50` for medium, and `25` for low. A low-confidence coupling finding is not a low-risk finding. Carry the numeric `confidence_anchor` into the plan's `## Assessment` section alongside the mapped level, so the raw lens judgement stays auditable after the merge.

## Reuse existing review assets

When the repository already ships a document-review skill with a subagent template, findings schema, and synthesis reference, dispatch through those rather than duplicating them. Keep this file authoritative for lens selection, the literal-match exclusion, and the `survives_rename` requirement.

## Disposition defaults

Coupling findings default to `Needs your decision` with `neutral-redesign` or `human-review`.

Never assign direct replacement, parameterization, or generalization to a coupling finding without first describing the assumption being changed and its breaking effects. Renaming a coupled implementation and calling it portable is the failure this review exists to prevent.

## State the limits

This review is non-deterministic and samples a large repository rather than reading every line. Record the lens set, the scope it received, and the paths it did not reach.

Absence of coupling findings is not evidence of portability. Report it as reviewed coverage, never as a clean result.

