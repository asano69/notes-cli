#!/usr/bin/env bash
# migrate-notes.sh
find "$(notes config home)" -name "*.md" | while read f; do
	# すでに新フォーマットならスキップ
	head -1 "$f" | grep -q "^---$" && continue

	title=$(sed -n '1p' "$f")
	category=$(grep "^- Category:" "$f" | sed 's/- Category: //')
	tags=$(grep "^- Tags:" "$f" | sed 's/- Tags: //' | tr -d ' ')
	created=$(grep "^- Created:" "$f" | sed 's/- Created: //')
	body=$(awk '/^$/{found++} found>=2{print}' "$f")

	{
		echo "---"
		echo "category: $category"
		echo "tags: [$tags]"
		echo "created: ${created%+*}" # タイムゾーン除去
		echo "---"
		echo "# $title"
		echo ""
		echo "$body"
	} >"${f}.new"
	mv "${f}.new" "$f"
done
