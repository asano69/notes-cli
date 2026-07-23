"""
Recursively scans markdown files for fenced code blocks (any language)
containing the context.WithTimeout + defer cancel() pattern with extra
leading indentation, and removes that indentation.

Usage:
    python fix_go_context_indent.py [root_dir] [--dry-run]

    root_dir : directory to scan (default: current directory)
    --dry-run: show what would change without modifying files
"""

import re
import sys
from pathlib import Path

# Matches any fenced code block, capturing any leading characters before each fence marker.
CODE_BLOCK_RE = re.compile(r"([^\n]*```[^\n]*\n)(.*?)([^\n]*```)", re.DOTALL)

# Lines that form the target pattern (after stripping leading whitespace)
PATTERN_LINES = (
    "ctx, cancel := context.WithTimeout(context.Background(),",
    "defer cancel()",
)


def has_indent(line: str) -> bool:
    return line != line.lstrip() and line.strip() != ""


def fix_block(block_body: str) -> tuple[str, bool]:
    """Remove common leading whitespace from the two target lines inside a block.

    Returns (fixed_body, changed).
    """
    lines = block_body.split("\n")
    changed = False

    i = 0
    while i < len(lines):
        stripped = lines[i].lstrip()
        # Detect the first line of the pattern
        if stripped.startswith(PATTERN_LINES[0]) and has_indent(lines[i]):
            # Check the next non-empty line for the second pattern line
            j = i + 1
            while j < len(lines) and lines[j].strip() == "":
                j += 1
            if j < len(lines) and lines[j].lstrip() == PATTERN_LINES[1] and has_indent(lines[j]):
                lines[i] = stripped
                lines[j] = lines[j].lstrip()
                changed = True
                i = j + 1
                continue
        i += 1

    return "\n".join(lines), changed


def fix_file(path: Path, dry_run: bool) -> bool:
    """Process one markdown file. Returns True if the file was (or would be) changed."""
    original = path.read_text(encoding="utf-8")
    result = []
    file_changed = False

    pos = 0
    for m in CODE_BLOCK_RE.finditer(original):
        result.append(original[pos : m.start()])
        open_fence, body, close_fence = m.group(1), m.group(2), m.group(3)
        fixed_body, changed = fix_block(body)
        # Strip everything before the ``` on each fence line
        fixed_open = re.sub(r"^[^\n`]*(`)", r"\1", open_fence)
        fixed_close = re.sub(r"^[^\n`]*(`)", r"\1", close_fence)
        if fixed_open != open_fence or fixed_close != close_fence:
            changed = True
        result.append(fixed_open + fixed_body + fixed_close)
        if changed:
            file_changed = True
        pos = m.end()
    result.append(original[pos:])

    if file_changed:
        print(f"{'[dry-run] would fix' if dry_run else 'fixed'}: {path}")
        if not dry_run:
            path.write_text("".join(result), encoding="utf-8")

    return file_changed


def main() -> None:
    args = sys.argv[1:]
    dry_run = "--dry-run" in args
    dirs = [a for a in args if not a.startswith("--")]
    root = Path(dirs[0]) if dirs else Path(".")

    md_files = list(root.rglob("*.md"))
    print(f"Scanning {len(md_files)} markdown file(s) under '{root}' ...\n")

    fixed_count = sum(fix_file(f, dry_run) for f in md_files)

    if fixed_count == 0:
        print("No files needed fixing.")
    else:
        print(f"\n{'Would fix' if dry_run else 'Fixed'} {fixed_count} file(s).")


if __name__ == "__main__":
    main()
