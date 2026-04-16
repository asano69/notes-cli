package notes

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

// TagAddCmd represents `notes tag-add` command.
// It opens a multi-select fzf picker, prompts for tags to add, then updates
// the tags field (inline format) in each selected note's frontmatter.
type TagAddCmd struct {
	cli      *kingpin.CmdClause
	Config   *Config
	Category string
	Tag      string
}

func (cmd *TagAddCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("tag-add", "Interactively select notes with fzf and add tags to them")
	cmd.cli.Flag("category", "Filter by category name with regular expression").Short('c').StringVar(&cmd.Category)
	cmd.cli.Flag("tag", "Filter by tag name with regular expression").Short('t').StringVar(&cmd.Tag)
}

func (cmd *TagAddCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// Do runs `notes tag-add` and returns an error if one occurs.
func (cmd *TagAddCmd) Do() error {
	notes, err := collectFilteredNotes(cmd.Config, cmd.Category, cmd.Tag)
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		return nil
	}

	selected, err := runFzf(cmd.Config, buildFzfInput(notes), fzfOptions{
		Multi:  true,
		Header: "TAB to multi-select | ENTER to add tags",
	})
	if err != nil || len(selected) == 0 {
		return err
	}

	fmt.Fprint(os.Stderr, "Tags to add (space-separated): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		return nil
	}

	newTags := splitTags(input)
	if len(newTags) == 0 {
		return nil
	}

	for _, rel := range selected {
		abs := relPathsToAbsPaths(cmd.Config, []string{rel})[0]
		note, err := LoadNote(abs, cmd.Config)
		if err != nil {
			return errors.Wrapf(err, "cannot load %q", rel)
		}

		merged := mergeTags(note.Tags, newTags)
		if err := updateTagsInFile(abs, merged); err != nil {
			return errors.Wrapf(err, "cannot update tags in %q", rel)
		}
		fmt.Printf("Updated: %s  tags: [%s]\n", rel, strings.Join(merged, ", "))
	}
	return nil
}

// splitTags splits a space- or comma-separated tag string into individual tags.
func splitTags(s string) []string {
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
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
