package notes

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// TagModCmd represents `notes tagmod` command. Each public fields represent options of the command.
// It renames a tag across all notes when "To" is given, or deletes it when "To" is empty
// (deletion requires the --force flag as a safety guard).
type TagModCmd struct {
	Config *Config
	// From is the tag name to rename or delete
	From string
	// To is the new tag name. When empty, the "From" tag is deleted instead of renamed
	To string
	// Force is a flag equivalent to --force. It is required to delete a tag (when To is empty)
	Force bool
}

// renameOrDeleteTag returns a copy of tags with "from" renamed to "to" (or removed when to is
// empty), deduplicating in case the result already contains "to". The second return value
// reports whether anything actually changed.
func renameOrDeleteTag(tags []string, from, to string) ([]string, bool) {
	changed := false
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		if t == from {
			changed = true
			if to == "" {
				continue // delete
			}
			t = to
		}
		if _, ok := seen[t]; ok {
			continue // drop duplicate caused by the rename
		}
		seen[t] = struct{}{}
		result = append(result, t)
	}
	return result, changed
}

// Do runs `notes tagmod` command and returns an error if one occurs
func (cmd *TagModCmd) Do() error {
	if cmd.To == "" && !cmd.Force {
		return errors.Errorf("Deleting tag '%s' requires --force. Run 'notes tagmod %s --force' to delete it", cmd.From, cmd.From)
	}

	cats, err := CollectCategories(cmd.Config, 0)
	if err != nil {
		return err
	}

	for _, cat := range cats {
		for _, p := range cat.NotePaths {
			note, err := LoadNote(p, cmd.Config)
			if err != nil {
				return err
			}

			newTags, changed := renameOrDeleteTag(note.Tags, cmd.From, cmd.To)
			if !changed {
				continue
			}

			if err := updateTagsInFile(p, newTags); err != nil {
				return errors.Wrapf(err, "cannot update tags in %q", p)
			}
			fmt.Printf("Updated: %s  tags: [%s]\n", note.RelFilePath(), strings.Join(newTags, ", "))
		}
	}

	return nil
}
