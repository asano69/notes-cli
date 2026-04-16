package notes

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

// DeleteCmd represents `notes delete` command.
// It opens a multi-select fzf picker and deletes the selected notes after
// asking for confirmation (skipped with --yes).
type DeleteCmd struct {
	cli      *kingpin.CmdClause
	Config   *Config
	Category string
	Tag      string
	Yes      bool
}

func (cmd *DeleteCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("delete", "Interactively select notes with fzf and delete them")
	cmd.cli.Flag("category", "Filter by category name with regular expression").Short('c').StringVar(&cmd.Category)
	cmd.cli.Flag("tag", "Filter by tag name with regular expression").Short('t').StringVar(&cmd.Tag)
	cmd.cli.Flag("yes", "Skip confirmation prompt").Short('y').BoolVar(&cmd.Yes)
}

func (cmd *DeleteCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// Do runs `notes delete` and returns an error if one occurs.
func (cmd *DeleteCmd) Do() error {
	notes, err := collectFilteredNotes(cmd.Config, cmd.Category, cmd.Tag)
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		return nil
	}

	selected, err := runFzf(cmd.Config, buildFzfInput(notes), fzfOptions{
		Multi:  true,
		Header: "TAB to multi-select | ENTER to confirm",
	})
	if err != nil || len(selected) == 0 {
		return err
	}

	if !cmd.Yes {
		fmt.Fprintf(os.Stderr, "Delete %d note(s)? [y/N] ", len(selected))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if !strings.EqualFold(strings.TrimSpace(scanner.Text()), "y") {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return nil
		}
	}

	for _, rel := range selected {
		abs := relPathsToAbsPaths(cmd.Config, []string{rel})[0]
		if err := os.Remove(abs); err != nil {
			return errors.Wrapf(err, "cannot delete %q", rel)
		}
		fmt.Println("Deleted:", rel)
	}
	return nil
}
