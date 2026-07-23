import sys
from pathlib import Path


def remove_one_leading_tab(text: str) -> str:
    """Remove one leading tab from each line."""
    lines = text.splitlines(keepends=True)
    result = []
    for line in lines:
        if line.startswith("\t"):
            line = line[1:]
        result.append(line)
    return "".join(result)


def process_files(directory: str) -> None:
    """Recursively process all .md files in the given directory."""
    for path in Path(directory).rglob("*.md"):
        original = path.read_text(encoding="utf-8")
        updated = remove_one_leading_tab(original)
        if updated != original:
            path.write_text(updated, encoding="utf-8")
            print(f"Updated: {path}")


if __name__ == "__main__":
    target = sys.argv[1] if len(sys.argv) > 1 in sys.argv else "."
    process_files(target)
