#!/usr/bin/env python3
"""Generate Hugo _index.md files for every non-hidden folder in a directory tree.

Usage:
    python generate_hugo_sections.py <root_folder>

For each folder found (including the root folder itself), an _index.md
is written (overwriting any existing one). The title is the folder's
own name; the category is its path relative to the root folder, with
components joined by '/' (e.g. "Design/test").
"""

import argparse
import os
from pathlib import Path
from zoneinfo import ZoneInfo
import datetime

TEMPLATE = """---
title: "{title}"
summary:
tags: []
categories: [{category}]
draft: false
date: {date}
lastmod: {date}
---
"""


def category_for(current: Path, root: Path) -> str:
    """Return the category path: folder names from root down to current, joined by '/'."""
    rel = current.relative_to(root)
    if rel == Path("."):
        # The root folder itself has no parent within the tree, so it acts as its own category.
        return current.name
    return rel.as_posix()


def generate_index_files(root: Path, date: str) -> None:
    """Walk the directory tree and write _index.md into every non-hidden folder."""
    for dirpath, dirnames, _ in os.walk(root):
        # Prevent os.walk from descending into hidden folders.
        dirnames[:] = [d for d in dirnames if not d.startswith(".")]

        current = Path(dirpath)
        if current.name.startswith("."):
            continue

        index_file = current / "_index.md"
        content = TEMPLATE.format(title=current.name, category=category_for(current, root), date=date)
        index_file.write_text(content, encoding="utf-8")
        print(f"Wrote {index_file}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate Hugo _index.md files recursively.")
    parser.add_argument("folder", type=Path, help="Root folder to process")
    args = parser.parse_args()

    if not args.folder.is_dir():
        raise SystemExit(f"Error: {args.folder} is not a directory")

    # Use the current JST time, accurate to the second, matching the front matter convention.
    now = datetime.datetime.now(ZoneInfo("Asia/Tokyo")).isoformat(timespec="seconds")
    generate_index_files(args.folder, now)


if __name__ == "__main__":
    main()
