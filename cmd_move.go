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

// MoveCmd represents `notes move` command.
// It opens a multi-select fzf picker, prompts for a destination category, then
// updates the category field in the frontmatter and moves each file.
type MoveCmd struct {
	cli      *kingpin.CmdClause
	Config   *Config
	Category string
	Tag      string
}

func (cmd *MoveCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("move", "Interactively select notes with fzf and move them to another category")
	cmd.cli.Flag("category", "Filter by category name with regular expression").Short('c').StringVar(&cmd.Category)
	cmd.cli.Flag("tag", "Filter by tag name with regular expression").Short('t').StringVar(&cmd.Tag)
}

func (cmd *MoveCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// Do runs `notes move` and returns an error if one occurs.
func (cmd *MoveCmd) Do() error {
	notes, err := collectFilteredNotes(cmd.Config, cmd.Category, cmd.Tag)
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		return nil
	}

	selected, err := runFzf(cmd.Config, buildFzfInput(notes), fzfOptions{
		Multi:  true,
		Header: "TAB to multi-select | ENTER to choose destination",
	})
	if err != nil || len(selected) == 0 {
		return err
	}

	fmt.Fprint(os.Stderr, "Move to category: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	destCat := strings.TrimSpace(scanner.Text())
	if destCat == "" {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		return nil
	}

	// Validate each segment of the destination category path.
	for _, part := range strings.Split(destCat, "/") {
		if err := validateDirname(part); err != nil {
			return errors.Wrapf(err, "invalid category name %q", part)
		}
	}

	destDir := filepath.Join(cmd.Config.HomePath, filepath.FromSlash(destCat))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.Wrapf(err, "cannot create category directory %q", destCat)
	}

	for _, rel := range selected {
		abs := relPathsToAbsPaths(cmd.Config, []string{rel})[0]
		base := filepath.Base(abs)
		dest := filepath.Join(destDir, base)

		if _, err := os.Stat(dest); err == nil {
			return errors.Errorf("cannot move: %q already exists in %q", base, destCat)
		}

		if err := updateCategoryInFile(abs, destCat); err != nil {
			return err
		}
		if err := os.Rename(abs, dest); err != nil {
			return errors.Wrapf(err, "cannot move %q", rel)
		}
		fmt.Printf("Moved: %s -> %s/%s\n", rel, destCat, base)
	}

	// Remove empty source directories (best-effort).
	for _, rel := range selected {
		srcDir := filepath.Dir(relPathsToAbsPaths(cmd.Config, []string{rel})[0])
		entries, err := os.ReadDir(srcDir)
		if err == nil && len(entries) == 0 && srcDir != cmd.Config.HomePath {
			os.Remove(srcDir) //nolint:errcheck
		}
	}

	return nil
}
