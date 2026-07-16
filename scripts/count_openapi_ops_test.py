#!/usr/bin/env python3
"""Unit tests for count_openapi_ops (F03 US3)."""
from __future__ import annotations

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts" / "count_openapi_ops.py"
OPENAPI_JSON = ROOT / "docs" / "api" / "openapi.json"
COVERAGE_MD = ROOT / "docs" / "api" / "coverage.md"

# Import helper
sys.path.insert(0, str(ROOT / "scripts"))
import count_openapi_ops  # noqa: E402


class CountOpsTest(unittest.TestCase):
    def test_canonical_snapshot_has_51_ops(self):
        doc = json.loads(OPENAPI_JSON.read_text(encoding="utf-8"))
        self.assertEqual(count_openapi_ops.count_operations(doc), 51)

    def test_ignores_x_extension_keys(self):
        doc = {
            "paths": {
                "/demo": {
                    "get": {"operationId": "getDemo"},
                    "x-internal": True,
                    "x-foo": {"bar": 1},
                }
            }
        }
        self.assertEqual(count_openapi_ops.count_operations(doc), 1)

    def test_mismatch_expected_exits_nonzero(self):
        rc = subprocess.run(
            [sys.executable, str(SCRIPT), str(OPENAPI_JSON), "--expected", "0"],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertNotEqual(rc.returncode, 0)
        self.assertIn("operations=", rc.stdout)

    def test_match_expected_exits_zero(self):
        rc = subprocess.run(
            [sys.executable, str(SCRIPT), str(OPENAPI_JSON), "--expected", "51"],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(rc.returncode, 0)

    def test_missing_coverage_md_detected(self):
        with tempfile.TemporaryDirectory() as td:
            missing = Path(td) / "coverage.md"
            rc = subprocess.run(
                [
                    sys.executable,
                    str(SCRIPT),
                    str(OPENAPI_JSON),
                    "--expected",
                    "51",
                    "--require-coverage-md",
                    str(missing),
                ],
                capture_output=True,
                text=True,
                check=False,
            )
            self.assertNotEqual(rc.returncode, 0)
            self.assertIn("missing coverage matrix", rc.stderr)

    def test_present_coverage_md_ok(self):
        self.assertTrue(COVERAGE_MD.is_file())
        rc = subprocess.run(
            [
                sys.executable,
                str(SCRIPT),
                str(OPENAPI_JSON),
                "--expected",
                "51",
                "--require-coverage-md",
                str(COVERAGE_MD),
            ],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(rc.returncode, 0)


if __name__ == "__main__":
    unittest.main()
