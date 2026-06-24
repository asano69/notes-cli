package notes

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// frontmatterFieldKind controls how a frontmatter field's value is rendered
// as a single "key: value" line.
type frontmatterFieldKind int

const (
	// quotedField renders its value as a double-quoted string, e.g. `title: "foo"`.
	quotedField frontmatterFieldKind = iota
	// listField renders its value as an inline list, e.g. `tags: [a, b]`.
	listField
	// rawField renders its value with no quoting, e.g. `date: 2026-06-21T14:00:46+09:00`.
	// An empty value renders as a bare `key: ` with nothing after it.
	rawField
)

// frontmatterField is one key of the YAML frontmatter schema notes-cli reads
// and writes, paired with how its value should be rendered.
type frontmatterField struct {
	key  string
	kind frontmatterFieldKind
}

// frontmatterSchema is the single source of truth for the YAML frontmatter
// schema: the keys notes-cli manages, their order, and how each value is
// rendered. note.Create() (writing a brand-new note) and FixCmd (repairing
// an existing note) both build their output from this table, so adding,
// removing, re-ordering, or re-formatting a field only needs to happen here.
var frontmatterSchema = []frontmatterField{
	{"title", quotedField},
	{"summary", quotedField},
	{"tags", listField},
	{"categories", listField},
	{"draft", rawField},
	{"date", rawField},
	{"lastmod", rawField},
}

// formatFrontmatterField renders one "key: value" line according to kind.
// scalar is used for quotedField/rawField, list is used for listField.
func formatFrontmatterField(kind frontmatterFieldKind, key, scalar string, list []string) string {
	switch kind {
	case quotedField:
		return key + ": " + quote(scalar)
	case listField:
		return key + ": [" + strings.Join(list, ", ") + "]"
	default: // rawField
		return key + ": " + scalar
	}
}

// renderFrontmatter builds a full "---\n...\n---\n" YAML frontmatter block
// from field values, in frontmatterSchema order. scalars holds values for
// quotedField/rawField keys; lists holds values for listField keys.
// A listField whose key is absent from lists is omitted entirely; this lets
// callers suppress optional list fields (e.g. categories) without touching
// the schema.
func renderFrontmatter(scalars map[string]string, lists map[string][]string) string {
	var b strings.Builder
	b.WriteString("---\n")
	for _, f := range frontmatterSchema {
		if f.kind == listField {
			list, ok := lists[f.key]
			if !ok {
				continue
			}
			b.WriteString(formatFrontmatterField(f.kind, f.key, "", list))
			b.WriteByte('\n')
			continue
		}
		b.WriteString(formatFrontmatterField(f.kind, f.key, scalars[f.key], nil))
		b.WriteByte('\n')
	}
	b.WriteString("---\n")
	return b.String()
}



// updateTagsInFile rewrites the tags field in the YAML frontmatter of the given
// note file. Both inline (tags: [a, b]) and block (tags:\n  - a\n  - b) formats
// are accepted on read; the result is always written in inline format.
func updateTagsInFile(path string, newTags []string) error {
	lines, err := readFileLines(path)
	if err != nil {
		return err
	}

	start, end, err := frontmatterBounds(lines)
	if err != nil {
		return errors.Wrapf(err, "cannot locate frontmatter in %q", path)
	}

	result := make([]string, 0, len(lines))
	i := 0
	for i < len(lines) {
		line := lines[i]
		inFM := i > start && i < end

		if inFM && line == "tags:" {
			// Block format: consume the following "  - ..." lines and emit inline.
			result = append(result, inlineTagsLine(newTags))
			i++
			for i < end && strings.HasPrefix(strings.TrimSpace(lines[i]), "- ") {
				i++
			}
			continue
		}
		if inFM && (strings.HasPrefix(line, "tags: ") || line == "tags: []") {
			result = append(result, inlineTagsLine(newTags))
			i++
			continue
		}
		result = append(result, line)
		i++
	}

	return writeFileLines(path, result)
}

// updateCategoryInFile rewrites the category field in the YAML frontmatter.
func updateCategoryInFile(path, newCategory string) error {
	lines, err := readFileLines(path)
	if err != nil {
		return err
	}

	start, end, err := frontmatterBounds(lines)
	if err != nil {
		return errors.Wrapf(err, "cannot locate frontmatter in %q", path)
	}

	for i, line := range lines {
		if i > start && i < end && strings.HasPrefix(line, "category: ") {
			lines[i] = "category: " + newCategory
			return writeFileLines(path, lines)
		}
	}
	return fmt.Errorf("category field not found in frontmatter of %q", path)
}

// frontmatterBounds returns the line indices of the opening and closing "---"
// markers. start and end are the indices of the "---" lines themselves.
func frontmatterBounds(lines []string) (start, end int, err error) {
	start = -1
	for i, line := range lines {
		if line == "---" {
			if start == -1 {
				start = i
			} else {
				return start, i, nil
			}
		}
	}
	return 0, 0, errors.New("frontmatter not found")
}

func inlineTagsLine(tags []string) string {
	return "tags: [" + strings.Join(tags, ", ") + "]"
}

func readFileLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeFileLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return w.Flush()
}
