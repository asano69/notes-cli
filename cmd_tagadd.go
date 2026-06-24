package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// AddTagCmd represents `notes tagadd` command. Unlike `tag-add` (which opens
// an interactive fzf picker), this command adds a single tag directly,
// without any interaction. The second argument ("target") is either:
//   - a path to a single note file (directories are not allowed), or
//   - a category name (directory path relative to NOTES_CLI_HOME,
//     e.g. "test/aaa"), in which case the tag is added to every note
//     belonging to that category.
type AddTagCmd struct {
	Config *Config
	// Tag is the tag name to add
	Tag string
	// Target is a path to a note file, or a category name
	Target string
}

// Do runs `notes tagadd` command and returns an error if one occurs
func (cmd *AddTagCmd) Do() error {
	tag := strings.TrimSpace(cmd.Tag)
	if tag == "" {
		return errors.New("Tag name cannot be empty")
	}

	// A target that exists as a regular file is treated as a single note
	// path. Anything else (a directory, or a path that doesn't exist) is
	// treated as a category name instead.
	if info, err := os.Stat(cmd.Target); err == nil && !info.IsDir() {
		return cmd.addToFile(cmd.Target, tag)
	}
	return cmd.addToCategory(cmd.Target, tag)
}

// addToFile adds tag to the single note file at path.
func (cmd *AddTagCmd) addToFile(path, tag string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrapf(err, "cannot resolve path %q", path)
	}

	note, err := LoadNote(abs, cmd.Config)
	if err != nil {
		return err
	}

	merged := mergeTags(note.Tags, []string{tag})
	if err := updateTagsInFile(abs, merged); err != nil {
		return errors.Wrapf(err, "cannot update tags in %q", path)
	}
	fmt.Printf("Updated: %s  tags: [%s]\n", note.RelFilePath(), strings.Join(merged, ", "))
	return nil
}

// addToCategory adds tag to every note belonging to the given category.
func (cmd *AddTagCmd) addToCategory(category, tag string) error {
	cats, err := CollectCategories(cmd.Config, 0)
	if err != nil {
		return err
	}

	cat, ok := cats[category]
	if !ok {
		ns := cats.Names()
		return errors.Errorf("Category '%s' does not exist. All categories are %s", category, strings.Join(ns, ", "))
	}

	notes, err := cat.Notes(cmd.Config)
	if err != nil {
		return err
	}

	for _, note := range notes {
		merged := mergeTags(note.Tags, []string{tag})
		if err := updateTagsInFile(note.FilePath(), merged); err != nil {
			return errors.Wrapf(err, "cannot update tags in %q", note.RelFilePath())
		}
		fmt.Printf("Updated: %s  tags: [%s]\n", note.RelFilePath(), strings.Join(merged, ", "))
	}
	return nil
}

// mergeTags returns existing tags with newTags appended, deduplicating by name.
func mergeTags(existing, add []string) []string {
	seen := make(map[string]struct{}, len(existing))
	result := make([]string, 0, len(existing)+len(add))
	for _, t := range existing {
		seen[t] = struct{}{}
		result = append(result, t)
	}
	for _, t := range add {
		if _, ok := seen[t]; !ok {
			seen[t] = struct{}{}
			result = append(result, t)
		}
	}
	return result
}
