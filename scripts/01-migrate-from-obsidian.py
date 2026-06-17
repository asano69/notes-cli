#!/usr/bin/env python3
"""Migrate markdown frontmatter to a new schema.

Old schema:
    category: <str>
    tags: [] or block list
    created: <datetime>

New schema:
    title: <from H1 or filename>
    date: <from created>
    summary:
    categories: [<from category>]
    tags: [...]
    code:

Usage:
    convert_frontmatter.py [--dry-run] <directory> [<directory> ...]
"""

import re
import sys
from pathlib import Path


def split_frontmatter(content):
    """Return (frontmatter_str, body) or (None, content) if no frontmatter."""
    match = re.match(r'^---\n(.*?)\n---\n?(.*)', content, re.DOTALL)
    if not match:
        return None, content
    return match.group(1), match.group(2)


def parse_field(fm_str, key):
    """Parse a scalar or list field from frontmatter text without PyYAML.

    Handles three YAML styles used by Obsidian:
        key: value      -> 'value'
        key: [a, b]     -> ['a', 'b']
        key:            -> []
          - a
          - b
    """
    pattern = rf'^{re.escape(key)}:([ \t]*.*)$'
    match = re.search(pattern, fm_str, re.MULTILINE)
    if not match:
        return None

    rest = match.group(1).strip()

    # Inline list: key: [a, b, c]
    if rest.startswith('[') and rest.endswith(']'):
        inner = rest[1:-1].strip()
        if not inner:
            return []
        return [item.strip() for item in inner.split(',')]

    # Scalar value on the same line
    if rest:
        return rest

    # Block list: indented "- item" lines following the key
    after = fm_str[match.end():]
    items = []
    for line in after.splitlines():
        block_match = re.match(r'^[ \t]+-[ \t]+(.*)', line)
        if block_match:
            items.append(block_match.group(1).strip())
        elif line.strip():
            break  # non-empty, non-list line: block ended
    return items


def extract_h1(body):
    """Remove the first H1 heading and return (title, remaining_body)."""
    match = re.search(r'^# (.+)$', body, re.MULTILINE)
    if not match:
        return None, body
    title = match.group(1).strip()
    body = body[:match.start()] + body[match.end():]
    return title, body


def inline_list(items):
    """Render a Python list as an inline YAML sequence: [a, b]."""
    return '[' + ', '.join(str(i) for i in items) + ']'


def convert_file(path, dry_run=False):
    """Convert one markdown file in place."""
    content = path.read_text(encoding='utf-8')

    fm_str, body = split_frontmatter(content)
    if fm_str is None:
        print(f'[skip] no frontmatter: {path}')
        return

    # Title: prefer H1 heading, fall back to filename stem
    h1_title, body = extract_h1(body)
    title = h1_title if h1_title else path.stem

    # Date: scalar string, preserves the original timezone offset
    date = parse_field(fm_str, 'created') or ''

    # category (scalar or list) -> categories list
    raw_cat = parse_field(fm_str, 'category')
    if isinstance(raw_cat, list):
        categories = raw_cat
    elif raw_cat:
        categories = [raw_cat]
    else:
        categories = []

    # tags: handles both inline [] and block list styles
    raw_tags = parse_field(fm_str, 'tags') or []
    tags = raw_tags if isinstance(raw_tags, list) else [raw_tags]

    new_fm = '\n'.join([
        '---',
        f'title: {title}',
        f'date: {date}',
        'summary:',
        f'categories: {inline_list(categories)}',
        f'tags: {inline_list(tags)}',
        'code:',
        '---',
    ])

    body = body.lstrip('\n')
    new_content = new_fm + '\n' + ('\n' + body if body.strip() else '')

    if dry_run:
        print(f'[dry-run] {path}')
        print(new_content)
        print()
        return

    path.write_text(new_content, encoding='utf-8')
    print(f'[done] {path}')


def main():
    args = sys.argv[1:]
    dry_run = '--dry-run' in args
    dirs = [a for a in args if not a.startswith('--')]

    if not dirs:
        print(f'Usage: {sys.argv[0]} [--dry-run] <directory> [<directory> ...]')
        sys.exit(1)

    for dir_arg in dirs:
        root = Path(dir_arg)
        if not root.is_dir():
            print(f'Error: not a directory: {root}', file=sys.stderr)
            sys.exit(1)
        for md_file in sorted(root.rglob('*.md')):
            convert_file(md_file, dry_run=dry_run)


if __name__ == '__main__':
    main()
