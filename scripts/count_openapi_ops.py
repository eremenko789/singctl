#!/usr/bin/env python3
"""Count OpenAPI HTTP operations and compare to expected (F03)."""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


def count_operations(doc: dict) -> int:
    paths = doc.get("paths") or {}
    n = 0
    for methods in paths.values():
        if not isinstance(methods, dict):
            continue
        for method in methods:
            if not str(method).startswith("x-"):
                n += 1
    return n


def load_json(path: Path) -> dict:
    with path.open(encoding="utf-8") as f:
        return json.load(f)


def main(argv: list[str] | None = None) -> int:
    p = argparse.ArgumentParser(description="Count OpenAPI operations")
    p.add_argument("openapi_json", type=Path, help="Path to openapi.json")
    p.add_argument(
        "--expected",
        type=int,
        default=None,
        help="If set, exit 1 when count != expected",
    )
    p.add_argument(
        "--require-coverage-md",
        type=Path,
        default=None,
        help="If set, require this file to exist",
    )
    args = p.parse_args(argv)

    try:
        doc = load_json(args.openapi_json)
    except (OSError, json.JSONDecodeError) as e:
        print(f"error reading OpenAPI JSON: {e}", file=sys.stderr)
        return 1

    n = count_operations(doc)
    if args.expected is not None:
        print(f"operations={n} expected={args.expected}")
        if n != args.expected:
            return 1
    else:
        print(f"operations={n}")

    if args.require_coverage_md is not None:
        if not args.require_coverage_md.is_file():
            print(
                f"missing coverage matrix: {args.require_coverage_md}",
                file=sys.stderr,
            )
            return 1
        print(f"OK: coverage matrix present at {args.require_coverage_md}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
