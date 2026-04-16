package notes

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

// RenameCmd represents `notes rename` command.
// It opens a fzf picker, then prompts for a new filename for each selected note.
type RenameCmd struct {
	cli      *kingpin.CmdClause
	Config   *Config
	Category string
	Tag      string
}

func (cmd *RenameCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("rename", "Interactively select notes with fzf and rename their files")
	cmd.cli.Flag("category", "Filter by category name with regular expression").Short('c').StringVar(&cmd.Category)
	cmd.cli.Flag("tag", "Filter by tag name with regular expression").Short('t').StringVar(&cmd.Tag)
}

func (cmd *RenameCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// Do runs `notes rename` and returns an error if one occurs.
func (cmd *RenameCmd) Do() error {
	notes, err := collectFilteredNotes(cmd.Config, cmd.Category, cmd.Tag)
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		return nil
	}

	selected, err := runFzf(cmd.Config, buildFzfInput(notes), fzfOptions{
		Multi:  true,
		Header: "TAB to multi-select | ENTER to rename",
	})
	if err != nil || len(selected) == 0 {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)
	for _, rel := range selected {
		abs := relPathsToAbsPaths(cmd.Config, []string{rel})[0]

		fmt.Fprintf(os.Stderr, "Rename %s -> ", rel)
		scanner.Scan()
		newName := strings.TrimSpace(scanner.Text())
		if newName == "" {
			fmt.Fprintln(os.Stderr, "  (skipped)")
			continue
		}
		if !strings.HasSuffix(newName, ".md") {
			newName += ".md"
		}

		dest := filepath.Join(filepath.Dir(abs), newName)
		if _, err := os.Stat(dest); err == nil {
			return errors.Errorf("cannot rename: %q already exists", newName)
		}
		if err := os.Rename(abs, dest); err != nil {
			return errors.Wrapf(err, "cannot rename %q", rel)
		}
		fmt.Printf("Renamed: %s -> %s\n", rel, newName)
	}
	return nil
}
