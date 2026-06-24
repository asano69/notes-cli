package notes

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// SectionsCmd represents `notes sections` command. Each public fields represent options of the command.
// Out field represents where this command should output.
type SectionsCmd struct {
	Config *Config
	// Out is a writer to write output of this command. Kind of stdout is expected
	Out io.Writer
}

// Do runs `notes sections` command and returns an error if occurs
func (cmd *SectionsCmd) Do() error {
	cats, err := CollectSections(cmd.Config, 0)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(cats))
	for c := range cats {
		names = append(names, c)
	}

	sort.Strings(names)

	_, err = fmt.Fprintln(cmd.Out, strings.Join(names, "\n"))
	return err
}
