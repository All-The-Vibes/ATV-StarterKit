#!/usr/bin/env python3
"""Route product files to coupling lenses and emit deterministic review batches."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any

DEFAULT_BATCH_SIZE = 40
ADVERSARIAL_SAMPLE = 60

# Files that cannot carry rename-surviving coupling. Each entry records why.
EXCLUSIONS: tuple[tuple[str, str], ...] = (
    # output/ and outputs/ are deliberately absent: in data projects they hold committed results.
    (r"(^|/)(dist|build|out|coverage|htmlcov|__pycache__|\.pytest_cache|\.next|\.turbo|\.svelte-kit)/", "generated-output"),
    (r"(^|/)(vendor|third_party|node_modules)/", "vendored"),
    (r"(^|/)[^/]*\.(lock)$", "lockfile"),
    (r"(^|/)(package-lock\.json|yarn\.lock|pnpm-lock\.yaml|poetry\.lock|Cargo\.lock|go\.sum)$", "lockfile"),
    (r"\.min\.(js|css)$", "minified"),
    (r"\.(map|snap)$", "generated-output"),
    # SVG stays reviewable: it is text, and logo geometry and brand hexes live in it.
    (r"\.(png|jpe?g|gif|webp|ico|bmp|tiff?)$", "image-asset"),
    (r"\.(woff2?|ttf|eot|otf)$", "font-asset"),
    # Distinct from binary-asset: these hold records, so exclusion means unread data, not absent data.
    (r"\.(xlsx?|xlsm|xlsb|parquet|avro|sav|dta|db|sqlite3?|mdb|accdb)$", "tabular-binary"),
    (r"\.(zip|gz|tgz|bz2|7z|rar|jar|war|whl|exe|dll|so|dylib|pdf|docx?|pptx?)$", "binary-asset"),
)

# Small, high-signal attribution surfaces a sample must never skip.
ALWAYS_REVIEW: tuple[str, ...] = (
    r"(^|/)(CODEOWNERS|AUTHORS|CONTRIBUTORS|MAINTAINERS|NOTICE|CITATION\.cff)$",
    r"(^|/)(LICENSE|COPYING)[^/]*$",
    r"(^|/)\.mailmap$",
    r"(^|/)catalog-info\.ya?ml$",
)

# First match wins, so order encodes precedence.
ROUTES: tuple[tuple[str, str], ...] = (
    # Data routes first: a ledger export under schemas/ is a record set, not a schema.
    (r"\.(csv|tsv|psv|jsonl|ndjson)$", "embedded-data"),
    (r"(^|/)(data|datasets?|fixtures?|seeds?|samples?|corpus|golden|evals?|evaluations?|baselines?|results?|outputs?)[^/]*/", "embedded-data"),
    (r"\.(css|scss|sass|less|styl|svg|drawio)$", "visual-identity"),
    (r"(^|/)(migrations?|alembic|schema|schemas)/", "data-model"),
    (r"(^|/)(models?|entities|entity|dto|dtos|contracts?|enums?)/", "data-model"),
    (r"\.(sql|prisma|graphql|gql|avsc|proto)$", "data-model"),
    (r"(^|/)[^/]*(schema|model|entity|contract|taxonomy|enum)[^/]*\.[a-z]+$", "data-model"),
    (r"(^|/)(auth|authz|authn|identity|rbac|permissions?|roles?|tenants?|security|middleware)/", "tenancy-authorization"),
    (r"(^|/)[^/]*(auth|role|permission|tenant|session|login|policy|principal)[^/]*\.[a-z]+$", "tenancy-authorization"),
    # Auth routes first, so a design-token file is never confused with a session-token file.
    (r"(^|/)(styles?|themes?|branding|design-tokens?|stylesheets)/", "visual-identity"),
    (r"(^|/)[^/]*(theme|palette|colou?rs?|typography|typeface|design-tokens?|tailwind\.config|brand)[^/]*\.[a-z]+$", "visual-identity"),
    (r"(^|/)(infra|infrastructure|terraform|bicep|deploy|deployment|charts?|k8s|kubernetes|helm|pipelines?|proxy|proxies|\.backstage|\.github)/", "integration"),
    (r"(^|/)(config|settings|clients?|gateway|integrations?|connectors?)/", "integration"),
    # Platform descriptors carry owners, cost codes, and internal URLs that no code search reaches.
    (r"(^|/)(catalog-info\.ya?ml|sonar-project\.properties|codecov\.ya?ml|\.snyk|renovate\.json5?|\.tool-versions|Procfile|app\.json)$", "integration"),
    (r"(Dockerfile|docker-compose[^/]*\.ya?ml|\.tf|\.tfvars|\.bicep)$", "integration"),
    # A corporate CA or proxy config couples the build to the customer's network, which no rename undoes.
    (r"\.(crt|cer|pem|der|p12|pfx|jks|keystore|truststore)$", "integration"),
    (r"(^|/)(Caddyfile|nginx\.conf|httpd\.conf|haproxy\.cfg|\.npmrc|\.pip/pip\.conf|pip\.conf)$", "integration"),
    (r"(^|/)\.env(\.[^/]*)?$", "integration"),
    (r"(^|/)(pyproject\.toml|package\.json|setup\.cfg|setup\.py|Cargo\.toml|go\.mod|pom\.xml|build\.gradle[^/]*|[^/]*\.(csproj|gemspec|podspec|nuspec))$", "integration"),
    (r"(^|/)[^/]*(config|settings|client|gateway|endpoint|provider|connection)[^/]*\.[a-z]+$", "integration"),
    # Last before the fallback: prose only claims a file no code lens already owns.
    (r"(^|/)(docs?|documentation|adrs?|rfcs?|handbook|runbooks?|wiki)/", "narrative-documentation"),
    (r"\.(md|mdx|rst|adoc|txt)$", "narrative-documentation"),
)

FALLBACK_LENS = "business-logic"
ADVERSARIAL_LENS = "adversarial-attribution"


def classify(path: str) -> tuple[str | None, str | None]:
    for pattern, reason in EXCLUSIONS:
        if re.search(pattern, path, re.IGNORECASE):
            return None, reason
    for pattern, lens in ROUTES:
        if re.search(pattern, path, re.IGNORECASE):
            return lens, None
    return FALLBACK_LENS, None


def sample(paths: list[str], size: int) -> list[str]:
    if len(paths) <= size:
        return list(paths)
    step = len(paths) / size
    return [paths[int(index * step)] for index in range(size)]


def build(manifest: list[str], batch_size: int, sample_size: int) -> dict[str, Any]:
    routed: dict[str, list[str]] = {}
    excluded: list[dict[str, str]] = []

    for path in sorted(manifest):
        lens, reason = classify(path)
        if lens is None:
            excluded.append({"path": path, "reason": reason or "excluded"})
            continue
        routed.setdefault(lens, []).append(path)

    reviewable = sorted(item for paths in routed.values() for item in paths)
    mandatory = [
        path
        for path in reviewable
        if any(re.search(pattern, path, re.IGNORECASE) for pattern in ALWAYS_REVIEW)
    ]
    routed[ADVERSARIAL_LENS] = sorted(set(sample(reviewable, sample_size)) | set(mandatory))

    batches: list[dict[str, Any]] = []
    for lens in sorted(routed):
        paths = routed[lens]
        for index in range(0, len(paths), batch_size):
            chunk = paths[index : index + batch_size]
            batches.append(
                {
                    "id": f"{lens}-{index // batch_size + 1:03d}",
                    "lens": lens,
                    "files": chunk,
                }
            )

    return {
        "batch_size": batch_size,
        "summary": {
            "product_files": len(manifest),
            "excluded": len(excluded),
            "reviewable": len(reviewable),
            "batches": len(batches),
            "by_lens": {lens: len(paths) for lens, paths in sorted(routed.items())},
        },
        "excluded": excluded,
        "batches": batches,
    }


def split(plan: dict[str, Any], directory: Path) -> list[Path]:
    """Write one dispatchable file per batch so each subagent reads only its own assignment."""
    directory.mkdir(parents=True, exist_ok=True)
    written: list[Path] = []
    for batch in plan["batches"]:
        target = directory / f"{batch['id']}.json"
        target.write_text(json.dumps(batch, indent=2) + "\n", encoding="utf-8")
        written.append(target)
    return written


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--report", required=True, type=Path, help="Scan report with coverage.product_files")
    parser.add_argument("--batch-size", type=int, default=DEFAULT_BATCH_SIZE)
    parser.add_argument("--sample-size", type=int, default=ADVERSARIAL_SAMPLE)
    parser.add_argument("--output", type=Path)
    parser.add_argument(
        "--split-dir",
        type=Path,
        help="Also write each batch to <dir>/<batch-id>.json for one-per-subagent dispatch",
    )
    arguments = parser.parse_args(argv)

    report = json.loads(arguments.report.read_text(encoding="utf-8"))
    manifest = report.get("coverage", {}).get("product_files")
    if not manifest:
        raise SystemExit(f"{arguments.report}: scan report has no coverage.product_files manifest")

    plan = build(manifest, arguments.batch_size, arguments.sample_size)
    text = json.dumps(plan, indent=2)
    if arguments.output:
        arguments.output.write_text(text, encoding="utf-8")
        print(f"planned {plan['summary']['batches']} batch(es) to {arguments.output}", file=sys.stderr)
    else:
        sys.stdout.write(text + "\n")

    if arguments.split_dir:
        written = split(plan, arguments.split_dir)
        print(f"split {len(written)} batch file(s) into {arguments.split_dir}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
