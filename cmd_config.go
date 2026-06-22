package notes

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"strings"
)

// ConfigCmd represents `notes config` command. Each public fields represent options of the command.
// Out field represents where this command should output.
type ConfigCmd struct {
	Config *Config
	// Name is a name of configuration. Must be one of "", "home", "git", "editor"
	Name string
	// Out is a writer to write output of this command. Kind of stdout is expected
	Out io.Writer
}

// Do runs `notes config` command and returns an error if occurs
func (cmd *ConfigCmd) Do() error {
	switch strings.ToLower(cmd.Name) {
	case "":
		fmt.Fprintf(
			cmd.Out,
			"HOME=%s\nGIT=%s\nEDITOR=%s\n",
			cmd.Config.HomePath,
			cmd.Config.GitPath,
			cmd.Config.EditorCmd,
		)
	case "home":
		fmt.Fprintln(cmd.Out, cmd.Config.HomePath)
	case "git":
		fmt.Fprintln(cmd.Out, cmd.Config.GitPath)
	case "editor":
		fmt.Fprintln(cmd.Out, cmd.Config.EditorCmd)
	default:
		return errors.Errorf("Unknown config name '%s'", cmd.Name)
	}
	return nil
}
