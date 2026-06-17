#!/usr/bin/env python3
"""Validate YAML front matter schema in Markdown files recursively.

No third-party dependencies — uses stdlib only.
"""

import re
import sys
from pathlib import Path

# All keys that must be present in the front matter
REQUIRED_KEYS = {"title", "date", "summary", "categories", "tags", "code"}

# Expected date format: 2026-05-31T12:05:53+09:00
DATE_PATTERN = re.compile(r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}$")


def parse_front_matter(text):
    """Parse simple YAML front matter without third-party libraries.

    Supports the subset of YAML used in Hugo/Obsidian front matter:
      - key: scalar value
      - key: [item1, item2]   (inline list)
      - key:                  (null / empty value)

    Returns a dict, or None if no front matter is found.
    """
    if not text.startswith("---"):
        return None
    end = text.find("\n---", 3)
    if end == -1:
        return None

    data = {}
    for line in text[3:end].splitlines():
        line = line.rstrip()
        if not line or line.startswith("#"):
            continue
        if ":" not in line:
            continue

        key, _, raw = line.partition(":")
        key = key.strip()
        raw = raw.strip()

        if raw.startswith("[") and raw.endswith("]"):
            # Inline list: [item1, item2] or []
            inner = raw[1:-1].strip()
            data[key] = [item.strip().strip("\"'") for item in inner.split(",")] if inner else []
        elif raw == "":
            data[key] = None
        else:
            data[key] = raw

    return data


def validate(data):
    """Validate a front matter dict. Returns a list of error strings."""
    errors = []

    # Check all required keys are present
    for key in sorted(REQUIRED_KEYS - set(data.keys())):
        errors.append(f"missing key: '{key}'")

    # Validate date format
    date_val = data.get("date")
    if date_val is None:
        pass  # already reported as missing key if absent
    elif not isinstance(date_val, str) or not DATE_PATTERN.match(date_val):
        errors.append(f"'date': invalid format '{date_val}' (expected: 2026-05-31T12:05:53+09:00)")

    # categories and tags must be lists (empty list [] is fine)
    for key in ("categories", "tags"):
        val = data.get(key)
        if val is not None and not isinstance(val, list):
            errors.append(f"'{key}' must be a list, got '{type(val).__name__}'")

    return errors


def check_file(path):
    """Check one Markdown file. Returns a list of error strings."""
    try:
        text = path.read_text(encoding="utf-8")
    except OSError as e:
        return [f"cannot read file: {e}"]

    data = parse_front_matter(text)

    if data is None:
        return ["no front matter found"]

    return validate(data)


def main():
    root = Path(sys.argv[1]) if len(sys.argv) > 1 else Path(".")
    files = sorted(root.rglob("*.md"))

    if not files:
        print(f"No .md files found under '{root}'")
        return 0

    total_errors = 0
    bad_files = 0

    for path in files:
        errors = check_file(path)
        if errors:
            bad_files += 1
            total_errors += len(errors)
            print(f"\n{path}")
            for msg in errors:
                print(f"  - {msg}")

    if total_errors == 0:
        print(f"All {len(files)} file(s) passed.")
    else:
        print(
            f"\n{len(files)} file(s) checked: "
            f"{bad_files} with errors, {total_errors} total error(s)."
        )

    return 1 if total_errors else 0


if __name__ == "__main__":
    sys.exit(main())
