#!/usr/bin/env python3
"""
logseq_to_obsidian.py
Converts a Logseq graph directory into an Obsidian vault.

Conversions performed:
  - pages/     -> <vault>/notes/
  - journals/  -> <vault>/journals/
  - assets/    -> <vault>/assets/  (copied as-is)
  - Logseq page properties (key:: value) -> YAML front matter
  - Block ID lines (id:: <uuid>) are removed
  - Logseq namespace filenames (a___b.md) -> subfolder (a/b.md)
  - Block embeds ((uuid)) are replaced with a plain comment
  - #+BEGIN_QUERY / #+END_QUERY blocks are wrapped in a code fence
  - Leading "- " bullet on every line is preserved (Obsidian handles bullets fine)
"""

import argparse
import re
import shutil
from pathlib import Path


# ---------------------------------------------------------------------------
# Regex patterns
# ---------------------------------------------------------------------------

# Matches a Logseq property line at the top of a file: "key:: value"
PROPERTY_LINE = re.compile(r"^([a-zA-Z][a-zA-Z0-9_-]*):: (.*)$")

# Matches a block-ID line that should be dropped entirely
BLOCK_ID_LINE = re.compile(r"^\s*id:: [0-9a-f-]{36}\s*$")

# Matches a block reference embed: ((uuid))
BLOCK_EMBED = re.compile(r"\(\([0-9a-f-]{36}\)\)")

# Matches the start/end of a Logseq query block
QUERY_BEGIN = re.compile(r"^\s*#\+BEGIN_QUERY\s*$", re.IGNORECASE)
QUERY_END   = re.compile(r"^\s*#\+END_QUERY\s*$",   re.IGNORECASE)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def namespace_to_path(stem: str) -> Path:
    """Convert a Logseq namespace filename stem (a___b___c) to a nested path (a/b/c)."""
    parts = stem.split("___")
    return Path(*parts) if len(parts) > 1 else Path(stem)


def extract_properties(lines: list[str]) -> tuple[dict[str, str], list[str]]:
    """
    Pull leading property lines out of the content.
    Returns (properties_dict, remaining_lines).
    Properties can appear on the first non-empty lines of the file.
    """
    props: dict[str, str] = {}
    i = 0
    # Skip a possible leading "- " bullet that Logseq puts before properties
    while i < len(lines):
        stripped = lines[i].strip().lstrip("- ").strip()
        m = PROPERTY_LINE.match(stripped)
        if m:
            props[m.group(1)] = m.group(2).strip()
            i += 1
        elif lines[i].strip() == "" and not props:
            # Allow blank lines before any property has been seen
            i += 1
        else:
            break
    return props, lines[i:]


def build_front_matter(props: dict[str, str]) -> list[str]:
    """Render a YAML front-matter block from a properties dict."""
    if not props:
        return []

    # Properties that hold comma-separated lists, mapped to their Obsidian key names.
    # Logseq uses "alias" (singular); Obsidian expects "aliases" (plural).
    LIST_PROPS = {"alias": "aliases", "tags": "tags"}

    lines = ["---"]
    for key, value in props.items():
        yaml_key = LIST_PROPS.get(key, key)

        if key in LIST_PROPS:
            # Split "Doggo, Woofer, Yapper" into a YAML sequence
            items = [v.strip() for v in value.split(",") if v.strip()]
            lines.append(f"{yaml_key}:")
            for item in items:
                lines.append(f"  - {item}")
        else:
            # Wrap values containing special YAML characters in quotes
            if any(c in value for c in (":", "#", "[", "]", "{", "}")):
                value = f'"{value}"'
            lines.append(f"{yaml_key}: {value}")

    lines.append("---")
    lines.append("")
    return lines


def convert_query_blocks(lines: list[str]) -> list[str]:
    """Wrap #+BEGIN_QUERY … #+END_QUERY in a markdown code fence."""
    result = []
    inside = False
    for line in lines:
        if QUERY_BEGIN.match(line):
            result.append("```logseq-query")
            inside = True
        elif QUERY_END.match(line):
            result.append("```")
            inside = False
        else:
            result.append(line)
    return result


def convert_block_embeds(line: str) -> str:
    """Replace ((uuid)) block references with a visible placeholder."""
    return BLOCK_EMBED.sub("*(block reference not resolved)*", line)


# Matches a top-level Logseq bullet: "- " with no leading whitespace
TOP_LEVEL_BULLET = re.compile(r"^- (.*)")

def strip_top_level_bullets(lines: list[str]) -> list[str]:
    """
    Remove the leading "- " from top-level (unindented) Logseq bullets.
    Nested bullets (indented with tabs or spaces) are kept as-is so that
    Obsidian still renders them as a proper nested list.
    """
    result = []
    for line in lines:
        m = TOP_LEVEL_BULLET.match(line)
        result.append(m.group(1) if m else line)
    return result


def convert_file_content(text: str) -> str:
    """Apply all in-file transformations and return the result."""
    lines = text.splitlines()

    # 1. Remove block-ID lines
    lines = [l for l in lines if not BLOCK_ID_LINE.match(l)]

    # 2. Extract leading properties -> YAML front matter
    props, lines = extract_properties(lines)
    front_matter = build_front_matter(props)

    # 3. Wrap Logseq query blocks
    lines = convert_query_blocks(lines)

    # 4. Strip first-level "- " bullets (nested bullets are preserved)
    lines = strip_top_level_bullets(lines)

    # 5. Replace block embeds inline
    lines = [convert_block_embeds(l) for l in lines]

    return "\n".join(front_matter + lines)


# ---------------------------------------------------------------------------
# Main conversion logic
# ---------------------------------------------------------------------------

def convert_graph(logseq_dir: Path, vault_dir: Path, dry_run: bool = False) -> None:
    logseq_dir = logseq_dir.resolve()
    vault_dir  = vault_dir.resolve()

    if not logseq_dir.is_dir():
        raise SystemExit(f"Error: Logseq directory not found: {logseq_dir}")

    mapping = {
        logseq_dir / "pages":   vault_dir / "notes",
        logseq_dir / "journals": vault_dir / "journals",
    }
    assets_src = logseq_dir / "assets"
    assets_dst = vault_dir / "assets"

    converted = 0
    copied    = 0
    skipped   = 0

    for src_root, dst_root in mapping.items():
        if not src_root.is_dir():
            print(f"  [skip] {src_root} not found")
            continue

        for src_file in src_root.rglob("*"):
            if not src_file.is_file():
                continue

            rel = src_file.relative_to(src_root)

            if src_file.suffix.lower() == ".md":
                # Resolve namespace folders for the stem
                nested = namespace_to_path(rel.stem)
                dst_file = dst_root / rel.parent / nested.parent / (nested.name + ".md")

                if not dry_run:
                    dst_file.parent.mkdir(parents=True, exist_ok=True)
                    raw  = src_file.read_text(encoding="utf-8")
                    out  = convert_file_content(raw)
                    dst_file.write_text(out, encoding="utf-8")

                print(f"  [md]   {src_file.relative_to(logseq_dir)}  ->  {dst_file.relative_to(vault_dir)}")
                converted += 1

            else:
                # Copy non-markdown files (org, etc.) unchanged
                dst_file = dst_root / rel
                if not dry_run:
                    dst_file.parent.mkdir(parents=True, exist_ok=True)
                    shutil.copy2(src_file, dst_file)
                print(f"  [copy] {src_file.relative_to(logseq_dir)}  ->  {dst_file.relative_to(vault_dir)}")
                copied += 1

    # Copy assets directory as-is
    if assets_src.is_dir():
        if not dry_run:
            shutil.copytree(assets_src, assets_dst, dirs_exist_ok=True)
        print(f"  [assets] {assets_src.relative_to(logseq_dir)}  ->  assets/")
        copied += 1
    else:
        skipped += 1

    print()
    print(f"Done.  converted={converted}  copied={copied}  skipped={skipped}")
    if dry_run:
        print("(dry-run: no files were written)")


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="Convert a Logseq graph directory to an Obsidian vault."
    )
    parser.add_argument("logseq_dir", help="Path to the Logseq graph root directory")
    parser.add_argument("vault_dir",  help="Path to the destination Obsidian vault")
    parser.add_argument(
        "--dry-run", action="store_true",
        help="Print what would be done without writing any files"
    )
    args = parser.parse_args()

    convert_graph(Path(args.logseq_dir), Path(args.vault_dir), dry_run=args.dry_run)


if __name__ == "__main__":
    main()
