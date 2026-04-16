package notes

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

// TagDelCmd represents `notes tag-del` command.
// Step 1: multi-select notes with fzf.
// Step 2: multi-select tags (from those notes) with a second fzf invocation.
// Step 3: remove the chosen tags from each selected note.
type TagDelCmd struct {
	cli      *kingpin.CmdClause
	Config   *Config
	Category string
	Tag      string
}

func (cmd *TagDelCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("tag-del", "Interactively select notes with fzf, then select tags to remove")
	cmd.cli.Flag("category", "Filter by category name with regular expression").Short('c').StringVar(&cmd.Category)
	cmd.cli.Flag("tag", "Filter by tag name with regular expression").Short('t').StringVar(&cmd.Tag)
}

func (cmd *TagDelCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// Do runs `notes tag-del` and returns an error if one occurs.
func (cmd *TagDelCmd) Do() error {
	notes, err := collectFilteredNotes(cmd.Config, cmd.Category, cmd.Tag)
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		return nil
	}

	// Step 1: pick notes.
	selectedRels, err := runFzf(cmd.Config, buildFzfInput(notes), fzfOptions{
		Multi:  true,
		Header: "TAB to multi-select notes | ENTER to choose tags to remove",
	})
	if err != nil || len(selectedRels) == 0 {
		return err
	}

	// Collect the union of tags from selected notes.
	absPaths := relPathsToAbsPaths(cmd.Config, selectedRels)
	tagSet := map[string]struct{}{}
	for _, abs := range absPaths {
		n, err := LoadNote(abs, cmd.Config)
		if err != nil {
			return errors.Wrapf(err, "cannot load %q", abs)
		}
		for _, t := range n.Tags {
			tagSet[t] = struct{}{}
		}
	}
	if len(tagSet) == 0 {
		fmt.Println("No tags to remove.")
		return nil
	}

	allTags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		allTags = append(allTags, t)
	}

	// Step 2: pick tags to remove.
	tagsToRemove, err := runFzfLines(cmd.Config, allTags, fzfOptions{
		Multi:  true,
		Prompt: "Tags to remove > ",
		Header: "TAB to multi-select | ENTER to confirm",
	})
	if err != nil || len(tagsToRemove) == 0 {
		return err
	}

	removeSet := make(map[string]struct{}, len(tagsToRemove))
	for _, t := range tagsToRemove {
		removeSet[t] = struct{}{}
	}

	// Step 3: apply.
	for _, abs := range absPaths {
		n, err := LoadNote(abs, cmd.Config)
		if err != nil {
			return errors.Wrapf(err, "cannot load %q", abs)
		}

		kept := make([]string, 0, len(n.Tags))
		for _, t := range n.Tags {
			if _, remove := removeSet[t]; !remove {
				kept = append(kept, t)
			}
		}

		rel := n.RelFilePath()
		if err := updateTagsInFile(abs, kept); err != nil {
			return errors.Wrapf(err, "cannot update tags in %q", rel)
		}
		fmt.Printf("Updated: %s  tags: [%s]\n", rel, strings.Join(kept, ", "))
	}
	return nil
}
