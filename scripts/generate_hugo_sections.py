#!/usr/bin/env python3
"""Generate Hugo _index.md files for every non-hidden folder in a directory tree.

Usage:
    python generate_hugo_sections.py <root_folder>

For each folder found (including the root folder itself), an _index.md
is written (overwriting any existing one) with the folder name used as
both the title and the category.
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
categories: [{title}]
draft: false
date: {date}
lastmod: {date}
---
"""


def generate_index_files(root: Path, date: str) -> None:
    """Walk the directory tree and write _index.md into every non-hidden folder."""
    for dirpath, dirnames, _ in os.walk(root):
        # Prevent os.walk from descending into hidden folders.
        dirnames[:] = [d for d in dirnames if not d.startswith(".")]

        current = Path(dirpath)
        if current.name.startswith("."):
            continue

        index_file = current / "_index.md"
        index_file.write_text(TEMPLATE.format(title=current.name, date=date), encoding="utf-8")
        print(f"Wrote {index_file}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate Hugo _index.md files recursively.")
    parser.add_argument("folder", type=Path, help="Root folder to process")
    args = parser.parse_args()

    if not args.folder.is_dir():
        raise SystemExit(f"Error: {args.folder} is not a directory")

    # Use midnight JST for the date, matching the existing front matter convention.
    today = datetime.datetime.now(ZoneInfo("Asia/Tokyo")).strftime("%Y-%m-%dT00:00:00+09:00")
    generate_index_files(args.folder, today)


if __name__ == "__main__":
    main()
