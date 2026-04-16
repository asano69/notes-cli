package notes

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

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
