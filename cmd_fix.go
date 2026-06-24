package notes

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// FixCmd represents `notes fix` command.
//
// It recursively walks the notes home directory, skipping hidden directories
// and files, and repairs each note's YAML frontmatter into the format
// notes-cli expects:
//
//	---
//	title: ""
//	summary: ""
//	tags: []
//	draft:
//	date:
//	lastmod:
//	---
//
// Missing keys are added with their default value. An unquoted 'title' or
// 'summary' is wrapped in double quotes. Block-style 'tags' lists are
// converted to the inline '[...]' style. An empty 'title' is derived from
// the file name, and an empty 'date' from the current time.
// 'categories' and the legacy 'category' field are removed: category is
// always derivable from the file's location relative to NOTES_CLI_HOME, so
// storing it in frontmatter is redundant.
// 'draft' and 'lastmod' are never modified, only added when missing. A note
// whose frontmatter cannot be repaired automatically is reported instead of
// being changed.
type FixCmd struct {
	Config *Config
	// DryRun prints what would be changed without modifying files
	DryRun bool
	// Out is a writer to write output of this command
	Out io.Writer
}

// Do runs `notes fix` command and returns an error if any occurs.
func (cmd *FixCmd) Do() error {
	paths, err := collectMarkdownFiles(cmd.Config.HomePath)
	if err != nil {
		return err
	}

	fixed, failed := 0, 0
	for _, p := range paths {
		changed, err := fixNoteFile(p, cmd.Config.HomePath, cmd.DryRun)
		if err != nil {
			fmt.Fprintf(cmd.Out, "%s\n  error: %s\n", p, err)
			failed++
			continue
		}
		if !changed {
			continue
		}
		if cmd.DryRun {
			fmt.Fprintf(cmd.Out, "%s\n  would be fixed\n", p)
		} else {
			fmt.Fprintf(cmd.Out, "%s\n  fixed\n", p)
		}
		fixed++
	}

	if cmd.DryRun {
		fmt.Fprintf(cmd.Out, "\n%d note(s) would be fixed, %d note(s) could not be fixed (dry-run: no files were modified).\n", fixed, failed)
	} else {
		fmt.Fprintf(cmd.Out, "\n%d note(s) fixed, %d note(s) could not be fixed.\n", fixed, failed)
	}

	return nil
}

// collectMarkdownFiles returns the paths of every ".md" file under root,
// skipping any directory or file whose name starts with "." (hidden).
func collectMarkdownFiles(root string) ([]string, error) {
	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		if info.IsDir() {
			if path != root && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".md") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "Cannot walk notes home directory")
	}
	return paths, nil
}

// fixNoteFile repairs the YAML frontmatter of the note at path, if needed.
// It reports whether the file was (or, in dry-run mode, would be) changed.
// When the frontmatter is broken in a way this command cannot repair
// automatically, an error is returned and the file is left untouched.
func fixNoteFile(path, home string, dryRun bool) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, errors.Wrap(err, "cannot read file")
	}

	lines := strings.Split(string(content), "\n")
	if lines[0] != "---" {
		return false, errors.New("file does not start with a '---' frontmatter delimiter")
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return false, errors.New("frontmatter is not closed with a '---' delimiter")
	}

	fixed, err := fixedFrontmatter(parseFrontmatterEntries(lines[1:end]), path, home)
	if err != nil {
		return false, err
	}

	newLines := append([]string{"---"}, fixed...)
	newLines = append(newLines, lines[end:]...)
	newContent := strings.Join(newLines, "\n")

	if newContent == string(content) {
		return false, nil
	}
	if dryRun {
		return true, nil
	}

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return false, errors.Wrap(err, "cannot write file")
	}
	return true, nil
}

// frontmatterEntry is one top-level key of YAML frontmatter, together with
// all of its raw lines: the "key: ..." line itself, plus any indented
// continuation lines that follow it (block-style list items, or further
// lines of a folded plain scalar).
type frontmatterEntry struct {
	key   string
	lines []string
}

// parseFrontmatterEntries splits frontmatter lines (the lines between the
// opening and closing '---' delimiters) into top-level entries. A new entry
// starts at every line of the form "key: ..." with no leading whitespace;
// indented lines belong to the entry above them.
func parseFrontmatterEntries(lines []string) []frontmatterEntry {
	var entries []frontmatterEntry
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if idx := strings.IndexByte(line, ':'); idx > 0 {
				entries = append(entries, frontmatterEntry{key: line[:idx], lines: []string{line}})
				continue
			}
		}
		if len(entries) > 0 {
			last := &entries[len(entries)-1]
			last.lines = append(last.lines, line)
		}
	}
	return entries
}

// scalarValue returns the value of a scalar entry (title, summary, date, ...)
// as a single line. Any folded continuation lines are joined onto the first
// line with a space, mirroring how YAML folds plain multi-line scalars.
func (e frontmatterEntry) scalarValue() string {
	value := strings.TrimSpace(e.lines[0][len(e.key)+1:])
	for _, l := range e.lines[1:] {
		value += " " + strings.TrimSpace(l)
	}
	return strings.TrimSpace(value)
}

// listItems returns the elements of a tags/categories entry, accepting both
// the inline style ("key: [a, b]") and the block style ("key:\n  - a\n  - b").
// Quotes around individual elements are stripped and empty elements are
// dropped. An entry with no value at all yields a nil slice.
func (e frontmatterEntry) listItems() ([]string, error) {
	head := strings.TrimSpace(e.lines[0][len(e.key)+1:])

	var raw []string
	switch {
	case len(e.lines) > 1:
		if head != "" {
			return nil, errors.Errorf("%q has both an inline value and block-style items", e.key)
		}
		for _, l := range e.lines[1:] {
			t := strings.TrimSpace(l)
			if t != "-" && !strings.HasPrefix(t, "- ") {
				return nil, errors.Errorf("cannot parse list item %q", l)
			}
			raw = append(raw, strings.TrimSpace(strings.TrimPrefix(t, "-")))
		}
	case head != "":
		if !strings.HasPrefix(head, "[") || !strings.HasSuffix(head, "]") {
			return nil, errors.Errorf("expected an inline or block style list for %q, got %q", e.key, head)
		}
		if inner := strings.TrimSpace(head[1 : len(head)-1]); inner != "" {
			raw = strings.Split(inner, ",")
		}
	}

	items := make([]string, 0, len(raw))
	for _, r := range raw {
		if v := unquote(strings.TrimSpace(r)); v != "" {
			items = append(items, v)
		}
	}
	return items, nil
}

// fixedFrontmatter builds the repaired frontmatter lines (without the
// surrounding '---' delimiters) for a note, given its parsed entries. It
// returns an error if some entry cannot be repaired automatically.
func fixedFrontmatter(entries []frontmatterEntry, path, home string) ([]string, error) {
	byKey := make(map[string]frontmatterEntry, len(entries))
	for _, e := range entries {
		byKey[e.key] = e
	}

	lines := make([]string, 0, len(frontmatterSchema)+len(entries))

	// title: quoted string, derived from the file name when empty
	title := ""
	if e, ok := byKey["title"]; ok {
		title = unquote(e.scalarValue())
	}
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	lines = append(lines, formatFrontmatterField(quotedField, "title", title, nil))

	// summary: quoted string, left empty when unset
	summary := ""
	if e, ok := byKey["summary"]; ok {
		summary = unquote(e.scalarValue())
	}
	lines = append(lines, formatFrontmatterField(quotedField, "summary", summary, nil))

	// tags: inline list, no derivation when empty
	tags, err := listItemsOf(byKey, "tags")
	if err != nil {
		return nil, err
	}
	lines = append(lines, formatFrontmatterField(listField, "tags", "", tags))

	// categories and the legacy category field are intentionally omitted.
	// Category is always derivable from the file's path relative to
	// NOTES_CLI_HOME, so storing it in frontmatter is redundant.

	// draft: never touched, only defaulted when missing
	lines = append(lines, untouchedOrDefault(byKey, "draft")...)

	// date: defaulted to now when empty, otherwise left untouched
	date := ""
	if e, ok := byKey["date"]; ok {
		date = e.scalarValue()
	}
	if date == "" {
		date = time.Now().Format(time.RFC3339)
	}
	lines = append(lines, formatFrontmatterField(rawField, "date", date, nil))

	// lastmod: never touched, only defaulted when missing
	lines = append(lines, untouchedOrDefault(byKey, "lastmod")...)

	// Any key outside the known schema is preserved verbatim, in its
	// original order, after the known keys above.
	// "category" (legacy singular) is also suppressed here: it is not in
	// frontmatterSchema (so isFrontmatterKey returns false for it), but
	// like "categories" it is now derived from the file path and should
	// be removed from existing notes.
	for _, e := range entries {
		if isFrontmatterKey(e.key) || e.key == "category" {
			continue
		}
		lines = append(lines, e.lines...)
	}

	return lines, nil
}

// listItemsOf returns the list items of key in byKey, or nil if key is absent.
func listItemsOf(byKey map[string]frontmatterEntry, key string) ([]string, error) {
	e, ok := byKey[key]
	if !ok {
		return nil, nil
	}
	items, err := e.listItems()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse %q", key)
	}
	return items, nil
}

// untouchedOrDefault returns the original lines of key in byKey unchanged,
// or a single default empty line ("key: ") when key is absent.
func untouchedOrDefault(byKey map[string]frontmatterEntry, key string) []string {
	if e, ok := byKey[key]; ok {
		return e.lines
	}
	return []string{key + ": "}
}

// isFrontmatterKey reports whether key is part of the known frontmatter schema.
func isFrontmatterKey(key string) bool {
	for _, f := range frontmatterSchema {
		if f.key == key {
			return true
		}
}
	return false
}

// quote wraps s in double quotes, escaping any backslash or double quote it
// already contains.
func quote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// unquote strips a matching pair of surrounding double or single quotes
// from s, if present. Unquoted input is returned unchanged.
func unquote(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
		if s[0] == '\'' && s[len(s)-1] == '\'' {
			return s[1 : len(s)-1]
		}
	}
	return s
}
