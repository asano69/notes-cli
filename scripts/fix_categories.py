#!/usr/bin/env python3
"""Fix the `categories` field in markdown YAML frontmatter so it matches
the file's directory path relative to a root directory.

Example frontmatter:

    ---
    title: "JavaScriptの歴史"
    summary:
    tags: []
    categories: [wiki/wiki-Front-end/JavaScript]
    draft:
    date: 2025-10-03T15:59:22+09:00
    lastmod:
    ---

For a file at <root>/wiki/wiki-Front-end/JavaScript/foo.md, the expected
value is `[wiki/wiki-Front-end/JavaScript]`. Files directly under <root>
expect `[]`.

Only the `categories:` line is touched; every other line is left
byte-for-byte identical. Files whose categories line is not a simple
flow-style list (`categories: [...]`) are skipped and reported instead of
being guessed at.

Usage:
    python fix_categories.py /path/to/folder            # fix in place
    python fix_categories.py /path/to/folder --dry-run   # preview only
"""

import argparse
import os
import re
import sys
from pathlib import Path

# Frontmatter block: "---\n" ... "---\n", group(1) is the content between.
FRONTMATTER_RE = re.compile(r"\A---\r?\n(.*?\r?\n)---\r?\n", re.DOTALL)

# A single-line flow-style categories field, e.g. "categories: [a/b/c]".
CATEGORIES_RE = re.compile(r"^categories:[ \t]*\[(.*)\][ \t]*\r?$", re.MULTILINE)


def iter_markdown_files(folder: Path):
    """Yield *.md files under folder, never descending into hidden
    directories (names starting with ".", e.g. .git, .hugo, .obsidian).
    Using os.walk (rather than Path.rglob) also keeps directories and files
    distinct, so a directory that happens to be named "*.md" is never
    mistaken for a file.
    """
    for dirpath, dirnames, filenames in os.walk(folder):
        dirnames[:] = [d for d in dirnames if not d.startswith(".")]
        for filename in filenames:
            if filename.endswith(".md") and not filename.startswith("."):
                yield Path(dirpath) / filename


def expected_categories(md_path: Path, root: Path) -> str:
    """Compute the categories value derived from the file's location."""
    rel_dir = md_path.parent.resolve().relative_to(root.resolve())
    if rel_dir == Path("."):
        return "[]"
    return f"[{rel_dir.as_posix()}]"


def fix_file(md_path: Path, root: Path, dry_run: bool) -> str:
    """Fix a single file. Returns a short status string."""
    text = md_path.read_text(encoding="utf-8")

    fm_match = FRONTMATTER_RE.match(text)
    if not fm_match:
        return "no-frontmatter"

    frontmatter = fm_match.group(1)
    cat_match = CATEGORIES_RE.search(frontmatter)
    if not cat_match:
        return "unparseable-categories"

    try:
        expected = expected_categories(md_path, root)
    except ValueError:
        return "outside-root"

    current = f"[{cat_match.group(1)}]"
    if current == expected:
        return "ok"

    new_line = f"categories: {expected}"
    new_frontmatter = (
        frontmatter[: cat_match.start()] + new_line + frontmatter[cat_match.end():]
    )
    new_text = text[: fm_match.start(1)] + new_frontmatter + text[fm_match.end(1):]

    if not dry_run:
        # Write to a temp file first, then atomically replace the original.
        tmp_path = md_path.with_suffix(md_path.suffix + ".tmp")
        tmp_path.write_text(new_text, encoding="utf-8")
        os.replace(tmp_path, md_path)

    return f"fixed:{current}->{expected}"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("folder", type=Path, help="Folder to scan recursively")
    parser.add_argument(
        "--root",
        type=Path,
        default=None,
        help="Base directory categories are computed relative to (default: same as 'folder')",
    )
    parser.add_argument("--dry-run", action="store_true", help="Show changes without writing files")
    args = parser.parse_args()

    if not args.folder.is_dir():
        sys.exit(f"Not a directory: {args.folder}")

    root = args.root if args.root is not None else args.folder

    fixed = skipped = 0
    for md_path in sorted(iter_markdown_files(args.folder)):
        result = fix_file(md_path, root, args.dry_run)

        if result.startswith("fixed:"):
            old, new = result[len("fixed:"):].split("->")
            print(f"[FIXED] {md_path}: {old} -> {new}")
            fixed += 1
        elif result == "unparseable-categories":
            print(f"[SKIP ] {md_path}: categories field is not a simple '[...]' list", file=sys.stderr)
            skipped += 1
        elif result == "no-frontmatter":
            print(f"[SKIP ] {md_path}: no YAML frontmatter found", file=sys.stderr)
            skipped += 1
        elif result == "outside-root":
            print(f"[SKIP ] {md_path}: file is not under --root", file=sys.stderr)
            skipped += 1
        # "ok" -> already correct, nothing to print

    mode = "(dry run, no files written) " if args.dry_run else ""
    print(f"\nDone {mode}- fixed: {fixed}, skipped: {skipped}")


if __name__ == "__main__":
    main()
