#!/usr/bin/env python3
"""Merge coupling-review lens output into one deduplicated finding set."""

from __future__ import annotations

import argparse
import json
import re
import sys
from collections import Counter
from pathlib import Path
from typing import Any

ANCHORS = (0, 25, 50, 75, 100)
ANCHOR_TO_CONFIDENCE = {100: "high", 75: "high", 50: "medium", 25: "low"}
LENSES = (
    "business-logic",
    "data-model",
    "tenancy-authorization",
    "integration",
    "visual-identity",
    "embedded-data",
    "narrative-documentation",
    "adversarial-attribution",
)
MAX_FINDINGS_PER_BATCH = 25
CONTRACT_BREACH = 2


def extract_json(text: str) -> str:
    """Subagents often wrap their answer in a fence or a sentence, so find the object."""
    fence = re.search(r"```(?:json)?\s*(.+?)```", text, re.DOTALL)
    if fence:
        return fence.group(1).strip()
    start = text.find("{")
    end = text.rfind("}")
    if start != -1 and end > start:
        return text[start : end + 1]
    return text


def load_lens(path: Path) -> dict[str, Any]:
    text = path.read_text(encoding="utf-8")
    try:
        payload = json.loads(text)
    except json.JSONDecodeError:
        try:
            payload = json.loads(extract_json(text))
        except json.JSONDecodeError as error:
            raise SystemExit(f"{path}: invalid JSON: {error}") from error
    if not isinstance(payload, dict):
        raise SystemExit(f"{path}: expected a JSON object")
    return payload


def normalize_evidence(value: Any) -> list[str]:
    if isinstance(value, list):
        return [str(item) for item in value]
    if value in (None, ""):
        return []
    return [str(value)]


def merge_key(finding: dict[str, Any]) -> tuple[str, str]:
    return (str(finding.get("path", "")), str(finding.get("anchor", "")))


def collect(paths: list[Path]) -> tuple[list[dict[str, Any]], list[str], dict[str, Any]]:
    merged: dict[tuple[str, str], dict[str, Any]] = {}
    residual: list[dict[str, Any]] = []
    problems: list[str] = []
    notes: list[str] = []
    coverage: dict[str, Any] = {}

    for path in paths:
        payload = load_lens(path)
        lens = str(payload.get("lens", path.stem))
        if lens not in LENSES:
            problems.append(f"{path.name}: unknown lens {lens!r}")

        batch_id = str(payload.get("batch_id") or lens)
        if batch_id in coverage:
            problems.append(f"{path.name}: batch {batch_id} was already reported by another file")

        findings = payload.get("findings") or []
        if len(findings) > MAX_FINDINGS_PER_BATCH:
            notes.append(f"{batch_id}: {len(findings)} findings exceeds the cap of {MAX_FINDINGS_PER_BATCH}")

        entry = dict(payload.get("coverage") or {})
        entry["lens"] = lens
        coverage[batch_id] = entry
        if not entry.get("files_reviewed"):
            problems.append(f"{batch_id}: reported no files_reviewed, so its coverage cannot be verified")

        for finding in findings:
            confidence = finding.get("confidence")
            if confidence not in ANCHORS:
                problems.append(f"{lens}: confidence {confidence!r} is not an anchor value")
                continue
            if confidence == 0:
                continue

            record = dict(finding)
            record["lens"] = lens
            record["evidence"] = normalize_evidence(finding.get("evidence"))
            record["confidence_anchor"] = confidence
            record["confidence"] = ANCHOR_TO_CONFIDENCE[confidence]

            if confidence == 25:
                residual.append(record)
                continue

            key = merge_key(record)
            existing = merged.get(key)
            if existing is None:
                record["coupling_types"] = sorted({str(record.get("coupling_type", lens))})
                merged[key] = record
                continue

            existing["coupling_types"] = sorted(
                set(existing["coupling_types"]) | {str(record.get("coupling_type", lens))}
            )
            existing["evidence"] = sorted(set(existing["evidence"]) | set(record["evidence"]))
            if record["confidence_anchor"] > existing["confidence_anchor"]:
                existing["confidence_anchor"] = record["confidence_anchor"]
                existing["confidence"] = record["confidence"]

    ordered = sorted(
        merged.values(),
        key=lambda item: (-item["confidence_anchor"], item.get("path", ""), item.get("anchor", "")),
    )
    return ordered, problems, {"per_lens": coverage, "residual_risks": residual, "notes": notes}


def enforce(coverage: dict[str, Any], plan_path: Path | None) -> dict[str, Any]:
    """Hold each batch to the file list it was assigned."""
    if plan_path is None:
        return {"enforced": False}

    plan = json.loads(plan_path.read_text(encoding="utf-8"))
    assigned = {str(batch["id"]): set(batch["files"]) for batch in plan.get("batches", [])}

    breaches: list[dict[str, Any]] = []
    for batch_id, files in sorted(assigned.items()):
        entry = coverage.get(batch_id)
        if entry is None:
            breaches.append({"batch": batch_id, "fault": "never reported", "unread": sorted(files)})
            continue
        unread = sorted(files - {str(item) for item in (entry.get("files_reviewed") or [])})
        if unread:
            breaches.append({"batch": batch_id, "fault": "incomplete", "unread": unread})

    return {
        "enforced": True,
        "batches_assigned": len(assigned),
        "batches_reported": len(coverage),
        "batches_complete": len(assigned) - len(breaches),
        "unexpected_batches": sorted(set(coverage) - set(assigned)),
        "breaches": breaches,
    }


def reconcile(coverage: dict[str, Any], report_path: Path | None, plan_path: Path | None) -> dict[str, Any]:
    reviewed: set[str] = set()
    per_lens: dict[str, set[str]] = {}
    for entry in coverage.values():
        files = {str(item) for item in (entry.get("files_reviewed") or [])}
        per_lens.setdefault(str(entry.get("lens", "unknown")), set()).update(files)
        reviewed |= files
    reviewed_by_lens = {lens: len(files) for lens, files in sorted(per_lens.items())}

    if report_path is None:
        return {"denominator": None, "files_reviewed": len(reviewed), "reviewed_by_lens": reviewed_by_lens}

    report = json.loads(report_path.read_text(encoding="utf-8"))
    manifest = report.get("coverage", {}).get("product_files")
    if not manifest:
        raise SystemExit(f"{report_path}: scan report has no coverage.product_files manifest")

    excluded: set[str] = set()
    if plan_path is not None:
        plan = json.loads(plan_path.read_text(encoding="utf-8"))
        excluded = {str(item["path"]) for item in plan.get("excluded", [])}

    product = set(manifest) - excluded
    covered = reviewed & product
    unreviewed = sorted(product - reviewed)
    return {
        "denominator": len(product),
        "excluded_by_rule": len(excluded),
        "files_reviewed": len(covered),
        "coverage_percent": round(100 * len(covered) / len(product), 1) if product else 100.0,
        "reviewed_by_lens": reviewed_by_lens,
        "reviewed_outside_manifest": sorted(reviewed - product - excluded),
        "unreviewed_count": len(unreviewed),
        "unreviewed": unreviewed,
    }


TOP_FINDINGS = 25


def render_summary(report: dict[str, Any]) -> None:
    contract = report["batch_contract"]
    coverage = report["file_coverage"]
    summary = report["summary"]

    if contract.get("enforced"):
        print(
            f"batches         {contract['batches_complete']}/{contract['batches_assigned']} complete, "
            f"{len(contract['breaches'])} breach(es)"
        )
    else:
        print("batches         not enforced (no --batch-plan)")
    if coverage.get("denominator") is not None:
        print(
            f"coverage        {coverage['files_reviewed']}/{coverage['denominator']} product files "
            f"({coverage['coverage_percent']}%), {coverage['excluded_by_rule']} excluded by rule"
        )
    print(
        f"findings        {summary['findings']} "
        f"({summary['by_confidence']['high']} high, {summary['by_confidence']['medium']} medium), "
        f"{summary['residual_risks']} residual"
    )
    print(
        "files by lens   "
        + ", ".join(f"{k}={v}" for k, v in coverage.get("reviewed_by_lens", {}).items())
        + " (adversarial-attribution re-reviews files another lens already owns)"
    )

    for breach in contract.get("breaches", []):
        print(f"  BREACH  {breach['batch']} {breach['fault']}, {len(breach['unread'])} file(s) unopened")

    if report["findings"]:
        print(f"\ntop findings by confidence (showing {min(TOP_FINDINGS, len(report['findings']))})")
        for finding in report["findings"][:TOP_FINDINGS]:
            types = ",".join(finding.get("coupling_types", []))
            print(f"  {finding['confidence_anchor']:>3}  {types:<24}  {finding.get('path')}:{finding.get('anchor')}")

    if report["notes"]:
        print("\nnotes")
        for note in report["notes"]:
            print(f"  {note}")

    if report["problems"]:
        print("\nproblems")
        for problem in report["problems"]:
            print(f"  {problem}")


def render_blocks(report: dict[str, Any]) -> None:
    blocking = [item for item in report["findings"] if item.get("disposition") == "[BLOCKS]"]
    if not blocking:
        print("no blocking couplings")
        return

    high = [item for item in blocking if item["confidence_anchor"] >= 75]
    print(f"blocking        {len(blocking)} ({len(high)} high confidence)")

    areas = Counter(str(item.get("path", "")).split("/")[0] for item in blocking)
    print("by area         " + ", ".join(f"{area}={count}" for area, count in areas.most_common()))

    remediations = Counter(str(item.get("remediation", "unspecified")) for item in blocking)
    print("by remediation  " + ", ".join(f"{name}={count}" for name, count in remediations.most_common()))

    owners = Counter(str(item.get("owner", "unassigned")) for item in blocking)
    print("by owner        " + ", ".join(f"{name}={count}" for name, count in owners.most_common()))

    print(f"\nhigh-confidence blocking couplings ({len(high)})")
    for item in high:
        location = f"{item.get('path')}:{item.get('line')}"
        print(f"  {location}  {item.get('anchor')}  [{item.get('remediation', 'unspecified')}]")
        matters = str(item.get("why_it_matters") or "").replace("\n", " ").strip()
        if matters:
            print(f"      {matters}")


DISPOSITION_MARKS = (("[BLOCKS]", "B"), ("[GATED]", "G"), ("[TUNE]", "T"))
DEFAULT_TREE_DEPTH = 3
TOP_UNEXAMINED = 5


def roll_up(paths: list[tuple[str, str]], depth: int) -> dict[str, Counter]:
    totals: dict[str, Counter] = {}
    for path, key in paths:
        parts = [part for part in str(path).split("/") if part][:-1] or ["(root)"]
        for index in range(min(len(parts), depth)):
            totals.setdefault("/".join(parts[: index + 1]), Counter())[key] += 1
    return totals


def render_tree(report: dict[str, Any], depth: int) -> None:
    findings = roll_up(
        [(item.get("path", ""), str(item.get("disposition", "[TUNE]"))) for item in report["findings"]],
        depth,
    )
    unread = roll_up([(path, "unread") for path in report["file_coverage"].get("unreviewed", [])], depth)

    if not findings:
        print("no findings to map")
    else:
        print(f"exposure map (depth {depth}; B=blocks G=gated T=tune)")
        for node in sorted(findings):
            counter = findings[node]
            cells = [f"{mark}{counter[key]}" for key, mark in DISPOSITION_MARKS if counter[key]]
            if node in unread:
                cells.append(f"unread={unread[node]['unread']}")
            label = "  " * node.count("/") + node.rsplit("/", 1)[-1] + "/"
            print(f"  {label:<44}{' '.join(cells)}")

    # A directory with no findings and many unread files is unexamined, not clean.
    blind = sorted(
        ((node, counter["unread"]) for node, counter in unread.items() if node not in findings),
        key=lambda item: (-item[1], item[0]),
    )[:TOP_UNEXAMINED]
    if blind:
        print("\nlargest unexamined directories")
        for node, count in blind:
            print(f"  {node + '/':<44}unread={count}")


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("lens_files", nargs="+", type=Path, help="JSON file per lens")
    parser.add_argument("--report", type=Path, help="Scan report supplying the product_files denominator")
    parser.add_argument("--batch-plan", type=Path, help="Batch plan the lenses must satisfy in full")
    parser.add_argument("--output", type=Path, help="Write the merged report here")
    parser.add_argument(
        "--summary",
        action="store_true",
        help="Print a human-readable rollup instead of the full JSON",
    )
    parser.add_argument(
        "--blocks",
        action="store_true",
        help="Print the blocking couplings grouped by area, remediation, and owner",
    )
    parser.add_argument(
        "--tree",
        nargs="?",
        type=int,
        const=DEFAULT_TREE_DEPTH,
        metavar="DEPTH",
        help="Print findings and unread files as an annotated directory tree",
    )
    arguments = parser.parse_args(argv)

    findings, problems, extra = collect(arguments.lens_files)
    contract = enforce(extra["per_lens"], arguments.batch_plan)
    reconciliation = reconcile(extra["per_lens"], arguments.report, arguments.batch_plan)
    if reconciliation.get("unreviewed_count"):
        problems.append(
            f"{reconciliation['unreviewed_count']} product files were never opened by any lens"
        )
    for breach in contract.get("breaches", []):
        problems.append(
            f"batch {breach['batch']} {breach['fault']}: {len(breach['unread'])} assigned files were not opened"
        )

    report = {
        "lenses_expected": list(LENSES),
        "lenses_received": sorted(reconciliation.get("reviewed_by_lens", {})),
        "lenses_missing": sorted(set(LENSES) - set(reconciliation.get("reviewed_by_lens", {}))),
        "batch_contract": contract,
        "file_coverage": reconciliation,
        "summary": {
            "findings": len(findings),
            "residual_risks": len(extra["residual_risks"]),
            "by_confidence": {
                level: sum(1 for item in findings if item["confidence"] == level)
                for level in ("high", "medium")
            },
        },
        "findings": findings,
        "residual_risks": extra["residual_risks"],
        "coverage": extra["per_lens"],
        "notes": extra["notes"],
        "problems": problems,
    }

    text = json.dumps(report, indent=2, sort_keys=False)
    if arguments.output:
        arguments.output.write_text(text, encoding="utf-8")
    elif not (arguments.summary or arguments.blocks or arguments.tree):
        sys.stdout.write(text + "\n")

    if arguments.summary:
        render_summary(report)
    if arguments.tree:
        if arguments.summary:
            print()
        render_tree(report, arguments.tree)
    if arguments.blocks:
        if arguments.summary or arguments.tree:
            print()
        render_blocks(report)

    if contract.get("breaches"):
        unread = sum(len(breach["unread"]) for breach in contract["breaches"])
        print(
            f"batch contract breached: {len(contract['breaches'])} of {contract['batches_assigned']} "
            f"batches left {unread} assigned files unopened. Re-dispatch those batches.",
            file=sys.stderr,
        )
        return CONTRACT_BREACH
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
