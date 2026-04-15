#!/usr/bin/env bash
NOTES_ROOT="$NOTES_CLI_HOME"
_list='notes ls -ro | awk "{print \$1 \"\t\" \$0}"'
_reload="reload(notes ls -ro | awk '{print \$1 \"\t\" \$0}')"
_preview="bat --color=always $NOTES_ROOT/{1}"
_fzf_opts='--delimiter="\t" --with-nth=2 --preview-window=up:60%'

cmd_open() {
	notes ls -ro |
		awk '{print $1 "\t" $0}' |
		fzf --delimiter="\t" --with-nth=2 \
			--preview="bat --color=always $NOTES_ROOT/{1}" \
			--preview-window=up:60% |
		awk -F'\t' '{print $1}' |
		xargs -I{} micro "$NOTES_ROOT/{}"
}

cmd_delete() {
	delete_notes() {
		local root="$1"
		shift
		read -r -p "Delete selected files? [y/N] " confirm </dev/tty
		[[ "$confirm" != [yY] ]] && return
		for rel in "$@"; do
			rm "$root/$rel"
		done
	}
	export -f delete_notes
	notes ls -ro |
		awk '{print $1 "\t" $0}' |
		fzf --delimiter="\t" --with-nth=2 \
			--multi \
			--preview="bat --color=always $NOTES_ROOT/{1}" \
			--preview-window=up:60% \
			--bind="enter:execute(bash -c 'delete_notes $NOTES_ROOT {+1}')+$_reload" \
			--bind="esc:abort"
}

cmd_rename() {
	rename_notes() {
		local root="$1"
		shift
		for rel in "$@"; do
			read -r -p "Rename $rel: " newname </dev/tty
			[[ -z "$newname" ]] && continue
			mv "$root/$rel" "$(dirname "$root/$rel")/${newname}.md"
		done
	}
	export -f rename_notes
	notes ls -ro |
		awk '{print $1 "\t" $0}' |
		fzf --delimiter="\t" --with-nth=2 \
			--multi \
			--preview="bat --color=always $NOTES_ROOT/{1}" \
			--preview-window=up:60% \
			--bind="enter:execute(bash -c 'rename_notes $NOTES_ROOT {+1}')+$_reload" \
			--bind="esc:abort"
}

cmd_move() {
	move_notes() {
		local root="$1"
		shift
		read -r -p "Move to folder: " folder </dev/tty
		[[ -z "$folder" ]] && return
		for rel in "$@"; do
			local file="$root/$rel"
			local base src_dir dest
			base=$(basename "$file")
			src_dir=$(dirname "$file")
			dest="$root/$folder"
			sed -i "s/^- Category: .*/- Category: $folder/" "$file"
			mkdir -p "$dest"
			mv "$file" "$dest/$base"
			if [[ -z "$(ls -A "$src_dir")" ]]; then
				cd ~ && rmdir "$src_dir"
			fi
		done
	}
	export -f move_notes
	notes ls -ro |
		awk '{print $1 "\t" $0}' |
		fzf --delimiter="\t" --with-nth=2 \
			--multi \
			--preview="bat --color=always $NOTES_ROOT/{1}" \
			--preview-window=up:60% \
			--bind="enter:execute(bash -c 'move_notes $NOTES_ROOT {+1}')+$_reload" \
			--bind="esc:abort"
}

cmd_tag_add() {
	add_tags() {
		local root="$1"
		local tmpfile="$2"
		stty sane </dev/tty 2>/dev/null
		read -r -p "追加するタグ (スペース区切り): " newtags </dev/tty
		[[ -z "$newtags" ]] && return
		while IFS=$'\t' read -r rel _rest; do
			local file="$root/$rel"
			[[ ! -f "$file" ]] && continue
			local current
			current=$(grep "^- Tags:" "$file" | sed 's/^- Tags: *//')
			local deduped
			deduped=$(printf '%s\n' $current $newtags |
				awk 'NF && !seen[$0]++' |
				tr '\n' ' ' |
				sed 's/ $//')
			if grep -q "^- Tags:" "$file"; then
				sed -i "s/^- Tags:.*/- Tags: $deduped/" "$file"
			else
				sed -i "/^- Created:/i - Tags: $deduped" "$file"
			fi
		done <"$tmpfile"
	}
	export -f add_tags
	notes ls -ro |
		awk '{print $1 "\t" $0}' |
		fzf --delimiter="\t" --with-nth=2 \
			--multi \
			--prompt="タグ追加 > " \
			--header="TAB で複数選択 | ENTER で確定" \
			--preview="bat --color=always $NOTES_ROOT/{1}" \
			--preview-window=up:60% \
			--bind="enter:execute(add_tags $NOTES_ROOT {+f})+$_reload" \
			--bind="esc:abort"
}

cmd_tag_del() {
	del_tags() {
		local root="$1"
		local tmpfile="$2"
		local all_tags
		all_tags=$(while IFS=$'\t' read -r rel _rest; do
			grep "^- Tags:" "$root/$rel" 2>/dev/null |
				sed 's/^- Tags: *//' |
				tr ' ' '\n'
		done <"$tmpfile" | sort -u | grep -v '^$')
		[[ -z "$all_tags" ]] && return
		stty sane </dev/tty 2>/dev/null
		local selected_tags
		selected_tags=$(printf '%s\n' "$all_tags" |
			fzf --multi \
				--prompt="削除するタグ > " \
				--header="TAB で複数選択 | ENTER で削除")
		[[ -z "$selected_tags" ]] && return
		while IFS=$'\t' read -r rel _rest; do
			local file="$root/$rel"
			[[ ! -f "$file" ]] && continue
			local current
			current=$(grep "^- Tags:" "$file" | sed 's/^- Tags: *//')
			local remaining
			remaining=$(printf '%s\n' $current |
				grep -vFxf <(printf '%s\n' "$selected_tags") |
				tr '\n' ' ' |
				sed 's/ $//')
			sed -i "s/^- Tags:.*/- Tags: $remaining/" "$file"
		done <"$tmpfile"
	}
	export -f del_tags
	notes ls -ro |
		awk '{print $1 "\t" $0}' |
		fzf --delimiter="\t" --with-nth=2 \
			--multi \
			--prompt="タグ削除対象ノート > " \
			--header="TAB で複数選択 | ENTER でタグ選択へ" \
			--preview="bat --color=always $NOTES_ROOT/{1}" \
			--preview-window=up:60% \
			--bind="enter:execute(del_tags $NOTES_ROOT {+f})+$_reload" \
			--bind="esc:abort"
}

usage() {
	echo "使い方: $(basename "$0") <command>"
	echo ""
	echo "Commands:"
	echo "  (e)dit     ノートを開く"
	echo "  (d)elete   ノートを削除する(複数選択可)"
	echo "  (r)ename   ノートのファイル名を変更する(複数選択可)"
	echo "  (m)ove     ノートのカテゴリを変更する(複数選択可)"
	echo "  (ta)tag-add  タグを追加する(複数選択可)"
	echo "  (td)tag-del  タグを削除する(複数選択可)"
}

case "${1:-}" in
e) cmd_open ;;
d) cmd_delete ;;
r) cmd_rename ;;
m) cmd_move ;;
ta) cmd_tag_add ;;
td) cmd_tag_del ;;
'') usage ;;
*) notes "$@" ;;
esac
