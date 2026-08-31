from __future__ import annotations

import importlib.util
import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest import mock
from pathlib import Path

SKILL_ROOT = Path(__file__).resolve().parents[1]
REPOSITORY_ROOT = SKILL_ROOT.parent.parent
PUBLIC_ROOT = REPOSITORY_ROOT if (REPOSITORY_ROOT / "README.md").is_file() else SKILL_ROOT
SCAN_SCRIPT = SKILL_ROOT / "scripts" / "scan-debranding.py"
READINESS_SCRIPT = SKILL_ROOT / "scripts" / "check-readiness.py"
MERGE_SCRIPT = SKILL_ROOT / "scripts" / "merge-lens-findings.py"
BATCH_SCRIPT = SKILL_ROOT / "scripts" / "plan-lens-batches.py"


def run_script(script: Path, *arguments: str, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, str(script), *arguments],
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
        check=False,
        env=env,
    )


def path_without_history_scanners() -> dict[str, str]:
    """Return an environment where no history scanner resolves, but git still does."""
    entries = os.environ.get("PATH", "").split(os.pathsep)
    keep = [
        entry
        for entry in entries
        if entry and not any(shutil.which(tool, path=entry) for tool in ("gitleaks", "trufflehog"))
    ]
    return {**os.environ, "PATH": os.pathsep.join(keep)}


def path_with_stub_scanner(directory: Path) -> dict[str, str]:
    """Return an environment where a scanner resolves but is never executed."""
    stub = directory / ("gitleaks.bat" if os.name == "nt" else "gitleaks")
    script = "@exit /b 1\n" if os.name == "nt" else "#!/bin/sh\nexit 1\n"
    stub.write_text(script, encoding="utf-8")
    stub.chmod(0o755)
    return {**os.environ, "PATH": os.pathsep.join([str(directory), os.environ.get("PATH", "")])}


@unittest.skipUnless(importlib.util.find_spec("json"), "standard library unavailable")
class DebrandingScriptTests(unittest.TestCase):
    def test_merge_deduplicates_lenses_and_flags_schema_breaches(self) -> None:
        def lens(name: str, findings: list[dict[str, object]]) -> dict[str, object]:
            return {
                "lens": name,
                "findings": findings,
                "coverage": {"files_reviewed": ["src/api/config.py"], "not_reviewed": ["src/workers/"]},
            }

        shared = {"path": "src/api/config.py", "anchor": "L208", "survives_rename": "hardcoded gateway region"}
        payloads = [
            lens("data-model", [
                {**shared, "coupling_type": "data-model", "confidence": 50, "evidence": ["a"]},
                {"path": "src/api/tax.py", "anchor": "L10", "coupling_type": "data-model", "confidence": 25, "evidence": "b"},
                {"path": "src/api/drop.py", "anchor": "L1", "coupling_type": "data-model", "confidence": 0, "evidence": []},
            ]),
            lens("integration", [
                {**shared, "coupling_type": "integration", "confidence": 75, "evidence": ["c"]},
                {"path": "src/api/x.py", "anchor": "L2", "coupling_type": "integration", "confidence": 60, "evidence": []},
            ]),
        ]

        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            files = []
            for payload in payloads:
                target = root / f"{payload['lens']}.json"
                target.write_text(json.dumps(payload), encoding="utf-8")
                files.append(str(target))

            result = run_script(MERGE_SCRIPT, *files)

        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)

        self.assertEqual(report["summary"]["findings"], 1)
        merged = report["findings"][0]
        self.assertEqual(merged["confidence_anchor"], 75)
        self.assertEqual(merged["confidence"], "high")
        self.assertEqual(merged["coupling_types"], ["data-model", "integration"])
        self.assertEqual(merged["evidence"], ["a", "c"])

        # Confidence 25 is preserved as residual risk, 0 is dropped entirely.
        self.assertEqual([item["path"] for item in report["residual_risks"]], ["src/api/tax.py"])
        self.assertEqual(report["residual_risks"][0]["evidence"], ["b"])

        self.assertIn("business-logic", report["lenses_missing"])
        self.assertTrue(any("not an anchor" in problem for problem in report["problems"]))

    def test_merge_reports_file_coverage_against_the_scan_manifest(self) -> None:
        lens_payload = {
            "lens": "integration",
            "findings": [],
            "coverage": {
                "files_reviewed": ["src/a.py", "src/b.py", "docs/pm/notes.md"],
                "not_reviewed": ["src/c.py"],
            },
        }
        scan_report = {
            "coverage": {"product_files": ["src/a.py", "src/b.py", "src/c.py", "src/d.py"]}
        }

        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            lens_file = root / "integration.json"
            lens_file.write_text(json.dumps(lens_payload), encoding="utf-8")
            report_file = root / "scan-report.json"
            report_file.write_text(json.dumps(scan_report), encoding="utf-8")

            result = run_script(MERGE_SCRIPT, str(lens_file), "--report", str(report_file))

        self.assertEqual(result.returncode, 0, result.stderr)
        coverage = json.loads(result.stdout)["file_coverage"]

        self.assertEqual(coverage["denominator"], 4)
        self.assertEqual(coverage["files_reviewed"], 2)
        self.assertEqual(coverage["coverage_percent"], 50.0)
        self.assertEqual(coverage["unreviewed"], ["src/c.py", "src/d.py"])
        self.assertEqual(coverage["reviewed_outside_manifest"], ["docs/pm/notes.md"])
        self.assertTrue(
            any("never opened by any lens" in problem for problem in json.loads(result.stdout)["problems"])
        )

    def test_merge_fails_when_a_batch_skips_files_it_was_assigned(self) -> None:
        plan = {
            "batches": [
                {"id": "integration-001", "lens": "integration", "files": ["src/a.py", "src/b.py"]},
                {"id": "data-model-002", "lens": "data-model", "files": ["src/c.py"]},
            ],
            "excluded": [{"path": "src/logo.png", "reason": "image-asset"}],
        }
        scan_report = {
            "coverage": {"product_files": ["src/a.py", "src/b.py", "src/c.py", "src/logo.png"]}
        }
        # Reviews one of its two files, and the second batch never reports at all.
        lens_payload = {
            "lens": "integration",
            "batch_id": "integration-001",
            "findings": [],
            "coverage": {"files_reviewed": ["src/a.py"], "not_reviewed": []},
        }

        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            lens_file = root / "lens-01.json"
            lens_file.write_text(json.dumps(lens_payload), encoding="utf-8")
            plan_file = root / "lens-batches.json"
            plan_file.write_text(json.dumps(plan), encoding="utf-8")
            report_file = root / "scan-report.json"
            report_file.write_text(json.dumps(scan_report), encoding="utf-8")

            result = run_script(
                MERGE_SCRIPT,
                str(lens_file),
                "--report", str(report_file),
                "--batch-plan", str(plan_file),
            )

        self.assertEqual(result.returncode, 2, "a batch shortfall must fail the run")
        self.assertIn("batch contract breached", result.stderr)
        report = json.loads(result.stdout)

        contract = report["batch_contract"]
        self.assertTrue(contract["enforced"])
        self.assertEqual(contract["batches_assigned"], 2)
        self.assertEqual(contract["batches_complete"], 0)

        breaches = {item["batch"]: item for item in contract["breaches"]}
        self.assertEqual(breaches["integration-001"]["fault"], "incomplete")
        self.assertEqual(breaches["integration-001"]["unread"], ["src/b.py"])
        self.assertEqual(breaches["data-model-002"]["fault"], "never reported")

        # The excluded asset leaves the denominator so full coverage stays reachable.
        self.assertEqual(report["file_coverage"]["denominator"], 3)
        self.assertEqual(report["file_coverage"]["excluded_by_rule"], 1)

    def test_merge_passes_when_every_batch_returns_its_full_assignment(self) -> None:
        plan = {
            "batches": [{"id": "integration-001", "lens": "integration", "files": ["src/a.py", "src/b.py"]}],
            "excluded": [],
        }
        scan_report = {"coverage": {"product_files": ["src/a.py", "src/b.py"]}}
        lens_payload = {
            "lens": "integration",
            "batch_id": "integration-001",
            "findings": [],
            "coverage": {"files_reviewed": ["src/a.py", "src/b.py"], "not_reviewed": []},
        }

        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            lens_file = root / "lens-01.json"
            lens_file.write_text(json.dumps(lens_payload), encoding="utf-8")
            plan_file = root / "lens-batches.json"
            plan_file.write_text(json.dumps(plan), encoding="utf-8")
            report_file = root / "scan-report.json"
            report_file.write_text(json.dumps(scan_report), encoding="utf-8")

            result = run_script(
                MERGE_SCRIPT,
                str(lens_file),
                "--report", str(report_file),
                "--batch-plan", str(plan_file),
            )

        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(report["batch_contract"]["breaches"], [])
        self.assertEqual(report["file_coverage"]["coverage_percent"], 100.0)

    def test_batch_planner_routes_every_file_once_and_records_exclusions(self) -> None:
        manifest = [
            "src/api/migrations/0001_init.sql",
            "src/api/models/order.py",
            "src/api/auth/roles.py",
            "src/api/config.py",
            "infra/main.bicep",
            "src/api/routers/health.py",
            "src/api/tasks/sync.py",
            "docs/adr/0004-data-mapping.md",
            "README.md",
            "docs/schemas/ledger.sql",
            "src/frontend/styles/theme.css",
            "src/frontend/assets/logo.svg",
            "tailwind.config.js",
            "src/api/auth/tokens.py",
            "sample-data/orders.csv",
            "src/frontend/contoso.crt",
            "Caddyfile",
            "src/api/pyproject.toml",
            ".github/CODEOWNERS",
            "catalog-info.yaml",
            "proxy/setup-proxy.ps1",
            "output/evaluations/run-01/example-service/result.md",
            "package-lock.json",
            "src/frontend/public/logo.png",
            "sample-data/archive.xlsx",
            "dist/bundle.min.js",
        ]
        scan_report = {"coverage": {"product_files": manifest}}

        with tempfile.TemporaryDirectory() as directory:
            report_file = Path(directory) / "scan-report.json"
            report_file.write_text(json.dumps(scan_report), encoding="utf-8")
            result = run_script(BATCH_SCRIPT, "--report", str(report_file), "--batch-size", "2")

        self.assertEqual(result.returncode, 0, result.stderr)
        plan = json.loads(result.stdout)

        excluded = {item["path"]: item["reason"] for item in plan["excluded"]}
        self.assertEqual(excluded["package-lock.json"], "lockfile")
        self.assertEqual(excluded["src/frontend/public/logo.png"], "image-asset")
        self.assertEqual(excluded["dist/bundle.min.js"], "generated-output")
        # Unread data is not absent data, so it gets a reason a human can act on.
        self.assertEqual(excluded["sample-data/archive.xlsx"], "tabular-binary")

        routed: dict[str, str] = {}
        for batch in plan["batches"]:
            self.assertLessEqual(len(batch["files"]), 2)
            if batch["lens"] == "adversarial-attribution":
                continue
            for path in batch["files"]:
                self.assertNotIn(path, routed, "a file must be routed to exactly one lens")
                routed[path] = batch["lens"]

        self.assertEqual(routed["src/api/migrations/0001_init.sql"], "data-model")
        self.assertEqual(routed["src/api/models/order.py"], "data-model")
        self.assertEqual(routed["src/api/auth/roles.py"], "tenancy-authorization")
        self.assertEqual(routed["src/api/config.py"], "integration")
        self.assertEqual(routed["infra/main.bicep"], "integration")
        self.assertEqual(routed["src/api/routers/health.py"], "business-logic")
        self.assertEqual(routed["src/api/tasks/sync.py"], "business-logic")
        self.assertEqual(routed["docs/adr/0004-data-mapping.md"], "narrative-documentation")
        self.assertEqual(routed["README.md"], "narrative-documentation")
        # Prose claims a file only when no code lens already owns it.
        self.assertEqual(routed["docs/schemas/ledger.sql"], "data-model")
        self.assertEqual(routed["src/frontend/styles/theme.css"], "visual-identity")
        self.assertEqual(routed["src/frontend/assets/logo.svg"], "visual-identity")
        self.assertEqual(routed["tailwind.config.js"], "visual-identity")
        # A session token is not a design token, so auth keeps precedence over the name match.
        self.assertEqual(routed["src/api/auth/tokens.py"], "tenancy-authorization")
        self.assertEqual(routed["sample-data/orders.csv"], "embedded-data")
        # A corporate trust bundle couples the build to one network, which is an integration concern.
        self.assertEqual(routed["src/frontend/contoso.crt"], "integration")
        self.assertEqual(routed["Caddyfile"], "integration")
        self.assertEqual(routed["src/api/pyproject.toml"], "integration")
        self.assertEqual(routed["catalog-info.yaml"], "integration")
        self.assertEqual(routed["proxy/setup-proxy.ps1"], "integration")
        # Tracked results ship, so a committed output tree is reviewed rather than dismissed.
        self.assertNotIn("output/evaluations/run-01/example-service/result.md", excluded)
        self.assertEqual(
            routed["output/evaluations/run-01/example-service/result.md"], "embedded-data"
        )

        adversarial = [
            path
            for batch in plan["batches"]
            if batch["lens"] == "adversarial-attribution"
            for path in batch["files"]
        ]
        self.assertIn(".github/CODEOWNERS", adversarial, "a sample must never skip ownership files")
        self.assertIn("catalog-info.yaml", adversarial)

        # Every reviewable file is routed, so exclusions plus routed equal the manifest.
        self.assertEqual(len(routed) + len(excluded), len(manifest))
        self.assertEqual(plan["summary"]["reviewable"], len(routed))

    def test_readiness_classifies_committed_data_without_reading_values(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            data = root / "sample-data"
            data.mkdir()
            secret_value = "Synthetic Sensitive Value"
            (data / "records.csv").write_text(
                f"entity_name,tax_id,quantity,notes\nExample Organization,TAX-123,12,{secret_value}\n",
                encoding="utf-8",
            )
            (data / "lookup.csv").write_text("code,label\nEX,Example\n", encoding="utf-8")
            (data / "records.xlsx").write_bytes(b"PK\x03\x04binary")
            subprocess.run(["git", "-C", str(root), "add", "."], check=True)

            result = run_script(READINESS_SCRIPT, "--root", str(root), "--skip-history-scan")

        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        datasets = {item["path"]: item for item in report["committed_datasets"]}

        self.assertEqual(report["summary"]["committed_datasets"], 3)
        self.assertEqual(report["summary"]["datasets_with_identity_columns"], 1)
        self.assertEqual(report["summary"]["unreadable_datasets"], 1)

        self.assertEqual(
            datasets["sample-data/records.csv"]["identity_columns"],
            ["entity_name", "tax_id"],
        )
        self.assertEqual(datasets["sample-data/lookup.csv"]["identity_columns"], [])
        self.assertFalse(datasets["sample-data/records.xlsx"]["readable"])

        # Column names classify the file; row values must never enter the report.
        self.assertNotIn(secret_value, result.stdout)
        self.assertNotIn("Example Organization", result.stdout)

    def test_readiness_reports_what_breaks_when_planning_records_are_removed(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            for relative, body in {
                "docs/plans/2026-01-01-001-shipping-plan.md": "# Plan\n",
                ".github/skills/ce-compound/SKILL.md": "Index docs/plans/ and docs/analysis/ here.\n",
                ".github/agents/researcher.agent.md": "Read docs/plans/ for prior art.\n",
                "src/app.py": "print('no planning reference')\n",
            }.items():
                target = root / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text(body, encoding="utf-8")
            subprocess.run(["git", "-C", str(root), "add", "."], check=True)

            result = run_script(READINESS_SCRIPT, "--root", str(root), "--skip-history-scan")

        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        referrers = {item["path"]: item for item in report["files_referencing_history"]}

        self.assertEqual(report["summary"]["files_referencing_history"], 2)
        self.assertEqual(report["summary"]["references_to_history"], 3)

        # Shipped tooling reads the records, so removing them breaks the published workflow.
        self.assertEqual(
            referrers[".github/skills/ce-compound/SKILL.md"]["targets"],
            ["docs/analysis", "docs/plans"],
        )
        self.assertEqual(referrers[".github/agents/researcher.agent.md"]["references"], 1)

        # The records themselves are not referrers, and unrelated code is not swept in.
        self.assertNotIn("docs/plans/2026-01-01-001-shipping-plan.md", referrers)
        self.assertNotIn("src/app.py", referrers)

    def test_readiness_classifies_workflow_artifacts_separately(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            for relative in (
                "docs/debranding/2026-01-01-001-release-plan.md",
                "docs/plans/2026-01-01-001-shipping-plan.md",
                "README.md",
            ):
                target = root / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text("# Record\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(root), "add", "."], check=True)

            result = run_script(READINESS_SCRIPT, "--root", str(root), "--skip-history-scan")

        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)

        self.assertEqual(report["summary"]["workflow_artifacts"], 1)
        self.assertEqual(
            report["workflow_artifacts"], ["docs/debranding/2026-01-01-001-release-plan.md"]
        )

    def test_readiness_honors_a_custom_workflow_artifact_path(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            plan = root / "engineering" / "plans" / "release-plan.md"
            plan.parent.mkdir(parents=True)
            plan.write_text("# Plan\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(root), "add", "."], check=True)

            result = run_script(
                READINESS_SCRIPT,
                "--root",
                str(root),
                "--skip-history-scan",
                "--workflow-artifact-path",
                "engineering/plans",
            )

        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(report["workflow_artifact_path"], "engineering/plans")
        self.assertEqual(report["workflow_artifacts"], ["engineering/plans/release-plan.md"])
        self.assertIn("docs/plans", report["history_paths"])
        self.assertIn("engineering/plans", report["history_paths"])

    def test_exposure_map_shows_findings_and_unread_files_by_directory(self) -> None:
        lens = {
            "batch_id": "embedded-data-001",
            "lens": "embedded-data",
            "coverage": {
                "files_reviewed": ["sample-data/records.csv", "src/prompts/reviewer.yaml"]
            },
            "findings": [
                {
                    "id": "DATA-001",
                    "path": "sample-data/records.csv",
                    "line": 1,
                    "anchor": "entity_name",
                    "disposition": "[BLOCKS]",
                    "confidence": 100,
                },
                {
                    "id": "LOGIC-001",
                    "path": "src/prompts/reviewer.yaml",
                    "line": 6,
                    "anchor": "review panel",
                    "disposition": "[TUNE]",
                    "confidence": 50,
                },
            ],
        }
        scan_report = {
            "coverage": {
                "product_files": [
                    "sample-data/records.csv",
                    "src/prompts/reviewer.yaml",
                    "src/frontend/public/logo.svg",
                ]
            }
        }

        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "lens.json").write_text(json.dumps(lens), encoding="utf-8")
            (root / "scan-report.json").write_text(json.dumps(scan_report), encoding="utf-8")
            result = run_script(
                MERGE_SCRIPT,
                str(root / "lens.json"),
                "--report",
                str(root / "scan-report.json"),
                "--tree",
            )

        self.assertIn("exposure map", result.stdout)
        self.assertIn("sample-data/", result.stdout)
        self.assertIn("B1", result.stdout)
        self.assertIn("T1", result.stdout)
        # An unexamined directory must be distinguishable from a clean one.
        self.assertIn("largest unexamined directories", result.stdout)
        self.assertIn("src/frontend/public/", result.stdout)
        self.assertIn("unread=1", result.stdout)

    def test_batch_ids_are_numbered_within_each_lens(self) -> None:
        manifest = [f"src/routers/r{index:02d}.py" for index in range(5)]
        manifest += [f"src/models/m{index:02d}.py" for index in range(3)]
        scan_report = {"coverage": {"product_files": manifest}}

        with tempfile.TemporaryDirectory() as directory:
            report_file = Path(directory) / "scan-report.json"
            report_file.write_text(json.dumps(scan_report), encoding="utf-8")
            result = run_script(BATCH_SCRIPT, "--report", str(report_file), "--batch-size", "2")

        self.assertEqual(result.returncode, 0, result.stderr)
        plan = json.loads(result.stdout)

        by_lens: dict[str, list[str]] = {}
        for batch in plan["batches"]:
            by_lens.setdefault(batch["lens"], []).append(batch["id"])

        # Each lens counts from 001, so the first business-logic batch is not named -003.
        for lens, ids in by_lens.items():
            self.assertEqual(ids, [f"{lens}-{n:03d}" for n in range(1, len(ids) + 1)])
        self.assertEqual(len({item["id"] for item in plan["batches"]}), len(plan["batches"]))

    def test_merge_accepts_lens_output_wrapped_in_a_markdown_fence(self) -> None:
        payload = {
            "lens": "integration",
            "batch_id": "integration-001",
            "findings": [],
            "coverage": {"files_reviewed": ["src/a.py"], "not_reviewed": []},
        }
        wrapped = "Here is the review.\n\n```json\n" + json.dumps(payload) + "\n```\n"

        with tempfile.TemporaryDirectory() as directory:
            lens_file = Path(directory) / "integration-001.json"
            lens_file.write_text(wrapped, encoding="utf-8")
            result = run_script(MERGE_SCRIPT, str(lens_file))

        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(report["coverage"]["integration-001"]["files_reviewed"], ["src/a.py"])

    def test_shared_skill_is_internal(self) -> None:
        skill = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")

        self.assertIn("name: solution-debranding", skill)
        self.assertIn("user-invocable: false", skill)
        self.assertIn("disable-model-invocation: true", skill)
        self.assertFalse((SKILL_ROOT.parent / "solution-debranding-core").exists())

    def test_stage_skills_are_invocable_and_delegate_to_the_shared_package(self) -> None:
        for mode in ("plan", "apply", "verify"):
            stage_root = SKILL_ROOT.parent / f"solution-debranding-{mode}"
            with self.subTest(mode=mode):
                self.assertTrue(stage_root.is_dir())
                stage = (stage_root / "SKILL.md").read_text(encoding="utf-8")

                self.assertIn(f"name: solution-debranding-{mode}", stage)
                self.assertIn("argument-hint:", stage)
                self.assertNotIn("user-invocable: false", stage)
                self.assertNotIn("disable-model-invocation", stage)
                self.assertIn("../solution-debranding/SKILL.md", stage)
                self.assertIn(f"`MODE={mode}`", stage)
                self.assertNotIn("#file:", stage)
                self.assertNotIn("${input:", stage)

    def test_merge_accepts_visual_identity_lens(self) -> None:
        payload = {
            "lens": "visual-identity",
            "batch_id": "visual-identity-001",
            "findings": [],
            "coverage": {"files_reviewed": ["src/styles/theme.css"], "not_reviewed": []},
        }
        with tempfile.TemporaryDirectory() as directory:
            lens_file = Path(directory) / "visual-identity-001.json"
            lens_file.write_text(json.dumps(payload), encoding="utf-8")
            result = run_script(MERGE_SCRIPT, str(lens_file))

        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(report["problems"], [])
        self.assertIn("visual-identity", report["lenses_expected"])

    def test_scan_redacts_evidence_and_keeps_ids_stable(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            source = root / "src" / "settings.txt"
            source.parent.mkdir()
            source.write_text("Customer=Contoso\nDomain=contoso.example\n", encoding="utf-8")

            first = run_script(
                SCAN_SCRIPT,
                "--root",
                str(root),
                "--source-customer",
                "Contoso",
                "--domain",
                "contoso.example",
            )
            second = run_script(
                SCAN_SCRIPT,
                "--root",
                str(root),
                "--source-customer",
                "Contoso",
                "--domain",
                "contoso.example",
            )

            self.assertEqual(first.returncode, 0, first.stderr)
            self.assertEqual(second.returncode, 0, second.stderr)
            first_report = json.loads(first.stdout)
            second_report = json.loads(second.stdout)
            self.assertEqual(
                [finding["id"] for finding in first_report["findings"]],
                [finding["id"] for finding in second_report["findings"]],
            )
            evidence = "\n".join(finding["evidence"] for finding in first_report["findings"])
            self.assertNotIn("Contoso", evidence)
            self.assertNotIn("contoso.example", evidence)
            self.assertIn("<source-customer:redacted>", evidence)

    def test_scan_never_copies_source_line_values_into_evidence(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            sensitive_value = "FAKE_RELEASE_TEST_ONLY"
            (root / "config.txt").write_text(
                f"API_KEY={sensitive_value} customer=Contoso\n",
                encoding="utf-8",
            )

            result = run_script(
                SCAN_SCRIPT,
                "--root", str(root),
                "--source-customer", "Contoso",
                "--no-derive-terms",
            )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn(sensitive_value, result.stdout)
        evidence = "\n".join(
            finding["evidence"] for finding in json.loads(result.stdout)["findings"]
        )
        self.assertIn("Matched <source-customer:redacted>", evidence)

    def test_scan_matches_term_at_underscore_boundaries_but_not_inside_words(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "code.py").write_text(
                "valid_contoso_report = True\n"
                "CONTOSO_ID = 42\n"
                "x = get_contoso()\n"
                "name = 'Contoso'\n"
                "unrelated = 'Acontosob'\n"
                "also_unrelated = 'microsoftcontoso'\n",
                encoding="utf-8",
            )

            result = run_script(
                SCAN_SCRIPT,
                "--root", str(root),
                "--source-customer", "Contoso",
                "--no-derive-terms",
            )

        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        matched_lines = {f["line"] for f in report["findings"] if f["category"] == "source-customer"}
        self.assertIn(1, matched_lines, "should match _contoso_ at underscore boundaries")
        self.assertIn(2, matched_lines, "should match CONTOSO_ at underscore boundary")
        self.assertIn(3, matched_lines, "should match _contoso( at underscore boundary")
        self.assertIn(4, matched_lines, "should match standalone 'Contoso'")
        self.assertNotIn(5, matched_lines, "should not match contoso inside Acontosob")
        self.assertNotIn(6, matched_lines, "should not match contoso inside microsoftcontoso")

    def test_scan_can_fail_verification_for_customer_branding_in_env_example(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            environment_template = root / ".env.example"
            environment_template.write_text(
                "GATEWAY_URL=https://api.contoso.example\n", encoding="utf-8"
            )

            blocked = run_script(
                SCAN_SCRIPT,
                "--root",
                str(root),
                "--source-customer",
                "Contoso",
                "--fail-on-blocking",
            )

            self.assertEqual(blocked.returncode, 1, blocked.stderr)
            blocked_report = json.loads(blocked.stdout)
            self.assertIn(".env.example", blocked_report["coverage"]["product_files"])
            self.assertIn(".env.example", blocked_report["summary"]["suspicious_files"])
            customer_findings = [
                finding
                for finding in blocked_report["findings"]
                if finding["category"] == "source-customer"
            ]
            self.assertEqual(len(customer_findings), 1)
            self.assertEqual(customer_findings[0]["path"], ".env.example")
            self.assertEqual(customer_findings[0]["disposition"], "[BLOCKS]")

            environment_template.write_text(
                "GATEWAY_URL=https://api.example.com\n", encoding="utf-8"
            )
            clean = run_script(
                SCAN_SCRIPT,
                "--root",
                str(root),
                "--source-customer",
                "Contoso",
                "--fail-on-blocking",
            )

            self.assertEqual(clean.returncode, 0, clean.stderr)
            clean_report = json.loads(clean.stdout)
            self.assertEqual(clean_report["summary"]["blocking"], 0)
            self.assertIn(".env.example", clean_report["summary"]["suspicious_files"])

    def test_scan_blocks_uuid_literals_across_code_and_markdown(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            script = root / "settings.py"
            readme = root / "README.md"
            tenant_id = "-".join(("8f14e45f", "ea5e", "4b1c", "8a56", "0c9f9d5f6a71"))
            client_id = "-".join(("45c48cce", "2e2d", "4f6f", "a4a2", "2f7c2de44f10"))
            script.write_text(
                f'TENANT_ID = "{tenant_id}"\n'
                f'AUTHORITY = "https://login.microsoftonline.com/{tenant_id}"\n',
                encoding="utf-8",
            )
            readme.write_text(f"Use application {client_id} for local access.\n", encoding="utf-8")

            blocked = run_script(
                SCAN_SCRIPT,
                "--root",
                str(root),
                "--source-customer",
                "Contoso",
                "--fail-on-blocking",
            )

            self.assertEqual(blocked.returncode, 1, blocked.stderr)
            report = json.loads(blocked.stdout)
            opaque_findings = [
                finding
                for finding in report["findings"]
                if finding["category"] == "hardcoded-opaque-identifier"
            ]
            self.assertEqual(
                [(finding["path"], finding["line"]) for finding in opaque_findings],
                [("README.md", 1), ("settings.py", 1), ("settings.py", 2)],
            )
            self.assertTrue(
                all(finding["disposition"] == "[BLOCKS]" for finding in opaque_findings)
            )
            evidence = "\n".join(finding["evidence"] for finding in report["findings"])
            self.assertNotIn(tenant_id, evidence)
            self.assertNotIn(client_id, evidence)
            self.assertTrue(
                any(
                    finding["category"] == "hardcoded-cloud-identity"
                    for finding in report["findings"]
                )
            )

            script.write_text(
                'TENANT_ID = os.environ["ENTRA_TENANT_ID"]\n'
                'CLIENT_ID = os.environ["ENTRA_CLIENT_ID"]\n'
                'AUTHORITY = f"https://login.microsoftonline.com/{TENANT_ID}"\n'
                'FALLBACK_AUTHORITY = "https://login.microsoftonline.com/common"\n',
                encoding="utf-8",
            )
            readme.write_text("Configure the application ID through the environment.\n", encoding="utf-8")
            clean = run_script(
                SCAN_SCRIPT,
                "--root",
                str(root),
                "--source-customer",
                "Contoso",
                "--fail-on-blocking",
            )

            self.assertEqual(clean.returncode, 0, clean.stderr)
            self.assertEqual(json.loads(clean.stdout)["summary"]["blocking"], 0)

    def test_scan_matches_terms_at_underscore_boundaries_but_not_inside_words(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "app.py"
            source.write_text(
                "atlas_ground_truth = load()\n"
                "_ATLAS_GT = atlas_ground_truth\n"
                "x = Atlassian.connect()\n"
                "name = 'Project Atlas'\n",
                encoding="utf-8",
            )

            result = run_script(
                SCAN_SCRIPT,
                "--root", str(root),
                "--source-customer", "Contoso",
                "--alias", "Atlas",
                "--no-derive-terms",
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            report = json.loads(result.stdout)
            alias_findings = [
                f for f in report["findings"]
                if f["category"] == "alias" and f["path"] == "app.py"
            ]
            matched_lines = {f["line"] for f in alias_findings}
            self.assertIn(1, matched_lines, "should match atlas at underscore boundary")
            self.assertIn(2, matched_lines, "should match ATLAS at underscore boundary")
            self.assertNotIn(3, matched_lines, "should not match inside Atlassian")
            self.assertIn(4, matched_lines, "should match standalone Atlas")

    def test_scan_tags_planning_artifacts_as_history_surface(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            product = root / "src" / "settings.txt"
            product.parent.mkdir()
            product.write_text("Customer=Contoso\n", encoding="utf-8")
            plan = root / "docs" / "plans" / "2026-01-01-001-contoso-plan.md"
            plan.parent.mkdir(parents=True)
            plan.write_text("Decision reviewed with Contoso stakeholders.\n", encoding="utf-8")

            result = run_script(SCAN_SCRIPT, "--root", str(root), "--source-customer", "Contoso")

            self.assertEqual(result.returncode, 0, result.stderr)
            report = json.loads(result.stdout)
            surfaces = {finding["path"]: finding["surface"] for finding in report["findings"]}
            self.assertEqual(surfaces["src/settings.txt"], "product")
            self.assertEqual(surfaces["docs/plans/2026-01-01-001-contoso-plan.md"], "history")
            history = [finding for finding in report["findings"] if finding["surface"] == "history"]
            self.assertTrue(history)
            for finding in history:
                self.assertEqual(finding["remediation"], "human-review")
                self.assertEqual(finding["disposition"], "[GATED]")
            self.assertEqual(report["summary"]["by_surface"]["history"], len(history))

            # Each record needs its own decision, so the list ships alongside the count.
            self.assertEqual(
                report["coverage"]["history_files"],
                ["docs/plans/2026-01-01-001-contoso-plan.md"],
            )
            self.assertNotIn(
                "docs/plans/2026-01-01-001-contoso-plan.md", report["coverage"]["product_files"]
            )

    def test_scan_derives_the_repository_slug_without_being_told(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory) / "contoso-region-widget-service"
            root.mkdir()
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            subprocess.run(
                [
                    "git",
                    "-C",
                    str(root),
                    "remote",
                    "add",
                    "origin",
                    "https://github.com/Contoso-Engineering/contoso-region-widget-service.git",
                ],
                check=True,
            )
            note = root / "docs" / "pm" / "sprint-01-summary.md"
            note.parent.mkdir(parents=True)
            note.write_text(
                "Filed under Contoso-Engineering/contoso-region-widget-service; the widget-service team owns it.\n",
                encoding="utf-8",
            )

            result = run_script(SCAN_SCRIPT, "--root", str(root), "--source-customer", "Contoso")

            self.assertEqual(result.returncode, 0, result.stderr)
            report = json.loads(result.stdout)
            derived = report["scan"]["derived_terms"]
            # The abbreviation the team uses in prose is never in the supplied term list.
            self.assertIn("widget-service", derived)
            self.assertIn("contoso-region-widget-service", derived)
            self.assertIn("Contoso-Engineering", derived)
            self.assertIn(
                "Contoso-Engineering/contoso-region-widget-service", derived
            )
            # A bare trailing token would match unrelated prose everywhere.
            self.assertNotIn("service", derived)
            categories = {finding["category"] for finding in report["findings"]}
            self.assertIn("resource-prefix", categories)

            without = run_script(
                SCAN_SCRIPT,
                "--root",
                str(root),
                "--source-customer",
                "Contoso",
                "--no-derive-terms",
            )

            self.assertEqual(without.returncode, 0, without.stderr)
            self.assertEqual(json.loads(without.stdout)["scan"]["derived_terms"], [])

            rendered = run_script(
                SCAN_SCRIPT, "--root", str(root), "--source-customer", "Contoso", "--summary"
            )

            # Inferred terms are a claim, so they have to be visible without opening the JSON.
            self.assertEqual(rendered.returncode, 0, rendered.stderr)
            self.assertIn("derived terms", rendered.stdout)
            self.assertIn("widget-service", rendered.stdout)
            self.assertIn("confirm before trusting", rendered.stdout)

    def test_summary_renders_the_results_worth_hand_indexing_for(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            certificate = root / "src" / "corp.crt"
            certificate.parent.mkdir()
            certificate.write_text("-----BEGIN CERTIFICATE-----\n", encoding="utf-8")
            (root / "src" / "settings.txt").write_text("Owner=Contoso\n", encoding="utf-8")

            result = run_script(
                SCAN_SCRIPT, "--root", str(root), "--source-customer", "Contoso", "--summary"
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("suspicious files", result.stdout)
            self.assertIn("src/corp.crt", result.stdout)

            payload = json.loads(
                run_script(
                    SCAN_SCRIPT, "--root", str(root), "--source-customer", "Contoso"
                ).stdout
            )
            self.assertIn("suspicious_files", payload["summary"])
            self.assertNotIn("suspicious_files", payload["coverage"])

    def test_scan_tags_planning_records_named_outside_a_history_directory(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            for relative in (
                "docs/UAT_Bug_Dashboard.md",
                "docs/architecture.md",
                "src/dashboard_service.py",
            ):
                target = root / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text("Owner=Contoso\n", encoding="utf-8")

            result = run_script(SCAN_SCRIPT, "--root", str(root), "--source-customer", "Contoso")

            self.assertEqual(result.returncode, 0, result.stderr)
            report = json.loads(result.stdout)
            surfaces = {finding["path"]: finding["surface"] for finding in report["findings"]}
            self.assertEqual(surfaces["docs/UAT_Bug_Dashboard.md"], "history")
            self.assertEqual(surfaces["docs/architecture.md"], "product")
            # The name rule is scoped to docs/, so product code keeps its surface.
            self.assertEqual(surfaces["src/dashboard_service.py"], "product")

    def test_scan_tags_agent_tooling_as_tooling_surface_outside_the_denominator(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            for relative in (
                "src/settings.txt",
                ".github/workflows/ci.yml",
                ".claude/skills/debrand/SKILL.md",
                ".github/prompts/review.prompt.md",
            ):
                target = root / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text("Owner=Contoso\n", encoding="utf-8")

            result = run_script(SCAN_SCRIPT, "--root", str(root), "--source-customer", "Contoso")

            self.assertEqual(result.returncode, 0, result.stderr)
            report = json.loads(result.stdout)
            surfaces = {finding["path"]: finding["surface"] for finding in report["findings"]}
            self.assertEqual(surfaces["src/settings.txt"], "product")
            # CI is real infrastructure, so it stays in the reviewed product surface.
            self.assertEqual(surfaces[".github/workflows/ci.yml"], "product")
            self.assertEqual(surfaces[".claude/skills/debrand/SKILL.md"], "tooling")
            self.assertEqual(surfaces[".github/prompts/review.prompt.md"], "tooling")

            coverage = report["coverage"]
            self.assertEqual(coverage["files_by_surface"]["tooling"], 2)
            self.assertEqual(
                coverage["product_files"], [".github/workflows/ci.yml", "src/settings.txt"]
            )
            self.assertEqual(
                coverage["tooling_files"],
                [".claude/skills/debrand/SKILL.md", ".github/prompts/review.prompt.md"],
            )
            self.assertEqual(
                coverage["files_total"],
                sum(coverage["files_by_surface"].values()),
            )

            # Tooling ships on release, so it stays blocking for the branding pass.
            tooling = [finding for finding in report["findings"] if finding["surface"] == "tooling"]
            self.assertEqual(report["summary"]["by_surface"]["tooling"], len(tooling))
            for finding in tooling:
                self.assertEqual(finding["disposition"], "[BLOCKS]")

    def test_batch_planner_writes_one_dispatchable_file_per_batch(self) -> None:
        manifest = ["src/a.py", "src/b.py", "src/c.py"]
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            report_file = root / "scan-report.json"
            report_file.write_text(
                json.dumps({"coverage": {"product_files": manifest}}), encoding="utf-8"
            )
            split_dir = root / "batches"
            result = run_script(
                BATCH_SCRIPT,
                "--report",
                str(report_file),
                "--batch-size",
                "2",
                "--output",
                str(root / "plan.json"),
                "--split-dir",
                str(split_dir),
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            plan = json.loads((root / "plan.json").read_text(encoding="utf-8"))
            written = sorted(path.name for path in split_dir.glob("*.json"))

            self.assertEqual(written, sorted(f"{batch['id']}.json" for batch in plan["batches"]))
            for batch in plan["batches"]:
                self.assertEqual(
                    json.loads((split_dir / f"{batch['id']}.json").read_text(encoding="utf-8")),
                    batch,
                )

    def test_merge_summary_reports_the_rollup_instead_of_json(self) -> None:
        lens_payload = {
            "lens": "integration",
            "batch_id": "integration-001",
            "findings": [
                {
                    "path": "src/a.py",
                    "anchor": "L4",
                    "coupling_type": "integration",
                    "confidence": 75,
                    "evidence": ["hardcoded region"],
                }
            ],
            "coverage": {"files_reviewed": ["src/a.py", "src/b.py"], "not_reviewed": []},
        }
        plan = {"batches": [{"id": "integration-001", "lens": "integration", "files": ["src/a.py", "src/b.py"]}]}
        scan_report = {"coverage": {"product_files": ["src/a.py", "src/b.py"]}}

        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            lens_file = root / "integration-001.json"
            lens_file.write_text(json.dumps(lens_payload), encoding="utf-8")
            plan_file = root / "plan.json"
            plan_file.write_text(json.dumps(plan), encoding="utf-8")
            report_file = root / "scan-report.json"
            report_file.write_text(json.dumps(scan_report), encoding="utf-8")

            result = run_script(
                MERGE_SCRIPT,
                str(lens_file),
                "--report",
                str(report_file),
                "--batch-plan",
                str(plan_file),
                "--summary",
            )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn('"findings":', result.stdout)
        self.assertIn("batches         1/1 complete", result.stdout)
        self.assertIn("coverage        2/2 product files (100.0%)", result.stdout)
        self.assertIn("src/a.py:L4", result.stdout)

    def test_merge_blocks_view_groups_blocking_couplings(self) -> None:
        lens_payload = {
            "lens": "integration",
            "batch_id": "integration-001",
            "findings": [
                {
                    "path": "src/gateway.py",
                    "line": 12,
                    "anchor": "InternalGateway",
                    "coupling_type": "integration",
                    "confidence": 100,
                    "evidence": ["calls an internal-only gateway"],
                    "disposition": "[BLOCKS]",
                    "remediation": "parameterization",
                    "owner": "platform-architecture",
                    "why_it_matters": "A reusing customer has no route to this host",
                },
                {
                    "path": "docs/notes.md",
                    "line": 3,
                    "anchor": "RegionNote",
                    "coupling_type": "integration",
                    "confidence": 50,
                    "evidence": ["mentions one region"],
                    "disposition": "[TUNE]",
                    "remediation": "generalization",
                },
            ],
            "coverage": {"files_reviewed": ["src/gateway.py", "docs/notes.md"], "not_reviewed": []},
        }

        with tempfile.TemporaryDirectory() as directory:
            lens_file = Path(directory) / "integration-001.json"
            lens_file.write_text(json.dumps(lens_payload), encoding="utf-8")
            result = run_script(MERGE_SCRIPT, str(lens_file), "--blocks")

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn('"findings":', result.stdout)
        self.assertIn("blocking        1 (1 high confidence)", result.stdout)
        self.assertIn("by area         src=1", result.stdout)
        self.assertIn("by remediation  parameterization=1", result.stdout)
        self.assertIn("by owner        platform-architecture=1", result.stdout)
        self.assertIn("src/gateway.py:12  InternalGateway  [parameterization]", result.stdout)
        self.assertIn("A reusing customer has no route to this host", result.stdout)
        self.assertNotIn("RegionNote", result.stdout)

    def test_delta_scan_includes_changed_and_untracked_files(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            subprocess.run(["git", "-C", str(root), "config", "user.email", "test@example.invalid"], check=True)
            subprocess.run(["git", "-C", str(root), "config", "user.name", "Test User"], check=True)
            tracked = root / "tracked.txt"
            tracked.write_text("neutral\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(root), "add", "tracked.txt"], check=True)
            subprocess.run(["git", "-C", str(root), "commit", "-qm", "baseline"], check=True)
            tracked.write_text("Contoso\n", encoding="utf-8")
            (root / "untracked.txt").write_text("Contoso\n", encoding="utf-8")

            result = run_script(
                SCAN_SCRIPT,
                "--root",
                str(root),
                "--source-customer",
                "Contoso",
                "--mode",
                "delta",
                "--base",
                "HEAD",
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            paths = {finding["path"] for finding in json.loads(result.stdout)["findings"]}
            self.assertEqual(paths, {"tracked.txt", "untracked.txt"})

    def test_readiness_reports_indicators_without_reading_suspicious_values(self) -> None:
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            subprocess.run(["git", "-C", str(root), "config", "user.email", "test@example.invalid"], check=True)
            subprocess.run(["git", "-C", str(root), "config", "user.name", "Test User"], check=True)
            (root / "README.md").write_text("# Example\n", encoding="utf-8")
            (root / "private.key").write_text("not-a-real-secret-value\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(root), "add", "."], check=True)
            subprocess.run(["git", "-C", str(root), "commit", "-qm", "baseline"], check=True)

            with tempfile.TemporaryDirectory() as stub_directory:
                result = run_script(
                    READINESS_SCRIPT,
                    "--root",
                    str(root),
                    "--skip-history-scan",
                    env=path_with_stub_scanner(Path(stub_directory)),
                )
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertNotIn("not-a-real-secret-value", result.stdout)
            report = json.loads(result.stdout)
            self.assertTrue(report["indicators"]["README"])
            self.assertEqual(report["suspicious_tracked_paths"], ["private.key"])
            self.assertFalse(report["history_scanning"]["performed"])
            self.assertEqual(report["history_scanning"]["status"], "skipped-by-request")

            # A missing scanner outranks the skip request, so the gap is never read as a choice.
            unavailable = run_script(
                READINESS_SCRIPT,
                "--root",
                str(root),
                "--skip-history-scan",
                env=path_without_history_scanners(),
            )

            self.assertEqual(unavailable.returncode, 0, unavailable.stderr)
            status = json.loads(unavailable.stdout)["history_scanning"]["status"]
            self.assertEqual(status, "unavailable-no-scanner-installed")

    def test_readiness_scans_history_when_a_scanner_is_installed(self) -> None:
        if shutil.which("gitleaks") is None and shutil.which("trufflehog") is None:
            self.skipTest("no history scanner installed")
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            subprocess.run(["git", "-C", str(root), "config", "user.email", "test@example.invalid"], check=True)
            subprocess.run(["git", "-C", str(root), "config", "user.name", "Test User"], check=True)
            (root / "README.md").write_text("# Example\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(root), "add", "."], check=True)
            subprocess.run(["git", "-C", str(root), "commit", "-qm", "baseline"], check=True)

            result = run_script(READINESS_SCRIPT, "--root", str(root))

            self.assertEqual(result.returncode, 0, result.stderr)
            history = json.loads(result.stdout)["history_scanning"]
            self.assertTrue(history["performed"], history)
            self.assertEqual(history["status"], "clean")
            self.assertIsNotNone(history["tool_version"])

    def test_gitleaks_summary_rejects_missing_or_malformed_reports(self) -> None:
        spec = importlib.util.spec_from_file_location("check_readiness", READINESS_SCRIPT)
        self.assertIsNotNone(spec)
        self.assertIsNotNone(spec.loader)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)

        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            missing = root / "missing.json"
            malformed = root / "malformed.json"
            wrong_shape = root / "wrong-shape.json"
            malformed.write_text("not json", encoding="utf-8")
            wrong_shape.write_text("{}", encoding="utf-8")

            for report in (missing, malformed, wrong_shape):
                with self.subTest(report=report.name):
                    with self.assertRaises(ValueError):
                        module.summarise_gitleaks(report)

    def test_readiness_reports_failed_gitleaks_without_a_report_as_an_error(self) -> None:
        spec = importlib.util.spec_from_file_location("check_readiness", READINESS_SCRIPT)
        self.assertIsNotNone(spec)
        self.assertIsNotNone(spec.loader)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)

        failed_scan = subprocess.CompletedProcess(
            args=["gitleaks"], returncode=1, stdout="", stderr=""
        )
        with mock.patch.object(module.shutil, "which", side_effect=lambda tool: tool == "gitleaks"):
            with mock.patch.object(module.subprocess, "run", return_value=failed_scan):
                history = module.scan_history(Path.cwd(), is_git=True, skip=False)

        self.assertFalse(history["performed"])
        self.assertEqual(history["status"], "error")
        self.assertEqual(history["exit_code"], 1)

    def test_readiness_reports_history_findings_without_exposing_the_secret(self) -> None:
        if shutil.which("gitleaks") is None:
            self.skipTest("gitleaks is not installed")
        # A token-shaped fixture with real entropy, so the scan is exercised end to end.
        leaked = "ghp_" + "qR7vX2mNbT4wLpZ9cE1yH6jK8sD3fG5aB0uI"
        with tempfile.TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            subprocess.run(["git", "-C", str(root), "config", "user.email", "test@example.invalid"], check=True)
            subprocess.run(["git", "-C", str(root), "config", "user.name", "Test User"], check=True)
            (root / "deploy.sh").write_text(f"{leaked}\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(root), "add", "."], check=True)
            subprocess.run(["git", "-C", str(root), "commit", "-qm", "baseline"], check=True)

            result = run_script(READINESS_SCRIPT, "--root", str(root))

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertNotIn(leaked, result.stdout)
            history = json.loads(result.stdout)["history_scanning"]
            self.assertEqual(history["status"], "findings")
            self.assertGreater(history["finding_count"], 0)
            self.assertIn("github-pat", history["findings_by_rule"])
            self.assertTrue(any("deploy.sh" in location for location in history["finding_locations"]))


if __name__ == "__main__":
    unittest.main()