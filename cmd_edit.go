package notes

import ()

// EditCmd represents `notes edit` command.
// It opens an interactive fzf picker and opens the selected note in the editor.
type EditCmd struct {
	Config   *Config
	Category string
	Tag      string
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
