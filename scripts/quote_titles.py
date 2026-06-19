#!/usr/bin/env python3
"""Quote unquoted title values in YAML front matter.

Works at the text level so it handles files with otherwise invalid YAML.

Usage:
    python quote_titles.py <file.md|dir> [...]
"""

import re
import sys
from pathlib import Path

_FM_RE = re.compile(r'\A---\n(.*?)\n---\n', re.DOTALL)
# Match a title line that has a non-empty value
_TITLE_RE = re.compile(r'^(title:[ \t]*)(\S[^\n]*)$', re.MULTILINE)


def _quote_if_needed(m: re.Match) -> str:
    prefix = m.group(1)
    value = m.group(2).rstrip()
    # Leave alone if already wrapped in matching " " or ' '
    if len(value) >= 2 and (
        (value[0] == '"' and value[-1] == '"')
        or (value[0] == "'" and value[-1] == "'")
    ):
        return m.group(0)
    # Escape any existing double quotes inside the value, then wrap
    escaped = value.replace('\\', '\\\\').replace('"', '\\"')
    return f'{prefix}"{escaped}"'


def process_file(path: Path) -> bool:
    text = path.read_text(encoding='utf-8')
    m = _FM_RE.match(text)
    if not m:
        return False

    fm_orig = m.group(1)
    fm_new = _TITLE_RE.sub(_quote_if_needed, fm_orig)

    if fm_new == fm_orig:
        return False  # nothing changed

    path.write_text('---\n' + fm_new + '\n---\n' + text[m.end():], encoding='utf-8')
    return True


def iter_targets(arg: str):
    """Yield .md files from a path argument (file or directory)."""
    p = Path(arg)
    if p.is_dir():
        yield from sorted(p.rglob('*.md'))
    elif p.is_file():
        yield p
    else:
        print(f'skip (not found): {p}', file=sys.stderr)


def main():
    if len(sys.argv) < 2:
        print('Usage: quote_titles.py <file.md|dir> [...]', file=sys.stderr)
        sys.exit(1)

    for arg in sys.argv[1:]:
        for p in iter_targets(arg):
            if process_file(p):
                print(f'quoted: {p}')


if __name__ == '__main__':
    main()
