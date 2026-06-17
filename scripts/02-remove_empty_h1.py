#!/usr/bin/env python3
"""Remove empty h1 lines (lines that are exactly '#' or '# ') from Markdown files."""

import sys
from pathlib import Path


def remove_empty_h1(path):
    with open(path, encoding="utf-8") as f:
        lines = f.readlines()

    filtered = [line for line in lines if line.rstrip("\n") not in ("#", "# ")]

    with open(path, "w", encoding="utf-8") as f:
        f.writelines(filtered)


def collect_md_files(arg):
    """Return all .md files for a given path (file or directory)."""
    p = Path(arg)
    if p.is_dir():
        return sorted(p.rglob("*.md"))
    return [p]


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <file.md|dir> [...]")
        sys.exit(1)

    for arg in sys.argv[1:]:
        for md_file in collect_md_files(arg):
            remove_empty_h1(md_file)
            print(f"Processed: {md_file}")
