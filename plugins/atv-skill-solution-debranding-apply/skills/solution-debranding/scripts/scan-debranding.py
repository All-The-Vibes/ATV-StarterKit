#!/usr/bin/env python3
"""Collect redacted, deterministic evidence of customer identifiers."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
from collections import Counter
from dataclasses import asdict, dataclass
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

SKIP_DIRECTORIES = {
    ".git",
    ".hg",
    ".svn",
    ".next",
    ".nuxt",
    ".pytest_cache",
    ".tox",
    ".venv",
    "__pycache__",
    "build",
    "coverage",
    "dist",
    "node_modules",
    "target",
    "vendor",
}
MAX_FILE_BYTES = 5 * 1024 * 1024
SUSPICIOUS_SUFFIXES = {".cer", ".crt", ".der", ".jks", ".key", ".p12", ".pem", ".pfx"}
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
# Agent tooling ships on public release, so it is branded, but it is not the product
# under review. Note that .github/workflows stays product: CI is real infrastructure.
DEFAULT_TOOLING_PATHS = (
    ".agents",
    ".atv",
    ".claude",
    ".context",
    ".github/agents",
    ".github/chatmodes",
    ".github/instructions",
    ".github/prompts",
    ".github/skills",
)
# Planning records do not always sit in a planning directory, so match the name too.
HISTORY_NAME_PATTERN = re.compile(
    r"(dashboard|sprint|retro|standup|brainstorm|ideation|post-?mortem|"
    r"meeting-?notes|uat|backlog|roadmap)",
    re.IGNORECASE,
)
HISTORY_NAME_ROOT = "docs"
# Single tokens like "api" or "svc" match everywhere, so derived variants must keep a separator.
DERIVED_TERM_MIN_LENGTH = 5
REPOSITORY_NAME_SEPARATORS = re.compile(r"[-_.]+")
UUID_LITERAL_PATTERN = re.compile(
    r"(?<![0-9a-f])[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}(?![0-9a-f])",
    re.IGNORECASE,
)
MICROSOFT_AUTHORITY_PATTERN = re.compile(
    r"https://login\.microsoftonline\.com/([^/'\"\s?#]+)",
    re.IGNORECASE,
)
GENERIC_MICROSOFT_AUTHORITIES = {"common", "consumers", "organizations"}
PARAMETERIZED_AUTHORITY_MARKERS = ("{", "}", "$", "%", "<", ">")


@dataclass(frozen=True)
class Term:
    category: str
    value: str
    confidence: str
    pattern: re.Pattern[str] | None = None


@dataclass(frozen=True)
class Finding:
    id: str
    category: str
    path: str
    surface: str
    line: int | None
    confidence: str
    evidence: str
    disposition: str
    remediation: str


TOP_PATHS = 40


def top_paths(findings: list[Finding], surface: str) -> list[dict[str, Any]]:
    counts = Counter(finding.path for finding in findings if finding.surface == surface)
    return [{"path": path, "findings": n} for path, n in counts.most_common(TOP_PATHS)]


def render_summary(report: dict[str, Any]) -> None:
    scan = report["scan"]
    coverage = report["coverage"]
    summary = report["summary"]

    print(f"root            {scan['root']}")
    print(f"mode            {scan['mode']}")
    print(f"files considered {scan['files_considered']}")
    print(
        f"files           {coverage['files_total']} total, "
        f"{coverage['files_by_surface']['product']} product, "
        f"{coverage['files_by_surface']['history']} history, "
        f"{coverage['files_by_surface']['tooling']} tooling, "
        f"{coverage['files_content_unread']} unread"
    )
    if coverage["skipped"]:
        print("skipped         " + ", ".join(f"{k}={v}" for k, v in coverage["skipped"].items()))
    print(
        f"findings        {summary['findings']} "
        f"({summary['by_surface']['product']} product, {summary['by_surface']['history']} history, "
        f"{summary['by_surface']['tooling']} tooling, "
        f"{summary['blocking']} blocking, {summary['gated']} gated)"
    )
    print("by category     " + ", ".join(f"{k}={v}" for k, v in summary["by_category"].items()))
    if scan["derived_terms"]:
        print("derived terms   " + ", ".join(scan["derived_terms"]) + " (inferred; confirm before trusting)")

    for label, key in (
        ("product", "top_product_files"),
        ("history", "top_history_files"),
        ("tooling", "top_tooling_files"),
    ):
        rows = summary[key]
        if not rows:
            continue
        print(f"\ntop {label} files by finding count")
        for row in rows:
            print(f"  {row['findings']:>5}  {row['path']}")

    if summary["suspicious_files"]:
        print("\nsuspicious files")
        for path in summary["suspicious_files"]:
            print(f"         {path}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path.cwd())
    parser.add_argument("--source-customer", required=True)
    parser.add_argument("--alias", action="append", default=[])
    parser.add_argument("--domain", action="append", default=[])
    parser.add_argument("--resource-prefix", action="append", default=[])
    parser.add_argument("--mode", choices=("full", "delta"), default="full")
    parser.add_argument("--base", default="origin/main", help="Git base for delta mode")
    parser.add_argument(
        "--no-derive-terms",
        action="store_true",
        help=(
            "Do not infer terms from the Git remote and directory name. "
            "Derivation is on by default because a repository's own slug is branding "
            "that no supplied term list tends to include."
        ),
    )
    parser.add_argument(
        "--history-path",
        action="append",
        default=[],
        help=(
            "Repository-relative directory holding planning or history artifacts. "
            f"Findings there are tagged surface=history. Defaults to {', '.join(DEFAULT_HISTORY_PATHS)}."
        ),
    )
    parser.add_argument(
        "--tooling-path",
        action="append",
        default=[],
        help=(
            "Repository-relative directory holding agent tooling that ships with the repository. "
            "Findings there are tagged surface=tooling: still branded, but outside the "
            f"coupling-review denominator. Defaults to {', '.join(DEFAULT_TOOLING_PATHS)}."
        ),
    )
    parser.add_argument("--output", type=Path)
    parser.add_argument(
        "--summary",
        action="store_true",
        help="Print a human-readable rollup instead of the full JSON",
    )
    parser.add_argument(
        "--fail-on-blocking",
        action="store_true",
        help="Exit 1 when the scan reports one or more blocking findings",
    )
    return parser.parse_args()


def normalize_prefixes(values: list[str], defaults: tuple[str, ...]) -> tuple[str, ...]:
    selected = [*defaults, *values]
    normalized = {value.strip().strip("/").casefold() for value in selected}
    return tuple(sorted(prefix for prefix in normalized if prefix))


def matches_prefix(lowered: str, prefixes: tuple[str, ...]) -> bool:
    return any(lowered == prefix or lowered.startswith(f"{prefix}/") for prefix in prefixes)


def surface_for(relative: str, history: tuple[str, ...], tooling: tuple[str, ...]) -> str:
    lowered = relative.casefold()
    if matches_prefix(lowered, tooling):
        return "tooling"
    if matches_prefix(lowered, history):
        return "history"
    if matches_prefix(lowered, (HISTORY_NAME_ROOT,)) and HISTORY_NAME_PATTERN.search(
        lowered.rsplit("/", 1)[-1]
    ):
        return "history"
    return "product"


def run_git(root: Path, *args: str) -> list[str]:
    result = subprocess.run(
        ["git", "-C", str(root), *args],
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
        check=False,
    )
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or "git command failed")
    return [line for line in result.stdout.splitlines() if line]


def full_files(root: Path) -> list[Path]:
    files: list[Path] = []
    for current, directories, names in os.walk(root):
        directories[:] = sorted(name for name in directories if name not in SKIP_DIRECTORIES)
        current_path = Path(current)
        files.extend(current_path / name for name in sorted(names))
    return files


def delta_files(root: Path, base: str) -> list[Path]:
    changed = set(run_git(root, "diff", "--name-only", "--diff-filter=ACMR", base))
    changed.update(run_git(root, "ls-files", "--others", "--exclude-standard"))
    return sorted(root / relative for relative in changed if (root / relative).is_file())


def parse_remote_slug(url: str) -> tuple[str, str] | None:
    """Return (owner, repository) from an https or ssh remote URL."""
    cleaned = url.strip().removesuffix(".git")
    if "://" in cleaned:
        path = cleaned.split("://", 1)[1].split("/", 1)[-1]
    elif "@" in cleaned and ":" in cleaned:
        path = cleaned.split(":", 1)[1]
    else:
        path = cleaned
    segments = [segment for segment in path.split("/") if segment]
    if len(segments) < 2:
        return None
    return segments[-2], segments[-1]


def repository_name_variants(name: str) -> list[str]:
    """Return shortened forms of a repository name, longest first.

    Teams abbreviate their own repository in prose, so `acme-platform-widget-svc`
    also circulates as `platform-widget-svc` and `widget-svc`.
    """
    segments = [segment for segment in REPOSITORY_NAME_SEPARATORS.split(name) if segment]
    variants = []
    for index in range(1, len(segments)):
        candidate = "-".join(segments[index:])
        if "-" in candidate and len(candidate) >= DERIVED_TERM_MIN_LENGTH:
            variants.append(candidate)
    return variants


def derive_repository_terms(root: Path) -> list[Term]:
    """Infer terms from the repository's own identity rather than its contents."""
    candidates: list[Term] = []
    names: list[str] = []
    try:
        remotes = run_git(root, "remote")
    except (OSError, RuntimeError):
        remotes = []
    remote = "origin" if "origin" in remotes else next(iter(remotes), None)
    if remote:
        try:
            urls = run_git(root, "remote", "get-url", remote)
        except (OSError, RuntimeError):
            urls = []
        for url in urls:
            parsed = parse_remote_slug(url)
            if not parsed:
                continue
            owner, repository = parsed
            candidates.append(Term("alias", owner, "medium"))
            candidates.append(Term("alias", f"{owner}/{repository}", "medium"))
            names.append(repository)
    names.append(root.name)
    for name in names:
        if len(name) >= DERIVED_TERM_MIN_LENGTH:
            candidates.append(Term("resource-prefix", name, "medium"))
        candidates.extend(
            Term("resource-prefix", variant, "low") for variant in repository_name_variants(name)
        )
    return candidates


def build_terms(args: argparse.Namespace, derived: list[Term] | None = None) -> list[Term]:
    raw_terms = list(derived or [])
    raw_terms.append(Term("source-customer", args.source_customer, "high"))
    raw_terms.extend(Term("alias", value, "medium") for value in args.alias)
    raw_terms.extend(Term("domain", value, "high") for value in args.domain)
    raw_terms.extend(Term("resource-prefix", value, "medium") for value in args.resource_prefix)
    unique: dict[tuple[str, str], Term] = {}
    for term in raw_terms:
        value = term.value.strip()
        if value:
            escaped = re.escape(value)
            # Treat _ as a boundary so codenames in snake_case identifiers are caught.
            pattern = re.compile(rf"(?<![a-zA-Z0-9]){escaped}(?![a-zA-Z0-9])", re.IGNORECASE)
            unique[(term.category, value.casefold())] = Term(term.category, value, term.confidence, pattern)
    return sorted(unique.values(), key=lambda item: (item.category, item.value.casefold()))


def stable_id(category: str, path: str, line: int | None, kind: str) -> str:
    value = f"{category}\0{path}\0{line or 0}\0{kind}".encode()
    return "DBR-" + hashlib.sha256(value).hexdigest()[:12].upper()


def scan_path_name(relative: str, terms: list[Term], surface: str) -> list[Finding]:
    findings: list[Finding] = []
    for term in terms:
        if term.value.casefold() in relative.casefold():
            gated = surface == "history" or term.category in {"domain", "resource-prefix"}
            findings.append(
                Finding(
                    stable_id(term.category, relative, None, "path"),
                    term.category,
                    relative,
                    surface,
                    None,
                    term.confidence,
                    f"Path contains <{term.category}:redacted>",
                    "[GATED]" if gated else "[BLOCKS]",
                    "human-review" if gated else "direct-replacement",
                )
            )
    return findings


def scan_content(path: Path, relative: str, terms: list[Term], surface: str) -> tuple[list[Finding], str | None]:
    try:
        if path.stat().st_size > MAX_FILE_BYTES:
            return [], "file-too-large"
        raw = path.read_bytes()
    except OSError:
        return [], "unreadable"
    if b"\x00" in raw[:8192]:
        return [], "binary"
    text = raw.decode("utf-8", errors="replace")
    findings: list[Finding] = []
    for line_number, line in enumerate(text.splitlines(), start=1):
        if UUID_LITERAL_PATTERN.search(line):
            findings.append(
                Finding(
                    stable_id("hardcoded-opaque-identifier", relative, line_number, "structural-content"),
                    "hardcoded-opaque-identifier",
                    relative,
                    surface,
                    line_number,
                    "medium",
                    "Line contains a UUID-shaped literal; value redacted",
                    "[GATED]" if surface == "history" else "[BLOCKS]",
                    "human-review" if surface == "history" else "classify-or-parameterize",
                )
            )
        authority = MICROSOFT_AUTHORITY_PATTERN.search(line)
        if authority:
            authority_tenant = authority.group(1)
            is_parameterized = any(
                marker in authority_tenant for marker in PARAMETERIZED_AUTHORITY_MARKERS
            )
            if (
                authority_tenant.casefold() not in GENERIC_MICROSOFT_AUTHORITIES
                and not is_parameterized
            ):
                findings.append(
                    Finding(
                        stable_id("hardcoded-cloud-identity", relative, line_number, "structural-content"),
                        "hardcoded-cloud-identity",
                        relative,
                        surface,
                        line_number,
                        "high",
                        "Line hardcodes a tenant-specific Microsoft authority; value redacted",
                        "[GATED]" if surface == "history" else "[BLOCKS]",
                        "human-review" if surface == "history" else "parameterize",
                    )
                )
        for term in terms:
            if not (term.pattern and term.pattern.search(line)):
                continue
            gated = surface == "history" or term.category in {"domain", "resource-prefix"}
            findings.append(
                Finding(
                    stable_id(term.category, relative, line_number, "content"),
                    term.category,
                    relative,
                    surface,
                    line_number,
                    term.confidence,
                    f"Matched <{term.category}:redacted>; source line withheld",
                    "[GATED]" if gated else "[BLOCKS]",
                    "human-review" if gated else "classify",
                )
            )
    return findings, None


def suspicious_file_finding(path: Path, relative: str, surface: str) -> Finding | None:
    lower_name = path.name.casefold()
    suspicious = path.suffix.casefold() in SUSPICIOUS_SUFFIXES or lower_name.startswith(".env")
    if not suspicious:
        return None
    return Finding(
        stable_id("suspicious-file", relative, None, "path"),
        "suspicious-file",
        relative,
        surface,
        None,
        "medium",
        "Path matches a credential, certificate, key, or environment-file pattern",
        "[GATED]",
        "human-review",
    )


def main() -> int:
    args = parse_args()
    root = args.root.resolve()
    if not root.is_dir():
        print(f"error: repository root does not exist: {root}", file=sys.stderr)
        return 2
    derived = [] if args.no_derive_terms else derive_repository_terms(root)
    terms = build_terms(args, derived)
    derived_values = sorted({term.value for term in derived})
    prefixes = normalize_prefixes(args.history_path, DEFAULT_HISTORY_PATHS)
    tooling_prefixes = normalize_prefixes(args.tooling_path, DEFAULT_TOOLING_PATHS)
    try:
        files = full_files(root) if args.mode == "full" else delta_files(root, args.base)
    except (OSError, RuntimeError) as error:
        print(f"error: unable to select files: {error}", file=sys.stderr)
        return 2

    findings: list[Finding] = []
    skipped: dict[str, int] = {}
    product_files: list[str] = []
    history_files: list[str] = []
    tooling_files: list[str] = []
    unread_files = 0
    for path in files:
        try:
            relative = path.relative_to(root).as_posix()
        except ValueError:
            continue
        surface = surface_for(relative, prefixes, tooling_prefixes)
        if surface == "history":
            history_files.append(relative)
        elif surface == "tooling":
            tooling_files.append(relative)
        else:
            product_files.append(relative)
        findings.extend(scan_path_name(relative, terms, surface))
        suspicious = suspicious_file_finding(path, relative, surface)
        if suspicious:
            findings.append(suspicious)
        content_findings, skip_reason = scan_content(path, relative, terms, surface)
        findings.extend(content_findings)
        if skip_reason:
            skipped[skip_reason] = skipped.get(skip_reason, 0) + 1
            unread_files += 1

    unique = {finding.id: finding for finding in findings}
    ordered = sorted(unique.values(), key=lambda item: (item.path.casefold(), item.line or 0, item.id))
    report = {
        "schema_version": 1,
        "scan": {
            "root": str(root),
            "mode": args.mode,
            "base": args.base if args.mode == "delta" else None,
            "scanned_at": datetime.now(UTC).isoformat(timespec="seconds"),
            "files_considered": len(files),
            "history_paths": list(prefixes),
            "tooling_paths": list(tooling_prefixes),
            "terms_by_category": {
                category: sum(1 for term in terms if term.category == category)
                for category in sorted({term.category for term in terms})
            },
            # Inferred, so confirm them: a remote's owner is often not the org named in the tree.
            "derived_terms": derived_values,
        },
        "coverage": {
            "files_total": len(product_files) + len(history_files) + len(tooling_files),
            "files_content_read": (
                len(product_files) + len(history_files) + len(tooling_files) - unread_files
            ),
            "files_content_unread": unread_files,
            "files_by_surface": {
                "product": len(product_files),
                "history": len(history_files),
                "tooling": len(tooling_files),
            },
            "skipped": dict(sorted(skipped.items())),
            "product_files": sorted(product_files),
            # Each of these needs a keep, anonymize, relocate, or remove decision.
            "history_files": sorted(history_files),
            "tooling_files": sorted(tooling_files),
            "limitations": [
                "Content evidence is redacted.",
                "Binary, image, generated, ignored, Git LFS, and Git-history review may require separate tools.",
                "Findings require semantic classification before remediation.",
                "Findings tagged surface=history are planning or history artifacts. Never rewrite them in place.",
                "Findings tagged surface=tooling are agent tooling that ships with the repository. They still need renaming, but they are not product code and are not coupling-reviewed.",
                "product_files is the review denominator. A lens that reads fewer of them has not covered the repository.",
            ],
        },
        "summary": {
            "findings": len(ordered),
            "blocking": sum(1 for finding in ordered if finding.disposition == "[BLOCKS]"),
            "gated": sum(1 for finding in ordered if finding.disposition == "[GATED]"),
            "by_surface": {
                "product": sum(1 for finding in ordered if finding.surface == "product"),
                "history": sum(1 for finding in ordered if finding.surface == "history"),
                "tooling": sum(1 for finding in ordered if finding.surface == "tooling"),
            },
            "by_category": {
                category: sum(1 for finding in ordered if finding.category == category)
                for category in sorted({finding.category for finding in ordered})
            },
            "top_product_files": top_paths(ordered, "product"),
            "top_history_files": top_paths(ordered, "history"),
            "top_tooling_files": top_paths(ordered, "tooling"),
            "suspicious_files": sorted(
                {finding.path for finding in ordered if finding.category == "suspicious-file"}
            ),
        },
        "findings": [asdict(finding) for finding in ordered],
    }
    rendered = json.dumps(report, indent=2, ensure_ascii=True) + "\n"
    if args.output:
        output = args.output if args.output.is_absolute() else Path.cwd() / args.output
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_text(rendered, encoding="utf-8")
        print(f"wrote {len(ordered)} redacted finding(s) to {output}")
    elif not args.summary:
        sys.stdout.write(rendered)

    if args.summary:
        render_summary(report)
    if args.fail_on_blocking and report["summary"]["blocking"]:
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())