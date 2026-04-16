package notes

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

// EditCmd represents `notes edit` command.
// It opens an interactive fzf picker and opens the selected note in the editor.
type EditCmd struct {
	cli      *kingpin.CmdClause
	Config   *Config
	Category string
	Tag      string
}

func (cmd *EditCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("edit", "Interactively select a note with fzf and open it in the editor")
	cmd.cli.Flag("category", "Filter by category name with regular expression").Short('c').StringVar(&cmd.Category)
	cmd.cli.Flag("tag", "Filter by tag name with regular expression").Short('t').StringVar(&cmd.Tag)
}

func (cmd *EditCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// Do runs `notes edit` and returns an error if one occurs.
func (cmd *EditCmd) Do() error {
	notes, err := collectFilteredNotes(cmd.Config, cmd.Category, cmd.Tag)
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		return nil
	}

	selected, err := runFzf(cmd.Config, buildFzfInput(notes), fzfOptions{})
	if err != nil || len(selected) == 0 {
		return err
	}

	return openEditor(cmd.Config, relPathsToAbsPaths(cmd.Config, selected)...)
}
