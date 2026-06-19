#!/usr/bin/env python3
import os
import sys
import re


def _strip_quotes(s):
    # Removes a single layer of matching single/double quotes, e.g.
    # '"Tech"' -> 'Tech'. Leaves unquoted text untouched.
    s = s.strip()
    if len(s) >= 2 and s[0] == s[-1] and s[0] in ('"', "'"):
        return s[1:-1]
    return s


def _split_flow_list(inner):
    # Splits the inside of a "[a, b, c]" flow list on commas, without
    # breaking on commas that appear inside quoted items.
    items = re.findall(r'"[^"]*"|\'[^\']*\'|[^,]+', inner)
    return [_strip_quotes(item) for item in items if item.strip()]


def _parse_block_list(lines):
    # Parses "  - item" style lines into a plain list of strings.
    items = []
    for line in lines:
        stripped = line.strip()
        if stripped.startswith('-'):
            items.append(_strip_quotes(stripped[1:].strip()))
    return items


def extract_field_block(text, key):
    # Grab the key's line plus any indented continuation lines that
    # follow it (i.e. a block-style YAML list). This is what the old
    # single-line regex was missing, which is why multi-line
    # categories/tags lists used to break it.
    pattern = rf'^{re.escape(key)}:.*(?:\n[ \t]+\S.*)*'
    m = re.search(pattern, text, re.M)
    return m.group(0) if m else None


def parse_field(text, key):
    # Returns (value, found). `found` is True if the key exists at all
    # (even with an empty/null value); `value` is the value parsed as
    # a real Python object (str or list) instead of raw text, covering
    # flow style ([a, b]), block style (- a / - b), and plain scalars.
    chunk = extract_field_block(text, key)
    if chunk is None:
        return None, False

    lines = chunk.split('\n')
    m = re.match(rf'^{re.escape(key)}:\s*(.*)$', lines[0])
    if not m:
        return None, False
    inline_value = m.group(1).strip()

    list_lines = [l for l in lines[1:] if re.match(r'^\s*-(\s|$)', l)]
    if list_lines:
        return _parse_block_list(list_lines), True

    if not inline_value:
        return None, True

    if inline_value.startswith('[') and inline_value.endswith(']'):
        return _split_flow_list(inline_value[1:-1]), True

    return _strip_quotes(inline_value), True


def format_field(key, value):
    # Lists are re-emitted as an indented YAML block; everything else
    # is written as a plain "key: value" line, same as before.
    if isinstance(value, list):
        if not value:
            return f"{key}: []"
        lines = [f"{key}:"] + [f"  - {item}" for item in value]
        return "\n".join(lines)
    if value is None:
        return f"{key}: "
    return f"{key}: {value}"


def convert_front_matter(text):
    # title/date/summary are kept as plain single-line scalars, same
    # approach as the original script (they are not the source of the
    # reported bug).
    # [ \t]* (not \s*) on purpose: \s also matches newlines, which let an
    # empty value here swallow the next line's content entirely.
    title_m = re.search(r'^title:[ \t]*(.*)', text, re.M)
    date_m = re.search(r'^date:[ \t]*(.*)', text, re.M)
    summary_m = re.search(r'^summary:[ \t]*(.*)', text, re.M)

    # categories/tags are parsed as real YAML so flow-style ([a, b]),
    # block-style (- a / - b), and plain scalar forms all work.
    categories, categories_found = parse_field(text, 'categories')
    tags, tags_found = parse_field(text, 'tags')

    if not (title_m and date_m and summary_m and categories_found and tags_found):
        return None

    title = title_m.group(1)
    date = date_m.group(1)
    summary = summary_m.group(1)

    new_yaml = "---\n"
    new_yaml += f"title: {title}\n"
    new_yaml += f"summary: {summary}\n"
    new_yaml += format_field('tags', tags) + "\n"
    new_yaml += format_field('categories', categories) + "\n"
    new_yaml += "draft: \n"
    new_yaml += f"date: {date}\n"
    new_yaml += "lastmod: \n"
    new_yaml += "---"
    return new_yaml


def process_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    pattern = r'^---\n(.*?)\n---'
    match = re.search(pattern, content, re.DOTALL | re.M)

    if match:
        front_matter = match.group(1)
        new_front_matter = convert_front_matter(front_matter)

        if new_front_matter:
            new_content = content.replace(match.group(0), new_front_matter, 1)
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(new_content)
            print(f"[変換成功] {file_path}")
        else:
            print(f"[スキップ] 必要な項目が不足しています: {file_path}")
    else:
        print(f"[スキップ] フロントマターが見つかりません: {file_path}")


def main(target_dir):
    target_extensions = ('.md', '.markdown', '.html')

    for root, dirs, files in os.walk(target_dir):
        for file in files:
            if file.endswith(target_extensions):
                file_path = os.path.join(root, file)
                try:
                    process_file(file_path)
                except Exception as e:
                    print(f"[エラー] {file_path}: {e}")


if __name__ == "__main__":
    # sys.argv[1:] gets the list of command-line arguments
    args = sys.argv[1:]

    if not args:
        print("エラー: 対象のフォルダパスが指定されていません。")
        print("使い方: python convert_yaml.py [対象フォルダのパス]")
        print("（例: python convert_yaml.py ./content/posts ）")
        sys.exit(1)

    target_directory = args[0]

    # Check whether the given path actually exists
    if not os.path.exists(target_directory):
        print(f"エラー: 指定されたパスが見つかりません: {target_directory}")
        sys.exit(1)

    print(f"「{target_directory}」内のファイルを再帰的に処理します...")
    main(target_directory)
    print("処理が完了しました。")
