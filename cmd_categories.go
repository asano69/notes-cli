package notes

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// CategoriesCmd represents `notes categories` command. Each public fields represent options of the command.
// Out field represents where this command should output.
type CategoriesCmd struct {
	Config *Config
	// Out is a writer to write output of this command. Kind of stdout is expected
	Out io.Writer
}

// Do runs `notes categories` command and returns an error if occurs
func (cmd *CategoriesCmd) Do() error {
	cats, err := CollectCategories(cmd.Config, 0)
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
