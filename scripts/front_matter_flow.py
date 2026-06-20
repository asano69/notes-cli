#!/usr/bin/env python3
# Wrap YAML front matter lists in inline flow style (square brackets)
"""Convert YAML front matter block-style lists to inline flow style.

Usage:
    python front_matter_flow.py file.md [file2.md ...]
    python front_matter_flow.py content/**/*.md
"""
import sys
import re
from io import StringIO
from pathlib import Path
from ruamel.yaml import YAML
from ruamel.yaml.comments import CommentedSeq

# Match a front matter block at the very start of the file
_FM_RE = re.compile(r'\A---\n(.*?)\n---\n', re.DOTALL)


def _to_flow(seq):
    """Wrap a list in a flow-style CommentedSeq."""
    cs = CommentedSeq(seq)
    cs.fa.set_flow_style()
    return cs


def _convert(data):
    """Recursively set flow style on all list values in place."""
    if isinstance(data, dict):
        for k in data:
            v = data[k]
            if isinstance(v, list):
                data[k] = _to_flow(v)
            elif isinstance(v, dict):
                _convert(v)


def _make_yaml():
    yaml = YAML()
    yaml.preserve_quotes = True
    # Keep timestamps as plain strings so dates are not reformatted
    yaml.constructor.yaml_constructors.pop('tag:yaml.org,2002:timestamp', None)
    # Represent None as empty (e.g. "summary: " not "summary: null")
    yaml.representer.add_representer(
        type(None),
        lambda d, _: d.represent_scalar('tag:yaml.org,2002:null', ''),
    )
    return yaml


def process_file(path: Path) -> bool:
    text = path.read_text(encoding='utf-8')
    m = _FM_RE.match(text)
    if not m:
        return False
    yaml = _make_yaml()
    try:
        data = yaml.load(m.group(1))
    except Exception as e:
        print(f'parse error (skipped): {path}: {e}', file=sys.stderr)
        return False
    if not isinstance(data, dict):
        return False
    _convert(data)
    buf = StringIO()
    yaml.dump(data, buf)
    path.write_text(
        '---\n' + buf.getvalue().rstrip('\n') + '\n---\n' + text[m.end():],
        encoding='utf-8',
    )
    return True


def iter_targets(arg: str):
    """Yield .md files from a path argument (file or directory), skipping hidden directories."""
    p = Path(arg)
    if p.is_dir():
        for path in sorted(p.rglob('*.md')):
            dirs = path.relative_to(p).parts[:-1]
            if any(part.startswith('.') for part in dirs):
                continue
            yield path
    elif p.is_file():
        yield p
    else:
        print(f'skip (not found): {p}', file=sys.stderr)


def main():
    if len(sys.argv) < 2:
        print('Usage: front_matter_flow.py <file.md|dir> [...]', file=sys.stderr)
        sys.exit(1)
    for arg in sys.argv[1:]:
        for p in iter_targets(arg):
            if process_file(p):
                print(f'converted: {p}')
            else:
                print(f'no front matter: {p}')


if __name__ == '__main__':
    main()
