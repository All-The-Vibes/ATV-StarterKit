#!/usr/bin/env python3
"""Report informational public-release readiness indicators as JSON."""

from __future__ import annotations

import argparse
import json
import re
import shutil
import subprocess
import sys
import tempfile
from collections import Counter
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

SUSPICIOUS_PATH = re.compile(
    r"(^|/)(\.env(?:\..+)?|.*\.(?:cer|crt|der|jks|key|p12|pem|pfx)|"
    r"id_(?:rsa|dsa|ecdsa|ed25519).*|.*credentials.*\.(?:json|xml|ya?ml))$",
    re.IGNORECASE,
)
SAFE_SAMPLE = re.compile(r"\.(?:example|sample|template)(?:\.|$)", re.IGNORECASE)
DATASET_PATH = re.compile(
    r"\.(?:csv|tsv|psv|jsonl|ndjson|parquet|avro|xlsx?|xlsm|xlsb|sav|dta|db|sqlite3?|mdb|accdb)$",
    re.IGNORECASE,
)
TEXT_TABULAR = (".csv", ".tsv", ".psv")
TEXT_RECORDS = (".jsonl", ".ndjson")
IDENTITY_COLUMN = re.compile(
    r"(name|email|phone|mobile|address|postcode|user_?id|userid|employee|emp_?id|staff|"
    r"national_?id|tax_?id|passport|registration_?(?:id|number)|dob|birth|salary|"
    r"customer|client|entity|account_?holder|ip_?address)",
    re.IGNORECASE,
)
MAX_REPORTED_DATASETS = 50
MAX_REPORTED_COLUMNS = 25
HEADER_READ_BYTES = 16384
HISTORY_SCANNERS = ("gitleaks", "trufflehog")
HISTORY_SCAN_TIMEOUT_SECONDS = 900
MAX_REPORTED_LOCATIONS = 50
DEFAULT_HISTORY_PATHS = (
    "docs/debranding",
    "docs/plans",
    "docs/pm",
    "docs/analysis",
    "docs/brainstorms",
    "docs/dashboard",
    "docs/ideation",
    "docs/solutions",
    "docs/verification",
)
REFERENCE_READ_BYTES = 1024 * 1024
MAX_REPORTED_REFERRERS = 50
DEFAULT_WORKFLOW_ARTIFACT_PATH = "docs/debranding"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path.cwd())
    parser.add_argument("--output", type=Path)
    parser.add_argument(
        "--skip-history-scan",
        action="store_true",
        help="Report the installed scanner without running it; history stays uncertified.",
    )
    parser.add_argument(
        "--history-path",
        action="append",
        default=[],
        help=(
            "Repository-relative directory holding planning records. "
            f"Defaults to {', '.join(DEFAULT_HISTORY_PATHS)}."
        ),
    )
    parser.add_argument(
        "--workflow-artifact-path",
        default=DEFAULT_WORKFLOW_ARTIFACT_PATH,
        help="Repository-relative directory containing this workflow's plans.",
    )
    return parser.parse_args()


def git_lines(root: Path, *args: str) -> list[str] | None:
    result = subprocess.run(
        ["git", "-C", str(root), *args],
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
        check=False,
    )
    if result.returncode != 0:
        return None
    return [line for line in result.stdout.splitlines() if line]


def exists_any(root: Path, candidates: tuple[str, ...]) -> bool:
    return any((root / candidate).exists() for candidate in candidates)


def has_tests(root: Path, tracked: list[str]) -> bool:
    if exists_any(root, ("test", "tests", "spec")):
        return True
    pattern = re.compile(
        r"((^|/)test_[^/]+\.py$|_test\.(go|py|rb)$|\.test\.(ts|js|tsx|jsx)$|_spec\.rb$)"
    )
    return any(pattern.search(path) for path in tracked)


def dataset_columns(file: Path, suffix: str) -> list[str] | None:
    """Return column names only. Row values never leave this function."""
    try:
        with file.open("r", encoding="utf-8", errors="replace", newline="") as handle:
            header = handle.read(HEADER_READ_BYTES).splitlines()[0]
    except (OSError, IndexError):
        return None
    if suffix in TEXT_RECORDS:
        try:
            record = json.loads(header)
        except json.JSONDecodeError:
            return None
        return sorted(record) if isinstance(record, dict) else None
    delimiter = "\t" if suffix == ".tsv" else "|" if suffix == ".psv" else ","
    return [column.strip().strip('"') for column in header.split(delimiter) if column.strip()]


def inspect_datasets(root: Path, tracked: list[str]) -> list[dict[str, Any]]:
    datasets: list[dict[str, Any]] = []
    for path in sorted(tracked):
        if not DATASET_PATH.search(path) or SAFE_SAMPLE.search(path):
            continue
        file = root / path
        suffix = file.suffix.lower()
        columns = (
            dataset_columns(file, suffix)
            if suffix in TEXT_TABULAR + TEXT_RECORDS and file.is_file()
            else None
        )
        datasets.append(
            {
                "path": path,
                "bytes": file.stat().st_size if file.is_file() else None,
                "readable": columns is not None,
                "columns": len(columns) if columns else None,
                "identity_columns": sorted(
                    {column for column in columns if IDENTITY_COLUMN.search(column)}
                )[:MAX_REPORTED_COLUMNS]
                if columns
                else [],
            }
        )
    return datasets[:MAX_REPORTED_DATASETS]


def find_history_references(
    root: Path, tracked: list[str], prefixes: tuple[str, ...]
) -> list[dict[str, Any]]:
    """Return files outside the planning records that point into them."""
    referrers: list[dict[str, Any]] = []
    for relative in tracked:
        lowered = relative.lower()
        if any(lowered.startswith(f"{prefix}/") for prefix in prefixes):
            continue
        file = root / relative
        try:
            raw = file.read_bytes()[:REFERENCE_READ_BYTES]
        except OSError:
            continue
        if b"\x00" in raw[:8192]:
            continue
        text = raw.decode("utf-8", errors="replace").lower()
        matched = {prefix: text.count(f"{prefix}/") for prefix in prefixes}
        matched = {prefix: count for prefix, count in matched.items() if count}
        if matched:
            referrers.append(
                {
                    "path": relative,
                    "references": sum(matched.values()),
                    "targets": sorted(matched),
                }
            )
    referrers.sort(key=lambda item: (-item["references"], item["path"]))
    return referrers


def tool_version(tool: str) -> str | None:
    arguments = ["version"] if tool == "gitleaks" else ["--version"]
    try:
        result = subprocess.run(
            [tool, *arguments],
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
            check=False,
            timeout=60,
        )
    except (OSError, subprocess.SubprocessError):
        return None
    reported = [line.strip() for line in (result.stdout + result.stderr).splitlines() if line.strip()]
    return reported[0] if reported else None


def summarise_gitleaks(report_path: Path) -> tuple[int, dict[str, int], list[str]]:
    """Return counts and locations only; secret values never leave the report file."""
    try:
        parsed = json.loads(report_path.read_text(encoding="utf-8") or "[]")
    except (OSError, json.JSONDecodeError) as error:
        raise ValueError("Gitleaks did not produce a readable JSON report") from error
    if not isinstance(parsed, list):
        raise ValueError("Gitleaks report must contain a JSON list")
    findings = parsed
    by_rule = Counter(str(finding.get("RuleID") or "unknown") for finding in findings)
    locations = sorted(
        {
            f"{finding.get('File') or 'unknown'}@{str(finding.get('Commit') or 'working-tree')[:12]}"
            for finding in findings
        }
    )
    return len(findings), dict(by_rule.most_common()), locations[:MAX_REPORTED_LOCATIONS]


def summarise_trufflehog(stdout: str) -> tuple[int, dict[str, int], list[str]]:
    detectors: Counter[str] = Counter()
    locations: set[str] = set()
    total = 0
    for line in stdout.splitlines():
        line = line.strip()
        if not line.startswith("{"):
            continue
        try:
            finding = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not isinstance(finding, dict) or "DetectorName" not in finding:
            continue
        total += 1
        detectors[str(finding.get("DetectorName") or "unknown")] += 1
        source = finding.get("SourceMetadata")
        git = source.get("Data", {}).get("Git", {}) if isinstance(source, dict) else {}
        if isinstance(git, dict):
            locations.add(f"{git.get('file') or 'unknown'}@{str(git.get('commit') or 'unknown')[:12]}")
    return total, dict(detectors.most_common()), sorted(locations)[:MAX_REPORTED_LOCATIONS]


def scan_history(root: Path, is_git: bool, skip: bool) -> dict[str, object]:
    available = [tool for tool in HISTORY_SCANNERS if shutil.which(tool)]
    result: dict[str, object] = {"available_tools": available, "performed": False, "tool": None}
    if not is_git:
        return {**result, "status": "unavailable-not-a-git-repository"}
    if not available:
        return {**result, "status": "unavailable-no-scanner-installed"}
    if skip:
        return {**result, "status": "skipped-by-request"}

    tool = available[0]
    scratch = Path(tempfile.mkdtemp(prefix="debranding-history-"))
    report_path = scratch / "history-findings.json"
    if tool == "gitleaks":
        command = [
            "gitleaks", "git", str(root), "--redact", "--no-banner",
            "--report-format", "json", "--report-path", str(report_path),
        ]
    else:
        command = ["trufflehog", "git", f"file://{root}", "--json", "--no-update"]

    try:
        completed = subprocess.run(
            command,
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
            check=False,
            timeout=HISTORY_SCAN_TIMEOUT_SECONDS,
        )
    except subprocess.TimeoutExpired:
        return {**result, "tool": tool, "status": "timed-out", "timeout_seconds": HISTORY_SCAN_TIMEOUT_SECONDS}
    except OSError as error:
        return {**result, "tool": tool, "status": "error", "detail": str(error)}

    if tool == "gitleaks":
        try:
            count, by_rule, locations = summarise_gitleaks(report_path)
        except ValueError as error:
            return {
                **result,
                "tool": tool,
                "status": "error",
                "exit_code": completed.returncode,
                "detail": str(error),
            }
        failed = completed.returncode not in (0, 1) or (
            completed.returncode == 1 and count == 0
        )
    else:
        count, by_rule, locations = summarise_trufflehog(completed.stdout)
        failed = completed.returncode != 0 and count == 0

    if failed:
        return {
            **result,
            "tool": tool,
            "status": "error",
            "exit_code": completed.returncode,
            "detail": completed.stderr.strip()[:500],
        }
    return {
        **result,
        "tool": tool,
        "tool_version": tool_version(tool),
        "performed": True,
        "status": "findings" if count else "clean",
        "finding_count": count,
        "findings_by_rule": by_rule,
        "finding_locations": locations,
        "report_path": str(report_path) if tool == "gitleaks" else None,
    }


def main() -> int:
    args = parse_args()
    root = args.root.resolve()
    if not root.is_dir():
        print(f"error: repository root does not exist: {root}", file=sys.stderr)
        return 2

    tracked = git_lines(root, "ls-files")
    is_git = tracked is not None
    tracked = tracked or []
    tags = git_lines(root, "tag", "--list", "v[0-9]*") or []
    suspicious = sorted(
        path for path in tracked if SUSPICIOUS_PATH.search(path) and not SAFE_SAMPLE.search(path)
    )
    workflow_artifact_path = (
        args.workflow_artifact_path.strip("/").lower() or DEFAULT_WORKFLOW_ARTIFACT_PATH
    )
    history_prefixes = tuple(
        sorted(
            {
                *(prefix.strip("/").lower() for prefix in DEFAULT_HISTORY_PATHS),
                *(prefix.strip("/").lower() for prefix in args.history_path),
                workflow_artifact_path,
            }
        )
    )
    datasets = inspect_datasets(root, tracked)
    history_referrers = find_history_references(root, tracked, history_prefixes)
    workflow_artifacts = sorted(
        path for path in tracked if path.lower().startswith(f"{workflow_artifact_path}/")
    )
    indicators = {
        "README": exists_any(root, ("README.md", "README.rst", "README.txt", "README")),
        "LICENSE": exists_any(root, ("LICENSE", "LICENSE.md", "LICENSE.txt", "COPYING", "COPYING.md")),
        "CONTRIBUTING": exists_any(root, ("CONTRIBUTING.md", "CONTRIBUTING.rst", ".github/CONTRIBUTING.md")),
        "SECURITY policy": exists_any(root, ("SECURITY.md", ".github/SECURITY.md")),
        "Code of conduct": exists_any(root, ("CODE_OF_CONDUCT.md", ".github/CODE_OF_CONDUCT.md")),
        ".gitignore": (root / ".gitignore").exists(),
        ".editorconfig": (root / ".editorconfig").exists(),
        "CI workflows": any((root / ".github/workflows").glob("*.y*ml")),
        "Automated dependency updates": exists_any(
            root,
            (
                ".github/dependabot.yml",
                ".github/dependabot.yaml",
                "renovate.json",
                "renovate.json5",
                ".github/renovate.json",
                ".github/renovate.json5",
            ),
        ),
        "Tests": has_tests(root, tracked),
        "Semver tags": any(re.fullmatch(r"v\d+\.\d+\.\d+(?:[-+].+)?", tag) for tag in tags),
    }
    history_scanning = scan_history(root, is_git, args.skip_history_scan)
    report = {
        "schema_version": 1,
        "scan": {
            "root": str(root),
            "scanned_at": datetime.now(UTC).isoformat(timespec="seconds"),
            "git_repository": is_git,
        },
        "indicators": indicators,
        "summary": {
            "present": sum(indicators.values()),
            "total": len(indicators),
            "suspicious_tracked_paths": len(suspicious),
            "committed_datasets": len(datasets),
            "datasets_with_identity_columns": sum(1 for item in datasets if item["identity_columns"]),
            "unreadable_datasets": sum(1 for item in datasets if not item["readable"]),
            "history_scan_status": history_scanning["status"],
            "files_referencing_history": len(history_referrers),
            "references_to_history": sum(item["references"] for item in history_referrers),
            "workflow_artifacts": len(workflow_artifacts),
        },
        "suspicious_tracked_paths": suspicious,
        "committed_datasets": datasets,
        "history_paths": list(history_prefixes),
        "workflow_artifact_path": workflow_artifact_path,
        "files_referencing_history": history_referrers[:MAX_REPORTED_REFERRERS],
        # Expected in the working tree; listed so the published tree can be checked against it.
        "workflow_artifacts": workflow_artifacts,
        "history_scanning": history_scanning,
        "manual_external_surfaces": [
            "Branches and tags",
            "CI logs, caches, and artifacts",
            "Releases and attached files",
            "Issues, pull requests, discussions, and wiki",
            "Package and container registries",
            "CI/CD variables, environments, service connections, and secrets",
            "External documentation, monitoring, and telemetry",
            "Deployed resources, DNS, certificates, and identities",
        ],
        "limitations": [
            "Indicators are discussion prompts, not an automatic release verdict.",
            "Suspicious paths are reported without reading or exposing their values.",
            "History findings are reported as rule, path, and commit only, never as values.",
            "A clean history scan covers reachable commits only, not external surfaces.",
        ],
    }
    rendered = json.dumps(report, indent=2, ensure_ascii=True) + "\n"
    if args.output:
        output = args.output if args.output.is_absolute() else Path.cwd() / args.output
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_text(rendered, encoding="utf-8")
        print(f"wrote readiness indicators to {output}")
    else:
        sys.stdout.write(rendered)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())